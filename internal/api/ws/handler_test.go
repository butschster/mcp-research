package ws

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

type fakeValidator struct {
	valid map[string]string // token -> user id
	// broken makes every lookup fail to reach an answer, which is a different
	// thing from answering "no".
	broken bool
}

func (f fakeValidator) UserIDForToken(_ context.Context, token string) string {
	if f.broken {
		return ""
	}
	return f.valid[token]
}

func (f fakeValidator) ValidateCredential(_ context.Context, token string) (string, bool) {
	if f.broken {
		return "", false
	}
	return f.valid[token], true
}

func quietHub() *Hub { return NewHub(slog.New(slog.NewTextHandler(io.Discard, nil))) }

// The handshake is where a credential is demanded. It used to demand nothing:
// the endpoint accepted every connection and the hub sent every event to it.
func TestHandshake_RejectsWithoutAValidCredential(t *testing.T) {
	hub := quietHub()
	hub.SetAuthorizer(fakeAuth{}, true)
	hub.SetTokenValidator(fakeValidator{valid: map[string]string{"good": "alice"}})
	handler := HandleWebSocket(hub, "")

	cases := map[string]string{
		"no token":      "",
		"unknown token": "?token=nope",
		"empty token":   "?token=",
	}
	for name, query := range cases {
		rec := httptest.NewRecorder()
		handler(rec, httptest.NewRequest(http.MethodGet, "/ws"+query, nil))
		if rec.Code != http.StatusUnauthorized {
			t.Errorf("%s: status %d, want %d", name, rec.Code, http.StatusUnauthorized)
		}
	}
}

// A hub with auth on and no validator wired must refuse rather than wave people
// through — the same fail-closed direction the delivery rule takes.
func TestHandshake_RejectsWhenNoValidatorIsConfigured(t *testing.T) {
	hub := quietHub()
	hub.SetAuthorizer(fakeAuth{}, true)

	rec := httptest.NewRecorder()
	HandleWebSocket(hub, "")(rec, httptest.NewRequest(http.MethodGet, "/ws?token=good", nil))
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

// With auth off there is no credential to ask for, and this is the mode the
// local single-binary run has always used.
func TestHandshake_LocalModeNeedsNoToken(t *testing.T) {
	hub := quietHub()
	rec := httptest.NewRecorder()
	HandleWebSocket(hub, "")(rec, httptest.NewRequest(http.MethodGet, "/ws", nil))
	// httptest's ResponseRecorder cannot be hijacked, so the upgrade itself
	// fails — but not with 401, which is the thing being asserted.
	if rec.Code == http.StatusUnauthorized {
		t.Error("local mode demanded a credential")
	}
}

func TestBearerToken_ReadsHeaderThenQuery(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/ws?token=from-query", nil)
	if got := bearerToken(r); got != "from-query" {
		t.Errorf("query: got %q", got)
	}

	r.Header.Set("Authorization", "Bearer from-header")
	if got := bearerToken(r); got != "from-header" {
		t.Errorf("header should win: got %q", got)
	}
}

// This endpoint used to accept every origin. With auth off — the default for a
// local run — that let any page the user happened to visit open a socket to
// their server and read the whole event stream.
func TestAllowOrigin(t *testing.T) {
	cases := []struct {
		name    string
		host    string
		baseURL string
		origin  string
		want    bool
	}{
		{"no origin at all (curl, MCP client)", "research.example.com", "", "", true},
		{"same origin as the request", "research.example.com", "", "http://research.example.com", true},
		{"the configured public URL", "research.example.com", "https://mcp.example.com", "https://mcp.example.com", true},
		{"a page on the open internet", "research.example.com", "", "https://evil.example", false},
		{"a lookalike host", "research.example.com", "https://mcp.example.com", "https://mcp.example.com.evil.example", false},
		{"garbage", "research.example.com", "", "://", false},

		// The Nuxt dev server runs on another port against a local binary.
		{"the dev server, talking to a local binary", "localhost:8088", "", "http://localhost:3000", true},
		{"loopback by address, locally", "127.0.0.1:8088", "", "http://127.0.0.1:3000", true},

		// But on a public host a loopback origin is evidence of nothing:
		// localhost:3000 is the default of nearly every JS dev server, and with
		// auth off this check is the only thing standing in the way.
		{"a local dev server aimed at a public host", "research.example.com", "", "http://localhost:3000", false},
		{"the same by address", "research.example.com", "", "http://127.0.0.1:5173", false},
	}
	for _, tc := range cases {
		r := httptest.NewRequest(http.MethodGet, "/ws", nil)
		r.Host = tc.host
		if tc.origin != "" {
			r.Header.Set("Origin", tc.origin)
		}
		if got := allowOrigin(tc.baseURL, r); got != tc.want {
			t.Errorf("%s: allowed=%v, want %v", tc.name, got, tc.want)
		}
	}
}
