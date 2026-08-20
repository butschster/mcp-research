package handlers

import (
	"context"

	"github.com/butschster/mcp-research/internal/storage"
)

// resolveAuthorName turns a user id into something a reader recognises — but
// only for somebody entitled to recognise it.
//
// The right to read a record does not carry the right to know who made it.
// Those two come apart through ordinary use: a research moves to another team
// and every member of the new one inherits work by people they have never
// shared a team with; a member leaves and whoever joins afterwards reads their
// name off everything they touched. Membership in the team that owns the
// research NOW is the boundary, and it is checked here rather than assumed from
// the read that got the caller this far.
//
// The email is deliberately not a fallback. It would fire for exactly the
// accounts that never set a name — API keys, OAuth clients, --default-user —
// turning a display nicety into a bulk address disclosure.
//
// Everything else fails quietly: a deleted user, a database hiccup or auth
// being off all mean "no name", and every line that shows one stands without it.
func resolveAuthorName(ctx context.Context, users *storage.UserRepository, teams *storage.TeamRepository, teamID, userID string) string {
	if users == nil || teams == nil || userID == "" || teamID == "" {
		return ""
	}
	if _, ok, err := teams.FindRole(ctx, teamID, userID); err != nil || !ok {
		return ""
	}
	user, err := users.FindByID(ctx, userID)
	if err != nil || user == nil {
		return ""
	}
	return user.Name
}
