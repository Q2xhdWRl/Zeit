package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/newa/zeiterfassung/internal/model"
)

// WorkScheduleRepository handles database operations for work schedules.
type WorkScheduleRepository struct {
	db *pgxpool.Pool
}

// NewWorkScheduleRepository creates a new WorkScheduleRepository.
func NewWorkScheduleRepository(db *pgxpool.Pool) *WorkScheduleRepository {
	return &WorkScheduleRepository{db: db}
}

const scheduleColumns = `id, user_id, valid_from, weekly_hours, monday_hours, tuesday_hours, wednesday_hours, thursday_hours, friday_hours, saturday_hours, sunday_hours, created_at, updated_at`

func scanSchedule(row pgx.Row) (*model.WorkSchedule, error) {
	var s model.WorkSchedule
	err := row.Scan(&s.ID, &s.UserID, &s.ValidFrom, &s.WeeklyHours, &s.MondayHours, &s.TuesdayHours, &s.WednesdayHours, &s.ThursdayHours, &s.FridayHours, &s.SaturdayHours, &s.SundayHours, &s.CreatedAt, &s.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// FindActiveSchedule returns the work schedule active on a given date for a user.
func (r *WorkScheduleRepository) FindActiveSchedule(ctx context.Context, userID uuid.UUID, date time.Time) (*model.WorkSchedule, error) {
	return scanSchedule(r.db.QueryRow(ctx,
		`SELECT `+scheduleColumns+` FROM work_schedules
		 WHERE user_id = $1 AND valid_from <= $2
		 ORDER BY valid_from DESC LIMIT 1`,
		userID, date))
}

// Upsert creates or updates a work schedule for a user starting on a date.
func (r *WorkScheduleRepository) Upsert(ctx context.Context, userID uuid.UUID, validFrom time.Time, weeklyHours, mon, tue, wed, thu, fri, sat, sun float64) (*model.WorkSchedule, error) {
	return scanSchedule(r.db.QueryRow(ctx,
		`INSERT INTO work_schedules (user_id, valid_from, weekly_hours, monday_hours, tuesday_hours, wednesday_hours, thursday_hours, friday_hours, saturday_hours, sunday_hours)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		 ON CONFLICT (user_id, valid_from) DO UPDATE SET
		   weekly_hours = $3, monday_hours = $4, tuesday_hours = $5, wednesday_hours = $6,
		   thursday_hours = $7, friday_hours = $8, saturday_hours = $9, sunday_hours = $10, updated_at = NOW()
		 RETURNING `+scheduleColumns,
		userID, validFrom, weeklyHours, mon, tue, wed, thu, fri, sat, sun))
}

// ListByUser returns all schedules for a user ordered by valid_from desc.
func (r *WorkScheduleRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]model.WorkSchedule, error) {
	rows, err := r.db.Query(ctx,
		`SELECT `+scheduleColumns+` FROM work_schedules WHERE user_id = $1 ORDER BY valid_from DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var schedules []model.WorkSchedule
	for rows.Next() {
		var s model.WorkSchedule
		if err := rows.Scan(&s.ID, &s.UserID, &s.ValidFrom, &s.WeeklyHours, &s.MondayHours, &s.TuesdayHours, &s.WednesdayHours, &s.ThursdayHours, &s.FridayHours, &s.SaturdayHours, &s.SundayHours, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		schedules = append(schedules, s)
	}
	return schedules, rows.Err()
}
