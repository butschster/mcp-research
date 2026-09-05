package service

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/dovod-app/app/internal/auth"
	"github.com/dovod-app/app/internal/domain"
	"github.com/dovod-app/app/internal/storage"
	"github.com/uptrace/bun"
)

// resumeFixture builds the whole cast the summary reads from, wired the way
// main.go wires it.
type resumeFixture struct {
	db      *bun.DB
	resume  *ResumeService
	tasks   *TaskService
	entries *EntryService
	session *SessionService
	anns    *AnnotationService
	notify  *mockNotifier
}

func newResumeFixture(t *testing.T) *resumeFixture {
	t.Helper()
	db := setupTestDB(t)
	log := slog.Default()
	notifier := &mockNotifier{}
	access := testAccess(db)

	entryRepo := storage.NewEntryRepository(db)
	researchRepo := storage.NewResearchRepository(db)
	sessionRepo := storage.NewSessionRepository(db)
	taskRepo := storage.NewTaskRepository(db)
	questionRepo := storage.NewQuestionRepository(db)
	annRepo := storage.NewAnnotationRepository(db)
	revRepo := storage.NewEntryRevisionRepository(db)

	entrySvc := NewEntryService(entryRepo, storage.NewSectionRepository(db), researchRepo, access,
		sessionRepo, storage.NewBlockRepository(db), revRepo,
		storage.NewCrossRefRepository(db), storage.NewExternalLinkRepository(db), notifier, log)
	researchSvc := NewResearchService(researchRepo, storage.NewSectionRepository(db),
		storage.NewTeamRepository(db), access, notifier, log)
	sessionSvc := NewSessionService(db, sessionRepo, questionRepo, researchRepo, access, entrySvc, notifier, log)
	taskSvc := NewTaskService(taskRepo, researchRepo, access, entrySvc, notifier, log)
	annSvc := NewAnnotationService(annRepo, entryRepo, revRepo, access, entrySvc, entrySvc, notifier, log)

	return &resumeFixture{
		db:     db,
		resume: NewResumeService(researchSvc, sessionRepo, taskRepo, questionRepo, annRepo, entryRepo, revRepo, access, log),
		tasks:  taskSvc, entries: entrySvc, session: sessionSvc, anns: annSvc, notify: notifier,
	}
}

func TestResume_EmptyResearch(t *testing.T) {
	f := newResumeFixture(t)
	ctx := context.Background()
	r := createTestResearch(t, f.db)

	out, err := f.resume.Get(ctx, r.ID, ResumeRequest{})
	if err != nil {
		t.Fatalf("resume: %v", err)
	}

	if out.SchemaVersion != domain.ResumeSchemaVersion {
		t.Errorf("schema_version = %d", out.SchemaVersion)
	}
	if out.Research.Code != r.Code || out.Research.ID != r.ID {
		t.Errorf("research identity = %+v", out.Research)
	}
	// Empty groups must still carry a total and a way in. A group with a nil
	// total would be indistinguishable from one whose count failed.
	for name, g := range map[string]struct {
		total   *int
		hasMore bool
		items   int
	}{
		"in_progress": {out.Work.InProgress.Total, out.Work.InProgress.HasMore, len(out.Work.InProgress.Items)},
		"pending":     {out.Work.Pending.Total, out.Work.Pending.HasMore, len(out.Work.Pending.Items)},
		"to_work":     {out.Annotations.ToWork.Total, out.Annotations.ToWork.HasMore, len(out.Annotations.ToWork.Items)},
		"recent":      {out.RecentEntries.Total, out.RecentEntries.HasMore, len(out.RecentEntries.Items)},
	} {
		if g.total == nil || *g.total != 0 || g.hasMore || g.items != 0 {
			t.Errorf("%s: empty group should report zero and no more, got total=%v has_more=%v items=%d", name, g.total, g.hasMore, g.items)
		}
	}
	if len(out.NextActions) != 0 {
		t.Errorf("nothing outstanding should propose nothing, got %+v", out.NextActions)
	}
	if out.Sessions.SelectionRequired || out.Sessions.SelectedID != "" {
		t.Errorf("no sessions: %+v", out.Sessions)
	}
	// An empty list, not null: a client reading `.length` must not break on the
	// research nobody has interviewed yet.
	if out.Sessions.Items == nil {
		t.Error("sessions.items is nil; it must serialise as an empty array")
	}
}

