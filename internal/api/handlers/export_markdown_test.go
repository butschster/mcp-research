package handlers

import (
	"strings"
	"testing"

	"github.com/butschster/mcp-research/internal/domain"
)

const artifactDoc = `<!doctype html>
<html><head><title>Chart</title><style>body{color:red}</style></head>
<body><div id="chart"></div><script>document.title = "x"</script></body></html>`

const blockDoc = `{"version":1,"blocks":[
	{"type":"heading","data":{"level":2,"text":"Findings"}},
	{"type":"paragraph","data":{"text":"Body with [[E2]]."}},
	{"type":"html","data":{"html":"<html><body>chart</body></html>","title":"Live chart"}}
]}`

func markdownFor(entries ...*domain.Entry) string {
	research := &domain.Research{ID: "r1", Name: "Research"}
	section := &domain.Section{ID: "s1", Name: "main"}
	return buildMarkdown(
		research,
		[]*domain.Section{section},
		map[string][]*domain.Entry{section.ID: entries},
		nil,
		nil,
	)
}

// An HTML document has no place in a markdown export: it is not markdown, it is
// not readable, and inlined it would leak its <style> and <script> into whatever
// renders the file. It is named instead.
func TestBuildMarkdown_HTMLIsNamedNotInlined(t *testing.T) {
	md := markdownFor(&domain.Entry{
		ID: "e1", Code: "E1", SectionID: "s1",
		Type: domain.EntryArtifact, Title: "Chart",
		Content: artifactDoc, Status: domain.EntryActive,
	})

	for _, leak := range []string{"<!doctype html>", "<script>", "<style>", "```html"} {
		if strings.Contains(md, leak) {
			t.Errorf("markdown export contains %q — the document should be named, not emitted:\n%s", leak, md)
		}
	}
	if !strings.Contains(md, "view it in the web UI") {
		t.Errorf("no note pointing at the web UI:\n%s", md)
	}
}

func TestBuildMarkdown_MarkdownEntryIsInlined(t *testing.T) {
	md := markdownFor(&domain.Entry{
		ID: "e1", Code: "E1", SectionID: "s1",
		Type: domain.EntryMarkdown, Title: "Note",
		Content: "# Heading\n\nbody", Status: domain.EntryActive,
	})

	if !strings.Contains(md, "# Heading\n\nbody") {
		t.Errorf("markdown content not inlined:\n%s", md)
	}
	if strings.Contains(md, "```html") {
		t.Error("markdown entry should not be fenced as html")
	}
}

// A block document becomes real markdown: prose, headings and lists survive, the
// stored JSON never appears, and the html block inside is named like a standalone
// artifact would be.
func TestBuildMarkdown_BlockDocumentIsSerialized(t *testing.T) {
	md := markdownFor(&domain.Entry{
		ID: "e1", Code: "E1", SectionID: "s1",
		Type: domain.EntryBlocks, Title: "Findings",
		Content: blockDoc, Status: domain.EntryActive,
	})

	for _, want := range []string{"## Findings", "Body with [[E2]].", "Live chart"} {
		if !strings.Contains(md, want) {
			t.Errorf("markdown is missing %q:\n%s", want, md)
		}
	}
	for _, leak := range []string{`"type"`, `"blocks"`, `{"version"`, "<html>"} {
		if strings.Contains(md, leak) {
			t.Errorf("markdown export leaks %q:\n%s", leak, md)
		}
	}
}

func TestBuildMarkdown_UnreadableBlockDocumentSaysSo(t *testing.T) {
	md := markdownFor(&domain.Entry{
		ID: "e1", Code: "E1", SectionID: "s1",
		Type: domain.EntryBlocks, Title: "Broken",
		Content: "{not json", Status: domain.EntryActive,
	})

	if strings.Contains(md, "{not json") {
		t.Errorf("the unparseable document was emitted verbatim:\n%s", md)
	}
	if !strings.Contains(md, "could not be read") {
		t.Errorf("no note explaining the gap:\n%s", md)
	}
}

func TestBuildSessionMarkdown_HTMLIsNamedNotInlined(t *testing.T) {
	md := buildSessionMarkdown(
		&domain.Research{ID: "r1", Name: "Research"},
		&domain.Session{ID: "ss1", Code: "SS1", Title: "Session"},
		nil,
		[]*domain.Entry{{
			ID: "e1", Code: "E1", SectionID: "s1",
			Type: domain.EntryArtifact, Title: "Chart",
			Content: artifactDoc, Status: domain.EntryActive,
		}},
		map[string]string{"s1": "main"},
	)

	if strings.Contains(md, "<!doctype html>") || strings.Contains(md, "```html") {
		t.Errorf("session export inlined the document:\n%s", md)
	}
	if !strings.Contains(md, "view it in the web UI") {
		t.Errorf("no note pointing at the web UI:\n%s", md)
	}
}

func TestBuildSessionMarkdown_BlockDocumentIsSerialized(t *testing.T) {
	md := buildSessionMarkdown(
		&domain.Research{ID: "r1", Name: "Research"},
		&domain.Session{ID: "ss1", Code: "SS1", Title: "Session"},
		nil,
		[]*domain.Entry{{
			ID: "e1", Code: "E1", SectionID: "s1",
			Type: domain.EntryBlocks, Title: "Findings",
			Content: blockDoc, Status: domain.EntryActive,
		}},
		map[string]string{"s1": "main"},
	)

	if !strings.Contains(md, "## Findings") {
		t.Errorf("block document not serialized:\n%s", md)
	}
	if strings.Contains(md, `"type"`) {
		t.Errorf("stored JSON leaked into the session export:\n%s", md)
	}
}

// exportEntries supplies content_markdown so the export pages need no knowledge
// of the block format.
func TestExportEntries_SuppliesContentMarkdown(t *testing.T) {
	out := exportEntries([]*domain.Entry{
		{ID: "e1", Type: domain.EntryBlocks, Content: blockDoc},
		{ID: "e2", Type: domain.EntryMarkdown, Content: "# plain"},
	})

	if len(out) != 2 {
		t.Fatalf("got %d entries, want 2", len(out))
	}
	if !strings.Contains(out[0].ContentMarkdown, "## Findings") {
		t.Errorf("blocks entry has no rendering: %q", out[0].ContentMarkdown)
	}
	if out[1].ContentMarkdown != "" {
		t.Errorf("markdown entry should not carry a rendering, got %q", out[1].ContentMarkdown)
	}
}
