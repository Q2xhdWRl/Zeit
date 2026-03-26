package model

import (
	"time"

	"github.com/google/uuid"
)

// AbsenceStatus represents the current status of an absence request.
type AbsenceStatus string

const (
	AbsenceStatusPending   AbsenceStatus = "pending"
	AbsenceStatusApproved  AbsenceStatus = "approved"
	AbsenceStatusRejected  AbsenceStatus = "rejected"
	AbsenceStatusCancelled AbsenceStatus = "cancelled"
)

// IsValid checks whether the status is a known value.
func (s AbsenceStatus) IsValid() bool {
	switch s {
	case AbsenceStatusPending, AbsenceStatusApproved, AbsenceStatusRejected, AbsenceStatusCancelled:
		return true
	}
	return false
}

// AbsenceType represents a category of absence (e.g. Urlaub, Krankheit).
type AbsenceType struct {
	ID               uuid.UUID `json:"id"`
	Name             string    `json:"name"`
	Color            string    `json:"color"`
	RequiresApproval bool      `json:"requires_approval"`
	CountsAsWork     bool      `json:"counts_as_work"`
	IsActive         bool      `json:"is_active"`
	SortOrder        int       `json:"sort_order"`
	CreatedAt        time.Time `json:"created_at"`
}

// Absence represents a user's absence request.
type Absence struct {
	ID            uuid.UUID     `json:"id"`
	UserID        uuid.UUID     `json:"user_id"`
	AbsenceTypeID uuid.UUID     `json:"absence_type_id"`
	StartDate     time.Time     `json:"start_date"`
	EndDate       time.Time     `json:"end_date"`
	Note          string        `json:"note"`
	Status        AbsenceStatus `json:"status"`
	ReviewedBy    *uuid.UUID    `json:"reviewed_by,omitempty"`
	ReviewedAt    *time.Time    `json:"reviewed_at,omitempty"`
	ReviewNote    string        `json:"review_note"`
	CreatedAt     time.Time     `json:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at"`
}

// WorkingDays returns the number of working days (Mo-Fr) in the absence period.
func (a *Absence) WorkingDays() int {
	count := 0
	for d := a.StartDate; !d.After(a.EndDate); d = d.AddDate(0, 0, 1) {
		wd := d.Weekday()
		if wd != time.Saturday && wd != time.Sunday {
			count++
		}
	}
	return count
}

// VacationEntitlement represents a user's vacation allowance for a given year.
type VacationEntitlement struct {
	ID            uuid.UUID `json:"id"`
	UserID        uuid.UUID `json:"user_id"`
	Year          int       `json:"year"`
	TotalDays     int       `json:"total_days"`
	CarryOverDays int       `json:"carry_over_days"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// VacationBalance holds the computed vacation balance for a user/year.
type VacationBalance struct {
	Year          int `json:"year"`
	TotalDays     int `json:"total_days"`
	CarryOverDays int `json:"carry_over_days"`
	UsedDays      int `json:"used_days"`
	PendingDays   int `json:"pending_days"`
	RemainingDays int `json:"remaining_days"`
}
