package orders

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

// MockOrderService is a mock implementation of OrderService for handler testing
type MockOrderService struct {
	createOrderFunc func(ctx context.Context, userID uuid.UUID, shippingAddress CreateOrderRequest) (*Order, []OrderItem, error)
	getOrderFunc    func(ctx context.Context, userID uuid.UUID, orderID uuid.UUID, isAdmin bool) (*Order, []OrderItem, error)
	listOrdersFunc  func(ctx context.Context, userID uuid.UUID, isAdmin bool, page, limit int) ([]Order, int64, error)
}

func (m *MockOrderService) CreateOrder(ctx context.Context, userID uuid.UUID, shippingAddress CreateOrderRequest) (*Order, []OrderItem, error) {
	if m.createOrderFunc != nil {
		return m.createOrderFunc(ctx, userID, shippingAddress)
	}
	return nil, nil, nil
}

func (m *MockOrderService) GetOrder(ctx context.Context, userID uuid.UUID, orderID uuid.UUID, isAdmin bool) (*Order, []OrderItem, error) {
	if m.getOrderFunc != nil {
		return m.getOrderFunc(ctx, userID, orderID, isAdmin)
	}
	return nil, nil, nil
}

func (m *MockOrderService) ListOrders(ctx context.Context, userID uuid.UUID, isAdmin bool, page, limit int) ([]Order, int64, error) {
	if m.listOrdersFunc != nil {
		return m.listOrdersFunc(ctx, userID, isAdmin, page, limit)
	}
	return nil, 0, nil
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

func TestOrderHandler_Create(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name           string
		mockFunc       func(ctx context.Context, userID uuid.UUID, shippingAddress CreateOrderRequest) (*Order, []OrderItem, error)
		expectedStatus int
		checkResponse  func(t *testing.T, body []byte)
	}{
		{
			name: "successful create order",
			mockFunc: func(ctx context.Context, uid uuid.UUID, shippingAddress CreateOrderRequest) (*Order, []OrderItem, error) {
				orderID := uuid.New()
				return &Order{
						ID:         orderID,
						UserID:     uid,
						TotalPrice: 99.99,
						Status:     StatusPending,
						CreatedAt:  time.Now(),
					}, []OrderItem{
						{
							ID:        uuid.New(),
							OrderID:   orderID,
							ProductID: uuid.New(),
							Quantity:  2,
							Price:     49.99,
						},
					}, nil
			},
			expectedStatus: http.StatusCreated,
			checkResponse: func(t *testing.T, body []byte) {
				var response OrderResponse
				if err := json.Unmarshal(body, &response); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				if response.TotalPrice != 99.99 {
					t.Errorf("Expected total price 99.99, got %f", response.TotalPrice)
				}
				if len(response.Items) != 1 {
					t.Errorf("Expected 1 item, got %d", len(response.Items))
				}
			},
		},
		{
			name: "empty cart",
			mockFunc: func(ctx context.Context, uid uuid.UUID, shippingAddress CreateOrderRequest) (*Order, []OrderItem, error) {
				return nil, nil, shared.ErrEmptyCart
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, body []byte) {
				var response shared.ErrorResponse
				if err := json.Unmarshal(body, &response); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				if response.Message != "cart is empty" {
					t.Errorf("Expected 'cart is empty', got %s", response.Message)
				}
			},
		},
		{
			name: "insufficient inventory",
			mockFunc: func(ctx context.Context, uid uuid.UUID, shippingAddress CreateOrderRequest) (*Order, []OrderItem, error) {
				return nil, nil, shared.ErrInsufficientInventory
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
			name:           "unauthorized - no user in context",
			mockFunc:       nil,
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockOrderService{createOrderFunc: tt.mockFunc}
			handler := NewHandler(mockService)

			var reqBody []byte
			if tt.name != "unauthorized - no user in context" {
				shippingAddress := CreateOrderRequest{
					ShippingStreet:  "123 Main St",
					ShippingCity:    "New York",
					ShippingZip:     "10001",
					ShippingCountry: "USA",
				}
				var err error
				reqBody, err = json.Marshal(shippingAddress)
				if err != nil {
					t.Fatalf("Failed to marshal request: %v", err)
				}
			}
			req := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewBuffer(reqBody))
			req.Header.Set("Content-Type", "application/json")
			if tt.name != "unauthorized - no user in context" {
				req = createRequestWithUser(req, userID, "customer")
			}
			rr := httptest.NewRecorder()

			handler.Create(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			if tt.checkResponse != nil {
				tt.checkResponse(t, rr.Body.Bytes())
			}
		})
	}
}

