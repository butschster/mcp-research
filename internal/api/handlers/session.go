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

	writeJSON(w, http.StatusOK, map[string]any{
		"data": result,
	})
}
