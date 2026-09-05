package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/dovod-app/app/internal/domain"
	"github.com/google/uuid"
)

// migrateResearchMemory runs under the migration lock, before the ledger entry
// is committed. SQL dialects share the exact same lossless JSON conversion.
func migrateResearchMemory(ctx context.Context, q Querier) error {
	type legacyResearch struct {
		id, instruction string
		memory          sql.NullString
	}
	rows, err := q.NewSelect().Table("researches").Column("id", "instruction", "memory").Rows(ctx)
	if err != nil {
		return err
	}
	var legacy []legacyResearch
	for rows.Next() {
		var r legacyResearch
		if err := rows.Scan(&r.id, &r.instruction, &r.memory); err != nil {
			rows.Close()
			return err
		}
		legacy = append(legacy, r)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return err
	}
	for _, r := range legacy {
		var notes []string
		if r.memory.Valid && r.memory.String != "" {
			if err := json.Unmarshal([]byte(r.memory.String), &notes); err != nil {
				return fmt.Errorf("research %s has invalid legacy memory; migration stopped without discarding it: %w", r.id, err)
			}
		}
		for i, text := range notes {
			item := domain.MemoryItem{Text: text, Author: "unknown"}
			if err := insertMemory(ctx, q, r.id, &item, int64(i)); err != nil {
				return err
			}
		}
		if r.instruction != "" {
			if err := importLegacyInstruction(ctx, q, r.id, r.instruction); err != nil {
				return err
			}
		}
	}
	// No live legacy columns: older binaries must not write a second source of
	// truth. Downgrades require the pre-upgrade backup, not a binary rollback.
	if _, err := q.ExecContext(ctx, "ALTER TABLE researches DROP COLUMN instruction"); err != nil {
		return err
	}
	_, err = q.ExecContext(ctx, "ALTER TABLE researches DROP COLUMN memory")
	return err
}

func importLegacyInstruction(ctx context.Context, q Querier, researchID, body string) error {
	slug := "legacy-instruction"
	for suffix := 2; ; suffix++ {
		exists, err := q.NewSelect().Table("skills").Where("research_id=? AND slug=?", researchID, slug).Exists(ctx)
		if err != nil {
			return err
		}
		if !exists {
			break
		}
		slug = fmt.Sprintf("legacy-instruction-%d", suffix)
	}
	id := uuid.NewString()
	now := time.Now().UTC().Format(time.DateTime)
	_, err := q.NewInsert().Table("skills").Model(&map[string]any{
		"id": id, "research_id": researchID, "slug": slug,
		"name": "Legacy research instruction", "tier": "private",
		"description": "Legacy research instruction — review and describe when to load it.",
		"body":        body, "needs_trigger": 1, "version": 1,
		"created_at": now, "updated_at": now,
	}).Exec(ctx)
	if err != nil {
		return err
	}
	_, err = q.NewInsert().Table("research_skills").Model(&map[string]any{
		"research_id": researchID, "skill_id": id, "via_template": 0, "attached_at": now,
	}).Exec(ctx)
	return err
}
