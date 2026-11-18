package netblock

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/netip"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/utmstack/UTMStack/agent/edr/cache"
	"github.com/utmstack/UTMStack/agent/edr/config"
)

func TestNetworkFeedReadersRejectRawExpandedAndRecordOverflow(t *testing.T) {
	if b, err := readNetworkFeed(strings.NewReader("1234"), 4); err != nil || string(b) != "1234" {
		t.Fatal(b, err)
	}
	if _, err := readNetworkFeed(strings.NewReader("12345"), 4); !errors.Is(err, ErrNetworkFeedBudget) {
		t.Fatal(err)
	}
	if inds, _, err := parseAccumulativeBounded(gz(strings.Repeat("# ignored\n", 10)), 1, "ip", 16, 2); !errors.Is(err, ErrNetworkFeedBudget) || inds != nil {
		t.Fatalf("expanded comments bypass %v %v", inds, err)
	}
	if inds, _, err := parseAccumulativeBounded(gz("192.0.2.1\n192.0.2.2\n192.0.2.3\n"), 1, "ip", 1024, 2); !errors.Is(err, ErrNetworkFeedBudget) || inds != nil {
		t.Fatalf("record cap bypass %v %v", inds, err)
	}
	daily := strings.Repeat(`{"value":"192.0.2.1","op":"add"}`+"\n", 3)
	if ops, _, err := parseDailyBounded(strings.NewReader(daily), 1024, 2); !errors.Is(err, ErrNetworkFeedBudget) || ops != nil {
		t.Fatalf("daily record cap bypass %v %v", ops, err)
	}
	if ops, err := ParseDaily(strings.NewReader(`{"value":"192.0.2.1","op":"del"}` + "\n" + strings.Repeat("x", maxNetworkFeedLineBytes+1))); err == nil || ops != nil {
		t.Fatal("daily parse returned a usable partial update")
	}
}