func TestResume_GroupsTotalsAndOrder(t *testing.T) {
	f := newResumeFixture(t)
	ctx := context.Background()
	r, section := createTestResearchWithSection(t, f.db)

	// Seven pending tasks so a limit of 3 leaves four behind.
	for i := 0; i < 7; i++ {
		if _, err := f.tasks.Create(ctx, CreateTaskRequest{
			ResearchID: r.ID, Title: "Pending task", Priority: domain.PriorityLow,
		}); err != nil {
			t.Fatalf("create task: %v", err)
		}
	}
	// One high-priority pending task must lead the group regardless of age.
	top, err := f.tasks.Create(ctx, CreateTaskRequest{ResearchID: r.ID, Title: "Urgent", Priority: domain.PriorityHigh})
	if err != nil {
		t.Fatalf("create task: %v", err)
	}
	running, err := f.tasks.Create(ctx, CreateTaskRequest{ResearchID: r.ID, Title: "Running"})
	if err != nil {
		t.Fatalf("create task: %v", err)
	}
	inProgress := domain.TaskInProgress
	if _, err := f.tasks.Update(ctx, running.ID, UpdateTaskRequest{Status: &inProgress}); err != nil {
		t.Fatalf("update task: %v", err)
	}
	blocked, err := f.tasks.Create(ctx, CreateTaskRequest{ResearchID: r.ID, Title: "Waiting on legal"})
	if err != nil {
		t.Fatalf("create task: %v", err)
	}
	blockedStatus := domain.TaskBlocked
	reason := "legal has not answered"
	if _, err := f.tasks.Update(ctx, blocked.ID, UpdateTaskRequest{Status: &blockedStatus, Result: &reason}); err != nil {
		t.Fatalf("update task: %v", err)
	}
	if _, err := f.entries.Create(ctx, CreateEntryRequest{
		ResearchID: r.ID, SectionID: section.ID, Title: "A finding", Content: "body",
	}); err != nil {
		t.Fatalf("create entry: %v", err)
	}

	out, err := f.resume.Get(ctx, r.Code, ResumeRequest{Limit: 3})
	if err != nil {
		t.Fatalf("resume: %v", err)
	}

	if got := out.Work.Pending.Returned; got != 3 {
		t.Errorf("pending returned = %d, want 3", got)
	}
	if out.Work.Pending.Total == nil || *out.Work.Pending.Total != 8 {
		t.Errorf("pending total = %v, want 8", out.Work.Pending.Total)
	}
	if !out.Work.Pending.HasMore {
		t.Error("pending has_more should be true: five were left behind")
	}
	if out.Work.Pending.More.Tool == "" || out.Work.Pending.More.Href == "" {
		t.Errorf("a shortened group must say where the rest is: %+v", out.Work.Pending.More)
	}
	if out.Work.Pending.Items[0].Code != top.Code {
		t.Errorf("pending[0] = %s, want the high-priority %s", out.Work.Pending.Items[0].Code, top.Code)
	}
	if out.Work.Blocked.Returned != 1 || out.Work.Blocked.Items[0].Note != reason {
		t.Errorf("blocked group should carry the recorded reason: %+v", out.Work.Blocked.Items)
	}
	if out.Work.InProgress.Returned != 1 || out.Work.InProgress.Items[0].Code != running.Code {
		t.Errorf("in_progress = %+v", out.Work.InProgress.Items)
	}
	if out.RecentEntries.Returned != 1 || out.RecentEntries.Items[0].Revision != 1 {
		t.Errorf("recent entries = %+v", out.RecentEntries.Items)
	}

	// The first suggestion is the work already open, and it names the evidence.
	if len(out.NextActions) == 0 {
		t.Fatal("expected next actions")
	}
	first := out.NextActions[0]
	if first.ReasonCode != domain.ReasonTaskInProgress || first.Target.Code != running.Code {
		t.Errorf("first action = %+v, want the in-progress task", first)
	}
	if first.Actor != domain.ResumeActorAgent || first.Reason == "" {
		t.Errorf("an action must say who acts and why: %+v", first)
	}
}

