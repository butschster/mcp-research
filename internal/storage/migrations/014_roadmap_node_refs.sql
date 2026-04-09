-- +migrate Up
ALTER TABLE roadmap_nodes ADD COLUMN ref_type TEXT;
ALTER TABLE roadmap_nodes ADD COLUMN ref_id TEXT;
ALTER TABLE roadmap_nodes ADD COLUMN metadata TEXT;

CREATE INDEX idx_roadmap_nodes_ref ON roadmap_nodes(ref_type, ref_id) WHERE ref_type IS NOT NULL;

-- +migrate Down
DROP INDEX IF EXISTS idx_roadmap_nodes_ref;
-- SQLite does not support DROP COLUMN; recreate table if needed.
