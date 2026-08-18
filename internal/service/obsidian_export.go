package service

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/butschster/mcp-research/internal/auth"
	"github.com/butschster/mcp-research/internal/domain"
	"github.com/butschster/mcp-research/internal/storage"
	"gopkg.in/yaml.v3"
)

// A research exported as an Obsidian vault: a folder per section, a note per
// entry, and the cross-references already written in the content resolving as
// links.
//
// The decision the whole file rests on: **a cross-reference is retargeted at the
// note's filename and keeps its code as the display text** — `[[E3]]` becomes
// `[[E3 — Pricing model|E3]]`. The reader sees exactly what the author wrote and
// the link resolves, which makes backlinks, unlinked mentions and the graph view
// work.
//
// The issue this was built from specified the opposite — leave `[[E3]]` alone
// and let `aliases: [E3]` in the frontmatter resolve it. A real vault says
// otherwise: an alias makes a note findable in the quick switcher, in search and
// in link autocomplete (where choosing a suggestion inserts a link to the
// filename), but a hand-written `[[E3]]` resolves against filenames only. Left
// alone it renders as an unresolved link and clicking it offers to create an
// empty note. Aliases are still written, for the switcher.
//
// Two consequences run through everything below:
//
//   - Retargeting happens on the way out, never in storage. This is a rendering,
//     like a block document becoming markdown, and re-exporting rebuilds it from
//     whatever the titles are then.
//   - Every referenceable entity needs a note, or a link has nowhere to land.
//     Entries, sessions, tasks and roadmaps each get one; a code naming something
//     outside the export gets a stub under _Unresolved/ that says so.

// VaultOptions selects which parts of a research are written.
type VaultOptions struct {
	Sessions  bool
	Tasks     bool
	Roadmaps  bool
	HTML      bool
	Revisions bool
	// RedactProvenance drops the facts about *how* an entry came to exist —
	// which session produced it, which revision it is on, who wrote that
	// revision. They are working process rather than findings, and a share link
	// publishes findings.
	//
	// It is phrased as a negative on purpose: the zero value has to be the
	// owner's full vault, so a caller that builds this struct without knowing
	// the field exists cannot silently strip an export.
	RedactProvenance bool
}

// DefaultVaultOptions writes everything a research holds except the revision
// tables, which are a curiosity in a vault rather than a working document.
func DefaultVaultOptions() VaultOptions {
	return VaultOptions{Sessions: true, Tasks: true, Roadmaps: true, HTML: true}
}

// VaultFile is one file in the archive. Path is relative to the archive root and
// always uses forward slashes, which is what the zip format wants.
type VaultFile struct {
	Path    string
	Content []byte
}

// Vault is the whole archive: a root folder name and its files.
type Vault struct {
	Root  string
	Files []VaultFile
}

type ObsidianService struct {
	research  *ResearchService
	section   *SectionService
	entries   *storage.EntryRepository
	session   *SessionService
	task      *TaskService
	roadmap   *RoadmapService
	revisions *storage.EntryRevisionRepository
	log       *slog.Logger
}

func NewObsidianService(
	research *ResearchService,
	section *SectionService,
	entries *storage.EntryRepository,
	session *SessionService,
	task *TaskService,
	roadmap *RoadmapService,
	revisions *storage.EntryRevisionRepository,
	log *slog.Logger,
) *ObsidianService {
	return &ObsidianService{
		research: research, section: section, entries: entries,
		session: session, task: task, roadmap: roadmap,
		revisions: revisions, log: log,
	}
}

// Vault builds the archive for one research.
//
// Ownership is settled by the first call: ResearchService.Get returns ErrNotFound
// for a research belonging to someone else, and every read below is by that
// research's id. That call is also what redacts `instruction` and `memory` for
// a share visitor, so the research this builds from is already the published
// one.
//
// What Get cannot do is narrow the *options*, which arrive from a query string.
// clampForShare does that, here rather than in the handler, for the same reason
// redaction lives inside Get: a second entry point that forgot would be a leak
// with no visible symptom.
func (s *ObsidianService) Vault(ctx context.Context, idOrCode string, opts VaultOptions) (*Vault, error) {
	research, err := s.research.Get(ctx, idOrCode)
	if err != nil {
		return nil, err
	}
	opts = clampForShare(ctx, opts)

	b := &vaultBuilder{
		opts:     opts,
		research: research,
		names:    map[string]string{},
		missing:  map[string]string{},
	}

	sections, err := s.section.List(ctx, research.ID)
	if err != nil {
		return nil, fmt.Errorf("list sections: %w", err)
	}
	entries, err := s.entries.FindByResearchWithContent(ctx, research.ID)
	if err != nil {
		return nil, fmt.Errorf("list entries: %w", err)
	}
	// The vault reads the repository directly, so the redaction that lives on
	// the service has to be applied by hand. Today the vault prints only keys
	// the section declares, and the section arrives redacted — but the values
	// are in memory either way, and anything that later renders them without
	// consulting the declaration would leak with no signal in review.
	redactEntriesForShare(ctx, entries)

	var sessions []*domain.Session
	questions := map[string][]*domain.Question{}
	sessionByID := map[string]*domain.Session{}
	// Sessions are read even when the folder is omitted: an entry's frontmatter
	// names the session that produced it, and that needs the code. When the
	// provenance is redacted too there is nothing left to resolve, so the read
	// is skipped rather than fetched and discarded.
	if opts.Sessions || !opts.RedactProvenance {
		sessions, err = s.session.ListByResearch(ctx, research.ID)
		if err != nil {
			return nil, fmt.Errorf("list sessions: %w", err)
		}
		for _, sess := range sessions {
			sessionByID[sess.ID] = sess
			if opts.Sessions {
				qs, err := s.session.ListQuestions(ctx, sess.ID, storage.QuestionFilter{})
				if err != nil {
					return nil, fmt.Errorf("list questions of %s: %w", sess.Code, err)
				}
				questions[sess.ID] = qs
			}
		}
	}

	var tasks []*domain.Task
	if opts.Tasks {
		tasks, err = s.task.List(ctx, research.ID, storage.TaskFilter{})
		if err != nil {
			return nil, fmt.Errorf("list tasks: %w", err)
		}
	}

	var roadmaps []*domain.Roadmap
	if opts.Roadmaps {
		list, err := s.roadmap.List(ctx, research.ID)
		if err != nil {
			return nil, fmt.Errorf("list roadmaps: %w", err)
		}
		for _, rm := range list {
			full, err := s.roadmap.Get(ctx, rm.ID)
			if err != nil {
				return nil, fmt.Errorf("get roadmap %s: %w", rm.Code, err)
			}
			roadmaps = append(roadmaps, full)
		}
	}

	revisions := map[string][]*domain.EntryRevision{}
	if opts.Revisions && s.revisions != nil {
		for _, e := range entries {
			list, err := s.revisions.ListByEntry(ctx, e.ID, 0)
			if err != nil {
				return nil, fmt.Errorf("list revisions of %s: %w", e.Code, err)
			}
			revisions[e.ID] = list
		}
	}

	b.build(sections, entries, sessions, questions, sessionByID, tasks, roadmaps, revisions)

	return &Vault{Root: b.rootName(), Files: b.files}, nil
}

