-- The last document revision a person actually saw.
--
-- This is server state rather than a browser preference: a reader may move
-- between devices, and two members of one team must never clear each other's
-- update queue. `viewer_id` is the lookup key. For an authenticated reader it
-- equals user_id; the one local, auth-disabled reader uses `local`, with no
-- user row to reference.
--
-- seen_revision deliberately is not a foreign key. Revision trimming may
-- remove old snapshots, while the read state still has to say that something
-- newer exists. The runtime trim query preserves revisions that are active
-- comparison bases, so the normal "show changes since I read this" path keeps
-- working even when a revision limit is configured.
CREATE TABLE IF NOT EXISTS entry_views (
    viewer_id     TEXT NOT NULL,
    user_id       TEXT REFERENCES users(id) ON DELETE CASCADE,
    entry_id      TEXT NOT NULL REFERENCES entries(id) ON DELETE CASCADE,
    seen_revision INTEGER NOT NULL CHECK (seen_revision > 0),
    seen_at       TEXT NOT NULL DEFAULT (datetime('now')),
    PRIMARY KEY (viewer_id, entry_id),
    CHECK (user_id IS NULL OR viewer_id = user_id)
);

-- The primary key starts with viewer_id; entry deletion needs the other
-- direction, and revision trimming asks which bases one entry must retain.
CREATE INDEX IF NOT EXISTS idx_entry_views_entry ON entry_views(entry_id);

-- Upgrading an established installation must not label its entire archive
-- "New". Existing members start at the current head; documents created or
-- changed after this migration are the first ones that enter their queue.
INSERT OR IGNORE INTO entry_views (viewer_id, user_id, entry_id, seen_revision, seen_at)
SELECT tm.user_id, tm.user_id, e.id, MAX(er.revision), datetime('now')
  FROM entries e
  JOIN researches r ON r.id = e.research_id
  JOIN team_members tm ON tm.team_id = r.team_id
  JOIN entry_revisions er ON er.entry_id = e.id
 GROUP BY tm.user_id, e.id;

-- Auth-disabled installations have no user or membership. They are one local
-- reader, so they get the same non-disruptive upgrade baseline.
INSERT OR IGNORE INTO entry_views (viewer_id, user_id, entry_id, seen_revision, seen_at)
SELECT 'local', NULL, e.id, MAX(er.revision), datetime('now')
  FROM entries e
  JOIN researches r ON r.id = e.research_id
  JOIN entry_revisions er ON er.entry_id = e.id
 WHERE r.team_id = 'team-local'
 GROUP BY e.id;
