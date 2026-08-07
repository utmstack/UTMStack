package incidents

import (
	"context"
	"fmt"
	"html"
	"strings"

	"github.com/threatwinds/go-sdk/catcher"
	mail_connectors "github.com/utmstack/utmstack/backend/internal/mail/connectors"
	appconfig_connectors "github.com/utmstack/utmstack/backend/modules/appconfig/connectors"
	"github.com/utmstack/utmstack/backend/modules/incidents/connectors"
	"github.com/utmstack/utmstack/backend/modules/incidents/domain"
)

const (
	ConfigKeyNotificationTo = "utmstack.incidents.notification_to"
	ConfigKeyNotificationCc = "utmstack.incidents.notification_cc"
)

// incidentMailer sends incident-created notifications to whoever the two config
// keys name. Nobody named means nobody notified: emptying the lists is how an
// admin turns these off, and it used to do the opposite — fall back to mailing
// every active user in the tenant.
type incidentMailer struct {
	mail  mail_connectors.MailService
	store appconfig_connectors.Store
}

// NewIncidentMailer wires the real mailer used at composition time.
func NewIncidentMailer(
	mail mail_connectors.MailService,
	store appconfig_connectors.Store,
) connectors.IncidentMailer {
	return &incidentMailer{mail: mail, store: store}
}

func (m *incidentMailer) SendIncidentCreated(ctx context.Context, incident domain.Incident) error {
	if m.mail == nil {
		catcher.Warn("incidents: mail service not configured — skipping notification", nil)
		return nil
	}
	to, cc, err := m.resolveRecipients(ctx)
	if err != nil {
		return fmt.Errorf("resolve incident notification recipients: %w", err)
	}
	if len(to) == 0 {
		// Not an error: no To configured is how notifications are switched off.
		// A Cc with no To is not a recipient list either — SMTP needs a To.
		return nil
	}
	subject, body := renderIncidentCreated(incident)
	return m.mail.SendMail(ctx, to, cc, subject, body, nil)
}

// resolveRecipients reads the two configured lists. There is no fallback: an
// address gets incident mail because somebody put it here.
func (m *incidentMailer) resolveRecipients(ctx context.Context) (to []string, cc []string, err error) {
	return readList(ctx, m.store, ConfigKeyNotificationTo),
		readList(ctx, m.store, ConfigKeyNotificationCc),
		nil
}

// readList reads a config key and splits it on commas. Missing/empty → nil.
func readList(ctx context.Context, store appconfig_connectors.Store, key string) []string {
	if store == nil {
		return nil
	}
	v, ok, err := store.GetString(ctx, key)
	if err != nil || !ok || strings.TrimSpace(v) == "" {
		return nil
	}
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

// renderIncidentCreated builds a subject + minimal HTML body for the notification.
func renderIncidentCreated(inc domain.Incident) (subject, body string) {
	subject = "New incident: " + inc.Name

	desc := ""
	if inc.Description != nil {
		desc = *inc.Description
	}
	// The severity is the word now, so the analyst reads "high" instead of the
	// "2" this used to print.
	sev := string(inc.Severity)
	if sev == "" {
		sev = "unknown"
	}
	body = fmt.Sprintf(
		`<html><body>`+
			`<h2>%s</h2>`+
			`<p><strong>Severity:</strong> %s</p>`+
			`<p><strong>Status:</strong> %s</p>`+
			`<p>%s</p>`+
			`</body></html>`,
		html.EscapeString(inc.Name),
		html.EscapeString(sev),
		html.EscapeString(string(inc.Status)),
		html.EscapeString(desc),
	)
	return subject, body
}
