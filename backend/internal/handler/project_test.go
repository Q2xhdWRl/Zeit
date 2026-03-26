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

func projectTestRouter(h *ProjectHandler, user *model.User) *chi.Mux {
	r := chi.NewRouter()
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := middleware.ContextWithUser(r.Context(), user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	})
	r.Post("/api/projects", h.Create)
	r.Put("/api/admin/projects/{projectID}", h.Update)
	return r
}

func TestProjectHandler_Create_MissingName(t *testing.T) {
	admin := &model.User{
		ID:         uuid.New(),
		Email:      "admin@test.com",
		GlobalRole: model.RoleAdmin,
		IsActive:   true,
	}

	h := &ProjectHandler{projectRepo: nil}
	router := projectTestRouter(h, admin)

	body, _ := json.Marshal(createProjectRequest{CustomerName: "NEWA"})
	req := httptest.NewRequest(http.MethodPost, "/api/projects", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestProjectHandler_Create_InvalidBody(t *testing.T) {
	admin := &model.User{
		ID:         uuid.New(),
		Email:      "admin@test.com",
		GlobalRole: model.RoleAdmin,
		IsActive:   true,
	}

	h := &ProjectHandler{projectRepo: nil}
	router := projectTestRouter(h, admin)

	req := httptest.NewRequest(http.MethodPost, "/api/projects", bytes.NewReader([]byte("not json")))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestProjectHandler_Update_InvalidProjectID(t *testing.T) {
	admin := &model.User{
		ID:         uuid.New(),
		Email:      "admin@test.com",
		GlobalRole: model.RoleAdmin,
		IsActive:   true,
	}

	h := &ProjectHandler{projectRepo: nil}
	router := projectTestRouter(h, admin)

	body, _ := json.Marshal(updateProjectRequest{Name: "New Name"})
	req := httptest.NewRequest(http.MethodPut, "/api/admin/projects/not-a-uuid", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestProjectHandler_Update_MissingName(t *testing.T) {
	admin := &model.User{
		ID:         uuid.New(),
		Email:      "admin@test.com",
		GlobalRole: model.RoleAdmin,
		IsActive:   true,
	}

	h := &ProjectHandler{projectRepo: nil}
	router := projectTestRouter(h, admin)

	body, _ := json.Marshal(updateProjectRequest{})
	req := httptest.NewRequest(http.MethodPut, "/api/admin/projects/"+uuid.New().String(), bytes.NewReader(body))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}
