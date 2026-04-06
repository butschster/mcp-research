package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/butschster/mcp-research/internal/domain"
)

type OAuthRepository struct {
	db *sql.DB
}

func NewOAuthRepository(db *sql.DB) *OAuthRepository {
	return &OAuthRepository{db: db}
}

// --- Clients ---

func (r *OAuthRepository) CreateClient(ctx context.Context, client *domain.OAuthClient, secretHash string) error {
	now := time.Now().UTC().Format(time.DateTime)
	var userID any
	if client.UserID != "" {
		userID = client.UserID
	}
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO oauth_clients (id, user_id, secret_hash, name, redirect_uris, created_at)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		client.ID, userID, secretHash, client.Name,
		marshalJSON(client.RedirectURIs), now,
	)
	if err != nil {
		return fmt.Errorf("insert oauth client: %w", err)
	}
	client.CreatedAt, _ = time.Parse(time.DateTime, now)
	return nil
}

func (r *OAuthRepository) FindClientByID(ctx context.Context, id string) (*domain.OAuthClient, string, error) {
	var client domain.OAuthClient
	var secretHash string
	var userID sql.NullString
	var redirectURIs sql.NullString
	var createdAt string
	err := r.db.QueryRowContext(ctx,
		`SELECT id, user_id, secret_hash, name, redirect_uris, created_at
		 FROM oauth_clients WHERE id=?`, id,
	).Scan(&client.ID, &userID, &secretHash, &client.Name, &redirectURIs, &createdAt)
	if err == sql.ErrNoRows {
		return nil, "", nil
	}
	if err != nil {
		return nil, "", fmt.Errorf("scan oauth client: %w", err)
	}
	if userID.Valid {
		client.UserID = userID.String
	}
	client.RedirectURIs = unmarshalStringSlice(redirectURIs)
	client.CreatedAt, _ = time.Parse(time.DateTime, createdAt)
	return &client, secretHash, nil
}

func (r *OAuthRepository) ListClientsByUser(ctx context.Context, userID string) ([]*domain.OAuthClient, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, user_id, name, redirect_uris, created_at
		 FROM oauth_clients WHERE user_id=? ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, fmt.Errorf("query oauth clients: %w", err)
	}
	defer rows.Close()

	var result []*domain.OAuthClient
	for rows.Next() {
		var client domain.OAuthClient
		var redirectURIs sql.NullString
		var createdAt string
		if err := rows.Scan(&client.ID, &client.UserID, &client.Name, &redirectURIs, &createdAt); err != nil {
			return nil, fmt.Errorf("scan oauth client row: %w", err)
		}
		client.RedirectURIs = unmarshalStringSlice(redirectURIs)
		client.CreatedAt, _ = time.Parse(time.DateTime, createdAt)
		result = append(result, &client)
	}
	return result, rows.Err()
}

// --- Authorization Codes ---

type OAuthCode struct {
	Code                string
	ClientID            string
	UserID              string
	RedirectURI         string
	Scope               string
	CodeChallenge       string
	CodeChallengeMethod string
	ExpiresAt           time.Time
}

func (r *OAuthRepository) CreateCode(ctx context.Context, code *OAuthCode) error {
	now := time.Now().UTC().Format(time.DateTime)
	expiresAt := code.ExpiresAt.UTC().Format(time.DateTime)
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO oauth_codes (code, client_id, user_id, redirect_uri, scope, code_challenge, code_challenge_method, expires_at, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		code.Code, code.ClientID, code.UserID, code.RedirectURI, code.Scope,
		code.CodeChallenge, code.CodeChallengeMethod, expiresAt, now,
	)
	return err
}

func (r *OAuthRepository) FindCode(ctx context.Context, code string) (*OAuthCode, error) {
	var c OAuthCode
	var expiresAt string
	err := r.db.QueryRowContext(ctx,
		`SELECT code, client_id, user_id, redirect_uri, scope, code_challenge, code_challenge_method, expires_at
		 FROM oauth_codes WHERE code=?`, code,
	).Scan(&c.Code, &c.ClientID, &c.UserID, &c.RedirectURI, &c.Scope, &c.CodeChallenge, &c.CodeChallengeMethod, &expiresAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("scan oauth code: %w", err)
	}
	c.ExpiresAt, _ = time.Parse(time.DateTime, expiresAt)
	return &c, nil
}

func (r *OAuthRepository) DeleteCode(ctx context.Context, code string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM oauth_codes WHERE code=?`, code)
	return err
}

func (r *OAuthRepository) CleanExpiredCodes(ctx context.Context) error {
	now := time.Now().UTC().Format(time.DateTime)
	_, err := r.db.ExecContext(ctx, `DELETE FROM oauth_codes WHERE expires_at < ?`, now)
	return err
}

// --- Tokens ---

type OAuthToken struct {
	ID               string
	ClientID         string
	UserID           string
	AccessTokenHash  string
	RefreshTokenHash string
	Scope            string
	ExpiresAt        time.Time
}

func (r *OAuthRepository) CreateToken(ctx context.Context, token *OAuthToken) error {
	now := time.Now().UTC().Format(time.DateTime)
	expiresAt := token.ExpiresAt.UTC().Format(time.DateTime)
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO oauth_tokens (id, client_id, user_id, access_token_hash, refresh_token_hash, scope, expires_at, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		token.ID, token.ClientID, token.UserID,
		token.AccessTokenHash, token.RefreshTokenHash,
		token.Scope, expiresAt, now,
	)
	return err
}

func (r *OAuthRepository) FindByAccessTokenHash(ctx context.Context, hash string) (*OAuthToken, error) {
	var t OAuthToken
	var expiresAt string
	err := r.db.QueryRowContext(ctx,
		`SELECT id, client_id, user_id, access_token_hash, refresh_token_hash, scope, expires_at
		 FROM oauth_tokens WHERE access_token_hash=?`, hash,
	).Scan(&t.ID, &t.ClientID, &t.UserID, &t.AccessTokenHash, &t.RefreshTokenHash, &t.Scope, &expiresAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("scan oauth token: %w", err)
	}
	t.ExpiresAt, _ = time.Parse(time.DateTime, expiresAt)
	return &t, nil
}

func (r *OAuthRepository) FindByRefreshTokenHash(ctx context.Context, hash string) (*OAuthToken, error) {
	var t OAuthToken
	var expiresAt string
	err := r.db.QueryRowContext(ctx,
		`SELECT id, client_id, user_id, access_token_hash, refresh_token_hash, scope, expires_at
		 FROM oauth_tokens WHERE refresh_token_hash=?`, hash,
	).Scan(&t.ID, &t.ClientID, &t.UserID, &t.AccessTokenHash, &t.RefreshTokenHash, &t.Scope, &expiresAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("scan oauth token: %w", err)
	}
	t.ExpiresAt, _ = time.Parse(time.DateTime, expiresAt)
	return &t, nil
}

func (r *OAuthRepository) DeleteToken(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM oauth_tokens WHERE id=?`, id)
	return err
}

func (r *OAuthRepository) CleanExpiredTokens(ctx context.Context) error {
	now := time.Now().UTC().Format(time.DateTime)
	_, err := r.db.ExecContext(ctx, `DELETE FROM oauth_tokens WHERE expires_at < ?`, now)
	return err
}
