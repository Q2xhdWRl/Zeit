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

// SessionRepository handles database operations for sessions.
type SessionRepository struct {
	db *pgxpool.Pool
}

// NewSessionRepository creates a new SessionRepository.
func NewSessionRepository(db *pgxpool.Pool) *SessionRepository {
	return &SessionRepository{db: db}
}

// Create inserts a new session into the database.
func (r *SessionRepository) Create(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) (*model.Session, error) {
	var s model.Session
	err := r.db.QueryRow(ctx,
		`INSERT INTO sessions (user_id, token_hash, expires_at)
		 VALUES ($1, $2, $3)
		 RETURNING id, user_id, token_hash, expires_at, created_at`,
		userID, tokenHash, expiresAt,
	).Scan(&s.ID, &s.UserID, &s.TokenHash, &s.ExpiresAt, &s.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// FindByTokenHash returns a valid (non-expired) session by its token hash.
func (r *SessionRepository) FindByTokenHash(ctx context.Context, tokenHash string) (*model.Session, error) {
	var s model.Session
	err := r.db.QueryRow(ctx,
		`SELECT id, user_id, token_hash, expires_at, created_at
		 FROM sessions
		 WHERE token_hash = $1 AND expires_at > NOW()`,
		tokenHash,
	).Scan(&s.ID, &s.UserID, &s.TokenHash, &s.ExpiresAt, &s.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// DeleteByID removes a session by its ID.
func (r *SessionRepository) DeleteByID(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `DELETE FROM sessions WHERE id = $1`, id)
	return err
}

// DeleteByUserID removes all sessions for a given user.
func (r *SessionRepository) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	_, err := r.db.Exec(ctx, `DELETE FROM sessions WHERE user_id = $1`, userID)
	return err
}

// DeleteExpired removes all expired sessions.
func (r *SessionRepository) DeleteExpired(ctx context.Context) (int64, error) {
	tag, err := r.db.Exec(ctx, `DELETE FROM sessions WHERE expires_at <= NOW()`)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}
