package tools

import (
	"context"
	"log/slog"

	"github.com/dovod-app/app/internal/domain"
	"github.com/dovod-app/app/internal/service"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type SkillUpdateInput struct {
	// Two ways in, because there are two kinds of caller. An agent working in a
	// research holds slugs and nothing else; one that has just written a team
	// skill holds the id it was handed back and no research to look it up
	// through.
	ResearchID *string `json:"research_id,omitempty" jsonschema:"Research to resolve the slug against. UUID or short code (R1). Give this with slug, or give skill_id"`
	Slug       *string `json:"slug,omitempty" jsonschema:"Slug of the skill to edit, from the skills index"`
	SkillID    *string `json:"skill_id,omitempty" jsonschema:"Id of the skill to edit, from skill_create or skill_list. Alternative to research_id plus slug"`

	Name        *string `json:"name,omitempty" jsonschema:"New name. The slug never follows a rename. Omit to leave it alone"`
	Description *string `json:"description,omitempty" jsonschema:"New trigger line, at most 200 characters. Omit to leave it alone"`
	Body        *string `json:"body,omitempty" jsonschema:"New markdown body, at most 16000 characters. Omit to leave it alone"`
}

// RegisterSkillUpdate edits a skill in place. Omitted fields are inherited from
// what is stored, so changing a trigger line does not mean resending a body the
// agent would have had to load first.
func RegisterSkillUpdate(srv *mcp.Server, skillSvc *service.SkillService, researchSvc *service.ResearchService, log *slog.Logger) {
	mcp.AddTool(srv, &mcp.Tool{
		Name: "skill_update",
		Description: "Edits a team or research-private skill in place. Send only the fields you are changing. " +
			"Built-in skills are refused: editing one is a fork, which is skill_fork. Every research following the skill sees the change.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input SkillUpdateInput) (*mcp.CallToolResult, any, error) {
		current, fail := resolveSkillArg(ctx, skillSvc, researchSvc, input.ResearchID, input.Slug, input.SkillID, true)
		if fail != nil {
			return fail()
		}
		if input.Name == nil && input.Description == nil && input.Body == nil {
			return validationErrorResult([]string{"nothing to change: send name, description or body"})
		}
		if current.Tier == domain.SkillBuiltin {
			// Not writeSkillError's `not_allowed` from the service, because the
			// service's sentence tells a human what forking means and this
			// caller needs the tool that does it.
			return errorResult("not_allowed: a built-in skill is rewritten from the binary on every upgrade, so an edit here would be lost. Call skill_fork with the same research_id and slug — it copies the built-in into your team, applies the edit, and moves this research's attachment onto the copy in one step.")
		}

		sk, err := skillSvc.Update(ctx, current.ID, skillInput(current, input.Name, input.Description, input.Body))
		if err != nil {
			log.Error("skill_update failed", "skill_id", current.ID, "error", err)
			return skillErrorResult(err)
		}

		out := skillPayload(sk)
		out["updated"] = true
		if sk.Tier == domain.SkillTeam {
			out["hint"] = "This skill is in the team library, so every research following it now reads the new text."
		}
		return successResult(out)
	})
}

// resolveSkillArg turns whichever address the caller gave into one skill.
//
// Four tools share it — skill_load, skill_update, skill_delete — because they
// share the problem: an agent working inside a research holds slugs and nothing
// else, while one that has just written a team skill holds the id it was handed
// back and has no research to look it up through.
//
// It returns a closure rather than an error so the two failure shapes stay
// distinct: a missing or contradictory argument is a validation error naming
// the fields, and everything else is the service's own refusal with its code.
//
// withBody decides which read is used. Load returns the body and Update
// inherits the fields it was not given, so both need it; Delete does not, and
// reading a 16000-character document to throw it away is the sort of cost that
// is invisible until a research holds forty of them.
func resolveSkillArg(
	ctx context.Context,
	skillSvc *service.SkillService,
	researchSvc *service.ResearchService,
	researchID, slug, skillID *string,
	withBody bool,
) (*domain.Skill, func() (*mcp.CallToolResult, any, error)) {
	hasID := skillID != nil && *skillID != ""
	hasSlug := slug != nil && *slug != "" && researchID != nil && *researchID != ""

	switch {
	case hasID && hasSlug:
		return nil, func() (*mcp.CallToolResult, any, error) {
			return validationErrorResult([]string{"give skill_id, or research_id with slug — not both"})
		}
	case !hasID && !hasSlug:
		return nil, func() (*mcp.CallToolResult, any, error) {
			return validationErrorResult([]string{"identify the skill: skill_id, or research_id together with slug"})
		}
	}

	if hasID {
		sk, err := skillSvc.Read(ctx, *skillID)
		if err != nil {
			// Deliberately not the slug form's invitation to go looking: by id,
			// "no such row" and "not yours" are one answer, and a hint would
			// separate them.
			return nil, func() (*mcp.CallToolResult, any, error) { return skillErrorResult(err) }
		}
		return sk, nil
	}

	rid, err := researchSvc.ResolveID(ctx, *researchID)
	if err != nil {
		return nil, func() (*mcp.CallToolResult, any, error) { return errorResult(err.Error()) }
	}
	var sk *domain.Skill
	if withBody {
		sk, err = skillSvc.Load(ctx, rid, *slug)
	} else {
		sk, err = skillSvc.ResolveSlug(ctx, rid, *slug)
	}
	if err != nil {
		return nil, func() (*mcp.CallToolResult, any, error) { return skillErrorForSlug(err, *slug) }
	}
	return sk, nil
}
