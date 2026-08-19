package tools

import (
	"context"
	"log/slog"

	"github.com/butschster/mcp-research/internal/domain"
	"github.com/butschster/mcp-research/internal/service"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type SkillDeleteInput struct {
	ResearchID *string `json:"research_id,omitempty" jsonschema:"Research to resolve the slug against. Accepts the UUID or the short code (R1). Give this with slug, or give skill_id"`
	Slug       *string `json:"slug,omitempty" jsonschema:"Slug of the skill to delete, as it appears in the skills index"`
	SkillID    *string `json:"skill_id,omitempty" jsonschema:"Id of the skill to delete. An alternative to research_id plus slug"`
}

// RegisterSkillDelete removes a skill from existence, as opposed to
// skill_detach, which removes it from one research.
//
// The distinction is the whole reason both tools exist. For a research-private
// skill they are the same act, because there is nowhere else for it to live.
// For a team skill they are not: detaching leaves it in the library for the
// next research, deleting takes it away from everyone.
func RegisterSkillDelete(srv *mcp.Server, skillSvc *service.SkillService, researchSvc *service.ResearchService, log *slog.Logger) {
	mcp.AddTool(srv, &mcp.Tool{
		Name: "skill_delete",
		Description: "Deletes a team or research-private skill outright and unrecoverably — skills have no revision history and nothing here restores one. " +
			"Use skill_detach instead to stop one research following a skill others still use, or skill_promote to lift a research-private skill into the team library rather than destroying it. " +
			"A team skill any research still follows is refused with skill_in_use; built-ins cannot be deleted.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input SkillDeleteInput) (*mcp.CallToolResult, any, error) {
		current, fail := resolveSkillArg(ctx, skillSvc, researchSvc, input.ResearchID, input.Slug, input.SkillID, false)
		if fail != nil {
			return fail()
		}
		if current.Tier == domain.SkillBuiltin {
			return errorResult("not_allowed: built-in skills ship with the server and cannot be deleted. Use skill_detach to stop a research following one.")
		}

		if err := skillSvc.Delete(ctx, current.ID); err != nil {
			log.Error("skill_delete failed", "skill_id", current.ID, "error", err)
			return skillErrorResult(err)
		}

		return successResult(map[string]any{
			"skill_id": current.ID,
			"slug":     current.Slug,
			"name":     current.Name,
			"tier":     current.Tier,
			"deleted":  true,
			// No hint. The warning that mattered is in the tool description,
			// where it can still change a decision; here it would be a sentence
			// about something already irreversible.
		})
	})
}
