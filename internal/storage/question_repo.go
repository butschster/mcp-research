package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/butschster/mcp-research/internal/domain"
)

type QuestionFilter struct {
	Status   *domain.QuestionStatus
	Area     *string
	Priority *domain.Priority
}

type QuestionRepository struct {
	db *sql.DB
}

func NewQuestionRepository(db *sql.DB) *QuestionRepository {
	return &QuestionRepository{db: db}
}

func (r *QuestionRepository) CreateBatch(ctx context.Context, questions []*domain.Question) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	for _, q := range questions {
		if q.ParentID != "" {
			continue
		}
		if err := r.insertTx(ctx, tx, q); err != nil {
			return err
		}
	}
	for _, q := range questions {
		if q.ParentID == "" {
			continue
		}
		if err := r.insertTx(ctx, tx, q); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *QuestionRepository) CreateBatchTx(ctx context.Context, tx *sql.Tx, questions []*domain.Question) error {
	for _, q := range questions {
		if q.ParentID != "" {
			continue
		}
		if err := r.insertTx(ctx, tx, q); err != nil {
			return err
		}
	}
	for _, q := range questions {
		if q.ParentID == "" {
			continue
		}
		if err := r.insertTx(ctx, tx, q); err != nil {
			return err
		}
	}
	return nil
}

func (r *QuestionRepository) insertTx(ctx context.Context, tx *sql.Tx, q *domain.Question) error {
	now := time.Now().UTC().Format(time.DateTime)

	if q.Code == "" {
		code, err := NextCode(ctx, tx, "questions", "Q", "session_id", q.SessionID)
		if err != nil {
			return fmt.Errorf("generate code: %w", err)
		}
		q.Code = code
	}

	var parentID *string
	if q.ParentID != "" {
		parentID = &q.ParentID
	}
	_, err := tx.ExecContext(ctx,
		`INSERT INTO questions (id, code, session_id, text, area, rationale, priority, status, answer, parent_id, position, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		q.ID, q.Code, q.SessionID, q.Text, q.Area, q.Rationale,
		q.Priority, q.Status, q.Answer, parentID, q.Position,
		now, now,
	)
	if err != nil {
		return fmt.Errorf("insert question: %w", err)
	}
	q.CreatedAt, _ = time.Parse(time.DateTime, now)
	q.UpdatedAt = q.CreatedAt
	return nil
}

func (r *QuestionRepository) Update(ctx context.Context, question *domain.Question) error {
	now := time.Now().UTC().Format(time.DateTime)
	var parentID *string
	if question.ParentID != "" {
		parentID = &question.ParentID
	}
	_, err := r.db.ExecContext(ctx,
		`UPDATE questions SET text=?, area=?, rationale=?, priority=?, status=?, answer=?, parent_id=?, position=?, code=?, updated_at=?
		 WHERE id=?`,
		question.Text, question.Area, question.Rationale,
		question.Priority, question.Status, question.Answer,
		parentID, question.Position, question.Code, now, question.ID,
	)
	if err != nil {
		return fmt.Errorf("update question: %w", err)
	}
	question.UpdatedAt, _ = time.Parse(time.DateTime, now)
	return nil
}

func (r *QuestionRepository) FindByID(ctx context.Context, id string) (*domain.Question, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, code, session_id, text, area, rationale, priority, status, answer, parent_id, position, created_at, updated_at
		 FROM questions WHERE id=?`, id)
	return r.scanQuestion(row)
}

func (r *QuestionRepository) FindBySession(ctx context.Context, sessionID string, filter QuestionFilter) ([]*domain.Question, error) {
	query := `SELECT id, code, session_id, text, area, rationale, priority, status, answer, parent_id, position, created_at, updated_at
		 FROM questions WHERE session_id=?`
	args := []any{sessionID}

	if filter.Status != nil {
		query += " AND status=?"
		args = append(args, *filter.Status)
	}
	if filter.Area != nil {
		query += " AND area=?"
		args = append(args, *filter.Area)
	}
	if filter.Priority != nil {
		query += " AND priority=?"
		args = append(args, *filter.Priority)
	}

	query += " ORDER BY position ASC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query questions: %w", err)
	}
	defer rows.Close()

	var result []*domain.Question
	for rows.Next() {
		q, err := r.scanQuestionRow(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, q)
	}
	return result, rows.Err()
}

func (r *QuestionRepository) CountByStatus(ctx context.Context, sessionID string) (map[domain.QuestionStatus]int, error) {
	rows, err := r.db.QueryContext(ctx,
		"SELECT status, COUNT(*) FROM questions WHERE session_id=? GROUP BY status", sessionID)
	if err != nil {
		return nil, fmt.Errorf("count questions: %w", err)
	}
	defer rows.Close()

	result := make(map[domain.QuestionStatus]int)
	for rows.Next() {
		var status domain.QuestionStatus
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			return nil, fmt.Errorf("scan count: %w", err)
		}
		result[status] = count
	}
	return result, rows.Err()
}

func (r *QuestionRepository) GetDepth(ctx context.Context, questionID string) (int, error) {
	depth := 0
	currentID := questionID

	for i := 0; i < 10; i++ {
		var parentID sql.NullString
		err := r.db.QueryRowContext(ctx, "SELECT parent_id FROM questions WHERE id=?", currentID).Scan(&parentID)
		if err != nil {
			return 0, fmt.Errorf("get parent: %w", err)
		}
		if !parentID.Valid || parentID.String == "" {
			return depth, nil
		}
		depth++
		currentID = parentID.String
	}

	return depth, nil
}

func (r *QuestionRepository) scanQuestion(row *sql.Row) (*domain.Question, error) {
	var q domain.Question
	var parentID sql.NullString
	var createdAt, updatedAt string

	err := row.Scan(
		&q.ID, &q.Code, &q.SessionID, &q.Text, &q.Area, &q.Rationale,
		&q.Priority, &q.Status, &q.Answer,
		&parentID, &q.Position,
		&createdAt, &updatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("scan question: %w", err)
	}
	if parentID.Valid {
		q.ParentID = parentID.String
	}
	q.CreatedAt, _ = time.Parse(time.DateTime, createdAt)
	q.UpdatedAt, _ = time.Parse(time.DateTime, updatedAt)
	return &q, nil
}

func (r *QuestionRepository) scanQuestionRow(rows *sql.Rows) (*domain.Question, error) {
	var q domain.Question
	var parentID sql.NullString
	var createdAt, updatedAt string

	err := rows.Scan(
		&q.ID, &q.Code, &q.SessionID, &q.Text, &q.Area, &q.Rationale,
		&q.Priority, &q.Status, &q.Answer,
		&parentID, &q.Position,
		&createdAt, &updatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("scan question row: %w", err)
	}
	if parentID.Valid {
		q.ParentID = parentID.String
	}
	q.CreatedAt, _ = time.Parse(time.DateTime, createdAt)
	q.UpdatedAt, _ = time.Parse(time.DateTime, updatedAt)
	return &q, nil
}
