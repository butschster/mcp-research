package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/dovod-app/app/internal/domain"
	"github.com/uptrace/bun"
)

type TemplateRepository struct {
	db *bun.DB
}

func NewTemplateRepository(db *bun.DB) *TemplateRepository {
	return &TemplateRepository{db: db}
}

// The body is never in this list. A template body is a whole methodology, and
// the whole shipped set arriving in a kickoff that needs one of them is the cost
// this feature exists to avoid. `body_words` gives a reader the size without the text.
const templateColumns = `t.id, COALESCE(t.team_id,''), COALESCE(t.user_id,''), t.slug, t.name,
	t.description, t.when_to_use, t.when_not_to_use, t.skills, COALESCE(t.forked_from,''),
	t.source, t.version, t.created_at, t.updated_at,
	(LENGTH(t.body) - LENGTH(REPLACE(REPLACE(t.body, char(10), ' '), ' ', '')) + 1)`

func (r *TemplateRepository) Create(ctx context.Context, tp *domain.Template) error {
	now := time.Now().UTC().Format(time.DateTime)
	skills, err := json.Marshal(tp.Skills)
	if err != nil {
		return fmt.Errorf("marshal template skills: %w", err)
	}
	// The caller says which. It used to be derived from the tier — global meant
	// built-in — which was true only while the binary was the sole way a global
	// row could appear. It no longer is: the operator can add one through the
	// API, and deriving it would have marked that row as ours and let the next
	// boot's refresh delete or overwrite it.
	source := tp.Source
	if source == "" {
		source = domain.TemplateSourceUser
	}
	_, err = r.db.NewInsert().Table("research_templates").Model(&map[string]any{
		"id":              tp.ID,
		"team_id":         nullable(tp.TeamID),
		"user_id":         nullable(tp.UserID),
		"slug":            tp.Slug,
		"name":            tp.Name,
		"description":     tp.Description,
		"when_to_use":     tp.WhenToUse,
		"when_not_to_use": tp.WhenNotToUse,
		"body":            tp.Body,
		"skills":          string(skills),
		"source":          string(source),
		"forked_from":     nullable(tp.ForkedFrom),
		"version":         tp.Version,
		"created_at":      now,
		"updated_at":      now,
	}).Exec(ctx)
	if err != nil {
		return fmt.Errorf("insert template: %w", err)
	}
	tp.Source = source
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
	res, err := r.db.NewUpdate().
		Table("research_templates").
		Set("name=?", tp.Name).
		Set("description=?", tp.Description).
		Set("when_to_use=?", tp.WhenToUse).
		Set("when_not_to_use=?", tp.WhenNotToUse).
		Set("body=?", tp.Body).
		Set("skills=?", string(skills)).
		Set("version=version+1").
		Set("updated_at=?", now).
		Where("id=?", tp.ID).
		Exec(ctx)
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
	res, err := r.db.NewDelete().Table("research_templates").Where("id=?", id).Exec(ctx)
	if err != nil {
		return fmt.Errorf("delete template: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *TemplateRepository) FindByID(ctx context.Context, id string) (*domain.Template, error) {
	row := selectRow(ctx, r.db.NewSelect().
		ColumnExpr(portableProjection(r.db, templateColumns)).
		TableExpr("research_templates t").
		Where("t.id=?", id))
	return scanTemplate(row)
}

// FindGlobalBySlug resolves a slug in the global tier, whatever wrote it. It
// deliberately cannot see a team's fork of the same slug — Fork copies what we
// publish, not what a team already edited.
func (r *TemplateRepository) FindGlobalBySlug(ctx context.Context, slug string) (*domain.Template, error) {
	row := selectRow(ctx, r.db.NewSelect().
		ColumnExpr(portableProjection(r.db, templateColumns)).
		TableExpr("research_templates t").
		Where("t.slug=? AND t.team_id IS NULL", slug))
	return scanTemplate(row)
}

// FindBuiltinBySlug is the boot-time lookup, and the narrower one on purpose.
//
// The refresh may only ever find a row it wrote itself. A global template the
// operator added through the API sits in the same tier under the same unique
// index, and matching on the tier alone would have let a later release quietly
// overwrite their text with ours the moment a shipped file happened to take the
// same slug. It cannot now: the insert collides with the index instead, and the
// loader reports a name clash it cannot resolve rather than resolving it in our
// favour.
func (r *TemplateRepository) FindBuiltinBySlug(ctx context.Context, slug string) (*domain.Template, error) {
	row := selectRow(ctx, r.db.NewSelect().
		ColumnExpr(portableProjection(r.db, templateColumns)).
		TableExpr("research_templates t").
		Where("t.slug=? AND t.team_id IS NULL AND t.source='builtin'", slug))
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
	row := selectRow(ctx, r.db.NewSelect().
		ColumnExpr(portableProjection(r.db, templateColumns)).
		TableExpr("research_templates t").
		Where("t.slug = ? AND ("+teamClause+" OR t.team_id IS NULL)", args...).
		OrderExpr("CASE WHEN t.team_id IS NULL THEN 1 ELSE 0 END, t.created_at, t.id").
		Limit(1))
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
	rows, err := r.db.NewSelect().
		ColumnExpr(portableProjection(r.db, templateColumns)).
		TableExpr("research_templates t").
		Where("("+teamClause+" OR (t.team_id IS NULL AND NOT ("+shadow+")))", args...).
		OrderExpr("CASE WHEN t.team_id IS NULL THEN 1 ELSE 0 END, t.name").
		Rows(ctx)
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
	rows, err := r.db.NewSelect().
		ColumnExpr(portableProjection(r.db, templateColumns)).
		TableExpr("research_templates t").
		Where("t.team_id = ?", teamID).
		OrderExpr("t.name").
		Rows(ctx)
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
	err := selectRow(ctx, r.db.NewSelect().
		ColumnExpr("body").
		TableExpr("research_templates").
		Where("id=?", id)).
		Scan(&body)
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
		where += " AND r.team_id IN (?)"
		args = append(args, bun.In(scope.TeamIDs))
	}
	var n int
	if err := r.db.NewSelect().
		ColumnExpr("COUNT(*)").
		TableExpr("researches r").
		Where(where, args...).
		Scan(ctx, &n); err != nil {
		return 0, fmt.Errorf("count researches from template: %w", err)
	}
	return n, nil
}

// SlugTaken answers within one tier, matching the two partial unique indexes.
func (r *TemplateRepository) SlugTaken(ctx context.Context, teamID, slug string) (bool, error) {
	query := r.db.NewSelect().ColumnExpr("1").Table("research_templates").Where("slug=?", slug)
	if teamID != "" {
		query.Where("team_id=?", teamID)
	} else {
		query.Where("team_id IS NULL")
	}
	var one int
	err := selectRow(ctx, query).Scan(&one)
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
	_, err := r.db.NewUpdate().
		Table("researches").
		Set("template_slug=?", slug).
		Set("template_version=?", version).
		Where("id=?", researchID).
		Exec(ctx)
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
	var skills, source, createdAt, updatedAt string
	err := s.Scan(&tp.ID, &tp.TeamID, &tp.UserID, &tp.Slug, &tp.Name, &tp.Description,
		&tp.WhenToUse, &tp.WhenNotToUse, &skills, &tp.ForkedFrom,
		&source, &tp.Version, &createdAt, &updatedAt, &tp.BodyWords)
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
	tp.Source = domain.TemplateSource(source)
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
		return "1=0", nil
	}
	return alias + ".team_id IN (?)", []any{bun.In(s.TeamIDs)}
}

// countWords matches what the SQL expression above counts — whitespace runs —
// rather than being more accurate than it. Two definitions of the same field
// that disagree by six percent is a worse answer than one rough definition
// everything shares: the create response and the next read would otherwise
// report different numbers for the same unchanged body.
func countWords(s string) int {
	return len(strings.Fields(s))
}
