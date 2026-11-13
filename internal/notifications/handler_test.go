package notifications

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/MarkyGitHub/e-commerce/internal/shared"
)

// MockUser implements shared.UserIDGetter for testing
type MockUser struct {
	ID    uuid.UUID
	Email string
	Role  string
}

func (u *MockUser) GetID() uuid.UUID {
	return u.ID
}

// MockNotificationService is a mock implementation of NotificationService
type MockNotificationService struct {
	getFunc        func(ctx context.Context, userID uuid.UUID) ([]Notification, error)
	markAsReadFunc func(ctx context.Context, userID uuid.UUID, notificationID uuid.UUID) error
}

func (m *MockNotificationService) CreateNotification(ctx context.Context, userID uuid.UUID, title, message string) (*Notification, error) {
	return nil, nil
}

func (m *MockNotificationService) CreateNotificationTx(ctx context.Context, tx *gorm.DB, userID uuid.UUID, title, message string) (*Notification, error) {
	return nil, nil
}

func (m *MockNotificationService) GetNotifications(ctx context.Context, userID uuid.UUID) ([]Notification, error) {
	if m.getFunc != nil {
		return m.getFunc(ctx, userID)
	}
	return nil, nil
}

func (m *MockNotificationService) MarkAsRead(ctx context.Context, userID uuid.UUID, notificationID uuid.UUID) error {
	if m.markAsReadFunc != nil {
		return m.markAsReadFunc(ctx, userID, notificationID)
	}
	return nil
}

// Helper to create request with user context
func requestWithUser(method, url string, user *MockUser) *http.Request {
	req := httptest.NewRequest(method, url, nil)
	ctx := context.WithValue(req.Context(), shared.UserContextKey, user)
	return req.WithContext(ctx)
}

func TestNotificationHandler_List(t *testing.T) {
	userID := uuid.New()
	user := &MockUser{ID: userID, Email: "test@example.com", Role: "customer"}

	tests := []struct {
		name           string
		user           *MockUser
		mockFunc       func(ctx context.Context, userID uuid.UUID) ([]Notification, error)
		expectedStatus int
		checkResponse  func(t *testing.T, body []byte)
	}{
		{
			name: "list notifications successfully",
			user: user,
			mockFunc: func(ctx context.Context, userID uuid.UUID) ([]Notification, error) {
				return []Notification{
					{ID: uuid.New(), UserID: userID, Message: "Order placed", IsRead: false, CreatedAt: time.Now()},
					{ID: uuid.New(), UserID: userID, Message: "Order shipped", IsRead: true, CreatedAt: time.Now()},
				}, nil
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body []byte) {
				var response []NotificationResponse
				if err := json.Unmarshal(body, &response); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				if len(response) != 2 {
					t.Errorf("Expected 2 notifications, got %d", len(response))
				}
			},
		},
		{
			name: "empty notifications",
			user: user,
			mockFunc: func(ctx context.Context, userID uuid.UUID) ([]Notification, error) {
				return []Notification{}, nil
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body []byte) {
				var response []NotificationResponse
				if err := json.Unmarshal(body, &response); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				if len(response) != 0 {
					t.Errorf("Expected 0 notifications, got %d", len(response))
				}
			},
		},
		{
			name:           "unauthorized - no user",
			user:           nil,
			mockFunc:       nil,
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockNotificationService{getFunc: tt.mockFunc}
			handler := NewHandler(mockService)

			var req *http.Request
			if tt.user != nil {
				req = requestWithUser(http.MethodGet, "/notifications", tt.user)
			} else {
				req = httptest.NewRequest(http.MethodGet, "/notifications", nil)
			}
			rr := httptest.NewRecorder()

			handler.List(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			if tt.checkResponse != nil {
				tt.checkResponse(t, rr.Body.Bytes())
			}
		})
	}
}

