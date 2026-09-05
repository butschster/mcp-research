package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sort"
	"time"

	"github.com/dovod-app/app/internal/auth"
	"github.com/dovod-app/app/internal/domain"
	"github.com/dovod-app/app/internal/storage"
)

// ResumeService answers one question: a new chat has opened on an existing
// research — what was being worked on, and what is the next defensible move?
//
// Three properties hold, and each of them is a decision rather than an
// implementation detail:
//
//   - **It is a projection, and it writes nothing.** No session is created, no
//     status moves, no document is marked seen. Opening a summary is not doing
//     the work, and a read that quietly acknowledged documents would destroy
//     the personal new/changed queue it sits next to.
//   - **It never guesses which session you meant.** The repository's "find the
//     active one" is a LIMIT 1 with no ORDER BY. With two sessions open, that
//     is a coin toss presented as a fact, so this returns both and says a
//     choice is required.
//   - **It never turns a failed read into an empty queue.** "No open tasks" is
//     a finding; "the tasks query failed" is an outage. Reporting the second as
//     the first is how an agent concludes the work is done and stops.
//
// It carries no memory, no instruction and no methodology. research_get owns
// those, and a second copy here would rebuild the always-loaded context this
// product spent a release removing.
type ResumeService struct {
	research    *ResearchService
	sessions    *storage.SessionRepository
	tasks       *storage.TaskRepository
	questions   *storage.QuestionRepository
	annotations *storage.AnnotationRepository
	entries     *storage.EntryRepository
	revisions   *storage.EntryRevisionRepository
	access      *Access
	log         *slog.Logger
}

func NewResumeService(
	research *ResearchService,
	sessions *storage.SessionRepository,
	tasks *storage.TaskRepository,
	questions *storage.QuestionRepository,
	annotations *storage.AnnotationRepository,
	entries *storage.EntryRepository,
	revisions *storage.EntryRevisionRepository,
	access *Access,
	log *slog.Logger,
) *ResumeService {
	return &ResumeService{
		research: research, sessions: sessions, tasks: tasks, questions: questions,
		annotations: annotations, entries: entries, revisions: revisions,
		access: access, log: log,
	}
}

// ResumeRequest is what the caller may steer: which session the summary is
// about, and how deep each queue goes.
type ResumeRequest struct {
	// SessionID is a UUID or an SS-code, resolved inside the chosen research.
	SessionID string
	// Limit is per group, clamped to [ResumeMinLimit, ResumeMaxLimit].
	Limit int
}

// Get builds the summary.
func (s *ResumeService) Get(ctx context.Context, researchIDOrCode string, req ResumeRequest) (*domain.ResearchResume, error) {
	// A share link is refused here rather than left to Access.Read, which would
	// allow it: a share resolves to viewer on this very research. The summary is
	// working process — what is unfinished, what a person disputed, what the
	// agent is meant to do next — and none of that is part of the read-only
	// result somebody published. Same rule as `instruction` and `memory`.
	if auth.ShareFromContext(ctx) != nil {
		return nil, ErrNotFound
	}

	research, err := s.research.Get(ctx, researchIDOrCode)
	if err != nil {
		return nil, err
	}
	limit := clampResumeLimit(req.Limit)

	out := &domain.ResearchResume{
		SchemaVersion: domain.ResumeSchemaVersion,
		// Not a sync cursor. Timestamps here have the precision the database
		// gives them and deletions leave no trace, so this says when the answer
		// was built and nothing more.
		GeneratedAt: time.Now().UTC(),
		Research: domain.ResumeResearch{
			ID: research.ID, Code: research.Code, Name: research.Name,
			Status: research.Status, Role: research.Role,
			CanWrite: s.access.Write(ctx, research.ID) == nil,
		},
	}

	selected, err := s.fillSessions(ctx, out, research, req.SessionID)
	if err != nil {
		return nil, err
	}
	if err := s.fillWork(ctx, out, research, limit); err != nil {
		return nil, err
	}
	if err := s.fillQuestions(ctx, out, research, selected, limit); err != nil {
		return nil, err
	}
	if err := s.fillAnnotations(ctx, out, research, limit); err != nil {
		return nil, err
	}
	if err := s.fillRecentEntries(ctx, out, research, limit); err != nil {
		return nil, err
	}

	out.NextActions = nextActions(out, research)
	fitResume(out)
	return out, nil
}

