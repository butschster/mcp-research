package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/butschster/mcp-research/internal/domain"
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
const researchColumns = `id, code, user_id, team_id, name, description, goal, status, instruction, memory, tags, created_at, updated_at, template_slug, template_version`

type ResearchRepository struct {
	db *sql.DB
}

func NewResearchRepository(db *sql.DB) *ResearchRepository {
	return &ResearchRepository{db: db}
}

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

	_, err := r.db.ExecContext(ctx,
		`INSERT INTO researches (id, code, user_id, team_id, name, description, goal, status, instruction, memory, tags, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		research.ID, research.Code, userID, teamID, research.Name, research.Description, research.Goal,
		research.Status, research.Instruction,
		marshalJSON(research.Memory), marshalJSON(research.Tags),
		now, now,
	)
	if err != nil {
		return fmt.Errorf("insert research: %w", err)
	}
	research.CreatedAt, _ = time.Parse(time.DateTime, now)
	research.UpdatedAt = research.CreatedAt
	return nil
}

func (r *ResearchRepository) Update(ctx context.Context, research *domain.Research) error {
	now := time.Now().UTC().Format(time.DateTime)
	_, err := r.db.ExecContext(ctx,
		`UPDATE researches SET name=?, description=?, goal=?, status=?, instruction=?, memory=?, tags=?, code=?, updated_at=?
		 WHERE id=?`,
		research.Name, research.Description, research.Goal,
		research.Status, research.Instruction,
		marshalJSON(research.Memory), marshalJSON(research.Tags),
		research.Code, now, research.ID,
	)
	if err != nil {
		return fmt.Errorf("update research: %w", err)
	}
	research.UpdatedAt, _ = time.Parse(time.DateTime, now)
	return nil
}

func (r *ResearchRepository) FindByID(ctx context.Context, id string) (*domain.Research, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT `+researchColumns+` FROM researches WHERE id=?`, id)
	return r.scanResearch(row)
}

func (r *ResearchRepository) FindByCode(ctx context.Context, code string) (*domain.Research, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT `+researchColumns+` FROM researches WHERE code=?`, code)
	return r.scanResearch(row)
}

func (r *ResearchRepository) FindAll(ctx context.Context, filter ResearchFilter) ([]*domain.Research, error) {
	query := `SELECT ` + researchColumns + ` FROM researches`
	var args []any
	var conditions []string

	if filter.Status != nil {
		conditions = append(conditions, "status=?")
		args = append(args, *filter.Status)
	}
	if filter.UserID != nil {
		conditions = append(conditions, "user_id=?")
		args = append(args, *filter.UserID)
	}
	if filter.MemberOf != nil {
		// The local team is included for every authenticated caller: it holds
		// researches created with no user at all, which everybody could see
		// before teams and which would otherwise vanish from every list.
		conditions = append(conditions,
			"(team_id = 'team-local' OR team_id IN (SELECT team_id FROM team_members WHERE user_id=?))")
		args = append(args, *filter.MemberOf)
	}
	if filter.TeamID != nil {
		conditions = append(conditions, "team_id=?")
		args = append(args, *filter.TeamID)
	}

	if len(conditions) > 0 {
		query += " WHERE " + conditions[0]
		for _, c := range conditions[1:] {
			query += " AND " + c
		}
	}

	query += " ORDER BY created_at DESC"

	rows, err := r.db.QueryContext(ctx, query, args...)
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
	return result, rows.Err()
}

func (r *ResearchRepository) Exists(ctx context.Context, id string) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM researches WHERE id=?", id).Scan(&count)
	return count > 0, err
}

// ClaimOrphanedResearches hands every ownerless research to the first user to
// register. It moves the team too: a research left in the local team would
// still be readable by an unauthenticated caller, which is the opposite of
// what turning auth on is for.
func (r *ResearchRepository) ClaimOrphanedResearches(ctx context.Context, userID, teamID string) (int64, error) {
	res, err := r.db.ExecContext(ctx,
		`UPDATE researches SET user_id=?, team_id=? WHERE user_id IS NULL`, userID, teamID)
	if err != nil {
		return 0, fmt.Errorf("claim orphaned researches: %w", err)
	}
	return res.RowsAffected()
}

// SetTeam moves a research to another team. This is the only way its access
// list changes, since access is the team's membership.
func (r *ResearchRepository) SetTeam(ctx context.Context, researchID, teamID string) error {
	now := time.Now().UTC().Format(time.DateTime)
	_, err := r.db.ExecContext(ctx,
		`UPDATE researches SET team_id=?, updated_at=? WHERE id=?`, teamID, now, researchID)
	if err != nil {
		return fmt.Errorf("set research team: %w", err)
	}
	return nil
}

// CountByTeam reports how many researches a team holds, which is what stops a
// team being deleted out from under its content.
func (r *ResearchRepository) CountByTeam(ctx context.Context, teamID string) (int, error) {
	var n int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM researches WHERE team_id=?`, teamID).Scan(&n)
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
	var memory, tags sql.NullString
	var createdAt, updatedAt string
	err := s.Scan(
		&res.ID, &res.Code, &userID, &teamID, &res.Name, &res.Description, &res.Goal,
		&res.Status, &res.Instruction,
		&memory, &tags,
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
	res.Memory = unmarshalStringSlice(memory)
	res.Tags = unmarshalStringSlice(tags)
	res.CreatedAt, _ = time.Parse(time.DateTime, createdAt)
	res.UpdatedAt, _ = time.Parse(time.DateTime, updatedAt)
	return &res, nil
}

func (r *ResearchRepository) scanResearch(row *sql.Row) (*domain.Research, error) {
	return scanResearchInto(row)
}

func (r *ResearchRepository) scanResearchRow(rows *sql.Rows) (*domain.Research, error) {
	return scanResearchInto(rows)
}
