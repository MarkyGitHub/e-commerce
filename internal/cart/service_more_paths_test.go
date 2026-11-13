package cart

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/MarkyGitHub/e-commerce/internal/shared"
	"gorm.io/gorm"
)

type cartRepoFindByUserErr struct{ err error }

func (c cartRepoFindByUserErr) FindOrCreateByUserID(ctx context.Context, userID uuid.UUID) (*Cart, error) {
	return &Cart{ID: uuid.New(), UserID: userID}, nil
}
func (c cartRepoFindByUserErr) FindByUserID(ctx context.Context, userID uuid.UUID) (*Cart, error) {
	return nil, c.err
}
func (c cartRepoFindByUserErr) Clear(ctx context.Context, cartID uuid.UUID) error { return nil }
func (c cartRepoFindByUserErr) ClearTx(ctx context.Context, tx *gorm.DB, cartID uuid.UUID) error {
	return nil
}

type cartRepoClearErr struct{ err error }

func (c cartRepoClearErr) FindOrCreateByUserID(ctx context.Context, userID uuid.UUID) (*Cart, error) {
	return &Cart{ID: uuid.New(), UserID: userID}, nil
}
func (c cartRepoClearErr) FindByUserID(ctx context.Context, userID uuid.UUID) (*Cart, error) {
	return &Cart{ID: uuid.New(), UserID: userID}, nil
}
func (c cartRepoClearErr) Clear(ctx context.Context, cartID uuid.UUID) error { return c.err }
func (c cartRepoClearErr) ClearTx(ctx context.Context, tx *gorm.DB, cartID uuid.UUID) error {
	return c.err
}

type cartItemRepoFindErr struct{ err error }

func (c cartItemRepoFindErr) Create(ctx context.Context, item *CartItem) error { return nil }
func (c cartItemRepoFindErr) FindByCartID(ctx context.Context, cartID uuid.UUID) ([]CartItem, error) {
	return nil, nil
}
func (c cartItemRepoFindErr) FindByID(ctx context.Context, id uuid.UUID) (*CartItem, error) {
	return nil, c.err
}
func (c cartItemRepoFindErr) Update(ctx context.Context, item *CartItem) error { return nil }
func (c cartItemRepoFindErr) Delete(ctx context.Context, id uuid.UUID) error   { return nil }
func (c cartItemRepoFindErr) FindByCartAndProduct(ctx context.Context, cartID, productID uuid.UUID) (*CartItem, error) {
	return nil, shared.ErrNotFound
}

type cartRepoFixed struct{ cart *Cart }

func (c cartRepoFixed) FindOrCreateByUserID(ctx context.Context, userID uuid.UUID) (*Cart, error) {
	return c.cart, nil
}
func (c cartRepoFixed) FindByUserID(ctx context.Context, userID uuid.UUID) (*Cart, error) {
	return c.cart, nil
}
func (c cartRepoFixed) Clear(ctx context.Context, cartID uuid.UUID) error { return nil }
func (c cartRepoFixed) ClearTx(ctx context.Context, tx *gorm.DB, cartID uuid.UUID) error {
	return nil
}

type cartItemRepoDeleteErrForCart struct {
	cartID uuid.UUID
	err    error
}

func (c cartItemRepoDeleteErrForCart) Create(ctx context.Context, item *CartItem) error { return nil }
func (c cartItemRepoDeleteErrForCart) FindByCartID(ctx context.Context, cartID uuid.UUID) ([]CartItem, error) {
	return nil, nil
}
func (c cartItemRepoDeleteErrForCart) FindByID(ctx context.Context, id uuid.UUID) (*CartItem, error) {
	return &CartItem{ID: id, CartID: c.cartID, ProductID: uuid.New(), Quantity: 1}, nil
}
func (c cartItemRepoDeleteErrForCart) Update(ctx context.Context, item *CartItem) error { return nil }
func (c cartItemRepoDeleteErrForCart) Delete(ctx context.Context, id uuid.UUID) error   { return c.err }
func (c cartItemRepoDeleteErrForCart) FindByCartAndProduct(ctx context.Context, cartID, productID uuid.UUID) (*CartItem, error) {
	return nil, shared.ErrNotFound
}

func TestCartService_RemoveItem_PropagatesCartRepoError(t *testing.T) {
	svc := NewCartService(cartRepoFindByUserErr{err: errors.New("cart")}, NewMockCartItemRepository(), NewMockInventoryRepository(), NewMockProductRepository(), NewMockNotificationService())
	if err := svc.RemoveItem(context.Background(), uuid.New(), uuid.New()); err == nil || err.Error() != "cart" {
		t.Fatalf("expected cart error, got %v", err)
	}
}

func TestCartService_RemoveItem_PropagatesItemLookupError(t *testing.T) {
	userID := uuid.New()
	cartRepo := NewMockCartRepository()
	_, _ = cartRepo.FindOrCreateByUserID(context.Background(), userID)

	svc := NewCartService(cartRepo, cartItemRepoFindErr{err: errors.New("item")}, NewMockInventoryRepository(), NewMockProductRepository(), NewMockNotificationService())
	if err := svc.RemoveItem(context.Background(), userID, uuid.New()); err == nil || err.Error() != "item" {
		t.Fatalf("expected item error, got %v", err)
	}
}

func TestCartService_RemoveItem_PropagatesDeleteError(t *testing.T) {
	userID := uuid.New()
	itemID := uuid.New()
	cartID := uuid.New()

	svc := NewCartService(
		cartRepoFixed{cart: &Cart{ID: cartID, UserID: userID}},
		cartItemRepoDeleteErrForCart{cartID: cartID, err: errors.New("delete")},
		NewMockInventoryRepository(),
		NewMockProductRepository(),
		NewMockNotificationService(),
	)

	if err := svc.RemoveItem(context.Background(), userID, itemID); err == nil || err.Error() != "delete" {
		t.Fatalf("expected delete error, got %v", err)
	}
}

func TestCartService_ClearCart_PropagatesFindError(t *testing.T) {
	svc := NewCartService(cartRepoFindByUserErr{err: errors.New("find")}, NewMockCartItemRepository(), NewMockInventoryRepository(), NewMockProductRepository(), NewMockNotificationService())
	if err := svc.ClearCart(context.Background(), uuid.New()); err == nil || err.Error() != "find" {
		t.Fatalf("expected find error, got %v", err)
	}
}

func TestCartService_ClearCart_PropagatesClearError(t *testing.T) {
	svc := NewCartService(cartRepoClearErr{err: errors.New("clear")}, NewMockCartItemRepository(), NewMockInventoryRepository(), NewMockProductRepository(), NewMockNotificationService())
	if err := svc.ClearCart(context.Background(), uuid.New()); err == nil || err.Error() != "clear" {
		t.Fatalf("expected clear error, got %v", err)
	}
}