func clampResumeLimit(n int) int {
	switch {
	// Zero is "did not say", which is the default. A negative number is a
	// caller who did say, wrongly, and is clamped to the floor like any other
	// out-of-range value.
	case n == 0:
		return domain.ResumeDefaultLimit
	case n < domain.ResumeMinLimit:
		return domain.ResumeMinLimit
	case n > domain.ResumeMaxLimit:
		return domain.ResumeMaxLimit
	}
	return n
}

// fillSessions lists every session and decides which one the summary is about.
func (s *ResumeService) fillSessions(ctx context.Context, out *domain.ResearchResume, research *domain.Research, want string) (*domain.Session, error) {
	sessions, err := s.sessions.FindByResearch(ctx, research.ID)
	if err != nil {
		return nil, fmt.Errorf("list sessions: %w", err)
	}

	// Deterministic order, newest first. FindByResearch sorts on `created_at`
	// alone, and that column is stored at second precision — so two sessions
	// created in the same second (closing one and opening the next is exactly
	// that) came back in whatever order the engine liked, and "the most recent
	// session" was a different session on SQLite than on PostgreSQL.
	sort.SliceStable(sessions, func(i, j int) bool {
		if !sessions[i].CreatedAt.Equal(sessions[j].CreatedAt) {
			return sessions[i].CreatedAt.After(sessions[j].CreatedAt)
		}
		return sessions[i].ID > sessions[j].ID
	})

	// An empty list, never nil: `"items": null` is not a list, and a client that
	// reads `.length` off it breaks on the one case most likely to occur — a
	// research nobody has interviewed yet.
	out.Sessions.Items = []domain.ResumeSession{}

	var active []*domain.Session
	for _, sess := range sessions {
		if sess.Status == domain.SessionActive {
			active = append(active, sess)
		}
		// Bounded like every other string here. A session's focus has no length
		// limit anywhere in the product, and an unbounded one used to eat the
		// whole size budget before the work queues were even measured — leaving
		// a summary that reported four open tasks and listed none of them.
		if len(out.Sessions.Items) < domain.ResumeMaxSessions {
			out.Sessions.Items = append(out.Sessions.Items, domain.ResumeSession{
				ID: sess.ID, Code: sess.Code, Title: trimPreview(sess.Title),
				Focus: trimPreview(sess.Focus), Status: sess.Status, UpdatedAt: sess.UpdatedAt,
			})
		}
	}
	out.Sessions.ActiveCount = len(active)

	// An explicitly named session is checked against this research. Answering
	// about somebody else's session because the id parsed is the failure the
	// check exists for; it reads as not found, like every other cross-research
	// reach in this product.
	if want != "" {
		for _, sess := range sessions {
			if sess.ID == want || (sess.Code != "" && sess.Code == want) {
				out.Sessions.SelectedID = sess.ID
				return sess, nil
			}
		}
		return nil, ErrNotFound
	}

	switch len(active) {
	case 1:
		out.Sessions.SelectedID = active[0].ID
		return active[0], nil
	case 0:
		// No active session: show the most recently *created* one — which is the
		// order FindByResearch returns — with its real status, so the caller sees
		// what was closed rather than an empty screen.
		// Starting a new one is a write and stays a separate, deliberate act.
		if len(sessions) > 0 {
			out.Sessions.SelectedID = sessions[0].ID
			return sessions[0], nil
		}
		return nil, nil
	default:
		// Two sessions open is a question, not a tie to break.
		out.Sessions.SelectionRequired = true
		return nil, nil
	}
}

func (s *ResumeService) fillWork(ctx context.Context, out *domain.ResearchResume, research *domain.Research, limit int) error {
	counts, err := s.tasks.CountByStatus(ctx, research.ID)
	if err != nil {
		return fmt.Errorf("count tasks: %w", err)
	}
	more := domain.ResumeMore{Tool: "task_list", Href: "/research/" + research.Code + "/tasks"}

	for _, group := range []struct {
		statuses []domain.TaskStatus
		total    int
		target   *domain.ResumeGroup[domain.ResumeTask]
	}{
		{[]domain.TaskStatus{domain.TaskInProgress}, counts[domain.TaskInProgress], &out.Work.InProgress},
		{[]domain.TaskStatus{domain.TaskBlocked}, counts[domain.TaskBlocked], &out.Work.Blocked},
		{[]domain.TaskStatus{domain.TaskPending}, counts[domain.TaskPending], &out.Work.Pending},
	} {
		tasks, err := s.tasks.FindForResume(ctx, research.ID, group.statuses, limit)
		if err != nil {
			return fmt.Errorf("list tasks: %w", err)
		}
		items := make([]domain.ResumeTask, 0, len(tasks))
		for _, t := range tasks {
			items = append(items, domain.ResumeTask{
				ID: t.ID, Code: t.Code, Title: trimPreview(t.Title),
				Status: t.Status, Priority: t.Priority,
				Note: trimPreview(taskNote(t)), UpdatedAt: t.UpdatedAt,
			})
		}
		*group.target = newResumeGroup(items, group.total, more)
	}
	return nil
}

