package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/newa/zeiterfassung/internal/model"
	"github.com/newa/zeiterfassung/internal/repository"
)

// TimeEntryService handles business logic for time entries.
type TimeEntryService struct {
	entryRepo   *repository.TimeEntryRepository
	projectRepo *repository.ProjectRepository
}

// NewTimeEntryService creates a new TimeEntryService.
func NewTimeEntryService(entryRepo *repository.TimeEntryRepository, projectRepo *repository.ProjectRepository) *TimeEntryService {
	return &TimeEntryService{entryRepo: entryRepo, projectRepo: projectRepo}
}

// CreateInput holds the data for creating a time entry.
type CreateInput struct {
	UserID       uuid.UUID
	EntryDate    string // YYYY-MM-DD
	StartTime    string // HH:MM
	EndTime      string // HH:MM
	BreakMinutes int
	ProjectID    *uuid.UUID
	Description  string
}

// Create validates and creates a new time entry.
func (s *TimeEntryService) Create(ctx context.Context, input CreateInput) (*model.TimeEntry, []ArbZGViolation, error) {
	entryDate, err := time.Parse("2006-01-02", input.EntryDate)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid entry_date format (expected YYYY-MM-DD): %w", err)
	}

	grossMinutes, err := ValidateTimeRange(input.StartTime, input.EndTime)
	if err != nil {
		return nil, nil, err
	}

	if input.BreakMinutes < 0 {
		return nil, nil, fmt.Errorf("break_minutes must be non-negative")
	}
	if input.BreakMinutes > grossMinutes {
		return nil, nil, fmt.Errorf("break_minutes cannot exceed total time span")
	}

	// Check for overlapping entries on the same day
	overlaps, err := s.entryRepo.HasOverlap(ctx, input.UserID, entryDate, input.StartTime, input.EndTime, uuid.Nil)
	if err != nil {
		return nil, nil, fmt.Errorf("overlap check failed: %w", err)
	}
	if overlaps {
		return nil, nil, fmt.Errorf("time entry overlaps with an existing entry")
	}

	// Validate project exists and is active
	if input.ProjectID != nil {
		project, err := s.projectRepo.FindByID(ctx, *input.ProjectID)
		if err != nil {
			return nil, nil, fmt.Errorf("project lookup failed: %w", err)
		}
		if project == nil {
			return nil, nil, fmt.Errorf("project not found")
		}
		if !project.IsActive {
			return nil, nil, fmt.Errorf("project is not active")
		}
	}

	// ArbZG: check all entries for this user on this day including the new one
	existingEntries, err := s.entryRepo.ListByUserAndDate(ctx, input.UserID, entryDate)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to fetch existing entries: %w", err)
	}

	totalGross := grossMinutes
	totalBreak := input.BreakMinutes
	for _, e := range existingEntries {
		eGross, _ := ValidateTimeRange(e.StartTime, e.EndTime)
		totalGross += eGross
		totalBreak += e.BreakMinutes
	}

	violations := ValidateArbZG(totalGross, totalBreak)
	for _, v := range violations {
		if v.Rule == "max_daily_hours" {
			return nil, violations, fmt.Errorf("daily limit exceeded: net working time for this day would exceed 10 hours (ArbZG §3)")
		}
	}

	entry, err := s.entryRepo.Create(ctx, input.UserID, entryDate, input.StartTime, input.EndTime, input.BreakMinutes, input.ProjectID, input.Description)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create time entry: %w", err)
	}

	// Audit log
	s.entryRepo.CreateAudit(ctx, entry.ID, input.UserID, "create", nil, entry, input.UserID)

	return entry, violations, nil
}

// UpdateInput holds the data for updating a time entry.
type UpdateInput struct {
	EntryID      uuid.UUID
	UserID       uuid.UUID // the authenticated user
	EntryDate    string
	StartTime    string
	EndTime      string
	BreakMinutes int
	ProjectID    *uuid.UUID
	Description  string
}

