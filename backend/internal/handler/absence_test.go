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

func absenceTestRouter(h *AbsenceHandler, user *model.User) *chi.Mux {
	r := chi.NewRouter()
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := middleware.ContextWithUser(r.Context(), user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	})
	r.Post("/api/absences", h.Create)
	r.Put("/api/absences/{absenceID}", h.Update)
	r.Delete("/api/absences/{absenceID}", h.Delete)
	r.Post("/api/absences/{absenceID}/cancel", h.Cancel)
	return r
}

func TestAbsenceHandler_Create_MissingFields(t *testing.T) {
	user := &model.User{
		ID:         uuid.New(),
		Email:      "user@test.com",
		GlobalRole: model.RoleUser,
		IsActive:   true,
	}

	h := &AbsenceHandler{svc: nil, absenceRepo: nil}
	router := absenceTestRouter(h, user)

	body, _ := json.Marshal(createAbsenceRequest{})
	req := httptest.NewRequest(http.MethodPost, "/api/absences", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestAbsenceHandler_Create_InvalidBody(t *testing.T) {
	user := &model.User{
		ID:         uuid.New(),
		Email:      "user@test.com",
		GlobalRole: model.RoleUser,
		IsActive:   true,
	}

	h := &AbsenceHandler{svc: nil, absenceRepo: nil}
	router := absenceTestRouter(h, user)

	req := httptest.NewRequest(http.MethodPost, "/api/absences", bytes.NewReader([]byte("not json")))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestAbsenceHandler_Update_InvalidAbsenceID(t *testing.T) {
	user := &model.User{
		ID:         uuid.New(),
		Email:      "user@test.com",
		GlobalRole: model.RoleUser,
		IsActive:   true,
	}

	h := &AbsenceHandler{svc: nil, absenceRepo: nil}
	router := absenceTestRouter(h, user)

	body, _ := json.Marshal(createAbsenceRequest{
		AbsenceTypeID: uuid.New().String(),
		StartDate:     "2026-03-25",
		EndDate:       "2026-03-25",
	})
	req := httptest.NewRequest(http.MethodPut, "/api/absences/not-a-uuid", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestAbsenceHandler_Delete_InvalidAbsenceID(t *testing.T) {
	user := &model.User{
		ID:         uuid.New(),
		Email:      "user@test.com",
		GlobalRole: model.RoleUser,
		IsActive:   true,
	}

	h := &AbsenceHandler{svc: nil, absenceRepo: nil}
	router := absenceTestRouter(h, user)

	req := httptest.NewRequest(http.MethodDelete, "/api/absences/not-a-uuid", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestAbsenceHandler_Cancel_InvalidAbsenceID(t *testing.T) {
	user := &model.User{
		ID:         uuid.New(),
		Email:      "user@test.com",
		GlobalRole: model.RoleUser,
		IsActive:   true,
	}

	h := &AbsenceHandler{svc: nil, absenceRepo: nil}
	router := absenceTestRouter(h, user)

	req := httptest.NewRequest(http.MethodPost, "/api/absences/not-a-uuid/cancel", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}
