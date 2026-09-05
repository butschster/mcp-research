package storage

import (
	"context"
	"testing"

	"github.com/butschster/mcp-research/internal/domain"
)

func TestMemoryRepository_AtomicUpdateAndForeignKeys(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()
	researches := NewResearchRepository(db)
	r := &domain.Research{ID: "memory-parent", Name: "Original", Status: domain.ResearchActive}
	if err := researches.Create(ctx, r); err != nil {
		t.Fatal(err)
	}
	r.Name = "must roll back"
	bad := &domain.MemoryItem{Text: "invalid reference", Author: "agent", SessionID: "missing-session"}
	if err := researches.UpdateWithMemory(ctx, r, bad); err == nil {
		t.Fatal("invalid foreign key accepted")
	}
	got, err := researches.FindByID(ctx, r.ID)
	if err != nil || got.Name != "Original" || len(got.Memory) != 0 {
		t.Fatalf("partial update: %+v %v", got, err)
	}
	session := &domain.Session{ID: "memory-session", ResearchID: r.ID, Title: "Actual session", Status: domain.SessionActive}
	if err := NewSessionRepository(db).Create(ctx, session); err != nil {
		t.Fatal(err)
	}
	item := &domain.MemoryItem{Text: "persistent note", Author: "user", SessionID: session.ID}
	if err := NewMemoryRepository(db).Create(ctx, r.ID, item); err != nil {
		t.Fatal(err)
	}
	if _, err := db.NewDelete().Table("sessions").Where("id=?", session.ID).Exec(ctx); err != nil {
		t.Fatal(err)
	}
	items, err := NewMemoryRepository(db).List(ctx, r.ID)
	if err != nil || len(items) != 1 || items[0].SessionID != "" || items[0].Text != "persistent note" {
		t.Fatalf("session deletion lost memory: %+v %v", items, err)
	}
	if _, err := db.NewDelete().Table("researches").Where("id=?", r.ID).Exec(ctx); err != nil {
		t.Fatal(err)
	}
	items, err = NewMemoryRepository(db).List(ctx, r.ID)
	if err != nil || len(items) != 0 {
		t.Fatalf("research deletion left memory: %+v %v", items, err)
	}
}
