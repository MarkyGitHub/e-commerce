package auth

import (
	"time"

	"github.com/google/uuid"
)

// User repräsentiert einen registrierten Benutzer
type User struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Username     string    `gorm:"not null" json:"username"`
	Email        string    `gorm:"uniqueIndex;not null" json:"email"`
	PasswordHash string    `gorm:"not null" json:"-"`
	Role         string    `gorm:"not null;default:customer" json:"role"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Session repräsentiert eine authentifizierte Benutzersitzung
type Session struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	Token     string    `gorm:"uniqueIndex;not null" json:"token"`
	ExpiresAt time.Time `gorm:"not null" json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

// RegisterRequest repräsentiert die Registrierungsanfrage
type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginRequest repräsentiert die Login-Anfrage
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginResponse repräsentiert die Login-Antwort
type LoginResponse struct {
	SessionToken string    `json:"session_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	User         User      `json:"user"`
}

// UserResponse repräsentiert einen Benutzer in API-Antworten
type UserResponse struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

// ToResponse konvertiert User zu UserResponse
func (u *User) ToResponse() UserResponse {
	return UserResponse{
		ID:        u.ID,
		Username:  u.Username,
		Email:     u.Email,
		Role:      u.Role,
		CreatedAt: u.CreatedAt,
	}
}

// IsAdmin gibt true zurück wenn Benutzer Admin ist
func (u *User) IsAdmin() bool {
	return u.Role == "admin"
}

// GetID gibt die Benutzer-ID zurück (implementiert shared.UserIDGetter)
func (u *User) GetID() uuid.UUID {
	return u.ID
}

// Role constants
const (
	RoleCustomer = "customer"
	RoleAdmin    = "admin"
)
