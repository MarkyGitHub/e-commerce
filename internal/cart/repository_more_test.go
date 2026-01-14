package cart

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

func TestCartRepository_GetDB(t *testing.T) {
	db := setupCartTestDB(t)
	repo := NewCartRepository(db)
	impl, ok := repo.(*CartRepositoryImpl)
	if !ok {
		t.Fatalf("expected *CartRepositoryImpl, got %T", repo)
	}
	if impl.GetDB() == nil {
		t.Fatalf("expected non-nil db")
	}
}

func TestCartItemRepository_Create_AssignsID(t *testing.T) {
	db := setupCartTestDB(t)
	repo := NewCartItemRepository(db)

	item := &CartItem{ID: uuid.Nil, CartID: uuid.New(), ProductID: uuid.New(), Quantity: 1}
	if err := repo.Create(context.Background(), item); err != nil {
		t.Fatalf("create: %v", err)
	}
	if item.ID == uuid.Nil {
		t.Fatalf("expected Create to assign a non-nil ID")
	}
}
