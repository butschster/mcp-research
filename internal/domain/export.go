package domain

import "time"

// ExportData is the portable representation of a complete research.
// It contains all data needed to recreate a research on another server.
type ExportData struct {
	Version    int            `json:"version"`
	ExportedAt time.Time      `json:"exported_at"`
	Research   ExportResearch `json:"research"`
}

type ExportResearch struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Goal        string         `json:"goal"`
	Status      ResearchStatus `json:"status"`
	Instruction string         `json:"instruction,omitempty"`
	Memory      []string       `json:"memory,omitempty"`
	Tags        []string       `json:"tags,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`

	Sections []ExportSection `json:"sections"`
	Sessions []ExportSession `json:"sessions,omitempty"`
	Tasks    []ExportTask    `json:"tasks,omitempty"`
	Roadmaps []ExportRoadmap `json:"roadmaps,omitempty"`
}

type ExportSection struct {
	Name        string        `json:"name"`
	DisplayName string        `json:"display_name,omitempty"`
	Description string        `json:"description,omitempty"`
	Status      SectionStatus `json:"status"`
	Position    int           `json:"position"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`

	Entries []ExportEntry `json:"entries,omitempty"`
}

type ExportEntry struct {
	Title       string      `json:"title"`
	Content     string      `json:"content"`
	Description string      `json:"description,omitempty"`
	Status      EntryStatus `json:"status"`
	Tags        []string    `json:"tags,omitempty"`
	SessionCode string      `json:"session_code,omitempty"` // links to ExportSession.Code
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}

type ExportSession struct {
	Code      string        `json:"code"` // for entry->session linking
	Title     string        `json:"title"`
	Focus     string        `json:"focus,omitempty"`
	Status    SessionStatus `json:"status"`
	Notes     string        `json:"notes,omitempty"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`

	Questions []ExportQuestion `json:"questions,omitempty"`
}

type ExportQuestion struct {
	Text      string         `json:"text"`
	Area      string         `json:"area,omitempty"`
	Rationale string         `json:"rationale,omitempty"`
	Priority  Priority       `json:"priority"`
	Status    QuestionStatus `json:"status"`
	Answer    string         `json:"answer,omitempty"`
	Position  int            `json:"position"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`

	Children []ExportQuestion `json:"children,omitempty"`
}

type ExportTask struct {
	Title       string     `json:"title"`
	Description string     `json:"description,omitempty"`
	Status      TaskStatus `json:"status"`
	Priority    Priority   `json:"priority"`
	Result      string     `json:"result,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

type ExportRoadmap struct {
	Title       string        `json:"title"`
	Description string        `json:"description,omitempty"`
	Statuses    []string      `json:"statuses,omitempty"`
	Status      RoadmapStatus `json:"status"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`

	Nodes []ExportRoadmapNode `json:"nodes,omitempty"`
	Edges []ExportRoadmapEdge `json:"edges,omitempty"`
}

type ExportRoadmapNode struct {
	Code        string    `json:"code"` // for edge/parent references
	Title       string    `json:"title"`
	Description string    `json:"description,omitempty"`
	NodeType    string    `json:"node_type"`
	Status      string    `json:"status,omitempty"`
	PositionX   float64   `json:"position_x"`
	PositionY   float64   `json:"position_y"`
	ParentCode  string    `json:"parent_code,omitempty"` // links to another node's Code
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ExportRoadmapEdge struct {
	SourceCode string `json:"source_code"` // node Code
	TargetCode string `json:"target_code"` // node Code
	Label      string `json:"label,omitempty"`
	EdgeType   string `json:"edge_type"`
}
