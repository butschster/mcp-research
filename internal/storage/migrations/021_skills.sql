-- A skill is a methodology document the agent opens when it decides it needs
-- it. It exists because `instruction` is one always-loaded field and
-- methodology is the wrong size for that: a jobs-to-be-done interview protocol
-- or a source-grading rubric is 800-2500 tokens, a research needs one or two of
-- them at the moment of use, and `instruction` is a per-row string that can
-- never be reused by the next research.
--
-- The division of labour, which every later change should be checked against:
-- `instruction` says what *this research* is; a skill says how a *kind of work*
-- is done.

CREATE TABLE IF NOT EXISTS skills (
    id TEXT PRIMARY KEY,

    -- Ownership is a three-way choice and exactly one of these is set.
    --
    --   both NULL  -> built-in: ships embedded in the binary, refreshed at
    --                 boot, never mutated in place. Editing one forks a team
    --                 copy, so an upgrade can never overwrite somebody's edits.
    --   team_id    -> team library: reusable across that team's researches.
    --   research_id-> research-private: belongs to one research and is never
    --                 offered in the library. This is where a rule that applies
    --                 only to this research lives, and it is what the
    --                 `instruction` migration writes into.
    --
    -- The tier column carries the same fact denormalised, because the index the
    -- agent sees is sorted by it on every read and a CASE over two nullable
    -- columns in that hot path is worse than one indexed string.
    team_id     TEXT REFERENCES teams(id)      ON DELETE CASCADE,
    research_id TEXT REFERENCES researches(id) ON DELETE CASCADE,

    -- Author. A record, never consulted for a permission decision — the same
    -- rule researches.user_id follows. Nullable because `auth_enabled: false`
    -- has no users at all.
    user_id TEXT REFERENCES users(id) ON DELETE SET NULL,

    slug TEXT NOT NULL,
    name TEXT NOT NULL,

    -- The only field that is always in the agent's context. It must say *when
    -- to load*, not what the body contains: the agent matches situations, and a
    -- description that describes gets a skill that never loads. Capped at 200
    -- characters in the service, with a field error rather than a silent
    -- truncation — a truncated trigger line is a skill that stops working.
    description TEXT NOT NULL DEFAULT '',

    -- Markdown. Never returned by research_get; only by skill_load and the
    -- single-skill read route.
    body TEXT NOT NULL DEFAULT '',

    tier TEXT NOT NULL DEFAULT 'team',   -- 'builtin' | 'team' | 'private'

    -- Product skills — how to manage a research, how to write entries, what an
    -- artifact is — go on nearly every research. Counting them would consume
    -- four of the six-skill budget before anybody chooses anything, and a
    -- template attaching two methodology skills would then breach the cap at
    -- creation every time. So they are ambient: always in the index, never
    -- counted, never detachable.
    ambient INTEGER NOT NULL DEFAULT 0,

    -- Slug of the built-in this row was forked from, so the UI can say "Team ·
    -- forked" and a future change can diff against the parent.
    forked_from TEXT,

    -- The instruction migration has to invent a name and a description for the
    -- text it moves. A machine-written trigger line is exactly the failure the
    -- description rule warns about, so flag what was generated: without this
    -- there is no way to surface the day-one work that actually matters.
    needs_trigger INTEGER NOT NULL DEFAULT 0,

    version    INTEGER NOT NULL DEFAULT 1,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now')),

    -- At most one owner. A row carrying both would be visible through the
    -- research arm of the scope check and invisible to the library listing,
    -- which filters on research_id IS NULL — an inconsistency that would be
    -- very hard to read back from a bug report.
    CHECK (team_id IS NULL OR research_id IS NULL)
);

-- A plain UNIQUE(team_id, slug) would not constrain built-ins at all: SQLite
-- treats NULLs as distinct in a UNIQUE constraint, so two rows of (NULL,
-- 'source-triage') both insert and the boot-time refresh would duplicate every
-- built-in on every upgrade. Three partial indexes instead, one per tier.
CREATE UNIQUE INDEX IF NOT EXISTS idx_skills_builtin_slug
    ON skills(slug) WHERE team_id IS NULL AND research_id IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_skills_team_slug
    ON skills(team_id, slug) WHERE team_id IS NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_skills_private_slug
    ON skills(research_id, slug) WHERE research_id IS NOT NULL;

-- What a research follows. Separate from the skill so the same team skill can
-- be attached to many researches without being copied — which is the whole
-- reason skills exist rather than a bigger instruction field.
CREATE TABLE IF NOT EXISTS research_skills (
    research_id TEXT NOT NULL REFERENCES researches(id) ON DELETE CASCADE,
    skill_id    TEXT NOT NULL REFERENCES skills(id)     ON DELETE CASCADE,
    -- Whether a template attached this rather than a human choosing it. The
    -- difference is worth showing in the UI and is free to record here.
    via_template INTEGER NOT NULL DEFAULT 0,
    attached_at  TEXT NOT NULL DEFAULT (datetime('now')),
    PRIMARY KEY (research_id, skill_id)
);

CREATE INDEX IF NOT EXISTS idx_research_skills_skill ON research_skills(skill_id);
