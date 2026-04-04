package tools

import (
	"context"
	"log/slog"

	"github.com/butschster/mcp-research/internal/domain"
	"github.com/butschster/mcp-research/internal/service"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type TaskCreateInput struct {
	ResearchID  string `json:"research_id" jsonschema:"ID of the research"`
	Title       string `json:"title" jsonschema:"Task title"`
	Description string `json:"description" jsonschema:"Detailed description of what needs to be done"`
	Priority    string `json:"priority" jsonschema:"Priority: high, medium, low. Default: medium"`
}

func RegisterTaskCreate(srv *mcp.Server, svc *service.TaskService, log *slog.Logger) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "task_create",
		Description: "Creates a new task in the research todo list. Use this to plan and track work items during research. Returns task_id.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input TaskCreateInput) (*mcp.CallToolResult, any, error) {
		var errs []string
		if input.ResearchID == "" {
			errs = append(errs, "research_id is required")
		}
		if input.Title == "" {
			errs = append(errs, "title is required")
		}
		if len(errs) > 0 {
			return validationErrorResult(errs)
		}

		task, err := svc.Create(ctx, service.CreateTaskRequest{
			ResearchID:  input.ResearchID,
			Title:       input.Title,
			Description: input.Description,
			Priority:    domain.Priority(input.Priority),
		})
		if err != nil {
			return errorResult(err.Error())
		}

		return successResult(map[string]any{
			"task_id":  task.ID,
			"title":    task.Title,
			"status":   task.Status,
			"priority": task.Priority,
		})
	})
}
