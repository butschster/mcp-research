package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/butschster/mcp-research/internal/auth"
	"github.com/butschster/mcp-research/internal/domain"
	"github.com/butschster/mcp-research/internal/storage"
	"github.com/google/uuid"
)

var (
	// ErrLastOwner guards the only irreversible mistake in team management: a
	// team with no owner can never be administered again, by anyone.
	ErrLastOwner = errors.New("a team must keep at least one owner")
	// ErrPersonalTeam is for the operations a personal team does not have. It
	// is where a solo user's researches live; deleting or leaving it would
	// strand them.
	ErrPersonalTeam  = errors.New("a personal team cannot be renamed, deleted or left")
	ErrTeamNotEmpty  = errors.New("move or delete the team's researches first")
	ErrInviteInvalid = errors.New("this invitation is no longer valid")
	ErrAlreadyMember = errors.New("already a member of this team")
	ErrNotMember     = errors.New("not a member of this team")
	ErrInvalidRole   = errors.New("role must be viewer, editor or owner")
	ErrTeamNameEmpty = errors.New("a team needs a name")
	// ErrNoAuth is for the team operations that are meaningless without an
	// identity — every one of them, since a team is a set of people.
	ErrNoAuth = errors.New("sign in to manage teams")
)

// inviteTTL is how long an invitation link works. Long enough to be passed
// along by hand over a weekend, short enough that a link pasted into a channel
// two years ago is not a way in.
const inviteTTL = 14 * 24 * time.Hour

// invitePrefix marks a token as this product's, so one pasted into the wrong
// field is recognisable rather than mysterious.
const invitePrefix = "mri_"

type TeamService struct {
	teams      *storage.TeamRepository
	invites    *storage.TeamInviteRepository
	users      *storage.UserRepository
	researches *storage.ResearchRepository
	events     EventNotifier
	log        *slog.Logger
}

func NewTeamService(
	teams *storage.TeamRepository,
	invites *storage.TeamInviteRepository,
	users *storage.UserRepository,
	researches *storage.ResearchRepository,
	events EventNotifier,
	log *slog.Logger,
) *TeamService {
	return &TeamService{teams: teams, invites: invites, users: users, researches: researches, events: events, log: log}
}

// --- Teams ---

func (s *TeamService) List(ctx context.Context) ([]*domain.Team, error) {
	uid := auth.UserIDFromContext(ctx)
	if uid == "" {
		return nil, ErrNoAuth
	}
	teams, err := s.teams.ListByUser(ctx, uid)
	if err != nil {
		return nil, err
	}
	if teams == nil {
		teams = []*domain.Team{}
	}
	return teams, nil
}

// Get returns a team the caller belongs to, with their role attached.
func (s *TeamService) Get(ctx context.Context, teamID string) (*domain.Team, error) {
	role, err := s.requireRole(ctx, teamID, domain.TeamViewer)
	if err != nil {
		return nil, err
	}
	team, err := s.teams.FindByID(ctx, teamID)
	if err != nil {
		return nil, err
	}
	if team == nil {
		return nil, ErrNotFound
	}
	team.Role = role
	return team, nil
}

func (s *TeamService) Create(ctx context.Context, name string) (*domain.Team, error) {
	uid := auth.UserIDFromContext(ctx)
	if uid == "" {
		return nil, ErrNoAuth
	}
	name = normalizeTitle(name)
	if name == "" {
		return nil, ErrTeamNameEmpty
	}

	team := &domain.Team{
		ID:        uuid.New().String(),
		Name:      name,
		CreatedBy: uid,
	}
	if err := s.teams.CreateWithOwner(ctx, team, uid); err != nil {
		return nil, err
	}
	team.MemberCount = 1
	s.notify("team.created", team.ID)
	return team, nil
}

func (s *TeamService) Rename(ctx context.Context, teamID, name string) (*domain.Team, error) {
	if _, err := s.requireRole(ctx, teamID, domain.TeamOwner); err != nil {
		return nil, err
	}
	team, err := s.teams.FindByID(ctx, teamID)
	if err != nil {
		return nil, err
	}
	if team == nil {
		return nil, ErrNotFound
	}
	if team.Personal {
		return nil, ErrPersonalTeam
	}
	name = normalizeTitle(name)
	if name == "" {
		return nil, ErrTeamNameEmpty
	}
	if err := s.teams.Rename(ctx, teamID, name); err != nil {
		return nil, err
	}
	team.Name = name
	team.Role = domain.TeamOwner
	s.notify("team.updated", teamID)
	return team, nil
}