func TestResume_SeveralActiveSessionsAreNotGuessed(t *testing.T) {
	f := newResumeFixture(t)
	ctx := context.Background()
	r := createTestResearch(t, f.db)

	first, _, err := f.session.Create(ctx, CreateSessionRequest{ResearchID: r.ID, Title: "Exploration"})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	if _, _, err := f.session.Create(ctx, CreateSessionRequest{ResearchID: r.ID, Title: "Deep dive"}); err != nil {
		t.Fatalf("create session: %v", err)
	}

	out, err := f.resume.Get(ctx, r.ID, ResumeRequest{})
	if err != nil {
		t.Fatalf("resume: %v", err)
	}
	if !out.Sessions.SelectionRequired || out.Sessions.SelectedID != "" {
		t.Fatalf("two active sessions must not be resolved silently: %+v", out.Sessions)
	}
	if out.Sessions.ActiveCount != 2 {
		t.Errorf("active_count = %d, want 2", out.Sessions.ActiveCount)
	}
	if len(out.NextActions) == 0 || out.NextActions[0].ReasonCode != domain.ReasonSessionAmbiguous {
		t.Fatalf("the ambiguity is the first thing to resolve: %+v", out.NextActions)
	}
	if out.NextActions[0].Actor != domain.ResumeActorHuman {
		t.Errorf("choosing a session is the person's call, got actor %q", out.NextActions[0].Actor)
	}
	// Naming one resolves it.
	picked, err := f.resume.Get(ctx, r.ID, ResumeRequest{SessionID: first.Code})
	if err != nil {
		t.Fatalf("resume with session: %v", err)
	}
	if picked.Sessions.SelectedID != first.ID || picked.Sessions.SelectionRequired {
		t.Errorf("named session should be selected: %+v", picked.Sessions)
	}
}

func TestResume_SessionFromAnotherResearchIsNotFound(t *testing.T) {
	f := newResumeFixture(t)
	ctx := context.Background()
	mine := createTestResearch(t, f.db)
	theirs := createTestResearch(t, f.db)

	other, _, err := f.session.Create(ctx, CreateSessionRequest{ResearchID: theirs.ID, Title: "Not yours"})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	if _, err := f.resume.Get(ctx, mine.ID, ResumeRequest{SessionID: other.ID}); !errors.Is(err, ErrNotFound) {
		t.Fatalf("cross-research session = %v, want ErrNotFound", err)
	}
}

func TestResume_QuestionsExcludeAnswered(t *testing.T) {
	f := newResumeFixture(t)
	ctx := context.Background()
	r := createTestResearch(t, f.db)

	session, questions, err := f.session.Create(ctx, CreateSessionRequest{
		ResearchID: r.ID, Title: "Interview",
		Questions: []CreateQuestionRequest{
			{Text: "First question", Priority: domain.PriorityHigh},
			{Text: "Second question", Priority: domain.PriorityMedium},
		},
	})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	answered := domain.QuestionAnswered
	answer := "already settled"
	if _, err := f.session.UpdateQuestion(ctx, questions[0].ID, &answered, &answer); err != nil {
		t.Fatalf("answer question: %v", err)
	}

	out, err := f.resume.Get(ctx, r.ID, ResumeRequest{})
	if err != nil {
		t.Fatalf("resume: %v", err)
	}
	if out.Questions.Open.Returned != 1 || out.Questions.Open.Items[0].Code != questions[1].Code {
		t.Fatalf("an answered question must not be proposed again: %+v", out.Questions.Open.Items)
	}
	if out.Questions.Open.Items[0].SessionCode != session.Code {
		t.Errorf("a question must carry the session it is read through, got %+v", out.Questions.Open.Items[0])
	}
}

