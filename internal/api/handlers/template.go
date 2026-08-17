package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/butschster/mcp-research/internal/domain"
	"github.com/butschster/mcp-research/internal/service"
)

type TemplateHandler struct {
	templates *service.TemplateService
	research  *service.ResearchService
	section   *service.SectionService
	skills    *service.SkillService
	log       *slog.Logger
}

func NewTemplateHandler(templates *service.TemplateService, research *service.ResearchService, section *service.SectionService, skills *service.SkillService, log *slog.Logger) *TemplateHandler {
	return &TemplateHandler{templates: templates, research: research, section: section, skills: skills, log: log}
}

type templateRequest struct {
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	WhenToUse    string   `json:"when_to_use"`
	WhenNotToUse string   `json:"when_not_to_use"`
	Body         string   `json:"body"`
	Skills       []string `json:"skills"`
	// TeamID names which library a fork goes into. It is required — there is no
	// personal-team fallback, because guessing which library somebody's copy of
	// a shipped methodology belongs in is the kind of guess that is wrong
	// silently. Creating a template names the team in the path instead.
	TeamID string `json:"team_id"`
}

func (r templateRequest) input() service.TemplateInput {
	return service.TemplateInput{
		Name: r.Name, Description: r.Description, WhenToUse: r.WhenToUse,
		WhenNotToUse: r.WhenNotToUse, Body: r.Body, Skills: r.Skills,
	}
}

// List returns every template the caller may use — no bodies.
func (h *TemplateHandler) List(w http.ResponseWriter, r *http.Request) {
	list, err := h.templates.List(r.Context())
	if err != nil {
		writeTemplateError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": emptyTemplates(list)})
}

// Get returns one template with its body.
func (h *TemplateHandler) Get(w http.ResponseWriter, r *http.Request) {
	tp, err := h.templates.Get(r.Context(), r.PathValue("slug"))
	if err != nil {
		writeTemplateError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": tp})
}

func (h *TemplateHandler) ListTeam(w http.ResponseWriter, r *http.Request) {
	list, err := h.templates.ListByTeam(r.Context(), r.PathValue("id"))
	if err != nil {
		writeTemplateError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": emptyTemplates(list)})
}

func (h *TemplateHandler) CreateTeam(w http.ResponseWriter, r *http.Request) {
	var req templateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	tp, err := h.templates.CreateTeam(r.Context(), r.PathValue("id"), req.input())
	if err != nil {
		writeTemplateError(w, err)
		return
	}
	out := map[string]any{"data": tp}
	// A slug that resolves to nothing is not fatal — a team may write a template
	// before the skill it names — but silence means the typo surfaces months
	// later as `skills_unavailable` on somebody's kickoff, which is the worst
	// possible moment and the furthest possible place from the author.
	if warnings := h.templates.UnknownSkills(r.Context(), tp); len(warnings) > 0 {
		out["warnings"] = warnings
	}
	writeJSON(w, http.StatusCreated, out)
}

// Fork copies a template that ships with the app into a team and applies the
// edit. The response says `forked` so a client can tell this from an update and
// follow the slug to the new row.
func (h *TemplateHandler) Fork(w http.ResponseWriter, r *http.Request) {
	var req templateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if req.TeamID == "" {
		// Without this the refusal was a bare 404, identical to "no such
		// template" and to "you are not in that team" — on a route naming a
		// template the caller has just successfully read.
		writeFieldError(w, "team_id is required: a fork belongs to a team, and there is no default", "team_id")
		return
	}
	tp, err := h.templates.Fork(r.Context(), req.TeamID, r.PathValue("slug"), req.input())
	if err != nil {
		writeTemplateError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": tp, "forked": true})
}

func (h *TemplateHandler) Update(w http.ResponseWriter, r *http.Request) {
	var req templateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	tp, err := h.templates.Update(r.Context(), r.PathValue("templateId"), req.input())
	if err != nil {
		writeTemplateError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": tp})
}

func (h *TemplateHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.templates.Delete(r.Context(), r.PathValue("templateId")); err != nil {
		writeTemplateError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// DraftFromResearch returns a *starting point*, not a saved structure.
//
// It is deliberately not a create call. "Save as template" cannot mean what it
// used to: a template carries no rows, so there is nothing to capture. What
// comes back is a body skeleton with the judgement left blank, for a human or an
// agent to write and then post.
func (h *TemplateHandler) DraftFromResearch(w http.ResponseWriter, r *http.Request) {
	id, err := h.research.ResolveID(r.Context(), r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	research, err := h.research.Get(r.Context(), id)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	sections, err := h.section.List(r.Context(), id)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	skills, err := h.skills.ListAttached(r.Context(), id)
	if err != nil {
		// A draft without the skill list is still a useful draft, and failing
		// the whole call because one list was unavailable helps nobody.
		h.log.Warn("template draft: skills unavailable", "research_id", id, "error", err)
	}
	draft := h.templates.DraftFromResearch(r.Context(), research, sections, skills)
	writeJSON(w, http.StatusOK, map[string]any{
		"data": map[string]any{
			"name":            draft.Name,
			"when_to_use":     draft.WhenToUse,
			"when_not_to_use": draft.WhenNotToUse,
			"body":            draft.Body,
			"skills":          emptyStrings(draft.Skills),
		},
		"hint": "This is a skeleton, not a methodology. Fill in what to ask before proposing anything, what good work looks like, and when it is finished — then POST it to /api/teams/{id}/templates.",
	})
}

func writeTemplateError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrTemplateSlugTaken):
		writeCodedError(w, http.StatusConflict, err.Error(), "slug_taken")
	case errors.Is(err, service.ErrTemplateGlobalWrite):
		writeCodedError(w, http.StatusForbidden, err.Error(), "not_allowed")
	case errors.Is(err, service.ErrTemplateNameEmpty):
		writeFieldError(w, err.Error(), "name")
	case errors.Is(err, service.ErrTemplateWhenEmpty), errors.Is(err, service.ErrTemplateCriterionLong):
		writeFieldError(w, err.Error(), "when_to_use")
	case errors.Is(err, service.ErrTemplateBodyEmpty), errors.Is(err, service.ErrTemplateBodyLong):
		writeFieldError(w, err.Error(), "body")
	default:
		writeServiceError(w, err)
	}
}

// emptyStrings keeps a JSON array out of `null`, so a client can round-trip the
// draft straight back into a create call.
func emptyStrings(list []string) []string {
	if list == nil {
		return []string{}
	}
	return list
}

func emptyTemplates(list []*domain.Template) []*domain.Template {
	if list == nil {
		return []*domain.Template{}
	}
	return list
}
