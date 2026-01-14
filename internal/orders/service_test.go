package orders

import (
	"context"
	"testing"
	"unsafe"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/mschl/ecommerce-api/internal/cart"
	"github.com/mschl/ecommerce-api/internal/inventory"
	"github.com/mschl/ecommerce-api/internal/notifications"
	"github.com/mschl/ecommerce-api/internal/products"
	"github.com/mschl/ecommerce-api/internal/shared"
)

// Mock repositories for order service testing

type MockOrderRepository struct {
	orders map[uuid.UUID]*Order
}

func NewMockOrderRepository() *MockOrderRepository {
	return &MockOrderRepository{orders: make(map[uuid.UUID]*Order)}
}

func (m *MockOrderRepository) Create(ctx context.Context, order *Order) error {
	if order.ID == uuid.Nil {
		order.ID = uuid.New()
	}
	m.orders[order.ID] = order
	return nil
}

func (m *MockOrderRepository) CreateTx(ctx context.Context, tx *gorm.DB, order *Order) error {
	return m.Create(ctx, order)
}

func (m *MockOrderRepository) FindByID(ctx context.Context, id uuid.UUID) (*Order, error) {
	if order, exists := m.orders[id]; exists {
		return order, nil
	}
	return nil, shared.ErrNotFound
}

func (m *MockOrderRepository) FindByUserID(ctx context.Context, userID uuid.UUID, offset, limit int) ([]Order, int64, error) {
	orders := make([]Order, 0)
	for _, o := range m.orders {
		if o.UserID == userID {
			orders = append(orders, *o)
		}
	}
	total := int64(len(orders))
	return orders, total, nil
}

func (m *MockOrderRepository) FindAll(ctx context.Context, offset, limit int) ([]Order, int64, error) {
	orders := make([]Order, 0, len(m.orders))
	for _, o := range m.orders {
		orders = append(orders, *o)
	}
	total := int64(len(orders))
	return orders, total, nil
}

type MockOrderItemRepository struct {
	items map[uuid.UUID]*OrderItem
}

func NewMockOrderItemRepository() *MockOrderItemRepository {
	return &MockOrderItemRepository{items: make(map[uuid.UUID]*OrderItem)}
}

func (m *MockOrderItemRepository) CreateBatch(ctx context.Context, items []OrderItem) error {
	for i := range items {
		if items[i].ID == uuid.Nil {
			items[i].ID = uuid.New()
		}
		m.items[items[i].ID] = &items[i]
	}
	return nil
}

func (m *MockOrderItemRepository) CreateBatchTx(ctx context.Context, tx *gorm.DB, items []OrderItem) error {
	return m.CreateBatch(ctx, items)
}

func (m *MockOrderItemRepository) FindByOrderID(ctx context.Context, orderID uuid.UUID) ([]OrderItem, error) {
	items := make([]OrderItem, 0)
	for _, item := range m.items {
		if item.OrderID == orderID {
			items = append(items, *item)
		}
	}
	return items, nil
}

type MockCartRepository struct {
	carts map[uuid.UUID]*cart.Cart
}

func NewMockCartRepository() *MockCartRepository {
	return &MockCartRepository{carts: make(map[uuid.UUID]*cart.Cart)}
}

func (m *MockCartRepository) FindByUserID(ctx context.Context, userID uuid.UUID) (*cart.Cart, error) {
	for _, c := range m.carts {
		if c.UserID == userID {
			return c, nil
		}
	}
	return nil, shared.ErrNotFound
}

func (m *MockCartRepository) FindOrCreateByUserID(ctx context.Context, userID uuid.UUID) (*cart.Cart, error) {
	for _, c := range m.carts {
		if c.UserID == userID {
			return c, nil
		}
	}
	c := &cart.Cart{
		ID:     uuid.New(),
		UserID: userID,
	}
	m.carts[c.ID] = c
	return c, nil
}

