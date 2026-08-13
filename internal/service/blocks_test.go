package service

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/butschster/mcp-research/internal/domain"
)

func normDoc(t *testing.T, raw string) *domain.BlockDocument {
	t.Helper()
	doc, err := NormalizeBlockDocument(raw)
	if err != nil {
		t.Fatalf("NormalizeBlockDocument: %v", err)
	}
	return doc
}

func types(doc *domain.BlockDocument) []string {
	out := make([]string, 0, len(doc.Blocks))
	for _, b := range doc.Blocks {
		out = append(out, string(b.Type))
	}
	return out
}

func TestNormalizeBlockDocument_Envelope(t *testing.T) {
	t.Run("accepts the full envelope", func(t *testing.T) {
		doc := normDoc(t, `{"version":1,"blocks":[{"type":"paragraph","data":{"text":"hi"}}]}`)
		if doc.Version != domain.BlockDocumentVersion {
			t.Errorf("Version = %d, want %d", doc.Version, domain.BlockDocumentVersion)
		}
	})

	t.Run("accepts a bare blocks array", func(t *testing.T) {
		doc := normDoc(t, `[{"type":"paragraph","data":{"text":"hi"}}]`)
		if len(doc.Blocks) != 1 || doc.Version != 1 {
			t.Errorf("got %d blocks, version %d", len(doc.Blocks), doc.Version)
		}
	})

	t.Run("stamps the current version over whatever was sent", func(t *testing.T) {
		doc := normDoc(t, `{"version":99,"blocks":[{"type":"divider","data":{}}]}`)
		if doc.Version != 1 {
			t.Errorf("Version = %d, want 1", doc.Version)
		}
	})

	t.Run("rejects input that is not a document at all", func(t *testing.T) {
		for _, raw := range []string{``, `   `, `"just a string"`, `# markdown heading`, `{`} {
			if _, err := NormalizeBlockDocument(raw); err == nil {
				t.Errorf("NormalizeBlockDocument(%q) = nil error, want one", raw)
			}
		}
	})

	t.Run("rejects a document whose every block was dropped", func(t *testing.T) {
		_, err := NormalizeBlockDocument(`{"blocks":[{"type":"nope","data":{}},{"type":"paragraph","data":{"text":""}}]}`)
		if err == nil {
			t.Fatal("want an error, got nil")
		}
		if !strings.Contains(err.Error(), "no valid blocks") {
			t.Errorf("error = %q, want it to say every block was dropped", err)
		}
	})
}

func TestNormalizeBlockDocument_DropsRatherThanFails(t *testing.T) {
	// The whole point of the format: a bad payload degrades, it does not 500.
	doc := normDoc(t, `{"blocks":[
		{"type":"paragraph","data":{"text":"kept"}},
		{"type":"unknown_future_block","data":{"whatever":1}},
		{"type":"heading","data":{"text":""}},
		{"type":"image","data":{"url":"javascript:alert(1)"}},
		{"type":"list","data":{"items":[]}},
		{"type":"code","data":{"code":"   "}},
		{"type":"divider","data":{}}
	]}`)
	if got := types(doc); len(got) != 2 || got[0] != "paragraph" || got[1] != "divider" {
		t.Errorf("kept %v, want [paragraph divider]", got)
	}
}

