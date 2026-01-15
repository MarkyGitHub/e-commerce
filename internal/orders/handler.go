package orders

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"

	"github.com/mschl/ecommerce-api/internal/middleware"
	"github.com/mschl/ecommerce-api/internal/shared"
)

// Handler verarbeitet Bestell-HTTP-Anfragen
type Handler struct {
	service OrderService
}

// NewHandler erstellt einen neuen Bestell-Handler
func NewHandler(service OrderService) *Handler {
	return &Handler{service: service}
}

// Create handles POST /orders
// @Summary Create order
// @Description Create an order from the current cart. Cart must not be empty and inventory must be sufficient. After successful order creation, the cart is cleared and a notification is sent.
// @Tags orders
// @Accept json
// @Produce json
// @Security SessionToken
// @Param X-Session-Token header string true "Session token"
// @Param request body CreateOrderRequest true "Order details with shipping address"
// @Success 201 {object} OrderResponse "Order created successfully"
// @Failure 400 {object} shared.ErrorResponse "Cart is empty, insufficient inventory, or invalid shipping address"
// @Failure 401 {object} shared.ErrorResponse "Missing or invalid session token"
// @Failure 500 {object} shared.ErrorResponse "Internal server error"
// @Router /orders [post]
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		shared.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		shared.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	order, items, err := h.service.CreateOrder(r.Context(), user.ID, req)
	if err != nil {
		h.handleError(w, err)
		return
	}

	shared.WriteJSON(w, http.StatusCreated, ToOrderResponseWithItems(order, items))
}

// List handles GET /orders
// @Summary List orders
// @Description Get a paginated list of orders. Customers see only their own orders, admins see all orders.
// @Tags orders
// @Produce json
// @Security SessionToken
// @Param X-Session-Token header string true "Session token"
// @Param page query int false "Page number" default(1) minimum(1)
// @Param limit query int false "Items per page" default(10) minimum(1) maximum(100)
// @Success 200 {object} shared.PaginatedResponse{data=[]OrderResponse} "List of orders"
// @Failure 401 {object} shared.ErrorResponse "Missing or invalid session token"
// @Failure 500 {object} shared.ErrorResponse "Internal server error"
// @Router /orders [get]
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		shared.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	page, limit := shared.ParsePagination(r)

	orders, totalCount, err := h.service.ListOrders(r.Context(), user.ID, user.IsAdmin(), page, limit)
	if err != nil {
		h.handleError(w, err)
		return
	}

	shared.WritePaginated(w, ToOrderResponseList(orders), totalCount)
}

// Get handles GET /orders/{id}
// @Summary Get order
// @Description Get order details by ID including all order items. Customers can only view their own orders, admins can view any order.
// @Tags orders
// @Produce json
// @Security SessionToken
// @Param X-Session-Token header string true "Session token"
// @Param id path string true "Order ID" format(uuid)
// @Success 200 {object} OrderResponse "Order details with items"
// @Failure 400 {object} shared.ErrorResponse "Invalid order ID format"
// @Failure 401 {object} shared.ErrorResponse "Missing or invalid session token"
// @Failure 403 {object} shared.ErrorResponse "Access denied (not your order)"
// @Failure 404 {object} shared.ErrorResponse "Order not found"
// @Failure 500 {object} shared.ErrorResponse "Internal server error"
// @Router /orders/{id} [get]
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		shared.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	orderID, err := h.parseUUID(r, "id")
	if err != nil {
		shared.WriteError(w, http.StatusBadRequest, "invalid order ID")
		return
	}

	order, items, err := h.service.GetOrder(r.Context(), user.ID, orderID, user.IsAdmin())
	if err != nil {
		h.handleError(w, err)
		return
	}

	shared.WriteJSON(w, http.StatusOK, ToOrderResponseWithItems(order, items))
}

// parseUUID extrahiert eine UUID aus dem URL-Pfad
func (h *Handler) parseUUID(r *http.Request, key string) (uuid.UUID, error) {
	idStr := r.PathValue(key)
	if idStr == "" {
		return uuid.Nil, errors.New("missing id")
	}
	return uuid.Parse(idStr)
}

// handleError konvertiert Service-Fehler zu HTTP-Antworten
func (h *Handler) handleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, shared.ErrNotFound):
		shared.WriteError(w, http.StatusNotFound, "order not found")
	case errors.Is(err, shared.ErrForbidden):
		shared.WriteError(w, http.StatusForbidden, "access denied")
	case errors.Is(err, shared.ErrEmptyCart):
		shared.WriteError(w, http.StatusBadRequest, "cart is empty")
	case errors.Is(err, shared.ErrInsufficientInventory):
		shared.WriteError(w, http.StatusBadRequest, "insufficient inventory")
	default:
		shared.WriteError(w, http.StatusInternalServerError, "internal server error")
	}
}
