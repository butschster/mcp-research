package service

import (
	"bytes"
	"compress/zlib"
	"encoding/base64"
	"encoding/json"
	"io"
	"strings"
	"testing"

	"github.com/dovod-app/app/internal/domain"
)

const richDoc = `{"version":1,"blocks":[
  {"type":"heading","data":{"level":2,"text":"Findings"}},
  {"type":"paragraph","data":{"text":"See [[E3]] and **note** the drift."}},
  {"type":"callout","data":{"variant":"warning","title":"Watch out","text":"Scores drift."}},
  {"type":"list","data":{"style":"ordered","items":["first","second"]}},
  {"type":"table","data":{"header":true,"rows":[["Model","Speed"],["Llama","96"]]}},
  {"type":"quote","data":{"text":"Measure twice.","cite":"folk wisdom"}},
  {"type":"code","data":{"language":"go","code":"fmt.Println(1)"}},
  {"type":"mermaid","data":{"code":"flowchart TD\n  A --> B","caption":"How it flows"}},
  {"type":"divider","data":{}},
  {"type":"image","data":{"url":"/media/a.png","alt":"chart","caption":"Throughput"}},
  {"type":"html","data":{"html":"<html><body>x</body></html>","title":"Live chart","caption":"interactive"}}
]}`

func TestBlockDocumentToMarkdown(t *testing.T) {
	doc := normDoc(t, richDoc)
	md := BlockDocumentToMarkdown(doc)

	t.Run("no JSON leaks into the output", func(t *testing.T) {
		for _, needle := range []string{`"type"`, `"data"`, `{"`, `version`} {
			if strings.Contains(md, needle) {
				t.Errorf("markdown contains %q — a serialized document leaked:\n%s", needle, md)
			}
		}
	})

	t.Run("every block contributes", func(t *testing.T) {
		for _, want := range []string{
			"## Findings",
			"See [[E3]] and **note** the drift.",
			"WARNING — Watch out",
			"1. first",
			"2. second",
			"| Model | Speed |",
			"| --- | --- |",
			"> Measure twice.",
			"> — folk wisdom",
			"```go",
			"fmt.Println(1)",
			"```mermaid",
			"flowchart TD",
			"*How it flows*",
			"---",
			"![chart](/media/a.png)",
			"*Throughput*",
			"Live chart",
		} {
			if !strings.Contains(md, want) {
				t.Errorf("markdown is missing %q:\n%s", want, md)
			}
		}
	})

	t.Run("cross-references survive so the export stays linkable", func(t *testing.T) {
		if !strings.Contains(md, "[[E3]]") {
			t.Error("[[E3]] was lost")
		}
	})

	t.Run("an html block is named, not dumped", func(t *testing.T) {
		if strings.Contains(md, "<html>") || strings.Contains(md, "<body>") {
			t.Errorf("raw HTML reached the markdown export:\n%s", md)
		}
		if !strings.Contains(md, "view in the web UI") {
			t.Error("the html block should say why it is not inline")
		}
	})

	t.Run("a table cell containing a pipe cannot break the table", func(t *testing.T) {
		d := normDoc(t, `{"blocks":[{"type":"table","data":{"rows":[["a|b","c"]]}}]}`)
		out := BlockDocumentToMarkdown(d)
		if !strings.Contains(out, `a\|b`) {
			t.Errorf("pipe was not escaped:\n%s", out)
		}
	})
}

func TestBlockPlainText(t *testing.T) {
	doc := normDoc(t, richDoc)
	text := BlockPlainText(doc)

	t.Run("prose is present", func(t *testing.T) {
		for _, want := range []string{"Findings", "See [[E3]]", "Watch out", "first", "Llama", "Measure twice", "chart", "Live chart"} {
			if !strings.Contains(text, want) {
				t.Errorf("plain text is missing %q:\n%s", want, text)
			}
		}
	})

	t.Run("code and html bodies are excluded", func(t *testing.T) {
		// A snippet that mentions a reference is not a reference, and indexing a
		// whole HTML document would swamp the search.
		for _, unwanted := range []string{"fmt.Println", "<body>"} {
			if strings.Contains(text, unwanted) {
				t.Errorf("plain text contains %q, which is not prose:\n%s", unwanted, text)
			}
		}
	})

	t.Run("no JSON punctuation", func(t *testing.T) {
		if strings.Contains(text, `"type"`) || strings.Contains(text, `"data"`) {
			t.Errorf("plain text looks like JSON:\n%s", text)
		}
	})
}

func TestBlockDocumentTitleAndDescription(t *testing.T) {
	t.Run("title comes from the first heading", func(t *testing.T) {
		doc := normDoc(t, richDoc)
		if got := BlockDocumentTitle(doc); got != "Findings" {
			t.Errorf("title = %q, want Findings", got)
		}
	})

	t.Run("without a heading it falls back to the first sentence", func(t *testing.T) {
		doc := normDoc(t, `{"blocks":[{"type":"paragraph","data":{"text":"A short lead. And more."}}]}`)
		if got := BlockDocumentTitle(doc); got != "A short lead" {
			t.Errorf("title = %q, want the first sentence", got)
		}
	})

	t.Run("a lone html block is named by its own title", func(t *testing.T) {
		doc := ArtifactToBlockDocument(`<html><head><title>Infra map</title>
<meta name="description" content="Two loops."></head><body>x</body></html>`)
		if got := BlockDocumentTitle(doc); got != "Infra map" {
			t.Errorf("title = %q, want Infra map", got)
		}
		if got := BlockDocumentDescription(doc, "Infra map"); got != "Two loops." {
			t.Errorf("description = %q, want the meta description", got)
		}
	})

	t.Run("description skips the paragraph already used as the title", func(t *testing.T) {
		doc := normDoc(t, `{"blocks":[
			{"type":"paragraph","data":{"text":"Lead sentence. Rest."}},
			{"type":"paragraph","data":{"text":"The actual summary."}}
		]}`)
		title := BlockDocumentTitle(doc)
		if got := BlockDocumentDescription(doc, title); got != "The actual summary." {
			t.Errorf("description = %q, want the second paragraph", got)
		}
	})
}

