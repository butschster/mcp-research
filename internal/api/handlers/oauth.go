package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"net/url"
	"strings"

	"github.com/butschster/mcp-research/internal/service"
)

type OAuthHandler struct {
	oauthSvc *service.OAuthService
	authSvc  *service.AuthService
	log      *slog.Logger
}

func NewOAuthHandler(oauthSvc *service.OAuthService, authSvc *service.AuthService, log *slog.Logger) *OAuthHandler {
	return &OAuthHandler{oauthSvc: oauthSvc, authSvc: authSvc, log: log}
}

// Authorize handles the OAuth2 authorization endpoint.
// GET  — renders login + consent HTML page
// POST — validates credentials, creates code, redirects to client
func (h *OAuthHandler) Authorize(w http.ResponseWriter, r *http.Request) {
	h.log.Info("oauth authorize", "method", r.Method, "url", r.URL.String(), "remote", r.RemoteAddr)

	clientID := r.URL.Query().Get("client_id")
	redirectURI := r.URL.Query().Get("redirect_uri")
	scope := r.URL.Query().Get("scope")
	state := r.URL.Query().Get("state")
	responseType := r.URL.Query().Get("response_type")
	codeChallenge := r.URL.Query().Get("code_challenge")
	codeChallengeMethod := r.URL.Query().Get("code_challenge_method")

	if clientID == "" || redirectURI == "" {
		h.renderAuthorizePage(w, clientID, redirectURI, scope, state, "client_id and redirect_uri are required")
		return
	}

	if responseType != "" && responseType != "code" {
		h.renderAuthorizePage(w, clientID, redirectURI, scope, state, "unsupported response_type (use 'code')")
		return
	}

	// Validate client exists and redirect_uri matches
	clientName, err := h.oauthSvc.ValidateClient(r.Context(), clientID, redirectURI)
	if err != nil {
		h.renderAuthorizePage(w, clientID, redirectURI, scope, state, err.Error())
		return
	}

	if r.Method == http.MethodGet {
		h.renderAuthorizePage(w, clientID, redirectURI, scope, state, "")
		return
	}

	// POST — process login + approve
	email := r.FormValue("email")
	password := r.FormValue("password")

	if email == "" || password == "" {
		h.renderAuthorizePage(w, clientID, redirectURI, scope, state, "Email and password are required")
		return
	}

	user, _, err := h.authSvc.Login(r.Context(), email, password)
	if err != nil {
		h.renderAuthorizePage(w, clientID, redirectURI, scope, state, "Invalid email or password")
		return
	}

	code, err := h.oauthSvc.Authorize(r.Context(), clientID, redirectURI, scope, user.ID, codeChallenge, codeChallengeMethod)
	if err != nil {
		h.log.Error("oauth authorize failed", "error", err)
		h.renderAuthorizePage(w, clientID, redirectURI, scope, state, "Authorization failed")
		return
	}

	// Redirect back to client with code
	u, _ := url.Parse(redirectURI)
	q := u.Query()
	q.Set("code", code)
	if state != "" {
		q.Set("state", state)
	}
	u.RawQuery = q.Encode()

	_ = clientName // used in template rendering
	http.Redirect(w, r, u.String(), http.StatusFound)
}

