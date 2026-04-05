-- Generalize crossrefs to support sources from entries, questions, and tasks.
-- SQLite doesn't support DROP COLUMN well, so we add new columns alongside.

ALTER TABLE crossrefs ADD COLUMN source_type TEXT NOT NULL DEFAULT 'entry';
ALTER TABLE crossrefs ADD COLUMN source_id TEXT NOT NULL DEFAULT '';

-- Backfill source_id from source_entry_id for existing records
UPDATE crossrefs SET source_id = source_entry_id WHERE source_id = '';

CREATE INDEX IF NOT EXISTS idx_crossrefs_source ON crossrefs(source_type, source_id);