// taskNote is the explanation to show beside a task.
//
// The result first, because a finished or abandoned task says what happened
// there; the description otherwise, because that is where the guides tell an
// agent to write why a task is blocked. A description identical to the title is
// no explanation at all and is dropped — repeating the title under the title
// looks like a rendering bug.
func taskNote(t *domain.Task) string {
	if t.Result != "" {
		return t.Result
	}
	if t.Description != "" && t.Description != t.Title {
		return t.Description
	}
	return ""
}

func (s *ResumeService) fillQuestions(ctx context.Context, out *domain.ResearchResume, research *domain.Research, session *domain.Session, limit int) error {
	// Deliberately empty rather than research-wide when no session is selected:
	// questions belong to a session, and merging two open interviews into one
	// list would invent a queue nobody is working.
	empty := domain.ResumeMore{Tool: "session_get", Href: "/research/" + research.Code + "/sessions"}
	if session == nil {
		out.Questions.Open = newResumeGroup([]domain.ResumeQuestion{}, 0, empty)
		out.Questions.Deferred = newResumeGroup([]domain.ResumeQuestion{}, 0, empty)
		return nil
	}

	counts, err := s.questions.CountByStatus(ctx, session.ID)
	if err != nil {
		return fmt.Errorf("count questions: %w", err)
	}
	more := domain.ResumeMore{
		Tool: "session_get",
		Href: "/research/" + research.Code + "/session/" + firstNonEmpty(session.Code, session.ID),
	}

	// Answered and skipped are absent by design: proposing a question somebody
	// already answered is the specific failure a continuation summary exists to
	// prevent.
	for _, group := range []struct {
		statuses []domain.QuestionStatus
		total    int
		target   *domain.ResumeGroup[domain.ResumeQuestion]
	}{
		{[]domain.QuestionStatus{domain.QuestionPending, domain.QuestionInProgress},
			counts[domain.QuestionPending] + counts[domain.QuestionInProgress], &out.Questions.Open},
		{[]domain.QuestionStatus{domain.QuestionDeferred}, counts[domain.QuestionDeferred], &out.Questions.Deferred},
	} {
		questions, err := s.questions.FindForResume(ctx, session.ID, group.statuses, limit)
		if err != nil {
			return fmt.Errorf("list questions: %w", err)
		}
		items := make([]domain.ResumeQuestion, 0, len(questions))
		for _, q := range questions {
			items = append(items, domain.ResumeQuestion{
				ID: q.ID, Code: q.Code, SessionID: session.ID, SessionCode: session.Code,
				Text: trimPreview(q.Text), Area: q.Area, Priority: q.Priority, Status: q.Status,
			})
		}
		*group.target = newResumeGroup(items, group.total, more)
	}
	return nil
}

func (s *ResumeService) fillAnnotations(ctx context.Context, out *domain.ResearchResume, research *domain.Research, limit int) error {
	counts, err := s.annotations.CountByStatus(ctx, research.ID)
	if err != nil {
		return fmt.Errorf("count annotations: %w", err)
	}
	more := domain.ResumeMore{Tool: "annotation_list", Href: "/research/" + research.Code + "/annotations"}

	// The split is the point. `open` is work an agent can do; `answered` is an
	// answer waiting for the person who raised the objection to accept it, and
	// the agent cannot accept its own answer. One combined count would tell an
	// agent to keep working a queue that is waiting on somebody else.
	for _, group := range []struct {
		status domain.AnnotationStatus
		target *domain.ResumeGroup[domain.ResumeAnnotation]
	}{
		{domain.AnnotationOpen, &out.Annotations.ToWork},
		{domain.AnnotationAnswered, &out.Annotations.AwaitingHuman},
	} {
		anns, err := s.annotations.FindForResume(ctx, research.ID, group.status, limit)
		if err != nil {
			return fmt.Errorf("list annotations: %w", err)
		}
		items := make([]domain.ResumeAnnotation, 0, len(anns))
		for _, a := range anns {
			items = append(items, domain.ResumeAnnotation{
				ID: a.ID, Code: a.Code, EntryID: a.EntryID, EntryCode: a.EntryCode,
				EntryTitle: trimPreview(a.EntryTitle), Kind: a.Kind, Status: a.Status,
				Quote: trimPreview(a.Quote.Exact), UpdatedAt: a.UpdatedAt,
			})
		}
		*group.target = newResumeGroup(items, counts[group.status], more)
	}
	return nil
}

