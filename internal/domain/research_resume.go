package domain

import "time"

// ResumeSchemaVersion is the DTO's version. A client that keeps a summary on
// screen across a reconnect compares it before merging.
const ResumeSchemaVersion = 1

// ResumeRecentWindow is what "recently" means for the changed-documents list.
//
// The window has to be stated somewhere, and stating it here is what lets the
// count beside that list be a count of documents that changed rather than of
// documents that exist. Two weeks is the span a person means by "while I was
// away"; a research untouched for longer shows an empty list, which is true.
const ResumeRecentWindow = 14 * 24 * time.Hour

const (
	// ResumeDefaultLimit is how many items a group carries when the caller says
	// nothing. Five is what fits a person's glance and an agent's budget; the
	// totals beside it say how much was left behind.
	ResumeDefaultLimit = 5
	ResumeMinLimit     = 1
	ResumeMaxLimit     = 15
	// ResumeMaxBytes caps the serialized payload. A summary that grows with the
	// research is the thing this feature exists to avoid.
	ResumeMaxBytes = 24 * 1024
	// ResumePreviewRunes bounds one quoted or titled string, counted in runes:
	// half a Cyrillic character is not a shorter string, it is a broken one.
	ResumePreviewRunes = 160
	// ResumeMaxSessions bounds the session list. It is generous because the
	// picker is built from it, and a research with more open threads than this
	// has a problem the summary cannot fix.
	ResumeMaxSessions = 12
)

// ResumeGroup is one queue in the summary: what was taken, how much there is,
// and whether the caller has seen all of it.
//
// `Total` is nullable on purpose. A group whose count could not be established
// says so rather than reporting zero — "no tasks" and "we did not manage to
// count the tasks" are different sentences, and the second one must never be
// rendered as the first.
type ResumeGroup[T any] struct {
	Items    []T  `json:"items"`
	Returned int  `json:"returned"`
	Total    *int `json:"total"`
	HasMore  bool `json:"has_more"`
	// More names where the full queue lives: a tool for an agent, a path for a
	// person. A truncated list that does not say how to finish reading it is a
	// list that claims to be complete.
	More ResumeMore `json:"more"`
}

// ResumeMore is how to open the whole of a group that was cut short.
type ResumeMore struct {
	Tool string `json:"tool,omitempty"`
	Href string `json:"href,omitempty"`
}

// ResumeResearch is the identity half of the summary. Constraints, memory and
// methodology are deliberately absent: research_get owns those, and copying
// them here would rebuild the always-loaded context this product removed.
type ResumeResearch struct {
	ID       string         `json:"id"`
	Code     string         `json:"code"`
	Name     string         `json:"name"`
	Status   ResearchStatus `json:"status"`
	Role     TeamRole       `json:"role,omitempty"`
	CanWrite bool           `json:"can_write"`
}

type ResumeSession struct {
	ID        string        `json:"id"`
	Code      string        `json:"code"`
	Title     string        `json:"title"`
	Focus     string        `json:"focus,omitempty"`
	Status    SessionStatus `json:"status"`
	UpdatedAt time.Time     `json:"updated_at"`
}

// ResumeSessions carries every session a continuation could mean.
//
// SelectionRequired is the whole point: two active sessions is a question for
// the caller, not a coin toss. The repository's "find an active one" is a
// LIMIT 1 with no ORDER BY, so picking silently would mean picking arbitrarily.
type ResumeSessions struct {
	Items             []ResumeSession `json:"items"`
	SelectedID        string          `json:"selected_id,omitempty"`
	SelectionRequired bool            `json:"selection_required"`
	// ActiveCount counts sessions still open, which is what makes the ambiguity
	// legible without the reader counting the list themselves.
	ActiveCount int `json:"active_count"`
}

