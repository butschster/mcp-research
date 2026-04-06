package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/butschster/mcp-research/internal/domain"
	"github.com/butschster/mcp-research/internal/service"
	"github.com/butschster/mcp-research/internal/storage"
)

type ResearchHandler struct {
	research *service.ResearchService
	section  *service.SectionService
	entry    *service.EntryService
	session  *service.SessionService
	log      *slog.Logger
}

func NewResearchHandler(research *service.ResearchService, section *service.SectionService, entry *service.EntryService, session *service.SessionService, log *slog.Logger) *ResearchHandler {
	return &ResearchHandler{research: research, section: section, entry: entry, session: session, log: log}
}

func (h *ResearchHandler) List(w http.ResponseWriter, r *http.Request) {
	filter := storage.ResearchFilter{}
	if s := r.URL.Query().Get("status"); s != "" {
		status := domain.ResearchStatus(s)
		filter.Status = &status
	}

	researches, err := h.research.List(r.Context(), filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
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
		writeError(w, http.StatusInternalServerError, err.Error())
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

	var latestSession *domain.Session
	if h.session != nil {
		latestSession, _ = h.session.FindLatest(r.Context(), id)
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"data": map[string]any{
			"research":       research,
			"sections":       sectionData,
			"active_session": latestSession,
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
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"data":  entries,
		"count": len(entries),
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
