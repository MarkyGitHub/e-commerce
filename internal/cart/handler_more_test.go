package cart

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/mschl/ecommerce-api/internal/shared"
)

func TestCartHandler_Clear_ServiceError(t *testing.T) {
	userID := uuid.New()
	mockService := &MockCartService{clearCartFunc: func(ctx context.Context, uid uuid.UUID) error {
		return errors.New("boom")
	}}
	h := NewHandler(mockService)

	req := httptest.NewRequest(http.MethodDelete, "/cart", nil)
	req = createRequestWithUser(req, userID, "customer")
	rr := httptest.NewRecorder()

	h.Clear(rr, req)
	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected %d, got %d", http.StatusInternalServerError, rr.Code)
	}
}

func TestCartHandler_RemoveItem_Forbidden(t *testing.T) {
	userID := uuid.New()
	itemID := uuid.New()

	mockService := &MockCartService{removeItemFunc: func(ctx context.Context, uid uuid.UUID, iid uuid.UUID) error {
		return shared.ErrForbidden
	}}
	h := NewHandler(mockService)

	req := httptest.NewRequest(http.MethodDelete, "/cart/items/"+itemID.String(), nil)
	req.SetPathValue("item_id", itemID.String())
	req = createRequestWithUser(req, userID, "customer")
	rr := httptest.NewRecorder()

	h.RemoveItem(rr, req)
	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected %d, got %d", http.StatusForbidden, rr.Code)
	}
}

func TestCartHandler_RemoveItem_UnknownError(t *testing.T) {
	userID := uuid.New()
	itemID := uuid.New()

	mockService := &MockCartService{removeItemFunc: func(ctx context.Context, uid uuid.UUID, iid uuid.UUID) error {
		return errors.New("boom")
	}}
	h := NewHandler(mockService)

	req := httptest.NewRequest(http.MethodDelete, "/cart/items/"+itemID.String(), nil)
	req.SetPathValue("item_id", itemID.String())
	req = createRequestWithUser(req, userID, "customer")
	rr := httptest.NewRecorder()

	h.RemoveItem(rr, req)
	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected %d, got %d", http.StatusInternalServerError, rr.Code)
	}
}

func TestCartHandler_parseUUID_MissingID(t *testing.T) {
	h := NewHandler(&MockCartService{})
	req := httptest.NewRequest(http.MethodGet, "/cart/items", nil)

	_, err := h.parseUUID(req, "item_id")
	if err == nil {
		t.Fatalf("expected error")
	}
}
