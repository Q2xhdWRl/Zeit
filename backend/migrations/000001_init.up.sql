-- 000001_init.up.sql
-- Initial schema: extensions and base setup

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "citext";

-- Metadata table for tracking migration state
CREATE TABLE IF NOT EXISTS schema_info (
    key   TEXT PRIMARY KEY,
    value TEXT NOT NULL
);

INSERT INTO schema_info (key, value)
VALUES ('version', '000001')
ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value;
