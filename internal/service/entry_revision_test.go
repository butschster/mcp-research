package service

import (
	"context"
	"log/slog"
	"strings"
	"testing"

	"github.com/butschster/mcp-research/internal/domain"
	"github.com/butschster/mcp-research/internal/storage"
)

// revisionFixture builds a service with sessions wired, which the revision code
// needs: a revision records the session that was active when it was written.
func revisionFixture(t *testing.T) (*EntryService, *SessionService, context.Context, *domain.Research, *domain.Section) {
	t.Helper()
	db := setupTestDB(t)
	log := slog.Default()

	researchRepo := storage.NewResearchRepository(db)
	sectionRepo := storage.NewSectionRepository(db)
	entryRepo := storage.NewEntryRepository(db)
	blockRepo := storage.NewBlockRepository(db)
	revisionRepo := storage.NewEntryRevisionRepository(db)
	crossrefRepo := storage.NewCrossRefRepository(db)
	sessionRepo := storage.NewSessionRepository(db)
	questionRepo := storage.NewQuestionRepository(db)

	researchSvc := NewResearchService(researchRepo, sectionRepo, storage.NewTeamRepository(db), testAccess(db), &mockNotifier{}, log)
	entrySvc := NewEntryService(entryRepo, sectionRepo, researchRepo, testAccess(db), sessionRepo, blockRepo, revisionRepo, crossrefRepo, nil, &mockNotifier{}, log)
	sessionSvc := NewSessionService(db, sessionRepo, questionRepo, researchRepo, testAccess(db), entrySvc, &mockNotifier{}, log)

	ctx := context.Background()
	research, sections, err := researchSvc.Create(ctx, CreateResearchRequest{
		Name: "R", Goal: "T",
		Sections: []CreateSectionRequest{{Name: "s1", DisplayName: "S1"}},
	})
	if err != nil {
		t.Fatalf("create research: %v", err)
	}
	return entrySvc, sessionSvc, ctx, research, sections[0]
}