func (m *MockCartRepository) Clear(ctx context.Context, cartID uuid.UUID) error {
	return nil
}

func (m *MockCartRepository) ClearTx(ctx context.Context, tx *gorm.DB, cartID uuid.UUID) error {
	return nil
}

type MockCartItemRepository struct {
	items map[uuid.UUID]*cart.CartItem
}

func NewMockCartItemRepository() *MockCartItemRepository {
	return &MockCartItemRepository{items: make(map[uuid.UUID]*cart.CartItem)}
}

func (m *MockCartItemRepository) FindByCartID(ctx context.Context, cartID uuid.UUID) ([]cart.CartItem, error) {
	items := make([]cart.CartItem, 0)
	for _, item := range m.items {
		if item.CartID == cartID {
			items = append(items, *item)
		}
	}
	return items, nil
}

func (m *MockCartItemRepository) Create(ctx context.Context, item *cart.CartItem) error {
	if item.ID == uuid.Nil {
		item.ID = uuid.New()
	}
	m.items[item.ID] = item
	return nil
}

func (m *MockCartItemRepository) FindByID(ctx context.Context, id uuid.UUID) (*cart.CartItem, error) {
	if item, exists := m.items[id]; exists {
		return item, nil
	}
	return nil, shared.ErrNotFound
}

func (m *MockCartItemRepository) Update(ctx context.Context, item *cart.CartItem) error {
	m.items[item.ID] = item
	return nil
}

func (m *MockCartItemRepository) Delete(ctx context.Context, id uuid.UUID) error {
	delete(m.items, id)
	return nil
}

func (m *MockCartItemRepository) FindByCartAndProduct(ctx context.Context, cartID, productID uuid.UUID) (*cart.CartItem, error) {
	for _, item := range m.items {
		if item.CartID == cartID && item.ProductID == productID {
			return item, nil
		}
	}
	return nil, shared.ErrNotFound
}

type MockProductRepository struct {
	products map[uuid.UUID]*products.Product
}

func NewMockProductRepository() *MockProductRepository {
	return &MockProductRepository{products: make(map[uuid.UUID]*products.Product)}
}

func (m *MockProductRepository) Create(ctx context.Context, product *products.Product) error {
	m.products[product.ID] = product
	return nil
}

func (m *MockProductRepository) FindByID(ctx context.Context, id uuid.UUID) (*products.Product, error) {
	if product, exists := m.products[id]; exists {
		return product, nil
	}
	return nil, shared.ErrNotFound
}

func (m *MockProductRepository) FindAll(ctx context.Context, offset, limit int, categoryID *uuid.UUID) ([]products.Product, int64, error) {
	return nil, 0, nil
}

func (m *MockProductRepository) Update(ctx context.Context, product *products.Product) error {
	return nil
}

func (m *MockProductRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return nil
}

type MockNotificationService struct {
	notifications map[uuid.UUID]*notifications.Notification
}

func NewMockNotificationService() *MockNotificationService {
	return &MockNotificationService{notifications: make(map[uuid.UUID]*notifications.Notification)}
}

func (m *MockNotificationService) CreateNotification(ctx context.Context, userID uuid.UUID, title, message string) (*notifications.Notification, error) {
	notif := &notifications.Notification{
		ID:      uuid.New(),
		UserID:  userID,
		Title:   title,
		Message: message,
		IsRead:  false,
	}
	m.notifications[notif.ID] = notif
	return notif, nil
}

func (m *MockNotificationService) CreateNotificationTx(ctx context.Context, tx *gorm.DB, userID uuid.UUID, title, message string) (*notifications.Notification, error) {
	return m.CreateNotification(ctx, userID, title, message)
}

func (m *MockNotificationService) GetNotifications(ctx context.Context, userID uuid.UUID) ([]notifications.Notification, error) {
	return nil, nil
}

func (m *MockNotificationService) MarkAsRead(ctx context.Context, userID uuid.UUID, notificationID uuid.UUID) error {
	return nil
}

