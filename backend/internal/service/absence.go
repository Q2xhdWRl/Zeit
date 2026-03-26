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

// AbsenceService handles business logic for absences.
type AbsenceService struct {
	absenceRepo *repository.AbsenceRepository
}

// NewAbsenceService creates a new AbsenceService.
func NewAbsenceService(absenceRepo *repository.AbsenceRepository) *AbsenceService {
	return &AbsenceService{absenceRepo: absenceRepo}
}

// CreateAbsenceInput holds the data for creating an absence.
type CreateAbsenceInput struct {
	UserID        uuid.UUID
	AbsenceTypeID string
	StartDate     string // YYYY-MM-DD
	EndDate       string // YYYY-MM-DD
	Note          string
}

// Create validates and creates a new absence request.
func (s *AbsenceService) Create(ctx context.Context, input CreateAbsenceInput) (*model.Absence, error) {
	typeID, err := uuid.Parse(input.AbsenceTypeID)
	if err != nil {
		return nil, fmt.Errorf("invalid absence_type_id")
	}

	startDate, err := time.Parse("2006-01-02", input.StartDate)
	if err != nil {
		return nil, fmt.Errorf("invalid start_date format (expected YYYY-MM-DD)")
	}
	endDate, err := time.Parse("2006-01-02", input.EndDate)
	if err != nil {
		return nil, fmt.Errorf("invalid end_date format (expected YYYY-MM-DD)")
	}

	if endDate.Before(startDate) {
		return nil, fmt.Errorf("end_date must be on or after start_date")
	}

	// Validate absence type exists and is active
	absenceType, err := s.absenceRepo.FindAbsenceTypeByID(ctx, typeID)
	if err != nil {
		return nil, fmt.Errorf("absence type lookup failed: %w", err)
	}
	if absenceType == nil {
		return nil, fmt.Errorf("absence type not found")
	}
	if !absenceType.IsActive {
		return nil, fmt.Errorf("absence type is not active")
	}

	// Check for overlapping absences
	overlaps, err := s.absenceRepo.HasOverlap(ctx, input.UserID, startDate, endDate, uuid.Nil)
	if err != nil {
		return nil, fmt.Errorf("overlap check failed: %w", err)
	}
	if overlaps {
		return nil, fmt.Errorf("absence overlaps with an existing request")
	}

	// Check vacation balance if this is a vacation type
	if absenceType.Name == "Urlaub" {
		if err := s.checkVacationBalance(ctx, input.UserID, typeID, startDate, endDate, uuid.Nil); err != nil {
			return nil, err
		}
	}

	// Determine initial status: auto-approve if no approval required
	status := model.AbsenceStatusPending
	if !absenceType.RequiresApproval {
		status = model.AbsenceStatusApproved
	}

	absence, err := s.absenceRepo.Create(ctx, input.UserID, typeID, startDate, endDate, input.Note, status)
	if err != nil {
		return nil, fmt.Errorf("failed to create absence: %w", err)
	}

	log.Info().
		Str("absence_id", absence.ID.String()).
		Str("user_id", input.UserID.String()).
		Str("type", absenceType.Name).
		Str("status", string(status)).
		Msg("absence created")

	return absence, nil
}

// UpdateAbsenceInput holds the data for updating a pending absence.
type UpdateAbsenceInput struct {
	AbsenceID     uuid.UUID
	UserID        uuid.UUID
	AbsenceTypeID string
	StartDate     string
	EndDate       string
	Note          string
}

