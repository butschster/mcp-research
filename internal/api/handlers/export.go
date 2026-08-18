package handlers

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/butschster/mcp-research/internal/auth"
	"github.com/butschster/mcp-research/internal/domain"
	"github.com/butschster/mcp-research/internal/service"
	"github.com/butschster/mcp-research/internal/storage"
)

type ExportHandler struct {
	research  *service.ResearchService
	section   *service.SectionService
	entry     *service.EntryService
	entries   *storage.EntryRepository
	session   *service.SessionService
	task      *service.TaskService
	exportSvc *service.ExportService
	obsidian  *service.ObsidianService
	roadmap   *service.RoadmapService
	log       *slog.Logger
}

func NewExportHandler(
	research *service.ResearchService,
	section *service.SectionService,
	entry *service.EntryService,
	entries *storage.EntryRepository,
	session *service.SessionService,
	task *service.TaskService,
	log *slog.Logger,
) *ExportHandler {
	return &ExportHandler{
		research: research, section: section,
		entry: entry, entries: entries,
		session: session, task: task, log: log,
	}
}

func (h *ExportHandler) SetExportService(svc *service.ExportService) {
	h.exportSvc = svc
}

// SetObsidianService enables `?format=obsidian`. Without it the route keeps
// working and the format reports itself as unconfigured.
func (h *ExportHandler) SetObsidianService(svc *service.ObsidianService) {
	h.obsidian = svc
}

// SetRoadmapService lets the export payload carry a roadmap count, which is
// what the download dialog needs to know whether the option applies.
func (h *ExportHandler) SetRoadmapService(svc *service.RoadmapService) {
	h.roadmap = svc
}

// Export returns all research data as JSON (for frontend rendering) and as markdown.
func (h *ExportHandler) Export(w http.ResponseWriter, r *http.Request) {
	idOrCode := r.PathValue("id")

	// The vault is offered to a share visitor too, on a link that includes
	// downloading. It used to be refused here on the grounds that it builds its
	// own payload rather than reusing this handler's — but it builds it from
	// `ResearchService.Get`, which is exactly where `instruction` and `memory`
	// are redacted, so the research it starts from was already the published
	// one. What genuinely was missing is that the vault's parts come from a
	// query string the visitor can type, and the include flags never saw it;
	// `service.clampForShare` narrows them now, next to the code that reads
	// them.
	share := auth.ShareFromContext(r.Context())
	if r.URL.Query().Get("format") == "obsidian" {
		// The vault export reads the research itself and does its own access
		// check, so it branches before this handler fetches anything.
		h.ExportObsidian(w, r)
		return
	}

	research, err := h.research.Get(r.Context(), idOrCode)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	researchID := research.ID

	sections, err := h.section.List(r.Context(), researchID)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	allEntries, err := h.entry.ListWithContent(r.Context(), researchID)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	// Group entries by section
	entriesBySection := make(map[string][]*domain.Entry)
	for _, e := range allEntries {
		entriesBySection[e.SectionID] = append(entriesBySection[e.SectionID], e)
	}

	// The include flags gate the routes that serve sessions and tasks, and they
	// have to gate this too: an export that carried the interview transcript
	// would hand over in one file exactly what the creator chose to leave out
	// of the pages.
	var sessionExports []sessionWithQuestions
	if share == nil || share.Include.Sessions {
		sessions, err := h.session.ListByResearch(r.Context(), researchID)
		if err != nil {
			writeServiceError(w, err)
			return
		}
		for _, sess := range sessions {
			qs, _ := h.session.ListQuestions(r.Context(), sess.ID, storage.QuestionFilter{})
			sessionExports = append(sessionExports, sessionWithQuestions{Session: sess, Questions: qs})
		}
	}

	var tasks []*domain.Task
	if share == nil || share.Include.Tasks {
		tasks, err = h.task.List(r.Context(), researchID, storage.TaskFilter{})
		if err != nil {
			writeServiceError(w, err)
			return
		}
	}

	// Build markdown
	md := buildMarkdown(research, sections, entriesBySection, sessionExports, tasks)

	// Check if raw markdown requested
	if r.URL.Query().Get("format") == "md" {
		w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
		w.Header().Set("Content-Disposition",
			contentDisposition(sanitizeFilename(research.Name)+".md"))
		w.Write([]byte(md))
		return
	}

	// JSON response with structured data + markdown
	type sectionExport struct {
		*domain.Section
		Entries []entryExport `json:"entries"`
	}
	var sectionData []sectionExport
	for _, s := range sections {
		sectionData = append(sectionData, sectionExport{
			Section: s,
			Entries: exportEntries(entriesBySection[s.ID]),
		})
	}

	// A count rather than the roadmaps themselves: this payload is already the
	// whole research, and the only caller that needs roadmaps here is a dialog
	// asking whether the option applies at all.
	//
	// It obeys the include flag too. "There is 1 roadmap" is exactly the fact a
	// link created with roadmaps switched off exists to withhold.
	roadmapCount := 0
	if h.roadmap != nil && (share == nil || share.Include.Roadmaps) {
		if rms, err := h.roadmap.List(r.Context(), researchID); err == nil {
			roadmapCount = len(rms)
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"research":      research,
		"sections":      sectionData,
		"sessions":      sessionExports,
		"tasks":         tasks,
		"roadmap_count": roadmapCount,
		"markdown":      md,
	})
}

