package domain

import "time"

// SkillTier is where a skill lives, and it is also its precedence: where two
// skills disagree, the higher tier wins. The index the agent reads is sorted by
// it on the server, not only on screen, so what the model is told about
// precedence matches the order it receives.
type SkillTier string

const (
	// SkillBuiltin ships embedded in the binary and is refreshed at boot.
	// Nobody owns it and nobody may edit it in place; an edit forks a team copy.
	SkillBuiltin SkillTier = "builtin"
	// SkillTeam belongs to a team and is reusable across its researches.
	SkillTeam SkillTier = "team"
	// SkillPrivate belongs to exactly one research and is never offered in the
	// library. It is where a rule that applies only to this research lives.
	SkillPrivate SkillTier = "private"
)

func ValidSkillTier(t SkillTier) bool {
	return t == SkillBuiltin || t == SkillTeam || t == SkillPrivate
}

// SkillRank orders the index: research-private first, then team, then built-in.
// A lower number wins a conflict.
func SkillRank(t SkillTier) int {
	switch t {
	case SkillPrivate:
		return 0
	case SkillTeam:
		return 1
	default:
		return 2
	}
}

// Limits the service enforces. They are here rather than in the service because
// the API, the MCP tool and the tests all quote them, and three copies of a
// budget is how they drift.
const (
	// SkillDescriptionMax is the cap on the one field that is always in the
	// agent's context. Six skills at 200 characters is the whole index budget.
	SkillDescriptionMax = 200
	// SkillBodyMax is a hard ceiling, not a target: a skill above it is two
	// skills. ~4000 tokens at the usual 4 characters per token.
	SkillBodyMax = 16000
	// SkillsPerResearch counts only skills somebody chose. Ambient product
	// skills are excluded, or the budget is spent before the user arrives.
	SkillsPerResearch = 6
)

type Skill struct {
	ID string `json:"id"`
	// TeamID and ResearchID are the ownership pair: exactly one is set, or
	// neither for a built-in. Tier carries the same fact in the shape callers
	// actually branch on.
	TeamID     string    `json:"team_id,omitempty"`
	ResearchID string    `json:"research_id,omitempty"`
	UserID     string    `json:"user_id,omitempty"`
	Slug       string    `json:"slug"`
	Name       string    `json:"name"`
	Tier       SkillTier `json:"tier"`
	// Description is the trigger line. It says when to load the skill, never
	// what the body contains.
	Description string `json:"description"`
	// Body is omitted from every list response. Only skill_load and the
	// single-skill read carry it, which is what keeps the index small.
	Body string `json:"body,omitempty"`
	// Ambient marks a product skill: always in the index, never counted against
	// the cap, never detachable.
	Ambient bool `json:"ambient,omitempty"`
	// ForkedFrom is the built-in slug this row copies, when it is a fork.
	ForkedFrom string `json:"forked_from,omitempty"`
	// NeedsTrigger says the description was written by a migration rather than
	// by a person, so the UI can ask somebody to fix it.
	NeedsTrigger bool `json:"needs_trigger,omitempty"`
	// BodyTokens is an estimate, shown so the cost of a skill is visible where
	// it is attached. Derived, never stored.
	BodyTokens int       `json:"body_tokens"`
	Version    int       `json:"version"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`

	// Attached is set on library listings so a client never has to work out
	// what is already on by set arithmetic.
	Attached bool `json:"attached,omitempty"`
	// ViaTemplate and AttachedAt are set on the attached list. They describe
	// the attachment, not the skill, and are absent everywhere else.
	ViaTemplate bool       `json:"via_template,omitempty"`
	AttachedAt  *time.Time `json:"attached_at,omitempty"`
}

// EstimateTokens is the same rough measure the UI shows: characters over four.
// It is honest about being an estimate and it is stable, which matters more
// here than accuracy — its job is to make a 4000-word skill look expensive.
func EstimateTokens(s string) int {
	if s == "" {
		return 0
	}
	return (len([]rune(s)) + 3) / 4
}
