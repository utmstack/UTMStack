package dto

import (
	"time"

	"github.com/utmstack/utmstack/backend/modules/notifications/domain"
)

type NotificationListQuery struct {
	Source *domain.NotificationSource
	Type   *domain.NotificationType
	Status *domain.NotificationStatus
	From   *time.Time
	To     *time.Time
	Read   *bool
	Page   int
	Size   int
	Sort   string
}
