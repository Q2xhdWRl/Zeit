-- =============================================================
-- Phase 6: Arbeitszeitmodell & Ueberstunden
-- =============================================================

-- ── Work Schedules (Soll-Stunden pro User) ──

CREATE TABLE work_schedules (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    valid_from DATE NOT NULL,
    weekly_hours NUMERIC(5,2) NOT NULL DEFAULT 40.00,
    monday_hours NUMERIC(4,2) NOT NULL DEFAULT 8.00,
    tuesday_hours NUMERIC(4,2) NOT NULL DEFAULT 8.00,
    wednesday_hours NUMERIC(4,2) NOT NULL DEFAULT 8.00,
    thursday_hours NUMERIC(4,2) NOT NULL DEFAULT 8.00,
    friday_hours NUMERIC(4,2) NOT NULL DEFAULT 8.00,
    saturday_hours NUMERIC(4,2) NOT NULL DEFAULT 0.00,
    sunday_hours NUMERIC(4,2) NOT NULL DEFAULT 0.00,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_weekly_hours_positive CHECK (weekly_hours >= 0),
    CONSTRAINT uq_schedule_user_from UNIQUE (user_id, valid_from)
);

CREATE INDEX idx_work_schedules_user ON work_schedules(user_id, valid_from DESC);
