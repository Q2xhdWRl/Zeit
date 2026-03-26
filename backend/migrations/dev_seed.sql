-- =============================================================
-- DEV SEED DATA — Nur fuer lokale Entwicklung!
-- Erstellt Testuser, Sessions, Teams, Projekte und Zeiteintraege.
--
-- Ausfuehren:
--   psql -U zeit_app -d zeiterfassung -f backend/migrations/dev_seed.sql
--
-- Session-Tokens fuer curl:
--   Admin:      dev-admin-token
--   Teamleiter: dev-leader-token
--   Benutzer:   dev-user-token
--
-- Beispiel:
--   curl -b "zeit_session=dev-admin-token" http://localhost:8080/api/time-entries
-- =============================================================

BEGIN;

-- ── Users ──

INSERT INTO users (id, email, display_name, global_role, is_active)
VALUES
  ('a0000000-0000-0000-0000-000000000001', 'admin@newa.test', 'Anna Admin', 'admin', true),
  ('a0000000-0000-0000-0000-000000000002', 'leader@newa.test', 'Lars Leiter', 'team_leader', true),
  ('a0000000-0000-0000-0000-000000000003', 'user@newa.test', 'Udo User', 'user', true)
ON CONFLICT (email) DO NOTHING;

-- ── Sessions (Token → SHA-256 Hash) ──
-- dev-admin-token  → 1734d503f6aa6a047c36d113cbad769f719c93784b469b771c4c3e7c63adbefd
-- dev-leader-token → 6dfd37eb4a13eaf973764040df4734ccf9c00f6580e102e71f6cd50fc1f1c344
-- dev-user-token   → ff194a51405eb34180b91ed9d9130ec5ddec108174c6806fc333ec3c33d83870

DELETE FROM sessions WHERE user_id IN (
  'a0000000-0000-0000-0000-000000000001',
  'a0000000-0000-0000-0000-000000000002',
  'a0000000-0000-0000-0000-000000000003'
);

INSERT INTO sessions (user_id, token_hash, expires_at)
VALUES
  ('a0000000-0000-0000-0000-000000000001', '1734d503f6aa6a047c36d113cbad769f719c93784b469b771c4c3e7c63adbefd', NOW() + INTERVAL '30 days'),
  ('a0000000-0000-0000-0000-000000000002', '6dfd37eb4a13eaf973764040df4734ccf9c00f6580e102e71f6cd50fc1f1c344', NOW() + INTERVAL '30 days'),
  ('a0000000-0000-0000-0000-000000000003', 'ff194a51405eb34180b91ed9d9130ec5ddec108174c6806fc333ec3c33d83870', NOW() + INTERVAL '30 days');

-- ── Teams ──

INSERT INTO teams (id, name, description)
VALUES
  ('b0000000-0000-0000-0000-000000000001', 'Entwicklung', 'Software-Entwicklungsteam'),
  ('b0000000-0000-0000-0000-000000000002', 'Vertrieb', 'Vertriebsteam')
ON CONFLICT (name) DO NOTHING;

INSERT INTO team_members (user_id, team_id, role)
VALUES
  ('a0000000-0000-0000-0000-000000000001', 'b0000000-0000-0000-0000-000000000001', 'admin'),
  ('a0000000-0000-0000-0000-000000000002', 'b0000000-0000-0000-0000-000000000001', 'team_leader'),
  ('a0000000-0000-0000-0000-000000000003', 'b0000000-0000-0000-0000-000000000001', 'user'),
  ('a0000000-0000-0000-0000-000000000002', 'b0000000-0000-0000-0000-000000000002', 'team_leader')
ON CONFLICT (user_id, team_id) DO NOTHING;

-- ── Projekte ──

INSERT INTO projects (id, name, customer_name, is_active)
VALUES
  ('c0000000-0000-0000-0000-000000000001', 'Zeiterfassung', 'NEWA', true),
  ('c0000000-0000-0000-0000-000000000002', 'Website Relaunch', 'Kunde A', true),
  ('c0000000-0000-0000-0000-000000000003', 'Altsystem', 'Intern', false)
