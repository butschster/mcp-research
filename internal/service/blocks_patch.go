package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/dovod-app/app/internal/domain"
)

// Block-level operations addressed by block id.
//
// The rule that separates this file from NormalizeBlockDocument: a whole-document
// write is FORGIVING — an unknown type or malformed data is dropped, because the
// author sent the entire article, still holds it, and can read back what stored.
// A patch is STRICT — it fails whole and writes nothing — because the caller did
// not send the document and therefore has nothing to compare against. A silently
// dropped op is indistinguishable from success, and the agent reports a fix it
// never made.
//
// In one line, for the docs: you are trusted when you write an article; you are
// held to your word when you point at a block.

// BlockOp is one operation of a patch.
type BlockOp struct {
	Op      string
	ID      string
	Type    string
	Data    map[string]any
	After   string
	Before  string
	At      string
	Item    string
	Checked *bool
}

// Op names.
const (
	OpUpdate   = "update"
	OpInsert   = "insert"
	OpDelete   = "delete"
	OpMove     = "move"
	OpSetState = "set_state"
)

// PatchBlocksRequest is one atomic change to a block document.
type PatchBlocksRequest struct {
	Ops []BlockOp
	// Rev, when set, must match the stored document's revision or the patch is
	// rejected. Structural ops should send it; a checkbox click cannot afford the
	// round trip to fetch one and does not need it — a set_state op names exactly
	// one item and cannot clobber prose.
	Rev string
}

// ErrConflict means the document changed since the caller read it.
var ErrConflict = fmt.Errorf("document changed since it was read")

// PatchBlocks applies ops to a blocks entry.
func (s *EntryService) PatchBlocks(ctx context.Context, entryID string, req PatchBlocksRequest) (*domain.Entry, error) {
	entry, err := s.entries.FindByID(ctx, entryID)
	if err != nil {
		return nil, fmt.Errorf("find entry: %w", err)
	}
	if entry == nil {
		return nil, ErrNotFound
	}
	// Ownership before anything that could describe the target: "no such block"
	// must be unreachable for someone who may not open the entry at all.
	if err := s.access.Write(ctx, entry.ResearchID); err != nil {
		return nil, err
	}
	if entry.Type != domain.EntryBlocks {
		return nil, fmt.Errorf("entry %s is a %s entry: block operations need entry_type=blocks", entry.Code, entry.Type)
	}
	if len(req.Ops) == 0 {
		return nil, fmt.Errorf("no ops given")
	}
	// A tick is not an edit to the document. Recording it as a revision would
	// bury the writes that changed what the entry says under a history of
	// checkbox clicks, and restoring one would be restoring nothing.
	note := s.resolveSession(ctx, entry, revisionNote{
		summary: summarizeOps(req.Ops),
		skip:    stateOnly(req.Ops),
	})

	// A tick moves no prose, so it can strand no mark, and reading every
	// annotation of the entry to prove that on each checkbox click is exactly
	// the cost stateOnly exists to avoid.
	var anchorsBefore map[string]domain.AnchorState
	if !stateOnly(req.Ops) {
		anchorsBefore = s.captureAnchors(ctx, entry)
	}

	report, err := s.mutateBlocks(ctx, entry, req.Rev, stateAuthoritative, note, func(doc *domain.BlockDocument) error {
		blocks, aerr := applyOps(doc.Blocks, req.Ops)
		if aerr != nil {
			return aerr
		}
		doc.Blocks = blocks
		return nil
	})
	if err != nil {
		return nil, err
	}
	entry.BlockReport = &report

	// A state-only patch changes no prose, so it cannot change what the document
	// contributes to the index or to cross-references. Skipping the rebuild is
	// what makes a checkbox click cheap: otherwise every tick would delete and
	// reinsert every crossref row of the entry to store one boolean.
	if !stateOnly(req.Ops) {
		s.updateCrossRefs(ctx, entry)
		s.updateExternalLinks(ctx, entry)
		entry.AnnReport = s.anchorReport(ctx, entry, anchorsBefore)
	}
	emit(ctx, s.events, Event{Type: "entry.updated", ResearchID: entry.ResearchID, EntityID: entry.ID, Entity: "entry"})
	return entry, nil
}

// summarizeOps labels a patch in the history by what it did, so a revision list
// reads as "Inserted 2 blocks, updated 1" rather than as a column of "Updated".
func summarizeOps(ops []BlockOp) string {
	counts := map[string]int{}
	order := []string{}
	for _, op := range ops {
		if _, seen := counts[op.Op]; !seen {
			order = append(order, op.Op)
		}
		counts[op.Op]++
	}
	verbs := map[string]string{
		OpUpdate: "updated", OpInsert: "inserted", OpDelete: "deleted",
		OpMove: "moved", OpSetState: "ticked",
	}
	parts := make([]string, 0, len(order))
	for _, op := range order {
		verb, ok := verbs[op]
		if !ok {
			verb = op
		}
		parts = append(parts, fmt.Sprintf("%s %d", verb, counts[op]))
	}
	if len(parts) == 0 {
		return ""
	}
	return "Patched blocks: " + strings.Join(parts, ", ")
}

