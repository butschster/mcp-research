package service

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"testing"

	"github.com/butschster/mcp-research/internal/domain"
	"github.com/butschster/mcp-research/internal/storage"
)

// The rules worth testing here are the ones that are cheap to break and
// expensive to notice: an upgrade duplicating or overwriting what we ship, the
// cap counting the wrong things, and a swap that leaves a research with fewer
// skills than it started with.

type skillKit struct {
	*roleKit
	skills *SkillService
	repo   *storage.SkillRepository
}

func newSkillKit(t *testing.T) *skillKit {
	t.Helper()
	k := newRoleKit(t)
	repo := storage.NewSkillRepository(k.db)
	return &skillKit{
		roleKit: k,
		skills: NewSkillService(repo, storage.NewResearchRepository(k.db),
			k.teamRepo, testAccess(k.db), k.events, slog.Default()),
		repo: repo,
	}
}

func skillInput(name string) SkillInput {
	return SkillInput{
		Name:        name,
		Description: "Use when the tests need a skill called " + name + ".",
		Body:        "# " + name + "\n\nSome methodology.",
	}
}

// chosen drops the ambient product skills, which every list now carries because
// they are on by definition. A test that means "what did this research choose"
// has to say so.
func chosen(list []*domain.Skill) []*domain.Skill {
	var out []*domain.Skill
	for _, sk := range list {
		if !sk.Ambient {
			out = append(out, sk)
		}
	}
	return out
}

func TestSkill_BuiltinsLoadAndDoNotDuplicateOnSecondBoot(t *testing.T) {
	k := newSkillKit(t)
	ctx := context.Background()

	first, err := k.skills.LoadBuiltinSkills(ctx)
	if err != nil {
		t.Fatalf("first boot: %v", err)
	}
	if first == 0 {
		t.Fatal("no built-in skills were loaded")
	}

	// The second boot is the case a plain UNIQUE(team_id, slug) would not have
	// caught, because SQLite treats NULLs as distinct.
	if _, err := k.skills.LoadBuiltinSkills(ctx); err != nil {
		t.Fatalf("second boot: %v", err)
	}

	var n int
	if err := k.db.QueryRow(`SELECT COUNT(*) FROM skills WHERE team_id IS NULL AND research_id IS NULL`).Scan(&n); err != nil {
		t.Fatalf("count builtins: %v", err)
	}
	if n != first {
		t.Fatalf("second boot changed the built-in count: %d then %d", first, n)
	}
}

func TestSkill_ShippedBuiltinsRespectTheirOwnBudgets(t *testing.T) {
	k := newSkillKit(t)
	if _, err := k.skills.LoadBuiltinSkills(context.Background()); err != nil {
		// LoadBuiltinSkills validates every file it reads, so a shipped skill
		// over the description cap fails the boot rather than silently eating
		// the index budget it exists to protect.
		t.Fatalf("a built-in we ship breaks its own rules: %v", err)
	}

	rows, err := k.db.Query(`SELECT slug, description, ambient FROM skills WHERE team_id IS NULL AND research_id IS NULL`)
	if err != nil {
		t.Fatalf("query builtins: %v", err)
	}
	defer rows.Close()

	var ambient int
	for rows.Next() {
		var slug, description string
		var isAmbient int
		if err := rows.Scan(&slug, &description, &isAmbient); err != nil {
			t.Fatalf("scan: %v", err)
		}
		if len([]rune(description)) > domain.SkillDescriptionMax {
			t.Errorf("%s: description is %d characters", slug, len([]rune(description)))
		}
		// A trigger line names a situation. A description that describes gets a
		// skill the agent never opens, so the shipped ones are held to it.
		if !strings.HasPrefix(strings.ToLower(description), "use when") {
			t.Errorf("%s: description does not start with a trigger: %q", slug, description)
		}
		ambient += isAmbient
	}
	if ambient == 0 {
		t.Error("no product skill is marked ambient; the cap will be full before a user chooses anything")
	}
}

