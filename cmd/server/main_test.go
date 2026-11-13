package main

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/MarkyGitHub/e-commerce/internal/config"
	"github.com/MarkyGitHub/e-commerce/internal/db"
	"github.com/MarkyGitHub/e-commerce/internal/testutil"
	"gorm.io/gorm"
)

func TestHealthCheck_OK(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()

	healthCheck(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
	}
	if rr.Body.String() != `{"status":"healthy"}` {
		t.Fatalf("unexpected body: %q", rr.Body.String())
	}
}

func mockConfig(port string) *config.Config {
	return &config.Config{
		AppEnv: "test",
		Server: config.ServerConfig{Port: port},
		DB:     config.DBConfig{},
	}
}

func TestRun_DefaultPort(t *testing.T) {
	testDB := testutil.SetupTestDB(t)
	if err := db.Migrate(testDB); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	var gotAddr string
	deps := serverDeps{
		loadConfig: func(string, string) (*config.Config, error) {
			return mockConfig("8080"), nil
		},
		connect: func(config.DBConfig) (*gorm.DB, error) {
			return testDB, nil
		},
		migrate: func(*gorm.DB) error { return nil },
		listen: func(addr string, h http.Handler) error {
			gotAddr = addr
			if h == nil {
				return errors.New("nil handler")
			}
			return nil
		},
	}

	if err := run(deps); err != nil {
		t.Fatalf("expected run() to succeed, got %v", err)
	}
	if gotAddr != ":8080" {
		t.Fatalf("expected addr :8080, got %q", gotAddr)
	}
}

func TestRun_CustomPort(t *testing.T) {
	testDB := testutil.SetupTestDB(t)

	var gotAddr string
	deps := serverDeps{
		loadConfig: func(string, string) (*config.Config, error) {
			return mockConfig("9090"), nil
		},
		connect: func(config.DBConfig) (*gorm.DB, error) {
			return testDB, nil
		},
		migrate: func(*gorm.DB) error { return nil },
		listen: func(addr string, h http.Handler) error {
			gotAddr = addr
			return nil
		},
	}

	if err := run(deps); err != nil {
		t.Fatalf("expected run() to succeed, got %v", err)
	}
	if gotAddr != ":9090" {
		t.Fatalf("expected addr :9090, got %q", gotAddr)
	}
}

func TestRun_ConfigError(t *testing.T) {
	deps := serverDeps{
		loadConfig: func(string, string) (*config.Config, error) {
			return nil, errors.New("config failed")
		},
		connect: func(config.DBConfig) (*gorm.DB, error) {
			return nil, nil
		},
		migrate: func(*gorm.DB) error { return nil },
		listen:  func(string, http.Handler) error { return nil },
	}

	if err := run(deps); err == nil {
		t.Fatalf("expected error")
	}
}

func TestRun_PropagatesErrors(t *testing.T) {
	deps := serverDeps{
		loadConfig: func(string, string) (*config.Config, error) {
			return mockConfig("8080"), nil
		},
		connect: func(config.DBConfig) (*gorm.DB, error) {
			return nil, errors.New("connect failed")
		},
		migrate: func(*gorm.DB) error { return nil },
		listen:  func(string, http.Handler) error { return nil },
	}

	if err := run(deps); err == nil {
		t.Fatalf("expected error")
	}
}

func TestRun_MigrateError(t *testing.T) {
	deps := serverDeps{
		loadConfig: func(string, string) (*config.Config, error) {
			return mockConfig("8080"), nil
		},
		connect: func(config.DBConfig) (*gorm.DB, error) {
			return &gorm.DB{}, nil
		},
		migrate: func(*gorm.DB) error { return errors.New("migrate failed") },
		listen:  func(string, http.Handler) error { return nil },
	}

	if err := run(deps); err == nil {
		t.Fatalf("expected error")
	}
}

func TestRun_ListenErrServerClosed_IsNil(t *testing.T) {
	deps := serverDeps{
		loadConfig: func(string, string) (*config.Config, error) {
			return mockConfig("8080"), nil
		},
		connect: func(config.DBConfig) (*gorm.DB, error) {
			return &gorm.DB{}, nil
		},
		migrate: func(*gorm.DB) error { return nil },
		listen:  func(string, http.Handler) error { return http.ErrServerClosed },
	}

	if err := run(deps); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestRun_ListenError(t *testing.T) {
	deps := serverDeps{
		loadConfig: func(string, string) (*config.Config, error) {
			return mockConfig("8080"), nil
		},
		connect: func(config.DBConfig) (*gorm.DB, error) {
			return &gorm.DB{}, nil
		},
		migrate: func(*gorm.DB) error { return nil },
		listen:  func(string, http.Handler) error { return errors.New("listen failed") },
	}

	if err := run(deps); err == nil {
		t.Fatalf("expected error")
	}
}

func TestMain_CallsRunMain(t *testing.T) {
	old := runMain
	t.Cleanup(func() { runMain = old })

	called := false
	runMain = func(deps serverDeps) error {
		called = true
		return nil
	}

	main()
	if !called {
		t.Fatalf("expected main() to call runMain")
	}
}
