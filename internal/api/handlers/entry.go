package handlers

import (
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
	log         *slog.Logger
}

func NewEntryHandler(entry *service.EntryService, researchSvc *service.ResearchService, entries *storage.EntryRepository, research *storage.ResearchRepository, log *slog.Logger) *EntryHandler {
	return &EntryHandler{entry: entry, researchSvc: researchSvc, entries: entries, research: research, log: log}
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
			err = fmt.Errorf("research not found")
		}
	} else {
		entry, err = h.entry.Get(r.Context(), idOrCode)
	}

	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, withProvenance(entryPayload(entry), h.entry.LatestRevision(r.Context(), entry)))
}

func (h *EntryHandler) GetByResearch(w http.ResponseWriter, r *http.Request) {
	researchIDOrCode := r.PathValue("id")

	research, err := h.research.FindByID(r.Context(), researchIDOrCode)
	if err == nil && research == nil {
		research, _ = h.research.FindByCode(r.Context(), researchIDOrCode)
	}
	if research == nil {
		writeError(w, http.StatusNotFound, "research not found")
		return
	}

	entry, err := h.entry.GetByIDOrCode(r.Context(), research.ID, r.PathValue("entryId"))
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, withProvenance(entryPayload(entry), h.entry.LatestRevision(r.Context(), entry)))
}

func (h *EntryHandler) GetRelatedByResearch(w http.ResponseWriter, r *http.Request) {
	researchIDOrCode := r.PathValue("id")

	research, err := h.research.FindByID(r.Context(), researchIDOrCode)
	if err == nil && research == nil {
		research, _ = h.research.FindByCode(r.Context(), researchIDOrCode)
	}
	if research == nil {
		writeError(w, http.StatusNotFound, "research not found")
		return
	}

	entry, err := h.entry.GetByIDOrCode(r.Context(), research.ID, r.PathValue("entryId"))
	if err != nil || entry == nil {
		writeError(w, http.StatusNotFound, "entry not found")
		return
	}

	related, err := h.entries.FindRelatedByTags(r.Context(), entry.ID, entry.Tags, auth.UserIDFromContext(r.Context()))
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
			writeError(w, http.StatusNotFound, "entry not found")
			return
		}
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, withProvenance(entryPayload(entry), h.entry.LatestRevision(r.Context(), entry)))
}

// ResolveResearchCode resolves a research short code to its ID and metadata.
func (h *EntryHandler) ResolveResearchCode(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")

	// Through the service: a short code is global, so resolving it from the
	// repository turned R1, R2, R3… into an enumeration of every user's
	// researches, ids and names included.
	research, err := h.researchSvc.Get(r.Context(), code)
	if err != nil || research == nil {
		writeError(w, http.StatusNotFound, "research not found")
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
			err = fmt.Errorf("research not found")
		}
	} else {
		entry, err = h.entry.Get(r.Context(), idOrCode)
	}

	if err != nil || entry == nil {
		writeError(w, http.StatusNotFound, "entry not found")
		return
	}

	related, err := h.entries.FindRelatedByTags(r.Context(), entry.ID, entry.Tags, auth.UserIDFromContext(r.Context()))
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
func withProvenance(payload map[string]any, rev *domain.EntryRevision) map[string]any {
	if rev == nil {
		return payload
	}
	payload["revision"] = rev.Revision
	payload["author_kind"] = rev.AuthorKind
	payload["revised_at"] = rev.CreatedAt
	if rev.SessionCode != "" {
		payload["revision_session"] = map[string]any{
			"code": rev.SessionCode, "title": rev.SessionTitle, "id": rev.SessionID,
		}
	}
	return payload
}
