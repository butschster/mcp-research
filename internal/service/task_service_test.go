package service

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/butschster/mcp-research/internal/domain"
	"github.com/butschster/mcp-research/internal/storage"
)

func TestTaskService_Create(t *testing.T) {
	db := setupTestDB(t)
	notifier := &mockNotifier{}
	svc := NewTaskService(storage.NewTaskRepository(db),
		storage.NewResearchRepository(db),
		testAccess(db),
		nil,
		notifier,
		slog.Default(),
	)
	ctx := context.Background()

	t.Run("creates task with default priority medium", func(t *testing.T) {
		r := createTestResearch(t, db)
		notifier.reset()

		task, err := svc.Create(ctx, CreateTaskRequest{
			ResearchID:  r.ID,
			Title:       "My Task",
			Description: "Do something",
		})
		if err != nil {
			t.Fatalf("create: %v", err)
		}
		if task.ID == "" {
			t.Fatal("expected non-empty ID")
		}
		if task.Status != domain.TaskPending {
			t.Errorf("expected status %q, got %q", domain.TaskPending, task.Status)
		}
		if task.Priority != domain.PriorityMedium {
			t.Errorf("expected priority %q, got %q", domain.PriorityMedium, task.Priority)
		}
		if !notifier.hasEvent("task.created") {
			t.Error("expected task.created event")
		}
	})

	t.Run("creates task with provided priority", func(t *testing.T) {
		r := createTestResearch(t, db)

		task, err := svc.Create(ctx, CreateTaskRequest{
			ResearchID:  r.ID,
			Title:       "High Priority Task",
			Description: "Urgent",
			Priority:    domain.PriorityHigh,
		})
		if err != nil {
			t.Fatalf("create: %v", err)
		}
		if task.Priority != domain.PriorityHigh {
			t.Errorf("expected priority %q, got %q", domain.PriorityHigh, task.Priority)
		}
	})

	t.Run("research not found", func(t *testing.T) {
		_, err := svc.Create(ctx, CreateTaskRequest{
			ResearchID:  "nonexistent",
			Title:       "Task",
			Description: "desc",
		})
		if err == nil {
			t.Fatal("expected error")
		}
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("expected error wrapping ErrNotFound, got %v", err)
		}
	})
}

func TestTaskService_Get(t *testing.T) {
	db := setupTestDB(t)
	svc := NewTaskService(storage.NewTaskRepository(db),
		storage.NewResearchRepository(db),
		testAccess(db),
		nil,
		&mockNotifier{},
		slog.Default(),
	)
	ctx := context.Background()

	t.Run("returns task by ID", func(t *testing.T) {
		r := createTestResearch(t, db)
		task, _ := svc.Create(ctx, CreateTaskRequest{
			ResearchID: r.ID, Title: "Get Test", Description: "d",
		})

		got, err := svc.Get(ctx, task.ID)
		if err != nil {
			t.Fatalf("get: %v", err)
		}
		if got.Title != "Get Test" {
			t.Errorf("expected title %q, got %q", "Get Test", got.Title)
		}
	})

	t.Run("not found", func(t *testing.T) {
		_, err := svc.Get(ctx, "nonexistent")
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})
}

func TestTaskService_List(t *testing.T) {
	db := setupTestDB(t)
	svc := NewTaskService(storage.NewTaskRepository(db),
		storage.NewResearchRepository(db),
		testAccess(db),
		nil,
		&mockNotifier{},
		slog.Default(),
	)
	ctx := context.Background()

	r := createTestResearch(t, db)
	_, _ = svc.Create(ctx, CreateTaskRequest{ResearchID: r.ID, Title: "T1", Description: "d", Priority: domain.PriorityHigh})
	_, _ = svc.Create(ctx, CreateTaskRequest{ResearchID: r.ID, Title: "T2", Description: "d", Priority: domain.PriorityLow})

	t.Run("returns all tasks", func(t *testing.T) {
		list, err := svc.List(ctx, r.ID, storage.TaskFilter{})
		if err != nil {
			t.Fatalf("list: %v", err)
		}
		if len(list) != 2 {
			t.Errorf("expected 2 tasks, got %d", len(list))
		}
	})

	t.Run("filters by status", func(t *testing.T) {
		list, err := svc.List(ctx, r.ID, storage.TaskFilter{Status: ptr(domain.TaskPending)})
		if err != nil {
			t.Fatalf("list: %v", err)
		}
		for _, task := range list {
			if task.Status != domain.TaskPending {
				t.Errorf("expected status %q, got %q", domain.TaskPending, task.Status)
			}
		}
	})

	t.Run("filters by priority", func(t *testing.T) {
		list, err := svc.List(ctx, r.ID, storage.TaskFilter{Priority: ptr(domain.PriorityHigh)})
		if err != nil {
			t.Fatalf("list: %v", err)
		}
		if len(list) != 1 {
			t.Errorf("expected 1 task, got %d", len(list))
		}
	})
}

