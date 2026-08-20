package domain

import "time"

// An annotation is a request for work addressed to a place in the text.
//
// That is the whole border with the rest of the product's trust machinery, and
// it is worth restating wherever this type is extended: evidence records what a
// block rests on, a claim status records how well established it is, an unknown
// records what we do not know. All three are written by the agent about its own
// knowledge. An annotation is the reader pointing at a sentence and saying what
// is wrong with it — and, unlike the other three, it has an addressee.

// AnnotationKind is what the mark asks for. The vocabulary is closed because
// each value is a different program for the agent working the queue; a
// free-form tag would turn "go work my annotations" into a mood.
type AnnotationKind string

const (
	// AnnotationVerify asks for a source: find one, cite it, or say plainly
	// that the claim could not be confirmed.
	AnnotationVerify AnnotationKind = "verify"
	// AnnotationDig asks for depth: a child entry, linked from the anchored
	// block so the document gains a way in rather than growing sideways.
	AnnotationDig AnnotationKind = "dig"
	// AnnotationDisagree records a dispute, and is the one that has to be
	// spelled out in every prompt that touches it: the agent must NOT edit the
	// text until the complaint goes away. A rewrite that satisfies the objection
	// destroys the disagreement instead of recording it. Both positions stay,
	// with who holds each.
	AnnotationDisagree AnnotationKind = "disagree"
)

func (k AnnotationKind) Valid() bool {
	switch k {
	case AnnotationVerify, AnnotationDig, AnnotationDisagree:
		return true
	}
	return false
}

// AnnotationStatus is the lifecycle. Deliberately four values and no more:
// priority, blocked, deferred and result all belong to Task, and adding any of
// them here builds a second todo list against the first.
type AnnotationStatus string

const (
	AnnotationOpen AnnotationStatus = "open"
	// AnnotationAnswered is as far as the agent may go. The work is done and
	// waiting for a person to agree that it is done.
	AnnotationAnswered AnnotationStatus = "answered"
	// AnnotationClosed is the human accepting the answer.
	AnnotationClosed AnnotationStatus = "closed"
	// AnnotationDismissed is the human withdrawing the mark — the doubt is no
	// longer worth holding. Kept rather than deleted: "we looked and decided it
	// did not matter" is a finding.
	AnnotationDismissed AnnotationStatus = "dismissed"
)

func (s AnnotationStatus) Valid() bool {
	switch s {
	case AnnotationOpen, AnnotationAnswered, AnnotationClosed, AnnotationDismissed:
		return true
	}
	return false
}

// Terminal reports whether the status is one only a human may set.
func (s AnnotationStatus) Terminal() bool {
	return s == AnnotationClosed || s == AnnotationDismissed
}

// AnchorState says where the marked text is now. It is computed on every read
// and never stored — a stored value goes stale the moment the agent writes.
type AnchorState string

const (
	// AnchorAnchored — the block is there and the quote is in it.
	AnchorAnchored AnchorState = "anchored"
	// AnchorDrifted — the block is there and the quote is gone. The most
	// valuable state in the set: the text under a mark changed, which is
	// precisely the moment a doubt is either answered or buried.
	AnchorDrifted AnchorState = "drifted"
	// AnchorMoved — the block is gone and the quote turned up elsewhere.
	// Carries the strategy and a confidence, because re-anchoring silently is
	// how an annotation ends up sitting on a sentence nobody marked.
	AnchorMoved AnchorState = "moved"
	// AnchorOrphaned — neither survives.
	//
	// Not garbage. It means the agent rewrote the paragraph somebody doubted:
	// either the rewrite answered the doubt or it buried an inconvenient claim,
	// and only a person can tell which. The row keeps its quote and its
	// anchored revision so entry_diff can show what happened.
	AnchorOrphaned AnchorState = "orphaned"
)

// Anchor strategies, reported alongside a resolved anchor so a reader can see
// how much to trust the placement.
const (
	AnchorByBlockID      = "block_id"
	AnchorByQuoteInBlock = "quote_in_block"
	AnchorByQuoteInDoc   = "quote_in_doc"
	AnchorNotFound       = "none"
)

const (
	// MaxQuoteLen caps the selection, in runes so a Cyrillic quote is not worth
	// half a Latin one. A selection past this is a request to annotate a
	// section, and the block it lives in is the honest unit for that.
	MaxQuoteLen = 2000
	// MaxQuoteContext is how much either side is kept to tell two identical
	// sentences apart. Enough to disambiguate, short enough that re-anchoring
	// does not fail because the surroundings were edited.
	MaxQuoteContext = 120
	// MaxAnnotationBody is the note itself.
	MaxAnnotationBody = 5000
	// MaxAnnotationResolution is what the agent writes back.
	MaxAnnotationResolution = 5000
	// MaxAnnotationBatch is the ceiling on one queue read.
	//
	// It lives here and is enforced by the service rather than being asked for
	// in a prompt, because a prompt limit is what a model ignores under load.
	// The number is a workable review batch for a person, which is the real
	// constraint: the pass is only as large as the diff somebody will read.
	MaxAnnotationBatch = 15
)

