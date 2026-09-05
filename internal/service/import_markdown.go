package service

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/dovod-app/app/internal/auth"
	"github.com/dovod-app/app/internal/domain"
	"gopkg.in/yaml.v3"
)

// One markdown file, into one section.
//
// The other half of the single-document download. Export and import are offered
// as a capability at the user's own risk: nothing here undertakes to make the
// round trip lossless, and two kinds of loss are treated very differently.
//
//   - A file that arrives DEGRADED is the user's problem. A [[E3]] naming an
//     entry from the research it was written in, a heading level nobody
//     intended, an Obsidian link rewrite that cannot be undone — we say what we
//     found and let them decide. We do not reverse `linkify`, do not remap a
//     cross-reference, and do not go looking for what a code meant somewhere
//     else.
//   - A file that would CORRUPT OUR RECORDS is ours. Front matter is text and
//     can assert anything. Accepting `session:` or an author out of a file would
//     let anybody fabricate provenance in a product whose entire history model
//     rests on it. That is refused, and it is not a fidelity question.
//
// **A validation error here is cheap and correct, and everywhere else in this
// product it is not.** Every other write refuses nothing, because the author is
// a model in the middle of an interview and a rejection destroys answers a
// person has already given. Here a person is standing over the file with an
// undo. The asymmetry is about who is holding the keyboard, not about the data —
// so if these two paths are ever "unified", the interview is what breaks.

// MaxImportFileBytes caps one dropped file. Generous next to any document
// somebody actually wrote by hand, and small enough that a mis-drop of a
// database dump is refused before it is read into memory.
//
// Aliased from domain, which is where the caps the client is told about live.
const MaxImportFileBytes = domain.MaxImportFileBytes

var (
	// ErrImportNotMarkdown is a file we will not read at all.
	ErrImportNotMarkdown = errors.New("only a .md or .markdown file can be imported")
	ErrImportTooLarge    = fmt.Errorf("that file is larger than %d KB", MaxImportFileBytes/1024)
	ErrImportEmpty       = errors.New("that file has no content once its front matter is removed")
	// ErrImportBadFrontMatter carries the parser's own message, which names the
	// line. A 500 on a stray tab is the wrong answer to somebody's typo.
	ErrImportBadFrontMatter = errors.New("the front matter could not be read")
	// ErrImportRejected is the commit refusing what the preview had already
	// shown to be wrong.
	ErrImportRejected = errors.New("this document cannot be imported as given")
)

// ImportNote is one key the parser read and did not use as written.
type ImportNote struct {
	Key   string `json:"key"`
	Value string `json:"value,omitempty"`
	// Reason is written for the person holding the file, not for a log.
	Reason string `json:"reason"`
}

// FieldNote is a remark about one of the four keys that become the entry
// itself.
//
// Those four have nowhere else to be reported: `metadata_report` covers
// declared fields, and `ignored`/`refused` are keys we decline by policy. A
// status the file spells wrong is none of those — it is a value we read, could
// not use, and quietly replaced with a default, which is exactly the kind of
// loss somebody has to be told about before they commit.
type FieldNote struct {
	Key   string `json:"key"`
	Value string `json:"value,omitempty"`
	// Applied says whether the value in the payload came from the file.
	Applied bool   `json:"applied"`
	Reason  string `json:"reason,omitempty"`
}

// UnresolvedRef is one dead reference and how often the document makes it.
type UnresolvedRef struct {
	Ref   string `json:"ref"`
	Count int    `json:"count"`
}

// Where the title came from, because the last of the three is the one somebody
// will want to change.
const (
	TitleFromFrontMatter = "frontmatter"
	TitleFromHeading     = "heading"
	TitleFromFilename    = "filename"
)

