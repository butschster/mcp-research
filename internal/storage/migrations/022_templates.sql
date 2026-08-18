-- A template is a kickoff methodology the model reads, not a skeleton the code
-- clones. It carries no sections, no questions and no rows: a name, the criteria
-- an agent matches on, and a markdown body saying what to ask the human and what
-- structure to propose. The model then designs the research itself.
--
-- That is the whole difference from the first draft of this feature, and every
-- absent column below follows from it. There is no `sections`, no `questions`,
-- no placeholder syntax — a template that is prose needs none of them, and the
-- panel that reviewed the idea was unanimous that a pre-committed section list
-- is the thing that damages a research rather than the thing that helps it.
--
-- Against its sibling, migration 021: a template is read once, before a research
-- exists; a skill is read on demand, during the work.

CREATE TABLE IF NOT EXISTS research_templates (
    id TEXT PRIMARY KEY,

    -- NULL = global: shipped with the binary, owned by nobody, visible to
    -- everyone, refreshed at boot. Otherwise the team that wrote it.
    --
    -- Two tiers and no third. A team cannot publish its own template to the
    -- whole server: a template body steers a model, and lending one team's
    -- instructions to another team's kickoff needs a trust story this product
    -- does not have.
    team_id TEXT REFERENCES teams(id) ON DELETE CASCADE,

    -- Author. A record, never consulted for a permission decision.
    user_id TEXT REFERENCES users(id) ON DELETE SET NULL,

    slug TEXT NOT NULL,
    name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',

    -- The matching criteria, and they are not decoration: this is what an agent
    -- chooses on, and it is all it sees before deciding. Every expert shown the
    -- built-in list asked for the negative form as well — knowing when a
    -- methodology is wrong is what stops it being applied to everything.
    when_to_use     TEXT NOT NULL DEFAULT '',
    when_not_to_use TEXT NOT NULL DEFAULT '',

    -- The feature. Markdown: what to ask before proposing anything, what
    -- structure to suggest, what a good entry looks like, when it is finished.
    body TEXT NOT NULL DEFAULT '',

    -- Skill slugs attached to the research this template starts. Slugs only —
    -- a skill body copied in here goes stale the moment it is copied, which is
    -- the reason skills are a separate table rather than a field.
    skills TEXT NOT NULL DEFAULT '[]',

    source     TEXT NOT NULL DEFAULT 'user',  -- 'builtin' | 'user'
    forked_from TEXT,                          -- slug of the built-in this copies
    version    INTEGER NOT NULL DEFAULT 1,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

-- A plain UNIQUE(team_id, slug) does not constrain the global tier at all:
-- SQLite treats NULLs as distinct, so two rows of (NULL, 'literature-review')
-- would both insert and the boot-time refresh would duplicate every built-in on
-- every upgrade. Partial indexes instead, one per tier — the same trap and the
-- same fix as migration 021.
CREATE UNIQUE INDEX IF NOT EXISTS idx_templates_global_slug
    ON research_templates(slug) WHERE team_id IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_templates_team_slug
    ON research_templates(team_id, slug) WHERE team_id IS NOT NULL;

-- Which methodology a research was started from, and which version of it.
--
-- The version matters because built-ins are refreshed from the binary on every
-- boot: without it an upgrade silently changes the methodology under a research
-- already in flight, and nobody could tell afterwards which text was followed.
ALTER TABLE researches ADD COLUMN template_slug TEXT NOT NULL DEFAULT '';
ALTER TABLE researches ADD COLUMN template_version INTEGER NOT NULL DEFAULT 0;
