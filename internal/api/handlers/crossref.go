package handlers

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/butschster/mcp-research/internal/domain"
	"github.com/butschster/mcp-research/internal/service"
	"github.com/butschster/mcp-research/internal/storage"
)

type CrossRefHandler struct {
	crossrefs   *storage.CrossRefRepository
	entrySvc    *service.EntryService
	researchSvc *service.ResearchService
	roadmapSvc  *service.RoadmapService
	log         *slog.Logger
}

func NewCrossRefHandler(crossrefs *storage.CrossRefRepository, entrySvc *service.EntryService, researchSvc *service.ResearchService, log *slog.Logger) *CrossRefHandler {
	return &CrossRefHandler{crossrefs: crossrefs, entrySvc: entrySvc, researchSvc: researchSvc, log: log}
}

func (h *CrossRefHandler) SetRoadmapService(svc *service.RoadmapService) {
	h.roadmapSvc = svc
}

// ListForResearch returns all stored cross-references for a research.
func (h *CrossRefHandler) ListForResearch(w http.ResponseWriter, r *http.Request) {
	researchID, err := h.researchSvc.ResolveID(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	refs, err := h.crossrefs.FindByResearch(r.Context(), researchID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"data":  refs,
		"count": len(refs),
	})
}

// GetForEntry returns outgoing and incoming cross-references for a specific entry,
// enriched with entry titles, codes, and research names.
func (h *CrossRefHandler) GetForEntry(w http.ResponseWriter, r *http.Request) {
	entryID := r.PathValue("id")

	// Resolve entry by ID or code
	researchIDOrCode := r.URL.Query().Get("research")
	entry, err := h.resolveEntry(r.Context(), entryID, researchIDOrCode)
	if err != nil {
		writeError(w, http.StatusNotFound, "entry not found")
		return
	}

	// Outgoing: this entry references others
	outgoing, err := h.crossrefs.FindBySourceEntry(r.Context(), entry.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Incoming: others reference this entry
	incoming, err := h.crossrefs.FindByTargetEntry(r.Context(), entry.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Enrich with metadata
	type enrichedRef struct {
		SourceType       string `json:"source_type"`
		SourceID         string `json:"source_id"`
		SourceResearchID string `json:"source_research_id"`
		TargetEntryID    string `json:"target_entry_id,omitempty"`
		TargetResearchID string `json:"target_research_id,omitempty"`
		TargetRoadmapID  string `json:"target_roadmap_id,omitempty"`
		TargetNodeID     string `json:"target_node_id,omitempty"`
		TargetRef        string `json:"target_ref"`
		Resolved         bool   `json:"resolved"`
		// Enriched fields
		EntryTitle   string `json:"entry_title,omitempty"`
		EntryCode    string `json:"entry_code,omitempty"`
		ResearchName string `json:"research_name,omitempty"`
		ResearchCode string `json:"research_code,omitempty"`
		RoadmapTitle string `json:"roadmap_title,omitempty"`
		RoadmapCode  string `json:"roadmap_code,omitempty"`
	}

	enrichOutgoing := make([]enrichedRef, 0, len(outgoing))
	for _, ref := range outgoing {
		er := enrichedRef{
			SourceType:       ref.SourceType,
			SourceID:         ref.SourceID,
			SourceResearchID: ref.SourceResearchID,
			TargetEntryID:    ref.TargetEntryID,
			TargetResearchID: ref.TargetResearchID,
			TargetRoadmapID:  ref.TargetRoadmapID,
			TargetNodeID:     ref.TargetNodeID,
			TargetRef:        ref.TargetRef,
			Resolved:         ref.Resolved,
		}
		if ref.TargetEntryID != "" {
			if te, _ := h.entrySvc.Get(r.Context(), ref.TargetEntryID); te != nil {
				er.EntryTitle = te.Title
				er.EntryCode = te.Code
			}
		}
		if ref.TargetRoadmapID != "" && h.roadmapSvc != nil {
			if rm, _ := h.roadmapSvc.Get(r.Context(), ref.TargetRoadmapID); rm != nil {
				er.RoadmapTitle = rm.Title
				er.RoadmapCode = rm.Code
			}
		}
		targetResID := ref.TargetResearchID
		if targetResID == "" {
			targetResID = ref.SourceResearchID
		}
		if targetResID != "" {
			if tr, _ := h.researchSvc.Get(r.Context(), targetResID); tr != nil {
				er.ResearchName = tr.Name
				er.ResearchCode = tr.Code
			}
		}
		enrichOutgoing = append(enrichOutgoing, er)
	}

	enrichIncoming := make([]enrichedRef, 0, len(incoming))
	for _, ref := range incoming {
		er := enrichedRef{
			SourceType:       ref.SourceType,
			SourceID:         ref.SourceID,
			SourceResearchID: ref.SourceResearchID,
			TargetEntryID:    ref.TargetEntryID,
			TargetResearchID: ref.TargetResearchID,
			TargetRef:        ref.TargetRef,
			Resolved:         ref.Resolved,
		}
		// For incoming, the "linked entry" is the source
		if ref.SourceType == "entry" && ref.SourceID != "" {
			if se, _ := h.entrySvc.Get(r.Context(), ref.SourceID); se != nil {
				er.EntryTitle = se.Title
				er.EntryCode = se.Code
			}
		}
		if ref.SourceResearchID != "" {
			if sr, _ := h.researchSvc.Get(r.Context(), ref.SourceResearchID); sr != nil {
				er.ResearchName = sr.Name
				er.ResearchCode = sr.Code
			}
		}
		enrichIncoming = append(enrichIncoming, er)
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"outgoing": enrichOutgoing,
		"incoming": enrichIncoming,
	})
}

func (h *CrossRefHandler) resolveEntry(ctx context.Context, idOrCode, researchIDOrCode string) (*domain.Entry, error) {
	if researchIDOrCode != "" {
		researchID, err := h.researchSvc.ResolveID(ctx, researchIDOrCode)
		if err != nil {
			return nil, err
		}
		return h.entrySvc.GetByIDOrCode(ctx, researchID, idOrCode)
	}
	return h.entrySvc.Get(ctx, idOrCode)
}

// Rebuild rescans all entries in a research and rebuilds cross-references.
func (h *CrossRefHandler) Rebuild(w http.ResponseWriter, r *http.Request) {
	researchID, err := h.researchSvc.ResolveID(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	count, err := h.entrySvc.RebuildCrossRefs(r.Context(), researchID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"rebuilt": count,
		"status":  "ok",
	})
}
