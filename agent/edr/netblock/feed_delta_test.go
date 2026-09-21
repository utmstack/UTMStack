// edr/netblock/feed_delta_test.go
package netblock

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/netip"
	"path/filepath"
	"testing"

	"github.com/utmstack/UTMStack/agent/edr/cache"
	"github.com/utmstack/UTMStack/agent/edr/config"
)

// TestUpdateDeltaOnceAppliesAddsAndDels bootstraps the store from an accumulative
// (gzip) snapshot, then applies a PLAIN ndjson daily delta (add + del) and asserts
// the live set reflects both operations.
func TestUpdateDeltaOnceAppliesAddsAndDels(t *testing.T) {
	c, _ := cache.Open(filepath.Join(t.TempDir(), "edr.db"))
	defer c.Close()
	cfg := config.Default()
	cfg.Server = "utm.example.com"
	cfg.Blocklist.Levels = []int{1}
	cfg.Blocklist.IndicatorTypes = []string{"ip"}
	store := NewStore()

	var applied []Indicator
	f := NewFeed(cfg, c, store, func(inds []Indicator) { applied = inds })

	// Bootstrap: accumulative gz list with 1.1.1.1 and 2.2.2.2.
	boot := gzList(t, "1.1.1.1\n2.2.2.2\n")
	f.Fetch = func(string) (io.ReadCloser, error) { return io.NopCloser(bytes.NewReader(boot)), nil }
	f.FetchChecksum = func(string) (string, error) { return sha(boot), nil }
	if err := f.UpdateOnce(context.Background()); err != nil {
		t.Fatal(err)
	}
	if _, ok := store.Match(netip.MustParseAddr("1.1.1.1")); !ok {
		t.Fatal("bootstrap: 1.1.1.1 should be blocked")
	}

	// Daily delta: PLAIN ndjson (NOT gzip) — add 3.3.3.3, del 1.1.1.1.
	daily := []byte(`{"value":"3.3.3.3","type":"ip","op":"add"}` + "\n" +
		`{"value":"1.1.1.1","type":"ip","op":"del"}` + "\n")
	f.Fetch = func(string) (io.ReadCloser, error) { return io.NopCloser(bytes.NewReader(daily)), nil }
	f.FetchChecksum = func(string) (string, error) { return sha(daily), nil }
	if err := f.updateDeltaOnce(context.Background()); err != nil {
		t.Fatal(err)
	}

	if _, ok := store.Match(netip.MustParseAddr("1.1.1.1")); ok {
		t.Fatal("1.1.1.1 should be removed by the daily del")
	}
	if _, ok := store.Match(netip.MustParseAddr("3.3.3.3")); !ok {
		t.Fatal("3.3.3.3 should be added by the daily add")
	}
	if _, ok := store.Match(netip.MustParseAddr("2.2.2.2")); !ok {
		t.Fatal("2.2.2.2 should remain")
	}
	// applyFn saw the merged set (2 indicators: 2.2.2.2 + 3.3.3.3).
	if len(applied) != 2 {
		t.Fatalf("apply called with %d indicators, want 2", len(applied))
	}
}

// TestUpdateDeltaOnceTotalFailureKeepsLastGood asserts a daily-delta cycle where
// every artifact fetch fails returns an error and never wipes the live set.
func TestUpdateDeltaOnceTotalFailureKeepsLastGood(t *testing.T) {
	c, _ := cache.Open(filepath.Join(t.TempDir(), "edr.db"))
	defer c.Close()
	cfg := config.Default()
	cfg.Server = "utm.example.com"
	cfg.Blocklist.Levels = []int{1}
	cfg.Blocklist.IndicatorTypes = []string{"ip"}
	store := NewStore()

	applyCalls := 0
	f := NewFeed(cfg, c, store, func([]Indicator) { applyCalls++ })

	boot := gzList(t, "1.1.1.1\n2.2.2.2\n")
	f.Fetch = func(string) (io.ReadCloser, error) { return io.NopCloser(bytes.NewReader(boot)), nil }
	f.FetchChecksum = func(string) (string, error) { return sha(boot), nil }
	if err := f.UpdateOnce(context.Background()); err != nil {
		t.Fatal(err)
	}
	goodCalls := applyCalls

	// Mirror goes dark for the daily artifacts.
	f.Fetch = func(string) (io.ReadCloser, error) { return nil, fmt.Errorf("mirror unreachable") }
	if err := f.updateDeltaOnce(context.Background()); err == nil {
		t.Fatal("updateDeltaOnce must return an error when all daily artifacts fail")
	}
	if _, ok := store.Match(netip.MustParseAddr("1.1.1.1")); !ok {
		t.Fatal("1.1.1.1 must remain blocked after a failed delta")
	}
	if _, ok := store.Match(netip.MustParseAddr("2.2.2.2")); !ok {
		t.Fatal("2.2.2.2 must remain blocked after a failed delta")
	}
	if applyCalls != goodCalls {
		t.Fatalf("applyFn invoked %d extra time(s) after total delta failure; must not re-apply", applyCalls-goodCalls)
	}
}

// TestUpdateDeltaOnceErrorsWithoutServer asserts that with no server/mirror
// configured the delta path returns an error without attempting any fetch.
func TestUpdateDeltaOnceErrorsWithoutServer(t *testing.T) {
	c, _ := cache.Open(filepath.Join(t.TempDir(), "edr.db"))
	defer c.Close()
	cfg := config.Default()
	cfg.Server = ""
	store := NewStore()
	f := NewFeed(cfg, c, store, nil)
	called := false
	f.Fetch = func(string) (io.ReadCloser, error) { called = true; return nil, nil }
	if err := f.updateDeltaOnce(context.Background()); err == nil {
		t.Fatal("expected error when no mirror URL can be derived")
	}
	if called {
		t.Fatal("must not fetch without a mirror URL")
	}
}