func TestSkill_CapCountsChosenSkillsAndIgnoresAmbientOnes(t *testing.T) {
	k := newSkillKit(t)
	owner, _, research, _, _ := k.sharedResearch(t, domain.TeamEditor)
	if _, err := k.skills.LoadBuiltinSkills(context.Background()); err != nil {
		t.Fatalf("load builtins: %v", err)
	}

	// Every ambient product skill attaches, however many there are.
	lib, err := k.skills.ListLibrary(owner, research.ID, "")
	if err != nil {
		t.Fatalf("library: %v", err)
	}
	var ambientAttached int
	for _, sk := range lib {
		if !sk.Ambient {
			continue
		}
		if _, err := k.skills.Attach(owner, research.ID, sk.Slug, false); err != nil {
			t.Fatalf("attach ambient %s: %v", sk.Slug, err)
		}
		ambientAttached++
	}
	if ambientAttached == 0 {
		t.Fatal("no ambient skills in the library; this test proves nothing")
	}

	// The budget is untouched by them: a full six chosen skills still fit.
	for i := 0; i < domain.SkillsPerResearch; i++ {
		if _, err := k.skills.CreatePrivate(owner, research.ID, skillInput(chosenName(i))); err != nil {
			t.Fatalf("private skill %d refused with %d ambient attached: %v", i, ambientAttached, err)
		}
	}

	_, err = k.skills.CreatePrivate(owner, research.ID, skillInput("one too many"))
	if !errors.Is(err, ErrSkillCapReached) {
		t.Fatalf("seventh chosen skill: want ErrSkillCapReached, got %v", err)
	}
}

func TestSkill_AmbientSkillCannotBeDetached(t *testing.T) {
	k := newSkillKit(t)
	owner, _, research, _, _ := k.sharedResearch(t, domain.TeamEditor)
	if _, err := k.skills.LoadBuiltinSkills(context.Background()); err != nil {
		t.Fatalf("load builtins: %v", err)
	}

	// The usual case: it is in the list without ever having been attached, so
	// the attachment lookup misses it. Answering "not found" about a row the
	// caller is looking at is the wrong refusal.
	if err := k.skills.Detach(owner, research.ID, "writing-entries"); !errors.Is(err, ErrSkillAmbient) {
		t.Fatalf("detaching an always-on skill: want ErrSkillAmbient, got %v", err)
	}

	sk, err := k.skills.Attach(owner, research.ID, "managing-a-research", false)
	if err != nil {
		t.Fatalf("attach: %v", err)
	}
	if !sk.Ambient {
		t.Fatal("managing-a-research should ship as ambient")
	}
	if err := k.skills.Detach(owner, research.ID, sk.Slug); !errors.Is(err, ErrSkillAmbient) {
		t.Fatalf("detach ambient: want ErrSkillAmbient, got %v", err)
	}
}

func TestSkill_EditingABuiltinForksAndLeavesTheOriginalAlone(t *testing.T) {
	k := newSkillKit(t)
	owner, _, research, _, teamID := k.sharedResearch(t, domain.TeamEditor)
	if _, err := k.skills.LoadBuiltinSkills(context.Background()); err != nil {
		t.Fatalf("load builtins: %v", err)
	}
	if _, err := k.skills.Attach(owner, research.ID, "evidence-grading", false); err != nil {
		t.Fatalf("attach: %v", err)
	}

	forked, err := k.skills.Fork(owner, research.ID, "evidence-grading", SkillInput{
		Name:        "Grading evidence, our way",
		Description: "Use when recording a claim that rests on a source in our house style.",
		Body:        "Only measured evidence counts here.",
	})
	if err != nil {
		t.Fatalf("fork: %v", err)
	}
	if forked.Tier != domain.SkillTeam || forked.TeamID != teamID {
		t.Fatalf("fork landed in the wrong place: tier=%s team=%s", forked.Tier, forked.TeamID)
	}
	if forked.ForkedFrom != "evidence-grading" {
		t.Errorf("fork does not record its parent: %q", forked.ForkedFrom)
	}

	// The built-in is untouched — this is the rule that lets an upgrade refresh
	// what we ship without destroying somebody's edits.
	original, err := k.repo.FindBuiltinBySlug(context.Background(), "evidence-grading")
	if err != nil || original == nil {
		t.Fatalf("built-in gone after fork: %v", err)
	}
	if original.Name == forked.Name {
		t.Fatal("the fork overwrote the built-in")
	}

	// The attachment moved, and it moved without passing through a state where
	// the research followed one fewer skill.
	attached := chosen(mustList(t, k, owner, research.ID))
	if len(attached) != 1 || attached[0].ID != forked.ID {
		t.Fatalf("attachment did not follow the fork: %+v", attached)
	}
}

