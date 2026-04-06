-- External links extracted from markdown content
CREATE TABLE IF NOT EXISTS external_links (
    id TEXT PRIMARY KEY,
    source_type TEXT NOT NULL DEFAULT 'entry',
    source_id TEXT NOT NULL,
    research_id TEXT NOT NULL REFERENCES researches(id) ON DELETE CASCADE,
    url TEXT NOT NULL,
    title TEXT NOT NULL DEFAULT '',
    domain TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_external_links_research ON external_links(research_id);
CREATE INDEX IF NOT EXISTS idx_external_links_source ON external_links(source_type, source_id);
CREATE INDEX IF NOT EXISTS idx_external_links_domain ON external_links(domain);
