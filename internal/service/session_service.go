package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/dovod-app/app/internal/domain"
	"github.com/dovod-app/app/internal/storage"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
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
	Notes   *string // replace notes
	AddNote *string // append to notes
}

type SessionWithQuestions struct {
	Session   *domain.Session
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
	db         *bun.DB
	sessions   *storage.SessionRepository
	questions  *storage.QuestionRepository
	researches *storage.ResearchRepository
	access     *Access
	crossrefs  CrossRefParser
	events     EventNotifier
	log        *slog.Logger
}

func NewSessionService(db *bun.DB, sessions *storage.SessionRepository, questions *storage.QuestionRepository, researches *storage.ResearchRepository, access *Access, crossrefs CrossRefParser, events EventNotifier, log *slog.Logger) *SessionService {
	return &SessionService{db: db, sessions: sessions, questions: questions, researches: researches, access: access, crossrefs: crossrefs, events: events, log: log}
}

func (s *SessionService) Create(ctx context.Context, req CreateSessionRequest) (*domain.Session, []*domain.Question, error) {
	if err := s.access.Write(ctx, req.ResearchID); err != nil {
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
		Title:      normalizeTitle(req.Title),
		Focus:      normalizeContent(req.Focus),
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
			Text:      normalizeContent(qr.Text),
			Area:      qr.Area,
			Rationale: normalizeContent(qr.Rationale),
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

	emit(ctx, s.events, Event{Type: "session.created", ResearchID: session.ResearchID, EntityID: session.ID, Entity: "session"})
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
	if err := s.access.Read(ctx, session.ResearchID); err != nil {
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
	if err := s.access.Read(ctx, researchID); err != nil {
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
	if err := s.access.Write(ctx, session.ResearchID); err != nil {
		return nil, err
	}

	if req.Title != nil {
		session.Title = normalizeTitle(*req.Title)
	}
	if req.Focus != nil {
		session.Focus = normalizeContent(*req.Focus)
	}
	if req.Status != nil {
		session.Status = *req.Status
	}
	if req.Notes != nil {
		session.Notes = normalizeContent(*req.Notes)
	}
	if req.AddNote != nil {
		note := normalizeContent(*req.AddNote)
		if session.Notes != "" {
			session.Notes += "\n"
		}
		session.Notes += note
	}

	if err := s.sessions.Update(ctx, session); err != nil {
		return nil, fmt.Errorf("update session: %w", err)
	}

	emit(ctx, s.events, Event{Type: "session.updated", ResearchID: session.ResearchID, EntityID: session.ID, Entity: "session"})
	return session, nil
}

func (s *SessionService) ListByResearch(ctx context.Context, researchID string) ([]*domain.Session, error) {
	if err := s.access.Read(ctx, researchID); err != nil {
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
	if err := s.access.Write(ctx, session.ResearchID); err != nil {
		return nil, err
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
			Text:      normalizeContent(qr.Text),
			Area:      qr.Area,
			Rationale: normalizeContent(qr.Rationale),
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

	// One event per question, each naming the question. This used to send a
	// single event carrying the *session* id, which made twelve new questions
	// indistinguishable from one and left no way to react to a particular one —
	// every other event in the system names the thing that changed.
	for _, q := range questions {
		emit(ctx, s.events, Event{Type: "question.created", ResearchID: session.ResearchID, EntityID: q.ID, Entity: "question"})
	}
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

	// `question_update` takes a caller-supplied question id and writes an
	// answer to it. Without this, holding a question's UUID was enough to
	// write into someone else's research — the same hole its sibling
	// ListQuestions carried, one method over.
	session, err := s.sessions.FindByID(ctx, question.SessionID)
	if err != nil {
		return nil, fmt.Errorf("find session: %w", err)
	}
	if session == nil {
		return nil, ErrNotFound
	}
	if err := s.access.Write(ctx, session.ResearchID); err != nil {
		return nil, err
	}

	if status != nil {
		question.Status = *status
	}
	if answer != nil {
		question.Answer = normalizeContent(*answer)
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
		s.crossrefs.ParseCrossRefs(ctx, "question", question.ID, session.ResearchID, question.Answer)
	}

	emit(ctx, s.events, Event{Type: "question.updated", ResearchID: session.ResearchID, EntityID: question.ID, Entity: "question"})
	return question, nil
}

// ListQuestions returns a session's questions, for a caller who owns the
// research the session belongs to.
//
// The ownership check is the point: `question_list` passes a caller-supplied
// session id straight through, so without it anyone holding a session's UUID
// could read another user's questions and answers. Every sibling method checks;
// this one did not, and a read path is exactly where that is hardest to notice.
func (s *SessionService) ListQuestions(ctx context.Context, sessionID string, filter storage.QuestionFilter) ([]*domain.Question, error) {
	session, err := s.sessions.FindByID(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("find session: %w", err)
	}
	if session == nil {
		return nil, ErrNotFound
	}
	if err := s.access.Read(ctx, session.ResearchID); err != nil {
		return nil, ErrNotFound
	}
	return s.questions.FindBySession(ctx, sessionID, filter)
}
