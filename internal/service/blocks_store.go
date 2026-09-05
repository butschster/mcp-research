package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/dovod-app/app/internal/domain"
	"github.com/dovod-app/app/internal/storage"
)

// Blocks of a `blocks` entry live in entry_blocks, one row each. This file owns
// the two directions — document to rows, rows to document — and the rule that
// makes the whole thing worth doing: state a human put on a block survives an
// agent rewriting the document around it.
//
// entries.content is written from the same rows in the same transaction. It is
// a projection: everything that reads an entry (search, exports, the portable
// format, the web UI) keeps reading one string and never learns about rows.

// blockStateKey is where per-item state sits in an assembled block's data. It is
// server-owned: the normalizer strips it from anything an author sends, so the
// field has exactly one origin and "carry it forward" is unambiguous.
const blockStateKey = "state"

// DocumentRev identifies a stored document. A content hash rather than a column:
// it needs no migration, cannot drift from what it describes, and changes if and
// only if the document changed. updated_at cannot do this job — it is stored at
// second granularity, and two ticks inside one second are indistinguishable.
func DocumentRev(content string) string {
	sum := sha256.Sum256([]byte(content))
	return hex.EncodeToString(sum[:])[:12]
}

// splitState separates a block's prose from its server-owned state.
func splitState(data map[string]any) (map[string]any, string) {
	if len(data) == 0 {
		return map[string]any{}, ""
	}
	raw, ok := data[blockStateKey]
	if !ok {
		return data, ""
	}
	clean := make(map[string]any, len(data))
	for k, v := range data {
		if k != blockStateKey {
			clean[k] = v
		}
	}
	m, ok := raw.(map[string]any)
	if !ok || len(m) == 0 {
		return clean, ""
	}
	b, err := json.Marshal(m)
	if err != nil {
		return clean, ""
	}
	return clean, string(b)
}

func parseState(raw string) map[string]any {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	var m map[string]any
	if err := json.Unmarshal([]byte(raw), &m); err != nil {
		return nil
	}
	return m
}

// rowsToDocument rebuilds the document a reader sees, state included.
func rowsToDocument(rows []storage.BlockRow) *domain.BlockDocument {
	doc := &domain.BlockDocument{Version: domain.BlockDocumentVersion}
	for _, r := range rows {
		var data map[string]any
		if err := json.Unmarshal([]byte(r.Data), &data); err != nil || data == nil {
			data = map[string]any{}
		}
		if st := parseState(r.State); len(st) > 0 {
			data[blockStateKey] = st
		}
		doc.Blocks = append(doc.Blocks, domain.Block{
			ID:   r.BlockID,
			Type: domain.BlockType(r.Type),
			Data: data,
		})
	}
	return doc
}

// stateByItemText indexes a previous document's state by the item text it was
// attached to. This is the fallback identity: when an agent regenerates a
// checklist and mints fresh block ids, matching by id finds nothing, and the
// text of an item the human ticked is the only thing left to recognize it by.
// It is weaker than an id — reword an item and its tick is lost — but it rescues
// the case that actually happens.
func stateByItemText(rows []storage.BlockRow) map[string]bool {
	out := map[string]bool{}
	for _, r := range rows {
		st := parseState(r.State)
		if len(st) == 0 {
			continue
		}
		var data map[string]any
		if err := json.Unmarshal([]byte(r.Data), &data); err != nil {
			continue
		}
		for key, text := range itemTexts(data) {
			if checked, ok := st[key].(bool); ok && checked {
				out[normalizeItemText(text)] = true
			}
		}
	}
	return out
}

// itemTexts maps item key to item text for a block that has items.
func itemTexts(data map[string]any) map[string]string {
	out := map[string]string{}
	items, _ := data["items"].([]any)
	for _, it := range items {
		m, ok := it.(map[string]any)
		if !ok {
			continue
		}
		key, _ := m["key"].(string)
		text, _ := m["text"].(string)
		if key != "" {
			out[key] = text
		}
	}
	return out
}

func normalizeItemText(s string) string {
	return strings.ToLower(strings.Join(strings.Fields(s), " "))
}

// documentToRows converts a normalized document into rows, carrying state
// forward from prev. Matching is by block id first — the identity the format is
// built on — then by item text for blocks whose id was not recognized.
// stateSource says where the state in doc came from.
//
// A patch works on the document the server itself assembled, so its state IS the
// truth — including the absence of a tick, which is how unticking the last item
// of a block is expressed. Author content is the opposite: state was stripped on
// the way in, so anything stored has to be carried forward.
type stateSource int

