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
	case domain.BlockTaskRef:
		// The author's own writing, and the codes so a search for T4 finds the
		// document that plans around it. The task TITLES are deliberately absent:
		// they live in `tasks`, they are searchable there, and a copy taken at
		// write time would go stale the moment somebody renames the task — the
		// projection is computed from the document alone and has no way to learn
		// that it did.
		out := []string{str(d, "note")}
		for _, ref := range taskRefCodes(d) {
			// A code goes in as a reference, so the block that IS a reference
			// files one: the task's page can then find the document that plans
			// around it, and so can the graph. A uuid stays bare — `[[uuid]]` is
			// not a syntax this product resolves, and filing it would only add an
			// unresolved row.
			if taskRefPattern.MatchString(ref) {
				out = append(out, "[["+ref+"]]")
				continue
			}
			out = append(out, ref)
		}
		return out
	case domain.BlockTranscript:
		out := []string{str(d, "title")}
		// The speaker leads the run it belongs to, and only the run — the same
		// grouping the renderer draws. Repeating the name on every line indexed a
		// string the page never shows, and a mark made across two consecutive
		// turns by one person was born drifted because the haystack had a "Peter"
		// between them that the reader's selection did not.
		//
		// It still answers "who said the thing about the gateway": the name sits
		// immediately before the lines it introduces.
		prev := ""
		for i, t := range transcriptTurns(d) {
			if t.Speaker != "" && (i == 0 || t.Speaker != prev) {
				out = append(out, t.Speaker)
			}
			if t.Speaker != "" {
				prev = t.Speaker
			}
			out = append(out, t.Text)
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
		case domain.BlockTranscript:
			// A document that is only a transcript still names itself. Without
			// this it would fall through to the untitled case, and "Untitled" is
			// what an entry gets when nobody could think of anything — not when
			// the author wrote the name at the top of the block.
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
	// A document with no prose — a lone html block, a lone transcript — still has
	// the author's framing to fall back on.
	for _, blk := range doc.Blocks {
		switch blk.Type {
		case domain.BlockHTML:
			if c := str(blk.Data, "caption"); c != "" {
				return clampStr(c, 500)
			}
		case domain.BlockTaskRef:
			if n := str(blk.Data, "note"); n != "" {
				return clampStr(n, 500)
			}
		case domain.BlockTranscript:
			// The opening line of a conversation is what it is about often
			// enough to beat showing nothing.
			if turns := transcriptTurns(blk.Data); len(turns) > 0 {
				return clampStr(turns[0].Text, 500)
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

// MarkdownOptions tunes the projection for one export target. The zero value is
// what the markdown and session exports have always produced.
type MarkdownOptions struct {
	// HTMLBlock renders an html block. Nil keeps the default — the block is
	// named rather than inlined, because a wall of markup is not markdown. The
	// Obsidian export overrides it: there the document is written beside the
	// note as a real file, so it can be linked instead of described.
	HTMLBlock func(blk domain.Block) string
	// OmitMermaidLink drops the live-editor line under a diagram. It exists for
	// readers that draw mermaid themselves, where the link is a kilobyte of
	// base64 in the middle of a note and buys nothing.
	OmitMermaidLink bool
	// Tasks resolves a task_ref block's references to the tasks themselves.
	//
	// Nil is a supported answer, not a missing dependency: a revision diff is a
	// snapshot of a document rather than of the board, and an export that a share
	// link forbids tasks to is handed no tasks to resolve from. Without it the
	// block still exports — as a list of references rather than as a checklist,
	// which is honest: nothing here knows whether T4 is done.
	Tasks TaskRefResolver
}

// TaskRefRow is one resolved row of a task_ref block.
type TaskRefRow struct {
	// Ref is the reference as the document wrote it — a code or a uuid.
	Ref    string
	Code   string
	Title  string
	Status string
	Done   bool
}

// TaskRefResolver turns the references in a task_ref block into rows.
//
// It is handed a list rather than asked one reference at a time so a caller can
// answer from a slice it already loaded, which is what keeps this out of the
// database: every export that resolves task refs had already read the research's
// tasks for its own sake, and — crucially — had already decided whether the
// caller is allowed to see them.
type TaskRefResolver func(refs []string) []TaskRefRow

// NewTaskRefResolver builds a resolver over tasks the caller already holds.
//
// A reference that matches nothing is DROPPED rather than rendered as a ghost:
// a deleted task is not an unfinished one, and a row that cannot be ticked is
// worse than no row. That is also what makes an empty task list safe — a share
// link with tasks switched off resolves nothing and the block degrades to its
// note and its codes, rather than leaking a title through an export.
func NewTaskRefResolver(tasks []*domain.Task) TaskRefResolver {
	byRef := make(map[string]*domain.Task, len(tasks)*2)
	for _, t := range tasks {
		if t == nil {
			continue
		}
		if t.Code != "" {
			byRef[strings.ToUpper(t.Code)] = t
		}
		byRef[strings.ToLower(t.ID)] = t
	}
	return func(refs []string) []TaskRefRow {
		out := make([]TaskRefRow, 0, len(refs))
		for _, ref := range refs {
			t := byRef[strings.ToUpper(ref)]
			if t == nil {
				t = byRef[strings.ToLower(ref)]
			}
			if t == nil {
				continue
			}
			out = append(out, TaskRefRow{
				Ref:    ref,
				Code:   t.Code,
				Title:  t.Title,
				Status: string(t.Status),
				Done:   t.Status == domain.TaskCompleted,
			})
		}
		return out
	}
}

// BlockDocumentToMarkdown serializes a document so the research and session
// exports produce a readable file instead of raw JSON.
func BlockDocumentToMarkdown(doc *domain.BlockDocument) string {
	return BlockDocumentToMarkdownWith(doc, MarkdownOptions{})
}

// BlockDocumentToMarkdownWith is BlockDocumentToMarkdown with the per-target
// overrides applied. One serializer, so a block type added to it appears in
// every export at once.
func BlockDocumentToMarkdownWith(doc *domain.BlockDocument, opts MarkdownOptions) string {
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

		case domain.BlockTaskRef:
			b.WriteString(taskRefMarkdown(d, opts.Tasks))

		case domain.BlockTranscript:
			// One paragraph per turn rather than one line: a turn may itself
			// contain a newline, and a hard line break would then split it in a
			// place the speaker did not.
			if t := str(d, "title"); t != "" {
				b.WriteString("**" + t + "**\n\n")
			}
			if n := intOr(d, truncatedKey, 0); n > 0 {
				b.WriteString(fmt.Sprintf("*%d further turns were not stored — this transcript is past the %d-turn cap.*\n\n", n, domain.MaxTranscriptTurns))
			}
			for _, turn := range transcriptTurns(d) {
				var lead string
				if turn.Speaker != "" {
					lead = "**" + turn.Speaker + "**"
				}
				if turn.Stamp != "" {
					if lead != "" {
						lead += " "
					}
					lead += "*(" + turn.Stamp + ")*"
				}
				if lead != "" {
					lead += ": "
				}
				b.WriteString(lead + turn.Text + "\n\n")
			}

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
			if !opts.OmitMermaidLink {
				b.WriteString(mermaidLiveLine(code))
			}
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
			if opts.HTMLBlock != nil {
				b.WriteString(opts.HTMLBlock(blk))
				continue
			}
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

// TranscriptTurn is one line of a transcript block.
type TranscriptTurn struct {
	Speaker string
	// Stamp is `ts` as the author wrote it. Named apart from the field because
	// `ts` says nothing at a call site, and because nothing here parses it.
	Stamp string
	Text  string
}

// transcriptTurns reads the turns out of a normalized transcript block.
func transcriptTurns(d map[string]any) []TranscriptTurn {
	raw, _ := d["turns"].([]any)
	out := make([]TranscriptTurn, 0, len(raw))
	for _, it := range raw {
		m, ok := it.(map[string]any)
		if !ok {
			continue
		}
		text, _ := m["text"].(string)
		if text == "" {
			continue
		}
		speaker, _ := m["speaker"].(string)
		ts, _ := m["ts"].(string)
		out = append(out, TranscriptTurn{Speaker: speaker, Stamp: ts, Text: text})
	}
	return out
}

// taskRefCodes reads the references out of a normalized task_ref block.
func taskRefCodes(d map[string]any) []string {
	raw, _ := d["tasks"].([]any)
	out := make([]string, 0, len(raw))
	for _, it := range raw {
		if s, ok := it.(string); ok && s != "" {
			out = append(out, s)
		}
	}
	return out
}

// taskRefMarkdown writes a task_ref block, with or without a resolver.
//
// Resolved, it is a GitHub task list: a reader who has never seen this app still
// sees what is done. Unresolved, it is a reference list — `[[T4]]` is the syntax
// this product links with everywhere else, so the vault turns it into a link to
// the task note and a plain reader sees the code it can look up. What it is NOT
// is a row of empty checkboxes: an export that says every task is open, having
// never asked, is worse than one that does not claim to know.
func taskRefMarkdown(d map[string]any, resolve TaskRefResolver) string {
	var b strings.Builder
	if note := str(d, "note"); note != "" {
		b.WriteString(note + "\n\n")
	}
	refs := taskRefCodes(d)

	if n := intOr(d, truncatedKey, 0); n > 0 {
		b.WriteString(fmt.Sprintf("*%d further references were not stored — this list is past the %d-task cap.*\n\n", n, domain.MaxTaskRefs))
	}

	if resolve == nil {
		for _, ref := range refs {
			b.WriteString("- [[" + ref + "]]\n")
		}
		b.WriteString("\n")
		return b.String()
	}

	rows := resolve(refs)
	if len(rows) == 0 {
		// Every reference is gone. Say so once rather than emitting an empty
		// bullet list, which reads as "no work" instead of "nothing left to show".
		b.WriteString("*No tasks — every reference in this list has been removed.*\n\n")
		return b.String()
	}
	if boolOr(d, "show_progress", true) {
		done := 0
		for _, r := range rows {
			if r.Done {
				done++
			}
		}
		// "done", not "closed": the status enum is `completed`, and `closed` is
		// already the annotation vocabulary. One word meaning two things across
		// two surfaces is how a reader learns to distrust both.
		b.WriteString(fmt.Sprintf("*%d of %d done*\n\n", done, len(rows)))
	}
	for _, r := range rows {
		mark := " "
		if r.Done {
			mark = "x"
		}
		label := r.Title
		if r.Code != "" {
			label = r.Code + " — " + r.Title
		}
		// A status a checkbox cannot express is named beside it. Ticked or not is
		// the whole of what `- [ ]` can say, and `blocked` said as `- [ ]` is a
		// lie by omission — it reads as "nobody has got to it".
		if r.Status != "" && r.Status != string(domain.TaskCompleted) && r.Status != string(domain.TaskPending) {
			label += " *(" + strings.ReplaceAll(r.Status, "_", " ") + ")*"
		}
		b.WriteString("- [" + mark + "] " + label + "\n")
	}
	b.WriteString("\n")
	return b.String()
}