// Token handles POST /auth/token — exchanges code for tokens.
// Accepts both application/x-www-form-urlencoded and application/json.
func (h *OAuthHandler) Token(w http.ResponseWriter, r *http.Request) {
	h.log.Info("oauth token request", "method", r.Method, "content_type", r.Header.Get("Content-Type"), "remote", r.RemoteAddr, "has_basic_auth", r.Header.Get("Authorization") != "")

	var grantType, code, clientID, clientSecret, redirectURI, codeVerifier, refreshTokenIn string

	contentType := r.Header.Get("Content-Type")
	if strings.HasPrefix(contentType, "application/json") {
		var body struct {
			GrantType    string `json:"grant_type"`
			Code         string `json:"code"`
			ClientID     string `json:"client_id"`
			ClientSecret string `json:"client_secret"`
			RedirectURI  string `json:"redirect_uri"`
			CodeVerifier string `json:"code_verifier"`
			RefreshToken string `json:"refresh_token"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON")
			return
		}
		grantType = body.GrantType
		code = body.Code
		clientID = body.ClientID
		clientSecret = body.ClientSecret
		redirectURI = body.RedirectURI
		codeVerifier = body.CodeVerifier
		refreshTokenIn = body.RefreshToken
	} else {
		r.ParseForm()
		grantType = r.FormValue("grant_type")
		code = r.FormValue("code")
		clientID = r.FormValue("client_id")
		clientSecret = r.FormValue("client_secret")
		redirectURI = r.FormValue("redirect_uri")
		codeVerifier = r.FormValue("code_verifier")
		refreshTokenIn = r.FormValue("refresh_token")
	}

	if grantType != "authorization_code" && grantType != "refresh_token" {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("unsupported grant_type: %s", grantType))
		return
	}

	// Also support Basic auth for client credentials
	if clientID == "" || clientSecret == "" {
		if u, p, ok := r.BasicAuth(); ok {
			clientID = u
			clientSecret = p
		}
	}

	h.log.Info("oauth token params", "grant_type", grantType, "client_id", clientID, "has_code", code != "", "has_secret", clientSecret != "", "redirect_uri", redirectURI)

	if clientID == "" || clientSecret == "" {
		h.log.Warn("oauth token: missing client credentials", "has_client_id", clientID != "", "has_secret", clientSecret != "")
		writeError(w, http.StatusBadRequest, "client_id and client_secret are required")
		return
	}

	var (
		accessToken  string
		refreshToken string
		expiresIn    int
		err          error
	)
	if grantType == "refresh_token" {
		if refreshTokenIn == "" {
			writeError(w, http.StatusBadRequest, "refresh_token is required")
			return
		}
		accessToken, refreshToken, expiresIn, err = h.oauthSvc.Refresh(r.Context(), refreshTokenIn, clientID, clientSecret)
	} else {
		if code == "" {
			h.log.Warn("oauth token: missing code")
			writeError(w, http.StatusBadRequest, "code is required")
			return
		}
		accessToken, refreshToken, expiresIn, err = h.oauthSvc.Exchange(r.Context(), code, clientID, clientSecret, redirectURI, codeVerifier)
	}
	if err != nil {
		h.log.Error("oauth token request failed", "error", err, "grant_type", grantType, "client_id", clientID, "redirect_uri", redirectURI)
		if errors.Is(err, service.ErrInvalidClient) || errors.Is(err, service.ErrInvalidCode) ||
			errors.Is(err, service.ErrInvalidRedirectURI) || errors.Is(err, service.ErrInvalidRefresh) {
			writeError(w, http.StatusBadRequest, err.Error())
		} else {
			writeError(w, http.StatusInternalServerError, "token exchange failed")
		}
		return
	}

	h.log.Info("oauth token exchange success", "client_id", clientID)

	w.Header().Set("Cache-Control", "no-store")
	writeJSON(w, http.StatusOK, map[string]any{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"token_type":    "Bearer",
		"expires_in":    expiresIn,
	})
}

// RegisterClient handles POST /auth/register — RFC 7591 Dynamic Client Registration.
// ChatGPT and other MCP clients use this to auto-create OAuth clients.
func (h *OAuthHandler) RegisterClient(w http.ResponseWriter, r *http.Request) {
	h.log.Info("oauth DCR register", "method", r.Method, "remote", r.RemoteAddr)

	var input struct {
		ClientName              string   `json:"client_name"`
		RedirectURIs            []string `json:"redirect_uris"`
		GrantTypes              []string `json:"grant_types"`
		ResponseTypes           []string `json:"response_types"`
		TokenEndpointAuthMethod string   `json:"token_endpoint_auth_method"`
	}
	if !decodeJSON(w, r, &input) {
		return
	}

	name := input.ClientName
	if name == "" {
		name = "OAuth Client"
	}
	if len(input.RedirectURIs) == 0 {
		writeError(w, http.StatusBadRequest, "redirect_uris is required")
		return
	}

	client, secret, err := h.oauthSvc.RegisterClient(r.Context(), name, input.RedirectURIs)
	if err != nil {
		h.log.Error("DCR registration failed", "error", err)
		writeError(w, http.StatusInternalServerError, "client registration failed")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"client_id":                  client.ID,
		"client_secret":              secret,
		"client_name":                client.Name,
		"redirect_uris":              client.RedirectURIs,
		"grant_types":                []string{"authorization_code", "refresh_token"},
		"response_types":             []string{"code"},
		"token_endpoint_auth_method": "client_secret_post",
	})
}

// OAuthMetadataHandler handles GET /.well-known/oauth-authorization-server
func OAuthMetadataHandler(baseURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"issuer":                                baseURL,
			"authorization_endpoint":                baseURL + "/auth/authorize",
			"token_endpoint":                        baseURL + "/auth/token",
			"registration_endpoint":                 baseURL + "/auth/register",
			"response_types_supported":              []string{"code"},
			"grant_types_supported":                 []string{"authorization_code", "refresh_token"},
			"token_endpoint_auth_methods_supported": []string{"client_secret_post", "client_secret_basic"},
			"code_challenge_methods_supported":      []string{"S256"},
			"scopes_supported":                      []string{"read", "write"},
		})
	}
}

// OAuthProtectedResourceHandler handles GET /.well-known/oauth-protected-resource
// Tells MCP clients which auth server protects the MCP endpoint.
func OAuthProtectedResourceHandler(baseURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"resource":              baseURL + "/mcp",
			"authorization_servers": []string{baseURL},
			"scopes_supported":      []string{"read", "write"},
		})
	}
}

type authorizePageData struct {
	ClientID    string
	RedirectURI string
	Scope       string
	State       string
	Error       string
}

func (h *OAuthHandler) renderAuthorizePage(w http.ResponseWriter, clientID, redirectURI, scope, state, errMsg string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if errMsg != "" && (clientID == "" || redirectURI == "") {
		w.WriteHeader(http.StatusBadRequest)
	}

	data := authorizePageData{
		ClientID:    clientID,
		RedirectURI: redirectURI,
		Scope:       scope,
		State:       state,
		Error:       errMsg,
	}

	authorizeTmpl.Execute(w, data)
}

var authorizeTmpl = template.Must(template.New("authorize").Parse(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>Authorize — Dovod</title>
<script>try { document.documentElement.dataset.theme = localStorage.getItem('dovod-theme') === 'dark' ? 'dark' : 'light'; } catch {}</script>
<style>
  :root { color-scheme: light; --bg: #f6f3ec; --surface: #fffdf8; --ink: #243d32; --muted: #606c5a; --primary: #243d32; --on-primary: #f6f3ec; --hover: #35523c; --border: #a7b29e; --line: #d9ded0; --error: #a63d32; --error-bg: #f8eae3; --error-line: #d8b4ab; }
  :root[data-theme="dark"] { color-scheme: dark; --bg: #0d1117; --surface: #151b23; --ink: #f0f6fc; --muted: #9198a1; --primary: #c7d9a7; --on-primary: #0d1117; --hover: #d8e7bf; --border: #3d444d; --line: #30363d; --error: #ef8880; --error-bg: #2a1b20; --error-line: #704046; }
  *, *::before, *::after { box-sizing: border-box; margin: 0; padding: 0; }
  body { font-family: 'Segoe UI', sans-serif; background: var(--bg); color: var(--ink); display: flex; justify-content: center; align-items: center; min-height: 100dvh; padding: 1.5rem; }
  .card { width: 100%; max-width: 420px; }
  .brand { display: block; width: 128px; color: var(--ink); margin-bottom: 3rem; }
  .brand svg { display: block; width: 100%; height: auto; }
  h1 { font-size: 2rem; font-weight: 700; letter-spacing: -.04em; margin-bottom: .75rem; }
  .subtitle { color: var(--muted); margin-bottom: 2rem; font-size: 1rem; line-height: 1.6; }
  .error { background: var(--error-bg); color: var(--error); border: 1px solid var(--error-line); padding: .75rem 1rem; border-radius: 6px; font-size: .9rem; line-height: 1.6; margin-bottom: 1.25rem; }
  label { display: block; font-size: .9rem; font-weight: 500; margin-bottom: .5rem; }
  input { width: 100%; min-height: 52px; padding: .75rem 1rem; background: var(--surface); color: var(--ink); border: 1px solid var(--border); border-radius: 6px; font-size: 1rem; margin-bottom: 1.25rem; font-family: inherit; }
  :focus-visible { outline: 2px solid var(--ink); outline-offset: 3px; }
  input:focus { outline: 2px solid var(--ink); outline-offset: -1px; }
  button { min-height: 48px; padding: .75rem 1.5rem; background: var(--primary); color: var(--on-primary); border: none; border-radius: 6px; font-size: 1rem; font-weight: 500; cursor: pointer; font-family: inherit; }
  button:hover { background: var(--hover); }
  .info { margin-top: 1.5rem; padding-top: 1.5rem; border-top: 1px solid var(--line); font-size: .85rem; color: var(--muted); line-height: 1.6; }
</style>
</head>
<body>
<div class="card">
  <a href="/" class="brand" aria-label="Dovod home"><svg xmlns="http://www.w3.org/2000/svg" viewBox="-1 -1 378.19048 102" fill="currentColor" role="img" aria-label="Dovod"><g transform="translate(-3.08123 98.03922) scale(0.1400560 -0.1400560)"><path d="M223 -14Q166 -14 120.5 17Q75 48 48.5 111Q22 174 22 270Q22 361 46 421.5Q70 482 114 512Q158 542 217 542Q263 542 298.5 523Q334 504 357 463Q380 422 389 356H398Q393 390 388 423Q383 456 380.5 487Q378 518 378 545V700H540V186V0H399V163H392Q381 105 358.5 65.5Q336 26 302.5 6Q269 -14 223 -14ZM280 115Q299 115 317 123.5Q335 132 349 150Q363 168 371.5 195Q380 222 380 259V266Q380 296 374 319.5Q368 343 358 359.5Q348 376 335 386.5Q322 397 307.5 402Q293 407 279 407Q248 407 227 386Q206 365 196 331.5Q186 298 186 259Q186 220 197 187Q208 154 229 134.5Q250 115 280 115Z"/><path d="M840 -14Q764 -14 705 16.5Q646 47 613 108.5Q580 170 580 263Q580 360 614.5 421.5Q649 483 708.5 512.5Q768 542 843 542Q921 542 979.5 511.5Q1038 481 1070.5 419.5Q1103 358 1103 264Q1103 167 1068.5 105.5Q1034 44 974.5 15Q915 -14 840 -14ZM844 105Q875 105 896 122.5Q917 140 928 174Q939 208 939 255Q939 307 927 342.5Q915 378 892.5 397Q870 416 838 416Q808 416 787 398.5Q766 381 755 347Q744 313 744 265Q744 189 770.5 147Q797 105 844 105Z"/><path d="M1269 0 1110 528H1282L1373 134H1377L1472 528H1634L1475 0Z"/><path d="M1901 -14Q1825 -14 1766 16.5Q1707 47 1674 108.5Q1641 170 1641 263Q1641 360 1675.5 421.5Q1710 483 1769.5 512.5Q1829 542 1904 542Q1982 542 2040.5 511.5Q2099 481 2131.5 419.5Q2164 358 2164 264Q2164 167 2129.5 105.5Q2095 44 2035.5 15Q1976 -14 1901 -14ZM1905 105Q1936 105 1957 122.5Q1978 140 1989 174Q2000 208 2000 255Q2000 307 1988 342.5Q1976 378 1953.5 397Q1931 416 1899 416Q1869 416 1848 398.5Q1827 381 1816 347Q1805 313 1805 265Q1805 189 1831.5 147Q1858 105 1905 105Z"/><path d="M2391 -14Q2334 -14 2288.5 17Q2243 48 2216.5 111Q2190 174 2190 270Q2190 361 2214 421.5Q2238 482 2282 512Q2326 542 2385 542Q2431 542 2466.5 523Q2502 504 2525 463Q2548 422 2557 356H2566Q2561 390 2556 423Q2551 456 2548.5 487Q2546 518 2546 545V700H2708V186V0H2567V163H2560Q2549 105 2526.5 65.5Q2504 26 2470.5 6Q2437 -14 2391 -14ZM2448 115Q2467 115 2485 123.5Q2503 132 2517 150Q2531 168 2539.5 195Q2548 222 2548 259V266Q2548 296 2542 319.5Q2536 343 2526 359.5Q2516 376 2503 386.5Q2490 397 2475.5 402Q2461 407 2447 407Q2416 407 2395 386Q2374 365 2364 331.5Q2354 298 2354 259Q2354 220 2365 187Q2376 154 2397 134.5Q2418 115 2448 115Z"/></g></svg></a>
  <h1>Connect to Dovod</h1>
  <p class="subtitle">Sign in to authorize access to your Dovod account.</p>
  {{if .Error}}<div class="error" role="alert">{{.Error}}</div>{{end}}
  <form method="POST">
    <label for="email">Email</label>
    <input type="email" id="email" name="email" required autocomplete="email" placeholder="you@example.com">
    <label for="password">Password</label>
    <input type="password" id="password" name="password" required autocomplete="current-password" placeholder="Your password">
    <button type="submit">Sign in &amp; Authorize</button>
  </form>
  <p class="info">You will be redirected back to the application.</p>
</div>
</body>
</html>`))
