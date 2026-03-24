-- 000001_init.down.sql
-- Rollback: remove extensions and metadata

DROP TABLE IF EXISTS schema_info;
DROP EXTENSION IF EXISTS "citext";
DROP EXTENSION IF EXISTS "uuid-ossp";
