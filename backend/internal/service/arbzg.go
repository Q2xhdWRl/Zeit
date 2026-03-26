package service

import (
	"fmt"
	"time"
)

// ArbZG (Arbeitszeitgesetz) validation rules.
const (
	ArbZGMaxDailyHours       = 10
	ArbZGBreakThreshold6h    = 6 * 60  // 360 minutes
	ArbZGBreakRequired6h     = 30      // minutes
	ArbZGBreakThreshold9h    = 9 * 60  // 540 minutes
	ArbZGBreakRequired9h     = 45      // minutes
	ArbZGMaxDailyMinutes     = ArbZGMaxDailyHours * 60
)

// ArbZGViolation describes a specific violation of working time regulations.
type ArbZGViolation struct {
	Rule    string `json:"rule"`
	Message string `json:"message"`
}

// ValidateArbZG checks a day's total work and break minutes against ArbZG rules.
// grossMinutes is the total time span (end - start) without subtracting breaks.
// breakMinutes is the declared break duration.
func ValidateArbZG(grossMinutes, breakMinutes int) []ArbZGViolation {
	var violations []ArbZGViolation
	netMinutes := grossMinutes - breakMinutes

	// Rule 1: Maximum 10 hours of net work per day
	if netMinutes > ArbZGMaxDailyMinutes {
		violations = append(violations, ArbZGViolation{
			Rule:    "max_daily_hours",
			Message: fmt.Sprintf("Net working time %s exceeds maximum of %d hours", formatMinutes(netMinutes), ArbZGMaxDailyHours),
		})
	}

	// Rule 2: > 6h work requires at least 30 min break
	if netMinutes > ArbZGBreakThreshold6h && breakMinutes < ArbZGBreakRequired6h {
		violations = append(violations, ArbZGViolation{
			Rule:    "break_6h",
			Message: fmt.Sprintf("Working more than 6 hours requires at least %d minutes break (got %d)", ArbZGBreakRequired6h, breakMinutes),
		})
	}

	// Rule 3: > 9h work requires at least 45 min break
	if netMinutes > ArbZGBreakThreshold9h && breakMinutes < ArbZGBreakRequired9h {
		violations = append(violations, ArbZGViolation{
			Rule:    "break_9h",
			Message: fmt.Sprintf("Working more than 9 hours requires at least %d minutes break (got %d)", ArbZGBreakRequired9h, breakMinutes),
		})
	}

	return violations
}

// ValidateTimeRange checks that start < end and returns gross minutes.
func ValidateTimeRange(startTime, endTime string) (int, error) {
	start, err := parseTime(startTime)
	if err != nil {
		return 0, fmt.Errorf("invalid start_time: %w", err)
	}
	end, err := parseTime(endTime)
	if err != nil {
		return 0, fmt.Errorf("invalid end_time: %w", err)
	}
	if !end.After(start) {
		return 0, fmt.Errorf("end_time must be after start_time")
	}
	return int(end.Sub(start).Minutes()), nil
}

func parseTime(s string) (time.Time, error) {
	t, err := time.Parse("15:04", s)
	if err != nil {
		t, err = time.Parse("15:04:05", s)
	}
	return t, err
}

func formatMinutes(m int) string {
	return fmt.Sprintf("%dh%02dm", m/60, m%60)
}
