package tools

import (
	"context"
	"log/slog"

	"github.com/dovod-app/app/internal/service"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type SkillLoadInput struct {
	ResearchID *string `json:"research_id,omitempty" jsonschema:"Research whose skills you are working under. Accepts the UUID or the short code (R1). Give this with slug, or give skill_id"`
	// One slug, never a list. There is deliberately no batch form: loading
	// every skill at once is exactly what this design exists to prevent, and an
	// array parameter is an invitation to do it.
	Slug *string `json:"slug,omitempty" jsonschema:"Slug of the skill to open, from the skills index in research_get"`
	// The second address exists for a skill that has no research to be reached
	// through: one just written into a team library by skill_create, which
	// hands back an id and attaches it to nothing.
	SkillID *string `json:"skill_id,omitempty" jsonschema:"Id of the skill to open, from skill_create or skill_list. An alternative to research_id plus slug"`
}

func RegisterSkillLoad(srv *mcp.Server, skillSvc *service.SkillService, researchSvc *service.ResearchService, log *slog.Logger) {
	mcp.AddTool(srv, &mcp.Tool{
		Name: "skill_load",
		Description: "Opens one skill and returns its full text. " +
			"research_get lists the skills attached to a research with a name and a line saying when to use each; call this when you are about to do the work one of them names, not while orienting. One skill at a time. " +
			"Address it by research_id plus slug, or by skill_id for one that is in a team library and attached to nothing yet.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input SkillLoadInput) (*mcp.CallToolResult, any, error) {
		// The slug branch resolves the research first: the index the agent read
		// this slug out of came from research_get, which accepts a short code,
		// so an agent that passes `R7` straight back in would otherwise be told
		// the research does not exist.
		sk, fail := resolveSkillArg(ctx, skillSvc, researchSvc, input.ResearchID, input.Slug, input.SkillID, true)
		if fail != nil {
			return fail()
		}

		return successResult(map[string]any{
			"slug":        sk.Slug,
			"name":        sk.Name,
			"tier":        sk.Tier,
			"description": sk.Description,
			"body":        sk.Body,
			// Version and updated_at let an agent holding an older body know it
			// has moved on. Both are stored and bumped on every edit; without
			// them a second skill_load was the only way to find out.
			"version":    sk.Version,
			"updated_at": sk.UpdatedAt,
			// Precedence is expressed by tier and nothing else, so it is
			// restated on every load rather than left to the agent to remember
			// from a document it may not have read this session.
			"precedence": "A skill attached to this research directly overrides a team skill, which overrides a built-in. Where two skills conflict, follow the higher one.",
		})
	})
}
