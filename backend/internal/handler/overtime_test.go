package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/newa/zeiterfassung/internal/middleware"
	"github.com/newa/zeiterfassung/internal/model"
)

func overtimeTestRouter(h *OvertimeHandler, user *model.User) *chi.Mux {
	r := chi.NewRouter()
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := middleware.ContextWithUser(r.Context(), user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	})
	r.Get("/api/overtime", h.OvertimeSummary)
	r.Get("/api/overtime/trend", h.OvertimeTrend)
	r.Get("/api/dashboard", h.Dashboard)
	r.Get("/api/work-schedule", h.GetSchedule)
	r.Put("/api/admin/work-schedules", h.UpsertSchedule)
	r.Get("/api/teams/{teamID}/availability", h.TeamAvailability)
	return r
}

func TestOvertimeHandler_TeamAvailability_InvalidTeamID(t *testing.T) {
	user := &model.User{
		ID:         uuid.New(),
		GlobalRole: model.RoleTeamLeader,
		IsActive:   true,
	}

	h := &OvertimeHandler{svc: nil, scheduleRepo: nil}
	router := overtimeTestRouter(h, user)

	req := httptest.NewRequest(http.MethodGet, "/api/teams/not-a-uuid/availability?from=2026-01-01&to=2026-01-07", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestOvertimeHandler_UpsertSchedule_InvalidBody(t *testing.T) {
	user := &model.User{
		ID:         uuid.New(),
		GlobalRole: model.RoleAdmin,
		IsActive:   true,
	}

	h := &OvertimeHandler{svc: nil, scheduleRepo: nil}
	router := overtimeTestRouter(h, user)

	req := httptest.NewRequest(http.MethodPut, "/api/admin/work-schedules", bytes.NewReader([]byte("not json")))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestOvertimeHandler_UpsertSchedule_InvalidUserID(t *testing.T) {
	user := &model.User{
		ID:         uuid.New(),
		GlobalRole: model.RoleAdmin,
		IsActive:   true,
	}

	h := &OvertimeHandler{svc: nil, scheduleRepo: nil}
	router := overtimeTestRouter(h, user)

	body, _ := json.Marshal(upsertScheduleRequest{
		UserID:      "not-a-uuid",
		ValidFrom:   "2026-01-01",
		WeeklyHours: 40,
	})

	req := httptest.NewRequest(http.MethodPut, "/api/admin/work-schedules", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestOvertimeHandler_UpsertSchedule_InvalidDate(t *testing.T) {
	user := &model.User{
		ID:         uuid.New(),
		GlobalRole: model.RoleAdmin,
		IsActive:   true,
	}

	h := &OvertimeHandler{svc: nil, scheduleRepo: nil}
	router := overtimeTestRouter(h, user)

	body, _ := json.Marshal(upsertScheduleRequest{
		UserID:      uuid.New().String(),
		ValidFrom:   "not-a-date",
		WeeklyHours: 40,
	})

	req := httptest.NewRequest(http.MethodPut, "/api/admin/work-schedules", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestOvertimeHandler_UpsertSchedule_NegativeWeeklyHours(t *testing.T) {
	user := &model.User{
		ID:         uuid.New(),
		GlobalRole: model.RoleAdmin,
		IsActive:   true,
	}

	h := &OvertimeHandler{svc: nil, scheduleRepo: nil}
	router := overtimeTestRouter(h, user)

	body, _ := json.Marshal(upsertScheduleRequest{
		UserID:      uuid.New().String(),
		ValidFrom:   "2026-01-01",
		WeeklyHours: -10,
	})

	req := httptest.NewRequest(http.MethodPut, "/api/admin/work-schedules", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}
