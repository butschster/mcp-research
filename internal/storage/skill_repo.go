package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/butschster/mcp-research/internal/domain"
	"github.com/uptrace/bun"
)

type SkillRepository struct {
	db *bun.DB
}

func NewSkillRepository(db *bun.DB) *SkillRepository {
	return &SkillRepository{db: db}
}

// A slug is unique only within its tier scope — a built-in `source-triage` and
// a team's own `source-triage` are different rows on purpose, because forking a
// built-in keeps its slug. So every slug lookup in this file is scoped to one
// research: its private skills, then its team's, then the built-ins, in that
// order. That is the same order the index is sorted in, which is why a slug
// resolves to exactly the skill the agent was looking at.
//
// Management by slug alone would be ambiguous, so the write routes address a
// skill by id instead.
//
// body_tokens is computed here rather than selected: a list has to show what a
// skill costs without shipping the text it costs it in, and a stored column
// would be one more thing to keep in step with the body.
const skillColumns = `s.id, COALESCE(s.team_id,''), COALESCE(s.research_id,''), COALESCE(s.user_id,''),
	s.slug, s.name, s.tier, s.description, s.ambient, COALESCE(s.forked_from,''),
	s.needs_trigger, s.version, s.created_at, s.updated_at, (LENGTH(s.body)+3)/4`

// skillScope is the WHERE fragment naming every skill a research may see: its
// own private ones, the library of the team that owns it, and every built-in.
// It takes the research id twice and nothing else — the team is resolved by
// joining rather than by the caller passing one it might have got wrong.
const skillScope = `(
	s.research_id = ?
	OR (s.team_id IS NOT NULL AND s.team_id = (SELECT team_id FROM researches WHERE id = ?))
	OR (s.team_id IS NULL AND s.research_id IS NULL)
)`

// tierRank orders by precedence rather than alphabetically. Written as SQL so
// the database does the sort and every caller gets the same order.
const tierRank = `CASE s.tier WHEN 'private' THEN 0 WHEN 'team' THEN 1 ELSE 2 END`

func (r *SkillRepository) Create(ctx context.Context, sk *domain.Skill) error {
	now := time.Now().UTC().Format(time.DateTime)
	_, err := r.db.NewInsert().Table("skills").Model(&map[string]any{
		"id":            sk.ID,
		"team_id":       nullable(sk.TeamID),
		"research_id":   nullable(sk.ResearchID),
		"user_id":       nullable(sk.UserID),
		"slug":          sk.Slug,
		"name":          sk.Name,
		"tier":          string(sk.Tier),
		"description":   sk.Description,
		"body":          sk.Body,
		"ambient":       boolToInt(sk.Ambient),
		"forked_from":   nullable(sk.ForkedFrom),
		"needs_trigger": boolToInt(sk.NeedsTrigger),
		"version":       sk.Version,
		"created_at":    now,
		"updated_at":    now,
	}).Exec(ctx)
	if err != nil {
		return fmt.Errorf("insert skill: %w", err)
	}
	sk.CreatedAt, _ = time.Parse(time.DateTime, now)
	sk.UpdatedAt = sk.CreatedAt
	sk.BodyTokens = domain.EstimateTokens(sk.Body)
	return nil
}

func (r *SkillRepository) Update(ctx context.Context, sk *domain.Skill) error {
	now := time.Now().UTC().Format(time.DateTime)
	res, err := r.db.NewUpdate().
		Table("skills").
		Set("name=?", sk.Name).
		Set("description=?", sk.Description).
		Set("body=?", sk.Body).
		Set("needs_trigger=?", boolToInt(sk.NeedsTrigger)).
		Set("version=version+1").
		Set("updated_at=?", now).
		Where("id=?", sk.ID).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("update skill: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return sql.ErrNoRows
	}
	sk.UpdatedAt, _ = time.Parse(time.DateTime, now)
	sk.Version++
	sk.BodyTokens = domain.EstimateTokens(sk.Body)
	return nil
}

// SetAmbient is separate from Update because it is structural rather than
// editorial: it decides whether a skill counts against a research's budget, and
// nothing a user edits should be able to change that.
func (r *SkillRepository) SetAmbient(ctx context.Context, id string, ambient bool) error {
	_, err := r.db.NewUpdate().Table("skills").Set("ambient=?", boolToInt(ambient)).Where("id=?", id).Exec(ctx)
	if err != nil {
		return fmt.Errorf("set skill ambient: %w", err)
	}
	return nil
}

