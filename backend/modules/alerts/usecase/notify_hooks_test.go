package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/utmstack/utmstack/backend/modules/alerts/connectors"
	"github.com/utmstack/utmstack/backend/modules/alerts/domain"
	"github.com/utmstack/utmstack/backend/pkg/authz"
)

// store is a fake of the alert table that counts how it is read.
type store struct {
	connectors.AlertRepository
	alerts []domain.UtmAlert
	// missing is how many looks come back empty before the alert shows up, as with a row not yet readable.
	missing int
	looks   int
	windows []time.Duration
	broad   int // reads that were not bounded by time
}

func (s *store) SearchByIDs(context.Context, []string) ([]domain.UtmAlert, error) {
	s.broad++

	return s.alerts, nil
}

func (s *store) SearchRecentByIDs(_ context.Context, _ []string, within time.Duration) ([]domain.UtmAlert, error) {
	s.looks++
	s.windows = append(s.windows, within)

	if s.missing > 0 {
		s.missing--

		return nil, nil
	}

	return s.alerts, nil
}

type fakeMailer struct {
	enabled bool
	sent    int
	err     error
}

func (m *fakeMailer) Enabled(context.Context) bool { return m.enabled }

func (m *fakeMailer) SendAlertRaised(context.Context, domain.UtmAlert) error {
	m.sent++

	return m.err
}

func inTenant(t string) context.Context { return authz.WithTenantID(context.Background(), t) }

func usecaseFor(s *store, m connectors.AlertMailer) *alertUsecase {
	return &alertUsecase{repo: s, mailer: m}
}

func fast(t *testing.T) {
	t.Helper()

	old := notifyRetryDelays
	notifyRetryDelays = []time.Duration{time.Millisecond, time.Millisecond, time.Millisecond, time.Millisecond}

	t.Cleanup(func() { notifyRetryDelays = old })
}

// The table is keyed by name and time: finding an alert by id reads a large share of it. The hooks start
// from the id and the tenant, so nothing is read for them.
func TestHooksAreToldWithoutReadingTheAlertAtAll(t *testing.T) {
	s := &store{alerts: []domain.UtmAlert{{ID: "a1"}}}
	u := usecaseFor(s, nil)

	var got [][2]string

	u.AddRaisedHook(func(_ context.Context, tenant, id string) error { got = append(got, [2]string{tenant, id}); return nil })
	u.AddRaisedHook(func(_ context.Context, _, id string) error { got = append(got, [2]string{"second", id}); return nil })

	if err := u.NotifyRaised(inTenant("t1"), "a1"); err != nil {
		t.Fatal(err)
	}

	if len(got) != 2 || got[0] != [2]string{"t1", "a1"} || got[1][0] != "second" {
		t.Errorf("hooks saw %v", got)
	}

	if s.looks != 0 || s.broad != 0 {
		t.Errorf("the table was read (%d bounded, %d unbounded) for something that needed no alert", s.looks, s.broad)
	}
}

func TestNobodyToWriteToMeansTheAlertIsNeverRead(t *testing.T) {
	s := &store{alerts: []domain.UtmAlert{{ID: "a1"}}}
	m := &fakeMailer{enabled: false}

	if err := usecaseFor(s, m).NotifyRaised(inTenant("t1"), "a1"); err != nil {
		t.Fatal(err)
	}

	if s.looks != 0 || s.broad != 0 || m.sent != 0 {
		t.Errorf("read %d+%d times, sent %d", s.looks, s.broad, m.sent)
	}
}

func TestTheEmailReadsTheAlertBoundedByTimeAndSendsIt(t *testing.T) {
	s := &store{alerts: []domain.UtmAlert{{ID: "a1"}}}
	m := &fakeMailer{enabled: true}

	if err := usecaseFor(s, m).NotifyRaised(inTenant("t1"), "a1"); err != nil {
		t.Fatal(err)
	}

	if m.sent != 1 || s.looks != 1 || s.broad != 0 {
		t.Errorf("sent %d after %d bounded and %d unbounded reads", m.sent, s.looks, s.broad)
	}

	if s.windows[0] != notifyWindow {
		t.Errorf("window %v", s.windows[0])
	}
}

func TestAnEchoOfAGroupedAlertIsNotMailed(t *testing.T) {
	s := &store{alerts: []domain.UtmAlert{{ID: "a2", ParentID: "a1"}}}
	m := &fakeMailer{enabled: true}

	if err := usecaseFor(s, m).NotifyRaised(inTenant("t1"), "a2"); err != nil || m.sent != 0 {
		t.Errorf("mailed %d, err %v", m.sent, err)
	}
}

func TestAFailingMailerDoesNotKeepTheHooksFromRunning(t *testing.T) {
	mailErr, hookErr := errors.New("smtp is down"), errors.New("nats is down")
	m := &fakeMailer{enabled: true, err: mailErr}
	called := 0

	u := usecaseFor(&store{alerts: []domain.UtmAlert{{ID: "a1"}}}, m)
	u.AddRaisedHook(func(context.Context, string, string) error { called++; return hookErr })

	err := u.NotifyRaised(inTenant("t1"), "a1")

	if m.sent != 1 || called != 1 {
		t.Errorf("mailed %d, hooked %d", m.sent, called)
	}

	if !errors.Is(err, mailErr) || !errors.Is(err, hookErr) {
		t.Errorf("both failures must be reported: %v", err)
	}
}

func TestAnAlertNotReadableYetIsLookedForAgainBeforeGivingUp(t *testing.T) {
	fast(t)

	s := &store{alerts: []domain.UtmAlert{{ID: "a1"}}, missing: 3}
	m := &fakeMailer{enabled: true}

	if err := usecaseFor(s, m).NotifyRaised(inTenant("t1"), "a1"); err != nil {
		t.Fatalf("an alert that appears on the fourth look must be mailed: %v", err)
	}

	if m.sent != 1 || s.looks != 4 {
		t.Errorf("mailed %d after %d looks", m.sent, s.looks)
	}
}

func TestAnAlertThatNeverAppearsIsReportedAfterTheLastLook(t *testing.T) {
	fast(t)

	s := &store{missing: 99}
	m := &fakeMailer{enabled: true}

	if err := usecaseFor(s, m).NotifyRaised(inTenant("t1"), "a1"); !errors.Is(err, domain.ErrAlertNotFound) || s.looks != 5 || m.sent != 0 {
		t.Errorf("err %v after %d looks, mailed %d", err, s.looks, m.sent)
	}
}

func TestTheLookingStopsWhenTheCallerGivesUp(t *testing.T) {
	old := notifyRetryDelays
	notifyRetryDelays = []time.Duration{time.Hour}

	defer func() { notifyRetryDelays = old }()

	ctx, cancel := context.WithTimeout(inTenant("t1"), 30*time.Millisecond)
	defer cancel()

	if err := usecaseFor(&store{missing: 99}, &fakeMailer{enabled: true}).NotifyRaised(ctx, "a1"); !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("err %v", err)
	}
}

func TestWithNothingToTellNothingHappens(t *testing.T) {
	s := &store{}

	if err := (&alertUsecase{repo: s}).NotifyRaised(inTenant("t1"), "a1"); err != nil || s.looks != 0 {
		t.Errorf("err %v, looks %d", err, s.looks)
	}

	if err := usecaseFor(s, &fakeMailer{enabled: true}).NotifyRaised(inTenant("t1"), ""); err != nil || s.looks != 0 {
		t.Errorf("an empty id: err %v, looks %d", err, s.looks)
	}
}