func newBoundedFeedTest(t *testing.T) (*Feed, *cache.Cache, *int) {
	t.Helper()
	c, err := cache.Open(filepath.Join(t.TempDir(), "edr.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = c.Close() })
	cfg := config.Default()
	cfg.Server = "mirror.invalid"
	cfg.Blocklist.IndicatorTypes = []string{"ip"}
	cfg.Blocklist.Levels = []int{1}
	calls := new(int)
	f := NewFeed(cfg, c, NewStore(), func([]Indicator) { *calls++ })
	limits := defaultNetworkFeedLimits()
	limits.RawArtifact = 2 << 20
	limits.RawTotal = 4 << 20
	limits.ExpandedArtifact = 2 << 20
	limits.ExpandedTotal = 4 << 20
	limits.Indicators = 2
	limits.DailyOps = 4
	f.limits = &limits
	payload := gzList(t, "198.51.100.4\n")
	f.Fetch = func(string) (io.ReadCloser, error) { return io.NopCloser(bytes.NewReader(payload)), nil }
	f.FetchChecksum = func(string) (string, error) { return sha(payload), nil }
	if err := f.UpdateOnce(context.Background()); err != nil {
		t.Fatal(err)
	}
	*calls = 0
	return f, c, calls
}

func assertBoundedFeedLastGood(t *testing.T, f *Feed, c *cache.Cache, calls int, checksum string) {
	t.Helper()
	if calls != 0 || f.store.Count() != 1 {
		t.Fatalf("partial feed published: callbacks=%d count=%d", calls, f.store.Count())
	}
	if _, ok := f.store.Match(netip.MustParseAddr("198.51.100.4")); !ok {
		t.Fatal("last-good indicator lost")
	}
	rows, err := c.ListIndicators()
	if err != nil || len(rows) != 1 || rows[0].Value != "198.51.100.4" {
		t.Fatalf("durable last-good changed %v %v", rows, err)
	}
	state, err := c.GetFeedState("level1/ip")
	if err != nil || state.Checksum != checksum {
		t.Fatalf("failed cycle advanced durable source state %+v %v", state, err)
	}
	if f.HealthError() == nil {
		t.Fatal("feed budget/parse error absent from health")
	}
}

func TestNetworkFeedAggregateBoundsPreserveLastGoodAtomically(t *testing.T) {
	for _, kind := range []string{"raw", "expanded", "indicators"} {
		t.Run(kind, func(t *testing.T) {
			f, c, calls := newBoundedFeedTest(t)
			state, _ := c.GetFeedState("level1/ip")
			f.cfg.Blocklist.IndicatorTypes = []string{"ip", "domain"}
			ip := gzList(t, "192.0.2.1\n192.0.2.2\n")
			domain := gzList(t, "new.example\n")
			if kind == "raw" {
				f.limits.RawArtifact = 64
				f.limits.RawTotal = int64(len(ip) + len(domain) - 1)
			}
			if kind == "expanded" {
				f.limits.ExpandedArtifact = 64
				f.limits.ExpandedTotal = 25
			}
			f.Fetch = func(url string) (io.ReadCloser, error) {
				payload := ip
				if strings.HasSuffix(url, "/domain") {
					payload = domain
				}
				return io.NopCloser(bytes.NewReader(payload)), nil
			}
			f.FetchChecksum = func(url string) (string, error) {
				payload := ip
				if strings.HasSuffix(url, "/domain.sha256") {
					payload = domain
				}
				return sha(payload), nil
			}
			if err := f.UpdateOnce(context.Background()); !errors.Is(err, ErrNetworkFeedBudget) {
				t.Fatalf("aggregate %s cap not reported: %v", kind, err)
			}
			assertBoundedFeedLastGood(t, f, c, *calls, state.Checksum)
		})
	}
}

func TestNetworkDailyBudgetAndParserErrorsDoNotApplyPartialDeletes(t *testing.T) {
	for _, kind := range []string{"records", "parser", "ops", "empty-checksum"} {
		t.Run(kind, func(t *testing.T) {
			f, c, calls := newBoundedFeedTest(t)
			state, _ := c.GetFeedState("level1/ip")
			payload := []byte(`{"value":"192.0.2.1","op":"add"}` + "\n" + `{"value":"192.0.2.2","op":"add"}` + "\n")
			if kind == "parser" {
				payload = []byte(`{"value":"198.51.100.4","op":"del"}` + "\n" + strings.Repeat("x", maxNetworkFeedLineBytes+1))
			}
			if kind == "ops" {
				payload = bytes.Repeat([]byte(`{"value":"198.51.100.4","op":"del"}`+"\n"), 5)
			}
			f.Fetch = func(string) (io.ReadCloser, error) { return io.NopCloser(bytes.NewReader(payload)), nil }
			f.FetchChecksum = func(string) (string, error) {
				if kind == "empty-checksum" {
					return "", nil
				}
				return sha(payload), nil
			}
			if err := f.updateDeltaOnce(context.Background()); err == nil {
				t.Fatalf("daily %s failure hidden", kind)
			}
			assertBoundedFeedLastGood(t, f, c, *calls, state.Checksum)
		})
	}
}

type feedRoundTripFunc func(*http.Request) (*http.Response, error)

func (f feedRoundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestNetworkFeedHTTPHonorsCancellationAndDeclaredByteCap(t *testing.T) {
	stopped := make(chan struct{})
	f := &Feed{httpClient: &http.Client{Transport: feedRoundTripFunc(func(r *http.Request) (*http.Response, error) {
		<-r.Context().Done()
		close(stopped)
		return nil, r.Context().Err()
	})}}
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	if body, err := f.httpFetch(ctx, "https://mirror.invalid/feed"); err == nil {
		body.Close()
		t.Fatal("HTTP cancellation ignored")
	}
	select {
	case <-stopped:
	case <-time.After(time.Second):
		t.Fatal("cancelled request remained active")
	}
	f.httpClient = &http.Client{Transport: feedRoundTripFunc(func(r *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, ContentLength: MaxNetworkFeedRawBytes + 1, Body: io.NopCloser(strings.NewReader("")), Header: make(http.Header)}, nil
	})}
	if body, err := f.httpFetch(context.Background(), "https://mirror.invalid/feed"); !errors.Is(err, ErrNetworkFeedBudget) {
		if body != nil {
			body.Close()
		}
		t.Fatalf("declared size cap ignored: %v", err)
	}
}

func TestNetworkFeedSourceFanoutIsBounded(t *testing.T) {
	if err := feedArtifactCount([]string{"ip", "domain", "hostname"}, []int{1, 2, 3}, 8); !errors.Is(err, ErrNetworkFeedBudget) {
		t.Fatal(err)
	}
	if err := feedArtifactCount([]string{"../bad"}, []int{1}, 64); err == nil {
		t.Fatal("unsafe artifact type accepted")
	}
}
