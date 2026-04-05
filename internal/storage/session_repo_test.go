package storage

import (
	"context"
	"testing"

	"github.com/butschster/mcp-research/internal/domain"
	"github.com/google/uuid"
)

func TestSessionRepository_CreateAndFindByID(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()
	researchRepo := NewResearchRepository(db)
	research := createTestResearch(t, researchRepo, ctx)
	repo := NewSessionRepository(db)

	s := &domain.Session{
		ID:         uuid.New().String(),
		ResearchID: research.ID,
		Title:      "Interview Session",
		Focus:      "User experience",
		Status:     domain.SessionActive,
		Notes:      "Some notes",
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
	if found.Title != "Interview Session" {
		t.Errorf("Title: got %s, want 'Interview Session'", found.Title)
	}
	if found.Focus != "User experience" {
		t.Errorf("Focus: got %s, want 'User experience'", found.Focus)
	}
	if found.Notes != "Some notes" {
		t.Errorf("Notes: got %s, want 'Some notes'", found.Notes)
	}
}

func TestSessionRepository_CreateTx(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()
	researchRepo := NewResearchRepository(db)
	research := createTestResearch(t, researchRepo, ctx)
	repo := NewSessionRepository(db)

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("BeginTx: %v", err)
	}

	s := &domain.Session{
		ID:         uuid.New().String(),
		ResearchID: research.ID,
		Title:      "TX Session",
		Focus:      "Testing",
		Status:     domain.SessionActive,
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
	if found.Title != "TX Session" {
		t.Errorf("Title: got %s, want 'TX Session'", found.Title)
	}
}

func TestSessionRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()
	researchRepo := NewResearchRepository(db)
	research := createTestResearch(t, researchRepo, ctx)
	repo := NewSessionRepository(db)

	s := &domain.Session{
		ID:         uuid.New().String(),
		ResearchID: research.ID,
		Title:      "Original",
		Focus:      "Original focus",
		Status:     domain.SessionActive,
		Notes:      "",
	}
	if err := repo.Create(ctx, s); err != nil {
		t.Fatalf("Create: %v", err)
	}

	s.Title = "Updated"
	s.Focus = "Updated focus"
	s.Status = domain.SessionCompleted
	s.Notes = "Final notes"

	if err := repo.Update(ctx, s); err != nil {
		t.Fatalf("Update: %v", err)
	}

	found, err := repo.FindByID(ctx, s.ID)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if found.Title != "Updated" {
		t.Errorf("Title: got %s, want Updated", found.Title)
	}
	if found.Focus != "Updated focus" {
		t.Errorf("Focus: got %s, want 'Updated focus'", found.Focus)
	}
	if found.Status != domain.SessionCompleted {
		t.Errorf("Status: got %s, want completed", found.Status)
	}
	if found.Notes != "Final notes" {
		t.Errorf("Notes: got %s, want 'Final notes'", found.Notes)
	}
}

func TestSessionRepository_FindByResearch(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()
	researchRepo := NewResearchRepository(db)
	research := createTestResearch(t, researchRepo, ctx)
	repo := NewSessionRepository(db)

	sessions := []*domain.Session{
		{ID: uuid.New().String(), ResearchID: research.ID, Title: "First", Focus: "f1", Status: domain.SessionActive},
		{ID: uuid.New().String(), ResearchID: research.ID, Title: "Second", Focus: "f2", Status: domain.SessionCompleted},
		{ID: uuid.New().String(), ResearchID: research.ID, Title: "Third", Focus: "f3", Status: domain.SessionActive},
	}
	for _, s := range sessions {
		if err := repo.Create(ctx, s); err != nil {
			t.Fatalf("Create: %v", err)
		}
	}

	found, err := repo.FindByResearch(ctx, research.ID)
	if err != nil {
		t.Fatalf("FindByResearch: %v", err)
	}
	if len(found) != 3 {
		t.Fatalf("expected 3 sessions, got %d", len(found))
	}

	// Verify all titles are present (ordering may not be deterministic for same-second timestamps)
	titles := make(map[string]bool)
	for _, s := range found {
		titles[s.Title] = true
	}
	for _, expected := range []string{"First", "Second", "Third"} {
		if !titles[expected] {
			t.Errorf("missing session with title %s", expected)
		}
	}
}

func TestSessionRepository_FindActive(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()
	researchRepo := NewResearchRepository(db)
	research := createTestResearch(t, researchRepo, ctx)
	repo := NewSessionRepository(db)

	// No active sessions initially
	active, err := repo.FindActive(ctx, research.ID)
	if err != nil {
		t.Fatalf("FindActive: %v", err)
	}
	if active != nil {
		t.Error("expected nil when no active sessions")
	}

	// Create a completed session
	s1 := &domain.Session{
		ID:         uuid.New().String(),
		ResearchID: research.ID,
		Title:      "Completed",
		Status:     domain.SessionCompleted,
	}
	if err := repo.Create(ctx, s1); err != nil {
		t.Fatalf("Create: %v", err)
	}

	active, err = repo.FindActive(ctx, research.ID)
	if err != nil {
		t.Fatalf("FindActive: %v", err)
	}
	if active != nil {
		t.Error("expected nil when only completed sessions exist")
	}

	// Create an active session
	s2 := &domain.Session{
		ID:         uuid.New().String(),
		ResearchID: research.ID,
		Title:      "Active One",
		Status:     domain.SessionActive,
	}
	if err := repo.Create(ctx, s2); err != nil {
		t.Fatalf("Create: %v", err)
	}

	active, err = repo.FindActive(ctx, research.ID)
	if err != nil {
		t.Fatalf("FindActive: %v", err)
	}
	if active == nil {
		t.Fatal("expected active session, got nil")
	}
	if active.Title != "Active One" {
		t.Errorf("Title: got %s, want 'Active One'", active.Title)
	}
}

func TestSessionRepository_FindByID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := NewSessionRepository(db)
	ctx := context.Background()

	found, err := repo.FindByID(ctx, "nonexistent-id")
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if found != nil {
		t.Error("expected nil for non-existent record")
	}
}
