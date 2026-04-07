package mcp

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/butschster/mcp-research/internal/auth"
	"github.com/butschster/mcp-research/internal/domain"
	"github.com/butschster/mcp-research/internal/service"
	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

type Server struct {
	server   *sdkmcp.Server
	research *service.ResearchService
	section  *service.SectionService
	entry    *service.EntryService
	session  *service.SessionService
	task     *service.TaskService
	roadmap  *service.RoadmapService
	log      *slog.Logger
}

func NewServer(
	research *service.ResearchService,
	section *service.SectionService,
	entry *service.EntryService,
	session *service.SessionService,
	task *service.TaskService,
	roadmap *service.RoadmapService,
	log *slog.Logger,
	version string,
) *Server {
	s := &Server{
		server: sdkmcp.NewServer(&sdkmcp.Implementation{
			Name:    "mcp-research",
			Version: version,
		}, &sdkmcp.ServerOptions{
			Logger:       log,
			Instructions: "MCP Research Server — AI-driven structured research sessions. Use research/initialize prompt to start. Use task_create/task_list to manage your todo list.",
		}),
		research: research,
		section:  section,
		entry:    entry,
		session:  session,
		task:     task,
		roadmap:  roadmap,
		log:      log,
	}

	s.registerTools()
	s.registerPrompts()

	return s
}

// RunStdio runs the MCP server over stdin/stdout.
// If defaultUser is provided, all operations are scoped to that user.
func (s *Server) RunStdio(ctx context.Context, defaultUser *domain.User) error {
	if defaultUser != nil {
		ctx = auth.WithUser(ctx, defaultUser)
		s.log.Info("stdio: running as user", "email", defaultUser.Email, "id", defaultUser.ID)
	}
	return s.server.Run(ctx, &sdkmcp.StdioTransport{})
}

func (s *Server) RunSSE(ctx context.Context, port int, authSvc *service.AuthService, baseURL string) error {
	sseHandler := sdkmcp.NewSSEHandler(func(r *http.Request) *sdkmcp.Server {
		return s.server
	}, nil)

	var handler http.Handler = sseHandler

	// Wrap with auth middleware when auth service is available
	if authSvc != nil {
		handler = sseAuthMiddleware(authSvc, baseURL, sseHandler)
	}

	addr := fmt.Sprintf(":%d", port)
	s.log.Info("MCP SSE server listening", "addr", addr)

	srv := &http.Server{Addr: addr, Handler: handler}
	go func() {
		<-ctx.Done()
		srv.Close()
	}()

	return srv.ListenAndServe()
}

func (s *Server) MCPServer() *sdkmcp.Server {
	return s.server
}

// StreamableHTTPHandler returns an http.Handler for the Streamable HTTP transport.
// Mount this at /mcp on the API server.
func (s *Server) StreamableHTTPHandler() http.Handler {
	return sdkmcp.NewStreamableHTTPHandler(func(r *http.Request) *sdkmcp.Server {
		return s.server
	}, nil)
}

// sseAuthMiddleware extracts bearer token from SSE requests and injects user into context.
// Returns WWW-Authenticate header on 401 to enable OAuth discovery per MCP spec.
func sseAuthMiddleware(authSvc *service.AuthService, baseURL string, next http.Handler) http.Handler {
	resourceMetadataURL := baseURL + "/.well-known/oauth-protected-resource"

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		slog.Info("sse request", "method", r.Method, "path", r.URL.Path, "remote", r.RemoteAddr, "has_auth", r.Header.Get("Authorization") != "")

		token := ""
		if authHeader := r.Header.Get("Authorization"); len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			token = authHeader[7:]
		}
		if token == "" {
			token = r.URL.Query().Get("token")
		}

		if token == "" {
			slog.Info("sse auth: no token", "path", r.URL.Path)
			w.Header().Set("WWW-Authenticate", fmt.Sprintf(`Bearer resource_metadata="%s"`, resourceMetadataURL))
			http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
			return
		}

		user, err := authSvc.ValidateToken(r.Context(), token)
		if err != nil || user == nil {
			slog.Info("sse auth: invalid token", "path", r.URL.Path, "error", err)
			w.Header().Set("WWW-Authenticate", fmt.Sprintf(`Bearer resource_metadata="%s"`, resourceMetadataURL))
			http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
			return
		}

		ctx := auth.WithUser(r.Context(), user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