// clampForShare narrows the requested vault to what the link actually publishes.
//
// The options come from a query string, and a share visitor can type one. The
// include flags gate the *routes* that serve sessions, tasks and roadmaps, and
// the vault is a fourth way to the same data that those flags never saw — so
// without this, `?sessions=true` on a link created with sessions switched off
// would hand over the interview transcript in a zip.
//
// Revisions and provenance are refused outright rather than gated: there is no
// flag that publishes them, because who edited what, when, and from which
// session is working process. It is the same rule the shared entry pages follow.
func clampForShare(ctx context.Context, opts VaultOptions) VaultOptions {
	share := auth.ShareFromContext(ctx)
	if share == nil {
		return opts
	}
	opts.Sessions = opts.Sessions && share.Include.Sessions
	opts.Tasks = opts.Tasks && share.Include.Tasks
	opts.Roadmaps = opts.Roadmaps && share.Include.Roadmaps
	opts.Revisions = false
	opts.RedactProvenance = true
	return opts
}

// ── the builder ──

type vaultBuilder struct {
	opts     VaultOptions
	research *domain.Research
	files    []VaultFile
	// names maps a short code to the note that answers to it, by filename. It is
	// what makes a link resolve — see linkify.
	names map[string]string
	// missing collects the codes that name something outside this export, so each
	// one gets a stub rather than a link into nowhere.
	missing map[string]string
}

func (b *vaultBuilder) rootName() string {
	return nameWithCode(b.research.Code, b.research.Name, "Research")
}

// addNote writes a markdown note, rewriting its cross-references into links
// Obsidian resolves. Everything a reader reads goes through here.
func (b *vaultBuilder) addNote(path string, body string) {
	b.files = append(b.files, VaultFile{Path: path, Content: []byte(b.linkify(body))})
}

// addFile writes content that must not be touched: an HTML document, and the
// stub notes, whose whole subject is a reference that could not be resolved.
func (b *vaultBuilder) addFile(path string, content string) {
	b.files = append(b.files, VaultFile{Path: path, Content: []byte(content)})
}

func (b *vaultBuilder) claim(code, note string) {
	if code != "" {
		b.names[code] = note
	}
}

// linkify turns `[[E3]]` into `[[E3 — Pricing model|E3]]`.
//
// This is the one thing the exporter changes about the text, and it took a real
// vault to learn that it has to. Obsidian's `aliases` frontmatter makes a note
// findable in the quick switcher and in autocomplete — where picking a
// suggestion inserts a link to the *filename* with the alias as display text —
// but a hand-written `[[E3]]` resolves against filenames only. Left alone it is
// an unresolved link, and clicking it offers to create an empty note.
//
// So the target becomes the filename and the display half keeps the code, which
// is what the author wrote and what the reader sees. Nothing is rewritten in
// storage: this is a rendering, the same way a block document is rendered to
// markdown on its way out, and re-exporting rebuilds it from the current titles.
func (b *vaultBuilder) linkify(text string) string {
	if !strings.Contains(text, "[[") {
		return text
	}
	return refPattern.ReplaceAllStringFunc(text, func(match string) string {
		ref := strings.TrimSpace(match[2 : len(match)-2])
		// A link that already carries display text is one this exporter wrote.
		if ref == "" || strings.Contains(ref, "|") {
			return match
		}
		target, ok := b.target(ref)
		if !ok || target == ref {
			return match
		}
		return "[[" + target + "|" + ref + "]]"
	})
}

// target resolves a reference to the note that should answer it.
func (b *vaultBuilder) target(ref string) (string, bool) {
	if name, ok := b.names[ref]; ok {
		return name, true
	}
	// Prose in double brackets is not a reference: only something shaped like a
	// short code earns a stub, and everything else is left exactly as written.
	if !codeRef.MatchString(ref) {
		return "", false
	}
	if name, ok := b.missing[ref]; ok {
		return name, true
	}
	// A bound on the damage: a research whose entries name hundreds of deleted
	// codes should not turn the vault into a folder of apologies.
	if len(b.missing) >= maxStubs {
		return "", false
	}
	name := sanitizeSegment(strings.ReplaceAll(ref, ":", " "), "ref")
	b.missing[ref] = name
	return name, true
}

