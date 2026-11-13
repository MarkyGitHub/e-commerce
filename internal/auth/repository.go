package auth

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/MarkyGitHub/e-commerce/internal/shared"
)

// UserRepository definiert Interface für Benutzer-Datenzugriff
type UserRepository interface {
	Create(ctx context.Context, user *User) error
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindByID(ctx context.Context, id uuid.UUID) (*User, error)
}

// SessionRepository definiert Interface für Session-Datenzugriff
type SessionRepository interface {
	Create(ctx context.Context, session *Session) error
	FindByToken(ctx context.Context, token string) (*Session, error)
	Delete(ctx context.Context, token string) error
}

// userRepository implementiert UserRepository mit GORM
type userRepository struct {
	db *gorm.DB
}

// NewUserRepository erstellt ein neues UserRepository
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *User) error {
	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}
	result := r.db.WithContext(ctx).Create(user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrDuplicatedKey) || 
		   (result.Error != nil && containsDuplicateError(result.Error.Error())) {
			return shared.ErrConflict
		}
		return result.Error
	}
	return nil
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*User, error) {
	var user User
	result := r.db.WithContext(ctx).Where("email = ?", email).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, shared.ErrNotFound
		}
		return nil, result.Error
	}
	return &user, nil
}

func (r *userRepository) FindByID(ctx context.Context, id uuid.UUID) (*User, error) {
	var user User
	result := r.db.WithContext(ctx).First(&user, "id = ?", id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, shared.ErrNotFound
		}
		return nil, result.Error
	}
	return &user, nil
}

// sessionRepository implementiert SessionRepository mit GORM
type sessionRepository struct {
	db *gorm.DB
}

// NewSessionRepository erstellt ein neues SessionRepository
func NewSessionRepository(db *gorm.DB) SessionRepository {
	return &sessionRepository{db: db}
}

func (r *sessionRepository) Create(ctx context.Context, session *Session) error {
	if session.ID == uuid.Nil {
		session.ID = uuid.New()
	}
	return r.db.WithContext(ctx).Create(session).Error
}

func (r *sessionRepository) FindByToken(ctx context.Context, token string) (*Session, error) {
	var session Session
	result := r.db.WithContext(ctx).Where("token = ?", token).First(&session)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, shared.ErrNotFound
		}
		return nil, result.Error
	}
	return &session, nil
}

func (r *sessionRepository) Delete(ctx context.Context, token string) error {
	result := r.db.WithContext(ctx).Where("token = ?", token).Delete(&Session{})
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// containsDuplicateError prüft ob Fehlermeldung auf Duplikat hinweist
func containsDuplicateError(errMsg string) bool {
	return len(errMsg) > 0 && (
		contains(errMsg, "duplicate key") ||
		contains(errMsg, "UNIQUE constraint failed") ||
		contains(errMsg, "Duplicate entry"))
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
