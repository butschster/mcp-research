package service

import (
	"context"
	"log/slog"
	"sort"
	"strings"
	"testing"

	"github.com/dovod-app/app/internal/domain"
	"github.com/dovod-app/app/internal/storage"
	"gopkg.in/yaml.v3"
)

type obsidianFixture struct {
	svc      *ObsidianService
	research *ResearchService
	section  *SectionService
	entry    *EntryService
	session  *SessionService
	task     *TaskService
	roadmap  *RoadmapService
}

func setupObsidian(t *testing.T) *obsidianFixture {
	t.Helper()
	db := setupTestDB(t)
	events := &mockNotifier{}
	log := slog.Default()

	researchRepo := storage.NewResearchRepository(db)
	sectionRepo := storage.NewSectionRepository(db)
	entryRepo := storage.NewEntryRepository(db)
	blockRepo := storage.NewBlockRepository(db)
	sessionRepo := storage.NewSessionRepository(db)
	questionRepo := storage.NewQuestionRepository(db)
	taskRepo := storage.NewTaskRepository(db)
	revisionRepo := storage.NewEntryRevisionRepository(db)
	crossrefRepo := storage.NewCrossRefRepository(db)
	externalLinkRepo := storage.NewExternalLinkRepository(db)
	roadmapRepo := storage.NewRoadmapRepository(db)
	roadmapNodeRepo := storage.NewRoadmapNodeRepository(db)
	roadmapEdgeRepo := storage.NewRoadmapEdgeRepository(db)

	researchSvc := NewResearchService(researchRepo, sectionRepo, storage.NewTeamRepository(db), testAccess(db), events, log)
	sectionSvc := NewSectionService(sectionRepo, entryRepo, researchRepo, testAccess(db), events, log)
	entrySvc := NewEntryService(entryRepo, sectionRepo, researchRepo, testAccess(db), sessionRepo, blockRepo, revisionRepo, crossrefRepo, externalLinkRepo, events, log)
	entrySvc.SetRoadmapRepos(roadmapRepo, roadmapNodeRepo)
	sessionSvc := NewSessionService(db, sessionRepo, questionRepo, researchRepo, testAccess(db), entrySvc, events, log)
	taskSvc := NewTaskService(taskRepo, researchRepo, testAccess(db), entrySvc, events, log)
	roadmapSvc := NewRoadmapService(roadmapRepo, roadmapNodeRepo, roadmapEdgeRepo, researchRepo, testAccess(db), events, log)
	roadmapSvc.SetRefResolvers(entryRepo, taskRepo, sessionRepo, questionRepo, sectionRepo)

	return &obsidianFixture{
		svc:      NewObsidianService(researchSvc, sectionSvc, entryRepo, sessionSvc, taskSvc, roadmapSvc, revisionRepo, log),
		research: researchSvc, section: sectionSvc, entry: entrySvc,
		session: sessionSvc, task: taskSvc, roadmap: roadmapSvc,
	}
}

