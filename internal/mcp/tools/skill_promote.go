package tools

import (
	"context"
	"log/slog"

	"github.com/dovod-app/app/internal/service"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type SkillPromoteInput struct {
	ResearchID string `json:"research_id" jsonschema:"Research the private skill belongs to. Accepts the UUID or the short code (R1)"`
	Slug       string `json:"slug" jsonschema:"Slug of the research-private skill to move into the team library"`
}

// RegisterSkillPromote moves a rule that turned out to be reusable up into the
// team's library, so the next research does not have to rediscover it.
//
// The private original is deleted rather than left behind: two rows with one
// slug in the same research resolve to the private one, so a promotion that
// kept it would appear to have done nothing.
func RegisterSkillPromote(srv *mcp.Server, skillSvc *service.SkillService, researchSvc *service.ResearchService, log *slog.Logger) {
	mcp.AddTool(srv, &mcp.Tool{
		Name: "skill_promote",
		Description: "Moves a research-private skill into the team library, where other researches can attach it, and keeps this research following it. " +
			"Use it when a rule written for one research turns out to describe how the team works generally.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input SkillPromoteInput) (*mcp.CallToolResult, any, error) {
		var errs []string
		if input.ResearchID == "" {
			errs = append(errs, "research_id is required")
		}
		if input.Slug == "" {
			errs = append(errs, "slug is required")
		}
		if len(errs) > 0 {
			return validationErrorResult(errs)
		}

		researchID, err := researchSvc.ResolveID(ctx, input.ResearchID)
		if err != nil {
			return errorResult(err.Error())
		}
		sk, err := skillSvc.Promote(ctx, researchID, input.Slug)
		if err != nil {
			log.Error("skill_promote failed", "slug", input.Slug, "error", err)
			// slug_taken here has one cause worth naming: the team library
			// already holds this slug, usually a fork of the same built-in that
			// this private copy came from. Renaming does not help — a slug is
			// fixed at creation and skill_update leaves it alone — so the way
			// out is to edit or delete the row already holding it, and an agent
			// told only "a skill with that name already exists here" will try
			// the rename and loop.
			if service.SkillErrorCode(err) == "slug_taken" {
				return errorResult("slug_taken: the team library already holds a skill with the slug " + input.Slug +
					" — most often a fork of the same built-in. A slug is fixed at creation and skill_update does not change it, so renaming will not clear this. " +
					"Find the team's row with skill_list(team_id) — team_list gives the id — and edit it in place with skill_update(skill_id). " +
					"Do NOT call skill_delete with this research_id and this slug: a slug resolves to what the research follows first, which is the private copy you are trying to save, and deleting it is not recoverable.")
			}
			return skillErrorForSlug(err, input.Slug)
		}

		out := skillPayload(sk)
		out["promoted"] = true
		out["hint"] = "It is a team skill now and this research still follows it. Editing it from here changes it for every research that attaches it."
		return successResult(out)
	})
}
