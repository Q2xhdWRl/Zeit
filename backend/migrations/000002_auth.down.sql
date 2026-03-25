-- 000002_auth.down.sql
-- Rollback authentication tables

DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS users;

UPDATE schema_info SET value = '000001' WHERE key = 'version';
