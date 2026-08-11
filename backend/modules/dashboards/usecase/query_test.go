package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/threatwinds/go-sdk/store"

	"github.com/utmstack/utmstack/backend/modules/dashboards/domain"
	"github.com/utmstack/utmstack/backend/pkg/authz"
)

const customerTenant = "8f1c1b8e-0000-4000-8000-000000000001"

type spyStore struct {
	called string
	scope  store.Scope
}

func (s *spyStore) Count(_ context.Context, sc store.Scope, _ []store.Filter) (int64, error) {
	s.called, s.scope = "Count", sc
	return 7, nil
}

func (s *spyStore) TopValues(_ context.Context, sc store.Scope, _ string, _ []store.Filter, _ int) ([]store.Bucket, error) {
	s.called, s.scope = "TopValues", sc
	return []store.Bucket{{Key: "a", Count: 1}}, nil
}

func (s *spyStore) Timeline(_ context.Context, sc store.Scope, _ []store.Filter, _ store.Interval) ([]store.Point, error) {
	s.called, s.scope = "Timeline", sc
	return []store.Point{{Count: 1}}, nil
}

func (s *spyStore) TimelineByField(_ context.Context, sc store.Scope, _ string, _ []store.Filter, _ store.Interval, _ int) ([]store.Series, error) {
	s.called, s.scope = "TimelineByField", sc
	return []store.Series{{Key: "a"}}, nil
}

func (s *spyStore) FetchPage(_ context.Context, sc store.Scope, _ []store.Filter, _ store.Page) ([]json.RawMessage, int64, error) {
	s.called, s.scope = "FetchPage", sc
	return []json.RawMessage{[]byte(`{}`)}, 1, nil
}

func run(t *testing.T, spec domain.Spec) (*spyStore, *Result, error) {
	t.Helper()
	st := &spyStore{}
	ctx := authz.WithTenantID(context.Background(), customerTenant)
	res, err := NewQueryService(st).Run(ctx, spec)
	return st, res, err
}

// Each chart is answered by the aggregation that fits it, so a widget never
// pulls rows it does not need to count them.
func TestEachChartUsesItsAggregation(t *testing.T) {
	cases := []struct {
		name string
		spec domain.Spec
		want string
	}{
		{"single number", domain.Spec{Dataset: "alerts", Chart: domain.ChartMetric}, "Count"},
		{"broken down", domain.Spec{Dataset: "alerts", Chart: domain.ChartCategory, Dimension: "name"}, "TopValues"},
		{"over time", domain.Spec{Dataset: "logs", Chart: domain.ChartTime, Interval: "1h"}, "Timeline"},
		{"over time, split", domain.Spec{Dataset: "logs", Chart: domain.ChartTime, Interval: "1h", Dimension: "dataType"}, "TimelineByField"},
		{"the rows", domain.Spec{Dataset: "logs", Chart: domain.ChartTable}, "FetchPage"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			st, _, err := run(t, c.spec)
			if err != nil {
				t.Fatalf("Run: %v", err)
			}
			if st.called != c.want {
				t.Errorf("used %s, want %s", st.called, c.want)
			}
		})
	}
}

// The tenant comes from the session. A stored visualization is the same
// question for everyone, and whose data answers it is not part of the question.
func TestTheScopeCarriesTheSessionsTenant(t *testing.T) {
	st, _, err := run(t, domain.Spec{Dataset: "alerts", Chart: domain.ChartMetric})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if st.scope.Tenant != customerTenant {
		t.Errorf("tenant = %q, want the session's", st.scope.Tenant)
	}
}

// Naming a dataset the schema does not have would otherwise reach the driver,
// and from there any table the connection can see — the server's own included.
func TestOnlyTheKnownDatasetsAreReachable(t *testing.T) {
	for _, ds := range []string{"system.query_log", "users", "", "Logs"} {
		_, _, err := run(t, domain.Spec{Dataset: ds, Chart: domain.ChartMetric})
		if err == nil {
			t.Errorf("dataset %q was accepted", ds)
		}
	}
}

func TestASpecThatMakesNoSenseIsRefused(t *testing.T) {
	cases := map[string]domain.Spec{
		"a breakdown with nothing to break down by": {Dataset: "logs", Chart: domain.ChartCategory},
		"a chart that does not exist":               {Dataset: "logs", Chart: "sankey"},
		"an aggregation that does not exist":        {Dataset: "logs", Chart: domain.ChartMetric, Metric: domain.Metric{Agg: "median"}},
		// The store counts records and nothing else. Answering an average with
		// a count would look perfectly reasonable on the screen.
		"an average, which the store cannot answer": {Dataset: "logs", Chart: domain.ChartMetric, Metric: domain.Metric{Agg: "avg", Field: "bytes"}},
	}

	for name, spec := range cases {
		t.Run(name, func(t *testing.T) {
			if _, _, err := run(t, spec); err == nil {
				t.Error("it was accepted")
			}
		})
	}
}

func TestAnUnknownFilterOperatorIsRefused(t *testing.T) {
	_, _, err := run(t, domain.Spec{
		Dataset: "logs", Chart: domain.ChartMetric,
		Filters: []domain.Filter{{Field: "x", Op: "regex_match", Value: ".*"}},
	})
	if !errors.Is(err, domain.ErrUnknownOp) {
		t.Errorf("err = %v, want ErrUnknownOp", err)
	}
}

// The answer crosses the wire to a browser, so its field names are a contract.
// The store's own types carry no JSON tags: returning them directly would put
// Go field names ("Key", "Count", "At") on the wire and the widget would draw
// nothing.
func TestTheAnswerIsSpelledTheWayTheClientReadsIt(t *testing.T) {
	total := int64(7)
	at := time.Date(2026, 8, 11, 0, 0, 0, 0, time.UTC)
	res := Result{
		Total:   &total,
		Buckets: []Bucket{{Key: "syslog", Count: 3}},
		Points:  []Point{{At: at, Count: 4}},
		Series:  []Series{{Key: "windows", Points: []Point{{At: at, Count: 5}}}},
	}

	b, err := json.Marshal(res)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	got := string(b)
	for _, want := range []string{
		`"total":7`,
		`"buckets":[{"key":"syslog","count":3}]`,
		`"points":[{"at":"2026-08-11T00:00:00Z","count":4}]`,
		`"series":[{"key":"windows","points":[{"at":"2026-08-11T00:00:00Z","count":5}]}]`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %s in %s", want, got)
		}
	}
}
