package db

import (
	"fmt"

	"gorm.io/gorm"

	"github.com/MarkyGitHub/e-commerce/internal/auth"
	"github.com/MarkyGitHub/e-commerce/internal/cart"
	"github.com/MarkyGitHub/e-commerce/internal/inventory"
	"github.com/MarkyGitHub/e-commerce/internal/notifications"
	"github.com/MarkyGitHub/e-commerce/internal/orders"
	"github.com/MarkyGitHub/e-commerce/internal/products"
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