// Delete removes a team once its researches have been moved elsewhere. The
// schema would cascade them away instead; refusing is the kinder answer,
// because "delete team" does not read as "delete everything in it".
func (s *TeamService) Delete(ctx context.Context, teamID string) error {
	if _, err := s.requireRole(ctx, teamID, domain.TeamOwner); err != nil {
		return err
	}
	team, err := s.teams.FindByID(ctx, teamID)
	if err != nil {
		return err
	}
	if team == nil {
		return ErrNotFound
	}
	if team.Personal {
		return ErrPersonalTeam
	}
	n, err := s.researches.CountByTeam(ctx, teamID)
	if err != nil {
		return err
	}
	if n > 0 {
		return ErrTeamNotEmpty
	}
	if err := s.teams.Delete(ctx, teamID); err != nil {
		return err
	}
	s.notify("team.deleted", teamID)
	return nil
}

// --- Members ---

func (s *TeamService) Members(ctx context.Context, teamID string) ([]*domain.TeamMember, error) {
	if _, err := s.requireRole(ctx, teamID, domain.TeamViewer); err != nil {
		return nil, err
	}
	members, err := s.teams.ListMembers(ctx, teamID)
	if err != nil {
		return nil, err
	}
	if members == nil {
		members = []*domain.TeamMember{}
	}
	return members, nil
}

func (s *TeamService) UpdateRole(ctx context.Context, teamID, userID string, role domain.TeamRole) error {
	if _, err := s.requireRole(ctx, teamID, domain.TeamOwner); err != nil {
		return err
	}
	if !domain.ValidTeamRole(role) {
		return ErrInvalidRole
	}

	current, ok, err := s.teams.FindRole(ctx, teamID, userID)
	if err != nil {
		return err
	}
	if !ok {
		return ErrNotMember
	}
	if current == role {
		return nil
	}
	if current == domain.TeamOwner {
		if err := s.guardLastOwner(ctx, teamID); err != nil {
			return err
		}
	}

	if err := s.teams.UpdateRole(ctx, teamID, userID, role); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotMember
		}
		return err
	}
	s.notify("team.member_role_changed", teamID)
	return nil
}

// RemoveMember drops someone from a team. It doubles as "leave team" when the
// caller names themselves, which is the same operation with a different
// permission: anyone may leave, only an owner may remove someone else.
func (s *TeamService) RemoveMember(ctx context.Context, teamID, userID string) error {
	uid := auth.UserIDFromContext(ctx)
	if uid == "" {
		return ErrNoAuth
	}

	leaving := uid == userID
	needed := domain.TeamOwner
	if leaving {
		needed = domain.TeamViewer
	}
	if _, err := s.requireRole(ctx, teamID, needed); err != nil {
		return err
	}

	team, err := s.teams.FindByID(ctx, teamID)
	if err != nil {
		return err
	}
	if team == nil {
		return ErrNotFound
	}
	if team.Personal {
		return ErrPersonalTeam
	}

	current, ok, err := s.teams.FindRole(ctx, teamID, userID)
	if err != nil {
		return err
	}
	if !ok {
		return ErrNotMember
	}
	if current == domain.TeamOwner {
		if err := s.guardLastOwner(ctx, teamID); err != nil {
			return err
		}
	}

	if err := s.teams.RemoveMember(ctx, teamID, userID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotMember
		}
		return err
	}
	s.notify("team.member_removed", teamID)
	return nil
}

// guardLastOwner refuses the change that would leave a team unadministrable.
func (s *TeamService) guardLastOwner(ctx context.Context, teamID string) error {
	owners, err := s.teams.CountOwners(ctx, teamID)
	if err != nil {
		return err
	}
	if owners <= 1 {
		return ErrLastOwner
	}
	return nil
}

// --- Invitations ---

// InviteResult carries the one and only sighting of the token. It is hashed at
// rest, so an owner who closes the dialog without copying the link has to
// issue a new invitation — the same trade the API keys already make.
type InviteResult struct {
	Invite *domain.TeamInvite `json:"invite"`
	Token  string             `json:"token"`
}

func (s *TeamService) Invites(ctx context.Context, teamID string) ([]*domain.TeamInvite, error) {
	if _, err := s.requireRole(ctx, teamID, domain.TeamOwner); err != nil {
		return nil, err
	}
	invites, err := s.invites.ListOpenByTeam(ctx, teamID)
	if err != nil {
		return nil, err
	}
	if invites == nil {
		invites = []*domain.TeamInvite{}
	}
	return invites, nil
}

