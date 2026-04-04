package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/butschster/mcp-research/internal/domain"
)

type SessionRepository struct {
	db *sql.DB
}

func NewSessionRepository(db *sql.DB) *SessionRepository {
	return &SessionRepository{db: db}
}

func (r *SessionRepository) Create(ctx context.Context, session *domain.Session) error {
	now := time.Now().UTC().Format(time.DateTime)
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO sessions (id, research_id, title, focus, status, notes, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		session.ID, session.ResearchID, session.Title, session.Focus,
		session.Status, session.Notes,
		now, now,
	)
	if err != nil {
		return fmt.Errorf("insert session: %w", err)
	}
	session.CreatedAt, _ = time.Parse(time.DateTime, now)
	session.UpdatedAt = session.CreatedAt
	return nil
}

func (r *SessionRepository) CreateTx(ctx context.Context, tx *sql.Tx, session *domain.Session) error {
	now := time.Now().UTC().Format(time.DateTime)
	_, err := tx.ExecContext(ctx,
		`INSERT INTO sessions (id, research_id, title, focus, status, notes, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		session.ID, session.ResearchID, session.Title, session.Focus,
		session.Status, session.Notes,
		now, now,
	)
	if err != nil {
		return fmt.Errorf("insert session: %w", err)
	}
	session.CreatedAt, _ = time.Parse(time.DateTime, now)
	session.UpdatedAt = session.CreatedAt
	return nil
}

func (r *SessionRepository) Update(ctx context.Context, session *domain.Session) error {
	now := time.Now().UTC().Format(time.DateTime)
	_, err := r.db.ExecContext(ctx,
		`UPDATE sessions SET title=?, focus=?, status=?, notes=?, updated_at=?
		 WHERE id=?`,
		session.Title, session.Focus, session.Status, session.Notes,
		now, session.ID,
	)
	if err != nil {
		return fmt.Errorf("update session: %w", err)
	}
	session.UpdatedAt, _ = time.Parse(time.DateTime, now)
	return nil
}

func (r *SessionRepository) FindByID(ctx context.Context, id string) (*domain.Session, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, research_id, title, focus, status, notes, created_at, updated_at
		 FROM sessions WHERE id=?`, id)
	return r.scanSession(row)
}

func (r *SessionRepository) FindByResearch(ctx context.Context, researchID string) ([]*domain.Session, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, research_id, title, focus, status, notes, created_at, updated_at
		 FROM sessions WHERE research_id=? ORDER BY created_at DESC`, researchID)
	if err != nil {
		return nil, fmt.Errorf("query sessions: %w", err)
	}
	defer rows.Close()

	var result []*domain.Session
	for rows.Next() {
		s, err := r.scanSessionRow(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, s)
	}
	return result, rows.Err()
}

// FindActive returns the active session for a research, or nil if none.
func (r *SessionRepository) FindActive(ctx context.Context, researchID string) (*domain.Session, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, research_id, title, focus, status, notes, created_at, updated_at
		 FROM sessions WHERE research_id=? AND status='active' LIMIT 1`, researchID)
	return r.scanSession(row)
}

func (r *SessionRepository) scanSession(row *sql.Row) (*domain.Session, error) {
	var s domain.Session
	var createdAt, updatedAt string
	err := row.Scan(
		&s.ID, &s.ResearchID, &s.Title, &s.Focus,
		&s.Status, &s.Notes,
		&createdAt, &updatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("scan session: %w", err)
	}
	s.CreatedAt, _ = time.Parse(time.DateTime, createdAt)
	s.UpdatedAt, _ = time.Parse(time.DateTime, updatedAt)
	return &s, nil
}

func (r *SessionRepository) scanSessionRow(rows *sql.Rows) (*domain.Session, error) {
	var s domain.Session
	var createdAt, updatedAt string
	err := rows.Scan(
		&s.ID, &s.ResearchID, &s.Title, &s.Focus,
		&s.Status, &s.Notes,
		&createdAt, &updatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("scan session row: %w", err)
	}
	s.CreatedAt, _ = time.Parse(time.DateTime, createdAt)
	s.UpdatedAt, _ = time.Parse(time.DateTime, updatedAt)
	return &s, nil
}
