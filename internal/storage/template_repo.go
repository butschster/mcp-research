package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/butschster/mcp-research/internal/domain"
)

type TemplateRepository struct {
	db *sql.DB
}

func NewTemplateRepository(db *sql.DB) *TemplateRepository {
	return &TemplateRepository{db: db}
}

// The body is never in this list. A template body is a whole methodology, and
// four of them arriving in a kickoff that needs one is the cost this feature
// exists to avoid. `body_words` gives a reader the size without the text.
const templateColumns = `t.id, COALESCE(t.team_id,''), COALESCE(t.user_id,''), t.slug, t.name,
	t.description, t.when_to_use, t.when_not_to_use, t.skills, COALESCE(t.forked_from,''),
	t.version, t.created_at, t.updated_at,
	(LENGTH(t.body) - LENGTH(REPLACE(REPLACE(t.body, char(10), ' '), ' ', '')) + 1)`

func (r *TemplateRepository) Create(ctx context.Context, tp *domain.Template) error {
	now := time.Now().UTC().Format(time.DateTime)
	skills, err := json.Marshal(tp.Skills)
	if err != nil {
		return fmt.Errorf("marshal template skills: %w", err)
	}
	source := "user"
	if tp.Tier == domain.TemplateGlobal {
		source = "builtin"
	}
	_, err = r.db.ExecContext(ctx,
		`INSERT INTO research_templates (id, team_id, user_id, slug, name, description,
		                                 when_to_use, when_not_to_use, body, skills,
		                                 source, forked_from, version, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		tp.ID, nullable(tp.TeamID), nullable(tp.UserID), tp.Slug, tp.Name, tp.Description,
		tp.WhenToUse, tp.WhenNotToUse, tp.Body, string(skills),
		source, nullable(tp.ForkedFrom), tp.Version, now, now)
	if err != nil {
		return fmt.Errorf("insert template: %w", err)
	}
	tp.CreatedAt, _ = time.Parse(time.DateTime, now)
	tp.UpdatedAt = tp.CreatedAt
	tp.BodyWords = countWords(tp.Body)
	return nil
}

func (r *TemplateRepository) Update(ctx context.Context, tp *domain.Template) error {
	now := time.Now().UTC().Format(time.DateTime)
	skills, err := json.Marshal(tp.Skills)
	if err != nil {
		return fmt.Errorf("marshal template skills: %w", err)
	}
	res, err := r.db.ExecContext(ctx,
		`UPDATE research_templates
		    SET name=?, description=?, when_to_use=?, when_not_to_use=?, body=?, skills=?,
		        version=version+1, updated_at=?
		  WHERE id=?`,
		tp.Name, tp.Description, tp.WhenToUse, tp.WhenNotToUse, tp.Body, string(skills), now, tp.ID)
	if err != nil {
		return fmt.Errorf("update template: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return sql.ErrNoRows
	}
	tp.UpdatedAt, _ = time.Parse(time.DateTime, now)
	tp.Version++
	tp.BodyWords = countWords(tp.Body)
	return nil
}

func (r *TemplateRepository) Delete(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM research_templates WHERE id=?`, id)
	if err != nil {
		return fmt.Errorf("delete template: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *TemplateRepository) FindByID(ctx context.Context, id string) (*domain.Template, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT `+templateColumns+` FROM research_templates t WHERE t.id=?`, id)
	return scanTemplate(row)
}

// FindGlobalBySlug is the boot-time lookup, and it deliberately cannot see a
// team's fork of the same slug: an upgrade refreshes what we ship and must
// never overwrite what somebody edited.
func (r *TemplateRepository) FindGlobalBySlug(ctx context.Context, slug string) (*domain.Template, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT `+templateColumns+` FROM research_templates t
		  WHERE t.slug=? AND t.team_id IS NULL`, slug)
	return scanTemplate(row)
}

// FindVisibleBySlug resolves a slug for one caller: their teams' templates
// first, then the global ones. A team's fork keeps its parent's slug, so the
// team tier has to win or forking would have no effect on the team that did it.
//
// The tiebreaker after the tier is not decoration. Somebody in two teams that
// have both forked the same global sees two rows with that slug, and without a
// stable order this returned whichever the planner felt like — so an agent could
// read one team's methodology and have the research resolve to the other's. It
// is still ambiguous by nature; that is why anything that must be exact uses the
// id instead.
//
// teamIDs comes from the caller's memberships rather than from a request, so a
// slug can never resolve into a team they are not in.
func (r *TemplateRepository) FindVisibleBySlug(ctx context.Context, scope TemplateScope, slug string) (*domain.Template, error) {
	args := []any{slug}
	teamClause, teamArgs := scope.clause()
	args = append(args, teamArgs...)
	row := r.db.QueryRowContext(ctx,
		`SELECT `+templateColumns+` FROM research_templates t
		  WHERE t.slug = ? AND (`+teamClause+` OR t.team_id IS NULL)
		  ORDER BY CASE WHEN t.team_id IS NULL THEN 1 ELSE 0 END, t.created_at, t.id
		  LIMIT 1`, args...)
	return scanTemplate(row)
}

// ListVisible is every template a caller may use: the global set plus the
// libraries of the teams they belong to. A team's fork shadows the global it
// copies, so the same slug is never offered twice.
func (r *TemplateRepository) ListVisible(ctx context.Context, scope TemplateScope) ([]*domain.Template, error) {
	teamClause, teamArgs := scope.clause()
	args := append([]any{}, teamArgs...)
	// The shadowing subquery repeats the team clause, so its arguments go in
	// twice. Note this hides a *global* behind a team row and nothing else: two
	// teams that both forked the same slug legitimately produce two rows, which
	// are two different methodologies that happen to share a name.
	shadowInner, shadowArgs := scope.clauseFor("f")
	shadow := `EXISTS (SELECT 1 FROM research_templates f WHERE f.slug = t.slug AND ` + shadowInner + `)`
	args = append(args, shadowArgs...)
	rows, err := r.db.QueryContext(ctx,
		`SELECT `+templateColumns+` FROM research_templates t
		  WHERE (`+teamClause+` OR (t.team_id IS NULL AND NOT (`+shadow+`)))
		  ORDER BY CASE WHEN t.team_id IS NULL THEN 1 ELSE 0 END, t.name`, args...)
	if err != nil {
		return nil, fmt.Errorf("query templates: %w", err)
	}
	defer rows.Close()

	var result []*domain.Template
	for rows.Next() {
		tp, err := scanTemplateRow(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, tp)
	}
	return result, rows.Err()
}

// ListByTeam is one team's own library, without the global set mixed in.
func (r *TemplateRepository) ListByTeam(ctx context.Context, teamID string) ([]*domain.Template, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT `+templateColumns+` FROM research_templates t WHERE t.team_id = ? ORDER BY t.name`, teamID)
	if err != nil {
		return nil, fmt.Errorf("query team templates: %w", err)
	}
	defer rows.Close()

	var result []*domain.Template
	for rows.Next() {
		tp, err := scanTemplateRow(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, tp)
	}
	return result, rows.Err()
}

// Body is read separately so that no list query can pick it up by accident.
func (r *TemplateRepository) Body(ctx context.Context, id string) (string, error) {
	var body string
	err := r.db.QueryRowContext(ctx, `SELECT body FROM research_templates WHERE id=?`, id).Scan(&body)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("read template body: %w", err)
	}
	return body, nil
}

