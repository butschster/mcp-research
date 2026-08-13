package tools

import (
	"context"
	"log/slog"

	"github.com/butschster/mcp-research/internal/domain"
	"github.com/butschster/mcp-research/internal/service"
	"github.com/butschster/mcp-research/internal/storage"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type TaskListInput struct {
	ResearchID string  `json:"research_id" jsonschema:"ID of the research"`
	Status     *string `json:"status" jsonschema:"Filter by status: pending, in_progress, blocked, completed, failed, deferred"`
	Priority   *string `json:"priority" jsonschema:"Filter by priority: high, medium, low"`
}

func RegisterTaskList(srv *mcp.Server, svc *service.TaskService, log *slog.Logger) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "task_list",
		Description: "Lists tasks in the research todo list. Returns tasks ordered by priority (high first), then by creation date. Use filters to find specific tasks.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input TaskListInput) (*mcp.CallToolResult, any, error) {
		if input.ResearchID == "" {
			return validationErrorResult([]string{"research_id is required"})
		}

		filter := storage.TaskFilter{}
		if v := derefStr(input.Status); v != "" {
			s := domain.TaskStatus(v)
			filter.Status = &s
		}
		if v := derefStr(input.Priority); v != "" {
			p := domain.Priority(v)
			filter.Priority = &p
		}

		tasks, err := svc.List(ctx, input.ResearchID, filter)
		if err != nil {
			return errorResult(err.Error())
		}

		counts, _ := svc.CountByStatus(ctx, input.ResearchID)
		total := 0
		for _, c := range counts {
			total += c
		}

		var items []map[string]any
		for _, t := range tasks {
			item := map[string]any{
				"id":          t.ID,
				"title":       t.Title,
				"description": t.Description,
				"status":      t.Status,
				"priority":    t.Priority,
			}
			if t.Result != "" {
				item["result"] = t.Result
			}
			items = append(items, item)
		}

		return successResult(map[string]any{
			"tasks": items,
			"count": len(items),
			"progress": map[string]any{
				"total":       total,
				"pending":     counts[domain.TaskPending] + counts[domain.TaskBlocked],
				"in_progress": counts[domain.TaskInProgress],
				"completed":   counts[domain.TaskCompleted],
				"failed":      counts[domain.TaskFailed],
				"deferred":    counts[domain.TaskDeferred],
			},
		})
	})
}
