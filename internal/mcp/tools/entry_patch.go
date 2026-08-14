package tools

import (
	"context"
	"log/slog"

	"github.com/butschster/mcp-research/internal/service"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// BlockData is a named map type so the field can be a pointer: an op that only
// deletes or ticks carries no data, and a plain map would still be flagged by
// the repo's nullable-input rule.
type BlockData map[string]any

type EntryPatchInput struct {
	EntryID string        `json:"entry_id" jsonschema:"ID of the blocks entry to patch"`
	Ops     []BlockOpSpec `json:"ops" jsonschema:"Operations applied in order as one atomic change. If any of them fails, nothing is written"`
	Rev     *string       `json:"rev" jsonschema:"Document revision from entry_read. When given, the patch is rejected if the document changed since you read it. Send it for structural edits"`
}

type BlockOpSpec struct {
	Op      string     `json:"op" jsonschema:"update, insert, delete, move or set_state"`
	ID      *string    `json:"id" jsonschema:"Block id to act on. Needed by update, delete, move and set_state. On insert, an id you choose for the new block so a later op can point at it"`
	Type    *string    `json:"type" jsonschema:"Block type, per /llms/blocks.md. Needed by insert; on update it changes the kind, omit to keep it"`
	Data    *BlockData `json:"data" jsonschema:"Block data for the type, replacing the block's data entirely. Needed by insert and update"`
	After   *string    `json:"after" jsonschema:"Place after this block id"`
	Before  *string    `json:"before" jsonschema:"Place before this block id"`
	At      *string    `json:"at" jsonschema:"Place at start or end. Used when neither after nor before is given; the default is end"`
	Item    *string    `json:"item" jsonschema:"set_state only: key of the checklist item to tick"`
	Checked *bool      `json:"checked" jsonschema:"set_state only: the new value. This is an absolute setting, not a toggle"`
}

func RegisterEntryPatch(srv *mcp.Server, svc *service.EntryService, log *slog.Logger) {
	mcp.AddTool(srv, &mcp.Tool{
		Name: "entry_patch",
		Description: "Edits one or more blocks of a blocks entry by id, instead of re-sending the whole document. " +
			"Ops apply in order as one atomic change: if any of them fails the whole patch is rejected and nothing is written. " +
			"Unlike entry_update, this is strict — an unknown block id or malformed data is an error, not a silently dropped block. " +
			"Read the entry first to learn block ids and the rev.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input EntryPatchInput) (*mcp.CallToolResult, any, error) {
		if input.EntryID == "" {
			return validationErrorResult([]string{"entry_id is required"})
		}
		if len(input.Ops) == 0 {
			return validationErrorResult([]string{"ops is required: give at least one operation"})
		}

		ops := make([]service.BlockOp, 0, len(input.Ops))
		for _, o := range input.Ops {
			op := service.BlockOp{
				Op:      o.Op,
				ID:      derefStr(o.ID),
				Type:    derefStr(o.Type),
				After:   derefStr(o.After),
				Before:  derefStr(o.Before),
				At:      derefStr(o.At),
				Item:    derefStr(o.Item),
				Checked: o.Checked,
			}
			if o.Data != nil {
				op.Data = map[string]any(*o.Data)
			}
			ops = append(ops, op)
		}

		entry, err := svc.PatchBlocks(ctx, input.EntryID, service.PatchBlocksRequest{
			Ops: ops,
			Rev: derefStr(input.Rev),
		})
		if err != nil {
			log.Error("entry_patch failed", "entry_id", input.EntryID, "error", err)
			return errorResult(err.Error())
		}

		result := map[string]any{
			"entry_id": entry.ID,
			"code":     entry.Code,
			"applied":  len(ops),
			"rev":      service.DocumentRev(entry.Content),
			"updated":  true,
		}
		if entry.BlockReport != nil {
			result["blocks"] = entry.BlockReport.Blocks
			result["state_preserved"] = entry.BlockReport.StatePreserved
			result["state_lost"] = entry.BlockReport.StateLost
		}
		return successResult(result)
	})
}
