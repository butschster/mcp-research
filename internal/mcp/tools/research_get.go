package tools

import (
	"context"
	"log/slog"

	"github.com/butschster/mcp-research/internal/service"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type ResearchGetInput struct {
	ResearchID string `json:"research_id" jsonschema:"ID of the research to retrieve"`
}

func RegisterResearchGet(srv *mcp.Server, researchSvc *service.ResearchService, sectionSvc *service.SectionService, sessionSvc *service.SessionService, log *slog.Logger) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "research_get",
		Description: "Returns full research context including sections with entry counts and active session. Use this to understand the current state of a research project.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input ResearchGetInput) (*mcp.CallToolResult, any, error) {
		if input.ResearchID == "" {
			return validationErrorResult([]string{"research_id is required"})
		}

		research, err := researchSvc.Get(ctx, input.ResearchID)
		if err != nil {
			return errorResult(err.Error())
		}

		sections, err := sectionSvc.List(ctx, input.ResearchID)
		if err != nil {
			return errorResult(err.Error())
		}

		var sectionData []map[string]any
		for _, s := range sections {
			count, _ := sectionSvc.CountEntries(ctx, s.ID)
			sectionData = append(sectionData, map[string]any{
				"id":            s.ID,
				"name":          s.Name,
				"display_name":  s.DisplayName,
				"description":   s.Description,
				"status":        s.Status,
				"position":      s.Position,
				"entries_count": count,
			})
		}

		var activeSession map[string]any
		if sessionSvc != nil {
			session, _ := sessionSvc.FindActive(ctx, input.ResearchID)
			if session != nil {
				activeSession = map[string]any{
					"id":     session.ID,
					"title":  session.Title,
					"focus":  session.Focus,
					"status": session.Status,
				}
			}
		}

		return successResult(map[string]any{
			"research":       research,
			"sections":       sectionData,
			"active_session": activeSession,
			"usage_hint":     "Use entry_create with section_id from the sections list above.",
		})
	})
}
