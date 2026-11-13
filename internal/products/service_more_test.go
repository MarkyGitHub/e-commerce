package products

import (
	"context"
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/MarkyGitHub/e-commerce/internal/notifications"
	"github.com/MarkyGitHub/e-commerce/internal/shared"
	"gorm.io/gorm"
)

type failingNotificationService struct{}

func (f failingNotificationService) CreateNotification(ctx context.Context, userID uuid.UUID, title, message string) (*notifications.Notification, error) {
	return nil, errors.New("notify")
}
func (f failingNotificationService) CreateNotificationTx(ctx context.Context, tx *gorm.DB, userID uuid.UUID, title, message string) (*notifications.Notification, error) {
	return nil, errors.New("notify")
}
func (f failingNotificationService) GetNotifications(ctx context.Context, userID uuid.UUID) ([]notifications.Notification, error) {
	return nil, nil
}
func (f failingNotificationService) MarkAsRead(ctx context.Context, userID uuid.UUID, notificationID uuid.UUID) error {
	return nil
}

type repoError struct{ err error }

func (r repoError) Create(ctx context.Context, product *Product) error                                   { return r.err }
func (r repoError) FindByID(ctx context.Context, id uuid.UUID) (*Product, error)                        { return nil, r.err }
func (r repoError) FindAll(ctx context.Context, offset, limit int, categoryID *uuid.UUID) ([]Product, int64, error) {
	return nil, 0, r.err
}
func (r repoError) Update(ctx context.Context, product *Product) error { return r.err }
func (r repoError) Delete(ctx context.Context, id uuid.UUID) error     { return r.err }

func TestProductService_CreateProduct_NotificationFailureDoesNotFail(t *testing.T) {
	repo := repoError{err: nil}
	svc := NewProductService(repo, failingNotificationService{})

	req := CreateProductRequest{Name: "N", Price: 1}
	p, err := svc.CreateProduct(context.Background(), req, uuid.New())
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if p == nil {
		t.Fatalf("expected product")
	}
}

func TestProductService_UpdateProduct_EmptyNameValidation(t *testing.T) {
	id := uuid.New()
	empty := ""

	repo := &MockProductRepository{products: map[uuid.UUID]*Product{id: {ID: id, Name: "old", Price: 1, IsActive: true}}}
	svc := NewProductService(repo, NewMockNotificationService())

	_, err := svc.UpdateProduct(context.Background(), id, UpdateProductRequest{Name: &empty})
	if err != ErrInvalidProductName {
		t.Fatalf("expected %v, got %v", ErrInvalidProductName, err)
	}
}

func TestProductRepository_Create_AutoGenerateID(t *testing.T) {
	db := setupProductsTestDB(t)
	repo := NewProductRepository(db)

	p := &Product{ID: uuid.Nil, Name: "P", Price: 1, IsActive: true}
	if err := repo.Create(context.Background(), p); err != nil {
		t.Fatalf("Create error: %v", err)
	}
	if p.ID == uuid.Nil {
		t.Fatalf("expected ID to be generated")
	}
}

func TestProductRepository_FindAll_CategoryFilter(t *testing.T) {
	db := setupProductsTestDB(t)
	repo := NewProductRepository(db)
	ctx := context.Background()

	cat1 := uuid.New()
	cat2 := uuid.New()
	p1 := &Product{ID: uuid.New(), Name: "A", Price: 1, CategoryID: &cat1, IsActive: true}
	p2 := &Product{ID: uuid.New(), Name: "B", Price: 1, CategoryID: &cat2, IsActive: true}
	_ = repo.Create(ctx, p1)
	_ = repo.Create(ctx, p2)

	items, total, err := repo.FindAll(ctx, 0, 10, &cat1)
	if err != nil {
		t.Fatalf("FindAll error: %v", err)
	}
	if total != 1 {
		t.Fatalf("expected total 1, got %d", total)
	}
	if len(items) != 1 || items[0].ID != p1.ID {
		t.Fatalf("expected only cat1 product")
	}
}

func TestProductHandler_List_InvalidCategoryID_Ignored(t *testing.T) {
	mockService := &MockProductService{listFunc: func(ctx context.Context, page, limit int, categoryID *uuid.UUID) ([]Product, int64, error) {
		if categoryID != nil {
			return nil, 0, shared.ErrBadRequest
		}
		return []Product{}, 0, nil
	}}
	h := NewHandler(mockService)

	req := httptest.NewRequest("GET", "/products?category_id=not-a-uuid", nil)
	rr := httptest.NewRecorder()

	h.List(rr, req)
	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}
