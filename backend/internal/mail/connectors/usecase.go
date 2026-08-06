package connectors

import (
	"context"

	"github.com/utmstack/utmstack/backend/internal/mail/domain"
)

type MailService interface {
	SendMail(ctx context.Context, address []string, subject, body string, attatchments []domain.Attatchment) error
	SendTemplateMail(ctx context.Context, address []string, subject, template string, vars map[string]string, locale string) error
	SendMailWithConfig(ctx context.Context, cfg *domain.EmailConfig, address []string, subject, body string, attatchments []domain.Attatchment) error
}
