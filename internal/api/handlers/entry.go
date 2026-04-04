package handlers

import (
	"log/slog"
	"net/http"

	"github.com/butschster/mcp-research/internal/service"
)

type EntryHandler struct {
	entry *service.EntryService
	log   *slog.Logger
}

func NewEntryHandler(entry *service.EntryService, log *slog.Logger) *EntryHandler {
	return &EntryHandler{entry: entry, log: log}
}

func (h *EntryHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	entry, err := h.entry.Get(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"data": entry,
	})
}
