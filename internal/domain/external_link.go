package domain

import "time"

// ExternalLink represents a URL extracted from markdown content.
type ExternalLink struct {
	ID         string    `json:"id"`
	SourceType string    `json:"source_type"` // "entry", "question", "task"
	SourceID   string    `json:"source_id"`
	ResearchID string    `json:"research_id"`
	URL        string    `json:"url"`
	Title      string    `json:"title"`                 // link text from markdown [title](url)
	Domain     string    `json:"domain"`                // extracted hostname
	EntryCode  string    `json:"entry_code,omitempty"`  // enriched: source entry code
	EntryTitle string    `json:"entry_title,omitempty"` // enriched: source entry title
	CreatedAt  time.Time `json:"created_at"`
}
