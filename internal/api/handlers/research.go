package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/butschster/mcp-research/internal/auth"
	"github.com/butschster/mcp-research/internal/domain"
	"github.com/butschster/mcp-research/internal/service"
	"github.com/butschster/mcp-research/internal/storage"
)

type ResearchHandler struct {
	research *service.ResearchService
	section  *service.SectionService
	entry    *service.EntryService
	entries  *storage.EntryRepository
	session  *service.SessionService
	shares   *service.ShareService
	log      *slog.Logger
}

func NewResearchHandler(research *service.ResearchService, section *service.SectionService, entry *service.EntryService, entries *storage.EntryRepository, session *service.SessionService, log *slog.Logger) *ResearchHandler {
	return &ResearchHandler{research: research, section: section, entry: entry, entries: entries, session: session, log: log}
}

// SetShareService adds the live-share count to the research payload. It is set
// after construction so the handler keeps working without it — the count is a
// badge, not the page.
func (h *ResearchHandler) SetShareService(svc *service.ShareService) {
	h.shares = svc
}

func (h *ResearchHandler) List(w http.ResponseWriter, r *http.Request) {
	filter := storage.ResearchFilter{}
	if s := r.URL.Query().Get("status"); s != "" {
		status := domain.ResearchStatus(s)
		filter.Status = &status
	}
	// `?team=` narrows the list to one team. It is a filter, not a scope: the
	// list is flat across every team the caller belongs to, and this is how a
	// reader with several of them looks at one at a time.
	if t := r.URL.Query().Get("team"); t != "" {
		filter.TeamID = &t
	}

	researches, err := h.research.List(r.Context(), filter)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"data":  researches,
		"count": len(researches),
	})
}

func (h *ResearchHandler) Get(w http.ResponseWriter, r *http.Request) {
	idOrCode := r.PathValue("id")

	research, err := h.research.Get(r.Context(), idOrCode)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	id := research.ID
	sections, err := h.section.List(r.Context(), id)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	type sectionWithCount struct {
		*domain.Section
		EntriesCount int `json:"entries_count"`
	}

	var sectionData []sectionWithCount
	for _, s := range sections {
		count, _ := h.section.CountEntries(r.Context(), s.ID)
		sectionData = append(sectionData, sectionWithCount{
			Section:      s,
			EntriesCount: count,
		})
	}

	// The newest session, unless the link that is asking excluded them.
	//
	// This route is not gated by the include flags — it is the research itself,
	// which every link carries — so the one part of its payload that is optional
	// has to gate itself. Without this a link created with sessions switched off
	// still handed over the latest session's title, focus and free-form notes on
	// the very first request the shared page makes.
	var latestSession *domain.Session
	if h.session != nil {
		sc := auth.ShareFromContext(r.Context())
		if sc == nil || sc.Include.Sessions {
			latestSession, _ = h.session.FindLatest(r.Context(), id)
		}
	}

	// How many links are handing this research out right now. Zero for anyone
	// who could not manage them anyway, which includes a share visitor reading
	// the page through one of them.
	activeShares := 0
	if h.shares != nil {
		activeShares = h.shares.ActiveCount(r.Context(), id)
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"data": map[string]any{
			"research":           research,
			"sections":           sectionData,
			"active_session":     latestSession,
			"active_share_count": activeShares,
		},
	})
}

func (h *ResearchHandler) ListSectionEntries(w http.ResponseWriter, r *http.Request) {
	idOrCode := r.PathValue("id")
	sectionID := r.PathValue("sectionId")

	researchID, err := h.research.ResolveID(r.Context(), idOrCode)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	entries, err := h.entry.List(r.Context(), researchID, sectionID, storage.EntryFilter{})
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"data":  entries,
		"count": len(entries),
	})
}

// ListAllEntries returns all entries for a research, with optional ?tag= filter.
func (h *ResearchHandler) ListAllEntries(w http.ResponseWriter, r *http.Request) {
	idOrCode := r.PathValue("id")

	researchID, err := h.research.ResolveID(r.Context(), idOrCode)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	filter := storage.EntryFilter{
		Tag:       r.URL.Query().Get("tag"),
		SessionID: r.URL.Query().Get("session"),
	}

	entries, err := h.entry.ListByResearch(r.Context(), researchID, filter)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"data":  entries,
		"count": len(entries),
	})
}

// ListTags returns all unique tags in a research with their entry counts.
func (h *ResearchHandler) ListTags(w http.ResponseWriter, r *http.Request) {
	idOrCode := r.PathValue("id")

	researchID, err := h.research.ResolveID(r.Context(), idOrCode)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	tags, err := h.entries.FindTagsByResearch(r.Context(), researchID)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"data":  tags,
		"count": len(tags),
	})
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]any{"error": msg})
}
