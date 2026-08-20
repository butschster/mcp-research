package service

import (
	"context"
	"log/slog"
	"testing"

	"github.com/butschster/mcp-research/internal/domain"
	"github.com/butschster/mcp-research/internal/storage"
)

func annotationFixture(t *testing.T) (*AnnotationService, *EntryService, context.Context, *domain.Research, *domain.Section) {
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
	annotationRepo := storage.NewAnnotationRepository(db)

	researchSvc := NewResearchService(researchRepo, sectionRepo, storage.NewTeamRepository(db), testAccess(db), &mockNotifier{}, log)
	entrySvc := NewEntryService(entryRepo, sectionRepo, researchRepo, testAccess(db), sessionRepo, blockRepo, revisionRepo, crossrefRepo, nil, &mockNotifier{}, log)
	entrySvc.SetAnnotations(annotationRepo)
	annSvc := NewAnnotationService(annotationRepo, entryRepo, revisionRepo, testAccess(db), entrySvc, entrySvc, &mockNotifier{}, log)

	ctx := context.Background()
	research, sections, err := researchSvc.Create(ctx, CreateResearchRequest{
		Name: "R", Goal: "T",
		Sections: []CreateSectionRequest{{Name: "s1", DisplayName: "S1"}},
	})
	if err != nil {
		t.Fatalf("create research: %v", err)
	}
	return annSvc, entrySvc, ctx, research, sections[0]
}

// blocksEntry writes a two-paragraph block document and returns it with the
// block ids the server minted.
func blocksEntry(t *testing.T, svc *EntryService, ctx context.Context, research *domain.Research, section *domain.Section, texts ...string) (*domain.Entry, []string) {
	t.Helper()
	doc := `{"blocks":[`
	for i, text := range texts {
		if i > 0 {
			doc += ","
		}
		doc += `{"type":"paragraph","data":{"text":"` + text + `"}}`
	}
	doc += `]}`

	entry, err := svc.Create(ctx, CreateEntryRequest{
		ResearchID: research.ID, SectionID: section.ID,
		Title: "Doc", Type: domain.EntryBlocks, Content: doc,
	})
	if err != nil {
		t.Fatalf("create entry: %v", err)
	}
	loaded, err := svc.LoadBlockDocument(ctx, entry)
	if err != nil {
		t.Fatalf("load document: %v", err)
	}
	ids := make([]string, 0, len(loaded.Blocks))
	for _, b := range loaded.Blocks {
		ids = append(ids, b.ID)
	}
	return entry, ids
}

func TestAnnotation_CreateAnchorsToBlock(t *testing.T) {
	svc, entrySvc, ctx, research, section := annotationFixture(t)
	entry, ids := blocksEntry(t, entrySvc, ctx, research, section, "Costs fall by 40 percent.", "Adoption is slow.")

	a, err := svc.Create(ctx, CreateAnnotationRequest{
		EntryID: entry.ID, BlockID: ids[0],
		Quote: domain.Quote{Exact: "fall by 40 percent"},
		Kind:  domain.AnnotationVerify, Body: "where is this from?",
	})
	if err != nil {
		t.Fatalf("create annotation: %v", err)
	}
	if a.Code != "A1" {
		t.Errorf("code = %q, want A1", a.Code)
	}
	if a.Status != domain.AnnotationOpen {
		t.Errorf("status = %q, want open", a.Status)
	}
	if a.Anchor == nil || a.Anchor.State != domain.AnchorAnchored {
		t.Fatalf("anchor = %+v, want anchored", a.Anchor)
	}
	if a.Anchor.Strategy != domain.AnchorByBlockID || a.Anchor.Confidence != 1 {
		t.Errorf("strategy/confidence = %q/%v, want block_id/1", a.Anchor.Strategy, a.Anchor.Confidence)
	}
}

func TestAnnotation_RejectsUnknownKind(t *testing.T) {
	svc, entrySvc, ctx, research, section := annotationFixture(t)
	entry, ids := blocksEntry(t, entrySvc, ctx, research, section, "Costs fall.")

	_, err := svc.Create(ctx, CreateAnnotationRequest{
		EntryID: entry.ID, BlockID: ids[0],
		Quote: domain.Quote{Exact: "Costs fall"}, Kind: "nitpick",
	})
	if err == nil {
		t.Fatal("want refusal for an unknown kind")
	}
}