func TestNotificationHandler_MarkAsRead(t *testing.T) {
	userID := uuid.New()
	user := &MockUser{ID: userID, Email: "test@example.com", Role: "customer"}
	notificationID := uuid.New()

	tests := []struct {
		name           string
		user           *MockUser
		notificationID string
		mockFunc       func(ctx context.Context, userID uuid.UUID, notificationID uuid.UUID) error
		expectedStatus int
		checkResponse  func(t *testing.T, body []byte)
	}{
		{
			name:           "mark as read successfully",
			user:           user,
			notificationID: notificationID.String(),
			mockFunc: func(ctx context.Context, userID uuid.UUID, nid uuid.UUID) error {
				return nil
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body []byte) {
				var response map[string]string
				if err := json.Unmarshal(body, &response); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				if response["message"] != "notification marked as read" {
					t.Errorf("Expected 'notification marked as read', got %s", response["message"])
				}
			},
		},
		{
			name:           "notification not found",
			user:           user,
			notificationID: uuid.New().String(),
			mockFunc: func(ctx context.Context, userID uuid.UUID, nid uuid.UUID) error {
				return shared.ErrNotFound
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "forbidden - not owner",
			user:           user,
			notificationID: uuid.New().String(),
			mockFunc: func(ctx context.Context, userID uuid.UUID, nid uuid.UUID) error {
				return shared.ErrForbidden
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "invalid UUID",
			user:           user,
			notificationID: "invalid-uuid",
			mockFunc:       nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "unauthorized - no user",
			user:           nil,
			notificationID: notificationID.String(),
			mockFunc:       nil,
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockNotificationService{markAsReadFunc: tt.mockFunc}
			handler := NewHandler(mockService)

			var req *http.Request
			if tt.user != nil {
				req = requestWithUser(http.MethodPut, "/notifications/"+tt.notificationID+"/read", tt.user)
			} else {
				req = httptest.NewRequest(http.MethodPut, "/notifications/"+tt.notificationID+"/read", nil)
			}
			req.SetPathValue("id", tt.notificationID)
			rr := httptest.NewRecorder()

			handler.MarkAsRead(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			if tt.checkResponse != nil {
				tt.checkResponse(t, rr.Body.Bytes())
			}
		})
	}
}

// Test notification flow
func TestNotificationHandler_Flow(t *testing.T) {
	userID := uuid.New()
	user := &MockUser{ID: userID, Email: "test@example.com", Role: "customer"}

	notifications := []Notification{
		{ID: uuid.New(), UserID: userID, Message: "Order #1 placed", IsRead: false, CreatedAt: time.Now()},
		{ID: uuid.New(), UserID: userID, Message: "Order #2 shipped", IsRead: false, CreatedAt: time.Now()},
	}

	mockService := &MockNotificationService{
		getFunc: func(ctx context.Context, uid uuid.UUID) ([]Notification, error) {
			if uid != userID {
				return []Notification{}, nil
			}
			return notifications, nil
		},
		markAsReadFunc: func(ctx context.Context, uid uuid.UUID, nid uuid.UUID) error {
			if uid != userID {
				return shared.ErrForbidden
			}
			for i, n := range notifications {
				if n.ID == nid {
					notifications[i].IsRead = true
					return nil
				}
			}
			return shared.ErrNotFound
		},
	}

	handler := NewHandler(mockService)

	// List notifications
	t.Run("list", func(t *testing.T) {
		req := requestWithUser(http.MethodGet, "/notifications", user)
		rr := httptest.NewRecorder()

		handler.List(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("List failed: %d", rr.Code)
		}

		var response []NotificationResponse
		if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal list response: %v", err)
		}
		if len(response) != 2 {
			t.Errorf("Expected 2 notifications, got %d", len(response))
		}
	})

	// Mark first notification as read
	t.Run("mark_as_read", func(t *testing.T) {
		req := requestWithUser(http.MethodPut, "/notifications/"+notifications[0].ID.String()+"/read", user)
		req.SetPathValue("id", notifications[0].ID.String())
		rr := httptest.NewRecorder()

		handler.MarkAsRead(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("MarkAsRead failed: %d", rr.Code)
		}
	})

	// Verify first notification is marked as read
	if !notifications[0].IsRead {
		t.Error("First notification should be marked as read")
	}
}

// TestNotificationHandler_List_InvalidUserContext tests List with invalid user context
func TestNotificationHandler_List_InvalidUserContext(t *testing.T) {
	mockService := &MockNotificationService{}
	handler := NewHandler(mockService)

	req := httptest.NewRequest(http.MethodGet, "/notifications", nil)
	// Invalid user context (not implementing GetID)
	ctx := context.WithValue(req.Context(), shared.UserContextKey, "invalid")
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.List(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, rr.Code)
	}
}

// TestNotificationHandler_MarkAsRead_InvalidUserContext tests MarkAsRead with invalid user context
func TestNotificationHandler_MarkAsRead_InvalidUserContext(t *testing.T) {
	mockService := &MockNotificationService{}
	handler := NewHandler(mockService)

	req := httptest.NewRequest(http.MethodPut, "/notifications/"+uuid.New().String()+"/read", nil)
	req.SetPathValue("id", uuid.New().String())
	// Invalid user context
	ctx := context.WithValue(req.Context(), shared.UserContextKey, "invalid")
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.MarkAsRead(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, rr.Code)
	}
}

// TestNotificationHandler_MarkAsRead_MissingPathValue tests MarkAsRead with missing path value
func TestNotificationHandler_MarkAsRead_MissingPathValue(t *testing.T) {
	mockService := &MockNotificationService{}
	handler := NewHandler(mockService)

	userID := uuid.New()
	user := &MockUser{ID: userID, Email: "test@example.com", Role: "customer"}
	req := requestWithUser(http.MethodPut, "/notifications//read", user)
	// No path value set
	rr := httptest.NewRecorder()

	handler.MarkAsRead(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

// TestNotificationHandler_HandleError tests all error handling paths
func TestNotificationHandler_HandleError(t *testing.T) {
	tests := []struct {
		name           string
		err            error
		expectedStatus int
		expectedMsg    string
	}{
		{
			name:           "ErrNotFound",
			err:            shared.ErrNotFound,
			expectedStatus: http.StatusNotFound,
			expectedMsg:    "notification not found",
		},
		{
			name:           "ErrForbidden",
			err:            shared.ErrForbidden,
			expectedStatus: http.StatusForbidden,
			expectedMsg:    "access denied",
		},
		{
			name:           "unknown error",
			err:            shared.ErrBadRequest,
			expectedStatus: http.StatusInternalServerError,
			expectedMsg:    "internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockNotificationService{
				getFunc: func(ctx context.Context, userID uuid.UUID) ([]Notification, error) {
					return nil, tt.err
				},
			}
			handler := NewHandler(mockService)

			userID := uuid.New()
			user := &MockUser{ID: userID, Email: "test@example.com", Role: "customer"}
			req := requestWithUser(http.MethodGet, "/notifications", user)
			rr := httptest.NewRecorder()

			handler.List(rr, req)

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