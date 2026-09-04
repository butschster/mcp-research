// Package testdb runs the same behavioral tests against an isolated database
// per fixture. Network credentials are used only to CREATE/DROP random test
// databases; the database named in the supplied DSN is never migrated or reset.
package testdb

import (
	"context"
	"database/sql"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/butschster/mcp-research/internal/config"
	"github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/uptrace/bun/driver/pgdriver"
)

func Config(t testing.TB) config.Config {
	t.Helper()
	driver := os.Getenv("TEST_DATABASE_DRIVER")
	if driver == "" || driver == "sqlite" {
		return config.Config{}
	}
	dsn := os.Getenv("TEST_DATABASE_DSN")
	if dsn == "" {
		t.Fatal("TEST_DATABASE_DSN is required when TEST_DATABASE_DRIVER is set")
	}
	name := "research_test_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	var admin *sql.DB
	var quote, isolatedDSN string
	switch driver {
	case "postgres":
		u, err := url.Parse(dsn)
		if err != nil {
			t.Fatal("invalid PostgreSQL test DSN")
		}
		admin = sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))
		u.Path = "/" + name
		isolatedDSN = u.String()
		quote = `"` + name + `"`
	case "mysql":
		c, err := mysql.ParseDSN(dsn)
		if err != nil {
			t.Fatal("invalid MySQL test DSN")
		}
		admin, err = sql.Open("mysql", dsn)
		if err != nil {
			t.Fatal(err)
		}
		c.DBName = name
		isolatedDSN = c.FormatDSN()
		quote = "`" + name + "`"
	default:
		t.Fatalf("unsupported TEST_DATABASE_DRIVER %q", driver)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if _, err := admin.ExecContext(ctx, "CREATE DATABASE "+quote); err != nil {
		admin.Close()
		t.Fatalf("create isolated test database: %v", err)
	}
	t.Cleanup(func() {
		defer admin.Close()
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if _, err := admin.ExecContext(ctx, "DROP DATABASE "+quote); err != nil {
			t.Errorf("drop isolated test database: %v", err)
		}
	})
	return config.Config{DBDriver: driver, DBDSN: isolatedDSN}
}
