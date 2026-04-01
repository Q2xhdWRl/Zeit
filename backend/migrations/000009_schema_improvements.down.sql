-- 000009 down: revert schema improvements

ALTER TABLE absences DROP CONSTRAINT IF EXISTS chk_review_both_or_neither;
DROP INDEX IF EXISTS idx_time_entry_audit_changed_at;
DROP INDEX IF EXISTS idx_absences_user_status;
ALTER TABLE time_entry_audit DROP CONSTRAINT IF EXISTS fk_audit_time_entry;

UPDATE schema_info SET value = '000008' WHERE key = 'version';
