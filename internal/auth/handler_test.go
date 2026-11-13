package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/MarkyGitHub/e-commerce/internal/shared"
)

// MockAuthService is a mock implementation of AuthService for handler testing
type MockAuthService struct {
	registerFunc        func(ctx context.Context, req RegisterRequest) (*User, error)
	loginFunc           func(ctx context.Context, req LoginRequest) (*LoginResponse, error)
	logoutFunc          func(ctx context.Context, token string) error
	validateSessionFunc func(ctx context.Context, token string) (*User, error)
}

func (m *MockAuthService) Register(ctx context.Context, req RegisterRequest) (*User, error) {
	if m.registerFunc != nil {
		return m.registerFunc(ctx, req)
	}
	return nil, nil
}

func (m *MockAuthService) Login(ctx context.Context, req LoginRequest) (*LoginResponse, error) {
	if m.loginFunc != nil {
		return m.loginFunc(ctx, req)
	}
	return nil, nil
}

func (m *MockAuthService) Logout(ctx context.Context, token string) error {
	if m.logoutFunc != nil {
		return m.logoutFunc(ctx, token)
	}
	return nil
}

func (m *MockAuthService) ValidateSession(ctx context.Context, token string) (*User, error) {
	if m.validateSessionFunc != nil {
		return m.validateSessionFunc(ctx, token)
	}
	return nil, nil
}

func TestAuthHandler_Register(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		mockFunc       func(ctx context.Context, req RegisterRequest) (*User, error)
		expectedStatus int
		checkResponse  func(t *testing.T, body []byte)
	}{
		{
			name: "successful registration",
			requestBody: RegisterRequest{
				Email:    "test@example.com",
				Password: "password123",
				Username:     "Test User",
			},
			mockFunc: func(ctx context.Context, req RegisterRequest) (*User, error) {
				return &User{
					ID:        uuid.New(),
					Email:     req.Email,
					Username:      req.Username,
					Role:      RoleCustomer,
					CreatedAt: time.Now(),
				}, nil
			},
			expectedStatus: http.StatusCreated,
			checkResponse: func(t *testing.T, body []byte) {
				var response UserResponse
				if err := json.Unmarshal(body, &response); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				if response.Email != "test@example.com" {
					t.Errorf("Expected email test@example.com, got %s", response.Email)
				}
				if response.Role != RoleCustomer {
					t.Errorf("Expected role customer, got %s", response.Role)
				}
			},
		},
		{
			name: "invalid email",
			requestBody: RegisterRequest{
				Email:    "invalid-email",
				Password: "password123",
				Username:     "Test User",
			},
			mockFunc: func(ctx context.Context, req RegisterRequest) (*User, error) {
				return nil, shared.ErrInvalidEmail
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, body []byte) {
				var response shared.ErrorResponse
				if err := json.Unmarshal(body, &response); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				if response.Message != "invalid email format" {
					t.Errorf("Expected 'invalid email format', got %s", response.Message)
				}
			},
		},
		{
			name: "password too short",
			requestBody: RegisterRequest{
				Email:    "test@example.com",
				Password: "short",
				Username:     "Test User",
			},
			mockFunc: func(ctx context.Context, req RegisterRequest) (*User, error) {
				return nil, shared.ErrPasswordTooShort
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, body []byte) {
				var response shared.ErrorResponse
				if err := json.Unmarshal(body, &response); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				if response.Message != "password must be at least 8 characters" {
					t.Errorf("Expected 'password must be at least 8 characters', got %s", response.Message)
				}
			},
		},
		{
			name: "email already registered",
			requestBody: RegisterRequest{
				Email:    "existing@example.com",
				Password: "password123",
				Username:     "Test User",
			},
			mockFunc: func(ctx context.Context, req RegisterRequest) (*User, error) {
				return nil, shared.ErrConflict
			},
			expectedStatus: http.StatusConflict,
			checkResponse: func(t *testing.T, body []byte) {
				var response shared.ErrorResponse
				if err := json.Unmarshal(body, &response); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				if response.Message != "email already registered" {
					t.Errorf("Expected 'email already registered', got %s", response.Message)
				}
			},
		},
		{
			name:           "invalid JSON body",
			requestBody:    "invalid json",
			mockFunc:       nil,
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, body []byte) {
				var response shared.ErrorResponse
				if err := json.Unmarshal(body, &response); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				if response.Message != "invalid request body" {
					t.Errorf("Expected 'invalid request body', got %s", response.Message)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockAuthService{registerFunc: tt.mockFunc}
			handler := NewHandler(mockService)

			var body []byte
			var err error
			if str, ok := tt.requestBody.(string); ok {
				body = []byte(str)
			} else {
				body, err = json.Marshal(tt.requestBody)
				if err != nil {
					t.Fatalf("Failed to marshal request: %v", err)
				}
			}

			req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			handler.Register(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			if tt.checkResponse != nil {
				tt.checkResponse(t, rr.Body.Bytes())
			}
		})
	}
}

