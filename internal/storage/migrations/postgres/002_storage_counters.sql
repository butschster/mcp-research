CREATE TABLE storage_counters (
    scope_key VARCHAR(512) PRIMARY KEY,
    value BIGINT NOT NULL DEFAULT 0
);
