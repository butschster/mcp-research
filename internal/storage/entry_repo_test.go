package storage

import (
	"context"
	"testing"

	"github.com/butschster/mcp-research/internal/domain"
	"github.com/google/uuid"
)

type entryTestFixture struct {
	db          *ResearchRepository
	sectionRepo *SectionRepository
	entryRepo   *EntryRepository
	research    *domain.Research
	section     *domain.Section
}

func setupEntryTest(t *testing.T) (*entryTestFixture, context.Context) {
	t.Helper()
	db := setupTestDB(t)
	ctx := context.Background()
	researchRepo := NewResearchRepository(db)
	research := createTestResearch(t, researchRepo, ctx)
	sectionRepo := NewSectionRepository(db)

	section := &domain.Section{
		ID:          uuid.New().String(),
		ResearchID:  research.ID,
		Name:        "test-section",
		DisplayName: "Test Section",
		Status:      domain.SectionActive,
		Position:    0,
	}
	if err := sectionRepo.Create(ctx, section); err != nil {
		t.Fatalf("create test section: %v", err)
	}

	return &entryTestFixture{
		db:          researchRepo,
		sectionRepo: sectionRepo,
		entryRepo:   NewEntryRepository(db),
		research:    research,
		section:     section,
	}, ctx
}

func TestEntryRepository_CreateAndFindByID(t *testing.T) {
	f, ctx := setupEntryTest(t)

	e1 := &domain.Entry{
		ID:          uuid.New().String(),
		ResearchID:  f.research.ID,
		SectionID:   f.section.ID,
		Title:       "First Entry",
		Content:     "First content",
		Description: "First description",
		Status:      domain.EntryDraft,
		Tags:        []string{"go"},
	}
	if err := f.entryRepo.Create(ctx, e1); err != nil {
		t.Fatalf("Create e1: %v", err)
	}
	if e1.Code != "E1" {
		t.Errorf("first code: got %s, want E1", e1.Code)
	}
	if e1.CreatedAt.IsZero() {
		t.Error("CreatedAt should not be zero after Create")
	}

	e2 := &domain.Entry{
		ID:         uuid.New().String(),
		ResearchID: f.research.ID,
		SectionID:  f.section.ID,
		Title:      "Second Entry",
		Content:    "Second content",
		Status:     domain.EntryDraft,
	}
	if err := f.entryRepo.Create(ctx, e2); err != nil {
		t.Fatalf("Create e2: %v", err)
	}
	if e2.Code != "E2" {
		t.Errorf("second code: got %s, want E2", e2.Code)
	}

	found, err := f.entryRepo.FindByID(ctx, e1.ID)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if found == nil {
		t.Fatal("FindByID returned nil")
	}
	if found.Title != "First Entry" {
		t.Errorf("Title: got %s, want 'First Entry'", found.Title)
	}
	if found.Content != "First content" {
		t.Errorf("Content: got %s, want 'First content'", found.Content)
	}
}

func TestEntryRepository_Update(t *testing.T) {
	f, ctx := setupEntryTest(t)

	e := &domain.Entry{
		ID:         uuid.New().String(),
		ResearchID: f.research.ID,
		SectionID:  f.section.ID,
		Title:      "Original",
		Content:    "Original content",
		Status:     domain.EntryDraft,
		Tags:       []string{"old"},
	}
	if err := f.entryRepo.Create(ctx, e); err != nil {
		t.Fatalf("Create: %v", err)
	}

	e.Title = "Updated"
	e.Content = "Updated content"
	e.Status = domain.EntryActive
	e.Tags = []string{"new", "updated"}

	if err := f.entryRepo.Update(ctx, e); err != nil {
		t.Fatalf("Update: %v", err)
	}

	found, err := f.entryRepo.FindByID(ctx, e.ID)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if found.Title != "Updated" {
		t.Errorf("Title: got %s, want Updated", found.Title)
	}
	if found.Content != "Updated content" {
		t.Errorf("Content: got %s, want 'Updated content'", found.Content)
	}
	if found.Status != domain.EntryActive {
		t.Errorf("Status: got %s, want active", found.Status)
	}
	if len(found.Tags) != 2 {
		t.Fatalf("Tags length: got %d, want 2", len(found.Tags))
	}
}

func TestEntryRepository_FindByCode(t *testing.T) {
	f, ctx := setupEntryTest(t)

	e := &domain.Entry{
		ID:         uuid.New().String(),
		ResearchID: f.research.ID,
		SectionID:  f.section.ID,
		Title:      "By Code",
		Content:    "content",
		Status:     domain.EntryDraft,
	}
	if err := f.entryRepo.Create(ctx, e); err != nil {
		t.Fatalf("Create: %v", err)
	}

	found, err := f.entryRepo.FindByCode(ctx, f.research.ID, e.Code)
	if err != nil {
		t.Fatalf("FindByCode: %v", err)
	}
	if found == nil {
		t.Fatal("FindByCode returned nil")
	}
	if found.ID != e.ID {
		t.Errorf("ID: got %s, want %s", found.ID, e.ID)
	}
}

