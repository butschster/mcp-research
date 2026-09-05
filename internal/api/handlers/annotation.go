package handlers

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/dovod-app/app/internal/domain"
	"github.com/dovod-app/app/internal/service"
	"github.com/dovod-app/app/internal/storage"
)

// The web is where a mark is born. Every route here is reachable from the
// browser, and Create in particular has no MCP counterpart on purpose — see
// AnnotationService.Create.

type AnnotationHandler struct {
	annotations *service.AnnotationService
	research    *service.ResearchService
	// users and teams resolve "who doubted this" to a name, under the rule in
	// resolveAuthorName. Both optional: with auth off there are no users, and a
	// mark then shows the glyph it has always shown.
	users *storage.UserRepository
	teams *storage.TeamRepository
	log   *slog.Logger
}

func NewAnnotationHandler(annotations *service.AnnotationService, research *service.ResearchService, users *storage.UserRepository, teams *storage.TeamRepository, log *slog.Logger) *AnnotationHandler {
	return &AnnotationHandler{annotations: annotations, research: research, users: users, teams: teams, log: log}
}

// ListByResearch is the queue page.
func (h *AnnotationHandler) ListByResearch(w http.ResponseWriter, r *http.Request) {
	researchID, err := h.research.ResolveID(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	filter := storage.AnnotationFilter{EntryID: r.URL.Query().Get("entry_id")}
	if s := r.URL.Query().Get("status"); s != "" {
		status := domain.AnnotationStatus(s)
		if !status.Valid() {
			writeError(w, http.StatusBadRequest, "status must be one of: open, answered, closed, dismissed")
			return
		}
		filter.Status = &status
	}
	if k := r.URL.Query().Get("kind"); k != "" {
		kind := domain.AnnotationKind(k)
		if !kind.Valid() {
			writeError(w, http.StatusBadRequest, "kind must be one of: verify, dig, disagree")
			return
		}
		filter.Kind = &kind
	}

	anns, err := h.annotations.ListByResearch(r.Context(), researchID, filter)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	// Anchor state is computed from the document, never stored, so this filter
	// runs after the read rather than in it. The orphan list is short by nature;
	// the cost of reading past it is the price of not persisting a value that
	// goes stale on the next agent write.
	if want := r.URL.Query().Get("anchor"); want != "" {
		anns = service.FilterByAnchor(anns, domain.AnchorState(want))
	}

	h.fillAuthorNames(r.Context(), researchID, anns)
	counts, _ := h.annotations.Counts(r.Context(), researchID)

	byAnchor := map[string]int{}
	byEntry := map[string]int{}
	for _, a := range anns {
		byEntry[a.EntryCode]++
		if a.Anchor != nil {
			byAnchor[string(a.Anchor.State)]++
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"data":  anns,
		"count": len(anns),
		"meta": map[string]any{
			"counts":    counts,
			"by_anchor": byAnchor,
			"by_entry":  byEntry,
		},
		// Kept alongside meta.counts because the research page already reads it
		// here; removing it is a frontend change, not a backend one.
		"counts": counts,
	})
}

// Bulk is the accept-all at the end of reviewing a pass.
func (h *AnnotationHandler) Bulk(w http.ResponseWriter, r *http.Request) {
	var input struct {
		IDs    []string `json:"ids"`
		Status string   `json:"status"`
		Reason string   `json:"reason"`
	}
	if !decodeJSON(w, r, &input) {
		return
	}

	researchID, err := h.research.ResolveID(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	results, err := h.annotations.Bulk(r.Context(), researchID, input.IDs, domain.AnnotationStatus(input.Status), input.Reason)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	applied := 0
	for _, res := range results {
		if res.OK {
			applied++
		}
	}
	// 200 with a per-row report rather than 204: a batch where two rows failed
	// is not a success, and it is not a failure either. The screen has to be
	// able to say "12 of 14" without asking again.
	writeJSON(w, http.StatusOK, map[string]any{
		"data":    results,
		"applied": applied,
		"total":   len(results),
	})
}

// fillAuthorNames resolves who left each mark, for readers entitled to know.
func (h *AnnotationHandler) fillAuthorNames(ctx context.Context, researchID string, anns []*domain.Annotation) {
	if h.users == nil || h.teams == nil || h.research == nil || len(anns) == 0 {
		return
	}
	// Read through the service, not the repository: it redacts the team fields
	// for a share visitor, so a link that reaches this code resolves no names at
	// all rather than resolving them against a team it cannot see.
	research, err := h.research.Get(ctx, researchID)
	if err != nil || research == nil {
		return
	}
	// One lookup per distinct author, not per mark: a queue is mostly one
	// person's marks, and this read happens on every page load.
	names := map[string]string{}
	for _, a := range anns {
		if a.UserID == "" {
			continue
		}
		name, seen := names[a.UserID]
		if !seen {
			name = resolveAuthorName(ctx, h.users, h.teams, research.TeamID, a.UserID)
			names[a.UserID] = name
		}
		a.AuthorName = name
	}
}

// ListByEntry is what the document page reads to draw its marks.
func (h *AnnotationHandler) ListByEntry(w http.ResponseWriter, r *http.Request) {
	anns, err := h.annotations.ListByEntry(r.Context(), r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": anns, "count": len(anns)})
}

func (h *AnnotationHandler) Create(w http.ResponseWriter, r *http.Request) {
	var input struct {
		BlockID string `json:"block_id"`
		Quote   struct {
			Exact  string `json:"exact"`
			Prefix string `json:"prefix"`
			Suffix string `json:"suffix"`
		} `json:"quote"`
		Kind string `json:"kind"`
		Body string `json:"body"`
	}
	if !decodeJSON(w, r, &input) {
		return
	}

	a, err := h.annotations.Create(r.Context(), service.CreateAnnotationRequest{
		EntryID: r.PathValue("id"),
		BlockID: input.BlockID,
		Quote: domain.Quote{
			Exact:  input.Quote.Exact,
			Prefix: input.Quote.Prefix,
			Suffix: input.Quote.Suffix,
		},
		Kind: domain.AnnotationKind(input.Kind),
		Body: input.Body,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": a})
}

// Update carries the human decisions: closing, dismissing, sending an answer
// back, editing the note, re-pointing a drifted mark by hand.
func (h *AnnotationHandler) Update(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Kind    *string `json:"kind"`
		Body    *string `json:"body"`
		Status  *string `json:"status"`
		Reason  *string `json:"reason"`
		TaskID  *string `json:"task_id"`
		BlockID *string `json:"block_id"`
		Quote   *struct {
			Exact  string `json:"exact"`
			Prefix string `json:"prefix"`
			Suffix string `json:"suffix"`
		} `json:"quote"`
	}
	if !decodeJSON(w, r, &input) {
		return
	}

	req := service.UpdateAnnotationRequest{
		Body:    input.Body,
		Reason:  input.Reason,
		TaskID:  input.TaskID,
		BlockID: input.BlockID,
	}
	if input.Kind != nil {
		k := domain.AnnotationKind(*input.Kind)
		req.Kind = &k
	}
	if input.Status != nil {
		s := domain.AnnotationStatus(*input.Status)
		req.Status = &s
	}
	if input.Quote != nil {
		req.Quote = &domain.Quote{
			Exact:  input.Quote.Exact,
			Prefix: input.Quote.Prefix,
			Suffix: input.Quote.Suffix,
		}
	}

	a, err := h.annotations.Update(r.Context(), r.PathValue("id"), req)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": a})
}

func (h *AnnotationHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.annotations.Delete(r.Context(), r.PathValue("id")); err != nil {
		writeServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
