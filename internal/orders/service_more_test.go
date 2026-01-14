package orders

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/mschl/ecommerce-api/internal/cart"
	"github.com/mschl/ecommerce-api/internal/products"
	"gorm.io/gorm"
)

func TestOrderService_CreateOrder_TransactionNotConfigured(t *testing.T) {
	userID := uuid.New()
	cartID := uuid.New()

	orderRepo := NewMockOrderRepository()
	orderItemRepo := NewMockOrderItemRepository()
	cartRepo := NewMockCartRepository()
	cartItemRepo := NewMockCartItemRepository()
	productRepo := NewMockProductRepository()
	inventoryRepo := NewMockInventoryRepository()
	notificationService := NewMockNotificationService()

	cartRepo.carts[cartID] = &cart.Cart{ID: cartID, UserID: userID}
	cartItemRepo.items[uuid.New()] = &cart.CartItem{ID: uuid.New(), CartID: cartID, ProductID: uuid.New(), Quantity: 1}

	svc := newTestOrderService(orderRepo, orderItemRepo, cartRepo, cartItemRepo, productRepo, inventoryRepo, notificationService)
	svc.transaction = nil

	addr := CreateOrderRequest{ShippingStreet: "s", ShippingCity: "c", ShippingZip: "z", ShippingCountry: "x"}
	_, _, err := svc.CreateOrder(context.Background(), userID, addr)
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestOrderService_CreateOrder_NonTxBranch(t *testing.T) {
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

	cartRepo.carts[cartID] = &cart.Cart{ID: cartID, UserID: userID}
	cartItemRepo.items[uuid.New()] = &cart.CartItem{ID: uuid.New(), CartID: cartID, ProductID: productID, Quantity: 2}

	productRepo.products[productID] = &products.Product{ID: productID, Name: "P", Price: 5.0}
	inventoryRepo.quantities[productID] = 10

	svc := newTestOrderService(orderRepo, orderItemRepo, cartRepo, cartItemRepo, productRepo, inventoryRepo, notificationService)
	// Force tx == nil path inside the transaction closure.
	svc.transaction = func(fn func(*gorm.DB) error) error { return fn(nil) }

	addr := CreateOrderRequest{ShippingStreet: "s", ShippingCity: "c", ShippingZip: "z", ShippingCountry: "x"}
	order, items, err := svc.CreateOrder(context.Background(), userID, addr)
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if order == nil || len(items) != 1 {
		t.Fatalf("expected order and 1 item")
	}
}

func TestNewOrderService_SetsTransactionHandler(t *testing.T) {
	svc := NewOrderService(&gorm.DB{}, nil, nil, nil, nil, nil, nil, nil)
	impl, ok := svc.(*orderService)
	if !ok {
		t.Fatalf("expected *orderService implementation")
	}
	if impl.transaction == nil {
		t.Fatalf("expected transaction handler to be configured")
	}
}