func (b *vaultBuilder) build(
	sections []*domain.Section,
	entries []*domain.Entry,
	sessions []*domain.Session,
	questions map[string][]*domain.Question,
	sessionByID map[string]*domain.Session,
	tasks []*domain.Task,
	roadmaps []*domain.Roadmap,
	revisions map[string][]*domain.EntryRevision,
) {
	// Every note's name is known before a single one is written, because a note
	// written first may link to one written last.
	// The index keeps the name a reader expects to find at the root of a folder.
	// A [[R5]] in the prose therefore targets "README", which resolves inside one
	// unzipped vault; the footers below use a relative path instead, so they hold
	// even when two exports share a vault and two README.md files exist.
	b.claim(b.research.Code, indexNote)
	for _, e := range entries {
		name := nameWithCode(e.Code, e.Title, "Entry")
		b.claim(e.Code, name)
		// [[R5:E1]] names the same entry as [[E1]] when R5 is this research —
		// the product resolves both, so the vault has to as well.
		if b.research.Code != "" && e.Code != "" {
			b.claim(b.research.Code+":"+e.Code, name)
		}
	}
	if b.opts.Sessions {
		for _, sess := range sessions {
			b.claim(sess.Code, nameWithCode(sess.Code, sess.Title, "Session"))
		}
	}
	for _, t := range tasks {
		b.claim(t.Code, nameWithCode(t.Code, t.Title, "Task"))
	}
	for _, rm := range roadmaps {
		name := nameWithCode(rm.Code, rm.Title, "Roadmap")
		b.claim(rm.Code, name)
		for _, n := range rm.Nodes {
			// A node link — [[RM1:N3]] — lands on the roadmap note, which is where
			// the node is drawn.
			b.claim(rm.Code+":"+n.Code, name)
		}
	}

	entriesBySection := map[string][]*domain.Entry{}
	for _, e := range entries {
		entriesBySection[e.SectionID] = append(entriesBySection[e.SectionID], e)
	}

	// Section folders are numbered so a file manager's alphabetical order is the
	// order the research was designed in. Obsidian sorts folders by name and
	// knows nothing about our `position`.
	folders := map[string]string{}
	for i, sec := range sections {
		name := sec.DisplayName
		if name == "" {
			name = sec.Name
		}
		folders[sec.ID] = fmt.Sprintf("%02d — %s", i+1, sanitizeSegment(name, "Section"))
	}

	sessionCode := func(id string) string {
		if sess := sessionByID[id]; sess != nil {
			return sess.Code
		}
		return ""
	}

	for _, sec := range sections {
		for _, e := range entriesBySection[sec.ID] {
			b.writeEntry(e, sec, folders[sec.ID], sessionCode(e.SessionID), revisions[e.ID], sessionCode)
		}
	}
	// An entry whose section vanished would otherwise be silently dropped.
	for _, e := range entries {
		if _, ok := folders[e.SectionID]; !ok {
			b.writeEntry(e, nil, "Unfiled", sessionCode(e.SessionID), revisions[e.ID], sessionCode)
		}
	}

	if b.opts.Sessions {
		for _, sess := range sessions {
			b.writeSession(sess, questions[sess.ID], entries)
		}
	}
	for _, t := range tasks {
		b.writeTask(t)
	}
	for _, rm := range roadmaps {
		b.writeRoadmap(rm)
	}

	b.writeReadme(sections, entriesBySection, folders, sessions, tasks, roadmaps)
	b.writeUnresolvedStubs()
}

// ── entries ──

func (b *vaultBuilder) writeEntry(
	e *domain.Entry,
	sec *domain.Section,
	folder string,
	sessCode string,
	revs []*domain.EntryRevision,
	sessionCode func(string) string,
) {
	fm := &frontmatter{}
	fm.add("code", e.Code)
	fm.add("title", e.Title)
	fm.add("aliases", aliasesFor(e.Code))
	fm.add("research", b.research.Code)
	if sec != nil {
		name := sec.DisplayName
		if name == "" {
			name = sec.Name
		}
		fm.add("section", name)
	}
	fm.add("type", string(e.Type))
	fm.add("status", string(e.Status))
	fm.add("tags", e.Tags)
	fm.add("created", stamp(e.CreatedAt))
	fm.add("updated", stamp(e.UpdatedAt))
	// Which session produced the entry is provenance, and a share does not carry
	// it. `sessCode` is already empty in that case — the sessions are not read —
	// but the guard is written out so the rule is visible where it applies.
	if sessCode != "" && !b.opts.RedactProvenance {
		fm.add("session", sessCode)
	}
	// User metadata last, so the eleven system keys keep their order and a
	// reader scanning two notes sees the same shape twice. Keys are refused at
	// declaration time when they collide with one of those, which is why
	// nothing here has to guard against overwriting them.
	addMetadataFrontmatter(fm, sec, e)
	if len(revs) > 0 {
		// The newest revision leads the list, which is what makes this the
		// entry's current provenance rather than its first.
		fm.add("revision", revs[0].Revision)
		fm.add("author", string(revs[0].AuthorKind))
	}

	var body strings.Builder
	body.WriteString(fm.render())
	body.WriteString(b.entryBody(e))

	// Every note links home. README → note works from the index; without this,
	// note → index means the file tree or the backlinks pane, which is the
	// navigation a reader performs dozens of times.
	footer := []string{"Research: " + mdLink(b.rootName(), "../"+indexNote+".md")}
	if sessCode != "" && b.opts.Sessions && !b.opts.RedactProvenance {
		footer = append(footer, "Session: "+b.wikilink(sessCode))
	}
	if len(revs) > 0 && b.opts.Revisions {
		b.addNote("_history/"+sanitizeSegment(e.Code, "entry")+".md", b.revisionTable(e, revs, sessionCode))
		footer = append(footer, "History: "+mdLink("Revisions", "../_history/"+sanitizeSegment(e.Code, "entry")+".md"))
	}
	// Obsidian cannot index a tag containing a space or a '#'. The exact tag
	// still goes in the frontmatter — mangling an author's vocabulary to fit a
	// reader is worse than saying it plainly here.
	if odd := unusableTags(e.Tags); len(odd) > 0 {
		footer = append(footer, "Tags Obsidian cannot index: "+strings.Join(quoteAll(odd), ", "))
	}
	body.WriteString("\n\n---\n\n" + strings.Join(footer, "  \n") + "\n")

	b.addNote(folder+"/"+nameWithCode(e.Code, e.Title, "Entry")+".md", body.String())
}

