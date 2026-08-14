package tools

import (
	"context"
	"errors"
	"log/slog"

	"github.com/butschster/mcp-research/internal/service"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type TeamListInput struct{}

// RegisterTeamList exposes the teams the caller belongs to.
//
// It exists so `research_create` has something to aim at: without a way to
// learn a team's id, an optional `team_id` parameter is a parameter nobody can
// fill in, and every research an agent makes lands in a personal team and has
// to be moved by hand.
func RegisterTeamList(srv *mcp.Server, svc *service.TeamService, log *slog.Logger) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "team_list",
		Description: "Lists the teams you belong to, with your role in each. A team owns researches: everyone in it can read them, and editors and owners can write. Pass a team id to research_create to put new work where your colleagues can see it.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input TeamListInput) (*mcp.CallToolResult, any, error) {
		teams, err := svc.List(ctx)
		if err != nil {
			// Auth is off: there are no users, so there are no teams to be in,
			// and saying so beats an error the agent cannot act on.
			if errors.Is(err, service.ErrNoAuth) {
				return successResult(map[string]any{
					"teams": []any{},
					"count": 0,
					"note":  "This server runs without accounts, so there are no teams. Every research is shared by whoever can reach it.",
				})
			}
			log.Error("team_list failed", "error", err)
			return errorResult(err.Error())
		}

		items := make([]map[string]any, 0, len(teams))
		for _, t := range teams {
			item := map[string]any{
				"team_id":    t.ID,
				"name":       t.Name,
				"role":       t.Role,
				"members":    t.MemberCount,
				"researches": t.ResearchCount,
			}
			if t.Personal {
				item["personal"] = true
			}
			if !t.Role.CanWrite() {
				item["access"] = "read-only"
			}
			items = append(items, item)
		}

		return successResult(map[string]any{"teams": items, "count": len(items)})
	})
}
