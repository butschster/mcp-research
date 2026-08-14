-- Blocks of a `blocks` entry become rows.
--
-- The row is the source of truth for a block's identity and for state a human
-- puts on it (checklist ticks), because that is the thing an agent rewriting the
-- document must not be able to destroy by accident.
--
-- entries.content stays, as a serialization of these rows written in the same
-- transaction. It is derived, never authoritative: search, both exports, the
-- portable format and the frontend all read it, and none of them should have to
-- know that blocks are rows.
--
-- research_id is denormalized from the entry so a block can be filtered by owner
-- in one query. It is filled from the entry row, never from a request, and it
-- never authorizes anything on its own — access is decided by the entry, exactly
-- as everywhere else. If entries ever become movable between researches, this
-- column has to move with them.
CREATE TABLE IF NOT EXISTS entry_blocks (
    entry_id    TEXT    NOT NULL REFERENCES entries(id)    ON DELETE CASCADE,
    research_id TEXT    NOT NULL REFERENCES researches(id) ON DELETE CASCADE,
    block_id    TEXT    NOT NULL,
    position    INTEGER NOT NULL,
    type        TEXT    NOT NULL,
    data        TEXT    NOT NULL DEFAULT '{}',
    state       TEXT    NOT NULL DEFAULT '',
    PRIMARY KEY (entry_id, block_id)
);

-- Render order. Every read of a document is this query.
CREATE INDEX IF NOT EXISTS idx_entry_blocks_order ON entry_blocks(entry_id, position);

-- Owner-scoped block queries: "find the block", rather than "find the entry".
CREATE INDEX IF NOT EXISTS idx_entry_blocks_research ON entry_blocks(research_id, type);

-- Backfill. Documents written before this migration live only in entries.content,
-- and a patch reads rows: without this, the first insert into such a document
-- would replace it with the one block that was inserted. Every stored document
-- already carries block ids (the normalizer fills them), so identity survives.
INSERT INTO entry_blocks (entry_id, research_id, block_id, position, type, data, state)
SELECT e.id,
       e.research_id,
       COALESCE(json_extract(b.value, '$.id'), lower(hex(randomblob(4)))),
       b.key,
       COALESCE(json_extract(b.value, '$.type'), 'paragraph'),
       COALESCE(json_extract(b.value, '$.data'), '{}'),
       ''
  FROM entries e, json_each(json_extract(e.content, '$.blocks')) b
 WHERE e.entry_type = 'blocks'
   AND json_valid(e.content)
   AND json_extract(e.content, '$.blocks') IS NOT NULL;
