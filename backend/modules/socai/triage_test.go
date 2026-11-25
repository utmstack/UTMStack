package socai

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/utmstack/go-sdk/v2/store"

	"github.com/utmstack/utmstack/backend/modules/socai/dto"
)

type fakeConfig struct {
	cfg *dto.ConfigResponse
	err error
}

func (f fakeConfig) Get(context.Context) (*dto.ConfigResponse, error) { return f.cfg, f.err }

type fakeDocs struct {
	alert    string
	related  map[string][]string // by filter field
	count    int64
	failByID error
	counted  []store.Filter
}

func (f *fakeDocs) FetchByID(_ context.Context, _ store.Scope, _ string) (json.RawMessage, error) {
	if f.failByID != nil {
		return nil, f.failByID
	}

	return json.RawMessage(f.alert), nil
}

func (f *fakeDocs) FetchN(_ context.Context, _ store.Scope, filters []store.Filter, _ int) ([]json.RawMessage, error) {
	var out []json.RawMessage
	for _, d := range f.related[filters[0].Field] {
		out = append(out, json.RawMessage(d))
	}

	return out, nil
}

func (f *fakeDocs) Count(_ context.Context, _ store.Scope, filters []store.Filter) (int64, error) {
	f.counted = filters

	return f.count, nil
}

type published struct {
	subject string
	data    []byte
}

type fakePub struct {
	got []published
	err error
}

func (p *fakePub) Publish(_ context.Context, subject string, data []byte) error {
	p.got = append(p.got, published{subject, data})

	return p.err
}

const bigAlert = `{"id":"a1","name":"Brute force","severity":"high","status":"Open","tenantId":"t1","notes":"",
 "adversary":{"ip":"9.9.9.9","user":"mallory","geolocation":{"country":"XX"}},
 "target":{"ip":"10.0.0.5","host":"srv"},
 "lastEvent":{"huge":"%s"},"events":[{"x":1},{"x":2}],"history":[{"user":"u"}]}`

func newTriage(cfg *dto.ConfigResponse, docs *fakeDocs, pub *fakePub, score ScoreFunc) *Triage {
	if score == nil {
		score = func(context.Context, string) (any, error) { return map[string]any{"score": 81}, nil }
	}

	return NewTriage(fakeConfig{cfg: cfg}, nil, score, docs, pub)
}

func on() *dto.ConfigResponse { return &dto.ConfigResponse{Configured: true, AutoAnalyze: true} }

func docs() *fakeDocs {
	return &fakeDocs{
		alert: strings.Replace(bigAlert, "%s", strings.Repeat("x", 5000), 1),
		related: map[string][]string{
			"adversary.ip": {`{"id":"a1","name":"itself"}`, `{"id":"a2","name":"Scan","adversary":{"ip":"9.9.9.9"},"lastEvent":{"x":1}}`},
			"target.ip":    {`{"id":"a3","name":"Probe","target":{"ip":"10.0.0.5"}}`},
		},
		count: 7,
	}
}

func decode(t *testing.T, p published) Message {
	t.Helper()

	var m Message
	if err := json.Unmarshal(p.data, &m); err != nil {
		t.Fatal(err)
	}

	return m
}

func TestANewAlertIsPublishedWithItsDossierAlreadyWorkedOut(t *testing.T) {
	pub := &fakePub{}
	d := docs()

	if err := newTriage(on(), d, pub, nil).AlertRaised(context.Background(), "t1", "a1"); err != nil {
		t.Fatal(err)
	}

	if len(pub.got) != 1 || pub.got[0].subject != AlertsSubject {
		t.Fatalf("published %+v", pub.got)
	}

	m := decode(t, pub.got[0])
	if m.TenantID != "t1" || m.AlertID != "a1" {
		t.Errorf("message %+v", m)
	}

	alert := m.Facts["alert"].(map[string]any)
	if alert["name"] != "Brute force" || alert["adversary"].(map[string]any)["ip"] != "9.9.9.9" {
		t.Errorf("alert %+v", alert)
	}

	for _, heavy := range []string{"lastEvent", "events", "history"} {
		if _, ok := alert[heavy]; ok {
			t.Errorf("%s must not travel in the message", heavy)
		}
	}

	if len(pub.got[0].data) > 4000 {
		t.Errorf("a %d byte message for an alert whose raw document had five thousand bytes of event", len(pub.got[0].data))
	}

	if m.Facts["score"].(map[string]any)["score"] != float64(81) {
		t.Errorf("score %+v", m.Facts["score"])
	}

	same := m.Facts["same_adversary"].([]any)
	if len(same) != 1 || same[0].(map[string]any)["id"] != "a2" {
		t.Errorf("the alert itself must not be its own neighbour: %+v", same)
	}

	if m.Facts["same_name_30d"] != float64(7) || len(m.Facts["same_target"].([]any)) != 1 {
		t.Errorf("facts %+v", m.Facts)
	}

	if d.counted[0].Field != "name" || d.counted[0].Value != "Brute force" {
		t.Errorf("the recurrence is counted by name: %+v", d.counted)
	}
}

