package storage

import (
	"context"
	"testing"

	"github.com/dovod-app/app/internal/domain"
	"github.com/google/uuid"
)

func TestResearchRepository_CreateAndFindByID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewResearchRepository(db)
	ctx := context.Background()

	r := &domain.Research{
		ID:          uuid.New().String(),
		Name:        "Test Research",
		Description: "A test research project",
		Goal:        "Test goal",
		Status:      domain.ResearchActive,
		Memory:      domain.Memory{{Text: "mem1", Author: "unknown"}, {Text: "mem2", Author: "unknown"}},
		Tags:        []string{"tag1", "tag2"},
	}

	if err := repo.Create(ctx, r); err != nil {
		t.Fatalf("Create: %v", err)
	}

	if r.CreatedAt.IsZero() {
		t.Error("CreatedAt should not be zero after Create")
	}
	if r.UpdatedAt.IsZero() {
		t.Error("UpdatedAt should not be zero after Create")
	}

	found, err := repo.FindByID(ctx, r.ID)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if found == nil {
		t.Fatal("FindByID returned nil")
	}

	if found.ID != r.ID {
		t.Errorf("ID: got %s, want %s", found.ID, r.ID)
	}
	if found.Name != r.Name {
		t.Errorf("Name: got %s, want %s", found.Name, r.Name)
	}
	if found.Description != r.Description {
		t.Errorf("Description: got %s, want %s", found.Description, r.Description)
	}
	if found.Goal != r.Goal {
		t.Errorf("Goal: got %s, want %s", found.Goal, r.Goal)
	}
	if found.Status != r.Status {
		t.Errorf("Status: got %s, want %s", found.Status, r.Status)
	}
}

func TestResearchRepository_AutoAssignsCode(t *testing.T) {
	db := setupTestDB(t)
	repo := NewResearchRepository(db)
	ctx := context.Background()

	r1 := &domain.Research{
		ID:     uuid.New().String(),
		Name:   "First",
		Status: domain.ResearchActive,
	}
	r2 := &domain.Research{
		ID:     uuid.New().String(),
		Name:   "Second",
		Status: domain.ResearchActive,
	}

	if err := repo.Create(ctx, r1); err != nil {
		t.Fatalf("Create r1: %v", err)
	}
	if r1.Code != "R1" {
		t.Errorf("first code: got %s, want R1", r1.Code)
	}

	if err := repo.Create(ctx, r2); err != nil {
		t.Fatalf("Create r2: %v", err)
	}
	if r2.Code != "R2" {
		t.Errorf("second code: got %s, want R2", r2.Code)
	}
}

func TestResearchRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	repo := NewResearchRepository(db)
	ctx := context.Background()

	r := &domain.Research{
		ID:     uuid.New().String(),
		Name:   "Original",
		Status: domain.ResearchActive,
	}
	if err := repo.Create(ctx, r); err != nil {
		t.Fatalf("Create: %v", err)
	}

	r.Name = "Updated"
	r.Description = "Updated description"
	r.Status = domain.ResearchCompleted

	if err := repo.Update(ctx, r); err != nil {
		t.Fatalf("Update: %v", err)
	}

	found, err := repo.FindByID(ctx, r.ID)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if found.Name != "Updated" {
		t.Errorf("Name: got %s, want Updated", found.Name)
	}
	if found.Description != "Updated description" {
		t.Errorf("Description: got %s, want 'Updated description'", found.Description)
	}
	if found.Status != domain.ResearchCompleted {
		t.Errorf("Status: got %s, want completed", found.Status)
	}
}

func TestResearchRepository_FindByCode(t *testing.T) {
	db := setupTestDB(t)
	repo := NewResearchRepository(db)
	ctx := context.Background()

	r := &domain.Research{
		ID:     uuid.New().String(),
		Name:   "By Code",
		Status: domain.ResearchActive,
	}
	if err := repo.Create(ctx, r); err != nil {
		t.Fatalf("Create: %v", err)
	}

	found, err := repo.FindByCode(ctx, r.Code)
	if err != nil {
		t.Fatalf("FindByCode: %v", err)
	}
	if found == nil {
		t.Fatal("FindByCode returned nil")
	}
	if found.ID != r.ID {
		t.Errorf("ID: got %s, want %s", found.ID, r.ID)
	}
}

