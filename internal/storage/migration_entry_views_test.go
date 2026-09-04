package storage

import (
	"context"
	"testing"
)

func TestMigration027_BaselinesExistingReadersWithoutCreatingAFalseQueue(t *testing.T) {
	db := migrateUpTo(t, "026_annotations")
	seed := []string{
		`INSERT INTO users (id, email, password_hash, name) VALUES ('u-owner', 'owner@test.com', 'x', 'Owner')`,
		`INSERT INTO users (id, email, password_hash, name) VALUES ('u-reader', 'reader@test.com', 'x', 'Reader')`,
		`INSERT INTO teams (id, name, personal, created_by) VALUES ('team-shared', 'Shared', 0, 'u-owner')`,
		`INSERT INTO team_members (team_id, user_id, role) VALUES ('team-shared', 'u-owner', 'owner')`,
		`INSERT INTO team_members (team_id, user_id, role) VALUES ('team-shared', 'u-reader', 'viewer')`,
		`INSERT INTO researches (id, code, user_id, team_id, name) VALUES ('r-shared', 'R1', 'u-owner', 'team-shared', 'Shared research')`,
		`INSERT INTO researches (id, code, team_id, name) VALUES ('r-local', 'R2', 'team-local', 'Local research')`,
		`INSERT INTO sections (id, code, research_id, name) VALUES ('s-shared', 'S1', 'r-shared', 'shared')`,
		`INSERT INTO sections (id, code, research_id, name) VALUES ('s-local', 'S1', 'r-local', 'local')`,
		`INSERT INTO entries (id, code, research_id, section_id, title) VALUES ('e-shared', 'E1', 'r-shared', 's-shared', 'Existing shared document')`,
		`INSERT INTO entries (id, code, research_id, section_id, title) VALUES ('e-local', 'E1', 'r-local', 's-local', 'Existing local document')`,
		`INSERT INTO entry_revisions (id, entry_id, research_id, revision) VALUES ('rev-s1', 'e-shared', 'r-shared', 1)`,
		`INSERT INTO entry_revisions (id, entry_id, research_id, revision) VALUES ('rev-s2', 'e-shared', 'r-shared', 2)`,
		`INSERT INTO entry_revisions (id, entry_id, research_id, revision) VALUES ('rev-l1', 'e-local', 'r-local', 1)`,
		`INSERT INTO entry_revisions (id, entry_id, research_id, revision) VALUES ('rev-l2', 'e-local', 'r-local', 2)`,
	}
	for _, statement := range seed {
		if _, err := db.Exec(statement); err != nil {
			t.Fatalf("seed: %v\n%s", err, statement)
		}
	}

	applyMigration(t, db, "027_entry_views.sql")

	for _, tc := range []struct {
		viewer string
		entry  string
	}{
		{viewer: "u-owner", entry: "e-shared"},
		{viewer: "u-reader", entry: "e-shared"},
		{viewer: "local", entry: "e-local"},
	} {
		var revision int
		if err := db.QueryRow(
			`SELECT seen_revision FROM entry_views WHERE viewer_id=? AND entry_id=?`,
			tc.viewer, tc.entry,
		).Scan(&revision); err != nil {
			t.Fatalf("baseline %s/%s: %v", tc.viewer, tc.entry, err)
		}
		if revision != 2 {
			t.Errorf("baseline %s/%s = r%d, want r2", tc.viewer, tc.entry, revision)
		}
	}

	repo := NewEntryViewRepository(db)
	updates, err := repo.ListUpdates(context.Background(), "u-reader", "r-shared")
	if err != nil {
		t.Fatalf("list updates: %v", err)
	}
	if len(updates) != 0 {
		t.Fatalf("upgrade created a false queue: %+v", updates)
	}

	if _, err := db.Exec(`INSERT INTO entry_revisions (id, entry_id, research_id, revision) VALUES ('rev-s3', 'e-shared', 'r-shared', 3)`); err != nil {
		t.Fatalf("append post-migration revision: %v", err)
	}
	updates, err = repo.ListUpdates(context.Background(), "u-reader", "r-shared")
	if err != nil {
		t.Fatalf("list changed updates: %v", err)
	}
	if len(updates) != 1 || updates[0].Kind != "changed" || updates[0].SeenRevision != 2 || updates[0].CurrentRevision != 3 {
		t.Fatalf("real post-upgrade change missing: %+v", updates)
	}
}
