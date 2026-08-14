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
	EntryType   *string          `json:"entry_type" jsonschema:"Optional new content kind: markdown, blocks, or artifact (sugar for one html block). Switching to blocks requires content in block form in the same call; switching from blocks to markdown converts what is stored"`
	Title       *string          `json:"title" jsonschema:"New title"`
	Content     *string          `json:"content" jsonschema:"Replace entire content"`
	Description *string          `json:"description" jsonschema:"New description"`
	Status      *string          `json:"status" jsonschema:"New status: draft, active, completed, archived"`
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
		Description: "Updates an entry. Only provided fields are updated. text_replace does surgical edits of a markdown entry and is refused on a blocks entry — use entry_patch there, which edits blocks by id.",
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

		var entryType *domain.EntryType
		if input.EntryType != nil {
			t := domain.EntryType(*input.EntryType)
			entryType = &t
		}

		entry, err := svc.Update(ctx, input.EntryID, service.UpdateEntryRequest{
			Type:        entryType,
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

		result := map[string]any{
			"entry_id": entry.ID,
			"title":    entry.Title,
			"status":   entry.Status,
			"updated":  true,
		}
		// The whole-document rewrite is exactly where a human's checklist ticks
		// die, and they die silently unless the writer is told.
		if r := entry.BlockReport; r != nil {
			result["blocks"] = r.Blocks
			result["blocks_reidentified"] = r.Reidentified
			result["state_preserved"] = r.StatePreserved
			result["state_lost"] = r.StateLost
			result["rev"] = service.DocumentRev(entry.Content)
		}
		return successResult(result)
	})
}
