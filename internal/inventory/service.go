package inventory

import (
	"context"

	"github.com/google/uuid"

	"github.com/mschl/ecommerce-api/internal/notifications"
	"github.com/mschl/ecommerce-api/internal/shared"
)

// InventoryService defines the interface for inventory business logic
type InventoryService interface {
	GetInventory(ctx context.Context, productID uuid.UUID) (*Inventory, error)
	UpdateInventory(ctx context.Context, productID uuid.UUID, quantity int) (*Inventory, error)
	IncrementInventory(ctx context.Context, productID uuid.UUID, incrementBy int, userID uuid.UUID) (*Inventory, error)
	CheckAvailability(ctx context.Context, productID uuid.UUID, quantity int) (bool, error)
	CreateInventory(ctx context.Context, productID uuid.UUID, quantity int) (*Inventory, error)
}

// inventoryService implements InventoryService
type inventoryService struct {
	repo                InventoryRepository
	notificationService notifications.NotificationService
}

// NewInventoryService creates a new InventoryService
func NewInventoryService(repo InventoryRepository, notificationService notifications.NotificationService) InventoryService {
	return &inventoryService{
		repo:                repo,
		notificationService: notificationService,
	}
}

func (s *inventoryService) GetInventory(ctx context.Context, productID uuid.UUID) (*Inventory, error) {
	return s.repo.FindByProductID(ctx, productID)
}

func (s *inventoryService) UpdateInventory(ctx context.Context, productID uuid.UUID, quantity int) (*Inventory, error) {
	// Validate quantity is non-negative
	if quantity < 0 {
		return nil, shared.ErrNegativeQuantity
	}

	inventory, err := s.repo.FindByProductID(ctx, productID)
	if err != nil {
		return nil, err
	}

	// Set quantity to new value
	inventory.Quantity = quantity

	if err := s.repo.Update(ctx, inventory); err != nil {
		return nil, err
	}

	return inventory, nil
}

func (s *inventoryService) IncrementInventory(ctx context.Context, productID uuid.UUID, incrementBy int, userID uuid.UUID) (*Inventory, error) {
	inventory, err := s.repo.FindByProductID(ctx, productID)
	if err != nil {
		return nil, err
	}

	// Increment quantity
	newQuantity := inventory.Quantity + incrementBy
	if newQuantity < 0 {
		return nil, shared.ErrNegativeQuantity
	}

	inventory.Quantity = newQuantity

	if err := s.repo.Update(ctx, inventory); err != nil {
		return nil, err
	}

	// Create notification for inventory update
	title := "Inventory Updated"
	message := "Inventory updated successfully."
	if _, err := s.notificationService.CreateNotification(ctx, userID, title, message); err != nil {
		// Log error but don't fail inventory update
	}

	return inventory, nil
}

func (s *inventoryService) CheckAvailability(ctx context.Context, productID uuid.UUID, quantity int) (bool, error) {
	inventory, err := s.repo.FindByProductID(ctx, productID)
	if err != nil {
		if err == shared.ErrNotFound {
			return false, nil
		}
		return false, err
	}
	return inventory.Quantity >= quantity, nil
}

func (s *inventoryService) CreateInventory(ctx context.Context, productID uuid.UUID, quantity int) (*Inventory, error) {
	if quantity < 0 {
		return nil, shared.ErrNegativeQuantity
	}

	inventory := &Inventory{
		ID:        uuid.New(),
		ProductID: productID,
		Quantity:  quantity,
	}

	if err := s.repo.Create(ctx, inventory); err != nil {
		return nil, err
	}

	return inventory, nil
}
