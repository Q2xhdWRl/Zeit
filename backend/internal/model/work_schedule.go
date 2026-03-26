package model

import (
	"time"

	"github.com/google/uuid"
)

// WorkSchedule represents a user's contracted working hours from a given date onwards.
type WorkSchedule struct {
	ID             uuid.UUID `json:"id"`
	UserID         uuid.UUID `json:"user_id"`
	ValidFrom      time.Time `json:"valid_from"`
	WeeklyHours    float64   `json:"weekly_hours"`
	MondayHours    float64   `json:"monday_hours"`
	TuesdayHours   float64   `json:"tuesday_hours"`
	WednesdayHours float64   `json:"wednesday_hours"`
	ThursdayHours  float64   `json:"thursday_hours"`
	FridayHours    float64   `json:"friday_hours"`
	SaturdayHours  float64   `json:"saturday_hours"`
	SundayHours    float64   `json:"sunday_hours"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// TargetMinutesForDay returns the target working minutes for a given weekday.
func (ws *WorkSchedule) TargetMinutesForDay(weekday time.Weekday) int {
	var hours float64
	switch weekday {
	case time.Monday:
		hours = ws.MondayHours
	case time.Tuesday:
		hours = ws.TuesdayHours
	case time.Wednesday:
		hours = ws.WednesdayHours
	case time.Thursday:
		hours = ws.ThursdayHours
	case time.Friday:
		hours = ws.FridayHours
	case time.Saturday:
		hours = ws.SaturdayHours
	case time.Sunday:
		hours = ws.SundayHours
	}
	return int(hours * 60)
}

// OvertimeSummary holds the computed overtime data for a period.
type OvertimeSummary struct {
	PeriodFrom    string `json:"period_from"`
	PeriodTo      string `json:"period_to"`
	TargetMinutes int    `json:"target_minutes"`
	ActualMinutes int    `json:"actual_minutes"`
	DiffMinutes   int    `json:"diff_minutes"`
}

// DayAvailability represents a team member's availability status for one day.
type DayAvailability struct {
	UserID      uuid.UUID `json:"user_id"`
	DisplayName string    `json:"display_name"`
	Date        string    `json:"date"`
	Status      string    `json:"status"` // "present", "absent", "homeoffice", "no_entry"
	AbsenceType string    `json:"absence_type,omitempty"`
	WorkMinutes int       `json:"work_minutes"`
}
