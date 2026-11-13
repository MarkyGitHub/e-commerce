package shared

import (
	"encoding/json"
	"net/http"
)

// ErrorResponse repräsentiert eine API-Fehlerantwort
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// PaginatedResponse repräsentiert eine paginierte API-Antwort
type PaginatedResponse struct {
	Items      interface{} `json:"items"`
	TotalCount int64       `json:"total_count"`
}

// WriteJSON schreibt eine JSON-Antwort mit gegebenem Statuscode
func WriteJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		_ = json.NewEncoder(w).Encode(data)
	}
}

// WriteError schreibt eine Fehlerantwort mit gegebenem Statuscode
func WriteError(w http.ResponseWriter, status int, message string) {
	WriteJSON(w, status, ErrorResponse{
		Error:   http.StatusText(status),
		Message: message,
	})
}

// WritePaginated schreibt eine paginierte Antwort
func WritePaginated(w http.ResponseWriter, data interface{}, totalCount int64) {
	WriteJSON(w, http.StatusOK, PaginatedResponse{
		Items:      data,
		TotalCount: totalCount,
	})
}
