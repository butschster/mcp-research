package service

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/butschster/mcp-research/internal/domain"
	"github.com/butschster/mcp-research/internal/storage"
)

func TestResearchService_Create(t *testing.T) {
	db := setupTestDB(t)
	notifier := &mockNotifier{}
	svc := NewResearchService(storage.NewResearchRepository(db),
		storage.NewSectionRepository(db),
		storage.NewTeamRepository(db),
		testAccess(db),
		notifier,
		slog.Default(),
	)
	ctx := context.Background()

	t.Run("creates research with default status active", func(t *testing.T) {
		notifier.reset()
		r, sections, err := svc.Create(ctx, CreateResearchRequest{
			Name:        "My Research",
			Description: "Description",
			Goal:        "Goal",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if r.ID == "" {
			t.Fatal("expected non-empty ID")
		}
		if r.Status != domain.ResearchActive {
			t.Errorf("expected status %q, got %q", domain.ResearchActive, r.Status)
		}
		if len(r.Memory) != 0 {
			t.Errorf("expected empty memory, got %v", r.Memory)
		}
		if r.Tags == nil || len(r.Tags) != 0 {
			t.Errorf("expected empty tags slice, got %v", r.Tags)
		}
		if len(sections) != 0 {
			t.Errorf("expected no sections, got %d", len(sections))
		}
		if !notifier.hasEvent("research.created") {
			t.Error("expected research.created event")
		}
	})

	t.Run("creates research with sections", func(t *testing.T) {
		notifier.reset()
		r, sections, err := svc.Create(ctx, CreateResearchRequest{
			Name:        "Research With Sections",
			Description: "Has sections",
			Goal:        "Test sections",
			Tags:        []string{"tag1", "tag2"},
			Sections: []CreateSectionRequest{
				{Name: "intro", DisplayName: "Introduction", Description: "Intro section", Position: 1},
				{Name: "methods", DisplayName: "Methods", Description: "Methods section", Position: 2},
			},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(sections) != 2 {
			t.Fatalf("expected 2 sections, got %d", len(sections))
		}
		for _, sec := range sections {
			if sec.ResearchID != r.ID {
				t.Errorf("section research_id %q != research ID %q", sec.ResearchID, r.ID)
			}
			if sec.Status != domain.SectionDraft {
				t.Errorf("expected section status %q, got %q", domain.SectionDraft, sec.Status)
			}
		}
		if len(r.Tags) != 2 {
			t.Errorf("expected 2 tags, got %d", len(r.Tags))
		}
	})
}

func TestResearchService_Get(t *testing.T) {
	db := setupTestDB(t)
	svc := NewResearchService(storage.NewResearchRepository(db),
		storage.NewSectionRepository(db),
		storage.NewTeamRepository(db),
		testAccess(db),
		&mockNotifier{},
		slog.Default(),
	)
	ctx := context.Background()

	t.Run("returns research by ID", func(t *testing.T) {
		r, _, err := svc.Create(ctx, CreateResearchRequest{Name: "Get Test", Description: "d", Goal: "g"})
		if err != nil {
			t.Fatalf("create: %v", err)
		}
		got, err := svc.Get(ctx, r.ID)
		if err != nil {
			t.Fatalf("get: %v", err)
		}
		if got.Name != "Get Test" {
			t.Errorf("expected name %q, got %q", "Get Test", got.Name)
		}
	})

	t.Run("returns ErrNotFound for missing ID", func(t *testing.T) {
		_, err := svc.Get(ctx, "nonexistent-id")
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})
}

func TestResearchService_List(t *testing.T) {
	db := setupTestDB(t)
	svc := NewResearchService(storage.NewResearchRepository(db),
		storage.NewSectionRepository(db),
		storage.NewTeamRepository(db),
		testAccess(db),
		&mockNotifier{},
		slog.Default(),
	)
	ctx := context.Background()

	_, _, _ = svc.Create(ctx, CreateResearchRequest{Name: "R1", Description: "d", Goal: "g"})
	_, _, _ = svc.Create(ctx, CreateResearchRequest{Name: "R2", Description: "d", Goal: "g"})

	t.Run("returns all", func(t *testing.T) {
		list, err := svc.List(ctx, storage.ResearchFilter{})
		if err != nil {
			t.Fatalf("list: %v", err)
		}
		if len(list) < 2 {
			t.Errorf("expected at least 2, got %d", len(list))
		}
	})

	t.Run("filters by status", func(t *testing.T) {
		status := domain.ResearchActive
		list, err := svc.List(ctx, storage.ResearchFilter{Status: &status})
		if err != nil {
			t.Fatalf("list: %v", err)
		}
		for _, r := range list {
			if r.Status != domain.ResearchActive {
				t.Errorf("expected active status, got %q", r.Status)
			}
		}
	})
}

func TestResearchService_Update(t *testing.T) {
	db := setupTestDB(t)
	notifier := &mockNotifier{}
	svc := NewResearchService(storage.NewResearchRepository(db),
		storage.NewSectionRepository(db),
		storage.NewTeamRepository(db),
		testAccess(db),
		notifier,
		slog.Default(),
	)
	ctx := context.Background()

	t.Run("partial updates", func(t *testing.T) {
		r, _, _ := svc.Create(ctx, CreateResearchRequest{Name: "Original", Description: "d", Goal: "g"})
		notifier.reset()

		updated, err := svc.Update(ctx, r.ID, UpdateResearchRequest{
			Name:        ptr("Updated Name"),
			Description: ptr("Updated Desc"),
			Goal:        ptr("Updated Goal"),
			Status:      ptr(domain.ResearchCompleted),
			Instruction: ptr("new instruction"),
			Tags:        []string{"new-tag"},
		})
		if err != nil {
			t.Fatalf("update: %v", err)
		}
		if updated.Name != "Updated Name" {
			t.Errorf("expected name %q, got %q", "Updated Name", updated.Name)
		}
		if updated.Description != "Updated Desc" {
			t.Errorf("expected description %q, got %q", "Updated Desc", updated.Description)
		}
		if updated.Goal != "Updated Goal" {
			t.Errorf("expected goal %q, got %q", "Updated Goal", updated.Goal)
		}
		if updated.Status != domain.ResearchCompleted {
			t.Errorf("expected status %q, got %q", domain.ResearchCompleted, updated.Status)
		}
		if updated.Instruction != "new instruction" {
			t.Errorf("expected instruction %q, got %q", "new instruction", updated.Instruction)
		}
		if len(updated.Tags) != 1 || updated.Tags[0] != "new-tag" {
			t.Errorf("expected tags [new-tag], got %v", updated.Tags)
		}
		if !notifier.hasEvent("research.updated") {
			t.Error("expected research.updated event")
		}
	})

	t.Run("not found", func(t *testing.T) {
		_, err := svc.Update(ctx, "nonexistent", UpdateResearchRequest{Name: ptr("x")})
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})

	t.Run("memory replace", func(t *testing.T) {
		r, _, _ := svc.Create(ctx, CreateResearchRequest{Name: "MemTest", Description: "d", Goal: "g"})
		updated, err := svc.Update(ctx, r.ID, UpdateResearchRequest{
			Memory: []string{"mem1", "mem2"},
		})
		if err != nil {
			t.Fatalf("update: %v", err)
		}
		if len(updated.Memory) != 2 {
			t.Errorf("expected 2 memory entries, got %d", len(updated.Memory))
		}
	})

	t.Run("add memory", func(t *testing.T) {
		r, _, _ := svc.Create(ctx, CreateResearchRequest{Name: "AddMemTest", Description: "d", Goal: "g"})
		_, _ = svc.Update(ctx, r.ID, UpdateResearchRequest{Memory: []string{"existing"}})
		updated, err := svc.Update(ctx, r.ID, UpdateResearchRequest{AddMemory: ptr("appended")})
		if err != nil {
			t.Fatalf("update: %v", err)
		}
		if len(updated.Memory) != 2 {
			t.Errorf("expected 2 memory entries, got %d", len(updated.Memory))
		}
		if updated.Memory[1] != "appended" {
			t.Errorf("expected second memory %q, got %q", "appended", updated.Memory[1])
		}
	})

	t.Run("mutual exclusion memory and add_memory", func(t *testing.T) {
		r, _, _ := svc.Create(ctx, CreateResearchRequest{Name: "MutExTest", Description: "d", Goal: "g"})
		_, err := svc.Update(ctx, r.ID, UpdateResearchRequest{
			Memory:    []string{"a"},
			AddMemory: ptr("b"),
		})
		if !errors.Is(err, ErrMutualExclusion) {
			t.Errorf("expected ErrMutualExclusion, got %v", err)
		}
	})
}

func TestResearchService_AddSection(t *testing.T) {
	db := setupTestDB(t)
	notifier := &mockNotifier{}
	svc := NewResearchService(storage.NewResearchRepository(db),
		storage.NewSectionRepository(db),
		storage.NewTeamRepository(db),
		testAccess(db),
		notifier,
		slog.Default(),
	)
	ctx := context.Background()

	t.Run("adds section to research", func(t *testing.T) {
		r, _, _ := svc.Create(ctx, CreateResearchRequest{Name: "SecTest", Description: "d", Goal: "g"})
		notifier.reset()

		sec, err := svc.AddSection(ctx, r.ID, CreateSectionRequest{
			Name: "new-section", DisplayName: "New Section", Description: "desc", Position: 1,
		})
		if err != nil {
			t.Fatalf("add section: %v", err)
		}
		if sec.ResearchID != r.ID {
			t.Errorf("expected research_id %q, got %q", r.ID, sec.ResearchID)
		}
		if sec.Status != domain.SectionDraft {
			t.Errorf("expected status %q, got %q", domain.SectionDraft, sec.Status)
		}
		if !notifier.hasEvent("section.created") {
			t.Error("expected section.created event")
		}
	})

	t.Run("duplicate name", func(t *testing.T) {
		r, _, _ := svc.Create(ctx, CreateResearchRequest{Name: "DupTest", Description: "d", Goal: "g"})
		_, _ = svc.AddSection(ctx, r.ID, CreateSectionRequest{Name: "dup", DisplayName: "Dup", Position: 1})
		_, err := svc.AddSection(ctx, r.ID, CreateSectionRequest{Name: "dup", DisplayName: "Dup2", Position: 2})
		if !errors.Is(err, ErrDuplicateSectionName) {
			t.Errorf("expected ErrDuplicateSectionName, got %v", err)
		}
	})

	t.Run("research not found", func(t *testing.T) {
		_, err := svc.AddSection(ctx, "nonexistent", CreateSectionRequest{Name: "x", DisplayName: "X", Position: 1})
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})
}
