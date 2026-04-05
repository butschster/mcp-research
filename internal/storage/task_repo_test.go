package storage

import (
	"context"
	"testing"
	"time"

	"github.com/butschster/mcp-research/internal/domain"
	"github.com/google/uuid"
)

func TestTaskRepository_CreateAndFindByID(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()
	researchRepo := NewResearchRepository(db)
	research := createTestResearch(t, researchRepo, ctx)
	repo := NewTaskRepository(db)

	task := &domain.Task{
		ID:          uuid.New().String(),
		ResearchID:  research.ID,
		Title:       "Test Task",
		Description: "Task description",
		Status:      domain.TaskPending,
		Priority:    domain.PriorityHigh,
		Result:      "",
	}

	if err := repo.Create(ctx, task); err != nil {
		t.Fatalf("Create: %v", err)
	}

	if task.CreatedAt.IsZero() {
		t.Error("CreatedAt should not be zero after Create")
	}
	if task.UpdatedAt.IsZero() {
		t.Error("UpdatedAt should not be zero after Create")
	}

	found, err := repo.FindByID(ctx, task.ID)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if found == nil {
		t.Fatal("FindByID returned nil")
	}
	if found.Title != "Test Task" {
		t.Errorf("Title: got %s, want 'Test Task'", found.Title)
	}
	if found.Priority != domain.PriorityHigh {
		t.Errorf("Priority: got %s, want high", found.Priority)
	}
	if found.CompletedAt != nil {
		t.Error("CompletedAt should be nil for pending task")
	}
}

func TestTaskRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()
	researchRepo := NewResearchRepository(db)
	research := createTestResearch(t, researchRepo, ctx)
	repo := NewTaskRepository(db)

	task := &domain.Task{
		ID:         uuid.New().String(),
		ResearchID: research.ID,
		Title:      "Original",
		Status:     domain.TaskPending,
		Priority:   domain.PriorityMedium,
	}
	if err := repo.Create(ctx, task); err != nil {
		t.Fatalf("Create: %v", err)
	}

	t.Run("update fields", func(t *testing.T) {
		task.Title = "Updated"
		task.Status = domain.TaskInProgress
		task.Priority = domain.PriorityHigh

		if err := repo.Update(ctx, task); err != nil {
			t.Fatalf("Update: %v", err)
		}

		found, err := repo.FindByID(ctx, task.ID)
		if err != nil {
			t.Fatalf("FindByID: %v", err)
		}
		if found.Title != "Updated" {
			t.Errorf("Title: got %s, want Updated", found.Title)
		}
		if found.Status != domain.TaskInProgress {
			t.Errorf("Status: got %s, want in_progress", found.Status)
		}
	})

	t.Run("set completed_at for completed status", func(t *testing.T) {
		now := time.Now().UTC()
		task.Status = domain.TaskCompleted
		task.Result = "Task completed successfully"
		task.CompletedAt = &now

		if err := repo.Update(ctx, task); err != nil {
			t.Fatalf("Update: %v", err)
		}

		found, err := repo.FindByID(ctx, task.ID)
		if err != nil {
			t.Fatalf("FindByID: %v", err)
		}
		if found.Status != domain.TaskCompleted {
			t.Errorf("Status: got %s, want completed", found.Status)
		}
		if found.CompletedAt == nil {
			t.Fatal("CompletedAt should not be nil for completed task")
		}
		if found.Result != "Task completed successfully" {
			t.Errorf("Result: got %s, want 'Task completed successfully'", found.Result)
		}
	})

	t.Run("set completed_at for failed status", func(t *testing.T) {
		failedTask := &domain.Task{
			ID:         uuid.New().String(),
			ResearchID: research.ID,
			Title:      "Failed Task",
			Status:     domain.TaskPending,
			Priority:   domain.PriorityLow,
		}
		if err := repo.Create(ctx, failedTask); err != nil {
			t.Fatalf("Create: %v", err)
		}

		now := time.Now().UTC()
		failedTask.Status = domain.TaskFailed
		failedTask.Result = "Task failed"
		failedTask.CompletedAt = &now

		if err := repo.Update(ctx, failedTask); err != nil {
			t.Fatalf("Update: %v", err)
		}

		found, err := repo.FindByID(ctx, failedTask.ID)
		if err != nil {
			t.Fatalf("FindByID: %v", err)
		}
		if found.CompletedAt == nil {
			t.Fatal("CompletedAt should not be nil for failed task")
		}
	})
}

