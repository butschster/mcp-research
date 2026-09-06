package tools

import (
	"context"
	"log/slog"

	"github.com/dovod-app/app/internal/domain"
	"github.com/dovod-app/app/internal/service"
	"github.com/dovod-app/app/internal/storage"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type ResearchListInput struct {
	Status *string `json:"status" jsonschema:"Filter by status: active, completed, or archived. Leave empty for all."`
}

func RegisterResearchList(srv *mcp.Server, svc *service.ResearchService, log *slog.Logger) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "research_list",
		Description: "Lists the research projects of every team you belong to, with an optional status filter. Returns id, name, goal, status and tags for each, plus the team name when it is shared and access: read-only when your role does not allow writing.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input ResearchListInput) (*mcp.CallToolResult, any, error) {
		filter := storage.ResearchFilter{}
		if s := derefStr(input.Status); s != "" {
			status := domain.ResearchStatus(s)
			filter.Status = &status
		}

		researches, err := svc.List(ctx, filter)
		if err != nil {
			return errorResult(err.Error())
		}

		var items []map[string]any
		for _, r := range researches {
			item := map[string]any{
				"id":     r.ID,
				"name":   r.Name,
				"goal":   r.Goal,
				"status": r.Status,
				"tags":   r.Tags,
			}
			// A shared research looks exactly like your own until you try to
			// write to one you may only read. Saying so here means the agent
			// can pick a target instead of discovering the refusal.
			if r.TeamName != "" && !r.TeamPersonal {
				item["team"] = r.TeamName
			}
			if r.Role != "" && !r.Role.CanWrite() {
				item["access"] = "read-only"
			}
			items = append(items, item)
		}

		return successResult(map[string]any{
			"researches": items,
			"count":      len(items),
		})
	})
}
