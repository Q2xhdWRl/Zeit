package model

import (
	"time"

	"github.com/google/uuid"
)

// Project represents a billable project that time entries can be assigned to.
type Project struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	CustomerName string    `json:"customer_name"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
}
