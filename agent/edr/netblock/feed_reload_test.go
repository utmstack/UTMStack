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
	"sync"
	"testing"
	"time"

	"github.com/utmstack/UTMStack/agent/edr/cache"
	"github.com/utmstack/UTMStack/agent/edr/config"
)

func TestFeedReloadCancelsOldSourceAndAppliesNewSourceWithoutNetworkSockets(t *testing.T) {
	f, _, calls := newBoundedFeedTest(t)
	f.Fetch = nil
	f.FetchChecksum = nil
	f.cfg.Blocklist.MirrorBaseURL = "https://old.invalid"
	started := make(chan struct{})
	var once sync.Once
	payload := gzList(t, "192.0.2.8\n")
	var mu sync.Mutex
	var newURLs []string
	f.httpClient = &http.Client{Transport: feedRoundTripFunc(func(r *http.Request) (*http.Response, error) {
		if r.URL.Host == "old.invalid" {
			once.Do(func() { close(started) })
			<-r.Context().Done()
			return nil, r.Context().Err()
		}
		if r.URL.Host != "new.invalid" {
			t.Errorf("unexpected source %s", r.URL.Host)
			return nil, errors.New("unexpected host")
		}
		mu.Lock()
		newURLs = append(newURLs, r.URL.String())
		mu.Unlock()
		body := payload
		if strings.HasSuffix(r.URL.Path, ".sha256") {
			body = []byte(sha(payload))
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(bytes.NewReader(body)), ContentLength: int64(len(body)), Header: make(http.Header)}, nil
	})}
	oldDone := make(chan error, 1)
	go func() { oldDone <- f.UpdateOnce(context.Background()) }()
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("old request did not begin")
	}
	next := f.cfg
	next.Blocklist.MirrorBaseURL = "https://new.invalid"
	next.Blocklist.Levels = []int{1}
	reloaded := make(chan struct{})
	go func() { f.Reload(next); close(reloaded) }()
	select {
	case <-reloaded:
	case <-time.After(time.Second):
		t.Fatal("reload did not cancel old request")
	}
	if err := <-oldDone; !errors.Is(err, context.Canceled) {
		t.Fatalf("old request error %v", err)
	}
	if *calls != 0 {
		t.Fatal("cancelled source published results")
	}
	next.Blocklist.Levels[0] = 3 // the stored config must own its input slice
	if err := f.UpdateOnce(context.Background()); err != nil {
		t.Fatal(err)
	}
	if _, ok := f.store.Match(netip.MustParseAddr("192.0.2.8")); !ok || *calls != 1 {
		t.Fatal("new source not committed")
	}
	mu.Lock()
	defer mu.Unlock()
	for _, url := range newURLs {
		if !strings.Contains(url, "/level1/") {
			t.Fatalf("caller mutated stored config: %s", url)
		}
	}
}

func TestFeedPersistenceFailureKeepsLiveAndDurableLastGood(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cache.db")
	c, err := cache.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	cfg := config.Default()
	cfg.Server = "mirror.invalid"
	cfg.Blocklist.IndicatorTypes = []string{"ip"}
	calls := 0
	f := NewFeed(cfg, c, NewStore(), func([]Indicator) { calls++ })
	payload := gzList(t, "192.0.2.1\n")
	f.Fetch = func(string) (io.ReadCloser, error) { return io.NopCloser(bytes.NewReader(payload)), nil }
	f.FetchChecksum = func(string) (string, error) { return sha(payload), nil }
	if err = f.UpdateOnce(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err = c.Close(); err != nil {
		t.Fatal(err)
	}
	payload = gzList(t, "192.0.2.2\n")
	if err = f.UpdateOnce(context.Background()); err == nil {
		t.Fatal("closed database write claimed success")
	}
	if _, ok := f.store.Match(netip.MustParseAddr("192.0.2.1")); !ok || f.store.Count() != 1 || calls != 1 || f.HealthError() == nil {
		t.Fatal("persistence failure published a new live set")
	}
	c, err = cache.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	rows, err := c.ListIndicators()
	if err != nil || len(rows) != 1 || rows[0].Value != "192.0.2.1" {
		t.Fatalf("durable last-good lost %v %v", rows, err)
	}
}

func TestDailyDeletionIsDurableAndEmptyLiveStoreWinsOnManagerReload(t *testing.T) {
	f, c, _ := newBoundedFeedTest(t)
	payload := []byte(`{"value":"198.51.100.4","type":"ip","op":"del"}` + "\n")
	f.Fetch = func(string) (io.ReadCloser, error) { return io.NopCloser(bytes.NewReader(payload)), nil }
	f.FetchChecksum = func(string) (string, error) { return sha(payload), nil }
	if err := f.updateDeltaOnce(context.Background()); err != nil {
		t.Fatal(err)
	}
	rows, err := c.ListIndicators()
	if err != nil || len(rows) != 0 || f.store.Count() != 0 {
		t.Fatalf("deleted IOC persisted: %v %v", rows, err)
	}
	// Simulate a historical upsert-only cache while a valid empty snapshot is
	// already live. A policy reload must use the published live snapshot.
	if err = c.PutIndicators([]cache.IndicatorRecord{{Value: "192.0.2.9", Type: "ip"}}); err != nil {
		t.Fatal(err)
	}
	b := newFake()
	cfg := config.Default()
	m := NewManager(Deps{Cfg: cfg, Cache: c, Spool: &memSpool{}, NewBlocker: func(string) (Blocker, error) { return b, nil }})
	m.store.Swap(nil)
	m.Reload(cfg, SystemNets{})
	if b.Count() != 0 {
		t.Fatal("policy reload resurrected stale cached indicators over a valid empty live set")
	}
}

func TestFeedRunReloadWakesNewSourceAndStopsOnCancellation(t *testing.T) {
	f, _, _ := newBoundedFeedTest(t)
	f.cfg.Blocklist.Enabled = false
	payload := gzList(t, "192.0.2.7\n")
	f.Fetch = func(string) (io.ReadCloser, error) { return io.NopCloser(bytes.NewReader(payload)), nil }
	f.FetchChecksum = func(string) (string, error) { return sha(payload), nil }
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan struct{})
	go func() { f.Run(ctx); close(done) }()
	next := f.cfg
	next.Blocklist.Enabled = true
	f.Reload(next)
	deadline := time.Now().Add(time.Second)
	for {
		if _, ok := f.store.Match(netip.MustParseAddr("192.0.2.7")); ok {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("reload did not wake immediate full sync")
		}
		time.Sleep(time.Millisecond)
	}
	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("feed run did not stop")
	}
}
