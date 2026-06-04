package connectors

import (
	"context"

	"github.com/utmstack/utmstack/backend/modules/notifications/domain"
	"github.com/utmstack/utmstack/backend/modules/notifications/dto"
)

type NotificationUsecase interface {
	Create(ctx context.Context, req dto.CreateNotificationRequest) (*domain.UtmNotification, error)
	List(ctx context.Context, q dto.NotificationListQuery) ([]domain.UtmNotification, int64, error)
	GetByID(ctx context.Context, id int64) (*domain.UtmNotification, error)
	MarkRead(ctx context.Context, id int64, read bool) (*domain.UtmNotification, error)
	UpdateStatus(ctx context.Context, id int64, status domain.NotificationStatus) (*domain.UtmNotification, error)
	MarkAllRead(ctx context.Context) error
	CountUnread(ctx context.Context) (int64, error)
	Delete(ctx context.Context, id int64) error
}

type Producer interface {
	Notify(ctx context.Context, source domain.NotificationSource, ntype domain.NotificationType, message string) error
}
