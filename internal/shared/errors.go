package shared

import (
	"errors"

	"github.com/google/uuid"
)

// contextKey is a type for context keys
type contextKey string

// UserContextKey is the key used to store the user in the request context
// Moved here to avoid import cycles
const UserContextKey contextKey = "user"

// UserIDGetter is a minimal interface to extract user ID from context
// This avoids import cycles when handlers need user ID
type UserIDGetter interface {
	GetID() uuid.UUID
}

// Common error types for the application
var (
	// ErrNotFound is returned when a resource is not found
	ErrNotFound = errors.New("resource not found")

	// ErrConflict is returned when there is a conflict (e.g., duplicate email)
	ErrConflict = errors.New("resource already exists")

	// ErrUnauthorized is returned when authentication fails
	ErrUnauthorized = errors.New("unauthorized")

	// ErrForbidden is returned when the user lacks permission
	ErrForbidden = errors.New("forbidden")

	// ErrBadRequest is returned when the request is invalid
	ErrBadRequest = errors.New("bad request")

	// ErrInvalidCredentials is returned when login credentials are invalid
	ErrInvalidCredentials = errors.New("invalid credentials")

	// ErrInvalidEmail is returned when email format is invalid
	ErrInvalidEmail = errors.New("invalid email format")

	// ErrPasswordTooShort is returned when password is too short
	ErrPasswordTooShort = errors.New("password must be at least 8 characters")

	// ErrInsufficientInventory is returned when inventory is insufficient
	ErrInsufficientInventory = errors.New("insufficient inventory")

	// ErrEmptyCart is returned when trying to create an order from an empty cart
	ErrEmptyCart = errors.New("cart is empty")

	// ErrNegativeQuantity is returned when quantity would become negative
	ErrNegativeQuantity = errors.New("quantity cannot be negative")
)
