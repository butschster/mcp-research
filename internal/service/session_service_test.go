package service

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/butschster/mcp-research/internal/domain"
	"github.com/butschster/mcp-research/internal/storage"
)

func TestSessionService_Create(t *testing.T) {
	db := setupTestDB(t)
	notifier := &mockNotifier{}
	svc := NewSessionService(
		db,
		storage.NewSessionRepository(db),
		storage.NewQuestionRepository(db),
		storage.NewResearchRepository(db),
		nil,
		notifier,
		slog.Default(),
	)
	ctx := context.Background()

	t.Run("creates session with questions", func(t *testing.T) {
		r := createTestResearch(t, db)
		notifier.reset()

		session, questions, err := svc.Create(ctx, CreateSessionRequest{
			ResearchID: r.ID,
			Title:      "Interview Session",
			Focus:      "User experience",
			Questions: []CreateQuestionRequest{
				{Text: "What do you think?", Area: "UX", Rationale: "baseline"},
				{Text: "How do you feel?", Area: "UX", Priority: domain.PriorityHigh},
			},
		})
		if err != nil {
			t.Fatalf("create: %v", err)
		}
		if session.ID == "" {
			t.Fatal("expected non-empty ID")
		}
		if session.Status != domain.SessionActive {
			t.Errorf("expected status %q, got %q", domain.SessionActive, session.Status)
		}
		if len(questions) != 2 {
			t.Fatalf("expected 2 questions, got %d", len(questions))
		}
		// First question should default to medium priority
		if questions[0].Priority != domain.PriorityMedium {
			t.Errorf("expected default priority %q, got %q", domain.PriorityMedium, questions[0].Priority)
		}
		// Second question should keep provided priority
		if questions[1].Priority != domain.PriorityHigh {
			t.Errorf("expected priority %q, got %q", domain.PriorityHigh, questions[1].Priority)
		}
		if !notifier.hasEvent("session.created") {
			t.Error("expected session.created event")
		}
	})

	t.Run("creates session without questions", func(t *testing.T) {
		r := createTestResearch(t, db)

		session, questions, err := svc.Create(ctx, CreateSessionRequest{
			ResearchID: r.ID,
			Title:      "Empty Session",
			Focus:      "General",
		})
		if err != nil {
			t.Fatalf("create: %v", err)
		}
		if session.Status != domain.SessionActive {
			t.Errorf("expected status %q, got %q", domain.SessionActive, session.Status)
		}
		if len(questions) != 0 {
			t.Errorf("expected 0 questions, got %d", len(questions))
		}
	})
}

func TestSessionService_Get(t *testing.T) {
	db := setupTestDB(t)
	svc := NewSessionService(
		db,
		storage.NewSessionRepository(db),
		storage.NewQuestionRepository(db),
		storage.NewResearchRepository(db),
		nil,
		&mockNotifier{},
		slog.Default(),
	)
	ctx := context.Background()

	t.Run("returns session with questions and progress", func(t *testing.T) {
		r := createTestResearch(t, db)
		session, _, _ := svc.Create(ctx, CreateSessionRequest{
			ResearchID: r.ID,
			Title:      "Get Test",
			Focus:      "Focus",
			Questions: []CreateQuestionRequest{
				{Text: "Q1", Area: "A"},
				{Text: "Q2", Area: "A"},
			},
		})

		got, err := svc.Get(ctx, session.ID)
		if err != nil {
			t.Fatalf("get: %v", err)
		}
		if got.Session.Title != "Get Test" {
			t.Errorf("expected title %q, got %q", "Get Test", got.Session.Title)
		}
		if len(got.Questions) != 2 {
			t.Errorf("expected 2 questions, got %d", len(got.Questions))
		}
		if got.Progress.Total != 2 {
			t.Errorf("expected total 2, got %d", got.Progress.Total)
		}
		if got.Progress.Pending != 2 {
			t.Errorf("expected pending 2, got %d", got.Progress.Pending)
		}
	})

	t.Run("not found", func(t *testing.T) {
		_, err := svc.Get(ctx, "nonexistent")
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})
}