func TestResearchRepository_FindAll(t *testing.T) {
	db := setupTestDB(t)
	repo := NewResearchRepository(db)
	ctx := context.Background()

	r1 := &domain.Research{ID: uuid.New().String(), Name: "Active1", Status: domain.ResearchActive}
	r2 := &domain.Research{ID: uuid.New().String(), Name: "Completed1", Status: domain.ResearchCompleted}
	r3 := &domain.Research{ID: uuid.New().String(), Name: "Active2", Status: domain.ResearchActive}

	for _, r := range []*domain.Research{r1, r2, r3} {
		if err := repo.Create(ctx, r); err != nil {
			t.Fatalf("Create: %v", err)
		}
	}

	t.Run("no filter", func(t *testing.T) {
		all, err := repo.FindAll(ctx, ResearchFilter{})
		if err != nil {
			t.Fatalf("FindAll: %v", err)
		}
		if len(all) != 3 {
			t.Errorf("expected 3 results, got %d", len(all))
		}
	})

	t.Run("filter by status", func(t *testing.T) {
		status := domain.ResearchActive
		active, err := repo.FindAll(ctx, ResearchFilter{Status: &status})
		if err != nil {
			t.Fatalf("FindAll with filter: %v", err)
		}
		if len(active) != 2 {
			t.Errorf("expected 2 active results, got %d", len(active))
		}
		for _, r := range active {
			if r.Status != domain.ResearchActive {
				t.Errorf("expected active status, got %s", r.Status)
			}
		}
	})
}

func TestResearchRepository_Exists(t *testing.T) {
	db := setupTestDB(t)
	repo := NewResearchRepository(db)
	ctx := context.Background()

	r := &domain.Research{ID: uuid.New().String(), Name: "Exists Test", Status: domain.ResearchActive}
	if err := repo.Create(ctx, r); err != nil {
		t.Fatalf("Create: %v", err)
	}

	exists, err := repo.Exists(ctx, r.ID)
	if err != nil {
		t.Fatalf("Exists: %v", err)
	}
	if !exists {
		t.Error("expected Exists to return true")
	}

	exists, err = repo.Exists(ctx, "nonexistent-id")
	if err != nil {
		t.Fatalf("Exists nonexistent: %v", err)
	}
	if exists {
		t.Error("expected Exists to return false for nonexistent ID")
	}
}

func TestResearchRepository_FindByID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := NewResearchRepository(db)
	ctx := context.Background()

	found, err := repo.FindByID(ctx, "nonexistent-id")
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if found != nil {
		t.Error("expected nil for non-existent record")
	}
}

func TestResearchRepository_MemoryAndTags(t *testing.T) {
	db := setupTestDB(t)
	repo := NewResearchRepository(db)
	ctx := context.Background()

	t.Run("non-empty memory and tags", func(t *testing.T) {
		r := &domain.Research{
			ID:     uuid.New().String(),
			Name:   "With Memory",
			Status: domain.ResearchActive,
			Memory: domain.Memory{{Text: "fact1", Author: "unknown"}, {Text: "fact2", Author: "unknown"}},
			Tags:   []string{"go", "testing"},
		}
		if err := repo.Create(ctx, r); err != nil {
			t.Fatalf("Create: %v", err)
		}

		found, err := repo.FindByID(ctx, r.ID)
		if err != nil {
			t.Fatalf("FindByID: %v", err)
		}
		if len(found.Memory) != 2 {
			t.Fatalf("Memory length: got %d, want 2", len(found.Memory))
		}
		if found.Memory[0].Text != "fact1" || found.Memory[1].Text != "fact2" {
			t.Errorf("Memory: got %v, want [fact1 fact2]", found.Memory)
		}
		if len(found.Tags) != 2 {
			t.Fatalf("Tags length: got %d, want 2", len(found.Tags))
		}
		if found.Tags[0] != "go" || found.Tags[1] != "testing" {
			t.Errorf("Tags: got %v, want [go testing]", found.Tags)
		}
	})

	t.Run("nil memory and tags become empty slices", func(t *testing.T) {
		r := &domain.Research{
			ID:     uuid.New().String(),
			Name:   "No Memory",
			Status: domain.ResearchActive,
		}
		if err := repo.Create(ctx, r); err != nil {
			t.Fatalf("Create: %v", err)
		}

		found, err := repo.FindByID(ctx, r.ID)
		if err != nil {
			t.Fatalf("FindByID: %v", err)
		}
		if found.Memory == nil {
			t.Error("Memory should not be nil (should be empty slice)")
		}
		if len(found.Memory) != 0 {
			t.Errorf("Memory length: got %d, want 0", len(found.Memory))
		}
		if found.Tags == nil {
			t.Error("Tags should not be nil (should be empty slice)")
		}
		if len(found.Tags) != 0 {
			t.Errorf("Tags length: got %d, want 0", len(found.Tags))
		}
	})
}
