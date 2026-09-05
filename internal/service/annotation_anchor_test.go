package service

import (
	"strings"
	"testing"

	"github.com/dovod-app/app/internal/domain"
)

func para(id, text string) domain.Block {
	return domain.Block{ID: id, Type: domain.BlockParagraph, Data: map[string]any{"text": text}}
}

func doc(blocks ...domain.Block) *domain.BlockDocument {
	return &domain.BlockDocument{Version: domain.BlockDocumentVersion, Blocks: blocks}
}

func mark(blockID, quote string) *domain.Annotation {
	return &domain.Annotation{
		ID: "a-" + blockID + "-" + quote, Code: "A1",
		BlockID: blockID, Quote: domain.Quote{Exact: quote},
	}
}

// The block is there and so is the sentence.
func TestAnchor_Anchored(t *testing.T) {
	a := mark("b1", "costs fall by 40 percent")
	ResolveAnchors([]*domain.Annotation{a}, doc(para("b1", "Our costs fall by 40 percent per year.")), "")

	if a.Anchor.State != domain.AnchorAnchored {
		t.Fatalf("state = %q, want anchored", a.Anchor.State)
	}
	if a.Anchor.BlockType != string(domain.BlockParagraph) {
		t.Errorf("block_type = %q, want paragraph", a.Anchor.BlockType)
	}
}

// The address survived, the sentence did not. This is the state the whole
// feature exists to surface: the paragraph somebody doubted was rewritten under
// the mark.
func TestAnchor_DriftedWhenTextChangesUnderIt(t *testing.T) {
	a := mark("b1", "costs fall by 40 percent")
	ResolveAnchors([]*domain.Annotation{a}, doc(para("b1", "Our costs fall somewhat.")), "")

	if a.Anchor.State != domain.AnchorDrifted {
		t.Fatalf("state = %q, want drifted", a.Anchor.State)
	}
	if a.Anchor.Text == "" {
		t.Error("drifted mark carries no current text, so nobody can see what replaced it")
	}
}

// The block id was minted afresh by a rewrite; the sentence survived elsewhere.
func TestAnchor_MovedCarriesConfidence(t *testing.T) {
	a := mark("gone", "costs fall by 40 percent")
	a.Quote.Prefix = "Our"
	a.Quote.Suffix = "per year."
	ResolveAnchors([]*domain.Annotation{a}, doc(
		para("b9", "Adoption is slow."),
		para("b7", "Our costs fall by 40 percent per year."),
	), "")

	if a.Anchor.State != domain.AnchorMoved {
		t.Fatalf("state = %q, want moved", a.Anchor.State)
	}
	if a.Anchor.BlockID != "b7" {
		t.Errorf("block = %q, want b7", a.Anchor.BlockID)
	}
	// Context that still holds is what separates "the same sentence" from "an
	// identical string somewhere else".
	if a.Anchor.Confidence < 0.9 {
		t.Errorf("confidence = %v, want high when the surrounding words match", a.Anchor.Confidence)
	}
}

// A quote found with none of its original context is still a find, but a weaker
// one, and the number has to say so.
func TestAnchor_MovedWithoutContextIsLessConfident(t *testing.T) {
	a := mark("gone", "costs fall by 40 percent")
	a.Quote.Prefix = "Historically"
	ResolveAnchors([]*domain.Annotation{a}, doc(para("b7", "Our costs fall by 40 percent per year.")), "")

	if a.Anchor.State != domain.AnchorMoved {
		t.Fatalf("state = %q, want moved", a.Anchor.State)
	}
	if a.Anchor.Confidence >= 0.9 {
		t.Errorf("confidence = %v, want lowered when the context no longer matches", a.Anchor.Confidence)
	}
}

// The sentence is gone from the document entirely. The mark is kept: only a
// person can say whether the rewrite answered the doubt or buried it.
func TestAnchor_Orphaned(t *testing.T) {
	a := mark("gone", "costs fall by 40 percent")
	ResolveAnchors([]*domain.Annotation{a}, doc(para("b7", "Adoption is slow.")), "")

	if a.Anchor.State != domain.AnchorOrphaned {
		t.Fatalf("state = %q, want orphaned", a.Anchor.State)
	}
	if a.Anchor.BlockIndex != -1 {
		t.Errorf("block_index = %d, want -1 so it does not sort under the first paragraph", a.Anchor.BlockIndex)
	}
}

