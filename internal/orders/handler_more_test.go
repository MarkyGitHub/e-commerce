package orders

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/mschl/ecommerce-api/internal/auth"
	"github.com/mschl/ecommerce-api/internal/shared"
)

func TestOrderHandler_Create_InvalidJSONBody(t *testing.T) {
	h := NewHandler(&MockOrderService{})

	req := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewBuffer([]byte("not-json")))
	req.Header.Set("Content-Type", "application/json")
	req = createRequestWithUser(req, uuid.New(), "customer")
	rr := httptest.NewRecorder()

	h.Create(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestOrderHandler_Create_ServiceError(t *testing.T) {
	mockService := &MockOrderService{createOrderFunc: func(ctx context.Context, userID uuid.UUID, shippingAddress CreateOrderRequest) (*Order, []OrderItem, error) {
		return nil, nil, errors.New("boom")
	}}
	h := NewHandler(mockService)

	req := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewBufferString(`{"shipping_street":"x","shipping_city":"y","shipping_zip":"z","shipping_country":"c"}`))
	req.Header.Set("Content-Type", "application/json")
	req = createRequestWithUser(req, uuid.New(), "customer")
	rr := httptest.NewRecorder()

	h.Create(rr, req)
	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected %d, got %d", http.StatusInternalServerError, rr.Code)
	}
}

func TestOrderHandler_List_ServiceError(t *testing.T) {
	mockService := &MockOrderService{listOrdersFunc: func(ctx context.Context, userID uuid.UUID, isAdmin bool, page, limit int) ([]Order, int64, error) {
		return nil, 0, errors.New("boom")
	}}
	h := NewHandler(mockService)

	req := httptest.NewRequest(http.MethodGet, "/orders?page=1&limit=10", nil)
	// Admin path and customer path both should surface error the same.
	user := &auth.User{ID: uuid.New(), Role: auth.RoleCustomer}
	req = req.WithContext(context.WithValue(req.Context(), shared.UserContextKey, user))
	rr := httptest.NewRecorder()

	h.List(rr, req)
	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected %d, got %d", http.StatusInternalServerError, rr.Code)
	}
}

func TestOrderHandler_parseUUID_MissingID(t *testing.T) {
	h := NewHandler(&MockOrderService{})
	req := httptest.NewRequest(http.MethodGet, "/orders", nil)

	_, err := h.parseUUID(req, "id")
	if err == nil {
		t.Fatalf("expected error")
	}
}
