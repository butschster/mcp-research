package storage

import (
	"log/slog"
	"testing"

	"github.com/dovod-app/app/internal/testdb"
	"github.com/uptrace/bun"
)

func setupTestDB(t *testing.T) *bun.DB {
	t.Helper()
	db, err := NewDB(testdb.Config(t), slog.Default())
	if err != nil {
		t.Fatalf("setup test db: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}
