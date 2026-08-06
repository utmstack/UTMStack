package usecase

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	mail_connectors "github.com/utmstack/utmstack/backend/internal/mail/connectors"
	mail_domain "github.com/utmstack/utmstack/backend/internal/mail/domain"
	"github.com/utmstack/utmstack/backend/modules/appconfig/connectors"
	"github.com/utmstack/utmstack/backend/modules/appconfig/domain"
	"github.com/utmstack/utmstack/backend/modules/appconfig/dto"
	"github.com/utmstack/utmstack/backend/pkg/secret"
)

type service struct {
	repo   connectors.Repository
	cipher *secret.Cipher
	mailer mail_connectors.MailService
	brand  connectors.BrandNameProvider
}

func New(repo connectors.Repository, cipher *secret.Cipher, brand connectors.BrandNameProvider) *service {
	return &service{repo: repo, cipher: cipher, brand: brand}
}

func (s *service) SetMailer(m mail_connectors.MailService) { s.mailer = m }

func (s *service) List(ctx context.Context) ([]dto.ConfigResponse, error) {
	rows, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]dto.ConfigResponse, 0, len(rows))
	for _, r := range rows {
		out = append(out, s.toResponse(r))
	}
	return out, nil
}

func (s *service) Get(ctx context.Context, key string) (*dto.ConfigResponse, error) {
	row, err := s.repo.GetByKey(ctx, key)
	if err != nil {
		return nil, err
	}
	if row == nil {
		return nil, nil
	}
	resp := s.toResponse(*row)
	return &resp, nil
}

// Update changes the value (and optionally metadata) of an existing, seeded
// parameter. It never creates a row: an unknown key returns (nil, nil) so the
// handler can answer 404.
func (s *service) Update(ctx context.Context, actor, key string, input dto.UpsertRequest) (*dto.ConfigResponse, error) {
	existing, err := s.repo.GetByKey(ctx, key)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, nil
	}

	row := *existing
	now := time.Now().UTC()
	row.ModificationTime = &now
	row.ModificationUser = actor

	if input.IsSecret != nil && *input.IsSecret {
		row.SetIsSecret()
	}
	if input.Label != "" {
		row.ConfParamLarge = input.Label
	}
	if input.Description != "" {
		row.ConfParamDescription = input.Description
	}

	if input.Value == "" {
		row.ConfParamValue = ""
	} else if row.IsSecret() {
		enc, err := s.cipher.Encrypt(input.Value)
		if err != nil {
			return nil, err
		}
		row.ConfParamValue = enc
	} else {
		row.ConfParamValue = input.Value
	}

	if err := s.repo.Save(ctx, &row); err != nil {
		return nil, err
	}
	resp := s.toResponse(row)
	return &resp, nil
}

func (s *service) GetString(ctx context.Context, key string) (string, bool, error) {
	row, err := s.repo.GetByKey(ctx, key)
	if err != nil {
		return "", false, err
	}
	if row == nil || row.ConfParamValue == "" {
		return "", false, nil
	}
	if row.IsSecret() {
		plain, err := s.cipher.Decrypt(row.ConfParamValue)
		if err != nil {
			return "", false, err
		}
		return plain, true, nil
	}
	return row.ConfParamValue, true, nil
}

// SetString persists a value onto an existing seeded parameter (used by the mail
// provider when saving SMTP settings). Unknown keys are rejected since
// parameters are seeded, never created here.
func (s *service) SetString(ctx context.Context, key, value string, opts connectors.SetOpts) error {
	existing, err := s.repo.GetByKey(ctx, key)
	if err != nil {
		return err
	}
	if existing == nil {
		return fmt.Errorf("config parameter %q does not exist", key)
	}

	row := *existing
	stored := value
	if opts.IsSecret {
		row.SetIsSecret()
		if value != "" {
			enc, err := s.cipher.Encrypt(value)
			if err != nil {
				return err
			}
			stored = enc
		}
	}
	row.ConfParamValue = stored
	if opts.Label != "" {
		row.ConfParamLarge = opts.Label
	}
	if opts.Description != "" {
		row.ConfParamDescription = opts.Description
	}
	now := time.Now().UTC()
	row.ModificationTime = &now
	row.ModificationUser = "system"
	return s.repo.Save(ctx, &row)
}

// CheckMail attempts to send a verification email through each supplied
// configuration. The recipient is the config's own From address (self-ping):
// success means the SMTP server accepted the message. The first failure is
// returned with the offending index so the caller can identify which
// configuration is broken.
func (s *service) CheckMail(ctx context.Context, configs []domain.MailConfig) error {
	if s.mailer == nil {
		return fmt.Errorf("mailer not configured")
	}
	if len(configs) == 0 {
		return fmt.Errorf("no mail configurations supplied")
	}
	body := fmt.Sprintf(`<html><body><p>%s mail configuration test — if you received this, your SMTP settings work.</p></body></html>`, s.brand.ProductName(ctx))
	for i, cfg := range configs {
		if cfg.From == "" {
			return fmt.Errorf("config[%d]: from address is required to send a test message", i)
		}
		// Let the UI test a previously-saved config without re-entering the secret:
		// an empty password falls back to the stored (decrypted) one.
		if cfg.Password == "" {
			if stored, ok, _ := s.GetString(ctx, "utmstack.mail.password"); ok {
				cfg.Password = stored
			}
		}
		ec := toEmailConfig(cfg)
		if err := s.mailer.SendMailWithConfig(ctx, &ec, []string{cfg.From}, nil, "Mail configuration test", body, nil); err != nil {
			return fmt.Errorf("config[%d] (%s:%d): %w", i, cfg.Host, cfg.Port, err)
		}
	}
	return nil
}

// GetDateFormat returns the org-wide timestamp display preference (timezone +
// format), falling back to UTC/medium. Read-only, non-secret — exposed publicly
// so every client can render timestamps consistently.
func (s *service) GetDateFormat(ctx context.Context) dto.DateFormatResponse {
	out := dto.DateFormatResponse{Timezone: "UTC", DateFormat: "medium"}
	if v, ok, _ := s.GetString(ctx, "utmstack.time.zone"); ok && v != "" {
		out.Timezone = v
	}
	if v, ok, _ := s.GetString(ctx, "utmstack.time.dateformat"); ok && v != "" {
		out.DateFormat = v
	}
	return out
}

func toEmailConfig(c domain.MailConfig) mail_domain.EmailConfig {
	smtpAuth := ""
	if at := strings.ToLower(strings.TrimSpace(c.AuthType)); at != "" && at != "none" && at != "false" {
		smtpAuth = "true"
	}
	return mail_domain.EmailConfig{
		Host:     c.Host,
		Port:     strconv.Itoa(c.Port),
		Username: c.Username,
		Password: c.Password,
		From:     c.From,
		SmtpAuth: smtpAuth,
	}
}

func (s *service) toResponse(c domain.Config) dto.ConfigResponse {
	r := dto.ConfigResponse{
		Key:         c.ConfParamShort,
		IsSecret:    c.IsSecret(),
		IsSet:       c.ConfParamValue != "",
		Label:       c.ConfParamLarge,
		Description: c.ConfParamDescription,
		UpdatedBy:   c.ModificationUser,
	}
	if c.ModificationTime != nil {
		r.UpdatedAt = *c.ModificationTime
	}
	if c.ConfParamValue == "" {
		return r
	}
	if !c.IsSecret() {
		r.Value = c.ConfParamValue
	}
	return r
}
