package shared

import (
	"net/http"
	"strconv"
)

const (
	// DefaultPage ist die Standard-Seitennummer
	DefaultPage = 1
	// DefaultLimit ist die Standard-Anzahl pro Seite
	DefaultLimit = 10
	// MaxLimit ist die maximale Anzahl pro Seite
	MaxLimit = 100
)

// ParsePagination extrahiert Paginierungs-Parameter aus der Anfrage
func ParsePagination(r *http.Request) (page, limit int) {
	page = DefaultPage
	limit = DefaultLimit

	if p := r.URL.Query().Get("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
			if limit > MaxLimit {
				limit = MaxLimit
			}
		}
	}

	return page, limit
}

// CalculateOffset berechnet den Datenbank-Offset aus Seite und Limit
func CalculateOffset(page, limit int) int {
	return (page - 1) * limit
}
