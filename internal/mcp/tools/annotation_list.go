package tools

import (
	"context"
	"log/slog"

	"github.com/butschster/mcp-research/internal/domain"
	"github.com/butschster/mcp-research/internal/service"
	"github.com/butschster/mcp-research/internal/storage"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type AnnotationListInput struct {
	ResearchID string  `json:"research_id" jsonschema:"ID of the research"`
	Status     *string `json:"status" jsonschema:"Filter by status: open, answered, closed, dismissed. Default: open"`
	Kind       *string `json:"kind" jsonschema:"Filter by kind: verify, dig, disagree"`
	EntryID    *string `json:"entry_id" jsonschema:"Only marks in this entry"`
	Limit      *int    `json:"limit" jsonschema:"How many to take. Capped by the server at 15 — a pass is only as large as the diff a person will read"`
	Offset     *int    `json:"offset" jsonschema:"Skip this many, to continue a previous batch"`
}

func RegisterAnnotationList(srv *mcp.Server, svc *service.AnnotationService, log *slog.Logger) {
	mcp.AddTool(srv, &mcp.Tool{
		Name: "annotation_list",
		Description: `Lists the marks a person left on places in the documents — the queue you work when asked to go through their annotations.

Each mark carries the quote it was attached to and the current text of the block it sits in, so you do NOT need to read the documents to triage. Read the entry only for the ones you are about to work.

The kind says what is being asked for, and they are different jobs:
  verify   — find a source and cite it, or say plainly that you could not confirm it
  dig      — write a child entry and link it from the anchored block with [[E19]]
  disagree — do NOT rewrite the text to make the objection go away. Record both positions and who holds each. Rewriting destroys the disagreement instead of recording it.

Anchor state says where the marked text is now:
  anchored — the text is where it was
  drifted  — the block is there, the sentence changed under the mark
  moved    — the sentence turned up elsewhere; check the confidence before trusting it
  orphaned — the text is gone. Do not guess what it said: use entry_history and entry_diff from anchored_revision, or ask.

Take a batch from ONE entry where you can — the response groups by entry for exactly that reason. Answer with annotation_answer. You cannot close a mark; a person does that.`,
	}, func(ctx context.Context, req *mcp.CallToolRequest, input AnnotationListInput) (*mcp.CallToolResult, any, error) {
		if input.ResearchID == "" {
			return validationErrorResult([]string{"research_id is required"})
		}

		filter := storage.AnnotationFilter{
			EntryID: derefStr(input.EntryID),
			Limit:   domain.MaxAnnotationBatch,
			Offset:  derefInt(input.Offset),
		}
		if input.Limit != nil && *input.Limit > 0 {
			filter.Limit = *input.Limit
		}

		// Open by default. The queue is a work list, and a caller who asked for
		// "the annotations" means the ones still wanting an answer.
		status := domain.AnnotationStatus(derefStr(input.Status))
		if status == "" {
			status = domain.AnnotationOpen
		}
		if !status.Valid() {
			return validationErrorResult([]string{"status must be one of: open, answered, closed, dismissed"})
		}
		filter.Status = &status

		if k := derefStr(input.Kind); k != "" {
			kind := domain.AnnotationKind(k)
			if !kind.Valid() {
				return validationErrorResult([]string{"kind must be one of: verify, dig, disagree"})
			}
			filter.Kind = &kind
		}

		anns, err := svc.ListByResearch(ctx, input.ResearchID, filter)
		if err != nil {
			return errorResult(err.Error())
		}

		byEntry := map[string]int{}
		byKind := map[string]int{}
		items := make([]map[string]any, 0, len(anns))
		for _, a := range anns {
			byEntry[a.EntryCode]++
			byKind[string(a.Kind)]++

			item := map[string]any{
				"annotation_id": a.ID,
				"code":          a.Code,
				"entry_id":      a.EntryID,
				"entry_code":    a.EntryCode,
				"entry_title":   a.EntryTitle,
				// A markdown document has no blocks, so its marks are held by
				// text alone and drift more easily. An absent block_id says the
				// same thing only if you know that rule; this says it outright.
				"entry_type": a.EntryType,
				"kind":       a.Kind,
				"status":     a.Status,
				"quote":      a.Quote.Exact,
				"note":       a.Body,
				"attempts":   a.Attempts,
			}
			if a.BlockID != "" {
				item["block_id"] = a.BlockID
			}
			if a.AnchoredRevision > 0 {
				item["anchored_revision"] = a.AnchoredRevision
			}
			if a.Anchor != nil {
				anchor := map[string]any{
					"state":      a.Anchor.State,
					"strategy":   a.Anchor.Strategy,
					"confidence": a.Anchor.Confidence,
				}
				// The block's current text is the whole point of the queue read:
				// it is what lets a mark be judged without opening the document.
				if a.Anchor.Text != "" {
					anchor["block_text"] = a.Anchor.Text
				}
				item["anchor"] = anchor
			}
			// Prior attempts matter more than they look: a mark that came back
			// carries the reason it was rejected, and repeating the rejected
			// answer is the most common way a pass wastes itself.
			if a.Resolution != "" {
				item["previous_resolution"] = a.Resolution
			}
			// And the objections themselves. The answer that was refused and the
			// reason it was refused are different things — the column exists so
			// one does not overwrite the other — and an agent handed only its own
			// rejected answer has no way to do anything different the second time.
			if len(a.Rejections) > 0 {
				reasons := make([]map[string]any, 0, len(a.Rejections))
				for _, r := range a.Rejections {
					entry := map[string]any{"reason": r.Reason}
					if r.Revision > 0 {
						entry["revision"] = r.Revision
					}
					reasons = append(reasons, entry)
				}
				item["rejections"] = reasons
			}
			items = append(items, item)
		}

		out := map[string]any{
			"annotations": items,
			"count":       len(items),
			"by_entry":    byEntry,
			"by_kind":     byKind,
		}
		if len(items) == filter.Limit {
			out["hint"] = "This is a full batch — there may be more. Finish these, have the person accept them, then continue with offset."
		}
		return successResult(out)
	})
}
