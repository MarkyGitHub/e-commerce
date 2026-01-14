package inventory

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/mschl/ecommerce-api/internal/products"
	"github.com/mschl/ecommerce-api/internal/shared"
	"github.com/mschl/ecommerce-api/internal/testutil"
)

func setupInventoryTestDB(t *testing.T) *gorm.DB {
	testDB := testutil.SetupTestDB(t)
	if err := testDB.AutoMigrate(&products.Product{}, &Inventory{}); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}
	return testDB
}

func TestInventoryRepository_Create(t *testing.T) {
	testDB := setupInventoryTestDB(t)
	productRepo := products.NewProductRepository(testDB)
	inventoryRepo := NewInventoryRepository(testDB)
	ctx := context.Background()

	// Create test product
	product := &products.Product{
		ID:   uuid.New(),
		Name: "Test Product",
		Price: 19.99,
	}
	if err := productRepo.Create(ctx, product); err != nil {
		t.Fatalf("Failed to create test product: %v", err)
	}

	tests := []struct {
		name      string
		inventory *Inventory
		wantErr   bool
	}{
		{
			name: "create valid inventory",
			inventory: &Inventory{
				ID:        uuid.New(),
				ProductID: product.ID,
				Quantity:  100,
			},
			wantErr: false,
		},
		{
			name: "create inventory with zero quantity",
			inventory: &Inventory{
				ID:        uuid.New(),
				ProductID: uuid.New(), // Different product
				Quantity:  0,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create product for inventory if needed
			if tt.inventory.ProductID != product.ID {
				testProduct := &products.Product{
					ID:   tt.inventory.ProductID,
					Name: "Test Product 2",
					Price: 9.99,
				}
				if err := productRepo.Create(ctx, testProduct); err != nil {
					t.Fatalf("Failed to create test product: %v", err)
				}
			}

			err := inventoryRepo.Create(ctx, tt.inventory)
			if (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				// Verify inventory was created
				found, err := inventoryRepo.FindByProductID(ctx, tt.inventory.ProductID)
				if err != nil {
					t.Errorf("FindByProductID() error = %v", err)
				}
				if found.Quantity != tt.inventory.Quantity {
					t.Errorf("FindByProductID() quantity = %d, want %d", found.Quantity, tt.inventory.Quantity)
				}
			}
		})
	}
}

func TestInventoryRepository_FindByProductID(t *testing.T) {
	testDB := setupInventoryTestDB(t)
	productRepo := products.NewProductRepository(testDB)
	inventoryRepo := NewInventoryRepository(testDB)
	ctx := context.Background()

	// Create test product and inventory
	product := &products.Product{
		ID:   uuid.New(),
		Name: "Find Product",
		Price: 19.99,
	}
	if err := productRepo.Create(ctx, product); err != nil {
		t.Fatalf("Failed to create test product: %v", err)
	}

	inventory := &Inventory{
		ID:        uuid.New(),
		ProductID: product.ID,
		Quantity:  50,
	}
	if err := inventoryRepo.Create(ctx, inventory); err != nil {
		t.Fatalf("Failed to create test inventory: %v", err)
	}

	tests := []struct {
		name      string
		productID uuid.UUID
		wantErr   error
		wantQty   int
	}{
		{
			name:      "find existing inventory",
			productID: product.ID,
			wantErr:   nil,
			wantQty:   50,
		},
		{
			name:      "inventory not found",
			productID: uuid.New(),
			wantErr:   shared.ErrNotFound,
			wantQty:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			found, err := inventoryRepo.FindByProductID(ctx, tt.productID)
			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Errorf("FindByProductID() error = %v, want %v", err, tt.wantErr)
				}
				if found != nil {
					t.Errorf("FindByProductID() found = %v, want nil", found)
				}
			} else {
				if err != nil {
					t.Errorf("FindByProductID() error = %v, want nil", err)
				}
				if found.Quantity != tt.wantQty {
					t.Errorf("FindByProductID() quantity = %d, want %d", found.Quantity, tt.wantQty)
				}
			}
		})
	}
}

func TestInventoryRepository_Update(t *testing.T) {
	testDB := setupInventoryTestDB(t)
	productRepo := products.NewProductRepository(testDB)
	inventoryRepo := NewInventoryRepository(testDB)
	ctx := context.Background()

	// Create test product and inventory
	product := &products.Product{
		ID:   uuid.New(),
		Name: "Update Product",
		Price: 19.99,
	}
	if err := productRepo.Create(ctx, product); err != nil {
		t.Fatalf("Failed to create test product: %v", err)
	}

	inventory := &Inventory{
		ID:        uuid.New(),
		ProductID: product.ID,
		Quantity:  100,
	}
	if err := inventoryRepo.Create(ctx, inventory); err != nil {
		t.Fatalf("Failed to create test inventory: %v", err)
	}

	// Update inventory
	inventory.Quantity = 200
	if err := inventoryRepo.Update(ctx, inventory); err != nil {
		t.Errorf("Update() error = %v, want nil", err)
	}

	// Verify update
	updated, err := inventoryRepo.FindByProductID(ctx, product.ID)
	if err != nil {
		t.Fatalf("FindByProductID() error = %v", err)
	}
	if updated.Quantity != 200 {
		t.Errorf("Update() quantity = %d, want 200", updated.Quantity)
	}
}

