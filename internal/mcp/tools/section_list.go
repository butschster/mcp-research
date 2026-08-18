package tools

import (
	"context"
	"log/slog"

	"github.com/butschster/mcp-research/internal/service"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type SectionListInput struct {
	ResearchID string `json:"research_id" jsonschema:"ID of the research"`
}

func RegisterSectionList(srv *mcp.Server, svc *service.SectionService, log *slog.Logger) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "section_list",
		Description: "Lists all sections within a research project with entry counts per section. Returns sections ordered by position.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input SectionListInput) (*mcp.CallToolResult, any, error) {
		if input.ResearchID == "" {
			return validationErrorResult([]string{"research_id is required"})
		}

		sections, err := svc.List(ctx, input.ResearchID)
		if err != nil {
			return errorResult(err.Error())
		}

		var items []map[string]any
		for _, s := range sections {
			count, _ := svc.CountEntries(ctx, s.ID)
			item := map[string]any{
				"id":            s.ID,
				"name":          s.Name,
				"display_name":  s.DisplayName,
				"description":   s.Description,
				"status":        s.Status,
				"position":      s.Position,
				"entries_count": count,
				"spec_version":  s.SpecVersion,
			}
			// Only when the section declares something. Most sections are topics
			// rather than document classes and declare nothing, and an empty
			// field_spec on every one of them is noise in a payload the
			// conductor reads on every call.
			if len(s.FieldSpec) > 0 {
				item["field_spec"] = s.FieldSpec
			}

			items = append(items, item)
		}

		return successResult(map[string]any{
			"sections": items,
			"count":    len(items),
		})
	})
}
