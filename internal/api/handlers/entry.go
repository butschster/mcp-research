package handlers

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/butschster/mcp-research/internal/auth"
	"github.com/butschster/mcp-research/internal/domain"
	"github.com/butschster/mcp-research/internal/service"
	"github.com/butschster/mcp-research/internal/storage"
)

type EntryHandler struct {
	entry       *service.EntryService
	researchSvc *service.ResearchService
	entries     *storage.EntryRepository
	research    *storage.ResearchRepository
	// users resolves the name behind a revision's user id, for the one line the
	// page shows about who wrote a document. Optional: with `auth_enabled:
	// false` there are no users at all, and the provenance line falls back to
	// the author kind it has always shown.
	users *storage.UserRepository
	log   *slog.Logger
}

func NewEntryHandler(entry *service.EntryService, researchSvc *service.ResearchService, entries *storage.EntryRepository, research *storage.ResearchRepository, users *storage.UserRepository, log *slog.Logger) *EntryHandler {
	return &EntryHandler{entry: entry, researchSvc: researchSvc, entries: entries, research: research, users: users, log: log}
}

func (h *EntryHandler) Get(w http.ResponseWriter, r *http.Request) {
	idOrCode := r.PathValue("id")

	var entry *domain.Entry
	var err error

	// If research query param is provided, resolve entry by code within that research
	if researchIDOrCode := r.URL.Query().Get("research"); researchIDOrCode != "" {
		// Resolve research first
		research, rErr := h.research.FindByID(r.Context(), researchIDOrCode)
		if rErr == nil && research == nil {
			research, _ = h.research.FindByCode(r.Context(), researchIDOrCode)
		}
		if research != nil {
			entry, err = h.entry.GetByIDOrCode(r.Context(), research.ID, idOrCode)
		} else {
			err = fmt.Errorf("not found")
		}
	} else {
		entry, err = h.entry.Get(r.Context(), idOrCode)
	}

	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, withProvenance(entryPayload(entry), h.entry.LatestRevision(r.Context(), entry), r.Context(), h.authorName))
}

func (h *EntryHandler) GetByResearch(w http.ResponseWriter, r *http.Request) {
	researchIDOrCode := r.PathValue("id")

	research, err := h.research.FindByID(r.Context(), researchIDOrCode)
	if err == nil && research == nil {
		research, _ = h.research.FindByCode(r.Context(), researchIDOrCode)
	}
	if research == nil {
		writeError(w, http.StatusNotFound, "not found")
		return
	}

	entry, err := h.entry.GetByIDOrCode(r.Context(), research.ID, r.PathValue("entryId"))
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, withProvenance(entryPayload(entry), h.entry.LatestRevision(r.Context(), entry), r.Context(), h.authorName))
}

func (h *EntryHandler) GetRelatedByResearch(w http.ResponseWriter, r *http.Request) {
	researchIDOrCode := r.PathValue("id")

	research, err := h.research.FindByID(r.Context(), researchIDOrCode)
	if err == nil && research == nil {
		research, _ = h.research.FindByCode(r.Context(), researchIDOrCode)
	}
	if research == nil {
		writeError(w, http.StatusNotFound, "not found")
		return
	}

	entry, err := h.entry.GetByIDOrCode(r.Context(), research.ID, r.PathValue("entryId"))
	if err != nil || entry == nil {
		writeError(w, http.StatusNotFound, "not found")
		return
	}

	related, err := h.entries.FindRelatedByTags(r.Context(), entry.ID, entry.Tags,
		auth.UserIDFromContext(r.Context()), shareResearchID(r.Context()))
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"data":  related,
		"count": len(related),
	})
}

// ResolveCode resolves a short code reference like "E3" (within research) or handles
// cross-research resolution via query param ?research_code=R2
func (h *EntryHandler) ResolveCode(w http.ResponseWriter, r *http.Request) {
	researchID := r.PathValue("id")
	entryCode := r.PathValue("code")

	// Through the service, never the repository: EntryService.GetByIDOrCode is
	// what checks ownership. Resolving a code straight from the repo returned
	// any user's entry, content included, to anyone who knew the research id.
	entry, err := h.entry.GetByIDOrCode(r.Context(), researchID, entryCode)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			writeError(w, http.StatusNotFound, "not found")
			return
		}
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, withProvenance(entryPayload(entry), h.entry.LatestRevision(r.Context(), entry), r.Context(), h.authorName))
}

