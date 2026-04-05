package handlers

import (
	"log/slog"
	"net/http"

	"github.com/butschster/mcp-research/internal/service"
	"github.com/butschster/mcp-research/internal/storage"
)

type TaskHandler struct {
	task     *service.TaskService
	research *service.ResearchService
	log      *slog.Logger
}

func NewTaskHandler(task *service.TaskService, research *service.ResearchService, log *slog.Logger) *TaskHandler {
	return &TaskHandler{task: task, research: research, log: log}
}

func (h *TaskHandler) ListByResearch(w http.ResponseWriter, r *http.Request) {
	researchID, err := h.research.ResolveID(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	tasks, err := h.task.List(r.Context(), researchID, storage.TaskFilter{})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	counts, _ := h.task.CountByStatus(r.Context(), researchID)

	writeJSON(w, http.StatusOK, map[string]any{
		"data":   tasks,
		"count":  len(tasks),
		"counts": counts,
	})
}
