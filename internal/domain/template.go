package domain

import "time"

// TemplateTier is where a template lives. Unlike a skill's tier this is not a
// precedence — two templates never apply at once, because exactly one is chosen
// at kickoff and then the template is done. It is only ownership.
type TemplateTier string

const (
	// TemplateGlobal ships with the binary and is refreshed at boot. Nobody
	// owns it; editing one forks a team copy.
	TemplateGlobal TemplateTier = "global"
	// TemplateTeam belongs to a team and is theirs alone.
	TemplateTeam TemplateTier = "team"
)

// Limits the service enforces. The two matching fields are capped because they
// are what `template_list` returns to every kickoff — the body is not, so it is
// allowed to be a real document.
const (
	// TemplateCriterionMax bounds `when_to_use` and `when_not_to_use`. Six
	// templates times two lines is the whole of what an agent reads before
	// choosing, so it has to stay small.
	TemplateCriterionMax = 240
	// TemplateBodyMax is generous on purpose: a kickoff methodology is read
	// once, in full, by a model that has just started. It is not the always-in-
	// context budget a skill description lives under.
	TemplateBodyMax = 24000
)

// TemplateSkill is one entry of a template's skill list, resolved for display.
// `Missing` is the interesting one: a built-in naming an unknown skill fails the
// boot, but nothing stops a team skill being deleted after a template named it.
type TemplateSkill struct {
	Slug        string `json:"slug"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Ambient     bool   `json:"ambient,omitempty"`
	Missing     bool   `json:"missing,omitempty"`
}

type Template struct {
	ID string `json:"id"`
	// TeamID is empty for a global template. Tier carries the same fact in the
	// shape callers branch on.
	TeamID string       `json:"team_id,omitempty"`
	UserID string       `json:"user_id,omitempty"`
	Slug   string       `json:"slug"`
	Name   string       `json:"name"`
	Tier   TemplateTier `json:"tier"`
	// Description is one line for a picker. WhenToUse and WhenNotToUse are what
	// an agent actually matches on.
	Description  string `json:"description"`
	WhenToUse    string `json:"when_to_use"`
	WhenNotToUse string `json:"when_not_to_use"`
	// Body is omitted from every list response: it is the document, and
	// shipping it in a list would put four methodologies into a kickoff that
	// needs one.
	Body string `json:"body,omitempty"`
	// Skills are slugs attached to the research this template starts.
	Skills []string `json:"skills"`
	// ForkedFrom is the global slug this copies, when it is a fork.
	ForkedFrom string `json:"forked_from,omitempty"`
	// SkillsResolved is set on a single-template read, so a reader sees what
	// the methodology will attach rather than a list of slugs. Derived.
	SkillsResolved []TemplateSkill `json:"skills_resolved,omitempty"`
	// ResearchCount is how many researches this has started, among those the
	// caller can read. Derived, and computed only on a single read — omitted
	// rather than sent as zero from a list, because a list row reading
	// "used 0 times" for a template in daily use is worse than saying nothing.
	ResearchCount int `json:"research_count,omitempty"`
	// BodyWords is a size hint for a reader deciding whether to open it, and it
	// is an estimate: a list cannot afford to fetch the body, so it is counted
	// in SQL by whitespace rather than by tokenising. Derived, never stored.
	BodyWords int       `json:"body_words"`
	Version   int       `json:"version"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
