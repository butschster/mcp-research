package storage

import (
	"database/sql"
	"sort"
	"strings"
	"testing"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"
)

// The teams migration is the one place in this change where existing data can
// be lost: it decides, for every research already in the database, who is
// still allowed to open it. A fresh database proves nothing about that, so
// this test stops at the migration before it, puts real rows in, and then
// applies it.

// migrateUpTo applies migrations in order, up to and including `last`.
func migrateUpTo(t *testing.T, last string) *bun.DB {
	t.Helper()

	sqldb, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	db := bun.NewDB(sqldb, sqlitedialect.New())
	t.Cleanup(func() { db.Close() })
	db.SetMaxOpenConns(1)
	if _, err := db.Exec("PRAGMA foreign_keys=ON"); err != nil {
		t.Fatalf("foreign keys: %v", err)
	}

	for _, name := range migrationNames(t) {
		applyMigration(t, db, name)
		if strings.TrimSuffix(name, ".sql") == last {
			break
		}
	}
	return db
}

func migrationNames(t *testing.T) []string {
	t.Helper()
	entries, err := migrationsFS.ReadDir("migrations")
	if err != nil {
		t.Fatalf("read migrations: %v", err)
	}
	var names []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)
	return names
}

func applyMigration(t *testing.T, db *bun.DB, name string) {
	t.Helper()
	data, err := migrationsFS.ReadFile("migrations/" + name)
	if err != nil {
		t.Fatalf("read %s: %v", name, err)
	}
	if _, err := db.Exec(string(data)); err != nil {
		t.Fatalf("apply %s: %v", name, err)
	}
}

func TestMigration019_BackfillsTeamsForExistingData(t *testing.T) {
	db := migrateUpTo(t, "018_entry_revisions")

	seed := []string{
		`INSERT INTO users (id, email, password_hash, name, created_at, updated_at)
		 VALUES ('u-alice', 'alice@test.com', 'x', 'Alice', '2024-01-01 00:00:00', '2024-01-01 00:00:00')`,
		// A user with no display name: the personal team falls back to the
		// local part of the address rather than being blank.
		`INSERT INTO users (id, email, password_hash, name, created_at, updated_at)
		 VALUES ('u-bob', 'bob@test.com', 'x', '', '2024-01-02 00:00:00', '2024-01-02 00:00:00')`,
		`INSERT INTO researches (id, code, user_id, name, description, goal, status, instruction, memory, tags, created_at, updated_at)
		 VALUES ('r-alice', 'R1', 'u-alice', 'Alice work', '', '', 'active', '', '[]', '[]', '2024-01-01 00:00:00', '2024-01-01 00:00:00')`,
		`INSERT INTO researches (id, code, user_id, name, description, goal, status, instruction, memory, tags, created_at, updated_at)
		 VALUES ('r-bob', 'R2', 'u-bob', 'Bob work', '', '', 'active', '', '[]', '[]', '2024-01-02 00:00:00', '2024-01-02 00:00:00')`,
		// The local single-user case: no owner at all.
		`INSERT INTO researches (id, code, user_id, name, description, goal, status, instruction, memory, tags, created_at, updated_at)
		 VALUES ('r-orphan', 'R3', NULL, 'Local work', '', '', 'active', '', '[]', '[]', '2024-01-03 00:00:00', '2024-01-03 00:00:00')`,
	}
	for _, stmt := range seed {
		if _, err := db.Exec(stmt); err != nil {
			t.Fatalf("seed: %v", err)
		}
	}

	applyMigration(t, db, "019_teams.sql")

	t.Run("no research is left without a team", func(t *testing.T) {
		var n int
		if err := db.QueryRow(`SELECT COUNT(*) FROM researches WHERE team_id IS NULL`).Scan(&n); err != nil {
			t.Fatalf("query: %v", err)
		}
		if n != 0 {
			t.Fatalf("%d researches have no team and are now unreachable", n)
		}
	})

	t.Run("each owner becomes an owner of their personal team", func(t *testing.T) {
		for _, tc := range []struct{ user, research, teamName string }{
			{"u-alice", "r-alice", "Alice"},
			{"u-bob", "r-bob", "bob"},
		} {
			var teamID, name, role string
			var personal int
			err := db.QueryRow(
				`SELECT t.id, t.name, t.personal, m.role
				 FROM researches r
				 JOIN teams t ON t.id = r.team_id
				 JOIN team_members m ON m.team_id = t.id AND m.user_id = ?
				 WHERE r.id = ?`, tc.user, tc.research).Scan(&teamID, &name, &personal, &role)
			if err != nil {
				t.Fatalf("%s: %v", tc.user, err)
			}
			if personal != 1 {
				t.Errorf("%s: the backfilled team should be personal", tc.user)
			}
			if name != tc.teamName {
				t.Errorf("%s: team name = %q, want %q", tc.user, name, tc.teamName)
			}
			if role != "owner" {
				t.Errorf("%s: role = %q, want owner", tc.user, role)
			}
		}
	})

	t.Run("an ownerless research goes to the local team", func(t *testing.T) {
		var teamID string
		if err := db.QueryRow(`SELECT team_id FROM researches WHERE id='r-orphan'`).Scan(&teamID); err != nil {
			t.Fatalf("query: %v", err)
		}
		if teamID != "team-local" {
			t.Fatalf("team = %q, want team-local", teamID)
		}
		// It has no members on purpose: with auth off nobody is in the request
		// context, and with auth on the first registered user claims it.
		var members int
		if err := db.QueryRow(`SELECT COUNT(*) FROM team_members WHERE team_id='team-local'`).Scan(&members); err != nil {
			t.Fatalf("query: %v", err)
		}
		if members != 0 {
			t.Fatalf("the local team should have no members, got %d", members)
		}
	})

	// The runner records a version and never re-applies a file, so the ALTER
	// only ever runs once. The backfill is a different matter: the statements
	// are not wrapped in a transaction, so an interrupted upgrade can leave
	// some of them applied. Running them again must converge rather than give
	// a user two personal teams.
	t.Run("the backfill converges when it runs twice", func(t *testing.T) {
		before := counts(t, db)
		for _, stmt := range backfillStatements(t) {
			if _, err := db.Exec(stmt); err != nil {
				t.Fatalf("re-run backfill: %v\n%s", err, stmt)
			}
		}
		after := counts(t, db)
		if before != after {
			t.Fatalf("second run changed the data: %+v -> %+v", before, after)
		}
	})
}

