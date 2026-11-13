package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"

	"github.com/MarkyGitHub/e-commerce/internal/auth"
	"github.com/MarkyGitHub/e-commerce/internal/shared"
)

// MockAuthService is a mock implementation of auth.AuthService for middleware testing
type MockAuthService struct {
	validateSessionFunc func(ctx context.Context, token string) (*auth.User, error)
}

func (m *MockAuthService) Register(ctx context.Context, req auth.RegisterRequest) (*auth.User, error) {
	return nil, nil
}

func (m *MockAuthService) Login(ctx context.Context, req auth.LoginRequest) (*auth.LoginResponse, error) {
	return nil, nil
}

func (m *MockAuthService) Logout(ctx context.Context, token string) error {
	return nil
}

func (m *MockAuthService) ValidateSession(ctx context.Context, token string) (*auth.User, error) {
	if m.validateSessionFunc != nil {
		return m.validateSessionFunc(ctx, token)
	}
	return nil, nil
}

func TestAuthMiddleware_Authenticate(t *testing.T) {
	tests := []struct {
		name           string
		sessionToken   string
		mockFunc       func(ctx context.Context, token string) (*auth.User, error)
		expectedStatus int
		checkContext   func(t *testing.T, r *http.Request)
	}{
		{
			name:         "valid session token",
			sessionToken: "valid-token",
			mockFunc: func(ctx context.Context, token string) (*auth.User, error) {
				return &auth.User{
					ID:    uuid.New(),
					Email: "test@example.com",
					Username:  "Test User",
					Role:  auth.RoleCustomer,
				}, nil
			},
			expectedStatus: http.StatusOK,
			checkContext: func(t *testing.T, r *http.Request) {
				user, ok := GetUserFromContext(r.Context())
				if !ok {
					t.Error("User should be in context")
					return
				}
				if user.Email != "test@example.com" {
					t.Errorf("Expected email test@example.com, got %s", user.Email)
				}
			},
		},
		{
			name:           "missing session token",
			sessionToken:   "",
			mockFunc:       nil,
			expectedStatus: http.StatusUnauthorized,
			checkContext: func(t *testing.T, r *http.Request) {
				_, ok := GetUserFromContext(r.Context())
				if ok {
					t.Error("User should not be in context when token is missing")
				}
			},
		},
		{
			name:         "invalid session token",
			sessionToken: "invalid-token",
			mockFunc: func(ctx context.Context, token string) (*auth.User, error) {
				return nil, shared.ErrUnauthorized
			},
			expectedStatus: http.StatusUnauthorized,
			checkContext: func(t *testing.T, r *http.Request) {
				_, ok := GetUserFromContext(r.Context())
				if ok {
					t.Error("User should not be in context when token is invalid")
				}
			},
		},
		{
			name:         "expired session token",
			sessionToken: "expired-token",
			mockFunc: func(ctx context.Context, token string) (*auth.User, error) {
				return nil, shared.ErrUnauthorized
			},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockAuthService{validateSessionFunc: tt.mockFunc}
			middleware := NewAuthMiddleware(mockService)

			handlerCalled := false
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				handlerCalled = true
				if tt.checkContext != nil {
					tt.checkContext(t, r)
				}
				w.WriteHeader(http.StatusOK)
			})

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			if tt.sessionToken != "" {
				req.Header.Set("X-Session-Token", tt.sessionToken)
			}
			rr := httptest.NewRecorder()

			middleware.Authenticate(nextHandler).ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			if tt.expectedStatus == http.StatusOK && !handlerCalled {
				t.Error("Next handler should be called on successful authentication")
			}
			if tt.expectedStatus != http.StatusOK && handlerCalled {
				t.Error("Next handler should not be called on failed authentication")
			}
		})
	}
}

func TestGetUserFromContext(t *testing.T) {
	t.Run("user in context", func(t *testing.T) {
		user := &auth.User{
			ID:    uuid.New(),
			Email: "test@example.com",
		}
		ctx := context.WithValue(context.Background(), shared.UserContextKey, user)

		got, ok := GetUserFromContext(ctx)
		if !ok {
			t.Error("GetUserFromContext() should return true")
		}
		if got.ID != user.ID {
			t.Errorf("GetUserFromContext() user.ID = %v, want %v", got.ID, user.ID)
		}
	})

	t.Run("no user in context", func(t *testing.T) {
		ctx := context.Background()

		_, ok := GetUserFromContext(ctx)
		if ok {
			t.Error("GetUserFromContext() should return false when no user in context")
		}
	})

	t.Run("wrong type in context", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), shared.UserContextKey, "not a user")

		_, ok := GetUserFromContext(ctx)
		if ok {
			t.Error("GetUserFromContext() should return false when wrong type in context")
		}
	})
}

// Property 4: Protected Endpoints Require Authentication
func TestProperty_ProtectedEndpointsRequireAuthentication(t *testing.T) {
	mockService := &MockAuthService{
		validateSessionFunc: func(ctx context.Context, token string) (*auth.User, error) {
			if token == "valid-token" {
				return &auth.User{
					ID:    uuid.New(),
					Email: "test@example.com",
					Role:  auth.RoleCustomer,
				}, nil
			}
			return nil, shared.ErrUnauthorized
		},
	}
	middleware := NewAuthMiddleware(mockService)

	handlerCalled := false
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	})

	// Property: Request without token is rejected
	t.Run("no_token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		rr := httptest.NewRecorder()

		middleware.Authenticate(nextHandler).ServeHTTP(rr, req)

		if rr.Code != http.StatusUnauthorized {
			t.Errorf("Expected 401, got %d", rr.Code)
		}
		if handlerCalled {
			t.Error("Handler should not be called without token")
		}
	})

	// Property: Request with invalid token is rejected
	t.Run("invalid_token", func(t *testing.T) {
		handlerCalled = false
		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("X-Session-Token", "invalid-token")
		rr := httptest.NewRecorder()

		middleware.Authenticate(nextHandler).ServeHTTP(rr, req)

		if rr.Code != http.StatusUnauthorized {
			t.Errorf("Expected 401, got %d", rr.Code)
		}
		if handlerCalled {
			t.Error("Handler should not be called with invalid token")
		}
	})

	// Property: Request with valid token is allowed
	t.Run("valid_token", func(t *testing.T) {
		handlerCalled = false
		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("X-Session-Token", "valid-token")
		rr := httptest.NewRecorder()

		middleware.Authenticate(nextHandler).ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("Expected 200, got %d", rr.Code)
		}
		if !handlerCalled {
			t.Error("Handler should be called with valid token")
		}
	})
}