func TestRevisions_CreateWritesFirstRevision(t *testing.T) {
	svc, _, ctx, research, section := revisionFixture(t)

	entry, err := svc.Create(ctx, CreateEntryRequest{
		ResearchID: research.ID,
		SectionID:  section.ID,
		Title:      "Pricing",
		Content:    "Seat based above 50 users.",
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	_, revs, err := svc.History(ctx, entry.ID)
	if err != nil {
		t.Fatalf("history: %v", err)
	}
	if len(revs) != 1 {
		t.Fatalf("want 1 revision after create, got %d", len(revs))
	}
	if revs[0].Revision != 1 {
		t.Errorf("first revision must be numbered 1, got %d", revs[0].Revision)
	}
	if got := revisionContent(t, svc, ctx, entry.ID, 1); got != "Seat based above 50 users." {
		t.Errorf("revision content = %q", got)
	}
	if revs[0].AuthorKind != domain.AuthorAgent {
		t.Errorf("author = %q, want agent (the default for anything not saying otherwise)", revs[0].AuthorKind)
	}
}

func TestRevisions_UpdateAppends(t *testing.T) {
	svc, _, ctx, research, section := revisionFixture(t)
	entry := mustEntry(t, svc, ctx, research.ID, section.ID, "First version.")

	if _, err := svc.Update(ctx, entry.ID, UpdateEntryRequest{Content: ptr("Second version.")}); err != nil {
		t.Fatalf("update: %v", err)
	}

	_, revs, _ := svc.History(ctx, entry.ID)
	if len(revs) != 2 {
		t.Fatalf("want 2 revisions, got %d", len(revs))
	}
	// Newest first.
	if revs[0].Revision != 2 || revs[1].Revision != 1 {
		t.Fatalf("history must be newest first: got %d then %d", revs[0].Revision, revs[1].Revision)
	}
	if got := revisionContent(t, svc, ctx, entry.ID, 2); got != "Second version." {
		t.Errorf("revision 2 content = %q", got)
	}
	if got := revisionContent(t, svc, ctx, entry.ID, 1); got != "First version." {
		t.Errorf("revision 1 content = %q", got)
	}
	if revs[0].Summary == "" {
		t.Error("an update should describe itself in the history")
	}
}

func TestRevisions_NoOpUpdateWritesNothing(t *testing.T) {
	svc, _, ctx, research, section := revisionFixture(t)
	entry := mustEntry(t, svc, ctx, research.ID, section.ID, "Same text.")

	// The exact same content, twice more. An agent that re-writes identical text
	// must leave one revision behind, not three.
	for i := 0; i < 2; i++ {
		if _, err := svc.Update(ctx, entry.ID, UpdateEntryRequest{Content: ptr("Same text.")}); err != nil {
			t.Fatalf("update %d: %v", i, err)
		}
	}

	_, revs, _ := svc.History(ctx, entry.ID)
	if len(revs) != 1 {
		t.Fatalf("identical writes must not create revisions: got %d", len(revs))
	}
}

func TestRevisions_StatusChangeIsARevision(t *testing.T) {
	svc, _, ctx, research, section := revisionFixture(t)
	entry := mustEntry(t, svc, ctx, research.ID, section.ID, "Body.")

	status := domain.EntryCompleted
	if _, err := svc.Update(ctx, entry.ID, UpdateEntryRequest{Status: &status}); err != nil {
		t.Fatalf("update status: %v", err)
	}

	_, revs, _ := svc.History(ctx, entry.ID)
	if len(revs) != 2 {
		t.Fatalf("a status change is a change: want 2 revisions, got %d", len(revs))
	}
	if revs[0].Status != domain.EntryCompleted {
		t.Errorf("revision status = %q", revs[0].Status)
	}
}

func TestRevisions_Restore(t *testing.T) {
	svc, _, ctx, research, section := revisionFixture(t)
	entry := mustEntry(t, svc, ctx, research.ID, section.ID, "Original.")

	if _, err := svc.Update(ctx, entry.ID, UpdateEntryRequest{Content: ptr("Damaged.")}); err != nil {
		t.Fatalf("update: %v", err)
	}

	restored, err := svc.Restore(ctx, entry.ID, 1)
	if err != nil {
		t.Fatalf("restore: %v", err)
	}
	if restored.Content != "Original." {
		t.Errorf("restored content = %q, want the revision 1 text", restored.Content)
	}

	_, revs, _ := svc.History(ctx, entry.ID)
	if len(revs) != 3 {
		t.Fatalf("restoring must append, not rewrite: want 3 revisions, got %d", len(revs))
	}
	if revs[0].Revision != 3 {
		t.Errorf("newest revision = %d, want 3", revs[0].Revision)
	}
	if revs[0].AuthorKind != domain.AuthorRestore {
		t.Errorf("restore author = %q, want restore", revs[0].AuthorKind)
	}
	if !strings.Contains(revs[0].Summary, "Restored revision 1") {
		t.Errorf("restore summary = %q", revs[0].Summary)
	}
	// Revision 2 is still there: history is append-only.
	if got := revisionContent(t, svc, ctx, entry.ID, 2); got != "Damaged." {
		t.Errorf("revision 2 was rewritten: %q", got)
	}
}

func TestRevisions_AuthorKindFromContext(t *testing.T) {
	svc, _, ctx, research, section := revisionFixture(t)
	entry := mustEntry(t, svc, ctx, research.ID, section.ID, "Written by an agent.")

	human := WithAuthor(ctx, domain.AuthorHuman)
	if _, err := svc.Update(human, entry.ID, UpdateEntryRequest{Content: ptr("Corrected by a person.")}); err != nil {
		t.Fatalf("update: %v", err)
	}

	_, revs, _ := svc.History(ctx, entry.ID)
	if revs[0].AuthorKind != domain.AuthorHuman {
		t.Errorf("author = %q, want human", revs[0].AuthorKind)
	}
	if revs[1].AuthorKind != domain.AuthorAgent {
		t.Errorf("earlier author = %q, want agent", revs[1].AuthorKind)
	}
}

func TestRevisions_RecordsActiveSession(t *testing.T) {
	entrySvc, sessionSvc, ctx, research, section := revisionFixture(t)

	sess, _, err := sessionSvc.Create(ctx, CreateSessionRequest{
		ResearchID: research.ID, Title: "Deep dive", Focus: "pricing",
	})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}

	entry := mustEntry(t, entrySvc, ctx, research.ID, section.ID, "Findings.")

	_, revs, _ := entrySvc.History(ctx, entry.ID)
	if revs[0].SessionID != sess.ID {
		t.Errorf("revision session = %q, want the active session %q", revs[0].SessionID, sess.ID)
	}
	if revs[0].SessionCode != sess.Code {
		t.Errorf("session code was not resolved for display: %q", revs[0].SessionCode)
	}
}

func TestRevisions_SessionChanges(t *testing.T) {
	entrySvc, sessionSvc, ctx, research, section := revisionFixture(t)

	// An entry that predates the session.
	existing := mustEntry(t, entrySvc, ctx, research.ID, section.ID, "Old body.")

	sess, _, err := sessionSvc.Create(ctx, CreateSessionRequest{
		ResearchID: research.ID, Title: "Session two", Focus: "everything",
	})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}

	// One entry edited during the session, one created during it.
	if _, err := entrySvc.Update(ctx, existing.ID, UpdateEntryRequest{Content: ptr("Old body, now revised.")}); err != nil {
		t.Fatalf("update: %v", err)
	}
	fresh := mustEntry(t, entrySvc, ctx, research.ID, section.ID, "Brand new.")

	changes, err := entrySvc.SessionChanges(ctx, research.ID, sess.ID)
	if err != nil {
		t.Fatalf("session changes: %v", err)
	}
	if len(changes) != 2 {
		t.Fatalf("want 2 changed entries, got %d", len(changes))
	}

	byID := map[string]*SessionEntryChange{}
	for _, c := range changes {
		byID[c.EntryID] = c
	}

	edited := byID[existing.ID]
	if edited == nil || edited.Created {
		t.Fatalf("the pre-existing entry must be reported as modified, not created: %+v", edited)
	}
	if edited.Diff == nil || edited.Diff.Added == 0 {
		t.Errorf("a modified entry should carry a diff with additions: %+v", edited.Diff)
	}

	created := byID[fresh.ID]
	if created == nil || !created.Created {
		t.Fatalf("the new entry must be reported as created: %+v", created)
	}
}

