package orders

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func TestOrderRepository_CreateTx_AutoGenerateID(t *testing.T) {
	testDB := setupOrdersTestDB(t)
	orderRepo := NewOrderRepository(testDB).(*orderRepository)
	ctx := context.Background()

	order := &Order{ID: uuid.Nil, UserID: uuid.New(), TotalPrice: 1.23, Status: StatusPending}
	tx := testDB.Begin()
	if tx.Error != nil {
		t.Fatalf("begin tx: %v", tx.Error)
	}
	defer tx.Rollback()

	if err := orderRepo.CreateTx(ctx, tx, order); err != nil {
		t.Fatalf("CreateTx error: %v", err)
	}
	if order.ID == uuid.Nil {
		t.Fatalf("expected ID to be generated")
	}
}

func TestOrderRepository_GetDB(t *testing.T) {
	testDB := setupOrdersTestDB(t)
	repo := NewOrderRepository(testDB).(*orderRepository)
	if repo.GetDB() == nil {
		t.Fatalf("expected non-nil db")
	}
}

func TestOrderItemRepository_CreateBatchTx_AutoGenerateIDs(t *testing.T) {
	testDB := setupOrdersTestDB(t)
	repo := NewOrderItemRepository(testDB).(*orderItemRepository)
	ctx := context.Background()

	orderID := uuid.New()
	items := []OrderItem{{ID: uuid.Nil, OrderID: orderID, ProductID: uuid.New(), Quantity: 1, Price: 1.0}}

	var tx *gorm.DB = testDB
	if err := repo.CreateBatchTx(ctx, tx, items); err != nil {
		t.Fatalf("CreateBatchTx error: %v", err)
	}
	if items[0].ID == uuid.Nil {
		t.Fatalf("expected item ID to be generated")
	}
}