func (s *TeamService) CreateInvite(ctx context.Context, teamID, email string, role domain.TeamRole) (*InviteResult, error) {
	uid := auth.UserIDFromContext(ctx)
	if _, err := s.requireRole(ctx, teamID, domain.TeamOwner); err != nil {
		return nil, err
	}

	// A personal team refuses every removal, so letting someone in would be a
	// one-way door: neither the owner nor the joiner could undo it.
	team, err := s.teams.FindByID(ctx, teamID)
	if err != nil {
		return nil, err
	}
	if team == nil {
		return nil, ErrNotFound
	}
	if team.Personal {
		return nil, ErrPersonalTeam
	}
	if role == "" {
		role = domain.TeamViewer
	}
	if !domain.ValidTeamRole(role) {
		return nil, ErrInvalidRole
	}

	email = strings.TrimSpace(strings.ToLower(email))

	// Inviting someone who is already here is a mistake worth naming rather
	// than a link that will dead-end when they follow it.
	if email != "" {
		if user, err := s.users.FindByEmail(ctx, email); err == nil && user != nil {
			if _, ok, err := s.teams.FindRole(ctx, teamID, user.ID); err == nil && ok {
				return nil, ErrAlreadyMember
			}
		}
	}

	token := invitePrefix + strings.ReplaceAll(uuid.New().String(), "-", "") +
		strings.ReplaceAll(uuid.New().String(), "-", "")
	invite := &domain.TeamInvite{
		ID:        uuid.New().String(),
		TeamID:    teamID,
		Email:     email,
		Role:      role,
		InvitedBy: uid,
		ExpiresAt: time.Now().UTC().Add(inviteTTL),
	}
	if err := s.invites.Create(ctx, invite, auth.HashAPIKey(token)); err != nil {
		return nil, err
	}
	s.notify("team.invited", teamID)
	return &InviteResult{Invite: invite, Token: token}, nil
}

func (s *TeamService) RevokeInvite(ctx context.Context, inviteID string) error {
	invite, err := s.invites.FindByID(ctx, inviteID)
	if err != nil {
		return err
	}
	if invite == nil {
		return ErrNotFound
	}
	if _, err := s.requireRole(ctx, invite.TeamID, domain.TeamOwner); err != nil {
		return err
	}
	if err := s.invites.Revoke(ctx, inviteID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}
	s.notify("team.invite_revoked", invite.TeamID)
	return nil
}

// InviteStatus is why a link cannot be used, or that it can. The recipient's
// next action differs for each, so collapsing them into one 404 would leave
// them with a dead page and no idea what to do.
type InviteStatus string

const (
	InvitePending  InviteStatus = "pending"
	InviteExpired  InviteStatus = "expired"
	InviteRevoked  InviteStatus = "revoked"
	InviteAccepted InviteStatus = "accepted"
	InviteUnknown  InviteStatus = "unknown"
)

// InvitePreview is what the landing page renders before anyone commits to
// anything. It answers deliberately without requiring a session: a recipient
// who is not signed in still has to be told what they are being invited to,
// and by whom, before being asked to create an account.
type InvitePreview struct {
	Status      InviteStatus    `json:"status"`
	TeamID      string          `json:"team_id,omitempty"`
	TeamName    string          `json:"team_name,omitempty"`
	Role        domain.TeamRole `json:"role,omitempty"`
	Email       string          `json:"email,omitempty"`
	InviterName string          `json:"inviter_name,omitempty"`
	// InviterEmail is the only next action an expired link can offer.
	InviterEmail string     `json:"inviter_email,omitempty"`
	ExpiresAt    *time.Time `json:"expires_at,omitempty"`
	// AlreadyMember and EmailMatches are resolved only when the request
	// carries a session, so the page can say "you are already in this team" or
	// "this invitation was sent to someone else" instead of finding out on
	// submit.
	AlreadyMember bool `json:"already_member"`
	EmailMatches  bool `json:"email_matches"`
	SignedIn      bool `json:"signed_in"`
}

// PreviewInvite reads a link without consuming it. It takes no permission at
// all: whoever holds the token is the audience.
func (s *TeamService) PreviewInvite(ctx context.Context, token string) (*InvitePreview, error) {
	invite, err := s.invites.FindByHash(ctx, auth.HashAPIKey(token))
	if err != nil {
		return nil, err
	}
	if invite == nil {
		return &InvitePreview{Status: InviteUnknown}, nil
	}

	preview := &InvitePreview{
		Status:       InvitePending,
		TeamID:       invite.TeamID,
		TeamName:     invite.TeamName,
		Role:         invite.Role,
		Email:        invite.Email,
		InviterName:  invite.InviterName,
		InviterEmail: invite.InviterEmail,
	}
	expires := invite.ExpiresAt
	preview.ExpiresAt = &expires

	switch {
	case invite.AcceptedAt != nil:
		preview.Status = InviteAccepted
	case invite.RevokedAt != nil:
		// Revoking is the owner withdrawing the invitation, most often
		// because it went to the wrong person. Going on naming the team and
		// the sender to whoever holds the link would leave the disclosure
		// standing after the access was taken back.
		return &InvitePreview{Status: InviteRevoked}, nil
	case !invite.ExpiresAt.After(time.Now().UTC()):
		preview.Status = InviteExpired
	}

	if user := auth.UserFromContext(ctx); user != nil {
		preview.SignedIn = true
		preview.EmailMatches = invite.Email == "" || strings.EqualFold(invite.Email, user.Email)
		if _, ok, err := s.teams.FindRole(ctx, invite.TeamID, user.ID); err == nil && ok {
			preview.AlreadyMember = true
		}
	}
	return preview, nil
}

