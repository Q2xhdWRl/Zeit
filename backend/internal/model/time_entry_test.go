package model

import (
	"testing"

	"github.com/google/uuid"
)

func TestTimeEntry_WorkMinutes(t *testing.T) {
	tests := []struct {
		name     string
		start    string
		end      string
		brk      int
		expected int
	}{
		{"8h day no break", "08:00", "16:00", 0, 480},
		{"8h day with 30min break", "08:00", "16:30", 30, 480},
		{"short 2h block", "10:00", "12:00", 0, 120},
		{"with seconds format", "08:00:00", "17:00:00", 60, 480},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := TimeEntry{
				ID:           uuid.New(),
				StartTime:    tt.start,
				EndTime:      tt.end,
				BreakMinutes: tt.brk,
			}
			got, err := e.WorkMinutes()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, got)
			}
		})
	}
}

func TestTimeEntry_WorkMinutes_InvalidTime(t *testing.T) {
	e := TimeEntry{StartTime: "invalid", EndTime: "17:00"}
	_, err := e.WorkMinutes()
	if err == nil {
		t.Error("expected error for invalid time format")
	}
}
