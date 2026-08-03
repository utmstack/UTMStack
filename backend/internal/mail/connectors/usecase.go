package connectors

import (
	"context"

	"github.com/utmstack/utmstack/backend/internal/mail/domain"
)


type MailService interface {
	// SendMail sends body to `to` with optional cc recipients. cc may be nil/empty.
	// subject may be empty (compliance-report legacy senders pass their subject
	// text inside the body — new callers should pass a proper subject).
	SendMail(ctx context.Context, to []string, cc []string, subject, body string, attatchments []domain.Attatchment) error
	SendTemplateMail(ctx context.Context, to []string, template string, vars map[string]string, locale string) error
	// SendMailWithConfig sends a message using the supplied EmailConfig instead
	// of loading one from storage. Used by orchestration flows such as
	// "test current mail settings" where the config under test is not persisted.
	SendMailWithConfig(ctx context.Context, cfg *domain.EmailConfig, to []string, cc []string, subject, body string, attatchments []domain.Attatchment) error
}
