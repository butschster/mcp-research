package service

import (
	"errors"
	"testing"

	"github.com/butschster/mcp-research/internal/auth"
	"github.com/butschster/mcp-research/internal/domain"
	"github.com/butschster/mcp-research/internal/storage"
)

// Two queries decided who could see what without going through Access, and
// both were still comparing creators after the sweep: the search box and
// cross-reference resolution. Each was wrong twice over — it hid a colleague's
// work inside a shared team, and it exposed ownerless researches to people who
// could not open them.

func TestReach_SearchFollowsTheTeam(t *testing.T) {
	k := newRoleKit(t)
	owner, member, research, section, _ := k.sharedResearch(t, domain.TeamEditor)
	stranger := createTestUser(t, k.db, "stranger@test.com", "Stranger")

	if _, err := k.entry.Create(owner, CreateEntryRequest{
		ResearchID: research.ID, SectionID: section.ID,
		Title: "Zaphod deployment notes", Content: "The bastion is at 10.0.0.1",
	}); err != nil {
		t.Fatalf("seed entry: %v", err)
	}

	entries := storage.NewEntryRepository(k.db)
	memberID := auth.UserIDFromContext(member)

	found, err := entries.SearchEntries(member, "Zaphod", 20, memberID, "")
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(found) != 1 {
		t.Fatalf("a colleague in the same team got %d results, want 1 — the search box is the "+
			"main way into a shared research", len(found))
	}

	found, err = entries.SearchEntries(userCtx(stranger), "Zaphod", 20, stranger.ID, "")
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(found) != 0 {
		t.Fatalf("someone outside the team got %d results, want 0", len(found))
	}
}

func TestReach_CrossReferenceFollowsTheTeam(t *testing.T) {
	k := newRoleKit(t)
	owner, member, research, section, teamID := k.sharedResearch(t, domain.TeamEditor)

	target, err := k.entry.Create(owner, CreateEntryRequest{
		ResearchID: research.ID, SectionID: section.ID,
		Title: "The referenced entry", Content: "body",
	})
	if err != nil {
		t.Fatalf("seed target: %v", err)
	}

	// The member writes their own research in the same team and points at it.
	own, sections, err := k.research.Create(member, CreateResearchRequest{
		TeamID: teamID, Name: "Member's own", Goal: "G",
		Sections: []CreateSectionRequest{{Name: "s1", DisplayName: "S1"}},
	})
	if err != nil {
		t.Fatalf("create research: %v", err)
	}
	if _, err := k.entry.Create(member, CreateEntryRequest{
		ResearchID: own.ID, SectionID: sections[0].ID,
		Title: "Points across", Content: "See [[" + research.Code + ":" + target.Code + "]].",
	}); err != nil {
		t.Fatalf("create entry: %v", err)
	}

	refs, err := storage.NewCrossRefRepository(k.db).FindByResearch(member, own.ID)
	if err != nil {
		t.Fatalf("crossrefs: %v", err)
	}
	var resolved bool
	for _, ref := range refs {
		if ref.TargetEntryID == target.ID {
			resolved = true
		}
	}
	if !resolved {
		t.Fatal("a reference to a colleague's entry in the same team stayed unresolved — " +
			"it renders as plain text forever")
	}
}

// The local team holds researches created with nobody in the context. It has
// no members, so the membership rule alone would strand them: unreadable,
// unlistable, and unmovable, with no path in the product to get them back.
func TestReach_LocalTeamStaysReachable(t *testing.T) {
	k := newRoleKit(t)

	// No user in the context — the `auth_enabled: false` path.
	orphan, _, err := k.research.Create(t.Context(), CreateResearchRequest{Name: "Local work", Goal: "G"})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if orphan.TeamID != domain.LocalTeamID {
		t.Fatalf("team = %q, want %q", orphan.TeamID, domain.LocalTeamID)
	}

	// A user who registers later — not the first, so nothing claims it.
	createTestUser(t, k.db, "first@test.com", "First")
	second := userCtx(createTestUser(t, k.db, "second@test.com", "Second"))

	if _, err := k.research.Get(second, orphan.ID); err != nil {
		t.Fatalf("an authenticated user must still reach it: %v", err)
	}
	list, err := k.research.List(second, storage.ResearchFilter{})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	var listed bool
	for _, r := range list {
		if r.ID == orphan.ID {
			listed = true
		}
	}
	if !listed {
		t.Error("it is missing from the list, which is how it becomes invisible")
	}

	// Reading is all it grants. Writing would restore the pre-teams behaviour,
	// and moving it would let whoever gets there first take a research the
	// whole instance can currently see.
	if _, err := k.research.Update(second, orphan.ID, UpdateResearchRequest{Name: ptr("Mine now")}); !errors.Is(err, ErrForbidden) {
		t.Errorf("writing to a local-team research = %v, want ErrForbidden", err)
	}
	team, err := k.team.Create(second, "Rescue")
	if err != nil {
		t.Fatalf("create team: %v", err)
	}
	if err := k.team.TransferResearch(second, orphan.ID, team.ID); !errors.Is(err, ErrNotFound) {
		t.Errorf("claiming a local-team research = %v, want ErrNotFound", err)
	}
}

