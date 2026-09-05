package domain

import "time"

type ResearchStatus string

const (
	ResearchActive    ResearchStatus = "active"
	ResearchCompleted ResearchStatus = "completed"
	ResearchArchived  ResearchStatus = "archived"
)

type Research struct {
	ID   string `json:"id"`
	Code string `json:"code"`
	// UserID is who created the research. It is a record, not a permission:
	// access comes from TeamID and the caller's role in that team.
	UserID string `json:"user_id,omitempty"`
	// TeamID is the owning team — the thing authorization actually consults.
	TeamID string `json:"team_id,omitempty"`
	// TeamName and Role are resolved per request for display, so a reader can
	// see whose team a research is in and the UI can hide controls the caller
	// cannot use. Neither is stored on the research.
	TeamName string   `json:"team_name,omitempty"`
	Role     TeamRole `json:"role,omitempty"`
	// TeamPersonal lets the UI leave a solo user's own researches unbadged —
	// labelling every card with your own name is noise, not information.
	TeamPersonal bool           `json:"team_is_personal,omitempty"`
	Name         string         `json:"name"`
	Description  string         `json:"description"`
	Goal         string         `json:"goal"`
	Status       ResearchStatus `json:"status"`
	Memory       Memory         `json:"memory"`
	Tags         []string       `json:"tags"`
	// TemplateSlug and TemplateVersion record the methodology this research was
	// started from, and which version of it. The version is stored because
	// built-in templates are refreshed from the binary on every boot: without
	// it, an upgrade would silently change the text behind a research already
	// in flight and nobody could tell afterwards which one was followed.
	// TemplateName is resolved for display and is never stored.
	TemplateSlug    string    `json:"template_slug,omitempty"`
	TemplateVersion int       `json:"template_version,omitempty"`
	TemplateName    string    `json:"template_name,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type SectionStatus string

const (
	SectionDraft     SectionStatus = "draft"
	SectionActive    SectionStatus = "active"
	SectionCompleted SectionStatus = "completed"
	SectionArchived  SectionStatus = "archived"
)

type Section struct {
	ID          string        `json:"id"`
	Code        string        `json:"code"`
	ResearchID  string        `json:"research_id"`
	Name        string        `json:"name"`
	DisplayName string        `json:"display_name"`
	Description string        `json:"description"`
	Status      SectionStatus `json:"status"`
	Position    int           `json:"position"`
	// FieldSpec is what documents in this section record. Empty is the normal
	// case and means this section accepts no metadata at all — which is what
	// lets a section that is a topic rather than a document class carry on
	// behaving exactly as it did before the feature existed.
	FieldSpec []FieldSpec `json:"field_spec"`
	// SpecVersion is bumped on every change to FieldSpec. Sections are
	// overwritten in place, so without it nothing records which declaration a
	// given document was written under.
	SpecVersion int       `json:"spec_version"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
