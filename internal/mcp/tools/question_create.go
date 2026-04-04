package tools

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/butschster/mcp-research/internal/domain"
	"github.com/butschster/mcp-research/internal/service"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type QuestionCreateInput struct {
	SessionID string                   `json:"session_id" jsonschema:"ID of the session"`
	Questions []QuestionCreateSpecInput `json:"questions" jsonschema:"Questions to create"`
}

func RegisterQuestionCreate(srv *mcp.Server, svc *service.SessionService, log *slog.Logger) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "question_create",
		Description: "Creates a batch of questions within a session. Supports parent_id for follow-up questions (max 3 levels deep). Returns created question IDs.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input QuestionCreateInput) (*mcp.CallToolResult, any, error) {
		var errs []string
		if input.SessionID == "" {
			errs = append(errs, "session_id is required")
		}
		if len(input.Questions) == 0 {
			errs = append(errs, "at least one question is required")
		}
		for i, q := range input.Questions {
			if q.Text == "" {
				errs = append(errs, fmt.Sprintf("questions[%d].text is required", i))
			}
		}
		if len(errs) > 0 {
			return validationErrorResult(errs)
		}

		var requests []service.CreateQuestionRequest
		for _, q := range input.Questions {
			priority := domain.Priority(q.Priority)
			if priority == "" {
				priority = domain.PriorityMedium
			}
			requests = append(requests, service.CreateQuestionRequest{
				Text:      q.Text,
				Area:      q.Area,
				Rationale: q.Rationale,
				Priority:  priority,
				ParentID:  q.ParentID,
				Position:  q.Position,
			})
		}

		questions, err := svc.AddQuestions(ctx, input.SessionID, requests)
		if err != nil {
			return errorResult(err.Error())
		}

		var ids []string
		for _, q := range questions {
			ids = append(ids, q.ID)
		}

		return successResult(map[string]any{
			"question_ids": ids,
			"count":        len(ids),
		})
	})
}
