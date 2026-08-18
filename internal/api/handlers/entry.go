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
	// teams answers whether the caller's research and the revision's author
	// still belong together. Without it the name is handed to anybody who can
	// read the document, which is a wider audience than the team.
	teams *storage.TeamRepository
	log   *slog.Logger
}

func NewEntryHandler(entry *service.EntryService, researchSvc *service.ResearchService, entries *storage.EntryRepository, research *storage.ResearchRepository, users *storage.UserRepository, teams *storage.TeamRepository, log *slog.Logger) *EntryHandler {
	return &EntryHandler{entry: entry, researchSvc: researchSvc, entries: entries, research: research, users: users, teams: teams, log: log}
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

	writeJSON(w, http.StatusOK, withProvenance(entryPayload(entry), h.entry.LatestRevision(r.Context(), entry), r.Context(), entry.ResearchID, h.teamOf, h.authorName))
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

	writeJSON(w, http.StatusOK, withProvenance(entryPayload(entry), h.entry.LatestRevision(r.Context(), entry), r.Context(), entry.ResearchID, h.teamOf, h.authorName))
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

	writeJSON(w, http.StatusOK, withProvenance(entryPayload(entry), h.entry.LatestRevision(r.Context(), entry), r.Context(), entry.ResearchID, h.teamOf, h.authorName))
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

// authorName resolves a revision's user id to something a reader recognises —
// but only for somebody entitled to recognise it.
//
// The right to read a document does not carry the right to know who wrote it.
// Those two come apart through ordinary use: a research moves to another team
// and every member of the new one inherits entries written by people they have
// never shared a team with; a member leaves and the people who join afterwards
// read their name off every document they touched. Membership in the team that
// owns the research now is the boundary, and it is checked here rather than
// assumed from the entry read that got us this far.
//
// The email is deliberately not a fallback. It would fire for exactly the
// accounts that never set a name — API keys, OAuth clients, --default-user —
// turning a display nicety into a bulk address disclosure. A person with no
// name is "a person", which is what the glyph already says.
//
// Everything else fails quietly: a deleted user, a database hiccup or auth
// being off all mean "no name", and the line has always stood without one.
func (h *EntryHandler) authorName(ctx context.Context, teamID, userID string) string {
	if h.users == nil || h.teams == nil || userID == "" || teamID == "" {
		return ""
	}
	if _, ok, err := h.teams.FindRole(ctx, teamID, userID); err != nil || !ok {
		return ""
	}
	user, err := h.users.FindByID(ctx, userID)
	if err != nil || user == nil {
		return ""
	}
	return user.Name
}