func TestAuthHandler_Login(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		mockFunc       func(ctx context.Context, req LoginRequest) (*LoginResponse, error)
		expectedStatus int
		checkResponse  func(t *testing.T, body []byte)
	}{
		{
			name: "successful login",
			requestBody: LoginRequest{
				Email:    "test@example.com",
				Password: "password123",
			},
			mockFunc: func(ctx context.Context, req LoginRequest) (*LoginResponse, error) {
				return &LoginResponse{
					SessionToken:     "test-session-token",
					ExpiresAt: time.Now().Add(24 * time.Hour),
					User: User{
						ID:    uuid.New(),
						Email: req.Email,
						Username:  "Test User",
						Role:  RoleCustomer,
					},
				}, nil
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body []byte) {
				var response LoginResponse
				if err := json.Unmarshal(body, &response); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				if response.SessionToken != "test-session-token" {
					t.Errorf("Expected token 'test-session-token', got %s", response.SessionToken)
				}
				if response.User.Email != "test@example.com" {
					t.Errorf("Expected email test@example.com, got %s", response.User.Email)
				}
			},
		},
		{
			name: "invalid credentials",
			requestBody: LoginRequest{
				Email:    "test@example.com",
				Password: "wrongpassword",
			},
			mockFunc: func(ctx context.Context, req LoginRequest) (*LoginResponse, error) {
				return nil, shared.ErrInvalidCredentials
			},
			expectedStatus: http.StatusUnauthorized,
			checkResponse: func(t *testing.T, body []byte) {
				var response shared.ErrorResponse
				if err := json.Unmarshal(body, &response); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				if response.Message != "invalid credentials" {
					t.Errorf("Expected 'invalid credentials', got %s", response.Message)
				}
			},
		},
		{
			name:           "invalid JSON body",
			requestBody:    "invalid json",
			mockFunc:       nil,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockAuthService{loginFunc: tt.mockFunc}
			handler := NewHandler(mockService)

			var body []byte
			var err error
			if str, ok := tt.requestBody.(string); ok {
				body = []byte(str)
			} else {
				body, err = json.Marshal(tt.requestBody)
				if err != nil {
					t.Fatalf("Failed to marshal request: %v", err)
				}
			}

			req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			handler.Login(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			if tt.checkResponse != nil {
				tt.checkResponse(t, rr.Body.Bytes())
			}
		})
	}
}

func TestAuthHandler_Logout(t *testing.T) {
	tests := []struct {
		name           string
		sessionToken   string
		mockFunc       func(ctx context.Context, token string) error
		expectedStatus int
	}{
		{
			name:         "successful logout",
			sessionToken: "valid-session-token",
			mockFunc: func(ctx context.Context, token string) error {
				return nil
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name:         "invalid session token",
			sessionToken: "invalid-token",
			mockFunc: func(ctx context.Context, token string) error {
				return shared.ErrUnauthorized
			},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockAuthService{logoutFunc: tt.mockFunc}
			handler := NewHandler(mockService)

			req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
			req.Header.Set("X-Session-Token", tt.sessionToken)
			rr := httptest.NewRecorder()

			handler.Logout(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
			}
		})
	}
}

