package tools

import (
	"context"
	"log/slog"

	"github.com/dovod-app/app/internal/service"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type ResearchResumeInput struct {
	ResearchID string `json:"research_id" jsonschema:"ID or short code (R1) of the research to continue"`
	// Both are required-and-nullable, like nearly every other tool here: send
	// `null` rather than leaving the property out. The schema is generated from
	// these tags, and a client that omits a property gets a protocol error
	// before the handler is ever reached — so the wording has to say `null`.
	SessionID *string `json:"session_id" jsonschema:"Which session you are continuing — UUID or short code (SS1) inside this research. Send null: with one active session the server picks it, with several it returns the candidates and asks which"`
	Limit     *int    `json:"limit" jsonschema:"How many items per group. Send null for the default of 5; values outside 1-15 are clamped"`
}

func RegisterResearchResume(srv *mcp.Server, svc *service.ResumeService, log *slog.Logger) {
	openWorld := false
	mcp.AddTool(srv, &mcp.Tool{
		Name: "research_resume",
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint: true,
			// The tool reaches nothing outside this server's own database.
			OpenWorldHint: &openWorld,
		},
		Description: `Returns the outstanding work in a research: tasks in progress, blocked and waiting, the open questions of the session you are continuing, the marks a person left, the documents that changed most recently, and up to three candidate next actions with the reason for each.

Call it after research_get when a new chat opens on an existing research. research_get carries the constraints — structured memory, methodology, the skills index — and this carries the queue. Neither replaces the other, and this one deliberately repeats none of the first.

What it does NOT do, and what that means for you:
  - It writes nothing. No session is created, no status moves, nothing is marked as read. Starting or continuing work is still your explicit call.
  - It does not choose between two open sessions. With several active it returns them with selection_required and no selected_id: ask which one, or pass session_id.
  - Both optional inputs are required-and-nullable: send "session_id": null and "limit": null when you have nothing to say for them.
  - It is not a change log. recent_entries is what was touched most recently, not everything that happened, and a deleted document leaves no trace here.
  - It is a top-N per group. Each group carries returned, total and has_more with the tool to open the rest — never read an empty top-N as "the work is finished".

next_actions carries an actor. ` + "`agent`" + ` is yours to do. ` + "`human`" + ` is waiting on a person — an answered mark needs the person who raised it to accept it, and you cannot accept your own answer.

A person opens this with "Continue R1", or "Continue R1, session SS4" when several sessions are open — that is the sentence the web UI hands them. What it obliges you to do is written down rather than left to improvisation: /llms/conducting-research.md, "Picking Up a Research That Is Already Running".`,
	}, func(ctx context.Context, req *mcp.CallToolRequest, input ResearchResumeInput) (*mcp.CallToolResult, any, error) {
		if input.ResearchID == "" {
			return validationErrorResult([]string{"research_id is required"})
		}

		resume, err := svc.Get(ctx, input.ResearchID, service.ResumeRequest{
			SessionID: derefStr(input.SessionID),
			Limit:     derefInt(input.Limit),
		})
		if err != nil {
			return errorResult(err.Error())
		}
		return successResult(resume)
	})
}
