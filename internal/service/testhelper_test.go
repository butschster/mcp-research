package service

import (
	"context"
	"database/sql"
	"log/slog"
	"testing"

	"github.com/butschster/mcp-research/internal/config"
	"github.com/butschster/mcp-research/internal/domain"
	"github.com/butschster/mcp-research/internal/storage"
)

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := storage.NewDB(config.Config{}, slog.Default())
	if err != nil {
		t.Fatalf("setup test db: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

// mockNotifier records events for assertions.
type mockNotifier struct {
	events []Event
}

func (m *mockNotifier) Notify(e Event) {
	m.events = append(m.events, e)
}

func (m *mockNotifier) reset() { m.events = nil }

func (m *mockNotifier) hasEvent(eventType string) bool {
	for _, e := range m.events {
		if e.Type == eventType {
			return true
		}
	}
	return false
}

func ptr[T any](v T) *T { return &v }

// createTestResearch is a helper that creates a research record for FK constraints.
func createTestResearch(t *testing.T, db *sql.DB) *domain.Research {
	t.Helper()
	ctx := context.Background()
	repo := storage.NewResearchRepository(db)
	svc := NewResearchService(repo, storage.NewSectionRepository(db), &mockNotifier{}, slog.Default())
	r, _, err := svc.Create(ctx, CreateResearchRequest{
		Name:        "Test Research",
		Description: "A test research project",
		Goal:        "Testing",
	})
	if err != nil {
		t.Fatalf("create test research: %v", err)
	}
	return r
}

// createTestResearchWithSection creates a research with one section.
func createTestResearchWithSection(t *testing.T, db *sql.DB) (*domain.Research, *domain.Section) {
	t.Helper()
	ctx := context.Background()
	repo := storage.NewResearchRepository(db)
	svc := NewResearchService(repo, storage.NewSectionRepository(db), &mockNotifier{}, slog.Default())
	r, sections, err := svc.Create(ctx, CreateResearchRequest{
		Name:        "Test Research",
		Description: "A test research project",
		Goal:        "Testing",
		Sections: []CreateSectionRequest{
			{Name: "section-1", DisplayName: "Section One", Description: "First section", Position: 1},
		},
	})
	if err != nil {
		t.Fatalf("create test research with section: %v", err)
	}
	return r, sections[0]
}
