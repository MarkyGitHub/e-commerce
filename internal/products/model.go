package products

import (
	"time"

	"github.com/google/uuid"
)

// Product represents a product in the catalog
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

// CreateProductRequest represents the request to create a product
type CreateProductRequest struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Price       float64    `json:"price"`
	CategoryID  *uuid.UUID `json:"category_id"`
}

// UpdateProductRequest represents the request to update a product
type UpdateProductRequest struct {
	Name        *string     `json:"name,omitempty"`
	Description *string     `json:"description,omitempty"`
	Price       *float64    `json:"price,omitempty"`
	CategoryID  *uuid.UUID  `json:"category_id,omitempty"`
	IsActive    *bool       `json:"is_active,omitempty"`
}

// Validate validates the CreateProductRequest
func (r *CreateProductRequest) Validate() error {
	if r.Name == "" {
		return ErrInvalidProductName
	}
	if r.Price < 0 {
		return ErrInvalidProductPrice
	}
	return nil
}

// Product-specific errors
var (
	ErrInvalidProductName  = &ValidationError{Field: "name", Message: "product name is required"}
	ErrInvalidProductPrice = &ValidationError{Field: "price", Message: "product price cannot be negative"}
)

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}
