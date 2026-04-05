package tools

import (
	"context"
	"log/slog"

	"github.com/butschster/mcp-research/internal/domain"
	"github.com/butschster/mcp-research/internal/service"
	"github.com/butschster/mcp-research/internal/storage"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type EntryListInput struct {
	ResearchID string `json:"research_id" jsonschema:"ID of the research"`
	SectionID  string `json:"section_id" jsonschema:"ID of the section"`
	Status     string `json:"status" jsonschema:"Filter by status: draft, active, completed, archived"`
}

func RegisterEntryList(srv *mcp.Server, svc *service.EntryService, log *slog.Logger) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "entry_list",
		Description: "Lists entries within a section. Returns entries WITHOUT content for token efficiency. Use entry_read to get full content of a specific entry.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input EntryListInput) (*mcp.CallToolResult, any, error) {
		if input.ResearchID == "" || input.SectionID == "" {
			return validationErrorResult([]string{"research_id and section_id are required"})
		}

		filter := storage.EntryFilter{}
		if input.Status != "" {
			s := domain.EntryStatus(input.Status)
			filter.Status = &s
		}

		entries, err := svc.List(ctx, input.ResearchID, input.SectionID, filter)
		if err != nil {
			return errorResult(err.Error())
		}

		var items []map[string]any
		for _, e := range entries {
			items = append(items, map[string]any{
				"id":          e.ID,
				"code":        e.Code,
				"title":       e.Title,
				"description": e.Description,
				"status":      e.Status,
				"tags":        e.Tags,
				"created_at":  e.CreatedAt,
			})
		}

		return successResult(map[string]any{
			"entries": items,
			"count":   len(items),
			"hint":    "Use entry_read with entry_id to get full content.",
		})
	})
}
