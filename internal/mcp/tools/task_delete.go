package tools

import (
	"context"
	"log/slog"

	"github.com/dovod-app/app/internal/service"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type TaskDeleteInput struct {
	TaskID string `json:"task_id" jsonschema:"ID of the task to delete"`
}

func RegisterTaskDelete(srv *mcp.Server, svc *service.TaskService, log *slog.Logger) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "task_delete",
		Description: "Deletes a task from the research todo list. Use this to remove tasks that are no longer relevant.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input TaskDeleteInput) (*mcp.CallToolResult, any, error) {
		if input.TaskID == "" {
			return validationErrorResult([]string{"task_id is required"})
		}

		if err := svc.Delete(ctx, input.TaskID); err != nil {
			return errorResult(err.Error())
		}

		return successResult(map[string]any{
			"deleted": true,
		})
	})
}
