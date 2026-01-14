package notifications

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/mschl/ecommerce-api/internal/shared"
)

// NotificationService defines the interface for notification business logic
type NotificationService interface {
	CreateNotification(ctx context.Context, userID uuid.UUID, title, message string) (*Notification, error)
	CreateNotificationTx(ctx context.Context, tx *gorm.DB, userID uuid.UUID, title, message string) (*Notification, error)
	GetNotifications(ctx context.Context, userID uuid.UUID) ([]Notification, error)
	MarkAsRead(ctx context.Context, userID uuid.UUID, notificationID uuid.UUID) error
}

// notificationService implements NotificationService
type notificationService struct {
	repo NotificationRepository
}

// NewNotificationService creates a new NotificationService
func NewNotificationService(repo NotificationRepository) NotificationService {
	return &notificationService{repo: repo}
}

func (s *notificationService) CreateNotification(ctx context.Context, userID uuid.UUID, title, message string) (*Notification, error) {
	notification := &Notification{
		ID:      uuid.New(),
		UserID:  userID,
		Title:   title,
		Message: message,
		IsRead:  false,
	}

	if err := s.repo.Create(ctx, notification); err != nil {
		return nil, err
	}

	return notification, nil
}

func (s *notificationService) CreateNotificationTx(ctx context.Context, tx *gorm.DB, userID uuid.UUID, title, message string) (*Notification, error) {
	notification := &Notification{
		ID:      uuid.New(),
		UserID:  userID,
		Title:   title,
		Message: message,
		IsRead:  false,
	}

	if err := s.repo.CreateTx(ctx, tx, notification); err != nil {
		return nil, err
	}

	return notification, nil
}

func (s *notificationService) GetNotifications(ctx context.Context, userID uuid.UUID) ([]Notification, error) {
	return s.repo.FindByUserID(ctx, userID)
}

func (s *notificationService) MarkAsRead(ctx context.Context, userID uuid.UUID, notificationID uuid.UUID) error {
	// Get notification and verify ownership
	notification, err := s.repo.FindByID(ctx, notificationID)
	if err != nil {
		return err
	}

	if notification.UserID != userID {
		return shared.ErrForbidden
	}

	return s.repo.MarkAsRead(ctx, notificationID)
}
