package service

import (
	"context"
	"testing"

	"github.com/butschster/mcp-research/internal/auth"
	"github.com/butschster/mcp-research/internal/domain"
)

// The hub decides who may see an event from the fields in it, so an event with
// the wrong scope is a delivery bug that no amount of care in the hub can fix.
// These assert the envelope at the place it is written.

func (m *mockNotifier) find(eventType string) (Event, bool) {
	for _, e := range m.events {
		if e.Type == eventType {
			return e, true
		}
	}
	return Event{}, false
}

func (m *mockNotifier) all(eventType string) []Event {
	var out []Event
	for _, e := range m.events {
		if e.Type == eventType {
			out = append(out, e)
		}
	}
	return out
}

func TestEvent_CarriesTheResearchAndTheActor(t *testing.T) {
	k := newRoleKit(t)
	owner, _, research, section, _ := k.sharedResearch(t, domain.TeamEditor)
	ownerID := auth.UserIDFromContext(owner)

	// The tab making the write, as the HTTP layer records it from X-Client-Id.
	ctx := WithClientID(owner, "tab-7")

	k.events.reset()
	entry, err := k.entry.Create(ctx, CreateEntryRequest{
		ResearchID: research.ID,
		SectionID:  section.ID,
		Content:    "hello",
		Title:      "Hello",
	})
	if err != nil {
		t.Fatalf("create entry: %v", err)
	}

	e, ok := k.events.find("entry.created")
	if !ok {
		t.Fatal("no entry.created event")
	}
	// Without the research id the hub has nothing to check the event against
	// and, correctly, delivers it to nobody.
	if e.ResearchID != research.ID {
		t.Errorf("research_id = %q, want %q", e.ResearchID, research.ID)
	}
	if e.EntityID != entry.ID {
		t.Errorf("entity_id = %q, want the entry %q", e.EntityID, entry.ID)
	}
	if e.ActorUserID != ownerID {
		t.Errorf("actor_user_id = %q, want %q", e.ActorUserID, ownerID)
	}
	if e.ActorClientID != "tab-7" {
		t.Errorf("actor_client_id = %q, want %q — without it a tab cannot tell its own write apart", e.ActorClientID, "tab-7")
	}
}

// Every event that names a research must carry its id. One that does not is not
// merely untidy: it is dropped by the delivery rule and reaches nobody.
func TestEvent_EveryScopedEventNamesItsResearch(t *testing.T) {
	k := newRoleKit(t)
	owner, _, research, section, _ := k.sharedResearch(t, domain.TeamEditor)

	k.events.reset()

	if _, _, err := k.session.Create(owner, CreateSessionRequest{
		ResearchID: research.ID,
		Title:      "Kickoff",
		Questions:  []CreateQuestionRequest{{Text: "Why?"}, {Text: "How?"}},
	}); err != nil {
		t.Fatalf("create session: %v", err)
	}
	if _, err := k.task.Create(owner, CreateTaskRequest{ResearchID: research.ID, Title: "Do it"}); err != nil {
		t.Fatalf("create task: %v", err)
	}
	entry, err := k.entry.Create(owner, CreateEntryRequest{
		ResearchID: research.ID, SectionID: section.ID, Content: "x", Title: "X",
	})
	if err != nil {
		t.Fatalf("create entry: %v", err)
	}
	if _, err := k.entry.RebuildCrossRefs(owner, research.ID); err != nil {
		t.Fatalf("rebuild crossrefs: %v", err)
	}
	_ = entry

	if len(k.events.events) == 0 {
		t.Fatal("nothing was emitted at all")
	}
	for _, e := range k.events.events {
		if e.Entity == "team" || e.TargetUserID != "" {
			continue // scoped to a team, or addressed to one person
		}
		if e.ResearchID == "" {
			t.Errorf("%s (%s) carries no research id, so the hub delivers it to nobody", e.Type, e.Entity)
		}
	}

	// The link tables the graph and the mind map are drawn from were rewritten;
	// nothing else announces that.
	if _, ok := k.events.find("crossrefs.rebuilt"); !ok {
		t.Error("rebuilding cross-references told nobody")
	}
}

// Adding questions used to emit one event carrying the *session* id, which made
// twelve new questions indistinguishable from one and left no way to react to a
// particular one.
func TestEvent_EachNewQuestionNamesItself(t *testing.T) {
	k := newRoleKit(t)
	owner, _, research, _, _ := k.sharedResearch(t, domain.TeamEditor)

	session, _, err := k.session.Create(owner, CreateSessionRequest{ResearchID: research.ID, Title: "S"})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}

	k.events.reset()
	questions, err := k.session.AddQuestions(owner, session.ID, []CreateQuestionRequest{
		{Text: "One"}, {Text: "Two"}, {Text: "Three"},
	})
	if err != nil {
		t.Fatalf("add questions: %v", err)
	}

	events := k.events.all("question.created")
	if len(events) != len(questions) {
		t.Fatalf("got %d question.created events for %d questions", len(events), len(questions))
	}
	named := map[string]bool{}
	for _, e := range events {
		named[e.EntityID] = true
	}
	for _, q := range questions {
		if !named[q.ID] {
			t.Errorf("no event named question %q", q.ID)
		}
	}
}

