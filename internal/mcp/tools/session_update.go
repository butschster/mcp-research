package tools

import (
	"context"
	"log/slog"

	"github.com/butschster/mcp-research/internal/domain"
	"github.com/butschster/mcp-research/internal/service"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type SessionUpdateInput struct {
	SessionID string  `json:"session_id" jsonschema:"ID of the session to update"`
	Title     *string `json:"title" jsonschema:"New title"`
	Focus     *string `json:"focus" jsonschema:"New focus"`
	Status    *string `json:"status" jsonschema:"New status: active, completed, archived"`
	Notes     *string `json:"notes" jsonschema:"Replace session notes (mutually exclusive with add_note)"`
	AddNote   *string `json:"add_note" jsonschema:"Append to session notes (mutually exclusive with notes)"`
}

func RegisterSessionUpdate(srv *mcp.Server, svc *service.SessionService, log *slog.Logger) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "session_update",
		Description: "Updates a session. Supports notes (replace) and add_note (append) which are mutually exclusive. Only provided fields are updated.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input SessionUpdateInput) (*mcp.CallToolResult, any, error) {
		if input.SessionID == "" {
			return validationErrorResult([]string{"session_id is required"})
		}

		var status *domain.SessionStatus
		if input.Status != nil {
			s := domain.SessionStatus(*input.Status)
			status = &s
		}

		session, err := svc.Update(ctx, input.SessionID, service.UpdateSessionRequest{
			Title:   input.Title,
			Focus:   input.Focus,
			Status:  status,
			Notes:   input.Notes,
			AddNote: input.AddNote,
		})
		if err != nil {
			return errorResult(err.Error())
		}

		return successResult(map[string]any{
			"session_id": session.ID,
			"status":     session.Status,
			"updated":    true,
		})
	})
}