func TestSkill_CopyHereThenPromoteMovesBetweenTiers(t *testing.T) {
	k := newSkillKit(t)
	owner, _, research, _, teamID := k.sharedResearch(t, domain.TeamEditor)
	if _, err := k.skills.LoadBuiltinSkills(context.Background()); err != nil {
		t.Fatalf("load builtins: %v", err)
	}
	if _, err := k.skills.Attach(owner, research.ID, "structured-interviewing", false); err != nil {
		t.Fatalf("attach: %v", err)
	}

	private, err := k.skills.CopyHere(owner, research.ID, "structured-interviewing")
	if err != nil {
		t.Fatalf("copy: %v", err)
	}
	if private.Tier != domain.SkillPrivate || private.ResearchID != research.ID {
		t.Fatalf("copy is not research-private: %+v", private)
	}
	if private.Body == "" {
		t.Error("copy lost the body it was made from")
	}

	promoted, err := k.skills.Promote(owner, research.ID, private.Slug)
	if err != nil {
		t.Fatalf("promote: %v", err)
	}
	if promoted.Tier != domain.SkillTeam || promoted.TeamID != teamID {
		t.Fatalf("promotion landed in the wrong place: %+v", promoted)
	}

	// The private original is gone. Leaving it would shadow the promotion
	// forever, because a slug resolves to the private tier first.
	attached := chosen(mustList(t, k, owner, research.ID))
	if len(attached) != 1 {
		t.Fatalf("want one chosen skill after promotion, got %d", len(attached))
	}
	if attached[0].Tier != domain.SkillTeam {
		t.Fatalf("the research still follows the private copy: %+v", attached[0])
	}
}

func TestSkill_IndexIsOrderedByTierNotByWhenItWasAttached(t *testing.T) {
	k := newSkillKit(t)
	owner, _, research, _, _ := k.sharedResearch(t, domain.TeamEditor)
	if _, err := k.skills.LoadBuiltinSkills(context.Background()); err != nil {
		t.Fatalf("load builtins: %v", err)
	}

	// Built-in first, private second — the opposite of the order they must be
	// reported in. With instruction gone, precedence is expressed by tier and
	// nothing else, so an index in attachment order states the wrong rule.
	if _, err := k.skills.Attach(owner, research.ID, "evidence-grading", false); err != nil {
		t.Fatalf("attach builtin: %v", err)
	}
	if _, err := k.skills.CreatePrivate(owner, research.ID, skillInput("Local rule")); err != nil {
		t.Fatalf("create private: %v", err)
	}

	index := chosen(k.skills.Index(owner, research.ID))
	if len(index) != 2 {
		t.Fatalf("want 2 chosen skills in the index, got %d", len(index))
	}
	if index[0].Tier != domain.SkillPrivate {
		t.Fatalf("index is not tier-ordered: first is %s", index[0].Tier)
	}
	if domain.SkillRank(index[0].Tier) > domain.SkillRank(index[1].Tier) {
		t.Fatal("index rank order is inverted")
	}
}

