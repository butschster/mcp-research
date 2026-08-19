package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/butschster/mcp-research/internal/auth"
	"github.com/butschster/mcp-research/internal/domain"
)

// What is worth testing here is the classification, because it is where the two
// kinds of failure this feature distinguishes actually live: a key we ignore is
// a fidelity decision the user carries, and a key we refuse is a hole in the
// provenance model we would be opening ourselves.

type importKit struct {
	*roleKit
	owner   context.Context
	member  context.Context
	res     *domain.Research
	section *domain.Section
}

func newImportKit(t *testing.T, role domain.TeamRole) *importKit {
	t.Helper()
	k := newRoleKit(t)
	owner, member, res, section, _ := k.sharedResearch(t, role)
	return &importKit{roleKit: k, owner: owner, member: member, res: res, section: section}
}

func (k *importKit) declare(t *testing.T, specs ...domain.FieldSpec) {
	t.Helper()
	list := specs
	if _, err := k.roleKit.section.Update(k.owner, k.section.ID, UpdateSectionRequest{FieldSpec: &list}); err != nil {
		t.Fatalf("declare field spec: %v", err)
	}
}

func (k *importKit) preview(t *testing.T, name, body string) *MarkdownImport {
	t.Helper()
	got, err := k.entry.PreviewMarkdownImport(k.owner, k.section.ID, name, []byte(body))
	if err != nil {
		t.Fatalf("preview %s: %v", name, err)
	}
	return got
}

func noteKeys(notes []ImportNote) []string {
	out := make([]string, 0, len(notes))
	for _, n := range notes {
		out = append(out, n.Key)
	}
	return out
}

func has(list []string, want string) bool {
	for _, s := range list {
		if strings.EqualFold(s, want) {
			return true
		}
	}
	return false
}

// Provenance is the one thing a text file may not assert. Everything this
// product says about who did what and in which session rests on it being the
// system's own record.
func TestMarkdownImport_RefusesProvenanceOutOfAFile(t *testing.T) {
	k := newImportKit(t, domain.TeamEditor)

	got := k.preview(t, "note.md", `---
title: Smuggled
session: SS3
author: somebody@example.com
author_kind: human
revision: 7
revisions:
  - r: 1
history: made up
---

Body.
`)

	for _, key := range []string{"session", "author", "author_kind", "revision", "revisions", "history"} {
		if !has(noteKeys(got.Refused), key) {
			t.Errorf("%q was not refused; refused = %v", key, noteKeys(got.Refused))
		}
	}
	// Refused is not the same as silently dropped: every one carries a reason
	// for the person holding the file.
	for _, n := range got.Refused {
		if strings.TrimSpace(n.Reason) == "" {
			t.Errorf("refused %q with no reason", n.Key)
		}
	}
	if len(got.Metadata) != 0 {
		t.Errorf("a refused key reached the metadata: %v", got.Metadata)
	}
}

// A code, a timestamp and the research a file thinks it came from are facts
// this product owns. Reading them and saying so is right; using them is not.
func TestMarkdownImport_IgnoresWhatTheSystemOwns(t *testing.T) {
	k := newImportKit(t, domain.TeamEditor)

	got := k.preview(t, "note.md", `---
title: Recycled
code: E50
created: 2020-01-01T00:00:00Z
updated: 2020-01-02T00:00:00Z
research: R99
section: Somewhere else
type: blocks
aliases: [E50]
---

Body.
`)

	for _, key := range []string{"code", "created", "updated", "research", "section", "type", "aliases"} {
		if !has(noteKeys(got.Ignored), key) {
			t.Errorf("%q was not reported as ignored; ignored = %v", key, noteKeys(got.Ignored))
		}
	}
	// A mismatched research is a remark, never a refusal — moving a document
	// between researches on purpose is the ordinary case.
	if len(got.Warnings) == 0 {
		t.Error("a file claiming another research produced no warning")
	}

	entry, err := k.entry.ImportMarkdown(k.owner, ImportEntryRequest{
		SectionID: k.section.ID, Title: got.Title, Body: got.Body,
	})
	if err != nil {
		t.Fatalf("commit: %v", err)
	}
	if entry.Code == "E50" {
		t.Error("the file's code was used; codes are assigned by the research")
	}
	if entry.Type != domain.EntryMarkdown {
		t.Errorf("type = %q, want markdown — a file does not get to ask for a block document", entry.Type)
	}
}

