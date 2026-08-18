-- A section declares what its documents record, and an entry carries the values.
--
-- The motivating case is a real one: a research holding eighteen specifications
-- in one section, each opening with five lines of hand-typed prose metadata —
-- status, date, which services produce and consume the thing. None of it was
-- queryable, and the prose "Status: draft for review" disagreed with the stored
-- status of `active` while the human believed the prose.
--
-- Two decisions are encoded in the shape below and every later change should be
-- checked against them:
--
--   The vocabulary is closed. Only a key the section declares may be written.
--   That is what makes the section table view possible — a column per field
--   over the section's entries — and it is what lets a section that declares
--   nothing behave exactly as it did before this migration.
--
--   The spec is a lens; values outlive it. A schema edit never rewrites a
--   document. Removing a field leaves its values in place, adding one makes
--   existing documents incomplete rather than invalid, and completeness is
--   computed on read and never stored — a stored flag goes stale the moment
--   the spec changes.

-- [{key, label, type, required, repeated, options[], help}]. JSON and not rows
-- because the spec is read whole, written whole and validated whole: every
-- entry write is checked against exactly one version of it. The precedent is
-- roadmap node metadata — take its storage shape, and take its total absence
-- of validation as the warning.
ALTER TABLE sections ADD COLUMN field_spec TEXT NOT NULL DEFAULT '[]';

-- Bumped on every change to field_spec. Sections are overwritten in place —
-- id, name, display_name, description, status, position and two timestamps,
-- nothing else — so without this, "which rule was this document held to" has
-- no answer, and lazy top-up becomes a silent rewrite of history: a document
-- written under the old spec would be indistinguishable from one violating the
-- new. Same reasoning as researches.template_version.
ALTER TABLE sections ADD COLUMN spec_version INTEGER NOT NULL DEFAULT 0;

-- Values keyed by the field keys the section declares. A key that was declared
-- and later removed keeps its value here: deleting a field is a decision about
-- future collection, not a verdict on what was already recorded.
ALTER TABLE entries ADD COLUMN metadata TEXT NOT NULL DEFAULT '{}';

-- The spec version this entry's metadata was last validated against.
ALTER TABLE entries ADD COLUMN spec_version INTEGER NOT NULL DEFAULT 0;

-- The revision table is a fixed column set, so without these two a metadata
-- edit would leave no trace in history at all. Worse than invisible: an edit
-- that touches only metadata is judged a no-op by EntryRevision.SameContent
-- and no revision is written, so the change disappears rather than merely
-- going unrecorded. That comparison is extended in the same change as this.
ALTER TABLE entry_revisions ADD COLUMN metadata TEXT NOT NULL DEFAULT '{}';
ALTER TABLE entry_revisions ADD COLUMN spec_version INTEGER NOT NULL DEFAULT 0;

-- Revision 1 of every pre-existing entry claims '{}' by the column default,
-- which here means "predates the feature" rather than "was empty" — an honest
-- statement, and the only one available. No backfill can invent metadata that
-- was never collected.
