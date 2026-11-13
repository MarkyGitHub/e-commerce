package inventory

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/MarkyGitHub/e-commerce/internal/shared"
)

// MockUser implements shared.UserIDGetter for testing
type MockUser struct {
	ID uuid.UUID
}

func (u *MockUser) GetID() uuid.UUID {
	return u.ID
}

// Helper to add user context to request
func requestWithUserContext(r *http.Request, userID uuid.UUID) *http.Request {
	user := &MockUser{ID: userID}
	ctx := context.WithValue(r.Context(), shared.UserContextKey, user)
	return r.WithContext(ctx)
}

// MockInventoryService is a mock implementation of InventoryService for handler testing
type MockInventoryService struct {
	getFunc            func(ctx context.Context, productID uuid.UUID) (*Inventory, error)
	updateFunc         func(ctx context.Context, productID uuid.UUID, quantity int) (*Inventory, error)
	incrementFunc      func(ctx context.Context, productID uuid.UUID, incrementBy int, userID uuid.UUID) (*Inventory, error)
	checkAvailFunc     func(ctx context.Context, productID uuid.UUID, quantity int) (bool, error)
	createInventoryFunc func(ctx context.Context, productID uuid.UUID, quantity int) (*Inventory, error)
}

func (m *MockInventoryService) GetInventory(ctx context.Context, productID uuid.UUID) (*Inventory, error) {
	if m.getFunc != nil {
		return m.getFunc(ctx, productID)
	}
	return nil, nil
}

func (m *MockInventoryService) UpdateInventory(ctx context.Context, productID uuid.UUID, quantity int) (*Inventory, error) {
	if m.updateFunc != nil {
		return m.updateFunc(ctx, productID, quantity)
	}
	return nil, nil
}

func (m *MockInventoryService) CheckAvailability(ctx context.Context, productID uuid.UUID, quantity int) (bool, error) {
	if m.checkAvailFunc != nil {
		return m.checkAvailFunc(ctx, productID, quantity)
	}
	return false, nil
}

func (m *MockInventoryService) IncrementInventory(ctx context.Context, productID uuid.UUID, incrementBy int, userID uuid.UUID) (*Inventory, error) {
	if m.incrementFunc != nil {
		return m.incrementFunc(ctx, productID, incrementBy, userID)
	}
	return nil, nil
}

func (m *MockInventoryService) CreateInventory(ctx context.Context, productID uuid.UUID, quantity int) (*Inventory, error) {
	if m.createInventoryFunc != nil {
		return m.createInventoryFunc(ctx, productID, quantity)
	}
	return nil, nil
}

