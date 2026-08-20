-- A person reads a document the agent wrote, reaches a sentence they do not
-- believe, and marks it.
--
-- The border with everything else that records trust in this product, and the
-- thing every later change should be checked against: an annotation is a
-- **request for work addressed to a place in the text**, not a record of what we
-- know. Evidence says what a block rests on. An annotation says what is wrong
-- with a sentence, and expects somebody to do something about it.
--
-- Which is why the agent does not write these. An agent marking its own
-- uncertainty is model-kind evidence — automatic, complete and free. A
-- quota-limited agent-authored annotation would be the same concept modelled a
-- second time, worse.
CREATE TABLE IF NOT EXISTS annotations (
    id                TEXT PRIMARY KEY,
    -- A1, A2 — per research, like every other short code.
    code              TEXT NOT NULL,
    -- Denormalized from the entry so the queue is one query. Filled from the
    -- entry row, never from a request, and it authorizes nothing on its own:
    -- access is decided by the entry, exactly as everywhere else.
    research_id       TEXT NOT NULL REFERENCES researches(id) ON DELETE CASCADE,
    entry_id          TEXT NOT NULL REFERENCES entries(id)    ON DELETE CASCADE,

    -- The block is the address, the quote is the proof.
    --
    -- Empty for a markdown entry, which has no addressable blocks. That is not a
    -- degraded case to hide: the UI says the precision is lower rather than
    -- implying a block-level accuracy it does not have.
    block_id          TEXT NOT NULL DEFAULT '',
    -- What the human actually selected, plus enough on either side to tell two
    -- identical sentences apart. There are deliberately NO character offsets:
    -- an offset is a position in a string the agent is about to regenerate.
    -- The highlight on screen is computed on read as an occurrence of
    -- quote_exact, and is never stored as truth.
    quote_exact       TEXT NOT NULL,
    quote_prefix      TEXT NOT NULL DEFAULT '',
    quote_suffix      TEXT NOT NULL DEFAULT '',
    -- entry_revisions.revision as it stood when the mark was made. This is what
    -- makes an orphan investigable: entry_diff from here to now shows exactly
    -- what happened to the text somebody doubted.
    anchored_revision INTEGER NOT NULL DEFAULT 0,

    -- verify | dig | disagree. Closed vocabulary, because each one is a
    -- different program for the agent — a free-form tag would make "go work my
    -- annotations" a mood rather than an instruction.
    kind              TEXT NOT NULL,
    body              TEXT NOT NULL DEFAULT '',

    -- Human by definition today. The column exists anyway: the day an agent is
    -- allowed to mark something, the two must not share a lane, because a
    -- human mark is a scarce signal and an agent's is free.
    author_kind       TEXT NOT NULL DEFAULT 'human',
    user_id           TEXT,

    -- open → answered → closed | dismissed.
    --
    -- No priority, no blocked, no deferred, no result. The moment those appear
    -- this has become a second tasks table. An annotation that needs scheduling
    -- is promoted to a task explicitly — task_id below — and one task may close
    -- five of these.
    --
    -- The agent may reach `answered` and no further. `closed` and `dismissed`
    -- are the human's, always: an agent that can accept its own work is a
    -- self-confirmation machine with extra steps.
    status            TEXT NOT NULL DEFAULT 'open',
    resolution        TEXT NOT NULL DEFAULT '',
    resolved_revision INTEGER,
    -- The pass that answered it. A pass runs inside a session, so "what did
    -- session SS4 actually settle" is one query.
    session_id        TEXT,
    task_id           TEXT,
    -- Rejections, in full: [{reason, revision, at}].
    --
    -- A count alone cannot carry what the lifecycle asks for. The agent must
    -- read the reasons its previous answers were refused, plural — and there is
    -- exactly one text column otherwise, `resolution`, which holds the answer
    -- being refused. Writing the reason there would erase the thing being
    -- rejected, leaving the next attempt with neither.
    rejections        TEXT    NOT NULL DEFAULT '[]',
    attempts          INTEGER NOT NULL DEFAULT 0,

    created_at        TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at        TEXT NOT NULL DEFAULT (datetime('now')),
    answered_at       TEXT,
    closed_at         TEXT
);

-- The queue: "what is open in this research", the one read the agent makes.
CREATE INDEX IF NOT EXISTS idx_annotations_queue ON annotations(research_id, status);

-- "What is marked in this document" — the read every entry page makes.
CREATE INDEX IF NOT EXISTS idx_annotations_entry ON annotations(entry_id, block_id);