func stateOnly(ops []BlockOp) bool {
	for _, op := range ops {
		if op.Op != OpSetState {
			return false
		}
	}
	return true
}

// applyOps runs the ops in order against one working copy, so an op may target a
// block an earlier op inserted. Any failure returns an error and the caller
// writes nothing.
func applyOps(blocks []domain.Block, ops []BlockOp) ([]domain.Block, error) {
	work := make([]domain.Block, len(blocks))
	copy(work, blocks)

	for i, op := range ops {
		var err error
		switch op.Op {
		case OpUpdate:
			work, err = applyUpdate(work, op)
		case OpInsert:
			work, err = applyInsert(work, op)
		case OpDelete:
			work, err = applyDelete(work, op)
		case OpMove:
			work, err = applyMove(work, op)
		case OpSetState:
			work, err = applySetState(work, op)
		default:
			err = fmt.Errorf("unknown op %q: want %s, %s, %s, %s or %s",
				op.Op, OpUpdate, OpInsert, OpDelete, OpMove, OpSetState)
		}
		if err != nil {
			return nil, fmt.Errorf("op %d (%s): %w", i, op.Op, err)
		}
	}
	if len(work) == 0 {
		return nil, fmt.Errorf("a patch cannot leave the document empty; delete the entry instead")
	}
	if len(work) > domain.MaxBlocks {
		return nil, fmt.Errorf("the document would hold %d blocks; the cap is %d", len(work), domain.MaxBlocks)
	}
	return work, nil
}

func indexOf(blocks []domain.Block, id string) int {
	for i, b := range blocks {
		if b.ID == id {
			return i
		}
	}
	return -1
}

// notFound names the ids that do exist, so a caller can re-anchor without
// fetching the whole document again.
func notFound(blocks []domain.Block, id string) error {
	known := make([]string, 0, len(blocks))
	for i, b := range blocks {
		if i == 20 {
			known = append(known, "…")
			break
		}
		known = append(known, b.ID)
	}
	if id == "" {
		return fmt.Errorf("no block id given; the document holds %s", strings.Join(known, ", "))
	}
	return fmt.Errorf("no block %q in this document; it holds %s", id, strings.Join(known, ", "))
}

func applyUpdate(blocks []domain.Block, op BlockOp) ([]domain.Block, error) {
	i := indexOf(blocks, op.ID)
	if i < 0 {
		return nil, notFound(blocks, op.ID)
	}
	blockType := blocks[i].Type
	if op.Type != "" {
		blockType = domain.BlockType(op.Type)
	}
	if op.Data == nil {
		return nil, fmt.Errorf("update needs data")
	}
	data, ok := normalizeOne(blockType, op.Data)
	if !ok {
		return nil, fmt.Errorf("%s data rejected: see /llms/blocks.md for the fields %s needs", blockType, blockType)
	}
	// State belongs to the row and survives an update of the prose around it.
	if st, had := blocks[i].Data[blockStateKey]; had && blockType == blocks[i].Type {
		data[blockStateKey] = st
	}
	blocks[i].Type = blockType
	blocks[i].Data = data
	return blocks, nil
}

func applyInsert(blocks []domain.Block, op BlockOp) ([]domain.Block, error) {
	if op.Type == "" {
		return nil, fmt.Errorf("insert needs a type")
	}
	if op.Data == nil {
		return nil, fmt.Errorf("insert needs data")
	}
	data, ok := normalizeOne(domain.BlockType(op.Type), op.Data)
	if !ok {
		return nil, fmt.Errorf("%s data rejected: see /llms/blocks.md for the fields %s needs", op.Type, op.Type)
	}
	at, err := placement(blocks, op)
	if err != nil {
		return nil, err
	}
	seen := map[string]bool{}
	for _, b := range blocks {
		seen[b.ID] = true
	}
	// The schema promises the id you choose is the id a later op can point at.
	// Silently minting a different one broke that promise and made the next op
	// fail on an id the caller had every reason to believe existed.
	if op.ID != "" {
		if seen[op.ID] {
			return nil, fmt.Errorf("block id %q is already used in this document", op.ID)
		}
		if !blockIDPattern.MatchString(op.ID) {
			return nil, fmt.Errorf("block id %q is not usable: %s", op.ID, blockIDRule)
		}
	}
	block := domain.Block{ID: blockID(op.ID, seen), Type: domain.BlockType(op.Type), Data: data}

	out := make([]domain.Block, 0, len(blocks)+1)
	out = append(out, blocks[:at]...)
	out = append(out, block)
	out = append(out, blocks[at:]...)
	return out, nil
}

