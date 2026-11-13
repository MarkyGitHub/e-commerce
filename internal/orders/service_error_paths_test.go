package orders

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/MarkyGitHub/e-commerce/internal/cart"
	"github.com/MarkyGitHub/e-commerce/internal/notifications"
	"github.com/MarkyGitHub/e-commerce/internal/products"
	"github.com/MarkyGitHub/e-commerce/internal/shared"
	"gorm.io/gorm"
)

type failingNotifService struct{}

func (f failingNotifService) CreateNotification(ctx context.Context, userID uuid.UUID, title, message string) (*notifications.Notification, error) {
	return nil, errors.New("notif")
}
func (f failingNotifService) CreateNotificationTx(ctx context.Context, tx *gorm.DB, userID uuid.UUID, title, message string) (*notifications.Notification, error) {
	return nil, errors.New("notif")
}
func (f failingNotifService) GetNotifications(ctx context.Context, userID uuid.UUID) ([]notifications.Notification, error) {
	return nil, nil
}
func (f failingNotifService) MarkAsRead(ctx context.Context, userID uuid.UUID, notificationID uuid.UUID) error {
	return nil
}

func TestOrderService_CreateOrder_ProductNotFound(t *testing.T) {
	userID := uuid.New()
	productID := uuid.New()
	cartID := uuid.New()

	cartRepo := NewMockCartRepository()
	cartItemRepo := NewMockCartItemRepository()
	cartRepo.carts[cartID] = &cart.Cart{ID: cartID, UserID: userID}
	cartItemRepo.items[uuid.New()] = &cart.CartItem{ID: uuid.New(), CartID: cartID, ProductID: productID, Quantity: 1}

	orderRepo := NewMockOrderRepository()
	orderItemRepo := NewMockOrderItemRepository()
	productRepo := NewMockProductRepository() // product missing
	inventoryRepo := NewMockInventoryRepository()
	inventoryRepo.quantities[productID] = 10
	notif := NewMockNotificationService()

	svc := newTestOrderService(orderRepo, orderItemRepo, cartRepo, cartItemRepo, productRepo, inventoryRepo, notif)
	addr := CreateOrderRequest{ShippingStreet: "s", ShippingCity: "c", ShippingZip: "z", ShippingCountry: "x"}
	_, _, err := svc.CreateOrder(context.Background(), userID, addr)
	if err != shared.ErrNotFound {
		t.Fatalf("expected %v, got %v", shared.ErrNotFound, err)
	}
}

func TestOrderService_CreateOrder_InsufficientInventory(t *testing.T) {
	userID := uuid.New()
	productID := uuid.New()
	cartID := uuid.New()

	cartRepo := NewMockCartRepository()
	cartItemRepo := NewMockCartItemRepository()
	cartRepo.carts[cartID] = &cart.Cart{ID: cartID, UserID: userID}
	cartItemRepo.items[uuid.New()] = &cart.CartItem{ID: uuid.New(), CartID: cartID, ProductID: productID, Quantity: 5}

	orderRepo := NewMockOrderRepository()
	orderItemRepo := NewMockOrderItemRepository()
	productRepo := NewMockProductRepository()
	productRepo.products[productID] = &products.Product{ID: productID, Name: "P", Price: 1}
	inventoryRepo := NewMockInventoryRepository()
	inventoryRepo.quantities[productID] = 1
	notif := NewMockNotificationService()

	svc := newTestOrderService(orderRepo, orderItemRepo, cartRepo, cartItemRepo, productRepo, inventoryRepo, notif)
	addr := CreateOrderRequest{ShippingStreet: "s", ShippingCity: "c", ShippingZip: "z", ShippingCountry: "x"}
	_, _, err := svc.CreateOrder(context.Background(), userID, addr)
	if err != shared.ErrInsufficientInventory {
		t.Fatalf("expected %v, got %v", shared.ErrInsufficientInventory, err)
	}
}

func TestNewOrderService_TransactionExecutes(t *testing.T) {
	// Use a real sqlite DB so db.Transaction is functional.
	testDB := setupOrdersTestDB(t)
	svc := NewOrderService(testDB, nil, nil, nil, nil, nil, nil, nil).(*orderService)

	if err := svc.transaction(func(tx *gorm.DB) error { return nil }); err != nil {
		t.Fatalf("expected transaction to succeed, got %v", err)
	}
}

func TestOrderService_CreateOrder_NotificationFailureBubblesUp(t *testing.T) {
	userID := uuid.New()
	productID := uuid.New()
	cartID := uuid.New()

	cartRepo := NewMockCartRepository()
	cartItemRepo := NewMockCartItemRepository()
	cartRepo.carts[cartID] = &cart.Cart{ID: cartID, UserID: userID}
	cartItemRepo.items[uuid.New()] = &cart.CartItem{ID: uuid.New(), CartID: cartID, ProductID: productID, Quantity: 1}

	orderRepo := NewMockOrderRepository()
	orderItemRepo := NewMockOrderItemRepository()
	productRepo := NewMockProductRepository()
	productRepo.products[productID] = &products.Product{ID: productID, Name: "P", Price: 1}
	inventoryRepo := NewMockInventoryRepository()
	inventoryRepo.quantities[productID] = 10

	svc := &orderService{
		transaction: func(fn func(*gorm.DB) error) error { return fn(mockTransaction()) },
		orderRepo:   orderRepo,
		orderItemRepo: orderItemRepo,
		cartRepo:      cartRepo,
		cartItemRepo:  cartItemRepo,
		productRepo:   productRepo,
		inventoryRepo: inventoryRepo,
		notificationService: failingNotifService{},
	}

	addr := CreateOrderRequest{ShippingStreet: "s", ShippingCity: "c", ShippingZip: "z", ShippingCountry: "x"}
	_, _, err := svc.CreateOrder(context.Background(), userID, addr)
	if err == nil || err.Error() != "notif" {
		t.Fatalf("expected notif error, got %v", err)
	}
}
