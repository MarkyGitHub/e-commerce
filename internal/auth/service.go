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

// AuthService definiert Interface für Authentifizierung
type AuthService interface {
	Register(ctx context.Context, req RegisterRequest) (*User, error)
	Login(ctx context.Context, req LoginRequest) (*LoginResponse, error)
	Logout(ctx context.Context, token string) error
	ValidateSession(ctx context.Context, token string) (*User, error)
}

// authService implementiert AuthService
type authService struct {
	userRepo            UserRepository
	sessionRepo         SessionRepository
	notificationService notifications.NotificationService
}

// NewAuthService erstellt einen neuen AuthService
func NewAuthService(userRepo UserRepository, sessionRepo SessionRepository, notificationService notifications.NotificationService) AuthService {
	return &authService{
		userRepo:            userRepo,
		sessionRepo:         sessionRepo,
		notificationService: notificationService,
	}
}

// Register erstellt ein neues Benutzerkonto
func (s *authService) Register(ctx context.Context, req RegisterRequest) (*User, error) {
	// E-Mail-Format validieren
	if !isValidEmail(req.Email) {
		return nil, shared.ErrInvalidEmail
	}

	// Passwortlänge validieren
	if len(req.Password) < 8 {
		return nil, shared.ErrPasswordTooShort
	}

	// Benutzername validieren
	if req.Username == "" {
		return nil, shared.ErrBadRequest
	}

	// Passwort hashen
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

	// Benachrichtigung für Registrierung erstellen
	title := "Willkommen"
	message := "Dein Konto wurde erfolgreich erstellt."
	if _, err := s.notificationService.CreateNotification(ctx, user.ID, title, message); err != nil {
		// Fehler loggen, aber Registrierung nicht abbrechen
	}

	return user, nil
}

// Login authentifiziert Benutzer und erstellt Session
func (s *authService) Login(ctx context.Context, req LoginRequest) (*LoginResponse, error) {
	// Benutzer per E-Mail finden
	user, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		if err == shared.ErrNotFound {
			return nil, shared.ErrInvalidCredentials
		}
		return nil, err
	}

	// Passwort verifizieren
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, shared.ErrInvalidCredentials
	}

	// Session-Token generieren
	token, err := generateToken()
	if err != nil {
		return nil, err
	}

	// Session erstellen
	session := &Session{
		ID:        uuid.New(),
		UserID:    user.ID,
		Token:     token,
		ExpiresAt: time.Now().Add(24 * time.Hour), // 24-Stunden-Session
	}

	if err := s.sessionRepo.Create(ctx, session); err != nil {
		return nil, err
	}

	// Benachrichtigung für Login erstellen
	title := "Anmeldung"
	message := "Du hast dich erfolgreich angemeldet."
	if _, err := s.notificationService.CreateNotification(ctx, user.ID, title, message); err != nil {
		// Fehler loggen, aber Login nicht abbrechen
	}

	return &LoginResponse{
		SessionToken: session.Token,
		ExpiresAt:    session.ExpiresAt,
		User:         *user,
	}, nil
}

// Logout macht eine Session ungültig
func (s *authService) Logout(ctx context.Context, token string) error {
	// Session holen um User-ID vor dem Löschen zu finden
	session, err := s.sessionRepo.FindByToken(ctx, token)
	if err != nil {
		return err
	}

	// Session löschen
	if err := s.sessionRepo.Delete(ctx, token); err != nil {
		return err
	}

	// Benachrichtigung für Logout erstellen
	title := "Abmeldung"
	message := "Du hast dich erfolgreich abgemeldet."
	if _, err := s.notificationService.CreateNotification(ctx, session.UserID, title, message); err != nil {
		// Fehler loggen, aber Logout nicht abbrechen
	}

	return nil
}

// ValidateSession validiert Session-Token und gibt Benutzer zurück
func (s *authService) ValidateSession(ctx context.Context, token string) (*User, error) {
	session, err := s.sessionRepo.FindByToken(ctx, token)
	if err != nil {
		if err == shared.ErrNotFound {
			return nil, shared.ErrUnauthorized
		}
		return nil, err
	}

	// Prüfen ob Session abgelaufen ist
	if time.Now().After(session.ExpiresAt) {
		// Abgelaufene Session löschen
		s.sessionRepo.Delete(ctx, token)
		return nil, shared.ErrUnauthorized
	}

	// Benutzer holen
	user, err := s.userRepo.FindByID(ctx, session.UserID)
	if err != nil {
		return nil, shared.ErrUnauthorized
	}

	return user, nil
}

// isValidEmail validiert E-Mail-Format mit Regex
func isValidEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

// generateToken generiert einen kryptographisch sicheren Zufalls-Token
func generateToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
