-- 000009: Schema improvements (FK constraint, indexes, review constraint)

-- Clean up orphaned audit records before adding FK constraint
DELETE FROM time_entry_audit
WHERE time_entry_id NOT IN (SELECT id FROM time_entries);

-- FK constraint: time_entry_audit → time_entries
ALTER TABLE time_entry_audit
ADD CONSTRAINT fk_audit_time_entry
FOREIGN KEY (time_entry_id) REFERENCES time_entries(id) ON DELETE CASCADE;

-- Index for frequent absence queries by user + status
CREATE INDEX IF NOT EXISTS idx_absences_user_status ON absences(user_id, status);

-- Index for audit queries by timestamp
CREATE INDEX IF NOT EXISTS idx_time_entry_audit_changed_at ON time_entry_audit(changed_at DESC);

-- Ensure reviewed_by and reviewed_at are always set together
ALTER TABLE absences ADD CONSTRAINT chk_review_both_or_neither
CHECK (
  (reviewed_by IS NULL AND reviewed_at IS NULL)
  OR (reviewed_by IS NOT NULL AND reviewed_at IS NOT NULL)
);

UPDATE schema_info SET value = '000009' WHERE key = 'version';