// seedVault builds a research that exercises every branch of the exporter: two
// sections, a markdown entry and a blocks entry, a session with questions, a
// task, a roadmap, and cross-references that resolve and cross-references that
// do not.
func seedVault(t *testing.T, f *obsidianFixture) *domain.Research {
	t.Helper()
	ctx := context.Background()

	research, sections, err := f.research.Create(ctx, CreateResearchRequest{
		Name:        "Competitive Landscape",
		Goal:        "Understand the market",
		Description: "A description",
		Tags:        []string{"market", "two words"},
		Sections: []CreateSectionRequest{
			{Name: "market", DisplayName: "Market Analysis", Position: 1},
			{Name: "rivals", DisplayName: "Competitors", Position: 2},
		},
	})
	if err != nil {
		t.Fatalf("create research: %v", err)
	}

	sess, _, err := f.session.Create(ctx, CreateSessionRequest{
		ResearchID: research.ID,
		Title:      "Initial exploration",
		Focus:      "Pricing",
		Questions: []CreateQuestionRequest{
			{Text: "How is it priced?", Area: "pricing", Priority: domain.PriorityHigh},
		},
	})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}

	if _, err := f.entry.Create(ctx, CreateEntryRequest{
		ResearchID: research.ID,
		SectionID:  sections[0].ID,
		SessionID:  sess.ID,
		Title:      "Pricing: model",
		Content:    "Seats are billed monthly. See [[E2]], [[R2:E5]] and [[E9]].",
		Tags:       []string{"pricing"},
	}); err != nil {
		t.Fatalf("create entry 1: %v", err)
	}

	if _, err := f.entry.Create(ctx, CreateEntryRequest{
		ResearchID: research.ID,
		SectionID:  sections[1].ID,
		Type:       domain.EntryBlocks,
		Title:      "Direct competitors",
		Content: `{"version":1,"blocks":[
			{"type":"heading","data":{"level":2,"text":"Who they are"}},
			{"type":"paragraph","data":{"text":"Three of them matter."}},
			{"type":"checklist","data":{"title":"To verify","items":[{"key":"a","text":"pricing"},{"key":"b","text":"headcount"}]}},
			{"type":"mermaid","data":{"code":"flowchart TD\n  A --> B","caption":"Shape"}},
			{"type":"html","data":{"html":"<html><body>chart</body></html>","title":"Revenue chart","caption":"interactive"}}
		]}`,
	}); err != nil {
		t.Fatalf("create entry 2: %v", err)
	}

	if _, err := f.task.Create(ctx, CreateTaskRequest{
		ResearchID: research.ID,
		Title:      "Verify market share",
		Priority:   domain.PriorityHigh,
	}); err != nil {
		t.Fatalf("create task: %v", err)
	}

	if _, err := f.roadmap.Create(ctx, CreateRoadmapRequest{
		ResearchID: research.ID,
		Title:      "Migration plan",
		Nodes: []CreateRoadmapNodeRequest{
			{TempID: "a", Title: "Start"},
			{TempID: "b", Title: `Finish "it"`, Status: "done"},
		},
		Edges: []CreateRoadmapEdgeRequest{
			{SourceNodeRef: "a", TargetNodeRef: "b", Label: "then"},
		},
	}); err != nil {
		t.Fatalf("create roadmap: %v", err)
	}

	return research
}

func vaultOf(t *testing.T, f *obsidianFixture, id string, opts VaultOptions) *Vault {
	t.Helper()
	v, err := f.svc.Vault(context.Background(), id, opts)
	if err != nil {
		t.Fatalf("build vault: %v", err)
	}
	return v
}

func fileIn(v *Vault, path string) (string, bool) {
	for _, f := range v.Files {
		if f.Path == path {
			return string(f.Content), true
		}
	}
	return "", false
}

func pathsOf(v *Vault) []string {
	out := make([]string, 0, len(v.Files))
	for _, f := range v.Files {
		out = append(out, f.Path)
	}
	sort.Strings(out)
	return out
}

func TestObsidian_VaultLayout(t *testing.T) {
	f := setupObsidian(t)
	research := seedVault(t, f)
	v := vaultOf(t, f, research.ID, DefaultVaultOptions())

	if v.Root != "R1 — Competitive Landscape" {
		t.Errorf("root folder = %q", v.Root)
	}

	want := []string{
		"01 — Market Analysis/E1 — Pricing model.md",
		"02 — Competitors/E2 — Direct competitors.md",
		"README.md",
		"Roadmaps/RM1 — Migration plan.md",
		"Sessions/SS1 — Initial exploration.md",
		"Tasks/T1 — Verify market share.md",
	}
	got := pathsOf(v)
	for _, w := range want {
		if _, ok := fileIn(v, w); !ok {
			t.Errorf("missing %q\ngot: %v", w, got)
		}
	}

	// The colon in "Pricing: model" is gone from the filename — no filesystem in
	// play accepts it — but nothing else about the title is.
	for _, p := range got {
		if strings.ContainsAny(p[strings.LastIndex(p, "/")+1:], `:*?"<>|`) {
			t.Errorf("filename holds a character a filesystem refuses: %q", p)
		}
	}
}

