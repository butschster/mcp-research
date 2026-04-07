-- Roadmaps: visual graphs/learning paths within a research
CREATE TABLE IF NOT EXISTS roadmaps (
    id TEXT PRIMARY KEY,
    code TEXT NOT NULL DEFAULT '',
    research_id TEXT NOT NULL REFERENCES researches(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    statuses TEXT NOT NULL DEFAULT '[]',
    status TEXT NOT NULL DEFAULT 'active',
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_roadmaps_research ON roadmaps(research_id);

CREATE TABLE IF NOT EXISTS roadmap_nodes (
    id TEXT PRIMARY KEY,
    code TEXT NOT NULL DEFAULT '',
    roadmap_id TEXT NOT NULL REFERENCES roadmaps(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    node_type TEXT NOT NULL DEFAULT 'step',
    status TEXT NOT NULL DEFAULT '',
    position_x REAL NOT NULL DEFAULT 0,
    position_y REAL NOT NULL DEFAULT 0,
    parent_id TEXT DEFAULT NULL,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_roadmap_nodes_roadmap ON roadmap_nodes(roadmap_id);
CREATE INDEX IF NOT EXISTS idx_roadmap_nodes_parent ON roadmap_nodes(parent_id);

CREATE TABLE IF NOT EXISTS roadmap_edges (
    id TEXT PRIMARY KEY,
    roadmap_id TEXT NOT NULL REFERENCES roadmaps(id) ON DELETE CASCADE,
    source_node_id TEXT NOT NULL REFERENCES roadmap_nodes(id) ON DELETE CASCADE,
    target_node_id TEXT NOT NULL REFERENCES roadmap_nodes(id) ON DELETE CASCADE,
    label TEXT NOT NULL DEFAULT '',
    edge_type TEXT NOT NULL DEFAULT 'default',
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_roadmap_edges_roadmap ON roadmap_edges(roadmap_id);
CREATE INDEX IF NOT EXISTS idx_roadmap_edges_source ON roadmap_edges(source_node_id);
CREATE INDEX IF NOT EXISTS idx_roadmap_edges_target ON roadmap_edges(target_node_id);
