package service

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/butschster/mcp-research/internal/domain"
	"github.com/butschster/mcp-research/internal/storage"
)

func TestSectionService_List(t *testing.T) {
	db := setupTestDB(t)
	r, sec := createTestResearchWithSection(t, db)
	svc := NewSectionService(
		storage.NewSectionRepository(db),
		storage.NewEntryRepository(db),
		storage.NewResearchRepository(db),
		&mockNotifier{},
		slog.Default(),
	)
	ctx := context.Background()

	list, err := svc.List(ctx, r.ID)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 section, got %d", len(list))
	}
	if list[0].ID != sec.ID {
		t.Errorf("expected section ID %q, got %q", sec.ID, list[0].ID)
	}
}

func TestSectionService_Get(t *testing.T) {
	db := setupTestDB(t)
	_, sec := createTestResearchWithSection(t, db)
	svc := NewSectionService(
		storage.NewSectionRepository(db),
		storage.NewEntryRepository(db),
		storage.NewResearchRepository(db),
		&mockNotifier{},
		slog.Default(),
	)
	ctx := context.Background()

	t.Run("returns section by ID", func(t *testing.T) {
		got, err := svc.Get(ctx, sec.ID)
		if err != nil {
			t.Fatalf("get: %v", err)
		}
		if got.Name != "section-1" {
			t.Errorf("expected name %q, got %q", "section-1", got.Name)
		}
	})

	t.Run("not found", func(t *testing.T) {
		_, err := svc.Get(ctx, "nonexistent")
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})
}

func TestSectionService_Update(t *testing.T) {
	db := setupTestDB(t)
	notifier := &mockNotifier{}
	sectionRepo := storage.NewSectionRepository(db)
	entryRepo := storage.NewEntryRepository(db)
	svc := NewSectionService(sectionRepo, entryRepo, storage.NewResearchRepository(db), notifier, slog.Default())
	ctx := context.Background()

	t.Run("partial updates", func(t *testing.T) {
		_, sec := createTestResearchWithSection(t, db)
		notifier.reset()

		updated, err := svc.Update(ctx, sec.ID, UpdateSectionRequest{
			DisplayName: ptr("Updated Display"),
			Description: ptr("Updated Desc"),
			Status:      ptr(domain.SectionActive),
			Position:    ptr(5),
		})
		if err != nil {
			t.Fatalf("update: %v", err)
		}
		if updated.DisplayName != "Updated Display" {
			t.Errorf("expected display_name %q, got %q", "Updated Display", updated.DisplayName)
		}
		if updated.Description != "Updated Desc" {
			t.Errorf("expected description %q, got %q", "Updated Desc", updated.Description)
		}
		if updated.Status != domain.SectionActive {
			t.Errorf("expected status %q, got %q", domain.SectionActive, updated.Status)
		}
		if updated.Position != 5 {
			t.Errorf("expected position 5, got %d", updated.Position)
		}
		if !notifier.hasEvent("section.updated") {
			t.Error("expected section.updated event")
		}
	})

	t.Run("complete with no entries returns error", func(t *testing.T) {
		_, sec := createTestResearchWithSection(t, db)

		_, err := svc.Update(ctx, sec.ID, UpdateSectionRequest{
			Status: ptr(domain.SectionCompleted),
		})
		if !errors.Is(err, ErrSectionHasNoEntries) {
			t.Errorf("expected ErrSectionHasNoEntries, got %v", err)
		}
	})

	t.Run("complete with entries succeeds", func(t *testing.T) {
		r, sec := createTestResearchWithSection(t, db)

		// Create an entry in the section
		entrySvc := NewEntryService(entryRepo, sectionRepo, storage.NewResearchRepository(db), nil, storage.NewBlockRepository(db), storage.NewCrossRefRepository(db), nil, &mockNotifier{}, slog.Default())
		_, err := entrySvc.Create(ctx, CreateEntryRequest{
			ResearchID: r.ID,
			SectionID:  sec.ID,
			Content:    "Some content for the section",
		})
		if err != nil {
			t.Fatalf("create entry: %v", err)
		}

		updated, err := svc.Update(ctx, sec.ID, UpdateSectionRequest{
			Status: ptr(domain.SectionCompleted),
		})
		if err != nil {
			t.Fatalf("update: %v", err)
		}
		if updated.Status != domain.SectionCompleted {
			t.Errorf("expected status %q, got %q", domain.SectionCompleted, updated.Status)
		}
	})

	t.Run("not found", func(t *testing.T) {
		_, err := svc.Update(ctx, "nonexistent", UpdateSectionRequest{DisplayName: ptr("x")})
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})
}
