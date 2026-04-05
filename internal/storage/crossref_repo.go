package storage

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/butschster/mcp-research/internal/domain"
)

type CrossRefRepository struct {
	db *sql.DB
}

func NewCrossRefRepository(db *sql.DB) *CrossRefRepository {
	return &CrossRefRepository{db: db}
}

// ReplaceForEntry deletes all existing refs from this entry and inserts new ones.
func (r *CrossRefRepository) ReplaceForEntry(ctx context.Context, entryID string, refs []domain.CrossRef) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, "DELETE FROM crossrefs WHERE source_entry_id=?", entryID); err != nil {
		return fmt.Errorf("delete old crossrefs: %w", err)
	}

	stmt, err := tx.PrepareContext(ctx,
		`INSERT INTO crossrefs (source_entry_id, source_research_id, target_entry_id, target_research_id, target_ref, resolved)
		 VALUES (?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return fmt.Errorf("prepare insert: %w", err)
	}
	defer stmt.Close()

	for _, ref := range refs {
		resolved := 0
		if ref.Resolved {
			resolved = 1
		}
		var targetEntryID, targetResearchID *string
		if ref.TargetEntryID != "" {
			targetEntryID = &ref.TargetEntryID
		}
		if ref.TargetResearchID != "" {
			targetResearchID = &ref.TargetResearchID
		}
		if _, err := stmt.ExecContext(ctx,
			ref.SourceEntryID, ref.SourceResearchID,
			targetEntryID, targetResearchID,
			ref.TargetRef, resolved,
		); err != nil {
			return fmt.Errorf("insert crossref: %w", err)
		}
	}

	return tx.Commit()
}

// FindByResearch returns all cross-references where the source belongs to the given research.
func (r *CrossRefRepository) FindByResearch(ctx context.Context, researchID string) ([]domain.CrossRef, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT source_entry_id, source_research_id,
		        COALESCE(target_entry_id, ''), COALESCE(target_research_id, ''),
		        target_ref, resolved
		 FROM crossrefs WHERE source_research_id=?
		 ORDER BY created_at`, researchID)
	if err != nil {
		return nil, fmt.Errorf("query crossrefs: %w", err)
	}
	defer rows.Close()

	var result []domain.CrossRef
	for rows.Next() {
		var cr domain.CrossRef
		var resolved int
		if err := rows.Scan(
			&cr.SourceEntryID, &cr.SourceResearchID,
			&cr.TargetEntryID, &cr.TargetResearchID,
			&cr.TargetRef, &resolved,
		); err != nil {
			return nil, fmt.Errorf("scan crossref: %w", err)
		}
		cr.Resolved = resolved == 1
		result = append(result, cr)
	}
	return result, rows.Err()
}

// FindByTargetEntry returns all entries that reference the given entry (incoming links).
func (r *CrossRefRepository) FindByTargetEntry(ctx context.Context, entryID string) ([]domain.CrossRef, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT source_entry_id, source_research_id,
		        COALESCE(target_entry_id, ''), COALESCE(target_research_id, ''),
		        target_ref, resolved
		 FROM crossrefs WHERE target_entry_id=?
		 ORDER BY created_at`, entryID)
	if err != nil {
		return nil, fmt.Errorf("query incoming crossrefs: %w", err)
	}
	defer rows.Close()

	var result []domain.CrossRef
	for rows.Next() {
		var cr domain.CrossRef
		var resolved int
		if err := rows.Scan(
			&cr.SourceEntryID, &cr.SourceResearchID,
			&cr.TargetEntryID, &cr.TargetResearchID,
			&cr.TargetRef, &resolved,
		); err != nil {
			return nil, fmt.Errorf("scan crossref: %w", err)
		}
		cr.Resolved = resolved == 1
		result = append(result, cr)
	}
	return result, rows.Err()
}
