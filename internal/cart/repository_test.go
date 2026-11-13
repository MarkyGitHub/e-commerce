package cart

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/MarkyGitHub/e-commerce/internal/auth"
	"github.com/MarkyGitHub/e-commerce/internal/products"
	"github.com/MarkyGitHub/e-commerce/internal/shared"
	"github.com/MarkyGitHub/e-commerce/internal/testutil"
)

func setupCartTestDB(t *testing.T) *gorm.DB {
	testDB := testutil.SetupTestDB(t)
	if err := testDB.AutoMigrate(&auth.User{}, &products.Product{}, &Cart{}, &CartItem{}); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}
	return testDB
}

func TestCartRepository_FindOrCreateByUserID(t *testing.T) {
	testDB := setupCartTestDB(t)
	userRepo := auth.NewUserRepository(testDB)
	cartRepo := NewCartRepository(testDB)
	ctx := context.Background()

	// Create test user
	user := &auth.User{
		ID:           uuid.New(),
		Email:        "cart@example.com",
		PasswordHash: "hash",
		Username:         "Cart User",
		Role:         auth.RoleCustomer,
	}
	if err := userRepo.Create(ctx, user); err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// First call should create cart
	cart1, err := cartRepo.FindOrCreateByUserID(ctx, user.ID)
	if err != nil {
		t.Errorf("FindOrCreateByUserID() error = %v, want nil", err)
	}
	if cart1 == nil {
		t.Fatal("FindOrCreateByUserID() returned nil")
	}
	if cart1.UserID != user.ID {
		t.Errorf("FindOrCreateByUserID() userID = %v, want %v", cart1.UserID, user.ID)
	}

	// Second call should find existing cart
	cart2, err := cartRepo.FindOrCreateByUserID(ctx, user.ID)
	if err != nil {
		t.Errorf("FindOrCreateByUserID() second call error = %v, want nil", err)
	}
	if cart2.ID != cart1.ID {
		t.Errorf("FindOrCreateByUserID() second call returned different cart ID")
	}
}

func TestCartRepository_FindByUserID(t *testing.T) {
	testDB := setupCartTestDB(t)
	userRepo := auth.NewUserRepository(testDB)
	cartRepo := NewCartRepository(testDB)
	ctx := context.Background()

	// Create test user and cart
	user := &auth.User{
		ID:           uuid.New(),
		Email:        "findcart@example.com",
		PasswordHash: "hash",
		Username:         "Find Cart User",
		Role:         auth.RoleCustomer,
	}
	if err := userRepo.Create(ctx, user); err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	cart, err := cartRepo.FindOrCreateByUserID(ctx, user.ID)
	if err != nil {
		t.Fatalf("Failed to create cart: %v", err)
	}

	tests := []struct {
		name    string
		userID  uuid.UUID
		wantErr error
		wantID  uuid.UUID
	}{
		{
			name:    "find existing cart",
			userID:  user.ID,
			wantErr: nil,
			wantID:  cart.ID,
		},
		{
			name:    "cart not found",
			userID:  uuid.New(),
			wantErr: shared.ErrNotFound,
			wantID:  uuid.Nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			found, err := cartRepo.FindByUserID(ctx, tt.userID)
			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Errorf("FindByUserID() error = %v, want %v", err, tt.wantErr)
				}
				if found != nil {
					t.Errorf("FindByUserID() found = %v, want nil", found)
				}
			} else {
				if err != nil {
					t.Errorf("FindByUserID() error = %v, want nil", err)
				}
				if found.ID != tt.wantID {
					t.Errorf("FindByUserID() id = %v, want %v", found.ID, tt.wantID)
				}
			}
		})
	}
}