func (s *ResumeService) fillRecentEntries(ctx context.Context, out *domain.ResearchResume, research *domain.Research, limit int) error {
	// "Recently" is a stated window, not "the newest N of everything". Counting
	// every document in the research instead would label a research of two
	// hundred documents "Changed 200" on a day nobody touched it.
	since := time.Now().Add(-domain.ResumeRecentWindow)
	entries, err := s.entries.FindRecentlyUpdated(ctx, research.ID, since, limit)
	if err != nil {
		return fmt.Errorf("list recent entries: %w", err)
	}
	total, err := s.entries.CountUpdatedSince(ctx, research.ID, since)
	if err != nil {
		return fmt.Errorf("count entries: %w", err)
	}

	ids := make([]string, 0, len(entries))
	for _, e := range entries {
		ids = append(ids, e.ID)
	}
	// One query for every revision number on the page. A failure here costs the
	// numbers, not the list: the documents are still the right documents, and a
	// missing revision reads as "unknown", which it is.
	revisions, err := s.revisions.LatestByEntries(ctx, ids)
	if err != nil {
		s.log.Warn("resume: latest revisions unavailable", "research_id", research.ID, "error", err)
		revisions = map[string]storage.LatestRevision{}
	}

	items := make([]domain.ResumeEntry, 0, len(entries))
	for _, e := range entries {
		head := revisions[e.ID]
		items = append(items, domain.ResumeEntry{
			ID: e.ID, Code: e.Code, Title: trimPreview(e.Title), SectionID: e.SectionID,
			UpdatedAt: e.UpdatedAt, AuthorKind: head.AuthorKind, Revision: head.Revision,
		})
	}
	out.RecentEntries = newResumeGroup(items, total, domain.ResumeMore{
		Tool: "entry_list", Href: "/research/" + research.Code + "?section=__all__",
	})
	return nil
}

// newResumeGroup pairs what was taken with how much there is.
//
// `total` is a pointer in the DTO so an unknown count is distinguishable from
// zero; every caller here has a real COUNT, so it is always set.
func newResumeGroup[T any](items []T, total int, more domain.ResumeMore) domain.ResumeGroup[T] {
	if items == nil {
		items = []T{}
	}
	t := total
	return domain.ResumeGroup[T]{
		Items:    items,
		Returned: len(items),
		Total:    &t,
		HasMore:  total > len(items),
		More:     more,
	}
}

