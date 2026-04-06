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

	code, err := h.oauthSvc.Authorize(r.Context(), clientID, redirectURI, scope, user.ID)
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

	var grantType, code, clientID, clientSecret, redirectURI string

	contentType := r.Header.Get("Content-Type")
	if strings.HasPrefix(contentType, "application/json") {
		var body struct {
			GrantType    string `json:"grant_type"`
			Code         string `json:"code"`
			ClientID     string `json:"client_id"`
			ClientSecret string `json:"client_secret"`
			RedirectURI  string `json:"redirect_uri"`
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
	} else {
		r.ParseForm()
		grantType = r.FormValue("grant_type")
		code = r.FormValue("code")
		clientID = r.FormValue("client_id")
		clientSecret = r.FormValue("client_secret")
		redirectURI = r.FormValue("redirect_uri")
	}

	if grantType != "authorization_code" {
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

	if code == "" || clientID == "" || clientSecret == "" {
		h.log.Warn("oauth token: missing params", "has_code", code != "", "has_client_id", clientID != "", "has_secret", clientSecret != "")
		writeError(w, http.StatusBadRequest, "code, client_id, and client_secret are required")
		return
	}

	accessToken, refreshToken, expiresIn, err := h.oauthSvc.Exchange(r.Context(), code, clientID, clientSecret, redirectURI)
	if err != nil {
		h.log.Error("oauth token exchange failed", "error", err, "client_id", clientID, "redirect_uri", redirectURI)
		if errors.Is(err, service.ErrInvalidClient) || errors.Is(err, service.ErrInvalidCode) || errors.Is(err, service.ErrInvalidRedirectURI) {
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
		ClientName   string   `json:"client_name"`
		RedirectURIs []string `json:"redirect_uris"`
		GrantTypes   []string `json:"grant_types"`
		ResponseTypes []string `json:"response_types"`
		TokenEndpointAuthMethod string `json:"token_endpoint_auth_method"`
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
		"client_id":                client.ID,
		"client_secret":            secret,
		"client_name":              client.Name,
		"redirect_uris":            client.RedirectURIs,
		"grant_types":              []string{"authorization_code"},
		"response_types":           []string{"code"},
		"token_endpoint_auth_method": "client_secret_post",
	})
}

// OAuthMetadataHandler handles GET /.well-known/oauth-authorization-server
func OAuthMetadataHandler(baseURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"issuer":                 baseURL,
			"authorization_endpoint": baseURL + "/auth/authorize",
			"token_endpoint":         baseURL + "/auth/token",
			"registration_endpoint":  baseURL + "/auth/register",
			"response_types_supported": []string{"code"},
			"grant_types_supported":    []string{"authorization_code"},
			"token_endpoint_auth_methods_supported": []string{"client_secret_post", "client_secret_basic"},
			"code_challenge_methods_supported": []string{"S256"},
			"scopes_supported": []string{"read", "write"},
		})
	}
}

// OAuthProtectedResourceHandler handles GET /.well-known/oauth-protected-resource
// Tells MCP clients which auth server protects the MCP endpoint.
func OAuthProtectedResourceHandler(baseURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"resource":               baseURL + "/sse",
			"authorization_servers":  []string{baseURL},
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
<title>Authorize — MCP Research</title>
<style>
  *, *::before, *::after { box-sizing: border-box; margin: 0; padding: 0; }
  body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; background: #f5f5f5; color: #1a1a1a; display: flex; justify-content: center; align-items: center; min-height: 100vh; }
  .card { background: #fff; border: 1px solid #e5e5e5; border-radius: 12px; padding: 2rem; width: 100%; max-width: 400px; }
  h1 { font-size: 1.5rem; font-weight: 600; margin-bottom: .25rem; }
  .subtitle { color: #666; margin-bottom: 1.5rem; font-size: .9rem; }
  .error { background: #fef2f2; color: #dc2626; padding: .5rem .75rem; border-radius: 8px; font-size: .85rem; margin-bottom: 1rem; }
  label { display: block; font-size: .85rem; font-weight: 500; margin-bottom: .25rem; }
  input { width: 100%; padding: .5rem .75rem; border: 1px solid #d1d5db; border-radius: 8px; font-size: .95rem; margin-bottom: 1rem; font-family: inherit; }
  input:focus { outline: 2px solid #2563eb; outline-offset: -1px; }
  button { width: 100%; padding: .6rem 1rem; background: #2563eb; color: #fff; border: none; border-radius: 8px; font-size: .95rem; font-weight: 500; cursor: pointer; font-family: inherit; }
  button:hover { background: #1d4ed8; }
  .info { margin-top: 1rem; font-size: .8rem; color: #999; text-align: center; }
</style>
</head>
<body>
<div class="card">
  <h1>Sign in</h1>
  <p class="subtitle">Authorize access to your MCP Research account</p>
  {{if .Error}}<div class="error">{{.Error}}</div>{{end}}
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

