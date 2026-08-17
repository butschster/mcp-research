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

// A template produces no rows, so most of what could go wrong here is about
// visibility and about the boot: an upgrade duplicating or overwriting what we
// ship, a team's fork failing to shadow its parent, and a methodology naming a
// skill that does not exist.

type templateKit struct {
	*skillKit
	templates *TemplateService
	repo      *storage.TemplateRepository
}

func newTemplateKit(t *testing.T) *templateKit {
	t.Helper()
	k := newSkillKit(t)
	repo := storage.NewTemplateRepository(k.db)
	tpl := NewTemplateService(repo, k.repo, k.teamRepo, testAccess(k.db), slog.Default())
	// Wired exactly as main.go does it: attaching goes through the service that
	// owns the cap, not the repository behind it.
	tpl.SetSkillService(k.skills)
	return &templateKit{skillKit: k, templates: tpl, repo: repo}
}

// loaded brings up both halves in the order the binary does: templates name
// skills, so skills have to exist first.
func (k *templateKit) loaded(t *testing.T) {
	t.Helper()
	if _, err := k.skills.LoadBuiltinSkills(context.Background()); err != nil {
		t.Fatalf("load builtin skills: %v", err)
	}
	if _, err := k.templates.LoadBuiltinTemplates(context.Background()); err != nil {
		t.Fatalf("load builtin templates: %v", err)
	}
}

func templateInput(name string) TemplateInput {
	return TemplateInput{
		Name:      name,
		WhenToUse: "Use when the tests need a template called " + name + ".",
		Body:      "## Before you propose anything\n\nAsk something.",
	}
}

func TestTemplate_BuiltinsLoadAndDoNotDuplicateOnSecondBoot(t *testing.T) {
	k := newTemplateKit(t)
	ctx := context.Background()
	if _, err := k.skills.LoadBuiltinSkills(ctx); err != nil {
		t.Fatalf("skills: %v", err)
	}

	first, err := k.templates.LoadBuiltinTemplates(ctx)
	if err != nil {
		t.Fatalf("first boot: %v", err)
	}
	if first == 0 {
		t.Fatal("no built-in templates were loaded")
	}
	if _, err := k.templates.LoadBuiltinTemplates(ctx); err != nil {
		t.Fatalf("second boot: %v", err)
	}

	var n int
	if err := k.db.QueryRow(`SELECT COUNT(*) FROM research_templates WHERE team_id IS NULL`).Scan(&n); err != nil {
		t.Fatalf("count: %v", err)
	}
	if n != first {
		t.Fatalf("second boot changed the count: %d then %d", first, n)
	}
}

// The loader refuses a template naming a skill that is not there. A broken
// methodology found at startup costs a restart; found at a kickoff it costs
// somebody's research.
func TestTemplate_ABuiltinNamingAnUnknownSkillFailsTheBoot(t *testing.T) {
	k := newTemplateKit(t)
	// Deliberately without loading the skills first, which is what the wrong
	// boot order would look like.
	n, err := k.templates.LoadBuiltinTemplates(context.Background())
	if !errors.Is(err, ErrTemplateBuiltinLoad) {
		t.Fatalf("want ErrTemplateBuiltinLoad, got %v", err)
	}
	if !strings.Contains(err.Error(), "does not exist") {
		t.Errorf("the report does not name the problem: %v", err)
	}
	// The bad ones remove themselves and nothing else. Returning on the first
	// one left a server running with no methodologies at all.
	if n != 0 {
		t.Logf("loaded %d despite the failure, which is the point", n)
	}
}

// One unloadable file must not take the others with it.
func TestTemplate_ABadFileOnlyRemovesItself(t *testing.T) {
	k := newTemplateKit(t)
	ctx := context.Background()
	if _, err := k.skills.LoadBuiltinSkills(ctx); err != nil {
		t.Fatalf("skills: %v", err)
	}
	if _, err := k.templates.LoadBuiltinTemplates(ctx); err != nil {
		t.Fatalf("clean load: %v", err)
	}
	// Break one the way a typo would, then reload: the others must survive.
	if _, err := k.db.Exec(`DELETE FROM skills WHERE slug='structured-interviewing'`); err != nil {
		t.Fatalf("break: %v", err)
	}
	if _, err := k.db.Exec(`UPDATE research_templates SET body='changed' WHERE team_id IS NULL`); err != nil {
		t.Fatalf("dirty: %v", err)
	}
	n, err := k.templates.LoadBuiltinTemplates(ctx)
	if err == nil {
		t.Fatal("a template naming a missing skill loaded without complaint")
	}
	if n == 0 {
		t.Fatal("one bad file removed every methodology")
	}
}

