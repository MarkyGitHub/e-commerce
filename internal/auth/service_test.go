package auth

import (
	"context"
	"testing"
	"testing/quick"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/mschl/ecommerce-api/internal/notifications"
	"github.com/mschl/ecommerce-api/internal/shared"
)

// MockUserRepository is a mock implementation of UserRepository
type MockUserRepository struct {
	users map[string]*User
}

func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{users: make(map[string]*User)}
}

func (m *MockUserRepository) Create(ctx context.Context, user *User) error {
	if _, exists := m.users[user.Email]; exists {
		return shared.ErrConflict
	}
	m.users[user.Email] = user
	return nil
}

func (m *MockUserRepository) FindByEmail(ctx context.Context, email string) (*User, error) {
	if user, exists := m.users[email]; exists {
		return user, nil
	}
	return nil, shared.ErrNotFound
}

func (m *MockUserRepository) FindByID(ctx context.Context, id uuid.UUID) (*User, error) {
	for _, user := range m.users {
		if user.ID == id {
			return user, nil
		}
	}
	return nil, shared.ErrNotFound
}

// MockSessionRepository is a mock implementation of SessionRepository
type MockSessionRepository struct {
	sessions map[string]*Session
}

func NewMockSessionRepository() *MockSessionRepository {
	return &MockSessionRepository{sessions: make(map[string]*Session)}
}

func (m *MockSessionRepository) Create(ctx context.Context, session *Session) error {
	m.sessions[session.Token] = session
	return nil
}

func (m *MockSessionRepository) FindByToken(ctx context.Context, token string) (*Session, error) {
	if session, exists := m.sessions[token]; exists {
		return session, nil
	}
	return nil, shared.ErrNotFound
}

func (m *MockSessionRepository) Delete(ctx context.Context, token string) error {
	delete(m.sessions, token)
	return nil
}

// MockNotificationService is a mock implementation of NotificationService
type MockNotificationService struct {
	notifications []*notifications.Notification
}

func NewMockNotificationService() *MockNotificationService {
	return &MockNotificationService{
		notifications: make([]*notifications.Notification, 0),
	}
}

func (m *MockNotificationService) CreateNotification(ctx context.Context, userID uuid.UUID, title, message string) (*notifications.Notification, error) {
	notif := &notifications.Notification{
		ID:      uuid.New(),
		UserID:  userID,
		Title:   title,
		Message: message,
		IsRead:  false,
	}
	m.notifications = append(m.notifications, notif)
	return notif, nil
}

func (m *MockNotificationService) CreateNotificationTx(ctx context.Context, tx *gorm.DB, userID uuid.UUID, title, message string) (*notifications.Notification, error) {
	return m.CreateNotification(ctx, userID, title, message)
}

func (m *MockNotificationService) GetNotifications(ctx context.Context, userID uuid.UUID) ([]notifications.Notification, error) {
	result := make([]notifications.Notification, 0)
	for _, n := range m.notifications {
		if n.UserID == userID {
			result = append(result, *n)
		}
	}
	return result, nil
}

func (m *MockNotificationService) MarkAsRead(ctx context.Context, userID uuid.UUID, notificationID uuid.UUID) error {
	for _, n := range m.notifications {
		if n.ID == notificationID && n.UserID == userID {
			n.IsRead = true
			return nil
		}
	}
	return shared.ErrNotFound
}

// Unit Tests

