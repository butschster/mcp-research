package storage

import (
	"context"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log/slog"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/butschster/mcp-research/internal/config"
	"github.com/butschster/mcp-research/internal/domain"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect"
)

// These contracts run on every CI database, alongside all existing repository,
// service and HTTP authorization tests. SQL mocks would miss these differences.
func TestBunContract_TextAndUnchangedUpdate(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()
	users := NewUserRepository(db)
	text := "O'Reilly ? $1 \\ 日本語\n'); DROP TABLE users; --"
	u := &domain.User{ID: uuid.NewString(), Email: "quoted@test.invalid", Name: text, PasswordHash: "test"}
	if err := users.Create(ctx, u); err != nil {
		t.Fatal(err)
	}
	got, err := users.FindByID(ctx, u.ID)
	if err != nil || got == nil || got.Name != text {
		t.Fatalf("text roundtrip: %#v, %v", got, err)
	}
	teams := NewTeamRepository(db)
	team := &domain.Team{ID: uuid.NewString(), Name: text, CreatedBy: u.ID}
	if err := teams.CreateWithOwner(ctx, team, u.ID); err != nil {
		t.Fatal(err)
	}
	if err := teams.UpdateRole(ctx, team.ID, u.ID, domain.TeamOwner); err != nil {
		t.Fatalf("unchanged member must still exist: %v", err)
	}
	if err := teams.AddMember(ctx, team.ID, u.ID, domain.TeamEditor, u.ID); err != nil {
		t.Fatal(err)
	}
	if err := teams.AddMember(ctx, team.ID, "missing-user", domain.TeamEditor, u.ID); err == nil {
		t.Fatal("upsert swallowed foreign-key violation")
	}
}

func TestBunContract_FailedMigration(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()
	driver := "sqlite"
	if db.Dialect().Name() == dialect.PG {
		driver = "postgres"
	}
	if db.Dialect().Name() == dialect.MySQL {
		driver = "mysql"
	}
	version := "test_failure"
	err := applyMigrationVersion(ctx, db, driver, version, []byte("CREATE TABLE migration_probe (id INTEGER); INSERT INTO missing_migration_table VALUES (1);"))
	if err == nil {
		t.Fatal("injected migration must fail")
	}
	var n int
	if err := db.NewSelect().Table("schema_migrations").ColumnExpr("COUNT(*)").Where("version=?", version).Scan(ctx, &n); err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Fatal("failed migration recorded as successful")
	}
	if driver == "mysql" {
		if err := runMigrations(db, driver, slog.Default()); err == nil || !strings.Contains(err.Error(), "incomplete") {
			t.Fatalf("partial MySQL DDL must block restart: %v", err)
		}
	} else {
		if err := db.NewSelect().Table("migration_probe").ColumnExpr("COUNT(*)").Scan(ctx, &n); err == nil {
			t.Fatal("failed migration left a table behind")
		}
		if err := runMigrations(db, driver, slog.Default()); err != nil {
			t.Fatalf("restart after rollback: %v", err)
		}
	}
}

func TestSQLiteUpgradeFromPreBunPreservesCodes(t *testing.T) {
	db := migrateUpTo(t, "027_entry_views")
	ctx := context.Background()
	// Reproduce an old deployment's migration ledger and a user-assigned code.
	for _, name := range migrationNames(t) {
		if strings.HasPrefix(name, "028_") {
			break
		}
		if _, err := db.NewInsert().Model(&migrationRow{Version: strings.TrimSuffix(name, ".sql"), AppliedAt: "2026-01-01 00:00:00"}).Exec(ctx); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := db.Exec("INSERT INTO researches (id, code, name, team_id) VALUES ('legacy', 'R41', 'legacy content', 'team-local')"); err != nil {
		t.Fatal(err)
	}
	if err := runMigrations(db, "sqlite", slog.Default()); err != nil {
		t.Fatal(err)
	}
	repo := NewResearchRepository(db)
	legacy, err := repo.FindByID(ctx, "legacy")
	if err != nil || legacy == nil || legacy.Code != "R41" || legacy.Name != "legacy content" {
		t.Fatalf("legacy data changed: %#v %v", legacy, err)
	}
	r := &domain.Research{ID: uuid.NewString(), Name: "new", TeamID: "team-local", Status: domain.ResearchActive}
	if err := repo.Create(ctx, r); err != nil {
		t.Fatal(err)
	}
	if r.Code != "R42" {
		t.Fatalf("counter did not seed from legacy data: %s", r.Code)
	}
	if err := runMigrations(db, "sqlite", slog.Default()); err != nil {
		t.Fatal(err)
	}
}

func TestBunContract_ConcurrentCodes(t *testing.T) {
	db := setupTestDB(t)
	repo := NewResearchRepository(db)
	ctx := context.Background()
	const writers = 16
	var wg sync.WaitGroup
	errors := make(chan error, writers)
	codes := make(chan string, writers)
	for i := 0; i < writers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			r := &domain.Research{ID: uuid.NewString(), Name: "concurrent", TeamID: "team-local", Status: domain.ResearchActive}
			if err := repo.Create(ctx, r); err != nil {
				errors <- err
				return
			}
			codes <- r.Code
		}()
	}
	wg.Wait()
	close(errors)
	close(codes)
	for err := range errors {
		t.Error(err)
	}
	seen := map[string]bool{}
	for code := range codes {
		if seen[code] {
			t.Errorf("duplicate code %s", code)
		}
		seen[code] = true
	}
	if len(seen) != writers {
		t.Fatalf("got %d unique codes, want %d", len(seen), writers)
	}
}

