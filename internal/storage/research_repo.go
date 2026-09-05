package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/butschster/mcp-research/internal/domain"
	"github.com/uptrace/bun"
)

type ResearchFilter struct {
	Status *domain.ResearchStatus
	// UserID filters by creator — the person whose name is on the research.
	UserID *string
	// MemberOf filters by access: researches owned by a team this user belongs
	// to. This is what the research list uses; UserID would hide researches a
	// colleague created in a shared team.
	MemberOf *string
	// TeamID narrows the list to one team. It composes with MemberOf rather
	// than replacing it, so naming a team you are not in returns nothing
	// instead of someone else's work.
	TeamID *string
}

// researchColumns is the projection every research read shares. It is one
// constant because the column order has to match the scanner, and three
// hand-written copies is how a new column ends up read into the wrong field.
const researchColumns = `id, code, user_id, team_id, name, description, goal, status, tags, created_at, updated_at, template_slug, template_version`

type ResearchRepository struct {
	db *bun.DB
}

func NewResearchRepository(db *bun.DB) *ResearchRepository {
	return &ResearchRepository{db: db}
}

func (r *ResearchRepository) DB() *bun.DB { return r.db }

func (r *ResearchRepository) Create(ctx context.Context, research *domain.Research) error {
	now := time.Now().UTC().Format(time.DateTime)

	// Auto-assign short code
	if research.Code == "" {
		code, err := NextCodeGlobal(ctx, r.db, "researches", "R")
		if err != nil {
			return fmt.Errorf("generate code: %w", err)
		}
		research.Code = code
	}

	var userID any
	if research.UserID != "" {
		userID = research.UserID
	}
	var teamID any
	if research.TeamID != "" {
		teamID = research.TeamID
	}

	err := r.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		_, err := tx.NewInsert().Table("researches").Model(&map[string]any{
			"id":          research.ID,
			"code":        research.Code,
			"user_id":     userID,
			"team_id":     teamID,
			"name":        research.Name,
			"description": research.Description,
			"goal":        research.Goal,
			"status":      research.Status,
			"tags":        marshalJSON(research.Tags),
			"created_at":  now,
			"updated_at":  now,
		}).Exec(ctx)
		if err != nil {
			return err
		}
		for i := range research.Memory {
			if err := insertMemory(ctx, tx, research.ID, &research.Memory[i], int64(i)); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("insert research: %w", err)
	}
	research.CreatedAt, _ = time.Parse(time.DateTime, now)
	research.UpdatedAt = research.CreatedAt
	return nil
}

func (r *ResearchRepository) Update(ctx context.Context, research *domain.Research) error {
	return r.UpdateWithMemory(ctx, research, nil)
}

// UpdateWithMemory commits combined metadata/add_memory requests atomically.
// Memory already loaded in research is never written back.
func (r *ResearchRepository) UpdateWithMemory(ctx context.Context, research *domain.Research, item *domain.MemoryItem) error {
	now := time.Now().UTC().Format(time.DateTime)
	err := r.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		_, err := tx.NewUpdate().
			Table("researches").
			Set("name=?", research.Name).
			Set("description=?", research.Description).
			Set("goal=?", research.Goal).
			Set("status=?", research.Status).
			Set("tags=?", marshalJSON(research.Tags)).
			Set("code=?", research.Code).
			Set("updated_at=?", now).
			Where("id=?", research.ID).
			Exec(ctx)
		if err != nil {
			return err
		}
		if item != nil {
			return insertMemory(ctx, tx, research.ID, item, time.Now().UTC().UnixNano())
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("update research: %w", err)
	}
	research.UpdatedAt, _ = time.Parse(time.DateTime, now)
	return nil
}

func (r *ResearchRepository) FindByID(ctx context.Context, id string) (*domain.Research, error) {
	row := selectRow(ctx, r.db.NewSelect().ColumnExpr(researchColumns).TableExpr("researches").Where("id=?", id))
	res, err := r.scanResearch(row)
	if err == nil && res != nil {
		err = NewMemoryRepository(r.db).Hydrate(ctx, []*domain.Research{res})
	}
	return res, err
}

func (r *ResearchRepository) FindByCode(ctx context.Context, code string) (*domain.Research, error) {
	row := selectRow(ctx, r.db.NewSelect().ColumnExpr(researchColumns).TableExpr("researches").Where("code=?", code))
	res, err := r.scanResearch(row)
	if err == nil && res != nil {
		err = NewMemoryRepository(r.db).Hydrate(ctx, []*domain.Research{res})
	}
	return res, err
}

