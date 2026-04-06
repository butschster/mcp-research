package handlers

import (
	"log/slog"
	"net/http"

	"github.com/butschster/mcp-research/internal/service"
	"github.com/butschster/mcp-research/internal/storage"
)

type ExternalLinkHandler struct {
	links    *storage.ExternalLinkRepository
	research *service.ResearchService
	log      *slog.Logger
}

func NewExternalLinkHandler(links *storage.ExternalLinkRepository, research *service.ResearchService, log *slog.Logger) *ExternalLinkHandler {
	return &ExternalLinkHandler{links: links, research: research, log: log}
}

func (h *ExternalLinkHandler) ListByResearch(w http.ResponseWriter, r *http.Request) {
	researchID, err := h.research.ResolveID(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	links, err := h.links.FindByResearch(r.Context(), researchID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Group by domain
	domains := make(map[string][]any)
	var domainOrder []string
	for _, l := range links {
		if _, exists := domains[l.Domain]; !exists {
			domainOrder = append(domainOrder, l.Domain)
		}
		domains[l.Domain] = append(domains[l.Domain], map[string]any{
			"url":         l.URL,
			"title":       l.Title,
			"domain":      l.Domain,
			"source_type": l.SourceType,
			"source_id":   l.SourceID,
			"entry_code":  l.EntryCode,
			"entry_title": l.EntryTitle,
		})
	}

	var grouped []map[string]any
	for _, d := range domainOrder {
		grouped = append(grouped, map[string]any{
			"domain": d,
			"links":  domains[d],
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"data":  grouped,
		"total": len(links),
	})
}

func (h *ExternalLinkHandler) ListByEntry(w http.ResponseWriter, r *http.Request) {
	entryID := r.PathValue("id")

	links, err := h.links.FindBySource(r.Context(), "entry", entryID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"data":  links,
		"count": len(links),
	})
}