// entryBody renders the entry's content for a note.
//
// The vault's two differences from a loose file live here and nowhere else: it
// omits the mermaid live-editor link, because Obsidian draws mermaid itself and
// the link is a kilobyte of base64 in the middle of a note; and it writes an
// html block beside the vault as a real file and links it, because Obsidian
// sanitizes inline HTML and would show a broken shell of an artifact.
func (b *vaultBuilder) entryBody(e *domain.Entry) string {
	return entryMarkdownBody(e, MarkdownOptions{
		OmitMermaidLink: true,
		HTMLBlock:       func(blk domain.Block) string { return b.htmlBlock(e, blk) },
	})
}

func (b *vaultBuilder) htmlBlock(e *domain.Entry, blk domain.Block) string {
	title := str(blk.Data, "title")
	if title == "" {
		title = "HTML block"
	}
	caption := str(blk.Data, "caption")

	if !b.opts.HTML {
		out := "> [!info] " + title + " — interactive HTML, omitted from this export\n"
		if caption != "" {
			out += ">\n> " + caption + "\n"
		}
		return out + "\n"
	}

	name := nameWithCode(e.Code, title, "HTML")
	if blk.ID != "" {
		name += " (" + blk.ID + ")"
	}
	path := "_html/" + name + ".html"
	b.addFile(path, standaloneHTML(title, str(blk.Data, "html")))

	out := "> [!info] " + title + " — interactive HTML\n"
	if caption != "" {
		out += ">\n> " + caption + "\n"
	}
	out += ">\n> " + mdLink("Open in a browser", "../"+path) + "\n\n"
	return out
}

func (b *vaultBuilder) revisionTable(e *domain.Entry, revs []*domain.EntryRevision, sessionCode func(string) string) string {
	fm := &frontmatter{}
	fm.add("title", e.Code+" — revision history")
	fm.add("entry", e.Code)
	fm.add("research", b.research.Code)

	var body strings.Builder
	body.WriteString(fm.render())
	body.WriteString("Revision history of [[" + e.Code + "]] as it stood when the vault was exported.\n\n")
	body.WriteString("| Revision | Date | Author | Session | Summary |\n|---|---|---|---|---|\n")
	for _, r := range revs {
		summary := r.Summary
		if summary == "" && r.Revision == 1 {
			summary = "Created"
		}
		// The repository does not enrich session codes, so resolve them through the
		// map built from this research's own sessions. Looking the id up directly
		// is how a cross-user leak got in last time.
		body.WriteString(fmt.Sprintf("| %d | %s | %s | %s | %s |\n",
			r.Revision, stamp(r.CreatedAt), r.AuthorKind, sessionCode(r.SessionID), escapeCell(summary)))
	}
	return body.String()
}

// ── sessions ──

func (b *vaultBuilder) writeSession(sess *domain.Session, questions []*domain.Question, entries []*domain.Entry) {
	fm := &frontmatter{}
	fm.add("code", sess.Code)
	fm.add("title", sess.Title)
	fm.add("aliases", aliasesFor(sess.Code))
	fm.add("research", b.research.Code)
	fm.add("focus", sess.Focus)
	fm.add("status", string(sess.Status))
	fm.add("created", stamp(sess.CreatedAt))
	fm.add("updated", stamp(sess.UpdatedAt))

	var body strings.Builder
	body.WriteString(fm.render())
	if sess.Focus != "" {
		body.WriteString("> " + sess.Focus + "\n\n")
	}
	if sess.Notes != "" {
		body.WriteString("## Notes\n\n" + normalizeContent(sess.Notes) + "\n\n")
	}

	answered := 0
	for _, q := range questions {
		if q.Status == domain.QuestionAnswered {
			answered++
		}
	}
	body.WriteString(fmt.Sprintf("## Questions (%d answered of %d)\n\n", answered, len(questions)))
	if len(questions) == 0 {
		body.WriteString("*No questions in this session.*\n\n")
	}
	for _, q := range questions {
		heading := q.Text
		if q.Code != "" {
			heading = q.Code + " — " + q.Text
		}
		body.WriteString("### " + oneLine(heading) + "\n\n")
		meta := []string{"Status: " + string(q.Status), "Priority: " + string(q.Priority)}
		if q.Area != "" {
			meta = append(meta, "Area: "+q.Area)
		}
		body.WriteString("*" + strings.Join(meta, " · ") + "*\n\n")
		if q.Rationale != "" {
			body.WriteString(q.Rationale + "\n\n")
		}
		if q.Answer != "" {
			body.WriteString(normalizeContent(q.Answer) + "\n\n")
		}
	}

	var produced []string
	for _, e := range entries {
		if e.SessionID == sess.ID && e.Code != "" {
			produced = append(produced, "- "+b.wikilink(e.Code))
		}
	}
	if len(produced) > 0 {
		body.WriteString("## Entries produced in this session\n\n" + strings.Join(produced, "\n") + "\n")
	}
	body.WriteString("\n---\n\nResearch: " + mdLink(b.rootName(), "../"+indexNote+".md") + "\n")

	b.addNote("Sessions/"+nameWithCode(sess.Code, sess.Title, "Session")+".md", body.String())
}

