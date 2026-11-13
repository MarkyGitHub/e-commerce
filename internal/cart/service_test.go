package cart

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/MarkyGitHub/e-commerce/internal/inventory"
	"github.com/MarkyGitHub/e-commerce/internal/notifications"
	"github.com/MarkyGitHub/e-commerce/internal/products"
	"github.com/MarkyGitHub/e-commerce/internal/shared"
)

// Mock repositories

type MockCartRepository struct {
	carts map[uuid.UUID]*Cart
}

func NewMockCartRepository() *MockCartRepository {
	return &MockCartRepository{carts: make(map[uuid.UUID]*Cart)}
}

func (m *MockCartRepository) FindOrCreateByUserID(ctx context.Context, userID uuid.UUID) (*Cart, error) {
	for _, c := range m.carts {
		if c.UserID == userID {
			return c, nil
		}
	}
	cart := &Cart{
		ID:     uuid.New(),
		UserID: userID,
	}
	m.carts[cart.ID] = cart
	return cart, nil
}

func (m *MockCartRepository) FindByUserID(ctx context.Context, userID uuid.UUID) (*Cart, error) {
	for _, c := range m.carts {
		if c.UserID == userID {
			return c, nil
		}
	}
	return nil, shared.ErrNotFound
}

func (m *MockCartRepository) Clear(ctx context.Context, cartID uuid.UUID) error {
	return nil
}

func (m *MockCartRepository) ClearTx(ctx context.Context, tx *gorm.DB, cartID uuid.UUID) error {
	return nil
}

type MockCartItemRepository struct {
	items map[uuid.UUID]*CartItem
}

func NewMockCartItemRepository() *MockCartItemRepository {
	return &MockCartItemRepository{items: make(map[uuid.UUID]*CartItem)}
}

func (m *MockCartItemRepository) Create(ctx context.Context, item *CartItem) error {
	if item.ID == uuid.Nil {
		item.ID = uuid.New()
	}
	m.items[item.ID] = item
	return nil
}

func (m *MockCartItemRepository) FindByCartID(ctx context.Context, cartID uuid.UUID) ([]CartItem, error) {
	items := make([]CartItem, 0)
	for _, item := range m.items {
		if item.CartID == cartID {
			items = append(items, *item)
		}
	}
	return items, nil
}

func (m *MockCartItemRepository) FindByID(ctx context.Context, id uuid.UUID) (*CartItem, error) {
	if item, exists := m.items[id]; exists {
		return item, nil
	}
	return nil, shared.ErrNotFound
}

func (m *MockCartItemRepository) Update(ctx context.Context, item *CartItem) error {
	if _, exists := m.items[item.ID]; !exists {
		return shared.ErrNotFound
	}
	m.items[item.ID] = item
	return nil
}

func (m *MockCartItemRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if _, exists := m.items[id]; !exists {
		return shared.ErrNotFound
	}
	delete(m.items, id)
	return nil
}

func (m *MockCartItemRepository) FindByCartAndProduct(ctx context.Context, cartID, productID uuid.UUID) (*CartItem, error) {
	for _, item := range m.items {
		if item.CartID == cartID && item.ProductID == productID {
			return item, nil
		}
	}
	return nil, shared.ErrNotFound
}

type MockInventoryRepository struct {
	inventory map[uuid.UUID]*inventory.Inventory
}

func NewMockInventoryRepository() *MockInventoryRepository {
	return &MockInventoryRepository{inventory: make(map[uuid.UUID]*inventory.Inventory)}
}

func (m *MockInventoryRepository) Create(ctx context.Context, inv *inventory.Inventory) error {
	m.inventory[inv.ProductID] = inv
	return nil
}

func (m *MockInventoryRepository) FindByProductID(ctx context.Context, productID uuid.UUID) (*inventory.Inventory, error) {
	if inv, exists := m.inventory[productID]; exists {
		return inv, nil
	}
	return nil, shared.ErrNotFound
}

func (m *MockInventoryRepository) Update(ctx context.Context, inv *inventory.Inventory) error {
	m.inventory[inv.ProductID] = inv
	return nil
}

func (m *MockInventoryRepository) DecrementQuantity(ctx context.Context, productID uuid.UUID, amount int) error {
	return nil
}

