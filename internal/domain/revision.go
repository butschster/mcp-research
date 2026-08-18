package domain

import (
	"encoding/json"
	"sort"
	"time"
)

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
	// Metadata is the entry's section-declared values as they stood after this
	// write. Without it a metadata edit would leave no trace in history — and,
	// because SameContent decides whether a revision is written at all, it
	// would not merely go unrecorded, it would be judged a no-op and vanish.
	Metadata    map[string]any `json:"metadata,omitempty"`
	SpecVersion int            `json:"spec_version,omitempty"`
	AuthorKind  AuthorKind     `json:"author_kind"`
	SessionID   string         `json:"session_id,omitempty"`
	UserID      string         `json:"user_id,omitempty"`
	Summary     string         `json:"summary,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`

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
	return sameMetadata(r.Metadata, other.Metadata)
}

// sameMetadata compares two value maps by their JSON encoding.
//
// Comparing the encodings rather than the maps is deliberate: the values arrive
// as `any` from a JSON decode and from a database column, so a number is
// float64 down one path and could be int down another, and == on `any` would
// then report a change nobody made. Encoding both sides normalises that, and
// the maps are a dozen short values, so the cost does not matter.
func sameMetadata(a, b map[string]any) bool {
	if len(a) == 0 && len(b) == 0 {
		return true
	}
	if len(a) != len(b) {
		return false
	}
	ja, err := json.Marshal(sortedPairs(a))
	if err != nil {
		return false
	}
	jb, err := json.Marshal(sortedPairs(b))
	if err != nil {
		return false
	}
	return string(ja) == string(jb)
}

// sortedPairs renders a map as an ordered slice, because Go's JSON encoder
// sorts map keys but not the values inside them consistently enough to rely on
// for equality.
func sortedPairs(m map[string]any) [][2]any {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	out := make([][2]any, 0, len(keys))
	for _, k := range keys {
		out = append(out, [2]any{k, m[k]})
	}
	return out
}
