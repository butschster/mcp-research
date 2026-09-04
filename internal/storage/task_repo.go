package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/butschster/mcp-research/internal/domain"
	"github.com/uptrace/bun"
)

type TaskFilter struct {
	Status   *domain.TaskStatus
	Priority *domain.Priority
}

type TaskRepository struct {
	db *bun.DB
}

func NewTaskRepository(db *bun.DB) *TaskRepository {
	return &TaskRepository{db: db}
}

func (r *TaskRepository) Create(ctx context.Context, task *domain.Task) error {
	now := time.Now().UTC().Format(time.DateTime)

	if task.Code == "" {
		code, err := NextCode(ctx, r.db, "tasks", "T", "research_id", task.ResearchID)
		if err != nil {
			return fmt.Errorf("generate code: %w", err)
		}
		task.Code = code
	}

	_, err := r.db.NewInsert().Table("tasks").Model(&map[string]any{
		"id":          task.ID,
		"code":        task.Code,
		"research_id": task.ResearchID,
		"title":       task.Title,
		"description": task.Description,
		"status":      task.Status,
		"priority":    task.Priority,
		"result":      task.Result,
		"created_at":  now,
		"updated_at":  now,
	}).Exec(ctx)
	if err != nil {
		return fmt.Errorf("insert task: %w", err)
	}
	task.CreatedAt, _ = time.Parse(time.DateTime, now)
	task.UpdatedAt = task.CreatedAt
	return nil
}

func (r *TaskRepository) Update(ctx context.Context, task *domain.Task) error {
	now := time.Now().UTC().Format(time.DateTime)
	var completedAt *string
	if task.CompletedAt != nil {
		s := task.CompletedAt.UTC().Format(time.DateTime)
		completedAt = &s
	}
	_, err := r.db.NewUpdate().
		Table("tasks").
		Set("title=?", task.Title).
		Set("description=?", task.Description).
		Set("status=?", task.Status).
		Set("priority=?", task.Priority).
		Set("result=?", task.Result).
		Set("code=?", task.Code).
		Set("updated_at=?", now).
		Set("completed_at=?", completedAt).
		Where("id=?", task.ID).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("update task: %w", err)
	}
	task.UpdatedAt, _ = time.Parse(time.DateTime, now)
	return nil
}

func (r *TaskRepository) FindByID(ctx context.Context, id string) (*domain.Task, error) {
	row := selectRow(ctx, r.db.NewSelect().
		ColumnExpr("id, code, research_id, title, description, status, priority, result, created_at, updated_at, completed_at").
		TableExpr("tasks").
		Where("id=?", id))
	return r.scanTask(row)
}

// FindByCode resolves a short code inside one research. Codes are scoped to a
// research, so both halves are needed — `T1` names a different task in every
// research that has one.
func (r *TaskRepository) FindByCode(ctx context.Context, researchID, code string) (*domain.Task, error) {
	row := selectRow(ctx, r.db.NewSelect().
		ColumnExpr("id, code, research_id, title, description, status, priority, result, created_at, updated_at, completed_at").
		TableExpr("tasks").
		Where("research_id=? AND code=?", researchID, code))
	return r.scanTask(row)
}

func (r *TaskRepository) FindByResearch(ctx context.Context, researchID string, filter TaskFilter) ([]*domain.Task, error) {
	query := r.db.NewSelect().
		ColumnExpr("id, code, research_id, title, description, status, priority, result, created_at, updated_at, completed_at").
		Table("tasks").
		Where("research_id=?", researchID)

	if filter.Status != nil {
		query.Where("status=?", *filter.Status)
	}
	if filter.Priority != nil {
		query.Where("priority=?", *filter.Priority)
	}

	rows, err := query.OrderExpr("CASE priority WHEN 'high' THEN 0 WHEN 'medium' THEN 1 WHEN 'low' THEN 2 END, created_at ASC").
		Rows(ctx)
	if err != nil {
		return nil, fmt.Errorf("query tasks: %w", err)
	}
	defer rows.Close()

	var result []*domain.Task
	for rows.Next() {
		t, err := r.scanTaskRow(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, t)
	}
	return result, rows.Err()
}

func (r *TaskRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.NewDelete().Table("tasks").Where("id=?", id).Exec(ctx)
	return err
}

func (r *TaskRepository) CountByStatus(ctx context.Context, researchID string) (map[domain.TaskStatus]int, error) {
	rows, err := r.db.NewSelect().
		ColumnExpr("status, COUNT(*)").
		TableExpr("tasks").
		Where("research_id=?", researchID).
		GroupExpr("status").
		Rows(ctx)
	if err != nil {
		return nil, fmt.Errorf("count tasks: %w", err)
	}
	defer rows.Close()

	result := make(map[domain.TaskStatus]int)
	for rows.Next() {
		var status domain.TaskStatus
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			return nil, fmt.Errorf("scan count: %w", err)
		}
		result[status] = count
	}
	return result, rows.Err()
}

func (r *TaskRepository) scanTask(row scanner) (*domain.Task, error) {
	var t domain.Task
	var createdAt, updatedAt string
	var completedAt sql.NullString

	err := row.Scan(
		&t.ID, &t.Code, &t.ResearchID, &t.Title, &t.Description,
		&t.Status, &t.Priority, &t.Result,
		&createdAt, &updatedAt, &completedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("scan task: %w", err)
	}
	t.CreatedAt, _ = time.Parse(time.DateTime, createdAt)
	t.UpdatedAt, _ = time.Parse(time.DateTime, updatedAt)
	if completedAt.Valid {
		ct, _ := time.Parse(time.DateTime, completedAt.String)
		t.CompletedAt = &ct
	}
	return &t, nil
}

func (r *TaskRepository) scanTaskRow(rows *sql.Rows) (*domain.Task, error) {
	var t domain.Task
	var createdAt, updatedAt string
	var completedAt sql.NullString

	err := rows.Scan(
		&t.ID, &t.Code, &t.ResearchID, &t.Title, &t.Description,
		&t.Status, &t.Priority, &t.Result,
		&createdAt, &updatedAt, &completedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("scan task row: %w", err)
	}
	t.CreatedAt, _ = time.Parse(time.DateTime, createdAt)
	t.UpdatedAt, _ = time.Parse(time.DateTime, updatedAt)
	if completedAt.Valid {
		ct, _ := time.Parse(time.DateTime, completedAt.String)
		t.CompletedAt = &ct
	}
	return &t, nil
}
