package repository

import (
	"context"
	"strconv"

	"github.com/utmstack/utmstack/backend/internal/mail/connectors"
	"github.com/utmstack/utmstack/backend/internal/mail/domain"
	app_cfg_connectors "github.com/utmstack/utmstack/backend/modules/appconfig/connectors"
	"github.com/utmstack/utmstack/backend/pkg/constants"
)

type mailConfigurationRepository struct {
	store app_cfg_connectors.Store
}

func NewMailConfigurationRepository(store app_cfg_connectors.Store) connectors.MailConfigurationRepository {
	return &mailConfigurationRepository{store: store}
}

func (self *mailConfigurationRepository) GetMailConfiguration(ctx context.Context) (*domain.EmailConfig, error) {
	cfg := domain.EmailConfig{
		ProtocolValue: constants.PROP_EMAIL_PROTOCOL_VALUE,
		PortTlsValue:  strconv.Itoa(constants.PROP_EMAIL_PORT_TLS_VALUE),
		PortSslValue:  strconv.Itoa(constants.PROP_EMAIL_PORT_SSL_VALUE),
		PortNoneValue: strconv.Itoa(constants.PROP_EMAIL_PORT_NONE_VALUE),
	}

	for _, f := range []struct {
		key string
		dst *string
	}{
		{constants.PROP_MAIL_HOST, &cfg.Host},
		{constants.PROP_MAIL_PORT, &cfg.Port},
		{constants.PROP_MAIL_PASSWORD, &cfg.Password},
		{constants.PROP_MAIL_USERNAME, &cfg.Username},
		{constants.PROP_MAIL_ORGNAME, &cfg.Orgname},
		{constants.PROP_MAIL_SMTP_AUTH, &cfg.SmtpAuth},
		{constants.PROP_MAIL_FROM, &cfg.From},
		{constants.PROP_MAIL_BASE_URL, &cfg.BaseUrl},
	} {
		v, _, err := self.store.GetString(ctx, f.key)
		if err != nil {
			return nil, err
		}
		*f.dst = v
	}

	return &cfg, nil
}

func (self *mailConfigurationRepository) SetMailConfiguration(ctx context.Context, config domain.EmailConfig) error {
	entries := []struct {
		key      string
		value    string
		isSecret bool
		label    string
	}{
		{constants.PROP_MAIL_HOST, config.Host, false, "Mail Host"},
		{constants.PROP_MAIL_PORT, config.Port, false, "Mail Port"},
		{constants.PROP_MAIL_USERNAME, config.Username, false, "Mail Username"},
		{constants.PROP_MAIL_PASSWORD, config.Password, true, "Mail Password"},
		{constants.PROP_MAIL_ORGNAME, config.Orgname, false, "Mail Organization"},
		{constants.PROP_MAIL_SMTP_AUTH, config.SmtpAuth, false, "SMTP Auth"},
		{constants.PROP_MAIL_FROM, config.From, false, "Mail From"},
		{constants.PROP_MAIL_BASE_URL, config.BaseUrl, false, "Mail Base URL"},
	}

	for _, e := range entries {
		opts := app_cfg_connectors.SetOpts{IsSecret: e.isSecret, Label: e.label}
		if err := self.store.SetString(ctx, e.key, e.value, opts); err != nil {
			return err
		}
	}
	return nil
}
