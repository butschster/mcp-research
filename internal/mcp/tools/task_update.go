package tools

import (
	"context"
	"log/slog"

	"github.com/butschster/mcp-research/internal/domain"
	"github.com/butschster/mcp-research/internal/service"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type TaskUpdateInput struct {
	TaskID      string  `json:"task_id" jsonschema:"ID of the task to update"`
	Title       *string `json:"title" jsonschema:"New title"`
	Description *string `json:"description" jsonschema:"New description"`
	Status      *string `json:"status" jsonschema:"New status: pending, in_progress, blocked, completed, failed, deferred"`
	Priority    *string `json:"priority" jsonschema:"New priority: high, medium, low"`
	Result      *string `json:"result" jsonschema:"Result or outcome of the task (set when completing)"`
}

func RegisterTaskUpdate(srv *mcp.Server, svc *service.TaskService, log *slog.Logger) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "task_update",
		Description: "Updates a task in the research todo list. Use status to track progress. Set result when completing a task to record what was accomplished.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input TaskUpdateInput) (*mcp.CallToolResult, any, error) {
		if input.TaskID == "" {
			return validationErrorResult([]string{"task_id is required"})
		}

		var status *domain.TaskStatus
		if input.Status != nil {
			s := domain.TaskStatus(*input.Status)
			status = &s
		}

		var priority *domain.Priority
		if input.Priority != nil {
			p := domain.Priority(*input.Priority)
			priority = &p
		}

		task, err := svc.Update(ctx, input.TaskID, service.UpdateTaskRequest{
			Title:       input.Title,
			Description: input.Description,
			Status:      status,
			Priority:    priority,
			Result:      input.Result,
		})
		if err != nil {
			return errorResult(err.Error())
		}

		return successResult(map[string]any{
			"task_id":  task.ID,
			"title":    task.Title,
			"status":   task.Status,
			"priority": task.Priority,
			"updated":  true,
		})
	})
}
