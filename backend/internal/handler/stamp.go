package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/newa/zeiterfassung/internal/middleware"
	"github.com/newa/zeiterfassung/internal/repository"
	"github.com/newa/zeiterfassung/internal/service"
)

// StampHandler handles clock-in/out HTTP endpoints.
type StampHandler struct {
	stampRepo *repository.StampRepository
	entrySvc  *service.TimeEntryService
}

// NewStampHandler creates a new StampHandler.
func NewStampHandler(stampRepo *repository.StampRepository, entrySvc *service.TimeEntryService) *StampHandler {
	return &StampHandler{stampRepo: stampRepo, entrySvc: entrySvc}
}

type stampInRequest struct {
	ProjectID   *string `json:"project_id,omitempty"`
	Description string  `json:"description"`
}

// GetActive handles GET /api/stamp/active.
func (h *StampHandler) GetActive(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())

	stamp, err := h.stampRepo.GetActive(r.Context(), user.ID)
	if err != nil {
		log.Error().Err(err).Str("user_id", user.ID.String()).Msg("failed to get active stamp")
		ErrorJSON(w, http.StatusInternalServerError, "failed to get active stamp")
		return
	}

	if stamp == nil {
		JSON(w, http.StatusOK, nil)
		return
	}

	JSON(w, http.StatusOK, stamp)
}

// StampIn handles POST /api/stamp/in.
func (h *StampHandler) StampIn(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())

	var req stampInRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ErrorJSON(w, http.StatusBadRequest, "invalid request body")
		return
	}

	existing, err := h.stampRepo.GetActive(r.Context(), user.ID)
	if err != nil {
		log.Error().Err(err).Msg("failed to check existing stamp")
		ErrorJSON(w, http.StatusInternalServerError, "failed to check existing stamp")
		return
	}
	if existing != nil {
		ErrorJSON(w, http.StatusConflict, "already stamped in")
		return
	}

	var projectID *uuid.UUID
	if req.ProjectID != nil && *req.ProjectID != "" {
		pid, err := uuid.Parse(*req.ProjectID)
		if err != nil {
			ErrorJSON(w, http.StatusBadRequest, "invalid project_id")
			return
		}
		projectID = &pid
	}

	stamp, err := h.stampRepo.Create(r.Context(), user.ID, projectID, req.Description)
	if err != nil {
		log.Error().Err(err).Str("user_id", user.ID.String()).Msg("failed to stamp in")
		ErrorJSON(w, http.StatusInternalServerError, "failed to stamp in")
		return
	}

	log.Info().Str("user_id", user.ID.String()).Time("started_at", stamp.StartedAt).Msg("stamped in")
	JSON(w, http.StatusCreated, stamp)
}

// StampOut handles POST /api/stamp/out.
// Finalizes the active stamp by creating a time entry and deleting the stamp.
func (h *StampHandler) StampOut(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	now := time.Now()

	stamp, err := h.stampRepo.GetActive(r.Context(), user.ID)
	if err != nil {
		log.Error().Err(err).Msg("failed to get active stamp")
		ErrorJSON(w, http.StatusInternalServerError, "failed to get active stamp")
		return
	}
	if stamp == nil {
		ErrorJSON(w, http.StatusConflict, "not stamped in")
		return
	}

	// Finalize any ongoing break.
	if stamp.OnBreak() {
		elapsed := int(now.Sub(*stamp.BreakStart).Minutes())
		if err := h.stampRepo.AccumulateBreak(r.Context(), user.ID, elapsed); err != nil {
			log.Error().Err(err).Msg("failed to accumulate break")
			ErrorJSON(w, http.StatusInternalServerError, "failed to finalize break")
			return
		}
		stamp.BreakMinutes += elapsed
		stamp.BreakStart = nil
	}

	startTime := stamp.StartedAt.Format("15:04")
	endTime := now.Format("15:04")
	entryDate := stamp.StartedAt.Format("2006-01-02")

	// Ensure at least 1 minute of work.
	if startTime == endTime {
		ErrorJSON(w, http.StatusBadRequest, "stamp duration too short")
		return
	}

	entry, violations, err := h.entrySvc.Create(r.Context(), service.CreateInput{
		UserID:       user.ID,
		EntryDate:    entryDate,
		StartTime:    startTime,
		EndTime:      endTime,
		BreakMinutes: stamp.BreakMinutes,
		ProjectID:    stamp.ProjectID,
		Description:  stamp.Description,
	})
	if err != nil {
		log.Error().Err(err).Str("user_id", user.ID.String()).Msg("failed to create time entry from stamp")
		ErrorJSON(w, http.StatusBadRequest, fmt.Sprintf("stamp-out failed: %s", err.Error()))
		return
	}

	if err := h.stampRepo.Delete(r.Context(), user.ID); err != nil {
		log.Error().Err(err).Msg("failed to delete stamp after stamp-out")
		// Non-fatal: entry was already created.
	}

	log.Info().Str("user_id", user.ID.String()).Str("entry_id", entry.ID.String()).Msg("stamped out")
	JSON(w, http.StatusOK, map[string]any{"entry": entry, "warnings": violations})
}

// ToggleBreak handles POST /api/stamp/break.
// Starts a break if none is active, or ends the current break.
func (h *StampHandler) ToggleBreak(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	now := time.Now()

	stamp, err := h.stampRepo.GetActive(r.Context(), user.ID)
	if err != nil {
		log.Error().Err(err).Msg("failed to get active stamp")
		ErrorJSON(w, http.StatusInternalServerError, "failed to get active stamp")
		return
	}
	if stamp == nil {
		ErrorJSON(w, http.StatusConflict, "not stamped in")
		return
	}

	if stamp.OnBreak() {
		// End break: accumulate elapsed minutes.
		elapsed := int(now.Sub(*stamp.BreakStart).Minutes())
		if err := h.stampRepo.AccumulateBreak(r.Context(), user.ID, elapsed); err != nil {
			log.Error().Err(err).Msg("failed to end break")
			ErrorJSON(w, http.StatusInternalServerError, "failed to end break")
			return
		}
		stamp.BreakMinutes += elapsed
		stamp.BreakStart = nil
	} else {
		// Start break.
		if err := h.stampRepo.SetBreakStart(r.Context(), user.ID, now); err != nil {
			log.Error().Err(err).Msg("failed to start break")
			ErrorJSON(w, http.StatusInternalServerError, "failed to start break")
			return
		}
		stamp.BreakStart = &now
	}

	JSON(w, http.StatusOK, stamp)
}
