package service

import (
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/butschster/mcp-research/internal/domain"
)

// Where a mark is now.
//
// The rule this file implements, and the reason it is not a one-line string
// search: the block is the address and the quote is the proof, and when they
// disagree the disagreement IS the finding. A mark whose block still exists but
// whose sentence is gone has just been told something — an agent rewrote the
// text somebody doubted — and reporting that as "not found" throws the signal
// away.
//
// Nothing here is stored. The document is rewritten by an agent between reads,
// so a persisted anchor state is a claim about a document that no longer exists.

// AnchorText is how much of the anchored block's text travels with a mark, so a
// reader of the queue can judge it without opening the document.
const AnchorText = 1000

// ResolveAnchors places every mark in the document as it stands now.
//
// Matches are CONSUMED: two marks on the same repeated sentence resolve to two
// different occurrences, never the same one. Without this, a document with a
// boilerplate line repeated in three sections would show three marks stacked on
// the first copy — the same bug reviveByText already had to solve for checklist
// ticks, and solved the same way.
//
// Order matters and is not incidental. Exact block matches are placed first, so
// a mark whose block is intact keeps its own sentence, and a mark searching the
// whole document cannot steal it.
func ResolveAnchors(anns []*domain.Annotation, doc *domain.BlockDocument, markdown string) {
	idx := newAnchorIndex(doc, markdown)

	// Pass 1 — the block is still there. These have first claim on their text.
	deferred := make([]*domain.Annotation, 0, len(anns))
	for _, a := range anns {
		if a == nil {
			continue
		}
		if a.BlockID == "" {
			deferred = append(deferred, a)
			continue
		}
		text, ok := idx.text(a.BlockID)
		if !ok {
			// The block is gone. Searching the document for it belongs to pass 2,
			// after intact blocks have taken their own sentences.
			deferred = append(deferred, a)
			continue
		}
		if _, found := idx.take(a.BlockID, a.Quote.Exact, true); found {
			a.Anchor = &domain.Anchor{
				State:      domain.AnchorAnchored,
				Strategy:   domain.AnchorByBlockID,
				Confidence: 1,
				BlockID:    a.BlockID,
				BlockIndex: idx.index[a.BlockID],
				BlockType:  idx.kind[a.BlockID],
				Text:       anchorPreview(text, a.Quote.Exact),
			}
			continue
		}
		// The address survived, the sentence did not. This is `drifted`, and it
		// is the state worth surfacing loudest: the paragraph somebody doubted
		// was rewritten in place.
		a.Anchor = &domain.Anchor{
			State:      domain.AnchorDrifted,
			Strategy:   domain.AnchorByBlockID,
			Confidence: 0,
			BlockID:    a.BlockID,
			BlockIndex: idx.index[a.BlockID],
			BlockType:  idx.kind[a.BlockID],
			Text:       anchorPreview(text, a.Quote.Exact),
		}
	}

	// Pass 2 — no usable block. Find the sentence anywhere it survived.
	for _, a := range deferred {
		blockID, text, conf, found := idx.search(a.Quote)
		if !found {
			a.Anchor = &domain.Anchor{
				State:      domain.AnchorOrphaned,
				Strategy:   domain.AnchorNotFound,
				Confidence: 0,
				// An orphan has no place in the document. -1 rather than 0 so a
				// list sorted by position does not file it under the first
				// paragraph, where it reads as a mark on text it never touched.
				BlockIndex: -1,
			}
			continue
		}
		// A markdown entry never had a block id, so finding its quote is the
		// normal, healthy case rather than a recovery. Saying `moved` there
		// would report drift that never happened — and the confidence goes to 1
		// for the same reason: the context check exists to grade a RECOVERED
		// placement, and there is nothing here to have recovered from. Leaving
		// it at 0.6 published a mark that was both `anchored` and doubtful,
		// which is not a state this vocabulary has.
		state := domain.AnchorMoved
		strategy := domain.AnchorByQuoteInDoc
		if a.BlockID == "" {
			state = domain.AnchorAnchored
			strategy = domain.AnchorByQuoteInBlock
			conf = 1
		}
		a.Anchor = &domain.Anchor{
			State:      state,
			Strategy:   strategy,
			Confidence: conf,
			BlockID:    blockID,
			BlockIndex: idx.index[blockID],
			BlockType:  idx.kind[blockID],
			Text:       anchorPreview(text, a.Quote.Exact),
		}
	}
}