// An unchanged boot must not bump the version, because a research stamps the
// version it followed and a bumped one reads as "the methodology was rewritten".
func TestTemplate_AnUnchangedBootDoesNotBumpTheVersion(t *testing.T) {
	k := newTemplateKit(t)
	k.loaded(t)

	before, err := k.repo.FindGlobalBySlug(context.Background(), "literature-review")
	if err != nil || before == nil {
		t.Fatalf("missing: %v", err)
	}
	for i := 0; i < 3; i++ {
		if _, err := k.templates.LoadBuiltinTemplates(context.Background()); err != nil {
			t.Fatalf("reboot %d: %v", i, err)
		}
	}
	after, err := k.repo.FindGlobalBySlug(context.Background(), "literature-review")
	if err != nil || after == nil {
		t.Fatalf("missing after reboot: %v", err)
	}
	if after.Version != before.Version {
		t.Fatalf("three unchanged boots took the version from %d to %d", before.Version, after.Version)
	}
}

func TestTemplate_ShippedTemplatesRespectTheirOwnRules(t *testing.T) {
	k := newTemplateKit(t)
	k.loaded(t)

	rows, err := k.db.Query(`SELECT slug, when_to_use, when_not_to_use, body FROM research_templates WHERE team_id IS NULL`)
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	defer rows.Close()

	var seen int
	for rows.Next() {
		var slug, when, whenNot, body string
		if err := rows.Scan(&slug, &when, &whenNot, &body); err != nil {
			t.Fatalf("scan: %v", err)
		}
		seen++
		if len([]rune(when)) > domain.TemplateCriterionMax {
			t.Errorf("%s: when_to_use is %d characters", slug, len([]rune(when)))
		}
		if !strings.HasPrefix(strings.ToLower(when), "use when") {
			t.Errorf("%s: when_to_use is not a matching line: %q", slug, when)
		}
		// The negative form is what stops a methodology being applied to
		// everything, so a built-in without one is incomplete.
		if strings.TrimSpace(whenNot) == "" {
			t.Errorf("%s: no when_not_to_use", slug)
		}
		// The rule the whole feature turns on has to be in the text, because
		// the text is the only place it can be.
		if !strings.Contains(body, "Before you propose anything") {
			t.Errorf("%s: body does not tell the agent to ask before proposing", slug)
		}
	}
	if seen < 4 {
		t.Fatalf("want at least 4 shipped templates, got %d", seen)
	}
}

func TestTemplate_EditingAGlobalForksAndShadowsIt(t *testing.T) {
	k := newTemplateKit(t)
	owner, _, _, _, teamID := k.sharedResearch(t, domain.TeamEditor)
	k.loaded(t)

	forked, err := k.templates.Fork(owner, teamID, "literature-review", TemplateInput{
		Name: "Literature review, our way",
		Body: "## Before you propose anything\n\nOur rules.",
	})
	if err != nil {
		t.Fatalf("fork: %v", err)
	}
	if forked.Tier != domain.TemplateTeam || forked.TeamID != teamID {
		t.Fatalf("fork landed in the wrong place: %+v", forked)
	}
	// Omitted fields are inherited: an edit restating only the body must not
	// silently drop the line an agent matches on.
	if forked.WhenToUse == "" {
		t.Error("the fork lost its matching line")
	}

	original, err := k.repo.FindGlobalBySlug(context.Background(), "literature-review")
	if err != nil || original == nil {
		t.Fatalf("global gone after fork: %v", err)
	}
	if original.Name == forked.Name {
		t.Fatal("the fork overwrote the global")
	}

	// One row per slug in what the team sees, and it is theirs.
	list, err := k.templates.List(owner)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	var matches []*domain.Template
	for _, tp := range list {
		if tp.Slug == "literature-review" {
			matches = append(matches, tp)
		}
	}
	if len(matches) != 1 {
		t.Fatalf("want one row for a forked slug, got %d", len(matches))
	}
	if matches[0].Tier != domain.TemplateTeam {
		t.Fatalf("the global still shadows the fork: %+v", matches[0])
	}
	// And resolving the slug gets the team's text, not ours.
	got, err := k.templates.Get(owner, "literature-review")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if !strings.Contains(got.Body, "Our rules") {
		t.Error("the slug still resolves to the global body")
	}
}

