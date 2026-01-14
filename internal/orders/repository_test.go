package orders

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/mschl/ecommerce-api/internal/auth"
	"github.com/mschl/ecommerce-api/internal/products"
	"github.com/mschl/ecommerce-api/internal/shared"
	"github.com/mschl/ecommerce-api/internal/testutil"
)

func setupOrdersTestDB(t *testing.T) *gorm.DB {
	testDB := testutil.SetupTestDB(t)
	if err := testDB.AutoMigrate(&auth.User{}, &products.Product{}, &Order{}, &OrderItem{}); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}
	return testDB
}

func TestOrderRepository_Create(t *testing.T) {
	testDB := setupOrdersTestDB(t)
	userRepo := auth.NewUserRepository(testDB)
	orderRepo := NewOrderRepository(testDB)
	ctx := context.Background()

	// Create test user
	user := &auth.User{
		ID:           uuid.New(),
		Email:        "order@example.com",
		PasswordHash: "hash",
		Username:         "Order User",
		Role:         auth.RoleCustomer,
	}
	if err := userRepo.Create(ctx, user); err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	order := &Order{
		ID:         uuid.New(),
		UserID:     user.ID,
		TotalPrice: 99.99,
		Status:     StatusPending,
	}

	if err := orderRepo.Create(ctx, order); err != nil {
		t.Errorf("Create() error = %v, want nil", err)
	}

	// Verify order was created
	found, err := orderRepo.FindByID(ctx, order.ID)
	if err != nil {
		t.Errorf("FindByID() error = %v", err)
	}
	if found.TotalPrice != 99.99 {
		t.Errorf("FindByID() totalPrice = %f, want 99.99", found.TotalPrice)
	}
}

