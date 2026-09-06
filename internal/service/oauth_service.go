package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/dovod-app/app/internal/domain"
	"github.com/dovod-app/app/internal/storage"
	"github.com/google/uuid"
)

var (
	ErrInvalidClient      = errors.New("invalid client_id or client_secret")
	ErrInvalidCode        = errors.New("invalid or expired authorization code")
	ErrInvalidRedirectURI = errors.New("redirect_uri mismatch")
	ErrInvalidGrant       = errors.New("invalid grant_type")
	ErrInvalidRefresh     = errors.New("invalid or expired refresh_token")
)

const (
	// AccessTokenTTL is how long an access token is honoured. It is also the
	// `expires_in` the token endpoint reports, and the two must not drift: a
	// client that is told an hour and gets a week has no reason to refresh, and
	// a client told an hour that gets a minute loops.
	AccessTokenTTL = time.Hour
	// RefreshTokenTTL is measured from the row's created_at, and rotation
	// writes a new row — so a client that keeps refreshing keeps working, and
	// one that goes quiet for a month has to sign in again.
	RefreshTokenTTL = 30 * 24 * time.Hour
)

type OAuthService struct {
	repo *storage.OAuthRepository
	log  *slog.Logger
}

func NewOAuthService(repo *storage.OAuthRepository, log *slog.Logger) *OAuthService {
	return &OAuthService{repo: repo, log: log}
}

func (s *OAuthService) CreateClient(ctx context.Context, userID, name string, redirectURIs []string) (*domain.OAuthClient, string, error) {
	clientSecret := generateSecret(32)
	secretHash := hashSHA256(clientSecret)

	client := &domain.OAuthClient{
		ID:           uuid.New().String(),
		UserID:       userID,
		Name:         name,
		RedirectURIs: redirectURIs,
	}

	if err := s.repo.CreateClient(ctx, client, secretHash); err != nil {
		return nil, "", fmt.Errorf("create oauth client: %w", err)
	}

	return client, clientSecret, nil
}

func (s *OAuthService) ListClients(ctx context.Context, userID string) ([]*domain.OAuthClient, error) {
	return s.repo.ListClientsByUser(ctx, userID)
}

// RegisterClient implements RFC 7591 Dynamic Client Registration.
// Creates a client without an owner — used by external services like ChatGPT.
func (s *OAuthService) RegisterClient(ctx context.Context, name string, redirectURIs []string) (*domain.OAuthClient, string, error) {
	clientSecret := generateSecret(32)
	secretHash := hashSHA256(clientSecret)

	client := &domain.OAuthClient{
		ID:           uuid.New().String(),
		UserID:       "", // unowned — DCR client
		Name:         name,
		RedirectURIs: redirectURIs,
	}

	if err := s.repo.CreateClient(ctx, client, secretHash); err != nil {
		return nil, "", fmt.Errorf("register oauth client: %w", err)
	}

	return client, clientSecret, nil
}

// ValidateClient checks that a client exists and the redirect_uri is allowed. Returns client name.
func (s *OAuthService) ValidateClient(ctx context.Context, clientID, redirectURI string) (string, error) {
	client, _, err := s.repo.FindClientByID(ctx, clientID)
	if err != nil {
		return "", fmt.Errorf("find client: %w", err)
	}
	if client == nil {
		return "", ErrInvalidClient
	}
	if !containsURI(client.RedirectURIs, redirectURI) {
		return "", ErrInvalidRedirectURI
	}
	return client.Name, nil
}

// Authorize creates an authorization code for the given user and client.
func (s *OAuthService) Authorize(ctx context.Context, clientID, redirectURI, scope, userID, codeChallenge, codeChallengeMethod string) (string, error) {
	client, _, err := s.repo.FindClientByID(ctx, clientID)
	if err != nil {
		return "", fmt.Errorf("find client: %w", err)
	}
	if client == nil {
		return "", ErrInvalidClient
	}

	// Validate redirect_uri
	if !containsURI(client.RedirectURIs, redirectURI) {
		return "", ErrInvalidRedirectURI
	}

	code := generateSecret(32)
	oauthCode := &storage.OAuthCode{
		Code:                code,
		ClientID:            clientID,
		UserID:              userID,
		RedirectURI:         redirectURI,
		Scope:               scope,
		CodeChallenge:       codeChallenge,
		CodeChallengeMethod: codeChallengeMethod,
		ExpiresAt:           time.Now().Add(10 * time.Minute),
	}

	if err := s.repo.CreateCode(ctx, oauthCode); err != nil {
		return "", fmt.Errorf("create code: %w", err)
	}

	return code, nil
}

