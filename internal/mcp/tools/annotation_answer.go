package tools

import (
	"context"
	"log/slog"

	"github.com/butschster/mcp-research/internal/service"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type AnnotationAnswerInput struct {
	AnnotationID string  `json:"annotation_id" jsonschema:"ID of the annotation, from annotation_list"`
	Resolution   string  `json:"resolution" jsonschema:"What you did about it, and what it rests on. Name the entry, source or question you produced — [[E19]] and [[Q7]] become real links"`
	TaskID       *string `json:"task_id" jsonschema:"Task this mark was promoted to, when the work does not fit in a pass"`
}

func RegisterAnnotationAnswer(srv *mcp.Server, svc *service.AnnotationService, log *slog.Logger) {
	mcp.AddTool(srv, &mcp.Tool{
		Name: "annotation_answer",
		Description: `Records what you did about one mark and moves it to 'answered' — done and waiting for a person to agree that it is done.

You cannot close or dismiss a mark. That is deliberate: an agent that accepts its own work turns this into a machine that confirms itself. Answer, and let them accept.

Write a resolution somebody can check without re-doing the work: what you found, where it came from, and what you changed. Cite what you produced — [[E19]] for an entry you wrote, [[Q7]] for a question you asked — and those become real links.

When you CANNOT answer without them, do not invent one. Create a question in the active session with question_create, then answer here naming it: "needs them — asked [[Q9]]". That is the honest outcome, and it is how a mark stops going round in circles.

For a 'disagree' mark, answering does not mean the objection went away. Record both positions in the document and say here where you put them.`,
	}, func(ctx context.Context, req *mcp.CallToolRequest, input AnnotationAnswerInput) (*mcp.CallToolResult, any, error) {
		var errs []string
		if input.AnnotationID == "" {
			errs = append(errs, "annotation_id is required")
		}
		if input.Resolution == "" {
			errs = append(errs, "resolution is required — a mark moved to answered with nothing recorded cannot be reviewed")
		}
		if len(errs) > 0 {
			return validationErrorResult(errs)
		}

		a, err := svc.Answer(ctx, input.AnnotationID, service.AnswerAnnotationRequest{
			Resolution: input.Resolution,
			TaskID:     derefStr(input.TaskID),
		})
		if err != nil {
			return errorResult(err.Error())
		}

		out := map[string]any{
			"code":       a.Code,
			"status":     a.Status,
			"entry_code": a.EntryCode,
			"kind":       a.Kind,
		}
		if a.ResolvedRevision != nil {
			out["resolved_revision"] = *a.ResolvedRevision
		}
		return successResult(out)
	})
}
