package tools

import (
	"context"
	"log/slog"

	"github.com/butschster/mcp-research/internal/service"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type TemplateListInput struct{}

func RegisterTemplateList(srv *mcp.Server, templateSvc *service.TemplateService, log *slog.Logger) {
	mcp.AddTool(srv, &mcp.Tool{
		Name: "template_list",
		Description: "Lists the research methodologies available — the ones that ship with the app and any your teams have written. " +
			"Each says when it is worth using and when it is not. Call this at the start of a new research, after you know what decision the user is trying to make, and read the one that fits with template_get before proposing any structure.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input TemplateListInput) (*mcp.CallToolResult, any, error) {
		list, err := templateSvc.List(ctx)
		if err != nil {
			log.Error("template_list failed", "error", err)
			return errorResult(err.Error())
		}

		// Matching criteria only, never a body. Twenty-five methodologies arriving in
		// a kickoff that needs one is the cost this whole feature exists to
		// avoid.
		//
		// `source` is not sent either. It matters to a human deciding whether to
		// trust or edit a methodology, and not at all to the model choosing one:
		// what it matches on is the two criteria lines, and who wrote the text
		// does not change whether the text fits.
		out := make([]map[string]any, 0, len(list))
		for _, tp := range list {
			out = append(out, map[string]any{
				"slug":            tp.Slug,
				"name":            tp.Name,
				"tier":            tp.Tier,
				"description":     tp.Description,
				"when_to_use":     tp.WhenToUse,
				"when_not_to_use": tp.WhenNotToUse,
			})
		}
		return successResult(map[string]any{
			"templates": out,
			"usage_hint": "Ask what decision is waiting on this research before you offer any of these. " +
				"A structure shown before the user has said what they are deciding anchors them to it. " +
				"When one fits, call template_get and follow its text; when none does, design the research yourself and say so.",
		})
	})
}