const (
	stateFromAuthor stateSource = iota
	stateAuthoritative
)

func documentToRows(entry *domain.Entry, doc *domain.BlockDocument, prev []storage.BlockRow, src stateSource) ([]storage.BlockRow, domain.BlockSaveReport) {
	byID := make(map[string]storage.BlockRow, len(prev))
	for _, r := range prev {
		byID[r.BlockID] = r
	}
	byText := stateByItemText(prev)

	report := domain.BlockSaveReport{Blocks: len(doc.Blocks)}
	rows := make([]storage.BlockRow, 0, len(doc.Blocks))

	for i, b := range doc.Blocks {
		data, incoming := splitState(b.Data)
		raw, err := json.Marshal(data)
		if err != nil {
			raw = []byte("{}")
		}

		state := incoming
		if src == stateFromAuthor {
			if old, ok := byID[b.ID]; ok {
				if old.Type == string(b.Type) && state == "" {
					state = old.State
				}
			} else {
				report.Reidentified++
				// The id is new. If this block's items carry text the previous
				// document had ticked, the human's intent survives the rename.
				if state == "" {
					state = reviveByText(data, byText)
				}
			}
		}
		// A tick whose item is gone is not a tick. Dropping it here is what keeps
		// the report honest: an agent that kept the block id but re-keyed its
		// items used to be told nothing was lost while every box went blank.
		state = pruneState(state, data)
		rows = append(rows, storage.BlockRow{
			EntryID:    entry.ID,
			ResearchID: entry.ResearchID,
			BlockID:    b.ID,
			Position:   i,
			Type:       string(b.Type),
			Data:       string(raw),
			State:      state,
		})
	}

	report.StatePreserved, report.StateLost = countState(prev, rows)
	return rows, report
}

// reviveByText rebuilds a state map for a block whose id changed, from item text
// that was ticked in the previous document.
// reviveByText consumes what it matches: one tick in the old document may revive
// exactly one item. Otherwise a document rewritten into two checklists that share
// an item text would show work done twice that nobody did once.
func reviveByText(data map[string]any, byText map[string]bool) string {
	if len(byText) == 0 {
		return ""
	}
	state := map[string]any{}
	for key, text := range itemTexts(data) {
		norm := normalizeItemText(text)
		if norm == "" || !byText[norm] {
			continue
		}
		state[key] = true
		delete(byText, norm)
	}
	if len(state) == 0 {
		return ""
	}
	b, err := json.Marshal(state)
	if err != nil {
		return ""
	}
	return string(b)
}

// pruneState drops ticks whose item no longer exists in the block.
func pruneState(raw string, data map[string]any) string {
	st := parseState(raw)
	if len(st) == 0 {
		return ""
	}
	live := itemTexts(data)
	kept := map[string]any{}
	for key, v := range st {
		if _, ok := live[key]; !ok {
			continue
		}
		if b, isBool := v.(bool); isBool && b {
			kept[key] = true
		}
	}
	if len(kept) == 0 {
		return ""
	}
	b, err := json.Marshal(kept)
	if err != nil {
		return ""
	}
	return string(b)
}

// countState compares how many ticks existed before a write with how many
// survived it.
func countState(prev []storage.BlockRow, next []storage.BlockRow) (preserved, lost int) {
	before := 0
	for _, r := range prev {
		before += countChecked(r.State)
	}
	after := 0
	for _, r := range next {
		after += countChecked(r.State)
	}
	if after >= before {
		return before, 0
	}
	return after, before - after
}

func countChecked(raw string) int {
	n := 0
	for _, v := range parseState(raw) {
		if b, ok := v.(bool); ok && b {
			n++
		}
	}
	return n
}

// LoadBlockDocument reads an entry's blocks from their rows.
//
// A blocks entry with no rows falls back to its projection. Migration 017
// backfills every document that predates the rows, but a fallback that costs one
// parse is worth having: without it an entry that slipped through would look
// empty, and a patch that appends to an empty document REPLACES the article.
func (s *EntryService) LoadBlockDocument(ctx context.Context, entry *domain.Entry) (*domain.BlockDocument, error) {
	rows, err := s.blocks.FindByEntry(ctx, nil, entry.ID)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 && strings.TrimSpace(entry.Content) != "" {
		doc, perr := ParseStoredBlockDocument(entry.Content)
		if perr != nil {
			return nil, fmt.Errorf("entry %s has no blocks and its content is not a document: %w", entry.Code, perr)
		}
		return doc, nil
	}
	return rowsToDocument(rows), nil
}

