package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/butschster/mcp-research/internal/auth"
	"github.com/butschster/mcp-research/internal/domain"
	"github.com/butschster/mcp-research/internal/storage"
	"github.com/google/uuid"
)

var (
	ErrEmailTaken         = errors.New("email already registered")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrRegistrationClosed = errors.New("registration is not allowed")
	ErrPasswordTooShort   = errors.New("password must be at least 6 characters")
	ErrInvalidEmail       = errors.New("invalid email address")
)

type AuthService struct {
	users             *storage.UserRepository
	apiKeys           *storage.APIKeyRepository
	oauthRepo         *storage.OAuthRepository
	researches        *storage.ResearchRepository
	teams             *storage.TeamRepository
	jwt               *auth.JWTManager
	allowRegistration bool
	log               *slog.Logger
}

func NewAuthService(
	users *storage.UserRepository,
	apiKeys *storage.APIKeyRepository,
	oauthRepo *storage.OAuthRepository,
	researches *storage.ResearchRepository,
	teams *storage.TeamRepository,
	jwt *auth.JWTManager,
	allowRegistration bool,
	log *slog.Logger,
) *AuthService {
	return &AuthService{
		users:             users,
		apiKeys:           apiKeys,
		oauthRepo:         oauthRepo,
		researches:        researches,
		teams:             teams,
		jwt:               jwt,
		allowRegistration: allowRegistration,
		log:               log,
	}
}

func (s *AuthService) Register(ctx context.Context, email, password, name string) (*domain.User, string, error) {
	if !s.allowRegistration {
		return nil, "", ErrRegistrationClosed
	}
	return s.register(ctx, email, password, name)
}

// RegisterInvited creates an account for someone holding a valid invitation,
// whether or not open registration is on. The invite is the authorization:
// without this, turning registration off would make it impossible to bring
// anyone onto a team, which is the one thing invitations exist for.
//
// The caller must have already validated the invite.
func (s *AuthService) RegisterInvited(ctx context.Context, email, password, name string) (*domain.User, string, error) {
	return s.register(ctx, email, password, name)
}

func (s *AuthService) register(ctx context.Context, email, password, name string) (*domain.User, string, error) {
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

	// The personal team is not optional: without one the user cannot create a
	// research at all, so a failure here is a failed registration rather than
	// something to log and carry on from.
	team := &domain.Team{
		ID:        uuid.New().String(),
		Name:      personalTeamName(user),
		Personal:  true,
		CreatedBy: user.ID,
	}
	if err := s.teams.CreateWithOwner(ctx, team, user.ID); err != nil {
		return nil, "", fmt.Errorf("create personal team: %w", err)
	}

	// First user claims orphaned researches
	count, _ := s.users.Count(ctx)
	if count == 1 {
		if n, err := s.researches.ClaimOrphanedResearches(ctx, user.ID, team.ID); err == nil && n > 0 {
			s.log.Info("claimed orphaned researches for first user", "user", user.Email, "count", n)
		}
	}

	token, err := s.jwt.Generate(user.ID)
	if err != nil {
		return nil, "", fmt.Errorf("generate token: %w", err)
	}

	return user, token, nil
}

// personalTeamName is what the user will see in a team picker before they have
// ever thought about teams, so it is their own name rather than "Personal".
func personalTeamName(user *domain.User) string {
	if name := strings.TrimSpace(user.Name); name != "" {
		return name
	}
	if i := strings.Index(user.Email, "@"); i > 0 {
		return user.Email[:i]
	}
	return user.Email
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

// IsSessionToken reports whether the token is a JWT issued to a browser session,
// as opposed to an API key or an OAuth access token held by a program. Callers
// use it to tell a person's write from a machine's; it says nothing about
// whether the token is authorized, which is ValidateToken's job.
func (s *AuthService) IsSessionToken(token string) bool {
	if s == nil || token == "" {
		return false
	}
	_, err := s.jwt.Validate(token)
	return err == nil
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
		// An access token that is past its expires_at is not a credential. The
		// token endpoint has always announced `expires_in`, and nothing until
		// now enforced it: a token minted for an hour authenticated forever.
		// The client's way back is the refresh_token grant.
		if time.Now().After(oauthToken.ExpiresAt) {
			return nil, nil
		}
		return s.users.FindByID(ctx, oauthToken.UserID)
	}

	return nil, nil
}

// UserIDForToken resolves a credential to a user id, or "" for anything it
// does not accept. It is what the WebSocket handshake authenticates with —
// the same JWT, API key or OAuth token every other route takes.
func (s *AuthService) UserIDForToken(ctx context.Context, token string) string {
	id, _ := s.ValidateCredential(ctx, token)
	return id
}

// ValidateCredential is UserIDForToken with the two failures told apart.
//
// `ok` is false when the answer could not be determined — the lookup itself
// failed. That is not a revocation, and treating it as one is expensive: the
// WebSocket re-checks live connections on a timer, against a database that
// permits one connection at a time, so a single busy moment would close a
// perfectly good socket and tell its owner they had been signed out.
func (s *AuthService) ValidateCredential(ctx context.Context, token string) (string, bool) {
	if s == nil {
		return "", false
	}
	user, err := s.ValidateToken(ctx, token)
	if err != nil {
		return "", false
	}
	if user == nil {
		return "", true
	}
	return user.ID, true
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
