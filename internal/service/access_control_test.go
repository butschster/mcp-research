package service

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/butschster/mcp-research/internal/auth"
	"github.com/butschster/mcp-research/internal/domain"
	"github.com/butschster/mcp-research/internal/storage"
	"github.com/google/uuid"
)

// userCtx creates a context with the given user.
func userCtx(user *domain.User) context.Context {
	return auth.WithUser(context.Background(), user)
}

// setupTwoUsers creates two users directly in the DB and returns their contexts.
func setupTwoUsers(t *testing.T, db *storage.UserRepository) (*domain.User, *domain.User) {
	t.Helper()
	ctx := context.Background()

	userA := &domain.User{ID: uuid.New().String(), Email: "alice@test.com", PasswordHash: "x", Name: "Alice"}
	userB := &domain.User{ID: uuid.New().String(), Email: "bob@test.com", PasswordHash: "x", Name: "Bob"}

	if err := db.Create(ctx, userA); err != nil {
		t.Fatalf("create user A: %v", err)
	}
	if err := db.Create(ctx, userB); err != nil {
		t.Fatalf("create user B: %v", err)
	}
	return userA, userB
}

func TestAccessControl_Research(t *testing.T) {
	db := setupTestDB(t)
	notifier := &mockNotifier{}
	log := slog.Default()

	userRepo := storage.NewUserRepository(db)
	researchRepo := storage.NewResearchRepository(db)
	sectionRepo := storage.NewSectionRepository(db)
	researchSvc := NewResearchService(researchRepo, sectionRepo, notifier, log)

	userA, userB := setupTwoUsers(t, userRepo)
	ctxA := userCtx(userA)
	ctxB := userCtx(userB)

	// User A creates a research
	research, _, err := researchSvc.Create(ctxA, CreateResearchRequest{
		Name: "Alice's Research",
		Goal: "Testing access control",
		Sections: []CreateSectionRequest{
			{Name: "section-1", DisplayName: "Section 1"},
		},
	})
	if err != nil {
		t.Fatalf("create research: %v", err)
	}

	// Verify research is owned by user A
	if research.UserID != userA.ID {
		t.Fatalf("expected research.UserID=%s, got %s", userA.ID, research.UserID)
	}

	t.Run("owner can access research", func(t *testing.T) {
		r, err := researchSvc.Get(ctxA, research.ID)
		if err != nil {
			t.Fatalf("owner should access own research: %v", err)
		}
		if r.ID != research.ID {
			t.Fatal("wrong research returned")
		}
	})

	t.Run("other user cannot access research by ID", func(t *testing.T) {
		_, err := researchSvc.Get(ctxB, research.ID)
		if !errors.Is(err, ErrNotFound) {
			t.Fatalf("expected ErrNotFound, got: %v", err)
		}
	})

	t.Run("other user cannot access research by code", func(t *testing.T) {
		_, err := researchSvc.Get(ctxB, research.Code)
		if !errors.Is(err, ErrNotFound) {
			t.Fatalf("expected ErrNotFound, got: %v", err)
		}
	})

	t.Run("list only shows own researches", func(t *testing.T) {
		// User B creates their own research
		_, _, _ = researchSvc.Create(ctxB, CreateResearchRequest{Name: "Bob's Research", Goal: "Bob's goal"})

		listA, _ := researchSvc.List(ctxA, storage.ResearchFilter{})
		listB, _ := researchSvc.List(ctxB, storage.ResearchFilter{})

		if len(listA) != 1 {
			t.Fatalf("user A should see 1 research, got %d", len(listA))
		}
		if listA[0].Name != "Alice's Research" {
			t.Fatal("user A sees wrong research")
		}
		if len(listB) != 1 {
			t.Fatalf("user B should see 1 research, got %d", len(listB))
		}
		if listB[0].Name != "Bob's Research" {
			t.Fatal("user B sees wrong research")
		}
	})

	t.Run("other user cannot update research", func(t *testing.T) {
		newName := "Hacked!"
		_, err := researchSvc.Update(ctxB, research.ID, UpdateResearchRequest{Name: &newName})
		if !errors.Is(err, ErrNotFound) {
			t.Fatalf("expected ErrNotFound, got: %v", err)
		}
	})

	t.Run("other user cannot add section", func(t *testing.T) {
		_, err := researchSvc.AddSection(ctxB, research.ID, CreateSectionRequest{
			Name: "hacked-section", DisplayName: "Hacked",
		})
		if !errors.Is(err, ErrNotFound) {
			t.Fatalf("expected ErrNotFound, got: %v", err)
		}
	})
}

