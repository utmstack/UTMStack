package repository

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	mail_connectors "github.com/utmstack/utmstack/backend/internal/mail/connectors"
	appconfig_connectors "github.com/utmstack/utmstack/backend/modules/appconfig/connectors"
	"github.com/utmstack/utmstack/backend/modules/iam/connectors"
	"github.com/utmstack/utmstack/backend/pkg/constants/templates"
)

type userInvitationMailer struct {
	mail       mail_connectors.MailService
	configRepo mail_connectors.MailConfigurationRepository
	brand      appconfig_connectors.BrandNameProvider
}

func NewUserInvitationMailer(
	mail mail_connectors.MailService,
	configRepo mail_connectors.MailConfigurationRepository,
	brand appconfig_connectors.BrandNameProvider,
) connectors.UserInvitationMailer {
	return &userInvitationMailer{mail: mail, configRepo: configRepo, brand: brand}
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
		"FirstName":   firstName,
		"InviteURL":   inviteURL,
		"ProductName": m.brand.ProductName(ctx),
	}, "")
}
