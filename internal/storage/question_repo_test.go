package storage

import (
	"context"
	"testing"
	"time"

	"github.com/butschster/mcp-research/internal/domain"
	"github.com/google/uuid"
)

type questionTestFixture struct {
	questionRepo *QuestionRepository
	research     *domain.Research
	session      *domain.Session
}

func setupQuestionTest(t *testing.T) (*questionTestFixture, context.Context) {
	t.Helper()
	db := setupTestDB(t)
	ctx := context.Background()
	researchRepo := NewResearchRepository(db)
	research := createTestResearch(t, researchRepo, ctx)

	sessionRepo := NewSessionRepository(db)
	session := &domain.Session{
		ID:         uuid.New().String(),
		ResearchID: research.ID,
		Title:      "Test Session",
		Focus:      "Testing",
		Status:     domain.SessionActive,
	}
	if err := sessionRepo.Create(ctx, session); err != nil {
		t.Fatalf("create test session: %v", err)
	}

	return &questionTestFixture{
		questionRepo: NewQuestionRepository(db),
		research:     research,
		session:      session,
	}, ctx
}

func TestQuestionRepository_CreateBatch(t *testing.T) {
	f, ctx := setupQuestionTest(t)

	parentID := uuid.New().String()
	childID := uuid.New().String()

	questions := []*domain.Question{
		{
			ID:        parentID,
			SessionID: f.session.ID,
			Text:      "Parent question?",
			Area:      "methodology",
			Rationale: "Need to understand approach",
			Priority:  domain.PriorityHigh,
			Status:    domain.QuestionPending,
			Position:  0,
		},
		{
			ID:        childID,
			SessionID: f.session.ID,
			Text:      "Child question?",
			Area:      "methodology",
			Rationale: "Follow-up",
			Priority:  domain.PriorityMedium,
			Status:    domain.QuestionPending,
			ParentID:  parentID,
			Position:  1,
		},
	}

	if err := f.questionRepo.CreateBatch(ctx, questions); err != nil {
		t.Fatalf("CreateBatch: %v", err)
	}

	// Verify parent
	parent, err := f.questionRepo.FindByID(ctx, parentID)
	if err != nil {
		t.Fatalf("FindByID parent: %v", err)
	}
	if parent == nil {
		t.Fatal("parent not found")
	}
	if parent.Text != "Parent question?" {
		t.Errorf("parent Text: got %s, want 'Parent question?'", parent.Text)
	}
	if parent.CreatedAt.IsZero() {
		t.Error("parent CreatedAt should not be zero")
	}
	if parent.ParentID != "" {
		t.Errorf("parent ParentID: got %s, want empty", parent.ParentID)
	}

	// Verify child
	child, err := f.questionRepo.FindByID(ctx, childID)
	if err != nil {
		t.Fatalf("FindByID child: %v", err)
	}
	if child == nil {
		t.Fatal("child not found")
	}
	if child.ParentID != parentID {
		t.Errorf("child ParentID: got %s, want %s", child.ParentID, parentID)
	}
}

func TestQuestionRepository_CreateBatchTx(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()
	researchRepo := NewResearchRepository(db)
	research := createTestResearch(t, researchRepo, ctx)

	sessionRepo := NewSessionRepository(db)
	session := &domain.Session{
		ID:         uuid.New().String(),
		ResearchID: research.ID,
		Title:      "TX Session",
		Focus:      "Testing",
		Status:     domain.SessionActive,
	}
	if err := sessionRepo.Create(ctx, session); err != nil {
		t.Fatalf("create session: %v", err)
	}

	questionRepo := NewQuestionRepository(db)

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("BeginTx: %v", err)
	}

	questions := []*domain.Question{
		{
			ID:        uuid.New().String(),
			SessionID: session.ID,
			Text:      "TX question?",
			Area:      "design",
			Priority:  domain.PriorityLow,
			Status:    domain.QuestionPending,
			Position:  0,
		},
	}

	if err := questionRepo.CreateBatchTx(ctx, tx, questions); err != nil {
		tx.Rollback()
		t.Fatalf("CreateBatchTx: %v", err)
	}

	if err := tx.Commit(); err != nil {
		t.Fatalf("Commit: %v", err)
	}

	found, err := questionRepo.FindByID(ctx, questions[0].ID)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if found == nil {
		t.Fatal("question not found after CreateBatchTx + Commit")
	}
	if found.Text != "TX question?" {
		t.Errorf("Text: got %s, want 'TX question?'", found.Text)
	}
}

