package storage

import (
	"context"
	"testing"

	"github.com/butschster/mcp-research/internal/domain"
	"github.com/google/uuid"
)

func TestCrossRefRepository_ReplaceAndFind(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	researchRepo := NewResearchRepository(db)
	sectionRepo := NewSectionRepository(db)
	entryRepo := NewEntryRepository(db)
	crossrefRepo := NewCrossRefRepository(db)

	r := &domain.Research{ID: uuid.New().String(), Name: "Test", Status: domain.ResearchActive, Memory: []string{}, Tags: []string{}}
	researchRepo.Create(ctx, r)

	s := &domain.Section{ID: uuid.New().String(), ResearchID: r.ID, Name: "s1", Status: domain.SectionDraft}
	sectionRepo.Create(ctx, s)

	e1 := &domain.Entry{ID: uuid.New().String(), ResearchID: r.ID, SectionID: s.ID, Content: "entry 1", Status: domain.EntryDraft, Tags: []string{}}
	entryRepo.Create(ctx, e1)

	e2 := &domain.Entry{ID: uuid.New().String(), ResearchID: r.ID, SectionID: s.ID, Content: "entry 2", Status: domain.EntryDraft, Tags: []string{}}
	entryRepo.Create(ctx, e2)

	// Insert crossrefs from entry source
	refs := []domain.CrossRef{
		{SourceType: "entry", SourceID: e1.ID, SourceResearchID: r.ID, TargetEntryID: e2.ID, TargetResearchID: r.ID, TargetRef: "E2", Resolved: true},
	}
	if err := crossrefRepo.ReplaceForSource(ctx, "entry", e1.ID, refs); err != nil {
		t.Fatal(err)
	}

	// Find by research
	found, err := crossrefRepo.FindByResearch(ctx, r.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(found) != 1 {
		t.Fatalf("FindByResearch: got %d, want 1", len(found))
	}
	if found[0].SourceType != "entry" {
		t.Errorf("SourceType: got %q, want entry", found[0].SourceType)
	}
	if found[0].TargetRef != "E2" {
		t.Errorf("TargetRef: got %q, want E2", found[0].TargetRef)
	}
	if !found[0].Resolved {
		t.Error("should be resolved")
	}

	// Find by target entry
	incoming, err := crossrefRepo.FindByTargetEntry(ctx, e2.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(incoming) != 1 {
		t.Fatalf("FindByTargetEntry: got %d, want 1", len(incoming))
	}

	// Replace clears old refs
	if err := crossrefRepo.ReplaceForSource(ctx, "entry", e1.ID, nil); err != nil {
		t.Fatal(err)
	}
	found, _ = crossrefRepo.FindByResearch(ctx, r.ID)
	if len(found) != 0 {
		t.Errorf("after replace with nil: got %d, want 0", len(found))
	}
}

func TestCrossRefRepository_QuestionSource(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	researchRepo := NewResearchRepository(db)
	sectionRepo := NewSectionRepository(db)
	entryRepo := NewEntryRepository(db)
	crossrefRepo := NewCrossRefRepository(db)

	r := &domain.Research{ID: uuid.New().String(), Name: "Test", Status: domain.ResearchActive, Memory: []string{}, Tags: []string{}}
	researchRepo.Create(ctx, r)

	s := &domain.Section{ID: uuid.New().String(), ResearchID: r.ID, Name: "s1", Status: domain.SectionDraft}
	sectionRepo.Create(ctx, s)

	e1 := &domain.Entry{ID: uuid.New().String(), ResearchID: r.ID, SectionID: s.ID, Content: "entry", Status: domain.EntryDraft, Tags: []string{}}
	entryRepo.Create(ctx, e1)

	questionID := uuid.New().String()

	// Insert crossref from question source (source_entry_id will be NULL)
	refs := []domain.CrossRef{
		{SourceType: "question", SourceID: questionID, SourceResearchID: r.ID, TargetEntryID: e1.ID, TargetResearchID: r.ID, TargetRef: "E1", Resolved: true},
	}
	if err := crossrefRepo.ReplaceForSource(ctx, "question", questionID, refs); err != nil {
		t.Fatal(err)
	}

	found, err := crossrefRepo.FindByResearch(ctx, r.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(found) != 1 {
		t.Fatalf("got %d crossrefs, want 1", len(found))
	}
	if found[0].SourceType != "question" {
		t.Errorf("SourceType: got %q, want question", found[0].SourceType)
	}
}
