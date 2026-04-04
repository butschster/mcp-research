package storage

import (
	"database/sql"
	"embed"
	"fmt"
	"log/slog"
	"strings"

	"github.com/butschster/mcp-research/internal/config"

	_ "modernc.org/sqlite"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

func NewDB(cfg config.Config, log *slog.Logger) (*sql.DB, error) {
	dsn := ":memory:"
	if cfg.DBPath != "" {
		dsn = cfg.DBPath
	} else {
		log.Warn("in-memory database: all data will be lost on restart", "hint", "use --db ./research.db for persistence")
	}

	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	if cfg.DBPath != "" {
		if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
			return nil, fmt.Errorf("set WAL mode: %w", err)
		}
		if _, err := db.Exec("PRAGMA synchronous=NORMAL"); err != nil {
			return nil, fmt.Errorf("set synchronous: %w", err)
		}
		if _, err := db.Exec("PRAGMA cache_size=1000"); err != nil {
			return nil, fmt.Errorf("set cache_size: %w", err)
		}
	}

	if _, err := db.Exec("PRAGMA foreign_keys=ON"); err != nil {
		return nil, fmt.Errorf("enable foreign keys: %w", err)
	}

	db.SetMaxOpenConns(1)

	if err := runMigrations(db, log); err != nil {
		return nil, fmt.Errorf("migrations: %w", err)
	}

	return db, nil
}

func runMigrations(db *sql.DB, log *slog.Logger) error {
	entries, err := migrationsFS.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("read migrations dir: %w", err)
	}

	// Ensure schema_migrations table exists (bootstrap)
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		version TEXT PRIMARY KEY,
		applied_at TEXT NOT NULL DEFAULT (datetime('now'))
	)`); err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".sql") {
			continue
		}

		version := strings.TrimSuffix(name, ".sql")

		var count int
		if err := db.QueryRow("SELECT COUNT(*) FROM schema_migrations WHERE version = ?", version).Scan(&count); err != nil {
			return fmt.Errorf("check migration %s: %w", version, err)
		}
		if count > 0 {
			continue
		}

		data, err := migrationsFS.ReadFile("migrations/" + name)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", name, err)
		}

		if _, err := db.Exec(string(data)); err != nil {
			return fmt.Errorf("apply migration %s: %w", name, err)
		}

		if _, err := db.Exec("INSERT INTO schema_migrations (version) VALUES (?)", version); err != nil {
			return fmt.Errorf("record migration %s: %w", version, err)
		}

		log.Info("applied migration", "version", version)
	}

	return nil
}
