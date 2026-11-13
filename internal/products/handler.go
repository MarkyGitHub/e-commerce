package products

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"

	"github.com/MarkyGitHub/e-commerce/internal/shared"
)

// Handler verarbeitet Produkt-HTTP-Anfragen
type Handler struct {
	service ProductService
}

// NewHandler erstellt einen neuen Produkt-Handler
func NewHandler(service ProductService) *Handler {
	return &Handler{service: service}
}

// List handles GET /products
// @Summary List products
// @Description Get a paginated list of all products. Public endpoint, no authentication required. Optionally filter by category_id.
// @Tags products
// @Produce json
// @Param category_id query string false "Category ID to filter by" format(uuid)
// @Param page query int false "Page number" default(1) minimum(1)
// @Param limit query int false "Items per page" default(10) minimum(1) maximum(100)
// @Success 200 {object} shared.PaginatedResponse{items=[]Product} "List of products"
// @Failure 500 {object} shared.ErrorResponse "Internal server error"
// @Router /products [get]
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	page, limit := shared.ParsePagination(r)

	// Optionalen Kategoriefilter parsen
	var categoryID *uuid.UUID
	if categoryIDStr := r.URL.Query().Get("category_id"); categoryIDStr != "" {
		id, err := uuid.Parse(categoryIDStr)
		if err == nil {
			categoryID = &id
		}
	}

	products, totalCount, err := h.service.ListProducts(r.Context(), page, limit, categoryID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	shared.WritePaginated(w, products, totalCount)
}

// Get handles GET /products/{id}
// @Summary Get product
// @Description Get a product by ID. Public endpoint, no authentication required.
// @Tags products
// @Produce json
// @Param id path string true "Product ID" format(uuid)
// @Success 200 {object} Product "Product details"
// @Failure 400 {object} shared.ErrorResponse "Invalid product ID format"
// @Failure 404 {object} shared.ErrorResponse "Product not found"
// @Failure 500 {object} shared.ErrorResponse "Internal server error"
// @Router /products/{id} [get]
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := h.parseUUID(r, "id")
	if err != nil {
		shared.WriteError(w, http.StatusBadRequest, "invalid product ID")
		return
	}

	product, err := h.service.GetProduct(r.Context(), id)
	if err != nil {
		h.handleError(w, err)
		return
	}

	shared.WriteJSON(w, http.StatusOK, product)
}

// Create handles POST /products
// @Summary Create product
// @Description Create a new product. Requires admin role. Product name and price are required. Price must be non-negative.
// @Tags products
// @Accept json
// @Produce json
// @Security SessionToken
// @Param X-Session-Token header string true "Session token"
// @Param request body CreateProductRequest true "Product details"
// @Success 201 {object} Product "Product created successfully"
// @Failure 400 {object} shared.ErrorResponse "Invalid request body or validation error"
// @Failure 401 {object} shared.ErrorResponse "Missing or invalid session token"
// @Failure 403 {object} shared.ErrorResponse "Insufficient permissions (admin required)"
// @Failure 500 {object} shared.ErrorResponse "Internal server error"
// @Router /products [post]
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userVal := r.Context().Value(shared.UserContextKey)
	if userVal == nil {
		shared.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	// Interface nutzen um User ID zu bekommen - funktioniert mit auth.User und Test-Mocks
	user, ok := userVal.(shared.UserIDGetter)
	if !ok {
		shared.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req CreateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		shared.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	product, err := h.service.CreateProduct(r.Context(), req, user.GetID())
	if err != nil {
		h.handleError(w, err)
		return
	}

	shared.WriteJSON(w, http.StatusCreated, product)
}

// Update handles PUT /products/{id}
// @Summary Update product
// @Description Update a product. Requires admin role. Only provided fields will be updated.
// @Tags products
// @Accept json
// @Produce json
// @Security SessionToken
// @Param X-Session-Token header string true "Session token"
// @Param id path string true "Product ID" format(uuid)
// @Param request body UpdateProductRequest true "Product update details"
// @Success 200 {object} Product "Product updated successfully"
// @Failure 400 {object} shared.ErrorResponse "Invalid request body or validation error"
// @Failure 401 {object} shared.ErrorResponse "Missing or invalid session token"
// @Failure 403 {object} shared.ErrorResponse "Insufficient permissions (admin required)"
// @Failure 404 {object} shared.ErrorResponse "Product not found"
// @Failure 500 {object} shared.ErrorResponse "Internal server error"
// @Router /products/{id} [put]
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := h.parseUUID(r, "id")
	if err != nil {
		shared.WriteError(w, http.StatusBadRequest, "invalid product ID")
		return
	}

	var req UpdateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		shared.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	product, err := h.service.UpdateProduct(r.Context(), id, req)
	if err != nil {
		h.handleError(w, err)
		return
	}

	shared.WriteJSON(w, http.StatusOK, product)
}

// Delete handles DELETE /products/{id}
// @Summary Delete product
// @Description Delete a product. Requires admin role. This action cannot be undone.
// @Tags products
// @Produce json
// @Security SessionToken
// @Param X-Session-Token header string true "Session token"
// @Param id path string true "Product ID" format(uuid)
// @Success 200 {object} map[string]string "Product deleted successfully"
// @Failure 400 {object} shared.ErrorResponse "Invalid product ID format"
// @Failure 401 {object} shared.ErrorResponse "Missing or invalid session token"
// @Failure 403 {object} shared.ErrorResponse "Insufficient permissions (admin required)"
// @Failure 404 {object} shared.ErrorResponse "Product not found"
// @Failure 500 {object} shared.ErrorResponse "Internal server error"
// @Router /products/{id} [delete]
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := h.parseUUID(r, "id")
	if err != nil {
		shared.WriteError(w, http.StatusBadRequest, "invalid product ID")
		return
	}

	if err := h.service.DeleteProduct(r.Context(), id); err != nil {
		h.handleError(w, err)
		return
	}

	shared.WriteJSON(w, http.StatusOK, map[string]string{"message": "product deleted"})
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
	var validationErr *ValidationError
	if errors.As(err, &validationErr) {
		shared.WriteError(w, http.StatusBadRequest, validationErr.Message)
		return
	}

	switch {
	case errors.Is(err, shared.ErrNotFound):
		shared.WriteError(w, http.StatusNotFound, "product not found")
	case errors.Is(err, shared.ErrBadRequest):
		shared.WriteError(w, http.StatusBadRequest, err.Error())
	default:
		shared.WriteError(w, http.StatusInternalServerError, "internal server error")
	}
}