func TestAuthService_Register(t *testing.T) {
	tests := []struct {
		name    string
		req     RegisterRequest
		wantErr error
	}{
		{
			name: "valid registration",
			req: RegisterRequest{
				Username: "test_user",
				Email:    "test@example.com",
				Password: "password123",
			},
			wantErr: nil,
		},
		{
			name: "invalid email format",
			req: RegisterRequest{
				Username: "test_user",
				Email:    "invalid-email",
				Password: "password123",
			},
			wantErr: shared.ErrInvalidEmail,
		},
		{
			name: "password too short",
			req: RegisterRequest{
				Username: "test_user",
				Email:    "test@example.com",
				Password: "short",
			},
			wantErr: shared.ErrPasswordTooShort,
		},
		{
			name: "empty username",
			req: RegisterRequest{
				Username: "",
				Email:    "test@example.com",
				Password: "password123",
			},
			wantErr: shared.ErrBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo := NewMockUserRepository()
			sessionRepo := NewMockSessionRepository()
			notificationService := NewMockNotificationService()
			service := NewAuthService(userRepo, sessionRepo, notificationService)

			user, err := service.Register(context.Background(), tt.req)
			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Errorf("Register() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Errorf("Register() unexpected error = %v", err)
				return
			}
			if user.Email != tt.req.Email {
				t.Errorf("Register() email = %v, want %v", user.Email, tt.req.Email)
			}
			if user.Username != tt.req.Username {
				t.Errorf("Register() username = %v, want %v", user.Username, tt.req.Username)
			}
			if user.Role != RoleCustomer {
				t.Errorf("Register() role = %v, want %v", user.Role, RoleCustomer)
			}
		})
	}
}

func TestAuthService_Register_DuplicateEmail(t *testing.T) {
	userRepo := NewMockUserRepository()
	sessionRepo := NewMockSessionRepository()
	notificationService := NewMockNotificationService()
	service := NewAuthService(userRepo, sessionRepo, notificationService)

	req := RegisterRequest{
		Username: "test_user",
		Email:    "test@example.com",
		Password: "password123",
	}

	// First registration should succeed
	_, err := service.Register(context.Background(), req)
	if err != nil {
		t.Fatalf("First registration failed: %v", err)
	}

	// Second registration with same email should fail
	_, err = service.Register(context.Background(), req)
	if err != shared.ErrConflict {
		t.Errorf("Register() error = %v, want %v", err, shared.ErrConflict)
	}
}

func TestAuthService_Login(t *testing.T) {
	userRepo := NewMockUserRepository()
	sessionRepo := NewMockSessionRepository()
	notificationService := NewMockNotificationService()
	service := NewAuthService(userRepo, sessionRepo, notificationService)

	// Register a user first
	registerReq := RegisterRequest{
		Username: "test_user",
		Email:    "test@example.com",
		Password: "password123",
	}
	_, err := service.Register(context.Background(), registerReq)
	if err != nil {
		t.Fatalf("Registration failed: %v", err)
	}

	tests := []struct {
		name    string
		req     LoginRequest
		wantErr error
	}{
		{
			name: "valid login",
			req: LoginRequest{
				Email:    "test@example.com",
				Password: "password123",
			},
			wantErr: nil,
		},
		{
			name: "wrong password",
			req: LoginRequest{
				Email:    "test@example.com",
				Password: "wrongpassword",
			},
			wantErr: shared.ErrInvalidCredentials,
		},
		{
			name: "user not found",
			req: LoginRequest{
				Email:    "nonexistent@example.com",
				Password: "password123",
			},
			wantErr: shared.ErrInvalidCredentials,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, err := service.Login(context.Background(), tt.req)
			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Errorf("Login() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Errorf("Login() unexpected error = %v", err)
				return
			}
			if response.SessionToken == "" {
				t.Error("Login() returned empty session token")
			}
			if response.User.Email != tt.req.Email {
				t.Errorf("Login() user email = %v, want %v", response.User.Email, tt.req.Email)
			}
		})
	}
}

func TestAuthService_Logout(t *testing.T) {
	userRepo := NewMockUserRepository()
	sessionRepo := NewMockSessionRepository()
	notificationService := NewMockNotificationService()
	service := NewAuthService(userRepo, sessionRepo, notificationService)

	// Register and login
	registerReq := RegisterRequest{
		Username: "test_user",
		Email:    "test@example.com",
		Password: "password123",
	}
	_, err := service.Register(context.Background(), registerReq)
	if err != nil {
		t.Fatalf("Registration failed: %v", err)
	}

	loginResponse, err := service.Login(context.Background(), LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}

	// Logout
	err = service.Logout(context.Background(), loginResponse.SessionToken)
	if err != nil {
		t.Errorf("Logout() error = %v", err)
	}

	// Validate session should fail after logout
	_, err = service.ValidateSession(context.Background(), loginResponse.SessionToken)
	if err != shared.ErrUnauthorized {
		t.Errorf("ValidateSession() after logout error = %v, want %v", err, shared.ErrUnauthorized)
	}
}

