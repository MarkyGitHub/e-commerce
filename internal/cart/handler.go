package cart

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"

	"github.com/MarkyGitHub/e-commerce/internal/middleware"
	"github.com/MarkyGitHub/e-commerce/internal/shared"
)

// Handler verarbeitet Warenkorb-HTTP-Anfragen
type Handler struct {
	service CartService
}

// NewHandler erstellt einen neuen Warenkorb-Handler
func NewHandler(service CartService) *Handler {
	return &Handler{service: service}
}

// Get handles GET /cart
// @Summary Get cart
// @Description Get the current authenticated user's shopping cart with all items. If cart is empty, returns empty items array.
// @Tags cart
// @Produce json
// @Security SessionToken
// @Param X-Session-Token header string true "Session token"
// @Success 200 {object} CartResponse "Cart details with items"
// @Failure 401 {object} shared.ErrorResponse "Missing or invalid session token"
// @Failure 500 {object} shared.ErrorResponse "Internal server error"
// @Router /cart [get]
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		shared.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	cart, items, err := h.service.GetCart(r.Context(), user.ID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	shared.WriteJSON(w, http.StatusOK, ToCartResponse(cart, items))
}

// AddItem handles POST /cart/items
// @Summary Add item to cart
// @Description Add a product to the cart. If the product already exists in the cart, the quantity will be updated. Quantity must be positive and inventory must be sufficient.
// @Tags cart
// @Accept json
// @Produce json
// @Security SessionToken
// @Param X-Session-Token header string true "Session token"
// @Param request body AddItemRequest true "Item details"
// @Success 201 {object} CartItemResponse "Item added to cart successfully"
// @Failure 400 {object} shared.ErrorResponse "Invalid request body, validation error, insufficient inventory, or invalid quantity"
// @Failure 401 {object} shared.ErrorResponse "Missing or invalid session token"
// @Failure 404 {object} shared.ErrorResponse "Product not found"
// @Failure 500 {object} shared.ErrorResponse "Internal server error"
// @Router /cart/items [post]
func (h *Handler) AddItem(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		shared.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req AddItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		shared.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Quantity <= 0 {
		shared.WriteError(w, http.StatusBadRequest, "quantity must be positive")
		return
	}

	item, err := h.service.AddItem(r.Context(), user.ID, req)
	if err != nil {
		h.handleError(w, err)
		return
	}

	shared.WriteJSON(w, http.StatusCreated, item.ToResponse())
}

// UpdateItem handles PUT /cart/items/{item_id}
// @Summary Update cart item
// @Description Update the quantity of a cart item. Quantity must be positive and inventory must be sufficient. Users can only update their own cart items.
// @Tags cart
// @Accept json
// @Produce json
// @Security SessionToken
// @Param X-Session-Token header string true "Session token"
// @Param item_id path string true "Cart item ID" format(uuid)
// @Param request body UpdateItemRequest true "Update details"
// @Success 200 {object} CartItemResponse "Cart item updated successfully"
// @Failure 400 {object} shared.ErrorResponse "Invalid request body, validation error, insufficient inventory, or invalid quantity"
// @Failure 401 {object} shared.ErrorResponse "Missing or invalid session token"
// @Failure 403 {object} shared.ErrorResponse "Access denied (not your cart item)"
// @Failure 404 {object} shared.ErrorResponse "Cart item not found"
// @Failure 500 {object} shared.ErrorResponse "Internal server error"
// @Router /cart/items/{item_id} [put]
func (h *Handler) UpdateItem(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		shared.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	itemID, err := h.parseUUID(r, "item_id")
	if err != nil {
		shared.WriteError(w, http.StatusBadRequest, "invalid item ID")
		return
	}

	var req UpdateItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		shared.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Quantity <= 0 {
		shared.WriteError(w, http.StatusBadRequest, "quantity must be positive")
		return
	}

	item, err := h.service.UpdateItem(r.Context(), user.ID, itemID, req.Quantity)
	if err != nil {
		h.handleError(w, err)
		return
	}

	shared.WriteJSON(w, http.StatusOK, item.ToResponse())
}

// RemoveItem handles DELETE /cart/items/{item_id}
// @Summary Remove cart item
// @Description Remove an item from the cart. Users can only remove items from their own cart.
// @Tags cart
// @Produce json
// @Security SessionToken
// @Param X-Session-Token header string true "Session token"
// @Param item_id path string true "Cart item ID" format(uuid)
// @Success 200 {object} map[string]string "Item removed successfully"
// @Failure 400 {object} shared.ErrorResponse "Invalid item ID format"
// @Failure 401 {object} shared.ErrorResponse "Missing or invalid session token"
// @Failure 403 {object} shared.ErrorResponse "Access denied (not your cart item)"
// @Failure 404 {object} shared.ErrorResponse "Cart item not found"
// @Failure 500 {object} shared.ErrorResponse "Internal server error"
// @Router /cart/items/{item_id} [delete]
func (h *Handler) RemoveItem(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		shared.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	itemID, err := h.parseUUID(r, "item_id")
	if err != nil {
		shared.WriteError(w, http.StatusBadRequest, "invalid item ID")
		return
	}

	if err := h.service.RemoveItem(r.Context(), user.ID, itemID); err != nil {
		h.handleError(w, err)
		return
	}

	shared.WriteJSON(w, http.StatusOK, map[string]string{"message": "item removed"})
}

// Clear handles DELETE /cart
// @Summary Clear cart
// @Description Remove all items from the current user's cart. This action cannot be undone.
// @Tags cart
// @Produce json
// @Security SessionToken
// @Param X-Session-Token header string true "Session token"
// @Success 200 {object} map[string]string "Cart cleared successfully"
// @Failure 401 {object} shared.ErrorResponse "Missing or invalid session token"
// @Failure 500 {object} shared.ErrorResponse "Internal server error"
// @Router /cart [delete]
func (h *Handler) Clear(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		shared.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	if err := h.service.ClearCart(r.Context(), user.ID); err != nil {
		h.handleError(w, err)
		return
	}

	shared.WriteJSON(w, http.StatusOK, map[string]string{"message": "cart cleared"})
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
		shared.WriteError(w, http.StatusNotFound, "resource not found")
	case errors.Is(err, shared.ErrForbidden):
		shared.WriteError(w, http.StatusForbidden, "access denied")
	case errors.Is(err, shared.ErrInsufficientInventory):
		shared.WriteError(w, http.StatusBadRequest, "insufficient inventory")
	default:
		shared.WriteError(w, http.StatusInternalServerError, "internal server error")
	}
}
