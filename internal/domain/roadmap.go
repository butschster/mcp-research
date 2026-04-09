package domain

import "time"

type RoadmapStatus string

const (
	RoadmapActive    RoadmapStatus = "active"
	RoadmapCompleted RoadmapStatus = "completed"
	RoadmapArchived  RoadmapStatus = "archived"
)

type Roadmap struct {
	ID          string        `json:"id"`
	Code        string        `json:"code"`
	ResearchID  string        `json:"research_id"`
	Title       string        `json:"title"`
	Description string        `json:"description,omitempty"`
	Statuses    []string      `json:"statuses"`
	Status      RoadmapStatus `json:"status"`
	Nodes       []*RoadmapNode `json:"nodes,omitempty"`
	Edges       []*RoadmapEdge `json:"edges,omitempty"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
}

// RoadmapNodeRefType defines what entity a node references.
type RoadmapNodeRefType string

const (
	RefTypeEntry    RoadmapNodeRefType = "entry"
	RefTypeTask     RoadmapNodeRefType = "task"
	RefTypeSession  RoadmapNodeRefType = "session"
	RefTypeResearch RoadmapNodeRefType = "research"
	RefTypeQuestion RoadmapNodeRefType = "question"
)

// RoadmapNodeRefData holds resolved data from the referenced entity.
// Populated at read time (lazy sync), never stored in DB.
type RoadmapNodeRefData struct {
	Title       string `json:"title,omitempty"`
	Status      string `json:"status,omitempty"`
	Code        string `json:"code,omitempty"`
	Description string `json:"description,omitempty"`
	ResearchID  string `json:"research_id,omitempty"`
	// Entry-specific
	SectionName string `json:"section_name,omitempty"`
	Content     string `json:"content,omitempty"`
	// Task-specific
	Priority string `json:"priority,omitempty"`
	Result   string `json:"result,omitempty"`
	// Session-specific
	TotalQuestions    int `json:"total_questions,omitempty"`
	AnsweredQuestions int `json:"answered_questions,omitempty"`
	// Research-specific
	SectionCount int `json:"section_count,omitempty"`
	EntryCount   int `json:"entry_count,omitempty"`
}

type RoadmapNode struct {
	ID          string    `json:"id"`
	Code        string    `json:"code"`
	RoadmapID   string    `json:"roadmap_id"`
	Title       string    `json:"title"`
	Description string    `json:"description,omitempty"`
	NodeType    string    `json:"node_type"`
	Status      string    `json:"status,omitempty"`
	PositionX   float64   `json:"position_x"`
	PositionY   float64   `json:"position_y"`
	ParentID    string    `json:"parent_id,omitempty"`
	// Reference fields: link node to a research entity
	RefType  string `json:"ref_type,omitempty"`
	RefID    string `json:"ref_id,omitempty"`
	Metadata string `json:"metadata,omitempty"` // JSON blob for node-type-specific data (checklist items, URL, etc.)
	// Resolved reference data (populated at read time, not stored)
	RefData *RoadmapNodeRefData `json:"ref_data,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type RoadmapEdge struct {
	ID           string    `json:"id"`
	RoadmapID    string    `json:"roadmap_id"`
	SourceNodeID string    `json:"source_node_id"`
	TargetNodeID string    `json:"target_node_id"`
	Label        string    `json:"label,omitempty"`
	EdgeType     string    `json:"edge_type"`
	CreatedAt    time.Time `json:"created_at"`
}