func (r *ResearchRepository) FindAll(ctx context.Context, filter ResearchFilter) ([]*domain.Research, error) {
	query := r.db.NewSelect().ColumnExpr(researchColumns).Table("researches")

	if filter.Status != nil {
		query.Where("status=?", *filter.Status)
	}
	if filter.UserID != nil {
		query.Where("user_id=?", *filter.UserID)
	}
	if filter.MemberOf != nil {
		// The local team is included for every authenticated caller: it holds
		// researches created with no user at all, which everybody could see
		// before teams and which would otherwise vanish from every list.
		query.Where("(team_id = 'team-local' OR team_id IN (?))",
			r.db.NewSelect().Column("team_id").Table("team_members").Where("user_id=?", *filter.MemberOf))
	}
	if filter.TeamID != nil {
		query.Where("team_id=?", *filter.TeamID)
	}

	rows, err := query.Order("created_at DESC").Rows(ctx)
	if err != nil {
		return nil, fmt.Errorf("query researches: %w", err)
	}
	defer rows.Close()

	var result []*domain.Research
	for rows.Next() {
		res, err := r.scanResearchRow(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, res)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	return result, NewMemoryRepository(r.db).Hydrate(ctx, result)
}

func (r *ResearchRepository) Exists(ctx context.Context, id string) (bool, error) {
	var count int
	err := selectRow(ctx, r.db.NewSelect().ColumnExpr("COUNT(*)").TableExpr("researches").Where("id=?", id)).Scan(&count)
	return count > 0, err
}

// ClaimOrphanedResearches hands every ownerless research to the first user to
// register. It moves the team too: a research left in the local team would
// still be readable by an unauthenticated caller, which is the opposite of
// what turning auth on is for.
func (r *ResearchRepository) ClaimOrphanedResearches(ctx context.Context, userID, teamID string) (int64, error) {
	res, err := r.db.NewUpdate().
		Table("researches").
		Set("user_id=?", userID).
		Set("team_id=?", teamID).
		Where("user_id IS NULL").
		Exec(ctx)
	if err != nil {
		return 0, fmt.Errorf("claim orphaned researches: %w", err)
	}
	return res.RowsAffected()
}

// SetTeam moves a research to another team. This is the only way its access
// list changes, since access is the team's membership.
func (r *ResearchRepository) SetTeam(ctx context.Context, researchID, teamID string) error {
	now := time.Now().UTC().Format(time.DateTime)
	_, err := r.db.NewUpdate().
		Table("researches").
		Set("team_id=?", teamID).
		Set("updated_at=?", now).
		Where("id=?", researchID).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("set research team: %w", err)
	}
	return nil
}

// CountByTeam reports how many researches a team holds, which is what stops a
// team being deleted out from under its content.
func (r *ResearchRepository) CountByTeam(ctx context.Context, teamID string) (int, error) {
	var n int
	err := selectRow(ctx, r.db.NewSelect().
		ColumnExpr("COUNT(*)").
		TableExpr("researches").
		Where("team_id=?", teamID)).
		Scan(&n)
	if err != nil {
		return 0, fmt.Errorf("count researches by team: %w", err)
	}
	return n, nil
}

// scanner is whatever a single row can be read from — sql.Row from a
// QueryRow, sql.Rows from a Query. One scanner for both keeps the field order
// in step with researchColumns.
type scanner interface {
	Scan(dest ...any) error
}

func scanResearchInto(s scanner) (*domain.Research, error) {
	var res domain.Research
	var userID, teamID sql.NullString
	var tags sql.NullString
	var createdAt, updatedAt string
	err := s.Scan(
		&res.ID, &res.Code, &userID, &teamID, &res.Name, &res.Description, &res.Goal,
		&res.Status, &tags,
		&createdAt, &updatedAt,
		&res.TemplateSlug, &res.TemplateVersion,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("scan research: %w", err)
	}
	if userID.Valid {
		res.UserID = userID.String
	}
	if teamID.Valid {
		res.TeamID = teamID.String
	}
	res.Memory = domain.Memory{}
	res.Tags = unmarshalStringSlice(tags)
	res.CreatedAt, _ = time.Parse(time.DateTime, createdAt)
	res.UpdatedAt, _ = time.Parse(time.DateTime, updatedAt)
	return &res, nil
}

func (r *ResearchRepository) scanResearch(row scanner) (*domain.Research, error) {
	return scanResearchInto(row)
}

func (r *ResearchRepository) scanResearchRow(rows *sql.Rows) (*domain.Research, error) {
	return scanResearchInto(rows)
}