func TestCartRepository_Clear(t *testing.T) {
	testDB := setupCartTestDB(t)
	userRepo := auth.NewUserRepository(testDB)
	cartRepo := NewCartRepository(testDB)
	cartItemRepo := NewCartItemRepository(testDB)
	productRepo := products.NewProductRepository(testDB)
	ctx := context.Background()

	// Create test user, product, cart, and items
	user := &auth.User{
		ID:           uuid.New(),
		Email:        "clear@example.com",
		PasswordHash: "hash",
		Username:         "Clear User",
		Role:         auth.RoleCustomer,
	}
	if err := userRepo.Create(ctx, user); err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	product := &products.Product{
		ID:    uuid.New(),
		Name:  "Clear Product",
		Price: 19.99,
	}
	if err := productRepo.Create(ctx, product); err != nil {
		t.Fatalf("Failed to create test product: %v", err)
	}

	cart, err := cartRepo.FindOrCreateByUserID(ctx, user.ID)
	if err != nil {
		t.Fatalf("Failed to create cart: %v", err)
	}

	// Add items
	item1 := &CartItem{
		ID:        uuid.New(),
		CartID:    cart.ID,
		ProductID: product.ID,
		Quantity:  2,
	}
	if err := cartItemRepo.Create(ctx, item1); err != nil {
		t.Fatalf("Failed to create cart item: %v", err)
	}

	// Verify items exist
	items, err := cartItemRepo.FindByCartID(ctx, cart.ID)
	if err != nil {
		t.Fatalf("FindByCartID() error = %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("Expected 1 item, got %d", len(items))
	}

	// Clear cart
	if err := cartRepo.Clear(ctx, cart.ID); err != nil {
		t.Errorf("Clear() error = %v, want nil", err)
	}

	// Verify items are cleared
	items, err = cartItemRepo.FindByCartID(ctx, cart.ID)
	if err != nil {
		t.Fatalf("FindByCartID() after clear error = %v", err)
	}
	if len(items) != 0 {
		t.Errorf("Clear() items count = %d, want 0", len(items))
	}
}

func TestCartItemRepository_Create(t *testing.T) {
	testDB := setupCartTestDB(t)
	userRepo := auth.NewUserRepository(testDB)
	cartRepo := NewCartRepository(testDB)
	cartItemRepo := NewCartItemRepository(testDB)
	productRepo := products.NewProductRepository(testDB)
	ctx := context.Background()

	// Setup
	user := &auth.User{
		ID:           uuid.New(),
		Email:        "item@example.com",
		PasswordHash: "hash",
		Username:         "Item User",
		Role:         auth.RoleCustomer,
	}
	if err := userRepo.Create(ctx, user); err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	product := &products.Product{
		ID:    uuid.New(),
		Name:  "Item Product",
		Price: 19.99,
	}
	if err := productRepo.Create(ctx, product); err != nil {
		t.Fatalf("Failed to create test product: %v", err)
	}

	cart, err := cartRepo.FindOrCreateByUserID(ctx, user.ID)
	if err != nil {
		t.Fatalf("Failed to create cart: %v", err)
	}

	// Create cart item
	item := &CartItem{
		ID:        uuid.New(),
		CartID:    cart.ID,
		ProductID: product.ID,
		Quantity:  3,
	}
	if err := cartItemRepo.Create(ctx, item); err != nil {
		t.Errorf("Create() error = %v, want nil", err)
	}

	// Verify item was created
	found, err := cartItemRepo.FindByID(ctx, item.ID)
	if err != nil {
		t.Errorf("FindByID() error = %v", err)
	}
	if found.Quantity != 3 {
		t.Errorf("FindByID() quantity = %d, want 3", found.Quantity)
	}
}

func TestCartItemRepository_FindByCartID(t *testing.T) {
	testDB := setupCartTestDB(t)
	userRepo := auth.NewUserRepository(testDB)
	cartRepo := NewCartRepository(testDB)
	cartItemRepo := NewCartItemRepository(testDB)
	productRepo := products.NewProductRepository(testDB)
	ctx := context.Background()

	// Setup
	user := &auth.User{
		ID:           uuid.New(),
		Email:        "finditems@example.com",
		PasswordHash: "hash",
		Username:         "Find Items User",
		Role:         auth.RoleCustomer,
	}
	if err := userRepo.Create(ctx, user); err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	product1 := &products.Product{ID: uuid.New(), Name: "Product 1", Price: 10.00}
	product2 := &products.Product{ID: uuid.New(), Name: "Product 2", Price: 20.00}
	if err := productRepo.Create(ctx, product1); err != nil {
		t.Fatalf("Failed to create product1: %v", err)
	}
	if err := productRepo.Create(ctx, product2); err != nil {
		t.Fatalf("Failed to create product2: %v", err)
	}

	cart, err := cartRepo.FindOrCreateByUserID(ctx, user.ID)
	if err != nil {
		t.Fatalf("Failed to create cart: %v", err)
	}

	// Create multiple items
	item1 := &CartItem{ID: uuid.New(), CartID: cart.ID, ProductID: product1.ID, Quantity: 1}
	item2 := &CartItem{ID: uuid.New(), CartID: cart.ID, ProductID: product2.ID, Quantity: 2}
	if err := cartItemRepo.Create(ctx, item1); err != nil {
		t.Fatalf("Failed to create item1: %v", err)
	}
	if err := cartItemRepo.Create(ctx, item2); err != nil {
		t.Fatalf("Failed to create item2: %v", err)
	}

	// Find all items
	items, err := cartItemRepo.FindByCartID(ctx, cart.ID)
	if err != nil {
		t.Errorf("FindByCartID() error = %v, want nil", err)
	}
	if len(items) != 2 {
		t.Errorf("FindByCartID() count = %d, want 2", len(items))
	}
}

func TestCartItemRepository_Update(t *testing.T) {
	testDB := setupCartTestDB(t)
	userRepo := auth.NewUserRepository(testDB)
	cartRepo := NewCartRepository(testDB)
	cartItemRepo := NewCartItemRepository(testDB)
	productRepo := products.NewProductRepository(testDB)
	ctx := context.Background()

	// Setup
	user := &auth.User{
		ID:           uuid.New(),
		Email:        "update@example.com",
		PasswordHash: "hash",
		Username:         "Update User",
		Role:         auth.RoleCustomer,
	}
	if err := userRepo.Create(ctx, user); err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	product := &products.Product{ID: uuid.New(), Name: "Update Product", Price: 19.99}
	if err := productRepo.Create(ctx, product); err != nil {
		t.Fatalf("Failed to create test product: %v", err)
	}

	cart, err := cartRepo.FindOrCreateByUserID(ctx, user.ID)
	if err != nil {
		t.Fatalf("Failed to create cart: %v", err)
	}

	item := &CartItem{ID: uuid.New(), CartID: cart.ID, ProductID: product.ID, Quantity: 1}
	if err := cartItemRepo.Create(ctx, item); err != nil {
		t.Fatalf("Failed to create cart item: %v", err)
	}

	// Update item
	item.Quantity = 5
	if err := cartItemRepo.Update(ctx, item); err != nil {
		t.Errorf("Update() error = %v, want nil", err)
	}

	// Verify update
	updated, err := cartItemRepo.FindByID(ctx, item.ID)
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}
	if updated.Quantity != 5 {
		t.Errorf("Update() quantity = %d, want 5", updated.Quantity)
	}
}

