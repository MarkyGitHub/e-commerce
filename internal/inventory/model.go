package inventory

import (
	"time"

	"github.com/google/uuid"
)

// Inventory represents the stock quantity for a product
type Inventory struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	ProductID uuid.UUID `gorm:"type:uuid;uniqueIndex;not null" json:"product_id"`
	Quantity  int       `gorm:"not null;default:0" json:"quantity"`
	UpdatedAt time.Time `json:"updated_at"`
}

// UpdateInventoryRequest represents the request to update inventory (PATCH - increment)
type UpdateInventoryRequest struct {
	IncrementBy int    `json:"increment_by"`
	Reason      string `json:"reason"`
}

// InventoryResponse represents the inventory in API responses
type InventoryResponse struct {
	ProductID uuid.UUID `json:"product_id"`
	Quantity  int       `json:"quantity"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ToResponse converts an Inventory to InventoryResponse
func (i *Inventory) ToResponse() InventoryResponse {
	return InventoryResponse{
		ProductID: i.ProductID,
		Quantity:  i.Quantity,
		UpdatedAt: i.UpdatedAt,
	}
}
