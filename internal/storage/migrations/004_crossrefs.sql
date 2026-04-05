-- Cross-references extracted from entry content
-- Rebuilt on entry create/update by parsing [[...]] patterns

CREATE TABLE IF NOT EXISTS crossrefs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    source_entry_id TEXT NOT NULL REFERENCES entries(id) ON DELETE CASCADE,
    source_research_id TEXT NOT NULL REFERENCES researches(id) ON DELETE CASCADE,
    target_entry_id TEXT DEFAULT NULL REFERENCES entries(id) ON DELETE CASCADE,
    target_research_id TEXT DEFAULT NULL REFERENCES researches(id) ON DELETE CASCADE,
    target_ref TEXT NOT NULL,
    resolved INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_crossrefs_source_entry ON crossrefs(source_entry_id);
CREATE INDEX IF NOT EXISTS idx_crossrefs_source_research ON crossrefs(source_research_id);
CREATE INDEX IF NOT EXISTS idx_crossrefs_target_entry ON crossrefs(target_entry_id);
CREATE INDEX IF NOT EXISTS idx_crossrefs_target_research ON crossrefs(target_research_id);
