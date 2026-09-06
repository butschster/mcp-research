package tools

import (
	"context"
	"log/slog"

	"github.com/dovod-app/app/internal/domain"
	"github.com/dovod-app/app/internal/service"
	"github.com/dovod-app/app/internal/storage"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type QuestionListInput struct {
	SessionID string  `json:"session_id" jsonschema:"ID of the session"`
	Status    *string `json:"status" jsonschema:"Filter by status: pending, in_progress, answered, deferred, skipped"`
	Area      *string `json:"area" jsonschema:"Filter by topic area"`
	Priority  *string `json:"priority" jsonschema:"Filter by priority: high, medium, low"`
}

func RegisterQuestionList(srv *mcp.Server, svc *service.SessionService, log *slog.Logger) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "question_list",
		Description: "Lists questions within a session with optional filters. Returns questions ordered by position.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input QuestionListInput) (*mcp.CallToolResult, any, error) {
		if input.SessionID == "" {
			return validationErrorResult([]string{"session_id is required"})
		}

		filter := storage.QuestionFilter{}
		if v := derefStr(input.Status); v != "" {
			s := domain.QuestionStatus(v)
			filter.Status = &s
		}
		if v := derefStr(input.Area); v != "" {
			filter.Area = &v
		}
		if v := derefStr(input.Priority); v != "" {
			p := domain.Priority(v)
			filter.Priority = &p
		}

		questions, err := svc.ListQuestions(ctx, input.SessionID, filter)
		if err != nil {
			return errorResult(err.Error())
		}

		var items []map[string]any
		for _, q := range questions {
			items = append(items, map[string]any{
				"id":        q.ID,
				"text":      q.Text,
				"area":      q.Area,
				"priority":  q.Priority,
				"status":    q.Status,
				"answer":    q.Answer,
				"parent_id": q.ParentID,
				"position":  q.Position,
			})
		}

		return successResult(map[string]any{
			"questions": items,
			"count":     len(items),
		})
	})
}
