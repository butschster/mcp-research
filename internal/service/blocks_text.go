package service

import (
	"fmt"
	"strings"

	"github.com/butschster/mcp-research/internal/domain"
)

// Text projections of a block document.
//
// Everything that used to read an entry's markdown as one string needs these:
// cross-reference extraction, external links, search, the auto title and
// description, and the markdown exports. Running any of them over the raw JSON
// would index keys and quote-escaped fragments instead of prose.

// BlockPlainText returns the document's prose, one block per line, for indexing
// and for [[E3]] extraction. Code and html bodies are deliberately excluded:
// they are not prose, and a snippet mentioning [[E3]] is not a reference.
func BlockPlainText(doc *domain.BlockDocument) string {
	if doc == nil {
		return ""
	}
	var b strings.Builder
	for _, blk := range doc.Blocks {
		for _, line := range blockTextLines(blk) {
			if line == "" {
				continue
			}
			b.WriteString(line)
			b.WriteString("\n")
		}
	}
	return strings.TrimRight(b.String(), "\n")
}

// blockTextLines is the per-type text contribution. A new text-bearing block
// type must be added here or its prose silently stops being searchable and its
// cross-references stop resolving.
func blockTextLines(blk domain.Block) []string {
	d := blk.Data
	switch blk.Type {
	case domain.BlockParagraph, domain.BlockQuote:
		return []string{str(d, "text"), str(d, "cite")}
	case domain.BlockHeading:
		return []string{str(d, "text")}
	case domain.BlockCallout:
		return []string{str(d, "title"), str(d, "text")}
	case domain.BlockList:
		items, _ := d["items"].([]any)
		out := make([]string, 0, len(items))
		for _, it := range items {
			if s, ok := it.(string); ok {
				out = append(out, s)
			}
		}
		return out
	case domain.BlockTable:
		rows, _ := d["rows"].([]any)
		var out []string
		for _, r := range rows {
			cells, _ := r.([]any)
			parts := make([]string, 0, len(cells))
			for _, c := range cells {
				if s, ok := c.(string); ok {
					parts = append(parts, s)
				}
			}
			out = append(out, strings.Join(parts, " "))
		}
		return out
	case domain.BlockChecklist:
		out := []string{str(d, "title")}
		for _, text := range checklistItems(d) {
			out = append(out, text.Text)
		}
		return out
	case domain.BlockImage:
		return []string{str(d, "alt"), str(d, "caption")}
	case domain.BlockHTML:
		// Only the author-written framing is prose; the document body is not.
		return []string{str(d, "title"), str(d, "caption")}
	case domain.BlockMermaid:
		// Same rule as code and html: the source is notation, the caption is the
		// author writing. Node labels look like prose but arrive wrapped in
		// syntax, and indexing them would put `-->` and `subgraph` in the index.
		return []string{str(d, "caption")}
	case domain.BlockCode, domain.BlockDivider:
		return nil
	}
	return nil
}

// BlockDocumentTitle derives an entry title: the first heading, else the first
// paragraph's opening sentence. Deriving it the markdown way would return a
// fragment of JSON.
func BlockDocumentTitle(doc *domain.BlockDocument) string {
	if doc == nil {
		return ""
	}
	for _, blk := range doc.Blocks {
		if blk.Type == domain.BlockHeading {
			if t := str(blk.Data, "text"); t != "" {
				return clampStr(t, 200)
			}
		}
	}
	for _, blk := range doc.Blocks {
		switch blk.Type {
		case domain.BlockParagraph:
			if t := firstSentence(str(blk.Data, "text")); t != "" {
				return t
			}
		case domain.BlockHTML:
			if t := str(blk.Data, "title"); t != "" {
				return clampStr(t, 200)
			}
		}
	}
	return ""
}

// BlockDocumentDescription derives a short summary: the first paragraph that is
// not already the title.
func BlockDocumentDescription(doc *domain.BlockDocument, title string) string {
	if doc == nil {
		return ""
	}
	for _, blk := range doc.Blocks {
		if blk.Type != domain.BlockParagraph {
			continue
		}
		text := str(blk.Data, "text")
		if text == "" || strings.HasPrefix(text, title) && title != "" {
			continue
		}
		return clampStr(text, 500)
	}
	// A document with no prose — a lone html block, say — still has the author's
	// framing to fall back on.
	for _, blk := range doc.Blocks {
		if blk.Type == domain.BlockHTML {
			if c := str(blk.Data, "caption"); c != "" {
				return clampStr(c, 500)
			}
		}
	}
	return ""
}

func firstSentence(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	if i := strings.IndexAny(s, ".!?\n"); i > 0 {
		s = s[:i]
	}
	return clampStr(s, 200)
}

