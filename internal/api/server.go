package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/butschster/mcp-research/internal/api/handlers"
	"github.com/butschster/mcp-research/internal/api/ws"
	"github.com/butschster/mcp-research/internal/auth"
	"github.com/butschster/mcp-research/internal/domain"
	"github.com/butschster/mcp-research/internal/service"
	"github.com/butschster/mcp-research/internal/storage"
)

type Server struct {
	mux  *http.ServeMux
	hub  *ws.Hub
	port int
	log  *slog.Logger
}

type ServerConfig struct {
	Port           int
	IsInMemory     bool
	APIToken       string
	AuthEnabled    bool
	BaseURL        string // Public base URL for OAuth metadata (e.g. https://mcp.example.com)
	OAuthSvc       *service.OAuthService
	AutoLoginToken string       // JWT for default user auto-login (empty = disabled)
	MCPHandler     http.Handler // Streamable HTTP MCP handler (mounted at /mcp)
	// Version is what the binary was built as. It reaches the web UI through
	// /api/health, which is the one endpoint every page can already call
	// without a credential.
	Version string
}

func NewServer(
	cfg ServerConfig,
	researchSvc *service.ResearchService,
	sectionSvc *service.SectionService,
	entrySvc *service.EntryService,
	sessionSvc *service.SessionService,
	taskSvc *service.TaskService,
	roadmapSvc *service.RoadmapService,
	exportSvc *service.ExportService,
	obsidianSvc *service.ObsidianService,
	teamSvc *service.TeamService,
	shareSvc *service.ShareService,
	skillSvc *service.SkillService,
	templateSvc *service.TemplateService,
	access *service.Access,
	authSvc *service.AuthService, // nil when auth disabled
	db *sql.DB,
	entryRepo *storage.EntryRepository,
	researchRepo *storage.ResearchRepository,
	crossrefRepo *storage.CrossRefRepository,
	externalLinkRepo *storage.ExternalLinkRepository,
	hub *ws.Hub,
	log *slog.Logger,
) *Server {
	mux := http.NewServeMux()

	rh := handlers.NewResearchHandler(researchSvc, sectionSvc, entrySvc, entryRepo, sessionSvc, log)
	eh := handlers.NewEntryHandler(entrySvc, researchSvc, entryRepo, researchRepo, log)
	sh := handlers.NewSessionHandler(sessionSvc, entrySvc, researchSvc, log)
	th := handlers.NewTaskHandler(taskSvc, researchSvc, log)
	rmh := handlers.NewRoadmapHandler(roadmapSvc, researchSvc, log)
	tmh := handlers.NewTeamHandler(teamSvc, researchSvc, log)
	shh := handlers.NewShareHandler(shareSvc, researchSvc, sectionSvc, log)
	skh := handlers.NewSkillHandler(skillSvc, researchSvc, log)
	tph := handlers.NewTemplateHandler(templateSvc, researchSvc, sectionSvc, skillSvc, log)
	rh.SetShareService(shareSvc)

	// Build auth middleware functions
	var requireAuth func(http.Handler) http.Handler
	var optionalAuth func(http.Handler) http.Handler

	if cfg.AuthEnabled && authSvc != nil {
		validator := &serviceTokenValidator{authSvc: authSvc}
		requireAuth = auth.RequireAuth(validator)
		optionalAuth = auth.OptionalAuth(validator)
	}

	// markAuthor records who is writing, for the revision history an entry write
	// leaves behind. The credential is the evidence: a JWT is a browser session,
	// so a person is typing; an API key, an OAuth token or the legacy write token
	// is a machine. With no credential at all this is a local run with auth off,
	// where the only thing on this port is the web UI.
	//
	// MCP needs no equivalent — service.AuthorFromContext defaults to agent,
	// which is what every MCP write is.
	markAuthor := func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			kind := domain.AuthorHuman
			if token := extractBearerToken(r); token != "" {
				if authSvc == nil || !authSvc.IsSessionToken(token) {
					kind = domain.AuthorAgent
				}
			}
			ctx := service.WithAuthor(r.Context(), kind)
			// Which tab is writing, so the event this produces can be
			// recognised by the tab that caused it and ignored there. Opaque to
			// the server; absent for every non-browser caller.
			ctx = service.WithClientID(ctx, r.Header.Get("X-Client-Id"))
			h.ServeHTTP(w, r.WithContext(ctx))
		})
	}

	// wrap applies auth to endpoints:
	// - auth_enabled: user-based auth
	// - api_token set: legacy bearer token
	// - neither: no auth
	wrap := func(h http.HandlerFunc) http.Handler {
		if requireAuth != nil {
			return markAuthor(requireAuth(http.HandlerFunc(h)))
		}
		if cfg.APIToken != "" {
			return markAuthor(bearerAuth(cfg.APIToken)(http.HandlerFunc(h)))
		}
		return markAuthor(h)
	}

	// wrapRead applies optional auth to read endpoints (user scoping when auth enabled)
	wrapRead := func(h http.HandlerFunc) http.Handler {
		if requireAuth != nil {
			return requireAuth(http.HandlerFunc(h))
		}
		return h
	}

	// wrapOperator serves the routes that can carry either kind of caller: a
	// person editing their team's template, or whoever runs this server writing
	// one every team will see.
	//
	// It has to be its own wrapper because the two credentials are checked by
	// different code. `wrap` sends everything through requireAuth once
	// auth_enabled is on, and requireAuth rejects the api_token — it is not a
	// JWT and not an API key, so the operator would get a 401 on a route that
	// exists for them. Here the api_token is recognised first and skips the user
	// check entirely, which is correct: the operator is not a user, has no team
	// and is in nobody's member list.
	//
	// It authorises nothing on its own. All it does is stamp the context; the
	// refusals live in TemplateService, so a second entry point cannot arrive
	// without one.
	asOperator := func(kind auth.OperatorKind, h http.HandlerFunc) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h(w, r.WithContext(auth.WithOperator(r.Context(), kind)))
		})
	}
	wrapOperator := func(h http.HandlerFunc) http.Handler {
		byToken := markAuthor(asOperator(auth.OperatorByToken, h))
		if requireAuth != nil {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if operatorCredential(cfg.APIToken, r) {
					byToken.ServeHTTP(w, r)
					return
				}
				// Not the operator: an ordinary authenticated caller, who may
				// still edit their own team's templates.
				markAuthor(requireAuth(http.HandlerFunc(h))).ServeHTTP(w, r)
			})
		}
		if cfg.APIToken != "" {
			return bearerAuth(cfg.APIToken)(byToken)
		}
		// Neither a token nor accounts: a local run, where every write is
		// already unauthenticated and there is no boundary for the operator to
		// prove themselves across. Refusing here would only mean the feature
		// cannot be tried without first inventing a credential. A different
		// kind, though — the token narrows what its holder can see, and doing
		// that here would hide a local user's own team templates from them.
		return markAuthor(asOperator(auth.OperatorNoBoundary, h))
	}

	// wrapReadOperator is the read side of the same story. Without it the
	// operator could write a global template and never read one back:
	// `wrapRead` sends everything through requireAuth, which rejects the
	// api_token — so the only id they would ever hold is the one their own
	// create call returned, and `PUT`/`DELETE` need an id.
	//
	// What they see is narrowed to the global tier by TemplateService, because
	// the token proves who runs the server and not membership of any team.
	wrapReadOperator := func(h http.HandlerFunc) http.Handler {
		if requireAuth == nil {
			return wrapRead(h)
		}
		byToken := asOperator(auth.OperatorByToken, h)
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if operatorCredential(cfg.APIToken, r) {
				byToken.ServeHTTP(w, r)
				return
			}
			requireAuth(http.HandlerFunc(h)).ServeHTTP(w, r)
		})
	}

	// wrapOptional attaches a session when the request carries one and lets it
	// through when it does not. Only the invitation preview uses it: someone
	// following a link has no account yet and still has to be told what they
	// are being invited to — but if they are signed in, the answer says more.
	wrapOptional := func(h http.HandlerFunc) http.Handler {
		if optionalAuth != nil {
			return optionalAuth(http.HandlerFunc(h))
		}
		return h
	}

	// --- Auth endpoints (only when auth enabled) ---
	if cfg.AuthEnabled && authSvc != nil {
		ah := handlers.NewAuthHandler(authSvc, teamSvc, cfg.AutoLoginToken, log)
		mux.HandleFunc("POST /api/auth/register", ah.Register)
		mux.HandleFunc("POST /api/auth/login", ah.Login)
		mux.Handle("GET /api/auth/me", requireAuth(http.HandlerFunc(ah.Me)))
		mux.Handle("POST /api/auth/api-keys", requireAuth(http.HandlerFunc(ah.CreateAPIKey)))
		mux.Handle("GET /api/auth/api-keys", requireAuth(http.HandlerFunc(ah.ListAPIKeys)))
		mux.Handle("DELETE /api/auth/api-keys/{id}", requireAuth(http.HandlerFunc(ah.DeleteAPIKey)))
		mux.HandleFunc("GET /api/auth/info", ah.AuthInfo)
	}

	// --- OAuth2 endpoints (only when auth enabled) ---
	if cfg.AuthEnabled && cfg.OAuthSvc != nil && authSvc != nil {
		oh := handlers.NewOAuthHandler(cfg.OAuthSvc, authSvc, log)

		// Standard OAuth2 paths (used by ChatGPT and other MCP clients)
		mux.HandleFunc("GET /auth/authorize", oh.Authorize)
		mux.HandleFunc("POST /auth/authorize", oh.Authorize)
		mux.HandleFunc("POST /auth/token", oh.Token)

		// RFC 7591 Dynamic Client Registration
		mux.HandleFunc("POST /auth/register", oh.RegisterClient)

		// OAuth2 Authorization Server Metadata (RFC 8414)
		baseURL := fmt.Sprintf("http://localhost:%d", cfg.Port)
		if cfg.BaseURL != "" {
			baseURL = cfg.BaseURL
		}
		mux.HandleFunc("GET /.well-known/oauth-authorization-server", handlers.OAuthMetadataHandler(baseURL))
		mux.HandleFunc("GET /.well-known/openid-configuration", handlers.OAuthMetadataHandler(baseURL))

		// OAuth2 Protected Resource Metadata (RFC 9728)
		mux.HandleFunc("GET /.well-known/oauth-protected-resource", handlers.OAuthProtectedResourceHandler(baseURL))
	}

	// --- Read endpoints ---
	mux.Handle("GET /api/researches", wrapRead(rh.List))
	mux.Handle("GET /api/researches/{id}", wrapRead(rh.Get))
	mux.Handle("GET /api/researches/{id}/sections/{sectionId}/entries", wrapRead(rh.ListSectionEntries))
	mux.Handle("GET /api/researches/{id}/entries", wrapRead(rh.ListAllEntries))
	mux.Handle("GET /api/researches/{id}/tags", wrapRead(rh.ListTags))
	exportHandler := handlers.NewExportHandler(researchSvc, sectionSvc, entrySvc, entryRepo, sessionSvc, taskSvc, log)
	exportHandler.SetExportService(exportSvc)
	exportHandler.SetObsidianService(obsidianSvc)
	exportHandler.SetRoadmapService(roadmapSvc)
	mux.Handle("GET /api/researches/{id}/export", wrapRead(exportHandler.Export))
	mux.Handle("GET /api/researches/{id}/sessions/{sessionId}/export", wrapRead(exportHandler.ExportSession))
	importHandler := handlers.NewImportHandler(exportSvc, log)
	mux.Handle("GET /api/researches/{id}/export/portable", wrapRead(exportHandler.ExportPortable))
	mux.Handle("POST /api/researches/import", wrap(importHandler.Import))
	mux.Handle("GET /api/entries/{id}", wrapRead(eh.Get))
	mux.Handle("GET /api/researches/{id}/entries/by-code/{code}", wrapRead(eh.ResolveCode))
	mux.Handle("GET /api/resolve/research/{code}", wrapRead(eh.ResolveResearchCode))
	crReadHandler := handlers.NewCrossRefHandler(crossrefRepo, entrySvc, researchSvc, access, log)
	crReadHandler.SetRoadmapService(roadmapSvc)
	mux.Handle("GET /api/researches/{id}/crossrefs", wrapRead(crReadHandler.ListForResearch))
	mux.Handle("GET /api/entries/{id}/crossrefs", wrapRead(crReadHandler.GetForEntry))
	elHandler := handlers.NewExternalLinkHandler(externalLinkRepo, researchSvc, entrySvc, log)
	mux.Handle("GET /api/researches/{id}/links", wrapRead(elHandler.ListByResearch))
	mux.Handle("GET /api/entries/{id}/links", wrapRead(elHandler.ListByEntry))
	mux.Handle("GET /api/entries/{id}/related", wrapRead(eh.GetRelated))
	revh := handlers.NewRevisionHandler(entrySvc, sessionSvc, researchSvc, log)
	mux.Handle("GET /api/entries/{id}/revisions", wrapRead(revh.List))
	mux.Handle("GET /api/entries/{id}/revisions/{revision}", wrapRead(revh.Get))
	mux.Handle("GET /api/entries/{id}/diff", wrapRead(revh.Diff))
	mux.Handle("POST /api/entries/{id}/revisions/{revision}/restore", wrap(revh.Restore))
	mux.Handle("GET /api/sessions/{id}/changes", wrapRead(revh.SessionChanges))
	mux.Handle("GET /api/search", wrapRead(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query().Get("q")
		if len(q) < 2 {
			writeJSON(w, http.StatusOK, map[string]any{"entries": []any{}, "researches": []any{}})
			return
		}
		// `research` scopes the search to one research, which is the only way
		// to find anything in a large one: tags are the sole in-research filter
		// and they exist only if the agent applied them.
		entries, err := entryRepo.SearchEntries(r.Context(), q, 20, auth.UserIDFromContext(r.Context()), r.URL.Query().Get("research"))
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
			return
		}
		// A nil slice encodes as `null`, and a client that reads `.length` off
		// it crashes on the one case that is most likely: no matches.
		if entries == nil {
			entries = []*domain.Entry{}
		}
		writeJSON(w, http.StatusOK, map[string]any{"entries": entries})
	}))
	mux.Handle("GET /api/researches/{id}/tasks", wrapRead(th.ListByResearch))
	mux.Handle("GET /api/researches/{id}/sessions", wrapRead(sh.ListByResearch))
	mux.Handle("GET /api/researches/{id}/sessions/{sessionId}", wrapRead(sh.Get))
	mux.Handle("GET /api/researches/{id}/entries/{entryId}", wrapRead(eh.GetByResearch))
	mux.Handle("GET /api/researches/{id}/roadmaps", wrapRead(rmh.ListByResearch))
	mux.Handle("GET /api/researches/{id}/roadmaps/{roadmapId}", wrapRead(rmh.GetByResearch))
	mux.Handle("GET /api/roadmaps/{id}", wrapRead(rmh.Get))

	// --- Graph endpoint ---
	gh := handlers.NewGraphHandler(researchSvc, sectionSvc, entrySvc, sessionSvc, taskSvc, entryRepo, crossrefRepo, access, log)
	mux.Handle("GET /api/researches/{id}/graph", wrapRead(gh.Get))

	// --- Write endpoints ---
	wh := handlers.NewWriteHandler(researchSvc, sectionSvc, entrySvc, sessionSvc, taskSvc, log)
	crh := handlers.NewCrossRefHandler(crossrefRepo, entrySvc, researchSvc, access, log)
	crh.SetRoadmapService(roadmapSvc)

	mux.Handle("POST /api/researches", wrap(wh.CreateResearch))
	mux.Handle("PUT /api/researches/{id}", wrap(wh.UpdateResearch))
	mux.Handle("POST /api/researches/{id}/sections", wrap(wh.AddSection))
	mux.Handle("PUT /api/sections/{sectionId}", wrap(wh.UpdateSection))
	mux.Handle("POST /api/entries", wrap(wh.CreateEntry))
	mux.Handle("PUT /api/entries/{id}", wrap(wh.UpdateEntry))
	mux.Handle("POST /api/entries/{id}/patch", wrap(wh.PatchEntry))
	mux.Handle("DELETE /api/entries/{id}", wrap(wh.DeleteEntry))
	mux.Handle("POST /api/tasks", wrap(wh.CreateTask))
	mux.Handle("PUT /api/tasks/{id}", wrap(wh.UpdateTask))
	mux.Handle("DELETE /api/tasks/{id}", wrap(wh.DeleteTask))
	mux.Handle("POST /api/sessions", wrap(wh.CreateSession))
	mux.Handle("PUT /api/sessions/{id}", wrap(wh.UpdateSession))
	mux.Handle("PUT /api/questions/{questionId}", wrap(wh.UpdateQuestion))
	mux.Handle("POST /api/sessions/{id}/questions", wrap(wh.AddQuestions))
	mux.Handle("POST /api/roadmaps", wrap(rmh.Create))
	mux.Handle("PUT /api/roadmaps/{id}", wrap(rmh.Update))
	mux.Handle("DELETE /api/roadmaps/{id}", wrap(rmh.Delete))
	mux.Handle("PUT /api/roadmap-nodes/{nodeId}", wrap(rmh.UpdateNode))
	mux.Handle("POST /api/researches/{id}/crossrefs/rebuild", wrap(crh.Rebuild))

	// --- Teams ---
	//
	// Every route here is scoped by membership inside TeamService, so wrapRead
	// and wrap do the same job they do elsewhere: put the caller in the
	// context. The one exception is the invite preview, which is reachable
	// without a session on purpose.
	mux.Handle("GET /api/teams", wrapRead(tmh.List))
	mux.Handle("GET /api/teams/{id}", wrapRead(tmh.Get))
	mux.Handle("POST /api/teams", wrap(tmh.Create))
	mux.Handle("PUT /api/teams/{id}", wrap(tmh.Update))
	mux.Handle("DELETE /api/teams/{id}", wrap(tmh.Delete))
	mux.Handle("GET /api/teams/{id}/members", wrapRead(tmh.Members))
	mux.Handle("PUT /api/teams/{id}/members/{userId}", wrap(tmh.UpdateMember))
	mux.Handle("DELETE /api/teams/{id}/members/{userId}", wrap(tmh.RemoveMember))
	mux.Handle("GET /api/teams/{id}/invites", wrapRead(tmh.Invites))
	mux.Handle("POST /api/teams/{id}/invites", wrap(tmh.CreateInvite))
	mux.Handle("DELETE /api/invites/{id}", wrap(tmh.RevokeInvite))
	mux.Handle("GET /api/invites/{token}", wrapOptional(tmh.PreviewInvite))
	mux.Handle("POST /api/invites/{token}/accept", wrap(tmh.AcceptInvite))
	mux.Handle("POST /api/researches/{id}/transfer", wrap(tmh.TransferResearch))
	mux.Handle("POST /api/teams/{id}/researches", wrap(tmh.AddResearches))

	// --- Skills ---
	//
	// A slug addresses a skill only inside a research, where the resolution
	// order is defined (private, then team, then built-in). Management is by id,
	// because a built-in and a team's fork of it share a slug on purpose.
	//
	// None of these are mounted on the shared sub-mux. A skill is a team's
	// methodology — the same class of working process as the instruction it
	// replaces, which redactForShare has always stripped — so a share link must
	// not reach one. SkillService.Load refuses a share context as well, because
	// a future route that forgot this should still fail closed.
	mux.Handle("GET /api/researches/{id}/skills", wrapRead(skh.ListAttached))
	mux.Handle("GET /api/researches/{id}/skills/library", wrapRead(skh.Library))
	mux.Handle("GET /api/researches/{id}/skills/{slug}", wrapRead(skh.Read))
	mux.Handle("POST /api/researches/{id}/skills", wrap(skh.Create))
	mux.Handle("DELETE /api/researches/{id}/skills/{slug}", wrap(skh.Detach))
	mux.Handle("PUT /api/researches/{id}/skills/{slug}", wrap(skh.Fork))
	mux.Handle("POST /api/researches/{id}/skills/{slug}/copy", wrap(skh.CopyHere))
	mux.Handle("POST /api/researches/{id}/skills/{slug}/promote", wrap(skh.Promote))
	mux.Handle("GET /api/teams/{id}/skills", wrapRead(skh.ListTeam))
	mux.Handle("POST /api/teams/{id}/skills", wrap(skh.CreateTeam))
	mux.Handle("GET /api/skills/{skillId}", wrapRead(skh.ReadByID))
	mux.Handle("PUT /api/skills/{skillId}", wrap(skh.Update))
	mux.Handle("DELETE /api/skills/{skillId}", wrap(skh.Delete))

	// --- Templates ---
	//
	// A template is a kickoff methodology, so it belongs to nobody or to a team
	// — never to a research. That is why none of these hang off
	// /api/researches/{id}: the one that does, `templates/draft`, reads a
	// research to *propose* a template and creates nothing.
	//
	// Not on the shared sub-mux, for the same reason skills are not: a
	// template body is a team's working process.
	mux.Handle("GET /api/templates", wrapReadOperator(tph.List))
	mux.Handle("GET /api/templates/{slug}", wrapReadOperator(tph.Get))
	// Read by slug, write by id — the same split skills use, and for the same
	// reason: a team's fork keeps its parent's slug, so a slug is an address
	// only within one caller's scope. Fork is a POST because `PUT {slug}/fork`
	// and `PUT {templateId}` are ambiguous to the router and neither is more
	// specific, which it refuses to route at all.
	// The server-wide library. `wrapOperator` on all three because a global
	// template and a team's live at the same routes and are told apart by the
	// row, not by the path — see the wrapper for why the api_token cannot go
	// through `wrap`.
	mux.Handle("POST /api/templates", wrapOperator(tph.CreateGlobal))
	mux.Handle("POST /api/templates/{slug}/fork", wrap(tph.Fork))
	mux.Handle("PUT /api/templates/{templateId}", wrapOperator(tph.Update))
	mux.Handle("DELETE /api/templates/{templateId}", wrapOperator(tph.Delete))
	mux.Handle("GET /api/teams/{id}/templates", wrapRead(tph.ListTeam))
	mux.Handle("POST /api/teams/{id}/templates", wrap(tph.CreateTeam))
	mux.Handle("GET /api/researches/{id}/templates/draft", wrapRead(tph.DraftFromResearch))

	// --- Share links ---
	//
	// Both halves live in share_routes.go: the owner's management routes here on
	// the authenticated API, and the visitor's read surface behind its own
	// prefix and its own middleware. Keeping the public prefix in one file is
	// the point — it is the only place in the product where data leaves the
	// owner boundary, and it should be readable in one sitting.
	registerShareRoutes(mux, shareDeps{
		shares:   shareSvc,
		research: rh,
		entry:    eh,
		session:  sh,
		task:     th,
		roadmap:  rmh,
		crossref: crReadHandler,
		links:    elHandler,
		export:   exportHandler,
		share:    shh,
	}, wrap, wrapRead)

	// Backfill short codes for all records missing them
	mux.Handle("POST /api/admin/backfill-codes", wrap(func(w http.ResponseWriter, r *http.Request) {
		count, err := storage.BackfillCodes(r.Context(), db)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"backfilled": count, "status": "ok"})
	}))

	if cfg.AuthEnabled {
		log.Info("auth: multi-user authentication enabled")
	} else if cfg.APIToken != "" {
		log.Info("write API: bearer token required")
	} else {
		log.Info("write API: no authentication (api_token not set)")
	}

	// WebSocket
	// The hub decides delivery per event, so this only has to establish who is
	// on the other end. With auth off it takes anyone, as it always has — but
	// the origin is checked either way, because "auth off" is a single-user
	// local run, not an invitation for any page in the browser to listen in.
	if authSvc != nil {
		hub.SetTokenValidator(authSvc)
	}
	mux.HandleFunc("/ws", ws.HandleWebSocket(hub, cfg.BaseURL))

	// Health
	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"status":       "ok",
			"version":      cfg.Version,
			"in_memory":    cfg.IsInMemory,
			"write_api":    cfg.APIToken != "" || cfg.AuthEnabled,
			"auth_enabled": cfg.AuthEnabled,
		})
	})

	// OpenAPI spec (auto-generated)
	mux.HandleFunc("GET /api/openapi.yaml", handleOpenAPI(cfg.APIToken != "" || cfg.AuthEnabled))

	// LLMs documentation
	llmsHandler := handleLLMSDocs()
	mux.Handle("GET /llms.txt", llmsHandler)
	mux.Handle("GET /llms/", http.StripPrefix("/", llmsHandler))

	// MCP Streamable HTTP transport (used by ChatGPT, Claude.ai)
	if cfg.MCPHandler != nil {
		var mcpEndpoint http.Handler = cfg.MCPHandler
		if requireAuth != nil {
			mcpEndpoint = requireAuth(cfg.MCPHandler)
		}
		mux.Handle("/mcp", mcpEndpoint)
		log.Info("MCP Streamable HTTP endpoint registered at /mcp")

		// Catch-all: serve MCP for POST/DELETE with MCP headers, static frontend for everything else
		static := staticHandler()
		mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Route MCP requests (POST/DELETE with JSON or MCP session header) to MCP handler
			if (r.Method == http.MethodPost || r.Method == http.MethodDelete) &&
				(r.Header.Get("Content-Type") == "application/json" || r.Header.Get("Mcp-Session-Id") != "") {
				mcpEndpoint.ServeHTTP(w, r)
				return
			}
			// GET with Accept: text/event-stream or MCP session → also MCP
			if r.Method == http.MethodGet && (r.Header.Get("Accept") == "text/event-stream" || r.Header.Get("Mcp-Session-Id") != "") {
				mcpEndpoint.ServeHTTP(w, r)
				return
			}
			static.ServeHTTP(w, r)
		}))
	} else {
		// Embedded frontend (catch-all, must be last)
		mux.Handle("/", staticHandler())
	}

	return &Server{mux: mux, hub: hub, port: cfg.Port, log: log}
}