func TestInventoryRepository_DecrementQuantity(t *testing.T) {
	testDB := setupInventoryTestDB(t)
	productRepo := products.NewProductRepository(testDB)
	inventoryRepo := NewInventoryRepository(testDB)
	ctx := context.Background()

	// Create test product and inventory
	product := &products.Product{
		ID:   uuid.New(),
		Name: "Decrement Product",
		Price: 19.99,
	}
	if err := productRepo.Create(ctx, product); err != nil {
		t.Fatalf("Failed to create test product: %v", err)
	}

	inventory := &Inventory{
		ID:        uuid.New(),
		ProductID: product.ID,
		Quantity:  100,
	}
	if err := inventoryRepo.Create(ctx, inventory); err != nil {
		t.Fatalf("Failed to create test inventory: %v", err)
	}

	tests := []struct {
		name      string
		amount    int
		wantErr   error
		wantQty   int
	}{
		{
			name:    "decrement valid amount",
			amount:  10,
			wantErr: nil,
			wantQty: 90,
		},
		{
			name:    "decrement more",
			amount:  20,
			wantErr: nil,
			wantQty: 70,
		},
		{
			name:    "decrement insufficient inventory",
			amount:  100,
			wantErr: shared.ErrInsufficientInventory,
			wantQty: 70, // Should remain unchanged
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := inventoryRepo.DecrementQuantity(ctx, product.ID, tt.amount)
			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Errorf("DecrementQuantity() error = %v, want %v", err, tt.wantErr)
				}
			} else {
				if err != nil {
					t.Errorf("DecrementQuantity() error = %v, want nil", err)
				}
				// Verify quantity
				found, err := inventoryRepo.FindByProductID(ctx, product.ID)
				if err != nil {
					t.Fatalf("FindByProductID() error = %v", err)
				}
				if found.Quantity != tt.wantQty {
					t.Errorf("DecrementQuantity() quantity = %d, want %d", found.Quantity, tt.wantQty)
				}
			}
		})
	}
}