func TestQuestionRepository_Update(t *testing.T) {
	f, ctx := setupQuestionTest(t)

	q := &domain.Question{
		ID:        uuid.New().String(),
		SessionID: f.session.ID,
		Text:      "Original question?",
		Area:      "design",
		Priority:  domain.PriorityMedium,
		Status:    domain.QuestionPending,
		Position:  0,
	}

	questions := []*domain.Question{q}
	if err := f.questionRepo.CreateBatch(ctx, questions); err != nil {
		t.Fatalf("CreateBatch: %v", err)
	}

	q.Text = "Updated question?"
	q.Status = domain.QuestionAnswered
	q.Answer = "This is the answer"
	q.Area = "implementation"

	if err := f.questionRepo.Update(ctx, q); err != nil {
		t.Fatalf("Update: %v", err)
	}

	found, err := f.questionRepo.FindByID(ctx, q.ID)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if found.Text != "Updated question?" {
		t.Errorf("Text: got %s, want 'Updated question?'", found.Text)
	}
	if found.Status != domain.QuestionAnswered {
		t.Errorf("Status: got %s, want answered", found.Status)
	}
	if found.Answer != "This is the answer" {
		t.Errorf("Answer: got %s, want 'This is the answer'", found.Answer)
	}
	if found.Area != "implementation" {
		t.Errorf("Area: got %s, want implementation", found.Area)
	}
}

func TestQuestionRepository_FindByID(t *testing.T) {
	f, ctx := setupQuestionTest(t)

	q := &domain.Question{
		ID:        uuid.New().String(),
		SessionID: f.session.ID,
		Text:      "Find me?",
		Area:      "testing",
		Rationale: "test rationale",
		Priority:  domain.PriorityHigh,
		Status:    domain.QuestionPending,
		Position:  0,
	}
	if err := f.questionRepo.CreateBatch(ctx, []*domain.Question{q}); err != nil {
		t.Fatalf("CreateBatch: %v", err)
	}

	found, err := f.questionRepo.FindByID(ctx, q.ID)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if found == nil {
		t.Fatal("FindByID returned nil")
	}
	if found.Rationale != "test rationale" {
		t.Errorf("Rationale: got %s, want 'test rationale'", found.Rationale)
	}
	if found.Priority != domain.PriorityHigh {
		t.Errorf("Priority: got %s, want high", found.Priority)
	}
}

func TestQuestionRepository_FindBySession(t *testing.T) {
	f, ctx := setupQuestionTest(t)

	questions := []*domain.Question{
		{ID: uuid.New().String(), SessionID: f.session.ID, Text: "Q1?", Area: "design", Priority: domain.PriorityHigh, Status: domain.QuestionPending, Position: 2},
		{ID: uuid.New().String(), SessionID: f.session.ID, Text: "Q2?", Area: "methodology", Priority: domain.PriorityMedium, Status: domain.QuestionAnswered, Position: 0},
		{ID: uuid.New().String(), SessionID: f.session.ID, Text: "Q3?", Area: "design", Priority: domain.PriorityLow, Status: domain.QuestionPending, Position: 1},
	}
	if err := f.questionRepo.CreateBatch(ctx, questions); err != nil {
		t.Fatalf("CreateBatch: %v", err)
	}

	t.Run("no filter", func(t *testing.T) {
		found, err := f.questionRepo.FindBySession(ctx, f.session.ID, QuestionFilter{})
		if err != nil {
			t.Fatalf("FindBySession: %v", err)
		}
		if len(found) != 3 {
			t.Fatalf("expected 3 questions, got %d", len(found))
		}
		// Ordered by position ASC
		if found[0].Text != "Q2?" {
			t.Errorf("first result: got %s, want Q2?", found[0].Text)
		}
		if found[1].Text != "Q3?" {
			t.Errorf("second result: got %s, want Q3?", found[1].Text)
		}
		if found[2].Text != "Q1?" {
			t.Errorf("third result: got %s, want Q1?", found[2].Text)
		}
	})

	t.Run("filter by status", func(t *testing.T) {
		status := domain.QuestionPending
		found, err := f.questionRepo.FindBySession(ctx, f.session.ID, QuestionFilter{Status: &status})
		if err != nil {
			t.Fatalf("FindBySession with status filter: %v", err)
		}
		if len(found) != 2 {
			t.Errorf("expected 2 pending questions, got %d", len(found))
		}
	})

	t.Run("filter by area", func(t *testing.T) {
		area := "design"
		found, err := f.questionRepo.FindBySession(ctx, f.session.ID, QuestionFilter{Area: &area})
		if err != nil {
			t.Fatalf("FindBySession with area filter: %v", err)
		}
		if len(found) != 2 {
			t.Errorf("expected 2 design questions, got %d", len(found))
		}
	})

	t.Run("filter by priority", func(t *testing.T) {
		priority := domain.PriorityHigh
		found, err := f.questionRepo.FindBySession(ctx, f.session.ID, QuestionFilter{Priority: &priority})
		if err != nil {
			t.Fatalf("FindBySession with priority filter: %v", err)
		}
		if len(found) != 1 {
			t.Fatalf("expected 1 high-priority question, got %d", len(found))
		}
		if found[0].Text != "Q1?" {
			t.Errorf("Text: got %s, want Q1?", found[0].Text)
		}
	})
}

