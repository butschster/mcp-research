package service

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/butschster/mcp-research/internal/domain"
	"github.com/butschster/mcp-research/internal/storage"
	"github.com/google/uuid"
)

type CreateSessionRequest struct {
	ResearchID string
	Title      string
	Focus      string
	Questions  []CreateQuestionRequest
}

type CreateQuestionRequest struct {
	Text      string
	Area      string
	Rationale string
	Priority  domain.Priority
	ParentID  string
	Position  int
}

type UpdateSessionRequest struct {
	Title   *string
	Focus   *string
	Status  *domain.SessionStatus
	Notes   *string  // replace notes
	AddNote *string  // append to notes
}

type SessionWithQuestions struct {
	Session  *domain.Session
	Questions []*domain.Question
	Progress  QuestionProgress
}

type QuestionProgress struct {
	Total    int `json:"total"`
	Answered int `json:"answered"`
	Pending  int `json:"pending"`
	Deferred int `json:"deferred"`
	Skipped  int `json:"skipped"`
}

type SessionService struct {
	db         *sql.DB
	sessions   *storage.SessionRepository
	questions  *storage.QuestionRepository
	researches *storage.ResearchRepository
	crossrefs  CrossRefParser
	events     EventNotifier
	log        *slog.Logger
}

func NewSessionService(db *sql.DB, sessions *storage.SessionRepository, questions *storage.QuestionRepository, researches *storage.ResearchRepository, crossrefs CrossRefParser, events EventNotifier, log *slog.Logger) *SessionService {
	return &SessionService{db: db, sessions: sessions, questions: questions, researches: researches, crossrefs: crossrefs, events: events, log: log}
}

