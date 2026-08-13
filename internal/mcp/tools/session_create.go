package tools

import (
	"context"
	"log/slog"

	"github.com/butschster/mcp-research/internal/domain"
	"github.com/butschster/mcp-research/internal/service"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type SessionCreateInput struct {
	ResearchID string                    `json:"research_id" jsonschema:"ID of the research"`
	Title      string                    `json:"title" jsonschema:"Session title"`
	Focus      *string                   `json:"focus" jsonschema:"Focus area for this session"`
	Questions  []QuestionCreateSpecInput `json:"questions" jsonschema:"Initial questions to create with the session"`
}

type QuestionCreateSpecInput struct {
	Text      string  `json:"text" jsonschema:"The question text"`
	Area      *string `json:"area" jsonschema:"Topic area"`
	Rationale *string `json:"rationale" jsonschema:"Why this question matters"`
	Priority  *string `json:"priority" jsonschema:"Priority: high, medium, low. Default: medium"`
	ParentID  *string `json:"parent_id" jsonschema:"Parent question ID for follow-ups"`
	Position  int     `json:"position" jsonschema:"Sort order"`
}

func RegisterSessionCreate(srv *mcp.Server, svc *service.SessionService, log *slog.Logger) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "session_create",
		Description: "Creates a new research session with optional initial questions. Session and questions are created atomically. Returns session_id and question count.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input SessionCreateInput) (*mcp.CallToolResult, any, error) {
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

		var questions []service.CreateQuestionRequest
		for _, q := range input.Questions {
			priority := domain.Priority(derefStr(q.Priority))
			if priority == "" {
				priority = domain.PriorityMedium
			}
			questions = append(questions, service.CreateQuestionRequest{
				Text:      q.Text,
				Area:      derefStr(q.Area),
				Rationale: derefStr(q.Rationale),
				Priority:  priority,
				ParentID:  derefStr(q.ParentID),
				Position:  q.Position,
			})
		}

		session, createdQuestions, err := svc.Create(ctx, service.CreateSessionRequest{
			ResearchID: input.ResearchID,
			Title:      input.Title,
			Focus:      derefStr(input.Focus),
			Questions:  questions,
		})
		if err != nil {
			return errorResult(err.Error())
		}

		return successResult(map[string]any{
			"session_id":        session.ID,
			"title":             session.Title,
			"status":            session.Status,
			"questions_created": len(createdQuestions),
		})
	})
}
