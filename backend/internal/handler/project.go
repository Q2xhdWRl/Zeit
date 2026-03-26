package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/newa/zeiterfassung/internal/repository"
)

// ProjectHandler handles HTTP requests for projects.
type ProjectHandler struct {
	projectRepo *repository.ProjectRepository
}

// NewProjectHandler creates a new ProjectHandler.
func NewProjectHandler(projectRepo *repository.ProjectRepository) *ProjectHandler {
	return &ProjectHandler{projectRepo: projectRepo}
}

type createProjectRequest struct {
	Name         string `json:"name"`
	CustomerName string `json:"customer_name"`
}

// Create handles POST /api/projects (admin only).
func (h *ProjectHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ErrorJSON(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name == "" {
		ErrorJSON(w, http.StatusBadRequest, "name is required")
		return
	}

	project, err := h.projectRepo.Create(r.Context(), req.Name, req.CustomerName)
	if err != nil {
		log.Error().Err(err).Msg("failed to create project")
		ErrorJSON(w, http.StatusInternalServerError, "failed to create project")
		return
	}

	JSON(w, http.StatusCreated, project)
}

// List handles GET /api/projects.
func (h *ProjectHandler) List(w http.ResponseWriter, r *http.Request) {
	projects, err := h.projectRepo.ListActive(r.Context())
	if err != nil {
		log.Error().Err(err).Msg("failed to list projects")
		ErrorJSON(w, http.StatusInternalServerError, "failed to list projects")
		return
	}
	JSON(w, http.StatusOK, projects)
}

// ListAll handles GET /api/admin/projects (includes inactive).
func (h *ProjectHandler) ListAll(w http.ResponseWriter, r *http.Request) {
	projects, err := h.projectRepo.ListAll(r.Context())
	if err != nil {
		log.Error().Err(err).Msg("failed to list all projects")
		ErrorJSON(w, http.StatusInternalServerError, "failed to list projects")
		return
	}
	JSON(w, http.StatusOK, projects)
}

type updateProjectRequest struct {
	Name         string `json:"name"`
	CustomerName string `json:"customer_name"`
	IsActive     *bool  `json:"is_active"`
}

// Update handles PUT /api/admin/projects/{projectID}.
func (h *ProjectHandler) Update(w http.ResponseWriter, r *http.Request) {
	projectID, err := uuid.Parse(chi.URLParam(r, "projectID"))
	if err != nil {
		ErrorJSON(w, http.StatusBadRequest, "invalid project ID")
		return
	}

	var req updateProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ErrorJSON(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name == "" {
		ErrorJSON(w, http.StatusBadRequest, "name is required")
		return
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	project, err := h.projectRepo.Update(r.Context(), projectID, req.Name, req.CustomerName, isActive)
	if err != nil {
		log.Error().Err(err).Msg("failed to update project")
		ErrorJSON(w, http.StatusInternalServerError, "failed to update project")
		return
	}
	if project == nil {
		ErrorJSON(w, http.StatusNotFound, "project not found")
		return
	}

	JSON(w, http.StatusOK, project)
}
