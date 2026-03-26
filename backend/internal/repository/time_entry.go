package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"

	"github.com/newa/zeiterfassung/internal/model"
)

// TimeEntryRepository handles database operations for time entries.
type TimeEntryRepository struct {
	db *pgxpool.Pool
}

// NewTimeEntryRepository creates a new TimeEntryRepository.
func NewTimeEntryRepository(db *pgxpool.Pool) *TimeEntryRepository {
	return &TimeEntryRepository{db: db}
}

const timeEntryColumns = `id, user_id, entry_date, start_time, end_time, break_minutes, project_id, description, insert_time, updated_at`

func scanTimeEntry(row pgx.Row) (*model.TimeEntry, error) {
	var e model.TimeEntry
	err := row.Scan(&e.ID, &e.UserID, &e.EntryDate, &e.StartTime, &e.EndTime, &e.BreakMinutes, &e.ProjectID, &e.Description, &e.InsertTime, &e.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &e, nil
}

// Create inserts a new time entry.
func (r *TimeEntryRepository) Create(ctx context.Context, userID uuid.UUID, entryDate time.Time, startTime, endTime string, breakMinutes int, projectID *uuid.UUID, description string) (*model.TimeEntry, error) {
	return scanTimeEntry(r.db.QueryRow(ctx,
		`INSERT INTO time_entries (user_id, entry_date, start_time, end_time, break_minutes, project_id, description)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 RETURNING `+timeEntryColumns,
		userID, entryDate, startTime, endTime, breakMinutes, projectID, description,
	))
}

// FindByID returns a time entry by its UUID.
func (r *TimeEntryRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.TimeEntry, error) {
	return scanTimeEntry(r.db.QueryRow(ctx,
		`SELECT `+timeEntryColumns+` FROM time_entries WHERE id = $1`, id,
	))
}

// ListByUserAndDate returns all time entries for a user on a given date.
func (r *TimeEntryRepository) ListByUserAndDate(ctx context.Context, userID uuid.UUID, date time.Time) ([]model.TimeEntry, error) {
	return r.queryEntries(ctx,
		`SELECT `+timeEntryColumns+` FROM time_entries WHERE user_id = $1 AND entry_date = $2 ORDER BY start_time`,
		userID, date,
	)
}

// ListByUserAndDateRange returns all time entries for a user in a date range (inclusive).
func (r *TimeEntryRepository) ListByUserAndDateRange(ctx context.Context, userID uuid.UUID, from, to time.Time) ([]model.TimeEntry, error) {
	return r.queryEntries(ctx,
		`SELECT `+timeEntryColumns+` FROM time_entries WHERE user_id = $1 AND entry_date >= $2 AND entry_date <= $3 ORDER BY entry_date, start_time`,
		userID, from, to,
	)
}

// ListByTeamAndDateRange returns time entries for all members of a team in a date range.
func (r *TimeEntryRepository) ListByTeamAndDateRange(ctx context.Context, teamID uuid.UUID, from, to time.Time) ([]model.TimeEntry, error) {
	return r.queryEntries(ctx,
		`SELECT te.`+timeEntryColumns+`
		 FROM time_entries te
		 JOIN team_members tm ON tm.user_id = te.user_id
		 WHERE tm.team_id = $1 AND te.entry_date >= $2 AND te.entry_date <= $3
		 ORDER BY te.entry_date, te.start_time`,
		teamID, from, to,
	)
}

// Update modifies an existing time entry.
func (r *TimeEntryRepository) Update(ctx context.Context, id uuid.UUID, entryDate time.Time, startTime, endTime string, breakMinutes int, projectID *uuid.UUID, description string) (*model.TimeEntry, error) {
	return scanTimeEntry(r.db.QueryRow(ctx,
		`UPDATE time_entries
		 SET entry_date = $1, start_time = $2, end_time = $3, break_minutes = $4, project_id = $5, description = $6, updated_at = NOW()
		 WHERE id = $7
		 RETURNING `+timeEntryColumns,
		entryDate, startTime, endTime, breakMinutes, projectID, description, id,
	))
}

// Delete removes a time entry.
func (r *TimeEntryRepository) Delete(ctx context.Context, id uuid.UUID) error {
	ct, err := r.db.Exec(ctx, `DELETE FROM time_entries WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

// HasOverlap checks if a new/updated time entry would overlap with existing entries for the same user and date.
// excludeID is the entry ID to exclude (for updates); pass uuid.Nil for new entries.
func (r *TimeEntryRepository) HasOverlap(ctx context.Context, userID uuid.UUID, entryDate time.Time, startTime, endTime string, excludeID uuid.UUID) (bool, error) {
	var count int
	err := r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM time_entries
		 WHERE user_id = $1 AND entry_date = $2
		   AND id != $3
		   AND start_time < $5 AND end_time > $4`,
		userID, entryDate, excludeID, startTime, endTime,
	).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// CreateAudit logs a change to the time_entry_audit table.
func (r *TimeEntryRepository) CreateAudit(ctx context.Context, entryID, userID uuid.UUID, action string, oldValues, newValues any, changedBy uuid.UUID) {
	_, err := r.db.Exec(ctx,
		`INSERT INTO time_entry_audit (time_entry_id, user_id, action, old_values, new_values, changed_by)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		entryID, userID, action, oldValues, newValues, changedBy,
	)
	if err != nil {
		log.Error().Err(err).Str("entry_id", entryID.String()).Msg("failed to create audit log")
	}
}

// DailySummary holds aggregated data for a single day.
type DailySummary struct {
	Date         time.Time `json:"date"`
	TotalMinutes int       `json:"total_minutes"`
	EntryCount   int       `json:"entry_count"`
}

// SummaryByDateRange returns per-day summaries for a user.
func (r *TimeEntryRepository) SummaryByDateRange(ctx context.Context, userID uuid.UUID, from, to time.Time) ([]DailySummary, error) {
	rows, err := r.db.Query(ctx,
		`SELECT entry_date,
		        SUM(EXTRACT(EPOCH FROM (end_time - start_time))/60 - break_minutes)::INT AS total_minutes,
		        COUNT(*) AS entry_count
		 FROM time_entries
		 WHERE user_id = $1 AND entry_date >= $2 AND entry_date <= $3
		 GROUP BY entry_date
		 ORDER BY entry_date`,
		userID, from, to,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var summaries []DailySummary
	for rows.Next() {
		var s DailySummary
		if err := rows.Scan(&s.Date, &s.TotalMinutes, &s.EntryCount); err != nil {
			return nil, err
		}
		summaries = append(summaries, s)
	}
	return summaries, rows.Err()
}

func (r *TimeEntryRepository) queryEntries(ctx context.Context, query string, args ...any) ([]model.TimeEntry, error) {
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []model.TimeEntry
	for rows.Next() {
		var e model.TimeEntry
		if err := rows.Scan(&e.ID, &e.UserID, &e.EntryDate, &e.StartTime, &e.EndTime, &e.BreakMinutes, &e.ProjectID, &e.Description, &e.InsertTime, &e.UpdatedAt); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	return entries, rows.Err()
}
