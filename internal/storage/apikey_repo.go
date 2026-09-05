package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/dovod-app/app/internal/domain"
	"github.com/uptrace/bun"
)

type APIKeyRepository struct {
	db *bun.DB
}

func NewAPIKeyRepository(db *bun.DB) *APIKeyRepository {
	return &APIKeyRepository{db: db}
}

func (r *APIKeyRepository) Create(ctx context.Context, key *domain.APIKey, keyHash string) error {
	now := time.Now().UTC().Format(time.DateTime)
	_, err := r.db.NewInsert().Table("api_keys").Model(&map[string]any{
		"id":         key.ID,
		"user_id":    key.UserID,
		"name":       key.Name,
		"key_hash":   keyHash,
		"key_prefix": key.KeyPrefix,
		"created_at": now,
	}).Exec(ctx)
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
	err := selectRow(ctx, r.db.NewSelect().
		ColumnExpr("id, user_id, name, key_prefix, last_used_at, expires_at, created_at").
		TableExpr("api_keys").
		Where("key_hash=?", keyHash)).
		Scan(&key.ID, &key.UserID, &key.Name, &key.KeyPrefix, &lastUsed, &expires, &createdAt)
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
	rows, err := r.db.NewSelect().
		ColumnExpr("id, user_id, name, key_prefix, last_used_at, expires_at, created_at").
		TableExpr("api_keys").
		Where("user_id=?", userID).
		OrderExpr("created_at DESC").
		Rows(ctx)
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
	res, err := r.db.NewDelete().Table("api_keys").Where("id=? AND user_id=?", id, userID).Exec(ctx)
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
	r.db.NewUpdate().Table("api_keys").Set("last_used_at=?", now).Where("id=?", id).Exec(ctx)
}
