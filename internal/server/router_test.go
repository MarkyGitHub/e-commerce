package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/MarkyGitHub/e-commerce/internal/db"
	"github.com/MarkyGitHub/e-commerce/internal/testutil"
)

func TestHealthCheck(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()

	HealthCheck(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
	}
	if got := rr.Body.String(); got != `{"status":"healthy"}` {
		t.Fatalf("unexpected body: %q", got)
	}
}

func TestBuildMux_HealthRoute(t *testing.T) {
	database := testutil.SetupTestDB(t)
	if err := db.Migrate(database); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	mux := BuildMux(database)
	if mux == nil {
		t.Fatal("BuildMux returned nil")
	}

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
	}
	if got := rr.Body.String(); got != `{"status":"healthy"}` {
		t.Fatalf("unexpected body: %q", got)
	}
}

func TestBuildMux_PublicAndProtectedRoutes(t *testing.T) {
	database := testutil.SetupTestDB(t)
	if err := db.Migrate(database); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	mux := BuildMux(database)

	tests := []struct {
		name       string
		method     string
		path       string
		wantStatus int
	}{
		{name: "list products public", method: http.MethodGet, path: "/products", wantStatus: http.StatusOK},
		{name: "create product requires auth", method: http.MethodPost, path: "/products", wantStatus: http.StatusUnauthorized},
		{name: "get cart requires auth", method: http.MethodGet, path: "/cart", wantStatus: http.StatusUnauthorized},
		{name: "list orders requires auth", method: http.MethodGet, path: "/orders", wantStatus: http.StatusUnauthorized},
		{name: "list notifications requires auth", method: http.MethodGet, path: "/notifications", wantStatus: http.StatusUnauthorized},
		{name: "swagger ui mounted", method: http.MethodGet, path: "/swagger/index.html", wantStatus: http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			rr := httptest.NewRecorder()
			mux.ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Fatalf("expected status %d, got %d (body=%q)", tt.wantStatus, rr.Code, rr.Body.String())
			}
		})
	}
}
