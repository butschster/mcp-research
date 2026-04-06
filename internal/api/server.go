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
	AutoLoginToken string // JWT for default user auto-login (empty = disabled)
	MCPHandler     http.Handler // Streamable HTTP MCP handler (mounted at /mcp)
}

func NewServer(
	cfg ServerConfig,
	researchSvc *service.ResearchService,
	sectionSvc *service.SectionService,
	entrySvc *service.EntryService,
	sessionSvc *service.SessionService,
	taskSvc *service.TaskService,
	authSvc *service.AuthService, // nil when auth disabled
	db *sql.DB,
	entryRepo *storage.EntryRepository,
	researchRepo *storage.ResearchRepository,
	crossrefRepo *storage.CrossRefRepository,
	hub *ws.Hub,
	log *slog.Logger,
) *Server {
	mux := http.NewServeMux()

	rh := handlers.NewResearchHandler(researchSvc, sectionSvc, entrySvc, entryRepo, sessionSvc, log)
	eh := handlers.NewEntryHandler(entrySvc, entryRepo, researchRepo, log)
	sh := handlers.NewSessionHandler(sessionSvc, researchSvc, log)
	th := handlers.NewTaskHandler(taskSvc, researchSvc, log)

	// Build auth middleware functions
	var requireAuth func(http.Handler) http.Handler
	var optionalAuth func(http.Handler) http.Handler

	if cfg.AuthEnabled && authSvc != nil {
		validator := &serviceTokenValidator{authSvc: authSvc}
		requireAuth = auth.RequireAuth(validator)
		optionalAuth = auth.OptionalAuth(validator)
	}

	// wrap applies auth to endpoints:
	// - auth_enabled: user-based auth
	// - api_token set: legacy bearer token
	// - neither: no auth
	wrap := func(h http.HandlerFunc) http.Handler {
		if requireAuth != nil {
			return requireAuth(http.HandlerFunc(h))
		}
		if cfg.APIToken != "" {
			return bearerAuth(cfg.APIToken)(http.HandlerFunc(h))
		}
		return h
	}

	// wrapRead applies optional auth to read endpoints (user scoping when auth enabled)
	wrapRead := func(h http.HandlerFunc) http.Handler {
		if optionalAuth != nil {
			return requireAuth(http.HandlerFunc(h))
		}
		return h
	}

	// --- Auth endpoints (only when auth enabled) ---
	if cfg.AuthEnabled && authSvc != nil {
		ah := handlers.NewAuthHandler(authSvc, cfg.AutoLoginToken, log)
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
	mux.Handle("GET /api/entries/{id}", wrapRead(eh.Get))
	mux.Handle("GET /api/researches/{id}/entries/by-code/{code}", wrapRead(eh.ResolveCode))
	mux.Handle("GET /api/resolve/research/{code}", wrapRead(eh.ResolveResearchCode))
	crReadHandler := handlers.NewCrossRefHandler(crossrefRepo, entrySvc, researchSvc, log)
	mux.Handle("GET /api/researches/{id}/crossrefs", wrapRead(crReadHandler.ListForResearch))
	mux.Handle("GET /api/entries/{id}/crossrefs", wrapRead(crReadHandler.GetForEntry))
	mux.Handle("GET /api/entries/{id}/related", wrapRead(eh.GetRelated))
	mux.Handle("GET /api/search", wrapRead(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query().Get("q")
		if len(q) < 2 {
			writeJSON(w, http.StatusOK, map[string]any{"entries": []any{}, "researches": []any{}})
			return
		}
		entries, err := entryRepo.SearchEntries(r.Context(), q, 20)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"entries": entries})
	}))
	mux.Handle("GET /api/researches/{id}/tasks", wrapRead(th.ListByResearch))
	mux.Handle("GET /api/researches/{id}/sessions", wrapRead(sh.ListByResearch))
	mux.Handle("GET /api/sessions/{id}", wrapRead(sh.Get))

	// --- Write endpoints ---
	wh := handlers.NewWriteHandler(researchSvc, sectionSvc, entrySvc, sessionSvc, taskSvc, log)
	crh := handlers.NewCrossRefHandler(crossrefRepo, entrySvc, researchSvc, log)

	mux.Handle("POST /api/researches", wrap(wh.CreateResearch))
	mux.Handle("PUT /api/researches/{id}", wrap(wh.UpdateResearch))
	mux.Handle("POST /api/researches/{id}/sections", wrap(wh.AddSection))
	mux.Handle("PUT /api/sections/{sectionId}", wrap(wh.UpdateSection))
	mux.Handle("POST /api/entries", wrap(wh.CreateEntry))
	mux.Handle("PUT /api/entries/{id}", wrap(wh.UpdateEntry))
	mux.Handle("POST /api/tasks", wrap(wh.CreateTask))
	mux.Handle("PUT /api/tasks/{id}", wrap(wh.UpdateTask))
	mux.Handle("DELETE /api/tasks/{id}", wrap(wh.DeleteTask))
	mux.Handle("POST /api/sessions", wrap(wh.CreateSession))
	mux.Handle("PUT /api/questions/{questionId}", wrap(wh.UpdateQuestion))
	mux.Handle("POST /api/sessions/{id}/questions", wrap(wh.AddQuestions))
	mux.Handle("POST /api/researches/{id}/crossrefs/rebuild", wrap(crh.Rebuild))

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
	mux.HandleFunc("/ws", ws.HandleWebSocket(hub))

	// Health
	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"status":       "ok",
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
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

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
