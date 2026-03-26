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

func timeEntryTestRouter(h *TimeEntryHandler, user *model.User) *chi.Mux {
	r := chi.NewRouter()
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := middleware.ContextWithUser(r.Context(), user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	})
	r.Post("/api/time-entries", h.Create)
	r.Put("/api/time-entries/{entryID}", h.Update)
	r.Delete("/api/time-entries/{entryID}", h.Delete)
	return r
}

func TestTimeEntryHandler_Create_MissingFields(t *testing.T) {
	user := &model.User{
		ID:         uuid.New(),
		Email:      "user@test.com",
		GlobalRole: model.RoleUser,
		IsActive:   true,
	}

	h := &TimeEntryHandler{svc: nil, entryRepo: nil}
	router := timeEntryTestRouter(h, user)

	// Empty body
	body, _ := json.Marshal(createTimeEntryRequest{})
	req := httptest.NewRequest(http.MethodPost, "/api/time-entries", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestTimeEntryHandler_Create_InvalidProjectID(t *testing.T) {
	user := &model.User{
		ID:         uuid.New(),
		Email:      "user@test.com",
		GlobalRole: model.RoleUser,
		IsActive:   true,
	}

	h := &TimeEntryHandler{svc: nil, entryRepo: nil}
	router := timeEntryTestRouter(h, user)

	badID := "not-a-uuid"
	body, _ := json.Marshal(map[string]any{
		"entry_date":    "2026-03-25",
		"start_time":    "08:00",
		"end_time":      "17:00",
		"break_minutes": 30,
		"project_id":    badID,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/time-entries", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestTimeEntryHandler_Update_InvalidEntryID(t *testing.T) {
	user := &model.User{
		ID:         uuid.New(),
		Email:      "user@test.com",
		GlobalRole: model.RoleUser,
		IsActive:   true,
	}

	h := &TimeEntryHandler{svc: nil, entryRepo: nil}
	router := timeEntryTestRouter(h, user)

	body, _ := json.Marshal(createTimeEntryRequest{
		EntryDate: "2026-03-25",
		StartTime: "08:00",
		EndTime:   "17:00",
	})
	req := httptest.NewRequest(http.MethodPut, "/api/time-entries/not-a-uuid", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestTimeEntryHandler_Delete_InvalidEntryID(t *testing.T) {
	user := &model.User{
		ID:         uuid.New(),
		Email:      "user@test.com",
		GlobalRole: model.RoleUser,
		IsActive:   true,
	}

	h := &TimeEntryHandler{svc: nil, entryRepo: nil}
	router := timeEntryTestRouter(h, user)

	req := httptest.NewRequest(http.MethodDelete, "/api/time-entries/not-a-uuid", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestTimeEntryHandler_Create_InvalidBody(t *testing.T) {
	user := &model.User{
		ID:         uuid.New(),
		Email:      "user@test.com",
		GlobalRole: model.RoleUser,
		IsActive:   true,
	}

	h := &TimeEntryHandler{svc: nil, entryRepo: nil}
	router := timeEntryTestRouter(h, user)

	req := httptest.NewRequest(http.MethodPost, "/api/time-entries", bytes.NewReader([]byte("not json")))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}
