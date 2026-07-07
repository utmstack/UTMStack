// edr/netblock/feed_test.go
package netblock

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
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
	cfg.Blocklist.Levels = []int{1}
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
	store := NewStore()
	f := NewFeed(config.Default(), c, store, func([]Indicator) {})
	var b bytes.Buffer
	gw := gzip.NewWriter(&b)
	gw.Write([]byte("1.2.3.4\n"))
	gw.Close()
	f.Fetch = func(string) (io.ReadCloser, error) { return io.NopCloser(bytes.NewReader(b.Bytes())), nil }
	f.FetchChecksum = func(string) (string, error) { return "deadbeef", nil } // wrong
	if err := f.UpdateOnce(context.Background()); err == nil {
		t.Fatal("expected checksum-mismatch error")
	}
	if store.Count() != 0 {
		t.Fatal("store must stay empty on checksum failure")
	}
}
