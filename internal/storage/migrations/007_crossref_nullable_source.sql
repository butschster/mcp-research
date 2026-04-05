-- Make source_entry_id nullable to support question/task crossrefs.
-- SQLite doesn't support ALTER COLUMN, so recreate the table.

CREATE TABLE crossrefs_new (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    source_type TEXT NOT NULL DEFAULT 'entry',
    source_id TEXT NOT NULL DEFAULT '',
    source_entry_id TEXT DEFAULT NULL,
    source_research_id TEXT NOT NULL,
    target_entry_id TEXT DEFAULT NULL,
    target_research_id TEXT DEFAULT NULL,
    target_ref TEXT NOT NULL,
    resolved INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

INSERT INTO crossrefs_new (id, source_type, source_id, source_entry_id, source_research_id, target_entry_id, target_research_id, target_ref, resolved, created_at)
  SELECT id, source_type, source_id,
         CASE WHEN source_entry_id = '' THEN NULL ELSE source_entry_id END,
         source_research_id, target_entry_id, target_research_id, target_ref, resolved, created_at
  FROM crossrefs;

DROP TABLE crossrefs;
ALTER TABLE crossrefs_new RENAME TO crossrefs;

CREATE INDEX IF NOT EXISTS idx_crossrefs_source ON crossrefs(source_type, source_id);
CREATE INDEX IF NOT EXISTS idx_crossrefs_source_research ON crossrefs(source_research_id);
CREATE INDEX IF NOT EXISTS idx_crossrefs_target_entry ON crossrefs(target_entry_id);
