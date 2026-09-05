package storage

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/butschster/mcp-research/internal/config"
	"github.com/go-sql-driver/mysql"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/mysqldialect"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/schema"
	_ "modernc.org/sqlite"
)

//go:embed migrations/*.sql migrations/postgres/*.sql migrations/mysql/*.sql
var migrationsFS embed.FS

// NewDB opens the configured database and wraps it with Bun's dialect-aware
// query builder. SQLite remains the zero-config default.
func NewDB(cfg config.Config, log *slog.Logger) (*bun.DB, error) {
	driver := normalizeDriver(cfg.DBDriver)
	sqldb, dialect, err := openSQLDB(driver, cfg, log)
	if err != nil {
		return nil, err
	}
	db := bun.NewDB(sqldb, dialect)
	if driver == "sqlite" {
		db.SetMaxOpenConns(1)
	} else {
		db.SetMaxOpenConns(10)
		db.SetMaxIdleConns(5)
		db.SetConnMaxLifetime(5 * time.Minute)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("connect %s database: %w", driver, err)
	}

	if driver == "sqlite" && !cfg.DatabaseInMemory() {
		if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
			_ = db.Close()
			return nil, fmt.Errorf("set WAL mode: %w", err)
		}
		if _, err := db.Exec("PRAGMA synchronous=NORMAL"); err != nil {
			_ = db.Close()
			return nil, fmt.Errorf("set synchronous: %w", err)
		}
		if _, err := db.Exec("PRAGMA cache_size=1000"); err != nil {
			_ = db.Close()
			return nil, fmt.Errorf("set cache_size: %w", err)
		}
	}

	if driver == "sqlite" {
		if _, err := db.Exec("PRAGMA foreign_keys=ON"); err != nil {
			_ = db.Close()
			return nil, fmt.Errorf("enable foreign keys: %w", err)
		}
		db.SetMaxOpenConns(1)
	}

	if err := runMigrations(db, driver, log); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("migrations: %w", err)
	}

	return db, nil
}

func normalizeDriver(driver string) string {
	switch strings.ToLower(strings.TrimSpace(driver)) {
	case "", "sqlite", "sqlite3":
		return "sqlite"
	case "postgres", "postgresql", "pg", "pgsql":
		return "postgres"
	case "mysql":
		return "mysql"
	default:
		return strings.ToLower(strings.TrimSpace(driver))
	}
}

func openSQLDB(driver string, cfg config.Config, log *slog.Logger) (*sql.DB, schema.Dialect, error) {
	switch driver {
	case "sqlite":
		dsn := cfg.DBDSN
		if dsn == "" {
			dsn = cfg.DBPath
		}
		if dsn == "" {
			dsn = ":memory:"
			log.Warn("in-memory database: all data will be lost on restart", "hint", "use --db ./research.db for persistence")
		}
		db, err := sql.Open("sqlite", dsn)
		return db, sqlitedialect.New(), wrapOpenError(driver, err)

	case "postgres":
		if cfg.DBDSN == "" {
			return nil, nil, fmt.Errorf("open postgres database: db_dsn/--db-dsn is required")
		}
		db, err := openPostgres(cfg.DBDSN)
		return db, pgdialect.New(), err

	case "mysql":
		if cfg.DBDSN == "" {
			return nil, nil, fmt.Errorf("open mysql database: db_dsn/--db-dsn is required")
		}
		mysqlCfg, err := mysql.ParseDSN(cfg.DBDSN)
		if err != nil {
			return nil, nil, fmt.Errorf("open mysql database: parse DSN: %w", err)
		}
		mysqlCfg.MultiStatements = true
		// Repository methods distinguish missing rows from unchanged rows.
		mysqlCfg.ClientFoundRows = true
		if mysqlCfg.Timeout == 0 {
			mysqlCfg.Timeout = 10 * time.Second
		}
		if mysqlCfg.ReadTimeout == 0 {
			mysqlCfg.ReadTimeout = 15 * time.Second
		}
		if mysqlCfg.WriteTimeout == 0 {
			mysqlCfg.WriteTimeout = 15 * time.Second
		}
		db, err := sql.Open("mysql", mysqlCfg.FormatDSN())
		return db, mysqldialect.New(), wrapOpenError(driver, err)

	default:
		return nil, nil, fmt.Errorf("unsupported database driver %q (want sqlite, postgres, or mysql)", driver)
	}
}