func TestMarkdownImport_MetadataIsValidatedAgainstTheSection(t *testing.T) {
	k := newImportKit(t, domain.TeamEditor)
	k.declare(t,
		domain.FieldSpec{Key: "stage", Label: "Stage", Type: domain.FieldEnum, Options: []string{"draft", "review"}},
		domain.FieldSpec{Key: "owner", Label: "Owner", Type: domain.FieldText},
	)

	got := k.preview(t, "note.md", `---
title: Declared and not
stage: review
owner: Alice
invented: whatever
---

Body.
`)

	if got.Metadata["stage"] != "review" || got.Metadata["owner"] != "Alice" {
		t.Errorf("declared values did not survive: %v", got.Metadata)
	}
	if len(got.Report.UnknownKeys) != 1 || got.Report.UnknownKeys[0].Key != "invented" {
		t.Errorf("unknown key not reported: %+v", got.Report.UnknownKeys)
	}
	if _, ok := got.Metadata["invented"]; ok {
		t.Error("an undeclared key was kept; the vocabulary is closed")
	}

	// And the commit refuses what the preview showed. This is the one write in
	// the product that does — see the comment at the top of import_markdown.go.
	_, err := k.entry.ImportMarkdown(k.owner, ImportEntryRequest{
		SectionID: k.section.ID, Title: "x", Body: "y",
		Metadata: map[string]any{"invented": "whatever"},
	})
	if !errors.Is(err, ErrImportRejected) {
		t.Errorf("commit with an undeclared key: want ErrImportRejected, got %v", err)
	}
}

func TestMarkdownImport_AFileWithNoFrontMatterIsOrdinary(t *testing.T) {
	k := newImportKit(t, domain.TeamEditor)

	got := k.preview(t, "raw.md", "## Работа с источниками\n\nПервый абзац.\n")
	if got.Title != "Работа с источниками" {
		t.Errorf("title = %q, want the first heading", got.Title)
	}
	if got.TitleSource != TitleFromHeading {
		t.Errorf("title_source = %q", got.TitleSource)
	}
	if got.Status != domain.EntryDraft {
		t.Errorf("status = %q, want draft", got.Status)
	}

	// No heading either: the filename is the last resort, and the download's
	// `E3 — Name.md` shape is unwound rather than kept in the title forever.
	got = k.preview(t, "E3 — Pricing model.md", "Just a paragraph.\n")
	if got.Title != "Pricing model" {
		t.Errorf("title = %q, want the filename without its code", got.Title)
	}
	if got.TitleSource != TitleFromFilename {
		t.Errorf("title_source = %q", got.TitleSource)
	}
}

func TestMarkdownImport_MalformedYamlNamesTheLine(t *testing.T) {
	k := newImportKit(t, domain.TeamEditor)

	_, err := k.entry.PreviewMarkdownImport(k.owner, k.section.ID, "bad.md", []byte("---\ntitle: ok\n  bad: \tindent\n---\n\nBody.\n"))
	if !errors.Is(err, ErrImportBadFrontMatter) {
		t.Fatalf("want ErrImportBadFrontMatter, got %v", err)
	}
	if !strings.Contains(err.Error(), "line") {
		t.Errorf("the message does not name a line, which is the only thing that makes it fixable: %v", err)
	}

	// An opening fence with no closing one would otherwise swallow the whole
	// document as front matter, or print YAML at the reader as prose.
	_, err = k.entry.PreviewMarkdownImport(k.owner, k.section.ID, "open.md", []byte("---\ntitle: ok\n\nBody with no closing fence.\n"))
	if !errors.Is(err, ErrImportBadFrontMatter) {
		t.Errorf("unclosed front matter: want ErrImportBadFrontMatter, got %v", err)
	}
}

