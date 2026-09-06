package service

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/dovod-app/app/internal/auth"
	"github.com/dovod-app/app/internal/storage"
	"github.com/uptrace/bun"
)

// The HTTP-level flow lives in internal/api/oauth_routes_test.go. What is here
// is the handful of decisions a request cannot reach through the mux: a code
// that has aged out, two exchanges racing on one code, and a token whose
// lifetime has to be read out of the database to be believed.

func newOAuthFixture(t *testing.T) (*OAuthService, *storage.OAuthRepository, *bun.DB, string) {
	t.Helper()
	db := setupTestDB(t)
	repo := storage.NewOAuthRepository(db)
	svc := NewOAuthService(repo, slog.New(slog.NewTextHandler(io.Discard, nil)))
	user := createTestUser(t, db, "oauth@test.com", "Person")
	return svc, repo, db, user.ID
}

const fixtureRedirect = "https://client.example.com/callback"

func challengeFor(verifier string) string {
	h := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(h[:])
}

func TestOAuthService_ValidateClient(t *testing.T) {
	svc, _, _, _ := newOAuthFixture(t)
	ctx := context.Background()
	client, _, err := svc.RegisterClient(ctx, "ChatGPT", []string{fixtureRedirect})
	if err != nil {
		t.Fatal(err)
	}

	if name, err := svc.ValidateClient(ctx, client.ID, fixtureRedirect); err != nil || name != "ChatGPT" {
		t.Fatalf("valid client: name %q err %v", name, err)
	}
	if _, err := svc.ValidateClient(ctx, "no-such-client", fixtureRedirect); !errors.Is(err, ErrInvalidClient) {
		t.Fatalf("unknown client: %v", err)
	}
	// A prefix of a registered URI is not a registered URI — the check is
	// equality, and a "starts with" would be an open redirect.
	for _, uri := range []string{
		fixtureRedirect + "/../evil",
		fixtureRedirect + "?next=https://evil.example.com",
		"https://client.example.com",
		"https://client.example.com.evil.test/callback",
	} {
		if _, err := svc.ValidateClient(ctx, client.ID, uri); !errors.Is(err, ErrInvalidRedirectURI) {
			t.Fatalf("redirect %q was accepted: %v", uri, err)
		}
	}
}

func TestOAuthService_ExpiredCodeIsRefused(t *testing.T) {
	svc, repo, db, userID := newOAuthFixture(t)
	ctx := context.Background()
	client, secret, err := svc.RegisterClient(ctx, "Client", []string{fixtureRedirect})
	if err != nil {
		t.Fatal(err)
	}
	code, err := svc.Authorize(ctx, client.ID, fixtureRedirect, "read", userID, "", "")
	if err != nil {
		t.Fatal(err)
	}

	// Codes live ten minutes; age this one out the way the clock would.
	past := time.Now().UTC().Add(-time.Minute).Format(time.DateTime)
	if _, err := db.NewUpdate().Table("oauth_codes").Set("expires_at=?", past).
		Where("code=?", code).Exec(ctx); err != nil {
		t.Fatal(err)
	}

	if _, _, _, err := svc.Exchange(ctx, code, client.ID, secret, fixtureRedirect, ""); !errors.Is(err, ErrInvalidCode) {
		t.Fatalf("expired code: %v", err)
	}
	// An expired code that is refused stays refused; it is not consumed by the
	// attempt, which is what CleanExpiredCodes is for.
	if found, err := repo.FindCode(ctx, code); err != nil || found == nil {
		t.Fatalf("refusing an expired code deleted it: %v", err)
	}
}

