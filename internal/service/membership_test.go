package service

import (
	"errors"
	"testing"
	"time"

	"github.com/dovod-app/app/internal/domain"
	"github.com/dovod-app/app/internal/storage"
)

func TestInvite_Lifecycle(t *testing.T) {
	k := newRoleKit(t)
	ownerUser := createTestUser(t, k.db, "owner@test.com", "Owner")
	guestUser := createTestUser(t, k.db, "guest@test.com", "Guest")
	owner, guest := userCtx(ownerUser), userCtx(guestUser)

	team, err := k.team.Create(owner, "Acme")
	if err != nil {
		t.Fatalf("create team: %v", err)
	}

	result, err := k.team.CreateInvite(owner, team.ID, "guest@test.com", domain.TeamEditor)
	if err != nil {
		t.Fatalf("create invite: %v", err)
	}
	if result.Token == "" {
		t.Fatal("the token is the whole invitation; it must come back once")
	}

	t.Run("the recipient can read it without an account", func(t *testing.T) {
		preview, err := k.team.PreviewInvite(t.Context(), result.Token)
		if err != nil {
			t.Fatalf("preview: %v", err)
		}
		if preview.Status != InvitePending {
			t.Fatalf("status = %s, want pending", preview.Status)
		}
		if preview.TeamName != "Acme" || preview.InviterName != "Owner" {
			t.Fatalf("a recipient must be told what they are joining, got %q by %q",
				preview.TeamName, preview.InviterName)
		}
		if preview.SignedIn {
			t.Error("nobody is signed in here")
		}
	})

	t.Run("signed in, it says more", func(t *testing.T) {
		preview, err := k.team.PreviewInvite(guest, result.Token)
		if err != nil {
			t.Fatalf("preview: %v", err)
		}
		if !preview.SignedIn || !preview.EmailMatches || preview.AlreadyMember {
			t.Fatalf("preview = %+v", preview)
		}
	})

	t.Run("accepting joins with the invited role", func(t *testing.T) {
		accepted, err := k.team.AcceptInvite(guest, result.Token)
		if err != nil {
			t.Fatalf("accept: %v", err)
		}
		if accepted.Role != domain.TeamEditor || accepted.TeamName != "Acme" {
			t.Fatalf("accept = %+v", accepted)
		}
		teams, err := k.team.List(guest)
		if err != nil {
			t.Fatalf("list teams: %v", err)
		}
		var found bool
		for _, tm := range teams {
			if tm.ID == team.ID {
				found = true
				if tm.Role != domain.TeamEditor {
					t.Fatalf("role = %s, want editor", tm.Role)
				}
			}
		}
		if !found {
			t.Fatal("the team is missing from the new member's list")
		}
	})

	t.Run("a link is single use", func(t *testing.T) {
		if _, err := k.team.AcceptInvite(guest, result.Token); !errors.Is(err, ErrAlreadyMember) {
			t.Fatalf("second accept = %v", err)
		}
		preview, _ := k.team.PreviewInvite(t.Context(), result.Token)
		if preview.Status != InviteAccepted {
			t.Fatalf("status = %s, want accepted", preview.Status)
		}
	})
}

