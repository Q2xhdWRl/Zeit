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

// userColumns lists the standard columns for user queries.
const userColumns = `id, email, display_name, azure_oid, global_role, is_active, created_at, last_login_at`

// scanUser scans a row into a User struct (must match userColumns order).
func scanUser(row pgx.Row) (*model.User, error) {
	var u model.User
	err := row.Scan(&u.ID, &u.Email, &u.DisplayName, &u.AzureOID, &u.GlobalRole, &u.IsActive, &u.CreatedAt, &u.LastLoginAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// FindByEmail returns a user by their email address (case-insensitive via CITEXT).
func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	return scanUser(r.db.QueryRow(ctx,
		`SELECT `+userColumns+` FROM users WHERE email = $1`, email,
	))
}

// FindByID returns a user by their UUID.
func (r *UserRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	return scanUser(r.db.QueryRow(ctx,
		`SELECT `+userColumns+` FROM users WHERE id = $1`, id,
	))
}

// ListAll returns all users ordered by display name.
func (r *UserRepository) ListAll(ctx context.Context) ([]model.User, error) {
	rows, err := r.db.Query(ctx,
		`SELECT `+userColumns+` FROM users ORDER BY display_name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []model.User
	for rows.Next() {
		var u model.User
		if err := rows.Scan(&u.ID, &u.Email, &u.DisplayName, &u.AzureOID, &u.GlobalRole, &u.IsActive, &u.CreatedAt, &u.LastLoginAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

// UpdateRole sets the global role for a user.
func (r *UserRepository) UpdateRole(ctx context.Context, id uuid.UUID, role model.UserRole) error {
	ct, err := r.db.Exec(ctx,
		`UPDATE users SET global_role = $1 WHERE id = $2`, role, id)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

// UpdateActive sets the is_active flag for a user.
func (r *UserRepository) UpdateActive(ctx context.Context, id uuid.UUID, active bool) error {
	ct, err := r.db.Exec(ctx,
		`UPDATE users SET is_active = $1 WHERE id = $2`, active, id)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
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
		 RETURNING `+userColumns,
		email, displayName, azureOID, now,
	).Scan(&u.ID, &u.Email, &u.DisplayName, &u.AzureOID, &u.GlobalRole, &u.IsActive, &u.CreatedAt, &u.LastLoginAt)
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