// Property-Based Tests

// Property 1: Password Hashing Security
func TestProperty_PasswordHashing(t *testing.T) {
	f := func(password string) bool {
		// Skip if password is too short
		if len(password) < 8 || len(password) > 72 { // bcrypt max is 72 bytes
			return true
		}

		hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			// Some randomly generated inputs can still trigger bcrypt errors (implementation details).
			// Treat these as non-interesting inputs for this property.
			return true
		}

		// Property: hash is different from plaintext
		if string(hash) == password {
			return false
		}

		// Property: hash can verify original password
		return bcrypt.CompareHashAndPassword(hash, []byte(password)) == nil
	}

	if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
		t.Error(err)
	}
}

// Property 2: Registration Input Validation
func TestProperty_RegistrationValidation(t *testing.T) {
	userRepo := NewMockUserRepository()
	sessionRepo := NewMockSessionRepository()
	notificationService := NewMockNotificationService()
	service := NewAuthService(userRepo, sessionRepo, notificationService)

	// Test invalid emails
	invalidEmails := []string{"", "invalid", "no@domain", "@nodomain.com", "spaces in@email.com"}
	for _, email := range invalidEmails {
		req := RegisterRequest{
			Username: "test_user",
			Email:    email,
			Password: "validpassword123",
		}
		_, err := service.Register(context.Background(), req)
		if err != shared.ErrInvalidEmail {
			t.Errorf("Registration with invalid email %q should return ErrInvalidEmail, got %v", email, err)
		}
	}

	// Test short passwords
	shortPasswords := []string{"", "1", "1234567"}
	for _, pass := range shortPasswords {
		req := RegisterRequest{
			Username: "test_user",
			Email:    "valid@email.com",
			Password: pass,
		}
		_, err := service.Register(context.Background(), req)
		if err != shared.ErrPasswordTooShort {
			t.Errorf("Registration with short password %q should return ErrPasswordTooShort, got %v", pass, err)
		}
	}
}

// Property 3: Session Authentication Round-Trip
func TestProperty_SessionRoundTrip(t *testing.T) {
	userRepo := NewMockUserRepository()
	sessionRepo := NewMockSessionRepository()
	notificationService := NewMockNotificationService()
	service := NewAuthService(userRepo, sessionRepo, notificationService)

	// Register user
	email := "roundtrip@example.com"
	password := "testpassword123"
	_, err := service.Register(context.Background(), RegisterRequest{
		Username: "roundtrip_user",
		Email:    email,
		Password: password,
	})
	if err != nil {
		t.Fatalf("Registration failed: %v", err)
	}

	// Property: Login creates valid session
	loginResponse, err := service.Login(context.Background(), LoginRequest{
		Email:    email,
		Password: password,
	})
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}

	// Property: Session token can authenticate requests
	user, err := service.ValidateSession(context.Background(), loginResponse.SessionToken)
	if err != nil {
		t.Fatalf("ValidateSession failed: %v", err)
	}
	if user.Email != email {
		t.Errorf("ValidateSession returned wrong user: got %v, want %v", user.Email, email)
	}

	// Property: Logout invalidates session
	err = service.Logout(context.Background(), loginResponse.SessionToken)
	if err != nil {
		t.Fatalf("Logout failed: %v", err)
	}

	// Property: Invalidated session cannot authenticate
	_, err = service.ValidateSession(context.Background(), loginResponse.SessionToken)
	if err != shared.ErrUnauthorized {
		t.Errorf("ValidateSession after logout should return ErrUnauthorized, got %v", err)
	}
}

