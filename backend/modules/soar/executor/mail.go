package executor

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	maildomain "github.com/utmstack/utmstack/backend/internal/mail/domain"
	soardomain "github.com/utmstack/utmstack/backend/modules/soar/domain"
)

// MailSender is the narrow slice of the mail service that the SOAR mail
// executor consumes. Keeps this package free of the broader mail import and
// lets tests swap in a fake.
type MailSender interface {
	SendMail(ctx context.Context, to []string, cc []string, subject, body string, attachments []maildomain.Attatchment) error
}

// Mail sends an email via the tenant's configured SMTP settings. Params are
// to, cc, subject, body — all $()-templates are already interpolated by the
// dispatcher before Execute runs.
// ponytail: no attachments (YAGNI); comma-split addresses, no header parser.
type Mail struct{ client MailSender }

func NewMail(c MailSender) *Mail { return &Mail{client: c} }

func (Mail) Type() string { return "mail" }

type mailParams struct {
	To      string `json:"to"`
	CC      string `json:"cc,omitempty"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

func (m *Mail) Execute(ctx context.Context, exec *soardomain.SoarExecution) (json.RawMessage, error) {
	if m.client == nil {
		return nil, errors.New("soar mail: client not configured")
	}
	var p mailParams
	if len(exec.Params) > 0 {
		if err := json.Unmarshal(exec.Params, &p); err != nil {
			return nil, fmt.Errorf("soar mail: params: %w", err)
		}
	}
	to := splitAddresses(p.To)
	if len(to) == 0 {
		return nil, errors.New("soar mail: at least one recipient is required")
	}
	if strings.TrimSpace(p.Subject) == "" {
		return nil, errors.New("soar mail: subject is required")
	}
	cc := splitAddresses(p.CC)
	if err := m.client.SendMail(ctx, to, cc, p.Subject, p.Body, nil); err != nil {
		return nil, fmt.Errorf("soar mail: send: %w", err)
	}
	exec.Result = fmt.Sprintf("sent email to %d recipient(s): %s", len(to), p.Subject)
	return nil, nil
}

func splitAddresses(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		if addr := strings.TrimSpace(part); addr != "" {
			out = append(out, addr)
		}
	}
	return out
}