func TestTaskRepository_FindByResearch(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()
	researchRepo := NewResearchRepository(db)
	research := createTestResearch(t, researchRepo, ctx)
	repo := NewTaskRepository(db)

	tasks := []*domain.Task{
		{ID: uuid.New().String(), ResearchID: research.ID, Title: "Low Pending", Status: domain.TaskPending, Priority: domain.PriorityLow},
		{ID: uuid.New().String(), ResearchID: research.ID, Title: "High Pending", Status: domain.TaskPending, Priority: domain.PriorityHigh},
		{ID: uuid.New().String(), ResearchID: research.ID, Title: "Medium Done", Status: domain.TaskCompleted, Priority: domain.PriorityMedium},
	}
	for _, task := range tasks {
		if err := repo.Create(ctx, task); err != nil {
			t.Fatalf("Create: %v", err)
		}
	}

	t.Run("no filter", func(t *testing.T) {
		found, err := repo.FindByResearch(ctx, research.ID, TaskFilter{})
		if err != nil {
			t.Fatalf("FindByResearch: %v", err)
		}
		if len(found) != 3 {
			t.Fatalf("expected 3 tasks, got %d", len(found))
		}
		// Ordered by priority (high, medium, low)
		if found[0].Title != "High Pending" {
			t.Errorf("first result: got %s, want 'High Pending'", found[0].Title)
		}
	})

	t.Run("filter by status", func(t *testing.T) {
		status := domain.TaskPending
		found, err := repo.FindByResearch(ctx, research.ID, TaskFilter{Status: &status})
		if err != nil {
			t.Fatalf("FindByResearch with status filter: %v", err)
		}
		if len(found) != 2 {
			t.Errorf("expected 2 pending tasks, got %d", len(found))
		}
	})

	t.Run("filter by priority", func(t *testing.T) {
		priority := domain.PriorityHigh
		found, err := repo.FindByResearch(ctx, research.ID, TaskFilter{Priority: &priority})
		if err != nil {
			t.Fatalf("FindByResearch with priority filter: %v", err)
		}
		if len(found) != 1 {
			t.Fatalf("expected 1 high-priority task, got %d", len(found))
		}
		if found[0].Title != "High Pending" {
			t.Errorf("Title: got %s, want 'High Pending'", found[0].Title)
		}
	})
}

func TestTaskRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()
	researchRepo := NewResearchRepository(db)
	research := createTestResearch(t, researchRepo, ctx)
	repo := NewTaskRepository(db)

	task := &domain.Task{
		ID:         uuid.New().String(),
		ResearchID: research.ID,
		Title:      "To Delete",
		Status:     domain.TaskPending,
		Priority:   domain.PriorityLow,
	}
	if err := repo.Create(ctx, task); err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := repo.Delete(ctx, task.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	found, err := repo.FindByID(ctx, task.ID)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if found != nil {
		t.Error("expected nil after Delete")
	}
}

func TestTaskRepository_CountByStatus(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()
	researchRepo := NewResearchRepository(db)
	research := createTestResearch(t, researchRepo, ctx)
	repo := NewTaskRepository(db)

	tasks := []*domain.Task{
		{ID: uuid.New().String(), ResearchID: research.ID, Title: "T1", Status: domain.TaskPending, Priority: domain.PriorityLow},
		{ID: uuid.New().String(), ResearchID: research.ID, Title: "T2", Status: domain.TaskPending, Priority: domain.PriorityMedium},
		{ID: uuid.New().String(), ResearchID: research.ID, Title: "T3", Status: domain.TaskCompleted, Priority: domain.PriorityHigh},
	}
	for _, task := range tasks {
		if err := repo.Create(ctx, task); err != nil {
			t.Fatalf("Create: %v", err)
		}
	}

	counts, err := repo.CountByStatus(ctx, research.ID)
	if err != nil {
		t.Fatalf("CountByStatus: %v", err)
	}

	if counts[domain.TaskPending] != 2 {
		t.Errorf("pending count: got %d, want 2", counts[domain.TaskPending])
	}
	if counts[domain.TaskCompleted] != 1 {
		t.Errorf("completed count: got %d, want 1", counts[domain.TaskCompleted])
	}
	if counts[domain.TaskInProgress] != 0 {
		t.Errorf("in_progress count: got %d, want 0", counts[domain.TaskInProgress])
	}
}

func TestTaskRepository_FindByID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := NewTaskRepository(db)
	ctx := context.Background()

	found, err := repo.FindByID(ctx, "nonexistent-id")
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if found != nil {
		t.Error("expected nil for non-existent record")
	}
}