func applyDelete(blocks []domain.Block, op BlockOp) ([]domain.Block, error) {
	i := indexOf(blocks, op.ID)
	if i < 0 {
		return nil, notFound(blocks, op.ID)
	}
	return append(blocks[:i:i], blocks[i+1:]...), nil
}

func applyMove(blocks []domain.Block, op BlockOp) ([]domain.Block, error) {
	i := indexOf(blocks, op.ID)
	if i < 0 {
		return nil, notFound(blocks, op.ID)
	}
	block := blocks[i]
	rest := append(blocks[:i:i], blocks[i+1:]...)
	at, err := placement(rest, op)
	if err != nil {
		return nil, err
	}
	out := make([]domain.Block, 0, len(blocks))
	out = append(out, rest[:at]...)
	out = append(out, block)
	out = append(out, rest[at:]...)
	return out, nil
}

// applySetState ticks or unticks one item of one block. Absolute, never a
// toggle: two clients flipping the same box must converge on what the humans
// meant rather than on how many messages arrived.
func applySetState(blocks []domain.Block, op BlockOp) ([]domain.Block, error) {
	i := indexOf(blocks, op.ID)
	if i < 0 {
		return nil, notFound(blocks, op.ID)
	}
	if blocks[i].Type != domain.BlockChecklist {
		// A task_ref looks tickable and is not: its boxes are the tasks, and the
		// only way to move one is to move the task. Said plainly here, because
		// "is a task_ref, not a checklist" leaves the caller to guess where the
		// state actually lives.
		if blocks[i].Type == domain.BlockTaskRef {
			return nil, fmt.Errorf("block %q is a %s: it has no state of its own — tick it with task_update, which is where the status lives", op.ID, domain.BlockTaskRef)
		}
		// Writing state onto a paragraph would leave a field its normalizer does
		// not know, and the next whole-document write would drop the block.
		return nil, fmt.Errorf("block %q is a %s, not a %s", op.ID, blocks[i].Type, domain.BlockChecklist)
	}
	if op.Item == "" {
		return nil, fmt.Errorf("set_state needs an item key")
	}
	if op.Checked == nil {
		return nil, fmt.Errorf("set_state needs checked")
	}
	known := false
	for _, item := range checklistItems(blocks[i].Data) {
		if item.Key == op.Item {
			known = true
			break
		}
	}
	if !known {
		return nil, fmt.Errorf("block %q has no item %q", op.ID, op.Item)
	}

	state, _ := blocks[i].Data[blockStateKey].(map[string]any)
	next := map[string]any{}
	for k, v := range state {
		next[k] = v
	}
	if *op.Checked {
		next[op.Item] = true
	} else {
		delete(next, op.Item)
	}
	data := map[string]any{}
	for k, v := range blocks[i].Data {
		data[k] = v
	}
	if len(next) > 0 {
		data[blockStateKey] = next
	} else {
		delete(data, blockStateKey)
	}
	blocks[i].Data = data
	return blocks, nil
}

// placement resolves an anchor to an index. Exactly one anchor, or none for
// "append" — two anchors is an error rather than a precedence rule nobody reads.
func placement(blocks []domain.Block, op BlockOp) (int, error) {
	given := 0
	for _, a := range []string{op.After, op.Before, op.At} {
		if a != "" {
			given++
		}
	}
	if given > 1 {
		return 0, fmt.Errorf("give one of after, before or at")
	}
	switch {
	case op.After != "":
		i := indexOf(blocks, op.After)
		if i < 0 {
			return 0, notFound(blocks, op.After)
		}
		return i + 1, nil
	case op.Before != "":
		i := indexOf(blocks, op.Before)
		if i < 0 {
			return 0, notFound(blocks, op.Before)
		}
		return i, nil
	case op.At == "start":
		return 0, nil
	case op.At == "end", op.At == "":
		return len(blocks), nil
	default:
		return 0, fmt.Errorf("at must be start or end, got %q", op.At)
	}
}

// normalizeOne runs a single block's data through the same per-type normalizer a
// whole-document write uses. Only the reaction to rejection differs: there it
// drops the block, here it fails the patch.
func normalizeOne(t domain.BlockType, data map[string]any) (map[string]any, bool) {
	norm, ok := blockNormalizers[t]
	if !ok {
		return nil, false
	}
	return norm(data)
}
