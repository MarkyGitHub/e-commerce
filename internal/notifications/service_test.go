package notifications

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/MarkyGitHub/e-commerce/internal/shared"
)

// MockNotificationRepository is a mock implementation of NotificationRepository
type MockNotificationRepository struct {
	notifications map[uuid.UUID]*Notification
}

func NewMockNotificationRepository() *MockNotificationRepository {
	return &MockNotificationRepository{notifications: make(map[uuid.UUID]*Notification)}
}

func (m *MockNotificationRepository) Create(ctx context.Context, notification *Notification) error {
	if notification.ID == uuid.Nil {
		notification.ID = uuid.New()
	}
	m.notifications[notification.ID] = notification
	return nil
}

func (m *MockNotificationRepository) CreateTx(ctx context.Context, tx *gorm.DB, notification *Notification) error {
	return m.Create(ctx, notification)
}

func (m *MockNotificationRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]Notification, error) {
	notifications := make([]Notification, 0)
	for _, n := range m.notifications {
		if n.UserID == userID {
			notifications = append(notifications, *n)
		}
	}
	return notifications, nil
}

func (m *MockNotificationRepository) FindByID(ctx context.Context, id uuid.UUID) (*Notification, error) {
	if notification, exists := m.notifications[id]; exists {
		return notification, nil
	}
	return nil, shared.ErrNotFound
}

func (m *MockNotificationRepository) MarkAsRead(ctx context.Context, id uuid.UUID) error {
	if notification, exists := m.notifications[id]; exists {
		notification.IsRead = true
		return nil
	}
	return shared.ErrNotFound
}

// Unit Tests

func TestNotificationService_CreateNotification(t *testing.T) {
	repo := NewMockNotificationRepository()
	service := NewNotificationService(repo)

	userID := uuid.New()
	title := "Test Title"
	message := "Test notification"

	notification, err := service.CreateNotification(context.Background(), userID, title, message)
	if err != nil {
		t.Errorf("CreateNotification() error = %v", err)
	}
	if notification == nil {
		t.Fatal("CreateNotification() returned nil")
	}
	if notification.UserID != userID {
		t.Errorf("CreateNotification() userID = %v, want %v", notification.UserID, userID)
	}
	if notification.Title != title {
		t.Errorf("CreateNotification() title = %v, want %v", notification.Title, title)
	}
	if notification.Message != message {
		t.Errorf("CreateNotification() message = %v, want %v", notification.Message, message)
	}
	if notification.IsRead {
		t.Error("CreateNotification() IsRead should be false")
	}
}

func TestNotificationService_GetNotifications(t *testing.T) {
	repo := NewMockNotificationRepository()
	service := NewNotificationService(repo)

	userID1 := uuid.New()
	userID2 := uuid.New()

	// Create notifications for user1
	_, err := service.CreateNotification(context.Background(), userID1, "Title 1", "Message 1")
	if err != nil {
		t.Fatalf("Failed to create notification: %v", err)
	}
	_, err = service.CreateNotification(context.Background(), userID1, "Title 2", "Message 2")
	if err != nil {
		t.Fatalf("Failed to create notification: %v", err)
	}

	// Create notification for user2
	_, err = service.CreateNotification(context.Background(), userID2, "Title 3", "Message 3")
	if err != nil {
		t.Fatalf("Failed to create notification: %v", err)
	}

	// Get notifications for user1
	notifications, err := service.GetNotifications(context.Background(), userID1)
	if err != nil {
		t.Errorf("GetNotifications() error = %v", err)
	}
	if len(notifications) != 2 {
		t.Errorf("GetNotifications() count = %d, want 2", len(notifications))
	}

	// Get notifications for user2
	notifications, err = service.GetNotifications(context.Background(), userID2)
	if err != nil {
		t.Errorf("GetNotifications() error = %v", err)
	}
	if len(notifications) != 1 {
		t.Errorf("GetNotifications() count = %d, want 1", len(notifications))
	}
}