// anchorIndex holds the document's text once, per block, and remembers which
// occurrences have already been claimed.
type anchorIndex struct {
	order []string
	texts map[string]string // block id -> text as written
	norm  map[string]string // block id -> normalized text, what matching uses
	// index and kind travel with a placement so a list of marks can be read in
	// the order of the document rather than the order they were made, and so a
	// mark on a table is not drawn like a mark on a paragraph.
	index map[string]int
	kind  map[string]string
	taken map[string]map[int]bool
}

// markdownAnchorBlock is the pseudo-block id for an entry that has no blocks.
// It never leaves this file: a markdown annotation stores an empty block id and
// reports one.
const markdownAnchorBlock = ""

func newAnchorIndex(doc *domain.BlockDocument, markdown string) *anchorIndex {
	idx := &anchorIndex{
		texts: map[string]string{},
		norm:  map[string]string{},
		index: map[string]int{},
		kind:  map[string]string{},
		taken: map[string]map[int]bool{},
	}
	if doc != nil {
		// The position recorded is the block's place in the DOCUMENT, not in the
		// annotatable subset: blocks that carry no prose are skipped for matching
		// but still occupy a place on the page, and reporting the subset index
		// would put a mark in the wrong half of the document.
		for i, blk := range doc.Blocks {
			if blk.ID == "" {
				continue
			}
			text := blockAnchorText(blk)
			if strings.TrimSpace(text) == "" {
				continue
			}
			idx.order = append(idx.order, blk.ID)
			idx.texts[blk.ID] = text
			idx.norm[blk.ID] = normalizeQuote(text)
			idx.index[blk.ID] = i
			idx.kind[blk.ID] = string(blk.Type)
		}
		return idx
	}
	if strings.TrimSpace(markdown) != "" {
		// Same rule as a block: the reader selects rendered text. A markdown
		// document carries more syntax than a block field — headings, list
		// bullets, fences — and this removes only the inline part, which is why
		// a mark on a markdown document is the less precise of the two and the
		// UI says so.
		text := StripInlineMarkdown(markdown)
		idx.order = append(idx.order, markdownAnchorBlock)
		idx.texts[markdownAnchorBlock] = text
		idx.norm[markdownAnchorBlock] = normalizeQuote(text)
	}
	return idx
}

func (i *anchorIndex) text(blockID string) (string, bool) {
	t, ok := i.texts[blockID]
	return t, ok
}

// take claims the first unclaimed occurrence of quote inside one block.
//
// When every occurrence is already claimed, the LAST one is shared rather than
// refused. Consumption exists so two marks on a sentence that repeats in a block
// land on different copies of it — not to punish two marks on the same sentence,
// which is an ordinary thing to do: "find me a source for this" and "and I do
// not believe it" are two marks on one claim. Refusing the second reported it as
// `drifted`, which told the reader the text had changed when nothing had.
// share says what to do when every occurrence in the block is already claimed:
// a mark that names this block may have the last one, a mark merely searching
// for its text may not — it should keep looking in blocks that still have a free
// copy, and only settle for a shared one when nothing anywhere is free.
func (i *anchorIndex) take(blockID, quote string, share bool) (int, bool) {
	hay, ok := i.norm[blockID]
	if !ok {
		return 0, false
	}
	needle := normalizeQuote(quote)
	if needle == "" {
		return 0, false
	}
	claimed := i.taken[blockID]
	from := 0
	last := -1
	for {
		rel := strings.Index(hay[from:], needle)
		if rel < 0 {
			if share && last >= 0 {
				return last, true
			}
			return 0, false
		}
		pos := from + rel
		last = pos
		if !claimed[pos] {
			if claimed == nil {
				claimed = map[int]bool{}
				i.taken[blockID] = claimed
			}
			claimed[pos] = true
			return pos, true
		}
		from = pos + 1
	}
}

