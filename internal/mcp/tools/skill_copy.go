package tools

import (
	"context"
	"log/slog"

	"github.com/dovod-app/app/internal/domain"
	"github.com/dovod-app/app/internal/service"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type SkillCopyInput struct {
	ResearchID string `json:"research_id" jsonschema:"Research that gets its own copy. Accepts the UUID or the short code (R1)"`
	Slug       string `json:"slug" jsonschema:"Slug of the team or built-in skill to copy down"`
}

// RegisterSkillCopy takes a shared skill private, so it can be bent to one
// research without changing it for anyone else.
func RegisterSkillCopy(srv *mcp.Server, skillSvc *service.SkillService, researchSvc *service.ResearchService, log *slog.Logger) {
	mcp.AddTool(srv, &mcp.Tool{
		Name: "skill_copy",
		Description: "Copies a team or built-in skill into this research as a private one, and moves the attachment onto the copy in one step. " +
			"Use it when a shared methodology is nearly right and this research needs it changed — editing the shared original would change it for every research that follows it.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input SkillCopyInput) (*mcp.CallToolResult, any, error) {
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
		before, resolveErr := skillSvc.ResolveSlug(ctx, researchID, input.Slug)

		sk, err := skillSvc.CopyHere(ctx, researchID, input.Slug)
		if err != nil {
			log.Error("skill_copy failed", "slug", input.Slug, "error", err)
			return skillErrorForSlug(err, input.Slug)
		}

		out := skillPayload(sk)
		// CopyHere answers with the original when it is already private, which
		// is the right thing for it to do and the wrong thing to report as a
		// copy. The agent would otherwise believe it had made something.
		if resolveErr == nil && before != nil && before.Tier == domain.SkillPrivate {
			out["copied"] = false
			out["hint"] = "Already private to this research — nothing was copied. Edit it with skill_update."
			return successResult(out)
		}
		out["copied"] = true
		out["hint"] = "This research now follows its own copy. Edit it with skill_update; skill_promote puts it back in the team library if it turns out to be reusable."
		return successResult(out)
	})
}
