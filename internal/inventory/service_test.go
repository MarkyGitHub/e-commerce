package inventory

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/mschl/ecommerce-api/internal/notifications"
	"github.com/mschl/ecommerce-api/internal/shared"
)

// MockInventoryRepository is a mock implementation of InventoryRepository
type MockInventoryRepository struct {
	inventory map[uuid.UUID]*Inventory
}

func NewMockInventoryRepository() *MockInventoryRepository {
	return &MockInventoryRepository{inventory: make(map[uuid.UUID]*Inventory)}
}

func (m *MockInventoryRepository) Create(ctx context.Context, inv *Inventory) error {
	if inv.ID == uuid.Nil {
		inv.ID = uuid.New()
	}
	m.inventory[inv.ProductID] = inv
	return nil
}

func (m *MockInventoryRepository) FindByProductID(ctx context.Context, productID uuid.UUID) (*Inventory, error) {
	if inv, exists := m.inventory[productID]; exists {
		return inv, nil
	}
	return nil, shared.ErrNotFound
}

func (m *MockInventoryRepository) Update(ctx context.Context, inv *Inventory) error {
	if _, exists := m.inventory[inv.ProductID]; !exists {
		return shared.ErrNotFound
	}
	m.inventory[inv.ProductID] = inv
	return nil
}

func (m *MockInventoryRepository) DecrementQuantity(ctx context.Context, productID uuid.UUID, amount int) error {
	inv, exists := m.inventory[productID]
	if !exists {
		return shared.ErrNotFound
	}
	if inv.Quantity < amount {
		return shared.ErrInsufficientInventory
	}
	inv.Quantity -= amount
	return nil
}

func (m *MockInventoryRepository) DecrementQuantityTx(ctx context.Context, tx *gorm.DB, productID uuid.UUID, amount int) error {
	return m.DecrementQuantity(ctx, productID, amount)
}

// MockNotificationService is a mock implementation of NotificationService
type MockNotificationService struct {
	notifications []*notifications.Notification
}

func NewMockNotificationService() *MockNotificationService {
	return &MockNotificationService{
		notifications: make([]*notifications.Notification, 0),
	}
}

func (m *MockNotificationService) CreateNotification(ctx context.Context, userID uuid.UUID, title, message string) (*notifications.Notification, error) {
	notif := &notifications.Notification{
		ID:      uuid.New(),
		UserID:  userID,
		Title:   title,
		Message: message,
		IsRead:  false,
	}
	m.notifications = append(m.notifications, notif)
	return notif, nil
}

func (m *MockNotificationService) CreateNotificationTx(ctx context.Context, tx *gorm.DB, userID uuid.UUID, title, message string) (*notifications.Notification, error) {
	return m.CreateNotification(ctx, userID, title, message)
}

func (m *MockNotificationService) GetNotifications(ctx context.Context, userID uuid.UUID) ([]notifications.Notification, error) {
	result := make([]notifications.Notification, 0)
	for _, n := range m.notifications {
		if n.UserID == userID {
			result = append(result, *n)
		}
	}
	return result, nil
}

func (m *MockNotificationService) MarkAsRead(ctx context.Context, userID uuid.UUID, notificationID uuid.UUID) error {
	for _, n := range m.notifications {
		if n.ID == notificationID && n.UserID == userID {
			n.IsRead = true
			return nil
		}
	}
	return shared.ErrNotFound
}

// Unit Tests

func TestInventoryService_CreateInventory(t *testing.T) {
	tests := []struct {
		name      string
		productID uuid.UUID
		quantity  int
		wantErr   error
	}{
		{
			name:      "valid inventory",
			productID: uuid.New(),
			quantity:  100,
			wantErr:   nil,
		},
		{
			name:      "zero quantity",
			productID: uuid.New(),
			quantity:  0,
			wantErr:   nil,
		},
		{
			name:      "negative quantity",
			productID: uuid.New(),
			quantity:  -10,
			wantErr:   shared.ErrNegativeQuantity,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewMockInventoryRepository()
			notificationService := NewMockNotificationService()
			service := NewInventoryService(repo, notificationService)

			inv, err := service.CreateInventory(context.Background(), tt.productID, tt.quantity)
			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Errorf("CreateInventory() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Errorf("CreateInventory() unexpected error = %v", err)
				return
			}
			if inv.Quantity != tt.quantity {
				t.Errorf("CreateInventory() quantity = %v, want %v", inv.Quantity, tt.quantity)
			}
		})
	}
}