// MarkdownImport is a file read, with nothing written.
//
// It is deliberately not an entry. What comes back is a proposal the human
// accepts, edits or throws away, and the commit takes the fields rather than
// the file — so an edit made in the preview survives, and so the commit has no
// parameter through which a refused key could return.
type MarkdownImport struct {
	// Filename is echoed back because the title may have come from it.
	Filename    string                `json:"filename"`
	Title       string                `json:"title"`
	TitleSource string                `json:"title_source"`
	Description string                `json:"description,omitempty"`
	Status      domain.EntryStatus    `json:"status"`
	Tags        []string              `json:"tags"`
	Body        string                `json:"body"`
	BodyLines   int                   `json:"body_lines"`
	Metadata    map[string]any        `json:"metadata,omitempty"`
	Report      domain.MetadataReport `json:"metadata_report"`
	// Fields are remarks about title, description, status and tags.
	Fields []FieldNote `json:"fields,omitempty"`
	// Ignored was read and not used: the system owns those facts.
	Ignored []ImportNote `json:"ignored,omitempty"`
	// Refused is provenance a text file does not get to assert.
	Refused []ImportNote `json:"refused,omitempty"`
	// UnresolvedRefs are [[...]] codes that name nothing in this research, with
	// how many times each appears. They are listed and left exactly as written —
	// repairing a reference would mean guessing what a code meant in the
	// research it came from.
	UnresolvedRefs []UnresolvedRef `json:"unresolved_refs,omitempty"`
	// Warnings are true but not actionable by us: the file naming a different
	// research or section than the one it was dropped on.
	Warnings []string `json:"warnings,omitempty"`
}

// importIgnoredKeys are read, reported and dropped. Each one is a fact the
// system owns; a file asserting it is stale at best and a collision at worst.
var importIgnoredKeys = map[string]string{
	"code":     "a code is assigned by the research and is unique inside it, so a file cannot bring its own",
	"created":  "recorded when the document is written here",
	"updated":  "recorded when the document is next changed here",
	"research": "the research is the one you dropped the file into",
	"section":  "the section is the one you dropped the file into",
	"type":     "a file imports as markdown; block documents are written, not uploaded",
	"aliases":  "written by the Obsidian export so links resolve in a vault, and meaningless here",
}

// importRefusedKeys are provenance. Everything this product says about who did
// what, when and in which session is its own assertion, and the whole history
// model rests on that being true.
var importRefusedKeys = map[string]string{
	"session":     "which session produced a document is recorded here, never claimed by a file",
	"author":      "who wrote a document is recorded here, never claimed by a file",
	"author_kind": "who wrote a document is recorded here, never claimed by a file",
	"revision":    "revision history is this product's record of what happened, not a field",
	"revisions":   "revision history is this product's record of what happened, not a field",
	"history":     "revision history is this product's record of what happened, not a field",
}

// PreviewMarkdownImport reads a file and writes nothing.
func (s *EntryService) PreviewMarkdownImport(ctx context.Context, sectionID, filename string, data []byte) (*MarkdownImport, error) {
	section, err := s.importTarget(ctx, sectionID)
	if err != nil {
		return nil, err
	}
	return s.parseMarkdownFile(ctx, section, filename, data)
}

// ImportEntryRequest is the accepted preview, possibly edited.
//
// It carries no session, no author, no code and no timestamps — not because a
// caller is trusted not to send them, but because there is nowhere to put them.
// A refusal enforced by the shape of a struct cannot be forgotten by the next
// endpoint.
type ImportEntryRequest struct {
	SectionID   string
	Title       string
	Description string
	Status      domain.EntryStatus
	Tags        []string
	Body        string
	Metadata    map[string]any
}

// ImportMarkdown commits one document into a section.
func (s *EntryService) ImportMarkdown(ctx context.Context, req ImportEntryRequest) (*domain.Entry, error) {
	section, err := s.importTarget(ctx, req.SectionID)
	if err != nil {
		return nil, err
	}

	title := normalizeTitle(req.Title)
	if title == "" {
		return nil, fmt.Errorf("%w: it needs a title", ErrImportRejected)
	}
	if strings.TrimSpace(req.Body) == "" {
		return nil, fmt.Errorf("%w: it has no content", ErrImportRejected)
	}
	if req.Status != "" && !validEntryStatus(req.Status) {
		return nil, fmt.Errorf("%w: %q is not a status", ErrImportRejected, req.Status)
	}

	// Strict, unlike every other write. The preview already showed the person
	// exactly this, so anything still wrong at commit is either an edit they
	// made or a client sending what the preview told it not to.
	if _, report := domain.ValidateMetadata(section.FieldSpec, req.Metadata); len(report.UnknownKeys)+len(report.InvalidValues) > 0 {
		return nil, fmt.Errorf("%w: %s", ErrImportRejected, describeMetadataProblems(report))
	}

	// The active session still attaches, if there is one.
	//
	// What the issue refuses is the FILE asserting a session; this is the
	// system recording the one that was open when a person dropped it, exactly
	// as it does for every other entry, and `author_kind: import` on revision 1
	// keeps the two distinguishable. Leaving it out would need a new field on
	// the shared create request for one caller, and would make an imported
	// document the only kind of entry a session's change list cannot see.
	return s.Create(WithAuthor(ctx, domain.AuthorImport), CreateEntryRequest{
		ResearchID:  section.ResearchID,
		SectionID:   section.ID,
		Type:        domain.EntryMarkdown,
		Title:       title,
		Description: req.Description,
		Status:      req.Status,
		Tags:        req.Tags,
		Content:     req.Body,
		Metadata:    req.Metadata,
		// The body came out of a file, so its backslashes are backslashes.
		Verbatim: true,
	})
}

