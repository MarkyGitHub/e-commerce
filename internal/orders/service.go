package orders

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/mschl/ecommerce-api/internal/cart"
	"github.com/mschl/ecommerce-api/internal/inventory"
	"github.com/mschl/ecommerce-api/internal/notifications"
	"github.com/mschl/ecommerce-api/internal/products"
	"github.com/mschl/ecommerce-api/internal/shared"
)

// OrderService definiert Interface für Bestell-Geschäftslogik
type OrderService interface {
	CreateOrder(ctx context.Context, userID uuid.UUID, shippingAddress CreateOrderRequest) (*Order, []OrderItem, error)
	GetOrder(ctx context.Context, userID uuid.UUID, orderID uuid.UUID, isAdmin bool) (*Order, []OrderItem, error)
	ListOrders(ctx context.Context, userID uuid.UUID, isAdmin bool, page, limit int) ([]Order, int64, error)
}

// orderService implementiert OrderService
type orderService struct {
	db                  *gorm.DB
	transaction         func(func(*gorm.DB) error) error
	orderRepo           OrderRepository
	orderItemRepo       OrderItemRepository
	cartRepo            cart.CartRepository
	cartItemRepo        cart.CartItemRepository
	productRepo         products.ProductRepository
	inventoryRepo       inventory.InventoryRepository
	notificationService notifications.NotificationService
}

// NewOrderService erstellt einen neuen OrderService
func NewOrderService(
	db *gorm.DB,
	orderRepo OrderRepository,
	orderItemRepo OrderItemRepository,
	cartRepo cart.CartRepository,
	cartItemRepo cart.CartItemRepository,
	productRepo products.ProductRepository,
	inventoryRepo inventory.InventoryRepository,
	notificationService notifications.NotificationService,
) OrderService {
	return &orderService{
		db: db,
		transaction: func(fn func(*gorm.DB) error) error {
			return db.Transaction(fn)
		},
		orderRepo:           orderRepo,
		orderItemRepo:       orderItemRepo,
		cartRepo:            cartRepo,
		cartItemRepo:        cartItemRepo,
		productRepo:         productRepo,
		inventoryRepo:       inventoryRepo,
		notificationService: notificationService,
	}
}

func (s *orderService) CreateOrder(ctx context.Context, userID uuid.UUID, shippingAddress CreateOrderRequest) (*Order, []OrderItem, error) {
	// Lieferadresse validieren
	if shippingAddress.ShippingStreet == "" {
		return nil, nil, shared.ErrBadRequest
	}
	if shippingAddress.ShippingCity == "" {
		return nil, nil, shared.ErrBadRequest
	}
	if shippingAddress.ShippingZip == "" {
		return nil, nil, shared.ErrBadRequest
	}
	if shippingAddress.ShippingCountry == "" {
		return nil, nil, shared.ErrBadRequest
	}

	// Warenkorb holen
	userCart, err := s.cartRepo.FindByUserID(ctx, userID)
	if err != nil {
		if err == shared.ErrNotFound {
			return nil, nil, shared.ErrEmptyCart
		}
		return nil, nil, err
	}

	// Warenkorb-Artikel holen
	cartItems, err := s.cartItemRepo.FindByCartID(ctx, userCart.ID)
	if err != nil {
		return nil, nil, err
	}

	if len(cartItems) == 0 {
		return nil, nil, shared.ErrEmptyCart
	}

	if s.transaction == nil {
		return nil, nil, fmt.Errorf("transaction handler is not configured")
	}

	// Transaktion starten
	var order *Order
	var orderItems []OrderItem

	err = s.transaction(func(tx *gorm.DB) error {
		var totalPrice float64

		// Bestellartikel erstellen und Summe berechnen
		orderItems = make([]OrderItem, len(cartItems))
		for i, cartItem := range cartItems {
			// Produktpreis holen
			product, err := s.productRepo.FindByID(ctx, cartItem.ProductID)
			if err != nil {
				return err
			}

			// Lagerbestand prüfen und reduzieren
			if tx != nil {
				if err := s.inventoryRepo.DecrementQuantityTx(ctx, tx, cartItem.ProductID, cartItem.Quantity); err != nil {
					return err
				}
			} else {
				if err := s.inventoryRepo.DecrementQuantity(ctx, cartItem.ProductID, cartItem.Quantity); err != nil {
					return err
				}
			}

			itemPrice := product.Price * float64(cartItem.Quantity)
			totalPrice += itemPrice

			orderItems[i] = OrderItem{
				ID:        uuid.New(),
				ProductID: cartItem.ProductID,
				Quantity:  cartItem.Quantity,
				Price:     product.Price,
			}
		}

		// Bestellung mit Lieferadresse erstellen
		order = &Order{
			ID:              uuid.New(),
			UserID:          userID,
			TotalPrice:      totalPrice,
			Status:          StatusPending,
			ShippingStreet:  shippingAddress.ShippingStreet,
			ShippingCity:    shippingAddress.ShippingCity,
			ShippingZip:     shippingAddress.ShippingZip,
			ShippingCountry: shippingAddress.ShippingCountry,
		}

		if tx != nil {
			if err := s.orderRepo.CreateTx(ctx, tx, order); err != nil {
				return err
			}

			// Bestell-ID auf Artikel setzen und erstellen
			for i := range orderItems {
				orderItems[i].OrderID = order.ID
			}

			if err := s.orderItemRepo.CreateBatchTx(ctx, tx, orderItems); err != nil {
				return err
			}

			// Warenkorb-Artikel löschen
			if err := s.cartRepo.ClearTx(ctx, tx, userCart.ID); err != nil {
				return err
			}

			// Benachrichtigung in Transaktion erstellen
			title := "Order Placed"
			message := fmt.Sprintf("Your order #%s has been placed successfully. Total: $%.2f", order.ID.String()[:8], order.TotalPrice)
			if _, notifErr := s.notificationService.CreateNotificationTx(ctx, tx, userID, title, message); notifErr != nil {
				return notifErr
			}
		} else {
			if err := s.orderRepo.Create(ctx, order); err != nil {
				return err
			}
			for i := range orderItems {
				orderItems[i].OrderID = order.ID
			}
			if err := s.orderItemRepo.CreateBatch(ctx, orderItems); err != nil {
				return err
			}
			if err := s.cartRepo.Clear(ctx, userCart.ID); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, nil, err
	}

	return order, orderItems, nil
}

func (s *orderService) GetOrder(ctx context.Context, userID uuid.UUID, orderID uuid.UUID, isAdmin bool) (*Order, []OrderItem, error) {
	order, err := s.orderRepo.FindByID(ctx, orderID)
	if err != nil {
		return nil, nil, err
	}

	// Zugriff prüfen: Benutzer sieht nur eigene Bestellungen außer Admin
	if !isAdmin && order.UserID != userID {
		return nil, nil, shared.ErrForbidden
	}

	items, err := s.orderItemRepo.FindByOrderID(ctx, orderID)
	if err != nil {
		return nil, nil, err
	}

	return order, items, nil
}

func (s *orderService) ListOrders(ctx context.Context, userID uuid.UUID, isAdmin bool, page, limit int) ([]Order, int64, error) {
	offset := (page - 1) * limit

	if isAdmin {
		return s.orderRepo.FindAll(ctx, offset, limit)
	}
	return s.orderRepo.FindByUserID(ctx, userID, offset, limit)
}
