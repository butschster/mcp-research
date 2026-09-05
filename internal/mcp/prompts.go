package mcp

import (
	"context"
	"strings"

	"github.com/dovod-app/app/internal/docs"
	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

func (s *Server) registerPrompts() {
	initText, _ := docs.ReadFile("research_initialize.md")
	conductText, _ := docs.ReadFile("research_conduct.md")

	s.server.AddPrompt(&sdkmcp.Prompt{
		Name:        "research/initialize",
		Description: "Interactively design and initialize a new research project. Guides you through defining goals, sections, and structure.",
		Arguments: []*sdkmcp.PromptArgument{
			{Name: "topic", Description: "Optional starting topic hint", Required: false},
		},
	}, func(ctx context.Context, req *sdkmcp.GetPromptRequest) (*sdkmcp.GetPromptResult, error) {
		text := string(initText)
		if topic := req.Params.Arguments["topic"]; topic != "" {
			text = strings.ReplaceAll(text, "{topic}", topic)
		}
		return &sdkmcp.GetPromptResult{
			Description: "Research initialization workflow",
			Messages: []*sdkmcp.PromptMessage{
				{Role: "user", Content: &sdkmcp.TextContent{Text: text}},
			},
		}, nil
	})

	s.server.AddPrompt(&sdkmcp.Prompt{
		Name:        "research/conduct",
		Description: "Systematically execute a research project by interviewing the user and creating entries. Requires an existing research_id.",
		Arguments: []*sdkmcp.PromptArgument{
			{Name: "research_id", Description: "ID of the research to conduct", Required: true},
		},
	}, func(ctx context.Context, req *sdkmcp.GetPromptRequest) (*sdkmcp.GetPromptResult, error) {
		text := string(conductText)
		if rid := req.Params.Arguments["research_id"]; rid != "" {
			text = strings.ReplaceAll(text, "{research_id}", rid)
		}
		return &sdkmcp.GetPromptResult{
			Description: "Research execution workflow",
			Messages: []*sdkmcp.PromptMessage{
				{Role: "user", Content: &sdkmcp.TextContent{Text: text}},
			},
		}, nil
	})
}