func TestInvite_DeadLinksSayWhy(t *testing.T) {
	k := newRoleKit(t)
	ownerUser := createTestUser(t, k.db, "owner@test.com", "Owner")
	guest := userCtx(createTestUser(t, k.db, "guest@test.com", "Guest"))
	owner := userCtx(ownerUser)

	team, err := k.team.Create(owner, "Acme")
	if err != nil {
		t.Fatalf("create team: %v", err)
	}

	t.Run("unknown", func(t *testing.T) {
		preview, err := k.team.PreviewInvite(t.Context(), "mri_nothing")
		if err != nil {
			t.Fatalf("preview: %v", err)
		}
		if preview.Status != InviteUnknown {
			t.Fatalf("status = %s, want unknown", preview.Status)
		}
	})

	t.Run("revoked", func(t *testing.T) {
		result, err := k.team.CreateInvite(owner, team.ID, "a@test.com", domain.TeamViewer)
		if err != nil {
			t.Fatalf("create invite: %v", err)
		}
		if err := k.team.RevokeInvite(owner, result.Invite.ID); err != nil {
			t.Fatalf("revoke: %v", err)
		}
		preview, _ := k.team.PreviewInvite(t.Context(), result.Token)
		if preview.Status != InviteRevoked {
			t.Fatalf("status = %s, want revoked", preview.Status)
		}
		if _, err := k.team.AcceptInvite(guest, result.Token); !errors.Is(err, ErrInviteInvalid) {
			t.Fatalf("accepting a revoked link = %v", err)
		}
	})

	t.Run("expired", func(t *testing.T) {
		result, err := k.team.CreateInvite(owner, team.ID, "b@test.com", domain.TeamViewer)
		if err != nil {
			t.Fatalf("create invite: %v", err)
		}
		// Reach past the service to age it: the alternative is a fourteen-day
		// test.
		if _, err := k.db.ExecContext(t.Context(),
			`UPDATE team_invites SET expires_at=? WHERE id=?`,
			time.Now().UTC().Add(-time.Hour).Format(time.DateTime), result.Invite.ID); err != nil {
			t.Fatalf("age the invite: %v", err)
		}
		preview, _ := k.team.PreviewInvite(t.Context(), result.Token)
		if preview.Status != InviteExpired {
			t.Fatalf("status = %s, want expired", preview.Status)
		}
		if _, err := k.team.AcceptInvite(guest, result.Token); !errors.Is(err, ErrInviteInvalid) {
			t.Fatalf("accepting an expired link = %v", err)
		}
	})

	// A revoked link is the owner's own decision and needs no reminder. An
	// expired one is the opposite: the recipient is looking at "expired" and
	// the owner is wondering why nobody joined, so the row has to stay.
	t.Run("the owner still sees the lapsed invite, not the revoked one", func(t *testing.T) {
		invites, err := k.team.Invites(owner, team.ID)
		if err != nil {
			t.Fatalf("invites: %v", err)
		}
		if len(invites) != 1 {
			t.Fatalf("want 1 open invite, got %d", len(invites))
		}
		if invites[0].Email != "b@test.com" {
			t.Fatalf("the surviving row should be the expired one, got %s", invites[0].Email)
		}
		if invites[0].Pending(time.Now().UTC()) {
			t.Error("it should read as expired")
		}
	})
}

// A mismatched address is allowed on purpose: people are invited at work and
// signed in personally all the time, and the token is the capability. The
// preview says so; it does not refuse.
func TestInvite_MismatchedEmailIsFlaggedNotBlocked(t *testing.T) {
	k := newRoleKit(t)
	owner := userCtx(createTestUser(t, k.db, "owner@test.com", "Owner"))
	other := userCtx(createTestUser(t, k.db, "personal@test.com", "Other"))

	team, _ := k.team.Create(owner, "Acme")
	result, err := k.team.CreateInvite(owner, team.ID, "work@test.com", domain.TeamViewer)
	if err != nil {
		t.Fatalf("create invite: %v", err)
	}

	preview, err := k.team.PreviewInvite(other, result.Token)
	if err != nil {
		t.Fatalf("preview: %v", err)
	}
	if preview.EmailMatches {
		t.Error("the addresses differ and the preview should say so")
	}
	if _, err := k.team.AcceptInvite(other, result.Token); err != nil {
		t.Fatalf("a mismatch must not block: %v", err)
	}
}

func TestTeam_LastOwnerIsProtected(t *testing.T) {
	k := newRoleKit(t)
	ownerUser := createTestUser(t, k.db, "owner@test.com", "Owner")
	editorUser := createTestUser(t, k.db, "editor@test.com", "Editor")
	owner := userCtx(ownerUser)

	team, _ := k.team.Create(owner, "Acme")
	addToTeam(t, k.db, team.ID, editorUser.ID, domain.TeamEditor)

	if err := k.team.UpdateRole(owner, team.ID, ownerUser.ID, domain.TeamEditor); !errors.Is(err, ErrLastOwner) {
		t.Errorf("demoting the last owner = %v", err)
	}
	if err := k.team.RemoveMember(owner, team.ID, ownerUser.ID); !errors.Is(err, ErrLastOwner) {
		t.Errorf("the last owner leaving = %v", err)
	}

	// With a second owner in place, both become ordinary operations.
	if err := k.team.UpdateRole(owner, team.ID, editorUser.ID, domain.TeamOwner); err != nil {
		t.Fatalf("promote: %v", err)
	}
	if err := k.team.RemoveMember(owner, team.ID, ownerUser.ID); err != nil {
		t.Fatalf("leaving with a second owner: %v", err)
	}
}

