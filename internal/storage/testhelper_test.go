package storage

import (
	"database/sql"
	"log/slog"
	"testing"

	"github.com/butschster/mcp-research/internal/config"
)

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := NewDB(config.Config{}, slog.Default())
	if err != nil {
		t.Fatalf("setup test db: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}
