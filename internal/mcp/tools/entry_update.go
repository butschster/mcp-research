package tools

import (
	"context"
	"log/slog"

	"github.com/butschster/mcp-research/internal/domain"
	"github.com/butschster/mcp-research/internal/service"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type EntryUpdateInput struct {
	EntryID     string           `json:"entry_id" jsonschema:"ID of the entry to update"`
	Title       *string          `json:"title" jsonschema:"New title"`
	Content     *string          `json:"content" jsonschema:"Replace entire content"`
	Description *string          `json:"description" jsonschema:"New description"`
	Status      *string          `json:"status" jsonschema:"New status. Must be one of the section's allowed_entry_statuses."`
	Tags        []string         `json:"tags" jsonschema:"Replace tags"`
	TextReplace *TextReplaceSpec `json:"text_replace" jsonschema:"Replace first occurrence of 'from' with 'to' in content"`
	SessionID   *string          `json:"session_id" jsonschema:"Link entry to a session (pass empty string to unlink)"`
}

type TextReplaceSpec struct {
	From string `json:"from" jsonschema:"Text to find"`
	To   string `json:"to" jsonschema:"Replacement text"`
}

func RegisterEntryUpdate(srv *mcp.Server, svc *service.EntryService, log *slog.Logger) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "entry_update",
		Description: "Updates an entry. Supports text_replace for surgical content edits (replaces first occurrence of 'from' with 'to'). Only provided fields are updated.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input EntryUpdateInput) (*mcp.CallToolResult, any, error) {
		if input.EntryID == "" {
			return validationErrorResult([]string{"entry_id is required"})
		}

		var status *domain.EntryStatus
		if input.Status != nil {
			s := domain.EntryStatus(*input.Status)
			status = &s
		}

		var textReplace *service.TextReplace
		if input.TextReplace != nil {
			textReplace = &service.TextReplace{
				From: input.TextReplace.From,
				To:   input.TextReplace.To,
			}
		}

		entry, err := svc.Update(ctx, input.EntryID, service.UpdateEntryRequest{
			Title:       input.Title,
			Content:     input.Content,
			Description: input.Description,
			Status:      status,
			Tags:        input.Tags,
			TextReplace: textReplace,
			SessionID:   input.SessionID,
		})
		if err != nil {
			return errorResult(err.Error())
		}

		return successResult(map[string]any{
			"entry_id": entry.ID,
			"title":    entry.Title,
			"status":   entry.Status,
			"updated":  true,
		})
	})
}