func TestInventoryService_UpdateInventory(t *testing.T) {
	repo := NewMockInventoryRepository()
	notificationService := NewMockNotificationService()
	service := NewInventoryService(repo, notificationService)

	// Create inventory first
	productID := uuid.New()
	_, err := service.CreateInventory(context.Background(), productID, 50)
	if err != nil {
		t.Fatalf("Failed to create inventory: %v", err)
	}

	tests := []struct {
		name     string
		quantity int
		wantErr  error
	}{
		{
			name:     "update to higher quantity",
			quantity: 100,
			wantErr:  nil,
		},
		{
			name:     "update to zero",
			quantity: 0,
			wantErr:  nil,
		},
		{
			name:     "update to negative",
			quantity: -10,
			wantErr:  shared.ErrNegativeQuantity,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inv, err := service.UpdateInventory(context.Background(), productID, tt.quantity)
			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Errorf("UpdateInventory() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Errorf("UpdateInventory() unexpected error = %v", err)
				return
			}
			if inv.Quantity != tt.quantity {
				t.Errorf("UpdateInventory() quantity = %v, want %v", inv.Quantity, tt.quantity)
			}
		})
	}
}

func TestInventoryService_UpdateInventory_NotFound(t *testing.T) {
	repo := NewMockInventoryRepository()
	notificationService := NewMockNotificationService()
	service := NewInventoryService(repo, notificationService)

	_, err := service.UpdateInventory(context.Background(), uuid.New(), 100)
	if err != shared.ErrNotFound {
		t.Errorf("UpdateInventory() error = %v, want %v", err, shared.ErrNotFound)
	}
}

func TestInventoryService_CheckAvailability(t *testing.T) {
	repo := NewMockInventoryRepository()
	notificationService := NewMockNotificationService()
	service := NewInventoryService(repo, notificationService)

	// Create inventory
	productID := uuid.New()
	_, err := service.CreateInventory(context.Background(), productID, 50)
	if err != nil {
		t.Fatalf("Failed to create inventory: %v", err)
	}

	tests := []struct {
		name     string
		quantity int
		want     bool
	}{
		{
			name:     "available - less than stock",
			quantity: 30,
			want:     true,
		},
		{
			name:     "available - exact stock",
			quantity: 50,
			want:     true,
		},
		{
			name:     "not available - more than stock",
			quantity: 51,
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			available, err := service.CheckAvailability(context.Background(), productID, tt.quantity)
			if err != nil {
				t.Errorf("CheckAvailability() error = %v", err)
				return
			}
			if available != tt.want {
				t.Errorf("CheckAvailability() = %v, want %v", available, tt.want)
			}
		})
	}
}

func TestInventoryService_CheckAvailability_NotFound(t *testing.T) {
	repo := NewMockInventoryRepository()
	notificationService := NewMockNotificationService()
	service := NewInventoryService(repo, notificationService)

	available, err := service.CheckAvailability(context.Background(), uuid.New(), 10)
	if err != nil {
		t.Errorf("CheckAvailability() unexpected error = %v", err)
	}
	if available {
		t.Error("CheckAvailability() should return false for non-existing product")
	}
}

// TestInventoryService_CheckAvailability_ZeroQuantity tests CheckAvailability with zero quantity request
func TestInventoryService_CheckAvailability_ZeroQuantity(t *testing.T) {
	repo := NewMockInventoryRepository()
	notificationService := NewMockNotificationService()
	service := NewInventoryService(repo, notificationService)

	// Create inventory
	productID := uuid.New()
	_, err := service.CreateInventory(context.Background(), productID, 50)
	if err != nil {
		t.Fatalf("Failed to create inventory: %v", err)
	}

	// Check availability for zero quantity (should always be true if inventory exists)
	available, err := service.CheckAvailability(context.Background(), productID, 0)
	if err != nil {
		t.Errorf("CheckAvailability() error = %v, want nil", err)
	}
	if !available {
		t.Error("CheckAvailability() should return true for zero quantity when inventory exists")
	}
}