func TestOAuthService_PKCEMethods(t *testing.T) {
	svc, _, _, userID := newOAuthFixture(t)
	ctx := context.Background()
	client, secret, err := svc.RegisterClient(ctx, "Client", []string{fixtureRedirect})
	if err != nil {
		t.Fatal(err)
	}
	const verifier = "a-verifier-of-a-perfectly-reasonable-length"

	cases := []struct {
		name      string
		challenge string
		method    string
		verifier  string
		wantErr   bool
	}{
		{"S256 matching", challengeFor(verifier), "S256", verifier, false},
		{"S256 mismatched", challengeFor(verifier), "S256", "something-else", true},
		{"plain matching", verifier, "plain", verifier, false},
		{"plain mismatched", verifier, "plain", "something-else", true},
		{"absent method defaults to plain", verifier, "", verifier, false},
		// A method nobody advertises must fail closed. Falling back to the
		// plain comparison would let a challenge stored as a SHA-256 digest be
		// satisfied by presenting the digest itself.
		{"unknown method", challengeFor(verifier), "S512", verifier, true},
		{"unknown method, digest replayed", challengeFor(verifier), "S512", challengeFor(verifier), true},
		// No challenge at all: PKCE was never requested, so nothing to check.
		{"no challenge", "", "", "", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			code, err := svc.Authorize(ctx, client.ID, fixtureRedirect, "read", userID, tc.challenge, tc.method)
			if err != nil {
				t.Fatal(err)
			}
			_, _, _, err = svc.Exchange(ctx, code, client.ID, secret, fixtureRedirect, tc.verifier)
			if tc.wantErr && !errors.Is(err, ErrInvalidCode) {
				t.Fatalf("want refusal, got %v", err)
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("want success, got %v", err)
			}
		})
	}
}

// TestOAuthService_ConcurrentExchange — the lookup and the delete are separate
// statements, so two requests arriving together both pass the lookup. Exactly
// one may end up with tokens.
func TestOAuthService_ConcurrentExchange(t *testing.T) {
	svc, _, db, userID := newOAuthFixture(t)
	ctx := context.Background()
	client, secret, err := svc.RegisterClient(ctx, "Client", []string{fixtureRedirect})
	if err != nil {
		t.Fatal(err)
	}
	code, err := svc.Authorize(ctx, client.ID, fixtureRedirect, "read", userID, "", "")
	if err != nil {
		t.Fatal(err)
	}

	const racers = 8
	var (
		wg      sync.WaitGroup
		mu      sync.Mutex
		granted int
	)
	start := make(chan struct{})
	for i := 0; i < racers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			if access, _, _, err := svc.Exchange(ctx, code, client.ID, secret, fixtureRedirect, ""); err == nil && access != "" {
				mu.Lock()
				granted++
				mu.Unlock()
			}
		}()
	}
	close(start)
	wg.Wait()

	if granted != 1 {
		t.Fatalf("%d of %d exchanges succeeded on one code, want 1", granted, racers)
	}
	var tokens int
	if err := db.NewSelect().ColumnExpr("count(*)").TableExpr("oauth_tokens").Scan(ctx, &tokens); err != nil {
		t.Fatal(err)
	}
	if tokens != 1 {
		t.Fatalf("%d token rows exist for one authorization", tokens)
	}
}

func TestOAuthService_AccessTokenLifetime(t *testing.T) {
	svc, repo, _, userID := newOAuthFixture(t)
	ctx := context.Background()
	client, secret, err := svc.RegisterClient(ctx, "Client", []string{fixtureRedirect})
	if err != nil {
		t.Fatal(err)
	}
	code, err := svc.Authorize(ctx, client.ID, fixtureRedirect, "read write", userID, "", "")
	if err != nil {
		t.Fatal(err)
	}
	access, refresh, expiresIn, err := svc.Exchange(ctx, code, client.ID, secret, fixtureRedirect, "")
	if err != nil {
		t.Fatal(err)
	}
	if expiresIn != int(AccessTokenTTL/time.Second) {
		t.Fatalf("expires_in %d, want %d", expiresIn, int(AccessTokenTTL/time.Second))
	}

	// What the client is told and what the database enforces have to agree.
	stored, err := repo.FindByAccessTokenHash(ctx, auth.HashAPIKey(access))
	if err != nil || stored == nil {
		t.Fatalf("stored token: %v", err)
	}
	if d := time.Until(stored.ExpiresAt); d > AccessTokenTTL+time.Minute || d < AccessTokenTTL-time.Minute {
		t.Fatalf("stored expires_at is %v away, expires_in said %ds", d, expiresIn)
	}
	if stored.Scope != "read write" {
		t.Fatalf("scope %q was not carried from the authorization", stored.Scope)
	}

	// Refresh keeps the scope and the user, and retires the row it came from.
	newAccess, _, _, err := svc.Refresh(ctx, refresh, client.ID, secret)
	if err != nil {
		t.Fatal(err)
	}
	rotated, err := repo.FindByAccessTokenHash(ctx, auth.HashAPIKey(newAccess))
	if err != nil || rotated == nil {
		t.Fatalf("rotated token: %v", err)
	}
	if rotated.UserID != userID || rotated.Scope != "read write" {
		t.Fatalf("rotation lost the grant: user %q scope %q", rotated.UserID, rotated.Scope)
	}
	if old, _ := repo.FindByAccessTokenHash(ctx, auth.HashAPIKey(access)); old != nil {
		t.Fatal("the refreshed-away access token is still stored")
	}
}

