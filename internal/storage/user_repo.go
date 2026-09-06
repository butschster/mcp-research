package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/dovod-app/app/internal/domain"
	"github.com/uptrace/bun"
)

type UserRepository struct {
	db *bun.DB
}

func NewUserRepository(db *bun.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	now := time.Now().UTC().Format(time.DateTime)
	_, err := r.db.NewInsert().Table("users").Model(&map[string]any{
		"id":            user.ID,
		"email":         user.Email,
		"password_hash": user.PasswordHash,
		"name":          user.Name,
		"created_at":    now,
		"updated_at":    now,
	}).Exec(ctx)
	if err != nil {
		return fmt.Errorf("insert user: %w", err)
	}
	user.CreatedAt, _ = time.Parse(time.DateTime, now)
	user.UpdatedAt = user.CreatedAt
	return nil
}

func (r *UserRepository) FindByID(ctx context.Context, id string) (*domain.User, error) {
	row := selectRow(ctx, r.db.NewSelect().
		ColumnExpr("id, email, password_hash, name, created_at, updated_at").
		TableExpr("users").
		Where("id=?", id))
	return r.scanUser(row)
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	row := selectRow(ctx, r.db.NewSelect().
		ColumnExpr("id, email, password_hash, name, created_at, updated_at").
		TableExpr("users").
		Where("email=?", email))
	return r.scanUser(row)
}

func (r *UserRepository) Count(ctx context.Context) (int, error) {
	var count int
	err := selectRow(ctx, r.db.NewSelect().ColumnExpr("COUNT(*)").TableExpr("users")).Scan(&count)
	return count, err
}

func (r *UserRepository) scanUser(row scanner) (*domain.User, error) {
	var u domain.User
	var createdAt, updatedAt string
	err := row.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Name, &createdAt, &updatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("scan user: %w", err)
	}
	u.CreatedAt, _ = time.Parse(time.DateTime, createdAt)
	u.UpdatedAt, _ = time.Parse(time.DateTime, updatedAt)
	return &u, nil
}
