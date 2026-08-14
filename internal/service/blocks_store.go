package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/butschster/mcp-research/internal/domain"
	"github.com/butschster/mcp-research/internal/storage"
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
func documentToRows(entry *domain.Entry, doc *domain.BlockDocument, prev []storage.BlockRow) ([]storage.BlockRow, domain.BlockSaveReport) {
	byID := make(map[string]storage.BlockRow, len(prev))
	for _, r := range prev {
		byID[r.BlockID] = r
	}
	byText := stateByItemText(prev)

	report := domain.BlockSaveReport{Blocks: len(doc.Blocks)}
	rows := make([]storage.BlockRow, 0, len(doc.Blocks))
	carriedText := map[string]bool{}

	for i, b := range doc.Blocks {
		data, incoming := splitState(b.Data)
		raw, err := json.Marshal(data)
		if err != nil {
			raw = []byte("{}")
		}

		state := incoming
		if old, ok := byID[b.ID]; ok {
			if old.Type == string(b.Type) && state == "" {
				state = old.State
			}
		} else {
			report.Reidentified++
			// The id is new. If this block's items carry text the previous
			// document had ticked, the human's intent survives the rename.
			if state == "" {
				if revived := reviveByText(data, byText, carriedText); revived != "" {
					state = revived
				}
			}
		}
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
func reviveByText(data map[string]any, byText, used map[string]bool) string {
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
		used[norm] = true
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
func (s *EntryService) LoadBlockDocument(ctx context.Context, entryID string) (*domain.BlockDocument, error) {
	rows, err := s.blocks.FindByEntry(ctx, nil, entryID)
	if err != nil {
		return nil, err
	}
	return rowsToDocument(rows), nil
}

// saveBlockDocument writes a document as rows and projects it back into
// entries.content, both inside one transaction. A half-written document is the
// worst thing this product could store, so the two never happen separately.
func (s *EntryService) saveBlockDocument(ctx context.Context, entry *domain.Entry, doc *domain.BlockDocument) (domain.BlockSaveReport, error) {
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
	rows, rep := documentToRows(entry, doc, prev)
	report = rep

	if err := s.blocks.ReplaceForEntry(ctx, tx, entry.ID, rows); err != nil {
		return report, err
	}

	// The projection is what every reader sees, so it is built from the rows
	// just written rather than from the document that produced them.
	projection, err := MarshalBlockDocument(rowsToDocument(rows))
	if err != nil {
		return report, err
	}
	entry.Content = projection
	if err := s.entries.UpdateTx(ctx, tx, entry); err != nil {
		return report, fmt.Errorf("update entry: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return report, fmt.Errorf("commit: %w", err)
	}
	return report, nil
}
