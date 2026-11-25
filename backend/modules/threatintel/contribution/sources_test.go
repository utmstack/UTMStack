package contribution

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/utmstack/go-sdk/v2/store"
)

type fakeFetcher struct {
	raw   json.RawMessage
	err   error
	scope store.Scope
	id    string
}

func (f *fakeFetcher) FetchByID(_ context.Context, scope store.Scope, id string) (json.RawMessage, error) {
	f.scope, f.id = scope, id

	return f.raw, f.err
}

func TestAnAlertIsReadAcrossTenants(t *testing.T) {
	f := &fakeFetcher{raw: json.RawMessage(`{"id":"a-1","name":"x"}`)}

	alert, err := NewAlertSource(f).AlertByID(context.Background(), "a-1")
	if err != nil || alert.ID != "a-1" {
		t.Fatalf("%+v %v", alert, err)
	}

	if f.scope.Tenant != store.AllTenants || f.id != "a-1" {
		t.Errorf("scope %+v id %q", f.scope, f.id)
	}
}

func TestAnAlertThatCannotBeReadIsAnError(t *testing.T) {
	for name, f := range map[string]*fakeFetcher{
		"fetch":  {err: errors.New("gone")},
		"decode": {raw: json.RawMessage(`not json`)},
		"no id":  {raw: json.RawMessage(`{"name":"x"}`)},
	} {
		if _, err := NewAlertSource(f).AlertByID(context.Background(), "a-1"); err == nil {
			t.Errorf("%s: expected an error", name)
		}
	}
}

func TestAnInvalidIncidentIdIsRejectedBeforeAnyRead(t *testing.T) {
	if _, err := NewIncidentSource(nil, nil).IncidentAlerts(context.Background(), "nope"); err == nil {
		t.Error("expected an error")
	}
}
