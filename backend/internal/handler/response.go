package handler

import (
	"encoding/json"
	"net/http"
)

// ErrorResponse represents a structured API error.
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// JSON writes a JSON response with the given status code.
func JSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

// ErrorJSON writes a JSON error response.
func ErrorJSON(w http.ResponseWriter, status int, errMsg string) {
	JSON(w, status, ErrorResponse{Error: http.StatusText(status), Message: errMsg})
}