// importTarget resolves the section and checks the caller may write to the
// research that owns it.
//
// A share visitor is refused here as well as by the route being absent from the
// shared sub-mux: this is the only write in the product reachable by section id
// alone, and one day somebody will mount something under /api/shared/ that
// takes one.
func (s *EntryService) importTarget(ctx context.Context, sectionID string) (*domain.Section, error) {
	if auth.ShareFromContext(ctx) != nil {
		return nil, ErrNotFound
	}
	section, err := s.sections.FindByID(ctx, sectionID)
	if err != nil {
		return nil, fmt.Errorf("find section: %w", err)
	}
	if section == nil {
		return nil, ErrNotFound
	}
	if err := s.access.Write(ctx, section.ResearchID); err != nil {
		return nil, err
	}
	return section, nil
}

func validEntryStatus(st domain.EntryStatus) bool {
	switch st {
	case domain.EntryDraft, domain.EntryActive, domain.EntryCompleted, domain.EntryArchived:
		return true
	}
	return false
}

func describeMetadataProblems(r domain.MetadataReport) string {
	var parts []string
	for _, i := range r.UnknownKeys {
		parts = append(parts, fmt.Sprintf("%s is not a field this section records", i.Key))
	}
	for _, i := range r.InvalidValues {
		parts = append(parts, fmt.Sprintf("%s: %s", i.Key, i.Reason))
	}
	return strings.Join(parts, "; ")
}

