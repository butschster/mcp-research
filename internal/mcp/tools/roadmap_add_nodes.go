package tools

import (
	"context"
	"log/slog"

	"github.com/dovod-app/app/internal/service"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type RoadmapAddNodesInput struct {
	RoadmapID string              `json:"roadmap_id" jsonschema:"ID of the roadmap to extend"`
	Nodes     []RoadmapCreateNode `json:"nodes" jsonschema:"New nodes to add"`
	Edges     []RoadmapCreateEdge `json:"edges" jsonschema:"New edges to add (can reference temp_ids of new nodes or IDs of existing nodes)"`
}

func RegisterRoadmapAddNodes(srv *mcp.Server, svc *service.RoadmapService, log *slog.Logger) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "roadmap_add_nodes",
		Description: "Adds new nodes and edges to an existing roadmap. Use temp_id on new nodes to reference them in new edges. Edges can also connect to existing node IDs.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input RoadmapAddNodesInput) (*mcp.CallToolResult, any, error) {
		if input.RoadmapID == "" {
			return validationErrorResult([]string{"roadmap_id is required"})
		}
		if len(input.Nodes) == 0 && len(input.Edges) == 0 {
			return validationErrorResult([]string{"at least one node or edge is required"})
		}

		var nodes []service.CreateRoadmapNodeRequest
		for _, n := range input.Nodes {
			nodes = append(nodes, nodeFromInput(n))
		}

		var edges []service.CreateRoadmapEdgeRequest
		for _, e := range input.Edges {
			edges = append(edges, edgeFromInput(e))
		}

		rm, err := svc.AddNodes(ctx, input.RoadmapID, nodes, edges)
		if err != nil {
			return errorResult(err.Error())
		}

		return successResult(map[string]any{
			"roadmap_id":  rm.ID,
			"total_nodes": len(rm.Nodes),
			"total_edges": len(rm.Edges),
		})
	})
}
