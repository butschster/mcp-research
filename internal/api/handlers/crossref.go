package handlers

import (
	"log/slog"
	"net/http"

	"github.com/butschster/mcp-research/internal/service"
	"github.com/butschster/mcp-research/internal/storage"
)

type CrossRefHandler struct {
	crossrefs   *storage.CrossRefRepository
	entrySvc    *service.EntryService
	researchSvc *service.ResearchService
	log         *slog.Logger
}

func NewCrossRefHandler(crossrefs *storage.CrossRefRepository, entrySvc *service.EntryService, researchSvc *service.ResearchService, log *slog.Logger) *CrossRefHandler {
	return &CrossRefHandler{crossrefs: crossrefs, entrySvc: entrySvc, researchSvc: researchSvc, log: log}
}

// ListForResearch returns all stored cross-references for a research.
func (h *CrossRefHandler) ListForResearch(w http.ResponseWriter, r *http.Request) {
	researchID, err := h.researchSvc.ResolveID(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	refs, err := h.crossrefs.FindByResearch(r.Context(), researchID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"data":  refs,
		"count": len(refs),
	})
}

// Rebuild rescans all entries in a research and rebuilds cross-references.
func (h *CrossRefHandler) Rebuild(w http.ResponseWriter, r *http.Request) {
	researchID, err := h.researchSvc.ResolveID(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	count, err := h.entrySvc.RebuildCrossRefs(r.Context(), researchID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"rebuilt": count,
		"status":  "ok",
	})
}