// parseMarkdownFile is the whole parser: front matter, key classification,
// metadata validation and the reference scan.
func (s *EntryService) parseMarkdownFile(ctx context.Context, section *domain.Section, filename string, data []byte) (*MarkdownImport, error) {
	if len(data) > MaxImportFileBytes {
		return nil, ErrImportTooLarge
	}
	if !looksLikeMarkdown(filename) {
		return nil, ErrImportNotMarkdown
	}

	// sanitizeUTF8 and nothing else. normalizeContent expands a literal `\n`,
	// which is right for an MCP client sending escaped newlines inside JSON and
	// wrong for a file, where two characters are two characters — most often
	// inside a code block, where expanding them rewrites the code the document
	// is about.
	src := sanitizeUTF8(strings.ReplaceAll(string(data), "\r\n", "\n"))

	front, body, err := splitFrontMatter(src)
	if err != nil {
		return nil, err
	}

	out := &MarkdownImport{
		Filename: filename,
		Body:     strings.TrimSpace(body),
		Tags:     []string{},
		Status:   domain.EntryDraft,
	}
	if out.Body == "" {
		return nil, ErrImportEmpty
	}

	keys, err := parseFrontMatter(front)
	if err != nil {
		return nil, err
	}

	// Whatever is left after the three closed sets is a candidate metadata
	// value. Sorted, so a report reads the same twice.
	candidate := map[string]any{}
	names := make([]string, 0, len(keys))
	for k := range keys {
		names = append(names, k)
	}
	sort.Strings(names)

	for _, k := range names {
		v := keys[k]
		lower := strings.ToLower(k)

		if reason, refused := importRefusedKeys[lower]; refused {
			out.Refused = append(out.Refused, ImportNote{Key: k, Value: noteString(v), Reason: reason})
			continue
		}
		if reason, ignored := importIgnoredKeys[lower]; ignored {
			out.Ignored = append(out.Ignored, ImportNote{Key: k, Value: noteString(v), Reason: reason})
			continue
		}

		switch lower {
		case "title":
			out.Title = normalizeTitle(scalarString(v))
			if out.Title != "" {
				out.TitleSource = TitleFromFrontMatter
			}
		case "description":
			out.Description = sanitizeUTF8(strings.TrimSpace(scalarString(v)))
		case "status":
			st := domain.EntryStatus(strings.ToLower(strings.TrimSpace(scalarString(v))))
			if validEntryStatus(st) {
				out.Status = st
			} else {
				out.Fields = append(out.Fields, FieldNote{
					Key: "status", Value: noteString(v), Applied: false,
					Reason: "not one of draft, active, completed or archived — importing as a draft",
				})
			}
		case "tags":
			out.Tags = stringList(v)
			if _, list := v.([]any); !list && len(out.Tags) > 0 {
				// A hand-written `tags: a, b` is accepted, because people write
				// it and refusing would be pedantry — but split on commas, which
				// is a guess worth admitting to.
				out.Fields = append(out.Fields, FieldNote{
					Key: "tags", Value: noteString(v), Applied: true,
					Reason: "written as one line rather than a YAML list, so it was split on commas",
				})
			}
		default:
			// Rendered to text before the validator sees it. Every field type
			// except `number` is checked as a string, and a person writing
			// `due: 2024-01-02` unquoted — which is how YAML is written
			// everywhere — was being told their valid date was "expected text".
			candidate[k] = frontMatterValue(v)
		}
	}

	// The same validator every other write uses, so a value the import accepts
	// is a value the section would have accepted anyway.
	stored, report := domain.ValidateMetadata(section.FieldSpec, candidate)
	// A value the validator could not take is dropped here, unlike everywhere
	// else in the product.
	//
	// ValidateMetadata keeps one — storing it verbatim and flagging it — because
	// discarding a person's answer to protect a type is the same mistake as
	// refusing the write, when the writer is a model mid-interview. But this
	// preview feeds a commit that refuses, and the dialog makes only the title
	// editable, so keeping it produced a document that could never be imported:
	// every press of the button returned the same 400 about a value the person
	// had no control that could clear. Dropped and reported is the only shape
	// with a way forward in it.
	for _, bad := range report.InvalidValues {
		delete(stored, bad.Key)
	}
	out.Metadata = stored
	// The validator reports keys; a person looking at a file they did not write
	// also needs what the file said, or "stage is not one of the options" sends
	// them back to the file to find out which option it tried.
	for i, issue := range report.UnknownKeys {
		report.UnknownKeys[i].Value = noteString(candidate[issue.Key])
	}
	for i, issue := range report.InvalidValues {
		report.InvalidValues[i].Value = noteString(candidate[issue.Key])
	}
	out.Report = report

	if out.Title == "" {
		if h := firstHeading(out.Body); h != "" {
			out.Title, out.TitleSource = h, TitleFromHeading
		} else {
			out.Title, out.TitleSource = titleFromFilename(filename), TitleFromFilename
			out.Fields = append(out.Fields, FieldNote{
				Key: "title", Value: out.Title, Applied: true,
				Reason: "the file names no title and the document has no heading, so this came from the filename",
			})
		}
	}
	if out.Description == "" {
		// What Create is about to derive, shown before it does. Leaving it blank
		// meant the dialog showed no description and the document came out with
		// one — on the one screen whose entire contract is that nothing happens
		// that it did not show.
		out.Description = autoDescription(out.Body)
	}
	out.BodyLines = strings.Count(out.Body, "\n") + 1

	out.Warnings = importWarnings(ctx, s, section, out.Ignored)
	out.UnresolvedRefs = s.unresolvedRefs(ctx, section.ResearchID,
		out.Title+"\n"+out.Description+"\n"+out.Body)

	return out, nil
}