// ── tasks ──

func (b *vaultBuilder) writeTask(t *domain.Task) {
	fm := &frontmatter{}
	fm.add("code", t.Code)
	fm.add("title", t.Title)
	fm.add("aliases", aliasesFor(t.Code))
	fm.add("research", b.research.Code)
	fm.add("status", string(t.Status))
	fm.add("priority", string(t.Priority))
	fm.add("created", stamp(t.CreatedAt))
	fm.add("updated", stamp(t.UpdatedAt))
	if t.CompletedAt != nil {
		fm.add("completed", stamp(*t.CompletedAt))
	}

	var body strings.Builder
	body.WriteString(fm.render())
	mark := " "
	switch t.Status {
	case domain.TaskCompleted:
		mark = "x"
	case domain.TaskFailed:
		mark = "!"
	}
	body.WriteString("- [" + mark + "] " + oneLine(t.Title) + "\n\n")
	if t.Description != "" {
		body.WriteString(normalizeContent(t.Description) + "\n\n")
	}
	if t.Result != "" {
		body.WriteString("## Result\n\n" + normalizeContent(t.Result) + "\n")
	}
	body.WriteString("\n---\n\nResearch: " + mdLink(b.rootName(), "../"+indexNote+".md") + "\n")

	b.addNote("Tasks/"+nameWithCode(t.Code, t.Title, "Task")+".md", body.String())
}

// ── roadmaps ──

func (b *vaultBuilder) writeRoadmap(rm *domain.Roadmap) {
	fm := &frontmatter{}
	fm.add("code", rm.Code)
	fm.add("title", rm.Title)
	fm.add("aliases", roadmapAliases(rm))
	fm.add("research", b.research.Code)
	fm.add("status", string(rm.Status))
	fm.add("statuses", rm.Statuses)
	fm.add("created", stamp(rm.CreatedAt))
	fm.add("updated", stamp(rm.UpdatedAt))

	var body strings.Builder
	body.WriteString(fm.render())
	if rm.Description != "" {
		body.WriteString(normalizeContent(rm.Description) + "\n\n")
	}
	// A mermaid fence is the one place where the vault reads better than this
	// product's own markdown export: Obsidian draws it.
	if diagram := roadmapMermaid(rm); diagram != "" {
		body.WriteString("```mermaid\n" + diagram + "```\n\n")
	}

	if len(rm.Nodes) > 0 {
		body.WriteString("## Nodes\n\n")
		for _, n := range rm.Nodes {
			body.WriteString("### " + oneLine(nodeLabel(n)) + "\n\n")
			var meta []string
			if n.NodeType != "" {
				meta = append(meta, "Type: "+n.NodeType)
			}
			if n.Status != "" {
				meta = append(meta, "Status: "+n.Status)
			}
			if ref := b.nodeRef(n); ref != "" {
				meta = append(meta, ref)
			}
			if len(meta) > 0 {
				body.WriteString("*" + strings.Join(meta, " · ") + "*\n\n")
			}
			if n.Description != "" {
				body.WriteString(normalizeContent(n.Description) + "\n\n")
			}
		}
	}

	body.WriteString("\n---\n\nResearch: " + mdLink(b.rootName(), "../"+indexNote+".md") + "\n")

	b.addNote("Roadmaps/"+nameWithCode(rm.Code, rm.Title, "Roadmap")+".md", body.String())
}

// nodeRef renders what a node points at.
//
// A code alone is not enough to link with: RefData carries the research the
// referenced entity belongs to, and A's roadmap pointing at B's E1 would
// otherwise produce a link onto A's own E1 — a wrong link that looks right.
// Anything this vault cannot resolve is named in plain text instead, which is
// also the answer for a question: it is a heading inside a session note, not a
// note of its own.
func (b *vaultBuilder) nodeRef(n *domain.RoadmapNode) string {
	if n.RefData == nil {
		return ""
	}
	d := n.RefData
	if d.ResearchID != "" && d.ResearchID != b.research.ID {
		if d.Title != "" {
			return "References " + oneLine(d.Title) + " in another research"
		}
		return ""
	}
	if _, ok := b.names[d.Code]; !ok {
		if d.Title != "" {
			return "References " + oneLine(d.Title)
		}
		return ""
	}
	return "References " + b.wikilink(d.Code)
}

func nodeLabel(n *domain.RoadmapNode) string {
	if n.Code == "" {
		return n.Title
	}
	return n.Code + " — " + n.Title
}

func roadmapMermaid(rm *domain.Roadmap) string {
	if len(rm.Nodes) == 0 {
		return ""
	}
	id := map[string]string{}
	for i, n := range rm.Nodes {
		key := n.Code
		if key == "" {
			key = fmt.Sprintf("n%d", i+1)
		}
		id[n.ID] = key
	}

	var b strings.Builder
	b.WriteString("flowchart TD\n")
	for _, n := range rm.Nodes {
		label := n.Title
		if n.Status != "" {
			label += " (" + n.Status + ")"
		}
		b.WriteString("  " + id[n.ID] + "[\"" + mermaidLabel(label) + "\"]\n")
	}
	drawn := map[string]bool{}
	for _, e := range rm.Edges {
		src, dst := id[e.SourceNodeID], id[e.TargetNodeID]
		if src == "" || dst == "" {
			continue
		}
		drawn[src+">"+dst] = true
		if e.Label != "" {
			b.WriteString("  " + src + " -->|" + mermaidLabel(e.Label) + "| " + dst + "\n")
		} else {
			b.WriteString("  " + src + " --> " + dst + "\n")
		}
	}
	// A nesting relationship carries no edge of its own; without this a tree of
	// nodes draws as a row of orphans.
	for _, n := range rm.Nodes {
		if n.ParentID == "" {
			continue
		}
		src, dst := id[n.ParentID], id[n.ID]
		if src == "" || dst == "" || drawn[src+">"+dst] {
			continue
		}
		b.WriteString("  " + src + " -.-> " + dst + "\n")
	}
	return b.String()
}

