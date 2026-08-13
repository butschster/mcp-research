package domain

import "time"

type EntryStatus string

const (
	EntryDraft     EntryStatus = "draft"
	EntryActive    EntryStatus = "active"
	EntryCompleted EntryStatus = "completed"
	EntryArchived  EntryStatus = "archived"
)

// EntryType selects how Content is interpreted when rendered.
type EntryType string

const (
	// EntryMarkdown is the default: Content is markdown.
	EntryMarkdown EntryType = "markdown"
	// EntryBlocks means Content is a block document — see block.go. Structured,
	// safe to render, indexable and exportable to markdown.
	EntryBlocks EntryType = "blocks"
	// EntryArtifact is accepted on input as sugar for a blocks document holding a
	// single `html` block, and is normalized to EntryBlocks on the way in. It is
	// no longer a stored type. Kept so clients that learned it keep working.
	EntryArtifact EntryType = "artifact"
)

// Valid reports whether the type may be supplied by a caller.
func (t EntryType) Valid() bool {
	return t == EntryMarkdown || t == EntryBlocks || t == EntryArtifact
}

// Stored reports whether the type may be persisted. EntryArtifact is an input
// alias only — storing it would leave two rendering paths for one thing.
func (t EntryType) Stored() bool {
	return t == EntryMarkdown || t == EntryBlocks
}

type Entry struct {
	ID          string      `json:"id"`
	Code        string      `json:"code"`
	ResearchID  string      `json:"research_id"`
	SectionID   string      `json:"section_id"`
	SessionID   string      `json:"session_id,omitempty"`
	Type        EntryType   `json:"entry_type"`
	Title       string      `json:"title"`
	Content     string      `json:"content,omitempty"`
	Description string      `json:"description"`
	Status      EntryStatus `json:"status"`
	Tags        []string    `json:"tags"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}
