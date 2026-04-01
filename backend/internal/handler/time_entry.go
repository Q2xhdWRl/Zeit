package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/newa/zeiterfassung/internal/config"
	"github.com/newa/zeiterfassung/internal/middleware"
	"github.com/newa/zeiterfassung/internal/repository"
	"github.com/newa/zeiterfassung/internal/service"
)

// TimeEntryHandler handles HTTP requests for time entries.
type TimeEntryHandler struct {
	svc       *service.TimeEntryService
	entryRepo *repository.TimeEntryRepository
}

// NewTimeEntryHandler creates a new TimeEntryHandler.
func NewTimeEntryHandler(svc *service.TimeEntryService, entryRepo *repository.TimeEntryRepository) *TimeEntryHandler {
	return &TimeEntryHandler{svc: svc, entryRepo: entryRepo}
}

type createTimeEntryRequest struct {
	EntryDate    string  `json:"entry_date"`
	StartTime    string  `json:"start_time"`
	EndTime      string  `json:"end_time"`
	BreakMinutes int     `json:"break_minutes"`
	ProjectID    *string `json:"project_id,omitempty"`
	Description  string  `json:"description"`
}

type timeEntryResponse struct {
	Entry      any                     `json:"entry"`
	Warnings   []service.ArbZGViolation `json:"warnings,omitempty"`
}

// Create handles POST /api/time-entries.
func (h *TimeEntryHandler) Create(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())

	var req createTimeEntryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ErrorJSON(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.EntryDate == "" || req.StartTime == "" || req.EndTime == "" {
		ErrorJSON(w, http.StatusBadRequest, "entry_date, start_time, and end_time are required")
		return
	}

	var projectID *uuid.UUID
	if req.ProjectID != nil && *req.ProjectID != "" {
		pid, err := uuid.Parse(*req.ProjectID)
		if err != nil {
			ErrorJSON(w, http.StatusBadRequest, "invalid project_id")
			return
		}
		projectID = &pid
	}

	entry, violations, err := h.svc.Create(r.Context(), service.CreateInput{
		UserID:       user.ID,
		EntryDate:    req.EntryDate,
		StartTime:    req.StartTime,
		EndTime:      req.EndTime,
		BreakMinutes: req.BreakMinutes,
		ProjectID:    projectID,
		Description:  req.Description,
	})
	if err != nil {
		log.Error().Err(err).Str("user_id", user.ID.String()).Msg("failed to create time entry")
		ErrorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	JSON(w, http.StatusCreated, timeEntryResponse{Entry: entry, Warnings: violations})
}

// Update handles PUT /api/time-entries/{entryID}.
func (h *TimeEntryHandler) Update(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())

	entryID, err := uuid.Parse(chi.URLParam(r, "entryID"))
	if err != nil {
		ErrorJSON(w, http.StatusBadRequest, "invalid entry ID")
		return
	}

	var req createTimeEntryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ErrorJSON(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.EntryDate == "" || req.StartTime == "" || req.EndTime == "" {
		ErrorJSON(w, http.StatusBadRequest, "entry_date, start_time, and end_time are required")
		return
	}

	var projectID *uuid.UUID
	if req.ProjectID != nil && *req.ProjectID != "" {
		pid, err := uuid.Parse(*req.ProjectID)
		if err != nil {
			ErrorJSON(w, http.StatusBadRequest, "invalid project_id")
			return
		}
		projectID = &pid
	}

	entry, violations, err := h.svc.Update(r.Context(), service.UpdateInput{
		EntryID:      entryID,
		UserID:       user.ID,
		EntryDate:    req.EntryDate,
		StartTime:    req.StartTime,
		EndTime:      req.EndTime,
		BreakMinutes: req.BreakMinutes,
		ProjectID:    projectID,
		Description:  req.Description,
	})
	if err != nil {
		log.Error().Err(err).Str("entry_id", entryID.String()).Msg("failed to update time entry")
		ErrorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	JSON(w, http.StatusOK, timeEntryResponse{Entry: entry, Warnings: violations})
}

// Delete handles DELETE /api/time-entries/{entryID}.
func (h *TimeEntryHandler) Delete(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())

	entryID, err := uuid.Parse(chi.URLParam(r, "entryID"))
	if err != nil {
		ErrorJSON(w, http.StatusBadRequest, "invalid entry ID")
		return
	}

	if err := h.svc.Delete(r.Context(), entryID, user.ID); err != nil {
		log.Error().Err(err).Str("entry_id", entryID.String()).Msg("failed to delete time entry")
		ErrorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	JSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// ListMy handles GET /api/time-entries?from=YYYY-MM-DD&to=YYYY-MM-DD.
func (h *TimeEntryHandler) ListMy(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())

	from, to, err := parseDateRange(r)
	if err != nil {
		ErrorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	entries, err := h.entryRepo.ListByUserAndDateRange(r.Context(), user.ID, from, to)
	if err != nil {
		log.Error().Err(err).Msg("failed to list time entries")
		ErrorJSON(w, http.StatusInternalServerError, "failed to list time entries")
		return
	}
	JSON(w, http.StatusOK, entries)
}

// ListByTeam handles GET /api/time-entries/team/{teamID}?from=YYYY-MM-DD&to=YYYY-MM-DD.
func (h *TimeEntryHandler) ListByTeam(w http.ResponseWriter, r *http.Request) {
	teamID, err := uuid.Parse(chi.URLParam(r, "teamID"))
	if err != nil {
		ErrorJSON(w, http.StatusBadRequest, "invalid team ID")
		return
	}

	from, to, err := parseDateRange(r)
	if err != nil {
		ErrorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	entries, err := h.entryRepo.ListByTeamAndDateRange(r.Context(), teamID, from, to)
	if err != nil {
		log.Error().Err(err).Msg("failed to list team time entries")
		ErrorJSON(w, http.StatusInternalServerError, "failed to list team time entries")
		return
	}

	JSON(w, http.StatusOK, entries)
}

// Summary handles GET /api/time-entries/summary?from=YYYY-MM-DD&to=YYYY-MM-DD.
func (h *TimeEntryHandler) Summary(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())

	from, to, err := parseDateRange(r)
	if err != nil {
		ErrorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	summaries, err := h.entryRepo.SummaryByDateRange(r.Context(), user.ID, from, to)
	if err != nil {
		log.Error().Err(err).Msg("failed to get time entry summary")
		ErrorJSON(w, http.StatusInternalServerError, "failed to get summary")
		return
	}

	JSON(w, http.StatusOK, summaries)
}

func parseDateRange(r *http.Request) (time.Time, time.Time, error) {
	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")

	if fromStr == "" || toStr == "" {
		// Default to current week (Monday to Sunday)
		now := time.Now()
		weekday := int(now.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		from := now.AddDate(0, 0, -(weekday - 1))
		to := from.AddDate(0, 0, 6)
		return truncateDate(from), truncateDate(to), nil
	}

	from, err := time.ParseInLocation("2006-01-02", fromStr, config.AppLocation)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	to, err := time.ParseInLocation("2006-01-02", toStr, config.AppLocation)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	if to.Before(from) {
		return time.Time{}, time.Time{}, fmt.Errorf("to must be on or after from")
	}
	if to.Sub(from) > 366*24*time.Hour {
		return time.Time{}, time.Time{}, fmt.Errorf("date range cannot exceed 366 days")
	}
	return from, to, nil
}

func truncateDate(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}