func (m *MockInventoryRepository) DecrementQuantityTx(ctx context.Context, tx *gorm.DB, productID uuid.UUID, amount int) error {
	return nil
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

// Tests

func TestCartService_GetCart(t *testing.T) {
	cartRepo := NewMockCartRepository()
	cartItemRepo := NewMockCartItemRepository()
	invRepo := NewMockInventoryRepository()
	prodRepo := NewMockProductRepository()
	notificationService := NewMockNotificationService()
	service := NewCartService(cartRepo, cartItemRepo, invRepo, prodRepo, notificationService)

	userID := uuid.New()

	// Get cart (should create new one)
	cart, items, err := service.GetCart(context.Background(), userID)
	if err != nil {
		t.Fatalf("GetCart() error = %v", err)
	}
	if cart.UserID != userID {
		t.Errorf("GetCart() cart.UserID = %v, want %v", cart.UserID, userID)
	}
	if len(items) != 0 {
		t.Errorf("GetCart() items = %v, want empty", len(items))
	}

	// Get cart again (should return existing)
	cart2, _, err := service.GetCart(context.Background(), userID)
	if err != nil {
		t.Fatalf("GetCart() error = %v", err)
	}
	if cart2.ID != cart.ID {
		t.Error("GetCart() should return same cart")
	}
}

func TestCartService_AddItem(t *testing.T) {
	cartRepo := NewMockCartRepository()
	cartItemRepo := NewMockCartItemRepository()
	invRepo := NewMockInventoryRepository()
	prodRepo := NewMockProductRepository()
		notificationService := NewMockNotificationService()
		service := NewCartService(cartRepo, cartItemRepo, invRepo, prodRepo, notificationService)

	userID := uuid.New()
	productID := uuid.New()

	// Add product and inventory
	prodRepo.products[productID] = &products.Product{
		ID:    productID,
		Name:  "Test Product",
		Price: 19.99,
	}
	invRepo.inventory[productID] = &inventory.Inventory{
		ProductID: productID,
		Quantity:  100,
	}

	// Add item to cart
	item, err := service.AddItem(context.Background(), userID, AddItemRequest{
		ProductID: productID,
		Quantity:  2,
	})
	if err != nil {
		t.Fatalf("AddItem() error = %v", err)
	}
	if item.ProductID != productID {
		t.Errorf("AddItem() productID = %v, want %v", item.ProductID, productID)
	}
	if item.Quantity != 2 {
		t.Errorf("AddItem() quantity = %v, want %v", item.Quantity, 2)
	}
}

func TestCartService_AddItem_ProductNotFound(t *testing.T) {
	cartRepo := NewMockCartRepository()
	cartItemRepo := NewMockCartItemRepository()
	invRepo := NewMockInventoryRepository()
	prodRepo := NewMockProductRepository()
		notificationService := NewMockNotificationService()
		service := NewCartService(cartRepo, cartItemRepo, invRepo, prodRepo, notificationService)

	userID := uuid.New()

	_, err := service.AddItem(context.Background(), userID, AddItemRequest{
		ProductID: uuid.New(),
		Quantity:  1,
	})
	if err != shared.ErrNotFound {
		t.Errorf("AddItem() error = %v, want %v", err, shared.ErrNotFound)
	}
}

func TestCartService_AddItem_InsufficientInventory(t *testing.T) {
	cartRepo := NewMockCartRepository()
	cartItemRepo := NewMockCartItemRepository()
	invRepo := NewMockInventoryRepository()
	prodRepo := NewMockProductRepository()
		notificationService := NewMockNotificationService()
		service := NewCartService(cartRepo, cartItemRepo, invRepo, prodRepo, notificationService)

	userID := uuid.New()
	productID := uuid.New()

	// Add product with limited inventory
	prodRepo.products[productID] = &products.Product{
		ID:    productID,
		Name:  "Limited Product",
		Price: 19.99,
	}
	invRepo.inventory[productID] = &inventory.Inventory{
		ProductID: productID,
		Quantity:  5,
	}

	// Try to add more than available
	_, err := service.AddItem(context.Background(), userID, AddItemRequest{
		ProductID: productID,
		Quantity:  10,
	})
	if err != shared.ErrInsufficientInventory {
		t.Errorf("AddItem() error = %v, want %v", err, shared.ErrInsufficientInventory)
	}
}

func TestCartService_UpdateItem(t *testing.T) {
	cartRepo := NewMockCartRepository()
	cartItemRepo := NewMockCartItemRepository()
	invRepo := NewMockInventoryRepository()
	prodRepo := NewMockProductRepository()
		notificationService := NewMockNotificationService()
		service := NewCartService(cartRepo, cartItemRepo, invRepo, prodRepo, notificationService)

	userID := uuid.New()
	productID := uuid.New()

	// Setup product and inventory
	prodRepo.products[productID] = &products.Product{
		ID:    productID,
		Name:  "Test Product",
		Price: 19.99,
	}
	invRepo.inventory[productID] = &inventory.Inventory{
		ProductID: productID,
		Quantity:  100,
	}

	// Add item first
	item, err := service.AddItem(context.Background(), userID, AddItemRequest{
		ProductID: productID,
		Quantity:  2,
	})
	if err != nil {
		t.Fatalf("AddItem() error = %v", err)
	}

	// Update quantity
	updated, err := service.UpdateItem(context.Background(), userID, item.ID, 5)
	if err != nil {
		t.Fatalf("UpdateItem() error = %v", err)
	}
	if updated.Quantity != 5 {
		t.Errorf("UpdateItem() quantity = %v, want %v", updated.Quantity, 5)
	}
}

func TestCartService_RemoveItem(t *testing.T) {
	cartRepo := NewMockCartRepository()
	cartItemRepo := NewMockCartItemRepository()
	invRepo := NewMockInventoryRepository()
	prodRepo := NewMockProductRepository()
		notificationService := NewMockNotificationService()
		service := NewCartService(cartRepo, cartItemRepo, invRepo, prodRepo, notificationService)

	userID := uuid.New()
	productID := uuid.New()

	// Setup product and inventory
	prodRepo.products[productID] = &products.Product{
		ID:    productID,
		Name:  "Test Product",
		Price: 19.99,
	}
	invRepo.inventory[productID] = &inventory.Inventory{
		ProductID: productID,
		Quantity:  100,
	}

	// Add item first
	item, err := service.AddItem(context.Background(), userID, AddItemRequest{
		ProductID: productID,
		Quantity:  2,
	})
	if err != nil {
		t.Fatalf("AddItem() error = %v", err)
	}

	// Remove item
	err = service.RemoveItem(context.Background(), userID, item.ID)
	if err != nil {
		t.Errorf("RemoveItem() error = %v", err)
	}

	// Verify removal - should get empty cart
	cart, items, err := service.GetCart(context.Background(), userID)
	if err != nil {
		t.Fatalf("GetCart() error = %v", err)
	}
	_ = cart // cart exists but should have no items after removal
	if len(items) != 0 {
		t.Errorf("GetCart() after remove items = %v, want 0", len(items))
	}
}

// Property 8: Cart Operations Consistency
func TestProperty_CartOperationsConsistency(t *testing.T) {
	cartRepo := NewMockCartRepository()
	cartItemRepo := NewMockCartItemRepository()
	invRepo := NewMockInventoryRepository()
	prodRepo := NewMockProductRepository()
		notificationService := NewMockNotificationService()
		service := NewCartService(cartRepo, cartItemRepo, invRepo, prodRepo, notificationService)

	userID := uuid.New()
	productID := uuid.New()

	// Setup
	prodRepo.products[productID] = &products.Product{
		ID:    productID,
		Name:  "Test Product",
		Price: 10.00,
	}
	invRepo.inventory[productID] = &inventory.Inventory{
		ProductID: productID,
		Quantity:  100,
	}

	// Property: Adding increases cart contents
	item, err := service.AddItem(context.Background(), userID, AddItemRequest{
		ProductID: productID,
		Quantity:  3,
	})
	if err != nil {
		t.Fatalf("AddItem failed: %v", err)
	}

	_, items, _ := service.GetCart(context.Background(), userID)
	if len(items) != 1 {
		t.Errorf("After add, cart should have 1 item, got %d", len(items))
	}
	if items[0].Quantity != 3 {
		t.Errorf("After add, quantity should be 3, got %d", items[0].Quantity)
	}

	// Property: Updating changes quantity
	_, err = service.UpdateItem(context.Background(), userID, item.ID, 5)
	if err != nil {
		t.Fatalf("UpdateItem failed: %v", err)
	}

	_, items, _ = service.GetCart(context.Background(), userID)
	if items[0].Quantity != 5 {
		t.Errorf("After update, quantity should be 5, got %d", items[0].Quantity)
	}

	// Property: Removing decreases cart contents
	err = service.RemoveItem(context.Background(), userID, item.ID)
	if err != nil {
		t.Fatalf("RemoveItem failed: %v", err)
	}

	_, items, _ = service.GetCart(context.Background(), userID)
	if len(items) != 0 {
		t.Errorf("After remove, cart should be empty, got %d items", len(items))
	}
}

// Property 9: Cart Inventory Validation
func TestProperty_CartInventoryValidation(t *testing.T) {
	cartRepo := NewMockCartRepository()
	cartItemRepo := NewMockCartItemRepository()
	invRepo := NewMockInventoryRepository()
	prodRepo := NewMockProductRepository()
		notificationService := NewMockNotificationService()
		service := NewCartService(cartRepo, cartItemRepo, invRepo, prodRepo, notificationService)

	userID := uuid.New()
	productID := uuid.New()

	// Setup with limited inventory
	prodRepo.products[productID] = &products.Product{
		ID:    productID,
		Name:  "Limited Product",
		Price: 10.00,
	}
	invRepo.inventory[productID] = &inventory.Inventory{
		ProductID: productID,
		Quantity:  10,
	}

	// Property: Can add up to inventory limit
	_, err := service.AddItem(context.Background(), userID, AddItemRequest{
		ProductID: productID,
		Quantity:  10,
	})
	if err != nil {
		t.Errorf("Adding exact inventory amount should succeed: %v", err)
	}

	// Create fresh cart for next test
	userID2 := uuid.New()

	// Property: Cannot exceed inventory
	_, err = service.AddItem(context.Background(), userID2, AddItemRequest{
		ProductID: productID,
		Quantity:  11,
	})
	if err != shared.ErrInsufficientInventory {
		t.Errorf("Exceeding inventory should return ErrInsufficientInventory, got %v", err)
	}
}

// TestCartService_UpdateItem_Forbidden tests UpdateItem forbidden access
func TestCartService_UpdateItem_Forbidden(t *testing.T) {
	cartRepo := NewMockCartRepository()
	cartItemRepo := NewMockCartItemRepository()
	invRepo := NewMockInventoryRepository()
	prodRepo := NewMockProductRepository()
	notificationService := NewMockNotificationService()
	service := NewCartService(cartRepo, cartItemRepo, invRepo, prodRepo, notificationService)

	userID := uuid.New()
	cartID := uuid.New()
	otherCartID := uuid.New()
	itemID := uuid.New()
	productID := uuid.New()

	// Create cart for user
	cartRepo.carts[cartID] = &Cart{ID: cartID, UserID: userID}

	// Create item in different cart
	cartItemRepo.items[itemID] = &CartItem{
		ID:        itemID,
		CartID:    otherCartID,
		ProductID: productID,
		Quantity:  1,
	}

	_, err := service.UpdateItem(context.Background(), userID, itemID, 5)
	if err != shared.ErrForbidden {
		t.Errorf("UpdateItem() error = %v, wantErr %v", err, shared.ErrForbidden)
	}
}

// TestCartService_UpdateItem_InsufficientInventory tests UpdateItem with insufficient inventory
func TestCartService_UpdateItem_InsufficientInventory(t *testing.T) {
	cartRepo := NewMockCartRepository()
	cartItemRepo := NewMockCartItemRepository()
	invRepo := NewMockInventoryRepository()
	prodRepo := NewMockProductRepository()
	notificationService := NewMockNotificationService()
	service := NewCartService(cartRepo, cartItemRepo, invRepo, prodRepo, notificationService)

	userID := uuid.New()
	cartID := uuid.New()
	itemID := uuid.New()
	productID := uuid.New()

	// Setup cart and item
	cartRepo.carts[cartID] = &Cart{ID: cartID, UserID: userID}
	cartItemRepo.items[itemID] = &CartItem{
		ID:        itemID,
		CartID:    cartID,
		ProductID: productID,
		Quantity:  1,
	}

	// Setup inventory with limited quantity
	invRepo.inventory[productID] = &inventory.Inventory{
		ProductID: productID,
		Quantity:  5,
	}

	// Try to update to more than available
	_, err := service.UpdateItem(context.Background(), userID, itemID, 10)
	if err != shared.ErrInsufficientInventory {
		t.Errorf("UpdateItem() error = %v, wantErr %v", err, shared.ErrInsufficientInventory)
	}
}

// TestCartService_AddItem_UpdateExisting tests AddItem when product already exists in cart
func TestCartService_AddItem_UpdateExisting(t *testing.T) {
	cartRepo := NewMockCartRepository()
	cartItemRepo := NewMockCartItemRepository()
	invRepo := NewMockInventoryRepository()
	prodRepo := NewMockProductRepository()
	notificationService := NewMockNotificationService()
	service := NewCartService(cartRepo, cartItemRepo, invRepo, prodRepo, notificationService)

	userID := uuid.New()
	productID := uuid.New()

	// Setup product and inventory
	prodRepo.products[productID] = &products.Product{
		ID:    productID,
		Name:  "Test Product",
		Price: 19.99,
	}
	invRepo.inventory[productID] = &inventory.Inventory{
		ProductID: productID,
		Quantity:  100,
	}

	// Add item first time
	item1, err := service.AddItem(context.Background(), userID, AddItemRequest{
		ProductID: productID,
		Quantity:  2,
	})
	if err != nil {
		t.Fatalf("AddItem() first time error = %v", err)
	}

	// Add same product again - should update quantity
	item2, err := service.AddItem(context.Background(), userID, AddItemRequest{
		ProductID: productID,
		Quantity:  3,
	})
	if err != nil {
		t.Fatalf("AddItem() second time error = %v", err)
	}

	// Should be the same item with updated quantity
	if item2.ID != item1.ID {
		t.Error("AddItem() should return same item when product already in cart")
	}
	if item2.Quantity != 5 { // 2 + 3
		t.Errorf("AddItem() quantity = %v, want %v", item2.Quantity, 5)
	}
}

// TestCartService_RemoveItem_Forbidden tests RemoveItem forbidden access
func TestCartService_RemoveItem_Forbidden(t *testing.T) {
	cartRepo := NewMockCartRepository()
	cartItemRepo := NewMockCartItemRepository()
	invRepo := NewMockInventoryRepository()
	prodRepo := NewMockProductRepository()
	notificationService := NewMockNotificationService()
	service := NewCartService(cartRepo, cartItemRepo, invRepo, prodRepo, notificationService)

	userID := uuid.New()
	cartID := uuid.New()
	otherCartID := uuid.New()
	itemID := uuid.New()

	cartRepo.carts[cartID] = &Cart{ID: cartID, UserID: userID}
	cartItemRepo.items[itemID] = &CartItem{
		ID:     itemID,
		CartID: otherCartID,
	}

	err := service.RemoveItem(context.Background(), userID, itemID)
	if err != shared.ErrForbidden {
		t.Errorf("RemoveItem() error = %v, wantErr %v", err, shared.ErrForbidden)
	}
}

// TestCartService_ClearCart_NotFound tests ClearCart when cart doesn't exist
func TestCartService_ClearCart_NotFound(t *testing.T) {
	cartRepo := NewMockCartRepository()
	cartItemRepo := NewMockCartItemRepository()
	invRepo := NewMockInventoryRepository()
	prodRepo := NewMockProductRepository()
	notificationService := NewMockNotificationService()
	service := NewCartService(cartRepo, cartItemRepo, invRepo, prodRepo, notificationService)

	userID := uuid.New()

	// ClearCart should succeed even if cart doesn't exist (returns nil)
	err := service.ClearCart(context.Background(), userID)
	if err != nil {
		t.Errorf("ClearCart() when cart not found error = %v, want nil", err)
	}
}