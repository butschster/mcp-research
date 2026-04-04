package tools

import (
	"context"
	"log/slog"

	"github.com/butschster/mcp-research/internal/domain"
	"github.com/butschster/mcp-research/internal/service"
	"github.com/butschster/mcp-research/internal/storage"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type ResearchListInput struct {
	Status string `json:"status" jsonschema:"Filter by status: active, completed, or archived. Leave empty for all."`
}

func RegisterResearchList(srv *mcp.Server, svc *service.ResearchService, log *slog.Logger) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "research_list",
		Description: "Lists all research projects with optional status filter. Returns id, name, goal, status, and tags for each research.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input ResearchListInput) (*mcp.CallToolResult, any, error) {
		filter := storage.ResearchFilter{}
		if input.Status != "" {
			status := domain.ResearchStatus(input.Status)
			filter.Status = &status
		}

		researches, err := svc.List(ctx, filter)
		if err != nil {
			return errorResult(err.Error())
		}

		var items []map[string]any
		for _, r := range researches {
			items = append(items, map[string]any{
				"id":     r.ID,
				"name":   r.Name,
				"goal":   r.Goal,
				"status": r.Status,
				"tags":   r.Tags,
			})
		}

		return successResult(map[string]any{
			"researches": items,
			"count":      len(items),
		})
	})
}
