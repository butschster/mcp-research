package tools

import (
	"context"
	"log/slog"

	"github.com/butschster/mcp-research/internal/service"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type EntryReadInput struct {
	EntryID string `json:"entry_id" jsonschema:"ID of the entry to read"`
}

func RegisterEntryRead(srv *mcp.Server, svc *service.EntryService, log *slog.Logger) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "entry_read",
		Description: "Reads a single entry with full markdown content. Use this after entry_list to get the complete content of a specific entry.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input EntryReadInput) (*mcp.CallToolResult, any, error) {
		if input.EntryID == "" {
			return validationErrorResult([]string{"entry_id is required"})
		}

		entry, err := svc.Get(ctx, input.EntryID)
		if err != nil {
			return errorResult(err.Error())
		}

		return successResult(entry)
	})
}