func TestInventoryHandler_Get(t *testing.T) {
	validProductID := uuid.New()

	tests := []struct {
		name           string
		productID      string
		mockFunc       func(ctx context.Context, productID uuid.UUID) (*Inventory, error)
		expectedStatus int
		checkResponse  func(t *testing.T, body []byte)
	}{
		{
			name:      "get existing inventory",
			productID: validProductID.String(),
			mockFunc: func(ctx context.Context, productID uuid.UUID) (*Inventory, error) {
				return &Inventory{
					ID:        uuid.New(),
					ProductID: productID,
					Quantity:  100,
					UpdatedAt: time.Now(),
				}, nil
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body []byte) {
				var response InventoryResponse
				if err := json.Unmarshal(body, &response); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				if response.Quantity != 100 {
					t.Errorf("Expected quantity 100, got %d", response.Quantity)
				}
				if response.ProductID != validProductID {
					t.Errorf("Expected product ID %s, got %s", validProductID, response.ProductID)
				}
			},
		},
		{
			name:      "inventory not found",
			productID: uuid.New().String(),
			mockFunc: func(ctx context.Context, productID uuid.UUID) (*Inventory, error) {
				return nil, shared.ErrNotFound
			},
			expectedStatus: http.StatusNotFound,
			checkResponse: func(t *testing.T, body []byte) {
				var response shared.ErrorResponse
				if err := json.Unmarshal(body, &response); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				if response.Message != "inventory not found" {
					t.Errorf("Expected 'inventory not found', got %s", response.Message)
				}
			},
		},
		{
			name:           "invalid UUID",
			productID:      "invalid-uuid",
			mockFunc:       nil,
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, body []byte) {
				var response shared.ErrorResponse
				if err := json.Unmarshal(body, &response); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				if response.Message != "invalid product ID" {
					t.Errorf("Expected 'invalid product ID', got %s", response.Message)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockInventoryService{getFunc: tt.mockFunc}
			handler := NewHandler(mockService)

			req := httptest.NewRequest(http.MethodGet, "/inventory/"+tt.productID, nil)
			req.SetPathValue("product_id", tt.productID)
			rr := httptest.NewRecorder()

			handler.Get(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			if tt.checkResponse != nil {
				tt.checkResponse(t, rr.Body.Bytes())
			}
		})
	}
}

func TestInventoryHandler_Update(t *testing.T) {
	validProductID := uuid.New()

	tests := []struct {
		name           string
		productID      string
		requestBody    interface{}
		mockFunc       func(ctx context.Context, productID uuid.UUID, incrementBy int, userID uuid.UUID) (*Inventory, error)
		expectedStatus int
		checkResponse  func(t *testing.T, body []byte)
	}{
		{
			name:        "update inventory successfully",
			productID:   validProductID.String(),
			requestBody: UpdateInventoryRequest{IncrementBy: 200, Reason: "Restock"},
			mockFunc: func(ctx context.Context, productID uuid.UUID, incrementBy int, userID uuid.UUID) (*Inventory, error) {
				return &Inventory{
					ID:        uuid.New(),
					ProductID: productID,
					Quantity:  100 + incrementBy, // Simulate incrementing from 100
					UpdatedAt: time.Now(),
				}, nil
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body []byte) {
				var response InventoryResponse
				if err := json.Unmarshal(body, &response); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				if response.Quantity != 300 { // 100 base + 200 increment
					t.Errorf("Expected quantity 300, got %d", response.Quantity)
				}
			},
		},
		{
			name:        "update to zero quantity",
			productID:   validProductID.String(),
			requestBody: UpdateInventoryRequest{IncrementBy: 0, Reason: "No change"},
			mockFunc: func(ctx context.Context, productID uuid.UUID, incrementBy int, userID uuid.UUID) (*Inventory, error) {
				return &Inventory{
					ID:        uuid.New(),
					ProductID: productID,
					Quantity:  100 + incrementBy, // Simulate incrementing from 100
					UpdatedAt: time.Now(),
				}, nil
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body []byte) {
				var response InventoryResponse
				if err := json.Unmarshal(body, &response); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				if response.Quantity != 100 { // 100 base + 0 increment
					t.Errorf("Expected quantity 100, got %d", response.Quantity)
				}
			},
		},
		{
			name:        "negative quantity rejected",
			productID:   validProductID.String(),
			requestBody: UpdateInventoryRequest{IncrementBy: -10, Reason: "Return"},
			mockFunc: func(ctx context.Context, productID uuid.UUID, incrementBy int, userID uuid.UUID) (*Inventory, error) {
				return nil, shared.ErrNegativeQuantity
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, body []byte) {
				var response shared.ErrorResponse
				if err := json.Unmarshal(body, &response); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				if response.Message != "quantity cannot be negative" {
					t.Errorf("Expected 'quantity cannot be negative', got %s", response.Message)
				}
			},
		},
		{
			name:        "inventory not found",
			productID:   uuid.New().String(),
			requestBody: UpdateInventoryRequest{IncrementBy: 100, Reason: "Restock"},
			mockFunc: func(ctx context.Context, productID uuid.UUID, incrementBy int, userID uuid.UUID) (*Inventory, error) {
				return nil, shared.ErrNotFound
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "invalid UUID",
			productID:      "invalid-uuid",
			requestBody:    UpdateInventoryRequest{IncrementBy: 100, Reason: "Restock"},
			mockFunc:       nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid JSON body",
			productID:      validProductID.String(),
			requestBody:    "invalid json",
			mockFunc:       nil,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockInventoryService{incrementFunc: tt.mockFunc}
			handler := NewHandler(mockService)

			var body []byte
			var err error
			if str, ok := tt.requestBody.(string); ok {
				body = []byte(str)
			} else {
				body, err = json.Marshal(tt.requestBody)
				if err != nil {
					t.Fatalf("Failed to marshal request: %v", err)
				}
			}

			req := httptest.NewRequest(http.MethodPut, "/inventory/"+tt.productID, bytes.NewReader(body))
			req.SetPathValue("product_id", tt.productID)
			req.Header.Set("Content-Type", "application/json")
			req = requestWithUserContext(req, uuid.New()) // Add user context
			rr := httptest.NewRecorder()

			handler.Update(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			if tt.checkResponse != nil {
				tt.checkResponse(t, rr.Body.Bytes())
			}
		})
	}
}

// Test inventory update flow
func TestInventoryHandler_UpdateFlow(t *testing.T) {
	productID := uuid.New()
	inventory := map[uuid.UUID]*Inventory{
		productID: {
			ID:        uuid.New(),
			ProductID: productID,
			Quantity:  50,
			UpdatedAt: time.Now(),
		},
	}

	mockService := &MockInventoryService{
		getFunc: func(ctx context.Context, pid uuid.UUID) (*Inventory, error) {
			if inv, ok := inventory[pid]; ok {
				return inv, nil
			}
			return nil, shared.ErrNotFound
		},
		incrementFunc: func(ctx context.Context, pid uuid.UUID, incrementBy int, userID uuid.UUID) (*Inventory, error) {
			inv, ok := inventory[pid]
			if !ok {
				return nil, shared.ErrNotFound
			}
			newQuantity := inv.Quantity + incrementBy
			if newQuantity < 0 {
				return nil, shared.ErrNegativeQuantity
			}
			inv.Quantity = newQuantity
			inv.UpdatedAt = time.Now()
			return inv, nil
		},
	}

	handler := NewHandler(mockService)

	// Get initial inventory
	t.Run("get_initial", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/inventory/"+productID.String(), nil)
		req.SetPathValue("product_id", productID.String())
		rr := httptest.NewRecorder()

		handler.Get(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("Get failed: %d", rr.Code)
		}

		var response InventoryResponse
		if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal get response: %v", err)
		}
		if response.Quantity != 50 {
			t.Errorf("Expected initial quantity 50, got %d", response.Quantity)
		}
	})

	// Update inventory
	t.Run("update", func(t *testing.T) {
		body, _ := json.Marshal(UpdateInventoryRequest{IncrementBy: 150, Reason: "Restock"})
		req := httptest.NewRequest(http.MethodPut, "/inventory/"+productID.String(), bytes.NewReader(body))
		req.SetPathValue("product_id", productID.String())
		req.Header.Set("Content-Type", "application/json")
		req = requestWithUserContext(req, uuid.New()) // Add user context
		rr := httptest.NewRecorder()

		handler.Update(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("Update failed: %d", rr.Code)
		}

		var response InventoryResponse
		if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal update response: %v", err)
		}
		if response.Quantity != 200 { // 50 base + 150 increment
			t.Errorf("Expected updated quantity 200, got %d", response.Quantity)
		}
	})

	// Verify update persisted
	t.Run("verify_update", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/inventory/"+productID.String(), nil)
		req.SetPathValue("product_id", productID.String())
		rr := httptest.NewRecorder()

		handler.Get(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("Get after update failed: %d", rr.Code)
		}

		var response InventoryResponse
		if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal verify response: %v", err)
		}
		if response.Quantity != 200 { // 50 base + 150 increment
			t.Errorf("Expected quantity 200 after update, got %d", response.Quantity)
		}
	})
}