func TestQuestionRepository_CountByStatus(t *testing.T) {
	f, ctx := setupQuestionTest(t)

	questions := []*domain.Question{
		{ID: uuid.New().String(), SessionID: f.session.ID, Text: "Q1?", Area: "a", Priority: domain.PriorityHigh, Status: domain.QuestionPending, Position: 0},
		{ID: uuid.New().String(), SessionID: f.session.ID, Text: "Q2?", Area: "a", Priority: domain.PriorityMedium, Status: domain.QuestionPending, Position: 1},
		{ID: uuid.New().String(), SessionID: f.session.ID, Text: "Q3?", Area: "a", Priority: domain.PriorityLow, Status: domain.QuestionAnswered, Position: 2},
	}
	if err := f.questionRepo.CreateBatch(ctx, questions); err != nil {
		t.Fatalf("CreateBatch: %v", err)
	}

	counts, err := f.questionRepo.CountByStatus(ctx, f.session.ID)
	if err != nil {
		t.Fatalf("CountByStatus: %v", err)
	}
	if counts[domain.QuestionPending] != 2 {
		t.Errorf("pending count: got %d, want 2", counts[domain.QuestionPending])
	}
	if counts[domain.QuestionAnswered] != 1 {
		t.Errorf("answered count: got %d, want 1", counts[domain.QuestionAnswered])
	}
	if counts[domain.QuestionSkipped] != 0 {
		t.Errorf("skipped count: got %d, want 0", counts[domain.QuestionSkipped])
	}
}

func TestQuestionRepository_GetDepth(t *testing.T) {
	f, ctx := setupQuestionTest(t)

	rootID := uuid.New().String()
	childID := uuid.New().String()
	grandchildID := uuid.New().String()

	// Create root, child, then grandchild in separate batches to respect parent ordering
	root := []*domain.Question{
		{ID: rootID, SessionID: f.session.ID, Text: "Root?", Area: "a", Priority: domain.PriorityHigh, Status: domain.QuestionPending, Position: 0},
	}
	if err := f.questionRepo.CreateBatch(ctx, root); err != nil {
		t.Fatalf("CreateBatch root: %v", err)
	}

	child := []*domain.Question{
		{ID: childID, SessionID: f.session.ID, Text: "Child?", Area: "a", Priority: domain.PriorityMedium, Status: domain.QuestionPending, ParentID: rootID, Position: 1},
	}
	if err := f.questionRepo.CreateBatch(ctx, child); err != nil {
		t.Fatalf("CreateBatch child: %v", err)
	}

	grandchild := []*domain.Question{
		{ID: grandchildID, SessionID: f.session.ID, Text: "Grandchild?", Area: "a", Priority: domain.PriorityLow, Status: domain.QuestionPending, ParentID: childID, Position: 2},
	}
	if err := f.questionRepo.CreateBatch(ctx, grandchild); err != nil {
		t.Fatalf("CreateBatch grandchild: %v", err)
	}

	tests := []struct {
		name     string
		id       string
		expected int
	}{
		{"root depth is 0", rootID, 0},
		{"child depth is 1", childID, 1},
		{"grandchild depth is 2", grandchildID, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			depth, err := f.questionRepo.GetDepth(ctx, tt.id)
			if err != nil {
				t.Fatalf("GetDepth: %v", err)
			}
			if depth != tt.expected {
				t.Errorf("depth: got %d, want %d", depth, tt.expected)
			}
		})
	}
}

func TestQuestionRepository_FindByID_NotFound(t *testing.T) {
	f, ctx := setupQuestionTest(t)

	found, err := f.questionRepo.FindByID(ctx, "nonexistent-id")
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if found != nil {
		t.Error("expected nil for non-existent record")
	}
}

// Ensure time import is used (for potential completedAt scenarios)
var _ = time.Now
