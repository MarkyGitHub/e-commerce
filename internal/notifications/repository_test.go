package notifications

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/mschl/ecommerce-api/internal/shared"
	"github.com/mschl/ecommerce-api/internal/testutil"
)

// TestUser is a local user struct for repository tests to avoid import cycle with auth package
type TestUser struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey"`
	Username     string    `gorm:"not null"`
	Email        string    `gorm:"uniqueIndex;not null"`
	PasswordHash string    `gorm:"not null"`
	Role         string    `gorm:"not null;default:customer"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (TestUser) TableName() string {
	return "users"
}

func setupNotificationsTestDB(t *testing.T) *gorm.DB {
	testDB := testutil.SetupTestDB(t)
	if err := testDB.AutoMigrate(&TestUser{}, &Notification{}); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}
	return testDB
}

func TestNotificationRepository_Create(t *testing.T) {
	testDB := setupNotificationsTestDB(t)
	notificationRepo := NewNotificationRepository(testDB)
	ctx := context.Background()

	// Create test user
	user := &TestUser{
		ID:           uuid.New(),
		Email:        "notify@example.com",
		PasswordHash: "hash",
		Username:     "Notify User",
		Role:         "customer",
	}
	if err := testDB.WithContext(ctx).Create(user).Error; err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	notification := &Notification{
		ID:      uuid.New(),
		UserID:  user.ID,
		Message: "Test notification",
		IsRead:  false,
	}

	if err := notificationRepo.Create(ctx, notification); err != nil {
		t.Errorf("Create() error = %v, want nil", err)
	}

	// Verify notification was created
	found, err := notificationRepo.FindByID(ctx, notification.ID)
	if err != nil {
		t.Errorf("FindByID() error = %v", err)
	}
	if found.Message != "Test notification" {
		t.Errorf("FindByID() message = %s, want Test notification", found.Message)
	}
	if found.IsRead {
		t.Error("FindByID() IsRead = true, want false")
	}
}

func TestNotificationRepository_CreateTx(t *testing.T) {
	testDB := setupNotificationsTestDB(t)
	notificationRepo := NewNotificationRepository(testDB)
	ctx := context.Background()

	// Create test user
	user := &TestUser{
		ID:           uuid.New(),
		Email:        "notifytx@example.com",
		PasswordHash: "hash",
		Username:     "Notify Tx User",
		Role:         "customer",
	}
	if err := testDB.WithContext(ctx).Create(user).Error; err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Test transaction-based create
	tx := testDB.Begin()
	defer tx.Rollback()

	notification := &Notification{
		ID:      uuid.New(),
		UserID:  user.ID,
		Message: "Tx notification",
		IsRead:  false,
	}

	if err := notificationRepo.CreateTx(ctx, tx, notification); err != nil {
		t.Errorf("CreateTx() error = %v, want nil", err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		t.Fatalf("Commit() error = %v", err)
	}

	// Verify notification was created
	found, err := notificationRepo.FindByID(ctx, notification.ID)
	if err != nil {
		t.Errorf("FindByID() error = %v", err)
	}
	if found.Message != "Tx notification" {
		t.Errorf("FindByID() message = %s, want Tx notification", found.Message)
	}
}

func TestNotificationRepository_FindByUserID(t *testing.T) {
	testDB := setupNotificationsTestDB(t)
	notificationRepo := NewNotificationRepository(testDB)
	ctx := context.Background()

	// Create test users
	user1 := &TestUser{
		ID:           uuid.New(),
		Email:        "notify1@example.com",
		PasswordHash: "hash",
		Username:     "Notify User 1",
		Role:         "customer",
	}
	user2 := &TestUser{
		ID:           uuid.New(),
		Email:        "notify2@example.com",
		PasswordHash: "hash",
		Username:     "Notify User 2",
		Role:         "customer",
	}
	if err := testDB.WithContext(ctx).Create(user1).Error; err != nil {
		t.Fatalf("Failed to create user1: %v", err)
	}
	if err := testDB.WithContext(ctx).Create(user2).Error; err != nil {
		t.Fatalf("Failed to create user2: %v", err)
	}

	// Create notifications for user1
	notif1 := &Notification{ID: uuid.New(), UserID: user1.ID, Message: "Message 1", IsRead: false}
	notif2 := &Notification{ID: uuid.New(), UserID: user1.ID, Message: "Message 2", IsRead: false}
	// Create notification for user2
	notif3 := &Notification{ID: uuid.New(), UserID: user2.ID, Message: "Message 3", IsRead: false}

	if err := notificationRepo.Create(ctx, notif1); err != nil {
		t.Fatalf("Failed to create notif1: %v", err)
	}
	if err := notificationRepo.Create(ctx, notif2); err != nil {
		t.Fatalf("Failed to create notif2: %v", err)
	}
	if err := notificationRepo.Create(ctx, notif3); err != nil {
		t.Fatalf("Failed to create notif3: %v", err)
	}

	// Find notifications for user1
	notifications, err := notificationRepo.FindByUserID(ctx, user1.ID)
	if err != nil {
		t.Errorf("FindByUserID() error = %v, want nil", err)
	}
	if len(notifications) != 2 {
		t.Errorf("FindByUserID() count = %d, want 2", len(notifications))
	}

	// Verify all notifications belong to user1
	for _, n := range notifications {
		if n.UserID != user1.ID {
			t.Errorf("FindByUserID() returned notification for wrong user: %v", n.UserID)
		}
	}

	// Verify notifications are ordered by created_at DESC (newest first)
	if len(notifications) >= 2 {
		if notifications[0].CreatedAt.Before(notifications[1].CreatedAt) {
			t.Error("FindByUserID() notifications should be ordered DESC (newest first)")
		}
	}
}

func TestNotificationRepository_FindByID(t *testing.T) {
	testDB := setupNotificationsTestDB(t)
	notificationRepo := NewNotificationRepository(testDB)
	ctx := context.Background()

	// Create test user and notification
	user := &TestUser{
		ID:           uuid.New(),
		Email:        "findnotify@example.com",
		PasswordHash: "hash",
		Username:     "Find Notify User",
		Role:         "customer",
	}
	if err := testDB.WithContext(ctx).Create(user).Error; err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	notification := &Notification{
		ID:      uuid.New(),
		UserID:  user.ID,
		Message: "Find notification",
		IsRead:  false,
	}
	if err := notificationRepo.Create(ctx, notification); err != nil {
		t.Fatalf("Failed to create test notification: %v", err)
	}

	tests := []struct {
		name    string
		id      uuid.UUID
		wantErr error
		wantID  uuid.UUID
	}{
		{
			name:    "find existing notification",
			id:      notification.ID,
			wantErr: nil,
			wantID:  notification.ID,
		},
		{
			name:    "notification not found",
			id:      uuid.New(),
			wantErr: shared.ErrNotFound,
			wantID:  uuid.Nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			found, err := notificationRepo.FindByID(ctx, tt.id)
			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Errorf("FindByID() error = %v, want %v", err, tt.wantErr)
				}
				if found != nil {
					t.Errorf("FindByID() found = %v, want nil", found)
				}
			} else {
				if err != nil {
					t.Errorf("FindByID() error = %v, want nil", err)
				}
				if found.ID != tt.wantID {
					t.Errorf("FindByID() id = %v, want %v", found.ID, tt.wantID)
				}
			}
		})
	}
}

func TestNotificationRepository_MarkAsRead(t *testing.T) {
	testDB := setupNotificationsTestDB(t)
	notificationRepo := NewNotificationRepository(testDB)
	ctx := context.Background()

	// Create test user and notification
	user := &TestUser{
		ID:           uuid.New(),
		Email:        "markread@example.com",
		PasswordHash: "hash",
		Username:     "Mark Read User",
		Role:         "customer",
	}
	if err := testDB.WithContext(ctx).Create(user).Error; err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	notification := &Notification{
		ID:      uuid.New(),
		UserID:  user.ID,
		Message: "Unread notification",
		IsRead:  false,
	}
	if err := notificationRepo.Create(ctx, notification); err != nil {
		t.Fatalf("Failed to create test notification: %v", err)
	}

	// Verify notification is unread
	found, err := notificationRepo.FindByID(ctx, notification.ID)
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}
	if found.IsRead {
		t.Error("Notification should be unread before MarkAsRead")
	}

	// Mark as read
	if err := notificationRepo.MarkAsRead(ctx, notification.ID); err != nil {
		t.Errorf("MarkAsRead() error = %v, want nil", err)
	}

	// Verify notification is read
	found, err = notificationRepo.FindByID(ctx, notification.ID)
	if err != nil {
		t.Fatalf("FindByID() after MarkAsRead error = %v", err)
	}
	if !found.IsRead {
		t.Error("Notification should be read after MarkAsRead")
	}

	// Test mark non-existent notification
	err = notificationRepo.MarkAsRead(ctx, uuid.New())
	if err != shared.ErrNotFound {
		t.Errorf("MarkAsRead() non-existent error = %v, want %v", err, shared.ErrNotFound)
	}
}
