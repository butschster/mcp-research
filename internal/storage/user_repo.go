package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/butschster/mcp-research/internal/domain"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	now := time.Now().UTC().Format(time.DateTime)
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO users (id, email, password_hash, name, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		user.ID, user.Email, user.PasswordHash, user.Name, now, now,
	)
	if err != nil {
		return fmt.Errorf("insert user: %w", err)
	}
	user.CreatedAt, _ = time.Parse(time.DateTime, now)
	user.UpdatedAt = user.CreatedAt
	return nil
}

func (r *UserRepository) FindByID(ctx context.Context, id string) (*domain.User, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, email, password_hash, name, created_at, updated_at FROM users WHERE id=?`, id)
	return r.scanUser(row)
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, email, password_hash, name, created_at, updated_at FROM users WHERE email=?`, email)
	return r.scanUser(row)
}

func (r *UserRepository) Count(ctx context.Context) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users").Scan(&count)
	return count, err
}

func (r *UserRepository) scanUser(row *sql.Row) (*domain.User, error) {
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