func TestSkill_ValidationRefusesWhatWouldBreakTheIndex(t *testing.T) {
	k := newSkillKit(t)
	owner, _, research, _, _ := k.sharedResearch(t, domain.TeamEditor)

	long := SkillInput{
		Name:        "Too chatty",
		Description: strings.Repeat("x", domain.SkillDescriptionMax+1),
		Body:        "body",
	}
	if _, err := k.skills.CreatePrivate(owner, research.ID, long); !errors.Is(err, ErrSkillDescriptionLong) {
		t.Fatalf("over-long description: want ErrSkillDescriptionLong, got %v", err)
	}

	// The cap is in runes: a Cyrillic trigger line is not half as long as a
	// Latin one, and counting bytes would refuse a legitimate Russian
	// description at half the stated limit.
	cyrillic := SkillInput{
		Name:        "Кириллица",
		Description: "Использовать, когда " + strings.Repeat("я", domain.SkillDescriptionMax-30),
		Body:        "тело",
	}
	if _, err := k.skills.CreatePrivate(owner, research.ID, cyrillic); err != nil {
		t.Fatalf("a Cyrillic description within the rune cap was refused: %v", err)
	}

	if _, err := k.skills.CreatePrivate(owner, research.ID, SkillInput{Name: "No body", Description: "Use when."}); !errors.Is(err, ErrSkillBodyEmpty) {
		t.Fatalf("empty body: want ErrSkillBodyEmpty, got %v", err)
	}
	if _, err := k.skills.CreatePrivate(owner, research.ID, SkillInput{Name: "  ", Description: "Use when.", Body: "b"}); !errors.Is(err, ErrSkillNameEmpty) {
		t.Fatalf("empty name: want ErrSkillNameEmpty, got %v", err)
	}
}

func TestSkill_ABuiltinIsNotEditableInPlace(t *testing.T) {
	k := newSkillKit(t)
	owner, _, _, _, _ := k.sharedResearch(t, domain.TeamEditor)
	if _, err := k.skills.LoadBuiltinSkills(context.Background()); err != nil {
		t.Fatalf("load builtins: %v", err)
	}
	builtin, err := k.repo.FindBuiltinBySlug(context.Background(), "writing-entries")
	if err != nil || builtin == nil {
		t.Fatalf("builtin missing: %v", err)
	}
	if _, err := k.skills.Update(owner, builtin.ID, skillInput("Rewritten")); !errors.Is(err, ErrSkillBuiltinWrite) {
		t.Fatalf("update builtin: want ErrSkillBuiltinWrite, got %v", err)
	}
	if err := k.skills.Delete(owner, builtin.ID); !errors.Is(err, ErrSkillBuiltinWrite) {
		t.Fatalf("delete builtin: want ErrSkillBuiltinWrite, got %v", err)
	}
}

func chosenName(i int) string {
	return "Chosen skill " + string(rune('A'+i))
}

func TestSkill_ViewerReadsAndCannotWrite(t *testing.T) {
	k := newSkillKit(t)
	owner, viewer, research, _, _ := k.sharedResearch(t, domain.TeamViewer)
	if _, err := k.skills.LoadBuiltinSkills(context.Background()); err != nil {
		t.Fatalf("load builtins: %v", err)
	}
	if _, err := k.skills.Attach(owner, research.ID, "evidence-grading", false); err != nil {
		t.Fatalf("attach: %v", err)
	}

	// A viewer may read the methodology: hiding the agent's configuration from
	// somebody who can already read every finding it produced protects nothing.
	if _, err := k.skills.ListAttached(viewer, research.ID); err != nil {
		t.Errorf("viewer cannot list skills: %v", err)
	}
	if _, err := k.skills.Load(viewer, research.ID, "evidence-grading"); err != nil {
		t.Errorf("viewer cannot read a skill body: %v", err)
	}

	for name, err := range map[string]error{
		"attach":  errOf(k.skills.Attach(viewer, research.ID, "writing-entries", false)),
		"detach":  k.skills.Detach(viewer, research.ID, "evidence-grading"),
		"create":  errOf(k.skills.CreatePrivate(viewer, research.ID, skillInput("Sneaky"))),
		"copy":    errOf(k.skills.CopyHere(viewer, research.ID, "evidence-grading")),
		"promote": errOf(k.skills.Promote(viewer, research.ID, "evidence-grading")),
		"fork":    errOf(k.skills.Fork(viewer, research.ID, "evidence-grading", skillInput("Mine"))),
	} {
		if !errors.Is(err, ErrForbidden) {
			t.Errorf("viewer %s: want ErrForbidden, got %v", name, err)
		}
	}
}

