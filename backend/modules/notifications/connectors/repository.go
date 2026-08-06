package connectors

import (
	"context"
	"time"

	"github.com/utmstack/utmstack/backend/modules/notifications/domain"
	"github.com/utmstack/utmstack/backend/modules/notifications/dto"
)

type NotificationRepository interface {
	Save(ctx context.Context, n *domain.UtmNotification) error
	FindByID(ctx context.Context, id int64) (*domain.UtmNotification, error)
	FindAll(ctx context.Context, q dto.NotificationListQuery) ([]domain.UtmNotification, int64, error)
	UpdateRead(ctx context.Context, id int64, read bool) (*domain.UtmNotification, error)
	UpdateStatus(ctx context.Context, id int64, status domain.NotificationStatus) (*domain.UtmNotification, error)
	MarkAllRead(ctx context.Context) (int64, error)
	CountUnread(ctx context.Context) (int64, error)
	Delete(ctx context.Context, id int64) error

	// DeleteOlderThan removes in bounded batches so a long-neglected table does
	// not become one enormous statement holding locks.
	DeleteOlderThan(ctx context.Context, cutoff time.Time, onlyRead bool, limit int) (int64, error)
}
