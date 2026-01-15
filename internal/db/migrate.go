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

// Migrate führt Datenbankmigrationen für alle Modelle aus
func Migrate(db *gorm.DB) error {
	err := db.AutoMigrate(
		// Auth-Modelle
		&auth.User{},
		&auth.Session{},
		// Produkte
		&products.Product{},
		// Bestand
		&inventory.Inventory{},
		// Warenkorb
		&cart.Cart{},
		&cart.CartItem{},
		// Bestellungen
		&orders.Order{},
		&orders.OrderItem{},
		// Benachrichtigungen
		&notifications.Notification{},
	)
	if err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}
