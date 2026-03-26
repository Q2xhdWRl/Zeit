package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/newa/zeiterfassung/internal/middleware"
	"github.com/newa/zeiterfassung/internal/model"
	"github.com/newa/zeiterfassung/internal/repository"
)

// TeamHandler handles team management endpoints.
type TeamHandler struct {
	teamRepo *repository.TeamRepository
}

// NewTeamHandler creates a new TeamHandler.
func NewTeamHandler(teamRepo *repository.TeamRepository) *TeamHandler {
	return &TeamHandler{teamRepo: teamRepo}
}

type createTeamRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// CreateTeam creates a new team (admin only).
func (h *TeamHandler) CreateTeam(w http.ResponseWriter, r *http.Request) {
	var req createTeamRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ErrorJSON(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name == "" {
		ErrorJSON(w, http.StatusBadRequest, "name is required")
		return
	}

	team, err := h.teamRepo.Create(r.Context(), req.Name, req.Description)
	if err != nil {
		log.Error().Err(err).Str("name", req.Name).Msg("failed to create team")
		ErrorJSON(w, http.StatusInternalServerError, "failed to create team")
		return
	}

	currentUser := middleware.UserFromContext(r.Context())
	log.Info().
		Str("admin_id", currentUser.ID.String()).
		Str("team_id", team.ID.String()).
		Str("team_name", team.Name).
		Msg("team created")

	JSON(w, http.StatusCreated, team)
}

// ListTeams returns all teams.
func (h *TeamHandler) ListTeams(w http.ResponseWriter, r *http.Request) {
	teams, err := h.teamRepo.ListAll(r.Context())
	if err != nil {
		log.Error().Err(err).Msg("failed to list teams")
		ErrorJSON(w, http.StatusInternalServerError, "failed to list teams")
		return
	}
	JSON(w, http.StatusOK, teams)
}

// GetTeam returns a single team by ID.
func (h *TeamHandler) GetTeam(w http.ResponseWriter, r *http.Request) {
	teamID, err := uuid.Parse(chi.URLParam(r, "teamID"))
	if err != nil {
		ErrorJSON(w, http.StatusBadRequest, "invalid team ID")
		return
	}

	team, err := h.teamRepo.FindByID(r.Context(), teamID)
	if err != nil {
		log.Error().Err(err).Msg("failed to get team")
		ErrorJSON(w, http.StatusInternalServerError, "failed to get team")
		return
	}
	if team == nil {
		ErrorJSON(w, http.StatusNotFound, "team not found")
		return
	}

	JSON(w, http.StatusOK, team)
}

type updateTeamRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// UpdateTeam modifies a team (admin only).
func (h *TeamHandler) UpdateTeam(w http.ResponseWriter, r *http.Request) {
	teamID, err := uuid.Parse(chi.URLParam(r, "teamID"))
	if err != nil {
		ErrorJSON(w, http.StatusBadRequest, "invalid team ID")
		return
	}

	var req updateTeamRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ErrorJSON(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name == "" {
		ErrorJSON(w, http.StatusBadRequest, "name is required")
		return
	}

	team, err := h.teamRepo.Update(r.Context(), teamID, req.Name, req.Description)
	if err != nil {
		log.Error().Err(err).Msg("failed to update team")
		ErrorJSON(w, http.StatusInternalServerError, "failed to update team")
		return
	}
	if team == nil {
		ErrorJSON(w, http.StatusNotFound, "team not found")
		return
	}

	JSON(w, http.StatusOK, team)
}

// DeleteTeam removes a team (admin only).
func (h *TeamHandler) DeleteTeam(w http.ResponseWriter, r *http.Request) {
	teamID, err := uuid.Parse(chi.URLParam(r, "teamID"))
	if err != nil {
		ErrorJSON(w, http.StatusBadRequest, "invalid team ID")
		return
	}

	if err := h.teamRepo.Delete(r.Context(), teamID); err != nil {
		log.Error().Err(err).Msg("failed to delete team")
		ErrorJSON(w, http.StatusInternalServerError, "failed to delete team")
		return
	}

	currentUser := middleware.UserFromContext(r.Context())
	log.Info().
		Str("admin_id", currentUser.ID.String()).
		Str("team_id", teamID.String()).
		Msg("team deleted")

	JSON(w, http.StatusOK, map[string]string{"status": "team_deleted"})
}

// ListMembers returns all members of a team.
func (h *TeamHandler) ListMembers(w http.ResponseWriter, r *http.Request) {
	teamID, err := uuid.Parse(chi.URLParam(r, "teamID"))
	if err != nil {
		ErrorJSON(w, http.StatusBadRequest, "invalid team ID")
		return
	}

	members, err := h.teamRepo.ListMembers(r.Context(), teamID)
	if err != nil {
		log.Error().Err(err).Msg("failed to list team members")
		ErrorJSON(w, http.StatusInternalServerError, "failed to list team members")
		return
	}

	JSON(w, http.StatusOK, members)
}

type addMemberRequest struct {
	UserID uuid.UUID      `json:"user_id"`
	Role   model.UserRole `json:"role"`
}

// AddMember adds a user to a team (admin or team leader).
func (h *TeamHandler) AddMember(w http.ResponseWriter, r *http.Request) {
	teamID, err := uuid.Parse(chi.URLParam(r, "teamID"))
	if err != nil {
		ErrorJSON(w, http.StatusBadRequest, "invalid team ID")
		return
	}

	var req addMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ErrorJSON(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.UserID == uuid.Nil {
		ErrorJSON(w, http.StatusBadRequest, "user_id is required")
		return
	}
	if !req.Role.IsValid() {
		ErrorJSON(w, http.StatusBadRequest, "invalid role: must be admin, team_leader, or user")
		return
	}

	if err := h.teamRepo.AddMember(r.Context(), teamID, req.UserID, req.Role); err != nil {
		log.Error().Err(err).Msg("failed to add team member")
		ErrorJSON(w, http.StatusInternalServerError, "failed to add team member")
		return
	}

	JSON(w, http.StatusCreated, map[string]string{"status": "member_added"})
}

// RemoveMember removes a user from a team (admin or team leader).
func (h *TeamHandler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	teamID, err := uuid.Parse(chi.URLParam(r, "teamID"))
	if err != nil {
		ErrorJSON(w, http.StatusBadRequest, "invalid team ID")
		return
	}

	userID, err := uuid.Parse(chi.URLParam(r, "userID"))
	if err != nil {
		ErrorJSON(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	if err := h.teamRepo.RemoveMember(r.Context(), teamID, userID); err != nil {
		log.Error().Err(err).Msg("failed to remove team member")
		ErrorJSON(w, http.StatusInternalServerError, "failed to remove team member")
		return
	}

	JSON(w, http.StatusOK, map[string]string{"status": "member_removed"})
}

// MyTeams returns the teams the current user belongs to.
func (h *TeamHandler) MyTeams(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	if user == nil {
		ErrorJSON(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	memberships, err := h.teamRepo.ListTeamsForUser(r.Context(), user.ID)
	if err != nil {
		log.Error().Err(err).Msg("failed to list user teams")
		ErrorJSON(w, http.StatusInternalServerError, "failed to list teams")
		return
	}

	JSON(w, http.StatusOK, memberships)
}
