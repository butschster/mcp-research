package domain

import "time"

// TeamRole is what a user may do inside a team, and therefore inside every
// research the team owns. There are three on purpose: a "commenter" role would
// be guessing at a comment design that does not exist yet.
type TeamRole string

const (
	// TeamViewer reads everything in the team's researches and exports them.
	TeamViewer TeamRole = "viewer"
	// TeamEditor is a viewer who may also create and change content.
	TeamEditor TeamRole = "editor"
	// TeamOwner is an editor who may also manage members, invites and the team
	// itself, and move researches between teams.
	TeamOwner TeamRole = "owner"
)

// ValidTeamRole reports whether a role read from input or from the database is
// one this product knows. An unknown role grants nothing.
func ValidTeamRole(r TeamRole) bool {
	switch r {
	case TeamViewer, TeamEditor, TeamOwner:
		return true
	}
	return false
}

// CanWrite reports whether the role may change content.
func (r TeamRole) CanWrite() bool { return r == TeamEditor || r == TeamOwner }

// CanAdmin reports whether the role may manage the team.
func (r TeamRole) CanAdmin() bool { return r == TeamOwner }

// LocalTeamID is the team that owns researches created with no user at all —
// the `auth_enabled: false` single-binary case. It has no members, and the
// access rules let an unauthenticated caller reach it, which is the access
// that mode has always had.
const LocalTeamID = "team-local"

type Team struct {
	ID string `json:"id"`
	// Name is the user's own name for a personal team, and whatever the
	// creator typed for any other.
	Name string `json:"name"`
	// Personal marks the team created alongside its user. It cannot be
	// deleted, renamed away from its owner, or left.
	Personal  bool      `json:"personal"`
	CreatedBy string    `json:"created_by,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Role is the requesting user's role in this team. It is resolved per
	// request rather than stored, so the frontend can hide controls the caller
	// cannot use instead of showing controls that fail on click.
	Role TeamRole `json:"role,omitempty"`
	// MemberCount and ResearchCount are summary fields for the team list.
	MemberCount   int `json:"member_count"`
	ResearchCount int `json:"research_count"`
}

type TeamMember struct {
	TeamID    string    `json:"team_id"`
	UserID    string    `json:"user_id"`
	Role      TeamRole  `json:"role"`
	InvitedBy string    `json:"invited_by,omitempty"`
	CreatedAt time.Time `json:"created_at"`

	// Joined from users, so a member list reads as people rather than as ids.
	Email string `json:"email,omitempty"`
	Name  string `json:"name,omitempty"`
}

type TeamInvite struct {
	ID     string   `json:"id"`
	TeamID string   `json:"team_id"`
	Email  string   `json:"email"`
	Role   TeamRole `json:"role"`
	// InvitedBy is the user id; InviterName is joined for display.
	InvitedBy string `json:"invited_by"`
	// InviterName and InviterEmail are joined for display. The address is the
	// only next action an expired invitation can offer its recipient.
	InviterName  string     `json:"inviter_name,omitempty"`
	InviterEmail string     `json:"inviter_email,omitempty"`
	TeamName     string     `json:"team_name,omitempty"`
	ExpiresAt    time.Time  `json:"expires_at"`
	AcceptedAt   *time.Time `json:"accepted_at,omitempty"`
	AcceptedBy   string     `json:"accepted_by,omitempty"`
	RevokedAt    *time.Time `json:"revoked_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
}

// Pending reports whether the invite can still be accepted at the given time.
// Expiry, revocation and prior acceptance are one question with one answer:
// callers must not each reimplement two thirds of it.
func (i *TeamInvite) Pending(now time.Time) bool {
	return i != nil && i.AcceptedAt == nil && i.RevokedAt == nil && i.ExpiresAt.After(now)
}
