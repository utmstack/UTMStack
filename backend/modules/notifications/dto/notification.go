package dto

import (
	"time"

	"github.com/utmstack/utmstack/backend/modules/notifications/domain"
)

type CreateNotificationRequest struct {
	Source  domain.NotificationSource `json:"source"  binding:"required,oneof=AS400 BIT_DEFENDER AWS AZURE OFFICE_365 SOPHOS GOOGLE EMAIL_SETTING ALERTS SYSTEM"`
	Type    domain.NotificationType   `json:"type"    binding:"required,oneof=INFO WARNING ERROR"`
	Message string                    `json:"message" binding:"required,min=1"`
}

type NotificationResponse struct {
	ID        int64                     `json:"id"`
	Source    domain.NotificationSource `json:"source"`
	Type      domain.NotificationType   `json:"type"`
	Message   string                    `json:"message"`
	CreatedAt time.Time                 `json:"createdAt"`
	UpdatedAt *time.Time                `json:"updatedAt,omitempty"`
	Read      bool                      `json:"read"`
	Status    domain.NotificationStatus `json:"status"`
}

func FromEntity(n *domain.UtmNotification) NotificationResponse {
	return NotificationResponse{
		ID:        n.ID,
		Source:    n.Source,
		Type:      n.Type,
		Message:   n.Message,
		CreatedAt: n.CreatedAt,
		UpdatedAt: n.UpdatedAt,
		Read:      n.Read,
		Status:    n.Status,
	}
}
