package incidents

import (
	"context"
	"fmt"
	"html"
	"strings"

	"github.com/threatwinds/go-sdk/catcher"
	mail_connectors "github.com/utmstack/utmstack/backend/internal/mail/connectors"
	appconfig_connectors "github.com/utmstack/utmstack/backend/modules/appconfig/connectors"
	iam_connectors "github.com/utmstack/utmstack/backend/modules/iam/connectors"
	"github.com/utmstack/utmstack/backend/modules/incidents/connectors"
	"github.com/utmstack/utmstack/backend/modules/incidents/domain"
)

// Config keys for the global alert/incident notification recipients. Both are
// comma-separated lists of email addresses; both are optional. See migration
// 000003_alert_notification_recipients.
const (
	ConfigKeyNotificationTo = "utmstack.alerts.notification_to"
	ConfigKeyNotificationCc = "utmstack.alerts.notification_cc"
)

// incidentMailer sends incident-created notifications. Recipients come from the
// two config keys above; when both are empty it falls back to every activated
// user.
type incidentMailer struct {
	mail     mail_connectors.MailService
	store    appconfig_connectors.Store
	userRepo iam_connectors.UserRepository
}

// NewIncidentMailer wires the real mailer used at composition time.
func NewIncidentMailer(
	mail mail_connectors.MailService,
	store appconfig_connectors.Store,
	userRepo iam_connectors.UserRepository,
) connectors.IncidentMailer {
	return &incidentMailer{mail: mail, store: store, userRepo: userRepo}
}

func (m *incidentMailer) SendIncidentCreated(ctx context.Context, incident domain.UtmIncident) error {
	if m.mail == nil {
		catcher.Warn("incidents: mail service not configured — skipping notification", nil)
		return nil
	}
	to, cc, err := m.resolveRecipients(ctx)
	if err != nil {
		return fmt.Errorf("resolve incident notification recipients: %w", err)
	}
	if len(to) == 0 {
		catcher.Warn("incidents: no recipients available — skipping notification", nil)
		return nil
	}
	subject, body := renderIncidentCreated(incident)
	return m.mail.SendMail(ctx, to, cc, subject, body, nil)
}

// resolveRecipients returns (to, cc) resolved from config; when BOTH config
// keys are empty it falls back to every activated user's email as `to` and no
// cc.
func (m *incidentMailer) resolveRecipients(ctx context.Context) ([]string, []string, error) {
	to := readList(ctx, m.store, ConfigKeyNotificationTo)
	cc := readList(ctx, m.store, ConfigKeyNotificationCc)
	if len(to) > 0 || len(cc) > 0 {
		return to, cc, nil
	}
	if m.userRepo == nil {
		return nil, nil, nil
	}
	// ponytail: single page, PageSize=200 (repo cap). Enough for every
	// production tenant we've seen; paginate if a deployment exceeds it.
	users, _, err := m.userRepo.List(ctx, iam_connectors.ListUsersFilter{PageSize: 200})
	if err != nil {
		return nil, nil, err
	}
	fallback := make([]string, 0, len(users))
	for _, u := range users {
		if !u.Activated || u.Email == "" {
			continue
		}
		fallback = append(fallback, u.Email)
	}
	return fallback, nil, nil
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
func renderIncidentCreated(inc domain.UtmIncident) (subject, body string) {
	subject = "New incident: " + inc.IncidentName

	desc := ""
	if inc.IncidentDescription != nil {
		desc = *inc.IncidentDescription
	}
	sev := "unknown"
	if inc.IncidentSeverity != nil {
		sev = fmt.Sprintf("%d", *inc.IncidentSeverity)
	}
	body = fmt.Sprintf(
		`<html><body>`+
			`<h2>%s</h2>`+
			`<p><strong>Severity:</strong> %s</p>`+
			`<p><strong>Status:</strong> %s</p>`+
			`<p>%s</p>`+
			`</body></html>`,
		html.EscapeString(inc.IncidentName),
		html.EscapeString(sev),
		html.EscapeString(inc.IncidentStatus),
		html.EscapeString(desc),
	)
	return subject, body
}
