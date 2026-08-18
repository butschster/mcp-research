-- Roadmap view modes: the same node-edge graph, laid out along an axis.
--
-- Until now a roadmap was a free-form graph and a node carried only position_x /
-- position_y. These four columns add the two axes a "full" roadmap wants — a
-- stage (a phase column) and a date (a point on a timeline) — plus, on the
-- roadmap itself, the ordered list of stage names and which view it opens in.
--
-- Every column is additive and defaulted, so an existing roadmap keeps behaving
-- exactly as before: view defaults to 'graph', stages to the empty list, and a
-- node's stage/node_date to empty. The graph layout reads none of them.

-- Which stage (column) a node sits in, and its point on a timeline.
--
-- `stage` is matched by string against roadmaps.stages the same way `status` is
-- matched against roadmaps.statuses — deliberately the same shape, so a node
-- whose stage is not in the list falls into a leftover column rather than
-- erroring, exactly as an unknown status already does. `node_date` is an ISO
-- 'YYYY-MM-DD' string or empty; empty means undated, and the timeline view sets
-- those aside rather than guessing a position.
ALTER TABLE roadmap_nodes ADD COLUMN stage TEXT NOT NULL DEFAULT '';
ALTER TABLE roadmap_nodes ADD COLUMN node_date TEXT NOT NULL DEFAULT '';

-- The ordered stage vocabulary, and the view the roadmap opens in.
--
-- `stages` is a JSON array of names, mirroring `statuses`: the order here is the
-- column order, and a name with no nodes is a legitimately empty column. `view`
-- is one of 'graph' | 'stages' | 'timeline'; the service validates it, the
-- frontend toggle overrides it locally. It is the roadmap's default, not a lock.
ALTER TABLE roadmaps ADD COLUMN stages TEXT NOT NULL DEFAULT '[]';
ALTER TABLE roadmaps ADD COLUMN view TEXT NOT NULL DEFAULT 'graph';
