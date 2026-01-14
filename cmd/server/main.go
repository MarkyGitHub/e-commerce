package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/mschl/ecommerce-api/internal/auth"
	"github.com/mschl/ecommerce-api/internal/cart"
	"github.com/mschl/ecommerce-api/internal/config"
	"github.com/mschl/ecommerce-api/internal/db"
	"github.com/mschl/ecommerce-api/internal/inventory"
	"github.com/mschl/ecommerce-api/internal/middleware"
	"github.com/mschl/ecommerce-api/internal/notifications"
	"github.com/mschl/ecommerce-api/internal/orders"
	"github.com/mschl/ecommerce-api/internal/products"
	"gorm.io/gorm"

	_ "github.com/mschl/ecommerce-api/docs/swagger" // Import generated docs
	httpSwagger "github.com/swaggo/http-swagger"
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
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"healthy"}`))
}

type serverDeps struct {
	loadConfig func(configDir string, defaultEnv string) (*config.Config, error)
	connect    func(config.DBConfig) (*gorm.DB, error)
	migrate    func(*gorm.DB) error
	listen     func(string, http.Handler) error
}

var runMain = run

func buildMux(database *gorm.DB) *http.ServeMux {
	// Initialize repositories
	userRepo := auth.NewUserRepository(database)
	sessionRepo := auth.NewSessionRepository(database)
	productRepo := products.NewProductRepository(database)
	inventoryRepo := inventory.NewInventoryRepository(database)
	cartRepo := cart.NewCartRepository(database)
	cartItemRepo := cart.NewCartItemRepository(database)
	orderRepo := orders.NewOrderRepository(database)
	orderItemRepo := orders.NewOrderItemRepository(database)
	notificationRepo := notifications.NewNotificationRepository(database)

	// Initialize services
	notificationService := notifications.NewNotificationService(notificationRepo)
	authService := auth.NewAuthService(userRepo, sessionRepo, notificationService)
	productService := products.NewProductService(productRepo, notificationService)
	inventoryService := inventory.NewInventoryService(inventoryRepo, notificationService)
	cartService := cart.NewCartService(cartRepo, cartItemRepo, inventoryRepo, productRepo, notificationService)
	orderService := orders.NewOrderService(
		database,
		orderRepo,
		orderItemRepo,
		cartRepo,
		cartItemRepo,
		productRepo,
		inventoryRepo,
		notificationService,
	)

	// Initialize handlers
	authHandler := auth.NewHandler(authService)
	productHandler := products.NewHandler(productService)
	inventoryHandler := inventory.NewHandler(inventoryService)
	cartHandler := cart.NewHandler(cartService)
	orderHandler := orders.NewHandler(orderService)
	notificationHandler := notifications.NewHandler(notificationService)

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(authService)

	// Create router
	mux := http.NewServeMux()

	// Auth routes (public)
	mux.HandleFunc("POST /auth/register", authHandler.Register)
	mux.HandleFunc("POST /auth/login", authHandler.Login)
	mux.Handle("POST /auth/logout", authMiddleware.Authenticate(http.HandlerFunc(authHandler.Logout)))

	// Product routes (public read, admin write)
	mux.HandleFunc("GET /products", productHandler.List)
	mux.HandleFunc("GET /products/{id}", productHandler.Get)
	mux.Handle("POST /products", authMiddleware.Authenticate(
		middleware.RequireAdmin(http.HandlerFunc(productHandler.Create))))
	mux.Handle("PUT /products/{id}", authMiddleware.Authenticate(
		middleware.RequireAdmin(http.HandlerFunc(productHandler.Update))))
	mux.Handle("DELETE /products/{id}", authMiddleware.Authenticate(
		middleware.RequireAdmin(http.HandlerFunc(productHandler.Delete))))

	// Inventory routes (authenticated read, admin write)
	mux.Handle("GET /inventory/{product_id}", authMiddleware.Authenticate(http.HandlerFunc(inventoryHandler.Get)))
	mux.Handle("PATCH /inventory/{product_id}", authMiddleware.Authenticate(
		middleware.RequireAdmin(http.HandlerFunc(inventoryHandler.Update))))

	// Cart routes (authenticated)
	mux.Handle("GET /cart", authMiddleware.Authenticate(http.HandlerFunc(cartHandler.Get)))
	mux.Handle("POST /cart/items", authMiddleware.Authenticate(http.HandlerFunc(cartHandler.AddItem)))
	mux.Handle("PUT /cart/items/{item_id}", authMiddleware.Authenticate(http.HandlerFunc(cartHandler.UpdateItem)))
	mux.Handle("DELETE /cart/items/{item_id}", authMiddleware.Authenticate(http.HandlerFunc(cartHandler.RemoveItem)))
	mux.Handle("DELETE /cart", authMiddleware.Authenticate(http.HandlerFunc(cartHandler.Clear)))

	// Order routes (authenticated)
	mux.Handle("POST /orders", authMiddleware.Authenticate(http.HandlerFunc(orderHandler.Create)))
	mux.Handle("GET /orders", authMiddleware.Authenticate(http.HandlerFunc(orderHandler.List)))
	mux.Handle("GET /orders/{id}", authMiddleware.Authenticate(http.HandlerFunc(orderHandler.Get)))

	// Notification routes (authenticated)
	mux.Handle("GET /notifications", authMiddleware.Authenticate(http.HandlerFunc(notificationHandler.List)))
	mux.Handle("PATCH /notifications/{id}/read", authMiddleware.Authenticate(http.HandlerFunc(notificationHandler.MarkAsRead)))

	// Health check
	mux.HandleFunc("GET /health", healthCheck)

	// Swagger UI
	mux.Handle("/swagger/", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))

	return mux
}

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

	mux := buildMux(database)

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
