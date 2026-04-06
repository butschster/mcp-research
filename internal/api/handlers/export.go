package handlers

import (
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/butschster/mcp-research/internal/domain"
	"github.com/butschster/mcp-research/internal/service"
	"github.com/butschster/mcp-research/internal/storage"
)

type ExportHandler struct {
	research *service.ResearchService
	section  *service.SectionService
	entry    *service.EntryService
	entries  *storage.EntryRepository
	session  *service.SessionService
	task     *service.TaskService
	log      *slog.Logger
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

// Export returns all research data as JSON (for frontend rendering) and as markdown.
func (h *ExportHandler) Export(w http.ResponseWriter, r *http.Request) {
	idOrCode := r.PathValue("id")

	research, err := h.research.Get(r.Context(), idOrCode)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	researchID := research.ID

	sections, err := h.section.List(r.Context(), researchID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	allEntries, err := h.entries.FindByResearchWithContent(r.Context(), researchID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Group entries by section
	entriesBySection := make(map[string][]*domain.Entry)
	for _, e := range allEntries {
		entriesBySection[e.SectionID] = append(entriesBySection[e.SectionID], e)
	}

	sessions, err := h.session.ListByResearch(r.Context(), researchID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Fetch questions for each session
	var sessionExports []sessionWithQuestions
	for _, sess := range sessions {
		qs, _ := h.session.ListQuestions(r.Context(), sess.ID, storage.QuestionFilter{})
		sessionExports = append(sessionExports, sessionWithQuestions{Session: sess, Questions: qs})
	}

	tasks, err := h.task.List(r.Context(), researchID, storage.TaskFilter{})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Build markdown
	md := buildMarkdown(research, sections, entriesBySection, sessionExports, tasks)

	// Check if raw markdown requested
	if r.URL.Query().Get("format") == "md" {
		w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
		w.Header().Set("Content-Disposition",
			fmt.Sprintf(`attachment; filename="%s.md"`, sanitizeFilename(research.Name)))
		w.Write([]byte(md))
		return
	}

	// JSON response with structured data + markdown
	type sectionExport struct {
		*domain.Section
		Entries []*domain.Entry `json:"entries"`
	}
	var sectionData []sectionExport
	for _, s := range sections {
		sectionData = append(sectionData, sectionExport{
			Section: s,
			Entries: entriesBySection[s.ID],
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"research": research,
		"sections": sectionData,
		"sessions": sessionExports,
		"tasks":    tasks,
		"markdown": md,
	})
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
			if e.Content != "" {
				b.WriteString(e.Content + "\n\n")
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

func sanitizeFilename(name string) string {
	r := strings.NewReplacer("/", "-", "\\", "-", ":", "-", "\"", "", "'", "", " ", "_")
	return r.Replace(name)
}