func TestEntryIndexText(t *testing.T) {
	t.Run("markdown entries index their content as before", func(t *testing.T) {
		e := &domain.Entry{Type: domain.EntryMarkdown, Content: "# Title\n\nSee [[E3]]."}
		if got := EntryIndexText(e); got != e.Content {
			t.Errorf("got %q, want the content unchanged", got)
		}
	})

	t.Run("blocks entries index prose, not JSON", func(t *testing.T) {
		raw, err := MarshalBlockDocument(normDoc(t, richDoc))
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
		e := &domain.Entry{Type: domain.EntryBlocks, Content: raw}
		got := EntryIndexText(e)
		if !strings.Contains(got, "[[E3]]") {
			t.Error("a cross-reference inside block text would not be extracted")
		}
		if strings.Contains(got, `"type"`) {
			t.Errorf("JSON keys would be indexed:\n%s", got)
		}
	})

	t.Run("an unreadable document yields nothing rather than JSON", func(t *testing.T) {
		e := &domain.Entry{Type: domain.EntryBlocks, Content: "{not json"}
		if got := EntryIndexText(e); got != "" {
			t.Errorf("got %q, want empty", got)
		}
	})
}

// The markdown export is the one place a tick leaves this app, and it silently
// showed every box empty: the export ran stored content back through the input
// normalizer, which strips server-owned state by design.
func TestBlockDocumentToMarkdown_ChecklistState(t *testing.T) {
	stored := `{"version":1,"blocks":[{"id":"cccc3333","type":"checklist","data":{
		"title":"Before the migration",
		"items":[{"key":"k1","text":"Back up"},{"key":"k2","text":"Dry run"}],
		"state":{"k1":true}}}]}`

	doc, err := ParseStoredBlockDocument(stored)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	md := BlockDocumentToMarkdown(doc)

	if !strings.Contains(md, "- [x] Back up") {
		t.Errorf("a ticked item is not marked done:\n%s", md)
	}
	if !strings.Contains(md, "- [ ] Dry run") {
		t.Errorf("an unticked item is not marked open:\n%s", md)
	}

	// And the input path must still refuse to take state from an author.
	normalized, err := NormalizeBlockDocument(stored)
	if err != nil {
		t.Fatalf("normalize: %v", err)
	}
	if _, has := normalized.Blocks[0].Data["state"]; has {
		t.Error("normalization kept author-sent state; it must have one origin")
	}
}

// The link under an exported diagram has to be a real one: the whole document is
// in the fragment, so a mistake in the encoding is invisible until someone opens
// it and sees the editor's "could not load" diagram.
func TestMermaidLiveURL_RoundTrips(t *testing.T) {
	code := "flowchart TD\n    A[Правка] --> B{entry_patch}\n    B -->|set_state| C[одна строка]"
	url := MermaidLiveURL(code)

	const prefix = "https://mermaid.live/view#pako:"
	if !strings.HasPrefix(url, prefix) {
		t.Fatalf("url = %q, want the pako form", url)
	}
	payload := strings.TrimPrefix(url, prefix)
	if strings.ContainsAny(payload, "+/=") {
		t.Errorf("payload is not url-safe base64: %q", payload)
	}

	raw, err := base64.RawURLEncoding.DecodeString(payload)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	zr, err := zlib.NewReader(bytes.NewReader(raw))
	if err != nil {
		t.Fatalf("the payload is not zlib, which is what pako.inflate expects: %v", err)
	}
	defer zr.Close()
	decoded, err := io.ReadAll(zr)
	if err != nil {
		t.Fatalf("inflate: %v", err)
	}

	var state struct {
		Code          string `json:"code"`
		Mermaid       string `json:"mermaid"`
		UpdateDiagram bool   `json:"updateDiagram"`
	}
	if err := json.Unmarshal(decoded, &state); err != nil {
		t.Fatalf("state is not the editor's shape: %v", err)
	}
	if state.Code != code {
		t.Errorf("code did not survive the round trip:\n got: %q\nwant: %q", state.Code, code)
	}
	// The editor reads `mermaid` as a JSON string, not an object.
	if !strings.Contains(state.Mermaid, "theme") || !state.UpdateDiagram {
		t.Errorf("state = %+v, want a theme config and updateDiagram", state)
	}

	if MermaidLiveURL("") != "" {
		t.Error("an empty source must produce no link rather than a broken one")
	}
}

func TestBlockDocumentToMarkdown_MermaidCarriesTheLink(t *testing.T) {
	doc := normDoc(t, `{"blocks":[{"type":"mermaid","data":{"code":"flowchart LR\n  A --> B","caption":"How it flows"}}]}`)
	md := BlockDocumentToMarkdown(doc)

	for _, want := range []string{"```mermaid", "flowchart LR", "*How it flows*", "[Open this diagram in mermaid.live](https://mermaid.live/view#pako:"} {
		if !strings.Contains(md, want) {
			t.Errorf("markdown is missing %q:\n%s", want, md)
		}
	}
	// A code block that merely declares mermaid is still a listing, not a figure.
	plain := normDoc(t, `{"blocks":[{"type":"code","data":{"language":"mermaid","code":"flowchart LR\n  A --> B"}}]}`)
	if strings.Contains(BlockDocumentToMarkdown(plain), "mermaid.live") {
		t.Error("a code block gained a diagram link; only the mermaid block should")
	}
}
