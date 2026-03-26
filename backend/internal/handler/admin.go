package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/newa/zeiterfassung/internal/model"
	"github.com/newa/zeiterfassung/internal/middleware"
	"github.com/newa/zeiterfassung/internal/repository"
)

// AdminHandler handles admin-only user management endpoints.
type AdminHandler struct {
	userRepo *repository.UserRepository
}

// NewAdminHandler creates a new AdminHandler.
func NewAdminHandler(userRepo *repository.UserRepository) *AdminHandler {
	return &AdminHandler{userRepo: userRepo}
}

// ListUsers returns all users (admin only).
func (h *AdminHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.userRepo.ListAll(r.Context())
	if err != nil {
		log.Error().Err(err).Msg("failed to list users")
		ErrorJSON(w, http.StatusInternalServerError, "failed to list users")
		return
	}
	JSON(w, http.StatusOK, users)
}

// GetUser returns a single user by ID (admin only).
func (h *AdminHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	userID, err := uuid.Parse(chi.URLParam(r, "userID"))
	if err != nil {
		ErrorJSON(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	user, err := h.userRepo.FindByID(r.Context(), userID)
	if err != nil {
		log.Error().Err(err).Msg("failed to get user")
		ErrorJSON(w, http.StatusInternalServerError, "failed to get user")
		return
	}
	if user == nil {
		ErrorJSON(w, http.StatusNotFound, "user not found")
		return
	}

	JSON(w, http.StatusOK, user)
}

type updateRoleRequest struct {
	Role model.UserRole `json:"role"`
}

// UpdateRole changes a user's global role (admin only).
func (h *AdminHandler) UpdateRole(w http.ResponseWriter, r *http.Request) {
	userID, err := uuid.Parse(chi.URLParam(r, "userID"))
	if err != nil {
		ErrorJSON(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	var req updateRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ErrorJSON(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if !req.Role.IsValid() {
		ErrorJSON(w, http.StatusBadRequest, "invalid role: must be admin, team_leader, or user")
		return
	}

	// Prevent admin from removing their own admin role
	currentUser := middleware.UserFromContext(r.Context())
	if currentUser.ID == userID && req.Role != model.RoleAdmin {
		ErrorJSON(w, http.StatusBadRequest, "cannot remove your own admin role")
		return
	}

	if err := h.userRepo.UpdateRole(r.Context(), userID, req.Role); err != nil {
		log.Error().Err(err).Str("user_id", userID.String()).Msg("failed to update role")
		ErrorJSON(w, http.StatusInternalServerError, "failed to update role")
		return
	}

	log.Info().
		Str("admin_id", currentUser.ID.String()).
		Str("target_user_id", userID.String()).
		Str("new_role", string(req.Role)).
		Msg("user role updated")

	JSON(w, http.StatusOK, map[string]string{"status": "role_updated"})
}

type updateActiveRequest struct {
	Active bool `json:"active"`
}

// UpdateActive activates or deactivates a user (admin only).
func (h *AdminHandler) UpdateActive(w http.ResponseWriter, r *http.Request) {
	userID, err := uuid.Parse(chi.URLParam(r, "userID"))
	if err != nil {
		ErrorJSON(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	var req updateActiveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ErrorJSON(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Prevent admin from deactivating themselves
	currentUser := middleware.UserFromContext(r.Context())
	if currentUser.ID == userID && !req.Active {
		ErrorJSON(w, http.StatusBadRequest, "cannot deactivate your own account")
		return
	}

	if err := h.userRepo.UpdateActive(r.Context(), userID, req.Active); err != nil {
		log.Error().Err(err).Str("user_id", userID.String()).Msg("failed to update active status")
		ErrorJSON(w, http.StatusInternalServerError, "failed to update active status")
		return
	}

	log.Info().
		Str("admin_id", currentUser.ID.String()).
		Str("target_user_id", userID.String()).
		Bool("active", req.Active).
		Msg("user active status updated")

	JSON(w, http.StatusOK, map[string]string{"status": "active_updated"})
}