// nextActions proposes at most three continuations, in a fixed order.
//
// The order is a visible sorting policy, not a judgement about what matters in
// this research: finish what is already open, settle what a person disputed,
// then start the highest-priority thing waiting, then continue the interview.
// Each candidate carries the fact it was derived from, so a reader can disagree
// with the policy without having to trust it.
func nextActions(out *domain.ResearchResume, research *domain.Research) []domain.ResumeAction {
	actions := make([]domain.ResumeAction, 0, 3)

	// A mark the agent has answered is waiting on a person, and it is the only
	// row on this list the reader themselves can clear. It used to appear only
	// when nothing else did, so on any busy research the one thing asking for
	// the reader went unmentioned while three agent tasks filled the slots. It
	// is reserved a place instead — last, so it never displaces agent work, and
	// carrying `actor: human` so an agent reading this does not take it on.
	reserved := 0
	if len(out.Annotations.AwaitingHuman.Items) > 0 {
		reserved = 1
	}

	// Ambiguity comes first and is the person's to resolve: every queue below
	// is a session's, and this is the summary saying which session it could not
	// choose for you.
	if out.Sessions.SelectionRequired {
		actions = append(actions, domain.ResumeAction{
			Kind:       "choose_session",
			Target:     domain.ResumeTarget{Type: "research", ID: research.ID, Code: research.Code, Title: research.Name},
			ReasonCode: domain.ReasonSessionAmbiguous,
			Reason:     fmt.Sprintf("%d sessions are active, say which one you are continuing", out.Sessions.ActiveCount),
			Actor:      domain.ResumeActorHuman,
			Tool:       "research_resume",
			Href:       "/research/" + research.Code + "/sessions",
		})
	}

	add := func(a domain.ResumeAction) {
		if len(actions) < 3-reserved {
			actions = append(actions, a)
		}
	}

	if len(out.Work.InProgress.Items) > 0 {
		t := out.Work.InProgress.Items[0]
		add(domain.ResumeAction{
			Kind:       "continue_task",
			Target:     domain.ResumeTarget{Type: "task", ID: t.ID, Code: t.Code, Title: t.Title},
			ReasonCode: domain.ReasonTaskInProgress,
			Reason:     "already in progress",
			Actor:      domain.ResumeActorAgent,
			Tool:       "task_update",
			Href:       "/research/" + research.Code + "/tasks",
		})
	}
	if len(out.Annotations.ToWork.Items) > 0 {
		a := out.Annotations.ToWork.Items[0]
		add(domain.ResumeAction{
			Kind:       "answer_annotation",
			Target:     domain.ResumeTarget{Type: "annotation", ID: a.ID, Code: a.Code, Title: a.EntryTitle, EntryCode: a.EntryCode},
			ReasonCode: domain.ReasonAnnotationOpen,
			Reason:     fmt.Sprintf("an open %s mark on %s", a.Kind, firstNonEmpty(a.EntryCode, "a document")),
			Actor:      domain.ResumeActorAgent,
			Tool:       "annotation_list",
			Href:       "/research/" + research.Code + "/annotations",
		})
	}
	if len(out.Work.Pending.Items) > 0 {
		t := out.Work.Pending.Items[0]
		add(domain.ResumeAction{
			Kind:       "start_task",
			Target:     domain.ResumeTarget{Type: "task", ID: t.ID, Code: t.Code, Title: t.Title},
			ReasonCode: domain.ReasonTaskPending,
			Reason:     fmt.Sprintf("the highest-priority task waiting (%s)", t.Priority),
			Actor:      domain.ResumeActorAgent,
			Tool:       "task_update",
			Href:       "/research/" + research.Code + "/tasks",
		})
	}
	if len(out.Questions.Open.Items) > 0 {
		q := out.Questions.Open.Items[0]
		add(domain.ResumeAction{
			Kind:       "answer_question",
			Target:     domain.ResumeTarget{Type: "question", ID: q.ID, Code: q.Code, Title: q.Text, SessionCode: q.SessionCode},
			ReasonCode: domain.ReasonQuestionOpen,
			Reason:     fmt.Sprintf("the next unanswered question in %s", firstNonEmpty(q.SessionCode, "this session")),
			Actor:      domain.ResumeActorAgent,
			Tool:       "question_update",
			Href:       "/research/" + research.Code + "/session/" + firstNonEmpty(q.SessionCode, q.SessionID),
		})
	}
	// The reserved place, taken last so it sits under the agent's work.
	if reserved > 0 {
		a := out.Annotations.AwaitingHuman.Items[0]
		actions = append(actions, domain.ResumeAction{
			Kind:       "await_review",
			Target:     domain.ResumeTarget{Type: "annotation", ID: a.ID, Code: a.Code, Title: a.EntryTitle, EntryCode: a.EntryCode},
			ReasonCode: domain.ReasonAnnotationAnswer,
			Reason:     "answered, waiting for you to accept or reject it",
			Actor:      domain.ResumeActorHuman,
			Href:       "/research/" + research.Code + "/annotations",
		})
	}
	return actions
}