func TestSessionService_Update(t *testing.T) {
	db := setupTestDB(t)
	notifier := &mockNotifier{}
	svc := NewSessionService(
		db,
		storage.NewSessionRepository(db),
		storage.NewQuestionRepository(db),
		storage.NewResearchRepository(db),
		nil,
		notifier,
		slog.Default(),
	)
	ctx := context.Background()

	t.Run("partial updates", func(t *testing.T) {
		r := createTestResearch(t, db)
		session, _, _ := svc.Create(ctx, CreateSessionRequest{
			ResearchID: r.ID, Title: "Original", Focus: "Original Focus",
		})
		notifier.reset()

		updated, err := svc.Update(ctx, session.ID, UpdateSessionRequest{
			Title:  ptr("Updated Title"),
			Focus:  ptr("Updated Focus"),
			Status: ptr(domain.SessionCompleted),
			Notes:  ptr("some notes"),
		})
		if err != nil {
			t.Fatalf("update: %v", err)
		}
		if updated.Title != "Updated Title" {
			t.Errorf("expected title %q, got %q", "Updated Title", updated.Title)
		}
		if updated.Focus != "Updated Focus" {
			t.Errorf("expected focus %q, got %q", "Updated Focus", updated.Focus)
		}
		if updated.Status != domain.SessionCompleted {
			t.Errorf("expected status %q, got %q", domain.SessionCompleted, updated.Status)
		}
		if updated.Notes != "some notes" {
			t.Errorf("expected notes %q, got %q", "some notes", updated.Notes)
		}
		if !notifier.hasEvent("session.updated") {
			t.Error("expected session.updated event")
		}
	})

	t.Run("add note appends", func(t *testing.T) {
		r := createTestResearch(t, db)
		session, _, _ := svc.Create(ctx, CreateSessionRequest{
			ResearchID: r.ID, Title: "Note Test", Focus: "f",
		})
		_, _ = svc.Update(ctx, session.ID, UpdateSessionRequest{Notes: ptr("first")})

		updated, err := svc.Update(ctx, session.ID, UpdateSessionRequest{AddNote: ptr("second")})
		if err != nil {
			t.Fatalf("update: %v", err)
		}
		if updated.Notes != "first\nsecond" {
			t.Errorf("expected notes %q, got %q", "first\nsecond", updated.Notes)
		}
	})

	t.Run("mutual exclusion notes and add_note", func(t *testing.T) {
		r := createTestResearch(t, db)
		session, _, _ := svc.Create(ctx, CreateSessionRequest{
			ResearchID: r.ID, Title: "MutEx", Focus: "f",
		})

		_, err := svc.Update(ctx, session.ID, UpdateSessionRequest{
			Notes:   ptr("a"),
			AddNote: ptr("b"),
		})
		if !errors.Is(err, ErrMutualExclusion) {
			t.Errorf("expected ErrMutualExclusion, got %v", err)
		}
	})

	t.Run("not found", func(t *testing.T) {
		_, err := svc.Update(ctx, "nonexistent", UpdateSessionRequest{Title: ptr("x")})
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})
}

func TestSessionService_ListByResearch(t *testing.T) {
	db := setupTestDB(t)
	svc := NewSessionService(
		db,
		storage.NewSessionRepository(db),
		storage.NewQuestionRepository(db),
		storage.NewResearchRepository(db),
		nil,
		&mockNotifier{},
		slog.Default(),
	)
	ctx := context.Background()

	r := createTestResearch(t, db)
	_, _, _ = svc.Create(ctx, CreateSessionRequest{ResearchID: r.ID, Title: "S1", Focus: "f"})
	_, _, _ = svc.Create(ctx, CreateSessionRequest{ResearchID: r.ID, Title: "S2", Focus: "f"})

	list, err := svc.ListByResearch(ctx, r.ID)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(list) != 2 {
		t.Errorf("expected 2 sessions, got %d", len(list))
	}
}

func TestSessionService_AddQuestions(t *testing.T) {
	db := setupTestDB(t)
	notifier := &mockNotifier{}
	svc := NewSessionService(
		db,
		storage.NewSessionRepository(db),
		storage.NewQuestionRepository(db),
		storage.NewResearchRepository(db),
		nil,
		notifier,
		slog.Default(),
	)
	ctx := context.Background()

	t.Run("adds questions to session", func(t *testing.T) {
		r := createTestResearch(t, db)
		session, _, _ := svc.Create(ctx, CreateSessionRequest{
			ResearchID: r.ID, Title: "AddQ", Focus: "f",
		})
		notifier.reset()

		questions, err := svc.AddQuestions(ctx, session.ID, []CreateQuestionRequest{
			{Text: "New Q1", Area: "A"},
			{Text: "New Q2", Area: "B", Priority: domain.PriorityLow},
		})
		if err != nil {
			t.Fatalf("add questions: %v", err)
		}
		if len(questions) != 2 {
			t.Fatalf("expected 2 questions, got %d", len(questions))
		}
		if questions[0].Priority != domain.PriorityMedium {
			t.Errorf("expected default priority medium, got %q", questions[0].Priority)
		}
		if questions[1].Priority != domain.PriorityLow {
			t.Errorf("expected priority low, got %q", questions[1].Priority)
		}
		if !notifier.hasEvent("question.created") {
			t.Error("expected question.created event")
		}
	})

	t.Run("depth limit", func(t *testing.T) {
		r := createTestResearch(t, db)
		session, _, _ := svc.Create(ctx, CreateSessionRequest{
			ResearchID: r.ID, Title: "Depth", Focus: "f",
			Questions: []CreateQuestionRequest{
				{Text: "Root Q", Area: "A"},
			},
		})

		// Get the root question
		swq, _ := svc.Get(ctx, session.ID)
		rootQ := swq.Questions[0]

		// Depth 1: child of root
		child1, err := svc.AddQuestions(ctx, session.ID, []CreateQuestionRequest{
			{Text: "Child 1", Area: "A", ParentID: rootQ.ID},
		})
		if err != nil {
			t.Fatalf("add child 1: %v", err)
		}

		// Depth 2: child of child1
		child2, err := svc.AddQuestions(ctx, session.ID, []CreateQuestionRequest{
			{Text: "Child 2", Area: "A", ParentID: child1[0].ID},
		})
		if err != nil {
			t.Fatalf("add child 2: %v", err)
		}

		// Depth 3: child of child2 - still allowed (depth of child2 is 2, < 3)
		child3, err := svc.AddQuestions(ctx, session.ID, []CreateQuestionRequest{
			{Text: "Child 3", Area: "A", ParentID: child2[0].ID},
		})
		if err != nil {
			t.Fatalf("add child 3: %v", err)
		}

		// Depth 4: child of child3 - should hit limit (depth of child3 is 3, >= 3)
		_, err = svc.AddQuestions(ctx, session.ID, []CreateQuestionRequest{
			{Text: "Child 4", Area: "A", ParentID: child3[0].ID},
		})
		if !errors.Is(err, ErrQuestionDepthLimit) {
			t.Errorf("expected ErrQuestionDepthLimit, got %v", err)
		}
	})
}