// Update validates and updates a pending absence.
func (s *AbsenceService) Update(ctx context.Context, input UpdateAbsenceInput) (*model.Absence, error) {
	existing, err := s.absenceRepo.FindByID(ctx, input.AbsenceID)
	if err != nil {
		return nil, fmt.Errorf("failed to find absence: %w", err)
	}
	if existing == nil {
		return nil, fmt.Errorf("absence not found")
	}
	if existing.UserID != input.UserID {
		return nil, fmt.Errorf("not authorized to update this absence")
	}
	if existing.Status != model.AbsenceStatusPending {
		return nil, fmt.Errorf("only pending absences can be updated")
	}

	typeID, err := uuid.Parse(input.AbsenceTypeID)
	if err != nil {
		return nil, fmt.Errorf("invalid absence_type_id")
	}

	startDate, err := time.Parse("2006-01-02", input.StartDate)
	if err != nil {
		return nil, fmt.Errorf("invalid start_date format")
	}
	endDate, err := time.Parse("2006-01-02", input.EndDate)
	if err != nil {
		return nil, fmt.Errorf("invalid end_date format")
	}
	if endDate.Before(startDate) {
		return nil, fmt.Errorf("end_date must be on or after start_date")
	}

	absenceType, err := s.absenceRepo.FindAbsenceTypeByID(ctx, typeID)
	if err != nil {
		return nil, fmt.Errorf("absence type lookup failed: %w", err)
	}
	if absenceType == nil {
		return nil, fmt.Errorf("absence type not found")
	}
	if !absenceType.IsActive {
		return nil, fmt.Errorf("absence type is not active")
	}

	// Check overlap excluding self
	overlaps, err := s.absenceRepo.HasOverlap(ctx, input.UserID, startDate, endDate, input.AbsenceID)
	if err != nil {
		return nil, fmt.Errorf("overlap check failed: %w", err)
	}
	if overlaps {
		return nil, fmt.Errorf("absence overlaps with an existing request")
	}

	if absenceType.Name == "Urlaub" {
		if err := s.checkVacationBalance(ctx, input.UserID, typeID, startDate, endDate, input.AbsenceID); err != nil {
			return nil, err
		}
	}

	updated, err := s.absenceRepo.Update(ctx, input.AbsenceID, typeID, startDate, endDate, input.Note)
	if err != nil {
		return nil, fmt.Errorf("failed to update absence: %w", err)
	}
	if updated == nil {
		return nil, fmt.Errorf("absence could not be updated (may no longer be pending)")
	}

	return updated, nil
}

// Delete removes a pending absence after authorization check.
func (s *AbsenceService) Delete(ctx context.Context, absenceID, userID uuid.UUID) error {
	existing, err := s.absenceRepo.FindByID(ctx, absenceID)
	if err != nil {
		return fmt.Errorf("failed to find absence: %w", err)
	}
	if existing == nil {
		return fmt.Errorf("absence not found")
	}
	if existing.UserID != userID {
		return fmt.Errorf("not authorized to delete this absence")
	}
	if existing.Status != model.AbsenceStatusPending {
		return fmt.Errorf("only pending absences can be deleted")
	}

	if err := s.absenceRepo.Delete(ctx, absenceID); err != nil {
		return fmt.Errorf("failed to delete absence: %w", err)
	}

	log.Info().
		Str("absence_id", absenceID.String()).
		Str("user_id", userID.String()).
		Msg("absence deleted")

	return nil
}

// ReviewInput holds the data for approving/rejecting an absence.
type ReviewInput struct {
	AbsenceID  uuid.UUID
	ReviewerID uuid.UUID
	Approve    bool
	ReviewNote string
}

// Review approves or rejects a pending absence.
func (s *AbsenceService) Review(ctx context.Context, input ReviewInput) (*model.Absence, error) {
	existing, err := s.absenceRepo.FindByID(ctx, input.AbsenceID)
	if err != nil {
		return nil, fmt.Errorf("failed to find absence: %w", err)
	}
	if existing == nil {
		return nil, fmt.Errorf("absence not found")
	}
	if existing.Status != model.AbsenceStatusPending {
		return nil, fmt.Errorf("only pending absences can be reviewed")
	}

	status := model.AbsenceStatusRejected
	if input.Approve {
		status = model.AbsenceStatusApproved
	}

	updated, err := s.absenceRepo.UpdateStatus(ctx, input.AbsenceID, status, input.ReviewerID, input.ReviewNote)
	if err != nil {
		return nil, fmt.Errorf("failed to update absence status: %w", err)
	}

	log.Info().
		Str("absence_id", input.AbsenceID.String()).
		Str("reviewer_id", input.ReviewerID.String()).
		Str("status", string(status)).
		Msg("absence reviewed")

	return updated, nil
}

