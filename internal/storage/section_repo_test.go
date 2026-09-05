package storage

import (
	"context"
	"testing"
	"time"

	"github.com/dovod-app/app/internal/domain"
	"github.com/google/uuid"
)

func createTestResearch(t *testing.T, repo *ResearchRepository, ctx context.Context) *domain.Research {
	t.Helper()
	r := &domain.Research{
		ID:     uuid.New().String(),
		Name:   "Test Research",
		Status: domain.ResearchActive,
	}
	if err := repo.Create(ctx, r); err != nil {
		t.Fatalf("create test research: %v", err)
	}
	return r
}

func TestSectionRepository_CreateAndFindByID(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()
	researchRepo := NewResearchRepository(db)
	research := createTestResearch(t, researchRepo, ctx)
	repo := NewSectionRepository(db)

	s := &domain.Section{
		ID:          uuid.New().String(),
		ResearchID:  research.ID,
		Name:        "methodology",
		DisplayName: "Methodology",
		Description: "Research methodology section",
		Status:      domain.SectionActive,
		Position:    1,
	}

	if err := repo.Create(ctx, s); err != nil {
		t.Fatalf("Create: %v", err)
	}

	if s.CreatedAt.IsZero() {
		t.Error("CreatedAt should not be zero after Create")
	}
	if s.UpdatedAt.IsZero() {
		t.Error("UpdatedAt should not be zero after Create")
	}

	found, err := repo.FindByID(ctx, s.ID)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if found == nil {
		t.Fatal("FindByID returned nil")
	}
	if found.ID != s.ID {
		t.Errorf("ID: got %s, want %s", found.ID, s.ID)
	}
	if found.Name != "methodology" {
		t.Errorf("Name: got %s, want methodology", found.Name)
	}
	if found.DisplayName != "Methodology" {
		t.Errorf("DisplayName: got %s, want Methodology", found.DisplayName)
	}
	if found.Position != 1 {
		t.Errorf("Position: got %d, want 1", found.Position)
	}
}

func TestSectionRepository_CreateTx(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()
	researchRepo := NewResearchRepository(db)
	research := createTestResearch(t, researchRepo, ctx)
	repo := NewSectionRepository(db)

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("BeginTx: %v", err)
	}

	s := &domain.Section{
		ID:          uuid.New().String(),
		ResearchID:  research.ID,
		Name:        "tx-section",
		DisplayName: "TX Section",
		Status:      domain.SectionDraft,
		Position:    0,
	}

	if err := repo.CreateTx(ctx, tx, s); err != nil {
		tx.Rollback()
		t.Fatalf("CreateTx: %v", err)
	}

	if err := tx.Commit(); err != nil {
		t.Fatalf("Commit: %v", err)
	}

	found, err := repo.FindByID(ctx, s.ID)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if found == nil {
		t.Fatal("FindByID returned nil after CreateTx + Commit")
	}
	if found.Name != "tx-section" {
		t.Errorf("Name: got %s, want tx-section", found.Name)
	}
}

func TestSectionRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()
	researchRepo := NewResearchRepository(db)
	research := createTestResearch(t, researchRepo, ctx)
	repo := NewSectionRepository(db)

	s := &domain.Section{
		ID:          uuid.New().String(),
		ResearchID:  research.ID,
		Name:        "original",
		DisplayName: "Original",
		Status:      domain.SectionDraft,
		Position:    0,
	}
	if err := repo.Create(ctx, s); err != nil {
		t.Fatalf("Create: %v", err)
	}

	s.Name = "updated"
	s.DisplayName = "Updated"
	s.Status = domain.SectionActive
	s.Position = 5

	if err := repo.Update(ctx, s); err != nil {
		t.Fatalf("Update: %v", err)
	}

	found, err := repo.FindByID(ctx, s.ID)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if found.Name != "updated" {
		t.Errorf("Name: got %s, want updated", found.Name)
	}
	if found.DisplayName != "Updated" {
		t.Errorf("DisplayName: got %s, want Updated", found.DisplayName)
	}
	if found.Status != domain.SectionActive {
		t.Errorf("Status: got %s, want active", found.Status)
	}
	if found.Position != 5 {
		t.Errorf("Position: got %d, want 5", found.Position)
	}
}