// Integration-style test for the full auth flow
func TestAuthHandler_FullFlow(t *testing.T) {
	// Setup mock service that maintains state
	users := make(map[string]*User)
	sessions := make(map[string]string) // token -> email

	mockService := &MockAuthService{
		registerFunc: func(ctx context.Context, req RegisterRequest) (*User, error) {
			if _, exists := users[req.Email]; exists {
				return nil, shared.ErrConflict
			}
			user := &User{
				ID:        uuid.New(),
				Email:     req.Email,
				Username:      req.Username,
				Role:      RoleCustomer,
				CreatedAt: time.Now(),
			}
			users[req.Email] = user
			return user, nil
		},
		loginFunc: func(ctx context.Context, req LoginRequest) (*LoginResponse, error) {
			user, exists := users[req.Email]
			if !exists {
				return nil, shared.ErrInvalidCredentials
			}
			token := "session-" + req.Email
			sessions[token] = req.Email
			return &LoginResponse{
				SessionToken:     token,
				ExpiresAt: time.Now().Add(24 * time.Hour),
				User:      *user,
			}, nil
		},
		logoutFunc: func(ctx context.Context, token string) error {
			delete(sessions, token)
			return nil
		},
	}

	handler := NewHandler(mockService)

	// Step 1: Register
	t.Run("register", func(t *testing.T) {
		body, _ := json.Marshal(RegisterRequest{
			Email:    "flow@test.com",
			Password: "password123",
			Username:     "Flow Test",
		})
		req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler.Register(rr, req)

		if rr.Code != http.StatusCreated {
			t.Fatalf("Register failed: %d", rr.Code)
		}
	})

	// Step 2: Login
	var sessionToken string
	t.Run("login", func(t *testing.T) {
		body, _ := json.Marshal(LoginRequest{
			Email:    "flow@test.com",
			Password: "password123",
		})
		req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler.Login(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("Login failed: %d", rr.Code)
		}

		var response LoginResponse
		if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal login response: %v", err)
		}
		sessionToken = response.SessionToken
	})

	// Step 3: Logout
	t.Run("logout", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
		req.Header.Set("X-Session-Token", sessionToken)
		rr := httptest.NewRecorder()

		handler.Logout(rr, req)

		if rr.Code != http.StatusNoContent {
			t.Fatalf("Logout failed: %d", rr.Code)
		}
	})
}

// TestAuthHandler_HandleError tests all error handling paths
func TestAuthHandler_HandleError(t *testing.T) {
	tests := []struct {
		name           string
		err            error
		expectedStatus int
		expectedMsg    string
	}{
		{
			name:           "ErrInvalidEmail",
			err:            shared.ErrInvalidEmail,
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "invalid email format",
		},
		{
			name:           "ErrPasswordTooShort",
			err:            shared.ErrPasswordTooShort,
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "password must be at least 8 characters",
		},
		{
			name:           "ErrConflict",
			err:            shared.ErrConflict,
			expectedStatus: http.StatusConflict,
			expectedMsg:    "email already registered",
		},
		{
			name:           "ErrInvalidCredentials",
			err:            shared.ErrInvalidCredentials,
			expectedStatus: http.StatusUnauthorized,
			expectedMsg:    "invalid credentials",
		},
		{
			name:           "ErrUnauthorized",
			err:            shared.ErrUnauthorized,
			expectedStatus: http.StatusUnauthorized,
			expectedMsg:    "unauthorized",
		},
		{
			name:           "ErrBadRequest",
			err:            shared.ErrBadRequest,
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "invalid request",
		},
		{
			name:           "unknown error",
			err:            shared.ErrNotFound,
			expectedStatus: http.StatusInternalServerError,
			expectedMsg:    "internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockAuthService{
				registerFunc: func(ctx context.Context, req RegisterRequest) (*User, error) {
					return nil, tt.err
				},
			}
			handler := NewHandler(mockService)

			body, _ := json.Marshal(RegisterRequest{
				Email:    "test@example.com",
				Password: "password123",
				Username: "Test User",
			})
			req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			handler.Register(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			var response shared.ErrorResponse
			if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
				t.Fatalf("Failed to unmarshal response: %v", err)
			}
			if response.Message != tt.expectedMsg {
				t.Errorf("Expected message '%s', got '%s'", tt.expectedMsg, response.Message)
			}
		})
	}
}