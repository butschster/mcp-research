package tools

import (
	"errors"

	"github.com/butschster/mcp-research/internal/domain"
	"github.com/butschster/mcp-research/internal/service"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// skillPayload is the shape every skill tool answers with, and it never carries
// a body.
//
// That is the whole design of this feature restated at the transport: the index
// is small and the bodies are one `skill_load` away. A management tool that
// returned the text it had just written would put a 2500-token document into
// the context of an agent that was tidying up, which is exactly the cost the
// progressive-disclosure split exists to avoid.
func skillPayload(sk *domain.Skill) map[string]any {
	out := map[string]any{
		"skill_id":    sk.ID,
		"slug":        sk.Slug,
		"name":        sk.Name,
		"tier":        sk.Tier,
		"description": sk.Description,
		"version":     sk.Version,
	}
	if sk.BodyTokens > 0 {
		out["body_tokens"] = sk.BodyTokens
	}
	if sk.Ambient {
		// Said out loud rather than left to be inferred from the tier: an
		// ambient skill is outside the budget and cannot be detached, and both
		// facts are things an agent otherwise learns by being refused.
		out["ambient"] = true
	}
	if sk.ForkedFrom != "" {
		out["forked_from"] = sk.ForkedFrom
	}
	if sk.NeedsTrigger {
		out["needs_trigger"] = true
	}
	// Only the attached listing carries it, and it is the only ordering signal
	// there is when the cap is full and something has to go. Without it the
	// choice of what to drop is a guess between six lines that all look equally
	// current.
	if sk.AttachedAt != nil {
		out["attached_at"] = sk.AttachedAt
	}
	return out
}

// skillErrorResult refuses with the machine-readable code alongside the
// sentence.
//
// An agent reading only prose cannot tell "this research already follows six"
// from "that one is already on" — the first means pick something to drop, the
// second means carry on — and it will retry the wrong one. The code is the same
// vocabulary the REST API answers with, from the same switch.
func skillErrorResult(err error) (*mcp.CallToolResult, any, error) {
	code := service.SkillErrorCode(err)
	if code == "" {
		return errorResult(err.Error())
	}
	// The cap is the one refusal an agent meets mid-task with a plan it now
	// cannot carry out, and the sentence it used to get named no way forward.
	// The remedy lived in skill_list's *successful* answer — readable only by an
	// agent that had already done the thing that would have avoided the error.
	if code == "skill_cap_reached" {
		return errorResult(code + ": " + err.Error() +
			" Six is the whole budget and there is no way to raise it; the always-on product skills are outside it and cannot be dropped. Call skill_list to see what this research follows, then skill_detach to free a slot — and note that detaching a research-private skill deletes it.")
	}
	return errorResult(code + ": " + err.Error())
}

// skillErrorForSlug is skillErrorResult with the refusal that needs the slug in
// it to be worth anything.
//
// Every tool here addresses a skill by slug, and a wrong slug used to answer
// with the bare string "not found" — no code, unlike every other refusal in
// this feature, and no way forward. It is also the mistake most available to an
// agent: a slug is fixed when the skill is created, never tracks a later
// rename, and a name with no Latin letters produces something like
// `skill-7c07487a`. So a model that derives a slug from a name is wrong more or
// less at random, and the answer has to say that rather than imply the skill is
// missing.
//
// Only for the slug form. Addressed by id, "not found" and "not yours" are
// deliberately the same answer, and inviting the caller to go looking would
// turn that into a hint.
func skillErrorForSlug(err error, slug string) (*mcp.CallToolResult, any, error) {
	if errors.Is(err, service.ErrNotFound) {
		return errorResult("not_found: nothing in this research answers to the slug \"" + slug + "\". " +
			"A slug is fixed when a skill is created and does not follow a later rename, so it cannot be derived from a name — call skill_list for the ones this research can address.")
	}
	return skillErrorResult(err)
}

// skillInput builds the write payload, inheriting from what is already stored
// for every field the caller left out.
//
// Partial edits are the point. Over REST the frontend holds a whole form and
// sends all three fields; an agent fixing one trigger line holds nothing, and
// making it resend a 2000-word body to change a sentence means reading the body
// first — a load it did not need. `current` is nil when there is nothing to
// inherit from, which is a create.
func skillInput(current *domain.Skill, name, description, body *string) service.SkillInput {
	in := service.SkillInput{}
	if current != nil {
		in.Name, in.Description, in.Body = current.Name, current.Description, current.Body
	}
	if name != nil {
		in.Name = *name
	}
	if description != nil {
		in.Description = *description
	} else if current != nil {
		// Carried through so Update can tell a rewritten trigger line from one
		// that was merely read back out of the row and written again.
		in.DescriptionUntouched = true
	}
	if body != nil {
		in.Body = *body
	}
	return in
}