func TestObsidian_FrontmatterParsesEverywhere(t *testing.T) {
	f := setupObsidian(t)
	research := seedVault(t, f)
	v := vaultOf(t, f, research.ID, DefaultVaultOptions())

	for _, file := range v.Files {
		if !strings.HasSuffix(file.Path, ".md") {
			continue
		}
		body := string(file.Content)
		if !strings.HasPrefix(body, "---\n") {
			t.Errorf("%s has no frontmatter", file.Path)
			continue
		}
		end := strings.Index(body[4:], "\n---\n")
		if end < 0 {
			t.Errorf("%s frontmatter is not terminated", file.Path)
			continue
		}
		var fm map[string]any
		if err := yaml.Unmarshal([]byte(body[4:4+end]), &fm); err != nil {
			t.Errorf("%s frontmatter does not parse: %v\n%s", file.Path, err, body[:4+end])
		}
	}

	entry, _ := fileIn(v, "01 — Market Analysis/E1 — Pricing model.md")
	for _, want := range []string{
		"aliases: [E1]",
		"code: E1",
		"research: R1",
		"section: Market Analysis",
		"session: SS1",
	} {
		if !strings.Contains(entry, want) {
			t.Errorf("entry note is missing %q:\n%s", want, entry)
		}
	}
	// A title holding a colon has to come back quoted, or the note stops parsing.
	if !strings.Contains(entry, `'Pricing: model'`) && !strings.Contains(entry, `"Pricing: model"`) {
		t.Errorf("a title with a colon was not quoted:\n%s", entry)
	}
}

func TestObsidian_ReferencesResolve(t *testing.T) {
	f := setupObsidian(t)
	research := seedVault(t, f)
	v := vaultOf(t, f, research.ID, DefaultVaultOptions())

	entry, ok := fileIn(v, "01 — Market Analysis/E1 — Pricing model.md")
	if !ok {
		t.Fatal("entry note missing")
	}
	// A reference is retargeted at the note's filename and keeps its code as the
	// display half — an alias does not resolve a hand-written link, so this is
	// what makes it work. What the reader sees is unchanged.
	if !strings.Contains(entry, "See [[E2 — Direct competitors|E2]], [[R2 E5|R2:E5]] and [[E9]].") {
		t.Errorf("references were not retargeted at their notes:\n%s", entry)
	}
	// [[E9]] needs no display half: the stub it lands on is named E9.
	if strings.Contains(entry, "[[E9|") {
		t.Errorf("a link whose target already reads as its code should stay plain:\n%s", entry)
	}

	// E2 is in the vault, so it gets no stub.
	if _, ok := fileIn(v, "_Unresolved/E2.md"); ok {
		t.Error("a reference that resolves should not produce a stub")
	}
	// The two that point outside do, so the reader sees a reference rather than
	// a link that quietly goes nowhere.
	for path, want := range map[string]string{
		"_Unresolved/R2 E5.md": "an entry in another research",
		"_Unresolved/E9.md":    "an entry",
	} {
		body, ok := fileIn(v, path)
		if !ok {
			t.Errorf("missing stub %q, got %v", path, pathsOf(v))
			continue
		}
		if !strings.Contains(body, want) {
			t.Errorf("%s does not say what it points at:\n%s", path, body)
		}
	}
	// Quoted, because a bare colon inside a flow sequence is not valid YAML —
	// which is the whole reason the frontmatter goes through an encoder.
	if body, _ := fileIn(v, "_Unresolved/R2 E5.md"); !strings.Contains(body, `aliases: ['R2:E5']`) {
		t.Errorf("a stub must answer to the reference that produced it:\n%s", body)
	}
}

func TestObsidian_BlocksBecomeReadableNotes(t *testing.T) {
	f := setupObsidian(t)
	research := seedVault(t, f)
	v := vaultOf(t, f, research.ID, DefaultVaultOptions())

	note, ok := fileIn(v, "02 — Competitors/E2 — Direct competitors.md")
	if !ok {
		t.Fatal("blocks note missing")
	}
	for _, needle := range []string{`"type"`, `"blocks"`, `{"`} {
		if strings.Contains(note, needle) {
			t.Errorf("serialized JSON leaked into the note (%q):\n%s", needle, note)
		}
	}
	for _, want := range []string{"## Who they are", "- [ ] pricing", "```mermaid"} {
		if !strings.Contains(note, want) {
			t.Errorf("note is missing %q:\n%s", want, note)
		}
	}
	// Obsidian draws mermaid itself; the live-editor link would be a kilobyte of
	// base64 in the middle of a note.
	if strings.Contains(note, "mermaid.live") {
		t.Errorf("the live-editor link should be omitted in a vault:\n%s", note)
	}

	// The HTML block is a real file, and the note links to it rather than
	// describing it — Obsidian sanitizes inline HTML and runs no scripts.
	var htmlPath string
	for _, file := range v.Files {
		if strings.HasPrefix(file.Path, "_html/") {
			htmlPath = file.Path
		}
	}
	if htmlPath == "" {
		t.Fatalf("no html file written: %v", pathsOf(v))
	}
	if !strings.Contains(htmlPath, "E2 — Revenue chart") {
		t.Errorf("html file is not named after its entry and block: %q", htmlPath)
	}
	if !strings.Contains(note, "Open in a browser") || !strings.Contains(note, "../_html/") {
		t.Errorf("the note does not link to its html file:\n%s", note)
	}
	body, _ := fileIn(v, htmlPath)
	if !strings.Contains(body, "<body>chart</body>") {
		t.Errorf("the html document was altered:\n%s", body)
	}
}

