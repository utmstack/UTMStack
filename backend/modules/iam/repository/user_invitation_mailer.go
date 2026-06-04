package repository

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/utmstack/utmstack/backend/modules/iam/connectors"
	mail_connectors "github.com/utmstack/utmstack/backend/internal/mail/connectors"
	"github.com/utmstack/utmstack/backend/pkg/constants/templates"
)

type userInvitationMailer struct {
	mail       mail_connectors.MailService
	configRepo mail_connectors.MailConfigurationRepository
}

func NewUserInvitationMailer(
	mail mail_connectors.MailService,
	configRepo mail_connectors.MailConfigurationRepository,
) connectors.UserInvitationMailer {
	return &userInvitationMailer{mail: mail, configRepo: configRepo}
}

func (m *userInvitationMailer) SendInvitation(ctx context.Context, to, firstName, resetKey string) error {
	cfg, err := m.configRepo.GetMailConfiguration(ctx)
	if err != nil {
		return fmt.Errorf("load mail config: %w", err)
	}
	if cfg.BaseUrl == "" {
		return fmt.Errorf("mail base url is not configured")
	}
	inviteURL := fmt.Sprintf("%s/reset/finish?key=%s",
		strings.TrimRight(cfg.BaseUrl, "/"),
		url.QueryEscape(resetKey),
	)
	return m.mail.SendTemplateMail(ctx, []string{to}, templates.UserInvitation, map[string]string{
		"FirstName": firstName,
		"InviteURL": inviteURL,
	}, "")
}
