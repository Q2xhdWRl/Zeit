package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/newa/zeiterfassung/internal/config"
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

	// Reject stamp-in if the current time falls within an existing time entry for today.
	now := time.Now().In(config.AppLocation)
	nowHM := now.Format("15:04")
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, config.AppLocation)
	overlap, err := h.entrySvc.HasEntryAt(r.Context(), user.ID, today, nowHM)
	if err != nil {
		log.Error().Err(err).Msg("failed to check entry overlap for stamp-in")
		ErrorJSON(w, http.StatusInternalServerError, "failed to check existing entries")
		return
	}
	if overlap {
		ErrorJSON(w, http.StatusConflict, "time entry already exists for this time slot")
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
// Supports midnight-crossing stamps by splitting into two entries.
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

	// Use configured app timezone consistently for all date/time formatting.
	loc := config.AppLocation
	startLocal := stamp.StartedAt.In(loc)

	startHM := startLocal.Format("15:04")
	startDate := startLocal.Format("2006-01-02")

	// ArbZG §3: cap so that existing day net minutes + new net minutes ≤ 600.
	// This allows multiple stamp cycles per day while respecting the legal daily limit.
	const arbZGMaxMinutes = 600 // 10 hours
	startDateParsed, _ := time.ParseInLocation("2006-01-02", startDate, loc)
	existingNetMinutes, err := h.entrySvc.GetDayNetMinutes(r.Context(), user.ID, startDateParsed)
	if err != nil {
		log.Error().Err(err).Msg("failed to fetch day net minutes for ArbZG check")
		ErrorJSON(w, http.StatusInternalServerError, "failed to check daily work limit")
		return
	}

	grossMinutes := int(now.Sub(stamp.StartedAt).Minutes())
	arbZGCapped := false
	remainingBudget := arbZGMaxMinutes - existingNetMinutes
	maxGrossForThisStamp := remainingBudget + stamp.BreakMinutes
	if maxGrossForThisStamp < 1 {
		// Daily limit already exhausted by previous entries — discard the stuck stamp.
		_ = h.stampRepo.Delete(r.Context(), user.ID)
		log.Warn().
			Str("user_id", user.ID.String()).
			Int("existing_net_minutes", existingNetMinutes).
			Msg("stamp discarded: ArbZG §3 daily limit already reached")
		ErrorJSON(w, http.StatusConflict, "daily work limit (ArbZG §3) already reached")
		return
	}
	if grossMinutes > maxGrossForThisStamp {
		now = stamp.StartedAt.Add(time.Duration(maxGrossForThisStamp) * time.Minute)
		arbZGCapped = true
		log.Warn().
			Str("user_id", user.ID.String()).
			Int("gross_minutes", grossMinutes).
			Int("existing_net_minutes", existingNetMinutes).
			Int("remaining_budget", remainingBudget).
			Msg("stamp-out capped at ArbZG §3 daily limit")
	}

	nowLocal := now.In(loc)
	endHM := nowLocal.Format("15:04")
	endDate := nowLocal.Format("2006-01-02")

	// Ensure at least 1 minute of work.
	if startHM == endHM && startDate == endDate {
		ErrorJSON(w, http.StatusBadRequest, "stamp duration too short")
		return
	}


	var lastEntry any
	var lastViolations any

	if startDate != endDate {
		// Midnight crossing: split into two entries.
		midnight := time.Date(startLocal.Year(), startLocal.Month(), startLocal.Day()+1, 0, 0, 0, 0, loc)
		day1Minutes := int(midnight.Sub(startLocal).Minutes())
		totalMinutes := int(now.Sub(startLocal).Minutes())

		breakDay1, breakDay2 := stamp.BreakMinutes, 0
		if totalMinutes > 0 && stamp.BreakMinutes > 0 {
			breakDay1 = stamp.BreakMinutes * day1Minutes / totalMinutes
			breakDay2 = stamp.BreakMinutes - breakDay1
		}

		// Entry for start date: start → 23:59.
		_, _, err := h.entrySvc.Create(r.Context(), service.CreateInput{
			UserID:       user.ID,
			EntryDate:    startDate,
			StartTime:    startHM,
			EndTime:      "23:59",
			BreakMinutes: breakDay1,
			ProjectID:    stamp.ProjectID,
			Description:  stamp.Description,
		})
		if err != nil {
			log.Error().Err(err).Str("user_id", user.ID.String()).Msg("failed to create day-1 entry from midnight-crossing stamp")
			stampOutEntryError(w, err)
			return
		}

		// Entry for end date: 00:00 → end.
		entry2, violations2, err := h.entrySvc.Create(r.Context(), service.CreateInput{
			UserID:       user.ID,
			EntryDate:    endDate,
			StartTime:    "00:00",
			EndTime:      endHM,
			BreakMinutes: breakDay2,
			ProjectID:    stamp.ProjectID,
			Description:  stamp.Description,
		})
		if err != nil {
			log.Error().Err(err).Str("user_id", user.ID.String()).Msg("failed to create day-2 entry from midnight-crossing stamp")
			stampOutEntryError(w, err)
			return
		}
		lastEntry = entry2
		lastViolations = violations2
	} else {
		// Same-day stamp: normal entry creation.
		entry, violations, err := h.entrySvc.Create(r.Context(), service.CreateInput{
			UserID:       user.ID,
			EntryDate:    startDate,
			StartTime:    startHM,
			EndTime:      endHM,
			BreakMinutes: stamp.BreakMinutes,
			ProjectID:    stamp.ProjectID,
			Description:  stamp.Description,
		})
		if err != nil {
			log.Error().Err(err).Str("user_id", user.ID.String()).Msg("failed to create time entry from stamp")
			stampOutEntryError(w, err)
			return
		}
		lastEntry = entry
		lastViolations = violations
	}

	if err := h.stampRepo.Delete(r.Context(), user.ID); err != nil {
		log.Error().Err(err).Msg("failed to delete stamp after stamp-out")
		// Non-fatal: entry was already created.
	}

	log.Info().Str("user_id", user.ID.String()).Msg("stamped out")
	JSON(w, http.StatusOK, map[string]any{
		"entry":          lastEntry,
		"warnings":       lastViolations,
		"arbzg_capped":   arbZGCapped,
	})
}

