// edr/netblock/feed_test.go
package netblock

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/netip"
	"path/filepath"
	"testing"

	"github.com/utmstack/UTMStack/agent/edr/cache"
	"github.com/utmstack/UTMStack/agent/edr/config"
)

func sha(b []byte) string { s := sha256.Sum256(b); return hex.EncodeToString(s[:]) }

func TestFeedBootstrapAndChecksum(t *testing.T) {
	c, _ := cache.Open(filepath.Join(t.TempDir(), "edr.db"))
	defer c.Close()
	cfg := config.Default()
	cfg.Server = "utm.example.com"
	cfg.Blocklist.Levels = []int{1}
	cfg.Blocklist.IndicatorTypes = []string{"ip"}
	store := NewStore()
	var applied []Indicator
	f := NewFeed(cfg, c, store, func(inds []Indicator) { applied = inds })

	var gzbuf bytes.Buffer
	w := gzip.NewWriter(&gzbuf)
	w.Write([]byte("1.2.3.4\n5.6.7.8\n"))
	w.Close()
	body := gzbuf.Bytes()

	f.Fetch = func(url string) (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(body)), nil
	}
	f.FetchChecksum = func(url string) (string, error) { return sha(body), nil }

	if err := f.UpdateOnce(context.Background()); err != nil {
		t.Fatal(err)
	}
	if store.Count() != 2 {
		t.Fatalf("store count = %d, want 2", store.Count())
	}
	if len(applied) != 2 {
		t.Fatalf("apply called with %d indicators", len(applied))
	}
}

func TestFeedRejectsBadChecksum(t *testing.T) {
	c, _ := cache.Open(filepath.Join(t.TempDir(), "edr.db"))
	defer c.Close()
	cfg := config.Default()
	cfg.Server = "utm.example.com"
	cfg.Blocklist.IndicatorTypes = []string{"ip"}
	store := NewStore()
	f := NewFeed(cfg, c, store, func([]Indicator) {})
	var b bytes.Buffer
	gw := gzip.NewWriter(&b)
	gw.Write([]byte("1.2.3.4\n"))
	gw.Close()
	f.Fetch = func(string) (io.ReadCloser, error) { return io.NopCloser(bytes.NewReader(b.Bytes())), nil }
	f.FetchChecksum = func(string) (string, error) { return "deadbeef", nil } // wrong
	// Fetches are now best-effort per artifact: whether UpdateOnce returns an
	// error or nil, the key invariant is that an unverified artifact is never
	// swapped in — the store must stay empty on a checksum mismatch.
	_ = f.UpdateOnce(context.Background())
	if store.Count() != 0 {
		t.Fatal("store must stay empty on checksum failure")
	}
}

// TestUpdateOnceAllFailKeepsLastGood asserts the stale-never-empty invariant:
// when every feed artifact fetch fails (mirror outage), UpdateOnce must NOT wipe
// the live blocklist. Enforcement continues on the last-known-good set; the store,
// f.cur and applyFn are left untouched, and an error is returned so the caller can
// flag the feed as stale (lastSync must not advance).
func TestUpdateOnceAllFailKeepsLastGood(t *testing.T) {
	c, _ := cache.Open(filepath.Join(t.TempDir(), "edr.db"))
	defer c.Close()
	cfg := config.Default()
	cfg.Server = "utm.example.com"
	cfg.Blocklist.Levels = []int{1}
	cfg.Blocklist.IndicatorTypes = []string{"ip"}
	store := NewStore()

	var lastApplied []Indicator
	applyCalls := 0
	f := NewFeed(cfg, c, store, func(inds []Indicator) {
		lastApplied = inds
		applyCalls++
	})

	// One successful bootstrap: two indicators fetched + checksum-verified.
	body := gzList(t, "1.2.3.4\n5.6.7.8\n")
	f.Fetch = func(url string) (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(body)), nil
	}
	f.FetchChecksum = func(url string) (string, error) { return sha(body), nil }
	if err := f.UpdateOnce(context.Background()); err != nil {
		t.Fatal(err)
	}
	if store.Count() != 2 {
		t.Fatalf("after bootstrap store count = %d, want 2", store.Count())
	}
	if applyCalls != 1 || len(lastApplied) != 2 {
		t.Fatalf("bootstrap apply: calls=%d applied=%d, want calls=1 applied=2", applyCalls, len(lastApplied))
	}
	goodCalls := applyCalls
	f.mu.Lock()
	goodCurLen := len(f.cur)
	f.mu.Unlock()

	// Mirror goes dark: every artifact fetch errors → success == 0.
	f.Fetch = func(url string) (io.ReadCloser, error) {
		return nil, fmt.Errorf("mirror unreachable")
	}
	err := f.UpdateOnce(context.Background())
	if err == nil {
		t.Fatal("UpdateOnce must return an error when all artifacts fail")
	}
	// Last-known-good set still enforced.
	if store.Count() != 2 {
		t.Fatalf("store must keep last-known-good set; count = %d, want 2", store.Count())
	}
	if _, ok := store.Match(netip.MustParseAddr("1.2.3.4")); !ok {
		t.Fatal("1.2.3.4 must still be blocked after a failed update")
	}
	// applyFn must NOT have been re-invoked (never applied an empty set).
	if applyCalls != goodCalls {
		t.Fatalf("applyFn invoked %d extra time(s) after total failure; must not re-apply", applyCalls-goodCalls)
	}
	if len(lastApplied) != 2 {
		t.Fatalf("lastApplied set was overwritten: len=%d, want 2", len(lastApplied))
	}
	// f.cur unchanged.
	f.mu.Lock()
	gotCurLen := len(f.cur)
	f.mu.Unlock()
	if gotCurLen != goodCurLen {
		t.Fatalf("f.cur changed after failed update: %d != %d", gotCurLen, goodCurLen)
	}
}

func TestFeedAppliesDelta(t *testing.T) {
	c, _ := cache.Open(filepath.Join(t.TempDir(), "edr.db"))
	defer c.Close()
	cfg := config.Default()
	cfg.Blocklist.IndicatorTypes = []string{"ip"}
	store := NewStore()
	i1, _ := ParseIndicator("1.1.1.1", "ip", 1)
	i2, _ := ParseIndicator("2.2.2.2", "ip", 1)
	store.Swap([]Indicator{i1, i2})
	f := NewFeed(cfg, c, store, func([]Indicator) {})
	// del 1.1.1.1, add 3.3.3.3
	ops := []DailyOp{
		{Value: "1.1.1.1", Type: "ip", Op: "del"},
		{Value: "3.3.3.3", Type: "ip", Op: "add"},
	}
	f.applyDelta(ops)
	if _, ok := store.Match(netip.MustParseAddr("1.1.1.1")); ok {
		t.Fatal("1.1.1.1 should be removed")
	}
	if _, ok := store.Match(netip.MustParseAddr("3.3.3.3")); !ok {
		t.Fatal("3.3.3.3 should be added")
	}
	if _, ok := store.Match(netip.MustParseAddr("2.2.2.2")); !ok {
		t.Fatal("2.2.2.2 should remain")
	}
}
