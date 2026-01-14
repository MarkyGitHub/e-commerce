package cart

import (
	"context"

	"github.com/google/uuid"

	"github.com/mschl/ecommerce-api/internal/inventory"
	"github.com/mschl/ecommerce-api/internal/notifications"
	"github.com/mschl/ecommerce-api/internal/products"
	"github.com/mschl/ecommerce-api/internal/shared"
)

// CartService defines the interface for cart business logic
type CartService interface {
	GetCart(ctx context.Context, userID uuid.UUID) (*Cart, []CartItem, error)
	AddItem(ctx context.Context, userID uuid.UUID, req AddItemRequest) (*CartItem, error)
	UpdateItem(ctx context.Context, userID uuid.UUID, itemID uuid.UUID, quantity int) (*CartItem, error)
	RemoveItem(ctx context.Context, userID uuid.UUID, itemID uuid.UUID) error
	ClearCart(ctx context.Context, userID uuid.UUID) error
}

// cartService implements CartService
type cartService struct {
	cartRepo            CartRepository
	cartItemRepo        CartItemRepository
	inventoryRepo       inventory.InventoryRepository
	productRepo         products.ProductRepository
	notificationService notifications.NotificationService
}

// NewCartService creates a new CartService
func NewCartService(
	cartRepo CartRepository,
	cartItemRepo CartItemRepository,
	inventoryRepo inventory.InventoryRepository,
	productRepo products.ProductRepository,
	notificationService notifications.NotificationService,
) CartService {
	return &cartService{
		cartRepo:            cartRepo,
		cartItemRepo:        cartItemRepo,
		inventoryRepo:       inventoryRepo,
		productRepo:         productRepo,
		notificationService: notificationService,
	}
}

func (s *cartService) GetCart(ctx context.Context, userID uuid.UUID) (*Cart, []CartItem, error) {
	cart, err := s.cartRepo.FindOrCreateByUserID(ctx, userID)
	if err != nil {
		return nil, nil, err
	}

	items, err := s.cartItemRepo.FindByCartID(ctx, cart.ID)
	if err != nil {
		return nil, nil, err
	}

	return cart, items, nil
}

func (s *cartService) AddItem(ctx context.Context, userID uuid.UUID, req AddItemRequest) (*CartItem, error) {
	// Validate product exists
	_, err := s.productRepo.FindByID(ctx, req.ProductID)
	if err != nil {
		return nil, err
	}

	// Check inventory availability
	inv, err := s.inventoryRepo.FindByProductID(ctx, req.ProductID)
	if err != nil {
		if err == shared.ErrNotFound {
			return nil, shared.ErrInsufficientInventory
		}
		return nil, err
	}
	if inv.Quantity < req.Quantity {
		return nil, shared.ErrInsufficientInventory
	}

	// Get or create cart
	cart, err := s.cartRepo.FindOrCreateByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Check if product already in cart
	existingItem, err := s.cartItemRepo.FindByCartAndProduct(ctx, cart.ID, req.ProductID)
	if err == nil {
		// Update existing item
		newQuantity := existingItem.Quantity + req.Quantity
		if inv.Quantity < newQuantity {
			return nil, shared.ErrInsufficientInventory
		}
		existingItem.Quantity = newQuantity
		if err := s.cartItemRepo.Update(ctx, existingItem); err != nil {
			return nil, err
		}
		// Create notification for cart update
		title := "Cart Updated"
		message := "Cart item quantity updated successfully."
		if _, err := s.notificationService.CreateNotification(ctx, userID, title, message); err != nil {
			// Log error but don't fail cart update
		}
		return existingItem, nil
	}
	if err != shared.ErrNotFound {
		return nil, err
	}

	// Create new cart item
	item := &CartItem{
		ID:        uuid.New(),
		CartID:    cart.ID,
		ProductID: req.ProductID,
		Quantity:  req.Quantity,
	}
	if err := s.cartItemRepo.Create(ctx, item); err != nil {
		return nil, err
	}

	// Create notification for cart update
	title := "Cart Updated"
	message := "Item added to cart successfully."
	if _, err := s.notificationService.CreateNotification(ctx, userID, title, message); err != nil {
		// Log error but don't fail cart addition
	}

	return item, nil
}

func (s *cartService) UpdateItem(ctx context.Context, userID uuid.UUID, itemID uuid.UUID, quantity int) (*CartItem, error) {
	// Get cart
	cart, err := s.cartRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Get cart item
	item, err := s.cartItemRepo.FindByID(ctx, itemID)
	if err != nil {
		return nil, err
	}

	// Verify item belongs to user's cart
	if item.CartID != cart.ID {
		return nil, shared.ErrForbidden
	}

	// Check inventory availability
	inv, err := s.inventoryRepo.FindByProductID(ctx, item.ProductID)
	if err != nil {
		return nil, err
	}
	if inv.Quantity < quantity {
		return nil, shared.ErrInsufficientInventory
	}

	// Update quantity
	item.Quantity = quantity
	if err := s.cartItemRepo.Update(ctx, item); err != nil {
		return nil, err
	}

	// Create notification for cart update
	title := "Cart Updated"
	message := "Cart item quantity updated successfully."
	if _, err := s.notificationService.CreateNotification(ctx, userID, title, message); err != nil {
		// Log error but don't fail cart update
	}

	return item, nil
}

func (s *cartService) RemoveItem(ctx context.Context, userID uuid.UUID, itemID uuid.UUID) error {
	// Get cart
	cart, err := s.cartRepo.FindByUserID(ctx, userID)
	if err != nil {
		return err
	}

	// Get cart item
	item, err := s.cartItemRepo.FindByID(ctx, itemID)
	if err != nil {
		return err
	}

	// Verify item belongs to user's cart
	if item.CartID != cart.ID {
		return shared.ErrForbidden
	}

	return s.cartItemRepo.Delete(ctx, itemID)
}

func (s *cartService) ClearCart(ctx context.Context, userID uuid.UUID) error {
	cart, err := s.cartRepo.FindByUserID(ctx, userID)
	if err != nil {
		if err == shared.ErrNotFound {
			return nil // Nothing to clear
		}
		return err
	}

	return s.cartRepo.Clear(ctx, cart.ID)
}
