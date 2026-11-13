package server

import (
	"net/http"

	"github.com/mschl/ecommerce-api/internal/auth"
	"github.com/mschl/ecommerce-api/internal/cart"
	"github.com/mschl/ecommerce-api/internal/inventory"
	"github.com/mschl/ecommerce-api/internal/middleware"
	"github.com/mschl/ecommerce-api/internal/notifications"
	"github.com/mschl/ecommerce-api/internal/orders"
	"github.com/mschl/ecommerce-api/internal/products"
	"gorm.io/gorm"

	httpSwagger "github.com/swaggo/http-swagger"
)

// BuildMux baut den HTTP-Router für die Anwendung.
func BuildMux(database *gorm.DB) *http.ServeMux {
	// Repositories initialisieren
	userRepo := auth.NewUserRepository(database)
	sessionRepo := auth.NewSessionRepository(database)
	productRepo := products.NewProductRepository(database)
	inventoryRepo := inventory.NewInventoryRepository(database)
	cartRepo := cart.NewCartRepository(database)
	cartItemRepo := cart.NewCartItemRepository(database)
	orderRepo := orders.NewOrderRepository(database)
	orderItemRepo := orders.NewOrderItemRepository(database)
	notificationRepo := notifications.NewNotificationRepository(database)

	// Services initialisieren
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

	// Handler initialisieren
	authHandler := auth.NewHandler(authService)
	productHandler := products.NewHandler(productService)
	inventoryHandler := inventory.NewHandler(inventoryService)
	cartHandler := cart.NewHandler(cartService)
	orderHandler := orders.NewHandler(orderService)
	notificationHandler := notifications.NewHandler(notificationService)

	// Middleware initialisieren
	authMiddleware := middleware.NewAuthMiddleware(authService)

	// Router erstellen
	mux := http.NewServeMux()

	// Auth-Routen (öffentlich)
	mux.HandleFunc("POST /auth/register", authHandler.Register)
	mux.HandleFunc("POST /auth/login", authHandler.Login)
	mux.Handle("POST /auth/logout", authMiddleware.Authenticate(http.HandlerFunc(authHandler.Logout)))

	// Produkt-Routen (öffentlich lesen, Admin schreiben)
	mux.HandleFunc("GET /products", productHandler.List)
	mux.HandleFunc("GET /products/{id}", productHandler.Get)
	mux.Handle("POST /products", authMiddleware.Authenticate(
		middleware.RequireAdmin(http.HandlerFunc(productHandler.Create))))
	mux.Handle("PUT /products/{id}", authMiddleware.Authenticate(
		middleware.RequireAdmin(http.HandlerFunc(productHandler.Update))))
	mux.Handle("DELETE /products/{id}", authMiddleware.Authenticate(
		middleware.RequireAdmin(http.HandlerFunc(productHandler.Delete))))

	// Bestand-Routen (authentifiziert lesen, Admin schreiben)
	mux.Handle("GET /inventory/{product_id}", authMiddleware.Authenticate(http.HandlerFunc(inventoryHandler.Get)))
	mux.Handle("PATCH /inventory/{product_id}", authMiddleware.Authenticate(
		middleware.RequireAdmin(http.HandlerFunc(inventoryHandler.Update))))

	// Warenkorb-Routen (authentifiziert)
	mux.Handle("GET /cart", authMiddleware.Authenticate(http.HandlerFunc(cartHandler.Get)))
	mux.Handle("POST /cart/items", authMiddleware.Authenticate(http.HandlerFunc(cartHandler.AddItem)))
	mux.Handle("PUT /cart/items/{item_id}", authMiddleware.Authenticate(http.HandlerFunc(cartHandler.UpdateItem)))
	mux.Handle("DELETE /cart/items/{item_id}", authMiddleware.Authenticate(http.HandlerFunc(cartHandler.RemoveItem)))
	mux.Handle("DELETE /cart", authMiddleware.Authenticate(http.HandlerFunc(cartHandler.Clear)))

	// Bestell-Routen (authentifiziert)
	mux.Handle("POST /orders", authMiddleware.Authenticate(http.HandlerFunc(orderHandler.Create)))
	mux.Handle("GET /orders", authMiddleware.Authenticate(http.HandlerFunc(orderHandler.List)))
	mux.Handle("GET /orders/{id}", authMiddleware.Authenticate(http.HandlerFunc(orderHandler.Get)))

	// Benachrichtigungs-Routen (authentifiziert)
	mux.Handle("GET /notifications", authMiddleware.Authenticate(http.HandlerFunc(notificationHandler.List)))
	mux.Handle("PATCH /notifications/{id}/read", authMiddleware.Authenticate(http.HandlerFunc(notificationHandler.MarkAsRead)))

	// Gesundheitsprüfung
	mux.HandleFunc("GET /health", HealthCheck)

	// Swagger UI
	mux.Handle("/swagger/", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))

	return mux
}

// HealthCheck handelt GET /health.
func HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"healthy"}`))
}
