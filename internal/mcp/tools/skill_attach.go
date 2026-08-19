package tools

import (
	"context"
	"log/slog"

	"github.com/butschster/mcp-research/internal/service"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type SkillAttachInput struct {
	ResearchID string `json:"research_id" jsonschema:"Research that should follow the skill. Accepts the UUID or the short code (R1)"`
	Slug       string `json:"slug" jsonschema:"Slug of the skill to attach, from the available list in skill_list"`
}

func RegisterSkillAttach(srv *mcp.Server, skillSvc *service.SkillService, researchSvc *service.ResearchService, log *slog.Logger) {
	mcp.AddTool(srv, &mcp.Tool{
		Name: "skill_attach",
		Description: "Makes this research follow an existing skill, so its trigger line is in the index from now on. " +
			"Six chosen skills per research; the seventh is refused with skill_cap_reached. Product skills are always on and need no attaching.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input SkillAttachInput) (*mcp.CallToolResult, any, error) {
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
		// viaTemplate is false and cannot be set from here. The flag records
		// that a methodology chose a skill at creation, and a tool that let the
		// caller claim it would make the one distinction the attached list draws
		// — a template's choice against a person's — unreliable.
		sk, err := skillSvc.Attach(ctx, researchID, input.Slug, false)
		if err != nil {
			log.Error("skill_attach failed", "slug", input.Slug, "error", err)
			return skillErrorForSlug(err, input.Slug)
		}

		out := skillPayload(sk)
		out["attached"] = true
		out["hint"] = "Its trigger line is in the index from the next research_get. Read the body with skill_load when you reach the work it names."
		return successResult(out)
	})
}
