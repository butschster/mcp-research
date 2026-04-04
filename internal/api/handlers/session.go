package handlers

import (
	"log/slog"
	"net/http"

	"github.com/butschster/mcp-research/internal/service"
)

type SessionHandler struct {
	session *service.SessionService
	log     *slog.Logger
}

func NewSessionHandler(session *service.SessionService, log *slog.Logger) *SessionHandler {
	return &SessionHandler{session: session, log: log}
}

func (h *SessionHandler) ListByResearch(w http.ResponseWriter, r *http.Request) {
	researchID := r.PathValue("id")

	sessions, err := h.session.ListByResearch(r.Context(), researchID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"data":  sessions,
		"count": len(sessions),
	})
}

func (h *SessionHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	result, err := h.session.Get(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	// Group questions by status for the frontend
	grouped := make(map[string][]map[string]any)
	for _, q := range result.Questions {
		grouped[string(q.Status)] = append(grouped[string(q.Status)], map[string]any{
			"id":        q.ID,
			"text":      q.Text,
			"area":      q.Area,
			"rationale": q.Rationale,
			"priority":  q.Priority,
			"status":    q.Status,
			"answer":    q.Answer,
			"parent_id": q.ParentID,
			"position":  q.Position,
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"data": map[string]any{
			"session":   result.Session,
			"questions": grouped,
			"progress":  result.Progress,
		},
	})
}
