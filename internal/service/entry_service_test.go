package service

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"testing"

	"github.com/butschster/mcp-research/internal/domain"
	"github.com/butschster/mcp-research/internal/storage"
)

func newEntryService(db *sql.DB, n EventNotifier) *EntryService {
	return NewEntryService(storage.NewEntryRepository(db),
		storage.NewSectionRepository(db),
		storage.NewResearchRepository(db),
		testAccess(db),
		storage.NewSessionRepository(db),
		storage.NewBlockRepository(db),
		storage.NewEntryRevisionRepository(db),
		storage.NewCrossRefRepository(db),
		storage.NewExternalLinkRepository(db),
		n,
		slog.Default(),
	)
}

func TestEntryService_Create(t *testing.T) {
	db := setupTestDB(t)
	notifier := &mockNotifier{}
	svc := newEntryService(db, notifier)
	ctx := context.Background()

	t.Run("creates entry with auto title and description", func(t *testing.T) {
		r, sec := createTestResearchWithSection(t, db)
		notifier.reset()

		entry, err := svc.Create(ctx, CreateEntryRequest{
			ResearchID: r.ID,
			SectionID:  sec.ID,
			Content:    "# My Title\n\nSome body content here\nAnother line",
		})
		if err != nil {
			t.Fatalf("create: %v", err)
		}
		if entry.ID == "" {
			t.Fatal("expected non-empty ID")
		}
		if entry.Status != domain.EntryDraft {
			t.Errorf("expected status %q, got %q", domain.EntryDraft, entry.Status)
		}
		if entry.Title == "" {
			t.Error("expected auto-generated title")
		}
		if entry.ResearchID != r.ID {
			t.Errorf("expected research_id %q, got %q", r.ID, entry.ResearchID)
		}
		if entry.SectionID != sec.ID {
			t.Errorf("expected section_id %q, got %q", sec.ID, entry.SectionID)
		}
		if !notifier.hasEvent("entry.created") {
			t.Error("expected entry.created event")
		}
	})

	t.Run("creates entry with explicit title", func(t *testing.T) {
		r, sec := createTestResearchWithSection(t, db)

		entry, err := svc.Create(ctx, CreateEntryRequest{
			ResearchID: r.ID,
			SectionID:  sec.ID,
			Content:    "content here",
			Title:      "Explicit Title",
		})
		if err != nil {
			t.Fatalf("create: %v", err)
		}
		if entry.Title != "Explicit Title" {
			t.Errorf("expected title %q, got %q", "Explicit Title", entry.Title)
		}
	})

	t.Run("empty content returns error", func(t *testing.T) {
		r, sec := createTestResearchWithSection(t, db)

		_, err := svc.Create(ctx, CreateEntryRequest{
			ResearchID: r.ID,
			SectionID:  sec.ID,
			Content:    "",
		})
		if err == nil {
			t.Fatal("expected error for empty content")
		}
	})

	t.Run("research not found", func(t *testing.T) {
		_, sec := createTestResearchWithSection(t, db)

		_, err := svc.Create(ctx, CreateEntryRequest{
			ResearchID: "nonexistent-research",
			SectionID:  sec.ID,
			Content:    "content",
		})
		if err == nil {
			t.Fatal("expected error")
		}
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("expected error wrapping ErrNotFound, got %v", err)
		}
	})

	t.Run("section not found", func(t *testing.T) {
		r := createTestResearch(t, db)

		_, err := svc.Create(ctx, CreateEntryRequest{
			ResearchID: r.ID,
			SectionID:  "nonexistent-section",
			Content:    "content",
		})
		if err == nil {
			t.Fatal("expected error")
		}
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("expected error wrapping ErrNotFound, got %v", err)
		}
	})

	t.Run("section wrong research", func(t *testing.T) {
		_, sec := createTestResearchWithSection(t, db)
		otherResearch := createTestResearch(t, db)

		_, err := svc.Create(ctx, CreateEntryRequest{
			ResearchID: otherResearch.ID,
			SectionID:  sec.ID,
			Content:    "content",
		})
		if err == nil {
			t.Fatal("expected error for section not belonging to research")
		}
	})
}

func TestEntryService_Get(t *testing.T) {
	db := setupTestDB(t)
	svc := newEntryService(db, &mockNotifier{})
	ctx := context.Background()

	t.Run("returns entry by ID", func(t *testing.T) {
		r, sec := createTestResearchWithSection(t, db)
		entry, _ := svc.Create(ctx, CreateEntryRequest{
			ResearchID: r.ID, SectionID: sec.ID, Content: "test content",
		})

		got, err := svc.Get(ctx, entry.ID)
		if err != nil {
			t.Fatalf("get: %v", err)
		}
		if got.Content != "test content" {
			t.Errorf("expected content %q, got %q", "test content", got.Content)
		}
	})

	t.Run("not found", func(t *testing.T) {
		_, err := svc.Get(ctx, "nonexistent")
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})
}

