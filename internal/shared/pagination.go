package shared

import (
	"net/http"
	"strconv"
)

const (
	// DefaultPage is the default page number
	DefaultPage = 1
	// DefaultLimit is the default number of items per page
	DefaultLimit = 10
	// MaxLimit is the maximum number of items per page
	MaxLimit = 100
)

// ParsePagination extracts pagination parameters from the request
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

// CalculateOffset calculates the database offset from page and limit
func CalculateOffset(page, limit int) int {
	return (page - 1) * limit
}