// A personal team refuses every removal, so admitting anyone would be a door
// that locks behind them.
func TestReach_PersonalTeamTakesNoMembers(t *testing.T) {
	k := newRoleKit(t)
	user := createTestUser(t, k.db, "solo@test.com", "Solo")
	ctx := userCtx(user)

	personal, err := storage.NewTeamRepository(k.db).FindPersonal(t.Context(), user.ID)
	if err != nil || personal == nil {
		t.Fatalf("personal team: %v", err)
	}

	if _, err := k.team.CreateInvite(ctx, personal.ID, "someone@test.com", domain.TeamViewer); !errors.Is(err, ErrPersonalTeam) {
		t.Fatalf("inviting into a personal team = %v, want ErrPersonalTeam", err)
	}
}

func TestReach_TransferRefusesAnUnknownTeam(t *testing.T) {
	k := newRoleKit(t)
	research, _, err := k.research.Create(t.Context(), CreateResearchRequest{Name: "Local", Goal: "G"})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	// No caller, so the role checks are skipped — the target still has to
	// exist, or a typo comes back as a database error.
	if err := k.team.TransferResearch(t.Context(), research.ID, "no-such-team"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("transfer to a nonexistent team = %v, want ErrNotFound", err)
	}
}


// The reference a colleague writes has to work for whoever may follow it, not
// only for whoever wrote it. That is the whole point of a shared team, and it
// is the case the old author-time check could not express.
func TestReach_CrossTeamReferenceResolvesForAnEntitledReader(t *testing.T) {
	k := newRoleKit(t)
	owner, member, shared, section, teamID := k.sharedResearch(t, domain.TeamEditor)

	target, err := k.entry.Create(owner, CreateEntryRequest{
		ResearchID: shared.ID, SectionID: section.ID,
		Title: "The target", Content: "body",
	})
	if err != nil {
		t.Fatalf("seed target: %v", err)
	}

	// A second team, holding a research that points at the first one. The
	// member belongs to both; the owner belongs only to the first.
	second, err := k.team.Create(member, "Second team")
	if err != nil {
		t.Fatalf("create team: %v", err)
	}
	_ = teamID
	own, sections, err := k.research.Create(member, CreateResearchRequest{
		TeamID: second.ID, Name: "Points across teams", Goal: "G",
		Sections: []CreateSectionRequest{{Name: "s1", DisplayName: "S1"}},
	})
	if err != nil {
		t.Fatalf("create research: %v", err)
	}
	pointer, err := k.entry.Create(member, CreateEntryRequest{
		ResearchID: own.ID, SectionID: sections[0].ID,
		Title: "Pointer", Content: "See [[" + shared.Code + ":" + target.Code + "]].",
	})
	if err != nil {
		t.Fatalf("create entry: %v", err)
	}

	refs, err := storage.NewCrossRefRepository(k.db).FindBySourceEntry(member, pointer.ID)
	if err != nil {
		t.Fatalf("crossrefs: %v", err)
	}

	var resolvedForMember bool
	for _, r := range testAccess(k.db).VisibleCrossRefs(member, refs) {
		if r.TargetEntryID == target.ID && r.Resolved {
			resolvedForMember = true
		}
	}
	if !resolvedForMember {
		t.Error("someone in both teams should be able to follow their own reference")
	}

	// A stranger to both teams gets the same row with everything stripped.
	stranger := userCtx(createTestUser(t, k.db, "stranger@test.com", "Stranger"))
	for _, r := range testAccess(k.db).VisibleCrossRefs(stranger, refs) {
		if r.Resolved || r.TargetEntryID != "" || r.TargetResearchID != "" {
			t.Errorf("a stranger was handed a target: %+v", r)
		}
	}
}
