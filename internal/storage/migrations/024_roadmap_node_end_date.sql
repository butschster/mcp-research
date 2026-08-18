-- A roadmap node can now span a range on the timeline, not only sit at a point.
--
-- `node_date` is the start; this adds the optional end. A node with both renders
-- as a bar from start to end (a Gantt bar); a node with only `node_date` stays a
-- point exactly as before. Empty means "no end" — a point, not a zero-length bar.
--
-- Additive and defaulted, so every existing node keeps behaving as a point.
ALTER TABLE roadmap_nodes ADD COLUMN node_end_date TEXT NOT NULL DEFAULT '';
