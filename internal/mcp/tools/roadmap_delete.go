package tools

import (
	"context"
	"log/slog"

	"github.com/dovod-app/app/internal/service"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type RoadmapDeleteInput struct {
	RoadmapID string `json:"roadmap_id" jsonschema:"ID of the roadmap to delete"`
}

func RegisterRoadmapDelete(srv *mcp.Server, svc *service.RoadmapService, log *slog.Logger) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "roadmap_delete",
		Description: "Deletes a roadmap and all its nodes and edges. This action is irreversible.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input RoadmapDeleteInput) (*mcp.CallToolResult, any, error) {
		if input.RoadmapID == "" {
			return validationErrorResult([]string{"roadmap_id is required"})
		}

		if err := svc.Delete(ctx, input.RoadmapID); err != nil {
			return errorResult(err.Error())
		}

		return successResult(map[string]any{
			"deleted": true,
		})
	})
}
