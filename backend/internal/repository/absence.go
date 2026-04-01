package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/newa/zeiterfassung/internal/config"
	"github.com/newa/zeiterfassung/internal/model"
)

// AbsenceRepository handles database operations for absences and related entities.
type AbsenceRepository struct {
	db *pgxpool.Pool
}

// NewAbsenceRepository creates a new AbsenceRepository.
func NewAbsenceRepository(db *pgxpool.Pool) *AbsenceRepository {
	return &AbsenceRepository{db: db}
}

// ── Absence Types ──

const absenceTypeColumns = `id, name, color, requires_approval, counts_as_work, is_active, sort_order, created_at`

func scanAbsenceType(row pgx.Row) (*model.AbsenceType, error) {
	var t model.AbsenceType
	err := row.Scan(&t.ID, &t.Name, &t.Color, &t.RequiresApproval, &t.CountsAsWork, &t.IsActive, &t.SortOrder, &t.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// ListAbsenceTypes returns all active absence types ordered by sort_order.
func (r *AbsenceRepository) ListAbsenceTypes(ctx context.Context) ([]model.AbsenceType, error) {
	rows, err := r.db.Query(ctx,
		`SELECT `+absenceTypeColumns+` FROM absence_types WHERE is_active = true ORDER BY sort_order, name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var types []model.AbsenceType
	for rows.Next() {
		var t model.AbsenceType
		if err := rows.Scan(&t.ID, &t.Name, &t.Color, &t.RequiresApproval, &t.CountsAsWork, &t.IsActive, &t.SortOrder, &t.CreatedAt); err != nil {
			return nil, err
		}
		types = append(types, t)
	}
	return types, rows.Err()
}

// ListAllAbsenceTypes returns all absence types (including inactive) for admin.
func (r *AbsenceRepository) ListAllAbsenceTypes(ctx context.Context) ([]model.AbsenceType, error) {
	rows, err := r.db.Query(ctx,
		`SELECT `+absenceTypeColumns+` FROM absence_types ORDER BY sort_order, name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var types []model.AbsenceType
	for rows.Next() {
		var t model.AbsenceType
		if err := rows.Scan(&t.ID, &t.Name, &t.Color, &t.RequiresApproval, &t.CountsAsWork, &t.IsActive, &t.SortOrder, &t.CreatedAt); err != nil {
			return nil, err
		}
		types = append(types, t)
	}
	return types, rows.Err()
}

// FindAbsenceTypeByID returns an absence type by its UUID.
func (r *AbsenceRepository) FindAbsenceTypeByID(ctx context.Context, id uuid.UUID) (*model.AbsenceType, error) {
	return scanAbsenceType(r.db.QueryRow(ctx,
		`SELECT `+absenceTypeColumns+` FROM absence_types WHERE id = $1`, id))
}

// UpdateAbsenceType modifies an absence type.
func (r *AbsenceRepository) UpdateAbsenceType(ctx context.Context, id uuid.UUID, name, color string, requiresApproval, countsAsWork, isActive bool, sortOrder int) (*model.AbsenceType, error) {
	return scanAbsenceType(r.db.QueryRow(ctx,
		`UPDATE absence_types SET name = $1, color = $2, requires_approval = $3, counts_as_work = $4, is_active = $5, sort_order = $6
		 WHERE id = $7 RETURNING `+absenceTypeColumns,
		name, color, requiresApproval, countsAsWork, isActive, sortOrder, id))
}

// ── Absences ──

const absenceColumns = `id, user_id, absence_type_id, start_date, end_date, note, status, reviewed_by, reviewed_at, review_note, created_at, updated_at`

// absenceColumnsA qualifies every column with the "a" table alias, required for
// JOIN queries where other tables (e.g. team_members) expose the same column names.
const absenceColumnsA = `a.id, a.user_id, a.absence_type_id, a.start_date, a.end_date, a.note, a.status, a.reviewed_by, a.reviewed_at, a.review_note, a.created_at, a.updated_at`

func scanAbsence(row pgx.Row) (*model.Absence, error) {
	var a model.Absence
	err := row.Scan(&a.ID, &a.UserID, &a.AbsenceTypeID, &a.StartDate, &a.EndDate, &a.Note, &a.Status, &a.ReviewedBy, &a.ReviewedAt, &a.ReviewNote, &a.CreatedAt, &a.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &a, nil
}

// Create inserts a new absence request.
func (r *AbsenceRepository) Create(ctx context.Context, userID, absenceTypeID uuid.UUID, startDate, endDate time.Time, note string, status model.AbsenceStatus) (*model.Absence, error) {
	return scanAbsence(r.db.QueryRow(ctx,
		`INSERT INTO absences (user_id, absence_type_id, start_date, end_date, note, status)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING `+absenceColumns,
		userID, absenceTypeID, startDate, endDate, note, status))
}

// FindByID returns an absence by its UUID.
func (r *AbsenceRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Absence, error) {
	return scanAbsence(r.db.QueryRow(ctx,
		`SELECT `+absenceColumns+` FROM absences WHERE id = $1`, id))
}

// ListByUser returns absences for a user in a date range.
func (r *AbsenceRepository) ListByUser(ctx context.Context, userID uuid.UUID, from, to time.Time) ([]model.Absence, error) {
	return r.queryAbsences(ctx,
		`SELECT `+absenceColumns+` FROM absences
		 WHERE user_id = $1 AND start_date <= $3 AND end_date >= $2
		 ORDER BY start_date`,
		userID, from, to)
}

// ListByTeam returns absences for all members of a team in a date range.
func (r *AbsenceRepository) ListByTeam(ctx context.Context, teamID uuid.UUID, from, to time.Time) ([]model.Absence, error) {
	return r.queryAbsences(ctx,
		`SELECT `+absenceColumnsA+`
		 FROM absences a
		 JOIN team_members tm ON tm.user_id = a.user_id
		 WHERE tm.team_id = $1 AND a.start_date <= $3 AND a.end_date >= $2
		 ORDER BY a.start_date`,
		teamID, from, to)
}

// ListPendingForTeam returns pending absences for a team (for approval).
func (r *AbsenceRepository) ListPendingForTeam(ctx context.Context, teamID uuid.UUID) ([]model.Absence, error) {
	return r.queryAbsences(ctx,
		`SELECT `+absenceColumnsA+`
		 FROM absences a
		 JOIN team_members tm ON tm.user_id = a.user_id
		 WHERE tm.team_id = $1 AND a.status = 'pending'
		 ORDER BY a.created_at`,
		teamID)
}

// Update modifies a pending absence request.
func (r *AbsenceRepository) Update(ctx context.Context, id uuid.UUID, absenceTypeID uuid.UUID, startDate, endDate time.Time, note string) (*model.Absence, error) {
	return scanAbsence(r.db.QueryRow(ctx,
		`UPDATE absences SET absence_type_id = $1, start_date = $2, end_date = $3, note = $4, updated_at = NOW()
		 WHERE id = $5 AND status = 'pending'
		 RETURNING `+absenceColumns,
		absenceTypeID, startDate, endDate, note, id))
}

// UpdateStatus sets the status of an absence (approve/reject/cancel).
func (r *AbsenceRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status model.AbsenceStatus, reviewedBy uuid.UUID, reviewNote string) (*model.Absence, error) {
	return scanAbsence(r.db.QueryRow(ctx,
		`UPDATE absences SET status = $1, reviewed_by = $2, reviewed_at = NOW(), review_note = $3, updated_at = NOW()
		 WHERE id = $4
		 RETURNING `+absenceColumns,
		status, reviewedBy, reviewNote, id))
}

// Delete removes a pending absence.
func (r *AbsenceRepository) Delete(ctx context.Context, id uuid.UUID) error {
	ct, err := r.db.Exec(ctx, `DELETE FROM absences WHERE id = $1 AND status = 'pending'`, id)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

// HasOverlap checks if a new/updated absence would overlap with existing non-cancelled absences.
func (r *AbsenceRepository) HasOverlap(ctx context.Context, userID uuid.UUID, startDate, endDate time.Time, excludeID uuid.UUID) (bool, error) {
	var count int
	err := r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM absences
		 WHERE user_id = $1
		   AND id != $2
		   AND status != 'cancelled'
		   AND status != 'rejected'
		   AND start_date <= $4
		   AND end_date >= $3`,
		userID, excludeID, startDate, endDate).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// CountVacationDays counts approved + pending vacation days for a user in a given year.
// Returns (approved_days, pending_days).
func (r *AbsenceRepository) CountVacationDays(ctx context.Context, userID uuid.UUID, year int, vacationTypeID uuid.UUID) (int, int, error) {
	yearStart := time.Date(year, 1, 1, 0, 0, 0, 0, config.AppLocation)
	yearEnd := time.Date(year, 12, 31, 0, 0, 0, 0, config.AppLocation)

	rows, err := r.db.Query(ctx,
		`SELECT start_date, end_date, status FROM absences
		 WHERE user_id = $1 AND absence_type_id = $2
		   AND start_date <= $4 AND end_date >= $3
		   AND status IN ('approved', 'pending')`,
		userID, vacationTypeID, yearStart, yearEnd)
	if err != nil {
		return 0, 0, err
	}
	defer rows.Close()

	var approvedDays, pendingDays int
	for rows.Next() {
		var a model.Absence
		if err := rows.Scan(&a.StartDate, &a.EndDate, &a.Status); err != nil {
			return 0, 0, err
		}
		// Clamp to year boundaries
		if a.StartDate.Before(yearStart) {
			a.StartDate = yearStart
		}
		if a.EndDate.After(yearEnd) {
			a.EndDate = yearEnd
		}
		days := a.WorkingDays()
		if a.Status == model.AbsenceStatusApproved {
			approvedDays += days
		} else {
			pendingDays += days
		}
	}
	return approvedDays, pendingDays, rows.Err()
}

// ── Vacation Entitlements ──

const entitlementColumns = `id, user_id, year, total_days, carry_over_days, created_at, updated_at`

func scanEntitlement(row pgx.Row) (*model.VacationEntitlement, error) {
	var e model.VacationEntitlement
	err := row.Scan(&e.ID, &e.UserID, &e.Year, &e.TotalDays, &e.CarryOverDays, &e.CreatedAt, &e.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &e, nil
}

// FindEntitlement returns the vacation entitlement for a user and year.
func (r *AbsenceRepository) FindEntitlement(ctx context.Context, userID uuid.UUID, year int) (*model.VacationEntitlement, error) {
	return scanEntitlement(r.db.QueryRow(ctx,
		`SELECT `+entitlementColumns+` FROM vacation_entitlements WHERE user_id = $1 AND year = $2`,
		userID, year))
}

// UpsertEntitlement creates or updates a vacation entitlement.
func (r *AbsenceRepository) UpsertEntitlement(ctx context.Context, userID uuid.UUID, year, totalDays, carryOverDays int) (*model.VacationEntitlement, error) {
	return scanEntitlement(r.db.QueryRow(ctx,
		`INSERT INTO vacation_entitlements (user_id, year, total_days, carry_over_days)
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT (user_id, year) DO UPDATE SET total_days = $3, carry_over_days = $4, updated_at = NOW()
		 RETURNING `+entitlementColumns,
		userID, year, totalDays, carryOverDays))
}

// ListEntitlementsByUser returns all entitlements for a user.
func (r *AbsenceRepository) ListEntitlementsByUser(ctx context.Context, userID uuid.UUID) ([]model.VacationEntitlement, error) {
	rows, err := r.db.Query(ctx,
		`SELECT `+entitlementColumns+` FROM vacation_entitlements WHERE user_id = $1 ORDER BY year DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entitlements []model.VacationEntitlement
	for rows.Next() {
		var e model.VacationEntitlement
		if err := rows.Scan(&e.ID, &e.UserID, &e.Year, &e.TotalDays, &e.CarryOverDays, &e.CreatedAt, &e.UpdatedAt); err != nil {
			return nil, err
		}
		entitlements = append(entitlements, e)
	}
	return entitlements, rows.Err()
}

func (r *AbsenceRepository) queryAbsences(ctx context.Context, query string, args ...any) ([]model.Absence, error) {
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var absences []model.Absence
	for rows.Next() {
		var a model.Absence
		if err := rows.Scan(&a.ID, &a.UserID, &a.AbsenceTypeID, &a.StartDate, &a.EndDate, &a.Note, &a.Status, &a.ReviewedBy, &a.ReviewedAt, &a.ReviewNote, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, err
		}
		absences = append(absences, a)
	}
	return absences, rows.Err()
}
