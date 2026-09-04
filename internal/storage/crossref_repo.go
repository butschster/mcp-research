package storage

import (
	"context"
	"fmt"

	"github.com/butschster/mcp-research/internal/domain"
	"github.com/uptrace/bun"
)

type CrossRefRepository struct {
	db *bun.DB
}

func NewCrossRefRepository(db *bun.DB) *CrossRefRepository {
	return &CrossRefRepository{db: db}
}

// ReplaceForSource deletes all existing refs from this source and inserts new ones.
func (r *CrossRefRepository) ReplaceForSource(ctx context.Context, sourceType, sourceID string, refs []domain.CrossRef) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.NewDelete().
		Table("crossrefs").
		Where("source_type=? AND source_id=?", sourceType, sourceID).
		Exec(ctx); err != nil {
		return fmt.Errorf("delete old crossrefs: %w", err)
	}

	// Inserts below are built by Bun on the same transaction.

	for _, ref := range refs {
		resolved := 0
		if ref.Resolved {
			resolved = 1
		}
		var targetEntryID, targetResearchID, targetRoadmapID, targetNodeID *string
		if ref.TargetEntryID != "" {
			targetEntryID = &ref.TargetEntryID
		}
		if ref.TargetResearchID != "" {
			targetResearchID = &ref.TargetResearchID
		}
		if ref.TargetRoadmapID != "" {
			targetRoadmapID = &ref.TargetRoadmapID
		}
		if ref.TargetNodeID != "" {
			targetNodeID = &ref.TargetNodeID
		}
		// source_entry_id kept for backward compat (NULL for non-entry sources)
		var sourceEntryID *string
		if ref.SourceType == "entry" {
			sourceEntryID = &ref.SourceID
		}
		if _, err := tx.NewInsert().Table("crossrefs").Model(&map[string]any{
			"source_type":        ref.SourceType,
			"source_id":          ref.SourceID,
			"source_entry_id":    sourceEntryID,
			"source_research_id": ref.SourceResearchID,
			"target_entry_id":    targetEntryID,
			"target_research_id": targetResearchID,
			"target_roadmap_id":  targetRoadmapID,
			"target_node_id":     targetNodeID,
			"target_ref":         ref.TargetRef,
			"resolved":           resolved,
		}).Exec(ctx); err != nil {
			return fmt.Errorf("insert crossref: %w", err)
		}
	}

	return tx.Commit()
}

// FindByResearch returns all cross-references where the source belongs to the given research.
func (r *CrossRefRepository) FindByResearch(ctx context.Context, researchID string) ([]domain.CrossRef, error) {
	rows, err := r.db.NewSelect().
		ColumnExpr("source_type, source_id, source_research_id, COALESCE(target_entry_id, ''), COALESCE(target_research_id, ''), COALESCE(target_roadmap_id, ''), COALESCE(target_node_id, ''), target_ref, resolved").
		TableExpr("crossrefs").
		Where("source_research_id=?", researchID).
		OrderExpr("created_at").
		Rows(ctx)
	if err != nil {
		return nil, fmt.Errorf("query crossrefs: %w", err)
	}
	defer rows.Close()

	var result []domain.CrossRef
	for rows.Next() {
		var cr domain.CrossRef
		var resolved int
		if err := rows.Scan(
			&cr.SourceType, &cr.SourceID, &cr.SourceResearchID,
			&cr.TargetEntryID, &cr.TargetResearchID,
			&cr.TargetRoadmapID, &cr.TargetNodeID,
			&cr.TargetRef, &resolved,
		); err != nil {
			return nil, fmt.Errorf("scan crossref: %w", err)
		}
		cr.Resolved = resolved == 1
		result = append(result, cr)
	}
	return result, rows.Err()
}

// FindBySourceEntry returns all cross-references where the given entry is the source (outgoing links).
func (r *CrossRefRepository) FindBySourceEntry(ctx context.Context, entryID string) ([]domain.CrossRef, error) {
	rows, err := r.db.NewSelect().
		ColumnExpr("source_type, source_id, source_research_id, COALESCE(target_entry_id, ''), COALESCE(target_research_id, ''), COALESCE(target_roadmap_id, ''), COALESCE(target_node_id, ''), target_ref, resolved").
		TableExpr("crossrefs").
		Where("source_type='entry' AND source_id=?", entryID).
		OrderExpr("created_at").
		Rows(ctx)
	if err != nil {
		return nil, fmt.Errorf("query outgoing crossrefs: %w", err)
	}
	defer rows.Close()

	var result []domain.CrossRef
	for rows.Next() {
		var cr domain.CrossRef
		var resolved int
		if err := rows.Scan(
			&cr.SourceType, &cr.SourceID, &cr.SourceResearchID,
			&cr.TargetEntryID, &cr.TargetResearchID,
			&cr.TargetRoadmapID, &cr.TargetNodeID,
			&cr.TargetRef, &resolved,
		); err != nil {
			return nil, fmt.Errorf("scan crossref: %w", err)
		}
		cr.Resolved = resolved == 1
		result = append(result, cr)
	}
	return result, rows.Err()
}

// FindByTargetEntry returns all sources that reference the given entry (incoming links).
func (r *CrossRefRepository) FindByTargetEntry(ctx context.Context, entryID string) ([]domain.CrossRef, error) {
	rows, err := r.db.NewSelect().
		ColumnExpr("source_type, source_id, source_research_id, COALESCE(target_entry_id, ''), COALESCE(target_research_id, ''), COALESCE(target_roadmap_id, ''), COALESCE(target_node_id, ''), target_ref, resolved").
		TableExpr("crossrefs").
		Where("target_entry_id=?", entryID).
		OrderExpr("created_at").
		Rows(ctx)
	if err != nil {
		return nil, fmt.Errorf("query incoming crossrefs: %w", err)
	}
	defer rows.Close()

	var result []domain.CrossRef
	for rows.Next() {
		var cr domain.CrossRef
		var resolved int
		if err := rows.Scan(
			&cr.SourceType, &cr.SourceID, &cr.SourceResearchID,
			&cr.TargetEntryID, &cr.TargetResearchID,
			&cr.TargetRoadmapID, &cr.TargetNodeID,
			&cr.TargetRef, &resolved,
		); err != nil {
			return nil, fmt.Errorf("scan crossref: %w", err)
		}
		cr.Resolved = resolved == 1
		result = append(result, cr)
	}
	return result, rows.Err()
}
