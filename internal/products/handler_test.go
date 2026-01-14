package products

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/mschl/ecommerce-api/internal/shared"
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

// MockProductService is a mock implementation of ProductService for handler testing
type MockProductService struct {
	createFunc func(ctx context.Context, req CreateProductRequest, userID uuid.UUID) (*Product, error)
	getFunc    func(ctx context.Context, id uuid.UUID) (*Product, error)
	listFunc   func(ctx context.Context, page, limit int, categoryID *uuid.UUID) ([]Product, int64, error)
	updateFunc func(ctx context.Context, id uuid.UUID, req UpdateProductRequest) (*Product, error)
	deleteFunc func(ctx context.Context, id uuid.UUID) error
}

func (m *MockProductService) CreateProduct(ctx context.Context, req CreateProductRequest, userID uuid.UUID) (*Product, error) {
	if m.createFunc != nil {
		return m.createFunc(ctx, req, userID)
	}
	return nil, nil
}

func (m *MockProductService) GetProduct(ctx context.Context, id uuid.UUID) (*Product, error) {
	if m.getFunc != nil {
		return m.getFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockProductService) ListProducts(ctx context.Context, page, limit int, categoryID *uuid.UUID) ([]Product, int64, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx, page, limit, categoryID)
	}
	return nil, 0, nil
}

func (m *MockProductService) UpdateProduct(ctx context.Context, id uuid.UUID, req UpdateProductRequest) (*Product, error) {
	if m.updateFunc != nil {
		return m.updateFunc(ctx, id, req)
	}
	return nil, nil
}

func (m *MockProductService) DeleteProduct(ctx context.Context, id uuid.UUID) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, id)
	}
	return nil
}

