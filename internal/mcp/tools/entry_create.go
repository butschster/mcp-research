package tools

import (
	"context"
	"log/slog"

	"github.com/butschster/mcp-research/internal/domain"
	"github.com/butschster/mcp-research/internal/service"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type EntryCreateInput struct {
	ResearchID  string   `json:"research_id" jsonschema:"ID of the research"`
	SectionID   string   `json:"section_id" jsonschema:"ID of the section"`
	SessionID   *string  `json:"session_id" jsonschema:"Optional session ID to link this entry to a session"`
	EntryType   *string  `json:"entry_type" jsonschema:"Optional content kind. markdown (default): content is markdown. blocks: content is a block document {version:1,blocks:[{type,data}]} — an article of typed blocks (paragraph, heading, list, table, quote, code, callout, divider, image, html), which is how you mix prose with alerts and custom visuals. artifact: sugar for a blocks document holding one html block; pass the HTML document as content. See /llms/blocks.md for the block catalog"`
	Content     string   `json:"content" jsonschema:"Markdown content of the entry"`
	Title       *string  `json:"title" jsonschema:"Optional title (auto-generated from content if empty)"`
	Description *string  `json:"description" jsonschema:"Optional description (auto-generated from content if empty)"`
	Status      *string  `json:"status" jsonschema:"Entry status: draft, active, completed, archived. Default: draft"`
	Tags        []string `json:"tags" jsonschema:"Tags for categorization"`
}

func RegisterEntryCreate(srv *mcp.Server, svc *service.EntryService, log *slog.Logger) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "entry_create",
		Description: "Creates a new entry within a section. Title and description are auto-generated from content if not provided. Returns entry_id and auto-generated fields.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input EntryCreateInput) (*mcp.CallToolResult, any, error) {
		var errs []string
		if input.ResearchID == "" {
			errs = append(errs, "research_id is required")
		}
		if input.SectionID == "" {
			errs = append(errs, "section_id is required")
		}
		if input.Content == "" {
			errs = append(errs, "content is required")
		}
		if len(errs) > 0 {
			return validationErrorResult(errs)
		}

		var status domain.EntryStatus
		if s := derefStr(input.Status); s != "" {
			status = domain.EntryStatus(s)
		}

		entry, err := svc.Create(ctx, service.CreateEntryRequest{
			ResearchID:  input.ResearchID,
			SectionID:   input.SectionID,
			SessionID:   derefStr(input.SessionID),
			Type:        domain.EntryType(derefStr(input.EntryType)),
			Content:     input.Content,
			Title:       derefStr(input.Title),
			Description: derefStr(input.Description),
			Status:      status,
			Tags:        input.Tags,
		})
		if err != nil {
			return errorResult(err.Error())
		}

		return successResult(map[string]any{
			"entry_id":    entry.ID,
			"code":        entry.Code,
			"title":       entry.Title,
			"description": entry.Description,
			"status":      entry.Status,
			"hint":        "Use [[" + entry.Code + "]] in other entries to cross-reference this entry.",
		})
	})
}