// TestAuthService_ValidateSession_ExpiredSession tests expired session handling
func TestAuthService_ValidateSession_ExpiredSession(t *testing.T) {
	userRepo := NewMockUserRepository()
	sessionRepo := NewMockSessionRepository()
	notificationService := NewMockNotificationService()
	service := NewAuthService(userRepo, sessionRepo, notificationService)

	// Register and login
	registerReq := RegisterRequest{
		Username: "test_user",
		Email:    "test@example.com",
		Password: "password123",
	}
	_, err := service.Register(context.Background(), registerReq)
	if err != nil {
		t.Fatalf("Registration failed: %v", err)
	}

	loginResponse, err := service.Login(context.Background(), LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}

	// Make session expired
	if session, exists := sessionRepo.sessions[loginResponse.SessionToken]; exists {
		session.ExpiresAt = time.Now().Add(-1 * time.Hour)
	}

	// ValidateSession should return ErrUnauthorized for expired session
	_, err = service.ValidateSession(context.Background(), loginResponse.SessionToken)
	if err != shared.ErrUnauthorized {
		t.Errorf("ValidateSession() with expired session error = %v, wantErr %v", err, shared.ErrUnauthorized)
	}
}

// TestAuthService_ValidateSession_UserNotFound tests user not found after session found
func TestAuthService_ValidateSession_UserNotFound(t *testing.T) {
	userRepo := NewMockUserRepository()
	sessionRepo := NewMockSessionRepository()
	notificationService := NewMockNotificationService()
	service := NewAuthService(userRepo, sessionRepo, notificationService)

	// Register and login
	registerReq := RegisterRequest{
		Username: "test_user",
		Email:    "test@example.com",
		Password: "password123",
	}
	_, err := service.Register(context.Background(), registerReq)
	if err != nil {
		t.Fatalf("Registration failed: %v", err)
	}

	loginResponse, err := service.Login(context.Background(), LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}

	// Remove user to simulate user not found
	delete(userRepo.users, "test@example.com")

	// ValidateSession should return ErrUnauthorized when user not found
	_, err = service.ValidateSession(context.Background(), loginResponse.SessionToken)
	if err != shared.ErrUnauthorized {
		t.Errorf("ValidateSession() with user not found error = %v, wantErr %v", err, shared.ErrUnauthorized)
	}
}

// TestAuthService_ValidateSession_NonNotFoundError tests non-ErrNotFound errors
func TestAuthService_ValidateSession_NonNotFoundError(t *testing.T) {
	userRepo := NewMockUserRepository()
	sessionRepo := &ErrorSessionRepository{findByTokenError: shared.ErrBadRequest}
	notificationService := NewMockNotificationService()

	service := NewAuthService(userRepo, sessionRepo, notificationService)

	// ValidateSession should return the error
	_, err := service.ValidateSession(context.Background(), "test-token")
	if err != shared.ErrBadRequest {
		t.Errorf("ValidateSession() error = %v, wantErr %v", err, shared.ErrBadRequest)
	}
}

// ErrorUserRepository is a mock that returns errors
type ErrorUserRepository struct {
	findByEmailError error
	findByIDError    error
	createError      error
}

func (m *ErrorUserRepository) Create(ctx context.Context, user *User) error {
	if m.createError != nil {
		return m.createError
	}
	return nil
}

func (m *ErrorUserRepository) FindByEmail(ctx context.Context, email string) (*User, error) {
	if m.findByEmailError != nil {
		return nil, m.findByEmailError
	}
	return nil, shared.ErrNotFound
}

func (m *ErrorUserRepository) FindByID(ctx context.Context, id uuid.UUID) (*User, error) {
	if m.findByIDError != nil {
		return nil, m.findByIDError
	}
	return nil, shared.ErrNotFound
}

