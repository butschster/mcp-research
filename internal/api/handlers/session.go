package handlers

import (
	"log/slog"
	"net/http"

	"github.com/butschster/mcp-research/internal/service"
	"github.com/butschster/mcp-research/internal/storage"
)

type SessionHandler struct {
	session  *service.SessionService
	entry    *service.EntryService
	research *service.ResearchService
	log      *slog.Logger
}

func NewSessionHandler(session *service.SessionService, entry *service.EntryService, research *service.ResearchService, log *slog.Logger) *SessionHandler {
	return &SessionHandler{session: session, entry: entry, research: research, log: log}
}

func (h *SessionHandler) ListByResearch(w http.ResponseWriter, r *http.Request) {
	researchID, err := h.research.ResolveID(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

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
	researchID, err := h.research.ResolveID(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	sessionIDOrCode := r.PathValue("sessionId")
	result, err := h.session.GetByIDOrCode(r.Context(), researchID, sessionIDOrCode)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	grouped := make(map[string][]map[string]any)
	for _, q := range result.Questions {
		grouped[string(q.Status)] = append(grouped[string(q.Status)], map[string]any{
			"id":        q.ID,
			"code":      q.Code,
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

	// Fetch entries linked to this session
	entries, _ := h.entry.ListByResearch(r.Context(), researchID, storage.EntryFilter{
		SessionID: result.Session.ID,
	})

	writeJSON(w, http.StatusOK, map[string]any{
		"data": map[string]any{
			"session":   result.Session,
			"questions": grouped,
			"progress":  result.Progress,
			"entries":   entries,
		},
	})
}
