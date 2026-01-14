package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"regexp"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/mschl/ecommerce-api/internal/notifications"
	"github.com/mschl/ecommerce-api/internal/shared"
)

// AuthService defines the interface for authentication operations
type AuthService interface {
	Register(ctx context.Context, req RegisterRequest) (*User, error)
	Login(ctx context.Context, req LoginRequest) (*LoginResponse, error)
	Logout(ctx context.Context, token string) error
	ValidateSession(ctx context.Context, token string) (*User, error)
}

// authService implements AuthService
type authService struct {
	userRepo            UserRepository
	sessionRepo         SessionRepository
	notificationService notifications.NotificationService
}

// NewAuthService creates a new AuthService
func NewAuthService(userRepo UserRepository, sessionRepo SessionRepository, notificationService notifications.NotificationService) AuthService {
	return &authService{
		userRepo:            userRepo,
		sessionRepo:         sessionRepo,
		notificationService: notificationService,
	}
}

// Register creates a new user account
func (s *authService) Register(ctx context.Context, req RegisterRequest) (*User, error) {
	// Validate email format
	if !isValidEmail(req.Email) {
		return nil, shared.ErrInvalidEmail
	}

	// Validate password length
	if len(req.Password) < 8 {
		return nil, shared.ErrPasswordTooShort
	}

	// Validate username
	if req.Username == "" {
		return nil, shared.ErrBadRequest
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &User{
		ID:           uuid.New(),
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		Role:         RoleCustomer,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	// Create notification for registration
	title := "Welcome"
	message := "Your account has been successfully created."
	if _, err := s.notificationService.CreateNotification(ctx, user.ID, title, message); err != nil {
		// Log error but don't fail registration
		// In production, you might want to use a logger here
	}

	return user, nil
}

// Login authenticates a user and creates a session
func (s *authService) Login(ctx context.Context, req LoginRequest) (*LoginResponse, error) {
	// Find user by email
	user, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		if err == shared.ErrNotFound {
			return nil, shared.ErrInvalidCredentials
		}
		return nil, err
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, shared.ErrInvalidCredentials
	}

	// Generate session token
	token, err := generateToken()
	if err != nil {
		return nil, err
	}

	// Create session
	session := &Session{
		ID:        uuid.New(),
		UserID:    user.ID,
		Token:     token,
		ExpiresAt: time.Now().Add(24 * time.Hour), // 24 hour session
	}

	if err := s.sessionRepo.Create(ctx, session); err != nil {
		return nil, err
	}

	// Create notification for login
	title := "Login"
	message := "You have successfully logged in."
	if _, err := s.notificationService.CreateNotification(ctx, user.ID, title, message); err != nil {
		// Log error but don't fail login
	}

	return &LoginResponse{
		SessionToken: session.Token,
		ExpiresAt:    session.ExpiresAt,
		User:         *user,
	}, nil
}

// Logout invalidates a session
func (s *authService) Logout(ctx context.Context, token string) error {
	// Get session to find user ID before deleting
	session, err := s.sessionRepo.FindByToken(ctx, token)
	if err != nil {
		return err
	}

	// Delete session
	if err := s.sessionRepo.Delete(ctx, token); err != nil {
		return err
	}

	// Create notification for logout
	title := "Logout"
	message := "You have successfully logged out."
	if _, err := s.notificationService.CreateNotification(ctx, session.UserID, title, message); err != nil {
		// Log error but don't fail logout
	}

	return nil
}

// ValidateSession validates a session token and returns the associated user
func (s *authService) ValidateSession(ctx context.Context, token string) (*User, error) {
	session, err := s.sessionRepo.FindByToken(ctx, token)
	if err != nil {
		if err == shared.ErrNotFound {
			return nil, shared.ErrUnauthorized
		}
		return nil, err
	}

	// Check if session is expired
	if time.Now().After(session.ExpiresAt) {
		// Delete expired session
		s.sessionRepo.Delete(ctx, token)
		return nil, shared.ErrUnauthorized
	}

	// Get user
	user, err := s.userRepo.FindByID(ctx, session.UserID)
	if err != nil {
		return nil, shared.ErrUnauthorized
	}

	return user, nil
}

// isValidEmail validates email format using regex
func isValidEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

// generateToken generates a cryptographically secure random token
func generateToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