// ExportSession returns a single session (questions, notes, linked entries) as JSON
// and as markdown. Mirrors Export, but scoped to one session.
func (h *ExportHandler) ExportSession(w http.ResponseWriter, r *http.Request) {
	research, err := h.research.Get(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	sess, err := h.session.GetByIDOrCode(r.Context(), research.ID, r.PathValue("sessionId"))
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	questions, err := h.session.ListQuestions(r.Context(), sess.Session.ID, storage.QuestionFilter{})
	if err != nil {
		writeServiceError(w, err)
		return
	}

	// Entries produced during this session, grouped by their section. Through
	// the service, not the repository: the session export is not reachable from
	// a share today, and the first thing anyone will do when it is reachable is
	// mount the route rather than re-audit this read.
	allEntries, err := h.entry.ListWithContent(r.Context(), research.ID)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	var sessionEntries []*domain.Entry
	for _, e := range allEntries {
		if e.SessionID == sess.Session.ID {
			sessionEntries = append(sessionEntries, e)
		}
	}

	sections, err := h.section.List(r.Context(), research.ID)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	sectionNames := make(map[string]string, len(sections))
	for _, s := range sections {
		name := s.DisplayName
		if name == "" {
			name = s.Name
		}
		sectionNames[s.ID] = name
	}

	md := buildSessionMarkdown(research, sess.Session, questions, sessionEntries, sectionNames)

	if r.URL.Query().Get("format") == "md" {
		w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
		w.Header().Set("Content-Disposition",
			contentDisposition(sanitizeFilename(sess.Session.Title)+".md"))
		w.Write([]byte(md))
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"research":      research,
		"session":       sess.Session,
		"questions":     questions,
		"entries":       exportEntries(sessionEntries),
		"section_names": sectionNames,
		"markdown":      md,
	})
}

// ExportPortable returns a complete research as portable JSON (for import on another server).
func (h *ExportHandler) ExportPortable(w http.ResponseWriter, r *http.Request) {
	if h.exportSvc == nil {
		writeError(w, http.StatusInternalServerError, "export service not configured")
		return
	}

	idOrCode := r.PathValue("id")
	data, err := h.exportSvc.Export(r.Context(), idOrCode)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Content-Disposition",
		contentDisposition(sanitizeFilename(data.Research.Name)+".json"))
	writeJSON(w, http.StatusOK, data)
}

type sessionWithQuestions struct {
	*domain.Session
	Questions []*domain.Question `json:"questions"`
}