func TestEntryService_List(t *testing.T) {
	db := setupTestDB(t)
	svc := newEntryService(db, &mockNotifier{})
	ctx := context.Background()

	r, sec := createTestResearchWithSection(t, db)
	_, _ = svc.Create(ctx, CreateEntryRequest{ResearchID: r.ID, SectionID: sec.ID, Content: "entry one"})
	_, _ = svc.Create(ctx, CreateEntryRequest{ResearchID: r.ID, SectionID: sec.ID, Content: "entry two"})

	t.Run("lists entries for section", func(t *testing.T) {
		list, err := svc.List(ctx, r.ID, sec.ID, storage.EntryFilter{})
		if err != nil {
			t.Fatalf("list: %v", err)
		}
		if len(list) != 2 {
			t.Errorf("expected 2 entries, got %d", len(list))
		}
	})
}

func TestEntryService_ListByResearch(t *testing.T) {
	db := setupTestDB(t)
	svc := newEntryService(db, &mockNotifier{})
	ctx := context.Background()

	r, sec := createTestResearchWithSection(t, db)
	_, _ = svc.Create(ctx, CreateEntryRequest{ResearchID: r.ID, SectionID: sec.ID, Content: "entry one"})

	list, err := svc.ListByResearch(ctx, r.ID, storage.EntryFilter{})
	if err != nil {
		t.Fatalf("list by research: %v", err)
	}
	if len(list) != 1 {
		t.Errorf("expected 1 entry, got %d", len(list))
	}
}

func TestEntryService_Update(t *testing.T) {
	db := setupTestDB(t)
	notifier := &mockNotifier{}
	svc := newEntryService(db, notifier)
	ctx := context.Background()

	t.Run("partial updates", func(t *testing.T) {
		r, sec := createTestResearchWithSection(t, db)
		entry, _ := svc.Create(ctx, CreateEntryRequest{
			ResearchID: r.ID, SectionID: sec.ID, Content: "original content",
		})
		notifier.reset()

		updated, err := svc.Update(ctx, entry.ID, UpdateEntryRequest{
			Title:       ptr("New Title"),
			Content:     ptr("new content"),
			Description: ptr("new desc"),
			Status:      ptr(domain.EntryActive),
			Tags:        []string{"t1"},
		})
		if err != nil {
			t.Fatalf("update: %v", err)
		}
		if updated.Title != "New Title" {
			t.Errorf("expected title %q, got %q", "New Title", updated.Title)
		}
		if updated.Content != "new content" {
			t.Errorf("expected content %q, got %q", "new content", updated.Content)
		}
		if updated.Description != "new desc" {
			t.Errorf("expected description %q, got %q", "new desc", updated.Description)
		}
		if updated.Status != domain.EntryActive {
			t.Errorf("expected status %q, got %q", domain.EntryActive, updated.Status)
		}
		if len(updated.Tags) != 1 || updated.Tags[0] != "t1" {
			t.Errorf("expected tags [t1], got %v", updated.Tags)
		}
		if !notifier.hasEvent("entry.updated") {
			t.Error("expected entry.updated event")
		}
	})

	t.Run("text replace", func(t *testing.T) {
		r, sec := createTestResearchWithSection(t, db)
		entry, _ := svc.Create(ctx, CreateEntryRequest{
			ResearchID: r.ID, SectionID: sec.ID, Content: "hello world",
		})

		updated, err := svc.Update(ctx, entry.ID, UpdateEntryRequest{
			TextReplace: &TextReplace{From: "world", To: "universe"},
		})
		if err != nil {
			t.Fatalf("update: %v", err)
		}
		if updated.Content != "hello universe" {
			t.Errorf("expected content %q, got %q", "hello universe", updated.Content)
		}
	})

	t.Run("text replace not found", func(t *testing.T) {
		r, sec := createTestResearchWithSection(t, db)
		entry, _ := svc.Create(ctx, CreateEntryRequest{
			ResearchID: r.ID, SectionID: sec.ID, Content: "hello world",
		})

		_, err := svc.Update(ctx, entry.ID, UpdateEntryRequest{
			TextReplace: &TextReplace{From: "nonexistent", To: "replacement"},
		})
		if !errors.Is(err, ErrTextReplaceNotFound) {
			t.Errorf("expected ErrTextReplaceNotFound, got %v", err)
		}
	})

	t.Run("not found", func(t *testing.T) {
		_, err := svc.Update(ctx, "nonexistent", UpdateEntryRequest{Title: ptr("x")})
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})
}
