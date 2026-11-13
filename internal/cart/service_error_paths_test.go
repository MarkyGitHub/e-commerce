package cart

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/MarkyGitHub/e-commerce/internal/inventory"
	"github.com/MarkyGitHub/e-commerce/internal/shared"
	"gorm.io/gorm"
)

type failingCartRepo struct{ err error }

func (f failingCartRepo) FindOrCreateByUserID(ctx context.Context, userID uuid.UUID) (*Cart, error) {
	return nil, f.err
}
func (f failingCartRepo) FindByUserID(ctx context.Context, userID uuid.UUID) (*Cart, error) {
	return nil, f.err
}
func (f failingCartRepo) Clear(ctx context.Context, cartID uuid.UUID) error { return f.err }
func (f failingCartRepo) ClearTx(ctx context.Context, tx *gorm.DB, cartID uuid.UUID) error {
	return f.err
}

type failingCartItemRepoErr struct{ err error }

func (f failingCartItemRepoErr) Create(ctx context.Context, item *CartItem) error { return nil }
func (f failingCartItemRepoErr) FindByCartID(ctx context.Context, cartID uuid.UUID) ([]CartItem, error) {
	return nil, f.err
}
func (f failingCartItemRepoErr) FindByID(ctx context.Context, id uuid.UUID) (*CartItem, error) {
	return nil, f.err
}
func (f failingCartItemRepoErr) Update(ctx context.Context, item *CartItem) error { return f.err }
func (f failingCartItemRepoErr) Delete(ctx context.Context, id uuid.UUID) error   { return f.err }
func (f failingCartItemRepoErr) FindByCartAndProduct(ctx context.Context, cartID, productID uuid.UUID) (*CartItem, error) {
	return nil, shared.ErrNotFound
}

type failingInventoryRepo struct{ err error }

func (f failingInventoryRepo) Create(ctx context.Context, inv *inventory.Inventory) error { return nil }
func (f failingInventoryRepo) FindByProductID(ctx context.Context, productID uuid.UUID) (*inventory.Inventory, error) {
	return nil, f.err
}
func (f failingInventoryRepo) Update(ctx context.Context, inv *inventory.Inventory) error { return nil }
func (f failingInventoryRepo) DecrementQuantity(ctx context.Context, productID uuid.UUID, amount int) error {
	return nil
}
func (f failingInventoryRepo) DecrementQuantityTx(ctx context.Context, tx *gorm.DB, productID uuid.UUID, amount int) error {
	return nil
}

func TestCartService_GetCart_PropagatesCartRepoError(t *testing.T) {
	svc := NewCartService(failingCartRepo{err: errors.New("boom")}, NewMockCartItemRepository(), NewMockInventoryRepository(), NewMockProductRepository(), NewMockNotificationService())
	_, _, err := svc.GetCart(context.Background(), uuid.New())
	if err == nil || err.Error() != "boom" {
		t.Fatalf("expected boom error, got %v", err)
	}
}

func TestCartService_GetCart_PropagatesCartItemRepoError(t *testing.T) {
	userID := uuid.New()
	cartRepo := NewMockCartRepository()
	userCart, _ := cartRepo.FindOrCreateByUserID(context.Background(), userID)

	svc := NewCartService(cartRepo, failingCartItemRepoErr{err: errors.New("items")}, NewMockInventoryRepository(), NewMockProductRepository(), NewMockNotificationService())
	_, _, err := svc.GetCart(context.Background(), userCart.UserID)
	if err == nil || err.Error() != "items" {
		t.Fatalf("expected items error, got %v", err)
	}
}

func TestCartService_UpdateItem_PropagatesInventoryError(t *testing.T) {
	userID := uuid.New()
	productID := uuid.New()
	itemID := uuid.New()

	cartRepo := NewMockCartRepository()
	cartItemRepo := NewMockCartItemRepository()
	userCart, _ := cartRepo.FindOrCreateByUserID(context.Background(), userID)
	cartItemRepo.items[itemID] = &CartItem{ID: itemID, CartID: userCart.ID, ProductID: productID, Quantity: 1}

	svc := NewCartService(cartRepo, cartItemRepo, failingInventoryRepo{err: errors.New("inv")}, NewMockProductRepository(), NewMockNotificationService())
	_, err := svc.UpdateItem(context.Background(), userID, itemID, 1)
	if err == nil || err.Error() != "inv" {
		t.Fatalf("expected inv error, got %v", err)
	}
}
