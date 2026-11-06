package incidents

import (
	"context"
	"reflect"
	"testing"

	mail_connectors "github.com/utmstack/utmstack/backend/internal/mail/connectors"
	mail_domain "github.com/utmstack/utmstack/backend/internal/mail/domain"
	appconfig_connectors "github.com/utmstack/utmstack/backend/modules/appconfig/connectors"
	"github.com/utmstack/utmstack/backend/modules/incidents/domain"
)

// stubStore is a tiny in-memory Store for testing the resolver.
type stubStore struct{ values map[string]string }

func (s *stubStore) GetString(_ context.Context, key string) (string, bool, error) {
	v, ok := s.values[key]
	return v, ok, nil
}

func (s *stubStore) GetOwnString(_ context.Context, key string) (string, bool, error) {
	v, ok := s.values[key]
	return v, ok, nil
}

func (s *stubStore) SetString(_ context.Context, _, _ string, _ appconfig_connectors.SetOpts) error {
	return nil
}

func TestResolveRecipients(t *testing.T) {
	cases := []struct {
		name   string
		cfg    map[string]string
		wantTo []string
		wantCc []string
	}{
		{
			name:   "both lists are read as given",
			cfg:    map[string]string{ConfigKeyNotificationTo: "to1@x.com, to2@x.com", ConfigKeyNotificationCc: "cc1@x.com"},
			wantTo: []string{"to1@x.com", "to2@x.com"},
			wantCc: []string{"cc1@x.com"},
		},
		{
			name:   "blank entries and stray spaces are dropped",
			cfg:    map[string]string{ConfigKeyNotificationTo: " a@x.com , , b@x.com ,"},
			wantTo: []string{"a@x.com", "b@x.com"},
			wantCc: nil,
		},
		// Emptying the lists is how an admin turns incident mail off. It used to
		// fall back to every active user of the tenant, so the one action that
		// looks like "stop emailing" sent the mail to more people.
		{
			name:   "no addresses configured means nobody is notified",
			cfg:    map[string]string{},
			wantTo: nil,
			wantCc: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := &incidentMailer{store: &stubStore{values: tc.cfg}}
			to, cc, err := m.resolveRecipients(context.Background())
			if err != nil {
				t.Fatalf("unexpected err: %v", err)
			}
			if !slicesEqual(to, tc.wantTo) {
				t.Errorf("to: got %v want %v", to, tc.wantTo)
			}
			if !slicesEqual(cc, tc.wantCc) {
				t.Errorf("cc: got %v want %v", cc, tc.wantCc)
			}
		})
	}
}

// inheritingStubStore models the real appconfig Store's fallback behavior:
// GetString sees the tenant's own value or, failing that, the platform
// tenant's; GetOwnString only ever sees the tenant's own.
type inheritingStubStore struct {
	own, platform map[string]string
}

func (s *inheritingStubStore) GetString(_ context.Context, key string) (string, bool, error) {
	if v, ok := s.own[key]; ok {
		return v, true, nil
	}
	v, ok := s.platform[key]
	return v, ok, nil
}

func (s *inheritingStubStore) GetOwnString(_ context.Context, key string) (string, bool, error) {
	v, ok := s.own[key]
	return v, ok, nil
}

func (s *inheritingStubStore) SetString(_ context.Context, _, _ string, _ appconfig_connectors.SetOpts) error {
	return nil
}

// ACC-F33: a tenant that never configured its own notification recipients
// must not have its incident mail silently sent to whatever the platform
// tenant has configured — that would leak one tenant's incident details to
// addresses only the platform's own admin chose.
func TestResolveRecipientsDoesNotInheritFromPlatformTenant(t *testing.T) {
	store := &inheritingStubStore{
		own:      map[string]string{},
		platform: map[string]string{ConfigKeyNotificationTo: "platform-admin@other-org.com"},
	}
	m := &incidentMailer{store: store}

	to, cc, err := m.resolveRecipients(context.Background())
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if to != nil || cc != nil {
		t.Fatalf("resolveRecipients leaked the platform tenant's row: to=%v cc=%v", to, cc)
	}
}

// A Cc with no To is not a recipient list: SMTP has nobody to deliver to, and
// sending would mean inventing a To the admin never asked for.
func TestCcAloneSendsNothing(t *testing.T) {
	sent := false
	m := &incidentMailer{
		store: &stubStore{values: map[string]string{ConfigKeyNotificationCc: "cc@x.com"}},
		mail:  &stubMail{onSend: func() { sent = true }},
	}
	if err := m.SendIncidentCreated(context.Background(), sampleIncident()); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if sent {
		t.Error("sent mail with a Cc but no To")
	}
}

func slicesEqual(a, b []string) bool {
	if len(a) == 0 && len(b) == 0 {
		return true
	}
	return reflect.DeepEqual(a, b)
}

// stubMail records whether a send was attempted.
type stubMail struct {
	mail_connectors.MailService // embedded: any method this test does not stub panics
	onSend                      func()
}

func (s *stubMail) SendMail(_ context.Context, _ []string, _ []string, _, _ string, _ []mail_domain.Attatchment) error {
	s.onSend()
	return nil
}

func sampleIncident() domain.Incident {
	return domain.Incident{Name: "Something happened", Status: domain.StatusOpen, Severity: domain.SeverityHigh}
}