// fitResume brings the payload under the size cap without lying about what is
// in the research.
//
// The order of sacrifice is deliberate: previews first, then examples, and
// never the totals, the has_more flags or the links that lead to the full
// queues. An item that did not fit is not an item that does not exist, and the
// JSON stays valid at every step — a caller must never receive half a document.
func fitResume(out *domain.ResearchResume) {
	if resumeSize(out) <= domain.ResumeMaxBytes {
		return
	}
	out.Truncated = true
	out.Note = "Shortened to fit the response size limit; totals and links are unchanged."

	shortenResumePreviews(out, 60)
	if resumeSize(out) <= domain.ResumeMaxBytes {
		return
	}

	// The session list goes before the work does. It is context — which thread
	// you are on — and the queues are the answer to the question that was
	// asked; giving up the answer first is how the summary came back saying
	// "four tasks pending" and listing none of them.
	for i := range out.Sessions.Items {
		out.Sessions.Items[i].Focus = ""
	}
	if resumeSize(out) <= domain.ResumeMaxBytes {
		return
	}
	if len(out.Sessions.Items) > 1 {
		kept := out.Sessions.Items[:1]
		// Keep the one the summary is actually about, not merely the first.
		for _, item := range out.Sessions.Items {
			if item.ID == out.Sessions.SelectedID {
				kept = []domain.ResumeSession{item}
				break
			}
		}
		out.Sessions.Items = kept
		if resumeSize(out) <= domain.ResumeMaxBytes {
			return
		}
	}

	for _, keep := range []int{3, 1, 0} {
		capResumeGroups(out, keep)
		if resumeSize(out) <= domain.ResumeMaxBytes {
			return
		}
	}
}

func resumeSize(out *domain.ResearchResume) int {
	b, err := json.Marshal(out)
	if err != nil {
		// Unmarshalable is not "small": treat it as over the cap so the caller
		// gets the trimmed shape rather than a payload nobody measured.
		return domain.ResumeMaxBytes + 1
	}
	return len(b)
}

func shortenResumePreviews(out *domain.ResearchResume, runes int) {
	for i := range out.Work.InProgress.Items {
		out.Work.InProgress.Items[i].Title = trimRunes(out.Work.InProgress.Items[i].Title, runes)
	}
	for i := range out.Work.Blocked.Items {
		out.Work.Blocked.Items[i].Title = trimRunes(out.Work.Blocked.Items[i].Title, runes)
	}
	for i := range out.Work.Pending.Items {
		out.Work.Pending.Items[i].Title = trimRunes(out.Work.Pending.Items[i].Title, runes)
	}
	for i := range out.Questions.Open.Items {
		out.Questions.Open.Items[i].Text = trimRunes(out.Questions.Open.Items[i].Text, runes)
	}
	for i := range out.Questions.Deferred.Items {
		out.Questions.Deferred.Items[i].Text = trimRunes(out.Questions.Deferred.Items[i].Text, runes)
	}
	for i := range out.Annotations.ToWork.Items {
		out.Annotations.ToWork.Items[i].Quote = trimRunes(out.Annotations.ToWork.Items[i].Quote, runes)
		out.Annotations.ToWork.Items[i].EntryTitle = trimRunes(out.Annotations.ToWork.Items[i].EntryTitle, runes)
	}
	for i := range out.Annotations.AwaitingHuman.Items {
		out.Annotations.AwaitingHuman.Items[i].Quote = trimRunes(out.Annotations.AwaitingHuman.Items[i].Quote, runes)
		out.Annotations.AwaitingHuman.Items[i].EntryTitle = trimRunes(out.Annotations.AwaitingHuman.Items[i].EntryTitle, runes)
	}
	for i := range out.RecentEntries.Items {
		out.RecentEntries.Items[i].Title = trimRunes(out.RecentEntries.Items[i].Title, runes)
	}
}

func capResumeGroups(out *domain.ResearchResume, keep int) {
	capResumeGroup(&out.Work.InProgress, keep)
	capResumeGroup(&out.Work.Blocked, keep)
	capResumeGroup(&out.Work.Pending, keep)
	capResumeGroup(&out.Questions.Open, keep)
	capResumeGroup(&out.Questions.Deferred, keep)
	capResumeGroup(&out.Annotations.ToWork, keep)
	capResumeGroup(&out.Annotations.AwaitingHuman, keep)
	capResumeGroup(&out.RecentEntries, keep)
}

// capResumeGroup drops items and re-derives `returned` and `has_more` from the
// total, so a shortened list still reports how much there really is.
func capResumeGroup[T any](g *domain.ResumeGroup[T], keep int) {
	if len(g.Items) > keep {
		g.Items = g.Items[:keep]
	}
	g.Returned = len(g.Items)
	if g.Total != nil {
		g.HasMore = *g.Total > g.Returned
	} else {
		g.HasMore = true
	}
}

func trimPreview(s string) string { return trimRunes(s, domain.ResumePreviewRunes) }

// trimRunes cuts by rune, never by byte: half a Cyrillic character is not a
// shorter string, it is an invalid one.
func trimRunes(s string, max int) string {
	if max <= 0 {
		return ""
	}
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max]) + "…"
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