// MockInventoryRepository tracks inventory quantities in-memory
type MockInventoryRepository struct {
	quantities map[uuid.UUID]int
}

func NewMockInventoryRepository() *MockInventoryRepository {
	return &MockInventoryRepository{quantities: make(map[uuid.UUID]int)}
}

func (m *MockInventoryRepository) Create(ctx context.Context, inv *inventory.Inventory) error {
	m.quantities[inv.ProductID] = inv.Quantity
	return nil
}

func (m *MockInventoryRepository) FindByProductID(ctx context.Context, productID uuid.UUID) (*inventory.Inventory, error) {
	qty, ok := m.quantities[productID]
	if !ok {
		return nil, shared.ErrNotFound
	}
	return &inventory.Inventory{ProductID: productID, Quantity: qty}, nil
}

func (m *MockInventoryRepository) Update(ctx context.Context, inv *inventory.Inventory) error {
	m.quantities[inv.ProductID] = inv.Quantity
	return nil
}

func (m *MockInventoryRepository) DecrementQuantity(ctx context.Context, productID uuid.UUID, amount int) error {
	return m.DecrementQuantityTx(ctx, nil, productID, amount)
}

func (m *MockInventoryRepository) DecrementQuantityTx(ctx context.Context, tx *gorm.DB, productID uuid.UUID, amount int) error {
	qty, ok := m.quantities[productID]
	if !ok {
		return shared.ErrInsufficientInventory
	}
	if qty < amount {
		return shared.ErrInsufficientInventory
	}
	m.quantities[productID] = qty - amount
	return nil
}

// mockTransaction creates a non-nil *gorm.DB pointer for testing
// Since all repositories are mocked and don't use the actual tx, we just need a non-nil pointer
// This is safe because the mocked repositories ignore the tx parameter
func mockTransaction() *gorm.DB {
	// Create a non-nil pointer by using unsafe to point to a real struct
	// This is safe in tests because mocked repositories don't actually use the tx
	// We use a dummy struct to ensure we have a valid memory address
	type dummy struct{}
	d := &dummy{}
	return (*gorm.DB)(unsafe.Pointer(d))
}

// helper to build orderService with stubbed transaction and mocks
func newTestOrderService(
	orderRepo OrderRepository,
	orderItemRepo OrderItemRepository,
	cartRepo *MockCartRepository,
	cartItemRepo *MockCartItemRepository,
	productRepo *MockProductRepository,
	inventoryRepo *MockInventoryRepository,
	notificationService *MockNotificationService,
) *orderService {
	return &orderService{
		// Pass a non-nil transaction so the service creates notifications
		// All repositories are mocked and ignore the tx parameter anyway
		transaction:         func(fn func(*gorm.DB) error) error { return fn(mockTransaction()) },
		orderRepo:           orderRepo,
		orderItemRepo:       orderItemRepo,
		cartRepo:            cartRepo,
		cartItemRepo:        cartItemRepo,
		productRepo:         productRepo,
		inventoryRepo:       inventoryRepo,
		notificationService: notificationService,
	}
}

// Unit Tests