// A research-private skill is exactly where the old instruction text lands, so
// one research reaching another's is the leak this tier would cause.
func TestSkill_PrivateSkillsDoNotCrossResearches(t *testing.T) {
	k := newSkillKit(t)
	owner, _, mine, _, teamID := k.sharedResearch(t, domain.TeamEditor)

	theirs, _, err := k.research.Create(owner, CreateResearchRequest{
		TeamID: teamID, Name: "Second research", Goal: "Other",
	})
	if err != nil {
		t.Fatalf("create second research: %v", err)
	}

	secret, err := k.skills.CreatePrivate(owner, mine.ID, SkillInput{
		Name:        "House rule",
		Description: "Use when writing anything in this particular research.",
		Body:        "Something only this research's people should read.",
	})
	if err != nil {
		t.Fatalf("create private: %v", err)
	}

	// Not in the other research's library...
	lib, err := k.skills.ListLibrary(owner, theirs.ID, "")
	if err != nil {
		t.Fatalf("library: %v", err)
	}
	for _, sk := range lib {
		if sk.Slug == secret.Slug {
			t.Fatal("a research-private skill appeared in another research's library")
		}
	}
	// ...and not loadable or attachable from it either, even by the same person
	// who owns both. Scope here is the research, not the reader.
	if _, err := k.skills.Load(owner, theirs.ID, secret.Slug); !errors.Is(err, ErrNotFound) {
		t.Errorf("cross-research load: want ErrNotFound, got %v", err)
	}
	if _, err := k.skills.Attach(owner, theirs.ID, secret.Slug, false); !errors.Is(err, ErrNotFound) {
		t.Errorf("cross-research attach: want ErrNotFound, got %v", err)
	}
}

// A team library is a team's. Another team's research must not see it even
// though both live on the same server.
func TestSkill_TeamLibrariesDoNotCrossTeams(t *testing.T) {
	k := newSkillKit(t)
	ownerA, _, researchA, _, teamA := k.sharedResearch(t, domain.TeamEditor)

	strangerUser := createTestUser(t, k.db, "stranger@test.com", "Stranger")
	stranger := userCtx(strangerUser)
	teamB, err := k.team.Create(stranger, "Other team")
	if err != nil {
		t.Fatalf("create team: %v", err)
	}
	researchB, _, err := k.research.Create(stranger, CreateResearchRequest{
		TeamID: teamB.ID, Name: "Their research", Goal: "Theirs",
	})
	if err != nil {
		t.Fatalf("create research: %v", err)
	}

	theirs, err := k.skills.CreateTeam(ownerA, teamA, SkillInput{
		Name:        "Our playbook",
		Description: "Use when doing the thing our team does.",
		Body:        "Our hard-won process.",
	})
	if err != nil {
		t.Fatalf("create team skill: %v", err)
	}

	lib, err := k.skills.ListLibrary(stranger, researchB.ID, "")
	if err != nil {
		t.Fatalf("library: %v", err)
	}
	for _, sk := range lib {
		if sk.Slug == theirs.Slug {
			t.Fatal("one team's skill appeared in another team's library")
		}
	}
	if _, err := k.skills.Attach(stranger, researchB.ID, theirs.Slug, false); !errors.Is(err, ErrNotFound) {
		t.Errorf("cross-team attach: want ErrNotFound, got %v", err)
	}
	// And the stranger cannot reach team A's research at all — the pre-existing
	// rule, restated here so a skills route cannot become a way around it.
	if _, err := k.skills.ListAttached(stranger, researchA.ID); !errors.Is(err, ErrNotFound) {
		t.Errorf("cross-team list: want ErrNotFound, got %v", err)
	}
}

// errOf drops a value so a table can hold calls with different return shapes.
func errOf(_ *domain.Skill, err error) error { return err }