// mutateBlocks is the only way a block document changes.
//
// The document is loaded, changed and written inside ONE transaction. Reading it
// outside was a real lost update, not a theoretical one: two people ticking two
// items of the same checklist both read the same rows, and the second write
// carried the first one's box back to unticked while reporting success.
//
// expectedRev, when set, is checked against the document as it is inside the
// transaction — checking it against a copy read earlier would leave the same
// window open one layer up.
// note says how the resulting revision is labelled, or that none is wanted —
// creation records its own after the fact, and a checkbox tick records none.
func (s *EntryService) mutateBlocks(
	ctx context.Context,
	entry *domain.Entry,
	expectedRev string,
	src stateSource,
	note revisionNote,
	mutate func(doc *domain.BlockDocument) error,
) (domain.BlockSaveReport, error) {
	var report domain.BlockSaveReport

	tx, err := s.blocks.DB().BeginTx(ctx, nil)
	if err != nil {
		return report, fmt.Errorf("begin: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	prev, err := s.blocks.FindByEntry(ctx, tx, entry.ID)
	if err != nil {
		return report, err
	}

	doc := rowsToDocument(prev)
	if len(prev) == 0 && strings.TrimSpace(entry.Content) != "" {
		// Rows are missing for a document that predates them (migration 017
		// backfills, this covers whatever slipped past). Without it a patch would
		// append to an empty document and replace the article with one block.
		if parsed, perr := ParseStoredBlockDocument(entry.Content); perr == nil {
			doc = parsed
		}
	}

	if expectedRev != "" {
		current, merr := MarshalBlockDocument(doc)
		if merr != nil {
			return report, merr
		}
		if DocumentRev(current) != expectedRev {
			return report, ErrConflict
		}
	}

	if err := mutate(doc); err != nil {
		return report, err
	}

	rows, rep := documentToRows(entry, doc, prev, src)
	report = rep

	if err := s.blocks.ReplaceForEntry(ctx, tx, entry.ID, rows); err != nil {
		return report, err
	}

	// The projection is what every reader sees, so it is built from the rows just
	// written rather than from the document that produced them.
	projection, err := MarshalBlockDocument(rowsToDocument(rows))
	if err != nil {
		return report, err
	}
	entry.Content = projection
	if err := s.entries.UpdateTx(ctx, tx, entry); err != nil {
		return report, fmt.Errorf("update entry: %w", err)
	}
	// Inside the same transaction as the write it describes: a revision that
	// outlived a rolled-back write would be a snapshot of a document that never
	// existed.
	if err := s.recordRevision(ctx, tx, entry, note); err != nil {
		return report, err
	}
	if err := tx.Commit(); err != nil {
		return report, fmt.Errorf("commit: %w", err)
	}
	return report, nil
}

// saveBlockDocument replaces a document wholesale.
func (s *EntryService) saveBlockDocument(ctx context.Context, entry *domain.Entry, doc *domain.BlockDocument, src stateSource, note revisionNote) (domain.BlockSaveReport, error) {
	return s.mutateBlocks(ctx, entry, "", src, note, func(cur *domain.BlockDocument) error {
		cur.Version = doc.Version
		cur.Blocks = doc.Blocks
		return nil
	})
}

// carryAuthoredState copies state from a document as its author sent it onto the
// normalized one, matching by block id and keeping only keys the normalized block
// still has. Used on creation and import, where there is no stored state to carry
// forward and the file is the only place ticks can come from.
func carryAuthoredState(doc, authored *domain.BlockDocument) {
	if doc == nil || authored == nil {
		return
	}
	byID := make(map[string]domain.Block, len(authored.Blocks))
	for _, b := range authored.Blocks {
		if b.ID != "" {
			byID[b.ID] = b
		}
	}
	for i, b := range doc.Blocks {
		src, ok := byID[b.ID]
		if !ok || src.Type != b.Type {
			continue
		}
		raw, _ := src.Data[blockStateKey].(map[string]any)
		if len(raw) == 0 {
			continue
		}
		encoded, err := json.Marshal(raw)
		if err != nil {
			continue
		}
		if pruned := pruneState(string(encoded), b.Data); pruned != "" {
			state := parseState(pruned)
			doc.Blocks[i].Data[blockStateKey] = state
		}
	}
}
