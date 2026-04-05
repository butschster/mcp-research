package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/butschster/mcp-research/internal/domain"
	"github.com/butschster/mcp-research/internal/storage"
	"github.com/google/uuid"
)

type CreateTaskRequest struct {
	ResearchID  string
	Title       string
	Description string
	Priority    domain.Priority
}

type UpdateTaskRequest struct {
	Title       *string
	Description *string
	Status      *domain.TaskStatus
	Priority    *domain.Priority
	Result      *string
}

type TaskService struct {
	tasks      *storage.TaskRepository
	researches *storage.ResearchRepository
	crossrefs  CrossRefParser
	events     EventNotifier
	log        *slog.Logger
}

func NewTaskService(tasks *storage.TaskRepository, researches *storage.ResearchRepository, crossrefs CrossRefParser, events EventNotifier, log *slog.Logger) *TaskService {
	return &TaskService{tasks: tasks, researches: researches, crossrefs: crossrefs, events: events, log: log}
}

func (s *TaskService) Create(ctx context.Context, req CreateTaskRequest) (*domain.Task, error) {
	if err := validateResearchAccess(ctx, s.researches, req.ResearchID); err != nil {
		return nil, fmt.Errorf("research %s: %w", req.ResearchID, err)
	}

	priority := req.Priority
	if priority == "" {
		priority = domain.PriorityMedium
	}

	task := &domain.Task{
		ID:          uuid.New().String(),
		ResearchID:  req.ResearchID,
		Title:       req.Title,
		Description: req.Description,
		Status:      domain.TaskPending,
		Priority:    priority,
	}

	if err := s.tasks.Create(ctx, task); err != nil {
		return nil, fmt.Errorf("create task: %w", err)
	}

	s.events.Notify(Event{Type: "task.created", ResearchID: task.ResearchID, EntityID: task.ID, Entity: "task"})
	return task, nil
}

func (s *TaskService) Get(ctx context.Context, id string) (*domain.Task, error) {
	task, err := s.tasks.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("find task: %w", err)
	}
	if task == nil {
		return nil, ErrNotFound
	}
	if err := validateResearchAccess(ctx, s.researches, task.ResearchID); err != nil {
		return nil, ErrNotFound
	}
	return task, nil
}

func (s *TaskService) List(ctx context.Context, researchID string, filter storage.TaskFilter) ([]*domain.Task, error) {
	if err := validateResearchAccess(ctx, s.researches, researchID); err != nil {
		return nil, err
	}
	return s.tasks.FindByResearch(ctx, researchID, filter)
}

func (s *TaskService) Update(ctx context.Context, id string, req UpdateTaskRequest) (*domain.Task, error) {
	task, err := s.tasks.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("find task: %w", err)
	}
	if task == nil {
		return nil, ErrNotFound
	}
	if err := validateResearchAccess(ctx, s.researches, task.ResearchID); err != nil {
		return nil, ErrNotFound
	}

	if req.Title != nil {
		task.Title = *req.Title
	}
	if req.Description != nil {
		task.Description = *req.Description
	}
	if req.Priority != nil {
		task.Priority = *req.Priority
	}
	if req.Result != nil {
		task.Result = *req.Result
	}
	if req.Status != nil {
		task.Status = *req.Status
		// Auto-set completed_at
		if *req.Status == domain.TaskCompleted || *req.Status == domain.TaskFailed {
			now := time.Now().UTC()
			task.CompletedAt = &now
		} else {
			task.CompletedAt = nil
		}
	}

	if err := s.tasks.Update(ctx, task); err != nil {
		return nil, fmt.Errorf("update task: %w", err)
	}

	// Parse crossrefs from result and description
	if s.crossrefs != nil {
		text := task.Description + "\n" + task.Result
		s.crossrefs.ParseCrossRefs(ctx, "task", task.ID, task.ResearchID, text)
	}

	s.events.Notify(Event{Type: "task.updated", ResearchID: task.ResearchID, EntityID: task.ID, Entity: "task"})
	return task, nil
}

func (s *TaskService) Delete(ctx context.Context, id string) error {
	task, err := s.tasks.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("find task: %w", err)
	}
	if task == nil {
		return ErrNotFound
	}
	if err := validateResearchAccess(ctx, s.researches, task.ResearchID); err != nil {
		return ErrNotFound
	}
	if err := s.tasks.Delete(ctx, id); err != nil {
		return err
	}
	s.events.Notify(Event{Type: "task.deleted", ResearchID: task.ResearchID, EntityID: id, Entity: "task"})
	return nil
}

func (s *TaskService) CountByStatus(ctx context.Context, researchID string) (map[domain.TaskStatus]int, error) {
	return s.tasks.CountByStatus(ctx, researchID)
}
