package tools

import (
	"context"
	"log/slog"

	"github.com/dovod-app/app/internal/domain"
	"github.com/dovod-app/app/internal/service"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type QuestionUpdateInput struct {
	QuestionID string  `json:"question_id" jsonschema:"ID of the question to update"`
	Status     *string `json:"status" jsonschema:"New status: pending, in_progress, answered, deferred, skipped. Note: answered requires non-empty answer."`
	Answer     *string `json:"answer" jsonschema:"Answer text (required when status is answered)"`
}

func RegisterQuestionUpdate(srv *mcp.Server, svc *service.SessionService, log *slog.Logger) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "question_update",
		Description: "Updates a question's status and answer. Setting status to 'answered' requires a non-empty answer. Returns the updated question.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input QuestionUpdateInput) (*mcp.CallToolResult, any, error) {
		if input.QuestionID == "" {
			return validationErrorResult([]string{"question_id is required"})
		}

		var status *domain.QuestionStatus
		if input.Status != nil {
			s := domain.QuestionStatus(*input.Status)
			status = &s
		}

		question, err := svc.UpdateQuestion(ctx, input.QuestionID, status, input.Answer)
		if err != nil {
			return errorResult(err.Error())
		}

		return successResult(map[string]any{
			"question_id": question.ID,
			"status":      question.Status,
			"answer":      question.Answer,
			"updated":     true,
		})
	})
}
