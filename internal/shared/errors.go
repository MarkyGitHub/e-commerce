package shared

import (
	"errors"

	"github.com/google/uuid"
)

// contextKey ist ein Typ für Kontext-Schlüssel
type contextKey string

// UserContextKey ist der Schlüssel zum Speichern des Benutzers im Request-Kontext
// Hierher verschoben um Import-Zyklen zu vermeiden
const UserContextKey contextKey = "user"

// UserIDGetter ist minimales Interface zum Extrahieren der User ID aus Kontext
// Vermeidet Import-Zyklen wenn Handler User ID benötigen
type UserIDGetter interface {
	GetID() uuid.UUID
}

// Allgemeine Fehlertypen für die Anwendung
var (
	// ErrNotFound wenn Ressource nicht gefunden
	ErrNotFound = errors.New("resource not found")

	// ErrConflict bei Konflikten (z.B. doppelte Email)
	ErrConflict = errors.New("resource already exists")

	// ErrUnauthorized wenn Authentifizierung fehlschlägt
	ErrUnauthorized = errors.New("unauthorized")

	// ErrForbidden wenn Benutzer keine Berechtigung hat
	ErrForbidden = errors.New("forbidden")

	// ErrBadRequest bei ungültiger Anfrage
	ErrBadRequest = errors.New("bad request")

	// ErrInvalidCredentials bei ungültigen Anmeldedaten
	ErrInvalidCredentials = errors.New("invalid credentials")

	// ErrInvalidEmail bei ungültigem Email-Format
	ErrInvalidEmail = errors.New("invalid email format")

	// ErrPasswordTooShort wenn Passwort zu kurz
	ErrPasswordTooShort = errors.New("password must be at least 8 characters")

	// ErrInsufficientInventory bei unzureichendem Lagerbestand
	ErrInsufficientInventory = errors.New("insufficient inventory")

	// ErrEmptyCart bei Bestellung aus leerem Warenkorb
	ErrEmptyCart = errors.New("cart is empty")

	// ErrNegativeQuantity wenn Menge negativ wird
	ErrNegativeQuantity = errors.New("quantity cannot be negative")
)
