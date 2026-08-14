package service

import (
	"context"

	"github.com/butschster/mcp-research/internal/auth"
	"github.com/butschster/mcp-research/internal/domain"
	"github.com/butschster/mcp-research/internal/storage"
)

// Access decides what the caller may do to a research. It is the only place in
// the product where that decision is made, and every service is handed one at
// construction so the check cannot be forgotten on the next method: a service
// built without it fails on its first call rather than quietly letting
// everyone through.
//
// Access is the team's, not the user's. `researches.user_id` records who
// created a research and is never consulted here — two sources of truth for
// who may do what is how this class of bug happens.
//
// Two rules hold across every method:
//
//   - **Nobody in the context means no check.** That is the local
//     single-binary case (`auth_enabled: false`), where there are no users to
//     be wrong about. It is the behaviour this product has always had; turning
//     auth on is what takes it away.
//   - **A non-member gets ErrNotFound, never a refusal.** Confirming that a
//     research exists is itself information about someone else's work. A
//     member who merely lacks the right gets ErrForbidden, because hiding a
//     research from someone who can already read it protects nothing.
//
// There is deliberately no owner-only tier here. Everything that needs one —
// moving a research, managing who can see it — is a decision about the *team*,
// and TeamService.requireRole makes it.
type Access struct {
	teams *storage.TeamRepository
}

func NewAccess(teams *storage.TeamRepository) *Access {
	return &Access{teams: teams}
}

// Read allows any member of the owning team.
func (a *Access) Read(ctx context.Context, researchID string) error {
	_, err := a.Role(ctx, researchID)
	return err
}

// Write allows an editor or an owner to change content.
func (a *Access) Write(ctx context.Context, researchID string) error {
	role, err := a.Role(ctx, researchID)
	if err != nil {
		return err
	}
	// An empty role reaches here only when there is no authenticated caller,
	// which Role has already allowed.
	if role != "" && !role.CanWrite() {
		return ErrForbidden
	}
	return nil
}

// Role resolves the caller's role, or ErrNotFound when the research is absent
// or belongs to a team they are not in. An empty role with a nil error means
// there is no authenticated caller at all.
//
// Existence is checked even then. "No check" is about permission, not about
// whether the thing is there: without this, a bad id in local mode would slip
// past every service and surface as a foreign-key error from the database.
func (a *Access) Role(ctx context.Context, researchID string) (domain.TeamRole, error) {
	uid := auth.UserIDFromContext(ctx)
	role, teamID, exists, err := a.teams.ResearchRole(ctx, researchID, uid)
	if err != nil {
		return "", err
	}
	if !exists {
		return "", ErrNotFound
	}
	if uid == "" {
		return "", nil
	}
	// The local team holds researches created with no user at all, and it has
	// no members. Leaving them to the membership rule would strand them behind
	// a team nobody can join; granting the caller ownership would let the
	// first person to press "move to team" take them from everyone else, which
	// is a capability that did not exist before teams.
	//
	// So: readable, and no more. That is a tightening of what an ownerless
	// research allowed before — it was writable too — and the safe direction
	// to be wrong in. The first registration still claims them into a real
	// team, which is the path out.
	if teamID == domain.LocalTeamID {
		return domain.TeamViewer, nil
	}
	if role == "" || !domain.ValidTeamRole(role) {
		return "", ErrNotFound
	}
	return role, nil
}

// RoleOrEmpty is Role without the error, for decorating a response with what
// the caller may do. A failure to resolve is reported as no role, which the
// frontend renders as read-only — the safe direction to be wrong in.
func (a *Access) RoleOrEmpty(ctx context.Context, researchID string) domain.TeamRole {
	role, err := a.Role(ctx, researchID)
	if err != nil {
		return ""
	}
	if role == "" {
		// Local mode: there is no team to have a role in, and everything is
		// permitted, so the UI must not draw a read-only screen.
		return domain.TeamOwner
	}
	return role
}
