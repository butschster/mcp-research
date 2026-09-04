package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/butschster/mcp-research/internal/domain"
	"github.com/uptrace/bun"
)

type OAuthRepository struct {
	db *bun.DB
}

func NewOAuthRepository(db *bun.DB) *OAuthRepository {
	return &OAuthRepository{db: db}
}

// --- Clients ---

func (r *OAuthRepository) CreateClient(ctx context.Context, client *domain.OAuthClient, secretHash string) error {
	now := time.Now().UTC().Format(time.DateTime)
	var userID any
	if client.UserID != "" {
		userID = client.UserID
	}
	_, err := r.db.NewInsert().Table("oauth_clients").Model(&map[string]any{
		"id":            client.ID,
		"user_id":       userID,
		"secret_hash":   secretHash,
		"name":          client.Name,
		"redirect_uris": marshalJSON(client.RedirectURIs),
		"created_at":    now,
	}).Exec(ctx)
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
	err := selectRow(ctx, r.db.NewSelect().
		ColumnExpr("id, user_id, secret_hash, name, redirect_uris, created_at").
		TableExpr("oauth_clients").
		Where("id=?", id)).
		Scan(&client.ID, &userID, &secretHash, &client.Name, &redirectURIs, &createdAt)
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
	rows, err := r.db.NewSelect().
		ColumnExpr("id, user_id, name, redirect_uris, created_at").
		TableExpr("oauth_clients").
		Where("user_id=?", userID).
		OrderExpr("created_at DESC").
		Rows(ctx)
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
	_, err := r.db.NewInsert().Table("oauth_codes").Model(&map[string]any{
		"code":                  code.Code,
		"client_id":             code.ClientID,
		"user_id":               code.UserID,
		"redirect_uri":          code.RedirectURI,
		"scope":                 code.Scope,
		"code_challenge":        code.CodeChallenge,
		"code_challenge_method": code.CodeChallengeMethod,
		"expires_at":            expiresAt,
		"created_at":            now,
	}).Exec(ctx)
	return err
}

func (r *OAuthRepository) FindCode(ctx context.Context, code string) (*OAuthCode, error) {
	var c OAuthCode
	var expiresAt string
	err := selectRow(ctx, r.db.NewSelect().
		ColumnExpr("code, client_id, user_id, redirect_uri, scope, code_challenge, code_challenge_method, expires_at").
		TableExpr("oauth_codes").
		Where("code=?", code)).
		Scan(&c.Code, &c.ClientID, &c.UserID, &c.RedirectURI, &c.Scope, &c.CodeChallenge, &c.CodeChallengeMethod, &expiresAt)
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
	_, err := r.db.NewDelete().Table("oauth_codes").Where("code=?", code).Exec(ctx)
	return err
}

func (r *OAuthRepository) CleanExpiredCodes(ctx context.Context) error {
	now := time.Now().UTC().Format(time.DateTime)
	_, err := r.db.NewDelete().Table("oauth_codes").Where("expires_at < ?", now).Exec(ctx)
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
	_, err := r.db.NewInsert().Table("oauth_tokens").Model(&map[string]any{
		"id":                 token.ID,
		"client_id":          token.ClientID,
		"user_id":            token.UserID,
		"access_token_hash":  token.AccessTokenHash,
		"refresh_token_hash": token.RefreshTokenHash,
		"scope":              token.Scope,
		"expires_at":         expiresAt,
		"created_at":         now,
	}).Exec(ctx)
	return err
}

func (r *OAuthRepository) FindByAccessTokenHash(ctx context.Context, hash string) (*OAuthToken, error) {
	var t OAuthToken
	var expiresAt string
	err := selectRow(ctx, r.db.NewSelect().
		ColumnExpr("id, client_id, user_id, access_token_hash, refresh_token_hash, scope, expires_at").
		TableExpr("oauth_tokens").
		Where("access_token_hash=?", hash)).
		Scan(&t.ID, &t.ClientID, &t.UserID, &t.AccessTokenHash, &t.RefreshTokenHash, &t.Scope, &expiresAt)
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
	err := selectRow(ctx, r.db.NewSelect().
		ColumnExpr("id, client_id, user_id, access_token_hash, refresh_token_hash, scope, expires_at").
		TableExpr("oauth_tokens").
		Where("refresh_token_hash=?", hash)).
		Scan(&t.ID, &t.ClientID, &t.UserID, &t.AccessTokenHash, &t.RefreshTokenHash, &t.Scope, &expiresAt)
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
	_, err := r.db.NewDelete().Table("oauth_tokens").Where("id=?", id).Exec(ctx)
	return err
}

func (r *OAuthRepository) CleanExpiredTokens(ctx context.Context) error {
	now := time.Now().UTC().Format(time.DateTime)
	_, err := r.db.NewDelete().Table("oauth_tokens").Where("expires_at < ?", now).Exec(ctx)
	return err
}
