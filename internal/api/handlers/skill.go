package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/dovod-app/app/internal/domain"
	"github.com/dovod-app/app/internal/service"
)

type SkillHandler struct {
	skills   *service.SkillService
	research *service.ResearchService
	log      *slog.Logger
}

func NewSkillHandler(skills *service.SkillService, research *service.ResearchService, log *slog.Logger) *SkillHandler {
	return &SkillHandler{skills: skills, research: research, log: log}
}

type skillRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Body        string `json:"body"`
	// Slug attaches an existing skill instead of writing a new one. The two are
	// one endpoint because from the client's side they are one act — "this
	// research should follow that" — and splitting them would mean the UI
	// choosing a route before the user chooses a skill.
	Slug string `json:"slug"`
}

// ListAttached returns what the research follows, in precedence order.
func (h *SkillHandler) ListAttached(w http.ResponseWriter, r *http.Request) {
	id, err := h.research.ResolveID(r.Context(), r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	list, err := h.skills.ListAttached(r.Context(), id)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"data": emptyIfNil(list),
		// The cap travels with the list rather than being hard-coded in the
		// frontend: a limit the client believes and the server enforces will
		// disagree exactly once, at the worst possible moment.
		"cap": service.SkillCap(),
		// `chosen`, not `attached`: the list also carries the ambient product
		// skills, which are on by definition and outside the budget. A count
		// called `attached` sitting over seven rows and reading 6 is the sort of
		// thing a reader assumes is a bug in the list.
		"chosen": service.CountChosen(list),
	})
}

