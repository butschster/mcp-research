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
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
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
