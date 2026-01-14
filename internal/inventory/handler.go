package inventory

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"

	"github.com/mschl/ecommerce-api/internal/shared"
)

// Handler handles inventory HTTP requests
type Handler struct {
	service InventoryService
}

// NewHandler creates a new inventory handler
func NewHandler(service InventoryService) *Handler {
	return &Handler{service: service}
}

// Get handles GET /inventory/{product_id}
// @Summary Get inventory
// @Description Get inventory quantity for a product. Public endpoint, no authentication required.
// @Tags inventory
// @Produce json
// @Param product_id path string true "Product ID" format(uuid)
// @Success 200 {object} InventoryResponse "Inventory details"
// @Failure 400 {object} shared.ErrorResponse "Invalid product ID format"
// @Failure 404 {object} shared.ErrorResponse "Inventory not found for product"
// @Failure 500 {object} shared.ErrorResponse "Internal server error"
// @Router /inventory/{product_id} [get]
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	productID, err := h.parseUUID(r, "product_id")
	if err != nil {
		shared.WriteError(w, http.StatusBadRequest, "invalid product ID")
		return
	}

	inventory, err := h.service.GetInventory(r.Context(), productID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	shared.WriteJSON(w, http.StatusOK, inventory.ToResponse())
}

// Update handles PATCH /inventory/{product_id}
// @Summary Increment inventory
// @Description Increment inventory quantity for a product. Requires admin role. Increments the current quantity by the specified amount.
// @Tags inventory
// @Accept json
// @Produce json
// @Security SessionToken
// @Param X-Session-Token header string true "Session token"
// @Param product_id path string true "Product ID" format(uuid)
// @Param request body UpdateInventoryRequest true "Inventory increment details"
// @Success 200 {object} InventoryResponse "Inventory incremented successfully"
// @Failure 400 {object} shared.ErrorResponse "Invalid request body or validation error"
// @Failure 401 {object} shared.ErrorResponse "Missing or invalid session token"
// @Failure 403 {object} shared.ErrorResponse "Insufficient permissions (admin required)"
// @Failure 404 {object} shared.ErrorResponse "Inventory not found for product"
// @Failure 500 {object} shared.ErrorResponse "Internal server error"
// @Router /inventory/{product_id} [patch]
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	userVal := r.Context().Value(shared.UserContextKey)
	if userVal == nil {
		shared.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	// Use type assertion to get user ID without importing auth package
	type userWithID interface {
		GetID() uuid.UUID
	}
	user, ok := userVal.(userWithID)
	if !ok {
		shared.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	productID, err := h.parseUUID(r, "product_id")
	if err != nil {
		shared.WriteError(w, http.StatusBadRequest, "invalid product ID")
		return
	}

	var req UpdateInventoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		shared.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	inventory, err := h.service.IncrementInventory(r.Context(), productID, req.IncrementBy, user.GetID())
	if err != nil {
		h.handleError(w, err)
		return
	}

	shared.WriteJSON(w, http.StatusOK, inventory.ToResponse())
}

// parseUUID extracts a UUID from the URL path
func (h *Handler) parseUUID(r *http.Request, key string) (uuid.UUID, error) {
	idStr := r.PathValue(key)
	if idStr == "" {
		return uuid.Nil, errors.New("missing id")
	}
	return uuid.Parse(idStr)
}

// handleError converts service errors to HTTP responses
func (h *Handler) handleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, shared.ErrNotFound):
		shared.WriteError(w, http.StatusNotFound, "inventory not found")
	case errors.Is(err, shared.ErrNegativeQuantity):
		shared.WriteError(w, http.StatusBadRequest, "quantity cannot be negative")
	default:
		shared.WriteError(w, http.StatusInternalServerError, "internal server error")
	}
}
