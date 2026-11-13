package products

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/MarkyGitHub/e-commerce/internal/notifications"
	"github.com/MarkyGitHub/e-commerce/internal/shared"
)

// MockProductRepository is a mock implementation of ProductRepository
type MockProductRepository struct {
	products map[uuid.UUID]*Product
}

func NewMockProductRepository() *MockProductRepository {
	return &MockProductRepository{products: make(map[uuid.UUID]*Product)}
}

func (m *MockProductRepository) Create(ctx context.Context, product *Product) error {
	if product.ID == uuid.Nil {
		product.ID = uuid.New()
	}
	m.products[product.ID] = product
	return nil
}

func (m *MockProductRepository) FindByID(ctx context.Context, id uuid.UUID) (*Product, error) {
	if product, exists := m.products[id]; exists {
		return product, nil
	}
	return nil, shared.ErrNotFound
}

func (m *MockProductRepository) FindAll(ctx context.Context, offset, limit int, categoryID *uuid.UUID) ([]Product, int64, error) {
	products := make([]Product, 0)
	for _, p := range m.products {
		if categoryID != nil && (p.CategoryID == nil || *p.CategoryID != *categoryID) {
			continue
		}
		products = append(products, *p)
	}
	total := int64(len(products))
	
	// Apply pagination
	start := offset
	if start > len(products) {
		start = len(products)
	}
	end := start + limit
	if end > len(products) {
		end = len(products)
	}
	
	return products[start:end], total, nil
}

func (m *MockProductRepository) Update(ctx context.Context, product *Product) error {
	if _, exists := m.products[product.ID]; !exists {
		return shared.ErrNotFound
	}
	m.products[product.ID] = product
	return nil
}

func (m *MockProductRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if _, exists := m.products[id]; !exists {
		return shared.ErrNotFound
	}
	delete(m.products, id)
	return nil
}

// MockNotificationService is a mock implementation of NotificationService
type MockNotificationService struct {
	notifications []*notifications.Notification
}

func NewMockNotificationService() *MockNotificationService {
	return &MockNotificationService{
		notifications: make([]*notifications.Notification, 0),
	}
}

func (m *MockNotificationService) CreateNotification(ctx context.Context, userID uuid.UUID, title, message string) (*notifications.Notification, error) {
	notif := &notifications.Notification{
		ID:      uuid.New(),
		UserID:  userID,
		Title:   title,
		Message: message,
		IsRead:  false,
	}
	m.notifications = append(m.notifications, notif)
	return notif, nil
}

func (m *MockNotificationService) CreateNotificationTx(ctx context.Context, tx *gorm.DB, userID uuid.UUID, title, message string) (*notifications.Notification, error) {
	return m.CreateNotification(ctx, userID, title, message)
}

func (m *MockNotificationService) GetNotifications(ctx context.Context, userID uuid.UUID) ([]notifications.Notification, error) {
	result := make([]notifications.Notification, 0)
	for _, n := range m.notifications {
		if n.UserID == userID {
			result = append(result, *n)
		}
	}
	return result, nil
}

func (m *MockNotificationService) MarkAsRead(ctx context.Context, userID uuid.UUID, notificationID uuid.UUID) error {
	for _, n := range m.notifications {
		if n.ID == notificationID && n.UserID == userID {
			n.IsRead = true
			return nil
		}
	}
	return shared.ErrNotFound
}

// Unit Tests

