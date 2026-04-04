package domain

import "time"

type EntryStatus string

const (
	EntryDraft     EntryStatus = "draft"
	EntryActive    EntryStatus = "active"
	EntryCompleted EntryStatus = "completed"
	EntryArchived  EntryStatus = "archived"
)

type Entry struct {
	ID          string      `json:"id"`
	ResearchID  string      `json:"research_id"`
	SectionID   string      `json:"section_id"`
	Title       string      `json:"title"`
	Content     string      `json:"content,omitempty"`
	Description string      `json:"description"`
	Status      EntryStatus `json:"status"`
	Tags        []string    `json:"tags"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}
