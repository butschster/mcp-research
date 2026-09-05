package storage

import (
	"context"
	"log/slog"
	"reflect"
	"strings"
	"testing"

	"github.com/butschster/mcp-research/internal/domain"
	"github.com/uptrace/bun"
)

// Reconstruct the previous two columns on an isolated test database. Unlike a
// SQLite-only fixture, this exercises the deployed migration runner on every
// database in the CI matrix, including MySQL's dirty-marker failure path.
func memoryLegacyDB(t *testing.T) (*bun.DB, string) {
	t.Helper()
	db := setupTestDB(t)
	driver := "sqlite"
	if db.Dialect().Name().String() == "pg" {
		driver = "postgres"
	}
	if db.Dialect().Name().String() == "mysql" {
		driver = "mysql"
	}
	for _, statement := range []string{
		"DROP TABLE research_memory",
		"ALTER TABLE researches ADD COLUMN instruction TEXT",
		"ALTER TABLE researches ADD COLUMN memory TEXT",
		"DELETE FROM schema_migrations WHERE version IN ('029_research_memory', '003_research_memory')",
	} {
		if _, err := db.Exec(statement); err != nil {
			t.Fatal(err)
		}
	}
	return db, driver
}

func TestMemoryMigration_PreservesLegacyDataAndRestarts(t *testing.T) {
	db, driver := memoryLegacyDB(t)
	ctx := context.Background()
	body := strings.Repeat("Exact instruction ? ' 日本語\n", 1000)
	for _, row := range []map[string]any{
		{"id": "legacy", "name": "Old research", "instruction": body, "memory": `["duplicate", "duplicate", "", "O'Reilly ? 日本語\nnext line"]`},
		{"id": "empty", "name": "Empty research", "instruction": "", "memory": "[]"},
	} {
		if _, err := db.NewInsert().Table("researches").Model(&row).Exec(ctx); err != nil {
			t.Fatal(err)
		}
	}
	if err := NewSkillRepository(db).Create(ctx, &domain.Skill{ID: "existing", ResearchID: "legacy", Slug: "legacy-instruction", Name: "Existing", Tier: domain.SkillPrivate, Body: "do not replace", Version: 1}); err != nil {
		t.Fatal(err)
	}
	if err := runMigrations(db, driver, slog.Default()); err != nil {
		t.Fatal(err)
	}
	items, err := NewMemoryRepository(db).List(ctx, "legacy")
	if err != nil || len(items) != 4 {
		t.Fatalf("memory: %+v %v", items, err)
	}
	want := []string{"duplicate", "duplicate", "", "O'Reilly ? 日本語\nnext line"}
	seen := map[string]bool{}
	for i, item := range items {
		if item.Text != want[i] || item.ID == "" || seen[item.ID] || item.Author != "unknown" || item.CreatedAt != nil || item.SessionID != "" || item.Version != 1 {
			t.Fatalf("item %d: %+v", i, item)
		}
		seen[item.ID] = true
	}
	skills, err := NewResearchRepository(db).ExportPrivateSkills(ctx, "legacy")
	if err != nil || len(skills) != 2 {
		t.Fatalf("skills: %+v %v", skills, err)
	}
	if skills[0].Body != "do not replace" || skills[1].Slug != "legacy-instruction-2" || skills[1].Body != body || !skills[1].NeedsTrigger || !skills[1].Attached {
		t.Fatalf("instruction conversion: %+v", skills)
	}
	for _, column := range []string{"instruction", "memory"} {
		var ignored string
		if err := db.NewSelect().TableExpr("researches r").ColumnExpr("r."+column).Where("r.id='legacy'").Scan(ctx, &ignored); err == nil {
			t.Fatalf("legacy %s column still active", column)
		}
	}
	if err := runMigrations(db, driver, slog.Default()); err != nil {
		t.Fatal(err)
	}
	again, err := NewMemoryRepository(db).List(ctx, "legacy")
	if err != nil || !reflect.DeepEqual(items, again) {
		t.Fatalf("restart changed IDs or notes: %+v %v", again, err)
	}
	againSkills, err := NewResearchRepository(db).ExportPrivateSkills(ctx, "legacy")
	if err != nil || !reflect.DeepEqual(skills, againSkills) {
		t.Fatalf("restart changed skills: %+v %v", againSkills, err)
	}
}

func TestMemoryMigration_InvalidJSONFailsClosed(t *testing.T) {
	db, driver := memoryLegacyDB(t)
	ctx := context.Background()
	row := map[string]any{"id": "bad", "name": "Malformed", "instruction": "preserve me", "memory": `["note", 42]`}
	if _, err := db.NewInsert().Table("researches").Model(&row).Exec(ctx); err != nil {
		t.Fatal(err)
	}
	if err := runMigrations(db, driver, slog.Default()); err == nil {
		t.Fatal("malformed legacy JSON was silently discarded")
	}
	var memory, instruction string
	if err := db.NewSelect().Table("researches").Column("memory", "instruction").Where("id='bad'").Scan(ctx, &memory, &instruction); err != nil {
		t.Fatal(err)
	}
	if memory != row["memory"] || instruction != row["instruction"] {
		t.Fatal("failure modified original data")
	}
	var n int
	if err := db.NewSelect().Table("schema_migrations").ColumnExpr("COUNT(*)").Where("version IN ('029_research_memory', '003_research_memory')").Scan(ctx, &n); err != nil || n != 0 {
		t.Fatalf("failed migration committed: %d %v", n, err)
	}
	if driver == "mysql" {
		if err := runMigrations(db, driver, slog.Default()); err == nil || !strings.Contains(err.Error(), "incomplete") {
			t.Fatalf("dirty migration replayed: %v", err)
		}
	} else {
		if _, err := db.NewUpdate().Table("researches").Set("memory=?", `["repaired"]`).Where("id='bad'").Exec(ctx); err != nil {
			t.Fatal(err)
		}
		if err := runMigrations(db, driver, slog.Default()); err != nil {
			t.Fatalf("transaction rollback did not allow retry: %v", err)
		}
	}
}