func TestOAuthService_ExpiredRefreshTokenIsRefused(t *testing.T) {
	svc, _, db, userID := newOAuthFixture(t)
	ctx := context.Background()
	client, secret, err := svc.RegisterClient(ctx, "Client", []string{fixtureRedirect})
	if err != nil {
		t.Fatal(err)
	}
	code, err := svc.Authorize(ctx, client.ID, fixtureRedirect, "read", userID, "", "")
	if err != nil {
		t.Fatal(err)
	}
	_, refresh, _, err := svc.Exchange(ctx, code, client.ID, secret, fixtureRedirect, "")
	if err != nil {
		t.Fatal(err)
	}

	// The refresh token's life is measured from the row's created_at, so this
	// is how a month of silence looks.
	stale := time.Now().UTC().Add(-RefreshTokenTTL - time.Hour).Format(time.DateTime)
	if _, err := db.NewUpdate().Table("oauth_tokens").Set("created_at=?", stale).
		Where("refresh_token_hash=?", auth.HashAPIKey(refresh)).Exec(ctx); err != nil {
		t.Fatal(err)
	}

	if _, _, _, err := svc.Refresh(ctx, refresh, client.ID, secret); !errors.Is(err, ErrInvalidRefresh) {
		t.Fatalf("stale refresh token: %v", err)
	}
}

// TestOAuthService_AuthorizeBindsTheRedirect — the code carries the URI it was
// asked for, and the exchange has to be handed the same one. Otherwise a code
// intercepted at one callback is spendable from another.
func TestOAuthService_AuthorizeBindsTheRedirect(t *testing.T) {
	svc, _, _, userID := newOAuthFixture(t)
	ctx := context.Background()
	client, secret, err := svc.RegisterClient(ctx, "Client",
		[]string{fixtureRedirect, "https://client.example.com/second"})
	if err != nil {
		t.Fatal(err)
	}

	if _, err := svc.Authorize(ctx, client.ID, "https://evil.example.com/", "read", userID, "", ""); !errors.Is(err, ErrInvalidRedirectURI) {
		t.Fatalf("authorize accepted an unregistered redirect: %v", err)
	}

	code, err := svc.Authorize(ctx, client.ID, fixtureRedirect, "read", userID, "", "")
	if err != nil {
		t.Fatal(err)
	}
	// The second URI is registered for this client, and still wrong for this code.
	if _, _, _, err := svc.Exchange(ctx, code, client.ID, secret,
		"https://client.example.com/second", ""); !errors.Is(err, ErrInvalidRedirectURI) {
		t.Fatalf("exchange accepted a different registered uri: %v", err)
	}
}

// TestOAuthService_DCRClientsAreUnowned records the deliberate asymmetry:
// CreateClient belongs to a person and shows up in their list, a client
// registered through RFC 7591 belongs to nobody and must not appear in anyone's.
func TestOAuthService_DCRClientsAreUnowned(t *testing.T) {
	svc, _, db, userID := newOAuthFixture(t)
	ctx := context.Background()

	if _, _, err := svc.CreateClient(ctx, userID, "My own app", []string{fixtureRedirect}); err != nil {
		t.Fatal(err)
	}
	if _, _, err := svc.RegisterClient(ctx, "ChatGPT", []string{fixtureRedirect}); err != nil {
		t.Fatal(err)
	}

	mine, err := svc.ListClients(ctx, userID)
	if err != nil {
		t.Fatal(err)
	}
	if len(mine) != 1 || mine[0].Name != "My own app" {
		t.Fatalf("ListClients returned %d clients: %+v", len(mine), mine)
	}
	var total int
	if err := db.NewSelect().ColumnExpr("count(*)").TableExpr("oauth_clients").Scan(ctx, &total); err != nil {
		t.Fatal(err)
	}
	if total != 2 {
		t.Fatalf("%d clients stored, want 2", total)
	}
}
