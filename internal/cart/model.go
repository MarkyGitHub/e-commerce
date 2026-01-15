package cart

import (
	"time"

	"github.com/google/uuid"
)

// Cart repräsentiert einen Warenkorb für einen Benutzer
type Cart struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;uniqueIndex;not null" json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CartItem repräsentiert einen Artikel im Warenkorb
type CartItem struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	CartID    uuid.UUID `gorm:"type:uuid;not null;index" json:"cart_id"`
	ProductID uuid.UUID `gorm:"type:uuid;not null" json:"product_id"`
	Quantity  int       `gorm:"not null" json:"quantity"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// AddItemRequest repräsentiert Anfrage zum Hinzufügen eines Artikels
type AddItemRequest struct {
	ProductID uuid.UUID `json:"product_id"`
	Quantity  int       `json:"quantity"`
}

// UpdateItemRequest repräsentiert Anfrage zum Aktualisieren eines Warenkorb-Artikels
type UpdateItemRequest struct {
	Quantity int `json:"quantity"`
}

// CartResponse repräsentiert den Warenkorb in API-Antworten
type CartResponse struct {
	ID        uuid.UUID          `json:"id"`
	UserID    uuid.UUID          `json:"user_id"`
	Items     []CartItemResponse `json:"items"`
	CreatedAt time.Time          `json:"created_at"`
	UpdatedAt time.Time          `json:"updated_at"`
}

// CartItemResponse repräsentiert einen Warenkorb-Artikel in API-Antworten
type CartItemResponse struct {
	ID        uuid.UUID `json:"id"`
	ProductID uuid.UUID `json:"product_id"`
	Quantity  int       `json:"quantity"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ToResponse konvertiert ein CartItem zu CartItemResponse
func (ci *CartItem) ToResponse() CartItemResponse {
	return CartItemResponse{
		ID:        ci.ID,
		ProductID: ci.ProductID,
		Quantity:  ci.Quantity,
		CreatedAt: ci.CreatedAt,
		UpdatedAt: ci.UpdatedAt,
	}
}

// ToCartResponse konvertiert einen Cart und seine Artikel zu CartResponse
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