// Quote is what was selected, with enough context to find it again.
type Quote struct {
	Exact  string `json:"exact"`
	Prefix string `json:"prefix,omitempty"`
	Suffix string `json:"suffix,omitempty"`
}

// Anchor is the computed placement of a mark in the document as it stands now.
type Anchor struct {
	State AnchorState `json:"state"`
	// Strategy is how the placement was found. Reported always, including for
	// the trivial case, so a client never has to guess whether a match was
	// exact or approximate.
	Strategy string `json:"strategy"`
	// Confidence is 1 for an exact match and lower for a recovered one.
	Confidence float64 `json:"confidence"`
	// BlockID is where the mark resolves *now*, which differs from the stored
	// block_id exactly when the state is `moved`.
	BlockID string `json:"block_id,omitempty"`
	// BlockIndex is the block's position in render order, and BlockType what it
	// is. Both exist for the reader: a list of marks sorted by creation time
	// jumps around a document somebody is reading top to bottom, and a mark on a
	// table is not shown the way a mark on a paragraph is.
	BlockIndex int    `json:"block_index"`
	BlockType  string `json:"block_type,omitempty"`
	// Text is the current text of the anchored block, so a reader of the queue
	// can judge a mark without opening the document.
	Text string `json:"text,omitempty"`
}

// Rejection is one refusal of an answer, kept in full.
//
// The agent is required to read these before trying again; a mark that has
// collected two is one nobody can settle from inside the system, and the honest
// next move is a question rather than a third attempt.
type Rejection struct {
	Reason string `json:"reason"`
	// Revision is the document revision the refused answer left behind, so the
	// rejection can be read against what was actually written.
	Revision int       `json:"revision,omitempty"`
	At       time.Time `json:"at"`
}

// Annotation is one mark on one place in one document.
type Annotation struct {
	ID         string `json:"id"`
	Code       string `json:"code"`
	ResearchID string `json:"research_id"`
	EntryID    string `json:"entry_id"`
	// BlockID is empty for a markdown entry. The Anchor is then resolved by
	// quote alone, and says so.
	BlockID          string `json:"block_id,omitempty"`
	Quote            Quote  `json:"quote"`
	AnchoredRevision int    `json:"anchored_revision"`

	Kind       AnnotationKind `json:"kind"`
	Body       string         `json:"body,omitempty"`
	AuthorKind AuthorKind     `json:"author_kind"`
	UserID     string         `json:"user_id,omitempty"`

	Status           AnnotationStatus `json:"status"`
	Resolution       string           `json:"resolution,omitempty"`
	ResolvedRevision *int             `json:"resolved_revision,omitempty"`
	SessionID        string           `json:"session_id,omitempty"`
	TaskID           string           `json:"task_id,omitempty"`
	Rejections       []Rejection      `json:"rejections,omitempty"`
	Attempts         int              `json:"attempts"`

	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	AnsweredAt *time.Time `json:"answered_at,omitempty"`
	ClosedAt   *time.Time `json:"closed_at,omitempty"`

	// Anchor is computed on read. It is absent from list responses that did not
	// load the document, and a client must treat its absence as "unknown"
	// rather than as "anchored".
	Anchor *Anchor `json:"anchor,omitempty"`

	// EntryCode, EntryTitle and EntryType are filled for queue reads so the list
	// is legible without a join per row on the client.
	//
	// EntryType earns its place: an empty BlockID means both "this is a markdown
	// document, which has no blocks" and "the block this was pinned to is gone",
	// and those are different sentences on screen.
	EntryCode  string    `json:"entry_code,omitempty"`
	EntryTitle string    `json:"entry_title,omitempty"`
	EntryType  EntryType `json:"entry_type,omitempty"`
	// AuthorName is resolved for display. Everything else in this product shows
	// a name rather than an id, and a mark is the one record where "who doubted
	// this" is half the information.
	AuthorName string `json:"author_name,omitempty"`
}

// AnnotationReport tells a writer what its write did to marks it never
// mentioned. Attached to the response of an entry write and never stored — the
// same contract as BlockSaveReport, and for the same reason: an agent that
// buried a contested sentence has to be told in the call that did it.
type AnnotationReport struct {
	Drifted  []string `json:"annotations_drifted,omitempty"`
	Orphaned []string `json:"annotations_orphaned,omitempty"`
}

// Empty reports whether there is anything worth telling the writer.
func (r AnnotationReport) Empty() bool {
	return len(r.Drifted) == 0 && len(r.Orphaned) == 0
}
