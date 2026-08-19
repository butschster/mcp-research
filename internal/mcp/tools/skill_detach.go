package tools

import (
	"context"
	"log/slog"

	"github.com/butschster/mcp-research/internal/domain"
	"github.com/butschster/mcp-research/internal/service"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type SkillDetachInput struct {
	ResearchID string `json:"research_id" jsonschema:"Research that should stop following the skill. Accepts the UUID or the short code (R1)"`
	Slug       string `json:"slug" jsonschema:"Slug of the skill to detach, from the following list in skill_list"`
}

func RegisterSkillDetach(srv *mcp.Server, skillSvc *service.SkillService, researchSvc *service.ResearchService, log *slog.Logger) {
	mcp.AddTool(srv, &mcp.Tool{
		Name: "skill_detach",
		Description: "Stops this research following a skill, freeing one of its six slots. " +
			"A research-private skill is deleted by this, because it exists nowhere else; a team or built-in one stays in the library. Product skills are always on and refuse with not_allowed.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input SkillDetachInput) (*mcp.CallToolResult, any, error) {
		var errs []string
		if input.ResearchID == "" {
			errs = append(errs, "research_id is required")
		}
		if input.Slug == "" {
			errs = append(errs, "slug is required")
		}
		if len(errs) > 0 {
			return validationErrorResult(errs)
		}

		researchID, err := researchSvc.ResolveID(ctx, input.ResearchID)
		if err != nil {
			return errorResult(err.Error())
		}
		// Read before detaching, so the answer can say what was actually lost.
		// Detach deletes a private skill outright, and "detached" is a
		// misleading word for that — an agent that dropped a research-private
		// skill to free a slot has destroyed the only copy of it.
		before, resolveErr := skillSvc.ResolveSlug(ctx, researchID, input.Slug)

		if err := skillSvc.Detach(ctx, researchID, input.Slug); err != nil {
			log.Error("skill_detach failed", "slug", input.Slug, "error", err)
			return skillErrorForSlug(err, input.Slug)
		}

		out := map[string]any{"slug": input.Slug, "detached": true}
		if resolveErr == nil && before != nil {
			out["tier"] = before.Tier
			out["name"] = before.Name
			if before.Tier == domain.SkillPrivate {
				out["deleted"] = true
				out["hint"] = "That skill belonged to this research alone, so detaching it deleted it. It is not recoverable from the library."
			}
		}
		return successResult(out)
	})
}
