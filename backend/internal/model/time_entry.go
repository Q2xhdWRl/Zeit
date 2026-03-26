package model

import (
	"time"

	"github.com/google/uuid"
)

// TimeEntry represents a single time booking for a user on a given day.
type TimeEntry struct {
	ID           uuid.UUID  `json:"id"`
	UserID       uuid.UUID  `json:"user_id"`
	EntryDate    time.Time  `json:"entry_date"`
	StartTime    string     `json:"start_time"`
	EndTime      string     `json:"end_time"`
	BreakMinutes int        `json:"break_minutes"`
	ProjectID    *uuid.UUID `json:"project_id,omitempty"`
	Description  string     `json:"description"`
	InsertTime   time.Time  `json:"insert_time"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// WorkMinutes returns the net working minutes (end - start - break).
func (e *TimeEntry) WorkMinutes() (int, error) {
	start, err := time.Parse("15:04:05", e.StartTime)
	if err != nil {
		start, err = time.Parse("15:04", e.StartTime)
		if err != nil {
			return 0, err
		}
	}
	end, err := time.Parse("15:04:05", e.EndTime)
	if err != nil {
		end, err = time.Parse("15:04", e.EndTime)
		if err != nil {
			return 0, err
		}
	}
	total := int(end.Sub(start).Minutes()) - e.BreakMinutes
	return total, nil
}

// TimeEntryAudit records a change to a time entry.
type TimeEntryAudit struct {
	ID          uuid.UUID  `json:"id"`
	TimeEntryID uuid.UUID  `json:"time_entry_id"`
	UserID      uuid.UUID  `json:"user_id"`
	Action      string     `json:"action"`
	OldValues   any        `json:"old_values,omitempty"`
	NewValues   any        `json:"new_values,omitempty"`
	ChangedAt   time.Time  `json:"changed_at"`
	ChangedBy   uuid.UUID  `json:"changed_by"`
}
