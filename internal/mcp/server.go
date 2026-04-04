package mcp

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

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
	log      *slog.Logger
}

func NewServer(
	research *service.ResearchService,
	section *service.SectionService,
	entry *service.EntryService,
	session *service.SessionService,
	task *service.TaskService,
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
		log:      log,
	}

	s.registerTools()
	s.registerPrompts()

	return s
}

func (s *Server) RunStdio(ctx context.Context) error {
	return s.server.Run(ctx, &sdkmcp.StdioTransport{})
}

func (s *Server) RunSSE(ctx context.Context, port int) error {
	handler := sdkmcp.NewSSEHandler(func(r *http.Request) *sdkmcp.Server {
		return s.server
	}, nil)
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
