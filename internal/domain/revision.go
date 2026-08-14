package domain

import "time"

// AuthorKind says what wrote a revision. It is the field a reader looks at first
// when deciding how much to trust a document: "a person wrote this" and "a model
// wrote this" are different claims, and until now the product could not tell
// them apart.
type AuthorKind string

const (
	// AuthorAgent is a write that arrived over MCP, or over the REST API with a
	// machine credential.
	AuthorAgent AuthorKind = "agent"
	// AuthorHuman is a write from the web UI or from a browser session.
	AuthorHuman AuthorKind = "human"
	// AuthorImport is a write made by importing a research.
	AuthorImport AuthorKind = "import"
	// AuthorRestore is a write produced by restoring an earlier revision.
	AuthorRestore AuthorKind = "restore"
)

func (k AuthorKind) Valid() bool {
	switch k {
	case AuthorAgent, AuthorHuman, AuthorImport, AuthorRestore:
		return true
	}
	return false
}

// EntryRevision is one snapshot of an entry, as it looked after a write.
//
// Content is omitted from list responses — a history of a long entry is
// otherwise dominated by copies of it — and present when a single revision is
// read or diffed.
type EntryRevision struct {
	ID          string      `json:"id"`
	EntryID     string      `json:"entry_id"`
	ResearchID  string      `json:"research_id"`
	Revision    int         `json:"revision"`
	Title       string      `json:"title"`
	Description string      `json:"description"`
	Content     string      `json:"content,omitempty"`
	Type        EntryType   `json:"entry_type"`
	Status      EntryStatus `json:"status"`
	Tags        []string    `json:"tags"`
	AuthorKind  AuthorKind  `json:"author_kind"`
	SessionID   string      `json:"session_id,omitempty"`
	UserID      string      `json:"user_id,omitempty"`
	Summary     string      `json:"summary,omitempty"`
	CreatedAt   time.Time   `json:"created_at"`

	// Enriched on read, never stored.
	SessionCode  string `json:"session_code,omitempty"`
	SessionTitle string `json:"session_title,omitempty"`
}

// SameContent reports whether a revision would be an exact copy of what is
// already stored. A write that changes nothing must not produce a revision —
// an agent that rewrites the same paragraph three times in one session should
// leave one entry in the history, not three.
func (r *EntryRevision) SameContent(other *EntryRevision) bool {
	if r == nil || other == nil {
		return false
	}
	if r.Title != other.Title ||
		r.Description != other.Description ||
		r.Content != other.Content ||
		r.Type != other.Type ||
		r.Status != other.Status {
		return false
	}
	if len(r.Tags) != len(other.Tags) {
		return false
	}
	for i := range r.Tags {
		if r.Tags[i] != other.Tags[i] {
			return false
		}
	}
	return true
}
