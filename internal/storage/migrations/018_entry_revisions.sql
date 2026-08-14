-- Every version an entry ever had.
--
-- An entry is written by a language model across many sessions, and Update
-- overwrites content in place. Without this table the third session can quietly
-- make a document worse and nothing can say what it looked like before, or what
-- a given session actually changed.
--
-- Append-only. Restoring an old revision writes a NEW revision whose content
-- equals it, so history is never rewritten and a revision number is never reused.
--
-- Not to be confused with the document `rev` a blocks entry carries: that is a
-- content hash used for optimistic concurrency on a patch (service.DocumentRev).
-- A revision here is a numbered snapshot in time. Different jobs, similar words.
--
-- session_id is the session that was active when the write happened. It is what
-- makes "what did session SS3 do to this research" answerable, which is the more
-- interesting half of this feature — per-entry undo is the smaller half.
CREATE TABLE IF NOT EXISTS entry_revisions (
    id          TEXT PRIMARY KEY,
    entry_id    TEXT    NOT NULL REFERENCES entries(id)    ON DELETE CASCADE,
    research_id TEXT    NOT NULL REFERENCES researches(id) ON DELETE CASCADE,
    revision    INTEGER NOT NULL,
    title       TEXT    NOT NULL DEFAULT '',
    description TEXT    NOT NULL DEFAULT '',
    content     TEXT    NOT NULL DEFAULT '',
    entry_type  TEXT    NOT NULL DEFAULT 'markdown',
    status      TEXT    NOT NULL DEFAULT '',
    tags        TEXT    NOT NULL DEFAULT '[]',
    author_kind TEXT    NOT NULL DEFAULT 'agent',
    session_id  TEXT,
    user_id     TEXT,
    summary     TEXT    NOT NULL DEFAULT '',
    created_at  TEXT    NOT NULL DEFAULT (datetime('now')),
    UNIQUE(entry_id, revision)
);

-- The history query: newest first, per entry.
CREATE INDEX IF NOT EXISTS idx_entry_revisions_entry ON entry_revisions(entry_id, revision DESC);

-- "What did this session change" — one query over the whole research.
CREATE INDEX IF NOT EXISTS idx_entry_revisions_session ON entry_revisions(session_id);

CREATE INDEX IF NOT EXISTS idx_entry_revisions_research ON entry_revisions(research_id, created_at);

-- Backfill: every existing entry gets revision 1 holding what it looks like now.
--
-- History starts complete rather than starting empty. The alternative — writing
-- the first revision on the next update — would make an entry's oldest recorded
-- state the one produced by the change you were trying to inspect, which is the
-- opposite of what a history is for.
--
-- author_kind is 'agent' because that is what wrote essentially everything that
-- predates this migration, and claiming to know better would be a guess.
-- created_at is the entry's updated_at, not now: the snapshot describes when the
-- content was last written, not when this table learned about it.
INSERT INTO entry_revisions (id, entry_id, research_id, revision, title, description, content, entry_type, status, tags, author_kind, session_id, summary, created_at)
SELECT lower(hex(randomblob(16))),
       e.id,
       e.research_id,
       1,
       e.title,
       e.description,
       e.content,
       e.entry_type,
       e.status,
       e.tags,
       'agent',
       e.session_id,
       '',
       e.updated_at
  FROM entries e;