// search looks for the quote across the whole document, in render order, and
// reports how much to trust the placement.
//
// Confidence is the surrounding context, not the match itself: an exact
// sentence found somewhere else in the document is still an exact sentence, but
// a sentence found with the words that used to precede and follow it is the
// same sentence, and those are different claims.
func (i *anchorIndex) search(q domain.Quote) (string, string, float64, bool) {
	needle := normalizeQuote(q.Exact)
	if needle == "" {
		return "", "", 0, false
	}
	// Free occurrences first, across every block, before settling for one that
	// another mark already holds. Sharing early would take a copy of the
	// sentence away from the mark whose block still exists.
	for _, share := range []bool{false, true} {
		for _, blockID := range i.order {
			pos, ok := i.take(blockID, needle, share)
			if !ok {
				continue
			}
			conf := 0.6
			if contextHolds(i.norm[blockID], pos, len(needle), q) {
				conf = 0.9
			}
			return blockID, i.texts[blockID], conf, true
		}
	}
	return "", "", 0, false
}

// contextHolds reports whether the words around the match are the ones recorded
// when the mark was made.
func contextHolds(hay string, pos, length int, q domain.Quote) bool {
	prefix := normalizeQuote(q.Prefix)
	suffix := normalizeQuote(q.Suffix)
	if prefix == "" && suffix == "" {
		// Nothing was recorded, so nothing is confirmed. Reporting "context
		// matched" here would inflate confidence on the strength of no evidence.
		return false
	}
	if prefix != "" && !strings.HasSuffix(strings.TrimSpace(hay[:pos]), strings.TrimSpace(prefix)) {
		return false
	}
	end := pos + length
	if suffix != "" && end <= len(hay) && !strings.HasPrefix(strings.TrimSpace(hay[end:]), strings.TrimSpace(suffix)) {
		return false
	}
	return true
}

// normalizeQuote is the comparison form: case and whitespace are not what a
// reader means by "the same sentence". It matches normalizeItemText, which does
// the same job for checklist ticks — one notion of sameness in this package,
// not two.
func normalizeQuote(s string) string {
	return strings.ToLower(strings.Join(strings.Fields(s), " "))
}

// anchorPreview is the text that travels with a mark, windowed around the
// quote.
//
// A paragraph fits whole and this changes nothing. A markdown document does
// not: it has no blocks, so the whole document is the "block", and taking its
// first thousand characters sent the agent the opening of the file — usually
// not containing the marked sentence at all, and the same thousand characters
// repeated for every mark in that document. The promise this field makes is
// "judge the mark without opening the entry", and only a window around the
// quote keeps it.
func anchorPreview(text, quote string) string {
	if len(text) <= AnchorText {
		return text
	}
	at := strings.Index(normalizeQuote(text), normalizeQuote(quote))
	if at < 0 {
		// Nothing to centre on — the drifted and orphaned cases. The opening is
		// as good an excerpt as any.
		return clampStr(text, AnchorText)
	}
	// Centred, then snapped back inside the string. Byte offsets from the
	// normalized copy are approximate on multi-byte text; clampStr is
	// rune-aware, so the edges stay valid either way.
	start := at - AnchorText/3
	if start < 0 {
		start = 0
	}
	for start > 0 && start < len(text) && !utf8.RuneStart(text[start]) {
		start--
	}
	if start >= len(text) {
		return clampStr(text, AnchorText)
	}
	out := clampStr(text[start:], AnchorText)
	if start > 0 {
		out = "…" + out
	}
	return out
}

