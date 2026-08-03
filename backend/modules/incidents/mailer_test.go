package incidents

import (
	"context"
	"reflect"
	"testing"

	appconfig_connectors "github.com/utmstack/utmstack/backend/modules/appconfig/connectors"
	iam_connectors "github.com/utmstack/utmstack/backend/modules/iam/connectors"
	iam_domain "github.com/utmstack/utmstack/backend/modules/iam/domain"
)

// stubStore is a tiny in-memory Store for testing the resolver's fallback path.
type stubStore struct{ values map[string]string }

func (s *stubStore) GetString(_ context.Context, key string) (string, bool, error) {
	v, ok := s.values[key]
	return v, ok, nil
}

func (s *stubStore) SetString(_ context.Context, _, _ string, _ appconfig_connectors.SetOpts) error {
	return nil
}

// stubUserRepo returns a canned user list. Only List is exercised.
type stubUserRepo struct {
	iam_connectors.UserRepository // embed so any unused method panics if called
	users                         []iam_domain.User
}

func (s *stubUserRepo) List(_ context.Context, _ iam_connectors.ListUsersFilter) ([]iam_domain.User, int64, error) {
	return s.users, int64(len(s.users)), nil
}

func TestResolveRecipients(t *testing.T) {
	users := []iam_domain.User{
		{Email: "a@x.com", Activated: true},
		{Email: "b@x.com", Activated: false}, // deactivated → dropped
		{Email: "", Activated: true},         // no email → dropped
		{Email: "c@x.com", Activated: true},
	}

	cases := []struct {
		name    string
		cfg     map[string]string
		users   []iam_domain.User
		wantTo  []string
		wantCc  []string
	}{
		{
			name:   "explicit to and cc override user fallback",
			cfg:    map[string]string{ConfigKeyNotificationTo: "to1@x.com, to2@x.com", ConfigKeyNotificationCc: "cc1@x.com"},
			users:  users,
			wantTo: []string{"to1@x.com", "to2@x.com"},
			wantCc: []string{"cc1@x.com"},
		},
		{
			name:   "only cc set — no fallback",
			cfg:    map[string]string{ConfigKeyNotificationCc: "cc1@x.com"},
			users:  users,
			wantTo: nil,
			wantCc: []string{"cc1@x.com"},
		},
		{
			name:   "both empty → all activated users with non-empty email",
			cfg:    map[string]string{},
			users:  users,
			wantTo: []string{"a@x.com", "c@x.com"},
			wantCc: nil,
		},
		{
			name:   "both empty and no users → empty",
			cfg:    map[string]string{},
			users:  nil,
			wantTo: []string{},
			wantCc: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := &incidentMailer{
				store:    &stubStore{values: tc.cfg},
				userRepo: &stubUserRepo{users: tc.users},
			}
			to, cc, err := m.resolveRecipients(context.Background())
			if err != nil {
				t.Fatalf("unexpected err: %v", err)
			}
			// Compare with normalised nil/empty: treat empty slice == nil for equality.
			if !slicesEqual(to, tc.wantTo) {
				t.Errorf("to: got %v want %v", to, tc.wantTo)
			}
			if !slicesEqual(cc, tc.wantCc) {
				t.Errorf("cc: got %v want %v", cc, tc.wantCc)
			}
		})
	}
}

func slicesEqual(a, b []string) bool {
	if len(a) == 0 && len(b) == 0 {
		return true
	}
	return reflect.DeepEqual(a, b)
}
