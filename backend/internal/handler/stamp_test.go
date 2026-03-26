package handler

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/newa/zeiterfassung/internal/middleware"
	"github.com/newa/zeiterfassung/internal/model"
)

func stampTestRouter(h *StampHandler, user *model.User) *chi.Mux {
	r := chi.NewRouter()
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := middleware.ContextWithUser(r.Context(), user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	})
	r.Get("/api/stamp/active", h.GetActive)
	r.Post("/api/stamp/in", h.StampIn)
	r.Post("/api/stamp/out", h.StampOut)
	r.Post("/api/stamp/break", h.ToggleBreak)
	return r
}

func TestStampHandler_StampIn_InvalidRequestBody(t *testing.T) {
	user := &model.User{
		ID:         uuid.New(),
		Email:      "user@test.com",
		GlobalRole: model.RoleUser,
		IsActive:   true,
	}

	h := &StampHandler{stampRepo: nil, entrySvc: nil}
	router := stampTestRouter(h, user)

	req := httptest.NewRequest(http.MethodPost, "/api/stamp/in", bytes.NewReader([]byte("not json")))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid body, got %d", rec.Code)
	}
}
