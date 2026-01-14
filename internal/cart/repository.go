package cart

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/mschl/ecommerce-api/internal/shared"
)

// CartRepository defines the interface for cart data access
type CartRepository interface {
	FindOrCreateByUserID(ctx context.Context, userID uuid.UUID) (*Cart, error)
	FindByUserID(ctx context.Context, userID uuid.UUID) (*Cart, error)
	Clear(ctx context.Context, cartID uuid.UUID) error
	ClearTx(ctx context.Context, tx *gorm.DB, cartID uuid.UUID) error
}

// CartItemRepository defines the interface for cart item data access
type CartItemRepository interface {
	Create(ctx context.Context, item *CartItem) error
	FindByCartID(ctx context.Context, cartID uuid.UUID) ([]CartItem, error)
	FindByID(ctx context.Context, id uuid.UUID) (*CartItem, error)
	Update(ctx context.Context, item *CartItem) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindByCartAndProduct(ctx context.Context, cartID, productID uuid.UUID) (*CartItem, error)
}

// CartRepositoryImpl implements CartRepository using GORM
type CartRepositoryImpl struct {
	db *gorm.DB
}

// NewCartRepository creates a new CartRepository
func NewCartRepository(db *gorm.DB) CartRepository {
	return &CartRepositoryImpl{db: db}
}

func (r *CartRepositoryImpl) FindOrCreateByUserID(ctx context.Context, userID uuid.UUID) (*Cart, error) {
	var cart Cart
	result := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&cart)
	if result.Error == nil {
		return &cart, nil
	}
	if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, result.Error
	}

	// Create new cart
	cart = Cart{
		ID:     uuid.New(),
		UserID: userID,
	}
	if err := r.db.WithContext(ctx).Create(&cart).Error; err != nil {
		return nil, err
	}
	return &cart, nil
}

func (r *CartRepositoryImpl) FindByUserID(ctx context.Context, userID uuid.UUID) (*Cart, error) {
	var cart Cart
	result := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&cart)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, shared.ErrNotFound
		}
		return nil, result.Error
	}
	return &cart, nil
}

func (r *CartRepositoryImpl) Clear(ctx context.Context, cartID uuid.UUID) error {
	return r.ClearTx(ctx, r.db, cartID)
}

func (r *CartRepositoryImpl) ClearTx(ctx context.Context, tx *gorm.DB, cartID uuid.UUID) error {
	return tx.WithContext(ctx).Where("cart_id = ?", cartID).Delete(&CartItem{}).Error
}

// GetDB returns the underlying database connection
func (r *CartRepositoryImpl) GetDB() *gorm.DB {
	return r.db
}

// cartItemRepository implements CartItemRepository using GORM
type cartItemRepository struct {
	db *gorm.DB
}

// NewCartItemRepository creates a new CartItemRepository
func NewCartItemRepository(db *gorm.DB) CartItemRepository {
	return &cartItemRepository{db: db}
}

func (r *cartItemRepository) Create(ctx context.Context, item *CartItem) error {
	if item.ID == uuid.Nil {
		item.ID = uuid.New()
	}
	return r.db.WithContext(ctx).Create(item).Error
}

func (r *cartItemRepository) FindByCartID(ctx context.Context, cartID uuid.UUID) ([]CartItem, error) {
	var items []CartItem
	result := r.db.WithContext(ctx).Where("cart_id = ?", cartID).Find(&items)
	if result.Error != nil {
		return nil, result.Error
	}
	return items, nil
}

func (r *cartItemRepository) FindByID(ctx context.Context, id uuid.UUID) (*CartItem, error) {
	var item CartItem
	result := r.db.WithContext(ctx).First(&item, "id = ?", id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, shared.ErrNotFound
		}
		return nil, result.Error
	}
	return &item, nil
}

func (r *cartItemRepository) Update(ctx context.Context, item *CartItem) error {
	return r.db.WithContext(ctx).Save(item).Error
}

func (r *cartItemRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&CartItem{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return shared.ErrNotFound
	}
	return nil
}

func (r *cartItemRepository) FindByCartAndProduct(ctx context.Context, cartID, productID uuid.UUID) (*CartItem, error) {
	var item CartItem
	result := r.db.WithContext(ctx).Where("cart_id = ? AND product_id = ?", cartID, productID).First(&item)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, shared.ErrNotFound
		}
		return nil, result.Error
	}
	return &item, nil
}
