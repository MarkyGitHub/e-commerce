package orders

import (
	"github.com/google/uuid"
)

// CreateOrderRequest represents the request to create an order
type CreateOrderRequest struct {
	Items            []OrderItemRequest `json:"items"`
	ShippingStreet   string             `json:"shipping_street"`
	ShippingCity     string             `json:"shipping_city"`
	ShippingZip      string             `json:"shipping_zip"`
	ShippingCountry  string             `json:"shipping_country"`
}

// OrderItemRequest represents an item in the create order request
type OrderItemRequest struct {
	ProductID uuid.UUID `json:"product_id"`
	Quantity  int       `json:"quantity"`
}