func mermaidLabel(s string) string {
	s = oneLine(s)
	// Quotes end the label and a hash starts a directive; both are escaped the
	// way mermaid's own parser expects.
	s = strings.ReplaceAll(s, "#", "#35;")
	s = strings.ReplaceAll(s, "\"", "#quot;")
	return clampStr(s, 120)
}

// ── the root note ──

func (b *vaultBuilder) writeReadme(
	sections []*domain.Section,
	entriesBySection map[string][]*domain.Entry,
	folders map[string]string,
	sessions []*domain.Session,
	tasks []*domain.Task,
	roadmaps []*domain.Roadmap,
) {
	r := b.research
	fm := &frontmatter{}
	fm.add("code", r.Code)
	fm.add("title", r.Name)
	fm.add("aliases", aliasesFor(r.Code))
	fm.add("status", string(r.Status))
	fm.add("tags", r.Tags)
	fm.add("created", stamp(r.CreatedAt))
	fm.add("updated", stamp(r.UpdatedAt))
	fm.add("exported", stamp(time.Now().UTC()))

	var body strings.Builder
	body.WriteString(fm.render())
	body.WriteString("# " + oneLine(r.Name) + "\n\n")
	if r.Goal != "" {
		body.WriteString("> " + oneLine(r.Goal) + "\n\n")
	}
	if r.Description != "" {
		body.WriteString(normalizeContent(r.Description) + "\n\n")
	}

	// The table of contents: every document in the vault, in the order the
	// research is organized, each one a link. This is the note a reader opens
	// first, so it is the one place that has to be complete.
	body.WriteString("## Contents\n\n")
	if len(sections) == 0 {
		body.WriteString("*This research has no sections yet.*\n\n")
	}
	for _, sec := range sections {
		name := sec.DisplayName
		if name == "" {
			name = sec.Name
		}
		entries := entriesBySection[sec.ID]
		body.WriteString("### " + oneLine(folders[sec.ID]) + "\n\n")
		if sec.Description != "" {
			body.WriteString(oneLine(sec.Description) + "\n\n")
		}
		if len(entries) == 0 {
			body.WriteString("*No entries.*\n\n")
			continue
		}
		for _, e := range entries {
			body.WriteString("- " + b.wikilink(e.Code))
			if e.Status != domain.EntryActive {
				body.WriteString(" — *" + string(e.Status) + "*")
			}
			body.WriteString("\n")
		}
		body.WriteString("\n")
	}

	if b.opts.Sessions && len(sessions) > 0 {
		body.WriteString("### Sessions\n\n")
		for _, sess := range sessions {
			body.WriteString("- " + b.wikilink(sess.Code) +
				" — *" + string(sess.Status) + "*\n")
		}
		body.WriteString("\n")
	}
	if len(tasks) > 0 {
		body.WriteString("### Tasks\n\n")
		for _, t := range tasks {
			mark := " "
			if t.Status == domain.TaskCompleted {
				mark = "x"
			}
			body.WriteString("- [" + mark + "] " + b.wikilink(t.Code) +
				" — *" + string(t.Status) + "*\n")
		}
		body.WriteString("\n")
	}
	if len(roadmaps) > 0 {
		body.WriteString("### Roadmaps\n\n")
		for _, rm := range roadmaps {
			body.WriteString("- " + b.wikilink(rm.Code) +
				fmt.Sprintf(" — %d nodes\n", len(rm.Nodes)))
		}
		body.WriteString("\n")
	}

	if len(r.Memory) > 0 {
		body.WriteString("## Memory\n\n")
		for _, m := range r.Memory {
			body.WriteString("- " + oneLine(m) + "\n")
		}
		body.WriteString("\n")
	}
	if r.Instruction != "" {
		body.WriteString("## Instruction\n\n" + normalizeContent(r.Instruction) + "\n")
	}

	b.addNote(indexNote+".md", body.String())
}

// ── references that leave the vault ──

// codeRef matches what a short code looks like — R1, E12, SS3, RM1:N2, R2:E5 —
// so a stub is written for a real reference and never for prose that happens to
// sit inside double brackets.
var codeRef = regexp.MustCompile(`^[A-Z]{1,3}\d+(:[A-Z]{1,3}\d+)?$`)

// indexNote is the note at the root of the archive: the research's own page and
// the table of contents of everything else. Every note that links home links
// here, one directory up.
const indexNote = "README"

// maxStubs bounds a runaway: a research whose entries reference hundreds of
// deleted codes should not turn the vault into a folder of apologies.
const maxStubs = 200

// writeUnresolvedStubs gives every reference that names something outside this
// export a note to land on.
//
// Every link in the vault resolves — that is the point of writing these. A
// reference to another research, to a deleted entry, or to a folder the options
// left out is still a real reference, so it lands on a note that says what it
// points at rather than on Obsidian's offer to create an empty file.
//
// The set is whatever `linkify` could not resolve, so it is exact: no second
// scan of the output, and nothing here that nothing links to.
func (b *vaultBuilder) writeUnresolvedStubs() {
	refs := make([]string, 0, len(b.missing))
	for ref := range b.missing {
		refs = append(refs, ref)
	}
	sort.Strings(refs)

	for _, ref := range refs {
		fm := &frontmatter{}
		fm.add("title", ref)
		fm.add("aliases", []string{ref})
		fm.add("research", b.research.Code)
		fm.add("unresolved", true)

		body := fm.render() +
			"`" + ref + "` points at " + describeRef(ref) +
			" that is not part of this export.\n\n" +
			"Backlinks below show which notes reference it.\n"
		// addFile, not addNote: this note's subject is a reference that has
		// nowhere to go, and linkifying it would point it at itself.
		b.addFile("_Unresolved/"+b.missing[ref]+".md", body)
	}
}