func TestTemplate_AGlobalIsNotEditableInPlace(t *testing.T) {
	k := newTemplateKit(t)
	owner, _, _, _, _ := k.sharedResearch(t, domain.TeamEditor)
	k.loaded(t)

	global, err := k.repo.FindGlobalBySlug(context.Background(), "technology-comparison")
	if err != nil || global == nil {
		t.Fatalf("missing: %v", err)
	}
	if _, err := k.templates.Update(owner, global.ID, templateInput("Rewritten")); !errors.Is(err, ErrTemplateGlobalWrite) {
		t.Fatalf("update: want ErrTemplateGlobalWrite, got %v", err)
	}
	if err := k.templates.Delete(owner, global.ID); !errors.Is(err, ErrTemplateGlobalWrite) {
		t.Fatalf("delete: want ErrTemplateGlobalWrite, got %v", err)
	}
}

func TestTemplate_TeamLibrariesDoNotCrossTeams(t *testing.T) {
	k := newTemplateKit(t)
	ownerA, _, _, _, teamA := k.sharedResearch(t, domain.TeamEditor)
	k.loaded(t)

	stranger := userCtx(createTestUser(t, k.db, "stranger-tpl@test.com", "Stranger"))
	if _, err := k.team.Create(stranger, "Their team"); err != nil {
		t.Fatalf("create team: %v", err)
	}

	theirs, err := k.templates.CreateTeam(ownerA, teamA, TemplateInput{
		Name:      "Our playbook",
		WhenToUse: "Use when doing the thing our team does.",
		Body:      "## Before you propose anything\n\nOurs.",
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	list, err := k.templates.List(stranger)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	for _, tp := range list {
		if tp.Slug == theirs.Slug {
			t.Fatal("one team's template appeared in another's list")
		}
	}
	if _, err := k.templates.Get(stranger, theirs.Slug); !errors.Is(err, ErrNotFound) {
		t.Errorf("cross-team get: want ErrNotFound, got %v", err)
	}
	if _, err := k.templates.Update(stranger, theirs.ID, templateInput("Mine now")); !errors.Is(err, ErrNotFound) {
		t.Errorf("cross-team update: want ErrNotFound, got %v", err)
	}
	if err := k.templates.Delete(stranger, theirs.ID); !errors.Is(err, ErrNotFound) {
		t.Errorf("cross-team delete: want ErrNotFound, got %v", err)
	}
	// The global set is still everybody's.
	var globals int
	for _, tp := range list {
		if tp.Tier == domain.TemplateGlobal {
			globals++
		}
	}
	if globals < 4 {
		t.Errorf("a stranger sees only %d shipped templates", globals)
	}
}

func TestTemplate_ViewerReadsAndCannotWrite(t *testing.T) {
	k := newTemplateKit(t)
	owner, viewer, _, _, teamID := k.sharedResearch(t, domain.TeamViewer)
	k.loaded(t)

	if _, err := k.templates.List(viewer); err != nil {
		t.Errorf("viewer cannot list templates: %v", err)
	}
	if _, err := k.templates.Get(viewer, "technology-comparison"); err != nil {
		t.Errorf("viewer cannot read a template: %v", err)
	}
	if _, err := k.templates.CreateTeam(viewer, teamID, templateInput("Sneaky")); !errors.Is(err, ErrForbidden) {
		t.Errorf("viewer create: want ErrForbidden, got %v", err)
	}
	if _, err := k.templates.Fork(viewer, teamID, "literature-review", templateInput("Mine")); !errors.Is(err, ErrForbidden) {
		t.Errorf("viewer fork: want ErrForbidden, got %v", err)
	}
	mine, err := k.templates.CreateTeam(owner, teamID, templateInput("Owner's"))
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if _, err := k.templates.Update(viewer, mine.ID, templateInput("Changed")); !errors.Is(err, ErrForbidden) {
		t.Errorf("viewer update: want ErrForbidden, got %v", err)
	}
}

// Creating from a template does the one structural thing in the feature.
func TestTemplate_CreatingFromOneStampsAndAttachesItsSkills(t *testing.T) {
	k := newTemplateKit(t)
	owner, _, _, _, teamID := k.sharedResearch(t, domain.TeamEditor)
	k.loaded(t)

	research, _, err := k.research.Create(owner, CreateResearchRequest{
		TeamID: teamID, Name: "From a methodology", Goal: "Prove the stamp",
	})
	if err != nil {
		t.Fatalf("create research: %v", err)
	}

	res := k.templates.AttachSkills(owner, research.ID, "user-interview-study")
	if !res.Resolved {
		t.Fatal("a shipped template did not resolve")
	}
	if len(res.Missed) != 0 {
		t.Fatalf("a shipped template names skills that are not there: %v", res.Missed)
	}
	if len(res.Attached) == 0 {
		t.Fatal("the template attached nothing")
	}

	var slug string
	var version int
	if err := k.db.QueryRow(`SELECT template_slug, template_version FROM researches WHERE id=?`, research.ID).
		Scan(&slug, &version); err != nil {
		t.Fatalf("read stamp: %v", err)
	}
	if slug != "user-interview-study" {
		t.Errorf("template_slug not recorded: %q", slug)
	}
	// The version matters because built-ins refresh on every boot: without it
	// an upgrade silently changes the methodology under a running research.
	if version == 0 {
		t.Error("template_version not recorded")
	}

	// The skills arrived marked as the template's doing, not somebody's choice.
	list, err := k.skills.ListAttached(owner, research.ID)
	if err != nil {
		t.Fatalf("list skills: %v", err)
	}
	var viaTemplate int
	for _, sk := range list {
		if sk.ViaTemplate {
			viaTemplate++
		}
	}
	if viaTemplate == 0 {
		t.Error("skills attached by a template are not marked as such")
	}
}

// A missing skill is reported, not fatal: a research that exists with one skill
// short beats a create call rolled back because a later build dropped one.
func TestTemplate_AnUnavailableSkillIsReportedRatherThanFatal(t *testing.T) {
	k := newTemplateKit(t)
	owner, _, _, _, teamID := k.sharedResearch(t, domain.TeamEditor)
	k.loaded(t)

	tp, err := k.templates.CreateTeam(owner, teamID, TemplateInput{
		Name:      "Names a ghost",
		WhenToUse: "Use when testing what happens to a skill that is not there.",
		Body:      "## Before you propose anything\n\nAsk.",
		Skills:    []string{"evidence-grading", "no-such-skill"},
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	research, _, err := k.research.Create(owner, CreateResearchRequest{
		TeamID: teamID, Name: "Ghost", Goal: "g",
	})
	if err != nil {
		t.Fatalf("create research: %v", err)
	}

	res := k.templates.AttachSkills(owner, research.ID, tp.Slug)
	if len(res.Attached) != 1 || res.Attached[0] != "evidence-grading" {
		t.Errorf("the available skill did not attach: %v", res.Attached)
	}
	if len(res.Missed) != 1 || res.Missed[0] != "no-such-skill" {
		t.Errorf("the missing skill was not reported: %v", res.Missed)
	}
}

// A slug nobody can resolve — usually a name where a slug was wanted — used to
// produce a result identical to a successful attach, with the research left
// unstamped and nothing saying so.
func TestTemplate_AnUnresolvableSlugIsReportedNotSwallowed(t *testing.T) {
	k := newTemplateKit(t)
	owner, _, research, _, _ := k.sharedResearch(t, domain.TeamEditor)
	k.loaded(t)

	res := k.templates.AttachSkills(owner, research.ID, "Literature review")
	if res.Resolved {
		t.Fatal("a template name resolved as if it were a slug")
	}
	var slug string
	if err := k.db.QueryRow(`SELECT template_slug FROM researches WHERE id=?`, research.ID).Scan(&slug); err != nil {
		t.Fatalf("read stamp: %v", err)
	}
	if slug != "" {
		t.Errorf("an unresolved template still stamped the research: %q", slug)
	}
}

// A methodology naming more skills than a research may follow does not get to
// overrun the budget: attaching goes through the service that owns the cap.
func TestTemplate_ATemplateCannotOverrunTheSkillCap(t *testing.T) {
	k := newTemplateKit(t)
	owner, _, research, _, teamID := k.sharedResearch(t, domain.TeamEditor)
	k.loaded(t)

	var slugs []string
	for i := 0; i < domain.SkillsPerResearch+3; i++ {
		sk, err := k.skills.CreateTeam(owner, teamID, skillInput(chosenName(i)))
		if err != nil {
			t.Fatalf("create team skill %d: %v", i, err)
		}
		slugs = append(slugs, sk.Slug)
	}
	tp, err := k.templates.CreateTeam(owner, teamID, TemplateInput{
		Name:      "Greedy",
		WhenToUse: "Use when testing what a template may not do.",
		Body:      "## Before you propose anything\n\nAsk.",
		Skills:    slugs,
	})
	if err != nil {
		t.Fatalf("create template: %v", err)
	}

	res := k.templates.AttachSkills(owner, research.ID, tp.Slug)
	if len(res.Attached) > domain.SkillsPerResearch {
		t.Fatalf("a template attached %d skills over a cap of %d", len(res.Attached), domain.SkillsPerResearch)
	}
	if len(res.Missed) == 0 {
		t.Error("the skills that did not fit were not reported")
	}
	n, err := k.skillKit.repo.CountChosen(context.Background(), research.ID)
	if err != nil {
		t.Fatalf("count: %v", err)
	}
	if n > domain.SkillsPerResearch {
		t.Fatalf("the research is over budget at %d of %d", n, domain.SkillsPerResearch)
	}
}

// "Save as template" cannot mean what it used to. There are no rows to capture,
// so what comes back is a skeleton with the judgement left blank — and it says
// so rather than looking like a finished methodology.
func TestTemplate_DraftFromResearchIsASkeletonNotACapture(t *testing.T) {
	k := newTemplateKit(t)
	owner, _, research, section, _ := k.sharedResearch(t, domain.TeamEditor)
	k.loaded(t)
	if _, err := k.skills.Attach(owner, research.ID, "evidence-grading", false); err != nil {
		t.Fatalf("attach: %v", err)
	}
	skills, err := k.skills.ListAttached(owner, research.ID)
	if err != nil {
		t.Fatalf("skills: %v", err)
	}

	draft := k.templates.DraftFromResearch(owner, research, []*domain.Section{section}, skills)

	if !strings.Contains(draft.Body, section.DisplayName) {
		t.Error("the draft did not carry the sections the research grew")
	}
	if !strings.Contains(draft.Body, "Before you propose anything") {
		t.Error("the draft has no place for the questions, which is the part that matters")
	}
	// Blanks, not invented prose: a skeleton that reads as finished is worse
	// than one that is obviously unfinished.
	if !strings.Contains(draft.Body, "_Write") {
		t.Error("the draft does not ask for the judgement it cannot derive")
	}
	// The always-on product skills are not named: they are on everywhere
	// already, and listing them is noise the reader has to filter.
	for _, s := range draft.Skills {
		if s == "writing-entries" || s == "managing-a-research" {
			t.Errorf("the draft names an always-on skill: %s", s)
		}
	}
	if len(draft.Skills) != 1 || draft.Skills[0] != "evidence-grading" {
		t.Errorf("the draft did not carry the chosen skills: %v", draft.Skills)
	}
}

func TestTemplate_ValidationRefusesWhatCannotBeMatchedOn(t *testing.T) {
	k := newTemplateKit(t)
	owner, _, _, _, teamID := k.sharedResearch(t, domain.TeamEditor)

	if _, err := k.templates.CreateTeam(owner, teamID, TemplateInput{
		Name: "No matching line", Body: "something",
	}); !errors.Is(err, ErrTemplateWhenEmpty) {
		t.Errorf("missing when_to_use: got %v", err)
	}
	if _, err := k.templates.CreateTeam(owner, teamID, TemplateInput{
		Name: "No body", WhenToUse: "Use when.",
	}); !errors.Is(err, ErrTemplateBodyEmpty) {
		t.Errorf("missing body: got %v", err)
	}
	if _, err := k.templates.CreateTeam(owner, teamID, TemplateInput{
		Name: "Chatty", WhenToUse: strings.Repeat("x", domain.TemplateCriterionMax+1), Body: "b",
	}); !errors.Is(err, ErrTemplateCriterionLong) {
		t.Errorf("over-long criterion: got %v", err)
	}
	// Runes, not bytes: a Cyrillic matching line is not worth half a Latin one.
	if _, err := k.templates.CreateTeam(owner, teamID, TemplateInput{
		Name:      "Кириллица",
		WhenToUse: "Использовать, когда " + strings.Repeat("я", domain.TemplateCriterionMax-40),
		Body:      "тело",
	}); err != nil {
		t.Errorf("a Cyrillic matching line within the rune cap was refused: %v", err)
	}
}
