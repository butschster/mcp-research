package handlers

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/dovod-app/app/internal/auth"
	"github.com/dovod-app/app/internal/domain"
	"github.com/dovod-app/app/internal/service"
	"github.com/dovod-app/app/internal/storage"
)

type EntryHandler struct {
	entry       *service.EntryService
	views       *service.EntryViewService
	researchSvc *service.ResearchService
	entries     *storage.EntryRepository
	research    *storage.ResearchRepository
	// users resolves the name behind a revision's user id, for the one line the
	// page shows about who wrote a document. Optional: with `auth_enabled:
	// false` there are no users at all, and the provenance line falls back to
	// the author kind it has always shown.
	users *storage.UserRepository
	// teams answers whether the caller's research and the revision's author
	// still belong together. Without it the name is handed to anybody who can
	// read the document, which is a wider audience than the team.
	teams *storage.TeamRepository
	log   *slog.Logger
}

func NewEntryHandler(entry *service.EntryService, researchSvc *service.ResearchService, entries *storage.EntryRepository, research *storage.ResearchRepository, users *storage.UserRepository, teams *storage.TeamRepository, log *slog.Logger) *EntryHandler {
	return &EntryHandler{entry: entry, researchSvc: researchSvc, entries: entries, research: research, users: users, teams: teams, log: log}
}

// SetEntryViewService adds the caller's personal revision checkpoint to a
// document response. Kept optional so handler fixtures that do not exercise
// the update queue stay small.
func (h *EntryHandler) SetEntryViewService(views *service.EntryViewService) {
	h.views = views
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

	h.writeEntry(w, r, entry)
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

	h.writeEntry(w, r, entry)
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

	h.writeEntry(w, r, entry)
}

// writeEntry holds the snapshot boundary for the document and its read state.
// LatestSnapshot reads the document and immutable revision that supplies its
// number from one database snapshot; StateAt then receives that number too. A concurrent
// commit therefore returns either complete old or complete new content, never
// old content labelled as the new revision.
func (h *EntryHandler) writeEntry(w http.ResponseWriter, r *http.Request, entry *domain.Entry) {
	snapshot, rev, err := h.entry.LatestSnapshot(r.Context(), entry)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	payload := withProvenance(entryPayload(snapshot), rev, r.Context(), snapshot.ResearchID, h.teamOf, h.authorName)
	if h.views != nil && rev != nil {
		state, err := h.views.StateAt(r.Context(), snapshot, rev.Revision)
		if err != nil {
			writeServiceError(w, err)
			return
		}
		if state != nil {
			payload["view_state"] = state
		}
	}
	writeJSON(w, http.StatusOK, payload)
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
func withProvenance(payload map[string]any, rev *domain.EntryRevision, ctx context.Context, researchID string, teamOf func(context.Context, string) string, nameOf func(context.Context, string, string) string) map[string]any {
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
	// The name, when the caller is entitled to it. "Written by a person" answers
	// the question only halfway, and in a team the other half is the point of
	// asking — but only inside that team.
	// Both lookups are behind the guards above, so a share visitor and a
	// document with no revision cost nothing — the team query used to run and
	// be discarded.
	if nameOf != nil && teamOf != nil && rev.UserID != "" {
		if name := nameOf(ctx, teamOf(ctx, researchID), rev.UserID); name != "" {
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

// teamOf is the team that owns the research a document lives in — the scope
// within which knowing who wrote it is ordinary rather than a disclosure.
func (h *EntryHandler) teamOf(ctx context.Context, researchID string) string {
	if h.research == nil {
		return ""
	}
	research, err := h.research.FindByID(ctx, researchID)
	if err != nil || research == nil {
		return ""
	}
	return research.TeamID
}

// authorName resolves a revision's author for display. The rule it obeys — and
// why the name is not simply handed to anyone who can read the document — lives
// with resolveAuthorName, which the annotation queue shares.
func (h *EntryHandler) authorName(ctx context.Context, teamID, userID string) string {
	return resolveAuthorName(ctx, h.users, h.teams, teamID, userID)
}