// Two marks on the same repeated sentence must resolve to two different
// occurrences. Without consuming a match they stack on the first copy — the bug
// reviveByText already had to solve for checklist ticks.
func TestAnchor_MatchesAreConsumed(t *testing.T) {
	first := mark("b1", "see the appendix")
	first.ID = "first"
	second := mark("b2", "see the appendix")
	second.ID = "second"

	ResolveAnchors([]*domain.Annotation{first, second}, doc(
		para("b1", "For the method, see the appendix."),
		para("b2", "For the figures, see the appendix."),
	), "")

	if first.Anchor.BlockID != "b1" || second.Anchor.BlockID != "b2" {
		t.Fatalf("blocks = %q/%q, want b1/b2 — each mark keeps its own occurrence",
			first.Anchor.BlockID, second.Anchor.BlockID)
	}
}

// A mark whose block is intact is placed before one that is searching, so the
// searcher cannot steal a sentence that still has an owner.
func TestAnchor_IntactBlockWinsOverSearch(t *testing.T) {
	searching := mark("gone", "see the appendix")
	searching.ID = "searching"
	intact := mark("b1", "see the appendix")
	intact.ID = "intact"

	// Deliberately ordered with the searcher first: pass order, not slice order,
	// is what decides this.
	ResolveAnchors([]*domain.Annotation{searching, intact}, doc(
		para("b1", "For the method, see the appendix."),
		para("b2", "For the figures, see the appendix."),
	), "")

	if intact.Anchor.State != domain.AnchorAnchored || intact.Anchor.BlockID != "b1" {
		t.Fatalf("intact mark = %+v, want anchored to b1", intact.Anchor)
	}
	if searching.Anchor.BlockID == "b1" {
		t.Error("a searching mark took the sentence belonging to an intact one")
	}
}

// A markdown entry has no blocks. Finding its quote is the healthy case, not a
// recovery, and reporting `moved` there would claim drift that never happened.
func TestAnchor_MarkdownIsAnchoredNotMoved(t *testing.T) {
	a := &domain.Annotation{ID: "a1", Code: "A1", Quote: domain.Quote{Exact: "seat based"}}
	ResolveAnchors([]*domain.Annotation{a}, nil, "Pricing is seat based above 50 users.")

	if a.Anchor.State != domain.AnchorAnchored {
		t.Fatalf("state = %q, want anchored", a.Anchor.State)
	}
	if a.Anchor.Strategy != domain.AnchorByQuoteInBlock {
		t.Errorf("strategy = %q, want the quote strategy", a.Anchor.Strategy)
	}
}

// Case and spacing are not what a reader means by "the same sentence".
func TestAnchor_IgnoresCaseAndWhitespace(t *testing.T) {
	a := mark("b1", "Costs   fall by 40 PERCENT")
	ResolveAnchors([]*domain.Annotation{a}, doc(para("b1", "Our costs fall by 40 percent per year.")), "")

	if a.Anchor.State != domain.AnchorAnchored {
		t.Fatalf("state = %q, want anchored", a.Anchor.State)
	}
}

// An html block renders inside a sandboxed iframe, where a browser selection
// cannot be observed. Offering to anchor there would be a gesture the product
// cannot honour.
func TestAnchor_HTMLBodyIsNotAnchorable(t *testing.T) {
	a := mark("b1", "hidden in markup")
	ResolveAnchors([]*domain.Annotation{a}, doc(domain.Block{
		ID: "b1", Type: domain.BlockHTML,
		Data: map[string]any{"html": "<p>hidden in markup</p>", "title": "Chart"},
	}), "")

	if a.Anchor.State != domain.AnchorDrifted {
		t.Fatalf("state = %q — the html body must not be searchable text", a.Anchor.State)
	}
}

// Only a transition is news. A mark that was already orphaned before a write is
// not something that write did, and reporting it every time trains the reader to
// ignore the field.
func TestAnnotationReport_OnlyReportsTransitions(t *testing.T) {
	stale := mark("gone", "long deleted")
	stale.Code = "A1"
	fresh := mark("b1", "costs fall by 40 percent")
	fresh.Code = "A2"

	document := doc(para("b1", "Our costs fall by 40 percent per year."))
	ResolveAnchors([]*domain.Annotation{stale, fresh}, document, "")
	before := AnchorStates([]*domain.Annotation{stale, fresh})

	// The agent rewrites the paragraph the second mark sat on.
	rewritten := doc(para("b1", "Our costs are lower than they were."))
	ResolveAnchors([]*domain.Annotation{stale, fresh}, rewritten, "")

	report := AnnotationReportFor(before, []*domain.Annotation{stale, fresh})
	if len(report.Orphaned) != 0 {
		t.Errorf("orphaned = %v, want empty — A1 was already orphaned before this write", report.Orphaned)
	}
	if len(report.Drifted) != 1 || report.Drifted[0] != "A2" {
		t.Fatalf("drifted = %v, want [A2]", report.Drifted)
	}
}

