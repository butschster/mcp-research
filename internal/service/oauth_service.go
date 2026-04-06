package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/butschster/mcp-research/internal/domain"
	"github.com/butschster/mcp-research/internal/storage"
	"github.com/google/uuid"
)

var (
	ErrInvalidClient      = errors.New("invalid client_id or client_secret")
	ErrInvalidCode        = errors.New("invalid or expired authorization code")
	ErrInvalidRedirectURI = errors.New("redirect_uri mismatch")
	ErrInvalidGrant       = errors.New("invalid grant_type")
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
	if client == nil || hashSHA256(clientSecret) != secretHash {
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

	// Delete used code
	s.repo.DeleteCode(ctx, code)

	// Generate tokens
	accessToken = generateSecret(32)
	refreshToken = generateSecret(32)
	expiresIn = 3600 // 1 hour

	token := &storage.OAuthToken{
		ID:               uuid.New().String(),
		ClientID:         clientID,
		UserID:           oauthCode.UserID,
		AccessTokenHash:  hashSHA256(accessToken),
		RefreshTokenHash: hashSHA256(refreshToken),
		Scope:            oauthCode.Scope,
		ExpiresAt:        time.Now().Add(time.Duration(expiresIn) * time.Second),
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
	if method == "S256" {
		h := sha256.Sum256([]byte(verifier))
		computed := base64URLEncode(h[:])
		return computed == challenge
	}
	// plain method
	return verifier == challenge
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