func TestCartItemRepository_Delete(t *testing.T) {
	testDB := setupCartTestDB(t)
	userRepo := auth.NewUserRepository(testDB)
	cartRepo := NewCartRepository(testDB)
	cartItemRepo := NewCartItemRepository(testDB)
	productRepo := products.NewProductRepository(testDB)
	ctx := context.Background()

	// Setup
	user := &auth.User{
		ID:           uuid.New(),
		Email:        "delete@example.com",
		PasswordHash: "hash",
		Username:         "Delete User",
		Role:         auth.RoleCustomer,
	}
	if err := userRepo.Create(ctx, user); err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	product := &products.Product{ID: uuid.New(), Name: "Delete Product", Price: 19.99}
	if err := productRepo.Create(ctx, product); err != nil {
		t.Fatalf("Failed to create test product: %v", err)
	}

	cart, err := cartRepo.FindOrCreateByUserID(ctx, user.ID)
	if err != nil {
		t.Fatalf("Failed to create cart: %v", err)
	}

	item := &CartItem{ID: uuid.New(), CartID: cart.ID, ProductID: product.ID, Quantity: 1}
	if err := cartItemRepo.Create(ctx, item); err != nil {
		t.Fatalf("Failed to create cart item: %v", err)
	}

	// Verify item exists
	_, err = cartItemRepo.FindByID(ctx, item.ID)
	if err != nil {
		t.Fatalf("Item should exist before delete: %v", err)
	}

	// Delete item
	if err := cartItemRepo.Delete(ctx, item.ID); err != nil {
		t.Errorf("Delete() error = %v, want nil", err)
	}

	// Verify item is deleted
	_, err = cartItemRepo.FindByID(ctx, item.ID)
	if err != shared.ErrNotFound {
		t.Errorf("Delete() item should not exist after delete, got error = %v", err)
	}

	// Test delete non-existent item
	err = cartItemRepo.Delete(ctx, uuid.New())
	if err != shared.ErrNotFound {
		t.Errorf("Delete() non-existent error = %v, want %v", err, shared.ErrNotFound)
	}
}