// Update validates and updates an existing time entry.
func (s *TimeEntryService) Update(ctx context.Context, input UpdateInput) (*model.TimeEntry, []ArbZGViolation, error) {
	existing, err := s.entryRepo.FindByID(ctx, input.EntryID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to find entry: %w", err)
	}
	if existing == nil {
		return nil, nil, fmt.Errorf("time entry not found")
	}
	if existing.UserID != input.UserID {
		return nil, nil, fmt.Errorf("not authorized to update this entry")
	}

	entryDate, err := time.Parse("2006-01-02", input.EntryDate)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid entry_date format: %w", err)
	}

	grossMinutes, err := ValidateTimeRange(input.StartTime, input.EndTime)
	if err != nil {
		return nil, nil, err
	}

	if input.BreakMinutes < 0 {
		return nil, nil, fmt.Errorf("break_minutes must be non-negative")
	}
	if input.BreakMinutes > grossMinutes {
		return nil, nil, fmt.Errorf("break_minutes cannot exceed total time span")
	}

	// Check overlap excluding self
	overlaps, err := s.entryRepo.HasOverlap(ctx, input.UserID, entryDate, input.StartTime, input.EndTime, input.EntryID)
	if err != nil {
		return nil, nil, fmt.Errorf("overlap check failed: %w", err)
	}
	if overlaps {
		return nil, nil, fmt.Errorf("time entry overlaps with an existing entry")
	}

	if input.ProjectID != nil {
		project, err := s.projectRepo.FindByID(ctx, *input.ProjectID)
		if err != nil {
			return nil, nil, fmt.Errorf("project lookup failed: %w", err)
		}
		if project == nil {
			return nil, nil, fmt.Errorf("project not found")
		}
		if !project.IsActive {
			return nil, nil, fmt.Errorf("project is not active")
		}
	}

	// ArbZG: total for the day excluding this entry
	existingEntries, err := s.entryRepo.ListByUserAndDate(ctx, input.UserID, entryDate)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to fetch existing entries: %w", err)
	}

	totalGross := grossMinutes
	totalBreak := input.BreakMinutes
	for _, e := range existingEntries {
		if e.ID == input.EntryID {
			continue
		}
		eGross, _ := ValidateTimeRange(e.StartTime, e.EndTime)
		totalGross += eGross
		totalBreak += e.BreakMinutes
	}

	violations := ValidateArbZG(totalGross, totalBreak)
	for _, v := range violations {
		if v.Rule == "max_daily_hours" {
			return nil, violations, fmt.Errorf("daily limit exceeded: net working time for this day would exceed 10 hours (ArbZG §3)")
		}
	}

	updated, err := s.entryRepo.Update(ctx, input.EntryID, entryDate, input.StartTime, input.EndTime, input.BreakMinutes, input.ProjectID, input.Description)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to update time entry: %w", err)
	}

	s.entryRepo.CreateAudit(ctx, input.EntryID, existing.UserID, "update", existing, updated, input.UserID)

	return updated, violations, nil
}

// GetDayNetMinutes returns the total net worked minutes for a user on a given date
// across all existing time entries. Used by StampOut to enforce the daily ArbZG cap.
func (s *TimeEntryService) GetDayNetMinutes(ctx context.Context, userID uuid.UUID, date time.Time) (int, error) {
	entries, err := s.entryRepo.ListByUserAndDate(ctx, userID, date)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch day entries: %w", err)
	}
	total := 0
	for _, e := range entries {
		gross, _ := ValidateTimeRange(e.StartTime, e.EndTime)
		total += gross - e.BreakMinutes
	}
	return total, nil
}

// HasEntryAt returns true if the user already has a time entry on the given date
// whose time range includes the given clock time (HH:MM). Used to prevent stamp-in
// inside an existing entry.
func (s *TimeEntryService) HasEntryAt(ctx context.Context, userID uuid.UUID, date time.Time, clockHM string) (bool, error) {
	entries, err := s.entryRepo.ListByUserAndDate(ctx, userID, date)
	if err != nil {
		return false, fmt.Errorf("failed to fetch day entries: %w", err)
	}
	clock, err := parseTime(clockHM)
	if err != nil {
		return false, fmt.Errorf("invalid clockHM: %w", err)
	}
	for _, e := range entries {
		start, err := parseTime(e.StartTime)
		if err != nil {
			continue
		}
		end, err := parseTime(e.EndTime)
		if err != nil {
			continue
		}
		// Clock is inside [start, end)
		if !clock.Before(start) && clock.Before(end) {
			return true, nil
		}
	}
	return false, nil
}

// Delete removes a time entry after authorization check.
func (s *TimeEntryService) Delete(ctx context.Context, entryID, userID uuid.UUID) error {
	existing, err := s.entryRepo.FindByID(ctx, entryID)
	if err != nil {
		return fmt.Errorf("failed to find entry: %w", err)
	}
	if existing == nil {
		return fmt.Errorf("time entry not found")
	}
	if existing.UserID != userID {
		return fmt.Errorf("not authorized to delete this entry")
	}

	if err := s.entryRepo.Delete(ctx, entryID); err != nil {
		return fmt.Errorf("failed to delete time entry: %w", err)
	}

	s.entryRepo.CreateAudit(ctx, entryID, existing.UserID, "delete", existing, nil, userID)

	log.Info().
		Str("entry_id", entryID.String()).
		Str("user_id", userID.String()).
		Msg("time entry deleted")

	return nil
}
