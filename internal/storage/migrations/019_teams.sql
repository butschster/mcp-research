-- Teams own researches; a role in the team is what grants access.
--
-- Before this migration, `researches.user_id` was both "who created it" and
-- "who may touch it". The two are separated here: the column stays as the
-- creator record, and every permission decision moves to `team_members`. Two
-- sources of truth for who may do what is how this class of bug happens.

CREATE TABLE IF NOT EXISTS teams (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    -- A personal team is created with its user and cannot be deleted or left:
    -- it is where a solo user's researches live, and losing it would strand
    -- them. Everything else about it is an ordinary team.
    personal INTEGER NOT NULL DEFAULT 0,
    created_by TEXT DEFAULT NULL REFERENCES users(id) ON DELETE SET NULL,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS team_members (
    team_id TEXT NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role TEXT NOT NULL DEFAULT 'viewer',  -- viewer | editor | owner
    invited_by TEXT DEFAULT NULL REFERENCES users(id) ON DELETE SET NULL,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    PRIMARY KEY (team_id, user_id)
);

-- The list query is "researches of every team I am in", so the lookup runs
-- from the user, not from the team.
CREATE INDEX IF NOT EXISTS idx_team_members_user ON team_members(user_id);

-- An invite is a link, never an email: this product has no mail configuration
-- and adding SMTP would stop it being downloadable-and-run. The inviter passes
-- the link along by whatever means they already use.
CREATE TABLE IF NOT EXISTS team_invites (
    id TEXT PRIMARY KEY,
    team_id TEXT NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    email TEXT NOT NULL DEFAULT '',
    role TEXT NOT NULL DEFAULT 'viewer',
    -- SHA-256, following api_keys: a leaked database must not hand out working
    -- invitations. The token itself is shown exactly once, at creation.
    token_hash TEXT NOT NULL UNIQUE,
    invited_by TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expires_at TEXT NOT NULL,
    accepted_at TEXT,
    accepted_by TEXT DEFAULT NULL REFERENCES users(id) ON DELETE SET NULL,
    revoked_at TEXT,
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);
CREATE INDEX IF NOT EXISTS idx_team_invites_team ON team_invites(team_id);

ALTER TABLE researches ADD COLUMN team_id TEXT DEFAULT NULL REFERENCES teams(id) ON DELETE CASCADE;
CREATE INDEX IF NOT EXISTS idx_researches_team ON researches(team_id);

-- Backfill. Every existing user gets a personal team, so membership is the
-- single source of truth from the first request after the upgrade rather than
-- from the first registration.
--
-- The team id is derived from the user id instead of being random: the
-- migration then reruns to the same result, and a half-applied upgrade cannot
-- leave a user with two personal teams. Teams created by the running app use
-- UUIDs; both forms are only ever compared, never parsed.
INSERT INTO teams (id, name, personal, created_by, created_at, updated_at)
SELECT 'team-' || u.id,
       CASE WHEN trim(u.name) <> '' THEN u.name
            ELSE substr(u.email, 1, instr(u.email, '@') - 1) END,
       1, u.id, u.created_at, u.created_at
FROM users u
WHERE NOT EXISTS (SELECT 1 FROM teams t WHERE t.id = 'team-' || u.id);

INSERT OR IGNORE INTO team_members (team_id, user_id, role, created_at)
SELECT 'team-' || u.id, u.id, 'owner', u.created_at FROM users u;

UPDATE researches SET team_id = 'team-' || user_id
WHERE team_id IS NULL AND user_id IS NOT NULL AND user_id <> '';

-- Researches with no owner at all: the local single-user mode, where auth is
-- off and nobody is ever in the request context. They go to one team with no
-- members, which every unauthenticated caller is allowed to reach — exactly
-- the access that mode has always had. The first user to register claims both
-- the researches and this team's contents, as they already claimed orphans.
INSERT INTO teams (id, name, personal, created_at, updated_at)
SELECT 'team-local', 'Local', 0, datetime('now'), datetime('now')
WHERE NOT EXISTS (SELECT 1 FROM teams WHERE id = 'team-local');

UPDATE researches SET team_id = 'team-local' WHERE team_id IS NULL;
