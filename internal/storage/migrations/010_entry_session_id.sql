-- Add optional session_id to entries (links entry to the session during which it was created)
ALTER TABLE entries ADD COLUMN session_id TEXT DEFAULT NULL REFERENCES sessions(id) ON DELETE SET NULL;
CREATE INDEX IF NOT EXISTS idx_entries_session ON entries(session_id);
