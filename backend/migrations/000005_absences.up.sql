-- =============================================================
-- Phase 5: Abwesenheitsverwaltung (Absence Management)
-- =============================================================

-- ── Absence Types ──

CREATE TABLE absence_types (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name TEXT NOT NULL UNIQUE,
    color TEXT NOT NULL DEFAULT '#6b7280',
    requires_approval BOOLEAN NOT NULL DEFAULT true,
    counts_as_work BOOLEAN NOT NULL DEFAULT false,
    is_active BOOLEAN NOT NULL DEFAULT true,
    sort_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Default absence types
INSERT INTO absence_types (name, color, requires_approval, counts_as_work, sort_order) VALUES
    ('Urlaub',        '#22c55e', true,  false, 1),
    ('Krankheit',     '#ef4444', false, false, 2),
    ('Homeoffice',    '#3b82f6', false, true,  3),
    ('Sonderurlaub',  '#a855f7', true,  false, 4),
    ('Fortbildung',   '#f59e0b', true,  true,  5),
    ('Elternzeit',    '#ec4899', true,  false, 6),
    ('Unbezahlt',     '#6b7280', true,  false, 7);

-- ── Absences ──

CREATE TABLE absences (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    absence_type_id UUID NOT NULL REFERENCES absence_types(id),
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    note TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'rejected', 'cancelled')),
    reviewed_by UUID REFERENCES users(id),
    reviewed_at TIMESTAMPTZ,
    review_note TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_end_after_start CHECK (end_date >= start_date)
);

CREATE INDEX idx_absences_user_id ON absences(user_id);
CREATE INDEX idx_absences_dates ON absences(start_date, end_date);
CREATE INDEX idx_absences_status ON absences(status);
CREATE INDEX idx_absences_type ON absences(absence_type_id);

-- ── Vacation Entitlements ──

CREATE TABLE vacation_entitlements (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    year INT NOT NULL,
    total_days INT NOT NULL DEFAULT 30,
    carry_over_days INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_entitlement_user_year UNIQUE (user_id, year),
    CONSTRAINT chk_total_days_positive CHECK (total_days >= 0),
    CONSTRAINT chk_carry_over_non_negative CHECK (carry_over_days >= 0)
);

CREATE INDEX idx_vacation_entitlements_user_year ON vacation_entitlements(user_id, year);
