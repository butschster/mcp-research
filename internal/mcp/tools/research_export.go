package tools

import (
	"context"
	"log/slog"

	"github.com/butschster/mcp-research/internal/service"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type ResearchExportInput struct {
	ResearchID string `json:"research_id" jsonschema:"Research UUID or short code (e.g. R1)"`
}

func RegisterResearchExport(srv *mcp.Server, exportSvc *service.ExportService, log *slog.Logger) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "research_export",
		Description: "Export a complete research (sections, entries, sessions, questions, tasks, roadmaps) as portable JSON. Use this to transfer a research to another server via research_import.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input ResearchExportInput) (*mcp.CallToolResult, any, error) {
		if input.ResearchID == "" {
			return validationErrorResult([]string{"research_id is required"})
		}

		data, err := exportSvc.Export(ctx, input.ResearchID)
		if err != nil {
			log.Error("research_export failed", "error", err)
			return errorResult(err.Error())
		}

		return successResult(data)
	})
}