// Losing access is the one thing the ordinary scope cannot announce: by the time
// it happens, the person it concerns is exactly the person the rule refuses to
// deliver to.
func TestEvent_LosingAccessIsAddressedToThePersonWhoLostIt(t *testing.T) {
	k := newRoleKit(t)
	ownerUser := createTestUser(t, k.db, "rev-owner@test.com", "Owner")
	memberUser := createTestUser(t, k.db, "rev-member@test.com", "Member")
	owner := userCtx(ownerUser)

	team, err := k.team.Create(owner, "Northwind Ops")
	if err != nil {
		t.Fatalf("create team: %v", err)
	}
	if err := k.teamRepo.AddMember(context.Background(), team.ID, memberUser.ID, domain.TeamEditor, ownerUser.ID); err != nil {
		t.Fatalf("add member: %v", err)
	}

	k.events.reset()
	if err := k.team.RemoveMember(owner, team.ID, memberUser.ID); err != nil {
		t.Fatalf("remove member: %v", err)
	}

	e, ok := k.events.find("access.revoked")
	if !ok {
		t.Fatal("the person removed was told nothing; their tab keeps showing what they lost")
	}
	if e.TargetUserID != memberUser.ID {
		t.Errorf("target = %q, want the removed member %q", e.TargetUserID, memberUser.ID)
	}
	// They cannot look either of these up any more — the fetch would 404.
	if e.Name != "Northwind Ops" {
		t.Errorf("name = %q, want the team's name", e.Name)
	}
	if e.Reason != "removed_from_team" {
		t.Errorf("reason = %q", e.Reason)
	}

	// A demotion is not a removal, but it changes what the controls on screen
	// are allowed to do, and the team event does not say whose rights moved.
	k.events.reset()
	if err := k.teamRepo.AddMember(context.Background(), team.ID, memberUser.ID, domain.TeamEditor, ownerUser.ID); err != nil {
		t.Fatalf("re-add member: %v", err)
	}
	if err := k.team.UpdateRole(owner, team.ID, memberUser.ID, domain.TeamViewer); err != nil {
		t.Fatalf("update role: %v", err)
	}
	changed, ok := k.events.find("access.changed")
	if !ok {
		t.Fatal("a demoted member is not told, and keeps Edit and Delete on screen until they 403")
	}
	if changed.TargetUserID != memberUser.ID {
		t.Errorf("target = %q, want %q", changed.TargetUserID, memberUser.ID)
	}
}

// A team event is delivered by asking who is in the team, and deleting a team
// takes its membership rows with it — so team.deleted announced after the fact
// reaches nobody at all. Everyone who was in it is told directly.
func TestEvent_DeletingATeamStillReachesItsMembers(t *testing.T) {
	k := newRoleKit(t)
	ownerUser := createTestUser(t, k.db, "del-owner@test.com", "Owner")
	memberUser := createTestUser(t, k.db, "del-member@test.com", "Member")
	owner := userCtx(ownerUser)

	team, err := k.team.Create(owner, "Doomed")
	if err != nil {
		t.Fatalf("create team: %v", err)
	}
	if err := k.teamRepo.AddMember(context.Background(), team.ID, memberUser.ID, domain.TeamEditor, ownerUser.ID); err != nil {
		t.Fatalf("add member: %v", err)
	}

	k.events.reset()
	if err := k.team.Delete(owner, team.ID); err != nil {
		t.Fatalf("delete team: %v", err)
	}

	told := map[string]Event{}
	for _, e := range k.events.all("access.revoked") {
		told[e.TargetUserID] = e
	}
	for _, u := range []*domain.User{ownerUser, memberUser} {
		e, ok := told[u.ID]
		if !ok {
			t.Errorf("%s was not told the team they were in no longer exists", u.Name)
			continue
		}
		if e.Reason != "team_deleted" || e.Name != "Doomed" {
			t.Errorf("%s got %+v, want the team's name and why", u.Name, e)
		}
	}
}

// Someone in both teams keeps the research and must not be told they lost it.
func TestEvent_TransferOnlyRevokesForThoseActuallyLosingIt(t *testing.T) {
	k := newRoleKit(t)
	ownerUser := createTestUser(t, k.db, "tr-owner@test.com", "Owner")
	stayerUser := createTestUser(t, k.db, "tr-stayer@test.com", "Stayer")
	loserUser := createTestUser(t, k.db, "tr-loser@test.com", "Loser")
	owner := userCtx(ownerUser)

	from, err := k.team.Create(owner, "From")
	if err != nil {
		t.Fatalf("create from-team: %v", err)
	}
	to, err := k.team.Create(owner, "To")
	if err != nil {
		t.Fatalf("create to-team: %v", err)
	}
	ctx := context.Background()
	for _, m := range []struct {
		team string
		user string
	}{
		{from.ID, stayerUser.ID}, {to.ID, stayerUser.ID}, {from.ID, loserUser.ID},
	} {
		if err := k.teamRepo.AddMember(ctx, m.team, m.user, domain.TeamEditor, ownerUser.ID); err != nil {
			t.Fatalf("add member: %v", err)
		}
	}

	research, _, err := k.research.Create(owner, CreateResearchRequest{TeamID: from.ID, Name: "Payment rails"})
	if err != nil {
		t.Fatalf("create research: %v", err)
	}

	k.events.reset()
	if err := k.team.TransferResearch(owner, research.ID, to.ID); err != nil {
		t.Fatalf("transfer: %v", err)
	}

	told := map[string]Event{}
	for _, e := range k.events.all("access.revoked") {
		told[e.TargetUserID] = e
	}
	if _, ok := told[loserUser.ID]; !ok {
		t.Error("the member left behind was not told the research moved out from under them")
	}
	if _, ok := told[stayerUser.ID]; ok {
		t.Error("a member of both teams was told they lost a research they still have")
	}
	if _, ok := told[ownerUser.ID]; ok {
		t.Error("the person who performed the transfer was told they lost access to it")
	}
	if e := told[loserUser.ID]; e.Name != "Payment rails" || e.Reason != "research_transferred" {
		t.Errorf("event = %+v, want the research's name and why", e)
	}
}