func TestOrderRepository_CreateTx(t *testing.T) {
	testDB := setupOrdersTestDB(t)
	userRepo := auth.NewUserRepository(testDB)
	orderRepo := NewOrderRepository(testDB)
	ctx := context.Background()

	// Create test user
	user := &auth.User{
		ID:           uuid.New(),
		Email:        "ordertx@example.com",
		PasswordHash: "hash",
		Username:         "Order Tx User",
		Role:         auth.RoleCustomer,
	}
	if err := userRepo.Create(ctx, user); err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Test transaction-based create
	tx := testDB.Begin()
	defer tx.Rollback()

	order := &Order{
		ID:         uuid.New(),
		UserID:     user.ID,
		TotalPrice: 49.99,
		Status:     StatusPending,
	}

	if err := orderRepo.CreateTx(ctx, tx, order); err != nil {
		t.Errorf("CreateTx() error = %v, want nil", err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		t.Fatalf("Commit() error = %v", err)
	}

	// Verify order was created
	found, err := orderRepo.FindByID(ctx, order.ID)
	if err != nil {
		t.Errorf("FindByID() error = %v", err)
	}
	if found.TotalPrice != 49.99 {
		t.Errorf("FindByID() totalPrice = %f, want 49.99", found.TotalPrice)
	}
}

func TestOrderRepository_FindByID(t *testing.T) {
	testDB := setupOrdersTestDB(t)
	userRepo := auth.NewUserRepository(testDB)
	orderRepo := NewOrderRepository(testDB)
	ctx := context.Background()

	// Create test user and order
	user := &auth.User{
		ID:           uuid.New(),
		Email:        "findorder@example.com",
		PasswordHash: "hash",
		Username:         "Find Order User",
		Role:         auth.RoleCustomer,
	}
	if err := userRepo.Create(ctx, user); err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	order := &Order{
		ID:         uuid.New(),
		UserID:     user.ID,
		TotalPrice: 29.99,
		Status:     StatusPending,
	}
	if err := orderRepo.Create(ctx, order); err != nil {
		t.Fatalf("Failed to create test order: %v", err)
	}

	tests := []struct {
		name    string
		id      uuid.UUID
		wantErr error
		wantID  uuid.UUID
	}{
		{
			name:    "find existing order",
			id:      order.ID,
			wantErr: nil,
			wantID:  order.ID,
		},
		{
			name:    "order not found",
			id:      uuid.New(),
			wantErr: shared.ErrNotFound,
			wantID:  uuid.Nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			found, err := orderRepo.FindByID(ctx, tt.id)
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

func TestOrderRepository_FindByUserID(t *testing.T) {
	testDB := setupOrdersTestDB(t)
	userRepo := auth.NewUserRepository(testDB)
	orderRepo := NewOrderRepository(testDB)
	ctx := context.Background()

	// Create test users
	user1 := &auth.User{
		ID:           uuid.New(),
		Email:        "user1@example.com",
		PasswordHash: "hash",
		Username:         "User 1",
		Role:         auth.RoleCustomer,
	}
	user2 := &auth.User{
		ID:           uuid.New(),
		Email:        "user2@example.com",
		PasswordHash: "hash",
		Username:         "User 2",
		Role:         auth.RoleCustomer,
	}
	if err := userRepo.Create(ctx, user1); err != nil {
		t.Fatalf("Failed to create user1: %v", err)
	}
	if err := userRepo.Create(ctx, user2); err != nil {
		t.Fatalf("Failed to create user2: %v", err)
	}

	// Create orders for user1
	order1 := &Order{ID: uuid.New(), UserID: user1.ID, TotalPrice: 10.00, Status: StatusPending}
	order2 := &Order{ID: uuid.New(), UserID: user1.ID, TotalPrice: 20.00, Status: StatusPending}
	// Create order for user2
	order3 := &Order{ID: uuid.New(), UserID: user2.ID, TotalPrice: 30.00, Status: StatusPending}

	if err := orderRepo.Create(ctx, order1); err != nil {
		t.Fatalf("Failed to create order1: %v", err)
	}
	if err := orderRepo.Create(ctx, order2); err != nil {
		t.Fatalf("Failed to create order2: %v", err)
	}
	if err := orderRepo.Create(ctx, order3); err != nil {
		t.Fatalf("Failed to create order3: %v", err)
	}

	// Test pagination
	orders, total, err := orderRepo.FindByUserID(ctx, user1.ID, 0, 10)
	if err != nil {
		t.Errorf("FindByUserID() error = %v, want nil", err)
	}
	if total != 2 {
		t.Errorf("FindByUserID() total = %d, want 2", total)
	}
	if len(orders) != 2 {
		t.Errorf("FindByUserID() count = %d, want 2", len(orders))
	}

	// Verify all orders belong to user1
	for _, o := range orders {
		if o.UserID != user1.ID {
			t.Errorf("FindByUserID() returned order for wrong user: %v", o.UserID)
		}
	}
}

func TestOrderRepository_FindAll(t *testing.T) {
	testDB := setupOrdersTestDB(t)
	userRepo := auth.NewUserRepository(testDB)
	orderRepo := NewOrderRepository(testDB)
	ctx := context.Background()

	// Create test users and orders
	user1 := &auth.User{
		ID:           uuid.New(),
		Email:        "all1@example.com",
		PasswordHash: "hash",
		Username:         "All User 1",
		Role:         auth.RoleCustomer,
	}
	user2 := &auth.User{
		ID:           uuid.New(),
		Email:        "all2@example.com",
		PasswordHash: "hash",
		Username:         "All User 2",
		Role:         auth.RoleCustomer,
	}
	if err := userRepo.Create(ctx, user1); err != nil {
		t.Fatalf("Failed to create user1: %v", err)
	}
	if err := userRepo.Create(ctx, user2); err != nil {
		t.Fatalf("Failed to create user2: %v", err)
	}

	// Create orders
	order1 := &Order{ID: uuid.New(), UserID: user1.ID, TotalPrice: 10.00, Status: StatusPending}
	order2 := &Order{ID: uuid.New(), UserID: user2.ID, TotalPrice: 20.00, Status: StatusPending}
	order3 := &Order{ID: uuid.New(), UserID: user1.ID, TotalPrice: 30.00, Status: StatusPending}

	if err := orderRepo.Create(ctx, order1); err != nil {
		t.Fatalf("Failed to create order1: %v", err)
	}
	if err := orderRepo.Create(ctx, order2); err != nil {
		t.Fatalf("Failed to create order2: %v", err)
	}
	if err := orderRepo.Create(ctx, order3); err != nil {
		t.Fatalf("Failed to create order3: %v", err)
	}

	// Test FindAll
	orders, total, err := orderRepo.FindAll(ctx, 0, 10)
	if err != nil {
		t.Errorf("FindAll() error = %v, want nil", err)
	}
	if total != 3 {
		t.Errorf("FindAll() total = %d, want 3", total)
	}
	if len(orders) != 3 {
		t.Errorf("FindAll() count = %d, want 3", len(orders))
	}
}

func TestOrderItemRepository_CreateBatch(t *testing.T) {
	testDB := setupOrdersTestDB(t)
	userRepo := auth.NewUserRepository(testDB)
	orderRepo := NewOrderRepository(testDB)
	orderItemRepo := NewOrderItemRepository(testDB)
	productRepo := products.NewProductRepository(testDB)
	ctx := context.Background()

	// Setup
	user := &auth.User{
		ID:           uuid.New(),
		Email:        "items@example.com",
		PasswordHash: "hash",
		Username:         "Items User",
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

	order := &Order{
		ID:         uuid.New(),
		UserID:     user.ID,
		TotalPrice: 50.00,
		Status:     StatusPending,
	}
	if err := orderRepo.Create(ctx, order); err != nil {
		t.Fatalf("Failed to create order: %v", err)
	}

	// Create batch of order items
	items := []OrderItem{
		{ID: uuid.New(), OrderID: order.ID, ProductID: product1.ID, Quantity: 1, Price: 10.00},
		{ID: uuid.New(), OrderID: order.ID, ProductID: product2.ID, Quantity: 2, Price: 20.00},
	}

	if err := orderItemRepo.CreateBatch(ctx, items); err != nil {
		t.Errorf("CreateBatch() error = %v, want nil", err)
	}

	// Verify items were created
	found, err := orderItemRepo.FindByOrderID(ctx, order.ID)
	if err != nil {
		t.Errorf("FindByOrderID() error = %v", err)
	}
	if len(found) != 2 {
		t.Errorf("FindByOrderID() count = %d, want 2", len(found))
	}
}

func TestOrderItemRepository_CreateBatchTx(t *testing.T) {
	testDB := setupOrdersTestDB(t)
	userRepo := auth.NewUserRepository(testDB)
	orderRepo := NewOrderRepository(testDB)
	orderItemRepo := NewOrderItemRepository(testDB)
	productRepo := products.NewProductRepository(testDB)
	ctx := context.Background()

	// Setup
	user := &auth.User{
		ID:           uuid.New(),
		Email:        "itemstx@example.com",
		PasswordHash: "hash",
		Username:         "Items Tx User",
		Role:         auth.RoleCustomer,
	}
	if err := userRepo.Create(ctx, user); err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	product := &products.Product{ID: uuid.New(), Name: "Tx Product", Price: 15.00}
	if err := productRepo.Create(ctx, product); err != nil {
		t.Fatalf("Failed to create product: %v", err)
	}

	order := &Order{
		ID:         uuid.New(),
		UserID:     user.ID,
		TotalPrice: 30.00,
		Status:     StatusPending,
	}
	if err := orderRepo.Create(ctx, order); err != nil {
		t.Fatalf("Failed to create order: %v", err)
	}

	// Test transaction-based batch create
	tx := testDB.Begin()
	defer tx.Rollback()

	items := []OrderItem{
		{ID: uuid.New(), OrderID: order.ID, ProductID: product.ID, Quantity: 2, Price: 15.00},
	}

	if err := orderItemRepo.CreateBatchTx(ctx, tx, items); err != nil {
		t.Errorf("CreateBatchTx() error = %v, want nil", err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		t.Fatalf("Commit() error = %v", err)
	}

	// Verify items were created
	found, err := orderItemRepo.FindByOrderID(ctx, order.ID)
	if err != nil {
		t.Errorf("FindByOrderID() error = %v", err)
	}
	if len(found) != 1 {
		t.Errorf("FindByOrderID() count = %d, want 1", len(found))
	}
}

func TestOrderItemRepository_FindByOrderID(t *testing.T) {
	testDB := setupOrdersTestDB(t)
	userRepo := auth.NewUserRepository(testDB)
	orderRepo := NewOrderRepository(testDB)
	orderItemRepo := NewOrderItemRepository(testDB)
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

	order := &Order{
		ID:         uuid.New(),
		UserID:     user.ID,
		TotalPrice: 50.00,
		Status:     StatusPending,
	}
	if err := orderRepo.Create(ctx, order); err != nil {
		t.Fatalf("Failed to create order: %v", err)
	}

	// Create items
	items := []OrderItem{
		{ID: uuid.New(), OrderID: order.ID, ProductID: product1.ID, Quantity: 1, Price: 10.00},
		{ID: uuid.New(), OrderID: order.ID, ProductID: product2.ID, Quantity: 2, Price: 20.00},
	}
	if err := orderItemRepo.CreateBatch(ctx, items); err != nil {
		t.Fatalf("Failed to create items: %v", err)
	}

	// Find items by order ID
	found, err := orderItemRepo.FindByOrderID(ctx, order.ID)
	if err != nil {
		t.Errorf("FindByOrderID() error = %v, want nil", err)
	}
	if len(found) != 2 {
		t.Errorf("FindByOrderID() count = %d, want 2", len(found))
	}

	// Verify all items belong to the order
	for _, item := range found {
		if item.OrderID != order.ID {
			t.Errorf("FindByOrderID() returned item for wrong order: %v", item.OrderID)
		}
	}

	// Test empty order
	emptyOrder := &Order{
		ID:         uuid.New(),
		UserID:     user.ID,
		TotalPrice: 0.00,
		Status:     StatusPending,
	}
	if err := orderRepo.Create(ctx, emptyOrder); err != nil {
		t.Fatalf("Failed to create empty order: %v", err)
	}

	found, err = orderItemRepo.FindByOrderID(ctx, emptyOrder.ID)
	if err != nil {
		t.Errorf("FindByOrderID() empty order error = %v, want nil", err)
	}
	if len(found) != 0 {
		t.Errorf("FindByOrderID() empty order count = %d, want 0", len(found))
	}
}
