package tools

import (
	"context"
	"github.com/butschster/mcp-research/internal/service"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type ResearchMemoryInput struct {
	ResearchID string   `json:"research_id" jsonschema:"Research UUID or R code"`
	Action     string   `json:"action" jsonschema:"One of list, add, update, delete"`
	Text       string   `json:"text,omitempty" jsonschema:"Note text for add or update"`
	ItemID     string   `json:"item_id,omitempty" jsonschema:"Memory item ID for update"`
	Version    int      `json:"version,omitempty" jsonschema:"Current item version from research_get or list; required for update"`
	IDs        []string `json:"ids,omitempty" jsonschema:"Explicit memory item IDs to delete (1–500); never the entire array"`
	SessionID  string   `json:"session_id,omitempty" jsonschema:"Research session UUID or SS code for a new note's provenance"`
}

func RegisterResearchMemory(srv *mcp.Server, svc *service.ResearchService) {
	mcp.AddTool(srv, &mcp.Tool{Name: "research_memory", Description: "Manage research memory one item at a time. List returns IDs, provenance and versions. Updating an outdated version fails rather than overwriting someone else's edit. Delete removes only the selected IDs."},
		func(ctx context.Context, req *mcp.CallToolRequest, input ResearchMemoryInput) (*mcp.CallToolResult, any, error) {
			var err error
			switch input.Action {
			case "list":
				items, e := svc.ListMemory(ctx, input.ResearchID)
				if e != nil {
					return errorResult(e.Error())
				}
				return successResult(items)
			case "add":
				item, e := svc.AddMemory(ctx, input.ResearchID, input.Text, input.SessionID)
				if e != nil {
					return errorResult(e.Error())
				}
				return successResult(item)
			case "update":
				err = svc.UpdateMemory(ctx, input.ResearchID, input.ItemID, input.Text, input.Version)
			case "delete":
				err = svc.DeleteMemory(ctx, input.ResearchID, input.IDs)
			default:
				return validationErrorResult([]string{"action must be list, add, update, or delete"})
			}
			if err != nil {
				return errorResult(err.Error())
			}
			return successResult(map[string]any{"updated": true})
		})
}