// Property 7: Inventory Update Consistency
func TestProperty_InventoryUpdateConsistency(t *testing.T) {
	repo := NewMockInventoryRepository()
	notificationService := NewMockNotificationService()
	service := NewInventoryService(repo, notificationService)

	productID := uuid.New()

	// Create with initial quantity
	initial := 100
	_, err := service.CreateInventory(context.Background(), productID, initial)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Test various updates
	quantities := []int{50, 200, 0, 150}
	for _, q := range quantities {
		inv, err := service.UpdateInventory(context.Background(), productID, q)
		if err != nil {
			t.Errorf("Update to %d failed: %v", q, err)
			continue
		}
		if inv.Quantity != q {
			t.Errorf("After update quantity = %d, want %d", inv.Quantity, q)
		}

		// Verify get returns same value
		got, err := service.GetInventory(context.Background(), productID)
		if err != nil {
			t.Errorf("Get failed: %v", err)
			continue
		}
		if got.Quantity != q {
			t.Errorf("Get after update quantity = %d, want %d", got.Quantity, q)
		}
	}

	// Negative quantities should be rejected
	_, err = service.UpdateInventory(context.Background(), productID, -10)
	if err != shared.ErrNegativeQuantity {
		t.Errorf("Negative update error = %v, want %v", err, shared.ErrNegativeQuantity)
	}
}

// TestInventoryService_GetInventory tests GetInventory
func TestInventoryService_GetInventory(t *testing.T) {
	repo := NewMockInventoryRepository()
	notificationService := NewMockNotificationService()
	service := NewInventoryService(repo, notificationService)

	productID := uuid.New()
	_, err := service.CreateInventory(context.Background(), productID, 100)
	if err != nil {
		t.Fatalf("Failed to create inventory: %v", err)
	}

	inv, err := service.GetInventory(context.Background(), productID)
	if err != nil {
		t.Errorf("GetInventory() error = %v", err)
	}
	if inv.Quantity != 100 {
		t.Errorf("GetInventory() quantity = %v, want %v", inv.Quantity, 100)
	}
}

// TestInventoryService_GetInventory_NotFound tests GetInventory when inventory doesn't exist
func TestInventoryService_GetInventory_NotFound(t *testing.T) {
	repo := NewMockInventoryRepository()
	notificationService := NewMockNotificationService()
	service := NewInventoryService(repo, notificationService)

	_, err := service.GetInventory(context.Background(), uuid.New())
	if err != shared.ErrNotFound {
		t.Errorf("GetInventory() error = %v, want %v", err, shared.ErrNotFound)
	}
}

// TestInventoryService_IncrementInventory tests IncrementInventory
func TestInventoryService_IncrementInventory(t *testing.T) {
	repo := NewMockInventoryRepository()
	notificationService := NewMockNotificationService()
	service := NewInventoryService(repo, notificationService)

	productID := uuid.New()
	userID := uuid.New()
	_, err := service.CreateInventory(context.Background(), productID, 50)
	if err != nil {
		t.Fatalf("Failed to create inventory: %v", err)
	}

	tests := []struct {
		name        string
		incrementBy int
		wantQuantity int
		wantErr     error
	}{
		{
			name:        "positive increment",
			incrementBy: 30,
			wantQuantity: 80,
			wantErr:     nil,
		},
		{
			name:        "zero increment",
			incrementBy: 0,
			wantQuantity: 50,
			wantErr:     nil,
		},
		{
			name:        "negative increment that results in negative",
			incrementBy: -100,
			wantQuantity: 0,
			wantErr:     shared.ErrNegativeQuantity,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset inventory for each test
			_, err := service.UpdateInventory(context.Background(), productID, 50)
			if err != nil {
				t.Fatalf("Failed to reset inventory: %v", err)
			}

			inv, err := service.IncrementInventory(context.Background(), productID, tt.incrementBy, userID)
			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Errorf("IncrementInventory() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Errorf("IncrementInventory() unexpected error = %v", err)
				return
			}
			if inv.Quantity != tt.wantQuantity {
				t.Errorf("IncrementInventory() quantity = %v, want %v", inv.Quantity, tt.wantQuantity)
			}
		})
	}
}

// TestInventoryService_IncrementInventory_NotFound tests IncrementInventory when inventory doesn't exist
func TestInventoryService_IncrementInventory_NotFound(t *testing.T) {
	repo := NewMockInventoryRepository()
	notificationService := NewMockNotificationService()
	service := NewInventoryService(repo, notificationService)

	_, err := service.IncrementInventory(context.Background(), uuid.New(), 10, uuid.New())
	if err != shared.ErrNotFound {
		t.Errorf("IncrementInventory() error = %v, want %v", err, shared.ErrNotFound)
	}
}

// ErrorInventoryRepository is a mock that returns errors
type ErrorInventoryRepository struct {
	createError          error
	findByProductIDError error
	updateError          error
	decrementError       error
	findByProductIDFunc  func(ctx context.Context, productID uuid.UUID) (*Inventory, error)
}

func (m *ErrorInventoryRepository) Create(ctx context.Context, inv *Inventory) error {
	if m.createError != nil {
		return m.createError
	}
	return nil
}