func TestResume_AnnotationsSplitAgentWorkFromHumanReview(t *testing.T) {
	f := newResumeFixture(t)
	ctx := context.Background()
	r, section := createTestResearchWithSection(t, f.db)

	entry, err := f.entries.Create(ctx, CreateEntryRequest{
		ResearchID: r.ID, SectionID: section.ID, Title: "Claims", Content: "The market doubled in a year.",
	})
	if err != nil {
		t.Fatalf("create entry: %v", err)
	}
	open, err := f.anns.Create(ctx, CreateAnnotationRequest{
		EntryID: entry.ID, Kind: domain.AnnotationVerify,
		Quote: domain.Quote{Exact: "The market doubled"}, Body: "source?",
	})
	if err != nil {
		t.Fatalf("create annotation: %v", err)
	}
	settled, err := f.anns.Create(ctx, CreateAnnotationRequest{
		EntryID: entry.ID, Kind: domain.AnnotationDig,
		Quote: domain.Quote{Exact: "in a year"}, Body: "expand",
	})
	if err != nil {
		t.Fatalf("create annotation: %v", err)
	}
	if _, err := f.anns.Answer(ctx, settled.ID, AnswerAnnotationRequest{Resolution: "written up in E2"}); err != nil {
		t.Fatalf("answer annotation: %v", err)
	}

	out, err := f.resume.Get(ctx, r.ID, ResumeRequest{})
	if err != nil {
		t.Fatalf("resume: %v", err)
	}
	if out.Annotations.ToWork.Returned != 1 || out.Annotations.ToWork.Items[0].Code != open.Code {
		t.Fatalf("to_work = %+v, want only the open mark", out.Annotations.ToWork.Items)
	}
	if out.Annotations.AwaitingHuman.Returned != 1 || out.Annotations.AwaitingHuman.Items[0].Code != settled.Code {
		t.Fatalf("awaiting_human = %+v, want the answered mark", out.Annotations.AwaitingHuman.Items)
	}
	if out.Annotations.ToWork.Items[0].EntryCode != entry.Code || out.Annotations.ToWork.Items[0].Quote == "" {
		t.Errorf("a mark must name its document and quote: %+v", out.Annotations.ToWork.Items[0])
	}
	// The agent's own answer is never proposed back to the agent for acceptance,
	// and it is never buried either: it is the one row the reader can clear
	// themselves, so it holds a place even when agent work fills the list.
	var awaiting *domain.ResumeAction
	for i, a := range out.NextActions {
		if a.ReasonCode == domain.ReasonAnnotationAnswer {
			awaiting = &out.NextActions[i]
			if a.Actor != domain.ResumeActorHuman {
				t.Errorf("an answered mark waits on a person, got actor %q", a.Actor)
			}
		}
	}
	if awaiting == nil {
		t.Error("the mark waiting on a person was left out of the next actions entirely")
	} else if out.NextActions[len(out.NextActions)-1].ReasonCode != domain.ReasonAnnotationAnswer {
		t.Error("waiting-on-a-person should sit last, under the work the agent can do")
	}
}

func TestResume_HumanEditIsVisibleOnRecentEntries(t *testing.T) {
	f := newResumeFixture(t)
	ctx := context.Background()
	r, section := createTestResearchWithSection(t, f.db)

	entry, err := f.entries.Create(ctx, CreateEntryRequest{
		ResearchID: r.ID, SectionID: section.ID, Title: "Draft", Content: "written by an agent",
	})
	if err != nil {
		t.Fatalf("create entry: %v", err)
	}
	human := WithAuthor(ctx, domain.AuthorHuman)
	if _, err := f.entries.Update(human, entry.ID, UpdateEntryRequest{Content: ptr("corrected by a person")}); err != nil {
		t.Fatalf("update entry: %v", err)
	}

	out, err := f.resume.Get(ctx, r.ID, ResumeRequest{})
	if err != nil {
		t.Fatalf("resume: %v", err)
	}
	if out.RecentEntries.Returned != 1 {
		t.Fatalf("recent entries = %+v", out.RecentEntries.Items)
	}
	got := out.RecentEntries.Items[0]
	if got.AuthorKind != domain.AuthorHuman {
		t.Errorf("author_kind = %q, want human — a correction must not read as stale agent work", got.AuthorKind)
	}
	if got.Revision != 2 {
		t.Errorf("revision = %d, want 2", got.Revision)
	}
}

