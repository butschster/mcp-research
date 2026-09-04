package storage

import (
	"context"
	"fmt"

	"github.com/uptrace/bun"
)

// BlockRow is one block of a `blocks` entry.
//
// Data and State are JSON documents kept as strings: the per-type shape belongs
// to the service layer's normalizer, and storage has no business parsing it.
// State is separate from Data on purpose — it is the field an agent may not
// write, and keeping it in its own column is what makes "carry it forward" a
// column the reconcile simply does not touch.
type BlockRow struct {
	EntryID    string
	ResearchID string
	BlockID    string
	Position   int
	Type       string
	Data       string
	State      string
}

type BlockRepository struct {
	db *bun.DB
}

func NewBlockRepository(db *bun.DB) *BlockRepository {
	return &BlockRepository{db: db}
}

// DB exposes the handle so a service can open the transaction a multi-statement
// block write has to run in. Nothing else should reach for it.
func (r *BlockRepository) DB() *bun.DB { return r.db }

// FindByEntry returns a document's blocks in render order.
func (r *BlockRepository) FindByEntry(ctx context.Context, q Querier, entryID string) ([]BlockRow, error) {
	if q == nil {
		q = r.db
	}
	rows, err := q.NewSelect().
		ColumnExpr("entry_id, research_id, block_id, position, type, data, state").
		TableExpr("entry_blocks").
		Where("entry_id=?", entryID).
		OrderExpr("position").
		Rows(ctx)
	if err != nil {
		return nil, fmt.Errorf("find blocks: %w", err)
	}
	defer rows.Close()

	var out []BlockRow
	for rows.Next() {
		var b BlockRow
		if err := rows.Scan(&b.EntryID, &b.ResearchID, &b.BlockID, &b.Position, &b.Type, &b.Data, &b.State); err != nil {
			return nil, fmt.Errorf("scan block: %w", err)
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

// ReplaceForEntry makes the stored rows equal to blocks, in one pass.
//
// Delete-all-then-insert rather than a per-row diff: a document is at most 400
// blocks, the caller has already decided what every row should contain, and the
// whole thing runs inside one transaction — so the simple form is both atomic
// and cheap. What must NOT be lost across it is state, and that is the caller's
// job (see service.reconcileBlocks), because only the caller knows which old
// block a new one corresponds to.
func (r *BlockRepository) ReplaceForEntry(ctx context.Context, q Querier, entryID string, blocks []BlockRow) error {
	if q == nil {
		q = r.db
	}
	if _, err := q.NewDelete().Table("entry_blocks").Where("entry_id=?", entryID).Exec(ctx); err != nil {
		return fmt.Errorf("clear blocks: %w", err)
	}
	for i, b := range blocks {
		if _, err := q.NewInsert().Table("entry_blocks").Model(&map[string]any{
			"entry_id":    entryID,
			"research_id": b.ResearchID,
			"block_id":    b.BlockID,
			"position":    i,
			"type":        b.Type,
			"data":        b.Data,
			"state":       b.State,
		}).Exec(ctx); err != nil {
			return fmt.Errorf("insert block %d (%s): %w", i, b.BlockID, err)
		}
	}
	return nil
}

// SetState writes one block's state and nothing else. This is the checkbox path:
// it touches one row, leaves the document's prose alone, and cannot collide with
// a concurrent edit to a different block.
//
// Reports whether the block existed, so the caller can tell the difference
// between "ticked" and "the block is gone" rather than silently succeeding.
func (r *BlockRepository) SetState(ctx context.Context, q Querier, entryID, blockID, state string) (bool, error) {
	if q == nil {
		q = r.db
	}
	res, err := q.NewUpdate().
		Table("entry_blocks").
		Set("state=?", state).
		Where("entry_id=? AND block_id=?", entryID, blockID).
		Exec(ctx)
	if err != nil {
		return false, fmt.Errorf("set block state: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("set block state: %w", err)
	}
	return n > 0, nil
}

// DeleteForEntry drops a document's rows. Used when an entry stops being a block
// document; ordinary deletion is handled by the foreign key.
func (r *BlockRepository) DeleteForEntry(ctx context.Context, q Querier, entryID string) error {
	if q == nil {
		q = r.db
	}
	if _, err := q.NewDelete().Table("entry_blocks").Where("entry_id=?", entryID).Exec(ctx); err != nil {
		return fmt.Errorf("delete blocks: %w", err)
	}
	return nil
}