func TestObsidian_ReadmeIndexesEverything(t *testing.T) {
	f := setupObsidian(t)
	research := seedVault(t, f)
	v := vaultOf(t, f, research.ID, DefaultVaultOptions())

	readme, ok := fileIn(v, "README.md")
	if !ok {
		t.Fatal("README missing")
	}
	// Every document in the vault has to be reachable from the note a reader
	// opens first.
	for _, want := range []string{
		"## Contents",
		"### 01 — Market Analysis",
		"### 02 — Competitors",
		"[[E1 — Pricing model]]",
		"[[E2 — Direct competitors]]",
		"[[SS1 — Initial exploration]]",
		"[[T1 — Verify market share]]",
		"[[RM1 — Migration plan]]",
	} {
		if !strings.Contains(readme, want) {
			t.Errorf("README is missing %q:\n%s", want, readme)
		}
	}
	if !strings.Contains(readme, "aliases: [R1]") {
		t.Error("the index note should be findable as R1 in the quick switcher")
	}
}

func TestObsidian_SessionAndTaskNotes(t *testing.T) {
	f := setupObsidian(t)
	research := seedVault(t, f)
	v := vaultOf(t, f, research.ID, DefaultVaultOptions())

	sess, _ := fileIn(v, "Sessions/SS1 — Initial exploration.md")
	for _, want := range []string{
		"aliases: [SS1]",
		"How is it priced?",
		"## Entries produced in this session",
		"[[E1 — Pricing model]]",
	} {
		if !strings.Contains(sess, want) {
			t.Errorf("session note is missing %q:\n%s", want, sess)
		}
	}

	entry, _ := fileIn(v, "01 — Market Analysis/E1 — Pricing model.md")
	if !strings.Contains(entry, "Session: [[SS1 — Initial exploration]]") {
		t.Errorf("an entry should link back to the session that produced it:\n%s", entry)
	}
	// The way home is a relative path, not a name: two exports unzipped into one
	// vault hold two README.md files, and a name would not tell them apart.
	if !strings.Contains(entry, "Research: [R1 — Competitive Landscape](../README.md)") {
		t.Errorf("an entry should link home by path:\n%s", entry)
	}

	task, _ := fileIn(v, "Tasks/T1 — Verify market share.md")
	for _, want := range []string{"aliases: [T1]", "priority: high", "- [ ] Verify market share"} {
		if !strings.Contains(task, want) {
			t.Errorf("task note is missing %q:\n%s", want, task)
		}
	}
}

func TestObsidian_RoadmapDrawsAsMermaid(t *testing.T) {
	f := setupObsidian(t)
	research := seedVault(t, f)
	v := vaultOf(t, f, research.ID, DefaultVaultOptions())

	rm, ok := fileIn(v, "Roadmaps/RM1 — Migration plan.md")
	if !ok {
		t.Fatal("roadmap note missing")
	}
	for _, want := range []string{"```mermaid", "flowchart TD", "-->|then|"} {
		if !strings.Contains(rm, want) {
			t.Errorf("roadmap note is missing %q:\n%s", want, rm)
		}
	}
	// A quote in a node title would end the label and break the diagram — inside
	// the fence. Outside it the title is prose and stays as the author wrote it.
	diagram := rm[strings.Index(rm, "```mermaid"):]
	diagram = diagram[:strings.Index(diagram, "```\n\n")]
	if strings.Contains(diagram, `Finish "it"`) {
		t.Errorf("a quote reached a mermaid label unescaped:\n%s", diagram)
	}
	if !strings.Contains(diagram, "#quot;") {
		t.Errorf("the quote should be escaped the way mermaid expects:\n%s", diagram)
	}
	if !strings.Contains(rm, `### N2 — Finish "it"`) {
		t.Errorf("the node heading should keep the title as written:\n%s", rm)
	}
	// A node link is written [[RM1:N1]]; the roadmap note answers to it.
	if !strings.Contains(rm, "RM1:N1") {
		t.Errorf("roadmap should alias its own nodes:\n%s", rm)
	}
}

