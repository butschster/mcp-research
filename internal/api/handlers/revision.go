package handlers

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/butschster/mcp-research/internal/domain"
	"github.com/butschster/mcp-research/internal/service"
	"github.com/google/uuid"
)

// RevisionHandler serves an entry's history, the comparison of two of its
// revisions, restoration, and the roll-up of what a session changed.
type RevisionHandler struct {
	entry    *service.EntryService
	session  *service.SessionService
	research *service.ResearchService
	log      *slog.Logger
}

func NewRevisionHandler(entry *service.EntryService, session *service.SessionService, research *service.ResearchService, log *slog.Logger) *RevisionHandler {
	return &RevisionHandler{entry: entry, session: session, research: research, log: log}
}

// List returns an entry's revisions, newest first, without their content.
func (h *RevisionHandler) List(w http.ResponseWriter, r *http.Request) {
	entry, revs, err := h.entry.History(r.Context(), r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"data": map[string]any{
			"entry_id":   entry.ID,
			"entry_code": entry.Code,
			"title":      entry.Title,
			"revisions":  revs,
			"current":    currentRevision(revs),
		},
	})
}

// Get returns one revision with its content.
func (h *RevisionHandler) Get(w http.ResponseWriter, r *http.Request) {
	number, ok := revisionNumber(w, r.PathValue("revision"))
	if !ok {
		return
	}
	rev, err := h.entry.Revision(r.Context(), r.PathValue("id"), number)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": rev})
}

// Diff compares two revisions. Both bounds are optional: `to` defaults to the
// newest revision and `from` to the one before it, which makes the bare URL
// answer "what changed last".
func (h *RevisionHandler) Diff(w http.ResponseWriter, r *http.Request) {
	from, err := optionalInt(r.URL.Query().Get("from"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "from must be a revision number")
		return
	}
	to, err := optionalInt(r.URL.Query().Get("to"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "to must be a revision number")
		return
	}

	diff, derr := h.entry.Diff(r.Context(), r.PathValue("id"), from, to)
	if derr != nil {
		writeServiceError(w, derr)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": diff})
}

// Restore writes an earlier revision's content back onto the entry, as a new
// revision. History is never rewritten.
func (h *RevisionHandler) Restore(w http.ResponseWriter, r *http.Request) {
	number, ok := revisionNumber(w, r.PathValue("revision"))
	if !ok {
		return
	}
	entry, err := h.entry.Restore(r.Context(), r.PathValue("id"), number)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"data": map[string]any{
			"entry_id":          entry.ID,
			"entry_code":        entry.Code,
			"title":             entry.Title,
			"restored_revision": number,
		},
	})
}

// SessionChanges reports every entry a session created or modified.
//
// The path segment is a session UUID. A short code (`SS1`) is per research and
// therefore ambiguous on a route that names no research — resolving one globally
// would hand a caller whichever research got that code first, so a code is
// accepted only alongside `?research=`, which scopes it the way every other
// code-bearing route does.
func (h *RevisionHandler) SessionChanges(w http.ResponseWriter, r *http.Request) {
	idOrCode := r.PathValue("id")

	// Resolving the session first is what scopes this to a research the caller
	// may read: the session service refuses one they may not, and without that
	// step the revision query would be reachable by session id alone.
	var (
		sess *service.SessionWithQuestions
		err  error
	)
	if researchIDOrCode := r.URL.Query().Get("research"); researchIDOrCode != "" {
		var researchID string
		researchID, err = h.research.ResolveID(r.Context(), researchIDOrCode)
		if err != nil {
			writeServiceError(w, err)
			return
		}
		sess, err = h.session.GetByIDOrCode(r.Context(), researchID, idOrCode)
	} else if _, perr := uuid.Parse(idOrCode); perr != nil {
		// A code with no research to scope it by. Same 404 as a session that
		// does not exist — the caller learns nothing either way.
		writeError(w, http.StatusNotFound, "not found")
		return
	} else {
		sess, err = h.session.Get(r.Context(), idOrCode)
	}
	if err != nil {
		writeServiceError(w, err)
		return
	}

	// ?summary=1 answers the tab badge — how many entries this session touched —
	// without diffing any of them. The session page asks this on every load; the
	// full list, which costs an LCS per touched entry, waits until somebody
	// actually opens the tab.
	if r.URL.Query().Get("summary") == "1" {
		created, modified, err := h.entry.SessionChangeCounts(r.Context(), sess.Session.ResearchID, sess.Session.ID)
		if err != nil {
			writeServiceError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"data": map[string]any{
				"session_id": sess.Session.ID,
				"created":    created,
				"modified":   modified,
				"count":      created + modified,
			},
		})
		return
	}

	changes, err := h.entry.SessionChanges(r.Context(), sess.Session.ResearchID, sess.Session.ID)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	created, modified := 0, 0
	for _, c := range changes {
		if c.Created {
			created++
		} else {
			modified++
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"data": map[string]any{
			"session_id":    sess.Session.ID,
			"session_code":  sess.Session.Code,
			"session_title": sess.Session.Title,
			"created":       created,
			"modified":      modified,
			"changes":       changes,
		},
	})
}

func currentRevision(revs []*domain.EntryRevision) int {
	if len(revs) == 0 {
		return 0
	}
	return revs[0].Revision
}

func revisionNumber(w http.ResponseWriter, raw string) (int, bool) {
	n, err := strconv.Atoi(raw)
	if err != nil || n < 1 {
		writeError(w, http.StatusBadRequest, "revision must be a positive number")
		return 0, false
	}
	return n, true
}

func optionalInt(raw string) (int, error) {
	if raw == "" {
		return 0, nil
	}
	return strconv.Atoi(raw)
}