// ResearchCount is how many researches a template has started, counted only
// among those the caller can read.
//
// The scoping is not politeness: a global template is visible to everybody, so
// an unscoped count would tell one team how busy another has been.
func (r *TemplateRepository) ResearchCount(ctx context.Context, scope TemplateScope, tp *domain.Template) (int, error) {
	where := "r.template_slug = ?"
	args := []any{tp.Slug}

	// A research records the slug it followed, and a fork keeps its parent's
	// slug — so matching on the slug alone counted every research started from
	// a fork twice, once under the fork and once under the global it copies.
	//
	// A team template counts only its own team's researches. A global counts
	// only teams that have no fork of that slug, because a team holding one
	// followed their text and not ours.
	if tp.TeamID != "" {
		where += " AND r.team_id = ?"
		args = append(args, tp.TeamID)
	} else {
		where += ` AND NOT EXISTS (SELECT 1 FROM research_templates f
		                            WHERE f.slug = ? AND f.team_id = r.team_id)`
		args = append(args, tp.Slug)
	}

	if !scope.All {
		if len(scope.TeamIDs) == 0 {
			return 0, nil
		}
		where += " AND r.team_id IN (" + placeholders(len(scope.TeamIDs)) + ")"
		for _, id := range scope.TeamIDs {
			args = append(args, id)
		}
	}
	var n int
	if err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM researches r WHERE `+where, args...).Scan(&n); err != nil {
		return 0, fmt.Errorf("count researches from template: %w", err)
	}
	return n, nil
}

// SlugTaken answers within one tier, matching the two partial unique indexes.
func (r *TemplateRepository) SlugTaken(ctx context.Context, teamID, slug string) (bool, error) {
	query := `SELECT 1 FROM research_templates WHERE team_id IS NULL AND slug=?`
	args := []any{slug}
	if teamID != "" {
		query = `SELECT 1 FROM research_templates WHERE team_id=? AND slug=?`
		args = []any{teamID, slug}
	}
	var one int
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&one)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("check template slug: %w", err)
	}
	return true, nil
}

// StampResearch records which methodology a research was started from. The
// version is stored because built-ins are refreshed on every boot: without it,
// an upgrade would silently change the text behind a research already running.
func (r *TemplateRepository) StampResearch(ctx context.Context, researchID, slug string, version int) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE researches SET template_slug=?, template_version=? WHERE id=?`,
		slug, version, researchID)
	if err != nil {
		return fmt.Errorf("stamp research template: %w", err)
	}
	return nil
}

