package domain

import (
	"time"
)

type UtmNotification struct {
	ID        int64              `gorm:"primaryKey;autoIncrement"                       json:"id"`
	Source    NotificationSource `gorm:"column:source;size:64;not null;index"           json:"source"`
	Type      NotificationType   `gorm:"column:type;size:16;not null"                   json:"type"`
	Message   string             `gorm:"column:message;type:text;not null"              json:"message"`
	CreatedAt time.Time          `gorm:"column:created_at;not null;index:,sort:desc"    json:"createdAt"`
	UpdatedAt *time.Time         `gorm:"column:updated_at"                              json:"updatedAt,omitempty"`
	Read      bool               `gorm:"column:read;not null;default:false"             json:"read"`
	Status    NotificationStatus `gorm:"column:status;size:16;not null;default:ACTIVE"  json:"status"`
}

func (UtmNotification) TableName() string { return "utm_notification" }

