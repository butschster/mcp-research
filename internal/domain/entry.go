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
	// EntryArtifact means Content is a self-contained HTML document, rendered in a
	// sandboxed iframe rather than as markdown.
	EntryArtifact EntryType = "artifact"
)

func (t EntryType) Valid() bool {
	return t == EntryMarkdown || t == EntryArtifact
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
