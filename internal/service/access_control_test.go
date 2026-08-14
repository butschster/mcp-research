package service

import (
	"context"
	"errors"
	"log/slog"
	"strings"
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
	blockRepo := storage.NewBlockRepository(db)
	crossrefRepo := storage.NewCrossRefRepository(db)
	researchSvc := NewResearchService(researchRepo, sectionRepo, notifier, log)
	entrySvc := NewEntryService(entryRepo, sectionRepo, researchRepo, nil, blockRepo, storage.NewEntryRevisionRepository(db), crossrefRepo, nil, notifier, log)

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
	blockRepo := storage.NewBlockRepository(db)
	crossrefRepo := storage.NewCrossRefRepository(db)
	taskRepo := storage.NewTaskRepository(db)
	researchSvc := NewResearchService(researchRepo, sectionRepo, notifier, log)
	entrySvc := NewEntryService(entryRepo, sectionRepo, researchRepo, nil, blockRepo, storage.NewEntryRevisionRepository(db), crossrefRepo, nil, notifier, log)
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
	blockRepo := storage.NewBlockRepository(db)
	crossrefRepo := storage.NewCrossRefRepository(db)
	sessionRepo := storage.NewSessionRepository(db)
	questionRepo := storage.NewQuestionRepository(db)
	researchSvc := NewResearchService(researchRepo, sectionRepo, notifier, log)
	entrySvc := NewEntryService(entryRepo, sectionRepo, researchRepo, nil, blockRepo, storage.NewEntryRevisionRepository(db), crossrefRepo, nil, notifier, log)
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

// TestAccessControl_Revisions covers the history surface added with entry
// revisions. A revision holds the entry's full text, so every one of these
// paths is a copy of the content the ownership check exists to protect — and
// two of them (diff, restore) can also mutate or describe an entry by id alone.
func TestAccessControl_Revisions(t *testing.T) {
	db := setupTestDB(t)
	notifier := &mockNotifier{}
	log := slog.Default()

	userRepo := storage.NewUserRepository(db)
	researchRepo := storage.NewResearchRepository(db)
	sectionRepo := storage.NewSectionRepository(db)
	entryRepo := storage.NewEntryRepository(db)
	blockRepo := storage.NewBlockRepository(db)
	crossrefRepo := storage.NewCrossRefRepository(db)
	sessionRepo := storage.NewSessionRepository(db)
	questionRepo := storage.NewQuestionRepository(db)

	researchSvc := NewResearchService(researchRepo, sectionRepo, notifier, log)
	entrySvc := NewEntryService(entryRepo, sectionRepo, researchRepo, sessionRepo, blockRepo, storage.NewEntryRevisionRepository(db), crossrefRepo, nil, notifier, log)
	sessionSvc := NewSessionService(db, sessionRepo, questionRepo, researchRepo, entrySvc, notifier, log)

	userA, userB := setupTwoUsers(t, userRepo)
	ctxA := userCtx(userA)
	ctxB := userCtx(userB)

	research, sections, _ := researchSvc.Create(ctxA, CreateResearchRequest{
		Name: "Alice's Research", Goal: "Test",
		Sections: []CreateSectionRequest{{Name: "s1", DisplayName: "S1"}},
	})

	session, _, err := sessionSvc.Create(ctxA, CreateSessionRequest{
		ResearchID: research.ID, Title: "Private session", Focus: "secrets",
	})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}

	entry, err := entrySvc.Create(ctxA, CreateEntryRequest{
		ResearchID: research.ID,
		SectionID:  sections[0].ID,
		Content:    "# Secret entry\n\nThis is private.",
	})
	if err != nil {
		t.Fatalf("create entry: %v", err)
	}
	if _, err := entrySvc.Update(ctxA, entry.ID, UpdateEntryRequest{Content: ptr("# Secret entry\n\nStill private, now revised.")}); err != nil {
		t.Fatalf("update entry: %v", err)
	}

	t.Run("other user cannot list revisions", func(t *testing.T) {
		if _, _, err := entrySvc.History(ctxB, entry.ID); !errors.Is(err, ErrNotFound) {
			t.Fatalf("expected ErrNotFound, got: %v", err)
		}
	})

	t.Run("other user cannot read a revision", func(t *testing.T) {
		if _, err := entrySvc.Revision(ctxB, entry.ID, 1); !errors.Is(err, ErrNotFound) {
			t.Fatalf("expected ErrNotFound, got: %v", err)
		}
	})

	t.Run("other user cannot diff", func(t *testing.T) {
		if _, err := entrySvc.Diff(ctxB, entry.ID, 1, 2); !errors.Is(err, ErrNotFound) {
			t.Fatalf("expected ErrNotFound, got: %v", err)
		}
	})

	t.Run("other user cannot restore", func(t *testing.T) {
		if _, err := entrySvc.Restore(ctxB, entry.ID, 1); !errors.Is(err, ErrNotFound) {
			t.Fatalf("expected ErrNotFound, got: %v", err)
		}
		// And the entry is untouched.
		current, err := entrySvc.Get(ctxA, entry.ID)
		if err != nil {
			t.Fatalf("get entry: %v", err)
		}
		if !strings.Contains(current.Content, "now revised") {
			t.Fatalf("a refused restore still changed the entry: %q", current.Content)
		}
	})

	t.Run("other user cannot read session changes", func(t *testing.T) {
		if _, err := entrySvc.SessionChanges(ctxB, research.ID, session.ID); !errors.Is(err, ErrNotFound) {
			t.Fatalf("expected ErrNotFound, got: %v", err)
		}
	})

	// The leak this test exists for: entries.session_id is caller-supplied, and
	// the revision history resolves it to a code and a title. Pointing your own
	// entry at a session from someone else's research must not turn your own
	// history into a window onto theirs.
	t.Run("foreign session id on own entry is refused", func(t *testing.T) {
		otherResearch, otherSections, _ := researchSvc.Create(ctxB, CreateResearchRequest{
			Name: "Bob's Research", Goal: "Test",
			Sections: []CreateSectionRequest{{Name: "s1", DisplayName: "S1"}},
		})
		bobSession, _, err := sessionSvc.Create(ctxB, CreateSessionRequest{
			ResearchID: otherResearch.ID, Title: "Acquisition talks with Acme", Focus: "secret",
		})
		if err != nil {
			t.Fatalf("create bob session: %v", err)
		}
		_ = otherSections

		// A writes to A's own entry, which is fully authorized — only the
		// session_id points somewhere they may not see.
		foreign := bobSession.ID
		_, err = entrySvc.Update(ctxA, entry.ID, UpdateEntryRequest{
			Content:   ptr("# Secret entry\n\nProbing."),
			SessionID: &foreign,
		})
		if !errors.Is(err, ErrNotFound) {
			t.Fatalf("expected the write to be refused with ErrNotFound, got: %v", err)
		}

		// And even if such a row existed, the history must not resolve it.
		_, revs, err := entrySvc.History(ctxA, entry.ID)
		if err != nil {
			t.Fatalf("history: %v", err)
		}
		for _, rev := range revs {
			if rev.SessionTitle == "Acquisition talks with Acme" || rev.SessionID == bobSession.ID {
				t.Fatalf("revision r%d leaked another user's session: code=%q title=%q",
					rev.Revision, rev.SessionCode, rev.SessionTitle)
			}
		}
	})

	t.Run("session changes with a foreign session id returns nothing", func(t *testing.T) {
		otherResearch, _, _ := researchSvc.Create(ctxB, CreateResearchRequest{Name: "Bob's Other", Goal: "T"})
		bobSession, _, err := sessionSvc.Create(ctxB, CreateSessionRequest{
			ResearchID: otherResearch.ID, Title: "Bob's session", Focus: "x",
		})
		if err != nil {
			t.Fatalf("create bob session: %v", err)
		}

		// A's own research, B's session id: an empty result, not an error and
		// not another user's changes.
		changes, err := entrySvc.SessionChanges(ctxA, research.ID, bobSession.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(changes) != 0 {
			t.Fatalf("expected no changes, got %d", len(changes))
		}

		// B's research with A's session id: refused outright.
		if _, err := entrySvc.SessionChanges(ctxA, otherResearch.ID, session.ID); !errors.Is(err, ErrNotFound) {
			t.Fatalf("expected ErrNotFound, got: %v", err)
		}
	})

	t.Run("owner can read all of it", func(t *testing.T) {
		_, revs, err := entrySvc.History(ctxA, entry.ID)
		if err != nil || len(revs) != 2 {
			t.Fatalf("owner history: %d revisions, err=%v", len(revs), err)
		}
		if _, err := entrySvc.Diff(ctxA, entry.ID, 1, 2); err != nil {
			t.Fatalf("owner diff: %v", err)
		}
		changes, err := entrySvc.SessionChanges(ctxA, research.ID, session.ID)
		if err != nil {
			t.Fatalf("owner session changes: %v", err)
		}
		if len(changes) == 0 {
			t.Fatal("owner should see what the session changed")
		}
	})
}

func TestAccessControl_ObsidianVault(t *testing.T) {
	db := setupTestDB(t)
	notifier := &mockNotifier{}
	log := slog.Default()

	userRepo := storage.NewUserRepository(db)
	researchRepo := storage.NewResearchRepository(db)
	sectionRepo := storage.NewSectionRepository(db)
	entryRepo := storage.NewEntryRepository(db)
	blockRepo := storage.NewBlockRepository(db)
	crossrefRepo := storage.NewCrossRefRepository(db)
	sessionRepo := storage.NewSessionRepository(db)
	questionRepo := storage.NewQuestionRepository(db)
	taskRepo := storage.NewTaskRepository(db)
	roadmapRepo := storage.NewRoadmapRepository(db)
	roadmapNodeRepo := storage.NewRoadmapNodeRepository(db)
	roadmapEdgeRepo := storage.NewRoadmapEdgeRepository(db)
	revisionRepo := storage.NewEntryRevisionRepository(db)

	researchSvc := NewResearchService(researchRepo, sectionRepo, notifier, log)
	sectionSvc := NewSectionService(sectionRepo, entryRepo, researchRepo, notifier, log)
	entrySvc := NewEntryService(entryRepo, sectionRepo, researchRepo, sessionRepo, blockRepo, revisionRepo, crossrefRepo, nil, notifier, log)
	sessionSvc := NewSessionService(db, sessionRepo, questionRepo, researchRepo, entrySvc, notifier, log)
	taskSvc := NewTaskService(taskRepo, researchRepo, entrySvc, notifier, log)
	roadmapSvc := NewRoadmapService(roadmapRepo, roadmapNodeRepo, roadmapEdgeRepo, researchRepo, notifier, log)
	obsidianSvc := NewObsidianService(researchSvc, sectionSvc, entryRepo, sessionSvc, taskSvc, roadmapSvc, revisionRepo, log)

	userA, userB := setupTwoUsers(t, userRepo)
	ctxA, ctxB := userCtx(userA), userCtx(userB)

	research, sections, _ := researchSvc.Create(ctxA, CreateResearchRequest{
		Name: "Alice's Research", Goal: "Test",
		Sections: []CreateSectionRequest{{Name: "s1", DisplayName: "S1"}},
	})
	if _, err := entrySvc.Create(ctxA, CreateEntryRequest{
		ResearchID: research.ID,
		SectionID:  sections[0].ID,
		Title:      "Secret",
		Content:    "This is private.",
	}); err != nil {
		t.Fatalf("create entry: %v", err)
	}

	t.Run("the owner gets a vault", func(t *testing.T) {
		v, err := obsidianSvc.Vault(ctxA, research.ID, DefaultVaultOptions())
		if err != nil {
			t.Fatalf("owner cannot export: %v", err)
		}
		if len(v.Files) == 0 {
			t.Fatal("owner's vault is empty")
		}
	})

	// An export is the widest read in the product: one call returns every entry,
	// session, question, task and roadmap at once. If ownership fails anywhere,
	// it fails here first.
	t.Run("another user gets nothing", func(t *testing.T) {
		for _, id := range []string{research.ID, research.Code} {
			v, err := obsidianSvc.Vault(ctxB, id, DefaultVaultOptions())
			if err == nil {
				t.Fatalf("exported another user's research by %q: %d files", id, len(v.Files))
			}
			if !errors.Is(err, ErrNotFound) {
				t.Errorf("by %q: got %v, want ErrNotFound — a 403 would confirm the research exists", id, err)
			}
		}
	})
}

// TestAccessControl_ListQuestions covers the read path that had no owner check:
// question_list passes a caller-supplied session id straight to the service, so
// anyone holding another user's session UUID could read its questions and their
// answers.
func TestAccessControl_ListQuestions(t *testing.T) {
	db := setupTestDB(t)
	notifier := &mockNotifier{}
	log := slog.Default()

	userRepo := storage.NewUserRepository(db)
	researchRepo := storage.NewResearchRepository(db)
	sectionRepo := storage.NewSectionRepository(db)
	entryRepo := storage.NewEntryRepository(db)
	blockRepo := storage.NewBlockRepository(db)
	crossrefRepo := storage.NewCrossRefRepository(db)
	sessionRepo := storage.NewSessionRepository(db)
	questionRepo := storage.NewQuestionRepository(db)

	researchSvc := NewResearchService(researchRepo, sectionRepo, notifier, log)
	entrySvc := NewEntryService(entryRepo, sectionRepo, researchRepo, sessionRepo, blockRepo, storage.NewEntryRevisionRepository(db), crossrefRepo, nil, notifier, log)
	sessionSvc := NewSessionService(db, sessionRepo, questionRepo, researchRepo, entrySvc, notifier, log)

	userA, userB := setupTwoUsers(t, userRepo)
	ctxA, ctxB := userCtx(userA), userCtx(userB)

	research, _, _ := researchSvc.Create(ctxA, CreateResearchRequest{Name: "Alice's Research", Goal: "Test"})
	session, _, err := sessionSvc.Create(ctxA, CreateSessionRequest{
		ResearchID: research.ID,
		Title:      "Private session",
		Questions:  []CreateQuestionRequest{{Text: "ALICE PRIVATE QUESTION", Priority: domain.PriorityHigh}},
	})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}

	t.Run("the owner reads their questions", func(t *testing.T) {
		qs, err := sessionSvc.ListQuestions(ctxA, session.ID, storage.QuestionFilter{})
		if err != nil || len(qs) != 1 {
			t.Fatalf("owner got %d questions, err %v", len(qs), err)
		}
	})

	t.Run("another user holding the session id gets nothing", func(t *testing.T) {
		qs, err := sessionSvc.ListQuestions(ctxB, session.ID, storage.QuestionFilter{})
		if err == nil {
			t.Fatalf("read another user's questions: %d returned", len(qs))
		}
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("got %v, want ErrNotFound", err)
		}
		if len(qs) != 0 {
			t.Errorf("a refused read still returned %d questions", len(qs))
		}
	})
}