func describeRef(ref string) string {
	head := ref
	if i := strings.IndexByte(ref, ':'); i >= 0 {
		head = ref[:i]
	}
	switch {
	case strings.HasPrefix(ref, "RM"):
		return "a roadmap"
	case strings.HasPrefix(head, "R"):
		if strings.Contains(ref, ":") {
			return "an entry in another research"
		}
		return "another research"
	case strings.HasPrefix(ref, "SS"):
		return "a session"
	case strings.HasPrefix(ref, "E"):
		return "an entry"
	case strings.HasPrefix(ref, "T"):
		return "a task"
	}
	return "something"
}

// ── frontmatter ──

type frontmatter struct {
	pairs []fmPair
}

type fmPair struct {
	key string
	val any
}

// add skips empty values: a note full of blank keys reads as broken metadata
// rather than absent metadata.
// addDeclared emits a key even when it has no value, as an explicit null.
//
// `add` skips empties, which is right for a system key nobody declared. It is
// wrong for a field a section says its documents record: the key would vanish
// from the note, and a vault query for "documents missing this field" would
// find nothing — the documents worth finding would be exactly the invisible
// ones.
func (f *frontmatter) addDeclared(key string, val any) {
	if val == nil {
		f.pairs = append(f.pairs, fmPair{key: key, val: nil})
		return
	}
	if str, ok := val.(string); ok && strings.TrimSpace(str) == "" {
		f.pairs = append(f.pairs, fmPair{key: key, val: nil})
		return
	}
	f.pairs = append(f.pairs, fmPair{key: key, val: val})
}

func (f *frontmatter) add(key string, val any) {
	switch v := val.(type) {
	case string:
		if strings.TrimSpace(v) == "" {
			return
		}
		val = oneLine(v)
	case []string:
		if len(v) == 0 {
			return
		}
	}
	f.pairs = append(f.pairs, fmPair{key: key, val: val})
}

// render emits the YAML block through a real encoder. A title holding a colon,
// a quote or a leading dash must not be able to produce a file Obsidian refuses
// to parse, which is exactly what string concatenation would do.
func (f *frontmatter) render() string {
	if len(f.pairs) == 0 {
		return ""
	}
	root := &yaml.Node{Kind: yaml.MappingNode}
	for _, p := range f.pairs {
		var v yaml.Node
		if err := v.Encode(p.val); err != nil {
			continue
		}
		if v.Kind == yaml.SequenceNode {
			// aliases: [E1] rather than a three-line block — this is metadata a
			// human scrolls past.
			v.Style = yaml.FlowStyle
		}
		root.Content = append(root.Content,
			&yaml.Node{Kind: yaml.ScalarNode, Value: p.key}, &v)
	}
	out, err := yaml.Marshal(root)
	if err != nil {
		return ""
	}
	return "---\n" + string(out) + "---\n\n"
}

// ── names and links ──

// aliasesFor makes the note findable by its code in the quick switcher, in
// search and in link autocomplete. It does NOT resolve links — see linkify.
func aliasesFor(code string) []string {
	if code == "" {
		return nil
	}
	return []string{code}
}

func roadmapAliases(rm *domain.Roadmap) []string {
	if rm.Code == "" {
		return nil
	}
	out := []string{rm.Code}
	for _, n := range rm.Nodes {
		if n.Code != "" {
			out = append(out, rm.Code+":"+n.Code)
		}
	}
	return out
}

// wikilink links to the note a code names, by filename — the only target
// Obsidian resolves. The filename already carries the code, so no display half
// is needed.
func (b *vaultBuilder) wikilink(code string) string {
	if name, ok := b.names[code]; ok {
		return "[[" + name + "]]"
	}
	return "[[" + code + "]]"
}

func mdLink(text, target string) string {
	return "[" + text + "](" + escapeLinkTarget(target) + ")"
}

// escapeLinkTarget escapes only what would end a markdown link early. Cyrillic
// and other non-ASCII stay readable: Obsidian and every browser handle them.
func escapeLinkTarget(s string) string {
	r := strings.NewReplacer(" ", "%20", "(", "%28", ")", "%29")
	return r.Replace(s)
}

// nameWithCode builds "E1 — Title", the filename a human reads. The code leads,
// which is also what makes two entries with the same title impossible to
// collide.
func nameWithCode(code, title, fallback string) string {
	title = sanitizeSegment(title, "")
	switch {
	case code != "" && title != "":
		return sanitizeSegment(code, fallback) + " — " + title
	case code != "":
		return sanitizeSegment(code, fallback)
	case title != "":
		return title
	}
	return fallback
}

// forbidden is what a filesystem refuses, plus the characters that would turn a
// name into a path.
var forbidden = map[rune]bool{
	'/': true, '\\': true, ':': true, '*': true, '?': true,
	'"': true, '<': true, '>': true, '|': true,
	// Obsidian's own illegal characters. A '#' or '^' inside a filename would
	// also cut a wikilink short — everything after it reads as a heading or a
	// block reference.
	'#': true, '^': true, '[': true, ']': true,
}