// Reported, and left exactly as written. Repairing a reference would mean
// guessing what a code meant in the research the file came from.
func TestMarkdownImport_UnresolvableRefsAreReportedNotRepaired(t *testing.T) {
	k := newImportKit(t, domain.TeamEditor)

	real, err := k.entry.Create(k.owner, CreateEntryRequest{
		ResearchID: k.res.ID, SectionID: k.section.ID,
		Title: "Neighbour", Content: "Something here.",
	})
	if err != nil {
		t.Fatalf("create neighbour: %v", err)
	}

	body := "See [[" + real.Code + "]] and [[E9999]] and [[E9999]] again.\n"
	got := k.preview(t, "refs.md", body)

	if len(got.UnresolvedRefs) != 1 || got.UnresolvedRefs[0].Ref != "E9999" {
		t.Fatalf("unresolved = %v, want exactly [E9999] — the real one resolves and the duplicate is one problem", got.UnresolvedRefs)
	}
	if got.UnresolvedRefs[0].Count != 2 {
		t.Errorf("count = %d, want 2 — the same dead reference twice is one row and two occurrences", got.UnresolvedRefs[0].Count)
	}
	if !strings.Contains(got.Body, "[[E9999]]") {
		t.Error("the reference was rewritten; it must be left exactly as written")
	}
}

// The escape expansion every other write applies is wrong for a file: two
// characters in a file are two characters, and they are usually inside a code
// block explaining them.
func TestMarkdownImport_ABackslashInAFileIsABackslash(t *testing.T) {
	k := newImportKit(t, domain.TeamEditor)

	body := "```go\nfmt.Println(\"a\\nb\")\n```\n"
	got := k.preview(t, "code.md", "# Code\n\n"+body)
	entry, err := k.entry.ImportMarkdown(k.owner, ImportEntryRequest{
		SectionID: k.section.ID, Title: got.Title, Body: got.Body,
	})
	if err != nil {
		t.Fatalf("commit: %v", err)
	}
	if !strings.Contains(entry.Content, `\n`) {
		t.Errorf("the literal backslash-n was expanded, rewriting the code the document is about:\n%s", entry.Content)
	}
}

func TestMarkdownImport_RefusesWhatItWillNotRead(t *testing.T) {
	k := newImportKit(t, domain.TeamEditor)

	if _, err := k.entry.PreviewMarkdownImport(k.owner, k.section.ID, "notes.txt", []byte("hello")); !errors.Is(err, ErrImportNotMarkdown) {
		t.Errorf(".txt: want ErrImportNotMarkdown, got %v", err)
	}
	big := make([]byte, MaxImportFileBytes+1)
	for i := range big {
		big[i] = 'a'
	}
	if _, err := k.entry.PreviewMarkdownImport(k.owner, k.section.ID, "big.md", big); !errors.Is(err, ErrImportTooLarge) {
		t.Errorf("oversized: want ErrImportTooLarge, got %v", err)
	}
	if _, err := k.entry.PreviewMarkdownImport(k.owner, k.section.ID, "empty.md", []byte("---\ntitle: Only front matter\n---\n")); !errors.Is(err, ErrImportEmpty) {
		t.Errorf("front matter only: want ErrImportEmpty, got %v", err)
	}
}

func TestMarkdownImport_ViewerCannotImportAndAShareCannotReach(t *testing.T) {
	k := newImportKit(t, domain.TeamViewer)

	if _, err := k.entry.PreviewMarkdownImport(k.member, k.section.ID, "note.md", []byte("# Hi\n\nBody.\n")); !errors.Is(err, ErrForbidden) {
		t.Errorf("viewer preview: want ErrForbidden, got %v", err)
	}
	if _, err := k.entry.ImportMarkdown(k.member, ImportEntryRequest{SectionID: k.section.ID, Title: "x", Body: "y"}); !errors.Is(err, ErrForbidden) {
		t.Errorf("viewer commit: want ErrForbidden, got %v", err)
	}

	// The routes are absent from the shared sub-mux, and the service refuses a
	// share of its own accord as well — this is the only write reachable by a
	// bare section id, and one day something under /api/shared/ will take one.
	visitor := auth.WithShare(context.Background(), &auth.Share{ID: "s1", ResearchID: k.res.ID})
	if _, err := k.entry.PreviewMarkdownImport(visitor, k.section.ID, "note.md", []byte("# Hi\n\nBody.\n")); !errors.Is(err, ErrNotFound) {
		t.Errorf("share preview: want ErrNotFound, got %v", err)
	}
	if _, err := k.entry.ImportMarkdown(visitor, ImportEntryRequest{SectionID: k.section.ID, Title: "x", Body: "y"}); !errors.Is(err, ErrNotFound) {
		t.Errorf("share commit: want ErrNotFound, got %v", err)
	}
}