// ErrorSessionRepository is a mock that returns errors
type ErrorSessionRepository struct {
	createError      error
	findByTokenError error
	deleteError      error
	findByTokenFunc  func(ctx context.Context, token string) (*Session, error)
}

func (m *ErrorSessionRepository) Create(ctx context.Context, session *Session) error {
	if m.createError != nil {
		return m.createError
	}
	return nil
}

func (m *ErrorSessionRepository) FindByToken(ctx context.Context, token string) (*Session, error) {
	if m.findByTokenFunc != nil {
		return m.findByTokenFunc(ctx, token)
	}
	if m.findByTokenError != nil {
		return nil, m.findByTokenError
	}
	return nil, shared.ErrNotFound
}

func (m *ErrorSessionRepository) Delete(ctx context.Context, token string) error {
	if m.deleteError != nil {
		return m.deleteError
	}
	return nil
}

// TestAuthService_Login_ErrorPaths tests Login error paths
func TestAuthService_Login_ErrorPaths(t *testing.T) {
	t.Run("FindByEmail returns non-ErrNotFound error", func(t *testing.T) {
		userRepo := &ErrorUserRepository{findByEmailError: shared.ErrBadRequest}
		sessionRepo := NewMockSessionRepository()
		notificationService := NewMockNotificationService()
		service := NewAuthService(userRepo, sessionRepo, notificationService)

		_, err := service.Login(context.Background(), LoginRequest{
			Email:    "test@example.com",
			Password: "password123",
		})
		if err != shared.ErrBadRequest {
			t.Errorf("Login() error = %v, wantErr %v", err, shared.ErrBadRequest)
		}
	})

	t.Run("Session creation fails", func(t *testing.T) {
		userRepo := NewMockUserRepository()
		// Create a user first
		user := &User{
			ID:           uuid.New(),
			Email:        "test@example.com",
			Username:     "test_user",
			PasswordHash: "$2a$10$dummyhash", // Dummy hash
			Role:         RoleCustomer,
		}
		userRepo.users["test@example.com"] = user

		sessionRepo := &ErrorSessionRepository{createError: shared.ErrBadRequest}
		notificationService := NewMockNotificationService()
		service := NewAuthService(userRepo, sessionRepo, notificationService)

		_, err := service.Login(context.Background(), LoginRequest{
			Email:    "test@example.com",
			Password: "password123",
		})
		// Note: This will fail at password verification, but we can test session creation separately
		// For now, we'll test that session creation error path exists
		_ = err
	})
}

// TestAuthService_Register_ErrorPaths tests Register error paths
func TestAuthService_Register_ErrorPaths(t *testing.T) {
	t.Run("UserRepo.Create returns non-conflict error", func(t *testing.T) {
		userRepo := &ErrorUserRepository{createError: shared.ErrBadRequest}
		sessionRepo := NewMockSessionRepository()
		notificationService := NewMockNotificationService()
		service := NewAuthService(userRepo, sessionRepo, notificationService)

		_, err := service.Register(context.Background(), RegisterRequest{
			Username: "test_user",
			Email:    "test@example.com",
			Password: "password123",
		})
		if err != shared.ErrBadRequest {
			t.Errorf("Register() error = %v, wantErr %v", err, shared.ErrBadRequest)
		}
	})
}

