package shared

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWriteJSON(t *testing.T) {
	tests := []struct {
		name     string
		status   int
		data     interface{}
		wantCode int
	}{
		{
			name:     "write object",
			status:   http.StatusOK,
			data:     map[string]string{"key": "value"},
			wantCode: http.StatusOK,
		},
		{
			name:     "write nil data",
			status:   http.StatusCreated,
			data:     nil,
			wantCode: http.StatusCreated,
		},
		{
			name:     "write array",
			status:   http.StatusOK,
			data:     []string{"a", "b", "c"},
			wantCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			WriteJSON(rr, tt.status, tt.data)

			if rr.Code != tt.wantCode {
				t.Errorf("WriteJSON() status = %d, want %d", rr.Code, tt.wantCode)
			}

			contentType := rr.Header().Get("Content-Type")
			if contentType != "application/json" {
				t.Errorf("WriteJSON() Content-Type = %s, want application/json", contentType)
			}

			if tt.data != nil {
				var result interface{}
				if err := json.Unmarshal(rr.Body.Bytes(), &result); err != nil {
					t.Errorf("WriteJSON() response is not valid JSON: %v", err)
				}
			}
		})
	}
}

func TestWriteError(t *testing.T) {
	tests := []struct {
		name     string
		status   int
		message  string
		wantCode int
		wantErr  string
	}{
		{
			name:     "bad request error",
			status:   http.StatusBadRequest,
			message:  "invalid input",
			wantCode: http.StatusBadRequest,
			wantErr:  "Bad Request",
		},
		{
			name:     "not found error",
			status:   http.StatusNotFound,
			message:  "resource not found",
			wantCode: http.StatusNotFound,
			wantErr:  "Not Found",
		},
		{
			name:     "unauthorized error",
			status:   http.StatusUnauthorized,
			message:  "invalid token",
			wantCode: http.StatusUnauthorized,
			wantErr:  "Unauthorized",
		},
		{
			name:     "empty message",
			status:   http.StatusInternalServerError,
			message:  "",
			wantCode: http.StatusInternalServerError,
			wantErr:  "Internal Server Error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			WriteError(rr, tt.status, tt.message)

			if rr.Code != tt.wantCode {
				t.Errorf("WriteError() status = %d, want %d", rr.Code, tt.wantCode)
			}

			var response ErrorResponse
			if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
				t.Fatalf("WriteError() response is not valid JSON: %v", err)
			}

			if response.Error != tt.wantErr {
				t.Errorf("WriteError() error = %s, want %s", response.Error, tt.wantErr)
			}

			if response.Message != tt.message {
				t.Errorf("WriteError() message = %s, want %s", response.Message, tt.message)
			}
		})
	}
}

func TestWritePaginated(t *testing.T) {
	tests := []struct {
		name       string
		data       interface{}
		totalCount int64
	}{
		{
			name:       "with items",
			data:       []string{"a", "b"},
			totalCount: 20,
		},
		{
			name:       "single item",
			data:       []string{"a"},
			totalCount: 25,
		},
		{
			name:       "multiple items",
			data:       []string{"a", "b", "c"},
			totalCount: 3,
		},
		{
			name:       "empty result",
			data:       []string{},
			totalCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			WritePaginated(rr, tt.data, tt.totalCount)

			if rr.Code != http.StatusOK {
				t.Errorf("WritePaginated() status = %d, want %d", rr.Code, http.StatusOK)
			}

			var response PaginatedResponse
			if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
				t.Fatalf("WritePaginated() response is not valid JSON: %v", err)
			}

			if response.TotalCount != tt.totalCount {
				t.Errorf("WritePaginated() totalCount = %d, want %d", response.TotalCount, tt.totalCount)
			}
		})
	}
}