// Cancel allows a user to cancel their own approved absence.
func (s *AbsenceService) Cancel(ctx context.Context, absenceID, userID uuid.UUID) (*model.Absence, error) {
	existing, err := s.absenceRepo.FindByID(ctx, absenceID)
	if err != nil {
		return nil, fmt.Errorf("failed to find absence: %w", err)
	}
	if existing == nil {
		return nil, fmt.Errorf("absence not found")
	}
	if existing.UserID != userID {
		return nil, fmt.Errorf("not authorized to cancel this absence")
	}
	if existing.Status != model.AbsenceStatusApproved && existing.Status != model.AbsenceStatusPending {
		return nil, fmt.Errorf("only pending or approved absences can be cancelled")
	}

	updated, err := s.absenceRepo.UpdateStatus(ctx, absenceID, model.AbsenceStatusCancelled, userID, "")
	if err != nil {
		return nil, fmt.Errorf("failed to cancel absence: %w", err)
	}

	log.Info().
		Str("absence_id", absenceID.String()).
		Str("user_id", userID.String()).
		Msg("absence cancelled")

	return updated, nil
}

// GetVacationBalance computes the vacation balance for a user and year.
func (s *AbsenceService) GetVacationBalance(ctx context.Context, userID uuid.UUID, year int) (*model.VacationBalance, error) {
	entitlement, err := s.absenceRepo.FindEntitlement(ctx, userID, year)
	if err != nil {
		return nil, fmt.Errorf("failed to find entitlement: %w", err)
	}

	totalDays := 30 // default
	carryOver := 0
	if entitlement != nil {
		totalDays = entitlement.TotalDays
		carryOver = entitlement.CarryOverDays
	}

	// Find the vacation type ID
	types, err := s.absenceRepo.ListAbsenceTypes(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list absence types: %w", err)
	}
	var vacationTypeID uuid.UUID
	for _, t := range types {
		if t.Name == "Urlaub" {
			vacationTypeID = t.ID
			break
		}
	}
	if vacationTypeID == uuid.Nil {
		return &model.VacationBalance{
			Year:          year,
			TotalDays:     totalDays,
			CarryOverDays: carryOver,
			RemainingDays: totalDays + carryOver,
		}, nil
	}

	approvedDays, pendingDays, err := s.absenceRepo.CountVacationDays(ctx, userID, year, vacationTypeID)
	if err != nil {
		return nil, fmt.Errorf("failed to count vacation days: %w", err)
	}

	return &model.VacationBalance{
		Year:          year,
		TotalDays:     totalDays,
		CarryOverDays: carryOver,
		UsedDays:      approvedDays,
		PendingDays:   pendingDays,
		RemainingDays: totalDays + carryOver - approvedDays - pendingDays,
	}, nil
}

func (s *AbsenceService) checkVacationBalance(ctx context.Context, userID, vacationTypeID uuid.UUID, startDate, endDate time.Time, excludeID uuid.UUID) error {
	// Calculate working days for this request
	tempAbsence := &model.Absence{StartDate: startDate, EndDate: endDate}
	requestedDays := tempAbsence.WorkingDays()

	year := startDate.Year()
	entitlement, err := s.absenceRepo.FindEntitlement(ctx, userID, year)
	if err != nil {
		return fmt.Errorf("failed to check vacation entitlement: %w", err)
	}

	totalDays := 30
	carryOver := 0
	if entitlement != nil {
		totalDays = entitlement.TotalDays
		carryOver = entitlement.CarryOverDays
	}

	approvedDays, pendingDays, err := s.absenceRepo.CountVacationDays(ctx, userID, year, vacationTypeID)
	if err != nil {
		return fmt.Errorf("failed to count vacation days: %w", err)
	}

	// If updating, subtract the days of the existing absence from the count
	if excludeID != uuid.Nil {
		existing, err := s.absenceRepo.FindByID(ctx, excludeID)
		if err == nil && existing != nil {
			existingDays := existing.WorkingDays()
			if existing.Status == model.AbsenceStatusApproved {
				approvedDays -= existingDays
			} else if existing.Status == model.AbsenceStatusPending {
				pendingDays -= existingDays
			}
		}
	}

	available := totalDays + carryOver - approvedDays - pendingDays
	if requestedDays > available {
		return fmt.Errorf("insufficient vacation days: %d requested, %d available", requestedDays, available)
	}

	return nil
}
