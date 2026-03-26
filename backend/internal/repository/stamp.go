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

// StampRepository handles database operations for active stamps.
type StampRepository struct {
	db *pgxpool.Pool
}

// NewStampRepository creates a new StampRepository.
func NewStampRepository(db *pgxpool.Pool) *StampRepository {
	return &StampRepository{db: db}
}

const stampColumns = `user_id, started_at, break_start, break_minutes, project_id, description`

func scanStamp(row pgx.Row) (*model.ActiveStamp, error) {
	var s model.ActiveStamp
	err := row.Scan(&s.UserID, &s.StartedAt, &s.BreakStart, &s.BreakMinutes, &s.ProjectID, &s.Description)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// GetActive returns the active stamp for a user, or nil if none exists.
func (r *StampRepository) GetActive(ctx context.Context, userID uuid.UUID) (*model.ActiveStamp, error) {
	return scanStamp(r.db.QueryRow(ctx,
		`SELECT `+stampColumns+` FROM active_stamps WHERE user_id = $1`, userID,
	))
}

// Create inserts a new active stamp. Replaces any existing stamp for the user.
func (r *StampRepository) Create(ctx context.Context, userID uuid.UUID, projectID *uuid.UUID, description string) (*model.ActiveStamp, error) {
	return scanStamp(r.db.QueryRow(ctx,
		`INSERT INTO active_stamps (user_id, project_id, description)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (user_id) DO UPDATE
		   SET started_at    = NOW(),
		       break_start   = NULL,
		       break_minutes = 0,
		       project_id    = EXCLUDED.project_id,
		       description   = EXCLUDED.description
		 RETURNING `+stampColumns,
		userID, projectID, description,
	))
}

// SetBreakStart starts a break by recording the current time.
func (r *StampRepository) SetBreakStart(ctx context.Context, userID uuid.UUID, breakStart time.Time) error {
	_, err := r.db.Exec(ctx,
		`UPDATE active_stamps SET break_start = $1 WHERE user_id = $2`,
		breakStart, userID,
	)
	return err
}

// AccumulateBreak adds elapsed break minutes and clears break_start.
func (r *StampRepository) AccumulateBreak(ctx context.Context, userID uuid.UUID, additionalMinutes int) error {
	_, err := r.db.Exec(ctx,
		`UPDATE active_stamps
		 SET break_minutes = break_minutes + $1,
		     break_start   = NULL
		 WHERE user_id = $2`,
		additionalMinutes, userID,
	)
	return err
}

// Delete removes the active stamp for a user (called after stamp-out).
func (r *StampRepository) Delete(ctx context.Context, userID uuid.UUID) error {
	_, err := r.db.Exec(ctx,
		`DELETE FROM active_stamps WHERE user_id = $1`, userID,
	)
	return err
}
