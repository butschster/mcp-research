package domain

import "time"

type SessionStatus string

const (
	SessionActive    SessionStatus = "active"
	SessionCompleted SessionStatus = "completed"
	SessionArchived  SessionStatus = "archived"
)

type Session struct {
	ID         string        `json:"id"`
	Code       string        `json:"code"`
	ResearchID string        `json:"research_id"`
	Title      string        `json:"title"`
	Focus      string        `json:"focus"`
	Status     SessionStatus `json:"status"`
	Notes      string        `json:"notes"`
	CreatedAt  time.Time     `json:"created_at"`
	UpdatedAt  time.Time     `json:"updated_at"`
}

type QuestionStatus string

const (
	QuestionPending    QuestionStatus = "pending"
	QuestionInProgress QuestionStatus = "in_progress"
	QuestionAnswered   QuestionStatus = "answered"
	QuestionDeferred   QuestionStatus = "deferred"
	QuestionSkipped    QuestionStatus = "skipped"
)

type Priority string

const (
	PriorityHigh   Priority = "high"
	PriorityMedium Priority = "medium"
	PriorityLow    Priority = "low"
)

type Question struct {
	ID        string         `json:"id"`
	Code      string         `json:"code"`
	SessionID string         `json:"session_id"`
	Text      string         `json:"text"`
	Area      string         `json:"area"`
	Rationale string         `json:"rationale"`
	Priority  Priority       `json:"priority"`
	Status    QuestionStatus `json:"status"`
	Answer    string         `json:"answer"`
	ParentID  string         `json:"parent_id,omitempty"`
	Position  int            `json:"position"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}