// Library is what this research may attach — the built-ins and its team's
// library, each carrying whether it is already on.
func (h *SkillHandler) Library(w http.ResponseWriter, r *http.Request) {
	id, err := h.research.ResolveID(r.Context(), r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	list, err := h.skills.ListLibrary(r.Context(), id, r.URL.Query().Get("q"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": emptyIfNil(list)})
}

// Read returns one skill with its body. The body lives only on this route and
// on skill_load, which is what keeps the index the agent sees small.
func (h *SkillHandler) Read(w http.ResponseWriter, r *http.Request) {
	id, err := h.research.ResolveID(r.Context(), r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	sk, err := h.skills.Load(r.Context(), id, r.PathValue("slug"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": sk})
}

// Create either attaches an existing skill by slug or writes a research-private
// one, depending on which fields arrived.
func (h *SkillHandler) Create(w http.ResponseWriter, r *http.Request) {
	id, err := h.research.ResolveID(r.Context(), r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	var req skillRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	if req.Slug != "" {
		sk, err := h.skills.Attach(r.Context(), id, req.Slug, false)
		if err != nil {
			writeSkillError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, map[string]any{"data": sk})
		return
	}

	sk, err := h.skills.CreatePrivate(r.Context(), id, service.SkillInput{
		Name: req.Name, Description: req.Description, Body: req.Body,
	})
	if err != nil {
		writeSkillError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": sk})
}

func (h *SkillHandler) Detach(w http.ResponseWriter, r *http.Request) {
	id, err := h.research.ResolveID(r.Context(), r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	if err := h.skills.Detach(r.Context(), id, r.PathValue("slug")); err != nil {
		writeSkillError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Fork edits a built-in by copying it into the team and swapping the
// attachment. The response says `forked` so the client can tell this from an
// ordinary update and follow the slug to the new row.
func (h *SkillHandler) Fork(w http.ResponseWriter, r *http.Request) {
	id, err := h.research.ResolveID(r.Context(), r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	var req skillRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	sk, err := h.skills.Fork(r.Context(), id, r.PathValue("slug"), service.SkillInput{
		Name: req.Name, Description: req.Description, Body: req.Body,
	})
	if err != nil {
		writeSkillError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": sk, "forked": true})
}

func (h *SkillHandler) CopyHere(w http.ResponseWriter, r *http.Request) {
	id, err := h.research.ResolveID(r.Context(), r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	sk, err := h.skills.CopyHere(r.Context(), id, r.PathValue("slug"))
	if err != nil {
		writeSkillError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": sk})
}

func (h *SkillHandler) Promote(w http.ResponseWriter, r *http.Request) {
	id, err := h.research.ResolveID(r.Context(), r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	sk, err := h.skills.Promote(r.Context(), id, r.PathValue("slug"))
	if err != nil {
		writeSkillError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": sk})
}

// ReadByID returns one skill with its body, addressed the way Update and Delete
// are. Without it a maintainer could overwrite a team skill without being able
// to read it back.
func (h *SkillHandler) ReadByID(w http.ResponseWriter, r *http.Request) {
	sk, err := h.skills.Read(r.Context(), r.PathValue("skillId"))
	if err != nil {
		writeSkillError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": sk})
}

// ListTeam is a team's library without going through one of its researches — a
// team writing its first skill does not have one yet.
func (h *SkillHandler) ListTeam(w http.ResponseWriter, r *http.Request) {
	list, err := h.skills.ListTeam(r.Context(), r.PathValue("id"))
	if err != nil {
		writeSkillError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": emptyIfNil(list)})
}

// --- Team library and by-id management ---
//
// A slug is unique only within its tier scope, so it is not an address for
// management: the built-in `evidence-grading` and a team's fork of it share
// one. Everything that edits or deletes a skill addresses it by id, and slugs
// are used only inside a research, where the resolution order is defined.

func (h *SkillHandler) CreateTeam(w http.ResponseWriter, r *http.Request) {
	var req skillRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	sk, err := h.skills.CreateTeam(r.Context(), r.PathValue("id"), service.SkillInput{
		Name: req.Name, Description: req.Description, Body: req.Body,
	})
	if err != nil {
		writeSkillError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": sk})
}

func (h *SkillHandler) Update(w http.ResponseWriter, r *http.Request) {
	var req skillRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	sk, err := h.skills.Update(r.Context(), r.PathValue("skillId"), service.SkillInput{
		Name: req.Name, Description: req.Description, Body: req.Body,
	})
	if err != nil {
		writeSkillError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": sk})
}

func (h *SkillHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.skills.Delete(r.Context(), r.PathValue("skillId")); err != nil {
		writeSkillError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// writeSkillError adds the machine-readable code the two 409s need. The UI
// writes a different sentence for "six are already attached" than for "that one
// is already on", and picking between them by matching on prose is how a
// message edit silently changes behaviour.
func writeSkillError(w http.ResponseWriter, err error) {
	// The code itself comes from the service, because the MCP tools answer with
	// the same vocabulary and two switches over the same errors drift.
	switch code := service.SkillErrorCode(err); code {
	case "skill_cap_reached", "already_attached", "slug_taken", "skill_in_use":
		writeCodedError(w, http.StatusConflict, err.Error(), code)
		return
	case "not_allowed":
		writeCodedError(w, http.StatusForbidden, err.Error(), code)
		return
	}
	switch {
	case errors.Is(err, service.ErrSkillNameEmpty):
		writeFieldError(w, err.Error(), "name")
	case errors.Is(err, service.ErrSkillDescriptionLong),
		errors.Is(err, service.ErrSkillDescriptionEmpty):
		writeFieldError(w, err.Error(), "description")
	case errors.Is(err, service.ErrSkillBodyEmpty), errors.Is(err, service.ErrSkillBodyLong):
		writeFieldError(w, err.Error(), "body")
	default:
		writeServiceError(w, err)
	}
}

func writeCodedError(w http.ResponseWriter, status int, msg, code string) {
	writeJSON(w, status, map[string]any{"error": msg, "code": code})
}

// writeFieldError names the input that was wrong, so the form can put the
// message under it rather than at the bottom of the page.
func writeFieldError(w http.ResponseWriter, msg, field string) {
	writeJSON(w, http.StatusBadRequest, map[string]any{"error": msg, "field": field})
}

// emptyIfNil keeps a JSON array out of `null`, which every list consumer in the
// frontend would otherwise have to guard.
func emptyIfNil(list []*domain.Skill) []*domain.Skill {
	if list == nil {
		return []*domain.Skill{}
	}
	return list
}
