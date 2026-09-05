package storage

import (
	"context"
	"testing"

	"github.com/butschster/mcp-research/internal/domain"
	"github.com/google/uuid"
)

func TestNextCodeGlobal(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()
	repo := NewResearchRepository(db)

	// First research gets R1
	r1 := &domain.Research{ID: uuid.New().String(), Name: "R1", Status: domain.ResearchActive, Memory: domain.Memory{}, Tags: []string{}}
	if err := repo.Create(ctx, r1); err != nil {
		t.Fatal(err)
	}
	if r1.Code != "R1" {
		t.Errorf("first research code: got %q, want R1", r1.Code)
	}

	// Second research gets R2
	r2 := &domain.Research{ID: uuid.New().String(), Name: "R2", Status: domain.ResearchActive, Memory: domain.Memory{}, Tags: []string{}}
	if err := repo.Create(ctx, r2); err != nil {
		t.Fatal(err)
	}
	if r2.Code != "R2" {
		t.Errorf("second research code: got %q, want R2", r2.Code)
	}
}

func TestNextCodeScoped(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	researchRepo := NewResearchRepository(db)
	sectionRepo := NewSectionRepository(db)
	entryRepo := NewEntryRepository(db)

	r := &domain.Research{ID: uuid.New().String(), Name: "Test", Status: domain.ResearchActive, Memory: domain.Memory{}, Tags: []string{}}
	if err := researchRepo.Create(ctx, r); err != nil {
		t.Fatal(err)
	}

	// Sections get S1, S2 scoped to research
	s1 := &domain.Section{ID: uuid.New().String(), ResearchID: r.ID, Name: "s1", Status: domain.SectionDraft}
	if err := sectionRepo.Create(ctx, s1); err != nil {
		t.Fatal(err)
	}
	if s1.Code != "S1" {
		t.Errorf("first section code: got %q, want S1", s1.Code)
	}

	s2 := &domain.Section{ID: uuid.New().String(), ResearchID: r.ID, Name: "s2", Status: domain.SectionDraft}
	if err := sectionRepo.Create(ctx, s2); err != nil {
		t.Fatal(err)
	}
	if s2.Code != "S2" {
		t.Errorf("second section code: got %q, want S2", s2.Code)
	}

	// Entries get E1, E2 scoped to research
	e1 := &domain.Entry{ID: uuid.New().String(), ResearchID: r.ID, SectionID: s1.ID, Content: "test", Status: domain.EntryDraft, Tags: []string{}}
	if err := entryRepo.Create(ctx, e1); err != nil {
		t.Fatal(err)
	}
	if e1.Code != "E1" {
		t.Errorf("first entry code: got %q, want E1", e1.Code)
	}

	e2 := &domain.Entry{ID: uuid.New().String(), ResearchID: r.ID, SectionID: s1.ID, Content: "test2", Status: domain.EntryDraft, Tags: []string{}}
	if err := entryRepo.Create(ctx, e2); err != nil {
		t.Fatal(err)
	}
	if e2.Code != "E2" {
		t.Errorf("second entry code: got %q, want E2", e2.Code)
	}

	// Entries in different research start at E1 again
	r2 := &domain.Research{ID: uuid.New().String(), Name: "Other", Status: domain.ResearchActive, Memory: domain.Memory{}, Tags: []string{}}
	if err := researchRepo.Create(ctx, r2); err != nil {
		t.Fatal(err)
	}
	s3 := &domain.Section{ID: uuid.New().String(), ResearchID: r2.ID, Name: "s1", Status: domain.SectionDraft}
	if err := sectionRepo.Create(ctx, s3); err != nil {
		t.Fatal(err)
	}
	e3 := &domain.Entry{ID: uuid.New().String(), ResearchID: r2.ID, SectionID: s3.ID, Content: "other", Status: domain.EntryDraft, Tags: []string{}}
	if err := entryRepo.Create(ctx, e3); err != nil {
		t.Fatal(err)
	}
	if e3.Code != "E1" {
		t.Errorf("entry in different research: got %q, want E1", e3.Code)
	}
}

func TestBackfillCodes(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	// Insert research with empty code directly
	_, err := db.ExecContext(ctx, `INSERT INTO researches (id, code, name, status, tags) VALUES (?, '', 'No Code', 'active', '[]')`, uuid.New().String())
	if err != nil {
		t.Fatal(err)
	}

	n, err := BackfillCodes(ctx, db)
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Errorf("backfilled: got %d, want 1", n)
	}

	// Running again should backfill 0
	n, err = BackfillCodes(ctx, db)
	if err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Errorf("second backfill: got %d, want 0", n)
	}
}

func TestFindByCode(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()
	repo := NewResearchRepository(db)

	r := &domain.Research{ID: uuid.New().String(), Name: "Findable", Status: domain.ResearchActive, Memory: domain.Memory{}, Tags: []string{}}
	if err := repo.Create(ctx, r); err != nil {
		t.Fatal(err)
	}

	found, err := repo.FindByCode(ctx, r.Code)
	if err != nil {
		t.Fatal(err)
	}
	if found == nil {
		t.Fatal("FindByCode returned nil")
	}
	if found.ID != r.ID {
		t.Errorf("FindByCode ID: got %s, want %s", found.ID, r.ID)
	}

	// Not found
	notFound, err := repo.FindByCode(ctx, "R999")
	if err != nil {
		t.Fatal(err)
	}
	if notFound != nil {
		t.Error("FindByCode should return nil for unknown code")
	}
}