func TestEntryRepository_FindBySection(t *testing.T) {
	f, ctx := setupEntryTest(t)

	entries := []*domain.Entry{
		{ID: uuid.New().String(), ResearchID: f.research.ID, SectionID: f.section.ID, Title: "Draft", Content: "c1", Status: domain.EntryDraft},
		{ID: uuid.New().String(), ResearchID: f.research.ID, SectionID: f.section.ID, Title: "Active", Content: "c2", Status: domain.EntryActive},
	}
	for _, e := range entries {
		if err := f.entryRepo.Create(ctx, e); err != nil {
			t.Fatalf("Create: %v", err)
		}
	}

	t.Run("no filter", func(t *testing.T) {
		found, err := f.entryRepo.FindBySection(ctx, f.research.ID, f.section.ID, EntryFilter{})
		if err != nil {
			t.Fatalf("FindBySection: %v", err)
		}
		if len(found) != 2 {
			t.Errorf("expected 2 entries, got %d", len(found))
		}
		// Content should be empty (no content returned)
		for _, e := range found {
			if e.Content != "" {
				t.Errorf("Content should be empty in FindBySection, got %s", e.Content)
			}
		}
	})

	t.Run("filter by status", func(t *testing.T) {
		status := domain.EntryActive
		found, err := f.entryRepo.FindBySection(ctx, f.research.ID, f.section.ID, EntryFilter{Status: &status})
		if err != nil {
			t.Fatalf("FindBySection with filter: %v", err)
		}
		if len(found) != 1 {
			t.Fatalf("expected 1 entry, got %d", len(found))
		}
		if found[0].Title != "Active" {
			t.Errorf("Title: got %s, want Active", found[0].Title)
		}
	})
}

func TestEntryRepository_FindByResearch(t *testing.T) {
	f, ctx := setupEntryTest(t)

	entries := []*domain.Entry{
		{ID: uuid.New().String(), ResearchID: f.research.ID, SectionID: f.section.ID, Title: "E1", Content: "c1", Status: domain.EntryDraft},
		{ID: uuid.New().String(), ResearchID: f.research.ID, SectionID: f.section.ID, Title: "E2", Content: "c2", Status: domain.EntryActive},
		{ID: uuid.New().String(), ResearchID: f.research.ID, SectionID: f.section.ID, Title: "E3", Content: "c3", Status: domain.EntryDraft},
	}
	for _, e := range entries {
		if err := f.entryRepo.Create(ctx, e); err != nil {
			t.Fatalf("Create: %v", err)
		}
	}

	t.Run("no filter", func(t *testing.T) {
		found, err := f.entryRepo.FindByResearch(ctx, f.research.ID, EntryFilter{})
		if err != nil {
			t.Fatalf("FindByResearch: %v", err)
		}
		if len(found) != 3 {
			t.Errorf("expected 3 entries, got %d", len(found))
		}
		for _, e := range found {
			if e.Content != "" {
				t.Errorf("Content should be empty in FindByResearch, got %s", e.Content)
			}
		}
	})

	t.Run("filter by status", func(t *testing.T) {
		status := domain.EntryDraft
		found, err := f.entryRepo.FindByResearch(ctx, f.research.ID, EntryFilter{Status: &status})
		if err != nil {
			t.Fatalf("FindByResearch with filter: %v", err)
		}
		if len(found) != 2 {
			t.Errorf("expected 2 draft entries, got %d", len(found))
		}
	})
}

func TestEntryRepository_CountBySection(t *testing.T) {
	f, ctx := setupEntryTest(t)

	count, err := f.entryRepo.CountBySection(ctx, f.section.ID)
	if err != nil {
		t.Fatalf("CountBySection: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0, got %d", count)
	}

	e := &domain.Entry{
		ID:         uuid.New().String(),
		ResearchID: f.research.ID,
		SectionID:  f.section.ID,
		Title:      "Count Test",
		Content:    "content",
		Status:     domain.EntryDraft,
	}
	if err := f.entryRepo.Create(ctx, e); err != nil {
		t.Fatalf("Create: %v", err)
	}

	count, err = f.entryRepo.CountBySection(ctx, f.section.ID)
	if err != nil {
		t.Fatalf("CountBySection: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1, got %d", count)
	}
}

func TestEntryRepository_FindByID_NotFound(t *testing.T) {
	f, ctx := setupEntryTest(t)

	found, err := f.entryRepo.FindByID(ctx, "nonexistent-id")
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if found != nil {
		t.Error("expected nil for non-existent record")
	}
}
