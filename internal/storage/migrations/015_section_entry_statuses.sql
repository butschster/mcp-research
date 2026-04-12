-- Section-level allowed entry statuses (JSON array of strings).
-- NULL means "use defaults" (draft, active, completed, archived).
ALTER TABLE sections ADD COLUMN allowed_entry_statuses TEXT DEFAULT NULL;
