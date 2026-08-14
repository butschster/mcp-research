package tools

import (
	"context"
	"log/slog"

	"github.com/butschster/mcp-research/internal/domain"
	"github.com/butschster/mcp-research/internal/service"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type EntryHistoryInput struct {
	EntryID string `json:"entry_id" jsonschema:"ID of the entry whose history to read"`
	Limit   *int   `json:"limit,omitempty" jsonschema:"How many revisions to return, newest first (default 20)"`
}

func RegisterEntryHistory(srv *mcp.Server, svc *service.EntryService, log *slog.Logger) {
	mcp.AddTool(srv, &mcp.Tool{
		Name: "entry_history",
		Description: "Lists the revisions of an entry, newest first: who wrote each one (agent, human, import or restore), " +
			"during which session, and what it changed. Content is not included — use entry_diff to see what changed, " +
			"or entry_read for the current text. Read this before rewriting an entry another session wrote.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input EntryHistoryInput) (*mcp.CallToolResult, any, error) {
		if input.EntryID == "" {
			return validationErrorResult([]string{"entry_id is required"})
		}

		entry, revisions, err := svc.History(ctx, input.EntryID)
		if err != nil {
			return errorResult(err.Error())
		}

		limit := 20
		if input.Limit != nil && *input.Limit > 0 {
			limit = *input.Limit
		}
		truncated := false
		if len(revisions) > limit {
			revisions = revisions[:limit]
			truncated = true
		}

		out := map[string]any{
			"entry_id":   entry.ID,
			"entry_code": entry.Code,
			"title":      entry.Title,
			"revisions":  summarizeRevisions(revisions),
		}
		if truncated {
			out["truncated"] = true
			out["hint"] = "More revisions exist; raise limit to see them."
		}
		return successResult(out)
	})
}

// summarizeRevisions keeps a history listing readable in a tool result: the
// fields an agent decides on, and none of the ids it has no use for.
func summarizeRevisions(revisions []*domain.EntryRevision) []map[string]any {
	out := make([]map[string]any, 0, len(revisions))
	for _, rev := range revisions {
		row := map[string]any{
			"revision":    rev.Revision,
			"author":      rev.AuthorKind,
			"created_at":  rev.CreatedAt,
			"title":       rev.Title,
			"status":      rev.Status,
			"entry_type":  rev.Type,
			"tags":        rev.Tags,
			"description": rev.Description,
		}
		if rev.Summary != "" {
			row["summary"] = rev.Summary
		}
		if rev.SessionCode != "" {
			row["session"] = rev.SessionCode
		}
		out = append(out, row)
	}
	return out
}
