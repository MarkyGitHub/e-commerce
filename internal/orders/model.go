package orders

import (
	"time"

	"github.com/google/uuid"
)

// Order represents a customer order
type Order struct {
	ID              uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	UserID          uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	TotalPrice      float64   `gorm:"not null" json:"total_price"`
	Status          string    `gorm:"not null;default:pending" json:"status"`
	ShippingStreet  string    `gorm:"type:varchar(255);not null" json:"shipping_street"`
	ShippingCity    string    `gorm:"type:varchar(255);not null" json:"shipping_city"`
	ShippingZip     string    `gorm:"type:varchar(50);not null" json:"shipping_zip"`
	ShippingCountry string    `gorm:"type:varchar(100);not null" json:"shipping_country"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// OrderItem represents an item in an order
type OrderItem struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	OrderID   uuid.UUID `gorm:"type:uuid;not null;index" json:"order_id"`
	ProductID uuid.UUID `gorm:"type:uuid;not null" json:"product_id"`
	Quantity  int       `gorm:"not null" json:"quantity"`
	Price     float64   `gorm:"not null" json:"price"`
	CreatedAt time.Time `json:"created_at"`
}

// Order status constants
const (
	StatusPending    = "pending"
	StatusProcessing = "processing"
	StatusCompleted  = "completed"
	StatusCancelled  = "cancelled"
)

// OrderResponse represents an order in API responses
type OrderResponse struct {
	ID              uuid.UUID           `json:"id"`
	UserID          uuid.UUID           `json:"user_id"`
	TotalPrice      float64             `json:"total_price"`
	Status          string              `json:"status"`
	ShippingStreet  string              `json:"shipping_street"`
	ShippingCity    string              `json:"shipping_city"`
	ShippingZip     string              `json:"shipping_zip"`
	ShippingCountry string              `json:"shipping_country"`
	Items           []OrderItemResponse `json:"items,omitempty"`
	CreatedAt       time.Time           `json:"created_at"`
	UpdatedAt       time.Time           `json:"updated_at"`
}

// OrderItemResponse represents an order item in API responses
type OrderItemResponse struct {
	ID        uuid.UUID `json:"id"`
	ProductID uuid.UUID `json:"product_id"`
	Quantity  int       `json:"quantity"`
	Price     float64   `json:"price"`
	CreatedAt time.Time `json:"created_at"`
}

// ToResponse converts an Order to OrderResponse
func (o *Order) ToResponse() OrderResponse {
	return OrderResponse{
		ID:              o.ID,
		UserID:          o.UserID,
		TotalPrice:      o.TotalPrice,
		Status:          o.Status,
		ShippingStreet:  o.ShippingStreet,
		ShippingCity:    o.ShippingCity,
		ShippingZip:     o.ShippingZip,
		ShippingCountry: o.ShippingCountry,
		CreatedAt:       o.CreatedAt,
		UpdatedAt:       o.UpdatedAt,
	}
}

// ToResponse converts an OrderItem to OrderItemResponse
func (oi *OrderItem) ToResponse() OrderItemResponse {
	return OrderItemResponse{
		ID:        oi.ID,
		ProductID: oi.ProductID,
		Quantity:  oi.Quantity,
		Price:     oi.Price,
		CreatedAt: oi.CreatedAt,
	}
}

// ToOrderResponseWithItems converts an Order and its items to OrderResponse
func ToOrderResponseWithItems(order *Order, items []OrderItem) OrderResponse {
	response := order.ToResponse()
	response.Items = make([]OrderItemResponse, len(items))
	for i, item := range items {
		response.Items[i] = item.ToResponse()
	}
	return response
}

// ToOrderResponseList converts a slice of Orders to OrderResponses
func ToOrderResponseList(orders []Order) []OrderResponse {
	responses := make([]OrderResponse, len(orders))
	for i, order := range orders {
		responses[i] = order.ToResponse()
	}
	return responses
}
