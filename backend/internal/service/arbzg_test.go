package service

import (
	"testing"
)

func TestValidateArbZG_NoViolations(t *testing.T) {
	// 8h work, 30min break → fine
	violations := ValidateArbZG(510, 30) // 8.5h gross, 30min break = 8h net
	if len(violations) != 0 {
		t.Errorf("expected no violations, got %d: %v", len(violations), violations)
	}
}

func TestValidateArbZG_MaxDailyHours(t *testing.T) {
	// 11h net work
	violations := ValidateArbZG(11*60+45, 45) // 11h45m gross, 45min break = 11h net
	found := false
	for _, v := range violations {
		if v.Rule == "max_daily_hours" {
			found = true
		}
	}
	if !found {
		t.Error("expected max_daily_hours violation")
	}
}

func TestValidateArbZG_Break6h(t *testing.T) {
	// 7h work with only 15min break
	violations := ValidateArbZG(7*60+15, 15) // 7h15m gross, 15min break = 7h net
	found := false
	for _, v := range violations {
		if v.Rule == "break_6h" {
			found = true
		}
	}
	if !found {
		t.Error("expected break_6h violation")
	}
}

func TestValidateArbZG_Break9h(t *testing.T) {
	// 9.5h work with only 30min break
	violations := ValidateArbZG(10*60, 30) // 10h gross, 30min break = 9h30m net
	found := false
	for _, v := range violations {
		if v.Rule == "break_9h" {
			found = true
		}
	}
	if !found {
		t.Error("expected break_9h violation")
	}
}

func TestValidateArbZG_Exactly6h_NoBreakRequired(t *testing.T) {
	// Exactly 6h net = no break required (rule is > 6h)
	violations := ValidateArbZG(360, 0) // 6h gross, 0 break = 6h net
	for _, v := range violations {
		if v.Rule == "break_6h" {
			t.Error("should not require break for exactly 6h")
		}
	}
}

func TestValidateArbZG_Exactly10h_OK(t *testing.T) {
	// 10h net = exactly at the limit, should be OK
	violations := ValidateArbZG(10*60+45, 45) // 10h45m gross, 45min break = 10h net
	for _, v := range violations {
		if v.Rule == "max_daily_hours" {
			t.Error("should allow exactly 10h net work")
		}
	}
}

func TestValidateTimeRange_Valid(t *testing.T) {
	minutes, err := ValidateTimeRange("08:00", "17:00")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if minutes != 540 {
		t.Errorf("expected 540 minutes, got %d", minutes)
	}
}

func TestValidateTimeRange_EndBeforeStart(t *testing.T) {
	_, err := ValidateTimeRange("17:00", "08:00")
	if err == nil {
		t.Error("expected error for end before start")
	}
}

func TestValidateTimeRange_EqualTimes(t *testing.T) {
	_, err := ValidateTimeRange("08:00", "08:00")
	if err == nil {
		t.Error("expected error for equal times")
	}
}

func TestValidateTimeRange_InvalidFormat(t *testing.T) {
	_, err := ValidateTimeRange("abc", "17:00")
	if err == nil {
		t.Error("expected error for invalid format")
	}
}