func TestNotificationService_MarkAsRead(t *testing.T) {
	repo := NewMockNotificationRepository()
	service := NewNotificationService(repo)

	userID := uuid.New()
	otherUserID := uuid.New()

	// Create notification
	notification, err := service.CreateNotification(context.Background(), userID, "Test Title", "Test message")
	if err != nil {
		t.Fatalf("Failed to create notification: %v", err)
	}

	// Mark as read by owner
	err = service.MarkAsRead(context.Background(), userID, notification.ID)
	if err != nil {
		t.Errorf("MarkAsRead() error = %v", err)
	}

	// Verify it's marked as read
	got, err := repo.FindByID(context.Background(), notification.ID)
	if err != nil {
		t.Fatalf("Failed to get notification: %v", err)
	}
	if !got.IsRead {
		t.Error("MarkAsRead() notification should be marked as read")
	}

	// Try to mark as read by different user
	err = service.MarkAsRead(context.Background(), otherUserID, notification.ID)
	if err != shared.ErrForbidden {
		t.Errorf("MarkAsRead() by other user error = %v, want %v", err, shared.ErrForbidden)
	}
}

func TestNotificationService_MarkAsRead_NotFound(t *testing.T) {
	repo := NewMockNotificationRepository()
	service := NewNotificationService(repo)

	userID := uuid.New()
	nonExistentID := uuid.New()

	err := service.MarkAsRead(context.Background(), userID, nonExistentID)
	if err != shared.ErrNotFound {
		t.Errorf("MarkAsRead() error = %v, want %v", err, shared.ErrNotFound)
	}
}

// TestNotificationService_CreateNotificationTx tests CreateNotificationTx
func TestNotificationService_CreateNotificationTx(t *testing.T) {
	repo := NewMockNotificationRepository()
	service := NewNotificationService(repo)

	userID := uuid.New()
	title := "Test Title"
	message := "Test notification"
	var tx *gorm.DB // Mock transaction

	notification, err := service.CreateNotificationTx(context.Background(), tx, userID, title, message)
	if err != nil {
		t.Errorf("CreateNotificationTx() error = %v", err)
	}
	if notification == nil {
		t.Fatal("CreateNotificationTx() returned nil")
	}
	if notification.UserID != userID {
		t.Errorf("CreateNotificationTx() userID = %v, want %v", notification.UserID, userID)
	}
	if notification.Title != title {
		t.Errorf("CreateNotificationTx() title = %v, want %v", notification.Title, title)
	}
	if notification.Message != message {
		t.Errorf("CreateNotificationTx() message = %v, want %v", notification.Message, message)
	}
}

// Property 13: Notification Access Control
func TestProperty_NotificationAccessControl(t *testing.T) {
	repo := NewMockNotificationRepository()
	service := NewNotificationService(repo)

	userID1 := uuid.New()
	userID2 := uuid.New()

	// Create notification for user1
	notification, err := service.CreateNotification(context.Background(), userID1, "Title", "User1 notification")
	if err != nil {
		t.Fatalf("CreateNotification failed: %v", err)
	}

	// Property: Users can mark their own notifications as read
	err = service.MarkAsRead(context.Background(), userID1, notification.ID)
	if err != nil {
		t.Errorf("User should be able to mark own notification as read: %v", err)
	}

	// Create another notification for user1
	notification2, err := service.CreateNotification(context.Background(), userID1, "Title 2", "User1 notification 2")
	if err != nil {
		t.Fatalf("CreateNotification failed: %v", err)
	}

	// Property: Users cannot mark other users' notifications as read
	err = service.MarkAsRead(context.Background(), userID2, notification2.ID)
	if err != shared.ErrForbidden {
		t.Errorf("User should not be able to mark other user's notification as read, got: %v", err)
	}

	// Property: Users can only see their own notifications
	notifications, err := service.GetNotifications(context.Background(), userID1)
	if err != nil {
		t.Fatalf("GetNotifications failed: %v", err)
	}
	for _, n := range notifications {
		if n.UserID != userID1 {
			t.Errorf("User should only see own notifications, got notification for user %v", n.UserID)
		}
	}
}
