package tools

import (
	"context"
	"log/slog"

	"github.com/butschster/mcp-research/internal/service"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type SkillListInput struct {
	ResearchID *string `json:"research_id,omitempty" jsonschema:"Research whose methodology you are looking at. Accepts the UUID or the short code (R1). Give this or team_id"`
	// A team's library on its own terms, for a team whose skills are attached to
	// nothing yet — which is every team that has just written its first one.
	TeamID *string `json:"team_id,omitempty" jsonschema:"List a team's whole library instead, including skills no research follows. Get the id from team_list. Give this or research_id"`
	Query  *string `json:"query,omitempty" jsonschema:"Optional text to filter the research's library by name or trigger line. Ignored with team_id, which lists the team's skills whole"`
}

// RegisterSkillList is the call that precedes every other skill tool: what this
// research follows, what it could follow, and how much of the budget is left.
//
// research_get already carries the attached index, and it is repeated here on
// purpose. This is the tool an agent reaches for when it is deciding what to
// change, and a decision made from two calls — one for what is on, one for what
// is available — is a decision made from two moments.
func RegisterSkillList(srv *mcp.Server, skillSvc *service.SkillService, researchSvc *service.ResearchService, log *slog.Logger) {
	mcp.AddTool(srv, &mcp.Tool{
		Name: "skill_list",
		Description: "Lists the skills this research follows and the ones it could attach, with how many of its six slots are spent. " +
			"Call it before changing which methodology a research works by — not to read a skill, which is skill_load. " +
			"Pass team_id instead of research_id to list a team's whole library, including skills no research follows yet.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input SkillListInput) (*mcp.CallToolResult, any, error) {
		hasResearch := input.ResearchID != nil && *input.ResearchID != ""
		hasTeam := input.TeamID != nil && *input.TeamID != ""
		switch {
		case hasResearch && hasTeam:
			return validationErrorResult([]string{"give research_id or team_id, not both: one lists what a research follows, the other a team's whole library"})
		case !hasResearch && !hasTeam:
			return validationErrorResult([]string{"research_id or team_id is required"})
		}

		if hasTeam {
			// No cap and no `chosen` here: the budget belongs to a research, and
			// a number that looked like one over a team's library would be read
			// as a limit on how many skills a team may write. There is none.
			team, err := skillSvc.ListTeam(ctx, *input.TeamID)
			if err != nil {
				log.Error("skill_list team failed", "error", err)
				return skillErrorResult(err)
			}
			items := make([]map[string]any, 0, len(team))
			for _, sk := range team {
				items = append(items, skillPayload(sk))
			}
			return successResult(map[string]any{
				"library": items,
				"count":   len(items),
				"hint":    "A team's whole library, whether or not any research follows a given skill. Attach one to a research with skill_attach, read a body with skill_load.",
			})
		}

		researchID, err := researchSvc.ResolveID(ctx, *input.ResearchID)
		if err != nil {
			return errorResult(err.Error())
		}

		attached, err := skillSvc.ListAttached(ctx, researchID)
		if err != nil {
			log.Error("skill_list failed", "error", err)
			return skillErrorResult(err)
		}
		library, err := skillSvc.ListLibrary(ctx, researchID, derefStr(input.Query))
		if err != nil {
			log.Error("skill_list library failed", "error", err)
			return skillErrorResult(err)
		}

		following := make([]map[string]any, 0, len(attached))
		for _, sk := range attached {
			item := skillPayload(sk)
			if sk.ViaTemplate {
				// A methodology's choice reads differently from a person's, and
				// an agent tidying up should know which it is undoing.
				item["via_template"] = true
			}
			following = append(following, item)
		}

		available := make([]map[string]any, 0, len(library))
		for _, sk := range library {
			item := skillPayload(sk)
			if sk.Attached {
				item["attached"] = true
			}
			available = append(available, item)
		}

		chosen := service.CountChosen(attached)
		out := map[string]any{
			"following": following,
			"available": available,
			// Named for what it counts. The ambient product skills are in
			// `following` and outside the budget, so this number is legitimately
			// smaller than the list above it.
			"chosen": chosen,
			"cap":    service.SkillCap(),
			// One clause, and it is the only thing here the tool list does not
			// already say: the product skills are missing from `available`
			// because they are already on, not because they are unavailable.
			"hint": "`available` excludes the always-on product skills — they need no attaching.",
		}
		if chosen >= service.SkillCap() {
			out["cap_reached"] = true
			out["cap_hint"] = "Every slot is spent. Detach something before attaching or writing another; the ambient skills are not among them and cannot be dropped."
		}
		return successResult(out)
	})
}