func TestBunContract_RoadmapNodeOrderWithTimestampTies(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()
	research := &domain.Research{ID: uuid.NewString(), Name: "ordered", TeamID: "team-local", Status: domain.ResearchActive}
	if err := NewResearchRepository(db).Create(ctx, research); err != nil {
		t.Fatal(err)
	}
	roadmap := &domain.Roadmap{ID: uuid.NewString(), ResearchID: research.ID, Title: "ordered"}
	if err := NewRoadmapRepository(db).Create(ctx, roadmap); err != nil {
		t.Fatal(err)
	}
	repo := NewRoadmapNodeRepository(db)
	for i := 1; i <= 12; i++ {
		node := &domain.RoadmapNode{ID: uuid.NewString(), RoadmapID: roadmap.ID, Title: fmt.Sprintf("Node %d", i)}
		if err := repo.Create(ctx, node); err != nil {
			t.Fatal(err)
		}
	}
	// Force a tie independently of machine speed and wall-clock boundaries.
	if _, err := db.NewUpdate().Table("roadmap_nodes").Set("created_at=?", "2026-01-01 00:00:00").Where("roadmap_id=?", roadmap.ID).Exec(ctx); err != nil {
		t.Fatal(err)
	}
	nodes, err := repo.FindByRoadmap(ctx, roadmap.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(nodes) != 12 {
		t.Fatalf("got %d nodes, want 12", len(nodes))
	}
	for i, node := range nodes {
		if want := fmt.Sprintf("N%d", i+1); node.Code != want {
			t.Errorf("node %d: got %s, want %s", i, node.Code, want)
		}
	}
}

func TestBunContract_RollbackUsesCallerTransaction(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()
	r := &domain.Research{ID: uuid.NewString(), Name: "rollback", TeamID: "team-local", Status: domain.ResearchActive}
	if err := NewResearchRepository(db).Create(ctx, r); err != nil {
		t.Fatal(err)
	}
	section := &domain.Section{ID: uuid.NewString(), ResearchID: r.ID, Name: "rolled back", Status: domain.SectionDraft}
	want := errors.New("injected failure after section insert")
	err := db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		if err := NewSectionRepository(db).CreateTx(ctx, tx, section); err != nil {
			return err
		}
		return want
	})
	if !errors.Is(err, want) {
		t.Fatal(err)
	}
	var n int
	if err := db.NewSelect().Table("sections").ColumnExpr("COUNT(*)").Where("id=?", section.ID).Scan(ctx, &n); err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Fatal("section escaped the rolled-back transaction")
	}
}

func TestSQLiteReopenPreservesDataAndPragmas(t *testing.T) {
	cfg := config.Config{DBDSN: filepath.Join(t.TempDir(), "research.db")}
	db, err := NewDB(cfg, slog.Default())
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	u := &domain.User{ID: uuid.NewString(), Email: "persist@test.invalid", Name: "persistent", PasswordHash: "test"}
	if err := NewUserRepository(db).Create(ctx, u); err != nil {
		db.Close()
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	db, err = NewDB(cfg, slog.Default())
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	got, err := NewUserRepository(db).FindByID(ctx, u.ID)
	if err != nil || got == nil || got.Name != u.Name {
		t.Fatalf("reopen: %#v %v", got, err)
	}
	var mode string
	var foreignKeys int
	if err := db.QueryRow("PRAGMA journal_mode").Scan(&mode); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow("PRAGMA foreign_keys").Scan(&foreignKeys); err != nil {
		t.Fatal(err)
	}
	if mode != "wal" || foreignKeys != 1 {
		t.Fatalf("journal=%s foreign_keys=%d", mode, foreignKeys)
	}
}

func TestDatabaseRejectsInvalidConfiguration(t *testing.T) {
	for _, cfg := range []config.Config{{DBDriver: "unknown"}, {DBDriver: "postgres"}, {DBDriver: "mysql"}, {DBDriver: "postgres", DBDSN: "%bad"}} {
		t.Run(fmt.Sprintf("%s/%s", cfg.DBDriver, cfg.DBDSN), func(t *testing.T) {
			db, err := NewDB(cfg, slog.Default())
			if db != nil {
				db.Close()
			}
			if err == nil {
				t.Fatal("expected configuration error")
			}
		})
	}
}

func TestRepositoriesUseBunBuilders(t *testing.T) {
	files, err := filepath.Glob("*_repo.go")
	if err != nil {
		t.Fatal(err)
	}
	if len(files) == 0 {
		t.Fatal("no repository files checked")
	}
	for _, file := range files {
		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, file, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		ast.Inspect(f, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			sel, ok := call.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			switch sel.Sel.Name {
			case "ExecContext", "QueryContext", "QueryRowContext", "PrepareContext", "NewRaw":
				t.Errorf("%s: raw SQL belongs in a dialect adapter, not a repository", fset.Position(call.Pos()))
			}
			return true
		})
	}
}