// Discard handles DELETE /api/stamp/active — abandons the active stamp without creating a time entry.
func (h *StampHandler) Discard(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())

	stamp, err := h.stampRepo.GetActive(r.Context(), user.ID)
	if err != nil {
		log.Error().Err(err).Msg("failed to check active stamp")
		ErrorJSON(w, http.StatusInternalServerError, "failed to check stamp")
		return
	}
	if stamp == nil {
		ErrorJSON(w, http.StatusConflict, "not stamped in")
		return
	}

	if err := h.stampRepo.Delete(r.Context(), user.ID); err != nil {
		log.Error().Err(err).Str("user_id", user.ID.String()).Msg("failed to discard stamp")
		ErrorJSON(w, http.StatusInternalServerError, "failed to discard stamp")
		return
	}

	log.Info().Str("user_id", user.ID.String()).Msg("stamp discarded")
	JSON(w, http.StatusOK, map[string]string{"status": "discarded"})
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

// stampOutEntryError writes an appropriate error response for entry-creation failures during stamp-out.
// Overlap errors get a 409 so the frontend can distinguish them and offer to discard the stamp.
func stampOutEntryError(w http.ResponseWriter, err error) {
	msg := fmt.Sprintf("stamp-out failed: %s", err.Error())
	if strings.Contains(err.Error(), "overlaps") {
		ErrorJSON(w, http.StatusConflict, msg)
		return
	}
	ErrorJSON(w, http.StatusBadRequest, msg)
}