func (m *ErrorInventoryRepository) FindByProductID(ctx context.Context, productID uuid.UUID) (*Inventory, error) {
	if m.findByProductIDFunc != nil {
		return m.findByProductIDFunc(ctx, productID)
	}
	if m.findByProductIDError != nil {
		return nil, m.findByProductIDError
	}
	return nil, shared.ErrNotFound
}

func (m *ErrorInventoryRepository) Update(ctx context.Context, inv *Inventory) error {
	if m.updateError != nil {
		return m.updateError
	}
	return nil
}

func (m *ErrorInventoryRepository) DecrementQuantity(ctx context.Context, productID uuid.UUID, amount int) error {
	if m.decrementError != nil {
		return m.decrementError
	}
	return nil
}

func (m *ErrorInventoryRepository) DecrementQuantityTx(ctx context.Context, tx *gorm.DB, productID uuid.UUID, amount int) error {
	return m.DecrementQuantity(ctx, productID, amount)
}

// TestInventoryService_CheckAvailability_Error tests CheckAvailability with non-ErrNotFound error
func TestInventoryService_CheckAvailability_Error(t *testing.T) {
	repo := &ErrorInventoryRepository{findByProductIDError: shared.ErrBadRequest}
	notificationService := NewMockNotificationService()
	service := NewInventoryService(repo, notificationService)

	available, err := service.CheckAvailability(context.Background(), uuid.New(), 10)
	if err != shared.ErrBadRequest {
		t.Errorf("CheckAvailability() error = %v, wantErr %v", err, shared.ErrBadRequest)
	}
	if available {
		t.Error("CheckAvailability() should return false on error")
	}
}

// TestInventoryService_UpdateInventory_UpdateError tests UpdateInventory when Update fails
func TestInventoryService_UpdateInventory_UpdateError(t *testing.T) {
	normalRepo := NewMockInventoryRepository()
	notificationService := NewMockNotificationService()
	
	// Create inventory first
	productID := uuid.New()
	inv := &Inventory{
		ID:        uuid.New(),
		ProductID: productID,
		Quantity:  50,
	}
	normalRepo.inventory[productID] = inv

	// Create error repository that fails on Update
	errorRepo := &ErrorInventoryRepository{
		updateError: shared.ErrBadRequest,
	}
	errorRepo.findByProductIDFunc = func(ctx context.Context, pid uuid.UUID) (*Inventory, error) {
		return normalRepo.inventory[pid], nil
	}

	errorService := NewInventoryService(errorRepo, notificationService)
	_, err := errorService.UpdateInventory(context.Background(), productID, 100)
	if err != shared.ErrBadRequest {
		t.Errorf("UpdateInventory() error = %v, wantErr %v", err, shared.ErrBadRequest)
	}
}

// TestInventoryService_CreateInventory_CreateError tests CreateInventory when Create fails
func TestInventoryService_CreateInventory_CreateError(t *testing.T) {
	repo := &ErrorInventoryRepository{createError: shared.ErrBadRequest}
	notificationService := NewMockNotificationService()
	service := NewInventoryService(repo, notificationService)

	_, err := service.CreateInventory(context.Background(), uuid.New(), 50)
	if err != shared.ErrBadRequest {
		t.Errorf("CreateInventory() error = %v, wantErr %v", err, shared.ErrBadRequest)
	}
}

// TestInventoryService_IncrementInventory_UpdateError tests IncrementInventory when Update fails
func TestInventoryService_IncrementInventory_UpdateError(t *testing.T) {
	normalRepo := NewMockInventoryRepository()
	notificationService := NewMockNotificationService()
	
	// Create inventory first
	productID := uuid.New()
	inv := &Inventory{
		ID:        uuid.New(),
		ProductID: productID,
		Quantity:  50,
	}
	normalRepo.inventory[productID] = inv

	// Create error repository that fails on Update
	errorRepo := &ErrorInventoryRepository{
		updateError: shared.ErrBadRequest,
	}
	errorRepo.findByProductIDFunc = func(ctx context.Context, pid uuid.UUID) (*Inventory, error) {
		return normalRepo.inventory[pid], nil
	}

	errorService := NewInventoryService(errorRepo, notificationService)
	_, err := errorService.IncrementInventory(context.Background(), productID, 10, uuid.New())
	if err != shared.ErrBadRequest {
		t.Errorf("IncrementInventory() error = %v, wantErr %v", err, shared.ErrBadRequest)
	}
}

