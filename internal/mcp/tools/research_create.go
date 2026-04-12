package tools

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/butschster/mcp-research/internal/service"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type ResearchCreateInput struct {
	Name        string              `json:"name" jsonschema:"Research name"`
	Description string              `json:"description" jsonschema:"Brief description of the research"`
	Goal        string              `json:"goal" jsonschema:"What the research aims to achieve"`
	Tags        []string            `json:"tags" jsonschema:"Tags for categorization"`
	Sections    []SectionSpecInput  `json:"sections" jsonschema:"Initial sections to create"`
}

type SectionSpecInput struct {
	Name                 string   `json:"name" jsonschema:"Section slug (lowercase alphanumeric with hyphens/underscores)"`
	DisplayName          string   `json:"display_name" jsonschema:"Human-readable section name"`
	Description          string   `json:"description" jsonschema:"Section description"`
	Position             int      `json:"position" jsonschema:"Sort order (0-based)"`
	AllowedEntryStatuses []string `json:"allowed_entry_statuses" jsonschema:"Allowed statuses for entries in this section. First status is the default. If omitted, defaults to [draft, active, completed, archived]"`
}

func RegisterResearchCreate(srv *mcp.Server, svc *service.ResearchService, log *slog.Logger) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "research_create",
		Description: "Creates a new research project with optional initial sections. Returns the research_id and count of sections created.",
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
				Name:                 s.Name,
				DisplayName:          s.DisplayName,
				Description:          s.Description,
				Position:             s.Position,
				AllowedEntryStatuses: s.AllowedEntryStatuses,
			})
		}

		research, createdSections, err := svc.Create(ctx, service.CreateResearchRequest{
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

		return successResult(map[string]any{
			"research_id":      research.ID,
			"code":             research.Code,
			"name":             research.Name,
			"status":           research.Status,
			"sections_created": len(createdSections),
			"hint":             "Use [[" + research.Code + "]] to reference this research from other entries. Use [[" + research.Code + ":E1]] to reference a specific entry.",
		})
	})
}
