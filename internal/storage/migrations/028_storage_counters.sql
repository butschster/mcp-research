-- Persistent short-code allocation, seeded lazily from existing records.
CREATE TABLE storage_counters (
    scope_key VARCHAR(512) PRIMARY KEY,
    value BIGINT NOT NULL DEFAULT 0
);