func TestProductService_CreateProduct(t *testing.T) {
	tests := []struct {
		name    string
		req     CreateProductRequest
		wantErr bool
	}{
		{
			name: "valid product",
			req: CreateProductRequest{
				Name:        "Test Product",
				Description: "A test product",
				Price:       29.99,
			},
			wantErr: false,
		},
		{
			name: "empty name",
			req: CreateProductRequest{
				Name:        "",
				Description: "A test product",
				Price:       29.99,
			},
			wantErr: true,
		},
		{
			name: "negative price",
			req: CreateProductRequest{
				Name:        "Test Product",
				Description: "A test product",
				Price:       -10.00,
			},
			wantErr: true,
		},
		{
			name: "zero price is valid",
			req: CreateProductRequest{
				Name:        "Free Product",
				Description: "A free product",
				Price:       0,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewMockProductRepository()
			notificationService := NewMockNotificationService()
			service := NewProductService(repo, notificationService)

			userID := uuid.New() // Mock user ID for admin
			product, err := service.CreateProduct(context.Background(), tt.req, userID)
			if tt.wantErr {
				if err == nil {
					t.Error("CreateProduct() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("CreateProduct() unexpected error = %v", err)
				return
			}
			if product.Name != tt.req.Name {
				t.Errorf("CreateProduct() name = %v, want %v", product.Name, tt.req.Name)
			}
			if product.Price != tt.req.Price {
				t.Errorf("CreateProduct() price = %v, want %v", product.Price, tt.req.Price)
			}
		})
	}
}

func TestProductService_GetProduct(t *testing.T) {
	repo := NewMockProductRepository()
	notificationService := NewMockNotificationService()
	service := NewProductService(repo, notificationService)

	// Create a product
	userID := uuid.New()
	created, err := service.CreateProduct(context.Background(), CreateProductRequest{
		Name:  "Test Product",
		Price: 19.99,
	}, userID)
	if err != nil {
		t.Fatalf("Failed to create product: %v", err)
	}

	// Get existing product
	product, err := service.GetProduct(context.Background(), created.ID)
	if err != nil {
		t.Errorf("GetProduct() error = %v", err)
	}
	if product.ID != created.ID {
		t.Errorf("GetProduct() ID = %v, want %v", product.ID, created.ID)
	}

	// Get non-existing product
	_, err = service.GetProduct(context.Background(), uuid.New())
	if err != shared.ErrNotFound {
		t.Errorf("GetProduct() error = %v, want %v", err, shared.ErrNotFound)
	}
}

func TestProductService_UpdateProduct(t *testing.T) {
	repo := NewMockProductRepository()
	notificationService := NewMockNotificationService()
	service := NewProductService(repo, notificationService)

	// Create a product
	userID := uuid.New()
	created, err := service.CreateProduct(context.Background(), CreateProductRequest{
		Name:  "Original Name",
		Price: 19.99,
	}, userID)
	if err != nil {
		t.Fatalf("Failed to create product: %v", err)
	}

	// Update product
	newName := "Updated Name"
	newPrice := 29.99
	updated, err := service.UpdateProduct(context.Background(), created.ID, UpdateProductRequest{
		Name:  &newName,
		Price: &newPrice,
	})
	if err != nil {
		t.Errorf("UpdateProduct() error = %v", err)
	}
	if updated.Name != newName {
		t.Errorf("UpdateProduct() name = %v, want %v", updated.Name, newName)
	}
	if updated.Price != newPrice {
		t.Errorf("UpdateProduct() price = %v, want %v", updated.Price, newPrice)
	}

	// Update non-existing product
	_, err = service.UpdateProduct(context.Background(), uuid.New(), UpdateProductRequest{
		Name: &newName,
	})
	if err != shared.ErrNotFound {
		t.Errorf("UpdateProduct() error = %v, want %v", err, shared.ErrNotFound)
	}
}

func TestProductService_DeleteProduct(t *testing.T) {
	repo := NewMockProductRepository()
	notificationService := NewMockNotificationService()
	service := NewProductService(repo, notificationService)

	// Create a product
	userID := uuid.New()
	created, err := service.CreateProduct(context.Background(), CreateProductRequest{
		Name:  "To Delete",
		Price: 19.99,
	}, userID)
	if err != nil {
		t.Fatalf("Failed to create product: %v", err)
	}

	// Delete product
	err = service.DeleteProduct(context.Background(), created.ID)
	if err != nil {
		t.Errorf("DeleteProduct() error = %v", err)
	}

	// Verify deletion
	_, err = service.GetProduct(context.Background(), created.ID)
	if err != shared.ErrNotFound {
		t.Errorf("GetProduct() after delete error = %v, want %v", err, shared.ErrNotFound)
	}

	// Delete non-existing product
	err = service.DeleteProduct(context.Background(), uuid.New())
	if err != shared.ErrNotFound {
		t.Errorf("DeleteProduct() error = %v, want %v", err, shared.ErrNotFound)
	}
}

func TestProductService_ListProducts(t *testing.T) {
	repo := NewMockProductRepository()
	notificationService := NewMockNotificationService()
	service := NewProductService(repo, notificationService)

	// Create multiple products
	for i := 0; i < 5; i++ {
		userID := uuid.New()
		_, err := service.CreateProduct(context.Background(), CreateProductRequest{
			Name:  "Product",
			Price: float64(i * 10),
		}, userID)
		if err != nil {
			t.Fatalf("Failed to create product: %v", err)
		}
	}

	// List all products
	products, total, err := service.ListProducts(context.Background(), 1, 10, nil)
	if err != nil {
		t.Errorf("ListProducts() error = %v", err)
	}
	if total != 5 {
		t.Errorf("ListProducts() total = %v, want %v", total, 5)
	}
	if len(products) != 5 {
		t.Errorf("ListProducts() count = %v, want %v", len(products), 5)
	}

	// Test pagination
	products, _, err = service.ListProducts(context.Background(), 1, 2, nil)
	if err != nil {
		t.Errorf("ListProducts() error = %v", err)
	}
	if len(products) != 2 {
		t.Errorf("ListProducts() with limit 2 count = %v, want %v", len(products), 2)
	}
}

// Property 6: Product CRUD Round-Trip
func TestProperty_ProductCRUDRoundTrip(t *testing.T) {
	repo := NewMockProductRepository()
	notificationService := NewMockNotificationService()
	service := NewProductService(repo, notificationService)

	// Create
	req := CreateProductRequest{
		Name:        "Round Trip Product",
		Description: "Testing round trip",
		Price:       99.99,
	}
	userID := uuid.New()
	created, err := service.CreateProduct(context.Background(), req, userID)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if created.ID == uuid.Nil {
		t.Error("Created product should have UUID")
	}

	// Read
	retrieved, err := service.GetProduct(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if retrieved.Name != req.Name {
		t.Errorf("Retrieved name = %v, want %v", retrieved.Name, req.Name)
	}
	if retrieved.Price != req.Price {
		t.Errorf("Retrieved price = %v, want %v", retrieved.Price, req.Price)
	}

	// Update
	newName := "Updated Product"
	updated, err := service.UpdateProduct(context.Background(), created.ID, UpdateProductRequest{
		Name: &newName,
	})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if updated.Name != newName {
		t.Errorf("Updated name = %v, want %v", updated.Name, newName)
	}

	// Delete
	err = service.DeleteProduct(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify deletion
	_, err = service.GetProduct(context.Background(), created.ID)
	if err != shared.ErrNotFound {
		t.Errorf("Get after delete error = %v, want %v", err, shared.ErrNotFound)
	}
}

// TestProductService_UpdateProduct_ErrorPaths tests UpdateProduct error paths
func TestProductService_UpdateProduct_ErrorPaths(t *testing.T) {
	repo := NewMockProductRepository()
	notificationService := NewMockNotificationService()
	service := NewProductService(repo, notificationService)

	// Create a product
	userID := uuid.New()
	created, err := service.CreateProduct(context.Background(), CreateProductRequest{
		Name:  "Original Name",
		Price: 19.99,
	}, userID)
	if err != nil {
		t.Fatalf("Failed to create product: %v", err)
	}

	tests := []struct {
		name    string
		req     UpdateProductRequest
		wantErr error
	}{
		{
			name:    "empty name",
			req:     UpdateProductRequest{Name: stringPtr("")},
			wantErr: ErrInvalidProductName,
		},
		{
			name:    "negative price",
			req:     UpdateProductRequest{Price: floatPtr(-10.0)},
			wantErr: ErrInvalidProductPrice,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.UpdateProduct(context.Background(), created.ID, tt.req)
			if err != tt.wantErr {
				t.Errorf("UpdateProduct() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func stringPtr(s string) *string {
	return &s
}

func floatPtr(f float64) *float64 {
	return &f
}

// FailingNotificationService is a mock that always fails
type FailingNotificationService struct{}

func (m *FailingNotificationService) CreateNotification(ctx context.Context, userID uuid.UUID, title, message string) (*notifications.Notification, error) {
	return nil, shared.ErrBadRequest
}

func (m *FailingNotificationService) CreateNotificationTx(ctx context.Context, tx *gorm.DB, userID uuid.UUID, title, message string) (*notifications.Notification, error) {
	return nil, shared.ErrBadRequest
}

func (m *FailingNotificationService) GetNotifications(ctx context.Context, userID uuid.UUID) ([]notifications.Notification, error) {
	return nil, shared.ErrBadRequest
}

func (m *FailingNotificationService) MarkAsRead(ctx context.Context, userID uuid.UUID, notificationID uuid.UUID) error {
	return shared.ErrBadRequest
}

// TestProductService_NotificationErrors tests that notification errors don't fail operations
func TestProductService_NotificationErrors(t *testing.T) {
	repo := NewMockProductRepository()
	failingNotificationService := &FailingNotificationService{}
	service := NewProductService(repo, failingNotificationService)

	// CreateProduct should succeed even if notification fails
	userID := uuid.New()
	product, err := service.CreateProduct(context.Background(), CreateProductRequest{
		Name:  "Test Product",
		Price: 19.99,
	}, userID)
	if err != nil {
		t.Errorf("CreateProduct() should succeed even if notification fails, got error: %v", err)
	}
	if product == nil {
		t.Error("CreateProduct() should return product even if notification fails")
	}
}