// Exchange trades an authorization code for access + refresh tokens.
func (s *OAuthService) Exchange(ctx context.Context, code, clientID, clientSecret, redirectURI, codeVerifier string) (accessToken, refreshToken string, expiresIn int, err error) {
	// Validate client
	client, secretHash, err := s.repo.FindClientByID(ctx, clientID)
	if err != nil {
		return "", "", 0, fmt.Errorf("find client: %w", err)
	}
	if client == nil || subtle.ConstantTimeCompare([]byte(hashSHA256(clientSecret)), []byte(secretHash)) != 1 {
		return "", "", 0, ErrInvalidClient
	}

	// Validate code
	oauthCode, err := s.repo.FindCode(ctx, code)
	if err != nil {
		return "", "", 0, fmt.Errorf("find code: %w", err)
	}
	if oauthCode == nil || oauthCode.ClientID != clientID || time.Now().After(oauthCode.ExpiresAt) {
		return "", "", 0, ErrInvalidCode
	}
	if oauthCode.RedirectURI != redirectURI {
		return "", "", 0, ErrInvalidRedirectURI
	}

	// Validate PKCE code_verifier
	if oauthCode.CodeChallenge != "" {
		if codeVerifier == "" {
			return "", "", 0, fmt.Errorf("code_verifier required: %w", ErrInvalidCode)
		}
		if !verifyPKCE(oauthCode.CodeChallenge, oauthCode.CodeChallengeMethod, codeVerifier) {
			return "", "", 0, fmt.Errorf("PKCE verification failed: %w", ErrInvalidCode)
		}
	}

	// Burn the code before minting anything. Losing the race here means some
	// other request already exchanged this code, and the correct answer is the
	// same one a stranger replaying it gets.
	used, err := s.repo.ConsumeCode(ctx, code)
	if err != nil {
		return "", "", 0, fmt.Errorf("consume code: %w", err)
	}
	if !used {
		return "", "", 0, ErrInvalidCode
	}

	return s.issue(ctx, clientID, oauthCode.UserID, oauthCode.Scope)
}

// Refresh trades a refresh token for a new pair and retires the old one.
//
// Rotation is the point: the row carrying the presented refresh token is
// deleted, so replaying a stolen refresh token that its owner has already used
// fails. Without this grant an access token that actually expires would strand
// every client an hour after it signed in.
func (s *OAuthService) Refresh(ctx context.Context, refreshToken, clientID, clientSecret string) (accessToken, newRefreshToken string, expiresIn int, err error) {
	client, secretHash, err := s.repo.FindClientByID(ctx, clientID)
	if err != nil {
		return "", "", 0, fmt.Errorf("find client: %w", err)
	}
	if client == nil || subtle.ConstantTimeCompare([]byte(hashSHA256(clientSecret)), []byte(secretHash)) != 1 {
		return "", "", 0, ErrInvalidClient
	}

	token, err := s.repo.FindByRefreshTokenHash(ctx, hashSHA256(refreshToken))
	if err != nil {
		return "", "", 0, fmt.Errorf("find refresh token: %w", err)
	}
	if token == nil || token.ClientID != clientID || time.Now().After(token.IssuedAt.Add(RefreshTokenTTL)) {
		return "", "", 0, ErrInvalidRefresh
	}

	if err := s.repo.DeleteToken(ctx, token.ID); err != nil {
		return "", "", 0, fmt.Errorf("retire token: %w", err)
	}

	return s.issue(ctx, clientID, token.UserID, token.Scope)
}

func (s *OAuthService) issue(ctx context.Context, clientID, userID, scope string) (accessToken, refreshToken string, expiresIn int, err error) {
	accessToken = generateSecret(32)
	refreshToken = generateSecret(32)
	expiresIn = int(AccessTokenTTL / time.Second)

	token := &storage.OAuthToken{
		ID:               uuid.New().String(),
		ClientID:         clientID,
		UserID:           userID,
		AccessTokenHash:  hashSHA256(accessToken),
		RefreshTokenHash: hashSHA256(refreshToken),
		Scope:            scope,
		ExpiresAt:        time.Now().Add(AccessTokenTTL),
	}

	if err := s.repo.CreateToken(ctx, token); err != nil {
		return "", "", 0, fmt.Errorf("create token: %w", err)
	}

	return accessToken, refreshToken, expiresIn, nil
}

func generateSecret(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func hashSHA256(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}

// verifyPKCE validates the code_verifier against the stored code_challenge.
func verifyPKCE(challenge, method, verifier string) bool {
	switch method {
	case "S256":
		h := sha256.Sum256([]byte(verifier))
		return subtle.ConstantTimeCompare([]byte(base64URLEncode(h[:])), []byte(challenge)) == 1
	case "", "plain":
		// RFC 7636 defaults an absent method to "plain".
		return subtle.ConstantTimeCompare([]byte(verifier), []byte(challenge)) == 1
	default:
		// An unknown method is not a reason to fall back to the weaker check.
		return false
	}
}

func base64URLEncode(data []byte) string {
	s := base64.RawURLEncoding.EncodeToString(data)
	return s
}

func containsURI(uris []string, uri string) bool {
	for _, u := range uris {
		if u == uri {
			return true
		}
	}
	return false
}