func TestObsidian_OptionsOmitFolders(t *testing.T) {
	f := setupObsidian(t)
	research := seedVault(t, f)
	v := vaultOf(t, f, research.ID, VaultOptions{})

	for _, prefix := range []string{"Sessions/", "Tasks/", "Roadmaps/", "_html/"} {
		for _, file := range v.Files {
			if strings.HasPrefix(file.Path, prefix) {
				t.Errorf("%s was written although its option is off", file.Path)
			}
		}
	}
	// The entries themselves are never optional.
	if _, ok := fileIn(v, "01 — Market Analysis/E1 — Pricing model.md"); !ok {
		t.Error("entries must survive every option")
	}
	// With html off the block still says what it was, without a link to nothing.
	note, _ := fileIn(v, "02 — Competitors/E2 — Direct competitors.md")
	if !strings.Contains(note, "omitted from this export") {
		t.Errorf("an omitted html block should say so:\n%s", note)
	}
	if strings.Contains(note, "../_html/") {
		t.Errorf("a link to an omitted file was written:\n%s", note)
	}
	// A session note is gone, so nothing may link to one.
	entry, _ := fileIn(v, "01 — Market Analysis/E1 — Pricing model.md")
	if strings.Contains(entry, "Session: [[SS1 — Initial exploration]]") {
		t.Errorf("an entry links to a session note that was not exported:\n%s", entry)
	}
}

func TestObsidian_RevisionsAreOptional(t *testing.T) {
	f := setupObsidian(t)
	research := seedVault(t, f)

	off := vaultOf(t, f, research.ID, DefaultVaultOptions())
	for _, file := range off.Files {
		if strings.HasPrefix(file.Path, "_history/") {
			t.Errorf("%s was written with revisions off", file.Path)
		}
	}

	opts := DefaultVaultOptions()
	opts.Revisions = true
	on := vaultOf(t, f, research.ID, opts)
	body, ok := fileIn(on, "_history/E1.md")
	if !ok {
		t.Fatalf("no history note: %v", pathsOf(on))
	}
	for _, want := range []string{"| Revision |", "[[E1 — Pricing model|E1]]", "agent"} {
		if !strings.Contains(body, want) {
			t.Errorf("history note is missing %q:\n%s", want, body)
		}
	}
	entry, _ := fileIn(on, "01 — Market Analysis/E1 — Pricing model.md")
	for _, want := range []string{"revision: 1", "author: agent", "../_history/E1.md"} {
		if !strings.Contains(entry, want) {
			t.Errorf("entry note is missing %q with revisions on:\n%s", want, entry)
		}
	}
}

