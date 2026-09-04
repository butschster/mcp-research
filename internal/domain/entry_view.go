package domain

import "time"

// EntryView is the last numbered snapshot one reader actually saw.
//
// ViewerID is an internal lookup key and is never accepted from an API caller.
// It is a user id when accounts are enabled and `local` for the single reader
// of an auth-disabled installation.
type EntryView struct {
	ViewerID     string    `json:"-"`
	UserID       string    `json:"-"`
	EntryID      string    `json:"entry_id"`
	SeenRevision int       `json:"seen_revision"`
	SeenAt       time.Time `json:"seen_at"`
}

// EntryUpdate is one row in a reader's personal document-update queue.
// Content is intentionally absent: the queue stays cheap, and opening the
// document is what marks the exact revision that was rendered as seen.
type EntryUpdate struct {
	EntryID         string      `json:"entry_id"`
	EntryCode       string      `json:"entry_code"`
	ResearchID      string      `json:"research_id"`
	SectionID       string      `json:"section_id"`
	Title           string      `json:"title"`
	Description     string      `json:"description,omitempty"`
	Type            EntryType   `json:"entry_type"`
	Status          EntryStatus `json:"status"`
	CurrentRevision int         `json:"current_revision"`
	SeenRevision    int         `json:"seen_revision"`
	UnseenRevisions int         `json:"unseen_revisions"`
	Kind            string      `json:"kind"` // new | changed
	UpdatedAt       time.Time   `json:"updated_at"`
}

// EntryViewState decorates a single document response without changing the
// entry itself. The caller captures this pair before marking CurrentRevision
// seen, so it can still open the exact old-to-new comparison on this visit.
type EntryViewState struct {
	CurrentRevision int    `json:"current_revision"`
	SeenRevision    int    `json:"seen_revision"`
	UnseenRevisions int    `json:"unseen_revisions"`
	Kind            string `json:"kind"` // new | changed | seen
}

// SeenRevision is supplied by the UI when it marks a rendered snapshot seen.
// The server never substitutes the newest revision: doing so could consume a
// write that landed after the page fetched but before this request arrived.
type SeenRevision struct {
	EntryID  string `json:"entry_id"`
	Revision int    `json:"revision"`
}