func TestOrderService_CreateOrder(t *testing.T) {
	userID := uuid.New()
	productID := uuid.New()
	cartID := uuid.New()

	// Setup
	orderRepo := NewMockOrderRepository()
	orderItemRepo := NewMockOrderItemRepository()
	cartRepo := NewMockCartRepository()
	cartItemRepo := NewMockCartItemRepository()
	productRepo := NewMockProductRepository()
	inventoryRepo := NewMockInventoryRepository()
	notificationService := NewMockNotificationService()

	// Create cart
	userCart := &cart.Cart{
		ID:     cartID,
		UserID: userID,
	}
	cartRepo.carts[cartID] = userCart

	// Create product
	product := &products.Product{
		ID:    productID,
		Name:  "Test Product",
		Price: 19.99,
	}
	productRepo.products[productID] = product

	// Create inventory
	inventoryRepo.quantities[productID] = 5

	// Create cart item
	cartItem := &cart.CartItem{
		ID:        uuid.New(),
		CartID:    cartID,
		ProductID: productID,
		Quantity:  2,
	}
	cartItemRepo.items[cartItem.ID] = cartItem

	service := newTestOrderService(orderRepo, orderItemRepo, cartRepo, cartItemRepo, productRepo, inventoryRepo, notificationService)

	// Test successful order creation
	shippingAddress := CreateOrderRequest{
		ShippingStreet:  "123 Main St",
		ShippingCity:    "New York",
		ShippingZip:     "10001",
		ShippingCountry: "USA",
	}
	order, items, err := service.CreateOrder(context.Background(), userID, shippingAddress)
	if err != nil {
		t.Fatalf("CreateOrder() error = %v", err)
	}
	if order == nil {
		t.Fatal("CreateOrder() returned nil order")
	}
	if order.UserID != userID {
		t.Errorf("CreateOrder() order.UserID = %v, want %v", order.UserID, userID)
	}
	if order.TotalPrice != 39.98 { // 19.99 * 2
		t.Errorf("CreateOrder() order.TotalPrice = %f, want 39.98", order.TotalPrice)
	}
	if len(items) != 1 {
		t.Errorf("CreateOrder() items count = %d, want 1", len(items))
	}
}

func TestOrderService_CreateOrder_EmptyCart(t *testing.T) {
	userID := uuid.New()

	orderRepo := NewMockOrderRepository()
	orderItemRepo := NewMockOrderItemRepository()
	cartRepo := NewMockCartRepository()
	cartItemRepo := NewMockCartItemRepository()
	productRepo := NewMockProductRepository()
	inventoryRepo := NewMockInventoryRepository()
	notificationService := NewMockNotificationService()

	service := newTestOrderService(orderRepo, orderItemRepo, cartRepo, cartItemRepo, productRepo, inventoryRepo, notificationService)

	shippingAddress := CreateOrderRequest{
		ShippingStreet:  "123 Main St",
		ShippingCity:    "New York",
		ShippingZip:     "10001",
		ShippingCountry: "USA",
	}
	_, _, err := service.CreateOrder(context.Background(), userID, shippingAddress)
	if err != shared.ErrEmptyCart {
		t.Errorf("CreateOrder() error = %v, want %v", err, shared.ErrEmptyCart)
	}
}

func TestOrderService_GetOrder(t *testing.T) {
	userID := uuid.New()
	otherUserID := uuid.New()
	orderID := uuid.New()

	orderRepo := NewMockOrderRepository()
	orderItemRepo := NewMockOrderItemRepository()
	cartRepo := NewMockCartRepository()
	cartItemRepo := NewMockCartItemRepository()
	productRepo := NewMockProductRepository()
	inventoryRepo := NewMockInventoryRepository()
	notificationService := NewMockNotificationService()

	// Create order
	order := &Order{
		ID:     orderID,
		UserID: userID,
	}
	orderRepo.orders[orderID] = order

	service := newTestOrderService(orderRepo, orderItemRepo, cartRepo, cartItemRepo, productRepo, inventoryRepo, notificationService)

	// Test get own order
	gotOrder, _, err := service.GetOrder(context.Background(), userID, orderID, false)
	if err != nil {
		t.Errorf("GetOrder() error = %v", err)
	}
	if gotOrder.ID != orderID {
		t.Errorf("GetOrder() order.ID = %v, want %v", gotOrder.ID, orderID)
	}

	// Test customer cannot get other user's order
	_, _, err = service.GetOrder(context.Background(), otherUserID, orderID, false)
	if err != shared.ErrForbidden {
		t.Errorf("GetOrder() error = %v, want %v", err, shared.ErrForbidden)
	}

	// Test admin can get any order
	gotOrder, _, err = service.GetOrder(context.Background(), otherUserID, orderID, true)
	if err != nil {
		t.Errorf("GetOrder() as admin error = %v", err)
	}
	if gotOrder.ID != orderID {
		t.Errorf("GetOrder() as admin order.ID = %v, want %v", gotOrder.ID, orderID)
	}
}

