package cart

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/MarkyGitHub/e-commerce/internal/inventory"
	"github.com/MarkyGitHub/e-commerce/internal/products"
	"github.com/MarkyGitHub/e-commerce/internal/shared"
)

type failingCartItemRepo struct{ err error }

func (f failingCartItemRepo) Create(ctx context.Context, item *CartItem) error { return nil }
func (f failingCartItemRepo) FindByCartID(ctx context.Context, cartID uuid.UUID) ([]CartItem, error) {
	return nil, nil
}
func (f failingCartItemRepo) FindByID(ctx context.Context, id uuid.UUID) (*CartItem, error) {
	return nil, nil
}
func (f failingCartItemRepo) Update(ctx context.Context, item *CartItem) error { return nil }
func (f failingCartItemRepo) Delete(ctx context.Context, id uuid.UUID) error   { return nil }
func (f failingCartItemRepo) FindByCartAndProduct(ctx context.Context, cartID, productID uuid.UUID) (*CartItem, error) {
	return nil, f.err
}

func TestCartService_AddItem_InventoryNotFoundMapsToInsufficient(t *testing.T) {
	cartRepo := NewMockCartRepository()
	cartItemRepo := NewMockCartItemRepository()
	prodRepo := NewMockProductRepository()
	notificationService := NewMockNotificationService()

	userID := uuid.New()
	productID := uuid.New()
	prodRepo.products[productID] = &products.Product{ID: productID, Name: "P", Price: 1}

	service := NewCartService(cartRepo, cartItemRepo, NewMockInventoryRepository(), prodRepo, notificationService)

	_, err := service.AddItem(context.Background(), userID, AddItemRequest{ProductID: productID, Quantity: 1})
	if err != shared.ErrInsufficientInventory {
		t.Fatalf("expected %v, got %v", shared.ErrInsufficientInventory, err)
	}
}

func TestCartService_AddItem_ExistingItem_InsufficientForNewQuantity(t *testing.T) {
	cartRepo := NewMockCartRepository()
	cartItemRepo := NewMockCartItemRepository()
	invRepo := NewMockInventoryRepository()
	prodRepo := NewMockProductRepository()
	notificationService := NewMockNotificationService()
	service := NewCartService(cartRepo, cartItemRepo, invRepo, prodRepo, notificationService)

	userID := uuid.New()
	productID := uuid.New()

	// Product exists
	prodRepo.products[productID] = &products.Product{ID: productID, Name: "P", Price: 1}
	// Inventory has only 5
	invRepo.inventory[productID] = &inventory.Inventory{ID: uuid.New(), ProductID: productID, Quantity: 5}

	cart, err := cartRepo.FindOrCreateByUserID(context.Background(), userID)
	if err != nil {
		t.Fatalf("FindOrCreateByUserID error: %v", err)
	}

	// Existing item has quantity 4; adding 2 would exceed inventory.
	existingID := uuid.New()
	cartItemRepo.items[existingID] = &CartItem{ID: existingID, CartID: cart.ID, ProductID: productID, Quantity: 4}

	_, err = service.AddItem(context.Background(), userID, AddItemRequest{ProductID: productID, Quantity: 2})
	if err != shared.ErrInsufficientInventory {
		t.Fatalf("expected %v, got %v", shared.ErrInsufficientInventory, err)
	}
}

func TestCartService_AddItem_UnexpectedFindByCartAndProductError(t *testing.T) {
	cartRepo := NewMockCartRepository()
	invRepo := NewMockInventoryRepository()
	prodRepo := NewMockProductRepository()
	notificationService := NewMockNotificationService()

	userID := uuid.New()
	productID := uuid.New()
	prodRepo.products[productID] = &products.Product{ID: productID, Name: "P", Price: 1}
	invRepo.inventory[productID] = &inventory.Inventory{ID: uuid.New(), ProductID: productID, Quantity: 10}

	service := NewCartService(cartRepo, failingCartItemRepo{err: errors.New("boom")}, invRepo, prodRepo, notificationService)
	_, err := service.AddItem(context.Background(), userID, AddItemRequest{ProductID: productID, Quantity: 1})
	if err == nil || err.Error() != "boom" {
		t.Fatalf("expected boom error, got %v", err)
	}
}

func TestCartService_UpdateItem_ForbiddenWhenItemNotInUsersCart(t *testing.T) {
	cartRepo := NewMockCartRepository()
	cartItemRepo := NewMockCartItemRepository()
	invRepo := NewMockInventoryRepository()
	prodRepo := NewMockProductRepository()
	notificationService := NewMockNotificationService()
	service := NewCartService(cartRepo, cartItemRepo, invRepo, prodRepo, notificationService)

	user1 := uuid.New()
	user2 := uuid.New()
	productID := uuid.New()

	// Ensure user carts exist
	cart1, _ := cartRepo.FindOrCreateByUserID(context.Background(), user1)
	cart2, _ := cartRepo.FindOrCreateByUserID(context.Background(), user2)

	itemID := uuid.New()
	cartItemRepo.items[itemID] = &CartItem{ID: itemID, CartID: cart2.ID, ProductID: productID, Quantity: 1}
	invRepo.inventory[productID] = &inventory.Inventory{ID: uuid.New(), ProductID: productID, Quantity: 10}
	_ = cart1

	_, err := service.UpdateItem(context.Background(), user1, itemID, 2)
	if err != shared.ErrForbidden {
		t.Fatalf("expected %v, got %v", shared.ErrForbidden, err)
	}
}

func TestCartService_ClearCart_NotFoundIsNoop(t *testing.T) {
	service := NewCartService(NewMockCartRepository(), NewMockCartItemRepository(), NewMockInventoryRepository(), NewMockProductRepository(), NewMockNotificationService())

	if err := service.ClearCart(context.Background(), uuid.New()); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}
