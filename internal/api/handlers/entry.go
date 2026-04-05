package handlers

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/butschster/mcp-research/internal/domain"
	"github.com/butschster/mcp-research/internal/service"
	"github.com/butschster/mcp-research/internal/storage"
)

type EntryHandler struct {
	entry    *service.EntryService
	entries  *storage.EntryRepository
	research *storage.ResearchRepository
	log      *slog.Logger
}

func NewEntryHandler(entry *service.EntryService, entries *storage.EntryRepository, research *storage.ResearchRepository, log *slog.Logger) *EntryHandler {
	return &EntryHandler{entry: entry, entries: entries, research: research, log: log}
}

func (h *EntryHandler) Get(w http.ResponseWriter, r *http.Request) {
	idOrCode := r.PathValue("id")

	var entry *domain.Entry
	var err error

	// If research query param is provided, resolve entry by code within that research
	if researchIDOrCode := r.URL.Query().Get("research"); researchIDOrCode != "" {
		// Resolve research first
		research, rErr := h.research.FindByID(r.Context(), researchIDOrCode)
		if rErr == nil && research == nil {
			research, _ = h.research.FindByCode(r.Context(), researchIDOrCode)
		}
		if research != nil {
			entry, err = h.entry.GetByIDOrCode(r.Context(), research.ID, idOrCode)
		} else {
			err = fmt.Errorf("research not found")
		}
	} else {
		entry, err = h.entry.Get(r.Context(), idOrCode)
	}

	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"data": entry,
	})
}

// ResolveCode resolves a short code reference like "E3" (within research) or handles
// cross-research resolution via query param ?research_code=R2
func (h *EntryHandler) ResolveCode(w http.ResponseWriter, r *http.Request) {
	researchID := r.PathValue("id")
	entryCode := r.PathValue("code")

	entry, err := h.entries.FindByCode(r.Context(), researchID, entryCode)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if entry == nil {
		writeError(w, http.StatusNotFound, "entry not found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"data": entry,
	})
}

// ResolveResearchCode resolves a research short code to its ID and metadata.
func (h *EntryHandler) ResolveResearchCode(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")

	research, err := h.research.FindByCode(r.Context(), code)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if research == nil {
		writeError(w, http.StatusNotFound, "research not found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"data": map[string]any{
			"id":   research.ID,
			"code": research.Code,
			"name": research.Name,
		},
	})
}
