package db

import (
	"strings"
	"testing"

	"github.com/mschl/ecommerce-api/internal/config"
)

func TestConnectWithDSN_InvalidDSN_ReturnsError(t *testing.T) {
	// This should fail without needing a running database.
	if _, err := ConnectWithDSN("this is not a valid postgres dsn"); err == nil {
		t.Fatalf("expected error")
	}
}

func TestConnect_InvalidPortInDSN_ReturnsError(t *testing.T) {
	// Use an invalid port string so the driver fails fast during DSN parsing.
	cfg := config.DBConfig{
		Host:     "localhost",
		Port:     "not-a-port",
		User:     "x",
		Password: "x",
		DBName:   "x",
		SSLMode:  "disable",
	}
	if _, err := Connect(cfg); err == nil {
		t.Fatalf("expected error")
	}
}

func TestDBConfig_DSN(t *testing.T) {
	cfg := config.DBConfig{
		Host:     "db.example",
		Port:     "5433",
		User:     "alice",
		Password: "secret",
		DBName:   "shop",
		SSLMode:  "require",
	}

	dsn := cfg.DSN()
	wantParts := []string{
		"host=db.example",
		"port=5433",
		"user=alice",
		"password=secret",
		"dbname=shop",
		"sslmode=require",
	}
	for _, part := range wantParts {
		if !strings.Contains(dsn, part) {
			t.Fatalf("expected DSN to contain %q, got %q", part, dsn)
		}
	}
}
