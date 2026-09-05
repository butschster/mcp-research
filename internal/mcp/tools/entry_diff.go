package tools

import (
	"context"
	"log/slog"

	"github.com/dovod-app/app/internal/service"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type EntryDiffInput struct {
	EntryID string `json:"entry_id" jsonschema:"ID of the entry to compare"`
	From    *int   `json:"from,omitempty" jsonschema:"Revision to compare from (default: the one before 'to')"`
	To      *int   `json:"to,omitempty" jsonschema:"Revision to compare to (default: the newest revision)"`
}

func RegisterEntryDiff(srv *mcp.Server, svc *service.EntryService, log *slog.Logger) {
	mcp.AddTool(srv, &mcp.Tool{
		Name: "entry_diff",
		Description: "Compares two revisions of an entry and returns a unified diff. With no revision numbers it shows " +
			"the most recent change. Use it to see what a previous session did to a document before you rewrite it, " +
			"or to check what your own edit actually changed. Block documents are compared as markdown, not as JSON.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input EntryDiffInput) (*mcp.CallToolResult, any, error) {
		if input.EntryID == "" {
			return validationErrorResult([]string{"entry_id is required"})
		}

		from, to := 0, 0
		if input.From != nil {
			from = *input.From
		}
		if input.To != nil {
			to = *input.To
		}
		if from < 0 || to < 0 {
			return validationErrorResult([]string{"revision numbers start at 1"})
		}

		diff, err := svc.Diff(ctx, input.EntryID, from, to)
		if err != nil {
			return errorResult(err.Error())
		}

		out := map[string]any{
			"entry_id":   diff.EntryID,
			"entry_code": diff.EntryCode,
			"to":         diff.To.Revision,
			"author":     diff.To.AuthorKind,
			"changed_at": diff.To.CreatedAt,
			"summary":    diff.Summary,
			"added":      diff.Content.Added,
			"removed":    diff.Content.Removed,
			// The unified text rather than the structured line list: a diff is
			// read here, not rendered, and the +/- form is what a model has seen
			// a million of.
			"diff": diff.Content.Unified,
		}
		// From is nil when there is nothing before `to` — revision 1, which is
		// every entry that predates the revisions migration and every entry not
		// edited since it was written. A title change is reported only when
		// there is an earlier title to have changed from: reading From.Title in
		// the nil case panicked the process, and an MCP tool call runs on a
		// goroutine outside net/http's recovery, so it took the server down for
		// everyone rather than failing the one call.
		if diff.From != nil {
			out["from"] = diff.From.Revision
			if diff.Title != nil {
				out["title_before"] = diff.From.Title
				out["title_after"] = diff.To.Title
			}
		} else {
			out["from"] = nil
			out["note"] = "No earlier revision — this is the entry as it was first written."
		}
		if diff.Content.Truncated {
			out["truncated"] = true
			out["hint"] = "The documents were too large to align line by line; the diff shows a full replacement."
		}
		return successResult(out)
	})
}
