package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/butschster/mcp-research/internal/domain"
)

type TaskFilter struct {
	Status   *domain.TaskStatus
	Priority *domain.Priority
}

type TaskRepository struct {
	db *sql.DB
}

func NewTaskRepository(db *sql.DB) *TaskRepository {
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

	_, err := r.db.ExecContext(ctx,
		`INSERT INTO tasks (id, code, research_id, title, description, status, priority, result, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		task.ID, task.Code, task.ResearchID, task.Title, task.Description,
		task.Status, task.Priority, task.Result,
		now, now,
	)
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
	_, err := r.db.ExecContext(ctx,
		`UPDATE tasks SET title=?, description=?, status=?, priority=?, result=?, code=?, updated_at=?, completed_at=?
		 WHERE id=?`,
		task.Title, task.Description, task.Status, task.Priority, task.Result,
		task.Code, now, completedAt, task.ID,
	)
	if err != nil {
		return fmt.Errorf("update task: %w", err)
	}
	task.UpdatedAt, _ = time.Parse(time.DateTime, now)
	return nil
}

func (r *TaskRepository) FindByID(ctx context.Context, id string) (*domain.Task, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, code, research_id, title, description, status, priority, result, created_at, updated_at, completed_at
		 FROM tasks WHERE id=?`, id)
	return r.scanTask(row)
}

func (r *TaskRepository) FindByResearch(ctx context.Context, researchID string, filter TaskFilter) ([]*domain.Task, error) {
	query := `SELECT id, code, research_id, title, description, status, priority, result, created_at, updated_at, completed_at
		 FROM tasks WHERE research_id=?`
	args := []any{researchID}

	if filter.Status != nil {
		query += " AND status=?"
		args = append(args, *filter.Status)
	}
	if filter.Priority != nil {
		query += " AND priority=?"
		args = append(args, *filter.Priority)
	}

	query += " ORDER BY CASE priority WHEN 'high' THEN 0 WHEN 'medium' THEN 1 WHEN 'low' THEN 2 END, created_at ASC"

	rows, err := r.db.QueryContext(ctx, query, args...)
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
	_, err := r.db.ExecContext(ctx, "DELETE FROM tasks WHERE id=?", id)
	return err
}

func (r *TaskRepository) CountByStatus(ctx context.Context, researchID string) (map[domain.TaskStatus]int, error) {
	rows, err := r.db.QueryContext(ctx,
		"SELECT status, COUNT(*) FROM tasks WHERE research_id=? GROUP BY status", researchID)
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

func (r *TaskRepository) scanTask(row *sql.Row) (*domain.Task, error) {
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
