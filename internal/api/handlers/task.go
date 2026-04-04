package handlers

import (
	"log/slog"
	"net/http"

	"github.com/butschster/mcp-research/internal/service"
	"github.com/butschster/mcp-research/internal/storage"
)

type TaskHandler struct {
	task *service.TaskService
	log  *slog.Logger
}

func NewTaskHandler(task *service.TaskService, log *slog.Logger) *TaskHandler {
	return &TaskHandler{task: task, log: log}
}

func (h *TaskHandler) ListByResearch(w http.ResponseWriter, r *http.Request) {
	researchID := r.PathValue("id")

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
