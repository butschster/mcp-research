package handlers

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/butschster/mcp-research/internal/auth"
	"github.com/butschster/mcp-research/internal/domain"
	"github.com/butschster/mcp-research/internal/service"
	"github.com/butschster/mcp-research/internal/storage"
	"github.com/butschster/mcp-research/internal/testdb"
)

type authContractValidator struct{ svc *service.AuthService }

func (v authContractValidator) ValidateToken(r *http.Request) (*domain.User, error) {
	return v.svc.ValidateToken(r.Context(), strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "))
}

func authContractMux(t *testing.T, allowRegistration bool) (*http.ServeMux, *service.AuthService) {
	t.Helper()
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	db, err := storage.NewDB(testdb.Config(t), log)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	svc := service.NewAuthService(storage.NewUserRepository(db), storage.NewAPIKeyRepository(db), storage.NewOAuthRepository(db), storage.NewResearchRepository(db), storage.NewTeamRepository(db), auth.NewJWTManager("test-only-secret", time.Hour), allowRegistration, log)
	h := NewAuthHandler(svc, nil, "", log)
	mux := http.NewServeMux()
	mux.HandleFunc("POST /register", h.Register)
	mux.HandleFunc("POST /login", h.Login)
	mux.HandleFunc("GET /info", h.AuthInfo)
	// Leave the handlers' defensive 401 checks reachable too. With a valid
	// bearer token this is the real service and database authentication path.
	middleware := auth.OptionalAuth(authContractValidator{svc})
	for pattern, handler := range map[string]http.HandlerFunc{"GET /me": h.Me, "POST /keys": h.CreateAPIKey, "GET /keys": h.ListAPIKeys, "DELETE /keys/{id}": h.DeleteAPIKey} {
		mux.Handle(pattern, middleware(handler))
	}
	return mux, svc
}

func authContractRequest(t *testing.T, mux http.Handler, method, path, body, token string, want int) *httptest.ResponseRecorder {
	t.Helper()
	r := httptest.NewRequest(method, path, strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	if token != "" {
		r.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, r)
	if w.Code != want {
		t.Fatalf("%s %s: status %d want %d: %s", method, path, w.Code, want, w.Body.String())
	}
	return w
}

func TestAuthContract_RegistrationLoginAndKeyRevocation(t *testing.T) {
	mux, svc := authContractMux(t, true)
	register := func(email string) (domain.User, string) {
		body, _ := json.Marshal(map[string]string{"email": email, "password": "long-enough", "name": "O'Reilly"})
		w := authContractRequest(t, mux, "POST", "/register", string(body), "", 201)
		var result struct {
			User  domain.User `json:"user"`
			Token string      `json:"token"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
			t.Fatal(err)
		}
		if result.User.ID == "" || result.Token == "" || strings.Contains(w.Body.String(), "password_hash") {
			t.Fatalf("registration payload: %s", w.Body.String())
		}
		return result.User, result.Token
	}
	owner, ownerToken := register(" OWNER@example.invalid ")
	_, otherToken := register("other@example.invalid")
	if owner.Email != "owner@example.invalid" {
		t.Fatalf("email not normalized: %q", owner.Email)
	}
	authContractRequest(t, mux, "POST", "/register", `{"email":"owner@example.invalid","password":"long-enough"}`, "", 409)
	for _, body := range []string{`{`, `{"email":"bad","password":"long-enough"}`, `{"email":"new@example.invalid","password":"x"}`} {
		authContractRequest(t, mux, "POST", "/register", body, "", 400)
	}
	authContractRequest(t, mux, "POST", "/login", `{"email":"owner@example.invalid","password":"wrong"}`, "", 401)
	authContractRequest(t, mux, "POST", "/login", `{"email":"missing@example.invalid","password":"long-enough"}`, "", 401)
	authContractRequest(t, mux, "POST", "/login", `{`, "", 400)
	w := authContractRequest(t, mux, "POST", "/login", `{"email":"owner@example.invalid","password":"long-enough"}`, "", 200)
	var login struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &login); err != nil || login.Token == "" {
		t.Fatalf("login response: %s %v", w.Body.String(), err)
	}
	w = authContractRequest(t, mux, "GET", "/me", "", login.Token, 200)
	var me domain.User
	if err := json.Unmarshal(w.Body.Bytes(), &me); err != nil || me.ID != owner.ID {
		t.Fatalf("authenticated identity: %+v %v", me, err)
	}
	for _, tc := range []struct{ method, path string }{{"GET", "/me"}, {"POST", "/keys"}, {"GET", "/keys"}, {"DELETE", "/keys/missing"}} {
		for _, token := range []string{"", "not-a-token"} {
			authContractRequest(t, mux, tc.method, tc.path, `{}`, token, 401)
		}
	}
	authContractRequest(t, mux, "POST", "/keys", `{`, ownerToken, 400)
	w = authContractRequest(t, mux, "GET", "/keys", "", ownerToken, 200)
	if strings.TrimSpace(w.Body.String()) != "[]" {
		t.Fatalf("empty keys must be array: %s", w.Body.String())
	}
	w = authContractRequest(t, mux, "POST", "/keys", `{"name":"integration"}`, ownerToken, 201)
	var key struct {
		Key  string `json:"key"`
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &key); err != nil || key.Key == "" || key.ID == "" || key.Name != "integration" {
		t.Fatalf("key response: %s %v", w.Body.String(), err)
	}
	authContractRequest(t, mux, "GET", "/me", "", key.Key, 200)
	w = authContractRequest(t, mux, "GET", "/keys", "", ownerToken, 200)
	var keys []domain.APIKey
	if err := json.Unmarshal(w.Body.Bytes(), &keys); err != nil || len(keys) != 1 || keys[0].ID != key.ID {
		t.Fatalf("key list: %+v %v", keys, err)
	}
	if strings.Contains(w.Body.String(), key.Key) || strings.Contains(w.Body.String(), "key_hash") {
		t.Fatal("key list leaked credential")
	}
	// Last-used is intentionally asynchronous; wait for persistence instead
	// of depending on the relative scheduling of the handler and its worker.
	deadline := time.Now().Add(3 * time.Second)
	for {
		keys, err := svc.ListAPIKeys(t.Context(), owner.ID)
		if err != nil {
			t.Fatal(err)
		}
		if len(keys) == 1 && keys[0].LastUsedAt != nil {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("last-used timestamp was not persisted")
		}
		time.Sleep(10 * time.Millisecond)
	}
	w = authContractRequest(t, mux, "GET", "/keys", "", otherToken, 200)
	if strings.TrimSpace(w.Body.String()) != "[]" {
		t.Fatalf("other user sees keys: %s", w.Body.String())
	}
	authContractRequest(t, mux, "DELETE", "/keys/"+key.ID, "", otherToken, 404)
	authContractRequest(t, mux, "GET", "/me", "", key.Key, 200)
	authContractRequest(t, mux, "DELETE", "/keys/"+key.ID, "", ownerToken, 200)
	authContractRequest(t, mux, "GET", "/me", "", key.Key, 401)
	authContractRequest(t, mux, "DELETE", "/keys/"+key.ID, "", ownerToken, 404)
	if user, err := svc.ValidateToken(context.Background(), key.Key); err != nil || user != nil {
		t.Fatalf("revoked key resolves user: %+v %v", user, err)
	}
}

func TestAuthContract_RegistrationClosedAndInfo(t *testing.T) {
	mux, _ := authContractMux(t, false)
	authContractRequest(t, mux, "POST", "/register", `{"email":"new@example.invalid","password":"long-enough"}`, "", 403)
	w := authContractRequest(t, mux, "GET", "/info", "", "", 200)
	var info map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &info); err != nil {
		t.Fatal(err)
	}
	if info["auth_enabled"] != true || info["allow_registration"] != false || info["auto_login_token"] != nil {
		t.Fatalf("auth settings: %v", info)
	}
}
