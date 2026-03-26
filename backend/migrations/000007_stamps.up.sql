CREATE TABLE active_stamps (
    user_id       UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    started_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    break_start   TIMESTAMPTZ NULL,
    break_minutes INT         NOT NULL DEFAULT 0,
    project_id    UUID        REFERENCES projects(id) ON DELETE SET NULL NULL,
    description   TEXT        NOT NULL DEFAULT ''
);