ON CONFLICT (name) DO NOTHING;

-- ── Zeiteintraege (aktuelle Woche fuer alle 3 User) ──

-- Montag dieser Woche berechnen
DO $$
DECLARE
  mon DATE := date_trunc('week', CURRENT_DATE)::DATE;
BEGIN

  -- Admin: Mo-Fr
  INSERT INTO time_entries (user_id, entry_date, start_time, end_time, break_minutes, project_id, description)
  VALUES
    ('a0000000-0000-0000-0000-000000000001', mon,     '08:00', '17:00', 30, 'c0000000-0000-0000-0000-000000000001', 'Backend-Entwicklung'),
    ('a0000000-0000-0000-0000-000000000001', mon + 1, '08:30', '17:30', 30, 'c0000000-0000-0000-0000-000000000001', 'Code Review'),
    ('a0000000-0000-0000-0000-000000000001', mon + 2, '07:30', '16:00', 45, 'c0000000-0000-0000-0000-000000000002', 'Kundenmeeting + Entwicklung'),
    ('a0000000-0000-0000-0000-000000000001', mon + 3, '09:00', '18:00', 30, 'c0000000-0000-0000-0000-000000000001', 'Feature-Entwicklung'),
    ('a0000000-0000-0000-0000-000000000001', mon + 4, '08:00', '14:00', 0,  'c0000000-0000-0000-0000-000000000001', 'Freitag kurzer Tag');

  -- Teamleiter: Mo-Mi
  INSERT INTO time_entries (user_id, entry_date, start_time, end_time, break_minutes, project_id, description)
  VALUES
    ('a0000000-0000-0000-0000-000000000002', mon,     '09:00', '17:30', 30, 'c0000000-0000-0000-0000-000000000002', 'Sprint Planning'),
    ('a0000000-0000-0000-0000-000000000002', mon + 1, '08:00', '16:30', 30, 'c0000000-0000-0000-0000-000000000002', 'Entwicklung'),
    ('a0000000-0000-0000-0000-000000000002', mon + 2, '08:30', '17:00', 45, NULL, 'Team-Workshop');

  -- User: Mo-Di
  INSERT INTO time_entries (user_id, entry_date, start_time, end_time, break_minutes, project_id, description)
  VALUES
    ('a0000000-0000-0000-0000-000000000003', mon,     '08:00', '16:30', 30, 'c0000000-0000-0000-0000-000000000001', 'Frontend-Arbeit'),
    ('a0000000-0000-0000-0000-000000000003', mon + 1, '09:00', '17:00', 30, 'c0000000-0000-0000-0000-000000000001', 'Bugfixes');

END $$;

-- ── Abwesenheiten (Phase 5) ──

-- Vacation entitlements fuer 2026
INSERT INTO vacation_entitlements (user_id, year, total_days, carry_over_days)
VALUES
  ('a0000000-0000-0000-0000-000000000001', 2026, 30, 3),
  ('a0000000-0000-0000-0000-000000000002', 2026, 28, 2),
  ('a0000000-0000-0000-0000-000000000003', 2026, 30, 0)
ON CONFLICT (user_id, year) DO NOTHING;

-- Absences (use subquery for absence type IDs)
INSERT INTO absences (user_id, absence_type_id, start_date, end_date, note, status, reviewed_by, reviewed_at)
VALUES
  -- Admin: 1 Woche Urlaub (genehmigt)
  ('a0000000-0000-0000-0000-000000000001',
   (SELECT id FROM absence_types WHERE name = 'Urlaub'),
   '2026-04-06', '2026-04-10', 'Osterurlaub', 'approved',
   'a0000000-0000-0000-0000-000000000001', NOW()),
  -- Teamleiter: 2 Tage Krankheit (auto-genehmigt)
  ('a0000000-0000-0000-0000-000000000002',
   (SELECT id FROM absence_types WHERE name = 'Krankheit'),
   '2026-03-30', '2026-03-31', 'Erkaeltung', 'approved',
   NULL, NULL),
  -- User: Urlaub beantragt (pending)
  ('a0000000-0000-0000-0000-000000000003',
   (SELECT id FROM absence_types WHERE name = 'Urlaub'),
   '2026-04-20', '2026-04-24', 'Kurzurlaub', 'pending',
   NULL, NULL),
  -- User: Homeoffice naechste Woche (auto-genehmigt)
  ('a0000000-0000-0000-0000-000000000003',
   (SELECT id FROM absence_types WHERE name = 'Homeoffice'),
   '2026-03-30', '2026-03-30', '', 'approved',
   NULL, NULL);

