package tools

import (
	"context"
	"log/slog"

	"github.com/butschster/mcp-research/internal/service"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type RoadmapGetInput struct {
	RoadmapID string `json:"roadmap_id" jsonschema:"ID of the roadmap"`
}

func RegisterRoadmapGet(srv *mcp.Server, svc *service.RoadmapService, log *slog.Logger) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "roadmap_get",
		Description: "Returns a roadmap with all its nodes and edges. Use this to read the full graph structure.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input RoadmapGetInput) (*mcp.CallToolResult, any, error) {
		if input.RoadmapID == "" {
			return validationErrorResult([]string{"roadmap_id is required"})
		}

		rm, err := svc.Get(ctx, input.RoadmapID)
		if err != nil {
			return errorResult(err.Error())
		}

		nodes := make([]map[string]any, 0, len(rm.Nodes))
		for _, n := range rm.Nodes {
			node := map[string]any{
				"node_id":     n.ID,
				"code":        n.Code,
				"title":       n.Title,
				"description": n.Description,
				"node_type":   n.NodeType,
				"status":      n.Status,
				"position_x":  n.PositionX,
				"position_y":  n.PositionY,
			}
			if n.ParentID != "" {
				node["parent_id"] = n.ParentID
			}
			nodes = append(nodes, node)
		}

		edges := make([]map[string]any, 0, len(rm.Edges))
		for _, e := range rm.Edges {
			edge := map[string]any{
				"edge_id":   e.ID,
				"source":    e.SourceNodeID,
				"target":    e.TargetNodeID,
				"label":     e.Label,
				"edge_type": e.EdgeType,
			}
			edges = append(edges, edge)
		}

		return successResult(map[string]any{
			"roadmap_id":  rm.ID,
			"code":        rm.Code,
			"research_id": rm.ResearchID,
			"title":       rm.Title,
			"description": rm.Description,
			"statuses":    rm.Statuses,
			"status":      rm.Status,
			"nodes":       nodes,
			"edges":       edges,
		})
	})
}