func TestOrderService_ListOrders(t *testing.T) {
	userID1 := uuid.New()
	userID2 := uuid.New()

	orderRepo := NewMockOrderRepository()
	orderItemRepo := NewMockOrderItemRepository()
	cartRepo := NewMockCartRepository()
	cartItemRepo := NewMockCartItemRepository()
	productRepo := NewMockProductRepository()
	inventoryRepo := NewMockInventoryRepository()
	notificationService := NewMockNotificationService()

	// Create orders for different users
	order1 := &Order{ID: uuid.New(), UserID: userID1}
	order2 := &Order{ID: uuid.New(), UserID: userID2}
	orderRepo.orders[order1.ID] = order1
	orderRepo.orders[order2.ID] = order2

	service := newTestOrderService(orderRepo, orderItemRepo, cartRepo, cartItemRepo, productRepo, inventoryRepo, notificationService)

	// Test customer sees only own orders
	orders, total, err := service.ListOrders(context.Background(), userID1, false, 1, 10)
	if err != nil {
		t.Errorf("ListOrders() error = %v", err)
	}
	if total != 1 {
		t.Errorf("ListOrders() total = %d, want 1", total)
	}
	if len(orders) != 1 {
		t.Errorf("ListOrders() count = %d, want 1", len(orders))
	}
	if orders[0].UserID != userID1 {
		t.Errorf("ListOrders() order.UserID = %v, want %v", orders[0].UserID, userID1)
	}

	// Test admin sees all orders
	orders, total, err = service.ListOrders(context.Background(), userID1, true, 1, 10)
	if err != nil {
		t.Errorf("ListOrders() as admin error = %v", err)
	}
	if total != 2 {
		t.Errorf("ListOrders() as admin total = %d, want 2", total)
	}
	if len(orders) != 2 {
		t.Errorf("ListOrders() as admin count = %d, want 2", len(orders))
	}
}

// Property 10: Order Transaction Atomicity
func TestProperty_OrderTransactionAtomicity(t *testing.T) {
	userID := uuid.New()
	productID := uuid.New()
	cartID := uuid.New()

	orderRepo := NewMockOrderRepository()
	orderItemRepo := NewMockOrderItemRepository()
	cartRepo := NewMockCartRepository()
	cartItemRepo := NewMockCartItemRepository()
	productRepo := NewMockProductRepository()
	inventoryRepo := NewMockInventoryRepository()
	notificationService := NewMockNotificationService()

	// Setup cart with items
	userCart := &cart.Cart{ID: cartID, UserID: userID}
	cartRepo.carts[cartID] = userCart

	product := &products.Product{ID: productID, Name: "Test", Price: 10.00}
	productRepo.products[productID] = product

	// Inventory
	inventoryRepo.quantities[productID] = 3

	cartItem := &cart.CartItem{
		ID:        uuid.New(),
		CartID:    cartID,
		ProductID: productID,
		Quantity:  1,
	}
	cartItemRepo.items[cartItem.ID] = cartItem

	service := newTestOrderService(orderRepo, orderItemRepo, cartRepo, cartItemRepo, productRepo, inventoryRepo, notificationService)

	// Property: Order creation is atomic - either all succeeds or all fails
	// In this test, we verify that order, items, and notification are all created together
	shippingAddress := CreateOrderRequest{
		ShippingStreet:  "123 Main St",
		ShippingCity:    "New York",
		ShippingZip:     "10001",
		ShippingCountry: "USA",
	}
	order, items, err := service.CreateOrder(context.Background(), userID, shippingAddress)
	if err != nil {
		t.Fatalf("CreateOrder failed: %v", err)
	}

	// Verify order exists
	if order == nil {
		t.Error("Order should be created")
	}

	// Verify items exist
	if len(items) == 0 {
		t.Error("Order items should be created")
	}

	// Verify notification was created (via notification service)
	if len(notificationService.notifications) == 0 {
		t.Error("Notification should be created")
	}
}

