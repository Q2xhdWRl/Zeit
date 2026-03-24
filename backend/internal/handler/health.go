package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// HealthHandler provides health check endpoints.
type HealthHandler struct {
	db *pgxpool.Pool
}

// NewHealthHandler creates a new HealthHandler.
func NewHealthHandler(db *pgxpool.Pool) *HealthHandler {
	return &HealthHandler{db: db}
}

// HealthResponse represents the health check response body.
type HealthResponse struct {
	Status   string `json:"status"`
	Database string `json:"database"`
	Version  string `json:"version"`
}

// Check returns the overall system health status.
func (h *HealthHandler) Check(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	dbStatus := "up"
	if h.db == nil {
		dbStatus = "down"
	} else if err := h.db.Ping(ctx); err != nil {
		dbStatus = "down"
	}

	status := "healthy"
	httpStatus := http.StatusOK
	if dbStatus == "down" {
		status = "degraded"
		httpStatus = http.StatusServiceUnavailable
	}

	resp := HealthResponse{
		Status:   status,
		Database: dbStatus,
		Version:  "0.1.0",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)
	json.NewEncoder(w).Encode(resp)
}
