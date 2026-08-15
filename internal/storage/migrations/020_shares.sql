-- A share link is an unguessable, revocable, read-only capability over one
-- research. It is not an account, not an invitation and not a role: nobody logs
-- in, nothing is granted to a person, and the link names no user at all.
--
-- That is why this table sits beside team_members rather than inside it. A team
-- role is "who you are"; a share is "what this URL may see". Merging the two
-- would put an anonymous reader into the membership tables, which every query
-- in the product treats as a person with an account.

CREATE TABLE IF NOT EXISTS shares (
    id TEXT PRIMARY KEY,
    -- SHA-256, following api_keys and team_invites: a leaked database must not
    -- hand out working links. The token itself is shown exactly once, at
    -- creation, and cannot be recovered afterwards.
    token_hash TEXT NOT NULL UNIQUE,
    research_id TEXT NOT NULL REFERENCES researches(id) ON DELETE CASCADE,
    -- Who created it. A record, never a permission: the link keeps working
    -- after its creator signs out, and stops working when the row is revoked
    -- or the research is deleted. It is deleted with the user because a share
    -- nobody can find in any UI is a capability with no way to withdraw it.
    --
    -- Nullable, because `auth_enabled: false` has no users at all — the local
    -- single-binary case, where a share is still a perfectly ordinary thing to
    -- want. NOT NULL here made the whole feature unusable in that mode.
    user_id TEXT DEFAULT NULL REFERENCES users(id) ON DELETE CASCADE,
    -- scope/target_id exist for narrower shares (one session, one entry, one
    -- roadmap). Only 'research' is issued today; the columns are here so the
    -- narrower forms do not need a second migration, and the service refuses
    -- anything else rather than half-enforcing it.
    scope TEXT NOT NULL DEFAULT 'research',
    target_id TEXT NOT NULL DEFAULT '',
    label TEXT NOT NULL DEFAULT '',
    -- JSON flags: sessions, tasks, roadmaps, export. Stored as one column
    -- because they are read and written together and never queried on.
    include TEXT NOT NULL DEFAULT '{}',
    -- bcrypt, optional. A password is a second factor on the URL, not an
    -- identity — there is still nobody to be.
    password_hash TEXT NOT NULL DEFAULT '',
    expires_at TEXT,
    revoked_at TEXT,
    last_seen_at TEXT,
    view_count INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_shares_research ON shares(research_id);
