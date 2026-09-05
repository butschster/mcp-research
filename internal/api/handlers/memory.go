package handlers

import "net/http"

func (h *WriteHandler) ListMemory(w http.ResponseWriter, r *http.Request) {
	items, err := h.research.ListMemory(r.Context(), r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": items})
}

func (h *WriteHandler) AddMemory(w http.ResponseWriter, r *http.Request) {
	var input AddMemoryRequest
	if !decodeJSON(w, r, &input) {
		return
	}
	item, err := h.research.AddMemory(r.Context(), r.PathValue("id"), input.Text, input.SessionID)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": item})
}

func (h *WriteHandler) UpdateMemory(w http.ResponseWriter, r *http.Request) {
	var input UpdateMemoryRequest
	if !decodeJSON(w, r, &input) {
		return
	}
	if err := h.research.UpdateMemory(r.Context(), r.PathValue("id"), r.PathValue("itemId"), input.Text, input.Version); err != nil {
		writeServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *WriteHandler) DeleteMemory(w http.ResponseWriter, r *http.Request) {
	if err := h.research.DeleteMemory(r.Context(), r.PathValue("id"), []string{r.PathValue("itemId")}); err != nil {
		writeServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *WriteHandler) BulkDeleteMemory(w http.ResponseWriter, r *http.Request) {
	var input BulkDeleteMemoryRequest
	if !decodeJSON(w, r, &input) {
		return
	}
	if err := h.research.DeleteMemory(r.Context(), r.PathValue("id"), input.IDs); err != nil {
		writeServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