func TestAccessControl_Section(t *testing.T) {
	db := setupTestDB(t)
	notifier := &mockNotifier{}
	log := slog.Default()

	userRepo := storage.NewUserRepository(db)
	researchRepo := storage.NewResearchRepository(db)
	sectionRepo := storage.NewSectionRepository(db)
	entryRepo := storage.NewEntryRepository(db)
	researchSvc := NewResearchService(researchRepo, sectionRepo, notifier, log)
	sectionSvc := NewSectionService(sectionRepo, entryRepo, researchRepo, notifier, log)

	userA, userB := setupTwoUsers(t, userRepo)
	ctxA := userCtx(userA)
	ctxB := userCtx(userB)

	research, sections, _ := researchSvc.Create(ctxA, CreateResearchRequest{
		Name: "Alice's Research", Goal: "Test",
		Sections: []CreateSectionRequest{{Name: "s1", DisplayName: "S1"}},
	})
	section := sections[0]

	t.Run("owner can get section", func(t *testing.T) {
		s, err := sectionSvc.Get(ctxA, section.ID)
		if err != nil {
			t.Fatalf("owner should access section: %v", err)
		}
		if s.ID != section.ID {
			t.Fatal("wrong section")
		}
	})

	t.Run("other user cannot get section", func(t *testing.T) {
		_, err := sectionSvc.Get(ctxB, section.ID)
		if !errors.Is(err, ErrNotFound) {
			t.Fatalf("expected ErrNotFound, got: %v", err)
		}
	})

	t.Run("other user cannot list sections", func(t *testing.T) {
		list, err := sectionSvc.List(ctxB, research.ID)
		if err != nil && !errors.Is(err, ErrNotFound) {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(list) > 0 {
			t.Fatal("user B should not see sections from user A's research")
		}
	})

	t.Run("other user cannot update section", func(t *testing.T) {
		newName := "Hacked"
		_, err := sectionSvc.Update(ctxB, section.ID, UpdateSectionRequest{DisplayName: &newName})
		if !errors.Is(err, ErrNotFound) {
			t.Fatalf("expected ErrNotFound, got: %v", err)
		}
	})
}

func TestAccessControl_Entry(t *testing.T) {
	db := setupTestDB(t)
	notifier := &mockNotifier{}
	log := slog.Default()

	userRepo := storage.NewUserRepository(db)
	researchRepo := storage.NewResearchRepository(db)
	sectionRepo := storage.NewSectionRepository(db)
	entryRepo := storage.NewEntryRepository(db)
	crossrefRepo := storage.NewCrossRefRepository(db)
	researchSvc := NewResearchService(researchRepo, sectionRepo, notifier, log)
	entrySvc := NewEntryService(entryRepo, sectionRepo, researchRepo, crossrefRepo, notifier, log)

	userA, userB := setupTwoUsers(t, userRepo)
	ctxA := userCtx(userA)
	ctxB := userCtx(userB)

	research, sections, _ := researchSvc.Create(ctxA, CreateResearchRequest{
		Name: "Alice's Research", Goal: "Test",
		Sections: []CreateSectionRequest{{Name: "s1", DisplayName: "S1"}},
	})
	section := sections[0]

	// User A creates an entry
	entry, err := entrySvc.Create(ctxA, CreateEntryRequest{
		ResearchID: research.ID,
		SectionID:  section.ID,
		Content:    "# Secret entry\n\nThis is private.",
	})
	if err != nil {
		t.Fatalf("create entry: %v", err)
	}

	t.Run("other user cannot create entry in foreign research", func(t *testing.T) {
		_, err := entrySvc.Create(ctxB, CreateEntryRequest{
			ResearchID: research.ID,
			SectionID:  section.ID,
			Content:    "# Hacked",
		})
		if !errors.Is(err, ErrNotFound) {
			t.Fatalf("expected ErrNotFound, got: %v", err)
		}
	})

	t.Run("other user cannot get entry", func(t *testing.T) {
		_, err := entrySvc.Get(ctxB, entry.ID)
		if !errors.Is(err, ErrNotFound) {
			t.Fatalf("expected ErrNotFound, got: %v", err)
		}
	})

	t.Run("other user cannot update entry", func(t *testing.T) {
		newTitle := "Hacked"
		_, err := entrySvc.Update(ctxB, entry.ID, UpdateEntryRequest{Title: &newTitle})
		if !errors.Is(err, ErrNotFound) {
			t.Fatalf("expected ErrNotFound, got: %v", err)
		}
	})

	t.Run("other user cannot list entries", func(t *testing.T) {
		list, err := entrySvc.List(ctxB, research.ID, section.ID, storage.EntryFilter{})
		if err != nil && !errors.Is(err, ErrNotFound) {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(list) > 0 {
			t.Fatal("user B should not see entries from user A's research")
		}
	})

	t.Run("owner can access entry", func(t *testing.T) {
		e, err := entrySvc.Get(ctxA, entry.ID)
		if err != nil {
			t.Fatalf("owner should access entry: %v", err)
		}
		if e.ID != entry.ID {
			t.Fatal("wrong entry")
		}
	})
}

func TestAccessControl_Task(t *testing.T) {
	db := setupTestDB(t)
	notifier := &mockNotifier{}
	log := slog.Default()

	userRepo := storage.NewUserRepository(db)
	researchRepo := storage.NewResearchRepository(db)
	sectionRepo := storage.NewSectionRepository(db)
	entryRepo := storage.NewEntryRepository(db)
	crossrefRepo := storage.NewCrossRefRepository(db)
	taskRepo := storage.NewTaskRepository(db)
	researchSvc := NewResearchService(researchRepo, sectionRepo, notifier, log)
	entrySvc := NewEntryService(entryRepo, sectionRepo, researchRepo, crossrefRepo, notifier, log)
	taskSvc := NewTaskService(taskRepo, researchRepo, entrySvc, notifier, log)

	userA, userB := setupTwoUsers(t, userRepo)
	ctxA := userCtx(userA)
	ctxB := userCtx(userB)

	research, _, _ := researchSvc.Create(ctxA, CreateResearchRequest{
		Name: "Alice's Research", Goal: "Test",
	})

	// User A creates a task
	task, err := taskSvc.Create(ctxA, CreateTaskRequest{
		ResearchID: research.ID,
		Title:      "Secret task",
	})
	if err != nil {
		t.Fatalf("create task: %v", err)
	}

	t.Run("other user cannot create task in foreign research", func(t *testing.T) {
		_, err := taskSvc.Create(ctxB, CreateTaskRequest{
			ResearchID: research.ID,
			Title:      "Hacked task",
		})
		if !errors.Is(err, ErrNotFound) {
			t.Fatalf("expected ErrNotFound, got: %v", err)
		}
	})

	t.Run("other user cannot get task", func(t *testing.T) {
		_, err := taskSvc.Get(ctxB, task.ID)
		if !errors.Is(err, ErrNotFound) {
			t.Fatalf("expected ErrNotFound, got: %v", err)
		}
	})

	t.Run("other user cannot update task", func(t *testing.T) {
		newTitle := "Hacked"
		_, err := taskSvc.Update(ctxB, task.ID, UpdateTaskRequest{Title: &newTitle})
		if !errors.Is(err, ErrNotFound) {
			t.Fatalf("expected ErrNotFound, got: %v", err)
		}
	})

	t.Run("other user cannot delete task", func(t *testing.T) {
		err := taskSvc.Delete(ctxB, task.ID)
		if !errors.Is(err, ErrNotFound) {
			t.Fatalf("expected ErrNotFound, got: %v", err)
		}
	})

	t.Run("other user cannot list tasks", func(t *testing.T) {
		list, err := taskSvc.List(ctxB, research.ID, storage.TaskFilter{})
		if err != nil && !errors.Is(err, ErrNotFound) {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(list) > 0 {
			t.Fatal("user B should not see tasks from user A's research")
		}
	})

	t.Run("owner can access task", func(t *testing.T) {
		tk, err := taskSvc.Get(ctxA, task.ID)
		if err != nil {
			t.Fatalf("owner should access task: %v", err)
		}
		if tk.ID != task.ID {
			t.Fatal("wrong task")
		}
	})
}

func TestAccessControl_Session(t *testing.T) {
	db := setupTestDB(t)
	notifier := &mockNotifier{}
	log := slog.Default()

	userRepo := storage.NewUserRepository(db)
	researchRepo := storage.NewResearchRepository(db)
	sectionRepo := storage.NewSectionRepository(db)
	entryRepo := storage.NewEntryRepository(db)
	crossrefRepo := storage.NewCrossRefRepository(db)
	sessionRepo := storage.NewSessionRepository(db)
	questionRepo := storage.NewQuestionRepository(db)
	researchSvc := NewResearchService(researchRepo, sectionRepo, notifier, log)
	entrySvc := NewEntryService(entryRepo, sectionRepo, researchRepo, crossrefRepo, notifier, log)
	sessionSvc := NewSessionService(db, sessionRepo, questionRepo, researchRepo, entrySvc, notifier, log)

	userA, userB := setupTwoUsers(t, userRepo)
	ctxA := userCtx(userA)
	ctxB := userCtx(userB)

	research, _, _ := researchSvc.Create(ctxA, CreateResearchRequest{
		Name: "Alice's Research", Goal: "Test",
	})

	// User A creates a session
	session, _, err := sessionSvc.Create(ctxA, CreateSessionRequest{
		ResearchID: research.ID,
		Title:      "Interview",
		Questions: []CreateQuestionRequest{
			{Text: "What is X?", Priority: "high"},
		},
	})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}

	t.Run("other user cannot create session in foreign research", func(t *testing.T) {
		_, _, err := sessionSvc.Create(ctxB, CreateSessionRequest{
			ResearchID: research.ID,
			Title:      "Hacked session",
		})
		if !errors.Is(err, ErrNotFound) {
			t.Fatalf("expected ErrNotFound, got: %v", err)
		}
	})

	t.Run("other user cannot get session", func(t *testing.T) {
		_, err := sessionSvc.Get(ctxB, session.ID)
		if !errors.Is(err, ErrNotFound) {
			t.Fatalf("expected ErrNotFound, got: %v", err)
		}
	})

	t.Run("other user cannot update session", func(t *testing.T) {
		newTitle := "Hacked"
		_, err := sessionSvc.Update(ctxB, session.ID, UpdateSessionRequest{Title: &newTitle})
		if !errors.Is(err, ErrNotFound) {
			t.Fatalf("expected ErrNotFound, got: %v", err)
		}
	})

	t.Run("other user cannot list sessions", func(t *testing.T) {
		list, err := sessionSvc.ListByResearch(ctxB, research.ID)
		if err != nil && !errors.Is(err, ErrNotFound) {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(list) > 0 {
			t.Fatal("user B should not see sessions from user A's research")
		}
	})

	t.Run("other user cannot add questions", func(t *testing.T) {
		_, err := sessionSvc.AddQuestions(ctxB, session.ID, []CreateQuestionRequest{
			{Text: "Hacked question"},
		})
		if !errors.Is(err, ErrNotFound) {
			t.Fatalf("expected ErrNotFound, got: %v", err)
		}
	})

	t.Run("owner can access session", func(t *testing.T) {
		s, err := sessionSvc.Get(ctxA, session.ID)
		if err != nil {
			t.Fatalf("owner should access session: %v", err)
		}
		if s.Session.ID != session.ID {
			t.Fatal("wrong session")
		}
	})
}

func TestAccessControl_NoAuth(t *testing.T) {
	db := setupTestDB(t)
	notifier := &mockNotifier{}
	log := slog.Default()

	researchRepo := storage.NewResearchRepository(db)
	sectionRepo := storage.NewSectionRepository(db)
	researchSvc := NewResearchService(researchRepo, sectionRepo, notifier, log)

	// Without auth context (backward compat) — should work fine
	ctx := context.Background()

	research, _, err := researchSvc.Create(ctx, CreateResearchRequest{
		Name: "Public Research", Goal: "No auth",
	})
	if err != nil {
		t.Fatalf("create research without auth: %v", err)
	}
	if research.UserID != "" {
		t.Fatalf("expected empty UserID, got %s", research.UserID)
	}

	// Should be accessible without auth
	r, err := researchSvc.Get(ctx, research.ID)
	if err != nil {
		t.Fatalf("should access research without auth: %v", err)
	}
	if r.ID != research.ID {
		t.Fatal("wrong research")
	}

	// List without auth should return all
	list, err := researchSvc.List(ctx, storage.ResearchFilter{})
	if err != nil {
		t.Fatalf("list without auth: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 research, got %d", len(list))
	}
}
