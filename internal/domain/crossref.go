package domain

// CrossRef represents a link extracted from [[...]] patterns in content.
// Source can be an entry, question, or task.
type CrossRef struct {
	SourceType       string `json:"source_type"` // "entry", "question", "task"
	SourceID         string `json:"source_id"`
	SourceResearchID string `json:"source_research_id"`
	TargetEntryID    string `json:"target_entry_id"`
	TargetResearchID string `json:"target_research_id"`
	TargetRef        string `json:"target_ref"` // raw reference text, e.g. "E3" or "R2:E5"
	Resolved         bool   `json:"resolved"`
}
