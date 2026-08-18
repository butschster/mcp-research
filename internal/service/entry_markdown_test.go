package service

import (
	"strings"
	"testing"

	"github.com/butschster/mcp-research/internal/domain"
)

// A document leaving as a file. The rules worth pinning are the two the issue
// says we deliberately do not solve, and the one thing the file does promise.

func TestMarkdownExport_CarriesFrontMatterAndTheDeclaredFields(t *testing.T) {
	k := newMetaKit(t)
	k.declare(t,
		specStage(),
		domain.FieldSpec{Key: "produces", Label: "Produces", Type: domain.FieldText, Repeated: true},
		domain.FieldSpec{Key: "owner", Label: "Owner", Type: domain.FieldText, Required: true},
	)
	entry := k.write(t, map[string]any{
		"stage":    "in-review",
		"produces": []any{"scanner-watchdog"},
	})

	file, err := k.entry.MarkdownExport(k.ctx, entry.ID)
	if err != nil {
		t.Fatalf("export: %v", err)
	}

	for _, want := range []string{
		"code: E1",
		"research: R1",
		"section: Specifications",
		"stage: in-review",
		"produces: [scanner-watchdog]",
		// A declared field nobody answered is present and null, not absent:
		// otherwise a search for the documents missing it finds nothing, and
		// those are exactly the documents worth finding.
		"owner: null",
	} {
		if !strings.Contains(file.Content, want) {
			t.Errorf("front matter is missing %q:\n%s", want, file.Content)
		}
	}
	if !strings.HasPrefix(file.Content, "---\n") {
		t.Fatalf("the file does not open with front matter:\n%s", file.Content)
	}
	if !strings.Contains(file.Filename, "E1") || !strings.HasSuffix(file.Filename, ".md") {
		t.Fatalf("filename = %q", file.Filename)
	}
}

func TestMarkdownExport_DoesNotRewriteReferencesAndCarriesNoProvenance(t *testing.T) {
	k := newMetaKit(t)
	entry, err := k.entry.Create(k.ctx, CreateEntryRequest{
		ResearchID: k.research.ID, SectionID: k.section.ID,
		Title: "SPEC-01", Content: "Follows from [[E9]] and the registry.",
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	file, err := k.entry.MarkdownExport(k.ctx, entry.ID)
	if err != nil {
		t.Fatalf("export: %v", err)
	}

	// The vault rewrites this into `[[E9 — Some title|E9]]` so Obsidian resolves
	// it against a sibling note. A loose file has no siblings, so the rewrite
	// would point at a note that exists nowhere.
	if !strings.Contains(file.Content, "[[E9]]") {
		t.Fatalf("the reference was rewritten:\n%s", file.Content)
	}
	// Who wrote a document and at which revision is working process. It is the
	// one thing this product never lets out.
	for _, forbidden := range []string{"author:", "revision:", "session:"} {
		if strings.Contains(file.Content, forbidden) {
			t.Errorf("the file carries %q:\n%s", forbidden, file.Content)
		}
	}
}

func TestMarkdownExport_RendersABlockDocumentAsMarkdown(t *testing.T) {
	k := newMetaKit(t)
	entry, err := k.entry.Create(k.ctx, CreateEntryRequest{
		ResearchID: k.research.ID, SectionID: k.section.ID,
		Type:  domain.EntryBlocks,
		Title: "Blocks", Content: `{"version":1,"blocks":[{"type":"heading","data":{"text":"Payload","level":2}},{"type":"paragraph","data":{"text":"One line."}}]}`,
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	file, err := k.entry.MarkdownExport(k.ctx, entry.ID)
	if err != nil {
		t.Fatalf("export: %v", err)
	}
	if strings.Contains(file.Content, `"blocks"`) {
		t.Fatalf("the stored JSON leaked into the file:\n%s", file.Content)
	}
	if !strings.Contains(file.Content, "## Payload") || !strings.Contains(file.Content, "One line.") {
		t.Fatalf("the block document did not render:\n%s", file.Content)
	}
}
