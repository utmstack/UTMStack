package dto

import (
	"time"

	"github.com/utmstack/utmstack/backend/modules/notifications/domain"
	"github.com/utmstack/utmstack/backend/pkg/database"
)

type NotificationListQuery struct {
	Source *domain.NotificationSource
	Type   *domain.NotificationType
	Status *domain.NotificationStatus
	From   *time.Time
	To     *time.Time
	Read   *bool
	database.Params
	Sort string
}
