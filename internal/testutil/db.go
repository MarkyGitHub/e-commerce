package testutil

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var openPostgres = func(dsn string) (*gorm.DB, error) {
	return gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), // Silence logs in tests
	})
}

// SetupTestDB erstellt eine Test-Datenbankverbindung
// Aufrufer müssen Migrationen selbst ausführen um Import-Zyklen zu vermeiden
// Standardmäßig wird eine In-Memory SQLite-Datenbank verwendet, so dass Tests ohne externe Abhängigkeiten laufen.
// Für Tests gegen Postgres setze TEST_DB_DIALECT=postgres und DB_*-Umgebungsvariablen.
func SetupTestDB(t *testing.T) *gorm.DB {
	if os.Getenv("TEST_DB_DIALECT") != "postgres" {
		// Eindeutige In-Memory-DB pro Test um Testübergreifende Störungen zu vermeiden.
		dsn := fmt.Sprintf("file:ecommerce_test_%d?mode=memory&cache=shared", time.Now().UnixNano())
		gormDB, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
		})
		if err != nil {
			t.Fatalf("Failed to open sqlite test database: %v", err)
		}
		// Fremdschlüssel-Constraints aktivieren.
		gormDB.Exec("PRAGMA foreign_keys = ON")

		t.Cleanup(func() {
			sqlDB, _ := gormDB.DB()
			if sqlDB != nil {
				sqlDB.Close()
			}
		})
		return gormDB
	}

	// Test-Datenbankkonfiguration verwenden (hier dupliziert um Import-Zyklus zu vermeiden)
	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "ecommerce")
	password := getEnv("DB_PASSWORD", "ecommerce")
	dbName := getEnv("DB_NAME", "ecommerce_test")
	sslMode := getEnv("DB_SSLMODE", "disable")

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbName, sslMode,
	)

	gormDB, err := openPostgres(dsn)
	if err != nil {
		// Test überspringen falls Datenbank nicht verfügbar (z.B. nicht gestartet)
		t.Skipf("Skipping test: database not available: %v", err)
		return nil
	}

	// Nach Test aufräumen
	t.Cleanup(func() {
		// Alle Tabellen bereinigen
		gormDB.Exec("TRUNCATE TABLE users, sessions, products, inventory, carts, cart_items, orders, order_items, notifications RESTART IDENTITY CASCADE")
		sqlDB, _ := gormDB.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	})

	return gormDB
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
