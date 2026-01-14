package cart

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/mschl/ecommerce-api/internal/auth"
	"github.com/mschl/ecommerce-api/internal/shared"
)

// MockCartService is a mock implementation of CartService for handler testing
type MockCartService struct {
	getCartFunc    func(ctx context.Context, userID uuid.UUID) (*Cart, []CartItem, error)
	addItemFunc    func(ctx context.Context, userID uuid.UUID, req AddItemRequest) (*CartItem, error)
	updateItemFunc func(ctx context.Context, userID uuid.UUID, itemID uuid.UUID, quantity int) (*CartItem, error)
	removeItemFunc func(ctx context.Context, userID uuid.UUID, itemID uuid.UUID) error
	clearCartFunc  func(ctx context.Context, userID uuid.UUID) error
}

func (m *MockCartService) GetCart(ctx context.Context, userID uuid.UUID) (*Cart, []CartItem, error) {
	if m.getCartFunc != nil {
		return m.getCartFunc(ctx, userID)
	}
	return nil, nil, nil
}

func (m *MockCartService) AddItem(ctx context.Context, userID uuid.UUID, req AddItemRequest) (*CartItem, error) {
	if m.addItemFunc != nil {
		return m.addItemFunc(ctx, userID, req)
	}
	return nil, nil
}

func (m *MockCartService) UpdateItem(ctx context.Context, userID uuid.UUID, itemID uuid.UUID, quantity int) (*CartItem, error) {
	if m.updateItemFunc != nil {
		return m.updateItemFunc(ctx, userID, itemID, quantity)
	}
	return nil, nil
}

func (m *MockCartService) RemoveItem(ctx context.Context, userID uuid.UUID, itemID uuid.UUID) error {
	if m.removeItemFunc != nil {
		return m.removeItemFunc(ctx, userID, itemID)
	}
	return nil
}

func (m *MockCartService) ClearCart(ctx context.Context, userID uuid.UUID) error {
	if m.clearCartFunc != nil {
		return m.clearCartFunc(ctx, userID)
	}
	return nil
}

func createRequestWithUser(r *http.Request, userID uuid.UUID, role string) *http.Request {
	user := &auth.User{
		ID:    userID,
		Email: "test@example.com",
		Username:  "Test User",
		Role:  role,
	}
	ctx := context.WithValue(r.Context(), shared.UserContextKey, user)
	return r.WithContext(ctx)
}