// A team fork keeps its built-in's slug, so a second research in the same team
// that follows the *built-in* can be handed the fork by a slug lookup that only
// consults scope. The agent would then read a body that is not the one its index
// named — silently, and only for the research that never asked for the fork.
func TestSkill_AResearchReadsTheSkillItActuallyFollows(t *testing.T) {
	k := newSkillKit(t)
	owner, _, forking, _, teamID := k.sharedResearch(t, domain.TeamEditor)
	if _, err := k.skills.LoadBuiltinSkills(context.Background()); err != nil {
		t.Fatalf("load builtins: %v", err)
	}

	bystander, _, err := k.research.Create(owner, CreateResearchRequest{
		TeamID: teamID, Name: "Bystander", Goal: "Follows the built-in",
	})
	if err != nil {
		t.Fatalf("create bystander: %v", err)
	}
	if _, err := k.skills.Attach(owner, bystander.ID, "evidence-grading", false); err != nil {
		t.Fatalf("attach builtin to bystander: %v", err)
	}

	// The other research forks it, which creates a team skill with the same slug.
	if _, err := k.skills.Attach(owner, forking.ID, "evidence-grading", false); err != nil {
		t.Fatalf("attach: %v", err)
	}
	if _, err := k.skills.Fork(owner, forking.ID, "evidence-grading", SkillInput{
		Name:        "Our grading",
		Description: "Use when grading a source the way this team decided to.",
		Body:        "TEAM VERSION",
	}); err != nil {
		t.Fatalf("fork: %v", err)
	}

	// The bystander's index still says built-in...
	index := chosen(k.skills.Index(owner, bystander.ID))
	if len(index) != 1 || index[0].Tier != domain.SkillBuiltin {
		t.Fatalf("bystander index changed under it: %+v", index)
	}
	// ...so loading that slug must return the built-in's body, not the fork's.
	loaded, err := k.skills.Load(owner, bystander.ID, "evidence-grading")
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if loaded.Tier != domain.SkillBuiltin {
		t.Fatalf("bystander was handed the %s version of a skill it does not follow", loaded.Tier)
	}
	if loaded.Body == "TEAM VERSION" {
		t.Fatal("bystander read the fork's body")
	}

	// And the research that did fork gets its own version.
	forked, err := k.skills.Load(owner, forking.ID, "evidence-grading")
	if err != nil {
		t.Fatalf("load fork: %v", err)
	}
	if forked.Body != "TEAM VERSION" {
		t.Fatalf("the forking research did not get its fork: tier=%s", forked.Tier)
	}
}

// Moving a research to another team must not carry the first team's methodology
// with it. Attachments are validated when they are made, and a transfer changes
// the owning team underneath live ones — so the queries that read an attachment
// intersect with scope rather than trusting the row.
func TestSkill_ATransferDoesNotCarryTheOldTeamsSkills(t *testing.T) {
	k := newSkillKit(t)
	owner, _, research, _, teamA := k.sharedResearch(t, domain.TeamEditor)

	secret, err := k.skills.CreateTeam(owner, teamA, SkillInput{
		Name:        "Team A playbook",
		Description: "Use when doing the thing team A does.",
		Body:        "TEAM A PROPRIETARY METHODOLOGY",
	})
	if err != nil {
		t.Fatalf("create team skill: %v", err)
	}
	if _, err := k.skills.Attach(owner, research.ID, secret.Slug, false); err != nil {
		t.Fatalf("attach: %v", err)
	}

	teamB, err := k.team.Create(owner, "Team B")
	if err != nil {
		t.Fatalf("create team B: %v", err)
	}
	if _, err := k.db.Exec(`UPDATE researches SET team_id=? WHERE id=?`, teamB.ID, research.ID); err != nil {
		t.Fatalf("transfer: %v", err)
	}

	attached, err := k.skills.ListAttached(owner, research.ID)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	for _, sk := range attached {
		if sk.Slug == secret.Slug {
			t.Fatal("team A's skill followed the research into team B")
		}
	}
	if _, err := k.skills.Load(owner, research.ID, secret.Slug); !errors.Is(err, ErrNotFound) {
		t.Errorf("team A's body was readable from team B: %v", err)
	}
	// And it stops spending the budget it can no longer be seen in.
	n, err := k.repo.CountChosen(context.Background(), research.ID)
	if err != nil {
		t.Fatalf("count: %v", err)
	}
	if n != 0 {
		t.Errorf("an unreachable skill still counts against the cap: %d", n)
	}
}

