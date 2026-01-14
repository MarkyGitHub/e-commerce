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

// SetupTestDB creates a test database connection
// Callers must run migrations themselves to avoid import cycles
// By default, this uses an in-memory SQLite database so tests run without external dependencies.
// To run tests against Postgres instead, set TEST_DB_DIALECT=postgres and provide DB_* env vars.
func SetupTestDB(t *testing.T) *gorm.DB {
	if os.Getenv("TEST_DB_DIALECT") != "postgres" {
		// Unique in-memory DB per test to avoid cross-test interference.
		dsn := fmt.Sprintf("file:ecommerce_test_%d?mode=memory&cache=shared", time.Now().UnixNano())
		gormDB, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
		})
		if err != nil {
			t.Fatalf("Failed to open sqlite test database: %v", err)
		}
		// Ensure foreign key constraints are enforced.
		gormDB.Exec("PRAGMA foreign_keys = ON")

		t.Cleanup(func() {
			sqlDB, _ := gormDB.DB()
			if sqlDB != nil {
				sqlDB.Close()
			}
		})
		return gormDB
	}

	// Use test database configuration (duplicated here to avoid import cycle)
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
		// Skip test if database is not available (e.g., not running)
		t.Skipf("Skipping test: database not available: %v", err)
		return nil
	}

	// Clean up after test
	t.Cleanup(func() {
		// Clean up all tables
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
