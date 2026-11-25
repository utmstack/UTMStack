package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/utmstack/utmstack/backend/modules/alerts/connectors"
	"github.com/utmstack/utmstack/backend/modules/alerts/domain"
)

type findsAlerts struct {
	connectors.AlertRepository
	alerts []domain.UtmAlert
}

func (f findsAlerts) SearchByIDs(context.Context, []string) ([]domain.UtmAlert, error) {
	return f.alerts, nil
}

type countingMailer struct {
	sent int
	err  error
}

func (m *countingMailer) SendAlertRaised(context.Context, domain.UtmAlert) error {
	m.sent++

	return m.err
}

func newUsecase(alert domain.UtmAlert, mailer connectors.AlertMailer) *alertUsecase {
	return &alertUsecase{repo: findsAlerts{alerts: []domain.UtmAlert{alert}}, mailer: mailer}
}

func TestEveryHookIsToldAboutANewAlertWithItsTenant(t *testing.T) {
	var got [][2]string

	u := newUsecase(domain.UtmAlert{ID: "a1", TenantID: "t1"}, nil)
	u.AddRaisedHook(func(_ context.Context, tenant, id string) error { got = append(got, [2]string{tenant, id}); return nil })
	u.AddRaisedHook(func(_ context.Context, tenant, id string) error {
		got = append(got, [2]string{"second", id})
		return nil
	})

	if err := u.NotifyRaised(context.Background(), "a1"); err != nil {
		t.Fatal(err)
	}

	if len(got) != 2 || got[0] != [2]string{"t1", "a1"} || got[1][0] != "second" {
		t.Errorf("hooks saw %v", got)
	}
}

func TestAnEchoOfAGroupedAlertTellsNobody(t *testing.T) {
	m := &countingMailer{}
	called := 0

	u := newUsecase(domain.UtmAlert{ID: "a2", ParentID: "a1"}, m)
	u.AddRaisedHook(func(context.Context, string, string) error { called++; return nil })

	if err := u.NotifyRaised(context.Background(), "a2"); err != nil || m.sent != 0 || called != 0 {
		t.Errorf("mailed %d, hooked %d, err %v", m.sent, called, err)
	}
}

func TestAFailingMailerDoesNotKeepTheHooksFromRunningNorTheOtherWayRound(t *testing.T) {
	mailErr, hookErr := errors.New("smtp is down"), errors.New("nats is down")
	m := &countingMailer{err: mailErr}
	called := 0

	u := newUsecase(domain.UtmAlert{ID: "a1", TenantID: "t1"}, m)
	u.AddRaisedHook(func(context.Context, string, string) error { called++; return hookErr })

	err := u.NotifyRaised(context.Background(), "a1")

	if m.sent != 1 || called != 1 {
		t.Errorf("mailed %d, hooked %d", m.sent, called)
	}

	if !errors.Is(err, mailErr) || !errors.Is(err, hookErr) {
		t.Errorf("both failures must be reported: %v", err)
	}
}

func TestWithNothingToTellNothingIsRead(t *testing.T) {
	u := &alertUsecase{repo: findsAlerts{}}

	if err := u.NotifyRaised(context.Background(), "a1"); err != nil {
		t.Errorf("no mailer and no hooks is not an error: %v", err)
	}
}
