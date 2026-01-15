package notifications

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/mschl/ecommerce-api/internal/shared"
)

// NotificationRepository definiert Interface für Benachrichtigungs-Datenzugriff
type NotificationRepository interface {
	Create(ctx context.Context, notification *Notification) error
	CreateTx(ctx context.Context, tx *gorm.DB, notification *Notification) error
	FindByUserID(ctx context.Context, userID uuid.UUID) ([]Notification, error)
	FindByID(ctx context.Context, id uuid.UUID) (*Notification, error)
	MarkAsRead(ctx context.Context, id uuid.UUID) error
}

// notificationRepository implementiert NotificationRepository mit GORM
type notificationRepository struct {
	db *gorm.DB
}

// NewNotificationRepository erstellt ein neues NotificationRepository
func NewNotificationRepository(db *gorm.DB) NotificationRepository {
	return &notificationRepository{db: db}
}

func (r *notificationRepository) Create(ctx context.Context, notification *Notification) error {
	return r.CreateTx(ctx, r.db, notification)
}

func (r *notificationRepository) CreateTx(ctx context.Context, tx *gorm.DB, notification *Notification) error {
	if notification.ID == uuid.Nil {
		notification.ID = uuid.New()
	}
	return tx.WithContext(ctx).Create(notification).Error
}

func (r *notificationRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]Notification, error) {
	var notifications []Notification
	result := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&notifications)
	if result.Error != nil {
		return nil, result.Error
	}
	return notifications, nil
}

func (r *notificationRepository) FindByID(ctx context.Context, id uuid.UUID) (*Notification, error) {
	var notification Notification
	result := r.db.WithContext(ctx).First(&notification, "id = ?", id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, shared.ErrNotFound
		}
		return nil, result.Error
	}
	return &notification, nil
}

func (r *notificationRepository) MarkAsRead(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Model(&Notification{}).
		Where("id = ?", id).
		Update("is_read", true)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return shared.ErrNotFound
	}
	return nil
}