func (r *SkillRepository) Delete(ctx context.Context, id string) error {
	res, err := r.db.NewDelete().Table("skills").Where("id=?", id).Exec(ctx)
	if err != nil {
		return fmt.Errorf("delete skill: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *SkillRepository) FindByID(ctx context.Context, id string) (*domain.Skill, error) {
	row := selectRow(ctx, r.db.NewSelect().
		ColumnExpr(portableProjection(r.db, skillColumns)).
		TableExpr("skills s").
		Where("s.id=?", id))
	return scanSkill(row)
}

// Body is read separately from the row so that no list query can ever pick it
// up by accident. The index staying small is the whole point of this feature,
// and a `SELECT *` added later would silently undo it.
func (r *SkillRepository) Body(ctx context.Context, id string) (string, error) {
	var body string
	err := selectRow(ctx, r.db.NewSelect().ColumnExpr("body").TableExpr("skills").Where("id=?", id)).Scan(&body)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("read skill body: %w", err)
	}
	return body, nil
}

// FindInResearchScope resolves a slug the way the agent means it: the skill of
// that name this research can actually see, highest tier first.
func (r *SkillRepository) FindInResearchScope(ctx context.Context, researchID, slug string) (*domain.Skill, error) {
	row := selectRow(ctx, r.db.NewSelect().
		ColumnExpr(portableProjection(r.db, skillColumns)).
		TableExpr("skills s").
		Where("s.slug = ?", slug).
		Where(skillScope, researchID, researchID).
		OrderExpr(tierRank).
		Limit(1))
	return scanSkill(row)
}

// FindAttachedBySlug resolves a slug among the skills this research actually
// follows *and may still see* — scope is applied here for the same reason as in
// ListAttached: a transfer changes the owning team under a live attachment. This is the lookup that must be used for reading and for anything
// that acts on an attachment.
//
// Scope alone is not enough. A team skill sharing a slug with a built-in — which
// is exactly what forking produces — outranks it for every research in the team,
// including ones that attached the built-in and never asked for the fork. Those
// researches would list the built-in in their index and be handed the team's
// body by skill_load: the agent reading something other than what it was told it
// had.
func (r *SkillRepository) FindAttachedBySlug(ctx context.Context, researchID, slug string) (*domain.Skill, error) {
	row := selectRow(ctx, r.db.NewSelect().
		ColumnExpr(portableProjection(r.db, skillColumns)).
		TableExpr("research_skills rs JOIN skills s ON s.id = rs.skill_id").
		Where("rs.research_id = ?", researchID).
		Where("s.slug = ?", slug).
		Where(skillScope, researchID, researchID).
		OrderExpr(tierRank).
		Limit(1))
	return scanSkill(row)
}

// FindBuiltinBySlug is the boot-time lookup, and it deliberately cannot see a
// team's fork of the same slug: an upgrade refreshes what we ship and must
// never touch what somebody edited.
func (r *SkillRepository) FindBuiltinBySlug(ctx context.Context, slug string) (*domain.Skill, error) {
	row := selectRow(ctx, r.db.NewSelect().
		ColumnExpr(portableProjection(r.db, skillColumns)).
		TableExpr("skills s").
		Where("s.slug=? AND s.team_id IS NULL AND s.research_id IS NULL", slug))
	return scanSkill(row)
}

// ListAttached returns what a research follows, in precedence order.
//
// It intersects with skillScope, and that is not belt and braces. An attachment
// is checked when it is made, and `POST /api/researches/{id}/transfer` then
// moves the research to another team without touching research_skills — so the
// rows survive while the team that owns the skill does not. Without the
// intersection this query, and the body read beside it, handed the destination
// team the source team's private methodology.
//
// The sort is tier, then when it was attached, then name. The last of those is
// not decoration: attached_at has one-second granularity, so six skills
// attached by one template land in the same instant and the order within them
// would otherwise be whatever the query planner felt like — a list that
// rearranges itself between page loads.
func (r *SkillRepository) ListAttached(ctx context.Context, researchID string) ([]*domain.Skill, error) {
	rows, err := r.db.NewSelect().
		ColumnExpr(portableProjection(r.db, skillColumns)).
		ColumnExpr("rs.via_template, rs.attached_at").
		TableExpr("research_skills rs JOIN skills s ON s.id = rs.skill_id").
		Where("rs.research_id = ?", researchID).
		Where(skillScope, researchID, researchID).
		OrderExpr(tierRank + ", rs.attached_at, s.name").
		Rows(ctx)
	if err != nil {
		return nil, fmt.Errorf("query attached skills: %w", err)
	}
	defer rows.Close()

	var result []*domain.Skill
	for rows.Next() {
		var sk domain.Skill
		var viaTemplate int
		var attachedAt string
		if err := scanSkillInto(rows, &sk, &viaTemplate, &attachedAt); err != nil {
			return nil, err
		}
		sk.ViaTemplate = viaTemplate != 0
		if t, err := time.Parse(time.DateTime, attachedAt); err == nil {
			sk.AttachedAt = &t
		}
		result = append(result, &sk)
	}
	return result, rows.Err()
}

// ListLibrary is what a research may attach: the non-ambient built-ins and its
// team's library, each carrying whether it is already on. Research-private skills are
// excluded — one research's private skill is not another's to borrow, and the
// exclusion is here rather than in the service because a library query that
// forgets it is a cross-research leak.
func (r *SkillRepository) ListLibrary(ctx context.Context, researchID, q string) ([]*domain.Skill, error) {
	like := "%" + q + "%"
	rows, err := r.db.NewSelect().
		ColumnExpr(portableProjection(r.db, skillColumns)).
		ColumnExpr("CASE WHEN EXISTS(SELECT 1 FROM research_skills rs WHERE rs.skill_id = s.id AND rs.research_id = ?) THEN 1 ELSE 0 END AS attached", researchID).
		TableExpr("skills s").
		Where("s.research_id IS NULL").
		// Ambient skills are already attached; team forks shadow built-ins.
		Where("s.ambient = 0").
		Where("s.team_id IS NULL OR s.team_id = (SELECT team_id FROM researches WHERE id = ?)", researchID).
		Where(`NOT (s.team_id IS NULL AND EXISTS (
			SELECT 1 FROM skills f WHERE f.slug = s.slug
			AND f.team_id = (SELECT team_id FROM researches WHERE id = ?)))`, researchID).
		Where("? = '' OR LOWER(s.name) LIKE LOWER(?) OR LOWER(s.description) LIKE LOWER(?)", q, like, like).
		OrderExpr(tierRank + ", s.name").
		Rows(ctx)
	if err != nil {
		return nil, fmt.Errorf("query skill library: %w", err)
	}
	defer rows.Close()

	var result []*domain.Skill
	for rows.Next() {
		var sk domain.Skill
		var attached int
		if err := scanSkillInto(rows, &sk, &attached); err != nil {
			return nil, err
		}
		sk.Attached = attached != 0
		result = append(result, &sk)
	}
	return result, rows.Err()
}

func (r *SkillRepository) Attach(ctx context.Context, researchID, skillID string, viaTemplate bool) error {
	now := time.Now().UTC().Format(time.DateTime)
	_, err := onConflict(r.db.NewInsert().Table("research_skills").Model(&map[string]any{
		"research_id":  researchID,
		"skill_id":     skillID,
		"via_template": boolToInt(viaTemplate),
		"attached_at":  now,
	}), r.db, []string{"research_id", "skill_id"}).Exec(ctx)
	if err != nil {
		return fmt.Errorf("attach skill: %w", err)
	}
	return nil
}

func (r *SkillRepository) Detach(ctx context.Context, researchID, skillID string) error {
	res, err := r.db.NewDelete().
		Table("research_skills").
		Where("research_id=? AND skill_id=?", researchID, skillID).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("detach skill: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// Replace swaps one attachment for another in a single statement pair, which is
// what fork, copy and promote all need. Doing it as detach-then-attach would
// pass through a state where the research has one fewer skill, and a concurrent
// attach could take the freed slot and leave the research with a hole.
func (r *SkillRepository) Replace(ctx context.Context, researchID, oldSkillID, newSkillID string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin replace: %w", err)
	}
	defer tx.Rollback()

	var viaTemplate int
	var attachedAt string
	err = selectRow(ctx, tx.NewSelect().
		ColumnExpr("via_template, attached_at").
		TableExpr("research_skills").
		Where("research_id=? AND skill_id=?", researchID, oldSkillID)).
		Scan(&viaTemplate, &attachedAt)
	if err == sql.ErrNoRows {
		return sql.ErrNoRows
	}
	if err != nil {
		return fmt.Errorf("read attachment: %w", err)
	}
	if _, err := tx.NewDelete().
		Table("research_skills").
		Where("research_id=? AND skill_id=?", researchID, oldSkillID).
		Exec(ctx); err != nil {
		return fmt.Errorf("drop old attachment: %w", err)
	}
	// The original attachment time is carried over: the research has followed
	// this methodology since then, and a fork is not a fresh choice.
	if _, err := onConflict(tx.NewInsert().Table("research_skills").Model(&map[string]any{
		"research_id":  researchID,
		"skill_id":     newSkillID,
		"via_template": viaTemplate,
		"attached_at":  attachedAt,
	}), tx, []string{"research_id", "skill_id"}, "via_template", "attached_at").Exec(ctx); err != nil {
		return fmt.Errorf("write new attachment: %w", err)
	}
	return tx.Commit()
}

// CountChosen counts what the cap is actually a budget over. Ambient product
// skills are excluded here rather than at the call site, so a second caller
// cannot forget and quietly reintroduce the day-one full cap. Scope is applied
// too, so a skill the research can no longer see does not go on spending its
// budget.
func (r *SkillRepository) CountChosen(ctx context.Context, researchID string) (int, error) {
	var n int
	err := selectRow(ctx, r.db.NewSelect().
		ColumnExpr("COUNT(*)").
		TableExpr("research_skills rs JOIN skills s ON s.id = rs.skill_id").
		Where("rs.research_id = ?", researchID).
		Where("s.ambient = 0").
		Where(skillScope, researchID, researchID)).
		Scan(&n)
	if err != nil {
		return 0, fmt.Errorf("count attached skills: %w", err)
	}
	return n, nil
}

// UsageCount is how many researches follow a skill, so deleting one can say
// what it would take away instead of silently taking it.
func (r *SkillRepository) UsageCount(ctx context.Context, skillID string) (int, error) {
	var n int
	err := selectRow(ctx, r.db.NewSelect().
		ColumnExpr("COUNT(*)").
		TableExpr("research_skills").
		Where("skill_id = ?", skillID)).
		Scan(&n)
	if err != nil {
		return 0, fmt.Errorf("count skill usage: %w", err)
	}
	return n, nil
}

// SlugTaken answers within one tier scope, matching the three partial unique
// indexes. Passing the wrong scope here would report a clash that the database
// would not have raised, or miss one it would.
func (r *SkillRepository) SlugTaken(ctx context.Context, tier domain.SkillTier, teamID, researchID, slug string) (bool, error) {
	query := r.db.NewSelect().ColumnExpr("1").Table("skills").Where("slug=?", slug)
	switch tier {
	case domain.SkillPrivate:
		query.Where("research_id=?", researchID)
	case domain.SkillTeam:
		query.Where("team_id=?", teamID)
	default:
		query.Where("team_id IS NULL AND research_id IS NULL")
	}
	var one int
	err := selectRow(ctx, query).Scan(&one)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("check skill slug: %w", err)
	}
	return true, nil
}

func scanSkill(s scanner) (*domain.Skill, error) {
	var sk domain.Skill
	if err := scanSkillInto(s, &sk); err != nil {
		return nil, err
	}
	if sk.ID == "" {
		return nil, nil
	}
	return &sk, nil
}

// scanSkillInto takes the fixed skill columns and then whatever extra
// destinations a particular query appended, so the column list has one
// definition and the joins that widen it cannot fall out of step with it.
func scanSkillInto(s scanner, sk *domain.Skill, extra ...any) error {
	var tier, createdAt, updatedAt string
	var ambient, needsTrigger int
	dest := []any{
		&sk.ID, &sk.TeamID, &sk.ResearchID, &sk.UserID,
		&sk.Slug, &sk.Name, &tier, &sk.Description, &ambient, &sk.ForkedFrom,
		&needsTrigger, &sk.Version, &createdAt, &updatedAt, &sk.BodyTokens,
	}
	dest = append(dest, extra...)
	err := s.Scan(dest...)
	if err == sql.ErrNoRows {
		return nil
	}
	if err != nil {
		return fmt.Errorf("scan skill: %w", err)
	}
	sk.Tier = domain.SkillTier(tier)
	sk.Ambient = ambient != 0
	sk.NeedsTrigger = needsTrigger != 0
	sk.CreatedAt, _ = time.Parse(time.DateTime, createdAt)
	sk.UpdatedAt, _ = time.Parse(time.DateTime, updatedAt)
	return nil
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// ResearchesFollowing lists every research attached to a skill. It is read
// *before* a delete, because the attachment rows cascade away with the skill and
// afterwards there is nobody left to notify.
func (r *SkillRepository) ResearchesFollowing(ctx context.Context, skillID string) ([]string, error) {
	rows, err := r.db.NewSelect().
		ColumnExpr("research_id").
		TableExpr("research_skills").
		Where("skill_id = ?", skillID).
		Rows(ctx)
	if err != nil {
		return nil, fmt.Errorf("query skill followers: %w", err)
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan skill follower: %w", err)
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

// ListAmbient returns the product skills that are on by definition.
//
// "Always on" has to be true by construction. Requiring somebody to attach them
// made it a promise the product broke on every research nobody had curated: the
// index came back empty, so the agent never learned skills existed at all.
func (r *SkillRepository) ListAmbient(ctx context.Context) ([]*domain.Skill, error) {
	rows, err := r.db.NewSelect().
		ColumnExpr(portableProjection(r.db, skillColumns)).
		TableExpr("skills s").
		Where("s.ambient = 1 AND s.team_id IS NULL AND s.research_id IS NULL").
		OrderExpr("s.name").
		Rows(ctx)
	if err != nil {
		return nil, fmt.Errorf("query ambient skills: %w", err)
	}
	defer rows.Close()

	var result []*domain.Skill
	for rows.Next() {
		var sk domain.Skill
		if err := scanSkillInto(rows, &sk); err != nil {
			return nil, err
		}
		result = append(result, &sk)
	}
	return result, rows.Err()
}

// ListByTeam is a team's library on its own terms, for a team that has no
// research to browse it through yet.
func (r *SkillRepository) ListByTeam(ctx context.Context, teamID string) ([]*domain.Skill, error) {
	rows, err := r.db.NewSelect().
		ColumnExpr(portableProjection(r.db, skillColumns)).
		TableExpr("skills s").
		Where("s.team_id = ?", teamID).
		OrderExpr("s.name").
		Rows(ctx)
	if err != nil {
		return nil, fmt.Errorf("query team skills: %w", err)
	}
	defer rows.Close()

	var result []*domain.Skill
	for rows.Next() {
		var sk domain.Skill
		if err := scanSkillInto(rows, &sk); err != nil {
			return nil, err
		}
		result = append(result, &sk)
	}
	return result, rows.Err()
}

// BuiltinSkillExists answers whether a slug names a skill we ship. It is the
// boot-time check behind a template's skill list: a template naming a skill that
// does not exist is a broken methodology, and finding that out at startup is
// worth more than finding it out from somebody's kickoff.
func (r *SkillRepository) BuiltinSkillExists(ctx context.Context, slug string) (bool, error) {
	var one int
	err := selectRow(ctx, r.db.NewSelect().
		ColumnExpr("1").
		TableExpr("skills").
		Where("slug=? AND team_id IS NULL AND research_id IS NULL", slug)).
		Scan(&one)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("check builtin skill: %w", err)
	}
	return true, nil
}

// FindForTemplate resolves a skill slug the way a template means it: the owning
// team's copy if there is one, otherwise the built-in. There is no research
// here — a template is read before one exists — so the research tier is out of
// scope by construction.
func (r *SkillRepository) FindForTemplate(ctx context.Context, teamID, slug string) (*domain.Skill, error) {
	row := selectRow(ctx, r.db.NewSelect().
		ColumnExpr(portableProjection(r.db, skillColumns)).
		TableExpr("skills s").
		Where("s.slug = ? AND s.research_id IS NULL AND (s.team_id IS NULL OR s.team_id = ?)", slug, teamID).
		OrderExpr("CASE WHEN s.team_id IS NULL THEN 1 ELSE 0 END").
		Limit(1))
	return scanSkill(row)
}
