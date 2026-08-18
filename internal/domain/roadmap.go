package domain

import "time"

type RoadmapStatus string

const (
	RoadmapActive    RoadmapStatus = "active"
	RoadmapCompleted RoadmapStatus = "completed"
	RoadmapArchived  RoadmapStatus = "archived"
)

// RoadmapView is which layout a roadmap opens in. The same node-edge data is
// rendered three ways; the value is the default, not a lock — the frontend
// toggle overrides it locally.
type RoadmapView string

const (
	// RoadmapViewGraph is the original free-form node-edge graph (x/y positions).
	RoadmapViewGraph RoadmapView = "graph"
	// RoadmapViewStages groups nodes into ordered stage columns.
	RoadmapViewStages RoadmapView = "stages"
	// RoadmapViewTimeline places nodes left-to-right by node_date.
	RoadmapViewTimeline RoadmapView = "timeline"
)

type Roadmap struct {
	ID          string        `json:"id"`
	Code        string        `json:"code"`
	ResearchID  string        `json:"research_id"`
	Title       string        `json:"title"`
	Description string        `json:"description,omitempty"`
	Statuses    []string      `json:"statuses"`
	Status      RoadmapStatus `json:"status"`
	// Stages is the ordered stage vocabulary — the column order for the stages
	// view. A name here with no nodes is a legitimately empty column. It relates
	// to a node's Stage exactly as Statuses relates to a node's Status.
	Stages []string `json:"stages"`
	// View is the layout this roadmap opens in: graph / stages / timeline.
	View      RoadmapView    `json:"view"`
	Nodes     []*RoadmapNode `json:"nodes,omitempty"`
	Edges     []*RoadmapEdge `json:"edges,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
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
	ID          string  `json:"id"`
	Code        string  `json:"code"`
	RoadmapID   string  `json:"roadmap_id"`
	Title       string  `json:"title"`
	Description string  `json:"description,omitempty"`
	NodeType    string  `json:"node_type"`
	Status      string  `json:"status,omitempty"`
	PositionX   float64 `json:"position_x"`
	PositionY   float64 `json:"position_y"`
	ParentID    string  `json:"parent_id,omitempty"`
	// Reference fields: link node to a research entity
	RefType  string `json:"ref_type,omitempty"`
	RefID    string `json:"ref_id,omitempty"`
	Metadata string `json:"metadata,omitempty"` // JSON blob for node-type-specific data (checklist items, URL, etc.)
	// Stage is the column this node sits in for the stages view. Matched by
	// string against the roadmap's Stages, the same way Status matches Statuses:
	// a value not in the list falls into a leftover column rather than erroring.
	Stage string `json:"stage,omitempty"`
	// NodeDate places the node on the timeline view. ISO 'YYYY-MM-DD' or empty;
	// empty means undated and the timeline sets it aside rather than guessing.
	NodeDate string `json:"node_date,omitempty"`
	// Resolved reference data (populated at read time, not stored)
	RefData   *RoadmapNodeRefData `json:"ref_data,omitempty"`
	CreatedAt time.Time           `json:"created_at"`
	UpdatedAt time.Time           `json:"updated_at"`
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
