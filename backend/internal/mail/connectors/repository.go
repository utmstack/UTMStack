package connectors

import (
	"context"

	"github.com/utmstack/utmstack/backend/internal/mail/domain"
)

type MailConfigurationRepository interface{
	GetMailConfiguration(ctx context.Context) (*domain.EmailConfig,error)
	SetMailConfiguration(ctx context.Context, config domain.EmailConfig) error
}
