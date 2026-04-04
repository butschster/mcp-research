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
}

type ResearchRepository struct {
	db *sql.DB
}

func NewResearchRepository(db *sql.DB) *ResearchRepository {
	return &ResearchRepository{db: db}
}

func (r *ResearchRepository) Create(ctx context.Context, research *domain.Research) error {
	now := time.Now().UTC().Format(time.DateTime)
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO researches (id, name, description, goal, status, instruction, memory, tags, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		research.ID, research.Name, research.Description, research.Goal,
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
		`UPDATE researches SET name=?, description=?, goal=?, status=?, instruction=?, memory=?, tags=?, updated_at=?
		 WHERE id=?`,
		research.Name, research.Description, research.Goal,
		research.Status, research.Instruction,
		marshalJSON(research.Memory), marshalJSON(research.Tags),
		now, research.ID,
	)
	if err != nil {
		return fmt.Errorf("update research: %w", err)
	}
	research.UpdatedAt, _ = time.Parse(time.DateTime, now)
	return nil
}

func (r *ResearchRepository) FindByID(ctx context.Context, id string) (*domain.Research, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, name, description, goal, status, instruction, memory, tags, created_at, updated_at
		 FROM researches WHERE id=?`, id)
	return r.scanResearch(row)
}

func (r *ResearchRepository) FindAll(ctx context.Context, filter ResearchFilter) ([]*domain.Research, error) {
	query := `SELECT id, name, description, goal, status, instruction, memory, tags, created_at, updated_at FROM researches`
	var args []any

	if filter.Status != nil {
		query += " WHERE status=?"
		args = append(args, *filter.Status)
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

func (r *ResearchRepository) scanResearch(row *sql.Row) (*domain.Research, error) {
	var res domain.Research
	var memory, tags sql.NullString
	var createdAt, updatedAt string
	err := row.Scan(
		&res.ID, &res.Name, &res.Description, &res.Goal,
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
	res.Memory = unmarshalStringSlice(memory)
	res.Tags = unmarshalStringSlice(tags)
	res.CreatedAt, _ = time.Parse(time.DateTime, createdAt)
	res.UpdatedAt, _ = time.Parse(time.DateTime, updatedAt)
	return &res, nil
}

func (r *ResearchRepository) scanResearchRow(rows *sql.Rows) (*domain.Research, error) {
	var res domain.Research
	var memory, tags sql.NullString
	var createdAt, updatedAt string
	err := rows.Scan(
		&res.ID, &res.Name, &res.Description, &res.Goal,
		&res.Status, &res.Instruction,
		&memory, &tags,
		&createdAt, &updatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("scan research row: %w", err)
	}
	res.Memory = unmarshalStringSlice(memory)
	res.Tags = unmarshalStringSlice(tags)
	res.CreatedAt, _ = time.Parse(time.DateTime, createdAt)
	res.UpdatedAt, _ = time.Parse(time.DateTime, updatedAt)
	return &res, nil
}
