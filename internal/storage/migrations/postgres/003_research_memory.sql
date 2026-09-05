CREATE TABLE research_memory (
    id TEXT PRIMARY KEY,
    research_id TEXT NOT NULL REFERENCES researches(id) ON DELETE CASCADE,
    text TEXT NOT NULL,
    created_at TEXT,
    session_id TEXT REFERENCES sessions(id) ON DELETE SET NULL,
    author TEXT NOT NULL,
    version INTEGER NOT NULL DEFAULT 1,
    position BIGINT NOT NULL
);
CREATE INDEX idx_research_memory_research ON research_memory(research_id, position, id);
