package products

import (
	"context"

	"github.com/google/uuid"

	"github.com/MarkyGitHub/e-commerce/internal/notifications"
)

// ProductService definiert Interface für Produkt-Geschäftslogik
type ProductService interface {
	CreateProduct(ctx context.Context, req CreateProductRequest, userID uuid.UUID) (*Product, error)
	GetProduct(ctx context.Context, id uuid.UUID) (*Product, error)
	ListProducts(ctx context.Context, page, limit int, categoryID *uuid.UUID) ([]Product, int64, error)
	UpdateProduct(ctx context.Context, id uuid.UUID, req UpdateProductRequest) (*Product, error)
	DeleteProduct(ctx context.Context, id uuid.UUID) error
}

// productService implementiert ProductService
type productService struct {
	repo                ProductRepository
	notificationService notifications.NotificationService
}

// NewProductService erstellt einen neuen ProductService
func NewProductService(repo ProductRepository, notificationService notifications.NotificationService) ProductService {
	return &productService{
		repo:                repo,
		notificationService: notificationService,
	}
}

func (s *productService) CreateProduct(ctx context.Context, req CreateProductRequest, userID uuid.UUID) (*Product, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	product := &Product{
		ID:          uuid.New(),
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		CategoryID:  req.CategoryID,
		IsActive:    true,
	}

	if err := s.repo.Create(ctx, product); err != nil {
		return nil, err
	}

	// Benachrichtigung für Produkterstellung erstellen (Fehler ignorieren – Erstellung nicht abbrechen)
	title := "Product Created"
	message := "Product created successfully."
	_, _ = s.notificationService.CreateNotification(ctx, userID, title, message)

	return product, nil
}

func (s *productService) GetProduct(ctx context.Context, id uuid.UUID) (*Product, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *productService) ListProducts(ctx context.Context, page, limit int, categoryID *uuid.UUID) ([]Product, int64, error) {
	offset := (page - 1) * limit
	return s.repo.FindAll(ctx, offset, limit, categoryID)
}

func (s *productService) UpdateProduct(ctx context.Context, id uuid.UUID, req UpdateProductRequest) (*Product, error) {
	product, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Felder aktualisieren falls angegeben
	if req.Name != nil {
		if *req.Name == "" {
			return nil, ErrInvalidProductName
		}
		product.Name = *req.Name
	}
	if req.Description != nil {
		product.Description = *req.Description
	}
	if req.Price != nil {
		if *req.Price < 0 {
			return nil, ErrInvalidProductPrice
		}
		product.Price = *req.Price
	}
	if req.CategoryID != nil {
		product.CategoryID = req.CategoryID
	}
	if req.IsActive != nil {
		product.IsActive = *req.IsActive
	}

	if err := s.repo.Update(ctx, product); err != nil {
		return nil, err
	}

	return product, nil
}

func (s *productService) DeleteProduct(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}
