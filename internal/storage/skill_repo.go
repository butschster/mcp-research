package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/butschster/mcp-research/internal/domain"
)

type SkillRepository struct {
	db *sql.DB
}

func NewSkillRepository(db *sql.DB) *SkillRepository {
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
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO skills (id, team_id, research_id, user_id, slug, name, tier, description,
		                     body, ambient, forked_from, needs_trigger, version, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		sk.ID, nullable(sk.TeamID), nullable(sk.ResearchID), nullable(sk.UserID),
		sk.Slug, sk.Name, string(sk.Tier), sk.Description, sk.Body,
		boolToInt(sk.Ambient), nullable(sk.ForkedFrom), boolToInt(sk.NeedsTrigger),
		sk.Version, now, now)
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
	res, err := r.db.ExecContext(ctx,
		`UPDATE skills SET name=?, description=?, body=?, needs_trigger=?,
		        version=version+1, updated_at=?
		  WHERE id=?`,
		sk.Name, sk.Description, sk.Body, boolToInt(sk.NeedsTrigger), now, sk.ID)
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
	_, err := r.db.ExecContext(ctx, `UPDATE skills SET ambient=? WHERE id=?`, boolToInt(ambient), id)
	if err != nil {
		return fmt.Errorf("set skill ambient: %w", err)
	}
	return nil
}

func (r *SkillRepository) Delete(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM skills WHERE id=?`, id)
	if err != nil {
		return fmt.Errorf("delete skill: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *SkillRepository) FindByID(ctx context.Context, id string) (*domain.Skill, error) {
	row := r.db.QueryRowContext(ctx, `SELECT `+skillColumns+` FROM skills s WHERE s.id=?`, id)
	return scanSkill(row)
}

// Body is read separately from the row so that no list query can ever pick it
// up by accident. The index staying small is the whole point of this feature,
// and a `SELECT *` added later would silently undo it.
func (r *SkillRepository) Body(ctx context.Context, id string) (string, error) {
	var body string
	err := r.db.QueryRowContext(ctx, `SELECT body FROM skills WHERE id=?`, id).Scan(&body)
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
	row := r.db.QueryRowContext(ctx,
		`SELECT `+skillColumns+` FROM skills s
		  WHERE s.slug = ? AND `+skillScope+`
		  ORDER BY `+tierRank+` LIMIT 1`,
		slug, researchID, researchID)
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
	row := r.db.QueryRowContext(ctx,
		`SELECT `+skillColumns+`
		   FROM research_skills rs JOIN skills s ON s.id = rs.skill_id
		  WHERE rs.research_id = ? AND s.slug = ? AND `+skillScope+`
		  ORDER BY `+tierRank+` LIMIT 1`,
		researchID, slug, researchID, researchID)
	return scanSkill(row)
}

// FindBuiltinBySlug is the boot-time lookup, and it deliberately cannot see a
// team's fork of the same slug: an upgrade refreshes what we ship and must
// never touch what somebody edited.
func (r *SkillRepository) FindBuiltinBySlug(ctx context.Context, slug string) (*domain.Skill, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT `+skillColumns+` FROM skills s
		  WHERE s.slug=? AND s.team_id IS NULL AND s.research_id IS NULL`, slug)
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
	rows, err := r.db.QueryContext(ctx,
		`SELECT `+skillColumns+`, rs.via_template, rs.attached_at
		   FROM research_skills rs JOIN skills s ON s.id = rs.skill_id
		  WHERE rs.research_id = ? AND `+skillScope+`
		  ORDER BY `+tierRank+`, rs.attached_at, s.name`,
		researchID, researchID, researchID)
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