func TestObsidian_TagsObsidianCannotIndexAreNamed(t *testing.T) {
	f := setupObsidian(t)
	ctx := context.Background()
	research, sections, err := f.research.Create(ctx, CreateResearchRequest{
		Name:     "Tagged",
		Sections: []CreateSectionRequest{{Name: "s", DisplayName: "S", Position: 1}},
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if _, err := f.entry.Create(ctx, CreateEntryRequest{
		ResearchID: research.ID,
		SectionID:  sections[0].ID,
		Title:      "Tagged entry",
		Content:    "body",
		Tags:       []string{"clean", "two words"},
	}); err != nil {
		t.Fatalf("create entry: %v", err)
	}

	v := vaultOf(t, f, research.ID, DefaultVaultOptions())
	note, _ := fileIn(v, "01 — S/E1 — Tagged entry.md")
	// The exact tag stays in the frontmatter — mangling an author's vocabulary
	// to fit a reader is worse than saying plainly that it will not index.
	if !strings.Contains(note, "two words") {
		t.Errorf("the tag was dropped:\n%s", note)
	}
	if !strings.Contains(note, "Tags Obsidian cannot index") {
		t.Errorf("the note does not say the tag will not index:\n%s", note)
	}
}

func TestSanitizeSegment(t *testing.T) {
	long := strings.Repeat("ы", 400)
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"a colon is dropped rather than replaced", "Pricing: model", "Pricing model"},
		{"a path separator becomes a dash, not nothing", "часть 1/2", "часть 1-2"},
		{"both separators", "a/b\\c", "a-b-c"},
		{"windows reserved names get out of the way", "CON", "CON_"},
		{"reserved check is case-insensitive", "nul", "nul_"},
		{"a trailing dot is stripped by Windows anyway", "report.", "report"},
		{"whitespace collapses", "  a   b  ", "a b"},
		{"an empty name falls back", "***", "Entry"},
		{"newlines become spaces", "a\nb", "a b"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := sanitizeSegment(c.in, "Entry"); got != c.want {
				t.Errorf("sanitizeSegment(%q) = %q, want %q", c.in, got, c.want)
			}
		})
	}

	t.Run("a long name is cut on a rune boundary", func(t *testing.T) {
		got := sanitizeSegment(long, "Entry")
		if len(got) > maxSegment {
			t.Errorf("segment is %d bytes, over the %d cap", len(got), maxSegment)
		}
		// The ellipsis is the point: a name cut mid-title must not read as a
		// shorter title that says something else.
		if !strings.HasSuffix(got, "…") {
			t.Errorf("a truncated name should say so: %q", got)
		}
		if body := strings.TrimSuffix(got, "…"); strings.ContainsRune(got, '\uFFFD') || !strings.HasPrefix(long, body) {
			t.Errorf("a character was cut in half: %q", got)
		}
	})

	t.Run("Cyrillic survives", func(t *testing.T) {
		if got := sanitizeSegment("Ценообразование", "Entry"); got != "Ценообразование" {
			t.Errorf("Cyrillic was mangled: %q", got)
		}
	})

	t.Run("emoji alone is still a name", func(t *testing.T) {
		if got := sanitizeSegment("🔥", "Entry"); got != "🔥" {
			t.Errorf("emoji title = %q", got)
		}
	})
}

func TestObsidian_MissingResearch(t *testing.T) {
	f := setupObsidian(t)
	if _, err := f.svc.Vault(context.Background(), "R99", DefaultVaultOptions()); err == nil {
		t.Fatal("a research that does not exist should not produce a vault")
	}
}

// TestObsidian_EveryLinkHasATarget is the regression test for the bug a real
// vault found: every note answered to its code through `aliases`, and Obsidian
// resolved none of them — clicking a link in the index offered to create an
// empty note instead of opening the entry.
//
// So: every [[...]] the exporter writes must name a file that exists in the
// archive. Nothing else proves the vault is navigable.
func TestObsidian_EveryLinkHasATarget(t *testing.T) {
	f := setupObsidian(t)
	research := seedVault(t, f)

	for _, opts := range []VaultOptions{DefaultVaultOptions(), {}, {Sessions: true, Revisions: true}} {
		v := vaultOf(t, f, research.ID, opts)

		notes := map[string]bool{}
		for _, file := range v.Files {
			base := file.Path[strings.LastIndex(file.Path, "/")+1:]
			notes[strings.TrimSuffix(base, ".md")] = true
		}

		for _, file := range v.Files {
			if !strings.HasSuffix(file.Path, ".md") {
				continue
			}
			for _, m := range refPattern.FindAllStringSubmatch(string(file.Content), -1) {
				target := m[1]
				if i := strings.IndexByte(target, '|'); i >= 0 {
					target = target[:i]
				}
				if !notes[target] {
					t.Errorf("%s links to [[%s]], which is no file in the archive (opts %+v)",
						file.Path, target, opts)
				}
			}
		}
	}
}

func TestObsidian_LinksSurviveOmittedFolders(t *testing.T) {
	f := setupObsidian(t)
	research := seedVault(t, f)
	// Sessions off: the session note is gone, so a [[SS1]] anywhere in the
	// content has to land on a stub rather than on nothing.
	v := vaultOf(t, f, research.ID, VaultOptions{})

	if _, ok := fileIn(v, "_Unresolved/E2.md"); ok {
		t.Error("an entry is never omitted, so it never needs a stub")
	}
	entry, _ := fileIn(v, "01 — Market Analysis/E1 — Pricing model.md")
	if !strings.Contains(entry, "[[E2 — Direct competitors|E2]]") {
		t.Errorf("a link between two entries must survive every option:\n%s", entry)
	}
}
