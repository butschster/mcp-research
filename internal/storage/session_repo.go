package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/butschster/mcp-research/internal/domain"
	"github.com/uptrace/bun"
)

type SessionRepository struct {
	db *bun.DB
}

func NewSessionRepository(db *bun.DB) *SessionRepository {
	return &SessionRepository{db: db}
}

func (r *SessionRepository) Create(ctx context.Context, session *domain.Session) error {
	now := time.Now().UTC().Format(time.DateTime)

	if session.Code == "" {
		code, err := NextCode(ctx, r.db, "sessions", "SS", "research_id", session.ResearchID)
		if err != nil {
			return fmt.Errorf("generate code: %w", err)
		}
		session.Code = code
	}

	_, err := r.db.NewInsert().Table("sessions").Model(&map[string]any{
		"id":          session.ID,
		"code":        session.Code,
		"research_id": session.ResearchID,
		"title":       session.Title,
		"focus":       session.Focus,
		"status":      session.Status,
		"notes":       session.Notes,
		"created_at":  now,
		"updated_at":  now,
	}).Exec(ctx)
	if err != nil {
		return fmt.Errorf("insert session: %w", err)
	}
	session.CreatedAt, _ = time.Parse(time.DateTime, now)
	session.UpdatedAt = session.CreatedAt
	return nil
}

func (r *SessionRepository) CreateTx(ctx context.Context, tx bun.Tx, session *domain.Session) error {
	now := time.Now().UTC().Format(time.DateTime)

	if session.Code == "" {
		code, err := NextCode(ctx, tx, "sessions", "SS", "research_id", session.ResearchID)
		if err != nil {
			return fmt.Errorf("generate code: %w", err)
		}
		session.Code = code
	}

	_, err := tx.NewInsert().Table("sessions").Model(&map[string]any{
		"id":          session.ID,
		"code":        session.Code,
		"research_id": session.ResearchID,
		"title":       session.Title,
		"focus":       session.Focus,
		"status":      session.Status,
		"notes":       session.Notes,
		"created_at":  now,
		"updated_at":  now,
	}).Exec(ctx)
	if err != nil {
		return fmt.Errorf("insert session: %w", err)
	}
	session.CreatedAt, _ = time.Parse(time.DateTime, now)
	session.UpdatedAt = session.CreatedAt
	return nil
}

func (r *SessionRepository) Update(ctx context.Context, session *domain.Session) error {
	now := time.Now().UTC().Format(time.DateTime)
	_, err := r.db.NewUpdate().
		Table("sessions").
		Set("title=?", session.Title).
		Set("focus=?", session.Focus).
		Set("status=?", session.Status).
		Set("notes=?", session.Notes).
		Set("code=?", session.Code).
		Set("updated_at=?", now).
		Where("id=?", session.ID).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("update session: %w", err)
	}
	session.UpdatedAt, _ = time.Parse(time.DateTime, now)
	return nil
}

func (r *SessionRepository) FindByID(ctx context.Context, id string) (*domain.Session, error) {
	row := selectRow(ctx, r.db.NewSelect().
		ColumnExpr("id, code, research_id, title, focus, status, notes, created_at, updated_at").
		TableExpr("sessions").
		Where("id=?", id))
	return r.scanSession(row)
}

func (r *SessionRepository) FindByCode(ctx context.Context, code string) (*domain.Session, error) {
	row := selectRow(ctx, r.db.NewSelect().
		ColumnExpr("id, code, research_id, title, focus, status, notes, created_at, updated_at").
		TableExpr("sessions").
		Where("code=?", code))
	return r.scanSession(row)
}

func (r *SessionRepository) FindByCodeAndResearch(ctx context.Context, code string, researchID string) (*domain.Session, error) {
	row := selectRow(ctx, r.db.NewSelect().
		ColumnExpr("id, code, research_id, title, focus, status, notes, created_at, updated_at").
		TableExpr("sessions").
		Where("code=? AND research_id=?", code, researchID))
	return r.scanSession(row)
}

func (r *SessionRepository) FindByResearch(ctx context.Context, researchID string) ([]*domain.Session, error) {
	rows, err := r.db.NewSelect().
		ColumnExpr("id, code, research_id, title, focus, status, notes, created_at, updated_at").
		TableExpr("sessions").
		Where("research_id=?", researchID).
		OrderExpr("created_at DESC").
		Rows(ctx)
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

func (r *SessionRepository) FindActive(ctx context.Context, researchID string) (*domain.Session, error) {
	row := selectRow(ctx, r.db.NewSelect().
		ColumnExpr("id, code, research_id, title, focus, status, notes, created_at, updated_at").
		TableExpr("sessions").
		Where("research_id=? AND status='active'", researchID).
		Limit(1))
	return r.scanSession(row)
}

func (r *SessionRepository) FindLatest(ctx context.Context, researchID string) (*domain.Session, error) {
	row := selectRow(ctx, r.db.NewSelect().
		ColumnExpr("id, code, research_id, title, focus, status, notes, created_at, updated_at").
		TableExpr("sessions").
		Where("research_id=?", researchID).
		OrderExpr("created_at DESC").
		Limit(1))
	return r.scanSession(row)
}

func (r *SessionRepository) scanSession(row scanner) (*domain.Session, error) {
	var s domain.Session
	var createdAt, updatedAt string
	err := row.Scan(
		&s.ID, &s.Code, &s.ResearchID, &s.Title, &s.Focus,
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
		&s.ID, &s.Code, &s.ResearchID, &s.Title, &s.Focus,
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