func TestSectionRepository_FindByResearch(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()
	researchRepo := NewResearchRepository(db)
	research := createTestResearch(t, researchRepo, ctx)
	repo := NewSectionRepository(db)

	sections := []*domain.Section{
		{ID: uuid.New().String(), ResearchID: research.ID, Name: "c", DisplayName: "C", Status: domain.SectionActive, Position: 3},
		{ID: uuid.New().String(), ResearchID: research.ID, Name: "a", DisplayName: "A", Status: domain.SectionActive, Position: 1},
		{ID: uuid.New().String(), ResearchID: research.ID, Name: "b", DisplayName: "B", Status: domain.SectionActive, Position: 2},
	}
	for _, s := range sections {
		if err := repo.Create(ctx, s); err != nil {
			t.Fatalf("Create: %v", err)
		}
	}

	found, err := repo.FindByResearch(ctx, research.ID)
	if err != nil {
		t.Fatalf("FindByResearch: %v", err)
	}
	if len(found) != 3 {
		t.Fatalf("expected 3 sections, got %d", len(found))
	}

	// Should be ordered by position ASC
	if found[0].Name != "a" {
		t.Errorf("first section: got %s, want a", found[0].Name)
	}
	if found[1].Name != "b" {
		t.Errorf("second section: got %s, want b", found[1].Name)
	}
	if found[2].Name != "c" {
		t.Errorf("third section: got %s, want c", found[2].Name)
	}
}

func TestSectionRepository_FindByResearchAndName(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()
	researchRepo := NewResearchRepository(db)
	research := createTestResearch(t, researchRepo, ctx)
	repo := NewSectionRepository(db)

	s := &domain.Section{
		ID:          uuid.New().String(),
		ResearchID:  research.ID,
		Name:        "unique-name",
		DisplayName: "Unique",
		Status:      domain.SectionActive,
		Position:    0,
	}
	if err := repo.Create(ctx, s); err != nil {
		t.Fatalf("Create: %v", err)
	}

	found, err := repo.FindByResearchAndName(ctx, research.ID, "unique-name")
	if err != nil {
		t.Fatalf("FindByResearchAndName: %v", err)
	}
	if found == nil {
		t.Fatal("FindByResearchAndName returned nil")
	}
	if found.ID != s.ID {
		t.Errorf("ID: got %s, want %s", found.ID, s.ID)
	}

	notFound, err := repo.FindByResearchAndName(ctx, research.ID, "nonexistent")
	if err != nil {
		t.Fatalf("FindByResearchAndName nonexistent: %v", err)
	}
	if notFound != nil {
		t.Error("expected nil for nonexistent name")
	}
}

func TestSectionRepository_CountEntriesBySection(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()
	researchRepo := NewResearchRepository(db)
	research := createTestResearch(t, researchRepo, ctx)
	sectionRepo := NewSectionRepository(db)

	s := &domain.Section{
		ID:          uuid.New().String(),
		ResearchID:  research.ID,
		Name:        "entries-section",
		DisplayName: "Entries",
		Status:      domain.SectionActive,
		Position:    0,
	}
	if err := sectionRepo.Create(ctx, s); err != nil {
		t.Fatalf("Create section: %v", err)
	}

	// Count should be 0 initially
	count, err := sectionRepo.CountEntriesBySection(ctx, s.ID)
	if err != nil {
		t.Fatalf("CountEntriesBySection: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 entries, got %d", count)
	}

	// Insert an entry directly
	now := time.Now().UTC().Format(time.DateTime)
	_, err = db.ExecContext(ctx,
		`INSERT INTO entries (id, code, research_id, section_id, title, content, description, status, tags, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		uuid.New().String(), "E1", research.ID, s.ID, "Test Entry", "content", "desc", "draft", "[]", now, now,
	)
	if err != nil {
		t.Fatalf("insert entry: %v", err)
	}

	count, err = sectionRepo.CountEntriesBySection(ctx, s.ID)
	if err != nil {
		t.Fatalf("CountEntriesBySection after insert: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 entry, got %d", count)
	}
}

func TestSectionRepository_FindByID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := NewSectionRepository(db)
	ctx := context.Background()

	found, err := repo.FindByID(ctx, "nonexistent-id")
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if found != nil {
		t.Error("expected nil for non-existent record")
	}
}
