package model

import (
	"testing"
	"time"
)

func TestWorkSchedule_TargetMinutesForDay(t *testing.T) {
	ws := &WorkSchedule{
		MondayHours:    8,
		TuesdayHours:   8,
		WednesdayHours: 8,
		ThursdayHours:  8,
		FridayHours:    6,
		SaturdayHours:  0,
		SundayHours:    0,
	}

	tests := []struct {
		day  time.Weekday
		want int
	}{
		{time.Monday, 480},
		{time.Tuesday, 480},
		{time.Wednesday, 480},
		{time.Thursday, 480},
		{time.Friday, 360},
		{time.Saturday, 0},
		{time.Sunday, 0},
	}

	for _, tc := range tests {
		got := ws.TargetMinutesForDay(tc.day)
		if got != tc.want {
			t.Errorf("TargetMinutesForDay(%v) = %d, want %d", tc.day, got, tc.want)
		}
	}
}

func TestWorkSchedule_TargetMinutesPartTime(t *testing.T) {
	ws := &WorkSchedule{
		MondayHours:    4,
		TuesdayHours:   4,
		WednesdayHours: 4,
		ThursdayHours:  4,
		FridayHours:    4,
		SaturdayHours:  0,
		SundayHours:    0,
	}

	got := ws.TargetMinutesForDay(time.Monday)
	if got != 240 {
		t.Errorf("expected 240 minutes for part-time Monday, got %d", got)
	}
}