func TestCartItemRepository_FindByCartAndProduct(t *testing.T) {
	testDB := setupCartTestDB(t)
	userRepo := auth.NewUserRepository(testDB)
	cartRepo := NewCartRepository(testDB)
	cartItemRepo := NewCartItemRepository(testDB)
	productRepo := products.NewProductRepository(testDB)
	ctx := context.Background()

	// Setup
	user := &auth.User{
		ID:           uuid.New(),
		Email:        "findby@example.com",
		PasswordHash: "hash",
		Username:         "Find By User",
		Role:         auth.RoleCustomer,
	}
	if err := userRepo.Create(ctx, user); err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	product := &products.Product{ID: uuid.New(), Name: "Find By Product", Price: 19.99}
	if err := productRepo.Create(ctx, product); err != nil {
		t.Fatalf("Failed to create test product: %v", err)
	}

	cart, err := cartRepo.FindOrCreateByUserID(ctx, user.ID)
	if err != nil {
		t.Fatalf("Failed to create cart: %v", err)
	}

	item := &CartItem{ID: uuid.New(), CartID: cart.ID, ProductID: product.ID, Quantity: 2}
	if err := cartItemRepo.Create(ctx, item); err != nil {
		t.Fatalf("Failed to create cart item: %v", err)
	}

	tests := []struct {
		name      string
		cartID    uuid.UUID
		productID uuid.UUID
		wantErr   error
		wantQty   int
	}{
		{
			name:      "find existing item",
			cartID:    cart.ID,
			productID: product.ID,
			wantErr:   nil,
			wantQty:   2,
		},
		{
			name:      "item not found",
			cartID:    cart.ID,
			productID: uuid.New(),
			wantErr:   shared.ErrNotFound,
			wantQty:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			found, err := cartItemRepo.FindByCartAndProduct(ctx, tt.cartID, tt.productID)
			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Errorf("FindByCartAndProduct() error = %v, want %v", err, tt.wantErr)
				}
				if found != nil {
					t.Errorf("FindByCartAndProduct() found = %v, want nil", found)
				}
			} else {
				if err != nil {
					t.Errorf("FindByCartAndProduct() error = %v, want nil", err)
				}
				if found.Quantity != tt.wantQty {
					t.Errorf("FindByCartAndProduct() quantity = %d, want %d", found.Quantity, tt.wantQty)
				}
			}
		})
	}
}