// A reader selects what is on screen. Block text is stored as markdown and
// rendered without it, so comparing a selection against the source made every
// mark on a sentence containing bold, code, a link or a reference be born
// `drifted` — the state that means "an agent rewrote this" — while nothing had
// happened.
func TestAnchor_MatchesRenderedTextNotMarkdownSource(t *testing.T) {
	cases := []struct {
		name     string
		stored   string
		selected string
	}{
		{"bold", "The **p99** latency is 40 ms.", "The p99 latency is 40 ms."},
		{"link", "See [the docs](https://example.com/x) for details.", "See the docs for details."},
		{"code", "Set `max_conns` to one.", "Set max_conns to one."},
		{"crossref", "Measured in [[E7]] last week.", "Measured in E7 last week."},
		{"italic", "This is *probably* wrong.", "This is probably wrong."},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			a := mark("b1", tc.selected)
			ResolveAnchors([]*domain.Annotation{a}, doc(para("b1", tc.stored)), "")
			if a.Anchor.State != domain.AnchorAnchored {
				t.Fatalf("state = %q, want anchored — stored %q, selected %q",
					a.Anchor.State, tc.stored, tc.selected)
			}
		})
	}
}

// A markdown entry is stripped the same way.
func TestAnchor_MarkdownEntryStripsInlineMarkdown(t *testing.T) {
	a := &domain.Annotation{ID: "a1", Code: "A1", Quote: domain.Quote{Exact: "seat based"}}
	ResolveAnchors([]*domain.Annotation{a}, nil, "Pricing is **seat based** above 50 users.")

	if a.Anchor.State != domain.AnchorAnchored {
		t.Fatalf("state = %q, want anchored", a.Anchor.State)
	}
}

// Two marks on one sentence is an ordinary gesture — "find me a source" and "and
// I do not believe it". Consumption exists for a sentence that REPEATS, not to
// refuse the second reader.
func TestAnchor_TwoMarksOnTheSameSentenceBothAnchor(t *testing.T) {
	first := mark("b1", "costs fall by 40 percent")
	first.ID, first.Code = "first", "A1"
	second := mark("b1", "costs fall by 40 percent")
	second.ID, second.Code = "second", "A2"

	ResolveAnchors([]*domain.Annotation{first, second},
		doc(para("b1", "Our costs fall by 40 percent per year.")), "")

	for _, a := range []*domain.Annotation{first, second} {
		if a.Anchor.State != domain.AnchorAnchored {
			t.Errorf("%s = %q, want anchored — nothing was rewritten", a.Code, a.Anchor.State)
		}
	}
}

// But a sentence that genuinely repeats still gives each mark its own copy.
func TestAnchor_RepeatedSentenceStillSplits(t *testing.T) {
	first := mark("b1", "see the appendix")
	first.ID = "first"
	second := mark("b1", "see the appendix")
	second.ID = "second"

	ResolveAnchors([]*domain.Annotation{first, second},
		doc(para("b1", "For the method, see the appendix. For the figures, see the appendix.")), "")

	if first.Anchor.State != domain.AnchorAnchored || second.Anchor.State != domain.AnchorAnchored {
		t.Fatalf("states = %q/%q, want both anchored", first.Anchor.State, second.Anchor.State)
	}
}

// The preview exists so a mark can be judged without opening the document. On a
// markdown entry the whole document is the "block", and taking its opening sent
// the reader a thousand characters that usually did not contain the marked
// sentence — the same thousand for every mark in that document.
func TestAnchor_PreviewWindowsAroundTheQuote(t *testing.T) {
	filler := strings.Repeat("Background about the ingest path. ", 60)
	document := filler + "The p99 budget is 40 ms at 10k rps. " + filler

	a := &domain.Annotation{ID: "a1", Code: "A1", Quote: domain.Quote{Exact: "p99 budget is 40 ms"}}
	ResolveAnchors([]*domain.Annotation{a}, nil, document)

	if a.Anchor.State != domain.AnchorAnchored {
		t.Fatalf("state = %q, want anchored", a.Anchor.State)
	}
	if !strings.Contains(a.Anchor.Text, "p99 budget is 40 ms") {
		t.Fatalf("preview does not contain the marked sentence: %q", clampStr(a.Anchor.Text, 120))
	}
	if len([]rune(a.Anchor.Text)) > AnchorText+1 {
		t.Errorf("preview is %d runes, want at most %d", len([]rune(a.Anchor.Text)), AnchorText+1)
	}
}
