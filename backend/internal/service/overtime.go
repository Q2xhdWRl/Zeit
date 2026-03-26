package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/newa/zeiterfassung/internal/model"
	"github.com/newa/zeiterfassung/internal/repository"
)

// DefaultWorkSchedule is used when no schedule is configured for a user.
var DefaultWorkSchedule = model.WorkSchedule{
	WeeklyHours:    40,
	MondayHours:    8,
	TuesdayHours:   8,
	WednesdayHours: 8,
	ThursdayHours:  8,
	FridayHours:    8,
	SaturdayHours:  0,
	SundayHours:    0,
}

// OvertimeService computes overtime and team availability.
type OvertimeService struct {
	scheduleRepo *repository.WorkScheduleRepository
	entryRepo    *repository.TimeEntryRepository
	absenceRepo  *repository.AbsenceRepository
	teamRepo     *repository.TeamRepository
	userRepo     *repository.UserRepository
}

// NewOvertimeService creates a new OvertimeService.
func NewOvertimeService(
	scheduleRepo *repository.WorkScheduleRepository,
	entryRepo *repository.TimeEntryRepository,
	absenceRepo *repository.AbsenceRepository,
	teamRepo *repository.TeamRepository,
	userRepo *repository.UserRepository,
) *OvertimeService {
	return &OvertimeService{
		scheduleRepo: scheduleRepo,
		entryRepo:    entryRepo,
		absenceRepo:  absenceRepo,
		teamRepo:     teamRepo,
		userRepo:     userRepo,
	}
}

// GetOvertimeSummary computes the overtime for a user over a date range.
func (s *OvertimeService) GetOvertimeSummary(ctx context.Context, userID uuid.UUID, from, to time.Time) (*model.OvertimeSummary, error) {
	// Get work schedule
	schedule, err := s.scheduleRepo.FindActiveSchedule(ctx, userID, from)
	if err != nil {
		return nil, fmt.Errorf("failed to find work schedule: %w", err)
	}
	if schedule == nil {
		schedule = &DefaultWorkSchedule
	}

	// Get actual worked minutes
	entries, err := s.entryRepo.ListByUserAndDateRange(ctx, userID, from, to)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch time entries: %w", err)
	}

	actualMinutes := 0
	for _, e := range entries {
		wm, _ := e.WorkMinutes()
		actualMinutes += wm
	}

	// Get absences that count as work (e.g. Homeoffice, Fortbildung)
	absences, err := s.absenceRepo.ListByUser(ctx, userID, from, to)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch absences: %w", err)
	}

	// Load absence types for counts_as_work check
	absenceTypes, err := s.absenceRepo.ListAbsenceTypes(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch absence types: %w", err)
	}
	typeMap := make(map[uuid.UUID]model.AbsenceType, len(absenceTypes))
	for _, at := range absenceTypes {
		typeMap[at.ID] = at
	}

	// Build set of absent dates (approved, non-work absences)
	absentDates := make(map[string]bool)
	workAbsenceDates := make(map[string]bool)
	for _, a := range absences {
		if a.Status != model.AbsenceStatusApproved {
			continue
		}
		at := typeMap[a.AbsenceTypeID]
		startDate := a.StartDate
		if startDate.Before(from) {
			startDate = from
		}
		endDate := a.EndDate
		if endDate.After(to) {
			endDate = to
		}
		for d := startDate; !d.After(endDate); d = d.AddDate(0, 0, 1) {
			ds := d.Format("2006-01-02")
			if at.CountsAsWork {
				workAbsenceDates[ds] = true
			} else {
				absentDates[ds] = true
			}
		}
	}

	// Calculate target minutes (only working days, excluding non-work absence days)
	targetMinutes := 0
	for d := from; !d.After(to); d = d.AddDate(0, 0, 1) {
		ds := d.Format("2006-01-02")
		if absentDates[ds] {
			continue // non-work absence, no target
		}
		dayTarget := schedule.TargetMinutesForDay(d.Weekday())
		targetMinutes += dayTarget
		// If it's a work-absence day (homeoffice, etc.), add the target as actual
		if workAbsenceDates[ds] {
			actualMinutes += dayTarget
		}
	}

	return &model.OvertimeSummary{
		PeriodFrom:    from.Format("2006-01-02"),
		PeriodTo:      to.Format("2006-01-02"),
		TargetMinutes: targetMinutes,
		ActualMinutes: actualMinutes,
		DiffMinutes:   actualMinutes - targetMinutes,
	}, nil
}

