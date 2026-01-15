package orders

import (
	"github.com/google/uuid"
)

// CreateOrderRequest repräsentiert Anfrage zum Erstellen einer Bestellung
type CreateOrderRequest struct {
	Items            []OrderItemRequest `json:"items"`
	ShippingStreet   string             `json:"shipping_street"`
	ShippingCity     string             `json:"shipping_city"`
	ShippingZip      string             `json:"shipping_zip"`
	ShippingCountry  string             `json:"shipping_country"`
}

// OrderItemRequest repräsentiert einen Artikel in der Bestellanfrage
type OrderItemRequest struct {
	ProductID uuid.UUID `json:"product_id"`
	Quantity  int       `json:"quantity"`
}