// serviceTokenValidator adapts AuthService for the auth middleware.
type serviceTokenValidator struct {
	authSvc *service.AuthService
}

func (v *serviceTokenValidator) ValidateToken(r *http.Request) (*domain.User, error) {
	token := extractBearerToken(r)
	if token == "" {
		return nil, nil
	}
	return v.authSvc.ValidateToken(r.Context(), token)
}

func extractBearerToken(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		return authHeader[7:]
	}
	// Also check query param (for SSE connections)
	if t := r.URL.Query().Get("token"); t != "" {
		return t
	}
	return ""
}

// operatorCredential answers whether this request carries the instance
// api_token, and reads the **header only**.
//
// Deliberately not extractBearerToken, which also accepts `?token=` for SSE. The
// api_token is the longest-lived and highest-privilege secret on the instance —
// it comes out of the config file and nothing rotates it — and in a query string
// it lands verbatim in the nginx access log (`deploy/nginx/mcp-research.conf`
// logs `$request`), in browser history on the GET routes, and in the `Referer`
// of any outbound link. `bearerAuth` has always read the header only; these two
// wrappers are the same credential and must not be laxer than it.
func operatorCredential(configured string, r *http.Request) bool {
	if configured == "" {
		return false
	}
	a := r.Header.Get("Authorization")
	return len(a) > 7 && a[:7] == "Bearer " && a[7:] == configured
}

// bearerAuth returns middleware that validates Authorization: Bearer <token>.
func bearerAuth(token string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			a := r.Header.Get("Authorization")
			if a == "" || len(a) < 8 || a[:7] != "Bearer " || a[7:] != token {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "invalid or missing bearer token"})
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func (s *Server) Start(ctx context.Context) error {
	handler := corsMiddleware(s.mux)
	addr := fmt.Sprintf(":%d", s.port)
	s.log.Info("API server listening", "addr", addr)

	srv := &http.Server{Addr: addr, Handler: handler}
	go func() {
		<-ctx.Done()
		srv.Close()
	}()

	return srv.ListenAndServe()
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip CORS for WebSocket upgrade
		if r.Header.Get("Upgrade") == "websocket" {
			next.ServeHTTP(w, r)
			return
		}

		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		// X-Client-Id is not a simple header, so its presence alone makes every
		// request preflight — the GETs included. Omitting it here blocks the whole
		// cross-origin dev setup, not merely the writes.
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Client-Id")
		// Without this a cross-origin download cannot read the filename the
		// server chose — which is every request under `make frontend-dev`, where
		// the SPA is served from :3000 and the API from :8088.
		w.Header().Set("Access-Control-Expose-Headers", "Content-Disposition")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
