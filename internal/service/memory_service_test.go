package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"testing"

	"github.com/butschster/mcp-research/internal/domain"
	"github.com/butschster/mcp-research/internal/storage"
)

func TestMemory_RolesIsolationAndProvenance(t *testing.T) {
	for _, role := range []domain.TeamRole{domain.TeamOwner, domain.TeamEditor, domain.TeamViewer} {
		t.Run(string(role), func(t *testing.T) {
			k := newShareKit(t)
			owner, member, research, _, teamID := k.sharedResearch(t, role)
			session, _, err := k.session.Create(owner, CreateSessionRequest{ResearchID: research.ID, Title: "Actual session"})
			if err != nil {
				t.Fatal(err)
			}
			item, err := k.research.AddMemory(owner, research.Code, "agent note", session.Code)
			if err != nil {
				t.Fatal(err)
			}
			if item.ID == "" || item.CreatedAt == nil || item.Author != "agent" || item.SessionID != session.ID || item.SessionCode != session.Code || item.Version != 1 {
				t.Fatalf("provenance: %+v", item)
			}
			if _, err := k.research.ListMemory(member, research.Code); err != nil {
				t.Fatal(err)
			}
			_, addErr := k.research.AddMemory(WithAuthor(member, domain.AuthorHuman), research.ID, "human note", "")
			editErr := k.research.UpdateMemory(member, research.ID, item.ID, "edited", item.Version)
			deleteErr := k.research.DeleteMemory(member, research.ID, []string{item.ID})
			for _, err := range []error{addErr, editErr, deleteErr} {
				if role.CanWrite() && err != nil {
					t.Fatalf("writer refused: %v", err)
				}
				if !role.CanWrite() && !errors.Is(err, ErrForbidden) {
					t.Fatalf("viewer write: %v", err)
				}
			}
			if role.CanWrite() {
				items, err := k.research.ListMemory(owner, research.ID)
				if err != nil || len(items) != 1 || items[0].Author != "user" || items[0].SessionID != "" {
					t.Fatalf("human attribution: %+v %v", items, err)
				}
			}
			other, _, err := k.research.Create(owner, CreateResearchRequest{Name: "Other", TeamID: teamID})
			if err != nil {
				t.Fatal(err)
			}
			if _, err := k.research.AddMemory(owner, other.ID, "wrong session", session.ID); !errors.Is(err, ErrNotFound) {
				t.Fatalf("cross research session accepted: %v", err)
			}
			otherItem, err := k.research.AddMemory(owner, other.ID, "other note", "")
			if err != nil {
				t.Fatal(err)
			}
			if err := k.research.UpdateMemory(owner, research.ID, otherItem.ID, "leak", 1); !errors.Is(err, ErrNotFound) {
				t.Fatalf("cross research edit: %v", err)
			}
			if err := k.research.DeleteMemory(owner, research.ID, []string{otherItem.ID}); err != nil {
				t.Fatal(err)
			}
			items, _ := k.research.ListMemory(owner, other.ID)
			if len(items) != 1 || items[0].Text != "other note" {
				t.Fatal("cross research delete succeeded")
			}
			stranger := userCtx(createTestUser(t, k.db, "stranger@test.invalid", "Stranger"))
			if _, err := k.research.ListMemory(stranger, research.ID); !errors.Is(err, ErrNotFound) {
				t.Fatalf("nonmember memory: %v", err)
			}
			if _, err := k.research.AddMemory(stranger, research.ID, "x", ""); !errors.Is(err, ErrNotFound) {
				t.Fatalf("nonmember append: %v", err)
			}
			share, err := k.shares.Create(owner, research.ID, CreateShareRequest{Include: allIncluded()})
			if err != nil {
				t.Fatal(err)
			}
			visitor := visit(t, k, share.Token)
			if _, err := k.research.ListMemory(visitor, research.ID); !errors.Is(err, ErrForbidden) {
				t.Fatalf("shared memory: %v", err)
			}
			if _, err := k.research.AddMemory(visitor, research.ID, "x", ""); !errors.Is(err, ErrForbidden) {
				t.Fatalf("shared write: %v", err)
			}
		})
	}
}

func TestMemory_ConcurrentAppendEditDeleteAndMetadata(t *testing.T) {
	db := setupTestDB(t)
	svc := NewResearchService(storage.NewResearchRepository(db), storage.NewSectionRepository(db), storage.NewTeamRepository(db), testAccess(db), NoopNotifier{}, slog.Default())
	ctx := context.Background()
	r, _, err := svc.Create(ctx, CreateResearchRequest{Name: "Concurrent memory"})
	if err != nil {
		t.Fatal(err)
	}
	first, err := svc.AddMemory(ctx, r.ID, "delete only this", "")
	if err != nil {
		t.Fatal(err)
	}
	second, err := svc.AddMemory(ctx, r.ID, "edit only this", "")
	if err != nil {
		t.Fatal(err)
	}
	stale, err := svc.researches.FindByID(ctx, r.ID)
	if err != nil {
		t.Fatal(err)
	}
	const count = 24
	start := make(chan struct{})
	errorsCh := make(chan error, count+2)
	var wg sync.WaitGroup
	for i := 0; i < count+2; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			<-start
			var err error
			switch i {
			case count:
				err = svc.DeleteMemory(ctx, r.ID, []string{first.ID})
			case count + 1:
				err = svc.UpdateMemory(ctx, r.ID, second.ID, "edited", 1)
			default:
				_, err = svc.Update(ctx, r.ID, UpdateResearchRequest{AddMemory: ptr(fmt.Sprintf("note-%d", i))})
			}
			errorsCh <- err
		}(i)
	}
	close(start)
	wg.Wait()
	close(errorsCh)
	for err := range errorsCh {
		if err != nil {
			t.Fatal(err)
		}
	}
	stale.Name = "renamed using stale research snapshot"
	if err := svc.researches.Update(ctx, stale); err != nil {
		t.Fatal(err)
	}
	items, err := svc.ListMemory(ctx, r.ID)
	if err != nil || len(items) != count+1 {
		t.Fatalf("lost concurrent notes: %d %v", len(items), err)
	}
	texts := map[string]bool{}
	for _, item := range items {
		texts[item.Text] = true
		if item.ID == first.ID {
			t.Fatal("deleted memory resurrected")
		}
	}
	for i := 0; i < count; i++ {
		if !texts[fmt.Sprintf("note-%d", i)] {
			t.Fatalf("missing note %d", i)
		}
	}
	if !texts["edited"] {
		t.Fatal("edit overwritten")
	}
	if err := svc.UpdateMemory(ctx, r.ID, second.ID, "stale edit", 1); !errors.Is(err, ErrConflict) {
		t.Fatalf("stale write: %v", err)
	}
	if err := svc.UpdateMemory(ctx, r.ID, second.ID, "no version", 0); !errors.Is(err, ErrValidation) {
		t.Fatalf("missing version: %v", err)
	}
	if err := svc.DeleteMemory(ctx, r.ID, nil); !errors.Is(err, ErrValidation) {
		t.Fatalf("unscoped delete: %v", err)
	}
}
