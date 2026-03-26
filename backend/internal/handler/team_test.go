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

func teamTestRouter(h *TeamHandler, user *model.User) *chi.Mux {
	r := chi.NewRouter()
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := middleware.ContextWithUser(r.Context(), user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	})
	r.Post("/api/teams", h.CreateTeam)
	r.Get("/api/teams", h.ListTeams)
	r.Get("/api/teams/{teamID}", h.GetTeam)
	r.Put("/api/teams/{teamID}", h.UpdateTeam)
	r.Delete("/api/teams/{teamID}", h.DeleteTeam)
	r.Get("/api/teams/{teamID}/members", h.ListMembers)
	r.Post("/api/teams/{teamID}/members", h.AddMember)
	r.Delete("/api/teams/{teamID}/members/{userID}", h.RemoveMember)
	r.Get("/api/teams/my", h.MyTeams)
	return r
}

func TestTeamHandler_CreateTeam_MissingName(t *testing.T) {
	admin := &model.User{
		ID:         uuid.New(),
		Email:      "admin@test.com",
		GlobalRole: model.RoleAdmin,
		IsActive:   true,
	}

	h := &TeamHandler{teamRepo: nil}
	router := teamTestRouter(h, admin)

	body, _ := json.Marshal(createTeamRequest{})
	req := httptest.NewRequest(http.MethodPost, "/api/teams", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestTeamHandler_GetTeam_InvalidID(t *testing.T) {
	user := &model.User{
		ID:         uuid.New(),
		Email:      "user@test.com",
		GlobalRole: model.RoleUser,
		IsActive:   true,
	}

	h := &TeamHandler{teamRepo: nil}
	router := teamTestRouter(h, user)

	req := httptest.NewRequest(http.MethodGet, "/api/teams/not-a-uuid", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestTeamHandler_UpdateTeam_InvalidID(t *testing.T) {
	admin := &model.User{
		ID:         uuid.New(),
		Email:      "admin@test.com",
		GlobalRole: model.RoleAdmin,
		IsActive:   true,
	}

	h := &TeamHandler{teamRepo: nil}
	router := teamTestRouter(h, admin)

	body, _ := json.Marshal(updateTeamRequest{Name: "New Name"})
	req := httptest.NewRequest(http.MethodPut, "/api/teams/not-a-uuid", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestTeamHandler_UpdateTeam_MissingName(t *testing.T) {
	admin := &model.User{
		ID:         uuid.New(),
		Email:      "admin@test.com",
		GlobalRole: model.RoleAdmin,
		IsActive:   true,
	}

	h := &TeamHandler{teamRepo: nil}
	router := teamTestRouter(h, admin)

	body, _ := json.Marshal(updateTeamRequest{})
	req := httptest.NewRequest(http.MethodPut, "/api/teams/"+uuid.New().String(), bytes.NewReader(body))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestTeamHandler_DeleteTeam_InvalidID(t *testing.T) {
	admin := &model.User{
		ID:         uuid.New(),
		Email:      "admin@test.com",
		GlobalRole: model.RoleAdmin,
		IsActive:   true,
	}

	h := &TeamHandler{teamRepo: nil}
	router := teamTestRouter(h, admin)

	req := httptest.NewRequest(http.MethodDelete, "/api/teams/not-a-uuid", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestTeamHandler_AddMember_InvalidTeamID(t *testing.T) {
	admin := &model.User{
		ID:         uuid.New(),
		Email:      "admin@test.com",
		GlobalRole: model.RoleAdmin,
		IsActive:   true,
	}

	h := &TeamHandler{teamRepo: nil}
	router := teamTestRouter(h, admin)

	body, _ := json.Marshal(addMemberRequest{UserID: uuid.New(), Role: model.RoleUser})
	req := httptest.NewRequest(http.MethodPost, "/api/teams/not-a-uuid/members", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestTeamHandler_AddMember_MissingUserID(t *testing.T) {
	admin := &model.User{
		ID:         uuid.New(),
		Email:      "admin@test.com",
		GlobalRole: model.RoleAdmin,
		IsActive:   true,
	}

	h := &TeamHandler{teamRepo: nil}
	router := teamTestRouter(h, admin)

	body, _ := json.Marshal(addMemberRequest{Role: model.RoleUser})
	req := httptest.NewRequest(http.MethodPost, "/api/teams/"+uuid.New().String()+"/members", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestTeamHandler_AddMember_InvalidRole(t *testing.T) {
	admin := &model.User{
		ID:         uuid.New(),
		Email:      "admin@test.com",
		GlobalRole: model.RoleAdmin,
		IsActive:   true,
	}

	h := &TeamHandler{teamRepo: nil}
	router := teamTestRouter(h, admin)

	body, _ := json.Marshal(map[string]any{
		"user_id": uuid.New().String(),
		"role":    "superuser",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/teams/"+uuid.New().String()+"/members", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestTeamHandler_RemoveMember_InvalidTeamID(t *testing.T) {
	admin := &model.User{
		ID:         uuid.New(),
		Email:      "admin@test.com",
		GlobalRole: model.RoleAdmin,
		IsActive:   true,
	}

	h := &TeamHandler{teamRepo: nil}
	router := teamTestRouter(h, admin)

	req := httptest.NewRequest(http.MethodDelete, "/api/teams/not-a-uuid/members/"+uuid.New().String(), nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestTeamHandler_RemoveMember_InvalidUserID(t *testing.T) {
	admin := &model.User{
		ID:         uuid.New(),
		Email:      "admin@test.com",
		GlobalRole: model.RoleAdmin,
		IsActive:   true,
	}

	h := &TeamHandler{teamRepo: nil}
	router := teamTestRouter(h, admin)

	req := httptest.NewRequest(http.MethodDelete, "/api/teams/"+uuid.New().String()+"/members/not-a-uuid", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}
