package shared

import (
	"net/http"
	"net/url"
	"testing"
)

func TestParsePagination(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		wantPage int
		wantLimit int
	}{
		{
			name:      "default values",
			query:     "",
			wantPage:  DefaultPage,
			wantLimit: DefaultLimit,
		},
		{
			name:      "valid page and limit",
			query:     "page=2&limit=20",
			wantPage:  2,
			wantLimit: 20,
		},
		{
			name:      "only page specified",
			query:     "page=3",
			wantPage:  3,
			wantLimit: DefaultLimit,
		},
		{
			name:      "only limit specified",
			query:     "limit=25",
			wantPage:  DefaultPage,
			wantLimit: 25,
		},
		{
			name:      "limit exceeds max",
			query:     "limit=200",
			wantPage:  DefaultPage,
			wantLimit: MaxLimit,
		},
		{
			name:      "invalid page (negative)",
			query:     "page=-1",
			wantPage:  DefaultPage,
			wantLimit: DefaultLimit,
		},
		{
			name:      "invalid page (zero)",
			query:     "page=0",
			wantPage:  DefaultPage,
			wantLimit: DefaultLimit,
		},
		{
			name:      "invalid page (non-numeric)",
			query:     "page=abc",
			wantPage:  DefaultPage,
			wantLimit: DefaultLimit,
		},
		{
			name:      "invalid limit (negative)",
			query:     "limit=-5",
			wantPage:  DefaultPage,
			wantLimit: DefaultLimit,
		},
		{
			name:      "invalid limit (zero)",
			query:     "limit=0",
			wantPage:  DefaultPage,
			wantLimit: DefaultLimit,
		},
		{
			name:      "invalid limit (non-numeric)",
			query:     "limit=xyz",
			wantPage:  DefaultPage,
			wantLimit: DefaultLimit,
		},
		{
			name:      "limit at max boundary",
			query:     "limit=100",
			wantPage:  DefaultPage,
			wantLimit: MaxLimit,
		},
		{
			name:      "limit just below max",
			query:     "limit=99",
			wantPage:  DefaultPage,
			wantLimit: 99,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req *http.Request
			if tt.query != "" {
				u, _ := url.Parse("http://example.com/test?" + tt.query)
				req = &http.Request{URL: u}
			} else {
				u, _ := url.Parse("http://example.com/test")
				req = &http.Request{URL: u}
			}

			page, limit := ParsePagination(req)

			if page != tt.wantPage {
				t.Errorf("ParsePagination() page = %d, want %d", page, tt.wantPage)
			}

			if limit != tt.wantLimit {
				t.Errorf("ParsePagination() limit = %d, want %d", limit, tt.wantLimit)
			}
		})
	}
}

func TestCalculateOffset(t *testing.T) {
	tests := []struct {
		name   string
		page   int
		limit  int
		wantOffset int
	}{
		{
			name:   "first page",
			page:   1,
			limit:  10,
			wantOffset: 0,
		},
		{
			name:   "second page",
			page:   2,
			limit:  10,
			wantOffset: 10,
		},
		{
			name:   "third page",
			page:   3,
			limit:  20,
			wantOffset: 40,
		},
		{
			name:   "page with different limit",
			page:   5,
			limit:  25,
			wantOffset: 100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			offset := CalculateOffset(tt.page, tt.limit)
			if offset != tt.wantOffset {
				t.Errorf("CalculateOffset(%d, %d) = %d, want %d", tt.page, tt.limit, offset, tt.wantOffset)
			}
		})
	}
}
