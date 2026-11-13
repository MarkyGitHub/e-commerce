package orders

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/MarkyGitHub/e-commerce/internal/shared"
)

// OrderRepository definiert Interface für Bestell-Datenzugriff
type OrderRepository interface {
	Create(ctx context.Context, order *Order) error
	CreateTx(ctx context.Context, tx *gorm.DB, order *Order) error
	FindByID(ctx context.Context, id uuid.UUID) (*Order, error)
	FindByUserID(ctx context.Context, userID uuid.UUID, offset, limit int) ([]Order, int64, error)
	FindAll(ctx context.Context, offset, limit int) ([]Order, int64, error)
}

// OrderItemRepository definiert Interface für Bestellartikel-Datenzugriff
type OrderItemRepository interface {
	CreateBatch(ctx context.Context, items []OrderItem) error
	CreateBatchTx(ctx context.Context, tx *gorm.DB, items []OrderItem) error
	FindByOrderID(ctx context.Context, orderID uuid.UUID) ([]OrderItem, error)
}

// orderRepository implementiert OrderRepository mit GORM
type orderRepository struct {
	db *gorm.DB
}

// NewOrderRepository erstellt ein neues OrderRepository
func NewOrderRepository(db *gorm.DB) OrderRepository {
	return &orderRepository{db: db}
}

func (r *orderRepository) Create(ctx context.Context, order *Order) error {
	return r.CreateTx(ctx, r.db, order)
}

func (r *orderRepository) CreateTx(ctx context.Context, tx *gorm.DB, order *Order) error {
	if order.ID == uuid.Nil {
		order.ID = uuid.New()
	}
	return tx.WithContext(ctx).Create(order).Error
}

func (r *orderRepository) FindByID(ctx context.Context, id uuid.UUID) (*Order, error) {
	var order Order
	result := r.db.WithContext(ctx).First(&order, "id = ?", id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, shared.ErrNotFound
		}
		return nil, result.Error
	}
	return &order, nil
}

func (r *orderRepository) FindByUserID(ctx context.Context, userID uuid.UUID, offset, limit int) ([]Order, int64, error) {
	var orders []Order
	var totalCount int64

	// Gesamtzahl ermitteln
	if err := r.db.WithContext(ctx).Model(&Order{}).Where("user_id = ?", userID).Count(&totalCount).Error; err != nil {
		return nil, 0, err
	}

	// Paginierte Ergebnisse holen
	result := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&orders)

	if result.Error != nil {
		return nil, 0, result.Error
	}

	return orders, totalCount, nil
}

func (r *orderRepository) FindAll(ctx context.Context, offset, limit int) ([]Order, int64, error) {
	var orders []Order
	var totalCount int64

	// Gesamtzahl ermitteln
	if err := r.db.WithContext(ctx).Model(&Order{}).Count(&totalCount).Error; err != nil {
		return nil, 0, err
	}

	// Paginierte Ergebnisse holen
	result := r.db.WithContext(ctx).
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&orders)

	if result.Error != nil {
		return nil, 0, result.Error
	}

	return orders, totalCount, nil
}

// GetDB gibt die Datenbankverbindung zurück
func (r *orderRepository) GetDB() *gorm.DB {
	return r.db
}

// orderItemRepository implementiert OrderItemRepository mit GORM
type orderItemRepository struct {
	db *gorm.DB
}

// NewOrderItemRepository erstellt ein neues OrderItemRepository
func NewOrderItemRepository(db *gorm.DB) OrderItemRepository {
	return &orderItemRepository{db: db}
}

func (r *orderItemRepository) CreateBatch(ctx context.Context, items []OrderItem) error {
	return r.CreateBatchTx(ctx, r.db, items)
}

func (r *orderItemRepository) CreateBatchTx(ctx context.Context, tx *gorm.DB, items []OrderItem) error {
	for i := range items {
		if items[i].ID == uuid.Nil {
			items[i].ID = uuid.New()
		}
	}
	return tx.WithContext(ctx).Create(&items).Error
}

func (r *orderItemRepository) FindByOrderID(ctx context.Context, orderID uuid.UUID) ([]OrderItem, error) {
	var items []OrderItem
	result := r.db.WithContext(ctx).Where("order_id = ?", orderID).Find(&items)
	if result.Error != nil {
		return nil, result.Error
	}
	return items, nil
}