func buildMarkdown(
	research *domain.Research,
	sections []*domain.Section,
	entriesBySection map[string][]*domain.Entry,
	sessions []sessionWithQuestions,
	tasks []*domain.Task,
) string {
	var b strings.Builder

	// Title
	b.WriteString(fmt.Sprintf("# %s\n\n", research.Name))
	if research.Goal != "" {
		b.WriteString(fmt.Sprintf("> %s\n\n", research.Goal))
	}
	if research.Description != "" {
		b.WriteString(research.Description + "\n\n")
	}
	if len(research.Tags) > 0 {
		b.WriteString("**Tags:** " + strings.Join(research.Tags, ", ") + "\n\n")
	}

	b.WriteString("---\n\n")

	// Table of contents
	b.WriteString("## Table of Contents\n\n")
	for i, s := range sections {
		entries := entriesBySection[s.ID]
		name := s.DisplayName
		if name == "" {
			name = s.Name
		}
		b.WriteString(fmt.Sprintf("%d. **%s** (%d entries)\n", i+1, name, len(entries)))
	}
	if len(sessions) > 0 {
		b.WriteString(fmt.Sprintf("%d. **Sessions** (%d)\n", len(sections)+1, len(sessions)))
	}
	if len(tasks) > 0 {
		b.WriteString(fmt.Sprintf("%d. **Tasks** (%d)\n", len(sections)+2, len(tasks)))
	}
	b.WriteString("\n---\n\n")

	// Sections + entries
	for _, s := range sections {
		name := s.DisplayName
		if name == "" {
			name = s.Name
		}
		b.WriteString(fmt.Sprintf("## %s\n\n", name))
		if s.Description != "" {
			b.WriteString(s.Description + "\n\n")
		}

		entries := entriesBySection[s.ID]
		if len(entries) == 0 {
			b.WriteString("*No entries yet.*\n\n")
			continue
		}

		for _, e := range entries {
			b.WriteString(fmt.Sprintf("### %s", e.Title))
			if e.Code != "" {
				b.WriteString(fmt.Sprintf(" [%s]", e.Code))
			}
			b.WriteString("\n\n")
			if len(e.Tags) > 0 {
				b.WriteString("**Tags:** " + strings.Join(e.Tags, ", ") + "  \n")
			}
			b.WriteString(fmt.Sprintf("**Status:** %s\n\n", e.Status))
			// Only the fields the section declares, in the order it declares
			// them, so the same document reads the same way in every export.
			// Values under keys the section has since dropped are kept in the
			// database but not printed: they are no longer what this section
			// says a document records.
			if md := metadataMarkdown(s.FieldSpec, e.Metadata); md != "" {
				b.WriteString(md)
			}
			if body := entryMarkdown(e); body != "" {
				b.WriteString(body + "\n\n")
			}
		}
	}

	// Sessions + questions
	if len(sessions) > 0 {
		b.WriteString("---\n\n## Sessions\n\n")
		for _, sess := range sessions {
			b.WriteString(fmt.Sprintf("### %s\n\n", sess.Title))
			if sess.Focus != "" {
				b.WriteString(fmt.Sprintf("**Focus:** %s\n\n", sess.Focus))
			}
			if sess.Notes != "" {
				b.WriteString(fmt.Sprintf("**Notes:** %s\n\n", sess.Notes))
			}

			if len(sess.Questions) == 0 {
				continue
			}

			for _, q := range sess.Questions {
				status := string(q.Status)
				b.WriteString(fmt.Sprintf("#### Q: %s\n\n", q.Text))
				if q.Area != "" {
					b.WriteString(fmt.Sprintf("**Area:** %s | ", q.Area))
				}
				b.WriteString(fmt.Sprintf("**Priority:** %s | **Status:** %s\n\n", q.Priority, status))
				if q.Answer != "" {
					b.WriteString(q.Answer + "\n\n")
				}
			}
		}
	}

	// Tasks
	if len(tasks) > 0 {
		b.WriteString("---\n\n## Tasks\n\n")
		for _, t := range tasks {
			check := "[ ]"
			if t.Status == domain.TaskCompleted {
				check = "[x]"
			} else if t.Status == domain.TaskFailed {
				check = "[!]"
			}
			b.WriteString(fmt.Sprintf("- %s **%s**", check, t.Title))
			if t.Priority != "" {
				b.WriteString(fmt.Sprintf(" (%s)", t.Priority))
			}
			b.WriteString("\n")
			if t.Description != "" {
				b.WriteString(fmt.Sprintf("  - %s\n", t.Description))
			}
			if t.Result != "" {
				b.WriteString(fmt.Sprintf("  - **Result:** %s\n", t.Result))
			}
		}
		b.WriteString("\n")
	}

	return b.String()
}

func buildSessionMarkdown(
	research *domain.Research,
	sess *domain.Session,
	questions []*domain.Question,
	entries []*domain.Entry,
	sectionNames map[string]string,
) string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("# %s\n\n", sess.Title))
	b.WriteString(fmt.Sprintf("*Session of research: %s*\n\n", research.Name))
	if sess.Code != "" {
		b.WriteString(fmt.Sprintf("**Code:** %s  \n", sess.Code))
	}
	b.WriteString(fmt.Sprintf("**Status:** %s\n\n", sess.Status))
	if sess.Focus != "" {
		b.WriteString(fmt.Sprintf("> %s\n\n", sess.Focus))
	}
	if sess.Notes != "" {
		b.WriteString("## Notes\n\n" + sess.Notes + "\n\n")
	}

	b.WriteString("---\n\n")

	answered := 0
	for _, q := range questions {
		if q.Status == domain.QuestionAnswered {
			answered++
		}
	}
	b.WriteString(fmt.Sprintf("## Questions (%d answered of %d)\n\n", answered, len(questions)))

	if len(questions) == 0 {
		b.WriteString("*No questions in this session.*\n\n")
	}
	for _, q := range questions {
		b.WriteString(fmt.Sprintf("### Q: %s\n\n", q.Text))
		if q.Area != "" {
			b.WriteString(fmt.Sprintf("**Area:** %s | ", q.Area))
		}
		b.WriteString(fmt.Sprintf("**Priority:** %s | **Status:** %s\n\n", q.Priority, q.Status))
		if q.Rationale != "" {
			b.WriteString(fmt.Sprintf("*%s*\n\n", q.Rationale))
		}
		if q.Answer != "" {
			b.WriteString(q.Answer + "\n\n")
		}
	}

	if len(entries) > 0 {
		b.WriteString("---\n\n## Entries produced in this session\n\n")
		for _, e := range entries {
			b.WriteString(fmt.Sprintf("### %s", e.Title))
			if e.Code != "" {
				b.WriteString(fmt.Sprintf(" [%s]", e.Code))
			}
			b.WriteString("\n\n")
			if name := sectionNames[e.SectionID]; name != "" {
				b.WriteString(fmt.Sprintf("**Section:** %s  \n", name))
			}
			if len(e.Tags) > 0 {
				b.WriteString("**Tags:** " + strings.Join(e.Tags, ", ") + "  \n")
			}
			b.WriteString(fmt.Sprintf("**Status:** %s\n\n", e.Status))
			if body := entryMarkdown(e); body != "" {
				b.WriteString(body + "\n\n")
			}
		}
	}

	return b.String()
}