// Property 11: Order Access Control
func TestProperty_OrderAccessControl(t *testing.T) {
	userID1 := uuid.New()
	userID2 := uuid.New()
	orderID := uuid.New()

	orderRepo := NewMockOrderRepository()
	orderItemRepo := NewMockOrderItemRepository()
	cartRepo := NewMockCartRepository()
	cartItemRepo := NewMockCartItemRepository()
	productRepo := NewMockProductRepository()
	inventoryRepo := NewMockInventoryRepository()
	notificationService := NewMockNotificationService()

	// Create order for user1
	order := &Order{ID: orderID, UserID: userID1}
	orderRepo.orders[orderID] = order

	service := newTestOrderService(orderRepo, orderItemRepo, cartRepo, cartItemRepo, productRepo, inventoryRepo, notificationService)

	// Property: Users can only access their own orders
	_, _, err := service.GetOrder(context.Background(), userID1, orderID, false)
	if err != nil {
		t.Errorf("User should access own order: %v", err)
	}

	// Property: Users cannot access other users' orders
	_, _, err = service.GetOrder(context.Background(), userID2, orderID, false)
	if err != shared.ErrForbidden {
		t.Errorf("User should not access other's order, got: %v", err)
	}

	// Property: Admins can access any order
	_, _, err = service.GetOrder(context.Background(), userID2, orderID, true)
	if err != nil {
		t.Errorf("Admin should access any order: %v", err)
	}
}

// Property 12: Order Triggers Notification
func TestProperty_OrderTriggersNotification(t *testing.T) {
	userID := uuid.New()
	productID := uuid.New()
	cartID := uuid.New()

	orderRepo := NewMockOrderRepository()
	orderItemRepo := NewMockOrderItemRepository()
	cartRepo := NewMockCartRepository()
	cartItemRepo := NewMockCartItemRepository()
	productRepo := NewMockProductRepository()
	inventoryRepo := NewMockInventoryRepository()
	notificationService := NewMockNotificationService()

	// Setup
	userCart := &cart.Cart{ID: cartID, UserID: userID}
	cartRepo.carts[cartID] = userCart

	product := &products.Product{ID: productID, Name: "Test", Price: 10.00}
	productRepo.products[productID] = product

	// Inventory
	inventoryRepo.quantities[productID] = 2

	cartItem := &cart.CartItem{
		ID:        uuid.New(),
		CartID:    cartID,
		ProductID: productID,
		Quantity:  1,
	}
	cartItemRepo.items[cartItem.ID] = cartItem

	service := newTestOrderService(orderRepo, orderItemRepo, cartRepo, cartItemRepo, productRepo, inventoryRepo, notificationService)

	// Property: Creating an order triggers a notification
	initialNotificationCount := len(notificationService.notifications)
	shippingAddress := CreateOrderRequest{
		ShippingStreet:  "123 Main St",
		ShippingCity:    "New York",
		ShippingZip:     "10001",
		ShippingCountry: "USA",
	}
	_, _, err := service.CreateOrder(context.Background(), userID, shippingAddress)
	if err != nil {
		t.Fatalf("CreateOrder failed: %v", err)
	}

	// Verify notification was created
	if len(notificationService.notifications) != initialNotificationCount+1 {
		t.Errorf("Notification count = %d, want %d", len(notificationService.notifications), initialNotificationCount+1)
	}

	// Verify notification is for the correct user
	for _, notif := range notificationService.notifications {
		if notif.UserID == userID {
			return // Found notification for user
		}
	}
	t.Error("Notification should be created for the user who placed the order")
}