// ListLibrary is what a research may attach: the built-ins and its team's
// library, each carrying whether it is already on. Research-private skills are
// excluded — one research's private skill is not another's to borrow, and the
// exclusion is here rather than in the service because a library query that
// forgets it is a cross-research leak.
func (r *SkillRepository) ListLibrary(ctx context.Context, researchID, q string) ([]*domain.Skill, error) {
	like := "%" + q + "%"
	rows, err := r.db.QueryContext(ctx,
		`SELECT `+skillColumns+`,
		        EXISTS(SELECT 1 FROM research_skills rs
		                WHERE rs.skill_id = s.id AND rs.research_id = ?) AS attached
		   FROM skills s
		  WHERE s.research_id IS NULL
		    AND (s.team_id IS NULL
		         OR s.team_id = (SELECT team_id FROM researches WHERE id = ?))
		    -- A built-in whose slug this team has forked is shadowed: attaching
		    -- resolves by slug and lands on the fork, so listing it produces a
		    -- row with a live Attach button that always answers already_attached.
		    AND NOT (s.team_id IS NULL AND EXISTS (
		        SELECT 1 FROM skills f
		         WHERE f.slug = s.slug
		           AND f.team_id = (SELECT team_id FROM researches WHERE id = ?)))
		    AND (? = '' OR s.name LIKE ? OR s.description LIKE ?)
		  ORDER BY `+tierRank+`, s.name`,
		researchID, researchID, researchID, q, like, like)
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
	_, err := r.db.ExecContext(ctx,
		`INSERT OR IGNORE INTO research_skills (research_id, skill_id, via_template, attached_at)
		 VALUES (?, ?, ?, ?)`,
		researchID, skillID, boolToInt(viaTemplate), now)
	if err != nil {
		return fmt.Errorf("attach skill: %w", err)
	}
	return nil
}

func (r *SkillRepository) Detach(ctx context.Context, researchID, skillID string) error {
	res, err := r.db.ExecContext(ctx,
		`DELETE FROM research_skills WHERE research_id=? AND skill_id=?`, researchID, skillID)
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
	err = tx.QueryRowContext(ctx,
		`SELECT via_template, attached_at FROM research_skills WHERE research_id=? AND skill_id=?`,
		researchID, oldSkillID).Scan(&viaTemplate, &attachedAt)
	if err == sql.ErrNoRows {
		return sql.ErrNoRows
	}
	if err != nil {
		return fmt.Errorf("read attachment: %w", err)
	}
	if _, err := tx.ExecContext(ctx,
		`DELETE FROM research_skills WHERE research_id=? AND skill_id=?`,
		researchID, oldSkillID); err != nil {
		return fmt.Errorf("drop old attachment: %w", err)
	}
	// The original attachment time is carried over: the research has followed
	// this methodology since then, and a fork is not a fresh choice.
	if _, err := tx.ExecContext(ctx,
		`INSERT OR REPLACE INTO research_skills (research_id, skill_id, via_template, attached_at)
		 VALUES (?, ?, ?, ?)`,
		researchID, newSkillID, viaTemplate, attachedAt); err != nil {
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
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM research_skills rs JOIN skills s ON s.id = rs.skill_id
		  WHERE rs.research_id = ? AND s.ambient = 0 AND `+skillScope,
		researchID, researchID, researchID).Scan(&n)
	if err != nil {
		return 0, fmt.Errorf("count attached skills: %w", err)
	}
	return n, nil
}

// UsageCount is how many researches follow a skill, so deleting one can say
// what it would take away instead of silently taking it.
func (r *SkillRepository) UsageCount(ctx context.Context, skillID string) (int, error) {
	var n int
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM research_skills WHERE skill_id = ?`, skillID).Scan(&n)
	if err != nil {
		return 0, fmt.Errorf("count skill usage: %w", err)
	}
	return n, nil
}

// SlugTaken answers within one tier scope, matching the three partial unique
// indexes. Passing the wrong scope here would report a clash that the database
// would not have raised, or miss one it would.
func (r *SkillRepository) SlugTaken(ctx context.Context, tier domain.SkillTier, teamID, researchID, slug string) (bool, error) {
	var query string
	var args []any
	switch tier {
	case domain.SkillPrivate:
		query = `SELECT 1 FROM skills WHERE research_id=? AND slug=?`
		args = []any{researchID, slug}
	case domain.SkillTeam:
		query = `SELECT 1 FROM skills WHERE team_id=? AND slug=?`
		args = []any{teamID, slug}
	default:
		query = `SELECT 1 FROM skills WHERE team_id IS NULL AND research_id IS NULL AND slug=?`
		args = []any{slug}
	}
	var one int
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&one)
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
	rows, err := r.db.QueryContext(ctx,
		`SELECT research_id FROM research_skills WHERE skill_id = ?`, skillID)
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
	rows, err := r.db.QueryContext(ctx,
		`SELECT `+skillColumns+` FROM skills s
		  WHERE s.ambient = 1 AND s.team_id IS NULL AND s.research_id IS NULL
		  ORDER BY s.name`)
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
	rows, err := r.db.QueryContext(ctx,
		`SELECT `+skillColumns+` FROM skills s WHERE s.team_id = ? ORDER BY s.name`, teamID)
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
