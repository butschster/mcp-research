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
	"github.com/butschster/mcp-research/internal/storage"
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
	entryRepo *storage.EntryRepository,
	researchRepo *storage.ResearchRepository,
	crossrefRepo *storage.CrossRefRepository,
	hub *ws.Hub,
	isInMemory bool,
	apiToken string,
	log *slog.Logger,
) *Server {
	mux := http.NewServeMux()

	rh := handlers.NewResearchHandler(researchSvc, sectionSvc, entrySvc, sessionSvc, log)
	eh := handlers.NewEntryHandler(entrySvc, entryRepo, researchRepo, log)
	sh := handlers.NewSessionHandler(sessionSvc, log)
	th := handlers.NewTaskHandler(taskSvc, log)

	// --- Read-only endpoints (no auth) ---

	mux.HandleFunc("GET /api/researches", rh.List)
	mux.HandleFunc("GET /api/researches/{id}", rh.Get)
	mux.HandleFunc("GET /api/researches/{id}/sections/{sectionId}/entries", rh.ListSectionEntries)
	mux.HandleFunc("GET /api/entries/{id}", eh.Get)
	mux.HandleFunc("GET /api/researches/{id}/entries/by-code/{code}", eh.ResolveCode)
	mux.HandleFunc("GET /api/resolve/research/{code}", eh.ResolveResearchCode)
	mux.HandleFunc("GET /api/researches/{id}/crossrefs", handlers.NewCrossRefHandler(crossrefRepo, entrySvc, log).ListForResearch)
	mux.HandleFunc("GET /api/researches/{id}/tasks", th.ListByResearch)
	mux.HandleFunc("GET /api/researches/{id}/sessions", sh.ListByResearch)
	mux.HandleFunc("GET /api/sessions/{id}", sh.Get)

	// --- Write endpoints (auth only when api_token is configured) ---

	wh := handlers.NewWriteHandler(researchSvc, sectionSvc, entrySvc, sessionSvc, taskSvc, log)
	crh := handlers.NewCrossRefHandler(crossrefRepo, entrySvc, log)

	wrap := func(h http.HandlerFunc) http.Handler {
		if apiToken != "" {
			return bearerAuth(apiToken)(http.HandlerFunc(h))
		}
		return h
	}

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
	mux.Handle("POST /api/researches/{id}/crossrefs/rebuild", wrap(crh.Rebuild))

	if apiToken != "" {
		log.Info("write API: bearer token required")
	} else {
		log.Info("write API: no authentication (api_token not set)")
	}

	// WebSocket
	mux.HandleFunc("/ws", ws.HandleWebSocket(hub))

	// Health
	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"status":    "ok",
			"in_memory": isInMemory,
			"write_api": apiToken != "",
		})
	})

	// OpenAPI spec (auto-generated)
	mux.HandleFunc("GET /api/openapi.yaml", handleOpenAPI(apiToken != ""))

	// LLMs documentation
	llmsHandler := handleLLMSDocs()
	mux.Handle("GET /llms.txt", llmsHandler)
	mux.Handle("GET /llms/", http.StripPrefix("/", llmsHandler))

	// Embedded frontend (catch-all, must be last)
	mux.Handle("/", staticHandler())

	return &Server{mux: mux, hub: hub, port: port, log: log}
}

// bearerAuth returns middleware that validates Authorization: Bearer <token>.
func bearerAuth(token string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			auth := r.Header.Get("Authorization")
			if auth == "" || len(auth) < 8 || auth[:7] != "Bearer " || auth[7:] != token {
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