// TestInventoryHandler_Update_Unauthorized tests Update without user context
func TestInventoryHandler_Update_Unauthorized(t *testing.T) {
	mockService := &MockInventoryService{}
	handler := NewHandler(mockService)

	productID := uuid.New()
	body, _ := json.Marshal(UpdateInventoryRequest{IncrementBy: 100, Reason: "Restock"})
	req := httptest.NewRequest(http.MethodPatch, "/inventory/"+productID.String(), bytes.NewReader(body))
	req.SetPathValue("product_id", productID.String())
	req.Header.Set("Content-Type", "application/json")
	// No user context
	rr := httptest.NewRecorder()

	handler.Update(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, rr.Code)
	}
}

// TestInventoryHandler_Update_InvalidUserContext tests Update with invalid user context
func TestInventoryHandler_Update_InvalidUserContext(t *testing.T) {
	mockService := &MockInventoryService{}
	handler := NewHandler(mockService)

	productID := uuid.New()
	body, _ := json.Marshal(UpdateInventoryRequest{IncrementBy: 100, Reason: "Restock"})
	req := httptest.NewRequest(http.MethodPatch, "/inventory/"+productID.String(), bytes.NewReader(body))
	req.SetPathValue("product_id", productID.String())
	req.Header.Set("Content-Type", "application/json")
	// Invalid user context (not implementing GetID)
	ctx := context.WithValue(req.Context(), shared.UserContextKey, "invalid")
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.Update(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, rr.Code)
	}
}

