package ws

import (
	"context"
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/gorilla/websocket"
)

func newUpgrader(baseURL string) *websocket.Upgrader {
	return &websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return allowOrigin(baseURL, r) },
	}
}

// TokenValidator resolves a bearer token to a user id. It is the same
// credential the REST API takes — a JWT, an API key or an OAuth token.
//
// The hub keeps hold of one because the handshake is not the last time the
// question matters: a token can be deleted or expire while the socket it opened
// is still connected.
type TokenValidator interface {
	// UserIDForToken resolves a credential, or returns "" for one it rejects.
	UserIDForToken(ctx context.Context, token string) string
	// ValidateCredential is the same question with "no" and "could not tell"
	// distinguished. A live connection is re-checked on a timer against a
	// one-connection database, so a failed lookup must not read as a
	// revocation.
	ValidateCredential(ctx context.Context, token string) (userID string, decided bool)
}

// HandleWebSocket upgrades to a WebSocket and registers the connection.
//
// With auth enabled the connection must carry a token, and it arrives as a
// query parameter rather than a header: browsers cannot set headers on a
// WebSocket handshake. That is the same accommodation the SSE transport makes.
func HandleWebSocket(hub *Hub, baseURL string) http.HandlerFunc {
	upgrader := newUpgrader(baseURL)

	return func(w http.ResponseWriter, r *http.Request) {
		userID, token := "", ""
		if hub.AuthEnabled() {
			validator := hub.TokenValidator()
			token = bearerToken(r)
			if token == "" || validator == nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			if userID = validator.UserIDForToken(r.Context(), token); userID == "" {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}

		client := NewClient(hub, conn, userID, token)
		hub.Register(client)

		go client.WritePump()
		go client.ReadPump()
	}
}

func bearerToken(r *http.Request) string {
	if h := r.Header.Get("Authorization"); len(h) > 7 && h[:7] == "Bearer " {
		return h[7:]
	}
	return r.URL.Query().Get("token")
}

// allowOrigin decides whether a handshake may come from where it says it does.
//
// This endpoint used to accept every origin, which with auth off — the default
// for a local run — let any page the user happened to visit open a socket to
// their server and read the whole event stream. The browser attaches Origin
// itself and a page cannot forge it, so checking it is the whole defence.
func allowOrigin(baseURL string, r *http.Request) bool {
	origin := r.Header.Get("Origin")
	if origin == "" {
		// Not a browser. MCP clients, curl and the test suite send no Origin,
		// and none of them are the thing this check exists to stop.
		return true
	}

	u, err := url.Parse(origin)
	if err != nil || u.Host == "" {
		return false
	}
	if strings.EqualFold(u.Host, r.Host) {
		return true
	}
	// The Nuxt dev server runs on another port against this one, so a loopback
	// origin is allowed — but only when this server is itself reached over
	// loopback. On a public host, `localhost:3000` says nothing reassuring: it
	// is the default of nearly every JS dev server, and with auth off the
	// origin check is the whole defence.
	if isLoopback(u.Hostname()) && isLoopback(hostname(r.Host)) {
		return true
	}
	if baseURL != "" {
		if b, err := url.Parse(baseURL); err == nil && b.Host != "" && strings.EqualFold(b.Host, u.Host) {
			return true
		}
	}
	return false
}

func isLoopback(host string) bool {
	return host == "localhost" || host == "127.0.0.1" || host == "::1"
}

// hostname strips the port from a Host header, which unlike a URL's Hostname()
// has no parsed form to ask.
func hostname(host string) string {
	if h, _, err := net.SplitHostPort(host); err == nil {
		return h
	}
	return strings.Trim(host, "[]")
}