func TestCartHandler_Get(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name           string
		mockFunc       func(ctx context.Context, userID uuid.UUID) (*Cart, []CartItem, error)
		expectedStatus int
		checkResponse  func(t *testing.T, body []byte)
	}{
		{
			name: "successful get cart",
			mockFunc: func(ctx context.Context, uid uuid.UUID) (*Cart, []CartItem, error) {
				return &Cart{
					ID:     uuid.New(),
					UserID: uid,
				}, []CartItem{
					{
						ID:        uuid.New(),
						CartID:    uuid.New(),
						ProductID: uuid.New(),
						Quantity:  2,
					},
				}, nil
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body []byte) {
				var response CartResponse
				if err := json.Unmarshal(body, &response); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				if len(response.Items) != 1 {
					t.Errorf("Expected 1 item, got %d", len(response.Items))
				}
			},
		},
		{
			name: "empty cart",
			mockFunc: func(ctx context.Context, uid uuid.UUID) (*Cart, []CartItem, error) {
				return &Cart{
					ID:     uuid.New(),
					UserID: uid,
				}, []CartItem{}, nil
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body []byte) {
				var response CartResponse
				if err := json.Unmarshal(body, &response); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				if len(response.Items) != 0 {
					t.Errorf("Expected 0 items, got %d", len(response.Items))
				}
			},
		},
		{
			name: "unauthorized - no user in context",
			mockFunc: func(ctx context.Context, uid uuid.UUID) (*Cart, []CartItem, error) {
				return nil, nil, nil
			},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockCartService{getCartFunc: tt.mockFunc}
			handler := NewHandler(mockService)

			req := httptest.NewRequest(http.MethodGet, "/cart", nil)
			if tt.name != "unauthorized - no user in context" {
				req = createRequestWithUser(req, userID, "customer")
			}
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

func TestCartHandler_AddItem(t *testing.T) {
	userID := uuid.New()
	productID := uuid.New()

	tests := []struct {
		name           string
		requestBody    interface{}
		mockFunc       func(ctx context.Context, userID uuid.UUID, req AddItemRequest) (*CartItem, error)
		expectedStatus int
		checkResponse  func(t *testing.T, body []byte)
	}{
		{
			name: "successful add item",
			requestBody: AddItemRequest{
				ProductID: productID,
				Quantity:  2,
			},
			mockFunc: func(ctx context.Context, uid uuid.UUID, req AddItemRequest) (*CartItem, error) {
				return &CartItem{
					ID:        uuid.New(),
					CartID:    uuid.New(),
					ProductID: req.ProductID,
					Quantity:  req.Quantity,
					CreatedAt: time.Now(),
				}, nil
			},
			expectedStatus: http.StatusCreated,
			checkResponse: func(t *testing.T, body []byte) {
				var response CartItemResponse
				if err := json.Unmarshal(body, &response); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				if response.Quantity != 2 {
					t.Errorf("Expected quantity 2, got %d", response.Quantity)
				}
			},
		},
		{
			name: "zero quantity rejected",
			requestBody: AddItemRequest{
				ProductID: productID,
				Quantity:  0,
			},
			mockFunc:       nil,
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, body []byte) {
				var response shared.ErrorResponse
				if err := json.Unmarshal(body, &response); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				if response.Message != "quantity must be positive" {
					t.Errorf("Expected 'quantity must be positive', got %s", response.Message)
				}
			},
		},
		{
			name: "negative quantity rejected",
			requestBody: AddItemRequest{
				ProductID: productID,
				Quantity:  -1,
			},
			mockFunc:       nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "product not found",
			requestBody: AddItemRequest{
				ProductID: productID,
				Quantity:  2,
			},
			mockFunc: func(ctx context.Context, uid uuid.UUID, req AddItemRequest) (*CartItem, error) {
				return nil, shared.ErrNotFound
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name: "insufficient inventory",
			requestBody: AddItemRequest{
				ProductID: productID,
				Quantity:  100,
			},
			mockFunc: func(ctx context.Context, uid uuid.UUID, req AddItemRequest) (*CartItem, error) {
				return nil, shared.ErrInsufficientInventory
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, body []byte) {
				var response shared.ErrorResponse
				if err := json.Unmarshal(body, &response); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				if response.Message != "insufficient inventory" {
					t.Errorf("Expected 'insufficient inventory', got %s", response.Message)
				}
			},
		},
		{
			name:           "invalid JSON body",
			requestBody:    "invalid json",
			mockFunc:       nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "unauthorized - no user in context",
			requestBody:    AddItemRequest{ProductID: productID, Quantity: 1},
			mockFunc:       nil,
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockCartService{addItemFunc: tt.mockFunc}
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

			req := httptest.NewRequest(http.MethodPost, "/cart/items", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			if tt.name != "unauthorized - no user in context" {
				req = createRequestWithUser(req, userID, "customer")
			}
			rr := httptest.NewRecorder()

			handler.AddItem(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			if tt.checkResponse != nil {
				tt.checkResponse(t, rr.Body.Bytes())
			}
		})
	}
}

func TestCartHandler_UpdateItem(t *testing.T) {
	userID := uuid.New()
	itemID := uuid.New()

	tests := []struct {
		name           string
		itemID         string
		requestBody    interface{}
		mockFunc       func(ctx context.Context, userID uuid.UUID, itemID uuid.UUID, quantity int) (*CartItem, error)
		expectedStatus int
		checkResponse  func(t *testing.T, body []byte)
	}{
		{
			name:   "successful update item",
			itemID: itemID.String(),
			requestBody: UpdateItemRequest{
				Quantity: 5,
			},
			mockFunc: func(ctx context.Context, uid uuid.UUID, iid uuid.UUID, qty int) (*CartItem, error) {
				return &CartItem{
					ID:        iid,
					CartID:    uuid.New(),
					ProductID: uuid.New(),
					Quantity:  qty,
					UpdatedAt: time.Now(),
				}, nil
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body []byte) {
				var response CartItemResponse
				if err := json.Unmarshal(body, &response); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				if response.Quantity != 5 {
					t.Errorf("Expected quantity 5, got %d", response.Quantity)
				}
			},
		},
		{
			name:   "zero quantity rejected",
			itemID: itemID.String(),
			requestBody: UpdateItemRequest{
				Quantity: 0,
			},
			mockFunc:       nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "item not found",
			itemID: itemID.String(),
			requestBody: UpdateItemRequest{
				Quantity: 5,
			},
			mockFunc: func(ctx context.Context, uid uuid.UUID, iid uuid.UUID, qty int) (*CartItem, error) {
				return nil, shared.ErrNotFound
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "invalid UUID",
			itemID:         "invalid-uuid",
			requestBody:    UpdateItemRequest{Quantity: 5},
			mockFunc:       nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid JSON body",
			itemID:         itemID.String(),
			requestBody:    "invalid json",
			mockFunc:       nil,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockCartService{updateItemFunc: tt.mockFunc}
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

			req := httptest.NewRequest(http.MethodPut, "/cart/items/"+tt.itemID, bytes.NewReader(body))
			req.SetPathValue("item_id", tt.itemID)
			req.Header.Set("Content-Type", "application/json")
			req = createRequestWithUser(req, userID, "customer")
			rr := httptest.NewRecorder()

			handler.UpdateItem(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			if tt.checkResponse != nil {
				tt.checkResponse(t, rr.Body.Bytes())
			}
		})
	}
}

func TestCartHandler_RemoveItem(t *testing.T) {
	userID := uuid.New()
	itemID := uuid.New()

	tests := []struct {
		name           string
		itemID         string
		mockFunc       func(ctx context.Context, userID uuid.UUID, itemID uuid.UUID) error
		expectedStatus int
		checkResponse  func(t *testing.T, body []byte)
	}{
		{
			name:   "successful remove item",
			itemID: itemID.String(),
			mockFunc: func(ctx context.Context, uid uuid.UUID, iid uuid.UUID) error {
				return nil
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body []byte) {
				var response map[string]string
				if err := json.Unmarshal(body, &response); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				if response["message"] != "item removed" {
					t.Errorf("Expected 'item removed', got %s", response["message"])
				}
			},
		},
		{
			name:   "item not found",
			itemID: itemID.String(),
			mockFunc: func(ctx context.Context, uid uuid.UUID, iid uuid.UUID) error {
				return shared.ErrNotFound
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "invalid UUID",
			itemID:         "invalid-uuid",
			mockFunc:       nil,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockCartService{removeItemFunc: tt.mockFunc}
			handler := NewHandler(mockService)

			req := httptest.NewRequest(http.MethodDelete, "/cart/items/"+tt.itemID, nil)
			req.SetPathValue("item_id", tt.itemID)
			req = createRequestWithUser(req, userID, "customer")
			rr := httptest.NewRecorder()

			handler.RemoveItem(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			if tt.checkResponse != nil {
				tt.checkResponse(t, rr.Body.Bytes())
			}
		})
	}
}

func TestCartHandler_Clear(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name           string
		mockFunc       func(ctx context.Context, userID uuid.UUID) error
		expectedStatus int
		checkResponse  func(t *testing.T, body []byte)
	}{
		{
			name: "successful clear cart",
			mockFunc: func(ctx context.Context, uid uuid.UUID) error {
				return nil
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body []byte) {
				var response map[string]string
				if err := json.Unmarshal(body, &response); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				if response["message"] != "cart cleared" {
					t.Errorf("Expected 'cart cleared', got %s", response["message"])
				}
			},
		},
		{
			name:           "unauthorized - no user in context",
			mockFunc:       nil,
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockCartService{clearCartFunc: tt.mockFunc}
			handler := NewHandler(mockService)

			req := httptest.NewRequest(http.MethodDelete, "/cart", nil)
			if tt.name != "unauthorized - no user in context" {
				req = createRequestWithUser(req, userID, "customer")
			}
			rr := httptest.NewRecorder()

			handler.Clear(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			if tt.checkResponse != nil {
				tt.checkResponse(t, rr.Body.Bytes())
			}
		})
	}
}

// TestCartHandler_HandleError tests all error handling paths
func TestCartHandler_HandleError(t *testing.T) {
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
			expectedMsg:    "resource not found",
		},
		{
			name:           "ErrForbidden",
			err:            shared.ErrForbidden,
			expectedStatus: http.StatusForbidden,
			expectedMsg:    "access denied",
		},
		{
			name:           "ErrInsufficientInventory",
			err:            shared.ErrInsufficientInventory,
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "insufficient inventory",
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
			mockService := &MockCartService{
				getCartFunc: func(ctx context.Context, uid uuid.UUID) (*Cart, []CartItem, error) {
					return nil, nil, tt.err
				},
			}
			handler := NewHandler(mockService)

			req := httptest.NewRequest(http.MethodGet, "/cart", nil)
			req = createRequestWithUser(req, uuid.New(), "customer")
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