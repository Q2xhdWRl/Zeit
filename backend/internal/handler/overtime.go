package handler

import (
	"encoding/json"
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

// OvertimeHandler handles HTTP requests for overtime and team availability.
type OvertimeHandler struct {
	svc          *service.OvertimeService
	scheduleRepo *repository.WorkScheduleRepository
}

// NewOvertimeHandler creates a new OvertimeHandler.
func NewOvertimeHandler(svc *service.OvertimeService, scheduleRepo *repository.WorkScheduleRepository) *OvertimeHandler {
	return &OvertimeHandler{svc: svc, scheduleRepo: scheduleRepo}
}

// OvertimeSummary handles GET /api/overtime?from=YYYY-MM-DD&to=YYYY-MM-DD.
func (h *OvertimeHandler) OvertimeSummary(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())

	from, to, err := parseOvertimeDateRange(r)
	if err != nil {
		ErrorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	summary, err := h.svc.GetOvertimeSummary(r.Context(), user.ID, from, to)
	if err != nil {
		log.Error().Err(err).Msg("failed to get overtime summary")
		ErrorJSON(w, http.StatusInternalServerError, "failed to get overtime summary")
		return
	}

	JSON(w, http.StatusOK, summary)
}

// OvertimeTrend handles GET /api/overtime/trend?months=6.
func (h *OvertimeHandler) OvertimeTrend(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())

	now := time.Now()
	monthsBack := 6

	toMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, config.AppLocation)
	fromMonth := toMonth.AddDate(0, -(monthsBack - 1), 0)

	summaries, err := h.svc.MonthlyOvertimeTrend(r.Context(), user.ID, fromMonth, toMonth)
	if err != nil {
		log.Error().Err(err).Msg("failed to get overtime trend")
		ErrorJSON(w, http.StatusInternalServerError, "failed to get overtime trend")
		return
	}

	JSON(w, http.StatusOK, summaries)
}

// TeamAvailability handles GET /api/team/{teamID}/availability?from=YYYY-MM-DD&to=YYYY-MM-DD.
func (h *OvertimeHandler) TeamAvailability(w http.ResponseWriter, r *http.Request) {
	teamID, err := uuid.Parse(chi.URLParam(r, "teamID"))
	if err != nil {
		ErrorJSON(w, http.StatusBadRequest, "invalid team ID")
		return
	}

	from, to, err := parseOvertimeDateRange(r)
	if err != nil {
		ErrorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	availability, err := h.svc.TeamAvailability(r.Context(), teamID, from, to)
	if err != nil {
		log.Error().Err(err).Msg("failed to get team availability")
		ErrorJSON(w, http.StatusInternalServerError, "failed to get team availability")
		return
	}

	JSON(w, http.StatusOK, availability)
}

// Dashboard handles GET /api/dashboard.
func (h *OvertimeHandler) Dashboard(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())

	stats, err := h.svc.GetDashboardStats(r.Context(), user.ID)
	if err != nil {
		log.Error().Err(err).Msg("failed to get dashboard stats")
		ErrorJSON(w, http.StatusInternalServerError, "failed to get dashboard stats")
		return
	}

	JSON(w, http.StatusOK, stats)
}

// GetSchedule handles GET /api/work-schedule.
func (h *OvertimeHandler) GetSchedule(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())

	schedules, err := h.scheduleRepo.ListByUser(r.Context(), user.ID)
	if err != nil {
		log.Error().Err(err).Msg("failed to list work schedules")
		ErrorJSON(w, http.StatusInternalServerError, "failed to get work schedule")
		return
	}

	JSON(w, http.StatusOK, schedules)
}

type upsertScheduleRequest struct {
	UserID         string  `json:"user_id"`
	ValidFrom      string  `json:"valid_from"`
	WeeklyHours    float64 `json:"weekly_hours"`
	MondayHours    float64 `json:"monday_hours"`
	TuesdayHours   float64 `json:"tuesday_hours"`
	WednesdayHours float64 `json:"wednesday_hours"`
	ThursdayHours  float64 `json:"thursday_hours"`
	FridayHours    float64 `json:"friday_hours"`
	SaturdayHours  float64 `json:"saturday_hours"`
	SundayHours    float64 `json:"sunday_hours"`
}