// TestInventoryHandler_Get_MissingPathValue tests Get with missing path value
func TestInventoryHandler_Get_MissingPathValue(t *testing.T) {
	mockService := &MockInventoryService{}
	handler := NewHandler(mockService)

	req := httptest.NewRequest(http.MethodGet, "/inventory/", nil)
	// No path value set
	rr := httptest.NewRecorder()

	handler.Get(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

// TestInventoryHandler_Update_MissingPathValue tests Update with missing path value
func TestInventoryHandler_Update_MissingPathValue(t *testing.T) {
	mockService := &MockInventoryService{}
	handler := NewHandler(mockService)

	body, _ := json.Marshal(UpdateInventoryRequest{IncrementBy: 100, Reason: "Restock"})
	req := httptest.NewRequest(http.MethodPatch, "/inventory/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = requestWithUserContext(req, uuid.New())
	// No path value set
	rr := httptest.NewRecorder()

	handler.Update(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

// TestInventoryHandler_Update_InvalidJSON tests Update with invalid JSON
func TestInventoryHandler_Update_InvalidJSON(t *testing.T) {
	mockService := &MockInventoryService{}
	handler := NewHandler(mockService)

	req := httptest.NewRequest(http.MethodPatch, "/inventory/"+uuid.New().String(), bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("product_id", uuid.New().String())
	req = requestWithUserContext(req, uuid.New())
	rr := httptest.NewRecorder()

	handler.Update(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}

	var response shared.ErrorResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if response.Message != "invalid request body" {
		t.Errorf("Expected 'invalid request body', got %s", response.Message)
	}
}

// TestInventoryHandler_Update_UserContextTypeAssertionFailure tests Update when user context type assertion fails
func TestInventoryHandler_Update_UserContextTypeAssertionFailure(t *testing.T) {
	mockService := &MockInventoryService{}
	handler := NewHandler(mockService)

	body, _ := json.Marshal(UpdateInventoryRequest{IncrementBy: 100, Reason: "Restock"})
	req := httptest.NewRequest(http.MethodPatch, "/inventory/"+uuid.New().String(), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("product_id", uuid.New().String())
	// Add invalid user context (not implementing GetID())
	ctx := context.WithValue(req.Context(), shared.UserContextKey, "invalid-user")
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.Update(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, rr.Code)
	}
}

// TestInventoryHandler_HandleError tests all error handling paths
func TestInventoryHandler_HandleError(t *testing.T) {
	tests := []struct {
		name           string
		err            error
		expectedStatus int
		expectedMsg    string
	}{
		{
			name:           "ErrNotFound",
			err:            shared.ErrNotFound,
			expectedStatus: http.StatusNotFound,
			expectedMsg:    "inventory not found",
		},
		{
			name:           "ErrNegativeQuantity",
			err:            shared.ErrNegativeQuantity,
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "quantity cannot be negative",
		},
		{
			name:           "unknown error",
			err:            shared.ErrBadRequest,
			expectedStatus: http.StatusInternalServerError,
			expectedMsg:    "internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockInventoryService{
				getFunc: func(ctx context.Context, productID uuid.UUID) (*Inventory, error) {
					return nil, tt.err
				},
			}
			handler := NewHandler(mockService)

			req := httptest.NewRequest(http.MethodGet, "/inventory/"+uuid.New().String(), nil)
			req.SetPathValue("product_id", uuid.New().String())
			rr := httptest.NewRecorder()

			handler.Get(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			var response shared.ErrorResponse
			if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
				t.Fatalf("Failed to unmarshal response: %v", err)
			}
			if response.Message != tt.expectedMsg {
				t.Errorf("Expected message '%s', got '%s'", tt.expectedMsg, response.Message)
			}
		})
	}
}