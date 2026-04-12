package tools

import (
	"context"
	"log/slog"

	"github.com/butschster/mcp-research/internal/service"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type ResearchAddSectionInput struct {
	ResearchID           string   `json:"research_id" jsonschema:"ID of the research"`
	Name                 string   `json:"name" jsonschema:"Section slug (lowercase alphanumeric with hyphens/underscores)"`
	DisplayName          string   `json:"display_name" jsonschema:"Human-readable section name"`
	Description          string   `json:"description" jsonschema:"Section description"`
	Position             int      `json:"position" jsonschema:"Sort order (0-based)"`
	AllowedEntryStatuses []string `json:"allowed_entry_statuses" jsonschema:"Allowed statuses for entries in this section. First status is the default. If omitted, defaults to [draft, active, completed, archived]"`
}

func RegisterResearchAddSection(srv *mcp.Server, svc *service.ResearchService, log *slog.Logger) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "research_add_section",
		Description: "Adds a new section to an existing research project. Section name must be a lowercase slug (alphanumeric, hyphens, underscores). Returns section_id.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input ResearchAddSectionInput) (*mcp.CallToolResult, any, error) {
		var errs []string
		if input.ResearchID == "" {
			errs = append(errs, "research_id is required")
		}
		if input.Name == "" {
			errs = append(errs, "name is required")
		} else if !isValidSlug(input.Name) {
			errs = append(errs, "name must be lowercase alphanumeric with hyphens or underscores")
		}
		if input.DisplayName == "" {
			errs = append(errs, "display_name is required")
		}
		if len(errs) > 0 {
			return validationErrorResult(errs)
		}

		section, err := svc.AddSection(ctx, input.ResearchID, service.CreateSectionRequest{
			Name:                 input.Name,
			DisplayName:          input.DisplayName,
			Description:          input.Description,
			Position:             input.Position,
			AllowedEntryStatuses: input.AllowedEntryStatuses,
		})
		if err != nil {
			return errorResult(err.Error())
		}

		return successResult(map[string]any{
			"section_id":             section.ID,
			"name":                   section.Name,
			"display_name":           section.DisplayName,
			"status":                 section.Status,
			"allowed_entry_statuses": section.AllowedEntryStatuses,
		})
	})
}
