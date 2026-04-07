-- Add roadmap and node targets to cross-references.
-- Supports [[RM1]] and [[RM1:N3]] patterns.

ALTER TABLE crossrefs ADD COLUMN target_roadmap_id TEXT DEFAULT NULL;
ALTER TABLE crossrefs ADD COLUMN target_node_id TEXT DEFAULT NULL;

CREATE INDEX IF NOT EXISTS idx_crossrefs_target_roadmap ON crossrefs(target_roadmap_id);