// TestInventoryService_IncrementInventory_NegativeResult tests IncrementInventory when result would be negative
func TestInventoryService_IncrementInventory_NegativeResult(t *testing.T) {
	repo := NewMockInventoryRepository()
	notificationService := NewMockNotificationService()
	service := NewInventoryService(repo, notificationService)

	// Create inventory with low quantity
	productID := uuid.New()
	_, err := service.CreateInventory(context.Background(), productID, 5)
	if err != nil {
		t.Fatalf("Failed to create inventory: %v", err)
	}

	// Try to increment by negative amount that would make total negative
	_, err = service.IncrementInventory(context.Background(), productID, -10, uuid.New())
	if err != shared.ErrNegativeQuantity {
		t.Errorf("IncrementInventory() error = %v, wantErr %v", err, shared.ErrNegativeQuantity)
	}
}

// TestInventoryService_IncrementInventory_NegativeIncrementValidResult tests IncrementInventory with negative increment that results in valid quantity
func TestInventoryService_IncrementInventory_NegativeIncrementValidResult(t *testing.T) {
	repo := NewMockInventoryRepository()
	notificationService := NewMockNotificationService()
	service := NewInventoryService(repo, notificationService)

	// Create inventory
	productID := uuid.New()
	_, err := service.CreateInventory(context.Background(), productID, 50)
	if err != nil {
		t.Fatalf("Failed to create inventory: %v", err)
	}

	// Increment by negative amount that still results in positive quantity
	inv, err := service.IncrementInventory(context.Background(), productID, -20, uuid.New())
	if err != nil {
		t.Errorf("IncrementInventory() error = %v, want nil", err)
	}
	if inv.Quantity != 30 {
		t.Errorf("IncrementInventory() quantity = %d, want 30", inv.Quantity)
	}
}

// TestInventory_ToResponse tests the ToResponse method
func TestInventory_ToResponse(t *testing.T) {
	inv := &Inventory{
		ID:        uuid.New(),
		ProductID: uuid.New(),
		Quantity:  100,
		Reserved:  5,
		Location:  "A-01",
	}

	response := inv.ToResponse()
	if response.ProductID != inv.ProductID {
		t.Errorf("ToResponse() ProductID = %v, want %v", response.ProductID, inv.ProductID)
	}
	if response.Quantity != inv.Quantity {
		t.Errorf("ToResponse() Quantity = %d, want %d", response.Quantity, inv.Quantity)
	}
	if response.Reserved != inv.Reserved {
		t.Errorf("ToResponse() Reserved = %d, want %d", response.Reserved, inv.Reserved)
	}
	if response.Location != inv.Location {
		t.Errorf("ToResponse() Location = %s, want %s", response.Location, inv.Location)
	}
}

// FailingNotificationService is a mock that always fails
type FailingNotificationService struct{}

func (m *FailingNotificationService) CreateNotification(ctx context.Context, userID uuid.UUID, title, message string) (*notifications.Notification, error) {
	return nil, shared.ErrBadRequest
}

func (m *FailingNotificationService) CreateNotificationTx(ctx context.Context, tx *gorm.DB, userID uuid.UUID, title, message string) (*notifications.Notification, error) {
	return nil, shared.ErrBadRequest
}

func (m *FailingNotificationService) GetNotifications(ctx context.Context, userID uuid.UUID) ([]notifications.Notification, error) {
	return nil, shared.ErrBadRequest
}

func (m *FailingNotificationService) MarkAsRead(ctx context.Context, userID uuid.UUID, notificationID uuid.UUID) error {
	return shared.ErrBadRequest
}

// TestInventoryService_NotificationErrors tests that notification errors don't fail operations
func TestInventoryService_NotificationErrors(t *testing.T) {
	repo := NewMockInventoryRepository()
	failingNotificationService := &FailingNotificationService{}
	service := NewInventoryService(repo, failingNotificationService)

	productID := uuid.New()
	userID := uuid.New()
	_, err := service.CreateInventory(context.Background(), productID, 50)
	if err != nil {
		t.Fatalf("Failed to create inventory: %v", err)
	}

	// IncrementInventory should succeed even if notification fails
	inv, err := service.IncrementInventory(context.Background(), productID, 10, userID)
	if err != nil {
		t.Errorf("IncrementInventory() should succeed even if notification fails, got error: %v", err)
	}
	if inv == nil {
		t.Error("IncrementInventory() should return inventory even if notification fails")
	}
	if inv.Quantity != 60 {
		t.Errorf("IncrementInventory() quantity = %v, want %v", inv.Quantity, 60)
	}
}