// The tab badge reads its number from SessionChangeCounts, which skips every
// diff. If it ever disagrees with the list it labels, the badge is a lie about
// the screen underneath it — so the two are asserted against each other.
func TestRevisions_SessionChangeCountsMatchTheList(t *testing.T) {
	entrySvc, sessionSvc, ctx, research, section := revisionFixture(t)

	existing := mustEntry(t, entrySvc, ctx, research.ID, section.ID, "Old body.")

	sess, _, err := sessionSvc.Create(ctx, CreateSessionRequest{
		ResearchID: research.ID, Title: "Counting session", Focus: "everything",
	})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}

	if _, err := entrySvc.Update(ctx, existing.ID, UpdateEntryRequest{Content: ptr("Old body, now revised.")}); err != nil {
		t.Fatalf("update: %v", err)
	}
	mustEntry(t, entrySvc, ctx, research.ID, section.ID, "Brand new.")
	mustEntry(t, entrySvc, ctx, research.ID, section.ID, "Also brand new.")

	created, modified, err := entrySvc.SessionChangeCounts(ctx, research.ID, sess.ID)
	if err != nil {
		t.Fatalf("session change counts: %v", err)
	}
	if created != 2 || modified != 1 {
		t.Errorf("want 2 created and 1 modified, got %d and %d", created, modified)
	}

	changes, err := entrySvc.SessionChanges(ctx, research.ID, sess.ID)
	if err != nil {
		t.Fatalf("session changes: %v", err)
	}
	if created+modified != len(changes) {
		t.Errorf("count %d disagrees with the list it labels (%d cards)", created+modified, len(changes))
	}
	listCreated := 0
	for _, c := range changes {
		if c.Created {
			listCreated++
		}
	}
	if listCreated != created {
		t.Errorf("created count %d disagrees with the list's %d", created, listCreated)
	}
}

// A session in a research the caller cannot read must not leak its size either:
// the count is a smaller answer than the list, not a less protected one.
func TestRevisions_SessionChangeCountsRefuseAnotherResearch(t *testing.T) {
	entrySvc, sessionSvc, ctx, research, section := revisionFixture(t)
	mustEntry(t, entrySvc, ctx, research.ID, section.ID, "Body.")

	sess, _, err := sessionSvc.Create(ctx, CreateSessionRequest{
		ResearchID: research.ID, Title: "Scoped", Focus: "f",
	})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}

	if _, _, err := entrySvc.SessionChangeCounts(ctx, "00000000-0000-0000-0000-000000000000", sess.ID); err == nil {
		t.Error("counting a session through a research id that is not its own must fail")
	}
}

