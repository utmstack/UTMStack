package connectors

import (
	"context"

	"github.com/utmstack/utmstack/backend/modules/notifications/domain"
	"github.com/utmstack/utmstack/backend/modules/notifications/dto"
)

type NotificationRepository interface {
	Save(ctx context.Context, n *domain.UtmNotification) error
	FindByID(ctx context.Context, id int64) (*domain.UtmNotification, error)
	FindAll(ctx context.Context, q dto.NotificationListQuery) ([]domain.UtmNotification, int64, error)
	FindAllGrouped(ctx context.Context, q dto.NotificationListQuery) ([]domain.NotificationGroup, int64, error)
	UpdateRead(ctx context.Context, id int64, read bool) (*domain.UtmNotification, error)
	UpdateStatus(ctx context.Context, id int64, status domain.NotificationStatus) (*domain.UtmNotification, error)
	MarkAllRead(ctx context.Context) (int64, error)
	CountUnread(ctx context.Context) (int64, error)
	Delete(ctx context.Context, id int64) error
}