// Forking or copying a product skill must not quietly convert it into a chosen
// one: that bought a free slot in the budget and made "how to write an entry"
// switch-off-able.
func TestSkill_AForkOfAProductSkillIsStillAProductSkill(t *testing.T) {
	k := newSkillKit(t)
	owner, _, research, _, _ := k.sharedResearch(t, domain.TeamEditor)
	if _, err := k.skills.LoadBuiltinSkills(context.Background()); err != nil {
		t.Fatalf("load builtins: %v", err)
	}
	if _, err := k.skills.Attach(owner, research.ID, "writing-entries", false); err != nil {
		t.Fatalf("attach: %v", err)
	}
	for i := 0; i < domain.SkillsPerResearch; i++ {
		if _, err := k.skills.CreatePrivate(owner, research.ID, skillInput(chosenName(i))); err != nil {
			t.Fatalf("fill budget: %v", err)
		}
	}

	forked, err := k.skills.Fork(owner, research.ID, "writing-entries", SkillInput{
		Name:        "How we write entries",
		Description: "Use when writing an entry in this team's house style.",
		Body:        "Ours.",
	})
	if err != nil {
		t.Fatalf("fork at the cap should be fine — it replaces, not adds: %v", err)
	}
	if !forked.Ambient {
		t.Fatal("the fork lost its ambient flag and became a chosen skill")
	}
	n, err := k.repo.CountChosen(context.Background(), research.ID)
	if err != nil {
		t.Fatalf("count: %v", err)
	}
	if n != domain.SkillsPerResearch {
		t.Errorf("forking a product skill changed the chosen count to %d", n)
	}
	if err := k.skills.Detach(owner, research.ID, forked.Slug); !errors.Is(err, ErrSkillAmbient) {
		t.Errorf("the forked product skill became detachable: %v", err)
	}
}

// A research-private skill exists only for its attachment, so detaching it must
// remove it. Leaving the row made it unreachable by every route while still
// holding its slug, so writing it again failed with slug_taken against
// something the user could not see.
func TestSkill_DetachingAPrivateSkillRemovesIt(t *testing.T) {
	k := newSkillKit(t)
	owner, _, research, _, _ := k.sharedResearch(t, domain.TeamEditor)

	sk, err := k.skills.CreatePrivate(owner, research.ID, skillInput("House rule"))
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if err := k.skills.Detach(owner, research.ID, sk.Slug); err != nil {
		t.Fatalf("detach: %v", err)
	}
	gone, err := k.repo.FindByID(context.Background(), sk.ID)
	if err != nil {
		t.Fatalf("find: %v", err)
	}
	if gone != nil {
		t.Fatal("the private skill row outlived its only attachment")
	}
	// The slug is free again, which is the user-visible half of the bug.
	if _, err := k.skills.CreatePrivate(owner, research.ID, skillInput("House rule")); err != nil {
		t.Fatalf("rewriting the skill after detaching it: %v", err)
	}
}

// A team with no rows is not a team. Without the existence check the insert
// reached the foreign key and answered 500 where every other team-scoped route
// answers 404.
func TestSkill_WritingToATeamThatDoesNotExistIsNotFound(t *testing.T) {
	k := newSkillKit(t)
	if _, err := k.skills.CreateTeam(context.Background(), "no-such-team", skillInput("Ghost")); !errors.Is(err, ErrNotFound) {
		t.Fatalf("unknown team: want ErrNotFound, got %v", err)
	}
}

func mustList(t *testing.T, k *skillKit, ctx context.Context, researchID string) []*domain.Skill {
	t.Helper()
	list, err := k.skills.ListAttached(ctx, researchID)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	return list
}