// A markdown entry has no blocks to pin to, and saying otherwise would be a
// promise the document cannot keep.
func TestAnnotation_MarkdownEntryDropsBlockID(t *testing.T) {
	svc, entrySvc, ctx, research, section := annotationFixture(t)
	entry, err := entrySvc.Create(ctx, CreateEntryRequest{
		ResearchID: research.ID, SectionID: section.ID,
		Title: "Notes", Content: "Pricing is seat based above 50 users.",
	})
	if err != nil {
		t.Fatalf("create entry: %v", err)
	}

	a, err := svc.Create(ctx, CreateAnnotationRequest{
		EntryID: entry.ID, BlockID: "invented",
		Quote: domain.Quote{Exact: "seat based"}, Kind: domain.AnnotationDig,
	})
	if err != nil {
		t.Fatalf("create annotation: %v", err)
	}
	if a.BlockID != "" {
		t.Errorf("block_id = %q, want empty on a markdown entry", a.BlockID)
	}
	if a.Anchor == nil || a.Anchor.State != domain.AnchorAnchored {
		t.Fatalf("anchor = %+v, want anchored by quote", a.Anchor)
	}
}

// The rule the whole feature turns on: an agent may finish work, never accept
// it.
func TestAnnotation_AgentCannotClose(t *testing.T) {
	svc, entrySvc, ctx, research, section := annotationFixture(t)
	entry, ids := blocksEntry(t, entrySvc, ctx, research, section, "Costs fall by 40 percent.")

	a, err := svc.Create(ctx, CreateAnnotationRequest{
		EntryID: entry.ID, BlockID: ids[0],
		Quote: domain.Quote{Exact: "40 percent"}, Kind: domain.AnnotationVerify,
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	closed := domain.AnnotationClosed
	agentCtx := WithAuthor(ctx, domain.AuthorAgent)
	if _, err := svc.Update(agentCtx, a.ID, UpdateAnnotationRequest{Status: &closed}); err == nil {
		t.Fatal("an agent must not be able to close a mark")
	}
	if _, err := svc.Update(agentCtx, a.ID, UpdateAnnotationRequest{Status: ptr(domain.AnnotationDismissed)}); err == nil {
		t.Fatal("an agent must not be able to dismiss a mark")
	}

	humanCtx := WithAuthor(ctx, domain.AuthorHuman)
	got, err := svc.Update(humanCtx, a.ID, UpdateAnnotationRequest{Status: &closed})
	if err != nil {
		t.Fatalf("a person must be able to close: %v", err)
	}
	if got.Status != domain.AnnotationClosed || got.ClosedAt == nil {
		t.Errorf("status = %q closedAt = %v, want closed and stamped", got.Status, got.ClosedAt)
	}
}

func TestAnnotation_AnswerStopsAtAnswered(t *testing.T) {
	svc, entrySvc, ctx, research, section := annotationFixture(t)
	entry, ids := blocksEntry(t, entrySvc, ctx, research, section, "Costs fall by 40 percent.")

	a, _ := svc.Create(ctx, CreateAnnotationRequest{
		EntryID: entry.ID, BlockID: ids[0],
		Quote: domain.Quote{Exact: "40 percent"}, Kind: domain.AnnotationVerify,
	})

	answered, err := svc.Answer(WithAuthor(ctx, domain.AuthorAgent), a.ID, AnswerAnnotationRequest{
		Resolution: "vendor page, not measured. Recorded as unverified.",
	})
	if err != nil {
		t.Fatalf("answer: %v", err)
	}
	if answered.Status != domain.AnnotationAnswered {
		t.Errorf("status = %q, want answered", answered.Status)
	}
	if answered.AnsweredAt == nil {
		t.Error("answered_at not stamped")
	}

	if _, err := svc.Answer(ctx, a.ID, AnswerAnnotationRequest{}); err == nil {
		t.Error("an answer with no resolution must be refused")
	}
}

// A rejection must not overwrite the answer it rejects: the next attempt needs
// both.
func TestAnnotation_RejectionKeepsTheAnswer(t *testing.T) {
	svc, entrySvc, ctx, research, section := annotationFixture(t)
	entry, ids := blocksEntry(t, entrySvc, ctx, research, section, "Costs fall by 40 percent.")

	a, _ := svc.Create(ctx, CreateAnnotationRequest{
		EntryID: entry.ID, BlockID: ids[0],
		Quote: domain.Quote{Exact: "40 percent"}, Kind: domain.AnnotationVerify,
	})
	if _, err := svc.Answer(ctx, a.ID, AnswerAnnotationRequest{Resolution: "found it on their blog"}); err != nil {
		t.Fatalf("answer: %v", err)
	}

	reopened, err := svc.Update(WithAuthor(ctx, domain.AuthorHuman), a.ID, UpdateAnnotationRequest{
		Status: ptr(domain.AnnotationOpen),
		Reason: ptr("a vendor blog is not a measurement"),
	})
	if err != nil {
		t.Fatalf("reject: %v", err)
	}
	if reopened.Attempts != 1 {
		t.Errorf("attempts = %d, want 1", reopened.Attempts)
	}
	if reopened.Resolution != "found it on their blog" {
		t.Errorf("resolution = %q, want the refused answer kept", reopened.Resolution)
	}
	if len(reopened.Rejections) != 1 || reopened.Rejections[0].Reason != "a vendor blog is not a measurement" {
		t.Fatalf("rejections = %+v, want the reason recorded beside it", reopened.Rejections)
	}
}

func TestAnnotation_BulkReportsPerRow(t *testing.T) {
	svc, entrySvc, ctx, research, section := annotationFixture(t)
	entry, ids := blocksEntry(t, entrySvc, ctx, research, section, "Costs fall by 40 percent.")

	a, _ := svc.Create(ctx, CreateAnnotationRequest{
		EntryID: entry.ID, BlockID: ids[0],
		Quote: domain.Quote{Exact: "40 percent"}, Kind: domain.AnnotationVerify,
	})

	results, err := svc.Bulk(WithAuthor(ctx, domain.AuthorHuman), research.ID,
		[]string{a.ID, "no-such-id"}, domain.AnnotationClosed, "")
	if err != nil {
		t.Fatalf("bulk: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("results = %d, want one per id", len(results))
	}
	if !results[0].OK || results[0].Code != "A1" {
		t.Errorf("first = %+v, want ok", results[0])
	}
	// A partial failure has to be visible: "one of two" is neither success nor
	// failure, and the screen must be able to say which row missed.
	if results[1].OK || results[1].Error == "" {
		t.Errorf("second = %+v, want a reported failure", results[1])
	}
}

func TestAnnotation_QueueIsScopedToTheResearch(t *testing.T) {
	svc, entrySvc, ctx, research, section := annotationFixture(t)
	entry, ids := blocksEntry(t, entrySvc, ctx, research, section, "Costs fall.", "Adoption is slow.")

	for _, q := range []string{"Costs fall", "Adoption is slow"} {
		if _, err := svc.Create(ctx, CreateAnnotationRequest{
			EntryID: entry.ID, BlockID: ids[0],
			Quote: domain.Quote{Exact: q}, Kind: domain.AnnotationDig,
		}); err != nil {
			t.Fatalf("create %q: %v", q, err)
		}
	}

	open := domain.AnnotationOpen
	got, err := svc.ListByResearch(ctx, research.ID, storage.AnnotationFilter{Status: &open})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("listed %d, want 2", len(got))
	}
	for _, a := range got {
		if a.EntryCode == "" || a.EntryTitle == "" {
			t.Errorf("%s carries no entry identity: %+v", a.Code, a)
		}
		if a.EntryType != domain.EntryBlocks {
			t.Errorf("%s entry_type = %q, want blocks", a.Code, a.EntryType)
		}
	}
}

// The server caps a batch rather than trusting a caller to. A prompt limit is
// what a model ignores under load.
func TestAnnotation_BatchIsCappedByTheServer(t *testing.T) {
	svc, entrySvc, ctx, research, section := annotationFixture(t)
	entry, ids := blocksEntry(t, entrySvc, ctx, research, section, "One two three four five six seven eight.")

	for i := 0; i < domain.MaxAnnotationBatch+3; i++ {
		if _, err := svc.Create(ctx, CreateAnnotationRequest{
			EntryID: entry.ID, BlockID: ids[0],
			Quote: domain.Quote{Exact: "One two three"}, Kind: domain.AnnotationVerify,
		}); err != nil {
			t.Fatalf("create %d: %v", i, err)
		}
	}

	got, err := svc.ListByResearch(ctx, research.ID, storage.AnnotationFilter{Limit: 500})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(got) > domain.MaxAnnotationBatch {
		t.Errorf("returned %d, want at most %d", len(got), domain.MaxAnnotationBatch)
	}
}