func TestNothingIsSentWhenTheTenantDidNotAskForIt(t *testing.T) {
	for name, cfg := range map[string]*dto.ConfigResponse{
		"not configured":          {Configured: false, AutoAnalyze: true},
		"automatic is off":        {Configured: true, AutoAnalyze: false},
		"no configuration at all": {},
	} {
		pub := &fakePub{}

		if err := newTriage(cfg, docs(), pub, nil).AlertRaised(context.Background(), "t1", "a1"); err != nil || len(pub.got) != 0 {
			t.Errorf("%s: published %d, err %v", name, len(pub.got), err)
		}
	}
}

func TestATenantOverItsQuotaIsNotAnalysed(t *testing.T) {
	pub := &fakePub{}
	tr := newTriage(on(), docs(), pub, nil)
	tr.quota = &AIQuota{
		LimitOf: func(context.Context, string) (int, error) { return 5, nil },
		Consume: func(context.Context, string) (int64, error) { return 6, nil },
	}

	if err := tr.AlertRaised(context.Background(), "t1", "a1"); err != nil || len(pub.got) != 0 {
		t.Errorf("published %d, err %v", len(pub.got), err)
	}

	tr.quota.Consume = func(context.Context, string) (int64, error) { return 5, nil }

	if err := tr.AlertRaised(context.Background(), "t1", "a1"); err != nil || len(pub.got) != 1 {
		t.Errorf("at the limit it still fits: published %d, err %v", len(pub.got), err)
	}
}

func TestWithoutADossierTheAlertIsStillQueuedForTheServiceToFetch(t *testing.T) {
	pub := &fakePub{}
	d := docs()
	d.failByID = errors.New("clickhouse is down")

	if err := newTriage(on(), d, pub, nil).AlertRaised(context.Background(), "t1", "a1"); err != nil {
		t.Fatal(err)
	}

	m := decode(t, pub.got[0])
	if m.AlertID != "a1" || m.Facts != nil {
		t.Errorf("message %+v", m)
	}
}

func TestAFailedScoreOrRelatedLookupDoesNotLoseTheRest(t *testing.T) {
	pub := &fakePub{}
	score := func(context.Context, string) (any, error) { return nil, errors.New("scoring is down") }

	if err := newTriage(on(), docs(), pub, score).AlertRaised(context.Background(), "t1", "a1"); err != nil {
		t.Fatal(err)
	}

	m := decode(t, pub.got[0])
	if _, ok := m.Facts["score"]; ok || m.Facts["alert"] == nil || m.Facts["same_name_30d"] == nil {
		t.Errorf("facts %+v", m.Facts)
	}
}

func TestAnAlertThatCannotBeQueuedIsReported(t *testing.T) {
	pub := &fakePub{err: errors.New("nats is down")}

	if err := newTriage(on(), docs(), pub, nil).AlertRaised(context.Background(), "t1", "a1"); err == nil {
		t.Error("an alert due for triage that was not queued must be reported")
	}
}

func TestAHugeDossierIsDroppedBeforeTheMessageExceedsWhatNATSTakes(t *testing.T) {
	pub := &fakePub{}
	d := docs()
	d.alert = `{"id":"a1","name":"` + strings.Repeat("n", maxMessage+10) + `","tenantId":"t1"}`

	if err := newTriage(on(), d, pub, nil).AlertRaised(context.Background(), "t1", "a1"); err != nil {
		t.Fatal(err)
	}

	if m := decode(t, pub.got[0]); m.Facts != nil || len(pub.got[0].data) > 1000 {
		t.Errorf("facts kept: %v, %d bytes", m.Facts != nil, len(pub.got[0].data))
	}
}

func TestQuotaSemantics(t *testing.T) {
	limit := func(n int) *AIQuota {
		return &AIQuota{
			LimitOf: func(context.Context, string) (int, error) { return n, nil },
			Consume: func(context.Context, string) (int64, error) { return 3, nil },
		}
	}

	cases := map[string]struct {
		q    *AIQuota
		want bool
	}{
		"no quota configured":  {nil, true},
		"uncapped":             {limit(-1), true},
		"an allowance of none": {limit(0), false},
		"over":                 {limit(2), false},
		"within":               {limit(3), true},
		"limit unreadable": {&AIQuota{
			LimitOf: func(context.Context, string) (int, error) { return 0, errors.New("x") },
			Consume: func(context.Context, string) (int64, error) { return 99, nil },
		}, true},
	}

	for name, c := range cases {
		if got := c.q.Check(context.Background(), "t1").Allowed; got != c.want {
			t.Errorf("%s: allowed=%v, want %v", name, got, c.want)
		}
	}
}
