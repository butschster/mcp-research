-- Add entry type: 'markdown' (default) or 'artifact' (self-contained HTML document
-- rendered in a sandboxed iframe instead of as markdown).
ALTER TABLE entries ADD COLUMN entry_type TEXT NOT NULL DEFAULT 'markdown';
CREATE INDEX IF NOT EXISTS idx_entries_type ON entries(entry_type);
