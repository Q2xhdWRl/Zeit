package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/newa/zeiterfassung/internal/model"
)

// TeamRepository handles database operations for teams.
type TeamRepository struct {
	db *pgxpool.Pool
}

// NewTeamRepository creates a new TeamRepository.
func NewTeamRepository(db *pgxpool.Pool) *TeamRepository {
	return &TeamRepository{db: db}
}

// Create inserts a new team.
func (r *TeamRepository) Create(ctx context.Context, name, description string) (*model.Team, error) {
	var t model.Team
	err := r.db.QueryRow(ctx,
		`INSERT INTO teams (name, description) VALUES ($1, $2)
		 RETURNING id, name, description, created_at, updated_at`,
		name, description,
	).Scan(&t.ID, &t.Name, &t.Description, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// FindByID returns a team by its UUID.
func (r *TeamRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Team, error) {
	var t model.Team
	err := r.db.QueryRow(ctx,
		`SELECT id, name, description, created_at, updated_at FROM teams WHERE id = $1`, id,
	).Scan(&t.ID, &t.Name, &t.Description, &t.CreatedAt, &t.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// ListAll returns all teams ordered by name.
func (r *TeamRepository) ListAll(ctx context.Context) ([]model.Team, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, name, description, created_at, updated_at FROM teams ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var teams []model.Team
	for rows.Next() {
		var t model.Team
		if err := rows.Scan(&t.ID, &t.Name, &t.Description, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		teams = append(teams, t)
	}
	return teams, rows.Err()
}

// Update modifies a team's name and description.
func (r *TeamRepository) Update(ctx context.Context, id uuid.UUID, name, description string) (*model.Team, error) {
	var t model.Team
	err := r.db.QueryRow(ctx,
		`UPDATE teams SET name = $1, description = $2, updated_at = NOW()
		 WHERE id = $3
		 RETURNING id, name, description, created_at, updated_at`,
		name, description, id,
	).Scan(&t.ID, &t.Name, &t.Description, &t.CreatedAt, &t.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// Delete removes a team by its UUID.
func (r *TeamRepository) Delete(ctx context.Context, id uuid.UUID) error {
	ct, err := r.db.Exec(ctx, `DELETE FROM teams WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

// AddMember adds a user to a team with a given role.
func (r *TeamRepository) AddMember(ctx context.Context, teamID, userID uuid.UUID, role model.UserRole) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO team_members (team_id, user_id, role) VALUES ($1, $2, $3)
		 ON CONFLICT (user_id, team_id) DO UPDATE SET role = EXCLUDED.role`,
		teamID, userID, role,
	)
	return err
}

// RemoveMember removes a user from a team.
func (r *TeamRepository) RemoveMember(ctx context.Context, teamID, userID uuid.UUID) error {
	ct, err := r.db.Exec(ctx,
		`DELETE FROM team_members WHERE team_id = $1 AND user_id = $2`, teamID, userID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

// ListMembers returns all members of a team with user details.
func (r *TeamRepository) ListMembers(ctx context.Context, teamID uuid.UUID) ([]model.TeamMember, error) {
	rows, err := r.db.Query(ctx,
		`SELECT tm.user_id, tm.team_id, tm.role, tm.joined_at, u.display_name, u.email
		 FROM team_members tm
		 JOIN users u ON u.id = tm.user_id
		 WHERE tm.team_id = $1
		 ORDER BY u.display_name`, teamID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []model.TeamMember
	for rows.Next() {
		var m model.TeamMember
		if err := rows.Scan(&m.UserID, &m.TeamID, &m.Role, &m.JoinedAt, &m.DisplayName, &m.Email); err != nil {
			return nil, err
		}
		members = append(members, m)
	}
	return members, rows.Err()
}

// ListTeamsForUser returns all teams a user belongs to.
func (r *TeamRepository) ListTeamsForUser(ctx context.Context, userID uuid.UUID) ([]model.TeamMember, error) {
	rows, err := r.db.Query(ctx,
		`SELECT tm.user_id, tm.team_id, tm.role, tm.joined_at, t.name, ''
		 FROM team_members tm
		 JOIN teams t ON t.id = tm.team_id
		 WHERE tm.user_id = $1
		 ORDER BY t.name`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var memberships []model.TeamMember
	for rows.Next() {
		var m model.TeamMember
		if err := rows.Scan(&m.UserID, &m.TeamID, &m.Role, &m.JoinedAt, &m.DisplayName, &m.Email); err != nil {
			return nil, err
		}
		memberships = append(memberships, m)
	}
	return memberships, rows.Err()
}

// IsTeamLeader checks if a user is a team_leader or admin for a specific team.
func (r *TeamRepository) IsTeamLeader(ctx context.Context, teamID, userID uuid.UUID) (bool, error) {
	var role model.UserRole
	err := r.db.QueryRow(ctx,
		`SELECT role FROM team_members WHERE team_id = $1 AND user_id = $2`, teamID, userID,
	).Scan(&role)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return role == model.RoleTeamLeader || role == model.RoleAdmin, nil
}

// IsTeamLeaderOfUser checks if leaderID is a team_leader in any team that memberID belongs to.
// Used to authorize cross-team operations like absence review.
func (r *TeamRepository) IsTeamLeaderOfUser(ctx context.Context, leaderID, memberID uuid.UUID) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx,
		`SELECT EXISTS(
			SELECT 1 FROM team_members tm_leader
			JOIN team_members tm_member ON tm_leader.team_id = tm_member.team_id
			WHERE tm_leader.user_id = $1
			  AND tm_leader.role IN ('team_leader', 'admin')
			  AND tm_member.user_id = $2
		)`, leaderID, memberID,
	).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}