// A stranger must not learn that a section id is real.
func TestMarkdownImport_AStrangerGetsNotFound(t *testing.T) {
	k := newImportKit(t, domain.TeamEditor)
	stranger := userCtx(createTestUser(t, k.db, "stranger@test.com", "Stranger"))

	if _, err := k.entry.PreviewMarkdownImport(stranger, k.section.ID, "note.md", []byte("# Hi\n\nBody.\n")); !errors.Is(err, ErrNotFound) {
		t.Errorf("stranger preview: want ErrNotFound, got %v", err)
	}
	if _, err := k.entry.ImportMarkdown(stranger, ImportEntryRequest{SectionID: k.section.ID, Title: "x", Body: "y"}); !errors.Is(err, ErrNotFound) {
		t.Errorf("stranger commit: want ErrNotFound, got %v", err)
	}
}

// The round trip, asserting only what we promise.
//
// Title, description, status, tags and metadata survive download → import.
// Cross-references and identity are deliberately NOT asserted: a code is
// assigned by the research and `[[E3]]` means whatever E3 means where the file
// lands. Tightening this test into a promise about either would be inventing a
// guarantee the feature explicitly declines to make — see the top of
// import_markdown.go.
func TestMarkdownImport_RoundTripKeepsOnlyWhatWePromise(t *testing.T) {
	k := newImportKit(t, domain.TeamEditor)
	k.declare(t,
		domain.FieldSpec{Key: "stage", Label: "Stage", Type: domain.FieldEnum, Options: []string{"draft", "review"}},
	)

	original, err := k.entry.Create(k.owner, CreateEntryRequest{
		ResearchID: k.res.ID, SectionID: k.section.ID,
		Title:       "Ценообразование",
		Description: "Как мы считаем цену.",
		Status:      domain.EntryActive,
		Tags:        []string{"pricing", "модель"},
		Content:     "# Ценообразование\n\nПервый абзац с [[E9999]] внутри.\n",
		Metadata:    map[string]any{"stage": "review"},
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	file, err := k.entry.MarkdownExport(k.owner, original.ID)
	if err != nil {
		t.Fatalf("export: %v", err)
	}

	parsed, err := k.entry.PreviewMarkdownImport(k.owner, k.section.ID, file.Filename, []byte(file.Content))
	if err != nil {
		t.Fatalf("preview: %v", err)
	}
	imported, err := k.entry.ImportMarkdown(k.owner, ImportEntryRequest{
		SectionID: k.section.ID, Title: parsed.Title, Description: parsed.Description,
		Status: parsed.Status, Tags: parsed.Tags, Body: parsed.Body, Metadata: parsed.Metadata,
	})
	if err != nil {
		t.Fatalf("commit: %v", err)
	}

	if imported.Title != original.Title {
		t.Errorf("title = %q, want %q", imported.Title, original.Title)
	}
	if imported.Description != original.Description {
		t.Errorf("description = %q, want %q", imported.Description, original.Description)
	}
	if imported.Status != original.Status {
		t.Errorf("status = %q, want %q", imported.Status, original.Status)
	}
	if strings.Join(imported.Tags, ",") != strings.Join(original.Tags, ",") {
		t.Errorf("tags = %v, want %v", imported.Tags, original.Tags)
	}
	if imported.Metadata["stage"] != "review" {
		t.Errorf("metadata = %v, want stage=review", imported.Metadata)
	}
	// Identity is NOT preserved, and that is the design.
	if imported.Code == original.Code {
		t.Error("the copy took the original's code")
	}
}

// Revision 1 says the document arrived as a file. `import` is a value the
// revisions table has always known; nothing was writing it at this granularity
// until now.
func TestMarkdownImport_TheFirstRevisionSaysItWasImported(t *testing.T) {
	k := newImportKit(t, domain.TeamEditor)

	got := k.preview(t, "note.md", "# Imported\n\nBody.\n")
	entry, err := k.entry.ImportMarkdown(k.owner, ImportEntryRequest{
		SectionID: k.section.ID, Title: got.Title, Body: got.Body,
	})
	if err != nil {
		t.Fatalf("commit: %v", err)
	}

	_, revs, err := k.entry.History(k.owner, entry.ID)
	if err != nil {
		t.Fatalf("history: %v", err)
	}
	if len(revs) == 0 {
		t.Fatal("no revision recorded for an imported document")
	}
	first := revs[len(revs)-1]
	if first.AuthorKind != domain.AuthorImport {
		t.Errorf("author_kind = %q, want %q", first.AuthorKind, domain.AuthorImport)
	}
}

// The four keys that become the entry have their own note channel, because a
// value we read and could not use is neither a metadata problem nor a key we
// declined by policy — and replacing it with a default silently is the loss
// this whole preview exists to prevent.
func TestMarkdownImport_ReportsWhatItDidWithTheCoreFields(t *testing.T) {
	k := newImportKit(t, domain.TeamEditor)

	got := k.preview(t, "note.md", `---
title: Fine
status: shipped
tags: pricing, модель
---

Body.
`)

	byKey := map[string]FieldNote{}
	for _, f := range got.Fields {
		byKey[f.Key] = f
	}

	st, ok := byKey["status"]
	if !ok {
		t.Fatalf("a status the file spells wrong was not reported: %+v", got.Fields)
	}
	if st.Applied || st.Value != "shipped" {
		t.Errorf("status note = %+v, want applied=false and the value the file gave", st)
	}
	if got.Status != domain.EntryDraft {
		t.Errorf("status = %q, want draft", got.Status)
	}

	// A hand-written comma line is accepted, because people write it — and the
	// split is a guess, so it is admitted to.
	tg, ok := byKey["tags"]
	if !ok {
		t.Fatalf("a comma-separated tags line was split without saying so: %+v", got.Fields)
	}
	if !tg.Applied {
		t.Errorf("tags note = %+v, want applied=true", tg)
	}
	if len(got.Tags) != 2 || got.Tags[0] != "pricing" || got.Tags[1] != "модель" {
		t.Errorf("tags = %v", got.Tags)
	}

	// A title taken from the filename is the one the person will want to change,
	// so it is reported even though nothing went wrong.
	fromFile := k.preview(t, "Untitled note.md", "Just prose, no heading.\n")
	if fromFile.TitleSource != TitleFromFilename {
		t.Fatalf("title_source = %q", fromFile.TitleSource)
	}
	var sawTitle bool
	for _, f := range fromFile.Fields {
		if f.Key == "title" {
			sawTitle = true
		}
	}
	if !sawTitle {
		t.Error("a title guessed from the filename was not reported")
	}
}

// A rejected metadata value is reported with what the file actually said. The
// author of the file is usually not the person importing it.
func TestMarkdownImport_AReportSaysWhatTheFileClaimed(t *testing.T) {
	k := newImportKit(t, domain.TeamEditor)
	k.declare(t, domain.FieldSpec{
		Key: "stage", Label: "Stage", Type: domain.FieldEnum, Options: []string{"draft", "review"},
	})

	got := k.preview(t, "note.md", "---\ntitle: T\nstage: shipped\nmade_up: 7\n---\n\nBody.\n")

	if len(got.Report.InvalidValues) != 1 || got.Report.InvalidValues[0].Value != "shipped" {
		t.Errorf("invalid values = %+v, want the value the file gave", got.Report.InvalidValues)
	}
	if len(got.Report.UnknownKeys) != 1 || got.Report.UnknownKeys[0].Value != "7" {
		t.Errorf("unknown keys = %+v, want the value the file gave", got.Report.UnknownKeys)
	}
}

// scalarString is on two paths at once: rendering a value for a report, and
// producing the entry's own fields. Clamping it for the first cut the second.
func TestMarkdownImport_ALongTitleAndDescriptionSurviveWhole(t *testing.T) {
	k := newImportKit(t, domain.TeamEditor)

	title := strings.Repeat("Заголовок ", 30)         // ~300 runes
	desc := strings.Repeat("Описание документа ", 20) // ~380 runes
	got := k.preview(t, "long.md", "---\ntitle: "+title+"\ndescription: "+desc+"\n---\n\nBody.\n")

	if strings.Contains(got.Title, "…") {
		t.Errorf("title was clamped by the report formatter: %q", got.Title)
	}
	if strings.Contains(got.Description, "…") {
		t.Errorf("description was clamped by the report formatter: %q", got.Description)
	}
	if len([]rune(got.Description)) < 300 {
		t.Errorf("description is %d runes, want the whole thing", len([]rune(got.Description)))
	}

	// A report value is still clamped, which is what the clamp was for.
	noted := k.preview(t, "note.md", "---\ntitle: T\nsession: "+strings.Repeat("x", 400)+"\n---\n\nBody.\n")
	if len(noted.Refused) == 0 || !strings.HasSuffix(noted.Refused[0].Value, "…") {
		t.Errorf("a 400-character report value was not clamped: %+v", noted.Refused)
	}
}

// yaml.v3 resolves an unquoted scalar to its Go type, and the metadata
// validator speaks strings. A person writing the most ordinary front matter
// there is was told their valid date was not text.
func TestMarkdownImport_UnquotedYamlScalarsReachTheValidator(t *testing.T) {
	k := newImportKit(t, domain.TeamEditor)
	k.declare(t,
		domain.FieldSpec{Key: "due", Label: "Due", Type: domain.FieldDate},
		domain.FieldSpec{Key: "flag", Label: "Flag", Type: domain.FieldText},
		domain.FieldSpec{Key: "score", Label: "Score", Type: domain.FieldNumber},
	)

	got := k.preview(t, "note.md", "---\ntitle: T\ndue: 2024-01-02\nflag: true\nscore: 7\n---\n\nBody.\n")

	if len(got.Report.InvalidValues) != 0 {
		t.Fatalf("ordinary unquoted YAML was rejected: %+v", got.Report.InvalidValues)
	}
	if got.Metadata["due"] != "2024-01-02" {
		t.Errorf("due = %#v, want the date as written", got.Metadata["due"])
	}
	if got.Metadata["flag"] != "true" {
		t.Errorf("flag = %#v", got.Metadata["flag"])
	}
}

// ValidateMetadata keeps an invalid value on purpose — right for an agent
// mid-interview, and a dead end behind a commit that refuses. The preview drops
// it, so the document can be imported with the field empty and the report
// saying why.
func TestMarkdownImport_ARejectedValueIsNotADeadEnd(t *testing.T) {
	k := newImportKit(t, domain.TeamEditor)
	k.declare(t, domain.FieldSpec{
		Key: "stage", Label: "Stage", Type: domain.FieldEnum, Options: []string{"draft", "review"},
	})

	got := k.preview(t, "note.md", "---\ntitle: T\nstage: shipped\n---\n\nBody.\n")
	if len(got.Report.InvalidValues) != 1 {
		t.Fatalf("the bad value was not reported: %+v", got.Report)
	}
	if _, kept := got.Metadata["stage"]; kept {
		t.Error("the rejected value is still in the payload the dialog posts back, so the commit refuses it forever")
	}

	// The exact payload the preview produces has to be committable.
	entry, err := k.entry.ImportMarkdown(k.owner, ImportEntryRequest{
		SectionID: k.section.ID, Title: got.Title, Description: got.Description,
		Status: got.Status, Tags: got.Tags, Body: got.Body, Metadata: got.Metadata,
	})
	if err != nil {
		t.Fatalf("the preview's own payload was refused: %v", err)
	}
	if _, set := entry.Metadata["stage"]; set {
		t.Error("the rejected value was stored after all")
	}
}
