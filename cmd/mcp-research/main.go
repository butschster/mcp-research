package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/butschster/mcp-research/internal/api"
	"github.com/butschster/mcp-research/internal/api/ws"
	"github.com/butschster/mcp-research/internal/auth"
	"github.com/butschster/mcp-research/internal/config"
	"github.com/butschster/mcp-research/internal/domain"
	mcpserver "github.com/butschster/mcp-research/internal/mcp"
	"github.com/butschster/mcp-research/internal/service"
	"github.com/butschster/mcp-research/internal/storage"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	cfg := config.Load()

	if cfg.Version {
		fmt.Printf("mcp-research %s (commit: %s, built: %s)\n", version, commit, date)
		os.Exit(0)
	}

	logLevel := parseLogLevel(cfg.LogLevel)
	log := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: logLevel}))

	db, err := storage.NewDB(cfg, log)
	if err != nil {
		log.Error("failed to initialize database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// Backfill short codes for any records missing them (after migrations)
	if n, err := storage.BackfillCodes(context.Background(), db); err != nil {
		log.Error("failed to backfill codes", "error", err)
	} else if n > 0 {
		log.Info("backfilled short codes", "count", n)
	}

	// WebSocket hub + event notifier
	hub := ws.NewHub(log)
	events := ws.NewHubNotifier(hub)

	// Repositories
	researchRepo := storage.NewResearchRepository(db)
	sectionRepo := storage.NewSectionRepository(db)
	entryRepo := storage.NewEntryRepository(db)
	blockRepo := storage.NewBlockRepository(db)
	sessionRepo := storage.NewSessionRepository(db)
	questionRepo := storage.NewQuestionRepository(db)
	taskRepo := storage.NewTaskRepository(db)
	revisionRepo := storage.NewEntryRevisionRepository(db)
	crossrefRepo := storage.NewCrossRefRepository(db)
	externalLinkRepo := storage.NewExternalLinkRepository(db)
	roadmapRepo := storage.NewRoadmapRepository(db)
	roadmapNodeRepo := storage.NewRoadmapNodeRepository(db)
	roadmapEdgeRepo := storage.NewRoadmapEdgeRepository(db)
	teamRepo := storage.NewTeamRepository(db)
	teamInviteRepo := storage.NewTeamInviteRepository(db)
	userRepo := storage.NewUserRepository(db)

	// Every service asks the same guard what the caller may do, so there is one
	// place to get authorization wrong instead of eight.
	access := service.NewAccess(teamRepo)

	// Services
	researchSvc := service.NewResearchService(researchRepo, sectionRepo, teamRepo, access, events, log)
	sectionSvc := service.NewSectionService(sectionRepo, entryRepo, researchRepo, access, events, log)
	entrySvc := service.NewEntryService(entryRepo, sectionRepo, researchRepo, access, sessionRepo, blockRepo, revisionRepo, crossrefRepo, externalLinkRepo, events, log)
	entrySvc.SetRoadmapRepos(roadmapRepo, roadmapNodeRepo)
	entrySvc.SetRevisionLimit(cfg.RevisionLimit)
	sessionSvc := service.NewSessionService(db, sessionRepo, questionRepo, researchRepo, access, entrySvc, events, log)
	taskSvc := service.NewTaskService(taskRepo, researchRepo, access, entrySvc, events, log)
	roadmapSvc := service.NewRoadmapService(roadmapRepo, roadmapNodeRepo, roadmapEdgeRepo, researchRepo, access, events, log)
	roadmapSvc.SetRefResolvers(entryRepo, taskRepo, sessionRepo, questionRepo, sectionRepo)
	exportSvc := service.NewExportService(researchSvc, sectionSvc, entrySvc, entryRepo, sessionSvc, taskSvc, roadmapSvc, log)
	obsidianSvc := service.NewObsidianService(researchSvc, sectionSvc, entryRepo, sessionSvc, taskSvc, roadmapSvc, revisionRepo, log)
	teamSvc := service.NewTeamService(teamRepo, teamInviteRepo, userRepo, researchRepo, events, log)

	// Auth (optional)
	var authSvc *service.AuthService
	var oauthSvc *service.OAuthService
	var defaultUser *domain.User
	var autoLoginToken string
	if cfg.AuthEnabled {
		if cfg.JWTSecret == "" {
			cfg.JWTSecret = generateRandomSecret()
			log.Warn("no jwt_secret configured, generated random secret (will change on restart)")
		}

		apiKeyRepo := storage.NewAPIKeyRepository(db)
		oauthRepo := storage.NewOAuthRepository(db)
		jwtMgr := auth.NewJWTManager(cfg.JWTSecret, 30*24*time.Hour)

		authSvc = service.NewAuthService(userRepo, apiKeyRepo, oauthRepo, researchRepo, teamRepo, jwtMgr, cfg.AllowRegistration, log)
		oauthSvc = service.NewOAuthService(oauthRepo, log)

		// Resolve or auto-create default user
		if cfg.DefaultUser != "" {
			u, err := userRepo.FindByEmail(context.Background(), cfg.DefaultUser)
			if err != nil {
				log.Error("failed to find default user", "email", cfg.DefaultUser, "error", err)
				os.Exit(1)
			}
			if u == nil {
				// Auto-create for local development
				u, _, err = authSvc.Register(context.Background(), cfg.DefaultUser, generateRandomSecret(), "Default User")
				if err != nil {
					log.Error("failed to auto-create default user", "email", cfg.DefaultUser, "error", err)
					os.Exit(1)
				}
				log.Info("auto-created default user", "email", cfg.DefaultUser, "id", u.ID)
			}
			defaultUser = u
			// Generate auto-login token for Web UI
			if token, err := jwtMgr.Generate(u.ID); err == nil {
				autoLoginToken = token
			}
			log.Info("default user configured", "email", u.Email, "id", u.ID)
		}
	}

	// MCP Server
	srv := mcpserver.NewServer(researchSvc, sectionSvc, entrySvc, sessionSvc, taskSvc, roadmapSvc, exportSvc, log, version)
	srv.SetBaseURL(cfg.BaseURL)

	log.Info("mcp-research started",
		"version", version,
		"transport", cfg.Transport,
		"web_port", cfg.WebPort,
		"db", cfg.DBPath,
		"auth_enabled", cfg.AuthEnabled,
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start REST API + WebSocket server in background
	apiCfg := api.ServerConfig{
		Port:           cfg.WebPort,
		IsInMemory:     cfg.DBPath == "",
		APIToken:       cfg.APIToken,
		AuthEnabled:    cfg.AuthEnabled,
		BaseURL:        cfg.BaseURL,
		OAuthSvc:       oauthSvc,
		AutoLoginToken: autoLoginToken,
		MCPHandler:     srv.StreamableHTTPHandler(),
	}
	apiSrv := api.NewServer(apiCfg, researchSvc, sectionSvc, entrySvc, sessionSvc, taskSvc, roadmapSvc, exportSvc, obsidianSvc, teamSvc, authSvc, db, entryRepo, researchRepo, crossrefRepo, externalLinkRepo, hub, log)
	go func() {
		if err := apiSrv.Start(ctx); err != nil {
			log.Error("API server error", "error", err)
		}
	}()

	// Run MCP server (blocking)
	switch cfg.Transport {
	case "sse":
		if err := srv.RunSSE(ctx, cfg.MCPPort, authSvc, cfg.BaseURL); err != nil {
			log.Error("SSE server error", "error", err)
			os.Exit(1)
		}
	default:
		if err := srv.RunStdio(ctx, defaultUser); err != nil {
			log.Error("server error", "error", err)
			os.Exit(1)
		}
	}
}

func parseLogLevel(s string) slog.Level {
	switch s {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func generateRandomSecret() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}