func TestResume_IsReadOnly(t *testing.T) {
	f := newResumeFixture(t)
	ctx := context.Background()
	r := createTestResearch(t, f.db)
	if _, _, err := f.session.Create(ctx, CreateSessionRequest{ResearchID: r.ID, Title: "Interview"}); err != nil {
		t.Fatalf("create session: %v", err)
	}
	task, err := f.tasks.Create(ctx, CreateTaskRequest{ResearchID: r.ID, Title: "Something"})
	if err != nil {
		t.Fatalf("create task: %v", err)
	}

	f.notify.reset()
	if _, err := f.resume.Get(ctx, r.ID, ResumeRequest{}); err != nil {
		t.Fatalf("resume: %v", err)
	}
	if len(f.notify.events) != 0 {
		t.Errorf("a read emitted write events: %+v", f.notify.events)
	}

	after, err := f.tasks.Get(ctx, task.ID)
	if err != nil {
		t.Fatalf("get task: %v", err)
	}
	if after.Status != domain.TaskPending || !after.UpdatedAt.Equal(task.UpdatedAt) {
		t.Errorf("reading the summary changed the task: %+v", after)
	}
	sessions, err := f.session.ListByResearch(ctx, r.ID)
	if err != nil {
		t.Fatalf("list sessions: %v", err)
	}
	if len(sessions) != 1 {
		t.Errorf("reading the summary created a session: %d sessions", len(sessions))
	}
}

func TestResume_ShareLinkIsRefused(t *testing.T) {
	f := newResumeFixture(t)
	ctx := context.Background()
	r := createTestResearch(t, f.db)

	// The visitor holds a valid capability over this very research, so Access
	// alone would let them through as a viewer. The summary is working process
	// and refuses anyway — the same rule that strips instruction and memory.
	shared := auth.WithShare(ctx, &auth.Share{ResearchID: r.ID})
	if _, err := f.resume.Get(shared, r.ID, ResumeRequest{}); !errors.Is(err, ErrNotFound) {
		t.Fatalf("share resume = %v, want ErrNotFound", err)
	}
}

func TestResume_OtherTeamsResearchIsNotFound(t *testing.T) {
	f := newResumeFixture(t)
	ctx := context.Background()

	owner := createTestUser(t, f.db, "owner@example.com", "Owner")
	stranger := createTestUser(t, f.db, "stranger@example.com", "Stranger")

	researchSvc := NewResearchService(storage.NewResearchRepository(f.db), storage.NewSectionRepository(f.db),
		storage.NewTeamRepository(f.db), testAccess(f.db), &mockNotifier{}, slog.Default())
	r, _, err := researchSvc.Create(auth.WithUser(ctx, owner), CreateResearchRequest{Name: "Private", Goal: "x"})
	if err != nil {
		t.Fatalf("create research: %v", err)
	}

	if _, err := f.resume.Get(auth.WithUser(ctx, stranger), r.ID, ResumeRequest{}); !errors.Is(err, ErrNotFound) {
		t.Fatalf("stranger resume = %v, want ErrNotFound", err)
	}
	if _, err := f.resume.Get(auth.WithUser(ctx, owner), r.ID, ResumeRequest{}); err != nil {
		t.Fatalf("owner resume: %v", err)
	}
}

