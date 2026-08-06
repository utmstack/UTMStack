package usecase

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"html/template"
	"mime/multipart"
	"net/smtp"
	"net/textproto"
	"strings"

	"github.com/utmstack/utmstack/backend/internal/mail/connectors"
	"github.com/utmstack/utmstack/backend/internal/mail/domain"
)

type mailService struct {
	repo connectors.MailConfigurationRepository
}

func New(repo connectors.MailConfigurationRepository) connectors.MailService {
	return &mailService{repo: repo}
}

func (s *mailService) SendMail(ctx context.Context, to []string, cc []string, subject, body string, attatchments []domain.Attatchment) error {
	cfg, err := s.repo.GetMailConfiguration(ctx)
	if err != nil {
		return fmt.Errorf("load mail config: %w", err)
	}
	return s.SendMailWithConfig(ctx, cfg, to, cc, subject, body, attatchments)
}

func (s *mailService) SendMailWithConfig(ctx context.Context, cfg *domain.EmailConfig, to []string, cc []string, subject, body string, attatchments []domain.Attatchment) error {
	if cfg == nil {
		return fmt.Errorf("mail configuration is nil")
	}
	if cfg.Host == "" || cfg.Port == "" {
		return fmt.Errorf("mail configuration is incomplete")
	}
	to = trimAddresses(to)
	cc = trimAddresses(cc)
	if len(to) == 0 {
		return fmt.Errorf("no recipients")
	}

	msg, err := buildMessage(cfg, to, cc, subject, body, attatchments)
	if err != nil {
		return err
	}

	addr := cfg.Host + ":" + cfg.Port
	auth := smtpAuth(cfg)
	from := senderAddress(cfg)
	// net/smtp needs every recipient in the RCPT TO list — headers alone don't
	// deliver mail. Merge to+cc for the envelope; the Cc header stays for display.
	rcpt := append(append([]string(nil), to...), cc...)
	return smtp.SendMail(addr, auth, from, rcpt, msg)
}

func (s *mailService) SendTemplateMail(ctx context.Context, to []string, subject, tmpl string, vars map[string]string, locale string) error {
	tpl, err := template.New("mail").Parse(tmpl)
	if err != nil {
		return fmt.Errorf("parse template: %w", err)
	}
	var buf bytes.Buffer
	data := map[string]any{"Vars": vars, "Locale": locale}
	for k, v := range vars {
		data[k] = v
	}
	if err := tpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("render template: %w", err)
	}
	return s.SendMail(ctx, to, nil, subject, buf.String(), nil)
}

func smtpAuth(cfg *domain.EmailConfig) smtp.Auth {
	if !strings.EqualFold(cfg.SmtpAuth, "true") || cfg.Username == "" {
		return nil
	}
	return smtp.PlainAuth("", cfg.Username, cfg.Password, cfg.Host)
}

func senderAddress(cfg *domain.EmailConfig) string {
	if cfg.From != "" {
		return cfg.From
	}
	return cfg.Username
}

// trimAddresses drops empty strings from a recipient list without touching order.
func trimAddresses(in []string) []string {
	if len(in) == 0 {
		return nil
	}
	out := make([]string, 0, len(in))
	for _, a := range in {
		a = strings.TrimSpace(a)
		if a != "" {
			out = append(out, a)
		}
	}
	return out
}

func buildMessage(cfg *domain.EmailConfig, to []string, cc []string, subject, body string, attatchments []domain.Attatchment) ([]byte, error) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	headers := textproto.MIMEHeader{}
	headers.Set("From", senderAddress(cfg))
	headers.Set("To", strings.Join(to, ", "))
	if len(cc) > 0 {
		headers.Set("Cc", strings.Join(cc, ", "))
	}
	if subject != "" {
		headers.Set("Subject", subject)
	}
	headers.Set("MIME-Version", "1.0")
	headers.Set("Content-Type", "multipart/mixed; boundary="+writer.Boundary())
	if cfg.Orgname != "" {
		headers.Set("X-Organization", cfg.Orgname)
	}

	var head bytes.Buffer
	for k, vs := range headers {
		for _, v := range vs {
			fmt.Fprintf(&head, "%s: %s\r\n", k, v)
		}
	}
	head.WriteString("\r\n")

	bodyHeader := textproto.MIMEHeader{}
	bodyHeader.Set("Content-Type", "text/html; charset=UTF-8")
	bodyHeader.Set("Content-Transfer-Encoding", "quoted-printable")
	part, err := writer.CreatePart(bodyHeader)
	if err != nil {
		return nil, fmt.Errorf("create body part: %w", err)
	}
	if _, err := part.Write([]byte(body)); err != nil {
		return nil, fmt.Errorf("write body: %w", err)
	}

	for _, a := range attatchments {
		attHeader := textproto.MIMEHeader{}
		ct := a.ContentType
		if ct == "" {
			ct = "application/octet-stream"
		}
		attHeader.Set("Content-Type", ct)
		attHeader.Set("Content-Transfer-Encoding", "base64")
		attHeader.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", a.Filename))
		ap, err := writer.CreatePart(attHeader)
		if err != nil {
			return nil, fmt.Errorf("create attachment part: %w", err)
		}
		encoded := make([]byte, base64.StdEncoding.EncodedLen(len(a.Bytes)))
		base64.StdEncoding.Encode(encoded, a.Bytes)
		if _, err := ap.Write(encoded); err != nil {
			return nil, fmt.Errorf("write attachment: %w", err)
		}
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("close multipart: %w", err)
	}

	return append(head.Bytes(), buf.Bytes()...), nil
}