func (s *SessionService) Create(ctx context.Context, req CreateSessionRequest) (*domain.Session, []*domain.Question, error) {
	if err := validateResearchAccess(ctx, s.researches, req.ResearchID); err != nil {
		return nil, nil, fmt.Errorf("research %s: %w", req.ResearchID, err)
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	session := &domain.Session{
		ID:         uuid.New().String(),
		ResearchID: req.ResearchID,
		Title:      req.Title,
		Focus:      req.Focus,
		Status:     domain.SessionActive,
	}

	if err := s.sessions.CreateTx(ctx, tx, session); err != nil {
		return nil, nil, fmt.Errorf("create session: %w", err)
	}

	var questions []*domain.Question
	for _, qr := range req.Questions {
		q := &domain.Question{
			ID:        uuid.New().String(),
			SessionID: session.ID,
			Text:      qr.Text,
			Area:      qr.Area,
			Rationale: qr.Rationale,
			Priority:  qr.Priority,
			Status:    domain.QuestionPending,
			ParentID:  qr.ParentID,
			Position:  qr.Position,
		}
		if q.Priority == "" {
			q.Priority = domain.PriorityMedium
		}
		questions = append(questions, q)
	}

	if len(questions) > 0 {
		if err := s.questions.CreateBatchTx(ctx, tx, questions); err != nil {
			return nil, nil, fmt.Errorf("create questions: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, nil, fmt.Errorf("commit: %w", err)
	}

	s.events.Notify(Event{Type: "session.created", ResearchID: session.ResearchID, EntityID: session.ID, Entity: "session"})
	return session, questions, nil
}

func (s *SessionService) Get(ctx context.Context, id string) (*SessionWithQuestions, error) {
	session, err := s.sessions.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("find session: %w", err)
	}
	if session == nil {
		// Try by short code
		session, err = s.sessions.FindByCode(ctx, id)
		if err != nil {
			return nil, fmt.Errorf("find session by code: %w", err)
		}
	}
	if session == nil {
		return nil, ErrNotFound
	}
	id = session.ID
	if err := validateResearchAccess(ctx, s.researches, session.ResearchID); err != nil {
		return nil, ErrNotFound
	}

	questions, err := s.questions.FindBySession(ctx, id, storage.QuestionFilter{})
	if err != nil {
		return nil, fmt.Errorf("find questions: %w", err)
	}

	counts, err := s.questions.CountByStatus(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("count questions: %w", err)
	}

	total := 0
	for _, c := range counts {
		total += c
	}

	progress := QuestionProgress{
		Total:    total,
		Answered: counts[domain.QuestionAnswered],
		Pending:  counts[domain.QuestionPending] + counts[domain.QuestionInProgress],
		Deferred: counts[domain.QuestionDeferred],
		Skipped:  counts[domain.QuestionSkipped],
	}

	return &SessionWithQuestions{
		Session:   session,
		Questions: questions,
		Progress:  progress,
	}, nil
}

func (s *SessionService) GetByIDOrCode(ctx context.Context, researchID string, idOrCode string) (*SessionWithQuestions, error) {
	if err := validateResearchAccess(ctx, s.researches, researchID); err != nil {
		return nil, ErrNotFound
	}

	session, err := s.sessions.FindByID(ctx, idOrCode)
	if err != nil {
		return nil, fmt.Errorf("find session: %w", err)
	}
	if session == nil {
		session, err = s.sessions.FindByCodeAndResearch(ctx, idOrCode, researchID)
		if err != nil {
			return nil, fmt.Errorf("find session by code: %w", err)
		}
	}
	if session == nil {
		return nil, ErrNotFound
	}
	if session.ResearchID != researchID {
		return nil, ErrNotFound
	}

	questions, err := s.questions.FindBySession(ctx, session.ID, storage.QuestionFilter{})
	if err != nil {
		return nil, fmt.Errorf("find questions: %w", err)
	}

	counts, err := s.questions.CountByStatus(ctx, session.ID)
	if err != nil {
		return nil, fmt.Errorf("count questions: %w", err)
	}

	total := 0
	for _, c := range counts {
		total += c
	}

	progress := QuestionProgress{
		Total:    total,
		Answered: counts[domain.QuestionAnswered],
		Pending:  counts[domain.QuestionPending] + counts[domain.QuestionInProgress],
		Deferred: counts[domain.QuestionDeferred],
		Skipped:  counts[domain.QuestionSkipped],
	}

	return &SessionWithQuestions{
		Session:   session,
		Questions: questions,
		Progress:  progress,
	}, nil
}

func (s *SessionService) Update(ctx context.Context, id string, req UpdateSessionRequest) (*domain.Session, error) {
	if req.Notes != nil && req.AddNote != nil {
		return nil, fmt.Errorf("notes and add_note: %w", ErrMutualExclusion)
	}

	session, err := s.sessions.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("find session: %w", err)
	}
	if session == nil {
		return nil, ErrNotFound
	}
	if err := validateResearchAccess(ctx, s.researches, session.ResearchID); err != nil {
		return nil, ErrNotFound
	}

	if req.Title != nil {
		session.Title = *req.Title
	}
	if req.Focus != nil {
		session.Focus = *req.Focus
	}
	if req.Status != nil {
		session.Status = *req.Status
	}
	if req.Notes != nil {
		session.Notes = *req.Notes
	}
	if req.AddNote != nil {
		if session.Notes != "" {
			session.Notes += "\n"
		}
		session.Notes += *req.AddNote
	}

	if err := s.sessions.Update(ctx, session); err != nil {
		return nil, fmt.Errorf("update session: %w", err)
	}

	s.events.Notify(Event{Type: "session.updated", ResearchID: session.ResearchID, EntityID: session.ID, Entity: "session"})
	return session, nil
}

func (s *SessionService) ListByResearch(ctx context.Context, researchID string) ([]*domain.Session, error) {
	if err := validateResearchAccess(ctx, s.researches, researchID); err != nil {
		return nil, err
	}
	return s.sessions.FindByResearch(ctx, researchID)
}

func (s *SessionService) FindActive(ctx context.Context, researchID string) (*domain.Session, error) {
	return s.sessions.FindActive(ctx, researchID)
}

func (s *SessionService) FindLatest(ctx context.Context, researchID string) (*domain.Session, error) {
	return s.sessions.FindLatest(ctx, researchID)
}

func (s *SessionService) AddQuestions(ctx context.Context, sessionID string, requests []CreateQuestionRequest) ([]*domain.Question, error) {
	session, err := s.sessions.FindByID(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("find session: %w", err)
	}
	if session == nil {
		return nil, ErrNotFound
	}
	if err := validateResearchAccess(ctx, s.researches, session.ResearchID); err != nil {
		return nil, ErrNotFound
	}

	var questions []*domain.Question
	for _, qr := range requests {
		// Check depth limit for child questions
		if qr.ParentID != "" {
			depth, err := s.questions.GetDepth(ctx, qr.ParentID)
			if err != nil {
				return nil, fmt.Errorf("check depth: %w", err)
			}
			if depth >= 3 {
				return nil, ErrQuestionDepthLimit
			}
		}

		q := &domain.Question{
			ID:        uuid.New().String(),
			SessionID: sessionID,
			Text:      qr.Text,
			Area:      qr.Area,
			Rationale: qr.Rationale,
			Priority:  qr.Priority,
			Status:    domain.QuestionPending,
			ParentID:  qr.ParentID,
			Position:  qr.Position,
		}
		if q.Priority == "" {
			q.Priority = domain.PriorityMedium
		}
		questions = append(questions, q)
	}

	if err := s.questions.CreateBatch(ctx, questions); err != nil {
		return nil, fmt.Errorf("create questions: %w", err)
	}

	s.events.Notify(Event{Type: "question.created", ResearchID: session.ResearchID, EntityID: sessionID, Entity: "question"})
	return questions, nil
}

func (s *SessionService) UpdateQuestion(ctx context.Context, id string, status *domain.QuestionStatus, answer *string) (*domain.Question, error) {
	question, err := s.questions.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("find question: %w", err)
	}
	if question == nil {
		return nil, ErrNotFound
	}

	if status != nil {
		question.Status = *status
	}
	if answer != nil {
		question.Answer = *answer
	}

	// Validate: answered requires non-empty answer
	if question.Status == domain.QuestionAnswered && question.Answer == "" {
		return nil, ErrAnswerRequired
	}

	if err := s.questions.Update(ctx, question); err != nil {
		return nil, fmt.Errorf("update question: %w", err)
	}

	// Parse crossrefs from answer text
	if s.crossrefs != nil && question.Answer != "" {
		session, _ := s.sessions.FindByID(ctx, question.SessionID)
		if session != nil {
			s.crossrefs.ParseCrossRefs(ctx, "question", question.ID, session.ResearchID, question.Answer)
		}
	}

	s.events.Notify(Event{Type: "question.updated", EntityID: question.ID, Entity: "question"})
	return question, nil
}

func (s *SessionService) ListQuestions(ctx context.Context, sessionID string, filter storage.QuestionFilter) ([]*domain.Question, error) {
	return s.questions.FindBySession(ctx, sessionID, filter)
}
