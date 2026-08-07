package alerts

import (
	"context"
	"fmt"
	"html"
	"strings"

	mail_connectors "github.com/utmstack/utmstack/backend/internal/mail/connectors"
	"github.com/utmstack/utmstack/backend/modules/alerts/connectors"
	"github.com/utmstack/utmstack/backend/modules/alerts/domain"
	appconfig_connectors "github.com/utmstack/utmstack/backend/modules/appconfig/connectors"
)

const (
	ConfigKeyNotificationTo = "utmstack.alerts.notification_to"
	ConfigKeyNotificationCc = "utmstack.alerts.notification_cc"
)

const (
	ConfigKeyBaseURL      = "utmstack.mail.baseUrl"
	ConfigKeyOrganization = "utmstack.mail.organization"
)

type alertMailer struct {
	mail  mail_connectors.MailService
	store appconfig_connectors.Store
}

func NewAlertMailer(mail mail_connectors.MailService, store appconfig_connectors.Store) connectors.AlertMailer {
	return &alertMailer{mail: mail, store: store}
}

func (m *alertMailer) SendAlertRaised(ctx context.Context, alert domain.UtmAlert) error {
	if m.mail == nil {
		return nil
	}
	to := readList(ctx, m.store, ConfigKeyNotificationTo)
	if len(to) == 0 {
		return nil
	}
	cc := readList(ctx, m.store, ConfigKeyNotificationCc)

	subject, body := m.render(ctx, alert)
	return m.mail.SendMail(ctx, to, cc, subject, body, nil)
}

func (m *alertMailer) render(ctx context.Context, a domain.UtmAlert) (subject, body string) {
	org := readString(ctx, m.store, ConfigKeyOrganization)
	subject = fmt.Sprintf("[%s] %s", shortID(a.ID), a.Name)
	if org != "" {
		subject = org + " " + subject
	}

	rows := [][2]string{
		{"Severity", string(a.Severity)},
		{"Category", a.Category},
		{"Technique", a.Technique},
		{"Data source", a.DataSource},
		{"Data type", a.DataType},
		{"When", a.Timestamp},
	}
	if a.Adversary != nil && a.Adversary.Host != "" {
		rows = append(rows, [2]string{"Adversary", a.Adversary.Host})
	}
	if a.Target != nil && a.Target.Host != "" {
		rows = append(rows, [2]string{"Target", a.Target.Host})
	}

	var b strings.Builder
	b.WriteString("<html><body>")
	fmt.Fprintf(&b, "<h2>%s</h2>", html.EscapeString(a.Name))
	if a.Description != "" {
		fmt.Fprintf(&b, "<p>%s</p>", html.EscapeString(a.Description))
	}
	b.WriteString("<table cellpadding=\"4\">")
	for _, r := range rows {
		if r[1] == "" {
			continue
		}
		fmt.Fprintf(&b, "<tr><td><strong>%s</strong></td><td>%s</td></tr>",
			html.EscapeString(r[0]), html.EscapeString(r[1]))
	}
	b.WriteString("</table>")

	if base := strings.TrimRight(readString(ctx, m.store, ConfigKeyBaseURL), "/"); base != "" {
		link := fmt.Sprintf("%s/threat-management/alerts?alertId=%s", base, a.ID)
		fmt.Fprintf(&b, "<p><a href=\"%s\">Open in UTMStack</a></p>", html.EscapeString(link))
	}
	b.WriteString("</body></html>")

	return subject, b.String()
}

func shortID(id string) string {
	if len(id) > 8 {
		return id[:8]
	}
	return id
}

func readString(ctx context.Context, store appconfig_connectors.Store, key string) string {
	if store == nil {
		return ""
	}
	v, ok, err := store.GetString(ctx, key)
	if err != nil || !ok {
		return ""
	}
	return strings.TrimSpace(v)
}

func readList(ctx context.Context, store appconfig_connectors.Store, key string) []string {
	v := readString(ctx, store, key)
	if v == "" {
		return nil
	}
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}