// entryMarkdown renders an entry's body for a markdown export.
//
// A blocks entry stores JSON, so writing Content straight out would put a
// serialized document into the file where prose belongs. An html block is named
// rather than inlined: a wall of markup in an export is not readable, and it is
// not markdown either. Legacy `artifact` rows — those written before migration
// 016 ran — are treated the same way.
func entryMarkdown(e *domain.Entry) string {
	if e == nil || e.Content == "" {
		return ""
	}
	switch e.Type {
	case domain.EntryBlocks:
		doc, err := service.ParseStoredBlockDocument(e.Content)
		if err != nil {
			// Unreadable document: say so rather than emitting JSON or nothing.
			return "*This entry holds a block document that could not be read.*"
		}
		return strings.TrimRight(service.BlockDocumentToMarkdown(doc), "\n")
	case domain.EntryArtifact:
		return "*HTML artifact — view it in the web UI.*"
	default:
		return e.Content
	}
}

// entryExport carries the entry plus, for block documents, a markdown rendering
// so the export pages have something to display without duplicating the
// serializer in TypeScript.
type entryExport struct {
	*domain.Entry
	ContentMarkdown string `json:"content_markdown,omitempty"`
}

func exportEntries(entries []*domain.Entry) []entryExport {
	out := make([]entryExport, 0, len(entries))
	for _, e := range entries {
		ee := entryExport{Entry: e}
		if e.Type == domain.EntryBlocks || e.Type == domain.EntryArtifact {
			ee.ContentMarkdown = entryMarkdown(e)
		}
		out = append(out, ee)
	}
	return out
}

func sanitizeFilename(name string) string {
	r := strings.NewReplacer("/", "-", "\\", "-", ":", "-", "\"", "", "'", "", " ", "_")
	return r.Replace(name)
}

// metadataMarkdown renders a document's declared fields as a short list above
// its body.
//
// A declared field with no value is printed as an em dash rather than skipped.
// A reader of the exported file has no other way to learn that the field exists
// and nobody answered it, and that gap is usually the more interesting half —
// it is what "this document is incomplete" looks like on paper.
func metadataMarkdown(specs []domain.FieldSpec, values map[string]any) string {
	if len(specs) == 0 {
		return ""
	}
	var b strings.Builder
	for _, f := range specs {
		label := f.Label
		if label == "" {
			label = f.Key
		}
		v, recorded := values[f.Key]
		// "unknown" and "—" are different sentences, and only one of them means
		// somebody looked. Collapsing them here would have made a required field
		// answered with an explicit unknown read as complete in the app and
		// blank on paper.
		text := "—"
		switch {
		case recorded && v == nil:
			text = "unknown"
		case recorded:
			text = formatMetadataValue(v)
		}
		b.WriteString("**" + label + ":** " + text + "  \n")
	}
	return b.String() + "\n"
}

func formatMetadataValue(v any) string {
	switch t := v.(type) {
	case nil:
		return "—"
	case string:
		if strings.TrimSpace(t) == "" {
			return "—"
		}
		return t
	case float64:
		return strconv.FormatFloat(t, 'f', -1, 64)
	case []any:
		parts := make([]string, 0, len(t))
		for _, item := range t {
			parts = append(parts, formatMetadataValue(item))
		}
		if len(parts) == 0 {
			return "—"
		}
		return strings.Join(parts, ", ")
	default:
		return fmt.Sprint(t)
	}
}

// EntryMarkdown serves one document as a markdown file.
//
// The whole-research exports answer "give me everything"; this answers "give me
// this one", which until now had no answer at all.
func (h *ExportHandler) EntryMarkdown(w http.ResponseWriter, r *http.Request) {
	file, err := h.entry.MarkdownExport(r.Context(), r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}

	w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
	w.Header().Set("Content-Disposition", contentDisposition(file.Filename))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(file.Content))
}