func scanTemplate(s scanner) (*domain.Template, error) {
	tp, err := scanTemplateRow(s)
	if err != nil {
		return nil, err
	}
	if tp.ID == "" {
		return nil, nil
	}
	return tp, nil
}

func scanTemplateRow(s scanner) (*domain.Template, error) {
	var tp domain.Template
	var skills, createdAt, updatedAt string
	err := s.Scan(&tp.ID, &tp.TeamID, &tp.UserID, &tp.Slug, &tp.Name, &tp.Description,
		&tp.WhenToUse, &tp.WhenNotToUse, &skills, &tp.ForkedFrom,
		&tp.Version, &createdAt, &updatedAt, &tp.BodyWords)
	if err == sql.ErrNoRows {
		return &domain.Template{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("scan template: %w", err)
	}
	tp.Tier = domain.TemplateTeam
	if tp.TeamID == "" {
		tp.Tier = domain.TemplateGlobal
	}
	// A row whose skill list is unreadable falls back to none. Failing towards
	// fewer attachments is the safe direction.
	if err := json.Unmarshal([]byte(skills), &tp.Skills); err != nil || tp.Skills == nil {
		tp.Skills = []string{}
	}
	tp.CreatedAt, _ = time.Parse(time.DateTime, createdAt)
	tp.UpdatedAt, _ = time.Parse(time.DateTime, updatedAt)
	return &tp, nil
}

// TemplateScope is which team libraries a caller can see.
//
// It is a type rather than a slice because "everything" and "nothing" are both
// real answers and a bare empty slice cannot tell them apart. Local mode
// (`auth_enabled: false`) has no memberships at all, so enumerating the caller's
// teams there returns nothing — and a template written on that instance would
// have become invisible to the person who wrote it.
type TemplateScope struct {
	TeamIDs []string
	// All is local mode: nobody to scope to, so everything is visible. It is
	// the same rule the rest of the product follows when there is no caller.
	All bool
}

func (s TemplateScope) clause() (string, []any) { return s.clauseFor("t") }

// Sees answers for one row, so a lookup by id obeys the same scoping as the
// list rather than becoming a way around it. A global template — no team —
// is everybody's.
func (s TemplateScope) Sees(teamID string) bool {
	if teamID == "" {
		return true
	}
	if s.All {
		return true
	}
	for _, id := range s.TeamIDs {
		if id == teamID {
			return true
		}
	}
	return false
}

func (s TemplateScope) clauseFor(alias string) (string, []any) {
	if s.All {
		return alias + ".team_id IS NOT NULL", nil
	}
	if len(s.TeamIDs) == 0 {
		return "0", nil
	}
	args := make([]any, 0, len(s.TeamIDs))
	for _, id := range s.TeamIDs {
		args = append(args, id)
	}
	return alias + ".team_id IN (" + placeholders(len(s.TeamIDs)) + ")", args
}

func placeholders(n int) string {
	return strings.TrimSuffix(strings.Repeat("?,", n), ",")
}

// countWords matches what the SQL expression above counts — whitespace runs —
// rather than being more accurate than it. Two definitions of the same field
// that disagree by six percent is a worse answer than one rough definition
// everything shares: the create response and the next read would otherwise
// report different numbers for the same unchanged body.
func countWords(s string) int {
	return len(strings.Fields(s))
}
