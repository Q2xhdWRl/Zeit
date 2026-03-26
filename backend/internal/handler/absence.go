package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/newa/zeiterfassung/internal/middleware"
	"github.com/newa/zeiterfassung/internal/repository"
	"github.com/newa/zeiterfassung/internal/service"
)

// AbsenceHandler handles HTTP requests for absences.
type AbsenceHandler struct {
	svc         *service.AbsenceService
	absenceRepo *repository.AbsenceRepository
}

// NewAbsenceHandler creates a new AbsenceHandler.
func NewAbsenceHandler(svc *service.AbsenceService, absenceRepo *repository.AbsenceRepository) *AbsenceHandler {
	return &AbsenceHandler{svc: svc, absenceRepo: absenceRepo}
}

type createAbsenceRequest struct {
	AbsenceTypeID string `json:"absence_type_id"`
	StartDate     string `json:"start_date"`
	EndDate       string `json:"end_date"`
	Note          string `json:"note"`
}

type reviewAbsenceRequest struct {
	Approve    bool   `json:"approve"`
	ReviewNote string `json:"review_note"`
}

// Create handles POST /api/absences.
func (h *AbsenceHandler) Create(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())

	var req createAbsenceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ErrorJSON(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.AbsenceTypeID == "" || req.StartDate == "" || req.EndDate == "" {
		ErrorJSON(w, http.StatusBadRequest, "absence_type_id, start_date, and end_date are required")
		return
	}

	absence, err := h.svc.Create(r.Context(), service.CreateAbsenceInput{
		UserID:        user.ID,
		AbsenceTypeID: req.AbsenceTypeID,
		StartDate:     req.StartDate,
		EndDate:       req.EndDate,
		Note:          req.Note,
	})
	if err != nil {
		log.Error().Err(err).Str("user_id", user.ID.String()).Msg("failed to create absence")
		ErrorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	JSON(w, http.StatusCreated, absence)
}

// Update handles PUT /api/absences/{absenceID}.
func (h *AbsenceHandler) Update(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())

	absenceID, err := uuid.Parse(chi.URLParam(r, "absenceID"))
	if err != nil {
		ErrorJSON(w, http.StatusBadRequest, "invalid absence ID")
		return
	}

	var req createAbsenceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ErrorJSON(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.AbsenceTypeID == "" || req.StartDate == "" || req.EndDate == "" {
		ErrorJSON(w, http.StatusBadRequest, "absence_type_id, start_date, and end_date are required")
		return
	}

	absence, err := h.svc.Update(r.Context(), service.UpdateAbsenceInput{
		AbsenceID:     absenceID,
		UserID:        user.ID,
		AbsenceTypeID: req.AbsenceTypeID,
		StartDate:     req.StartDate,
		EndDate:       req.EndDate,
		Note:          req.Note,
	})
	if err != nil {
		log.Error().Err(err).Str("absence_id", absenceID.String()).Msg("failed to update absence")
		ErrorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	JSON(w, http.StatusOK, absence)
}

// Delete handles DELETE /api/absences/{absenceID}.
func (h *AbsenceHandler) Delete(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())

	absenceID, err := uuid.Parse(chi.URLParam(r, "absenceID"))
	if err != nil {
		ErrorJSON(w, http.StatusBadRequest, "invalid absence ID")
		return
	}

	if err := h.svc.Delete(r.Context(), absenceID, user.ID); err != nil {
		log.Error().Err(err).Str("absence_id", absenceID.String()).Msg("failed to delete absence")
		ErrorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	JSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// Cancel handles POST /api/absences/{absenceID}/cancel.
func (h *AbsenceHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())

	absenceID, err := uuid.Parse(chi.URLParam(r, "absenceID"))
	if err != nil {
		ErrorJSON(w, http.StatusBadRequest, "invalid absence ID")
		return
	}

	absence, err := h.svc.Cancel(r.Context(), absenceID, user.ID)
	if err != nil {
		log.Error().Err(err).Str("absence_id", absenceID.String()).Msg("failed to cancel absence")
		ErrorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	JSON(w, http.StatusOK, absence)
}

// ListMy handles GET /api/absences?from=YYYY-MM-DD&to=YYYY-MM-DD.
func (h *AbsenceHandler) ListMy(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())

	from, to, err := parseAbsenceDateRange(r)
	if err != nil {
		ErrorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	absences, err := h.absenceRepo.ListByUser(r.Context(), user.ID, from, to)
	if err != nil {
		log.Error().Err(err).Msg("failed to list absences")
		ErrorJSON(w, http.StatusInternalServerError, "failed to list absences")
		return
	}

	JSON(w, http.StatusOK, absences)
}

// ListByTeam handles GET /api/absences/team/{teamID}?from=YYYY-MM-DD&to=YYYY-MM-DD.
func (h *AbsenceHandler) ListByTeam(w http.ResponseWriter, r *http.Request) {
	teamID, err := uuid.Parse(chi.URLParam(r, "teamID"))
	if err != nil {
		ErrorJSON(w, http.StatusBadRequest, "invalid team ID")
		return
	}

	from, to, err := parseAbsenceDateRange(r)
	if err != nil {
		ErrorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	absences, err := h.absenceRepo.ListByTeam(r.Context(), teamID, from, to)
	if err != nil {
		log.Error().Err(err).Msg("failed to list team absences")
		ErrorJSON(w, http.StatusInternalServerError, "failed to list team absences")
		return
	}

	JSON(w, http.StatusOK, absences)
}

// ListPending handles GET /api/absences/team/{teamID}/pending.
func (h *AbsenceHandler) ListPending(w http.ResponseWriter, r *http.Request) {
	teamID, err := uuid.Parse(chi.URLParam(r, "teamID"))
	if err != nil {
		ErrorJSON(w, http.StatusBadRequest, "invalid team ID")
		return
	}

	absences, err := h.absenceRepo.ListPendingForTeam(r.Context(), teamID)
	if err != nil {
		log.Error().Err(err).Msg("failed to list pending absences")
		ErrorJSON(w, http.StatusInternalServerError, "failed to list pending absences")
		return
	}

	JSON(w, http.StatusOK, absences)
}

// Review handles PUT /api/absences/{absenceID}/review.
func (h *AbsenceHandler) Review(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())

	absenceID, err := uuid.Parse(chi.URLParam(r, "absenceID"))
	if err != nil {
		ErrorJSON(w, http.StatusBadRequest, "invalid absence ID")
		return
	}

	var req reviewAbsenceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ErrorJSON(w, http.StatusBadRequest, "invalid request body")
		return
	}

	absence, err := h.svc.Review(r.Context(), service.ReviewInput{
		AbsenceID:  absenceID,
		ReviewerID: user.ID,
		Approve:    req.Approve,
		ReviewNote: req.ReviewNote,
	})
	if err != nil {
		log.Error().Err(err).Str("absence_id", absenceID.String()).Msg("failed to review absence")
		ErrorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	JSON(w, http.StatusOK, absence)
}

// ListAbsenceTypes handles GET /api/absence-types.
func (h *AbsenceHandler) ListAbsenceTypes(w http.ResponseWriter, r *http.Request) {
	types, err := h.absenceRepo.ListAbsenceTypes(r.Context())
	if err != nil {
		log.Error().Err(err).Msg("failed to list absence types")
		ErrorJSON(w, http.StatusInternalServerError, "failed to list absence types")
		return
	}

	JSON(w, http.StatusOK, types)
}

// ListAllAbsenceTypes handles GET /api/admin/absence-types (includes inactive).
func (h *AbsenceHandler) ListAllAbsenceTypes(w http.ResponseWriter, r *http.Request) {
	types, err := h.absenceRepo.ListAllAbsenceTypes(r.Context())
	if err != nil {
		log.Error().Err(err).Msg("failed to list all absence types")
		ErrorJSON(w, http.StatusInternalServerError, "failed to list absence types")
		return
	}

	JSON(w, http.StatusOK, types)
}

type updateAbsenceTypeRequest struct {
	Name             string `json:"name"`
	Color            string `json:"color"`
	RequiresApproval bool   `json:"requires_approval"`
	CountsAsWork     bool   `json:"counts_as_work"`
	IsActive         bool   `json:"is_active"`
	SortOrder        int    `json:"sort_order"`
}

// UpdateAbsenceType handles PUT /api/admin/absence-types/{typeID}.
func (h *AbsenceHandler) UpdateAbsenceType(w http.ResponseWriter, r *http.Request) {
	typeID, err := uuid.Parse(chi.URLParam(r, "typeID"))
	if err != nil {
		ErrorJSON(w, http.StatusBadRequest, "invalid type ID")
		return
	}

	var req updateAbsenceTypeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ErrorJSON(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Name == "" {
		ErrorJSON(w, http.StatusBadRequest, "name is required")
		return
	}

	updated, err := h.absenceRepo.UpdateAbsenceType(r.Context(), typeID, req.Name, req.Color, req.RequiresApproval, req.CountsAsWork, req.IsActive, req.SortOrder)
	if err != nil {
		log.Error().Err(err).Str("type_id", typeID.String()).Msg("failed to update absence type")
		ErrorJSON(w, http.StatusInternalServerError, "failed to update absence type")
		return
	}

	JSON(w, http.StatusOK, updated)
}

// VacationBalance handles GET /api/absences/balance?year=YYYY.
func (h *AbsenceHandler) VacationBalance(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())

	yearStr := r.URL.Query().Get("year")
	year := time.Now().Year()
	if yearStr != "" {
		parsed, err := strconv.Atoi(yearStr)
		if err != nil {
			ErrorJSON(w, http.StatusBadRequest, "invalid year")
			return
		}
		year = parsed
	}

	balance, err := h.svc.GetVacationBalance(r.Context(), user.ID, year)
	if err != nil {
		log.Error().Err(err).Msg("failed to get vacation balance")
		ErrorJSON(w, http.StatusInternalServerError, "failed to get vacation balance")
		return
	}

	JSON(w, http.StatusOK, balance)
}

type upsertEntitlementRequest struct {
	UserID        string `json:"user_id"`
	Year          int    `json:"year"`
	TotalDays     int    `json:"total_days"`
	CarryOverDays int    `json:"carry_over_days"`
}

// UpsertEntitlement handles PUT /api/admin/entitlements.
func (h *AbsenceHandler) UpsertEntitlement(w http.ResponseWriter, r *http.Request) {
	var req upsertEntitlementRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ErrorJSON(w, http.StatusBadRequest, "invalid request body")
		return
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		ErrorJSON(w, http.StatusBadRequest, "invalid user_id")
		return
	}

	if req.Year < 2000 || req.Year > 2100 {
		ErrorJSON(w, http.StatusBadRequest, "year must be between 2000 and 2100")
		return
	}
	if req.TotalDays < 0 {
		ErrorJSON(w, http.StatusBadRequest, "total_days must be non-negative")
		return
	}

	entitlement, err := h.absenceRepo.UpsertEntitlement(r.Context(), userID, req.Year, req.TotalDays, req.CarryOverDays)
	if err != nil {
		log.Error().Err(err).Msg("failed to upsert entitlement")
		ErrorJSON(w, http.StatusInternalServerError, "failed to save entitlement")
		return
	}

	JSON(w, http.StatusOK, entitlement)
}

func parseAbsenceDateRange(r *http.Request) (time.Time, time.Time, error) {
	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")

	if fromStr == "" || toStr == "" {
		// Default to current month
		now := time.Now()
		from := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		to := from.AddDate(0, 1, -1)
		return from, to, nil
	}

	from, err := time.Parse("2006-01-02", fromStr)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	to, err := time.Parse("2006-01-02", toStr)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	return from, to, nil
}
