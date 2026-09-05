package tools

import (
	"context"
	"log/slog"

	"github.com/dovod-app/app/internal/domain"
	"github.com/dovod-app/app/internal/service"
	"github.com/dovod-app/app/internal/storage"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type EntryListInput struct {
	ResearchID string  `json:"research_id" jsonschema:"ID of the research"`
	SectionID  string  `json:"section_id" jsonschema:"ID of the section"`
	Status     *string `json:"status" jsonschema:"Filter by status: draft, active, completed, archived"`
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
		if v := derefStr(input.Status); v != "" {
			s := domain.EntryStatus(v)
			filter.Status = &s
		}

		entries, err := svc.List(ctx, input.ResearchID, input.SectionID, filter)
		if err != nil {
			return errorResult(err.Error())
		}

		var items []map[string]any
		for _, e := range entries {
			item := map[string]any{
				"id":          e.ID,
				"code":        e.Code,
				"title":       e.Title,
				"description": e.Description,
				"status":      e.Status,
				"tags":        e.Tags,
				"created_at":  e.CreatedAt,
			}
			// Content is left out of this listing on purpose; metadata is not
			// content. Without it "which specifications are still in review" cost
			// one entry_read per document, which is the question this feature was
			// built to answer.
			if len(e.Metadata) > 0 {
				item["metadata"] = e.Metadata
			}
			if e.MetaStatus != nil && !e.MetaStatus.Complete {
				item["metadata_missing"] = e.MetaStatus.MissingRequired
			}
			items = append(items, item)
		}

		return successResult(map[string]any{
			"entries": items,
			"count":   len(items),
			"hint":    "Use entry_read with entry_id to get full content.",
		})
	})
}
