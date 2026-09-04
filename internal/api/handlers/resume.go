package handlers

import (
	"net/http"
	"strconv"

	"github.com/butschster/mcp-research/internal/service"
)

// ResumeHandler serves the continuation summary to the web UI.
//
// It is a read, and one the page makes on every visit, so it does exactly what
// the MCP tool does and nothing more: no session is started, nothing is marked
// as seen, no status moves. The personal new/changed queue stays a separate
// request precisely so that opening this page cannot silently acknowledge
// documents nobody has read.
type ResumeHandler struct {
	resume *service.ResumeService
}

func NewResumeHandler(resume *service.ResumeService) *ResumeHandler {
	return &ResumeHandler{resume: resume}
}

func (h *ResumeHandler) Get(w http.ResponseWriter, r *http.Request) {
	req := service.ResumeRequest{SessionID: r.URL.Query().Get("session_id")}

	// A malformed limit is refused rather than silently defaulted: a caller who
	// sent `limit=abc` has a bug, and answering with five items pretends they
	// asked for five. Out-of-range values are a different case — the service
	// clamps those, because "as many as you can" is a reasonable thing to mean.
	if v := r.URL.Query().Get("limit"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			writeError(w, http.StatusBadRequest, "limit must be a number")
			return
		}
		req.Limit = n
	}

	// The id in the path may be a UUID or an R-code; the service resolves it,
	// checks access and refuses a share link. Errors go through the shared
	// mapper so an unknown research and someone else's research read alike.
	resume, err := h.resume.Get(r.Context(), r.PathValue("id"), req)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"data": resume})
}
