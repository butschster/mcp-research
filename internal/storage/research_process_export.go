package storage

import (
	"context"
	"time"

	"github.com/dovod-app/app/internal/domain"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

func (r *ResearchRepository) ExportPrivateSkills(ctx context.Context, researchID string) ([]domain.ExportPrivateSkill, error) {
	rows, err := r.db.NewSelect().TableExpr("skills s").
		ColumnExpr("s.slug, s.name, s.description, s.body, s.needs_trigger, CASE WHEN rs.skill_id IS NULL THEN 0 ELSE 1 END").
		Join("LEFT JOIN research_skills rs ON rs.skill_id=s.id AND rs.research_id=s.research_id").
		Where("s.research_id=?", researchID).OrderExpr("s.slug").Rows(ctx)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	skills := []domain.ExportPrivateSkill{}
	for rows.Next() {
		var skill domain.ExportPrivateSkill
		if err := rows.Scan(&skill.Slug, &skill.Name, &skill.Description, &skill.Body, &skill.NeedsTrigger, &skill.Attached); err != nil {
			return nil, err
		}
		skills = append(skills, skill)
	}
	return skills, rows.Err()
}

// ImportProcess is used only by portable import after access checking and
// session remapping. It preserves legacy content even above current UI caps.
func (r *ResearchRepository) ImportProcess(ctx context.Context, researchID, instruction string, memory domain.Memory, skills []domain.ExportPrivateSkill) error {
	return r.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		for i, item := range memory {
			item.ID = uuid.NewString()
			if item.Author == "" {
				item.Author = "unknown"
			}
			if err := insertMemory(ctx, tx, researchID, &item, int64(i)); err != nil {
				return err
			}
		}
		now := time.Now().UTC().Format(time.DateTime)
		for _, skill := range skills {
			id := uuid.NewString()
			_, err := tx.NewInsert().Table("skills").Model(&map[string]any{
				"id": id, "research_id": researchID, "slug": skill.Slug, "name": skill.Name,
				"description": skill.Description, "body": skill.Body, "tier": "private",
				"needs_trigger": boolToInt(skill.NeedsTrigger), "version": 1,
				"created_at": now, "updated_at": now,
			}).Exec(ctx)
			if err != nil {
				return err
			}
			if skill.Attached {
				_, err = tx.NewInsert().Table("research_skills").Model(&map[string]any{
					"research_id": researchID, "skill_id": id, "via_template": 0, "attached_at": now,
				}).Exec(ctx)
				if err != nil {
					return err
				}
			}
		}
		if instruction != "" {
			return importLegacyInstruction(ctx, tx, researchID, instruction)
		}
		return nil
	})
}