func TestRevisions_BlockPatchWritesRevisionButTickDoesNot(t *testing.T) {
	svc, ctx, entry := patchFixture(t)

	_, revs, err := svc.History(ctx, entry.ID)
	if err != nil {
		t.Fatalf("history: %v", err)
	}
	if len(revs) != 1 {
		t.Fatalf("a created block document has one revision, got %d", len(revs))
	}

	// A structural patch is an edit.
	if _, err := svc.PatchBlocks(ctx, entry.ID, PatchBlocksRequest{Ops: []BlockOp{{
		Op: OpInsert, Type: "paragraph", Data: map[string]any{"text": "One more note."}, At: "end",
	}}}); err != nil {
		t.Fatalf("patch insert: %v", err)
	}

	// Ticking a checkbox is not.
	checked := true
	if _, err := svc.PatchBlocks(ctx, entry.ID, PatchBlocksRequest{Ops: []BlockOp{{
		Op: OpSetState, ID: "cccc3333", Item: "k1", Checked: &checked,
	}}}); err != nil {
		t.Fatalf("patch tick: %v", err)
	}

	_, revs, _ = svc.History(ctx, entry.ID)
	if len(revs) != 2 {
		t.Fatalf("want 2 revisions (create + structural patch), got %d — a tick must not be a revision", len(revs))
	}
	if !strings.Contains(revs[0].Summary, "inserted") {
		t.Errorf("patch summary = %q, want it to name the op", revs[0].Summary)
	}
}

func TestLatestSnapshot_KeepsUnrevisionedChecklistState(t *testing.T) {
	svc, ctx, entry := patchFixture(t)

	checked := true
	if _, err := svc.PatchBlocks(ctx, entry.ID, PatchBlocksRequest{Ops: []BlockOp{{
		Op: OpSetState, ID: "cccc3333", Item: "k1", Checked: &checked,
	}}}); err != nil {
		t.Fatalf("tick: %v", err)
	}

	snapshot, rev, err := svc.LatestSnapshot(ctx, entry)
	if err != nil {
		t.Fatalf("latest snapshot: %v", err)
	}
	if rev == nil || rev.Revision != 1 {
		t.Fatalf("tick must keep revision 1, got %+v", rev)
	}
	doc, err := ParseStoredBlockDocument(snapshot.Content)
	if err != nil {
		t.Fatalf("parse snapshot: %v", err)
	}
	for _, block := range doc.Blocks {
		if block.ID == "cccc3333" && checklistItems(block.Data)[0].Checked {
			return
		}
	}
	t.Fatal("latest snapshot lost the current checklist tick")
}

func TestRevisions_RestoreKeepsBlockIDsAndTicks(t *testing.T) {
	svc, ctx, entry := patchFixture(t)

	// Tick an item, then damage the document.
	checked := true
	if _, err := svc.PatchBlocks(ctx, entry.ID, PatchBlocksRequest{Ops: []BlockOp{{
		Op: OpSetState, ID: "cccc3333", Item: "k1", Checked: &checked,
	}}}); err != nil {
		t.Fatalf("tick: %v", err)
	}
	if _, err := svc.PatchBlocks(ctx, entry.ID, PatchBlocksRequest{Ops: []BlockOp{{
		Op: OpDelete, ID: "bbbb2222",
	}}}); err != nil {
		t.Fatalf("delete block: %v", err)
	}

	if _, err := svc.Restore(ctx, entry.ID, 1); err != nil {
		t.Fatalf("restore: %v", err)
	}

	blocks := blocksOf(t, svc, ctx, entry.ID)
	ids := make([]string, 0, len(blocks))
	for _, b := range blocks {
		ids = append(ids, b.ID)
	}
	if len(blocks) != 3 {
		t.Fatalf("restore should bring the deleted block back: got %d blocks (%v)", len(blocks), ids)
	}
	for _, want := range []string{"aaaa1111", "bbbb2222", "cccc3333"} {
		found := false
		for _, id := range ids {
			if id == want {
				found = true
			}
		}
		if !found {
			t.Errorf("block id %q did not survive the restore; ids are %v", want, ids)
		}
	}

	// The tick belongs to the human, not to the document an agent wrote: a
	// restore of the prose must not silently untick their work.
	var ticked bool
	for _, b := range blocks {
		if b.ID != "cccc3333" {
			continue
		}
		state, _ := b.Data["state"].(map[string]any)
		if v, ok := state["k1"].(bool); ok && v {
			ticked = true
		}
	}
	if !ticked {
		t.Error("restoring the document dropped a checkbox a human had ticked")
	}
}