// pgdriver's option panics on malformed DSNs. Configuration mistakes must
// return an error without printing credentials in a panic message.
func openPostgres(dsn string) (db *sql.DB, err error) {
	defer func() {
		if recover() != nil {
			db = nil
			err = fmt.Errorf("invalid postgres DSN")
		}
	}()
	return sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn))), nil
}

func wrapOpenError(driver string, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("open %s database: %w", driver, err)
}

type migrationRow struct {
	bun.BaseModel `bun:"table:schema_migrations"`
	Version       string `bun:"version,pk"`
	AppliedAt     string `bun:"applied_at"`
}

func runMigrations(db *bun.DB, driver string, log *slog.Logger) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	conn, err := db.Conn(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()
	// A dedicated connection keeps the advisory lock for the full upgrade.
	switch driver {
	case "postgres":
		if _, err := conn.ExecContext(ctx, "SELECT pg_advisory_lock(732164901)"); err != nil {
			return err
		}
		defer conn.ExecContext(context.Background(), "SELECT pg_advisory_unlock(732164901)")
	case "mysql":
		var locked int
		if err := conn.QueryRowContext(ctx, "SELECT GET_LOCK(CONCAT('mcp-research:', MD5(DATABASE())), 60)").Scan(&locked); err != nil {
			return err
		}
		if locked != 1 {
			return fmt.Errorf("could not acquire migration lock")
		}
		defer conn.ExecContext(context.Background(), "SELECT RELEASE_LOCK(CONCAT('mcp-research:', MD5(DATABASE())))")
	}
	dir := "migrations"
	if driver != "sqlite" {
		dir += "/" + driver
	}
	entries, err := migrationsFS.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("read %s migrations dir: %w", driver, err)
	}

	// Ensure schema_migrations table exists (bootstrap)
	if _, err := conn.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (
		version VARCHAR(255) PRIMARY KEY,
		applied_at VARCHAR(32) NOT NULL
	)`); err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}
	// MySQL DDL commits implicitly. Persist an in-progress marker so a failed
	// upgrade requires explicit recovery instead of replaying partial DDL.
	if driver == "mysql" {
		if _, err := conn.ExecContext(ctx, "CREATE TABLE IF NOT EXISTS schema_migration_dirty (version VARCHAR(255) PRIMARY KEY)"); err != nil {
			return err
		}
		var dirty string
		err := conn.NewSelect().Column("version").Table("schema_migration_dirty").Limit(1).Scan(ctx, &dirty)
		if err == nil {
			return fmt.Errorf("MySQL migration %s is incomplete; restore or repair the schema before clearing schema_migration_dirty", dirty)
		}
		if err != sql.ErrNoRows {
			return err
		}
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
		if err := conn.NewSelect().ColumnExpr("COUNT(*)").Table("schema_migrations").Where("version=?", version).Scan(ctx, &count); err != nil {
			return fmt.Errorf("check migration %s: %w", version, err)
		}
		if count > 0 {
			continue
		}

		data, err := migrationsFS.ReadFile(dir + "/" + name)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", name, err)
		}

		if err := applyMigrationVersion(ctx, conn, driver, version, data); err != nil {
			return err
		}

		log.Info("applied migration", "version", version)
	}

	return nil
}

func applyMigrationVersion(ctx context.Context, conn Querier, driver, version string, data []byte) error {
	apply := func(q Querier) error {
		if _, err := q.ExecContext(ctx, string(data)); err != nil {
			return fmt.Errorf("apply migration %s: %w", version, err)
		}
		if version == "029_research_memory" || version == "003_research_memory" {
			if err := migrateResearchMemory(ctx, q); err != nil {
				return fmt.Errorf("migrate research memory: %w", err)
			}
		}
		_, err := q.NewInsert().Model(&migrationRow{Version: version, AppliedAt: time.Now().UTC().Format(time.RFC3339)}).Exec(ctx)
		return err
	}
	if driver != "mysql" {
		return conn.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error { return apply(tx) })
	}
	if _, err := conn.NewInsert().Table("schema_migration_dirty").Model(&map[string]any{"version": version}).Exec(ctx); err != nil {
		return err
	}
	if err := apply(conn); err != nil {
		return err
	}
	_, err := conn.NewDelete().Table("schema_migration_dirty").Where("version=?", version).Exec(ctx)
	return err
}
