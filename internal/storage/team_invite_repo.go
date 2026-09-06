package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/dovod-app/app/internal/domain"
	"github.com/uptrace/bun"
)

type TeamInviteRepository struct {
	db *bun.DB
}

func NewTeamInviteRepository(db *bun.DB) *TeamInviteRepository {
	return &TeamInviteRepository{db: db}
}

// inviteColumns joins the names an invite is displayed with. A recipient
// opening a link needs to read "Anna invited you to Acme", not two ids.
const inviteColumns = `i.id, i.team_id, i.email, i.role, i.invited_by, i.expires_at,
	i.accepted_at, i.accepted_by, i.revoked_at, i.created_at,
	COALESCE(t.name, ''), COALESCE(u.name, u.email, ''), COALESCE(u.email, '')`

const inviteFrom = ` FROM team_invites i
	LEFT JOIN teams t ON t.id = i.team_id
	LEFT JOIN users u ON u.id = i.invited_by`

func (r *TeamInviteRepository) Create(ctx context.Context, inv *domain.TeamInvite, tokenHash string) error {
	now := time.Now().UTC().Format(time.DateTime)
	_, err := r.db.NewInsert().Table("team_invites").Model(&map[string]any{
		"id":         inv.ID,
		"team_id":    inv.TeamID,
		"email":      inv.Email,
		"role":       inv.Role,
		"token_hash": tokenHash,
		"invited_by": inv.InvitedBy,
		"expires_at": inv.ExpiresAt.UTC().Format(time.DateTime),
		"created_at": now,
	}).Exec(ctx)
	if err != nil {
		return fmt.Errorf("insert invite: %w", err)
	}
	inv.CreatedAt, _ = time.Parse(time.DateTime, now)
	return nil
}

// FindByHash is the lookup a link performs. The plaintext token is never
// stored, so a leaked database hands out no working invitations.
func (r *TeamInviteRepository) FindByHash(ctx context.Context, tokenHash string) (*domain.TeamInvite, error) {
	row := selectRow(ctx, r.db.NewSelect().
		ColumnExpr(inviteColumns).
		TableExpr("team_invites i LEFT JOIN teams t ON t.id = i.team_id LEFT JOIN users u ON u.id = i.invited_by").
		Where("i.token_hash = ?", tokenHash))
	return scanInvite(row)
}

func (r *TeamInviteRepository) FindByID(ctx context.Context, id string) (*domain.TeamInvite, error) {
	row := selectRow(ctx, r.db.NewSelect().
		ColumnExpr(inviteColumns).
		TableExpr("team_invites i LEFT JOIN teams t ON t.id = i.team_id LEFT JOIN users u ON u.id = i.invited_by").
		Where("i.id = ?", id))
	return scanInvite(row)
}

// ListOpenByTeam returns the invitations that are still someone's business:
// not accepted, not revoked.
//
// Expired ones are included on purpose. An invitation that quietly lapsed is
// exactly what an owner needs to see — the recipient is looking at "expired"
// and the owner, with the row filtered away, has no idea why nobody joined.
func (r *TeamInviteRepository) ListOpenByTeam(ctx context.Context, teamID string) ([]*domain.TeamInvite, error) {
	rows, err := r.db.NewSelect().
		ColumnExpr(inviteColumns).
		TableExpr("team_invites i LEFT JOIN teams t ON t.id = i.team_id LEFT JOIN users u ON u.id = i.invited_by").
		Where("i.team_id = ? AND i.accepted_at IS NULL AND i.revoked_at IS NULL", teamID).
		OrderExpr("i.created_at DESC").
		Rows(ctx)
	if err != nil {
		return nil, fmt.Errorf("query invites: %w", err)
	}
	defer rows.Close()

	var result []*domain.TeamInvite
	for rows.Next() {
		inv, err := scanInvite(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, inv)
	}
	return result, rows.Err()
}

func (r *TeamInviteRepository) Revoke(ctx context.Context, id string) error {
	now := time.Now().UTC().Format(time.DateTime)
	res, err := r.db.NewUpdate().
		Table("team_invites").
		Set("revoked_at=?", now).
		Where("id=? AND revoked_at IS NULL", id).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("revoke invite: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// MarkAccepted closes the invite. The WHERE clause carries the guard so two
// people opening the same link cannot both consume it.
func (r *TeamInviteRepository) MarkAccepted(ctx context.Context, id, userID string) error {
	now := time.Now().UTC().Format(time.DateTime)
	res, err := r.db.NewUpdate().
		Table("team_invites").
		Set("accepted_at=?", now).
		Set("accepted_by=?", userID).
		Where("id=? AND accepted_at IS NULL AND revoked_at IS NULL", id).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("accept invite: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func scanInvite(s scanner) (*domain.TeamInvite, error) {
	var inv domain.TeamInvite
	var acceptedAt, acceptedBy, revokedAt sql.NullString
	var expiresAt, createdAt string
	err := s.Scan(&inv.ID, &inv.TeamID, &inv.Email, &inv.Role, &inv.InvitedBy, &expiresAt,
		&acceptedAt, &acceptedBy, &revokedAt, &createdAt,
		&inv.TeamName, &inv.InviterName, &inv.InviterEmail)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("scan invite: %w", err)
	}
	inv.ExpiresAt, _ = time.Parse(time.DateTime, expiresAt)
	inv.CreatedAt, _ = time.Parse(time.DateTime, createdAt)
	if acceptedAt.Valid {
		t, _ := time.Parse(time.DateTime, acceptedAt.String)
		inv.AcceptedAt = &t
	}
	if revokedAt.Valid {
		t, _ := time.Parse(time.DateTime, revokedAt.String)
		inv.RevokedAt = &t
	}
	if acceptedBy.Valid {
		inv.AcceptedBy = acceptedBy.String
	}
	return &inv, nil
}
