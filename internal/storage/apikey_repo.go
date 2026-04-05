package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/butschster/mcp-research/internal/domain"
)

type APIKeyRepository struct {
	db *sql.DB
}

func NewAPIKeyRepository(db *sql.DB) *APIKeyRepository {
	return &APIKeyRepository{db: db}
}

func (r *APIKeyRepository) Create(ctx context.Context, key *domain.APIKey, keyHash string) error {
	now := time.Now().UTC().Format(time.DateTime)
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO api_keys (id, user_id, name, key_hash, key_prefix, created_at)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		key.ID, key.UserID, key.Name, keyHash, key.KeyPrefix, now,
	)
	if err != nil {
		return fmt.Errorf("insert api key: %w", err)
	}
	key.CreatedAt, _ = time.Parse(time.DateTime, now)
	return nil
}

func (r *APIKeyRepository) FindByHash(ctx context.Context, keyHash string) (*domain.APIKey, error) {
	var key domain.APIKey
	var lastUsed, expires sql.NullString
	var createdAt string
	err := r.db.QueryRowContext(ctx,
		`SELECT id, user_id, name, key_prefix, last_used_at, expires_at, created_at
		 FROM api_keys WHERE key_hash=?`, keyHash,
	).Scan(&key.ID, &key.UserID, &key.Name, &key.KeyPrefix, &lastUsed, &expires, &createdAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("scan api key: %w", err)
	}
	key.CreatedAt, _ = time.Parse(time.DateTime, createdAt)
	if lastUsed.Valid {
		t, _ := time.Parse(time.DateTime, lastUsed.String)
		key.LastUsedAt = &t
	}
	if expires.Valid {
		t, _ := time.Parse(time.DateTime, expires.String)
		key.ExpiresAt = &t
	}
	return &key, nil
}

func (r *APIKeyRepository) ListByUser(ctx context.Context, userID string) ([]*domain.APIKey, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, user_id, name, key_prefix, last_used_at, expires_at, created_at
		 FROM api_keys WHERE user_id=? ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, fmt.Errorf("query api keys: %w", err)
	}
	defer rows.Close()

	var result []*domain.APIKey
	for rows.Next() {
		var key domain.APIKey
		var lastUsed, expires sql.NullString
		var createdAt string
		if err := rows.Scan(&key.ID, &key.UserID, &key.Name, &key.KeyPrefix, &lastUsed, &expires, &createdAt); err != nil {
			return nil, fmt.Errorf("scan api key row: %w", err)
		}
		key.CreatedAt, _ = time.Parse(time.DateTime, createdAt)
		if lastUsed.Valid {
			t, _ := time.Parse(time.DateTime, lastUsed.String)
			key.LastUsedAt = &t
		}
		if expires.Valid {
			t, _ := time.Parse(time.DateTime, expires.String)
			key.ExpiresAt = &t
		}
		result = append(result, &key)
	}
	return result, rows.Err()
}

func (r *APIKeyRepository) Delete(ctx context.Context, id, userID string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM api_keys WHERE id=? AND user_id=?`, id, userID)
	if err != nil {
		return fmt.Errorf("delete api key: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return fmt.Errorf("api key not found")
	}
	return nil
}

func (r *APIKeyRepository) TouchLastUsed(ctx context.Context, id string) {
	now := time.Now().UTC().Format(time.DateTime)
	r.db.ExecContext(ctx, `UPDATE api_keys SET last_used_at=? WHERE id=?`, now, id)
}
