package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/dovod-app/app/internal/auth"
	"github.com/dovod-app/app/internal/domain"
	"github.com/dovod-app/app/internal/service"
)

type ShareHandler struct {
	shares   *service.ShareService
	research *service.ResearchService
	section  *service.SectionService
	log      *slog.Logger
}

func NewShareHandler(shares *service.ShareService, research *service.ResearchService, section *service.SectionService, log *slog.Logger) *ShareHandler {
	return &ShareHandler{shares: shares, research: research, section: section, log: log}
}

// --- Owner side ---

type createShareRequest struct {
	Label   string `json:"label"`
	Include struct {
		Sessions *bool `json:"sessions"`
		Tasks    *bool `json:"tasks"`
		Roadmaps *bool `json:"roadmaps"`
		Export   *bool `json:"export"`
	} `json:"include"`
	// ExpiresInDays is a pointer so "never" and "not sent" stay distinguishable
	// from "in zero days".
	ExpiresInDays *int   `json:"expires_in_days"`
	Password      string `json:"password"`
}

// Create issues a link and returns it exactly once.
func (h *ShareHandler) Create(w http.ResponseWriter, r *http.Request) {
	researchID, err := h.research.ResolveID(r.Context(), r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}

	var req createShareRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && !errors.Is(err, io.EOF) {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	// Roadmaps are part of the research proper, so they default to in; the
	// other three are working state and default to out. A client that sends no
	// include object at all gets the same answer as one that sends an empty
	// one, which is what makes the flags safe to add to later.
	include := domain.ShareInclude{Roadmaps: true}
	if req.Include.Sessions != nil {
		include.Sessions = *req.Include.Sessions
	}
	if req.Include.Tasks != nil {
		include.Tasks = *req.Include.Tasks
	}
	if req.Include.Roadmaps != nil {
		include.Roadmaps = *req.Include.Roadmaps
	}
	if req.Include.Export != nil {
		include.Export = *req.Include.Export
	}

	result, err := h.shares.Create(r.Context(), researchID, service.CreateShareRequest{
		Label:         req.Label,
		Include:       include,
		ExpiresInDays: req.ExpiresInDays,
		Password:      req.Password,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"data": map[string]any{
			"share": result.Share,
			// The only response in the product that carries this. There is no
			// route that can produce it again.
			"token": result.Token,
			"url":   shareURL(r, result.Token),
		},
	})
}

func (h *ShareHandler) List(w http.ResponseWriter, r *http.Request) {
	researchID, err := h.research.ResolveID(r.Context(), r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	shares, err := h.shares.List(r.Context(), researchID)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": shares, "count": len(shares)})
}

func (h *ShareHandler) Revoke(w http.ResponseWriter, r *http.Request) {
	if err := h.shares.Revoke(r.Context(), r.PathValue("id")); err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"status": "revoked"})
}

// --- Visitor side ---

// Payload is the first request a shared page makes: what this link is, and the
// research it opens on. Everything after it goes through the ordinary read
// routes mounted under the same prefix.
func (h *ShareHandler) Payload(w http.ResponseWriter, r *http.Request) {
	sc := auth.ShareFromContext(r.Context())
	if sc == nil {
		// Unreachable through the mounted routes; a refusal rather than a nil
		// dereference if it ever becomes reachable.
		writeError(w, http.StatusNotFound, service.ErrShareUnavailable.Error())
		return
	}

	research, err := h.research.Get(r.Context(), sc.ResearchID)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	sections, err := h.section.List(r.Context(), research.ID)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	type sectionWithCount struct {
		*domain.Section
		EntriesCount int `json:"entries_count"`
	}
	sectionData := make([]sectionWithCount, 0, len(sections))
	for _, s := range sections {
		count, _ := h.section.CountEntries(r.Context(), s.ID)
		sectionData = append(sectionData, sectionWithCount{Section: s, EntriesCount: count})
	}

	// A visit, counted once.
	//
	// This route is also what the shared page refetches on every realtime event,
	// and counting those made the number mean something else entirely: a visitor
	// who left a tab open while the owner's agent wrote forty entries registered
	// as forty-one views, and `last_seen_at` tracked the owner's writing rather
	// than the client's reading. Only the page load asks for it.
	//
	// A visitor who strips the parameter under-counts their own visits, which is
	// the harmless direction to be wrong in — the owner learns less, and the
	// issue is explicit that finer tracking of a recipient is surveillance
	// rather than analytics.
	if r.URL.Query().Get("visit") == "1" {
		h.shares.Seen(r.Context(), sc.ID)
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"data": map[string]any{
			"share":    shareView(r.Context(), h.shares, sc),
			"research": research,
			"sections": sectionData,
		},
	})
}

type unlockRequest struct {
	Password string `json:"password"`
}

// Unlock exchanges a password for the value the visitor sends with every later
// request. It runs outside the share scope, because a locked link has not been
// resolved yet.
func (h *ShareHandler) Unlock(w http.ResponseWriter, r *http.Request) {
	var req unlockRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && !errors.Is(err, io.EOF) {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	unlock, err := h.shares.Unlock(r.Context(), r.PathValue("token"), req.Password)
	if err != nil {
		// Here — and only here — "locked" means the password was wrong rather
		// than missing. The visitor is standing at the prompt already; telling
		// them to open the prompt is not an answer.
		if errors.Is(err, service.ErrShareLocked) {
			writeJSON(w, http.StatusUnauthorized, map[string]any{
				"error":  "that password is not right",
				"reason": "invalid_password",
			})
			return
		}
		WriteShareError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": map[string]any{"unlock": unlock}})
}

// shareView is the metadata a visitor is allowed to know about their own link:
// what it is called, what it contains, who shared it, and when it lapses. Not
// the id of the row, not the creator's account, not the team.
func shareView(ctx context.Context, svc *service.ShareService, sc *auth.Share) map[string]any {
	meta := svc.Describe(ctx, sc.ID)
	view := map[string]any{
		"research_id": sc.ResearchID,
		"include":     sc.Include,
	}
	if meta == nil {
		return view
	}
	view["label"] = meta.Label
	view["owner_name"] = meta.CreatorName
	view["research_code"] = meta.ResearchCode
	view["expires_at"] = meta.ExpiresAt
	return view
}

// WriteShareError maps the share errors to statuses.
//
// Everything except "locked" collapses into one 404 on purpose: revoked,
// expired and never existed must be indistinguishable from outside, or the URL
// bar becomes a way to ask whether a link was ever real.
func WriteShareError(w http.ResponseWriter, err error) {
	switch {
	case err == nil:
		return
	case errors.Is(err, service.ErrShareLocked):
		writeJSON(w, http.StatusUnauthorized, map[string]any{
			"error":  err.Error(),
			"reason": "password_required",
		})
	case errors.Is(err, service.ErrShareUnavailable), errors.Is(err, service.ErrNotFound):
		writeError(w, http.StatusNotFound, service.ErrShareUnavailable.Error())
	default:
		writeServiceError(w, err)
	}
}

// shareURL rebuilds the link a visitor opens, so the dialog can show something
// copyable rather than making the browser guess.
func shareURL(r *http.Request, token string) string {
	scheme := "http"
	if r.TLS != nil || strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https") {
		scheme = "https"
	}
	return scheme + "://" + r.Host + "/s/" + token
}