// importWarnings says out loud when the file thinks it belongs somewhere else.
// It is a remark, never a refusal: a document moved between researches on
// purpose is the ordinary case, and refusing it would make the feature useless
// for the one job people actually have.
func importWarnings(ctx context.Context, s *EntryService, section *domain.Section, ignored []ImportNote) []string {
	var out []string
	for _, n := range ignored {
		switch strings.ToLower(n.Key) {
		case "research":
			if research, err := s.researches.FindByID(ctx, section.ResearchID); err == nil && research != nil {
				if n.Value != "" && !strings.EqualFold(n.Value, research.Code) {
					out = append(out, fmt.Sprintf("The file says it came from %s; you are importing it into %s. Any [[E…]] in it means codes from %s.", n.Value, research.Code, n.Value))
				}
			}
		case "section":
			name := section.DisplayName
			if name == "" {
				name = section.Name
			}
			if n.Value != "" && !strings.EqualFold(n.Value, name) {
				out = append(out, fmt.Sprintf("The file says it belongs to %q; you are importing it into %q.", n.Value, name))
			}
		}
	}
	return out
}

// unresolvedRefs lists the [[...]] codes that name nothing here.
//
// It resolves and throws the result away — the import writes no crossrefs of
// its own, because Create does that from the stored content a moment later.
// What this exists for is telling somebody, before they commit, that a third of
// the links in their document are about to point at nothing.
func (s *EntryService) unresolvedRefs(ctx context.Context, researchID, text string) []UnresolvedRef {
	var out []UnresolvedRef
	at := map[string]int{}
	// Through the reader's eyes before anything is reported.
	//
	// resolveRefs resolves a short code globally, on purpose — codes are global
	// and the reader, not the author, decides whether a resolved reference is
	// shown. Reporting the raw bit here inverted that into an oracle: a list of
	// what did NOT resolve is a list of what did, so uploading a file full of
	// `[[R1]] [[R1:E1]] [[R1:E2]]…` into your own section read back the entire
	// code space of every research on the instance. VisibleCrossRefs exists to
	// close exactly that, and this is the one consumer of resolveRefs that was
	// not going through it.
	for _, cr := range s.access.VisibleCrossRefs(ctx, s.resolveRefs(ctx, "entry", "", researchID, text)) {
		if cr.Resolved {
			continue
		}
		if i, seen := at[cr.TargetRef]; seen {
			out[i].Count++
			continue
		}
		at[cr.TargetRef] = len(out)
		out = append(out, UnresolvedRef{Ref: cr.TargetRef, Count: 1})
	}
	return out
}

// splitFrontMatter separates a leading YAML block from the body.
//
// A file with no front matter is the ordinary case and not an error: most
// markdown in the world has none, and the issue this implements says so.
func splitFrontMatter(src string) (front, body string, err error) {
	// A BOM survives a round trip through several editors and would otherwise
	// stop the opening fence from matching, sending every such file down the
	// no-front-matter path with its YAML printed as prose.
	src = strings.TrimPrefix(src, "\ufeff")
	if !strings.HasPrefix(src, "---\n") && src != "---" {
		return "", src, nil
	}
	rest := strings.TrimPrefix(src, "---\n")
	// The closing fence is a line of exactly `---` or `...`, which is what YAML
	// itself accepts as a document end.
	for _, fence := range []string{"\n---\n", "\n...\n"} {
		if i := strings.Index(rest, fence); i >= 0 {
			return rest[:i], rest[i+len(fence):], nil
		}
	}
	for _, fence := range []string{"\n---", "\n..."} {
		if strings.HasSuffix(rest, fence) {
			return strings.TrimSuffix(rest, fence), "", nil
		}
	}
	// An opening fence with no closing one. Treating the whole file as front
	// matter would swallow the document; treating it as prose prints YAML at
	// the reader. Saying so is the only honest answer.
	return "", "", fmt.Errorf("%w: it opens with --- but never closes the block", ErrImportBadFrontMatter)
}

func parseFrontMatter(front string) (map[string]any, error) {
	if strings.TrimSpace(front) == "" {
		return map[string]any{}, nil
	}
	var raw map[string]any
	if err := yaml.Unmarshal([]byte(front), &raw); err != nil {
		// yaml.v3's message names the line, which is the whole point of
		// surfacing it rather than a sentence of our own — but it counts from
		// the start of the block it was handed, and the person is looking at a
		// file whose first line is the opening `---`. A number that is reliably
		// one off is worse than no number: they fix the wrong line.
		return nil, fmt.Errorf("%w: %s", ErrImportBadFrontMatter, shiftYAMLLine(err.Error()))
	}
	if raw == nil {
		return map[string]any{}, nil
	}
	return raw, nil
}

