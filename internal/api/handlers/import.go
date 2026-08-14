package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/butschster/mcp-research/internal/domain"
	"github.com/butschster/mcp-research/internal/service"
)

type ImportHandler struct {
	export *service.ExportService
	log    *slog.Logger
}

func NewImportHandler(export *service.ExportService, log *slog.Logger) *ImportHandler {
	return &ImportHandler{export: export, log: log}
}

func (h *ImportHandler) Import(w http.ResponseWriter, r *http.Request) {
	var data domain.ExportData
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}

	research, err := h.export.Import(r.Context(), &data, r.URL.Query().Get("team"))
	if err != nil {
		h.log.Error("import failed", "error", err)
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"status":      "imported",
		"research_id": research.ID,
		"code":        research.Code,
		"name":        research.Name,
	})
}
