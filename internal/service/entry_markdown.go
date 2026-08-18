package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/butschster/mcp-research/internal/auth"
	"github.com/butschster/mcp-research/internal/domain"
)

// One document, as a file.
//
// Until now a document could only leave as part of a whole-research export —
// markdown, PDF, a vault, a JSON dump — and there was no way to take one and
// put it somewhere else. This is that, and it is deliberately small: the file
// lands in an editor we do not control, and guaranteeing what happens to it
// afterwards is a project. We offer the file.
//
// Two things it does NOT do, so nobody files them as bugs later:
//
//   - It does not rewrite `[[E3]]`. The vault turns that into
//     `[[E3 — Pricing model|E3]]` so Obsidian resolves it against a sibling
//     note; a loose file has no siblings, so the rewrite would point at a note
//     that exists nowhere. The reference is emitted exactly as written.
//   - It carries no provenance. Who wrote a document, in which session and at
//     which revision is working process, and it is the one thing this product
//     never lets out — a share link is refused it too.
//   - An html block is **named, not inlined**: `> **Revenue chart** —
//     interactive HTML, view in the web UI.` The vault writes the artifact
//     beside the note as a real file and links it; a loose file has no beside,
//     and a wall of markup is not markdown.

// MarkdownFile is one document rendered for download.
type MarkdownFile struct {
	// Filename is safe to put in a Content-Disposition header and on a disk.
	Filename string
	Content  string
}

// MarkdownExport renders a single entry as a markdown file with YAML front
// matter, the same front matter the Obsidian vault writes.
func (s *EntryService) MarkdownExport(ctx context.Context, entryID string) (*MarkdownFile, error) {
	// A share visitor is refused here, not only by the route being absent from
	// the shared sub-mux. The sub-mux is the boundary and stays the boundary —
	// but this is now a second renderer of a document's whole body, reachable by
	// id, and the next person mounting something under /api/shared/ should find
	// the refusal in the code rather than infer it from a test.
	if auth.ShareFromContext(ctx) != nil {
		return nil, ErrNotFound
	}

	entry, err := s.Get(ctx, entryID)
	if err != nil {
		return nil, err
	}

	research, err := s.researches.FindByID(ctx, entry.ResearchID)
	if err != nil {
		return nil, fmt.Errorf("find research: %w", err)
	}
	if research == nil {
		return nil, ErrNotFound
	}
	// The section is read for its declaration, not for its name alone: the
	// metadata keys a document carries are only meaningful in the order the
	// section declares them.
	section, err := s.sections.FindByID(ctx, entry.SectionID)
	if err != nil {
		return nil, fmt.Errorf("find section: %w", err)
	}

	fm := &frontmatter{}
	fm.add("code", entry.Code)
	fm.add("title", entry.Title)
	// `research` is the entire identity guarantee. A code means nothing outside
	// the research that issued it, so the file says which one it came from and
	// claims nothing more.
	fm.add("research", research.Code)
	if section != nil {
		name := section.DisplayName
		if name == "" {
			name = section.Name
		}
		fm.add("section", name)
	}
	fm.add("type", string(entry.Type))
	fm.add("status", string(entry.Status))
	fm.add("tags", entry.Tags)
	fm.add("created", stamp(entry.CreatedAt))
	fm.add("updated", stamp(entry.UpdatedAt))
	addMetadataFrontmatter(fm, section, entry)

	var b strings.Builder
	b.WriteString(fm.render())
	// No description. An entry's is derived from lines 2-5 of its own content
	// (`autoDescription`), so printing it above the body is the same sentences
	// twice — which is why the vault does not print it either, though it does
	// for a task and a roadmap, whose descriptions are written by hand.
	b.WriteString("\n" + entryMarkdownBody(entry, MarkdownOptions{}))

	return &MarkdownFile{
		Filename: markdownFilename(entry),
		Content:  b.String(),
	}, nil
}

// entryMarkdownBody renders a document's body for a markdown file.
//
// Shared by the vault note and the single-file download. It was written twice
// first — same guard, same branch, same "could not be read" sentence — and the
// compiler could not say so, because the two were methods on different
// receivers. That is the same failure that produced a duplicate
// contentDisposition one package over, one level down where nothing catches it.
//
// The options are the whole difference: the vault omits the mermaid live-editor
// link because Obsidian draws mermaid itself, and overrides HTMLBlock to write
// the artifact beside the note as a real file. A loose file takes the zero
// value, whose default names an html block rather than emitting it — a wall of
// markup is not markdown, and there is no beside to link to.
func entryMarkdownBody(entry *domain.Entry, opts MarkdownOptions) string {
	if entry.Content == "" {
		return ""
	}
	if entry.Type != domain.EntryBlocks {
		return strings.TrimRight(normalizeContent(entry.Content), "\n") + "\n"
	}
	doc, err := ParseStoredBlockDocument(entry.Content)
	if err != nil {
		return "*This entry holds a block document that could not be read.*\n"
	}
	return strings.TrimRight(BlockDocumentToMarkdownWith(doc, opts), "\n") + "\n"
}

// markdownFilename names the file after the code and the title.
//
// The code leads because it is what every link in this product is built from,
// and because a folder of downloads sorts by it into the order the research
// issued them.
func markdownFilename(entry *domain.Entry) string {
	title := sanitizeSegment(entry.Title, "entry")
	if entry.Code == "" {
		return title + ".md"
	}
	return sanitizeSegment(entry.Code, "entry") + " — " + title + ".md"
}
