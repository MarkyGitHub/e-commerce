package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGetEnv_DefaultAndOverride(t *testing.T) {
	key := "CONFIG_TEST_KEY"
	os.Unsetenv(key)

	if got := getEnv(key, "default"); got != "default" {
		t.Fatalf("expected default value, got %q", got)
	}

	t.Setenv(key, "override")
	if got := getEnv(key, "default"); got != "override" {
		t.Fatalf("expected override value, got %q", got)
	}
}

func TestLoad_FromEnvVars(t *testing.T) {
	t.Setenv("APP_ENV", "test")
	t.Setenv("SERVER_PORT", "9000")
	t.Setenv("DB_HOST", "db.example")
	t.Setenv("DB_PORT", "5433")
	t.Setenv("DB_USER", "alice")
	t.Setenv("DB_PASSWORD", "secret")
	t.Setenv("DB_NAME", "shop")
	t.Setenv("DB_SSLMODE", "require")

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.AppEnv != "test" {
		t.Errorf("expected AppEnv=test, got %q", cfg.AppEnv)
	}
	if cfg.Server.Port != "9000" {
		t.Errorf("expected Server.Port=9000, got %q", cfg.Server.Port)
	}
	if cfg.DB.Host != "db.example" {
		t.Errorf("expected DB.Host=db.example, got %q", cfg.DB.Host)
	}
	if cfg.DB.Port != "5433" {
		t.Errorf("expected DB.Port=5433, got %q", cfg.DB.Port)
	}
	if cfg.DB.User != "alice" {
		t.Errorf("expected DB.User=alice, got %q", cfg.DB.User)
	}
	if cfg.DB.Password != "secret" {
		t.Errorf("expected DB.Password=secret, got %q", cfg.DB.Password)
	}
	if cfg.DB.DBName != "shop" {
		t.Errorf("expected DB.DBName=shop, got %q", cfg.DB.DBName)
	}
	if cfg.DB.SSLMode != "require" {
		t.Errorf("expected DB.SSLMode=require, got %q", cfg.DB.SSLMode)
	}
}

func TestLoad_FromEnvFile(t *testing.T) {
	// Create a temporary .env file
	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, "test.env")
	content := `APP_ENV=fromfile
SERVER_PORT=7777
DB_HOST=filehost
DB_PORT=5555
DB_USER=fileuser
DB_PASSWORD=filepass
DB_NAME=filedb
DB_SSLMODE=verify-full
`
	if err := os.WriteFile(envFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write env file: %v", err)
	}

	// Clear any existing env vars
	os.Unsetenv("APP_ENV")
	os.Unsetenv("SERVER_PORT")
	os.Unsetenv("DB_HOST")
	os.Unsetenv("DB_PORT")
	os.Unsetenv("DB_USER")
	os.Unsetenv("DB_PASSWORD")
	os.Unsetenv("DB_NAME")
	os.Unsetenv("DB_SSLMODE")

	cfg, err := Load(envFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.AppEnv != "fromfile" {
		t.Errorf("expected AppEnv=fromfile, got %q", cfg.AppEnv)
	}
	if cfg.Server.Port != "7777" {
		t.Errorf("expected Server.Port=7777, got %q", cfg.Server.Port)
	}
	if cfg.DB.Host != "filehost" {
		t.Errorf("expected DB.Host=filehost, got %q", cfg.DB.Host)
	}
}

func TestLoad_InvalidFile(t *testing.T) {
	_, err := Load("/nonexistent/path/to/file.env")
	if err == nil {
		t.Fatalf("expected error for invalid file path")
	}
}

func TestDBConfig_DSN(t *testing.T) {
	cfg := DBConfig{
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

func TestLoadFromAppEnv(t *testing.T) {
	// Create a temporary configs directory
	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, "test.env")
	content := `APP_ENV=test
SERVER_PORT=8888
`
	if err := os.WriteFile(envFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write env file: %v", err)
	}

	// Clear existing env vars
	os.Unsetenv("APP_ENV")
	os.Unsetenv("SERVER_PORT")

	cfg, err := LoadFromAppEnv(tmpDir, "test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Server.Port != "8888" {
		t.Errorf("expected Server.Port=8888, got %q", cfg.Server.Port)
	}
}

func TestLoadFromAppEnv_FileNotExist(t *testing.T) {
	// When file doesn't exist, should load from environment with defaults
	os.Unsetenv("APP_ENV")
	os.Unsetenv("SERVER_PORT")

	cfg, err := LoadFromAppEnv("/nonexistent/path", "local")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should use defaults
	if cfg.Server.Port != "8080" {
		t.Errorf("expected default Server.Port=8080, got %q", cfg.Server.Port)
	}
}

func TestGetSessionTTLMinutes(t *testing.T) {
	os.Unsetenv("SESSION_TTL_MINUTES")

	// Default value
	if got := GetSessionTTLMinutes(); got != 60 {
		t.Errorf("expected default 60, got %d", got)
	}

	// Custom value
	t.Setenv("SESSION_TTL_MINUTES", "120")
	if got := GetSessionTTLMinutes(); got != 120 {
		t.Errorf("expected 120, got %d", got)
	}

	// Invalid value falls back to default
	t.Setenv("SESSION_TTL_MINUTES", "invalid")
	if got := GetSessionTTLMinutes(); got != 60 {
		t.Errorf("expected fallback 60, got %d", got)
	}
}