// windowsReserved names cannot be files on Windows whatever the extension.
var windowsReserved = map[string]bool{
	"CON": true, "PRN": true, "AUX": true, "NUL": true,
	"COM1": true, "COM2": true, "COM3": true, "COM4": true, "COM5": true,
	"COM6": true, "COM7": true, "COM8": true, "COM9": true,
	"LPT1": true, "LPT2": true, "LPT3": true, "LPT4": true, "LPT5": true,
	"LPT6": true, "LPT7": true, "LPT8": true, "LPT9": true,
}

// maxSegment bounds one path component in bytes. Two of them plus the archive
// root stay well inside the 255-byte limit every filesystem in play enforces.
const maxSegment = 100

// sanitizeSegment turns a title into a path component.
//
// Cyrillic and other non-ASCII stay: Obsidian handles them, and transliterating
// a Russian research into ASCII would make it unreadable to the person who
// wrote it.
func sanitizeSegment(s, fallback string) string {
	var b strings.Builder
	for _, r := range s {
		switch {
		case r == '/' || r == '\\':
			// A separator is dividing something — "часть 1/2" dropped to
			// "часть 12" reads as twelve. A dash keeps the division.
			b.WriteRune('-')
		case forbidden[r]:
			// Dropped rather than replaced: "Pricing: model" reads better as
			// "Pricing model" than as "Pricing- model".
			continue
		case r == '\n' || r == '\t' || unicode.IsControl(r):
			b.WriteRune(' ')
		default:
			b.WriteRune(r)
		}
	}
	out := strings.Join(strings.Fields(b.String()), " ")
	out = clampBytes(out, maxSegment)
	// A trailing dot or space is silently stripped by Windows, which would
	// desynchronize the name from the link that points at it.
	out = strings.Trim(out, " .")
	if out == "" {
		return fallback
	}
	if windowsReserved[strings.ToUpper(out)] {
		out += "_"
	}
	return out
}

// clampBytes cuts on a rune boundary: half a character in a filename is a file
// nothing can open.
func clampBytes(s string, max int) string {
	if len(s) <= max {
		return s
	}
	// Room for the ellipsis, which is what tells a reader the name is short of
	// the title rather than disagreeing with it — "порогом на пять" read as a
	// threshold of five until this was added.
	cut := s[:max-len("…")]
	for len(cut) > 0 && !utf8.ValidString(cut) {
		cut = cut[:len(cut)-1]
	}
	return strings.TrimRight(cut, " ") + "…"
}

// ── small helpers ──

func stamp(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format(time.RFC3339)
}

func oneLine(s string) string {
	return strings.Join(strings.Fields(strings.ReplaceAll(s, "\n", " ")), " ")
}

func escapeCell(s string) string {
	return strings.ReplaceAll(oneLine(s), "|", "\\|")
}

// unusableTags are the tags Obsidian's tag pane will not accept.
func unusableTags(tags []string) []string {
	var out []string
	for _, t := range tags {
		if strings.ContainsAny(t, " \t#") {
			out = append(out, t)
		}
	}
	return out
}

func quoteAll(in []string) []string {
	out := make([]string, 0, len(in))
	for _, s := range in {
		out = append(out, "\""+s+"\"")
	}
	return out
}

// standaloneHTML makes sure the file opens as a document. A block usually holds
// a complete page; one that does not would otherwise be a fragment the browser
// renders without a title or a charset.
func standaloneHTML(title, html string) string {
	trimmed := strings.TrimSpace(html)
	lower := strings.ToLower(trimmed)
	if strings.HasPrefix(lower, "<!doctype") || strings.HasPrefix(lower, "<html") {
		return trimmed + "\n"
	}
	return "<!doctype html>\n<html>\n<head>\n<meta charset=\"utf-8\">\n<title>" +
		htmlEscape(title) + "</title>\n</head>\n<body>\n" + trimmed + "\n</body>\n</html>\n"
}

func htmlEscape(s string) string {
	r := strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", "\"", "&quot;")
	return r.Replace(s)
}

// addMetadataFrontmatter writes a document's declared fields, in the order the
// section declares them.
//
// Shared by the vault and the single-file download, because these are the rules
// that drift: two copies would disagree about an unfilled field or an explicit
// unknown, and the disagreement would be invisible until somebody diffed two
// exports of the same document.
//
// It goes last, after the system keys, so a reader scanning two notes sees the
// same shape twice. Keys are refused at declaration time when they collide with
// one of those, which is why nothing here guards against overwriting them.
func addMetadataFrontmatter(fm *frontmatter, sec *domain.Section, e *domain.Entry) {
	if sec == nil {
		return
	}
	for _, f := range sec.FieldSpec {
		v, recorded := e.Metadata[f.Key]
		// A declared field nobody answered emits null, so a query for "documents
		// missing this" can find it. A field answered with an explicit unknown
		// emits the word, because somebody did look and that document is not the
		// one you are looking for.
		if recorded && v == nil {
			fm.addDeclared(f.Key, "unknown")
			continue
		}
		fm.addDeclared(f.Key, metadataForFrontmatter(f, v))
	}
}

// metadataForFrontmatter renders one value the way a vault query expects it.
//
// The rules are not cosmetic. A one-element list still has to be a YAML
// sequence, because a query filtering on membership fails against a string. A
// number has to stay unquoted or it sorts lexically. And a reference is emitted
// in its bracket form so Obsidian treats the property as a link rather than as
// text that happens to look like one.
func metadataForFrontmatter(spec domain.FieldSpec, v any) any {
	switch t := v.(type) {
	case nil:
		return nil
	case []any:
		out := make([]any, 0, len(t))
		for _, item := range t {
			out = append(out, metadataForFrontmatter(spec, item))
		}
		return out
	case string:
		if spec.Type == domain.FieldRef && t != "" {
			return "[[" + t + "]]"
		}
		return t
	default:
		return v
	}
}