// UpsertSchedule handles PUT /api/admin/work-schedules.
func (h *OvertimeHandler) UpsertSchedule(w http.ResponseWriter, r *http.Request) {
	var req upsertScheduleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ErrorJSON(w, http.StatusBadRequest, "invalid request body")
		return
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		ErrorJSON(w, http.StatusBadRequest, "invalid user_id")
		return
	}

	validFrom, err := time.Parse("2006-01-02", req.ValidFrom)
	if err != nil {
		ErrorJSON(w, http.StatusBadRequest, "invalid valid_from date")
		return
	}

	if req.WeeklyHours < 0 {
		ErrorJSON(w, http.StatusBadRequest, "weekly_hours must be non-negative")
		return
	}

	dailyHours := []float64{
		req.MondayHours, req.TuesdayHours, req.WednesdayHours,
		req.ThursdayHours, req.FridayHours, req.SaturdayHours, req.SundayHours,
	}
	sumDaily := 0.0
	for _, h := range dailyHours {
		if h < 0 || h > 24 {
			ErrorJSON(w, http.StatusBadRequest, "daily hours must be between 0 and 24")
			return
		}
		sumDaily += h
	}
	if diff := sumDaily - req.WeeklyHours; diff > 0.01 || diff < -0.01 {
		ErrorJSON(w, http.StatusBadRequest, "sum of daily hours must equal weekly_hours")
		return
	}

	schedule, err := h.scheduleRepo.Upsert(r.Context(), userID, validFrom,
		req.WeeklyHours, req.MondayHours, req.TuesdayHours, req.WednesdayHours,
		req.ThursdayHours, req.FridayHours, req.SaturdayHours, req.SundayHours)
	if err != nil {
		log.Error().Err(err).Msg("failed to upsert work schedule")
		ErrorJSON(w, http.StatusInternalServerError, "failed to save work schedule")
		return
	}

	JSON(w, http.StatusOK, schedule)
}

// TeamOvertimeSummary handles GET /api/overtime/team/{teamID}.
func (h *OvertimeHandler) TeamOvertimeSummary(w http.ResponseWriter, r *http.Request) {
	teamID, err := uuid.Parse(chi.URLParam(r, "teamID"))
	if err != nil {
		ErrorJSON(w, http.StatusBadRequest, "invalid team ID")
		return
	}

	from, to, err := parseOvertimeDateRange(r)
	if err != nil {
		ErrorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	summaries, err := h.svc.GetTeamOvertimeSummaries(r.Context(), teamID, from, to)
	if err != nil {
		log.Error().Err(err).Str("team_id", teamID.String()).Msg("failed to get team overtime summaries")
		ErrorJSON(w, http.StatusInternalServerError, "failed to get team overtime summaries")
		return
	}

	JSON(w, http.StatusOK, summaries)
}

// AdminOvertimeSummary handles GET /api/admin/overtime.
func (h *OvertimeHandler) AdminOvertimeSummary(w http.ResponseWriter, r *http.Request) {
	from, to, err := parseOvertimeDateRange(r)
	if err != nil {
		ErrorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	summaries, err := h.svc.GetAllUsersOvertimeSummaries(r.Context(), from, to)
	if err != nil {
		log.Error().Err(err).Msg("failed to get all users overtime summaries")
		ErrorJSON(w, http.StatusInternalServerError, "failed to get overtime summaries")
		return
	}

	JSON(w, http.StatusOK, summaries)
}

func parseOvertimeDateRange(r *http.Request) (time.Time, time.Time, error) {
	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")

	if fromStr == "" || toStr == "" {
		// Default to current month
		now := time.Now()
		from := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		to := from.AddDate(0, 1, -1)
		return from, to, nil
	}

	loc := config.AppLocation
	from, err := time.ParseInLocation("2006-01-02", fromStr, loc)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	to, err := time.ParseInLocation("2006-01-02", toStr, loc)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	return from, to, nil
}
