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

// UserRepository handles database operations for users.
type UserRepository struct {
	db *pgxpool.Pool
}

// NewUserRepository creates a new UserRepository.
func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

// FindByEmail returns a user by their email address (case-insensitive via CITEXT).
func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	var u model.User
	err := r.db.QueryRow(ctx,
		`SELECT id, email, display_name, azure_oid, is_active, created_at, last_login_at
		 FROM users WHERE email = $1`, email,
	).Scan(&u.ID, &u.Email, &u.DisplayName, &u.AzureOID, &u.IsActive, &u.CreatedAt, &u.LastLoginAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// FindByID returns a user by their UUID.
func (r *UserRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	var u model.User
	err := r.db.QueryRow(ctx,
		`SELECT id, email, display_name, azure_oid, is_active, created_at, last_login_at
		 FROM users WHERE id = $1`, id,
	).Scan(&u.ID, &u.Email, &u.DisplayName, &u.AzureOID, &u.IsActive, &u.CreatedAt, &u.LastLoginAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// UpsertByAzureOID creates or updates a user based on their Azure Object ID.
// On conflict (same azure_oid), it updates display_name, email, and last_login_at.
func (r *UserRepository) UpsertByAzureOID(ctx context.Context, email, displayName, azureOID string) (*model.User, error) {
	var u model.User
	now := time.Now()
	err := r.db.QueryRow(ctx,
		`INSERT INTO users (email, display_name, azure_oid, last_login_at)
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT (azure_oid) DO UPDATE SET
		   email = EXCLUDED.email,
		   display_name = EXCLUDED.display_name,
		   last_login_at = $4
		 RETURNING id, email, display_name, azure_oid, is_active, created_at, last_login_at`,
		email, displayName, azureOID, now,
	).Scan(&u.ID, &u.Email, &u.DisplayName, &u.AzureOID, &u.IsActive, &u.CreatedAt, &u.LastLoginAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// UpdateLastLogin sets the last_login_at timestamp for a user.
func (r *UserRepository) UpdateLastLogin(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx,
		`UPDATE users SET last_login_at = NOW() WHERE id = $1`, id,
	)
	return err
}