func TestNormalizeBlockDocument_Blocks(t *testing.T) {
	t.Run("heading clamps the level to 2..4", func(t *testing.T) {
		doc := normDoc(t, `{"blocks":[
			{"type":"heading","data":{"level":1,"text":"one"}},
			{"type":"heading","data":{"level":3,"text":"three"}},
			{"type":"heading","data":{"level":9,"text":"nine"}}
		]}`)
		want := []int{2, 3, 2}
		for i, b := range doc.Blocks {
			if got := intOr(b.Data, "level", 0); got != want[i] {
				t.Errorf("block %d level = %d, want %d", i, got, want[i])
			}
		}
	})

	t.Run("list keeps order, drops blanks and non-strings", func(t *testing.T) {
		doc := normDoc(t, `{"blocks":[{"type":"list","data":{"style":"ordered","items":["a","",42,"b"]}}]}`)
		items, _ := doc.Blocks[0].Data["items"].([]any)
		if len(items) != 2 || items[0] != "a" || items[1] != "b" {
			t.Errorf("items = %v, want [a b]", items)
		}
		if got := str(doc.Blocks[0].Data, "style"); got != "ordered" {
			t.Errorf("style = %q, want ordered", got)
		}
	})

	t.Run("unknown list style falls back to unordered", func(t *testing.T) {
		doc := normDoc(t, `{"blocks":[{"type":"list","data":{"style":"checklist","items":["a"]}}]}`)
		if got := str(doc.Blocks[0].Data, "style"); got != "unordered" {
			t.Errorf("style = %q, want unordered", got)
		}
	})

	t.Run("callout falls back to info and keeps the title", func(t *testing.T) {
		doc := normDoc(t, `{"blocks":[{"type":"callout","data":{"variant":"nuclear","title":"Heads up","text":"body"}}]}`)
		d := doc.Blocks[0].Data
		if str(d, "variant") != "info" || str(d, "title") != "Heads up" {
			t.Errorf("data = %v, want info variant with the title kept", d)
		}
	})

	t.Run("image accepts http and relative urls, rejects the rest", func(t *testing.T) {
		for _, tc := range []struct {
			url string
			ok  bool
		}{
			{"https://example.com/a.png", true},
			{"http://example.com/a.png", true},
			{"/media/a.png", true},
			{"//evil.example.com/a.png", false},
			{"javascript:alert(1)", false},
			{"data:image/svg+xml;base64,AAA", false},
			{"", false},
		} {
			raw := `{"blocks":[{"type":"image","data":{"url":` + mustJSON(tc.url) + `}}]}`
			_, err := NormalizeBlockDocument(raw)
			if tc.ok && err != nil {
				t.Errorf("url %q was dropped, want kept", tc.url)
			}
			if !tc.ok && err == nil {
				t.Errorf("url %q was kept, want dropped", tc.url)
			}
		}
	})

	t.Run("code is stored verbatim, backslashes intact", func(t *testing.T) {
		// A snippet may legitimately contain a backslash escape. Expanding it the
		// way paragraph text is expanded would corrupt the code.
		code := `fmt.Println("C:\notes\tmp")`
		raw := `{"blocks":[{"type":"code","data":{"language":"Go","code":` + mustJSON(code) + `}}]}`
		d := normDoc(t, raw).Blocks[0].Data
		if got := str(d, "code"); got != code {
			t.Errorf("code = %q, want it byte for byte as %q", got, code)
		}
		if got := str(d, "language"); got != "go" {
			t.Errorf("language = %q, want go", got)
		}
	})

	t.Run("paragraph expands escaped newlines the way markdown content does", func(t *testing.T) {
		doc := normDoc(t, `{"blocks":[{"type":"paragraph","data":{"text":"one\\ntwo"}}]}`)
		if got := str(doc.Blocks[0].Data, "text"); got != "one\ntwo" {
			t.Errorf("text = %q, want a real newline", got)
		}
	})

	t.Run("heading collapses a newline instead of splitting", func(t *testing.T) {
		doc := normDoc(t, `{"blocks":[{"type":"heading","data":{"text":"two\nlines"}}]}`)
		if got := str(doc.Blocks[0].Data, "text"); got != "two lines" {
			t.Errorf("text = %q, want it on one line", got)
		}
	})

	t.Run("table drops non-array rows and caps columns", func(t *testing.T) {
		doc := normDoc(t, `{"blocks":[{"type":"table","data":{"rows":[["a","b"],"nope",["c"]]}}]}`)
		rows, _ := doc.Blocks[0].Data["rows"].([]any)
		if len(rows) != 2 {
			t.Errorf("rows = %v, want 2", rows)
		}
	})

	t.Run("an unknown field is not carried into storage", func(t *testing.T) {
		// Blocks have no width tier: every block sits in the reading column, and a
		// leftover width from an older payload must not be persisted.
		doc := normDoc(t, `{"blocks":[{"type":"table","data":{"rows":[["a"]],"width":"wide","bogus":1}}]}`)
		d := doc.Blocks[0].Data
		for _, key := range []string{"width", "bogus"} {
			if _, ok := d[key]; ok {
				t.Errorf("data carries %q: %v", key, d)
			}
		}
	})
}

func TestNormalizeBlockDocument_Caps(t *testing.T) {
	t.Run("caps the number of blocks", func(t *testing.T) {
		var sb strings.Builder
		sb.WriteString(`{"blocks":[`)
		for i := 0; i < domain.MaxBlocks+50; i++ {
			if i > 0 {
				sb.WriteString(",")
			}
			sb.WriteString(`{"type":"divider","data":{}}`)
		}
		sb.WriteString(`]}`)
		doc := normDoc(t, sb.String())
		if len(doc.Blocks) != domain.MaxBlocks {
			t.Errorf("kept %d blocks, want the cap of %d", len(doc.Blocks), domain.MaxBlocks)
		}
	})

	t.Run("clamps long text without breaking UTF-8", func(t *testing.T) {
		long := strings.Repeat("я", domain.MaxBlockText)
		doc := normDoc(t, `{"blocks":[{"type":"paragraph","data":{"text":`+mustJSON(long)+`}}]}`)
		got := str(doc.Blocks[0].Data, "text")
		if len(got) > domain.MaxBlockText {
			t.Errorf("len = %d, want <= %d", len(got), domain.MaxBlockText)
		}
		if strings.ContainsRune(got, '\uFFFD') {
			t.Error("clamp produced a replacement character — it cut mid-rune")
		}
	})
}

func TestArtifactToBlockDocument(t *testing.T) {
	html := `<!doctype html><html><head><title>T</title></head><body>x</body></html>`
	doc := ArtifactToBlockDocument(html)
	if len(doc.Blocks) != 1 || doc.Blocks[0].Type != domain.BlockHTML {
		t.Fatalf("blocks = %v, want a single html block", types(doc))
	}
	d := doc.Blocks[0].Data
	if str(d, "html") != html {
		t.Error("html body was altered")
	}

	// And it must survive the round trip through storage.
	raw, err := MarshalBlockDocument(doc)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	back := normDoc(t, raw)
	if str(back.Blocks[0].Data, "html") != html {
		t.Error("html body changed through marshal/normalize")
	}
}

func mustJSON(s string) string {
	b, err := json.Marshal(s)
	if err != nil {
		panic(err)
	}
	return string(b)
}
