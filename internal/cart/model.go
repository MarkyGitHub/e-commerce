package cart

import (
	"time"

	"github.com/google/uuid"
)

// Cart represents a shopping cart for a user
type Cart struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;uniqueIndex;not null" json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CartItem represents an item in a shopping cart
type CartItem struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	CartID    uuid.UUID `gorm:"type:uuid;not null;index" json:"cart_id"`
	ProductID uuid.UUID `gorm:"type:uuid;not null" json:"product_id"`
	Quantity  int       `gorm:"not null" json:"quantity"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// AddItemRequest represents the request to add an item to the cart
type AddItemRequest struct {
	ProductID uuid.UUID `json:"product_id"`
	Quantity  int       `json:"quantity"`
}

// UpdateItemRequest represents the request to update a cart item
type UpdateItemRequest struct {
	Quantity int `json:"quantity"`
}

// CartResponse represents the cart in API responses
type CartResponse struct {
	ID        uuid.UUID          `json:"id"`
	UserID    uuid.UUID          `json:"user_id"`
	Items     []CartItemResponse `json:"items"`
	CreatedAt time.Time          `json:"created_at"`
	UpdatedAt time.Time          `json:"updated_at"`
}

// CartItemResponse represents a cart item in API responses
type CartItemResponse struct {
	ID        uuid.UUID `json:"id"`
	ProductID uuid.UUID `json:"product_id"`
	Quantity  int       `json:"quantity"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ToResponse converts a CartItem to CartItemResponse
func (ci *CartItem) ToResponse() CartItemResponse {
	return CartItemResponse{
		ID:        ci.ID,
		ProductID: ci.ProductID,
		Quantity:  ci.Quantity,
		CreatedAt: ci.CreatedAt,
		UpdatedAt: ci.UpdatedAt,
	}
}

// ToCartResponse converts a Cart and its items to CartResponse
func ToCartResponse(cart *Cart, items []CartItem) CartResponse {
	itemResponses := make([]CartItemResponse, len(items))
	for i, item := range items {
		itemResponses[i] = item.ToResponse()
	}
	return CartResponse{
		ID:        cart.ID,
		UserID:    cart.UserID,
		Items:     itemResponses,
		CreatedAt: cart.CreatedAt,
		UpdatedAt: cart.UpdatedAt,
	}
}
