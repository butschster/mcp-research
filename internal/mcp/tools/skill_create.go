package tools

import (
	"context"
	"log/slog"

	"github.com/butschster/mcp-research/internal/domain"
	"github.com/butschster/mcp-research/internal/service"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type SkillCreateInput struct {
	// Exactly one of these two decides the tier, which is why neither is a
	// flag on the other: "private or team" is not a property of a skill you
	// might toggle later, it is which thing owns it.
	ResearchID *string `json:"research_id,omitempty" jsonschema:"Write a skill belonging to this research alone, and attach it. UUID or short code (R1). Give this or team_id"`
	TeamID     *string `json:"team_id,omitempty" jsonschema:"Write a skill into this team's library instead, attached to nothing. Id from team_list. Give this or research_id"`

	// The reasoning behind these three lives in the refusals and in
	// /llms/skills.md. A schema is paid for on every call of every session; a
	// refusal is read once, by the caller that needs it.
	Name        string `json:"name" jsonschema:"Short name of the methodology, e.g. 'Grading evidence'"`
	Description string `json:"description" jsonschema:"The trigger line: when to load this, not what it contains. Start with 'Use when'. At most 200 characters"`
	Body        string `json:"body" jsonschema:"The methodology itself, in markdown. 600-2500 tokens; at most 16000 characters"`
}

// RegisterSkillCreate writes a new skill at one of the two authorable tiers.
//
// The built-in tier is not among them and never will be: those ship embedded in
// the binary and are rewritten on every boot, so a row written there would be
// destroyed by the next upgrade.
func RegisterSkillCreate(srv *mcp.Server, skillSvc *service.SkillService, researchSvc *service.ResearchService, log *slog.Logger) {
	mcp.AddTool(srv, &mcp.Tool{
		Name: "skill_create",
		Description: "Writes a new methodology skill. Give research_id for one that applies to this research alone (it is attached immediately and spends a slot), or team_id for one your colleagues' researches can attach too. " +
			"Use private skills for this research's working rules and team skills for reusable methodology; goal and description explain its scope.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input SkillCreateInput) (*mcp.CallToolResult, any, error) {
		var errs []string
		hasResearch := input.ResearchID != nil && *input.ResearchID != ""
		hasTeam := input.TeamID != nil && *input.TeamID != ""
		switch {
		case hasResearch && hasTeam:
			errs = append(errs, "give research_id or team_id, not both: they are two different owners")
		case !hasResearch && !hasTeam:
			errs = append(errs, "research_id or team_id is required: a skill belongs to one research or to a team's library")
		}
		if input.Name == "" {
			errs = append(errs, "name is required")
		}
		if input.Description == "" {
			errs = append(errs, "description is required: it is the line the agent decides from, so start it with \"Use when\"")
		}
		if input.Body == "" {
			errs = append(errs, "body is required")
		}
		if len(errs) > 0 {
			return validationErrorResult(errs)
		}

		in := service.SkillInput{Name: input.Name, Description: input.Description, Body: input.Body}

		if hasTeam {
			sk, err := skillSvc.CreateTeam(ctx, *input.TeamID, in)
			if err != nil {
				log.Error("skill_create team failed", "error", err)
				return skillErrorResult(err)
			}
			out := skillPayload(sk)
			out["created"] = true
			// Said explicitly because it is the one surprise in this tool: a
			// team skill is written and then followed by nobody. An agent that
			// wrote it for the research it is working on has to attach it.
			out["hint"] = "It is in the team library and attached to nothing. Call skill_attach with this slug to make a research follow it."
			return successResult(out)
		}

		researchID, err := researchSvc.ResolveID(ctx, *input.ResearchID)
		if err != nil {
			return errorResult(err.Error())
		}
		sk, err := skillSvc.CreatePrivate(ctx, researchID, in)
		if err != nil {
			log.Error("skill_create private failed", "error", err)
			return skillErrorResult(err)
		}
		out := skillPayload(sk)
		out["created"] = true
		out["attached"] = true
		out["hint"] = "A research-private skill is attached on creation and spends one of the six slots. It is never offered to another research; skill_promote moves it into the team library if it turns out to be reusable."
		if sk.Tier != domain.SkillPrivate {
			// Defensive: the tier is decided by the service, and a payload that
			// claimed private while the row said otherwise would mislead every
			// later decision about precedence.
			out["tier"] = sk.Tier
		}
		return successResult(out)
	})
}