func TestOrderHandler_Get(t *testing.T) {
	userID := uuid.New()
	orderID := uuid.New()
	otherUserID := uuid.New()

	tests := []struct {
		name           string
		orderID        string
		userRole       string
		mockFunc       func(ctx context.Context, userID uuid.UUID, orderID uuid.UUID, isAdmin bool) (*Order, []OrderItem, error)
		expectedStatus int
		checkResponse  func(t *testing.T, body []byte)
	}{
		{
			name:     "successful get own order",
			orderID:  orderID.String(),
			userRole: "customer",
			mockFunc: func(ctx context.Context, uid uuid.UUID, oid uuid.UUID, isAdmin bool) (*Order, []OrderItem, error) {
				return &Order{
						ID:         oid,
						UserID:     uid,
						TotalPrice: 99.99,
						Status:     StatusPending,
						CreatedAt:  time.Now(),
					}, []OrderItem{
						{
							ID:        uuid.New(),
							OrderID:   oid,
							ProductID: uuid.New(),
							Quantity:  2,
							Price:     49.99,
						},
					}, nil
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body []byte) {
				var response OrderResponse
				if err := json.Unmarshal(body, &response); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				if response.TotalPrice != 99.99 {
					t.Errorf("Expected total price 99.99, got %f", response.TotalPrice)
				}
			},
		},
		{
			name:     "admin can get any order",
			orderID:  orderID.String(),
			userRole: "admin",
			mockFunc: func(ctx context.Context, uid uuid.UUID, oid uuid.UUID, isAdmin bool) (*Order, []OrderItem, error) {
				return &Order{
					ID:         oid,
					UserID:     otherUserID, // Different user
					TotalPrice: 99.99,
					Status:     StatusPending,
					CreatedAt:  time.Now(),
				}, []OrderItem{}, nil
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:     "customer cannot get other user's order",
			orderID:  orderID.String(),
			userRole: "customer",
			mockFunc: func(ctx context.Context, uid uuid.UUID, oid uuid.UUID, isAdmin bool) (*Order, []OrderItem, error) {
				return nil, nil, shared.ErrForbidden
			},
			expectedStatus: http.StatusForbidden,
			checkResponse: func(t *testing.T, body []byte) {
				var response shared.ErrorResponse
				if err := json.Unmarshal(body, &response); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				if response.Message != "access denied" {
					t.Errorf("Expected 'access denied', got %s", response.Message)
				}
			},
		},
		{
			name:     "order not found",
			orderID:  uuid.New().String(),
			userRole: "customer",
			mockFunc: func(ctx context.Context, uid uuid.UUID, oid uuid.UUID, isAdmin bool) (*Order, []OrderItem, error) {
				return nil, nil, shared.ErrNotFound
			},
			expectedStatus: http.StatusNotFound,
			checkResponse: func(t *testing.T, body []byte) {
				var response shared.ErrorResponse
				if err := json.Unmarshal(body, &response); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				if response.Message != "order not found" {
					t.Errorf("Expected 'order not found', got %s", response.Message)
				}
			},
		},
		{
			name:           "invalid UUID",
			orderID:        "invalid-uuid",
			userRole:       "customer",
			mockFunc:       nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "unauthorized - no user in context",
			orderID:        orderID.String(),
			userRole:       "",
			mockFunc:       nil,
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockOrderService{getOrderFunc: tt.mockFunc}
			handler := NewHandler(mockService)

			req := httptest.NewRequest(http.MethodGet, "/orders/"+tt.orderID, nil)
			req.SetPathValue("id", tt.orderID)
			if tt.name != "unauthorized - no user in context" {
				req = createRequestWithUser(req, userID, tt.userRole)
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

func TestOrderHandler_List(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name           string
		queryParams    string
		userRole       string
		mockFunc       func(ctx context.Context, userID uuid.UUID, isAdmin bool, page, limit int) ([]Order, int64, error)
		expectedStatus int
		checkResponse  func(t *testing.T, body []byte)
	}{
		{
			name:        "customer lists own orders",
			queryParams: "",
			userRole:    "customer",
			mockFunc: func(ctx context.Context, uid uuid.UUID, isAdmin bool, page, limit int) ([]Order, int64, error) {
				if isAdmin {
					t.Error("Customer should not be admin")
				}
				return []Order{
					{
						ID:         uuid.New(),
						UserID:     uid,
						TotalPrice: 99.99,
						Status:     StatusPending,
						CreatedAt:  time.Now(),
					},
				}, 1, nil
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body []byte) {
				var response shared.PaginatedResponse
				if err := json.Unmarshal(body, &response); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				if response.TotalCount != 1 {
					t.Errorf("Expected total count 1, got %d", response.TotalCount)
				}
			},
		},
		{
			name:        "admin lists all orders",
			queryParams: "",
			userRole:    "admin",
			mockFunc: func(ctx context.Context, uid uuid.UUID, isAdmin bool, page, limit int) ([]Order, int64, error) {
				if !isAdmin {
					t.Error("Admin flag should be true")
				}
				return []Order{
					{
						ID:         uuid.New(),
						UserID:     uuid.New(),
						TotalPrice: 99.99,
						Status:     StatusPending,
						CreatedAt:  time.Now(),
					},
					{
						ID:         uuid.New(),
						UserID:     uuid.New(),
						TotalPrice: 149.99,
						Status:     StatusPending,
						CreatedAt:  time.Now(),
					},
				}, 2, nil
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body []byte) {
				var response shared.PaginatedResponse
				if err := json.Unmarshal(body, &response); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				if response.TotalCount != 2 {
					t.Errorf("Expected total count 2, got %d", response.TotalCount)
				}
			},
		},
		{
			name:        "list with pagination",
			queryParams: "?page=2&limit=5",
			userRole:    "customer",
			mockFunc: func(ctx context.Context, uid uuid.UUID, isAdmin bool, page, limit int) ([]Order, int64, error) {
				if page != 2 || limit != 5 {
					t.Errorf("Expected page=2, limit=5, got page=%d, limit=%d", page, limit)
				}
				return []Order{}, 10, nil
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body []byte) {
				var response shared.PaginatedResponse
				if err := json.Unmarshal(body, &response); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				if response.TotalCount != 10 {
					t.Errorf("Expected total count 10, got %d", response.TotalCount)
				}
			},
		},
		{
			name:           "unauthorized - no user in context",
			queryParams:    "",
			userRole:       "",
			mockFunc:       nil,
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockOrderService{listOrdersFunc: tt.mockFunc}
			handler := NewHandler(mockService)

			req := httptest.NewRequest(http.MethodGet, "/orders"+tt.queryParams, nil)
			if tt.name != "unauthorized - no user in context" {
				req = createRequestWithUser(req, userID, tt.userRole)
			}
			rr := httptest.NewRecorder()

			handler.List(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			if tt.checkResponse != nil {
				tt.checkResponse(t, rr.Body.Bytes())
			}
		})
	}
}

// TestOrderHandler_HandleError tests all error handling paths
func TestOrderHandler_HandleError(t *testing.T) {
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
			expectedMsg:    "order not found",
		},
		{
			name:           "ErrForbidden",
			err:            shared.ErrForbidden,
			expectedStatus: http.StatusForbidden,
			expectedMsg:    "access denied",
		},
		{
			name:           "ErrEmptyCart",
			err:            shared.ErrEmptyCart,
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "cart is empty",
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
			mockService := &MockOrderService{
				getOrderFunc: func(ctx context.Context, uid uuid.UUID, oid uuid.UUID, isAdmin bool) (*Order, []OrderItem, error) {
					return nil, nil, tt.err
				},
			}
			handler := NewHandler(mockService)

			userID := uuid.New()
			req := httptest.NewRequest(http.MethodGet, "/orders/"+uuid.New().String(), nil)
			req.SetPathValue("id", uuid.New().String())
			req = createRequestWithUser(req, userID, "customer")
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
