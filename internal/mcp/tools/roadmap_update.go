package tools

import (
	"context"
	"log/slog"

	"github.com/butschster/mcp-research/internal/domain"
	"github.com/butschster/mcp-research/internal/service"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type RoadmapUpdateInput struct {
	RoadmapID   string   `json:"roadmap_id" jsonschema:"ID of the roadmap"`
	Title       *string  `json:"title" jsonschema:"New title"`
	Description *string  `json:"description" jsonschema:"New description"`
	Statuses    []string `json:"statuses" jsonschema:"New list of available node statuses"`
	Status      *string  `json:"status" jsonschema:"Roadmap status: active, completed, archived"`
}

func RegisterRoadmapUpdate(srv *mcp.Server, svc *service.RoadmapService, log *slog.Logger) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "roadmap_update",
		Description: "Updates roadmap metadata (title, description, statuses, status).",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input RoadmapUpdateInput) (*mcp.CallToolResult, any, error) {
		if input.RoadmapID == "" {
			return validationErrorResult([]string{"roadmap_id is required"})
		}

		updateReq := service.UpdateRoadmapRequest{
			Title:       input.Title,
			Description: input.Description,
			Statuses:    input.Statuses,
		}
		if input.Status != nil {
			s := domain.RoadmapStatus(*input.Status)
			updateReq.Status = &s
		}

		rm, err := svc.Update(ctx, input.RoadmapID, updateReq)
		if err != nil {
			return errorResult(err.Error())
		}

		return successResult(map[string]any{
			"roadmap_id":  rm.ID,
			"title":       rm.Title,
			"description": rm.Description,
			"statuses":    rm.Statuses,
			"status":      rm.Status,
		})
	})
}