// AcceptResult is what the caller needs to land somewhere useful afterwards.
type AcceptResult struct {
	TeamID   string          `json:"team_id"`
	TeamName string          `json:"team_name"`
	Role     domain.TeamRole `json:"role"`
}

// AcceptInvite joins the signed-in user to the team.
//
// A mismatch between the invited address and the signed-in one is allowed on
// purpose: people are invited at a work address and signed in with a personal
// one all the time, and the token is the capability. The frontend warns; it
// does not block.
func (s *TeamService) AcceptInvite(ctx context.Context, token string) (*AcceptResult, error) {
	user := auth.UserFromContext(ctx)
	if user == nil {
		return nil, ErrNoAuth
	}

	invite, err := s.invites.FindByHash(ctx, auth.HashAPIKey(token))
	if err != nil {
		return nil, err
	}
	if invite == nil {
		return nil, ErrInviteInvalid
	}

	// Membership first, so someone reopening the link they already used is
	// told they are in rather than that their invitation is broken.
	if _, ok, err := s.teams.FindRole(ctx, invite.TeamID, user.ID); err == nil && ok {
		return nil, ErrAlreadyMember
	}
	if !invite.Pending(time.Now().UTC()) {
		return nil, ErrInviteInvalid
	}

	// Consuming the link and joining the team are one act. Split, a failure
	// on the second leaves the invitation spent and the person still outside,
	// with no way back in. The WHERE clause inside still decides the race
	// between two people opening the same link.
	if err := s.teams.AcceptInviteTx(ctx, invite.ID, invite.TeamID, user.ID, invite.Role, invite.InvitedBy); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrInviteInvalid
		}
		return nil, err
	}

	s.notify("team.member_added", invite.TeamID)
	return &AcceptResult{TeamID: invite.TeamID, TeamName: invite.TeamName, Role: invite.Role}, nil
}

// --- Research transfer ---

// TransferResearch moves a research to another team, which is the only way its
// audience changes. It takes ownership on the way out and write access on the
// way in: handing your work to a team you merely read would be a way to lose
// it, and pushing it into a team you have no say in would be a way to litter.
func (s *TeamService) TransferResearch(ctx context.Context, researchID, targetTeamID string) error {
	uid := auth.UserIDFromContext(ctx)
	research, err := s.researches.FindByID(ctx, researchID)
	if err != nil {
		return err
	}
	if research == nil {
		return ErrNotFound
	}

	if uid != "" {
		if _, err := s.requireRole(ctx, research.TeamID, domain.TeamOwner); err != nil {
			return err
		}
		role, ok, err := s.teams.FindRole(ctx, targetTeamID, uid)
		if err != nil {
			return err
		}
		if !ok {
			return ErrNotFound
		}
		if !role.CanWrite() {
			return ErrForbidden
		}
	}

	if research.TeamID == targetTeamID {
		return nil
	}
	// With no caller the role checks above are skipped, so the target still has
	// to be shown to exist — otherwise a typo comes back as a foreign-key
	// failure from the database.
	target, err := s.teams.FindByID(ctx, targetTeamID)
	if err != nil {
		return err
	}
	if target == nil {
		return ErrNotFound
	}
	if err := s.researches.SetTeam(ctx, researchID, targetTeamID); err != nil {
		return err
	}
	s.events.Notify(Event{Type: "research.transferred", ResearchID: researchID, EntityID: researchID, Entity: "research"})
	return nil
}

// requireRole is the team-level counterpart of Access: a non-member is told
// the team does not exist, and a member who is merely too junior is told no.
func (s *TeamService) requireRole(ctx context.Context, teamID string, needed domain.TeamRole) (domain.TeamRole, error) {
	uid := auth.UserIDFromContext(ctx)
	if uid == "" {
		return "", ErrNoAuth
	}
	role, ok, err := s.teams.FindRole(ctx, teamID, uid)
	if err != nil {
		return "", fmt.Errorf("find team role: %w", err)
	}
	if !ok || !domain.ValidTeamRole(role) {
		return "", ErrNotFound
	}
	switch needed {
	case domain.TeamOwner:
		if !role.CanAdmin() {
			return role, ErrForbidden
		}
	case domain.TeamEditor:
		if !role.CanWrite() {
			return role, ErrForbidden
		}
	}
	return role, nil
}

func (s *TeamService) notify(eventType, teamID string) {
	s.events.Notify(Event{Type: eventType, EntityID: teamID, Entity: "team"})
}
