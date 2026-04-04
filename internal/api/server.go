package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/butschster/mcp-research/internal/api/handlers"
	"github.com/butschster/mcp-research/internal/api/ws"
	"github.com/butschster/mcp-research/internal/service"
)

type Server struct {
	mux  *http.ServeMux
	hub  *ws.Hub
	port int
	log  *slog.Logger
}

func NewServer(
	port int,
	researchSvc *service.ResearchService,
	sectionSvc *service.SectionService,
	entrySvc *service.EntryService,
	sessionSvc *service.SessionService,
	taskSvc *service.TaskService,
	hub *ws.Hub,
	isInMemory bool,
	log *slog.Logger,
) *Server {
	mux := http.NewServeMux()

	rh := handlers.NewResearchHandler(researchSvc, sectionSvc, entrySvc, sessionSvc, log)
	eh := handlers.NewEntryHandler(entrySvc, log)
	sh := handlers.NewSessionHandler(sessionSvc, log)
	th := handlers.NewTaskHandler(taskSvc, log)

	// Research endpoints
	mux.HandleFunc("GET /api/researches", rh.List)
	mux.HandleFunc("GET /api/researches/{id}", rh.Get)
	mux.HandleFunc("GET /api/researches/{id}/sections/{sectionId}/entries", rh.ListSectionEntries)

	// Entry endpoints
	mux.HandleFunc("GET /api/entries/{id}", eh.Get)

	// Task endpoints
	mux.HandleFunc("GET /api/researches/{id}/tasks", th.ListByResearch)

	// Session endpoints
	mux.HandleFunc("GET /api/researches/{id}/sessions", sh.ListByResearch)
	mux.HandleFunc("GET /api/sessions/{id}", sh.Get)

	// WebSocket (no method prefix — upgrade needs raw handler)
	mux.HandleFunc("/ws", ws.HandleWebSocket(hub))

	// Health check
	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"status":    "ok",
			"in_memory": isInMemory,
		})
	})

	// Embedded frontend (catch-all, must be last)
	mux.Handle("/", staticHandler())

	return &Server{mux: mux, hub: hub, port: port, log: log}
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
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

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
