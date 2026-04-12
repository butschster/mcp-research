package domain

import "time"

type ResearchStatus string

const (
	ResearchActive    ResearchStatus = "active"
	ResearchCompleted ResearchStatus = "completed"
	ResearchArchived  ResearchStatus = "archived"
)

type Research struct {
	ID          string         `json:"id"`
	Code        string         `json:"code"`
	UserID      string         `json:"user_id,omitempty"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Goal        string         `json:"goal"`
	Status      ResearchStatus `json:"status"`
	Instruction string         `json:"instruction"`
	Memory      []string       `json:"memory"`
	Tags        []string       `json:"tags"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

type SectionStatus string

const (
	SectionDraft     SectionStatus = "draft"
	SectionActive    SectionStatus = "active"
	SectionCompleted SectionStatus = "completed"
	SectionArchived  SectionStatus = "archived"
)

// DefaultEntryStatuses are used when a section doesn't specify custom statuses.
var DefaultEntryStatuses = []string{
	string(EntryDraft),
	string(EntryActive),
	string(EntryCompleted),
	string(EntryArchived),
}

type Section struct {
	ID                   string        `json:"id"`
	Code                 string        `json:"code"`
	ResearchID           string        `json:"research_id"`
	Name                 string        `json:"name"`
	DisplayName          string        `json:"display_name"`
	Description          string        `json:"description"`
	Status               SectionStatus `json:"status"`
	Position             int           `json:"position"`
	AllowedEntryStatuses []string      `json:"allowed_entry_statuses"`
	CreatedAt            time.Time     `json:"created_at"`
	UpdatedAt            time.Time     `json:"updated_at"`
}

// DefaultEntryStatus returns the first allowed entry status (used as default for new entries).
func (s *Section) DefaultEntryStatus() EntryStatus {
	if len(s.AllowedEntryStatuses) > 0 {
		return EntryStatus(s.AllowedEntryStatuses[0])
	}
	return EntryDraft
}

// IsEntryStatusAllowed checks if a given status is in the section's allowed list.
func (s *Section) IsEntryStatusAllowed(status EntryStatus) bool {
	for _, allowed := range s.AllowedEntryStatuses {
		if allowed == string(status) {
			return true
		}
	}
	return false
}