// backfillStatements is the migration from the ALTER onwards — everything that
// moves data rather than changing the shape of a table.
func backfillStatements(t *testing.T) []string {
	t.Helper()
	data, err := migrationsFS.ReadFile("migrations/019_teams.sql")
	if err != nil {
		t.Fatalf("read migration: %v", err)
	}
	body := string(data)
	marker := "-- Backfill."
	i := strings.Index(body, marker)
	if i < 0 {
		t.Fatal("the backfill section is no longer marked; update this test with the migration")
	}

	// Comments go first: one of them contains a semicolon, and splitting on
	// semicolons would otherwise cut it loose as prose in the middle of SQL.
	var stmts []string
	for _, stmt := range strings.Split(stripComments(body[i:]), ";") {
		if strings.TrimSpace(stmt) != "" {
			stmts = append(stmts, stmt)
		}
	}
	return stmts
}

func stripComments(s string) string {
	var b strings.Builder
	for _, line := range strings.Split(s, "\n") {
		if !strings.HasPrefix(strings.TrimSpace(line), "--") {
			b.WriteString(line)
			b.WriteString("\n")
		}
	}
	return b.String()
}

type teamCounts struct{ Teams, Members, Researches int }

func counts(t *testing.T, db *bun.DB) teamCounts {
	t.Helper()
	var c teamCounts
	row := func(q string, dest *int) {
		if err := db.QueryRow(q).Scan(dest); err != nil {
			t.Fatalf("%s: %v", q, err)
		}
	}
	row(`SELECT COUNT(*) FROM teams`, &c.Teams)
	row(`SELECT COUNT(*) FROM team_members`, &c.Members)
	row(`SELECT COUNT(DISTINCT team_id) FROM researches`, &c.Researches)
	return c
}
