package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/butschster/mcp-research/internal/domain"
)

type ExternalLinkRepository struct {
	db *sql.DB
}

func NewExternalLinkRepository(db *sql.DB) *ExternalLinkRepository {
	return &ExternalLinkRepository{db: db}
}

// ReplaceForSource deletes all existing links from this source and inserts new ones.
func (r *ExternalLinkRepository) ReplaceForSource(ctx context.Context, sourceType, sourceID string, links []domain.ExternalLink) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, "DELETE FROM external_links WHERE source_type=? AND source_id=?", sourceType, sourceID); err != nil {
		return fmt.Errorf("delete old links: %w", err)
	}

	if len(links) == 0 {
		return tx.Commit()
	}

	stmt, err := tx.PrepareContext(ctx,
		`INSERT INTO external_links (id, source_type, source_id, research_id, url, title, domain, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return fmt.Errorf("prepare insert: %w", err)
	}
	defer stmt.Close()

	now := time.Now().UTC().Format(time.DateTime)
	for _, link := range links {
		if _, err := stmt.ExecContext(ctx,
			link.ID, link.SourceType, link.SourceID, link.ResearchID,
			link.URL, link.Title, link.Domain, now,
		); err != nil {
			return fmt.Errorf("insert link: %w", err)
		}
	}

	return tx.Commit()
}

// FindByResearch returns all external links for a research, ordered by domain then URL.
func (r *ExternalLinkRepository) FindByResearch(ctx context.Context, researchID string) ([]domain.ExternalLink, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT l.id, l.source_type, l.source_id, l.research_id, l.url, l.title, l.domain, l.created_at,
		        COALESCE(e.code, ''), COALESCE(e.title, '')
		 FROM external_links l
		 LEFT JOIN entries e ON l.source_type='entry' AND l.source_id=e.id
		 WHERE l.research_id=?
		 ORDER BY l.domain, l.url`, researchID)
	if err != nil {
		return nil, fmt.Errorf("query links: %w", err)
	}
	defer rows.Close()

	return r.scanRows(rows)
}

// FindBySource returns all external links from a specific source.
func (r *ExternalLinkRepository) FindBySource(ctx context.Context, sourceType, sourceID string) ([]domain.ExternalLink, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT l.id, l.source_type, l.source_id, l.research_id, l.url, l.title, l.domain, l.created_at,
		        COALESCE(e.code, ''), COALESCE(e.title, '')
		 FROM external_links l
		 LEFT JOIN entries e ON l.source_type='entry' AND l.source_id=e.id
		 WHERE l.source_type=? AND l.source_id=?
		 ORDER BY l.created_at`, sourceType, sourceID)
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