type ResumeTask struct {
	ID       string     `json:"id"`
	Code     string     `json:"code"`
	Title    string     `json:"title"`
	Status   TaskStatus `json:"status"`
	Priority Priority   `json:"priority"`
	// Note is whatever explanation the task carries: its result once there is
	// one, otherwise its description. Both are where a reason ends up in
	// practice — the guides tell an agent to say why a task is blocked in the
	// description, and `result` is written when the work finishes.
	//
	// It is deliberately not called `blocked_reason`: nothing forces either
	// field to be filled, and that name would read as a promise that a blocked
	// task always explains itself. Empty means nobody said why, which is itself
	// worth seeing.
	Note      string    `json:"note,omitempty"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ResumeQuestion struct {
	ID          string         `json:"id"`
	Code        string         `json:"code"`
	SessionID   string         `json:"session_id"`
	SessionCode string         `json:"session_code,omitempty"`
	Text        string         `json:"text"`
	Area        string         `json:"area,omitempty"`
	Priority    Priority       `json:"priority"`
	Status      QuestionStatus `json:"status"`
}

type ResumeAnnotation struct {
	ID         string           `json:"id"`
	Code       string           `json:"code"`
	EntryID    string           `json:"entry_id"`
	EntryCode  string           `json:"entry_code,omitempty"`
	EntryTitle string           `json:"entry_title,omitempty"`
	Kind       AnnotationKind   `json:"kind"`
	Status     AnnotationStatus `json:"status"`
	Quote      string           `json:"quote,omitempty"`
	UpdatedAt  time.Time        `json:"updated_at"`
}

type ResumeEntry struct {
	ID        string    `json:"id"`
	Code      string    `json:"code"`
	Title     string    `json:"title"`
	SectionID string    `json:"section_id"`
	UpdatedAt time.Time `json:"updated_at"`
	// AuthorKind is who wrote the newest revision — and the single most
	// load-bearing field in this list. A document a *human* edited after the
	// last session is a correction to build on, not stale work to redo, and an
	// agent that cannot tell the two apart overwrites the correction. Empty
	// when there is no revision row to ask.
	AuthorKind AuthorKind `json:"author_kind,omitempty"`
	// Revision is the newest numbered snapshot, so a caller can ask for the
	// change rather than re-reading the document. Zero means no revision row,
	// which is a document written before revisions existed.
	Revision int `json:"revision,omitempty"`
}

// ResumeActor says who the next action is waiting on. It is not a hint about
// importance: an answered annotation waits on a person because only a person
// accepts the work, and no amount of agent effort moves it.
type ResumeActor string

const (
	ResumeActorAgent ResumeActor = "agent"
	ResumeActorHuman ResumeActor = "human"
)

// ResumeReason is a closed vocabulary so a client can act on the reason
// without parsing the sentence beside it.
type ResumeReason string

const (
	ReasonTaskInProgress   ResumeReason = "task_in_progress"
	ReasonAnnotationOpen   ResumeReason = "annotation_open"
	ReasonTaskPending      ResumeReason = "task_pending"
	ReasonQuestionOpen     ResumeReason = "question_open"
	ReasonAnnotationAnswer ResumeReason = "annotation_awaiting_human"
	ReasonSessionAmbiguous ResumeReason = "session_selection_required"
)

// ResumeTarget names the entity an action is about, in both identities the
// product uses — the UUID a tool takes and the short code a URL is built from.
type ResumeTarget struct {
	Type  string `json:"type"`
	ID    string `json:"id,omitempty"`
	Code  string `json:"code,omitempty"`
	Title string `json:"title,omitempty"`
	// SessionCode and EntryCode are the parents a URL needs. A question is read
	// at /research/{R}/session/{SS}/question/{Q} and a mark is read on its
	// document, so an action naming only the question is an action nothing can
	// open.
	SessionCode string `json:"session_code,omitempty"`
	EntryCode   string `json:"entry_code,omitempty"`
}

// ResumeAction is one candidate continuation with the evidence for it. The
// reason is required: a suggestion a reader cannot check is a suggestion they
// have to trust, and this summary is deterministic precisely so they need not.
type ResumeAction struct {
	Kind       string       `json:"kind"`
	Target     ResumeTarget `json:"target"`
	ReasonCode ResumeReason `json:"reason_code"`
	Reason     string       `json:"reason"`
	Actor      ResumeActor  `json:"actor"`
	Tool       string       `json:"tool,omitempty"`
	Href       string       `json:"href,omitempty"`
}

// ResearchResume is the whole answer. It is a projection: nothing here is
// stored, and reading it changes nothing — no session is created, no status
// moves, no document is marked seen.
type ResearchResume struct {
	SchemaVersion int       `json:"schema_version"`
	GeneratedAt   time.Time `json:"generated_at"`

	Research ResumeResearch `json:"research"`
	Sessions ResumeSessions `json:"sessions"`

	Work struct {
		InProgress ResumeGroup[ResumeTask] `json:"in_progress"`
		Blocked    ResumeGroup[ResumeTask] `json:"blocked"`
		Pending    ResumeGroup[ResumeTask] `json:"pending"`
	} `json:"work"`

	Questions struct {
		Open     ResumeGroup[ResumeQuestion] `json:"open"`
		Deferred ResumeGroup[ResumeQuestion] `json:"deferred"`
	} `json:"questions"`

	Annotations struct {
		ToWork        ResumeGroup[ResumeAnnotation] `json:"to_work"`
		AwaitingHuman ResumeGroup[ResumeAnnotation] `json:"awaiting_human"`
	} `json:"annotations"`

	RecentEntries ResumeGroup[ResumeEntry] `json:"recent_entries"`

	NextActions []ResumeAction `json:"next_actions"`

	// Truncated is set when the size cap dropped content. Totals and has_more
	// survive truncation, so a group that lost items still reports how many
	// there are.
	Truncated bool `json:"truncated"`
	// Note explains a truncation or a degraded read in one sentence.
	Note string `json:"note,omitempty"`
}
