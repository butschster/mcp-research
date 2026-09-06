package tools

import (
	"context"
	"log/slog"
	"strings"

	"github.com/dovod-app/app/internal/service"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type ResearchExportInput struct {
	ResearchID string  `json:"research_id" jsonschema:"Research UUID or short code (e.g. R1)"`
	Format     *string `json:"format,omitempty" jsonschema:"portable (default) returns the research as JSON for research_import. obsidian returns a download link to a zip shaped like an Obsidian vault"`
}

// obsidianVault is what the tool returns for format=obsidian. A tool result
// cannot carry a zip, so it carries the link and a description of what is
// inside — enough for an agent to answer "give me this research for my notes
// app" with a link rather than an explanation.
type obsidianVault struct {
	Format      string `json:"format"`
	URL         string `json:"url"`
	Research    string `json:"research"`
	Contains    string `json:"contains"`
	Links       string `json:"links"`
	Auth        string `json:"auth,omitempty"`
	Description string `json:"description"`
}

func RegisterResearchExport(
	srv *mcp.Server,
	exportSvc *service.ExportService,
	researchSvc *service.ResearchService,
	baseURL func() string,
	log *slog.Logger,
) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "research_export",
		Description: "Export a complete research (sections, entries, sessions, questions, tasks, roadmaps). format=portable (default) returns JSON for transferring to another server via research_import. format=obsidian returns a link to download the research as an Obsidian vault: a folder per section, a markdown note per entry, and every [[E3]] cross-reference resolving as a link.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input ResearchExportInput) (*mcp.CallToolResult, any, error) {
		if input.ResearchID == "" {
			return validationErrorResult([]string{"research_id is required"})
		}

		format := "portable"
		if input.Format != nil {
			format = strings.ToLower(strings.TrimSpace(*input.Format))
		}

		switch format {
		case "", "portable", "json":
			data, err := exportSvc.Export(ctx, input.ResearchID)
			if err != nil {
				log.Error("research_export failed", "error", err)
				return errorResult(err.Error())
			}
			return successResult(data)

		case "obsidian", "vault", "zip":
			// Resolved here rather than by the download route so the agent is told
			// now that the research does not exist, instead of handing the user a
			// link that 404s.
			research, err := researchSvc.Get(ctx, input.ResearchID)
			if err != nil {
				log.Error("research_export failed", "error", err)
				return errorResult(err.Error())
			}
			code := research.Code
			if code == "" {
				code = research.ID
			}
			path := "/api/researches/" + code + "/export?format=obsidian"

			out := obsidianVault{
				Format:   "obsidian",
				URL:      strings.TrimRight(baseURL(), "/") + path,
				Research: research.Name,
				Contains: "README.md with a linked table of contents, a folder per section, a note per entry, plus Sessions, Tasks and Roadmaps folders and any HTML artifacts under _html/. The archive has no wrapper folder.",
				Links:    "Each [[E3]] is retargeted at the note it names and keeps its code as the display text, so links, backlinks and the graph view work in Obsidian.",
				Description: "Download the zip and unzip it into a folder inside an Obsidian vault. " +
					"Options: sessions, tasks, roadmaps, html (all on) and revisions (off) as query parameters.",
			}
			if baseURL() == "" {
				// A relative path is the honest answer when nothing configured a
				// public origin: inventing localhost would send the user nowhere.
				out.URL = path
				out.Description += " The URL is relative to this server; no public base URL is configured."
			}
			// A person, not an agent, will open this. Point them at the page they
			// are already signed in to rather than at a raw endpoint that answers
			// a browser with 401 JSON — and do not suggest putting a token in a
			// URL, which would land in logs and Referer headers.
			out.Auth = "This URL is fetched with credentials: send the API token or JWT as an Authorization: Bearer header. A person should instead open the research's Export page in the web UI, where the same archive is one click away."
			if base := strings.TrimRight(baseURL(), "/"); base != "" {
				out.Auth += " Page: " + base + "/research/" + code + "/export"
			}
			return successResult(out)
		}

		return validationErrorResult([]string{"format must be portable or obsidian"})
	})
}
