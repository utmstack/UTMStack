package repository

import (
	"context"

	"github.com/utmstack/utmstack/backend/modules/iam/connectors"
	mail_connectors "github.com/utmstack/utmstack/backend/internal/mail/connectors"
	"github.com/utmstack/utmstack/backend/pkg/constants/templates"
)

type tfaMailer struct {
	mail       mail_connectors.MailService
	configRepo mail_connectors.MailConfigurationRepository
}

func NewTfaMailer(
	mail mail_connectors.MailService,
	configRepo mail_connectors.MailConfigurationRepository,
) connectors.TfaMailer {
	return &tfaMailer{mail: mail, configRepo: configRepo}
}

func (m *tfaMailer) SendTfaCode(ctx context.Context, to, firstName, code string) error {
	return m.mail.SendTemplateMail(ctx, []string{to}, templates.TfaCode, map[string]string{
		"FirstName": firstName,
		"Code":      code,
	}, "")
}