// TestOrderService_CreateOrder_InvalidShippingAddress tests CreateOrder with invalid shipping address
func TestOrderService_CreateOrder_InvalidShippingAddress(t *testing.T) {
	orderRepo := NewMockOrderRepository()
	orderItemRepo := NewMockOrderItemRepository()
	cartRepo := NewMockCartRepository()
	cartItemRepo := NewMockCartItemRepository()
	productRepo := NewMockProductRepository()
	inventoryRepo := NewMockInventoryRepository()
	notificationService := NewMockNotificationService()
	service := newTestOrderService(orderRepo, orderItemRepo, cartRepo, cartItemRepo, productRepo, inventoryRepo, notificationService)

	userID := uuid.New()

	tests := []struct {
		name            string
		shippingAddress CreateOrderRequest
		wantErr         error
	}{
		{
			name: "missing street",
			shippingAddress: CreateOrderRequest{
				ShippingCity:    "City",
				ShippingZip:     "12345",
				ShippingCountry: "Country",
			},
			wantErr: shared.ErrBadRequest,
		},
		{
			name: "missing city",
			shippingAddress: CreateOrderRequest{
				ShippingStreet:  "Street",
				ShippingZip:     "12345",
				ShippingCountry: "Country",
			},
			wantErr: shared.ErrBadRequest,
		},
		{
			name: "missing zip",
			shippingAddress: CreateOrderRequest{
				ShippingStreet:  "Street",
				ShippingCity:    "City",
				ShippingCountry: "Country",
			},
			wantErr: shared.ErrBadRequest,
		},
		{
			name: "missing country",
			shippingAddress: CreateOrderRequest{
				ShippingStreet: "Street",
				ShippingCity:   "City",
				ShippingZip:    "12345",
			},
			wantErr: shared.ErrBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := service.CreateOrder(context.Background(), userID, tt.shippingAddress)
			if err != tt.wantErr {
				t.Errorf("CreateOrder() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestOrderService_CreateOrder_CartNotFound tests CreateOrder when cart doesn't exist
func TestOrderService_CreateOrder_CartNotFound(t *testing.T) {
	orderRepo := NewMockOrderRepository()
	orderItemRepo := NewMockOrderItemRepository()
	cartRepo := NewMockCartRepository()
	cartItemRepo := NewMockCartItemRepository()
	productRepo := NewMockProductRepository()
	inventoryRepo := NewMockInventoryRepository()
	notificationService := NewMockNotificationService()
	service := newTestOrderService(orderRepo, orderItemRepo, cartRepo, cartItemRepo, productRepo, inventoryRepo, notificationService)

	userID := uuid.New()
	shippingAddress := CreateOrderRequest{
		ShippingStreet:  "123 Main St",
		ShippingCity:    "City",
		ShippingZip:     "12345",
		ShippingCountry: "Country",
	}

	_, _, err := service.CreateOrder(context.Background(), userID, shippingAddress)
	if err != shared.ErrEmptyCart {
		t.Errorf("CreateOrder() error = %v, wantErr %v", err, shared.ErrEmptyCart)
	}
}

// TestOrderService_GetOrder_NotFound tests GetOrder when order doesn't exist
func TestOrderService_GetOrder_NotFound(t *testing.T) {
	orderRepo := NewMockOrderRepository()
	orderItemRepo := NewMockOrderItemRepository()
	cartRepo := NewMockCartRepository()
	cartItemRepo := NewMockCartItemRepository()
	productRepo := NewMockProductRepository()
	inventoryRepo := NewMockInventoryRepository()
	notificationService := NewMockNotificationService()
	service := newTestOrderService(orderRepo, orderItemRepo, cartRepo, cartItemRepo, productRepo, inventoryRepo, notificationService)

	userID := uuid.New()
	orderID := uuid.New()

	_, _, err := service.GetOrder(context.Background(), userID, orderID, false)
	if err != shared.ErrNotFound {
		t.Errorf("GetOrder() error = %v, wantErr %v", err, shared.ErrNotFound)
	}
}
