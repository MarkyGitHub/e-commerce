package products

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/MarkyGitHub/e-commerce/internal/shared"
)

// ProductRepository definiert Interface für Produkt-Datenzugriff
type ProductRepository interface {
	Create(ctx context.Context, product *Product) error
	FindByID(ctx context.Context, id uuid.UUID) (*Product, error)
	FindAll(ctx context.Context, offset, limit int, categoryID *uuid.UUID) ([]Product, int64, error)
	Update(ctx context.Context, product *Product) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// productRepository implementiert ProductRepository mit GORM
type productRepository struct {
	db *gorm.DB
}

// NewProductRepository erstellt ein neues ProductRepository
func NewProductRepository(db *gorm.DB) ProductRepository {
	return &productRepository{db: db}
}

func (r *productRepository) Create(ctx context.Context, product *Product) error {
	if product.ID == uuid.Nil {
		product.ID = uuid.New()
	}
	return r.db.WithContext(ctx).Create(product).Error
}

func (r *productRepository) FindByID(ctx context.Context, id uuid.UUID) (*Product, error) {
	var product Product
	result := r.db.WithContext(ctx).First(&product, "id = ?", id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, shared.ErrNotFound
		}
		return nil, result.Error
	}
	return &product, nil
}

func (r *productRepository) FindAll(ctx context.Context, offset, limit int, categoryID *uuid.UUID) ([]Product, int64, error) {
	var products []Product
	var totalCount int64

	query := r.db.WithContext(ctx).Model(&Product{}).Where("is_active = ?", true)

	// Nach Kategorie filtern falls angegeben
	if categoryID != nil {
		query = query.Where("category_id = ?", *categoryID)
	}

	// Gesamtzahl ermitteln
	if err := query.Count(&totalCount).Error; err != nil {
		return nil, 0, err
	}

	// Paginierte Ergebnisse holen
	result := query.
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&products)

	if result.Error != nil {
		return nil, 0, result.Error
	}

	return products, totalCount, nil
}

func (r *productRepository) Update(ctx context.Context, product *Product) error {
	// GORM's Save() will INSERT when the primary key does not exist.
	// For this repository contract, Update must not create new rows.
	var count int64
	if err := r.db.WithContext(ctx).Model(&Product{}).Where("id = ?", product.ID).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return shared.ErrNotFound
	}

	return r.db.WithContext(ctx).Save(product).Error
}

func (r *productRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&Product{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return shared.ErrNotFound
	}
	return nil
}
