package tools

import (
	"context"
	"log/slog"

	"github.com/butschster/mcp-research/internal/domain"
	"github.com/butschster/mcp-research/internal/service"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type ResearchUpdateInput struct {
	ResearchID  string   `json:"research_id" jsonschema:"ID of the research to update"`
	Name        *string  `json:"name" jsonschema:"New research name"`
	Description *string  `json:"description" jsonschema:"New description"`
	Goal        *string  `json:"goal" jsonschema:"New goal"`
	Status      *string  `json:"status" jsonschema:"New status: active, completed, or archived"`
	Instruction *string  `json:"instruction" jsonschema:"Working instructions for the LLM"`
	Tags        []string `json:"tags" jsonschema:"Replace tags (null=no change)"`
	Memory      []string `json:"memory" jsonschema:"Replace entire memory array (mutually exclusive with add_memory)"`
	AddMemory   *string  `json:"add_memory" jsonschema:"Append single entry to memory (mutually exclusive with memory)"`
}

func RegisterResearchUpdate(srv *mcp.Server, svc *service.ResearchService, log *slog.Logger) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "research_update",
		Description: "Updates a research project with partial updates. Supports memory (replace) and add_memory (append) which are mutually exclusive. Only provided fields are updated.",
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
			Instruction: input.Instruction,
			Tags:        input.Tags,
			Memory:      input.Memory,
			AddMemory:   input.AddMemory,
		})
		if err != nil {
			return errorResult(err.Error())
		}

		return successResult(map[string]any{
			"research_id": research.ID,
			"name":        research.Name,
			"status":      research.Status,
			"updated":     true,
		})
	})
}