func TestResume_RealisticPayloadStaysUnderTheCap(t *testing.T) {
	f := newResumeFixture(t)
	ctx := context.Background()
	r := createTestResearch(t, f.db)

	// The worst case a caller can ask for: the maximum limit, every title long
	// enough to be cut, in a script where one character is two bytes.
	long := strings.Repeat("длинный заголовок задачи ", 400)
	for i := 0; i < domain.ResumeMaxLimit*2; i++ {
		if _, err := f.tasks.Create(ctx, CreateTaskRequest{ResearchID: r.ID, Title: long}); err != nil {
			t.Fatalf("create task: %v", err)
		}
	}

	out, err := f.resume.Get(ctx, r.ID, ResumeRequest{Limit: domain.ResumeMaxLimit})
	if err != nil {
		t.Fatalf("resume: %v", err)
	}
	if size := resumeSize(out); size > domain.ResumeMaxBytes {
		t.Errorf("payload is %d bytes, cap is %d", size, domain.ResumeMaxBytes)
	}
	if out.Work.Pending.Total == nil || *out.Work.Pending.Total != domain.ResumeMaxLimit*2 {
		t.Errorf("pending total = %v, want %d", out.Work.Pending.Total, domain.ResumeMaxLimit*2)
	}
	for _, item := range out.Work.Pending.Items {
		if runes := []rune(item.Title); len(runes) > domain.ResumePreviewRunes+1 {
			t.Errorf("title kept %d runes, cap is %d", len(runes), domain.ResumePreviewRunes)
		}
		if !utf8ValidString(item.Title) {
			t.Errorf("truncated title is not valid UTF-8: %q", item.Title)
		}
	}
}

// The size policy itself, exercised on a payload built to breach the cap.
//
// It is tested here rather than through the service because the service already
// bounds every preview as it collects, so a real research cannot reach the cap
// without thousands of items — and the property worth pinning is what happens
// when something does: totals and links survive, items are what is given up.
func TestResume_FitDropsItemsButNeverTotals(t *testing.T) {
	long := strings.Repeat("очень длинный заголовок ", 200)
	out := &domain.ResearchResume{SchemaVersion: domain.ResumeSchemaVersion}

	total := 900
	items := make([]domain.ResumeTask, 0, 60)
	for i := 0; i < 60; i++ {
		items = append(items, domain.ResumeTask{
			ID: "id", Code: "T1", Title: long, Status: domain.TaskPending, Priority: domain.PriorityLow,
		})
	}
	more := domain.ResumeMore{Tool: "task_list", Href: "/research/R1/tasks"}
	out.Work.Pending = newResumeGroup(items, total, more)
	out.NextActions = []domain.ResumeAction{{
		Kind: "start_task", ReasonCode: domain.ReasonTaskPending,
		Reason: "T1 is the highest-priority task waiting (low).", Actor: domain.ResumeActorAgent,
	}}

	if resumeSize(out) <= domain.ResumeMaxBytes {
		t.Fatal("fixture is not over the cap; the test proves nothing")
	}
	fitResume(out)

	if size := resumeSize(out); size > domain.ResumeMaxBytes {
		t.Errorf("payload is %d bytes after fitting, cap is %d", size, domain.ResumeMaxBytes)
	}
	if !out.Truncated || out.Note == "" {
		t.Errorf("a shortened payload must say so: truncated=%v note=%q", out.Truncated, out.Note)
	}
	if out.Work.Pending.Total == nil || *out.Work.Pending.Total != total {
		t.Errorf("truncation lost the total: %v", out.Work.Pending.Total)
	}
	if !out.Work.Pending.HasMore {
		t.Error("a truncated group must report that more exists")
	}
	if out.Work.Pending.Returned != len(out.Work.Pending.Items) {
		t.Errorf("returned=%d does not match %d items kept", out.Work.Pending.Returned, len(out.Work.Pending.Items))
	}
	if out.Work.Pending.More.Tool == "" || out.Work.Pending.More.Href == "" {
		t.Error("the way to the rest of the queue must survive truncation")
	}
	if len(out.NextActions) == 0 {
		t.Error("truncation must not drop the next actions: they are the point of the summary")
	}
	for _, item := range out.Work.Pending.Items {
		if !utf8ValidString(item.Title) {
			t.Errorf("truncated title is not valid UTF-8: %q", item.Title)
		}
	}
}

