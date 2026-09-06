package handlers

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/dovod-app/app/internal/domain"
	"github.com/dovod-app/app/internal/service"
)

// TeamHandler serves team membership: who is in a team, what they may do, and
// the invitation links that bring people in.
type TeamHandler struct {
	teams    *service.TeamService
	research *service.ResearchService
	log      *slog.Logger
}

func NewTeamHandler(teams *service.TeamService, research *service.ResearchService, log *slog.Logger) *TeamHandler {
	return &TeamHandler{teams: teams, research: research, log: log}
}

func (h *TeamHandler) List(w http.ResponseWriter, r *http.Request) {
	teams, err := h.teams.List(r.Context())
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": teams, "count": len(teams)})
}

func (h *TeamHandler) Get(w http.ResponseWriter, r *http.Request) {
	team, err := h.teams.Get(r.Context(), r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": team})
}

func (h *TeamHandler) Create(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name string `json:"name"`
	}
	if !decodeJSON(w, r, &input) {
		return
	}
	team, err := h.teams.Create(r.Context(), input.Name)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": team})
}

func (h *TeamHandler) Update(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name string `json:"name"`
	}
	if !decodeJSON(w, r, &input) {
		return
	}
	team, err := h.teams.Rename(r.Context(), r.PathValue("id"), input.Name)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": team})
}

func (h *TeamHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.teams.Delete(r.Context(), r.PathValue("id")); err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": map[string]any{"deleted": true}})
}

// --- Members ---

func (h *TeamHandler) Members(w http.ResponseWriter, r *http.Request) {
	members, err := h.teams.Members(r.Context(), r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": members, "count": len(members)})
}

func (h *TeamHandler) UpdateMember(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Role string `json:"role"`
	}
	if !decodeJSON(w, r, &input) {
		return
	}
	err := h.teams.UpdateRole(r.Context(), r.PathValue("id"), r.PathValue("userId"), domain.TeamRole(input.Role))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": map[string]any{"role": input.Role}})
}

func (h *TeamHandler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	if err := h.teams.RemoveMember(r.Context(), r.PathValue("id"), r.PathValue("userId")); err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": map[string]any{"removed": true}})
}

// --- Invitations ---

func (h *TeamHandler) Invites(w http.ResponseWriter, r *http.Request) {
	invites, err := h.teams.Invites(r.Context(), r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": invites, "count": len(invites)})
}

// CreateInvite returns the link. The token appears in this response and
// nowhere else, ever: it is hashed at rest, so an owner who loses it issues a
// new invitation rather than looking the old one up.
func (h *TeamHandler) CreateInvite(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email string `json:"email"`
		Role  string `json:"role"`
	}
	if !decodeJSON(w, r, &input) {
		return
	}

	result, err := h.teams.CreateInvite(r.Context(), r.PathValue("id"), input.Email, domain.TeamRole(input.Role))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{
		"data": map[string]any{
			"invite": result.Invite,
			"token":  result.Token,
			"url":    inviteURL(r, result.Token),
		},
	})
}

func (h *TeamHandler) RevokeInvite(w http.ResponseWriter, r *http.Request) {
	if err := h.teams.RevokeInvite(r.Context(), r.PathValue("id")); err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": map[string]any{"revoked": true}})
}

// PreviewInvite is the one team route that answers without a session: the
// recipient has to be told what they are being invited to before they are
// asked to create an account. It reports why a link is dead rather than a bare
// 404, because "expired", "revoked" and "never existed" lead to three
// different next steps.
//
// When the request does carry a session it says more — whether they are
// already in the team, and whether they are signed in as the invited address.
func (h *TeamHandler) PreviewInvite(w http.ResponseWriter, r *http.Request) {
	preview, err := h.teams.PreviewInvite(r.Context(), r.PathValue("token"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": preview})
}

func (h *TeamHandler) AcceptInvite(w http.ResponseWriter, r *http.Request) {
	result, err := h.teams.AcceptInvite(r.Context(), r.PathValue("token"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": result})
}

// --- Research transfer ---

func (h *TeamHandler) TransferResearch(w http.ResponseWriter, r *http.Request) {
	var input struct {
		TeamID string `json:"team_id"`
	}
	if !decodeJSON(w, r, &input) {
		return
	}
	if input.TeamID == "" {
		writeError(w, http.StatusBadRequest, "team_id is required")
		return
	}

	researchID, err := h.research.ResolveID(r.Context(), r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	if err := h.teams.TransferResearch(r.Context(), researchID, input.TeamID); err != nil {
		writeServiceError(w, err)
		return
	}

	research, err := h.research.Get(r.Context(), researchID)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": research})
}

// AddResearches pulls researches into a team — the inverse of TransferResearch,
// which pushes one out of the team it is in. Same operation, opposite subject:
// this one is asked from the team, about a list.
func (h *TeamHandler) AddResearches(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ResearchIDs []string `json:"research_ids"`
	}
	if !decodeJSON(w, r, &input) {
		return
	}
	if len(input.ResearchIDs) == 0 {
		writeError(w, http.StatusBadRequest, "research_ids is required")
		return
	}

	// Short codes are what the UI holds, so they are resolved here rather than
	// making every caller look up a uuid first.
	ids := make([]string, 0, len(input.ResearchIDs))
	for _, ref := range input.ResearchIDs {
		id, err := h.research.ResolveID(r.Context(), ref)
		if err != nil {
			writeServiceError(w, err)
			return
		}
		ids = append(ids, id)
	}

	moved, err := h.teams.TransferResearches(r.Context(), r.PathValue("id"), ids)
	if err != nil {
		// A partial run is reported as a failure with its count, not as a
		// success: the reader asked for all of them.
		if moved > 0 {
			h.log.Warn("bulk transfer stopped part-way", "team_id", r.PathValue("id"), "moved", moved, "error", err)
		}
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": map[string]any{"moved": moved}})
}

// inviteURL builds the link an owner copies. It is derived from the request
// rather than from configuration so the link works on whatever host the owner
// is actually using — localhost during development, the real domain in
// production — instead of pointing somewhere they cannot reach.
func inviteURL(r *http.Request, token string) string {
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	if proto := r.Header.Get("X-Forwarded-Proto"); proto != "" {
		scheme = strings.Split(proto, ",")[0]
	}
	host := r.Host
	if fwd := r.Header.Get("X-Forwarded-Host"); fwd != "" {
		host = strings.Split(fwd, ",")[0]
	}
	return scheme + "://" + host + "/invite/" + token
}
