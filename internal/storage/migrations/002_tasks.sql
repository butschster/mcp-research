CREATE TABLE IF NOT EXISTS tasks (
    id TEXT PRIMARY KEY,
    research_id TEXT NOT NULL REFERENCES researches(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'pending',
    priority TEXT NOT NULL DEFAULT 'medium',
    result TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now')),
    completed_at TEXT DEFAULT NULL
);

CREATE INDEX IF NOT EXISTS idx_tasks_research ON tasks(research_id);
CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status);