func TestTeam_PersonalTeamIsFixed(t *testing.T) {
	k := newRoleKit(t)
	user := createTestUser(t, k.db, "solo@test.com", "Solo")
	ctx := userCtx(user)

	teams, err := k.team.List(ctx)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(teams) != 1 || !teams[0].Personal {
		t.Fatalf("a new user should have exactly one personal team, got %+v", teams)
	}
	personal := teams[0]

	if _, err := k.team.Rename(ctx, personal.ID, "Something else"); !errors.Is(err, ErrPersonalTeam) {
		t.Errorf("rename = %v", err)
	}
	if err := k.team.Delete(ctx, personal.ID); !errors.Is(err, ErrPersonalTeam) {
		t.Errorf("delete = %v", err)
	}
	if err := k.team.RemoveMember(ctx, personal.ID, user.ID); !errors.Is(err, ErrPersonalTeam) {
		t.Errorf("leave = %v", err)
	}
}

func TestTeam_DeleteRefusesWhileItHoldsWork(t *testing.T) {
	k := newRoleKit(t)
	owner := userCtx(createTestUser(t, k.db, "owner@test.com", "Owner"))

	team, _ := k.team.Create(owner, "Acme")
	research, _, err := k.research.Create(owner, CreateResearchRequest{
		TeamID: team.ID, Name: "Work", Goal: "G",
	})
	if err != nil {
		t.Fatalf("create research: %v", err)
	}

	if err := k.team.Delete(owner, team.ID); !errors.Is(err, ErrTeamNotEmpty) {
		t.Fatalf("deleting a team holding a research = %v", err)
	}

	// Moving the research out is what unblocks it.
	personal, err := storage.NewTeamRepository(k.db).FindPersonal(t.Context(), research.UserID)
	if err != nil || personal == nil {
		t.Fatalf("personal team: %v", err)
	}
	if err := k.team.TransferResearch(owner, research.ID, personal.ID); err != nil {
		t.Fatalf("transfer: %v", err)
	}
	if err := k.team.Delete(owner, team.ID); err != nil {
		t.Fatalf("delete after emptying: %v", err)
	}
}

func TestTeam_TransferMovesTheAudience(t *testing.T) {
	k := newRoleKit(t)
	ownerUser := createTestUser(t, k.db, "owner@test.com", "Owner")
	colleagueUser := createTestUser(t, k.db, "colleague@test.com", "Colleague")
	owner, colleague := userCtx(ownerUser), userCtx(colleagueUser)

	shared, _ := k.team.Create(owner, "Acme")
	addToTeam(t, k.db, shared.ID, colleagueUser.ID, domain.TeamViewer)

	// Starts private, in the owner's personal team.
	research, _, err := k.research.Create(owner, CreateResearchRequest{Name: "Private", Goal: "G"})
	if err != nil {
		t.Fatalf("create research: %v", err)
	}
	if _, err := k.research.Get(colleague, research.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("a colleague must not see a personal research: %v", err)
	}

	if err := k.team.TransferResearch(owner, research.ID, shared.ID); err != nil {
		t.Fatalf("transfer: %v", err)
	}

	got, err := k.research.Get(colleague, research.ID)
	if err != nil {
		t.Fatalf("after the transfer the colleague should see it: %v", err)
	}
	if got.Role != domain.TeamViewer {
		t.Fatalf("role = %s, want viewer", got.Role)
	}
	if got.TeamName != "Acme" {
		t.Fatalf("team name = %q, want Acme", got.TeamName)
	}
}