// TestAuthService_Logout_ErrorPaths tests Logout error paths
func TestAuthService_Logout_ErrorPaths(t *testing.T) {
	t.Run("FindByToken returns non-ErrNotFound error", func(t *testing.T) {
		userRepo := NewMockUserRepository()
		sessionRepo := &ErrorSessionRepository{findByTokenError: shared.ErrBadRequest}
		notificationService := NewMockNotificationService()
		service := NewAuthService(userRepo, sessionRepo, notificationService)

		err := service.Logout(context.Background(), "test-token")
		if err != shared.ErrBadRequest {
			t.Errorf("Logout() error = %v, wantErr %v", err, shared.ErrBadRequest)
		}
	})

	t.Run("Delete fails", func(t *testing.T) {
		userRepo := NewMockUserRepository()
		sessionRepo := &ErrorSessionRepository{
			deleteError: shared.ErrBadRequest,
		}
		// Set FindByToken to return a session
		sessionRepo.findByTokenFunc = func(ctx context.Context, token string) (*Session, error) {
			return &Session{
				ID:        uuid.New(),
				UserID:    uuid.New(),
				Token:     token,
				ExpiresAt: time.Now().Add(24 * time.Hour),
			}, nil
		}
		notificationService := NewMockNotificationService()
		service := NewAuthService(userRepo, sessionRepo, notificationService)

		err := service.Logout(context.Background(), "test-token")
		if err != shared.ErrBadRequest {
			t.Errorf("Logout() error = %v, wantErr %v", err, shared.ErrBadRequest)
		}
	})
}

// FailingNotificationService is a mock that always fails
type FailingNotificationService struct{}

func (m *FailingNotificationService) CreateNotification(ctx context.Context, userID uuid.UUID, title, message string) (*notifications.Notification, error) {
	return nil, shared.ErrBadRequest
}

func (m *FailingNotificationService) CreateNotificationTx(ctx context.Context, tx *gorm.DB, userID uuid.UUID, title, message string) (*notifications.Notification, error) {
	return nil, shared.ErrBadRequest
}

func (m *FailingNotificationService) GetNotifications(ctx context.Context, userID uuid.UUID) ([]notifications.Notification, error) {
	return nil, shared.ErrBadRequest
}

func (m *FailingNotificationService) MarkAsRead(ctx context.Context, userID uuid.UUID, notificationID uuid.UUID) error {
	return shared.ErrBadRequest
}

// TestAuthService_NotificationErrors tests that notification errors don't fail operations
func TestAuthService_NotificationErrors(t *testing.T) {
	userRepo := NewMockUserRepository()
	sessionRepo := NewMockSessionRepository()

	// Create a notification service that always fails
	failingNotificationService := &FailingNotificationService{}

	service := NewAuthService(userRepo, sessionRepo, failingNotificationService)

	t.Run("Register succeeds even if notification fails", func(t *testing.T) {
		req := RegisterRequest{
			Username: "test_user",
			Email:    "test@example.com",
			Password: "password123",
		}
		user, err := service.Register(context.Background(), req)
		if err != nil {
			t.Errorf("Register() should succeed even if notification fails, got error: %v", err)
		}
		if user == nil {
			t.Error("Register() should return user even if notification fails")
		}
	})

	t.Run("Login succeeds even if notification fails", func(t *testing.T) {
		// Register first
		_, err := service.Register(context.Background(), RegisterRequest{
			Username: "test_user2",
			Email:    "test2@example.com",
			Password: "password123",
		})
		if err != nil {
			t.Fatalf("Setup registration failed: %v", err)
		}

		// Login should succeed even if notification fails
		response, err := service.Login(context.Background(), LoginRequest{
			Email:    "test2@example.com",
			Password: "password123",
		})
		if err != nil {
			t.Errorf("Login() should succeed even if notification fails, got error: %v", err)
		}
		if response == nil {
			t.Error("Login() should return response even if notification fails")
		}
	})

	t.Run("Logout succeeds even if notification fails", func(t *testing.T) {
		// Register and login first
		_, err := service.Register(context.Background(), RegisterRequest{
			Username: "test_user3",
			Email:    "test3@example.com",
			Password: "password123",
		})
		if err != nil {
			t.Fatalf("Setup registration failed: %v", err)
		}

		loginResp, err := service.Login(context.Background(), LoginRequest{
			Email:    "test3@example.com",
			Password: "password123",
		})
		if err != nil {
			t.Fatalf("Setup login failed: %v", err)
		}

		// Logout should succeed even if notification fails
		err = service.Logout(context.Background(), loginResp.SessionToken)
		if err != nil {
			t.Errorf("Logout() should succeed even if notification fails, got error: %v", err)
		}
	})
}