func TestTaskService_Update(t *testing.T) {
	db := setupTestDB(t)
	notifier := &mockNotifier{}
	svc := NewTaskService(storage.NewTaskRepository(db),
		storage.NewResearchRepository(db),
		testAccess(db),
		nil,
		notifier,
		slog.Default(),
	)
	ctx := context.Background()

	t.Run("partial updates", func(t *testing.T) {
		r := createTestResearch(t, db)
		task, _ := svc.Create(ctx, CreateTaskRequest{ResearchID: r.ID, Title: "Original", Description: "d"})
		notifier.reset()

		updated, err := svc.Update(ctx, task.ID, UpdateTaskRequest{
			Title:       ptr("Updated"),
			Description: ptr("New Desc"),
			Priority:    ptr(domain.PriorityHigh),
			Result:      ptr("some result"),
		})
		if err != nil {
			t.Fatalf("update: %v", err)
		}
		if updated.Title != "Updated" {
			t.Errorf("expected title %q, got %q", "Updated", updated.Title)
		}
		if updated.Description != "New Desc" {
			t.Errorf("expected description %q, got %q", "New Desc", updated.Description)
		}
		if updated.Priority != domain.PriorityHigh {
			t.Errorf("expected priority %q, got %q", domain.PriorityHigh, updated.Priority)
		}
		if updated.Result != "some result" {
			t.Errorf("expected result %q, got %q", "some result", updated.Result)
		}
		if !notifier.hasEvent("task.updated") {
			t.Error("expected task.updated event")
		}
	})

	t.Run("update to completed sets CompletedAt", func(t *testing.T) {
		r := createTestResearch(t, db)
		task, _ := svc.Create(ctx, CreateTaskRequest{ResearchID: r.ID, Title: "Complete Me", Description: "d"})

		updated, err := svc.Update(ctx, task.ID, UpdateTaskRequest{
			Status: ptr(domain.TaskCompleted),
		})
		if err != nil {
			t.Fatalf("update: %v", err)
		}
		if updated.CompletedAt == nil {
			t.Error("expected CompletedAt to be set")
		}
	})

	t.Run("update to failed sets CompletedAt", func(t *testing.T) {
		r := createTestResearch(t, db)
		task, _ := svc.Create(ctx, CreateTaskRequest{ResearchID: r.ID, Title: "Fail Me", Description: "d"})

		updated, err := svc.Update(ctx, task.ID, UpdateTaskRequest{
			Status: ptr(domain.TaskFailed),
		})
		if err != nil {
			t.Fatalf("update: %v", err)
		}
		if updated.CompletedAt == nil {
			t.Error("expected CompletedAt to be set")
		}
	})

	t.Run("update away from completed clears CompletedAt", func(t *testing.T) {
		r := createTestResearch(t, db)
		task, _ := svc.Create(ctx, CreateTaskRequest{ResearchID: r.ID, Title: "Reopen", Description: "d"})
		_, _ = svc.Update(ctx, task.ID, UpdateTaskRequest{Status: ptr(domain.TaskCompleted)})

		updated, err := svc.Update(ctx, task.ID, UpdateTaskRequest{
			Status: ptr(domain.TaskPending),
		})
		if err != nil {
			t.Fatalf("update: %v", err)
		}
		if updated.CompletedAt != nil {
			t.Error("expected CompletedAt to be nil")
		}
	})

	t.Run("not found", func(t *testing.T) {
		_, err := svc.Update(ctx, "nonexistent", UpdateTaskRequest{Title: ptr("x")})
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})
}

func TestTaskService_Delete(t *testing.T) {
	db := setupTestDB(t)
	notifier := &mockNotifier{}
	svc := NewTaskService(storage.NewTaskRepository(db),
		storage.NewResearchRepository(db),
		testAccess(db),
		nil,
		notifier,
		slog.Default(),
	)
	ctx := context.Background()

	t.Run("deletes task", func(t *testing.T) {
		r := createTestResearch(t, db)
		task, _ := svc.Create(ctx, CreateTaskRequest{ResearchID: r.ID, Title: "Delete Me", Description: "d"})
		notifier.reset()

		err := svc.Delete(ctx, task.ID)
		if err != nil {
			t.Fatalf("delete: %v", err)
		}
		if !notifier.hasEvent("task.deleted") {
			t.Error("expected task.deleted event")
		}

		// Verify it's gone
		_, err = svc.Get(ctx, task.ID)
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("expected ErrNotFound after delete, got %v", err)
		}
	})

	t.Run("not found", func(t *testing.T) {
		err := svc.Delete(ctx, "nonexistent")
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})
}

func TestTaskService_CountByStatus(t *testing.T) {
	db := setupTestDB(t)
	svc := NewTaskService(storage.NewTaskRepository(db),
		storage.NewResearchRepository(db),
		testAccess(db),
		nil,
		&mockNotifier{},
		slog.Default(),
	)
	ctx := context.Background()

	r := createTestResearch(t, db)
	task1, _ := svc.Create(ctx, CreateTaskRequest{ResearchID: r.ID, Title: "T1", Description: "d"})
	_, _ = svc.Create(ctx, CreateTaskRequest{ResearchID: r.ID, Title: "T2", Description: "d"})
	_, _ = svc.Update(ctx, task1.ID, UpdateTaskRequest{Status: ptr(domain.TaskCompleted)})

	counts, err := svc.CountByStatus(ctx, r.ID)
	if err != nil {
		t.Fatalf("count by status: %v", err)
	}
	if counts[domain.TaskPending] != 1 {
		t.Errorf("expected 1 pending, got %d", counts[domain.TaskPending])
	}
	if counts[domain.TaskCompleted] != 1 {
		t.Errorf("expected 1 completed, got %d", counts[domain.TaskCompleted])
	}
}
