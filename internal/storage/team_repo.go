package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/butschster/mcp-research/internal/domain"
	"github.com/uptrace/bun"
)

type TeamRepository struct {
	db *bun.DB
}

func NewTeamRepository(db *bun.DB) *TeamRepository {
	return &TeamRepository{db: db}
}

// CreateWithOwner writes the team and its first membership together. Split in
// two, a failure between them leaves a team nobody can administer — including
// the person who just created it.
func (r *TeamRepository) CreateWithOwner(ctx context.Context, team *domain.Team, ownerID string) error {
	now := time.Now().UTC().Format(time.DateTime)

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	personal := 0
	if team.Personal {
		personal = 1
	}
	var createdBy any
	if team.CreatedBy != "" {
		createdBy = team.CreatedBy
	}

	if _, err := tx.NewInsert().Table("teams").Model(&map[string]any{
		"id":         team.ID,
		"name":       team.Name,
		"personal":   personal,
		"created_by": createdBy,
		"created_at": now,
		"updated_at": now,
	}).Exec(ctx); err != nil {
		return fmt.Errorf("insert team: %w", err)
	}

	if ownerID != "" {
		if _, err := onConflict(tx.NewInsert().Table("team_members").Model(&map[string]any{
			"team_id":    team.ID,
			"user_id":    ownerID,
			"role":       domain.TeamOwner,
			"created_at": now,
		}), tx, []string{"team_id", "user_id"}).Exec(ctx); err != nil {
			return fmt.Errorf("insert owner membership: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}

	team.CreatedAt, _ = time.Parse(time.DateTime, now)
	team.UpdatedAt = team.CreatedAt
	team.Role = domain.TeamOwner
	return nil
}

func (r *TeamRepository) FindByID(ctx context.Context, id string) (*domain.Team, error) {
	row := selectRow(ctx, r.db.NewSelect().
		ColumnExpr("id, name, personal, created_by, created_at, updated_at, (SELECT COUNT(*) FROM team_members x WHERE x.team_id = teams.id), (SELECT COUNT(*) FROM researches x WHERE x.team_id = teams.id)").
		TableExpr("teams").
		Where("id=?", id))
	var t domain.Team
	var personal int
	var createdBy sql.NullString
	var createdAt, updatedAt string
	err := row.Scan(&t.ID, &t.Name, &personal, &createdBy, &createdAt, &updatedAt,
		&t.MemberCount, &t.ResearchCount)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("scan team: %w", err)
	}
	t.Personal = personal == 1
	if createdBy.Valid {
		t.CreatedBy = createdBy.String
	}
	t.CreatedAt, _ = time.Parse(time.DateTime, createdAt)
	t.UpdatedAt, _ = time.Parse(time.DateTime, updatedAt)
	return &t, nil
}

// FindPersonal returns the team created alongside the user. Every user has
// one; a nil result means the account predates teams and needs one made.
func (r *TeamRepository) FindPersonal(ctx context.Context, userID string) (*domain.Team, error) {
	row := selectRow(ctx, r.db.NewSelect().
		ColumnExpr("t.id, t.name, t.personal, t.created_by, t.created_at, t.updated_at").
		TableExpr("teams t JOIN team_members m ON m.team_id = t.id AND m.user_id = ?", userID).
		Where("t.personal = 1").
		OrderExpr("t.created_at").
		Limit(1))
	var t domain.Team
	var personal int
	var createdBy sql.NullString
	var createdAt, updatedAt string
	err := row.Scan(&t.ID, &t.Name, &personal, &createdBy, &createdAt, &updatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("scan personal team: %w", err)
	}
	t.Personal = personal == 1
	if createdBy.Valid {
		t.CreatedBy = createdBy.String
	}
	t.CreatedAt, _ = time.Parse(time.DateTime, createdAt)
	t.UpdatedAt, _ = time.Parse(time.DateTime, updatedAt)
	t.Role = domain.TeamOwner
	return &t, nil
}

func (r *TeamRepository) Rename(ctx context.Context, id, name string) error {
	now := time.Now().UTC().Format(time.DateTime)
	_, err := r.db.NewUpdate().Table("teams").Set("name=?", name).Set("updated_at=?", now).Where("id=?", id).Exec(ctx)
	if err != nil {
		return fmt.Errorf("rename team: %w", err)
	}
	return nil
}

func (r *TeamRepository) Delete(ctx context.Context, id string) error {
	if _, err := r.db.NewDelete().Table("teams").Where("id=?", id).Exec(ctx); err != nil {
		return fmt.Errorf("delete team: %w", err)
	}
	return nil
}

// ListByUser returns every team the user belongs to, with their role and the
// counts the team list shows. The counts are subqueries rather than a second
// round trip because the list is rendered whole.
func (r *TeamRepository) ListByUser(ctx context.Context, userID string) ([]*domain.Team, error) {
	rows, err := r.db.NewSelect().
		ColumnExpr("t.id, t.name, t.personal, t.created_by, t.created_at, t.updated_at, m.role, (SELECT COUNT(*) FROM team_members x WHERE x.team_id = t.id), (SELECT COUNT(*) FROM researches x WHERE x.team_id = t.id)").
		TableExpr("teams t JOIN team_members m ON m.team_id = t.id").
		Where("m.user_id = ?", userID).
		OrderExpr("t.personal DESC, LOWER(t.name)").
		Rows(ctx)
	if err != nil {
		return nil, fmt.Errorf("query teams: %w", err)
	}
	defer rows.Close()

	var result []*domain.Team
	for rows.Next() {
		var t domain.Team
		var personal int
		var createdBy sql.NullString
		var createdAt, updatedAt string
		if err := rows.Scan(&t.ID, &t.Name, &personal, &createdBy, &createdAt, &updatedAt,
			&t.Role, &t.MemberCount, &t.ResearchCount); err != nil {
			return nil, fmt.Errorf("scan team row: %w", err)
		}
		t.Personal = personal == 1
		if createdBy.Valid {
			t.CreatedBy = createdBy.String
		}
		t.CreatedAt, _ = time.Parse(time.DateTime, createdAt)
		t.UpdatedAt, _ = time.Parse(time.DateTime, updatedAt)
		result = append(result, &t)
	}
	return result, rows.Err()
}

// FindRole returns the user's role in the team, and whether they are a member
// at all. An empty role with ok=false is "not a member", which every caller
// must treat as not-found rather than as forbidden.
func (r *TeamRepository) FindRole(ctx context.Context, teamID, userID string) (domain.TeamRole, bool, error) {
	var role domain.TeamRole
	err := selectRow(ctx, r.db.NewSelect().
		ColumnExpr("role").
		TableExpr("team_members").
		Where("team_id=? AND user_id=?", teamID, userID)).
		Scan(&role)
	if err == sql.ErrNoRows {
		return "", false, nil
	}
	if err != nil {
		return "", false, fmt.Errorf("find team role: %w", err)
	}
	return role, true, nil
}

// ResearchRole answers the only question authorization ever asks: for this
// research, what may this user do? One query, because it runs on every read
// and every write in the product.
//
// exists reports whether the research is there at all; an empty role on an
// existing research means the caller is not in its team.
func (r *TeamRepository) ResearchRole(ctx context.Context, researchID, userID string) (role domain.TeamRole, teamID string, exists bool, err error) {
	var team sql.NullString
	var found sql.NullString
	err = selectRow(ctx, r.db.NewSelect().
		ColumnExpr("r.team_id, m.role").
		TableExpr("researches r LEFT JOIN team_members m ON m.team_id = r.team_id AND m.user_id = ?", userID).
		Where("r.id = ?", researchID)).
		Scan(&team, &found)
	if err == sql.ErrNoRows {
		return "", "", false, nil
	}
	if err != nil {
		return "", "", false, fmt.Errorf("find research role: %w", err)
	}
	if team.Valid {
		teamID = team.String
	}
	if found.Valid {
		role = domain.TeamRole(found.String)
	}
	return role, teamID, true, nil
}

func (r *TeamRepository) AddMember(ctx context.Context, teamID, userID string, role domain.TeamRole, invitedBy string) error {
	now := time.Now().UTC().Format(time.DateTime)
	var by any
	if invitedBy != "" {
		by = invitedBy
	}
	_, err := onConflict(r.db.NewInsert().Table("team_members").Model(&map[string]any{
		"team_id":    teamID,
		"user_id":    userID,
		"role":       role,
		"invited_by": by,
		"created_at": now,
	}), r.db, []string{"team_id", "user_id"}, "role").Exec(ctx)
	if err != nil {
		return fmt.Errorf("add team member: %w", err)
	}
	return nil
}

// AcceptInviteTx consumes an invitation and adds its holder to the team in one
// transaction. The two halves cannot be split: an invitation marked used with
// nobody added is a link that has been spent and a person still locked out.
//
// The `accepted_at IS NULL` guard is what decides the race when two people open
// the same link; sql.ErrNoRows means somebody else got there first.
func (r *TeamRepository) AcceptInviteTx(ctx context.Context, inviteID, teamID, userID string, role domain.TeamRole, invitedBy string) error {
	now := time.Now().UTC().Format(time.DateTime)

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	res, err := tx.NewUpdate().
		Table("team_invites").
		Set("accepted_at=?", now).
		Set("accepted_by=?", userID).
		Where("id=? AND accepted_at IS NULL AND revoked_at IS NULL", inviteID).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("accept invite: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return sql.ErrNoRows
	}

	var by any
	if invitedBy != "" {
		by = invitedBy
	}
	if _, err := onConflict(tx.NewInsert().Table("team_members").Model(&map[string]any{
		"team_id":    teamID,
		"user_id":    userID,
		"role":       role,
		"invited_by": by,
		"created_at": now,
	}), tx, []string{"team_id", "user_id"}, "role").Exec(ctx); err != nil {
		return fmt.Errorf("add team member: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	return nil
}

func (r *TeamRepository) UpdateRole(ctx context.Context, teamID, userID string, role domain.TeamRole) error {
	res, err := r.db.NewUpdate().
		Table("team_members").
		Set("role=?", role).
		Where("team_id=? AND user_id=?", teamID, userID).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("update member role: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *TeamRepository) RemoveMember(ctx context.Context, teamID, userID string) error {
	res, err := r.db.NewDelete().Table("team_members").Where("team_id=? AND user_id=?", teamID, userID).Exec(ctx)
	if err != nil {
		return fmt.Errorf("remove team member: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// CountOwners is what stops a team losing its last administrator — by
// demotion, by removal, or by its owner leaving.
func (r *TeamRepository) CountOwners(ctx context.Context, teamID string) (int, error) {
	var n int
	err := selectRow(ctx, r.db.NewSelect().
		ColumnExpr("COUNT(*)").
		TableExpr("team_members").
		Where("team_id=? AND role=?", teamID, domain.TeamOwner)).
		Scan(&n)
	if err != nil {
		return 0, fmt.Errorf("count team owners: %w", err)
	}
	return n, nil
}

func (r *TeamRepository) ListMembers(ctx context.Context, teamID string) ([]*domain.TeamMember, error) {
	rows, err := r.db.NewSelect().
		ColumnExpr("m.team_id, m.user_id, m.role, m.invited_by, m.created_at, u.email, u.name").
		TableExpr("team_members m JOIN users u ON u.id = m.user_id").
		Where("m.team_id = ?", teamID).
		OrderExpr("CASE m.role WHEN 'owner' THEN 0 WHEN 'editor' THEN 1 ELSE 2 END, LOWER(u.name), LOWER(u.email)").
		Rows(ctx)
	if err != nil {
		return nil, fmt.Errorf("query team members: %w", err)
	}
	defer rows.Close()

	var result []*domain.TeamMember
	for rows.Next() {
		var m domain.TeamMember
		var invitedBy sql.NullString
		var createdAt string
		if err := rows.Scan(&m.TeamID, &m.UserID, &m.Role, &invitedBy, &createdAt, &m.Email, &m.Name); err != nil {
			return nil, fmt.Errorf("scan team member: %w", err)
		}
		if invitedBy.Valid {
			m.InvitedBy = invitedBy.String
		}
		m.CreatedAt, _ = time.Parse(time.DateTime, createdAt)
		result = append(result, &m)
	}
	return result, rows.Err()
}
