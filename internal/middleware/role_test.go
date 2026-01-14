package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"

	"github.com/mschl/ecommerce-api/internal/auth"
	"github.com/mschl/ecommerce-api/internal/shared"
)

func createRequestWithUser(r *http.Request, userID uuid.UUID, role string) *http.Request {
	user := &auth.User{
		ID:    userID,
		Email: "test@example.com",
		Username:  "Test User",
		Role:  role,
	}
	ctx := context.WithValue(r.Context(), shared.UserContextKey, user)
	return r.WithContext(ctx)
}

func TestRequireRole(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name           string
		userRole       string
		requiredRole   string
		expectedStatus int
		handlerCalled  bool
	}{
		{
			name:           "customer accessing customer endpoint",
			userRole:       auth.RoleCustomer,
			requiredRole:   auth.RoleCustomer,
			expectedStatus: http.StatusOK,
			handlerCalled:  true,
		},
		{
			name:           "admin accessing customer endpoint",
			userRole:       auth.RoleAdmin,
			requiredRole:   auth.RoleCustomer,
			expectedStatus: http.StatusOK,
			handlerCalled:  true,
		},
		{
			name:           "customer accessing admin endpoint",
			userRole:       auth.RoleCustomer,
			requiredRole:   auth.RoleAdmin,
			expectedStatus: http.StatusForbidden,
			handlerCalled:  false,
		},
		{
			name:           "admin accessing admin endpoint",
			userRole:       auth.RoleAdmin,
			requiredRole:   auth.RoleAdmin,
			expectedStatus: http.StatusOK,
			handlerCalled:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handlerCalled := false
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				handlerCalled = true
				w.WriteHeader(http.StatusOK)
			})

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req = createRequestWithUser(req, userID, tt.userRole)
			rr := httptest.NewRecorder()

			RequireRole(tt.requiredRole)(nextHandler).ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
			}
			if handlerCalled != tt.handlerCalled {
				t.Errorf("Handler called = %v, want %v", handlerCalled, tt.handlerCalled)
			}
		})
	}
}

func TestRequireRole_NoUserInContext(t *testing.T) {
	handlerCalled := false
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	// No user in context
	rr := httptest.NewRecorder()

	RequireRole(auth.RoleAdmin)(nextHandler).ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, rr.Code)
	}
	if handlerCalled {
		t.Error("Handler should not be called when user is not in context")
	}
}

func TestRequireAdmin(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name           string
		userRole       string
		expectedStatus int
		handlerCalled  bool
	}{
		{
			name:           "admin accessing admin endpoint",
			userRole:       auth.RoleAdmin,
			expectedStatus: http.StatusOK,
			handlerCalled:  true,
		},
		{
			name:           "customer accessing admin endpoint",
			userRole:       auth.RoleCustomer,
			expectedStatus: http.StatusForbidden,
			handlerCalled:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handlerCalled := false
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				handlerCalled = true
				w.WriteHeader(http.StatusOK)
			})

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req = createRequestWithUser(req, userID, tt.userRole)
			rr := httptest.NewRecorder()

			RequireAdmin(nextHandler).ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
			}
			if handlerCalled != tt.handlerCalled {
				t.Errorf("Handler called = %v, want %v", handlerCalled, tt.handlerCalled)
			}
		})
	}
}

// Property 5: Admin Endpoints Require Admin Role
func TestProperty_AdminEndpointsRequireAdminRole(t *testing.T) {
	userID := uuid.New()

	handlerCalled := false
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	})

	// Property: Customer cannot access admin endpoints
	t.Run("customer_denied", func(t *testing.T) {
		handlerCalled = false
		req := httptest.NewRequest(http.MethodGet, "/admin", nil)
		req = createRequestWithUser(req, userID, auth.RoleCustomer)
		rr := httptest.NewRecorder()

		RequireAdmin(nextHandler).ServeHTTP(rr, req)

		if rr.Code != http.StatusForbidden {
			t.Errorf("Expected 403, got %d", rr.Code)
		}
		if handlerCalled {
			t.Error("Handler should not be called for customer")
		}
	})

	// Property: Admin can access admin endpoints
	t.Run("admin_allowed", func(t *testing.T) {
		handlerCalled = false
		req := httptest.NewRequest(http.MethodGet, "/admin", nil)
		req = createRequestWithUser(req, userID, auth.RoleAdmin)
		rr := httptest.NewRecorder()

		RequireAdmin(nextHandler).ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("Expected 200, got %d", rr.Code)
		}
		if !handlerCalled {
			t.Error("Handler should be called for admin")
		}
	})

	// Property: Admin can also access customer endpoints
	t.Run("admin_can_access_customer_endpoints", func(t *testing.T) {
		handlerCalled = false
		req := httptest.NewRequest(http.MethodGet, "/customer", nil)
		req = createRequestWithUser(req, userID, auth.RoleAdmin)
		rr := httptest.NewRecorder()

		RequireRole(auth.RoleCustomer)(nextHandler).ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("Expected 200, got %d", rr.Code)
		}
		if !handlerCalled {
			t.Error("Handler should be called for admin accessing customer endpoint")
		}
	})
}