func utf8ValidString(s string) bool {
	for _, r := range s {
		if r == '\uFFFD' {
			return false
		}
	}
	return true
}

// A session with an enormous focus must not cost the caller their work queue.
//
// `focus` has no length limit anywhere in the product, and the size cap used to
// empty all eight groups before it touched the session list — so a summary came
// back reporting four pending tasks and listing none of them, which is the one
// thing the issue says a queue must never do.
func TestResume_LongSessionTextDoesNotEmptyTheQueues(t *testing.T) {
	f := newResumeFixture(t)
	ctx := context.Background()
	r := createTestResearch(t, f.db)

	huge := strings.Repeat("контекст сессии ", 900)
	for i := 0; i < 3; i++ {
		if _, _, err := f.session.Create(ctx, CreateSessionRequest{
			ResearchID: r.ID, Title: "Сессия", Focus: huge,
		}); err != nil {
			t.Fatalf("create session: %v", err)
		}
	}
	for i := 0; i < 4; i++ {
		if _, err := f.tasks.Create(ctx, CreateTaskRequest{ResearchID: r.ID, Title: "Открытая задача"}); err != nil {
			t.Fatalf("create task: %v", err)
		}
	}

	out, err := f.resume.Get(ctx, r.ID, ResumeRequest{})
	if err != nil {
		t.Fatalf("resume: %v", err)
	}
	if size := resumeSize(out); size > domain.ResumeMaxBytes {
		t.Errorf("payload is %d bytes, cap is %d", size, domain.ResumeMaxBytes)
	}
	if out.Work.Pending.Returned == 0 {
		t.Fatalf("the work queue was emptied to make room for session text: %+v", out.Work.Pending)
	}
	if out.Work.Pending.Total == nil || *out.Work.Pending.Total != 4 {
		t.Errorf("pending total = %v, want 4", out.Work.Pending.Total)
	}
	// Whatever the summary reports as returned is what it actually carries; a
	// chip saying "Todo 4" over a panel saying "the board is clear" is the
	// failure this asserts against.
	if out.Work.Pending.Returned != len(out.Work.Pending.Items) {
		t.Errorf("returned=%d but %d items", out.Work.Pending.Returned, len(out.Work.Pending.Items))
	}
	for _, item := range out.Sessions.Items {
		if runes := []rune(item.Focus); len(runes) > domain.ResumePreviewRunes+1 {
			t.Errorf("session focus kept %d runes, cap is %d", len(runes), domain.ResumePreviewRunes)
		}
	}
}

// "Changed" counts what changed, not what exists. A research full of documents
// nobody has touched reports zero, which is what lets the UI say "nothing
// waiting" instead of showing a permanent counter.
func TestResume_RecentEntriesCountOnlyRecentOnes(t *testing.T) {
	f := newResumeFixture(t)
	ctx := context.Background()
	r, section := createTestResearchWithSection(t, f.db)

	entry, err := f.entries.Create(ctx, CreateEntryRequest{
		ResearchID: r.ID, SectionID: section.ID, Title: "Старый документ", Content: "тело",
	})
	if err != nil {
		t.Fatalf("create entry: %v", err)
	}
	// Push it outside the window the summary calls "recently".
	old := time.Now().Add(-domain.ResumeRecentWindow - time.Hour)
	if _, err := f.db.NewUpdate().Table("entries").
		Set("updated_at = ?", old).Where("id = ?", entry.ID).Exec(ctx); err != nil {
		t.Fatalf("age the entry: %v", err)
	}

	out, err := f.resume.Get(ctx, r.ID, ResumeRequest{})
	if err != nil {
		t.Fatalf("resume: %v", err)
	}
	if out.RecentEntries.Total == nil || *out.RecentEntries.Total != 0 {
		t.Errorf("recent total = %v, want 0 — the document exists but nothing changed", out.RecentEntries.Total)
	}
	if out.RecentEntries.Returned != 0 || out.RecentEntries.HasMore {
		t.Errorf("recent entries = %+v, want an empty list with no more", out.RecentEntries)
	}
}