// blockAnchorText is the text a mark may be attached to.
//
// It reuses the projection that search and cross-reference extraction already
// use, which means code, mermaid sources and html bodies contribute nothing —
// and that is the intended answer, not an oversight. An html block renders
// inside a sandboxed iframe where a browser selection cannot be observed at all,
// so offering to annotate one would be a gesture the product cannot honour.
func blockAnchorText(blk domain.Block) string {
	lines := blockTextLines(blk)
	kept := lines[:0]
	for _, l := range lines {
		if strings.TrimSpace(l) != "" {
			kept = append(kept, StripInlineMarkdown(l))
		}
	}
	return strings.Join(kept, "\n")
}

// Inline markdown, removed the way the page removes it.
//
// This is the half of the matching rule that is easy to miss. Block text is
// STORED as `The **p99** latency is 40 ms` and RENDERED as `The p99 latency is
// 40 ms`; a person selects the rendered form, because that is the only form
// there is on screen. Comparing their selection against the source meant every
// mark on a sentence containing bold, inline code, a link or a cross-reference
// was born `drifted` — the state that means "an agent rewrote this" — while
// nothing had happened at all.
//
// The rules mirror `renderInline` in the frontend, in its order, and the two
// have to move together. Anything that changes what a reader sees without
// changing what is stored belongs here.
var (
	reMDLink   = regexp.MustCompile(`\[([^\]]+)\]\([^)\s]+\)`)
	reMDCode   = regexp.MustCompile("`([^`]+)`")
	reMDBold   = regexp.MustCompile(`\*\*([^*]+)\*\*`)
	reMDItalic = regexp.MustCompile(`(^|[^*])\*([^*]+)\*([^*]|$)`)
	reMDRef    = regexp.MustCompile(`\[\[([^\]]+)\]\]`)
)

// StripInlineMarkdown returns the text a reader sees for one stored field.
func StripInlineMarkdown(s string) string {
	if s == "" {
		return ""
	}
	// A cross-reference renders as its own code — [[E3]] becomes a link reading
	// "E3" — so the brackets go and the text stays. Done first, so a reference
	// inside a link label is not eaten by the link rule.
	s = reMDRef.ReplaceAllString(s, "$1")
	s = reMDLink.ReplaceAllString(s, "$1")
	s = reMDCode.ReplaceAllString(s, "$1")
	s = reMDBold.ReplaceAllString(s, "$1")
	s = reMDItalic.ReplaceAllString(s, "$1$2$3")
	return s
}

// AnnotationReportFor compares anchors before and after a write and says what
// the write did to marks the writer never mentioned.
//
// The contract mirrors BlockSaveReport exactly: an agent that buried a contested
// sentence is told so in the call that buried it, rather than the loss being
// discovered by a person weeks later.
func AnnotationReportFor(was map[string]domain.AnchorState, after []*domain.Annotation) domain.AnnotationReport {
	var report domain.AnnotationReport
	for _, a := range after {
		if a == nil || a.Anchor == nil {
			continue
		}
		// Only a transition is news. A mark that was already orphaned before
		// this write is not something this write did, and reporting it every
		// time would train the reader to ignore the field.
		prev, seen := was[a.ID]
		if seen && prev == a.Anchor.State {
			continue
		}
		switch a.Anchor.State {
		case domain.AnchorDrifted:
			report.Drifted = append(report.Drifted, a.Code)
		case domain.AnchorOrphaned:
			report.Orphaned = append(report.Orphaned, a.Code)
		}
	}
	return report
}

// AnchorStates snapshots where marks sit, for comparison after a write.
func AnchorStates(anns []*domain.Annotation) map[string]domain.AnchorState {
	out := make(map[string]domain.AnchorState, len(anns))
	for _, a := range anns {
		if a != nil && a.Anchor != nil {
			out[a.ID] = a.Anchor.State
		}
	}
	return out
}
