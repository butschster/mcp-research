package tools

import (
	"context"
	"log/slog"

	"github.com/butschster/mcp-research/internal/service"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type TemplateGetInput struct {
	Slug string `json:"slug" jsonschema:"Slug of the template to read, from template_list"`
}

func RegisterTemplateGet(srv *mcp.Server, templateSvc *service.TemplateService, log *slog.Logger) {
	mcp.AddTool(srv, &mcp.Tool{
		Name: "template_get",
		Description: "Returns one research methodology in full: what to ask the user before proposing anything, what structure to suggest, what a good entry looks like, and when the research is finished. " +
			"Follow it — it is instructions, not a form. Pass its slug to research_create so the research records which methodology it was started from.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input TemplateGetInput) (*mcp.CallToolResult, any, error) {
		if input.Slug == "" {
			return validationErrorResult([]string{"slug is required"})
		}
		tp, err := templateSvc.Get(ctx, input.Slug)
		if err != nil {
			log.Error("template_get failed", "error", err)
			return errorResult(err.Error())
		}
		return successResult(map[string]any{
			"slug":            tp.Slug,
			"name":            tp.Name,
			"tier":            tp.Tier,
			"when_to_use":     tp.WhenToUse,
			"when_not_to_use": tp.WhenNotToUse,
			"body":            tp.Body,
			// Resolved, not raw slugs: an agent choosing a methodology has to be
			// able to see that part of it is unbacked, and `missing` was
			// visible over REST and invisible here.
			"skills":  tp.SkillsResolved,
			"version": tp.Version,
			"usage_hint": "The structure below is a proposal, not a schema. Adapt it to what the user told you, " +
				"and create a section when you have something to put in it — an empty section is a standing " +
				"instruction to invent content for it. Pass template_slug to research_create when you create the research.",
		})
	})
}
