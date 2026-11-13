package products

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/MarkyGitHub/e-commerce/internal/shared"
)

func TestProductHandler_List_WithValidCategoryID(t *testing.T) {
	expectedCategoryID := uuid.New()
	mockService := &MockProductService{listFunc: func(ctx context.Context, page, limit int, categoryID *uuid.UUID) ([]Product, int64, error) {
		if categoryID == nil || *categoryID != expectedCategoryID {
			t.Fatalf("expected categoryID %v, got %v", expectedCategoryID, categoryID)
		}
		return []Product{}, 0, nil
	}}

	h := NewHandler(mockService)
	req := httptest.NewRequest(http.MethodGet, "/products?category_id="+expectedCategoryID.String(), nil)
	rr := httptest.NewRecorder()

	h.List(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d", http.StatusOK, rr.Code)
	}
}

func TestProductHandler_List_ServiceError(t *testing.T) {
	mockService := &MockProductService{listFunc: func(ctx context.Context, page, limit int, categoryID *uuid.UUID) ([]Product, int64, error) {
		return nil, 0, errors.New("boom")
	}}

	h := NewHandler(mockService)
	req := httptest.NewRequest(http.MethodGet, "/products", nil)
	rr := httptest.NewRecorder()

	h.List(rr, req)
	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected %d, got %d", http.StatusInternalServerError, rr.Code)
	}
}

func TestProductHandler_Create_Unauthorized_NoUserInContext(t *testing.T) {
	mockService := &MockProductService{}
	h := NewHandler(mockService)

	body, _ := json.Marshal(CreateProductRequest{Name: "X", Price: 1})
	req := httptest.NewRequest(http.MethodPost, "/products", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.Create(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected %d, got %d", http.StatusUnauthorized, rr.Code)
	}
}

func TestProductHandler_Create_Unauthorized_BadUserType(t *testing.T) {
	mockService := &MockProductService{}
	h := NewHandler(mockService)

	body, _ := json.Marshal(CreateProductRequest{Name: "X", Price: 1})
	req := httptest.NewRequest(http.MethodPost, "/products", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(context.WithValue(req.Context(), shared.UserContextKey, struct{}{}))
	rr := httptest.NewRecorder()

	h.Create(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected %d, got %d", http.StatusUnauthorized, rr.Code)
	}
}

func TestProductHandler_Update_InvalidJSONBody(t *testing.T) {
	mockService := &MockProductService{}
	h := NewHandler(mockService)

	id := uuid.New()
	req := httptest.NewRequest(http.MethodPut, "/products/"+id.String(), bytes.NewReader([]byte("not-json")))
	req.SetPathValue("id", id.String())
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.Update(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestProductHandler_parseUUID_MissingID(t *testing.T) {
	h := NewHandler(&MockProductService{})
	req := httptest.NewRequest(http.MethodGet, "/products", nil)

	_, err := h.parseUUID(req, "id")
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestProductHandler_HandleError_BadRequestEchoesMessage(t *testing.T) {
	h := NewHandler(&MockProductService{})
	rr := httptest.NewRecorder()

	h.handleError(rr, fmt.Errorf("bad input: %w", shared.ErrBadRequest))

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected %d, got %d", http.StatusBadRequest, rr.Code)
	}
	var resp shared.ErrorResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp.Message != "bad input: bad request" {
		t.Fatalf("expected echoed message, got %q", resp.Message)
	}
}