func TestSessionService_UpdateQuestion(t *testing.T) {
	db := setupTestDB(t)
	notifier := &mockNotifier{}
	svc := NewSessionService(
		db,
		storage.NewSessionRepository(db),
		storage.NewQuestionRepository(db),
		storage.NewResearchRepository(db),
		nil,
		notifier,
		slog.Default(),
	)
	ctx := context.Background()

	t.Run("updates status and answer", func(t *testing.T) {
		r := createTestResearch(t, db)
		session, questions, _ := svc.Create(ctx, CreateSessionRequest{
			ResearchID: r.ID, Title: "UQ", Focus: "f",
			Questions: []CreateQuestionRequest{{Text: "Q?", Area: "A"}},
		})
		_ = session
		notifier.reset()

		updated, err := svc.UpdateQuestion(ctx, questions[0].ID,
			ptr(domain.QuestionAnswered), ptr("The answer"))
		if err != nil {
			t.Fatalf("update question: %v", err)
		}
		if updated.Status != domain.QuestionAnswered {
			t.Errorf("expected status %q, got %q", domain.QuestionAnswered, updated.Status)
		}
		if updated.Answer != "The answer" {
			t.Errorf("expected answer %q, got %q", "The answer", updated.Answer)
		}
		if !notifier.hasEvent("question.updated") {
			t.Error("expected question.updated event")
		}
	})

	t.Run("answered without answer returns error", func(t *testing.T) {
		r := createTestResearch(t, db)
		_, questions, _ := svc.Create(ctx, CreateSessionRequest{
			ResearchID: r.ID, Title: "UQ2", Focus: "f",
			Questions: []CreateQuestionRequest{{Text: "Q?", Area: "A"}},
		})

		_, err := svc.UpdateQuestion(ctx, questions[0].ID,
			ptr(domain.QuestionAnswered), nil)
		if !errors.Is(err, ErrAnswerRequired) {
			t.Errorf("expected ErrAnswerRequired, got %v", err)
		}
	})

	t.Run("not found", func(t *testing.T) {
		_, err := svc.UpdateQuestion(ctx, "nonexistent", ptr(domain.QuestionAnswered), ptr("a"))
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})
}

func TestSessionService_ListQuestions(t *testing.T) {
	db := setupTestDB(t)
	svc := NewSessionService(
		db,
		storage.NewSessionRepository(db),
		storage.NewQuestionRepository(db),
		storage.NewResearchRepository(db),
		nil,
		&mockNotifier{},
		slog.Default(),
	)
	ctx := context.Background()

	r := createTestResearch(t, db)
	session, _, _ := svc.Create(ctx, CreateSessionRequest{
		ResearchID: r.ID, Title: "LQ", Focus: "f",
		Questions: []CreateQuestionRequest{
			{Text: "Q1", Area: "A", Priority: domain.PriorityHigh},
			{Text: "Q2", Area: "B", Priority: domain.PriorityLow},
		},
	})

	t.Run("returns all questions", func(t *testing.T) {
		list, err := svc.ListQuestions(ctx, session.ID, storage.QuestionFilter{})
		if err != nil {
			t.Fatalf("list: %v", err)
		}
		if len(list) != 2 {
			t.Errorf("expected 2 questions, got %d", len(list))
		}
	})

	t.Run("filters by priority", func(t *testing.T) {
		list, err := svc.ListQuestions(ctx, session.ID, storage.QuestionFilter{
			Priority: ptr(domain.PriorityHigh),
		})
		if err != nil {
			t.Fatalf("list: %v", err)
		}
		if len(list) != 1 {
			t.Errorf("expected 1 question, got %d", len(list))
		}
	})
}