// ResolveResearchCode resolves a research short code to its ID and metadata.
func (h *EntryHandler) ResolveResearchCode(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")

	// Through the service: a short code is global, so resolving it from the
	// repository turned R1, R2, R3… into an enumeration of every user's
	// researches, ids and names included.
	research, err := h.researchSvc.Get(r.Context(), code)
	if err != nil || research == nil {
		writeError(w, http.StatusNotFound, "not found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"data": map[string]any{
			"id":   research.ID,
			"code": research.Code,
			"name": research.Name,
		},
	})
}

// GetRelated returns entries that share tags with the given entry.
func (h *EntryHandler) GetRelated(w http.ResponseWriter, r *http.Request) {
	idOrCode := r.PathValue("id")

	var entry *domain.Entry
	var err error

	if researchIDOrCode := r.URL.Query().Get("research"); researchIDOrCode != "" {
		research, rErr := h.research.FindByID(r.Context(), researchIDOrCode)
		if rErr == nil && research == nil {
			research, _ = h.research.FindByCode(r.Context(), researchIDOrCode)
		}
		if research != nil {
			entry, err = h.entry.GetByIDOrCode(r.Context(), research.ID, idOrCode)
		} else {
			err = fmt.Errorf("not found")
		}
	} else {
		entry, err = h.entry.Get(r.Context(), idOrCode)
	}

	if err != nil || entry == nil {
		writeError(w, http.StatusNotFound, "not found")
		return
	}

	related, err := h.entries.FindRelatedByTags(r.Context(), entry.ID, entry.Tags,
		auth.UserIDFromContext(r.Context()), shareResearchID(r.Context()))
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"data":  related,
		"count": len(related),
	})
}

// entryPayload adds the document revision for a blocks entry, so a client can
// send it back with a patch and get a conflict rather than a silent overwrite.
func entryPayload(entry *domain.Entry) map[string]any {
	out := map[string]any{"data": entry}
	if entry != nil && entry.Type == domain.EntryBlocks {
		out["rev"] = service.DocumentRev(entry.Content)
	}
	return out
}

// withProvenance adds who last wrote the entry and when, so the page can say it
// without opening the history. `rev` here is the document hash a blocks entry
// carries; `revision` is the numbered snapshot — different things, and the two
// names sit side by side in this payload precisely because clients confuse them.
func withProvenance(payload map[string]any, rev *domain.EntryRevision, ctx context.Context, nameOf func(context.Context, string) string) map[string]any {
	if rev == nil {
		return payload
	}
	// A share visitor gets none of it.
	//
	// Provenance is who edited what, when, from which interview session — the
	// working process behind the result, of a piece with `instruction` and
	// `memory`, and the shared view deliberately has no history panel to show
	// it in. It also carried a concrete leak: `revision_session` names the
	// session's code and title, so a link created with sessions switched off
	// was still printing the interview's title on every entry page.
	if auth.ShareFromContext(ctx) != nil {
		return payload
	}
	payload["revision"] = rev.Revision
	payload["author_kind"] = rev.AuthorKind
	payload["revised_at"] = rev.CreatedAt
	// The name, when there is one. "Written by a person" answers the question
	// only halfway, and in a team the other half is the whole point of asking.
	if nameOf != nil && rev.UserID != "" {
		if name := nameOf(ctx, rev.UserID); name != "" {
			payload["author_name"] = name
		}
	}
	if rev.SessionCode != "" {
		payload["revision_session"] = map[string]any{
			"code": rev.SessionCode, "title": rev.SessionTitle, "id": rev.SessionID,
		}
	}
	return payload
}

// shareResearchID confines a query to the shared research, or returns "" for
// an ordinary reader whose membership does the confining. Related-by-tags is
// the one read in this file that is not routed through a service, so it is the
// one that has to say this out loud.
func shareResearchID(ctx context.Context) string {
	if sc := auth.ShareFromContext(ctx); sc != nil {
		return sc.ResearchID
	}
	return ""
}

// authorName resolves a revision's user id to something a reader recognises.
//
// It fails quietly: a deleted user, a database hiccup or auth being switched
// off all mean "no name", and the provenance line has always been able to stand
// without one. Nothing about a document should fail to render because the
// person who wrote it can no longer be looked up.
func (h *EntryHandler) authorName(ctx context.Context, userID string) string {
	if h.users == nil || userID == "" {
		return ""
	}
	user, err := h.users.FindByID(ctx, userID)
	if err != nil || user == nil {
		return ""
	}
	if user.Name != "" {
		return user.Name
	}
	return user.Email
}
