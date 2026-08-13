package handlers

import (
	"strings"
	"testing"

	"github.com/butschster/mcp-research/internal/domain"
)

const artifactDoc = `<!doctype html>
<html><head><title>Chart</title><style>body{color:red}</style></head>
<body><div id="chart"></div><script>document.title = "x"</script></body></html>`

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

func TestBuildMarkdown_ArtifactIsFenced(t *testing.T) {
	md := markdownFor(&domain.Entry{
		ID: "e1", Code: "E1", SectionID: "s1",
		Type: domain.EntryArtifact, Title: "Chart",
		Content: artifactDoc, Status: domain.EntryActive,
	})

	if strings.Contains(md, "\n\n<!doctype html>") {
		t.Error("artifact HTML is inlined into the markdown document")
	}
	if !strings.Contains(md, "```html\n"+artifactDoc+"\n```") {
		t.Errorf("artifact not wrapped in an html fence:\n%s", md)
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

func TestBuildMarkdown_ArtifactWithOwnFence(t *testing.T) {
	content := "<pre>```\ncode\n```</pre>"
	md := markdownFor(&domain.Entry{
		ID: "e1", Code: "E1", SectionID: "s1",
		Type: domain.EntryArtifact, Title: "Fenced",
		Content: content, Status: domain.EntryActive,
	})

	if !strings.Contains(md, "````html\n"+content+"\n````") {
		t.Errorf("fence not widened past the document's own backticks:\n%s", md)
	}
}

func TestBuildSessionMarkdown_ArtifactIsFenced(t *testing.T) {
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

	if !strings.Contains(md, "```html\n"+artifactDoc+"\n```") {
		t.Errorf("artifact not wrapped in an html fence:\n%s", md)
	}
}
