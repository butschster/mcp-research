package tools

import (
	"context"
	"errors"
	"log/slog"

	"github.com/dovod-app/app/internal/service"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type SkillForkInput struct {
	ResearchID string `json:"research_id" jsonschema:"Research whose attachment moves onto the copy. Accepts the UUID or the short code (R1)"`
	Slug       string `json:"slug" jsonschema:"Slug of the built-in skill to fork"`

	Name        *string `json:"name,omitempty" jsonschema:"New name for the copy. Omit to keep the built-in's"`
	Description *string `json:"description,omitempty" jsonschema:"New trigger line, at most 200 characters. Omit to keep the built-in's"`
	Body        *string `json:"body,omitempty" jsonschema:"New markdown body. Omit to keep the built-in's, which is what you want if you are about to edit it with skill_update"`
}

// RegisterSkillFork is how a built-in gets changed: not in place, which the
// next upgrade would undo, but as a team copy that keeps the same slug.
//
// The copy and the re-attachment happen together on purpose. Detach-then-attach
// could hit the six-skill cap between the two calls and leave the research
// following one fewer skill than it started with, which is a hole nobody would
// think to look for.
func RegisterSkillFork(srv *mcp.Server, skillSvc *service.SkillService, researchSvc *service.ResearchService, log *slog.Logger) {
	mcp.AddTool(srv, &mcp.Tool{
		Name: "skill_fork",
		Description: "Copies a built-in skill into your team's library so it can be edited, and moves this research's attachment onto the copy in one step. " +
			"The copy keeps the same slug — it is the same methodology, edited — and the original built-in is untouched. Send only the fields you are changing; the rest are inherited.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input SkillForkInput) (*mcp.CallToolResult, any, error) {
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
		// Every field may be omitted: the service inherits each one from the
		// built-in, including the body, which it reads itself. A fork sent with
		// nothing but a slug is a legitimate first step — take a copy now, edit
		// it with skill_update later.
		sk, err := skillSvc.Fork(ctx, researchID, input.Slug, skillInput(nil, input.Name, input.Description, input.Body))
		if err != nil {
			log.Error("skill_fork failed", "slug", input.Slug, "error", err)
			// Fork refuses anything that is not a built-in with a bare
			// ErrNotFound, which reads as "no such skill" — and an agent that
			// believes that goes looking for the slug it can plainly see in the
			// index. Every other wrong-kind refusal in this feature names the
			// tool that does work, so this one does too.
			if errors.Is(err, service.ErrNotFound) {
				if existing, resolveErr := skillSvc.ResolveSlug(ctx, researchID, input.Slug); resolveErr == nil && existing != nil {
					return errorResult("not_allowed: " + input.Slug + " is a " + string(existing.Tier) +
						" skill, and forking only applies to a built-in. A team skill is edited in place with skill_update; skill_copy takes one into this research as a private copy instead.")
				}
			}
			return skillErrorForSlug(err, input.Slug)
		}

		out := skillPayload(sk)
		out["forked"] = true
		out["hint"] = "The copy belongs to your team and this research now follows it instead of the built-in. Other researches keep following the original until they attach this slug themselves."
		return successResult(out)
	})
}