-- ── Arbeitszeitmodelle (Phase 6) ──

INSERT INTO work_schedules (user_id, valid_from, weekly_hours, monday_hours, tuesday_hours, wednesday_hours, thursday_hours, friday_hours, saturday_hours, sunday_hours)
VALUES
  -- Admin: Vollzeit 40h
  ('a0000000-0000-0000-0000-000000000001', '2026-01-01', 40.00, 8.00, 8.00, 8.00, 8.00, 8.00, 0.00, 0.00),
  -- Teamleiter: Vollzeit 40h
  ('a0000000-0000-0000-0000-000000000002', '2026-01-01', 40.00, 8.00, 8.00, 8.00, 8.00, 8.00, 0.00, 0.00),
  -- User: Teilzeit 30h (Mo-Fr je 6h)
  ('a0000000-0000-0000-0000-000000000003', '2026-01-01', 30.00, 6.00, 6.00, 6.00, 6.00, 6.00, 0.00, 0.00)
ON CONFLICT (user_id, valid_from) DO NOTHING;

COMMIT;

-- =============================================================
-- Fertig! Teste mit:
--
--   curl -b "zeit_session=dev-admin-token" http://localhost:8080/api/auth/me
--   curl -b "zeit_session=dev-admin-token" http://localhost:8080/api/time-entries
--   curl -b "zeit_session=dev-leader-token" http://localhost:8080/api/time-entries
--   curl -b "zeit_session=dev-user-token" http://localhost:8080/api/time-entries
--
--   # Neuen Eintrag erstellen:
--   curl -X POST -b "zeit_session=dev-admin-token" \
--     -H "Content-Type: application/json" \
--     -d '{"entry_date":"2026-03-25","start_time":"08:00","end_time":"17:00","break_minutes":30}' \
--     http://localhost:8080/api/time-entries
--
--   # ArbZG-Warnung provozieren (>6h, nur 10min Pause):
--   curl -X POST -b "zeit_session=dev-user-token" \
--     -H "Content-Type: application/json" \
--     -d '{"entry_date":"2026-03-28","start_time":"08:00","end_time":"17:00","break_minutes":10}' \
--     http://localhost:8080/api/time-entries
--
--   # Abwesenheiten:
--   curl -b "zeit_session=dev-admin-token" http://localhost:8080/api/absences
--   curl -b "zeit_session=dev-admin-token" http://localhost:8080/api/absence-types
--   curl -b "zeit_session=dev-admin-token" http://localhost:8080/api/absences/balance
--
--   # Neue Abwesenheit beantragen:
--   curl -X POST -b "zeit_session=dev-user-token" \
--     -H "Content-Type: application/json" \
--     -d '{"absence_type_id":"<URLAUB_TYPE_ID>","start_date":"2026-05-04","end_date":"2026-05-08","note":"Mai-Urlaub"}' \
--     http://localhost:8080/api/absences
--
--   # Ueberstunden & Dashboard (Phase 6):
--   curl -b "zeit_session=dev-admin-token" http://localhost:8080/api/dashboard
--   curl -b "zeit_session=dev-admin-token" http://localhost:8080/api/overtime
--   curl -b "zeit_session=dev-admin-token" http://localhost:8080/api/overtime/trend
--   curl -b "zeit_session=dev-admin-token" http://localhost:8080/api/work-schedule
--   curl -b "zeit_session=dev-leader-token" "http://localhost:8080/api/teams/b0000000-0000-0000-0000-000000000001/availability?from=2026-03-23&to=2026-03-27"
-- =============================================================
