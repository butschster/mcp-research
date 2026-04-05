package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/butschster/mcp-research/internal/auth"
	"github.com/butschster/mcp-research/internal/domain"
	"github.com/butschster/mcp-research/internal/storage"
	"github.com/google/uuid"
)

var (
	ErrEmailTaken          = errors.New("email already registered")
	ErrInvalidCredentials  = errors.New("invalid email or password")
	ErrRegistrationClosed  = errors.New("registration is not allowed")
	ErrPasswordTooShort    = errors.New("password must be at least 6 characters")
	ErrInvalidEmail        = errors.New("invalid email address")
)

type AuthService struct {
	users             *storage.UserRepository
	apiKeys           *storage.APIKeyRepository
	oauthRepo         *storage.OAuthRepository
	researches        *storage.ResearchRepository
	jwt               *auth.JWTManager
	allowRegistration bool
	log               *slog.Logger
}

func NewAuthService(
	users *storage.UserRepository,
	apiKeys *storage.APIKeyRepository,
	oauthRepo *storage.OAuthRepository,
	researches *storage.ResearchRepository,
	jwt *auth.JWTManager,
	allowRegistration bool,
	log *slog.Logger,
) *AuthService {
	return &AuthService{
		users:             users,
		apiKeys:           apiKeys,
		oauthRepo:         oauthRepo,
		researches:        researches,
		jwt:               jwt,
		allowRegistration: allowRegistration,
		log:               log,
	}
}

func (s *AuthService) Register(ctx context.Context, email, password, name string) (*domain.User, string, error) {
	if !s.allowRegistration {
		return nil, "", ErrRegistrationClosed
	}

	email = strings.TrimSpace(strings.ToLower(email))
	if !strings.Contains(email, "@") || !strings.Contains(email, ".") {
		return nil, "", ErrInvalidEmail
	}
	if len(password) < 6 {
		return nil, "", ErrPasswordTooShort
	}

	existing, err := s.users.FindByEmail(ctx, email)
	if err != nil {
		return nil, "", fmt.Errorf("check email: %w", err)
	}
	if existing != nil {
		return nil, "", ErrEmailTaken
	}

	hash, err := auth.HashPassword(password)
	if err != nil {
		return nil, "", fmt.Errorf("hash password: %w", err)
	}

	user := &domain.User{
		ID:           uuid.New().String(),
		Email:        email,
		PasswordHash: hash,
		Name:         name,
	}

	if err := s.users.Create(ctx, user); err != nil {
		return nil, "", fmt.Errorf("create user: %w", err)
	}

	// First user claims orphaned researches
	count, _ := s.users.Count(ctx)
	if count == 1 {
		if n, err := s.researches.ClaimOrphanedResearches(ctx, user.ID); err == nil && n > 0 {
			s.log.Info("claimed orphaned researches for first user", "user", user.Email, "count", n)
		}
	}

	token, err := s.jwt.Generate(user.ID)
	if err != nil {
		return nil, "", fmt.Errorf("generate token: %w", err)
	}

	return user, token, nil
}

func (s *AuthService) Login(ctx context.Context, email, password string) (*domain.User, string, error) {
	email = strings.TrimSpace(strings.ToLower(email))

	user, err := s.users.FindByEmail(ctx, email)
	if err != nil {
		return nil, "", fmt.Errorf("find user: %w", err)
	}
	if user == nil {
		return nil, "", ErrInvalidCredentials
	}

	if err := auth.CheckPassword(user.PasswordHash, password); err != nil {
		return nil, "", ErrInvalidCredentials
	}

	token, err := s.jwt.Generate(user.ID)
	if err != nil {
		return nil, "", fmt.Errorf("generate token: %w", err)
	}

	return user, token, nil
}

// ValidateToken tries JWT, then API key, then OAuth token. Returns the user or nil.
func (s *AuthService) ValidateToken(ctx context.Context, token string) (*domain.User, error) {
	if token == "" {
		return nil, nil
	}

	// Try JWT
	if userID, err := s.jwt.Validate(token); err == nil {
		return s.users.FindByID(ctx, userID)
	}

	// Try API key
	hash := auth.HashAPIKey(token)
	apiKey, err := s.apiKeys.FindByHash(ctx, hash)
	if err != nil {
		return nil, fmt.Errorf("find api key: %w", err)
	}
	if apiKey != nil {
		go s.apiKeys.TouchLastUsed(context.Background(), apiKey.ID)
		return s.users.FindByID(ctx, apiKey.UserID)
	}

	// Try OAuth access token
	oauthToken, err := s.oauthRepo.FindByAccessTokenHash(ctx, hash)
	if err != nil {
		return nil, fmt.Errorf("find oauth token: %w", err)
	}
	if oauthToken != nil {
		return s.users.FindByID(ctx, oauthToken.UserID)
	}

	return nil, nil
}

// CreateAPIKey generates a new API key for the user.
func (s *AuthService) CreateAPIKey(ctx context.Context, userID, name string) (string, *domain.APIKey, error) {
	plain, hash, prefix := auth.GenerateAPIKey()

	key := &domain.APIKey{
		ID:        uuid.New().String(),
		UserID:    userID,
		Name:      name,
		KeyPrefix: prefix,
	}

	if err := s.apiKeys.Create(ctx, key, hash); err != nil {
		return "", nil, fmt.Errorf("create api key: %w", err)
	}

	return plain, key, nil
}

func (s *AuthService) ListAPIKeys(ctx context.Context, userID string) ([]*domain.APIKey, error) {
	return s.apiKeys.ListByUser(ctx, userID)
}

func (s *AuthService) DeleteAPIKey(ctx context.Context, keyID, userID string) error {
	return s.apiKeys.Delete(ctx, keyID, userID)
}

func (s *AuthService) AllowRegistration() bool {
	return s.allowRegistration
}
