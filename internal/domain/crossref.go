package domain

// CrossRef represents a link from one entry to another, extracted from [[...]] patterns in content.
type CrossRef struct {
	SourceEntryID    string `json:"source_entry_id"`
	SourceResearchID string `json:"source_research_id"`
	TargetEntryID    string `json:"target_entry_id"`
	TargetResearchID string `json:"target_research_id"`
	TargetRef        string `json:"target_ref"` // raw reference text, e.g. "E3" or "R2:E5"
	Resolved         bool   `json:"resolved"`
}
