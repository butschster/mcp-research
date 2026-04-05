-- Add short codes for cross-referencing
-- Researches get globally unique codes: R1, R2, R3
-- Entries get codes scoped to their research: E1, E2, E3

ALTER TABLE researches ADD COLUMN code TEXT NOT NULL DEFAULT '';
ALTER TABLE entries ADD COLUMN code TEXT NOT NULL DEFAULT '';

CREATE UNIQUE INDEX IF NOT EXISTS idx_researches_code ON researches(code) WHERE code != '';
CREATE UNIQUE INDEX IF NOT EXISTS idx_entries_code_research ON entries(research_id, code) WHERE code != '';