// BlockDocumentToMarkdown serializes a document so the research and session
// exports produce a readable file instead of raw JSON.
func BlockDocumentToMarkdown(doc *domain.BlockDocument) string {
	if doc == nil {
		return ""
	}
	var b strings.Builder
	for _, blk := range doc.Blocks {
		d := blk.Data
		switch blk.Type {
		case domain.BlockHeading:
			level := intOr(d, "level", 2)
			b.WriteString(strings.Repeat("#", level) + " " + str(d, "text") + "\n\n")

		case domain.BlockParagraph:
			b.WriteString(str(d, "text") + "\n\n")

		case domain.BlockList:
			items, _ := d["items"].([]any)
			ordered := str(d, "style") == domain.ListOrdered
			for i, it := range items {
				s, _ := it.(string)
				if ordered {
					b.WriteString(fmt.Sprintf("%d. %s\n", i+1, s))
				} else {
					b.WriteString("- " + s + "\n")
				}
			}
			b.WriteString("\n")

		case domain.BlockTable:
			rows, _ := d["rows"].([]any)
			header := boolOr(d, "header", true)
			for i, r := range rows {
				cells, _ := r.([]any)
				parts := make([]string, 0, len(cells))
				for _, c := range cells {
					s, _ := c.(string)
					// A pipe inside a cell would break the table shape.
					parts = append(parts, strings.ReplaceAll(s, "|", "\\|"))
				}
				b.WriteString("| " + strings.Join(parts, " | ") + " |\n")
				if i == 0 && header {
					b.WriteString("|" + strings.Repeat(" --- |", len(parts)) + "\n")
				}
			}
			b.WriteString("\n")

		case domain.BlockQuote:
			for _, line := range strings.Split(str(d, "text"), "\n") {
				b.WriteString("> " + line + "\n")
			}
			if cite := str(d, "cite"); cite != "" {
				b.WriteString(">\n> — " + cite + "\n")
			}
			b.WriteString("\n")

		case domain.BlockCode:
			b.WriteString("```" + str(d, "language") + "\n" + str(d, "code") + "\n```\n\n")

		case domain.BlockChecklist:
			// GitHub task list syntax: the export is a working checklist, and a
			// reader who has never seen this app still sees what is done.
			if t := str(d, "title"); t != "" {
				b.WriteString("**" + t + "**\n\n")
			}
			for _, item := range checklistItems(d) {
				mark := " "
				if item.Checked {
					mark = "x"
				}
				b.WriteString("- [" + mark + "] " + item.Text + "\n")
			}
			b.WriteString("\n")

		case domain.BlockMermaid:
			// A mermaid fence is markdown that renders: GitHub and this app both
			// draw it, so the export loses nothing. The link below it is for every
			// reader whose viewer does not — the diagram rides in the URL fragment,
			// so nothing is uploaded by following it.
			code := str(d, "code")
			b.WriteString("```mermaid\n" + code + "\n```\n")
			if cap := str(d, "caption"); cap != "" {
				b.WriteString("*" + cap + "*\n")
			}
			b.WriteString(mermaidLiveLine(code))
			b.WriteString("\n")

		case domain.BlockCallout:
			// Markdown has no callout; a blockquote with a label keeps the intent.
			label := strings.ToUpper(str(d, "variant"))
			if t := str(d, "title"); t != "" {
				label += " — " + t
			}
			b.WriteString("> **" + label + "**\n>\n")
			for _, line := range strings.Split(str(d, "text"), "\n") {
				b.WriteString("> " + line + "\n")
			}
			b.WriteString("\n")

		case domain.BlockDivider:
			b.WriteString("---\n\n")

		case domain.BlockImage:
			alt := str(d, "alt")
			b.WriteString("![" + alt + "](" + str(d, "url") + ")\n")
			if cap := str(d, "caption"); cap != "" {
				b.WriteString("*" + cap + "*\n")
			}
			b.WriteString("\n")

		case domain.BlockHTML:
			// An HTML document cannot become markdown. Name it and keep the
			// author's framing so the export reads as a deliberate omission
			// rather than a hole.
			title := str(d, "title")
			if title == "" {
				title = "HTML block"
			}
			b.WriteString("> **" + title + "** — interactive HTML, view in the web UI.\n")
			if cap := str(d, "caption"); cap != "" {
				b.WriteString(">\n> " + cap + "\n")
			}
			b.WriteString("\n")
		}
	}
	return strings.TrimRight(b.String(), "\n") + "\n"
}

// ChecklistItem is one line of a checklist block, with the tick resolved from
// the block's state map.
type ChecklistItem struct {
	Key     string
	Text    string
	Checked bool
}

// checklistItems pairs items with their state, which is stored beside them
// rather than inside them.
func checklistItems(d map[string]any) []ChecklistItem {
	state, _ := d[blockStateKey].(map[string]any)
	items, _ := d["items"].([]any)
	out := make([]ChecklistItem, 0, len(items))
	for _, it := range items {
		m, ok := it.(map[string]any)
		if !ok {
			continue
		}
		key, _ := m["key"].(string)
		text, _ := m["text"].(string)
		checked := false
		if state != nil {
			checked, _ = state[key].(bool)
		}
		out = append(out, ChecklistItem{Key: key, Text: text, Checked: checked})
	}
	return out
}
