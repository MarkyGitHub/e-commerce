package inventory

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/mschl/ecommerce-api/internal/shared"
)

// InventoryRepository defines the interface for inventory data access
type InventoryRepository interface {
	Create(ctx context.Context, inventory *Inventory) error
	FindByProductID(ctx context.Context, productID uuid.UUID) (*Inventory, error)
	Update(ctx context.Context, inventory *Inventory) error
	DecrementQuantity(ctx context.Context, productID uuid.UUID, amount int) error
	DecrementQuantityTx(ctx context.Context, tx *gorm.DB, productID uuid.UUID, amount int) error
}

// InventoryRepositoryImpl implements InventoryRepository using GORM
type InventoryRepositoryImpl struct {
	db *gorm.DB
}

// NewInventoryRepository creates a new InventoryRepository
func NewInventoryRepository(db *gorm.DB) InventoryRepository {
	return &InventoryRepositoryImpl{db: db}
}

func (r *InventoryRepositoryImpl) Create(ctx context.Context, inventory *Inventory) error {
	if inventory.ID == uuid.Nil {
		inventory.ID = uuid.New()
	}
	return r.db.WithContext(ctx).Create(inventory).Error
}

func (r *InventoryRepositoryImpl) FindByProductID(ctx context.Context, productID uuid.UUID) (*Inventory, error) {
	var inventory Inventory
	result := r.db.WithContext(ctx).Where("product_id = ?", productID).First(&inventory)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, shared.ErrNotFound
		}
		return nil, result.Error
	}
	return &inventory, nil
}

func (r *InventoryRepositoryImpl) Update(ctx context.Context, inventory *Inventory) error {
	result := r.db.WithContext(ctx).Save(inventory)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (r *InventoryRepositoryImpl) DecrementQuantity(ctx context.Context, productID uuid.UUID, amount int) error {
	return r.DecrementQuantityTx(ctx, r.db, productID, amount)
}

func (r *InventoryRepositoryImpl) DecrementQuantityTx(ctx context.Context, tx *gorm.DB, productID uuid.UUID, amount int) error {
	result := tx.WithContext(ctx).
		Model(&Inventory{}).
		Where("product_id = ? AND quantity >= ?", productID, amount).
		Update("quantity", gorm.Expr("quantity - ?", amount))

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return shared.ErrInsufficientInventory
	}
	return nil
}

// GetDB returns the underlying database connection for transactions
func (r *InventoryRepositoryImpl) GetDB() *gorm.DB {
	return r.db
}
