package tools

import (
	"context"
	"log/slog"

	"github.com/dovod-app/app/internal/service"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type RoadmapListInput struct {
	ResearchID string `json:"research_id" jsonschema:"ID of the research"`
}

func RegisterRoadmapList(srv *mcp.Server, svc *service.RoadmapService, log *slog.Logger) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "roadmap_list",
		Description: "Lists all roadmaps for a research. Returns metadata without nodes/edges — use roadmap_get for full graph.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input RoadmapListInput) (*mcp.CallToolResult, any, error) {
		if input.ResearchID == "" {
			return validationErrorResult([]string{"research_id is required"})
		}

		roadmaps, err := svc.List(ctx, input.ResearchID)
		if err != nil {
			return errorResult(err.Error())
		}

		items := make([]map[string]any, 0, len(roadmaps))
		for _, rm := range roadmaps {
			items = append(items, map[string]any{
				"roadmap_id":  rm.ID,
				"code":        rm.Code,
				"title":       rm.Title,
				"description": rm.Description,
				"status":      rm.Status,
				"statuses":    rm.Statuses,
			})
		}

		return successResult(map[string]any{
			"research_id": input.ResearchID,
			"count":       len(items),
			"roadmaps":    items,
		})
	})
}
