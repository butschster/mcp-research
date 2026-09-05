package storage

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/dovod-app/app/internal/domain"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

var ErrMemoryConflict = errors.New("memory item changed; reload before editing")

type MemoryRepository struct{ db *bun.DB }

func NewMemoryRepository(db *bun.DB) *MemoryRepository { return &MemoryRepository{db: db} }

func insertMemory(ctx context.Context, q Querier, researchID string, item *domain.MemoryItem, position int64) error {
	if item.ID == "" {
		item.ID = uuid.NewString()
	}
	item.Version = 1
	var created any
	if item.CreatedAt != nil {
		created = item.CreatedAt.UTC().Format(time.RFC3339Nano)
	}
	_, err := q.NewInsert().Table("research_memory").Model(&map[string]any{
		"id": item.ID, "research_id": researchID, "text": item.Text,
		"created_at": created, "session_id": nullable(item.SessionID),
		"author": item.Author, "version": 1, "position": position,
	}).Exec(ctx)
	return err
}

func (r *MemoryRepository) Create(ctx context.Context, researchID string, item *domain.MemoryItem) error {
	return insertMemory(ctx, r.db, researchID, item, time.Now().UTC().UnixNano())
}

func (r *MemoryRepository) List(ctx context.Context, researchID string) (domain.Memory, error) {
	// Use the same batched projection for both individual and list reads.
	research := &domain.Research{ID: researchID}
	err := r.Hydrate(ctx, []*domain.Research{research})
	return research.Memory, err
}

func (r *MemoryRepository) Hydrate(ctx context.Context, researches []*domain.Research) error {
	if len(researches) == 0 {
		return nil
	}
	ids := make([]string, 0, len(researches))
	byID := make(map[string]*domain.Research, len(researches))
	for _, research := range researches {
		research.Memory = domain.Memory{}
		ids = append(ids, research.ID)
		byID[research.ID] = research
	}
	rows, err := r.db.NewSelect().TableExpr("research_memory m").
		ColumnExpr("m.research_id, m.id, m.text, m.created_at, COALESCE(m.session_id,''), COALESCE(s.code,''), m.author, m.version").
		Join("LEFT JOIN sessions s ON s.id=m.session_id AND s.research_id=m.research_id").
		Where("m.research_id IN (?)", bun.In(ids)).OrderExpr("m.position, m.id").Rows(ctx)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var researchID string
		var item domain.MemoryItem
		var created sql.NullString
		if err := rows.Scan(&researchID, &item.ID, &item.Text, &created, &item.SessionID, &item.SessionCode, &item.Author, &item.Version); err != nil {
			return err
		}
		if created.Valid {
			t, err := time.Parse(time.RFC3339Nano, created.String)
			if err != nil {
				return err
			}
			item.CreatedAt = &t
		}
		byID[researchID].Memory = append(byID[researchID].Memory, item)
	}
	return rows.Err()
}

func (r *MemoryRepository) Update(ctx context.Context, researchID, id, text string, version int) error {
	res, err := r.db.NewUpdate().Table("research_memory").Set("text=?", text).
		Set("version=version+1").Where("research_id=? AND id=? AND version=?", researchID, id, version).Exec(ctx)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		exists, err := r.db.NewSelect().Table("research_memory").Where("research_id=? AND id=?", researchID, id).Exists(ctx)
		if err != nil {
			return err
		}
		if !exists {
			return sql.ErrNoRows
		}
		return ErrMemoryConflict
	}
	return nil
}

// Delete only touches explicitly selected IDs in this research. An append that
// happens concurrently cannot be removed by a stale browser's selection.
func (r *MemoryRepository) Delete(ctx context.Context, researchID string, ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	_, err := r.db.NewDelete().Table("research_memory").Where("research_id=?", researchID).
		Where("id IN (?)", bun.In(ids)).Exec(ctx)
	return err
}