// MonthlyOvertimeTrend returns overtime summaries per month for a range.
func (s *OvertimeService) MonthlyOvertimeTrend(ctx context.Context, userID uuid.UUID, fromMonth, toMonth time.Time) ([]model.OvertimeSummary, error) {
	var summaries []model.OvertimeSummary

	current := time.Date(fromMonth.Year(), fromMonth.Month(), 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(toMonth.Year(), toMonth.Month(), 1, 0, 0, 0, 0, time.UTC)

	for !current.After(end) {
		monthStart := current
		monthEnd := time.Date(current.Year(), current.Month()+1, 0, 0, 0, 0, 0, time.UTC)

		summary, err := s.GetOvertimeSummary(ctx, userID, monthStart, monthEnd)
		if err != nil {
			return nil, err
		}
		summaries = append(summaries, *summary)
		current = current.AddDate(0, 1, 0)
	}

	return summaries, nil
}

// TeamAvailability computes the availability status for each team member for each day in a range.
func (s *OvertimeService) TeamAvailability(ctx context.Context, teamID uuid.UUID, from, to time.Time) ([]model.DayAvailability, error) {
	members, err := s.teamRepo.ListMembers(ctx, teamID)
	if err != nil {
		return nil, fmt.Errorf("failed to list team members: %w", err)
	}

	// Load absence types
	absenceTypes, err := s.absenceRepo.ListAbsenceTypes(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch absence types: %w", err)
	}
	typeMap := make(map[uuid.UUID]model.AbsenceType, len(absenceTypes))
	for _, at := range absenceTypes {
		typeMap[at.ID] = at
	}

	var result []model.DayAvailability

	for _, member := range members {
		user, err := s.userRepo.FindByID(ctx, member.UserID)
		if err != nil || user == nil {
			continue
		}

		entries, err := s.entryRepo.ListByUserAndDateRange(ctx, member.UserID, from, to)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch entries for user %s: %w", member.UserID, err)
		}

		absences, err := s.absenceRepo.ListByUser(ctx, member.UserID, from, to)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch absences for user %s: %w", member.UserID, err)
		}

		// Build maps for quick lookup
		entryMinutesByDate := make(map[string]int)
		for _, e := range entries {
			ds := e.EntryDate.Format("2006-01-02")
			wm, _ := e.WorkMinutes()
			entryMinutesByDate[ds] += wm
		}

		absenceByDate := make(map[string]*model.Absence)
		for i, a := range absences {
			if a.Status != model.AbsenceStatusApproved {
				continue
			}
			for d := a.StartDate; !d.After(a.EndDate); d = d.AddDate(0, 0, 1) {
				ds := d.Format("2006-01-02")
				absenceByDate[ds] = &absences[i]
			}
		}

		for d := from; !d.After(to); d = d.AddDate(0, 0, 1) {
			ds := d.Format("2006-01-02")
			wd := d.Weekday()
			if wd == time.Saturday || wd == time.Sunday {
				continue
			}

			avail := model.DayAvailability{
				UserID:      member.UserID,
				DisplayName: user.DisplayName,
				Date:        ds,
			}

			if a, ok := absenceByDate[ds]; ok {
				at := typeMap[a.AbsenceTypeID]
				if at.Name == "Homeoffice" {
					avail.Status = "homeoffice"
				} else {
					avail.Status = "absent"
				}
				avail.AbsenceType = at.Name
			} else if mins, ok := entryMinutesByDate[ds]; ok && mins > 0 {
				avail.Status = "present"
				avail.WorkMinutes = mins
			} else {
				avail.Status = "no_entry"
			}

			result = append(result, avail)
		}
	}

	return result, nil
}