func TestInventoryRepository_DecrementQuantityTx(t *testing.T) {
	testDB := setupInventoryTestDB(t)
	productRepo := products.NewProductRepository(testDB)
	inventoryRepo := NewInventoryRepository(testDB)
	ctx := context.Background()

	// Create test product and inventory
	product := &products.Product{
		ID:   uuid.New(),
		Name: "Tx Product",
		Price: 19.99,
	}
	if err := productRepo.Create(ctx, product); err != nil {
		t.Fatalf("Failed to create test product: %v", err)
	}

	inventory := &Inventory{
		ID:        uuid.New(),
		ProductID: product.ID,
		Quantity:  100,
	}
	if err := inventoryRepo.Create(ctx, inventory); err != nil {
		t.Fatalf("Failed to create test inventory: %v", err)
	}

	// Test transaction-based decrement
	tx := testDB.Begin()
	defer tx.Rollback()

	err := inventoryRepo.DecrementQuantityTx(ctx, tx, product.ID, 30)
	if err != nil {
		t.Errorf("DecrementQuantityTx() error = %v, want nil", err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		t.Fatalf("Commit() error = %v", err)
	}

	// Verify quantity was decremented
	found, err := inventoryRepo.FindByProductID(ctx, product.ID)
	if err != nil {
		t.Fatalf("FindByProductID() error = %v", err)
	}
	if found.Quantity != 70 {
		t.Errorf("DecrementQuantityTx() quantity = %d, want 70", found.Quantity)
	}
}

// TestInventoryRepository_Create_AutoGenerateID tests Create with nil ID (auto-generation)
func TestInventoryRepository_Create_AutoGenerateID(t *testing.T) {
	testDB := setupInventoryTestDB(t)
	productRepo := products.NewProductRepository(testDB)
	inventoryRepo := NewInventoryRepository(testDB)
	ctx := context.Background()

	// Create test product
	product := &products.Product{
		ID:   uuid.New(),
		Name: "Auto ID Product",
		Price: 19.99,
	}
	if err := productRepo.Create(ctx, product); err != nil {
		t.Fatalf("Failed to create test product: %v", err)
	}

	// Create inventory with nil ID (should be auto-generated)
	inventory := &Inventory{
		ID:        uuid.Nil, // Should be auto-generated
		ProductID: product.ID,
		Quantity:  100,
	}
	if err := inventoryRepo.Create(ctx, inventory); err != nil {
		t.Errorf("Create() error = %v, want nil", err)
	}
	if inventory.ID == uuid.Nil {
		t.Error("Create() should auto-generate ID when nil")
	}
}

// TestInventoryRepository_GetDB tests GetDB method
func TestInventoryRepository_GetDB(t *testing.T) {
	testDB := setupInventoryTestDB(t)
	inventoryRepo := NewInventoryRepository(testDB)

	// GetDB should return the underlying database connection
	db := inventoryRepo.(*InventoryRepositoryImpl).GetDB()
	if db == nil {
		t.Error("GetDB() should return non-nil database connection")
	}
	if db != testDB {
		t.Error("GetDB() should return the same database connection")
	}
}

// TestInventoryRepository_DecrementQuantityTx_InsufficientInventory tests DecrementQuantityTx with insufficient inventory
func TestInventoryRepository_DecrementQuantityTx_InsufficientInventory(t *testing.T) {
	testDB := setupInventoryTestDB(t)
	productRepo := products.NewProductRepository(testDB)
	inventoryRepo := NewInventoryRepository(testDB)
	ctx := context.Background()

	// Create test product and inventory
	product := &products.Product{
		ID:   uuid.New(),
		Name: "Insufficient Product",
		Price: 19.99,
	}
	if err := productRepo.Create(ctx, product); err != nil {
		t.Fatalf("Failed to create test product: %v", err)
	}

	inventory := &Inventory{
		ID:        uuid.New(),
		ProductID: product.ID,
		Quantity:  10,
	}
	if err := inventoryRepo.Create(ctx, inventory); err != nil {
		t.Fatalf("Failed to create test inventory: %v", err)
	}

	// Test transaction-based decrement with insufficient inventory
	tx := testDB.Begin()
	defer tx.Rollback()

	err := inventoryRepo.DecrementQuantityTx(ctx, tx, product.ID, 20)
	if err != shared.ErrInsufficientInventory {
		t.Errorf("DecrementQuantityTx() error = %v, want %v", err, shared.ErrInsufficientInventory)
	}

	// Rollback transaction
	tx.Rollback()

	// Verify quantity was not changed
	found, err := inventoryRepo.FindByProductID(ctx, product.ID)
	if err != nil {
		t.Fatalf("FindByProductID() error = %v", err)
	}
	if found.Quantity != 10 {
		t.Errorf("DecrementQuantityTx() quantity = %d, want 10 (unchanged)", found.Quantity)
	}
}

// TestInventoryRepository_DecrementQuantityTx_ExactQuantity tests DecrementQuantityTx with exact quantity
func TestInventoryRepository_DecrementQuantityTx_ExactQuantity(t *testing.T) {
	testDB := setupInventoryTestDB(t)
	productRepo := products.NewProductRepository(testDB)
	inventoryRepo := NewInventoryRepository(testDB)
	ctx := context.Background()

	// Create test product and inventory
	product := &products.Product{
		ID:   uuid.New(),
		Name: "Exact Quantity Product",
		Price: 19.99,
	}
	if err := productRepo.Create(ctx, product); err != nil {
		t.Fatalf("Failed to create test product: %v", err)
	}

	inventory := &Inventory{
		ID:        uuid.New(),
		ProductID: product.ID,
		Quantity:  30,
	}
	if err := inventoryRepo.Create(ctx, inventory); err != nil {
		t.Fatalf("Failed to create test inventory: %v", err)
	}

	// Test transaction-based decrement with exact quantity
	tx := testDB.Begin()
	defer tx.Rollback()

	err := inventoryRepo.DecrementQuantityTx(ctx, tx, product.ID, 30)
	if err != nil {
		t.Errorf("DecrementQuantityTx() error = %v, want nil", err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		t.Fatalf("Commit() error = %v", err)
	}

	// Verify quantity was decremented to zero
	found, err := inventoryRepo.FindByProductID(ctx, product.ID)
	if err != nil {
		t.Fatalf("FindByProductID() error = %v", err)
	}
	if found.Quantity != 0 {
		t.Errorf("DecrementQuantityTx() quantity = %d, want 0", found.Quantity)
	}
}

// TestInventoryRepository_DecrementQuantity_NonTx tests DecrementQuantity (non-transactional wrapper)
func TestInventoryRepository_DecrementQuantity_NonTx(t *testing.T) {
	testDB := setupInventoryTestDB(t)
	productRepo := products.NewProductRepository(testDB)
	inventoryRepo := NewInventoryRepository(testDB)
	ctx := context.Background()

	// Create test product and inventory
	product := &products.Product{
		ID:   uuid.New(),
		Name: "Decrement Product",
		Price: 19.99,
	}
	if err := productRepo.Create(ctx, product); err != nil {
		t.Fatalf("Failed to create test product: %v", err)
	}

	inventory := &Inventory{
		ID:        uuid.New(),
		ProductID: product.ID,
		Quantity:  50,
	}
	if err := inventoryRepo.Create(ctx, inventory); err != nil {
		t.Fatalf("Failed to create test inventory: %v", err)
	}

	// Test DecrementQuantity (non-transactional)
	err := inventoryRepo.DecrementQuantity(ctx, product.ID, 20)
	if err != nil {
		t.Errorf("DecrementQuantity() error = %v, want nil", err)
	}

	// Verify quantity was decremented
	found, err := inventoryRepo.FindByProductID(ctx, product.ID)
	if err != nil {
		t.Fatalf("FindByProductID() error = %v", err)
	}
	if found.Quantity != 30 {
		t.Errorf("DecrementQuantity() quantity = %d, want 30", found.Quantity)
	}
}