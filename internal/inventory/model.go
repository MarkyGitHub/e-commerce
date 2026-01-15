package inventory

import (
	"time"

	"github.com/google/uuid"
)

// Inventory repräsentiert den Lagerbestand für ein Produkt
type Inventory struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	ProductID uuid.UUID `gorm:"type:uuid;uniqueIndex;not null" json:"product_id"`
	Quantity  int       `gorm:"not null;default:0" json:"quantity"`
	UpdatedAt time.Time `json:"updated_at"`
}

// UpdateInventoryRequest repräsentiert Anfrage zum Aktualisieren des Inventars (PATCH - Erhöhung)
type UpdateInventoryRequest struct {
	IncrementBy int    `json:"increment_by"`
	Reason      string `json:"reason"`
}

// InventoryResponse repräsentiert das Inventar in API-Antworten
type InventoryResponse struct {
	ProductID uuid.UUID `json:"product_id"`
	Quantity  int       `json:"quantity"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ToResponse konvertiert ein Inventory zu InventoryResponse
func (i *Inventory) ToResponse() InventoryResponse {
	return InventoryResponse{
		ProductID: i.ProductID,
		Quantity:  i.Quantity,
		UpdatedAt: i.UpdatedAt,
	}
}
