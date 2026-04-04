package tools

import (
	"context"
	"log/slog"

	"github.com/butschster/mcp-research/internal/service"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type SessionGetInput struct {
	SessionID string `json:"session_id" jsonschema:"ID of the session to retrieve"`
}

func RegisterSessionGet(srv *mcp.Server, svc *service.SessionService, log *slog.Logger) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "session_get",
		Description: "Returns full session context with questions grouped by status and progress counters. Use this to understand the current state of a research session.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input SessionGetInput) (*mcp.CallToolResult, any, error) {
		if input.SessionID == "" {
			return validationErrorResult([]string{"session_id is required"})
		}

		result, err := svc.Get(ctx, input.SessionID)
		if err != nil {
			return errorResult(err.Error())
		}

		// Group questions by status
		grouped := make(map[string][]map[string]any)
		for _, q := range result.Questions {
			grouped[string(q.Status)] = append(grouped[string(q.Status)], map[string]any{
				"id":        q.ID,
				"text":      q.Text,
				"area":      q.Area,
				"priority":  q.Priority,
				"answer":    q.Answer,
				"parent_id": q.ParentID,
				"position":  q.Position,
			})
		}

		return successResult(map[string]any{
			"session":   result.Session,
			"questions": grouped,
			"progress":  result.Progress,
		})
	})
}