// The four product skills are on without anybody attaching them. Requiring an
// attach made "always on" false on every research nobody had curated: the index
// came back empty and the agent never learned skills existed.
func TestSkill_ProductSkillsAreOnWithoutBeingAttached(t *testing.T) {
	k := newSkillKit(t)
	owner, _, research, _, _ := k.sharedResearch(t, domain.TeamEditor)
	if _, err := k.skills.LoadBuiltinSkills(context.Background()); err != nil {
		t.Fatalf("load builtins: %v", err)
	}

	index := k.skills.Index(owner, research.ID)
	if len(index) == 0 {
		t.Fatal("a fresh research has an empty index, so the agent is never told skills exist")
	}
	for _, sk := range index {
		if !sk.Ambient {
			t.Errorf("%s is in the index of a research that chose nothing", sk.Slug)
		}
	}
	// And they cost nothing against the budget.
	n, err := k.repo.CountChosen(context.Background(), research.ID)
	if err != nil {
		t.Fatalf("count: %v", err)
	}
	if n != 0 {
		t.Errorf("product skills spent %d of the budget", n)
	}
	// The body is readable without an attachment, which is what makes the index
	// entry worth anything.
	if _, err := k.skills.Load(owner, research.ID, "writing-entries"); err != nil {
		t.Errorf("an always-on skill could not be loaded: %v", err)
	}
}

// A description is the only thing the agent decides from, so a skill saved
// without one can never load. It is required, not merely recommended.
func TestSkill_ADescriptionIsRequired(t *testing.T) {
	k := newSkillKit(t)
	owner, _, research, _, _ := k.sharedResearch(t, domain.TeamEditor)
	_, err := k.skills.CreatePrivate(owner, research.ID, SkillInput{Name: "Silent", Body: "text"})
	if !errors.Is(err, ErrSkillDescriptionEmpty) {
		t.Fatalf("want ErrSkillDescriptionEmpty, got %v", err)
	}
}

// A fork sent with only a body used to land with an empty description — an
// unloadable skill, produced by the operation whose purpose is to keep the
// parent's meaning.
func TestSkill_AForkInheritsWhatTheCallerDidNotRestate(t *testing.T) {
	k := newSkillKit(t)
	owner, _, research, _, _ := k.sharedResearch(t, domain.TeamEditor)
	if _, err := k.skills.LoadBuiltinSkills(context.Background()); err != nil {
		t.Fatalf("load builtins: %v", err)
	}
	parent, err := k.skills.Load(owner, research.ID, "evidence-grading")
	if err != nil {
		t.Fatalf("read parent: %v", err)
	}

	forked, err := k.skills.Fork(owner, research.ID, "evidence-grading", SkillInput{Body: "Only measured evidence."})
	if err != nil {
		t.Fatalf("fork: %v", err)
	}
	if forked.Description != parent.Description {
		t.Errorf("the fork lost its trigger line: %q", forked.Description)
	}
	if forked.Name != parent.Name {
		t.Errorf("the fork lost its name: %q", forked.Name)
	}
}

// A built-in the team has already forked must not be offered again: attaching
// resolves by slug and lands on the fork, so the row would carry a live Attach
// button that always answers already_attached.
func TestSkill_AForkedBuiltinIsNotOfferedTwice(t *testing.T) {
	k := newSkillKit(t)
	owner, _, research, _, _ := k.sharedResearch(t, domain.TeamEditor)
	if _, err := k.skills.LoadBuiltinSkills(context.Background()); err != nil {
		t.Fatalf("load builtins: %v", err)
	}
	if _, err := k.skills.Attach(owner, research.ID, "evidence-grading", false); err != nil {
		t.Fatalf("attach: %v", err)
	}
	if _, err := k.skills.Fork(owner, research.ID, "evidence-grading", SkillInput{Body: "Ours."}); err != nil {
		t.Fatalf("fork: %v", err)
	}

	lib, err := k.skills.ListLibrary(owner, research.ID, "")
	if err != nil {
		t.Fatalf("library: %v", err)
	}
	var seen int
	for _, sk := range lib {
		if sk.Slug == "evidence-grading" {
			seen++
			if sk.Tier != domain.SkillTeam {
				t.Errorf("the shadowed built-in is still listed: %s", sk.Tier)
			}
		}
	}
	if seen != 1 {
		t.Fatalf("want one row for a forked slug, got %d", seen)
	}
}
