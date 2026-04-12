package tools

import (
	"context"
	"log/slog"

	"github.com/butschster/mcp-research/internal/service"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type EntryDeleteInput struct {
	EntryID string `json:"entry_id" jsonschema:"ID of the entry to delete"`
}

func RegisterEntryDelete(srv *mcp.Server, svc *service.EntryService, log *slog.Logger) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "entry_delete",
		Description: "Deletes an entry from a research section. Use this to remove entries that are no longer relevant. Also cleans up associated cross-references and external links.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input EntryDeleteInput) (*mcp.CallToolResult, any, error) {
		if input.EntryID == "" {
			return validationErrorResult([]string{"entry_id is required"})
		}

		if err := svc.Delete(ctx, input.EntryID); err != nil {
			return errorResult(err.Error())
		}

		return successResult(map[string]any{
			"deleted": true,
		})
	})
}
