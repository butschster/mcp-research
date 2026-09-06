package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/dovod-app/app/internal/domain"
	"github.com/uptrace/bun"
)

type ExternalLinkRepository struct {
	db *bun.DB
}

func NewExternalLinkRepository(db *bun.DB) *ExternalLinkRepository {
	return &ExternalLinkRepository{db: db}
}

// ReplaceForSource deletes all existing links from this source and inserts new ones.
func (r *ExternalLinkRepository) ReplaceForSource(ctx context.Context, sourceType, sourceID string, links []domain.ExternalLink) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.NewDelete().
		Table("external_links").
		Where("source_type=? AND source_id=?", sourceType, sourceID).
		Exec(ctx); err != nil {
		return fmt.Errorf("delete old links: %w", err)
	}

	if len(links) == 0 {
		return tx.Commit()
	}

	// Inserts below are built by Bun on the same transaction.

	now := time.Now().UTC().Format(time.DateTime)
	for _, link := range links {
		if _, err := tx.NewInsert().Table("external_links").Model(&map[string]any{
			"id":          link.ID,
			"source_type": link.SourceType,
			"source_id":   link.SourceID,
			"research_id": link.ResearchID,
			"url":         link.URL,
			"title":       link.Title,
			"domain":      link.Domain,
			"created_at":  now,
		}).Exec(ctx); err != nil {
			return fmt.Errorf("insert link: %w", err)
		}
	}

	return tx.Commit()
}

// FindByResearch returns all external links for a research, ordered by domain then URL.
func (r *ExternalLinkRepository) FindByResearch(ctx context.Context, researchID string) ([]domain.ExternalLink, error) {
	rows, err := r.db.NewSelect().
		ColumnExpr("l.id, l.source_type, l.source_id, l.research_id, l.url, l.title, l.domain, l.created_at, COALESCE(e.code, ''), COALESCE(e.title, '')").
		TableExpr("external_links l LEFT JOIN entries e ON l.source_type='entry' AND l.source_id=e.id").
		Where("l.research_id=?", researchID).
		OrderExpr("l.domain, l.url").
		Rows(ctx)
	if err != nil {
		return nil, fmt.Errorf("query links: %w", err)
	}
	defer rows.Close()

	return r.scanRows(rows)
}

// FindBySource returns all external links from a specific source.
func (r *ExternalLinkRepository) FindBySource(ctx context.Context, sourceType, sourceID string) ([]domain.ExternalLink, error) {
	rows, err := r.db.NewSelect().
		ColumnExpr("l.id, l.source_type, l.source_id, l.research_id, l.url, l.title, l.domain, l.created_at, COALESCE(e.code, ''), COALESCE(e.title, '')").
		TableExpr("external_links l LEFT JOIN entries e ON l.source_type='entry' AND l.source_id=e.id").
		Where("l.source_type=? AND l.source_id=?", sourceType, sourceID).
		OrderExpr("l.created_at").
		Rows(ctx)
	if err != nil {
		return nil, fmt.Errorf("query links: %w", err)
	}
	defer rows.Close()

	return r.scanRows(rows)
}

func (r *ExternalLinkRepository) scanRows(rows *sql.Rows) ([]domain.ExternalLink, error) {
	var result []domain.ExternalLink
	for rows.Next() {
		var l domain.ExternalLink
		var createdAt string
		if err := rows.Scan(
			&l.ID, &l.SourceType, &l.SourceID, &l.ResearchID,
			&l.URL, &l.Title, &l.Domain, &createdAt,
			&l.EntryCode, &l.EntryTitle,
		); err != nil {
			return nil, fmt.Errorf("scan link: %w", err)
		}
		l.CreatedAt, _ = time.Parse(time.DateTime, createdAt)
		result = append(result, l)
	}
	return result, rows.Err()
}