func TestRevisions_BlockDiffIsMarkdownNotJSON(t *testing.T) {
	svc, ctx, entry := patchFixture(t)

	if _, err := svc.PatchBlocks(ctx, entry.ID, PatchBlocksRequest{Ops: []BlockOp{{
		Op: OpUpdate, ID: "bbbb2222", Data: map[string]any{"text": "Read this first."},
	}}}); err != nil {
		t.Fatalf("patch: %v", err)
	}

	diff, err := svc.Diff(ctx, entry.ID, 0, 0)
	if err != nil {
		t.Fatalf("diff: %v", err)
	}
	if strings.Contains(diff.Content.Unified, `"type":`) || strings.Contains(diff.Content.Unified, "bbbb2222") {
		t.Errorf("a block diff must compare rendered markdown, not the stored JSON:\n%s", diff.Content.Unified)
	}
	if !strings.Contains(diff.Content.Unified, "Read this first.") {
		t.Errorf("diff does not show the new text:\n%s", diff.Content.Unified)
	}
	if diff.From == nil || diff.From.Revision != 1 || diff.To.Revision != 2 {
		t.Errorf("bare diff should compare the last two revisions, got %v -> %d", diff.From, diff.To.Revision)
	}
}

func TestRevisions_ImportIsAttributedToImport(t *testing.T) {
	svc, _, ctx, research, section := revisionFixture(t)

	imported := WithAuthor(ctx, domain.AuthorImport)
	entry, err := svc.Create(imported, CreateEntryRequest{
		ResearchID: research.ID,
		SectionID:  section.ID,
		Content:    "Came from a file.",
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	_, revs, _ := svc.History(ctx, entry.ID)
	if revs[0].AuthorKind != domain.AuthorImport {
		t.Errorf("author = %q, want import", revs[0].AuthorKind)
	}
}

func TestRevisions_TrimKeepsFirstAndNewest(t *testing.T) {
	svc, _, ctx, research, section := revisionFixture(t)
	svc.SetRevisionLimit(3)

	entry := mustEntry(t, svc, ctx, research.ID, section.ID, "v1")
	for i := 2; i <= 8; i++ {
		body := strings.Repeat("x", i)
		if _, err := svc.Update(ctx, entry.ID, UpdateEntryRequest{Content: &body}); err != nil {
			t.Fatalf("update %d: %v", i, err)
		}
	}

	_, revs, _ := svc.History(ctx, entry.ID)
	if len(revs) != 4 {
		t.Fatalf("limit 3 keeps the newest three plus revision 1: got %d", len(revs))
	}
	last := revs[len(revs)-1]
	if last.Revision != 1 {
		t.Fatalf("revision 1 must survive trimming, oldest kept is %d", last.Revision)
	}
	if got := revisionContent(t, svc, ctx, entry.ID, 1); got != "v1" {
		t.Errorf("revision 1 content = %q, want the text it was created with", got)
	}
	if revs[0].Revision != 8 {
		t.Errorf("newest revision = %d, want 8", revs[0].Revision)
	}
}

func TestRevisions_DeletingEntryDropsItsHistory(t *testing.T) {
	svc, _, ctx, research, section := revisionFixture(t)
	entry := mustEntry(t, svc, ctx, research.ID, section.ID, "Body.")

	if err := svc.Delete(ctx, entry.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, _, err := svc.History(ctx, entry.ID); err == nil {
		t.Error("history of a deleted entry must not be readable")
	}
}

// --- helpers ---

func mustEntry(t *testing.T, svc *EntryService, ctx context.Context, researchID, sectionID, content string) *domain.Entry {
	t.Helper()
	entry, err := svc.Create(ctx, CreateEntryRequest{
		ResearchID: researchID,
		SectionID:  sectionID,
		Title:      "Entry",
		Content:    content,
	})
	if err != nil {
		t.Fatalf("create entry: %v", err)
	}
	return entry
}

// revisionContent reads one revision's body. History deliberately omits content
// — a listing of a long document would otherwise be mostly copies of it — so a
// test that cares about the text asks for the revision itself.
func revisionContent(t *testing.T, svc *EntryService, ctx context.Context, entryID string, number int) string {
	t.Helper()
	rev, err := svc.Revision(ctx, entryID, number)
	if err != nil {
		t.Fatalf("read revision %d: %v", number, err)
	}
	return rev.Content
}

func TestRevisions_HistoryOmitsContent(t *testing.T) {
	svc, _, ctx, research, section := revisionFixture(t)
	entry := mustEntry(t, svc, ctx, research.ID, section.ID, "A body long enough to notice.")

	_, revs, _ := svc.History(ctx, entry.ID)
	if revs[0].Content != "" {
		t.Error("a history listing must not carry entry content")
	}
	if rev, err := svc.Revision(ctx, entry.ID, 1); err != nil || rev.Content == "" {
		t.Errorf("reading one revision must return its content (err=%v)", err)
	}
}

// Deleting an entry takes its revisions with it, so a session that wrote only
// entries which were later deleted reports nothing. This is the honest current
// behaviour, asserted so that a change to it is a decision rather than a
// surprise: keeping the history of a deleted entry means keeping the row alive
// past its entry, which is a schema change.
func TestRevisions_SessionChangesDropDeletedEntries(t *testing.T) {
	entrySvc, sessionSvc, ctx, research, section := revisionFixture(t)

	sess, _, err := sessionSvc.Create(ctx, CreateSessionRequest{
		ResearchID: research.ID, Title: "Session", Focus: "x",
	})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}

	kept := mustEntry(t, entrySvc, ctx, research.ID, section.ID, "Survives.")
	doomed := mustEntry(t, entrySvc, ctx, research.ID, section.ID, "Deleted later.")

	if err := entrySvc.Delete(ctx, doomed.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}

	changes, err := entrySvc.SessionChanges(ctx, research.ID, sess.ID)
	if err != nil {
		t.Fatalf("session changes: %v", err)
	}
	if len(changes) != 1 || changes[0].EntryID != kept.ID {
		t.Fatalf("want only the surviving entry, got %d changes", len(changes))
	}
}

// An entry that names its own session must have its first revision filed under
// that session, not under whichever one happens to be active. Import creates
// sessions first, applies their statuses, then creates entries with explicit
// session ids — so getting this wrong files an entire imported research under
// the one session that came back active.
func TestRevisions_CreateHonoursExplicitSession(t *testing.T) {
	entrySvc, sessionSvc, ctx, research, section := revisionFixture(t)

	first, _, err := sessionSvc.Create(ctx, CreateSessionRequest{
		ResearchID: research.ID, Title: "Earlier session", Focus: "a",
	})
	if err != nil {
		t.Fatalf("create first session: %v", err)
	}
	completed := domain.SessionCompleted
	if _, err := sessionSvc.Update(ctx, first.ID, UpdateSessionRequest{Status: &completed}); err != nil {
		t.Fatalf("complete first session: %v", err)
	}

	active, _, err := sessionSvc.Create(ctx, CreateSessionRequest{
		ResearchID: research.ID, Title: "Active session", Focus: "b",
	})
	if err != nil {
		t.Fatalf("create active session: %v", err)
	}

	entry, err := entrySvc.Create(ctx, CreateEntryRequest{
		ResearchID: research.ID,
		SectionID:  section.ID,
		SessionID:  first.ID, // explicit, and not the active one
		Content:    "Written under the earlier session.",
	})
	if err != nil {
		t.Fatalf("create entry: %v", err)
	}

	_, revs, _ := entrySvc.History(ctx, entry.ID)
	if revs[0].SessionID != first.ID {
		t.Fatalf("revision 1 filed under %q, want the entry's own session %q", revs[0].SessionID, first.ID)
	}

	byFirst, err := entrySvc.SessionChanges(ctx, research.ID, first.ID)
	if err != nil || len(byFirst) != 1 {
		t.Fatalf("the earlier session should own the entry: %d changes, err=%v", len(byFirst), err)
	}
	byActive, err := entrySvc.SessionChanges(ctx, research.ID, active.ID)
	if err != nil {
		t.Fatalf("session changes: %v", err)
	}
	if len(byActive) != 0 {
		t.Fatalf("the active session claimed %d entries it did not write", len(byActive))
	}
}

func TestLatestSnapshot_KeepsEntrySessionSeparateFromAuthoringSession(t *testing.T) {
	entrySvc, sessionSvc, ctx, research, section := revisionFixture(t)

	linked, _, err := sessionSvc.Create(ctx, CreateSessionRequest{
		ResearchID: research.ID, Title: "Linked session", Focus: "a",
	})
	if err != nil {
		t.Fatalf("create linked session: %v", err)
	}
	completed := domain.SessionCompleted
	if _, err := sessionSvc.Update(ctx, linked.ID, UpdateSessionRequest{Status: &completed}); err != nil {
		t.Fatalf("complete linked session: %v", err)
	}
	authoring, _, err := sessionSvc.Create(ctx, CreateSessionRequest{
		ResearchID: research.ID, Title: "Authoring session", Focus: "b",
	})
	if err != nil {
		t.Fatalf("create authoring session: %v", err)
	}

	entry, err := entrySvc.Create(ctx, CreateEntryRequest{
		ResearchID: research.ID, SectionID: section.ID, SessionID: linked.ID, Content: "v1",
	})
	if err != nil {
		t.Fatalf("create entry: %v", err)
	}
	if _, err := entrySvc.Update(ctx, entry.ID, UpdateEntryRequest{Content: ptr("v2")}); err != nil {
		t.Fatalf("update entry: %v", err)
	}

	snapshot, rev, err := entrySvc.LatestSnapshot(ctx, entry)
	if err != nil {
		t.Fatalf("latest snapshot: %v", err)
	}
	if snapshot.SessionID != linked.ID {
		t.Fatalf("entry linked session = %q, want %q", snapshot.SessionID, linked.ID)
	}
	if rev == nil || rev.SessionID != authoring.ID {
		t.Fatalf("revision authoring session = %+v, want %q", rev, authoring.ID)
	}
}

// A metadata-only write — the shape every status flip produces — records a
// revision whose content diff is empty. `fields` is what makes that revision
// legible instead of rendering as "nothing changed".
func TestRevisions_DiffReportsFieldChanges(t *testing.T) {
	svc, _, ctx, research, section := revisionFixture(t)
	entry := mustEntry(t, svc, ctx, research.ID, section.ID, "Body stays put.")

	status := domain.EntryActive
	if _, err := svc.Update(ctx, entry.ID, UpdateEntryRequest{
		Status: &status,
		Tags:   []string{"pricing"},
	}); err != nil {
		t.Fatalf("update: %v", err)
	}

	diff, err := svc.Diff(ctx, entry.ID, 0, 0)
	if err != nil {
		t.Fatalf("diff: %v", err)
	}
	if diff.Content.Added != 0 || diff.Content.Removed != 0 {
		t.Fatalf("the body did not change: got +%d −%d", diff.Content.Added, diff.Content.Removed)
	}

	byField := map[string][2]string{}
	for _, f := range diff.Fields {
		byField[f.Field] = [2]string{f.Before, f.After}
	}
	if got := byField["status"]; got != [2]string{"draft", "active"} {
		t.Errorf("status change = %v, want draft→active", got)
	}
	if got := byField["tags"]; got != [2]string{"", "pricing"} {
		t.Errorf("tags change = %v", got)
	}

	// Revision 1 has no predecessor: reporting every field as changed-from-empty
	// would describe the entry's birth as an edit.
	first, err := svc.Diff(ctx, entry.ID, 0, 1)
	if err != nil {
		t.Fatalf("diff r1: %v", err)
	}
	if first.From != nil {
		t.Fatalf("revision 1 must have no base, got r%d", first.From.Revision)
	}
	if len(first.Fields) != 0 {
		t.Errorf("revision 1 reported %d field changes; it should report none", len(first.Fields))
	}
}
