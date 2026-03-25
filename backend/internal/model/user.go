package model

import (
	"time"

	"github.com/google/uuid"
)

// User represents an authenticated user in the system.
type User struct {
	ID          uuid.UUID  `json:"id"`
	Email       string     `json:"email"`
	DisplayName string     `json:"display_name"`
	AzureOID    *string    `json:"azure_oid,omitempty"`
	IsActive    bool       `json:"is_active"`
	CreatedAt   time.Time  `json:"created_at"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty"`
}
