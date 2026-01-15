package cart

import (
	"context"

	"github.com/google/uuid"

	"github.com/mschl/ecommerce-api/internal/inventory"
	"github.com/mschl/ecommerce-api/internal/notifications"
	"github.com/mschl/ecommerce-api/internal/products"
	"github.com/mschl/ecommerce-api/internal/shared"
)

// CartService definiert Interface für Warenkorb-Geschäftslogik
type CartService interface {
	GetCart(ctx context.Context, userID uuid.UUID) (*Cart, []CartItem, error)
	AddItem(ctx context.Context, userID uuid.UUID, req AddItemRequest) (*CartItem, error)
	UpdateItem(ctx context.Context, userID uuid.UUID, itemID uuid.UUID, quantity int) (*CartItem, error)
	RemoveItem(ctx context.Context, userID uuid.UUID, itemID uuid.UUID) error
	ClearCart(ctx context.Context, userID uuid.UUID) error
}

// cartService implementiert CartService
type cartService struct {
	cartRepo            CartRepository
	cartItemRepo        CartItemRepository
	inventoryRepo       inventory.InventoryRepository
	productRepo         products.ProductRepository
	notificationService notifications.NotificationService
}

// NewCartService erstellt einen neuen CartService
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
	// Prüfen ob Produkt existiert
	_, err := s.productRepo.FindByID(ctx, req.ProductID)
	if err != nil {
		return nil, err
	}

	// Lagerbestandsverfügbarkeit prüfen
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

	// Warenkorb holen oder erstellen
	cart, err := s.cartRepo.FindOrCreateByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Prüfen ob Produkt bereits im Warenkorb
	existingItem, err := s.cartItemRepo.FindByCartAndProduct(ctx, cart.ID, req.ProductID)
	if err == nil {
		// Bestehenden Artikel aktualisieren
		newQuantity := existingItem.Quantity + req.Quantity
		if inv.Quantity < newQuantity {
			return nil, shared.ErrInsufficientInventory
		}
		existingItem.Quantity = newQuantity
		if err := s.cartItemRepo.Update(ctx, existingItem); err != nil {
			return nil, err
		}
		// Benachrichtigung für Warenkorb-Aktualisierung erstellen
		title := "Cart Updated"
		message := "Cart item quantity updated successfully."
		if _, err := s.notificationService.CreateNotification(ctx, userID, title, message); err != nil {
			// Fehler loggen, aber Warenkorb-Aktualisierung nicht abbrechen
		}
		return existingItem, nil
	}
	if err != shared.ErrNotFound {
		return nil, err
	}

	// Neuen Warenkorb-Artikel erstellen
	item := &CartItem{
		ID:        uuid.New(),
		CartID:    cart.ID,
		ProductID: req.ProductID,
		Quantity:  req.Quantity,
	}
	if err := s.cartItemRepo.Create(ctx, item); err != nil {
		return nil, err
	}

	// Benachrichtigung für Warenkorb-Aktualisierung erstellen
	title := "Cart Updated"
	message := "Item added to cart successfully."
	if _, err := s.notificationService.CreateNotification(ctx, userID, title, message); err != nil {
		// Fehler loggen, aber Hinzufügen nicht abbrechen
	}

	return item, nil
}

func (s *cartService) UpdateItem(ctx context.Context, userID uuid.UUID, itemID uuid.UUID, quantity int) (*CartItem, error) {
	// Warenkorb holen
	cart, err := s.cartRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Warenkorb-Artikel holen
	item, err := s.cartItemRepo.FindByID(ctx, itemID)
	if err != nil {
		return nil, err
	}

	// Prüfen ob Artikel zum Benutzer-Warenkorb gehört
	if item.CartID != cart.ID {
		return nil, shared.ErrForbidden
	}

	// Lagerbestandsverfügbarkeit prüfen
	inv, err := s.inventoryRepo.FindByProductID(ctx, item.ProductID)
	if err != nil {
		return nil, err
	}
	if inv.Quantity < quantity {
		return nil, shared.ErrInsufficientInventory
	}

	// Menge aktualisieren
	item.Quantity = quantity
	if err := s.cartItemRepo.Update(ctx, item); err != nil {
		return nil, err
	}

	// Benachrichtigung für Warenkorb-Aktualisierung erstellen
	title := "Cart Updated"
	message := "Cart item quantity updated successfully."
	if _, err := s.notificationService.CreateNotification(ctx, userID, title, message); err != nil {
		// Fehler loggen, aber Warenkorb-Aktualisierung nicht abbrechen
	}

	return item, nil
}

func (s *cartService) RemoveItem(ctx context.Context, userID uuid.UUID, itemID uuid.UUID) error {
	// Warenkorb holen
	cart, err := s.cartRepo.FindByUserID(ctx, userID)
	if err != nil {
		return err
	}

	// Warenkorb-Artikel holen
	item, err := s.cartItemRepo.FindByID(ctx, itemID)
	if err != nil {
		return err
	}

	// Prüfen ob Artikel zum Benutzer-Warenkorb gehört
	if item.CartID != cart.ID {
		return shared.ErrForbidden
	}

	return s.cartItemRepo.Delete(ctx, itemID)
}

func (s *cartService) ClearCart(ctx context.Context, userID uuid.UUID) error {
	cart, err := s.cartRepo.FindByUserID(ctx, userID)
	if err != nil {
		if err == shared.ErrNotFound {
			return nil // Nichts zu löschen
		}
		return err
	}

	return s.cartRepo.Clear(ctx, cart.ID)
}