func TestTeam_BulkTransferMovesEveryOne(t *testing.T) {
	k := newRoleKit(t)
	ownerUser := createTestUser(t, k.db, "owner@test.com", "Owner")
	colleagueUser := createTestUser(t, k.db, "colleague@test.com", "Colleague")
	owner, colleague := userCtx(ownerUser), userCtx(colleagueUser)

	shared, _ := k.team.Create(owner, "Acme")
	addToTeam(t, k.db, shared.ID, colleagueUser.ID, domain.TeamViewer)

	var ids []string
	for _, name := range []string{"One", "Two", "Three"} {
		research, _, err := k.research.Create(owner, CreateResearchRequest{Name: name, Goal: "G"})
		if err != nil {
			t.Fatalf("create %s: %v", name, err)
		}
		ids = append(ids, research.ID)
	}

	// The same id twice: a form can repeat one, and it must not be counted twice.
	moved, err := k.team.TransferResearches(owner, shared.ID, append(ids, ids[0]))
	if err != nil {
		t.Fatalf("bulk transfer: %v", err)
	}
	if moved != 3 {
		t.Fatalf("moved = %d, want 3", moved)
	}
	for _, id := range ids {
		if _, err := k.research.Get(colleague, id); err != nil {
			t.Fatalf("colleague should see %s after the move: %v", id, err)
		}
	}

	// Asking again is a no-op, not a second move.
	again, err := k.team.TransferResearches(owner, shared.ID, ids)
	if err != nil {
		t.Fatalf("repeat: %v", err)
	}
	if again != 0 {
		t.Fatalf("repeat moved = %d, want 0", again)
	}
}

// One research the caller may not move must stop the whole run before anything
// moves — otherwise a rejected list half-applies and the reader is told only
// about the refusal.
func TestTeam_BulkTransferRefusesAllOrNothing(t *testing.T) {
	k := newRoleKit(t)
	ownerUser := createTestUser(t, k.db, "owner@test.com", "Owner")
	strangerUser := createTestUser(t, k.db, "stranger@test.com", "Stranger")
	owner, stranger := userCtx(ownerUser), userCtx(strangerUser)

	shared, _ := k.team.Create(owner, "Acme")

	mine, _, err := k.research.Create(owner, CreateResearchRequest{Name: "Mine", Goal: "G"})
	if err != nil {
		t.Fatalf("create mine: %v", err)
	}
	theirs, _, err := k.research.Create(stranger, CreateResearchRequest{Name: "Theirs", Goal: "G"})
	if err != nil {
		t.Fatalf("create theirs: %v", err)
	}

	// The stranger's research is second, so a run that moved as it went would
	// already have moved the first one.
	if _, err := k.team.TransferResearches(owner, shared.ID, []string{mine.ID, theirs.ID}); !errors.Is(err, ErrNotFound) {
		t.Fatalf("bulk transfer with a foreign research = %v, want ErrNotFound", err)
	}

	got, err := k.research.Get(owner, mine.ID)
	if err != nil {
		t.Fatalf("get mine: %v", err)
	}
	if got.TeamID == shared.ID {
		t.Fatal("the first research moved even though the run was refused")
	}
}

func TestTeam_CreatingInSomeoneElsesTeamIsRefused(t *testing.T) {
	k := newRoleKit(t)
	owner := userCtx(createTestUser(t, k.db, "owner@test.com", "Owner"))
	outsiderUser := createTestUser(t, k.db, "outsider@test.com", "Outsider")
	outsider := userCtx(outsiderUser)

	team, _ := k.team.Create(owner, "Acme")

	if _, _, err := k.research.Create(outsider, CreateResearchRequest{
		TeamID: team.ID, Name: "Trespass", Goal: "G",
	}); !errors.Is(err, ErrNotFound) {
		t.Fatalf("a non-member naming a team should get ErrNotFound, got %v", err)
	}

	// A viewer is a member, and still may not put work there.
	addToTeam(t, k.db, team.ID, outsiderUser.ID, domain.TeamViewer)
	if _, _, err := k.research.Create(outsider, CreateResearchRequest{
		TeamID: team.ID, Name: "Trespass", Goal: "G",
	}); !errors.Is(err, ErrForbidden) {
		t.Fatalf("a viewer creating in the team = %v", err)
	}
}
