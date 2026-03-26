package model

import (
	"testing"
	"time"
)

func TestAbsenceWorkingDays_SingleDay(t *testing.T) {
	// Wednesday 2026-03-25
	a := Absence{
		StartDate: time.Date(2026, 3, 25, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2026, 3, 25, 0, 0, 0, 0, time.UTC),
	}
	if got := a.WorkingDays(); got != 1 {
		t.Errorf("expected 1 working day, got %d", got)
	}
}

func TestAbsenceWorkingDays_FullWeek(t *testing.T) {
	// Monday to Friday 2026-03-23 to 2026-03-27
	a := Absence{
		StartDate: time.Date(2026, 3, 23, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2026, 3, 27, 0, 0, 0, 0, time.UTC),
	}
	if got := a.WorkingDays(); got != 5 {
		t.Errorf("expected 5 working days, got %d", got)
	}
}

func TestAbsenceWorkingDays_IncludesWeekend(t *testing.T) {
	// Thursday to Tuesday 2026-03-26 to 2026-03-31
	a := Absence{
		StartDate: time.Date(2026, 3, 26, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2026, 3, 31, 0, 0, 0, 0, time.UTC),
	}
	// Thu, Fri, (Sat, Sun), Mon, Tue = 4 working days
	if got := a.WorkingDays(); got != 4 {
		t.Errorf("expected 4 working days, got %d", got)
	}
}

func TestAbsenceWorkingDays_WeekendOnly(t *testing.T) {
	// Saturday to Sunday 2026-03-28 to 2026-03-29
	a := Absence{
		StartDate: time.Date(2026, 3, 28, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2026, 3, 29, 0, 0, 0, 0, time.UTC),
	}
	if got := a.WorkingDays(); got != 0 {
		t.Errorf("expected 0 working days, got %d", got)
	}
}

func TestAbsenceWorkingDays_TwoWeeks(t *testing.T) {
	// 2 full weeks: 2026-03-23 (Mon) to 2026-04-03 (Fri)
	a := Absence{
		StartDate: time.Date(2026, 3, 23, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2026, 4, 3, 0, 0, 0, 0, time.UTC),
	}
	if got := a.WorkingDays(); got != 10 {
		t.Errorf("expected 10 working days, got %d", got)
	}
}

func TestAbsenceStatus_IsValid(t *testing.T) {
	valid := []AbsenceStatus{AbsenceStatusPending, AbsenceStatusApproved, AbsenceStatusRejected, AbsenceStatusCancelled}
	for _, s := range valid {
		if !s.IsValid() {
			t.Errorf("expected %q to be valid", s)
		}
	}

	if AbsenceStatus("invalid").IsValid() {
		t.Error("expected 'invalid' to be invalid")
	}
}
