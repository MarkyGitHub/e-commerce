package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/MarkyGitHub/e-commerce/internal/config"
	"github.com/MarkyGitHub/e-commerce/internal/db"
	"github.com/MarkyGitHub/e-commerce/internal/server"
	"gorm.io/gorm"

	_ "github.com/MarkyGitHub/e-commerce/docs/swagger" // Generierte Docs importieren
)

// @title E-Commerce REST API
// @version 1.0
// @description A complete E-Commerce REST API built in Go following layered architecture (Handler → Service → Repository)
//
// @contact.name API Support
// @contact.email support@ecommerce.local
//
// @host localhost:8080
// @BasePath /
// @schemes http
//
// @securityDefinitions.apikey SessionToken
// @in header
// @name X-Session-Token
// @description Session token for authentication. Obtain by logging in via /auth/login endpoint.

// healthCheck handles GET /health
// @Summary Health check
// @Description Check if the API server is running and healthy
// @Tags utility
// @Produce json
// @Success 200 {object} map[string]string "Server is healthy"
// @Router /health [get]
func healthCheck(w http.ResponseWriter, r *http.Request) {
	server.HealthCheck(w, r)
}

type serverDeps struct {
	loadConfig func(configDir string, defaultEnv string) (*config.Config, error)
	connect    func(config.DBConfig) (*gorm.DB, error)
	migrate    func(*gorm.DB) error
	listen     func(string, http.Handler) error
}

var runMain = run

func run(deps serverDeps) error {
	cfg, err := deps.loadConfig("configs", "local")
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	database, err := deps.connect(cfg.DB)
	if err != nil {
		return fmt.Errorf("connect db: %w", err)
	}

	if err := deps.migrate(database); err != nil {
		return fmt.Errorf("run migrations: %w", err)
	}

	mux := server.BuildMux(database)

	addr := fmt.Sprintf(":%s", cfg.Server.Port)
	log.Printf("Server starting on %s (env: %s)", addr, cfg.AppEnv)

	if err := deps.listen(addr, mux); err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return fmt.Errorf("listen and serve: %w", err)
	}

	return nil
}

func defaultDeps() serverDeps {
	return serverDeps{
		loadConfig: config.LoadFromAppEnv,
		connect:    db.Connect,
		migrate:    db.Migrate,
		listen:     http.ListenAndServe,
	}
}

func main() {
	deps := defaultDeps()

	log.Println("Starting server")
	if err := runMain(deps); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
