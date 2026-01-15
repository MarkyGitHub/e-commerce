package products

import (
	"time"

	"github.com/google/uuid"
)

// Product repräsentiert ein Produkt im Katalog
type Product struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	Name        string     `gorm:"not null" json:"name"`
	Description string     `json:"description"`
	Price       float64    `gorm:"not null" json:"price"`
	CategoryID  *uuid.UUID `gorm:"type:uuid" json:"category_id"`
	IsActive    bool       `gorm:"not null;default:true" json:"is_active"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// CreateProductRequest repräsentiert Anfrage zum Erstellen eines Produkts
type CreateProductRequest struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Price       float64    `json:"price"`
	CategoryID  *uuid.UUID `json:"category_id"`
}

// UpdateProductRequest repräsentiert Anfrage zum Aktualisieren eines Produkts
type UpdateProductRequest struct {
	Name        *string     `json:"name,omitempty"`
	Description *string     `json:"description,omitempty"`
	Price       *float64    `json:"price,omitempty"`
	CategoryID  *uuid.UUID  `json:"category_id,omitempty"`
	IsActive    *bool       `json:"is_active,omitempty"`
}

// Validate validiert CreateProductRequest
func (r *CreateProductRequest) Validate() error {
	if r.Name == "" {
		return ErrInvalidProductName
	}
	if r.Price < 0 {
		return ErrInvalidProductPrice
	}
	return nil
}

// Produkt-spezifische Fehler
var (
	ErrInvalidProductName  = &ValidationError{Field: "name", Message: "product name is required"}
	ErrInvalidProductPrice = &ValidationError{Field: "price", Message: "product price cannot be negative"}
)

// ValidationError repräsentiert einen Validierungsfehler
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}
