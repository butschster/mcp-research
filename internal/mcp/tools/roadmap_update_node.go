package tools

import (
	"context"
	"log/slog"

	"github.com/butschster/mcp-research/internal/service"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type RoadmapUpdateNodeInput struct {
	NodeID      string   `json:"node_id" jsonschema:"ID of the node to update"`
	Title       *string  `json:"title" jsonschema:"New title"`
	Description *string  `json:"description" jsonschema:"New description text"`
	NodeType    *string  `json:"node_type" jsonschema:"New node type: step, milestone, decision, info, group"`
	Status      *string  `json:"status" jsonschema:"New status (must be from the roadmap's statuses list)"`
	PositionX   *float64 `json:"position_x" jsonschema:"New X position"`
	PositionY   *float64 `json:"position_y" jsonschema:"New Y position"`
	ParentID    *string  `json:"parent_id" jsonschema:"New parent node ID (empty string to clear)"`
}

func RegisterRoadmapUpdateNode(srv *mcp.Server, svc *service.RoadmapService, log *slog.Logger) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "roadmap_update_node",
		Description: "Updates a single roadmap node. Use this to change node status, content, type, or position.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input RoadmapUpdateNodeInput) (*mcp.CallToolResult, any, error) {
		if input.NodeID == "" {
			return validationErrorResult([]string{"node_id is required"})
		}

		node, err := svc.UpdateNode(ctx, input.NodeID, service.UpdateRoadmapNodeRequest{
			Title:       input.Title,
			Description: input.Description,
			NodeType:    input.NodeType,
			Status:      input.Status,
			PositionX:   input.PositionX,
			PositionY:   input.PositionY,
			ParentID:    input.ParentID,
		})
		if err != nil {
			return errorResult(err.Error())
		}

		return successResult(map[string]any{
			"node_id":   node.ID,
			"code":      node.Code,
			"title":     node.Title,
			"node_type": node.NodeType,
			"status":    node.Status,
		})
	})
}
