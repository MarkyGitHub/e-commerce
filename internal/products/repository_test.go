package products

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/mschl/ecommerce-api/internal/shared"
	"github.com/mschl/ecommerce-api/internal/testutil"
)

func setupProductsTestDB(t *testing.T) *gorm.DB {
	testDB := testutil.SetupTestDB(t)
	if err := testDB.AutoMigrate(&Product{}); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}
	return testDB
}

func TestProductRepository_Create(t *testing.T) {
	testDB := setupProductsTestDB(t)
	repo := NewProductRepository(testDB)
	ctx := context.Background()

	tests := []struct {
		name    string
		product *Product
		wantErr bool
	}{
		{
			name: "create valid product",
			product: &Product{
				ID:          uuid.New(),
				Name:        "Test Product",
				Description: "Test Description",
				Price:       19.99,
			},
			wantErr: false,
		},
		{
			name: "create product without description",
			product: &Product{
				ID:   uuid.New(),
				Name: "Simple Product",
				Price: 9.99,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.Create(ctx, tt.product)
			if (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				// Verify product was created
				found, err := repo.FindByID(ctx, tt.product.ID)
				if err != nil {
					t.Errorf("FindByID() error = %v", err)
				}
				if found.Name != tt.product.Name {
					t.Errorf("FindByID() name = %s, want %s", found.Name, tt.product.Name)
				}
			}
		})
	}
}

func TestProductRepository_FindByID(t *testing.T) {
	testDB := setupProductsTestDB(t)
	repo := NewProductRepository(testDB)
	ctx := context.Background()

	// Create test product
	product := &Product{
		ID:          uuid.New(),
		Name:        "Find Product",
		Description: "Find Description",
		Price:       29.99,
	}
	if err := repo.Create(ctx, product); err != nil {
		t.Fatalf("Failed to create test product: %v", err)
	}

	tests := []struct {
		name    string
		id      uuid.UUID
		wantErr error
		wantID  uuid.UUID
	}{
		{
			name:    "find existing product",
			id:      product.ID,
			wantErr: nil,
			wantID:  product.ID,
		},
		{
			name:    "product not found",
			id:      uuid.New(),
			wantErr: shared.ErrNotFound,
			wantID:  uuid.Nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			found, err := repo.FindByID(ctx, tt.id)
			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Errorf("FindByID() error = %v, want %v", err, tt.wantErr)
				}
				if found != nil {
					t.Errorf("FindByID() found = %v, want nil", found)
				}
			} else {
				if err != nil {
					t.Errorf("FindByID() error = %v, want nil", err)
				}
				if found.ID != tt.wantID {
					t.Errorf("FindByID() id = %v, want %v", found.ID, tt.wantID)
				}
			}
		})
	}
}

func TestProductRepository_FindAll(t *testing.T) {
	testDB := setupProductsTestDB(t)
	repo := NewProductRepository(testDB)
	ctx := context.Background()

	// Create multiple test products
	products := []*Product{
		{ID: uuid.New(), Name: "Product 1", Price: 10.00},
		{ID: uuid.New(), Name: "Product 2", Price: 20.00},
		{ID: uuid.New(), Name: "Product 3", Price: 30.00},
		{ID: uuid.New(), Name: "Product 4", Price: 40.00},
		{ID: uuid.New(), Name: "Product 5", Price: 50.00},
	}

	for _, p := range products {
		if err := repo.Create(ctx, p); err != nil {
			t.Fatalf("Failed to create test product: %v", err)
		}
	}

	tests := []struct {
		name       string
		offset     int
		limit      int
		wantCount  int
		wantTotal  int64
	}{
		{
			name:      "first page",
			offset:    0,
			limit:     2,
			wantCount: 2,
			wantTotal: 5,
		},
		{
			name:      "second page",
			offset:    2,
			limit:     2,
			wantCount: 2,
			wantTotal: 5,
		},
		{
			name:      "last page",
			offset:    4,
			limit:     2,
			wantCount: 1,
			wantTotal: 5,
		},
		{
			name:      "limit exceeds total",
			offset:    0,
			limit:     10,
			wantCount: 5,
			wantTotal: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, total, err := repo.FindAll(ctx, tt.offset, tt.limit, nil)
			if err != nil {
				t.Errorf("FindAll() error = %v, want nil", err)
			}
			if int64(len(results)) != int64(tt.wantCount) {
				t.Errorf("FindAll() count = %d, want %d", len(results), tt.wantCount)
			}
			if total != tt.wantTotal {
				t.Errorf("FindAll() total = %d, want %d", total, tt.wantTotal)
			}
		})
	}
}

func TestProductRepository_Update(t *testing.T) {
	testDB := setupProductsTestDB(t)
	repo := NewProductRepository(testDB)
	ctx := context.Background()

	// Create test product
	product := &Product{
		ID:          uuid.New(),
		Name:        "Original Name",
		Description: "Original Description",
		Price:       19.99,
	}
	if err := repo.Create(ctx, product); err != nil {
		t.Fatalf("Failed to create test product: %v", err)
	}

	// Update product
	product.Name = "Updated Name"
	product.Description = "Updated Description"
	product.Price = 29.99

	if err := repo.Update(ctx, product); err != nil {
		t.Errorf("Update() error = %v, want nil", err)
	}

	// Verify update
	updated, err := repo.FindByID(ctx, product.ID)
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}
	if updated.Name != "Updated Name" {
		t.Errorf("Update() name = %s, want Updated Name", updated.Name)
	}
	if updated.Price != 29.99 {
		t.Errorf("Update() price = %f, want 29.99", updated.Price)
	}

	// Test update non-existent product
	nonExistent := &Product{
		ID:    uuid.New(),
		Name:  "Non-existent",
		Price: 99.99,
	}
	err = repo.Update(ctx, nonExistent)
	if err != shared.ErrNotFound {
		t.Errorf("Update() non-existent error = %v, want %v", err, shared.ErrNotFound)
	}
}

func TestProductRepository_Delete(t *testing.T) {
	testDB := setupProductsTestDB(t)
	repo := NewProductRepository(testDB)
	ctx := context.Background()

	// Create test product
	product := &Product{
		ID:   uuid.New(),
		Name: "Delete Product",
		Price: 19.99,
	}
	if err := repo.Create(ctx, product); err != nil {
		t.Fatalf("Failed to create test product: %v", err)
	}

	// Verify product exists
	_, err := repo.FindByID(ctx, product.ID)
	if err != nil {
		t.Fatalf("Product should exist before delete: %v", err)
	}

	// Delete product
	if err := repo.Delete(ctx, product.ID); err != nil {
		t.Errorf("Delete() error = %v, want nil", err)
	}

	// Verify product is deleted
	_, err = repo.FindByID(ctx, product.ID)
	if err != shared.ErrNotFound {
		t.Errorf("Delete() product should not exist after delete, got error = %v", err)
	}

	// Test delete non-existent product
	err = repo.Delete(ctx, uuid.New())
	if err != shared.ErrNotFound {
		t.Errorf("Delete() non-existent error = %v, want %v", err, shared.ErrNotFound)
	}
}
