package model

import (
	"time"

	"github.com/google/uuid"
)

// ActiveStamp represents an in-progress work session that has been stamped in
// but not yet stamped out. At most one active stamp exists per user.
type ActiveStamp struct {
	UserID       uuid.UUID  `json:"user_id"`
	StartedAt    time.Time  `json:"started_at"`
	BreakStart   *time.Time `json:"break_start,omitempty"`
	BreakMinutes int        `json:"break_minutes"`
	ProjectID    *uuid.UUID `json:"project_id,omitempty"`
	Description  string     `json:"description"`
}

// OnBreak returns true when a break is currently active.
func (s *ActiveStamp) OnBreak() bool {
	return s.BreakStart != nil
}

// TotalBreakMinutes returns accumulated break minutes including any ongoing break.
func (s *ActiveStamp) TotalBreakMinutes(now time.Time) int {
	total := s.BreakMinutes
	if s.BreakStart != nil {
		total += int(now.Sub(*s.BreakStart).Minutes())
	}
	return total
}
