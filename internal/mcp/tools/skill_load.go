package tools

import (
	"context"
	"log/slog"

	"github.com/butschster/mcp-research/internal/service"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type SkillLoadInput struct {
	ResearchID string `json:"research_id" jsonschema:"Research whose skills you are working under. Accepts the UUID or the short code (R1)"`
	// One slug, never a list. There is deliberately no batch form: loading
	// every skill at once is exactly what this design exists to prevent, and an
	// array parameter is an invitation to do it.
	Slug string `json:"slug" jsonschema:"Slug of the skill to open, from the skills index in research_get"`
}

func RegisterSkillLoad(srv *mcp.Server, skillSvc *service.SkillService, researchSvc *service.ResearchService, log *slog.Logger) {
	mcp.AddTool(srv, &mcp.Tool{
		Name: "skill_load",
		Description: "Opens one skill and returns its full text. " +
			"research_get lists the skills attached to a research with a name and a line saying when to use each; call this when you are about to do the work one of them names, not while orienting. One skill at a time.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input SkillLoadInput) (*mcp.CallToolResult, any, error) {
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

		// The index the agent read this slug out of came from research_get,
		// which accepts a short code — so an agent that passes `R7` straight
		// back in would otherwise be told the research does not exist.
		researchID, err := researchSvc.ResolveID(ctx, input.ResearchID)
		if err != nil {
			return errorResult(err.Error())
		}

		sk, err := skillSvc.Load(ctx, researchID, input.Slug)
		if err != nil {
			log.Error("skill_load failed", "error", err)
			return errorResult(err.Error())
		}

		return successResult(map[string]any{
			"slug":        sk.Slug,
			"name":        sk.Name,
			"tier":        sk.Tier,
			"description": sk.Description,
			"body":        sk.Body,
			// Version and updated_at let an agent holding an older body know it
			// has moved on. Both are stored and bumped on every edit; without
			// them a second skill_load was the only way to find out.
			"version":    sk.Version,
			"updated_at": sk.UpdatedAt,
			// Precedence is expressed by tier and nothing else, so it is
			// restated on every load rather than left to the agent to remember
			// from a document it may not have read this session.
			"precedence": "A skill attached to this research directly overrides a team skill, which overrides a built-in. Where two skills conflict, follow the higher one.",
		})
	})
}
