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

// adminTestRouter creates a chi router with the admin handler mounted and user injection middleware.
func adminTestRouter(h *AdminHandler, user *model.User) *chi.Mux {
	r := chi.NewRouter()
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := middleware.ContextWithUser(r.Context(), user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	})
	r.Put("/api/admin/users/{userID}/role", h.UpdateRole)
	r.Put("/api/admin/users/{userID}/active", h.UpdateActive)
	return r
}

func TestAdminHandler_UpdateRole_RejectsSelfDemotion(t *testing.T) {
	adminUser := &model.User{
		ID:         uuid.New(),
		Email:      "admin@test.com",
		GlobalRole: model.RoleAdmin,
		IsActive:   true,
	}

	h := &AdminHandler{userRepo: nil}
	router := adminTestRouter(h, adminUser)

	body, _ := json.Marshal(updateRoleRequest{Role: model.RoleUser})
	req := httptest.NewRequest(http.MethodPut, "/api/admin/users/"+adminUser.ID.String()+"/role", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestAdminHandler_UpdateActive_RejectsSelfDeactivation(t *testing.T) {
	adminUser := &model.User{
		ID:         uuid.New(),
		Email:      "admin@test.com",
		GlobalRole: model.RoleAdmin,
		IsActive:   true,
	}

	h := &AdminHandler{userRepo: nil}
	router := adminTestRouter(h, adminUser)

	body, _ := json.Marshal(updateActiveRequest{Active: false})
	req := httptest.NewRequest(http.MethodPut, "/api/admin/users/"+adminUser.ID.String()+"/active", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestAdminHandler_UpdateRole_InvalidRole(t *testing.T) {
	adminUser := &model.User{
		ID:         uuid.New(),
		Email:      "admin@test.com",
		GlobalRole: model.RoleAdmin,
		IsActive:   true,
	}

	h := &AdminHandler{userRepo: nil}
	router := adminTestRouter(h, adminUser)

	body, _ := json.Marshal(map[string]string{"role": "superadmin"})
	targetID := uuid.New()
	req := httptest.NewRequest(http.MethodPut, "/api/admin/users/"+targetID.String()+"/role", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestAdminHandler_UpdateRole_InvalidUserID(t *testing.T) {
	adminUser := &model.User{
		ID:         uuid.New(),
		Email:      "admin@test.com",
		GlobalRole: model.RoleAdmin,
		IsActive:   true,
	}

	h := &AdminHandler{userRepo: nil}
	router := adminTestRouter(h, adminUser)

	body, _ := json.Marshal(updateRoleRequest{Role: model.RoleUser})
	req := httptest.NewRequest(http.MethodPut, "/api/admin/users/not-a-uuid/role", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}
