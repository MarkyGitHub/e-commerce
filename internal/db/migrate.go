package db

import (
	"fmt"

	"gorm.io/gorm"

	"github.com/mschl/ecommerce-api/internal/auth"
	"github.com/mschl/ecommerce-api/internal/cart"
	"github.com/mschl/ecommerce-api/internal/inventory"
	"github.com/mschl/ecommerce-api/internal/notifications"
	"github.com/mschl/ecommerce-api/internal/orders"
	"github.com/mschl/ecommerce-api/internal/products"
)

// Migrate runs database migrations for all models
func Migrate(db *gorm.DB) error {
	err := db.AutoMigrate(
		// Auth models
		&auth.User{},
		&auth.Session{},
		// Products
		&products.Product{},
		// Inventory
		&inventory.Inventory{},
		// Cart
		&cart.Cart{},
		&cart.CartItem{},
		// Orders
		&orders.Order{},
		&orders.OrderItem{},
		// Notifications
		&notifications.Notification{},
	)
	if err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}
