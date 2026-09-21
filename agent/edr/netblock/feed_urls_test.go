// utmstack-v12/agent/edr/netblock/feed_urls_test.go
package netblock

import (
	"bytes"
	"compress/gzip"
	"context"
	"io"
	"path/filepath"
	"testing"

	"github.com/utmstack/UTMStack/agent/edr/cache"
	"github.com/utmstack/UTMStack/agent/edr/config"
)

func gzList(t *testing.T, s string) []byte {
	t.Helper()
	var b bytes.Buffer
	w := gzip.NewWriter(&b)
	if _, err := w.Write([]byte(s)); err != nil {
		t.Fatal(err)
	}
	_ = w.Close()
	return b.Bytes()
}

func TestAutoDerivedURLsArePathBased(t *testing.T) {
	c, _ := cache.Open(filepath.Join(t.TempDir(), "edr.db"))
	defer c.Close()
	cfg := config.Default()
	cfg.Server = "utm.example.com"
	cfg.Blocklist.Levels = []int{1}
	cfg.Blocklist.IndicatorTypes = []string{"ip"}
	store := NewStore()
	f := NewFeed(cfg, c, store, nil)

	payload := gzList(t, "1.2.3.4\n")
	var fetched, checksummed []string
	f.Fetch = func(url string) (io.ReadCloser, error) {
		fetched = append(fetched, url)
		return io.NopCloser(bytes.NewReader(payload)), nil
	}
	f.FetchChecksum = func(url string) (string, error) {
		checksummed = append(checksummed, url)
		return sha(payload), nil
	}
	if err := f.UpdateOnce(context.Background()); err != nil {
		t.Fatal(err)
	}

	wantBase := "https://utm.example.com:9001/private/edr/feeds/v1"
	if len(fetched) != 1 || fetched[0] != wantBase+"/download/list/level1/accumulative/ip" {
		t.Fatalf("fetch urls = %v", fetched)
	}
	if len(checksummed) != 1 || checksummed[0] != wantBase+"/download/list/level1/accumulative/ip.sha256" {
		t.Fatalf("checksum urls = %v", checksummed)
	}
}

func TestUpdateOnceErrorsWithoutServerOrMirror(t *testing.T) {
	c, _ := cache.Open(filepath.Join(t.TempDir(), "edr.db"))
	defer c.Close()
	cfg := config.Default()
	cfg.Server = ""
	store := NewStore()
	f := NewFeed(cfg, c, store, nil)
	called := false
	f.Fetch = func(string) (io.ReadCloser, error) { called = true; return nil, nil }
	if err := f.UpdateOnce(context.Background()); err == nil {
		t.Fatal("expected error when no mirror URL can be derived")
	}
	if called {
		t.Fatal("must not fetch without a mirror URL")
	}
}
