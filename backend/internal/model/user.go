package model

import (
	"time"

	"github.com/google/uuid"
)

// UserRole represents the role a user can have globally or within a team.
type UserRole string

const (
	RoleAdmin      UserRole = "admin"
	RoleTeamLeader UserRole = "team_leader"
	RoleUser       UserRole = "user"
)

// IsValid checks whether the role is a known value.
func (r UserRole) IsValid() bool {
	switch r {
	case RoleAdmin, RoleTeamLeader, RoleUser:
		return true
	}
	return false
}

// User represents an authenticated user in the system.
type User struct {
	ID          uuid.UUID  `json:"id"`
	Email       string     `json:"email"`
	DisplayName string     `json:"display_name"`
	AzureOID    *string    `json:"azure_oid,omitempty"`
	GlobalRole  UserRole   `json:"global_role"`
	IsActive    bool       `json:"is_active"`
	CreatedAt   time.Time  `json:"created_at"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty"`
}

// IsAdmin returns true if the user has the global admin role.
func (u *User) IsAdmin() bool {
	return u.GlobalRole == RoleAdmin
}

// IsTeamLeader returns true if the user has the global team_leader role.
func (u *User) IsTeamLeader() bool {
	return u.GlobalRole == RoleTeamLeader
}
