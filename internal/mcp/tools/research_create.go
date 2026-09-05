package tools

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/dovod-app/app/internal/service"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type ResearchCreateInput struct {
	Name        string             `json:"name" jsonschema:"Research name"`
	Description string             `json:"description" jsonschema:"Brief description of the research"`
	Goal        string             `json:"goal" jsonschema:"What the research aims to achieve"`
	Tags        []string           `json:"tags" jsonschema:"Tags for categorization"`
	Sections    []SectionSpecInput `json:"sections" jsonschema:"Initial sections to create"`
	// Without this the agent's work always landed in the caller's personal
	// team and had to be moved by hand afterwards, which is the wrong default
	// for anyone who set a team up in order to work in it.
	TeamID *string `json:"team_id,omitempty" jsonschema:"Team to create it in. Use team_list to find one. Defaults to your personal team."`
	// The one structural thing a template does. Everything else about a
	// template is text the model read before getting here.
	TemplateSlug *string `json:"template_slug,omitempty" jsonschema:"Slug of the methodology you followed, from template_list. Records it on the research and attaches the skills that methodology names."`
}

type SectionSpecInput struct {
	Name        string `json:"name" jsonschema:"Section slug (lowercase alphanumeric with hyphens/underscores)"`
	DisplayName string `json:"display_name" jsonschema:"Human-readable section name"`
	Description string `json:"description" jsonschema:"Section description"`
	Position    int    `json:"position" jsonschema:"Sort order (0-based)"`
}

func RegisterResearchCreate(srv *mcp.Server, svc *service.ResearchService, templates *service.TemplateService, log *slog.Logger) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "research_create",
		Description: "Creates a new research project with optional initial sections, in your personal team unless team_id names another one you may write to (see team_list). Returns the research_id and count of sections created.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input ResearchCreateInput) (*mcp.CallToolResult, any, error) {
		var errs []string
		if input.Name == "" {
			errs = append(errs, "name is required")
		}
		for i, s := range input.Sections {
			if !isValidSlug(s.Name) {
				errs = append(errs, fmt.Sprintf("sections[%d].name must be lowercase alphanumeric with hyphens or underscores", i))
			}
		}
		if len(errs) > 0 {
			return validationErrorResult(errs)
		}

		var sections []service.CreateSectionRequest
		for _, s := range input.Sections {
			sections = append(sections, service.CreateSectionRequest{
				Name:        s.Name,
				DisplayName: s.DisplayName,
				Description: s.Description,
				Position:    s.Position,
			})
		}

		research, createdSections, err := svc.Create(ctx, service.CreateResearchRequest{
			TeamID:      derefStr(input.TeamID),
			Name:        input.Name,
			Description: input.Description,
			Goal:        input.Goal,
			Tags:        input.Tags,
			Sections:    sections,
		})
		if err != nil {
			log.Error("research_create failed", "error", err)
			return errorResult(err.Error())
		}

		// Recording the methodology and attaching its skills happens after the
		// research exists, and never fails the creation: a research that exists
		// with one skill missing is a better outcome than a create call rolled
		// back because a template named something a later build removed.
		var attach service.AttachResult
		requestedTemplate := derefStr(input.TemplateSlug)
		if requestedTemplate != "" && templates != nil {
			attach = templates.AttachSkills(ctx, research.ID, requestedTemplate)
		}

		result := map[string]any{
			"research_id":      research.ID,
			"code":             research.Code,
			"name":             research.Name,
			"status":           research.Status,
			"sections_created": len(createdSections),
			"hint":             "Use [[" + research.Code + "]] to reference this research from other entries. Use [[" + research.Code + ":E1]] to reference a specific entry.",
		}
		if research.TeamName != "" && !research.TeamPersonal {
			result["team"] = research.TeamName
		}
		if len(attach.Attached) > 0 {
			result["skills_attached"] = attach.Attached
		}
		// Reported rather than swallowed: the methodology named something the
		// research did not get, and the conductor should know which part of it
		// will not be available.
		if len(attach.Missed) > 0 {
			result["skills_unavailable"] = attach.Missed
		}
		// The failure that used to be invisible. A slug nobody can resolve —
		// usually a name where a slug was wanted — produced a response
		// identical to a successful one, with the research left unstamped.
		if requestedTemplate != "" && !attach.Resolved {
			result["template_unresolved"] = requestedTemplate
			result["template_hint"] = "No template of that name is visible to you, so nothing was recorded and no skills were attached. Use the slug from template_list."
		}
		return successResult(result)
	})
}