// DashboardStats holds aggregated data for the user's dashboard.
type DashboardStats struct {
	TodayMinutes     int                  `json:"today_minutes"`
	WeekMinutes      int                  `json:"week_minutes"`
	MonthOvertime    *model.OvertimeSummary `json:"month_overtime"`
	VacationBalance  *model.VacationBalance `json:"vacation_balance"`
	TeamPresentCount int                  `json:"team_present_count"`
	TeamTotalCount   int                  `json:"team_total_count"`
}

// GetDashboardStats returns aggregated dashboard data for a user.
func (s *OvertimeService) GetDashboardStats(ctx context.Context, userID uuid.UUID) (*DashboardStats, error) {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	// Today's entries
	todayEntries, err := s.entryRepo.ListByUserAndDate(ctx, userID, today)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch today's entries: %w", err)
	}
	todayMinutes := 0
	for _, e := range todayEntries {
		wm, _ := e.WorkMinutes()
		todayMinutes += wm
	}

	// This week
	weekday := int(now.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	weekStart := today.AddDate(0, 0, -(weekday - 1))
	weekEnd := weekStart.AddDate(0, 0, 6)

	weekEntries, err := s.entryRepo.ListByUserAndDateRange(ctx, userID, weekStart, weekEnd)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch week entries: %w", err)
	}
	weekMinutes := 0
	for _, e := range weekEntries {
		wm, _ := e.WorkMinutes()
		weekMinutes += wm
	}

	// Month overtime
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	monthEnd := monthStart.AddDate(0, 1, -1)
	monthOvertime, err := s.GetOvertimeSummary(ctx, userID, monthStart, monthEnd)
	if err != nil {
		monthOvertime = &model.OvertimeSummary{}
	}

	stats := &DashboardStats{
		TodayMinutes:  todayMinutes,
		WeekMinutes:   weekMinutes,
		MonthOvertime: monthOvertime,
	}

	return stats, nil
}

// UserOvertimeSummary pairs a user's display info with their overtime summary.
type UserOvertimeSummary struct {
	UserID      uuid.UUID            `json:"user_id"`
	DisplayName string               `json:"display_name"`
	Summary     *model.OvertimeSummary `json:"summary"`
}

// GetTeamOvertimeSummaries returns overtime for every member of the given team.
func (s *OvertimeService) GetTeamOvertimeSummaries(ctx context.Context, teamID uuid.UUID, from, to time.Time) ([]UserOvertimeSummary, error) {
	members, err := s.teamRepo.ListMembers(ctx, teamID)
	if err != nil {
		return nil, fmt.Errorf("failed to list team members: %w", err)
	}

	result := make([]UserOvertimeSummary, 0, len(members))
	for _, m := range members {
		user, err := s.userRepo.FindByID(ctx, m.UserID)
		if err != nil || user == nil {
			continue
		}
		summary, err := s.GetOvertimeSummary(ctx, m.UserID, from, to)
		if err != nil {
			return nil, fmt.Errorf("failed to get overtime for user %s: %w", m.UserID, err)
		}
		result = append(result, UserOvertimeSummary{
			UserID:      m.UserID,
			DisplayName: user.DisplayName,
			Summary:     summary,
		})
	}
	return result, nil
}

// GetAllUsersOvertimeSummaries returns overtime for every active user (admin use).
func (s *OvertimeService) GetAllUsersOvertimeSummaries(ctx context.Context, from, to time.Time) ([]UserOvertimeSummary, error) {
	users, err := s.userRepo.ListAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	result := make([]UserOvertimeSummary, 0, len(users))
	for _, u := range users {
		if !u.IsActive {
			continue
		}
		summary, err := s.GetOvertimeSummary(ctx, u.ID, from, to)
		if err != nil {
			return nil, fmt.Errorf("failed to get overtime for user %s: %w", u.ID, err)
		}
		result = append(result, UserOvertimeSummary{
			UserID:      u.ID,
			DisplayName: u.DisplayName,
			Summary:     summary,
		})
	}
	return result, nil
}