// yamlLine matches the "yaml: line N:" prefix yaml.v3 writes.
var yamlLine = regexp.MustCompile(`line (\d+)`)

// shiftYAMLLine renumbers to the file the person is looking at, which has the
// opening fence on line 1.
func shiftYAMLLine(msg string) string {
	return yamlLine.ReplaceAllStringFunc(msg, func(m string) string {
		n, err := strconv.Atoi(strings.TrimPrefix(m, "line "))
		if err != nil {
			return m
		}
		return fmt.Sprintf("line %d", n+1)
	})
}

// frontMatterValue renders a YAML value into what the metadata validator
// accepts: a string, or a list of strings for a repeated field. Numbers are
// left alone, because `number` is the one type checked numerically.
func frontMatterValue(v any) any {
	switch t := v.(type) {
	case []any:
		out := make([]any, 0, len(t))
		for _, e := range t {
			out = append(out, frontMatterValue(e))
		}
		return out
	case int, int64, float64:
		return t
	default:
		return scalarString(v)
	}
}

func looksLikeMarkdown(filename string) bool {
	switch strings.ToLower(filepath.Ext(filename)) {
	case ".md", ".markdown":
		return true
	}
	return false
}

// firstHeading takes a title from the document's own first ATX heading, at any
// level: a file whose top heading is `##` still has a name.
func firstHeading(body string) string {
	for _, line := range strings.Split(body, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "#") {
			continue
		}
		if h := normalizeTitle(strings.TrimLeft(line, "# ")); h != "" {
			return h
		}
	}
	return ""
}

// titleFromFilename is the last resort. The download names a file
// `E3 — Pricing model.md`, so the separator is stripped back off rather than
// left in the title of every document that makes the round trip.
func titleFromFilename(filename string) string {
	name := strings.TrimSuffix(filepath.Base(filename), filepath.Ext(filename))
	if _, after, found := strings.Cut(name, " — "); found && strings.TrimSpace(after) != "" {
		name = after
	}
	if t := normalizeTitle(name); t != "" {
		return t
	}
	return "Imported document"
}

// scalarString renders a front matter value as text, unclamped.
//
// It has to be unclamped because it is on the path that produces the entry's
// own fields, not only the path that describes them: a title or a description
// past the report's 120-rune ceiling was being stored cut, with an ellipsis and
// no note — losing exactly the hand-written description the download was taught
// to emit for. `noteString` is the clamped one, and it is for report values.
//
// yaml.v3 resolves an unquoted scalar to its Go type, so `due: 2024-01-02` is a
// time.Time and `flag: true` is a bool. Rendering them here is what lets an
// ordinary hand-written file reach the validator, which speaks strings.
func scalarString(v any) string {
	switch t := v.(type) {
	case nil:
		return ""
	case string:
		return t
	case time.Time:
		// The shape a date field expects, and the shape the export writes.
		if t.Hour() == 0 && t.Minute() == 0 && t.Second() == 0 {
			return t.Format("2006-01-02")
		}
		return t.Format(time.RFC3339)
	case []any:
		parts := make([]string, 0, len(t))
		for _, e := range t {
			parts = append(parts, scalarString(e))
		}
		return strings.Join(parts, ", ")
	case map[string]any:
		return fmt.Sprintf("(%d keys)", len(t))
	default:
		return fmt.Sprint(t)
	}
}

// noteString is scalarString with the ceiling a report needs. A front matter
// key holding a whole paragraph would otherwise push the reason off the screen.
func noteString(v any) string {
	s := oneLine(scalarString(v))
	if r := []rune(s); len(r) > 120 {
		return string(r[:120]) + "…"
	}
	return s
}

// stringList accepts both YAML shapes for tags: a sequence, and the
// comma-separated line people write by hand.
func stringList(v any) []string {
	out := []string{}
	switch t := v.(type) {
	case nil:
		return out
	case []any:
		for _, e := range t {
			if s := strings.TrimSpace(scalarString(e)); s != "" {
				out = append(out, s)
			}
		}
	case string:
		for _, part := range strings.Split(t, ",") {
			if s := strings.TrimSpace(part); s != "" {
				out = append(out, s)
			}
		}
	default:
		if s := strings.TrimSpace(scalarString(v)); s != "" {
			out = append(out, s)
		}
	}
	return out
}