func TestProductHandler_List(t *testing.T) {
	tests := []struct {
		name           string
		queryParams    string
		mockFunc       func(ctx context.Context, page, limit int, categoryID *uuid.UUID) ([]Product, int64, error)
		expectedStatus int
		checkResponse  func(t *testing.T, body []byte)
	}{
		{
			name:        "list products with defaults",
			queryParams: "",
			mockFunc: func(ctx context.Context, page, limit int, categoryID *uuid.UUID) ([]Product, int64, error) {
				return []Product{
					{ID: uuid.New(), Name: "Product 1", Price: 10.00},
					{ID: uuid.New(), Name: "Product 2", Price: 20.00},
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
			name:        "list products with pagination",
			queryParams: "?page=2&limit=5",
			mockFunc: func(ctx context.Context, page, limit int, categoryID *uuid.UUID) ([]Product, int64, error) {
				if page != 2 || limit != 5 {
					t.Errorf("Expected page=2, limit=5, got page=%d, limit=%d", page, limit)
				}
				return []Product{}, 10, nil
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
			name:        "empty product list",
			queryParams: "",
			mockFunc: func(ctx context.Context, page, limit int, categoryID *uuid.UUID) ([]Product, int64, error) {
				return []Product{}, 0, nil
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body []byte) {
				var response shared.PaginatedResponse
				if err := json.Unmarshal(body, &response); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				if response.TotalCount != 0 {
					t.Errorf("Expected total count 0, got %d", response.TotalCount)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockProductService{listFunc: tt.mockFunc}
			handler := NewHandler(mockService)

			req := httptest.NewRequest(http.MethodGet, "/products"+tt.queryParams, nil)
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

func TestProductHandler_Get(t *testing.T) {
	validID := uuid.New()

	tests := []struct {
		name           string
		productID      string
		mockFunc       func(ctx context.Context, id uuid.UUID) (*Product, error)
		expectedStatus int
		checkResponse  func(t *testing.T, body []byte)
	}{
		{
			name:      "get existing product",
			productID: validID.String(),
			mockFunc: func(ctx context.Context, id uuid.UUID) (*Product, error) {
				return &Product{
					ID:          id,
					Name:        "Test Product",
					Description: "A test product",
					Price:       29.99,
					CreatedAt:   time.Now(),
				}, nil
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body []byte) {
				var response Product
				if err := json.Unmarshal(body, &response); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				if response.Name != "Test Product" {
					t.Errorf("Expected name 'Test Product', got %s", response.Name)
				}
				if response.Price != 29.99 {
					t.Errorf("Expected price 29.99, got %f", response.Price)
				}
			},
		},
		{
			name:      "product not found",
			productID: uuid.New().String(),
			mockFunc: func(ctx context.Context, id uuid.UUID) (*Product, error) {
				return nil, shared.ErrNotFound
			},
			expectedStatus: http.StatusNotFound,
			checkResponse: func(t *testing.T, body []byte) {
				var response shared.ErrorResponse
				if err := json.Unmarshal(body, &response); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				if response.Message != "product not found" {
					t.Errorf("Expected 'product not found', got %s", response.Message)
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
			mockService := &MockProductService{getFunc: tt.mockFunc}
			handler := NewHandler(mockService)

			req := httptest.NewRequest(http.MethodGet, "/products/"+tt.productID, nil)
			req.SetPathValue("id", tt.productID)
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

func TestProductHandler_Create(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		mockFunc       func(ctx context.Context, req CreateProductRequest, userID uuid.UUID) (*Product, error)
		expectedStatus int
		checkResponse  func(t *testing.T, body []byte)
	}{
		{
			name: "create valid product",
			requestBody: CreateProductRequest{
				Name:        "New Product",
				Description: "A new product",
				Price:       49.99,
			},
			mockFunc: func(ctx context.Context, req CreateProductRequest, userID uuid.UUID) (*Product, error) {
				return &Product{
					ID:          uuid.New(),
					Name:        req.Name,
					Description: req.Description,
					Price:       req.Price,
					CreatedAt:   time.Now(),
				}, nil
			},
			expectedStatus: http.StatusCreated,
			checkResponse: func(t *testing.T, body []byte) {
				var response Product
				if err := json.Unmarshal(body, &response); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				if response.Name != "New Product" {
					t.Errorf("Expected name 'New Product', got %s", response.Name)
				}
				if response.Price != 49.99 {
					t.Errorf("Expected price 49.99, got %f", response.Price)
				}
			},
		},
		{
			name: "create product with empty name",
			requestBody: CreateProductRequest{
				Name:  "",
				Price: 49.99,
			},
			mockFunc: func(ctx context.Context, req CreateProductRequest, userID uuid.UUID) (*Product, error) {
				return nil, ErrInvalidProductName
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, body []byte) {
				var response shared.ErrorResponse
				if err := json.Unmarshal(body, &response); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				if response.Message != "product name is required" {
					t.Errorf("Expected 'product name is required', got %s", response.Message)
				}
			},
		},
		{
			name: "create product with negative price",
			requestBody: CreateProductRequest{
				Name:  "Test Product",
				Price: -10.00,
			},
			mockFunc: func(ctx context.Context, req CreateProductRequest, userID uuid.UUID) (*Product, error) {
				return nil, ErrInvalidProductPrice
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, body []byte) {
				var response shared.ErrorResponse
				if err := json.Unmarshal(body, &response); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				if response.Message != "product price cannot be negative" {
					t.Errorf("Expected 'product price cannot be negative', got %s", response.Message)
				}
			},
		},
		{
			name:           "invalid JSON body",
			requestBody:    "invalid json",
			mockFunc:       nil,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockProductService{createFunc: tt.mockFunc}
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

			req := httptest.NewRequest(http.MethodPost, "/products", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			req = requestWithUserContext(req, uuid.New()) // Add user context
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

func TestProductHandler_Update(t *testing.T) {
	validID := uuid.New()
	newName := "Updated Product"
	newPrice := 59.99

	tests := []struct {
		name           string
		productID      string
		requestBody    interface{}
		mockFunc       func(ctx context.Context, id uuid.UUID, req UpdateProductRequest) (*Product, error)
		expectedStatus int
		checkResponse  func(t *testing.T, body []byte)
	}{
		{
			name:      "update existing product",
			productID: validID.String(),
			requestBody: UpdateProductRequest{
				Name:  &newName,
				Price: &newPrice,
			},
			mockFunc: func(ctx context.Context, id uuid.UUID, req UpdateProductRequest) (*Product, error) {
				return &Product{
					ID:        id,
					Name:      *req.Name,
					Price:     *req.Price,
					UpdatedAt: time.Now(),
				}, nil
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body []byte) {
				var response Product
				if err := json.Unmarshal(body, &response); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				if response.Name != "Updated Product" {
					t.Errorf("Expected name 'Updated Product', got %s", response.Name)
				}
			},
		},
		{
			name:        "update non-existing product",
			productID:   uuid.New().String(),
			requestBody: UpdateProductRequest{Name: &newName},
			mockFunc: func(ctx context.Context, id uuid.UUID, req UpdateProductRequest) (*Product, error) {
				return nil, shared.ErrNotFound
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "invalid UUID",
			productID:      "invalid-uuid",
			requestBody:    UpdateProductRequest{Name: &newName},
			mockFunc:       nil,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockProductService{updateFunc: tt.mockFunc}
			handler := NewHandler(mockService)

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPut, "/products/"+tt.productID, bytes.NewReader(body))
			req.SetPathValue("id", tt.productID)
			req.Header.Set("Content-Type", "application/json")
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

func TestProductHandler_Delete(t *testing.T) {
	validID := uuid.New()

	tests := []struct {
		name           string
		productID      string
		mockFunc       func(ctx context.Context, id uuid.UUID) error
		expectedStatus int
		checkResponse  func(t *testing.T, body []byte)
	}{
		{
			name:      "delete existing product",
			productID: validID.String(),
			mockFunc: func(ctx context.Context, id uuid.UUID) error {
				return nil
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body []byte) {
				var response map[string]string
				if err := json.Unmarshal(body, &response); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				if response["message"] != "product deleted" {
					t.Errorf("Expected 'product deleted', got %s", response["message"])
				}
			},
		},
		{
			name:      "delete non-existing product",
			productID: uuid.New().String(),
			mockFunc: func(ctx context.Context, id uuid.UUID) error {
				return shared.ErrNotFound
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "invalid UUID",
			productID:      "invalid-uuid",
			mockFunc:       nil,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockProductService{deleteFunc: tt.mockFunc}
			handler := NewHandler(mockService)

			req := httptest.NewRequest(http.MethodDelete, "/products/"+tt.productID, nil)
			req.SetPathValue("id", tt.productID)
			rr := httptest.NewRecorder()

			handler.Delete(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			if tt.checkResponse != nil {
				tt.checkResponse(t, rr.Body.Bytes())
			}
		})
	}
}

// Test full CRUD flow
func TestProductHandler_CRUDFlow(t *testing.T) {
	products := make(map[uuid.UUID]*Product)

	mockService := &MockProductService{
		createFunc: func(ctx context.Context, req CreateProductRequest, userID uuid.UUID) (*Product, error) {
			p := &Product{
				ID:          uuid.New(),
				Name:        req.Name,
				Description: req.Description,
				Price:       req.Price,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			}
			products[p.ID] = p
			return p, nil
		},
		getFunc: func(ctx context.Context, id uuid.UUID) (*Product, error) {
			if p, ok := products[id]; ok {
				return p, nil
			}
			return nil, shared.ErrNotFound
		},
		listFunc: func(ctx context.Context, page, limit int, categoryID *uuid.UUID) ([]Product, int64, error) {
			list := make([]Product, 0, len(products))
			for _, p := range products {
				list = append(list, *p)
			}
			return list, int64(len(list)), nil
		},
		updateFunc: func(ctx context.Context, id uuid.UUID, req UpdateProductRequest) (*Product, error) {
			p, ok := products[id]
			if !ok {
				return nil, shared.ErrNotFound
			}
			if req.Name != nil {
				p.Name = *req.Name
			}
			if req.Price != nil {
				p.Price = *req.Price
			}
			p.UpdatedAt = time.Now()
			return p, nil
		},
		deleteFunc: func(ctx context.Context, id uuid.UUID) error {
			if _, ok := products[id]; !ok {
				return shared.ErrNotFound
			}
			delete(products, id)
			return nil
		},
	}

	handler := NewHandler(mockService)

	// Create
	var createdID uuid.UUID
	t.Run("create", func(t *testing.T) {
		body, _ := json.Marshal(CreateProductRequest{
			Name:        "CRUD Test Product",
			Description: "Testing CRUD",
			Price:       99.99,
		})
		req := httptest.NewRequest(http.MethodPost, "/products", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req = requestWithUserContext(req, uuid.New()) // Add user context
		rr := httptest.NewRecorder()

		handler.Create(rr, req)

		if rr.Code != http.StatusCreated {
			t.Fatalf("Create failed: %d", rr.Code)
		}

		var response Product
		json.Unmarshal(rr.Body.Bytes(), &response)
		createdID = response.ID
	})

	// Get
	t.Run("get", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/products/"+createdID.String(), nil)
		req.SetPathValue("id", createdID.String())
		rr := httptest.NewRecorder()

		handler.Get(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("Get failed: %d", rr.Code)
		}

		var response Product
		json.Unmarshal(rr.Body.Bytes(), &response)
		if response.Name != "CRUD Test Product" {
			t.Errorf("Expected 'CRUD Test Product', got %s", response.Name)
		}
	})

	// Update
	t.Run("update", func(t *testing.T) {
		newName := "Updated CRUD Product"
		body, _ := json.Marshal(UpdateProductRequest{Name: &newName})
		req := httptest.NewRequest(http.MethodPut, "/products/"+createdID.String(), bytes.NewReader(body))
		req.SetPathValue("id", createdID.String())
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler.Update(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("Update failed: %d", rr.Code)
		}

		var response Product
		json.Unmarshal(rr.Body.Bytes(), &response)
		if response.Name != "Updated CRUD Product" {
			t.Errorf("Expected 'Updated CRUD Product', got %s", response.Name)
		}
	})

	// List
	t.Run("list", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/products", nil)
		rr := httptest.NewRecorder()

		handler.List(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("List failed: %d", rr.Code)
		}

		var response shared.PaginatedResponse
		json.Unmarshal(rr.Body.Bytes(), &response)
		if response.TotalCount != 1 {
			t.Errorf("Expected 1 product, got %d", response.TotalCount)
		}
	})

	// Delete
	t.Run("delete", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/products/"+createdID.String(), nil)
		req.SetPathValue("id", createdID.String())
		rr := httptest.NewRecorder()

		handler.Delete(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("Delete failed: %d", rr.Code)
		}
	})

	// Verify deletion
	t.Run("get_after_delete", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/products/"+createdID.String(), nil)
		req.SetPathValue("id", createdID.String())
		rr := httptest.NewRecorder()

		handler.Get(rr, req)

		if rr.Code != http.StatusNotFound {
			t.Errorf("Expected 404 after delete, got %d", rr.Code)
		}
	})
}

// TestProductHandler_HandleError tests all error handling paths
func TestProductHandler_HandleError(t *testing.T) {
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
			expectedMsg:    "product not found",
		},
		{
			name:           "ErrInvalidProductName",
			err:            ErrInvalidProductName,
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "product name is required",
		},
		{
			name:           "ErrInvalidProductPrice",
			err:            ErrInvalidProductPrice,
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "product price cannot be negative",
		},
		{
			name:           "unknown error",
			err:            errors.New("boom"),
			expectedStatus: http.StatusInternalServerError,
			expectedMsg:    "internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockProductService{
				getFunc: func(ctx context.Context, id uuid.UUID) (*Product, error) {
					return nil, tt.err
				},
			}
			handler := NewHandler(mockService)

			req := httptest.NewRequest(http.MethodGet, "/products/"+uuid.New().String(), nil)
			req.SetPathValue("id", uuid.New().String())
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
