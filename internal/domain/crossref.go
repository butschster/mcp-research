package domain

// CrossRef represents a link extracted from [[...]] patterns in content.
// Source can be an entry, question, or task.
// Target can be an entry, research, roadmap, or roadmap node.
type CrossRef struct {
	SourceType       string `json:"source_type"` // "entry", "question", "task"
	SourceID         string `json:"source_id"`
	SourceResearchID string `json:"source_research_id"`
	TargetEntryID    string `json:"target_entry_id,omitempty"`
	TargetResearchID string `json:"target_research_id,omitempty"`
	TargetRoadmapID  string `json:"target_roadmap_id,omitempty"`
	TargetNodeID     string `json:"target_node_id,omitempty"`
	TargetRef        string `json:"target_ref"` // raw reference text, e.g. "E3", "R2:E5", "RM1", "RM1:N3"
	Resolved         bool   `json:"resolved"`
}
