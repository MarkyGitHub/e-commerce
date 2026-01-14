package orders

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/mschl/ecommerce-api/internal/testutil"
)

func TestOrderRepository_FindAll_CountError(t *testing.T) {
	db := testutil.SetupTestDB(t)
	repo := NewOrderRepository(db)

	_, _, err := repo.FindAll(context.Background(), 0, 10)
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestOrderRepository_FindAll_FindError_WhenColumnsMissing(t *testing.T) {
	db := testutil.SetupTestDB(t)

	// Create a deliberately incomplete orders table so Count succeeds,
	// but Find fails when selecting/ordering by missing columns.
	if err := db.Exec(`CREATE TABLE orders (id TEXT PRIMARY KEY, user_id TEXT NOT NULL)`).Error; err != nil {
		t.Fatalf("create table: %v", err)
	}

	repo := NewOrderRepository(db)
	_, _, err := repo.FindAll(context.Background(), 0, 10)
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestOrderRepository_FindByUserID_CountError(t *testing.T) {
	db := testutil.SetupTestDB(t)
	repo := NewOrderRepository(db)

	_, _, err := repo.FindByUserID(context.Background(), uuid.New(), 0, 10)
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestOrderRepository_FindByUserID_FindError_WhenColumnsMissing(t *testing.T) {
	db := testutil.SetupTestDB(t)
	if err := db.Exec(`CREATE TABLE orders (id TEXT PRIMARY KEY, user_id TEXT NOT NULL)`).Error; err != nil {
		t.Fatalf("create table: %v", err)
	}

	repo := NewOrderRepository(db)
	_, _, err := repo.FindByUserID(context.Background(), uuid.New(), 0, 10)
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestOrderRepository_FindByID_DBError(t *testing.T) {
	db := testutil.SetupTestDB(t)
	repo := NewOrderRepository(db)

	_, err := repo.FindByID(context.Background(), uuid.New())
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestOrderItemRepository_FindByOrderID_DBError(t *testing.T) {
	db := testutil.SetupTestDB(t)
	repo := NewOrderItemRepository(db)

	_, err := repo.FindByOrderID(context.Background(), uuid.New())
	if err == nil {
		t.Fatalf("expected error")
	}
}
