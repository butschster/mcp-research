package handlers

import (
	"net/http"

	"github.com/dovod-app/app/internal/domain"
	"github.com/dovod-app/app/internal/service"
)

// EntryViewHandler serves the personal, research-scoped queue of documents a
// reader has not seen at their current revision.
type EntryViewHandler struct {
	views    *service.EntryViewService
	research *service.ResearchService
}

func NewEntryViewHandler(views *service.EntryViewService, research *service.ResearchService) *EntryViewHandler {
	return &EntryViewHandler{views: views, research: research}
}

func (h *EntryViewHandler) List(w http.ResponseWriter, r *http.Request) {
	researchID, err := h.research.ResolveID(r.Context(), r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	updates, err := h.views.List(r.Context(), researchID)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": updates})
}

func (h *EntryViewHandler) MarkSeen(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Revision int `json:"revision"`
	}
	if !decodeJSON(w, r, &input) {
		return
	}
	if err := h.views.MarkSeen(r.Context(), r.PathValue("id"), input.Revision); err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"data": map[string]any{
			"entry_id": r.PathValue("id"),
			"revision": input.Revision,
		},
	})
}

func (h *EntryViewHandler) MarkAllSeen(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Entries []domain.SeenRevision `json:"entries"`
	}
	if !decodeJSON(w, r, &input) {
		return
	}
	researchID, err := h.research.ResolveID(r.Context(), r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	if err := h.views.MarkSeenMany(r.Context(), researchID, input.Entries); err != nil {
		writeServiceError(w, err)
		return
	}
	marked := make(map[string]struct{}, len(input.Entries))
	for _, target := range input.Entries {
		marked[target.EntryID] = struct{}{}
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"data": map[string]any{"marked": len(marked)},
	})
}
