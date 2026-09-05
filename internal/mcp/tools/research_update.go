package tools

import (
	"context"
	"log/slog"

	"github.com/dovod-app/app/internal/domain"
	"github.com/dovod-app/app/internal/service"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type ResearchUpdateInput struct {
	ResearchID  string   `json:"research_id" jsonschema:"ID of the research to update"`
	Name        *string  `json:"name" jsonschema:"New research name"`
	Description *string  `json:"description" jsonschema:"New description"`
	Goal        *string  `json:"goal" jsonschema:"New goal"`
	Status      *string  `json:"status" jsonschema:"New status: active, completed, or archived"`
	Tags        []string `json:"tags" jsonschema:"Replace tags (null=no change)"`
	AddMemory   *string  `json:"add_memory,omitempty" jsonschema:"Append one memory item. Use research_memory to edit or delete by ID; use private skills for instructions"`
	SessionID   string   `json:"session_id,omitempty" jsonschema:"Research session UUID or SS code for add_memory provenance; omit when not working in a research session"`
}

func RegisterResearchUpdate(srv *mcp.Server, svc *service.ResearchService, log *slog.Logger) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "research_update",
		Description: "Updates research metadata and optionally appends one memory item. Only provided fields are updated. Memory edits use research_memory; methodology belongs in private skills.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input ResearchUpdateInput) (*mcp.CallToolResult, any, error) {
		if input.ResearchID == "" {
			return validationErrorResult([]string{"research_id is required"})
		}

		var status *domain.ResearchStatus
		if input.Status != nil {
			s := domain.ResearchStatus(*input.Status)
			status = &s
		}

		research, err := svc.Update(ctx, input.ResearchID, service.UpdateResearchRequest{
			Name:        input.Name,
			Description: input.Description,
			Goal:        input.Goal,
			Status:      status,
			Tags:        input.Tags,
			AddMemory:   input.AddMemory,
			SessionID:   input.SessionID,
		})
		if err != nil {
			return errorResult(err.Error())
		}

		return successResult(map[string]any{
			"research_id": research.ID,
			"name":        research.Name,
			"status":      research.Status,
			"updated":     true,
			"memory":      research.Memory,
		})
	})
}
