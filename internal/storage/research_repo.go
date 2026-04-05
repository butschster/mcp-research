package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/butschster/mcp-research/internal/domain"
)

type ResearchFilter struct {
	Status *domain.ResearchStatus
	UserID *string
}

type ResearchRepository struct {
	db *sql.DB
}

func NewResearchRepository(db *sql.DB) *ResearchRepository {
	return &ResearchRepository{db: db}
}

func (r *ResearchRepository) Create(ctx context.Context, research *domain.Research) error {
	now := time.Now().UTC().Format(time.DateTime)

	// Auto-assign short code
	if research.Code == "" {
		code, err := NextCodeGlobal(ctx, r.db, "researches", "R")
		if err != nil {
			return fmt.Errorf("generate code: %w", err)
		}
		research.Code = code
	}

	var userID any
	if research.UserID != "" {
		userID = research.UserID
	}

	_, err := r.db.ExecContext(ctx,
		`INSERT INTO researches (id, code, user_id, name, description, goal, status, instruction, memory, tags, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		research.ID, research.Code, userID, research.Name, research.Description, research.Goal,
		research.Status, research.Instruction,
		marshalJSON(research.Memory), marshalJSON(research.Tags),
		now, now,
	)
	if err != nil {
		return fmt.Errorf("insert research: %w", err)
	}
	research.CreatedAt, _ = time.Parse(time.DateTime, now)
	research.UpdatedAt = research.CreatedAt
	return nil
}


func (r *ResearchRepository) Update(ctx context.Context, research *domain.Research) error {
	now := time.Now().UTC().Format(time.DateTime)
	_, err := r.db.ExecContext(ctx,
		`UPDATE researches SET name=?, description=?, goal=?, status=?, instruction=?, memory=?, tags=?, code=?, updated_at=?
		 WHERE id=?`,
		research.Name, research.Description, research.Goal,
		research.Status, research.Instruction,
		marshalJSON(research.Memory), marshalJSON(research.Tags),
		research.Code, now, research.ID,
	)
	if err != nil {
		return fmt.Errorf("update research: %w", err)
	}
	research.UpdatedAt, _ = time.Parse(time.DateTime, now)
	return nil
}

func (r *ResearchRepository) FindByID(ctx context.Context, id string) (*domain.Research, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, code, user_id, name, description, goal, status, instruction, memory, tags, created_at, updated_at
		 FROM researches WHERE id=?`, id)
	return r.scanResearch(row)
}

func (r *ResearchRepository) FindByCode(ctx context.Context, code string) (*domain.Research, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, code, user_id, name, description, goal, status, instruction, memory, tags, created_at, updated_at
		 FROM researches WHERE code=?`, code)
	return r.scanResearch(row)
}

func (r *ResearchRepository) FindAll(ctx context.Context, filter ResearchFilter) ([]*domain.Research, error) {
	query := `SELECT id, code, user_id, name, description, goal, status, instruction, memory, tags, created_at, updated_at FROM researches`
	var args []any
	var conditions []string

	if filter.Status != nil {
		conditions = append(conditions, "status=?")
		args = append(args, *filter.Status)
	}
	if filter.UserID != nil {
		conditions = append(conditions, "user_id=?")
		args = append(args, *filter.UserID)
	}

	if len(conditions) > 0 {
		query += " WHERE " + conditions[0]
		for _, c := range conditions[1:] {
			query += " AND " + c
		}
	}

	query += " ORDER BY created_at DESC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query researches: %w", err)
	}
	defer rows.Close()

	var result []*domain.Research
	for rows.Next() {
		res, err := r.scanResearchRow(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, res)
	}
	return result, rows.Err()
}

func (r *ResearchRepository) Exists(ctx context.Context, id string) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM researches WHERE id=?", id).Scan(&count)
	return count > 0, err
}

func (r *ResearchRepository) ClaimOrphanedResearches(ctx context.Context, userID string) (int64, error) {
	res, err := r.db.ExecContext(ctx, `UPDATE researches SET user_id=? WHERE user_id IS NULL`, userID)
	if err != nil {
		return 0, fmt.Errorf("claim orphaned researches: %w", err)
	}
	return res.RowsAffected()
}

func (r *ResearchRepository) scanResearch(row *sql.Row) (*domain.Research, error) {
	var res domain.Research
	var userID sql.NullString
	var memory, tags sql.NullString
	var createdAt, updatedAt string
	err := row.Scan(
		&res.ID, &res.Code, &userID, &res.Name, &res.Description, &res.Goal,
		&res.Status, &res.Instruction,
		&memory, &tags,
		&createdAt, &updatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("scan research: %w", err)
	}
	if userID.Valid {
		res.UserID = userID.String
	}
	res.Memory = unmarshalStringSlice(memory)
	res.Tags = unmarshalStringSlice(tags)
	res.CreatedAt, _ = time.Parse(time.DateTime, createdAt)
	res.UpdatedAt, _ = time.Parse(time.DateTime, updatedAt)
	return &res, nil
}

func (r *ResearchRepository) scanResearchRow(rows *sql.Rows) (*domain.Research, error) {
	var res domain.Research
	var userID sql.NullString
	var memory, tags sql.NullString
	var createdAt, updatedAt string
	err := rows.Scan(
		&res.ID, &res.Code, &userID, &res.Name, &res.Description, &res.Goal,
		&res.Status, &res.Instruction,
		&memory, &tags,
		&createdAt, &updatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("scan research row: %w", err)
	}
	if userID.Valid {
		res.UserID = userID.String
	}
	res.Memory = unmarshalStringSlice(memory)
	res.Tags = unmarshalStringSlice(tags)
	res.CreatedAt, _ = time.Parse(time.DateTime, createdAt)
	res.UpdatedAt, _ = time.Parse(time.DateTime, updatedAt)
	return &res, nil
}
