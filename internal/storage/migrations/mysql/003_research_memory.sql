CREATE TABLE research_memory (
    id VARCHAR(255) PRIMARY KEY,
    research_id VARCHAR(255) NOT NULL,
    text LONGTEXT NOT NULL,
    created_at TEXT,
    session_id VARCHAR(255),
    author VARCHAR(20) NOT NULL,
    version INTEGER NOT NULL DEFAULT 1,
    position BIGINT NOT NULL,
    FOREIGN KEY (research_id) REFERENCES researches(id) ON DELETE CASCADE,
    FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE SET NULL,
    INDEX idx_research_memory_research (research_id, position, id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;
