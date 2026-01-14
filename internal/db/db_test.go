package db

import (
	"os"
	"strings"
	"testing"
)

func TestGetEnv_DefaultAndOverride(t *testing.T) {
	key := "DB_GETENV_TEST"
	_ = os.Unsetenv(key)

	if got := getEnv(key, "default"); got != "default" {
		t.Fatalf("expected default value, got %q", got)
	}

	t.Setenv(key, "override")
	if got := getEnv(key, "default"); got != "override" {
		t.Fatalf("expected override value, got %q", got)
	}
}

func TestNewConfigFromEnv_AndDSN(t *testing.T) {
	t.Setenv("DB_HOST", "db.example")
	t.Setenv("DB_PORT", "5433")
	t.Setenv("DB_USER", "alice")
	t.Setenv("DB_PASSWORD", "secret")
	t.Setenv("DB_NAME", "shop")
	t.Setenv("DB_SSLMODE", "require")

	cfg := NewConfigFromEnv()
	if cfg.Host != "db.example" || cfg.Port != "5433" || cfg.User != "alice" || cfg.Password != "secret" || cfg.DBName != "shop" || cfg.SSLMode != "require" {
		t.Fatalf("unexpected config: %+v", cfg)
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

func TestConnectWithDSN_InvalidDSN_ReturnsError(t *testing.T) {
	// This should fail without needing a running database.
	if _, err := ConnectWithDSN("this is not a valid postgres dsn"); err == nil {
		t.Fatalf("expected error")
	}
}

func TestConnect_InvalidPortInDSN_ReturnsError(t *testing.T) {
	// Use an invalid port string so the driver fails fast during DSN parsing.
	cfg := Config{
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
