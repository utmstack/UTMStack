package repository

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

// The plugin reads this file and the backend writes it. Its shape is the
// contract between them: plugins.feeds.*, the same wrapper every other plugin
// config uses.
func TestTheFileIsShapedTheWayThePluginReadsIt(t *testing.T) {
	dir := t.TempDir()
	s := NewConfigStore(dir)

	if err := s.Update(func(c *FeedsConfig) {
		c.Enabled = true
		c.APIKey = "encrypted-key"
		c.APISecret = "encrypted-secret"
	}); err != nil {
		t.Fatalf("Update: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "system_plugins_feeds.yaml"))
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	for _, want := range []string{"plugins:", "feeds:", "enabled: true", "api_key: encrypted-key", "api_secret: encrypted-secret"} {
		if !strings.Contains(string(data), want) {
			t.Errorf("missing %q in:\n%s", want, data)
		}
	}
}

// Turning the contribution off must not throw away the credentials, or turning
// it back on would mean registering with ThreatWinds all over again.
func TestSwitchingItOffKeepsTheCredentials(t *testing.T) {
	s := NewConfigStore(t.TempDir())

	if err := s.Update(func(c *FeedsConfig) {
		c.Enabled = true
		c.APIKey = "k"
		c.APISecret = "s"
	}); err != nil {
		t.Fatal(err)
	}
	if err := s.Update(func(c *FeedsConfig) { c.Enabled = false }); err != nil {
		t.Fatal(err)
	}

	got, err := s.Load()
	if err != nil {
		t.Fatal(err)
	}
	if got.Enabled {
		t.Error("it stayed on")
	}
	if got.APIKey != "k" || got.APISecret != "s" {
		t.Errorf("credentials = %q/%q, want them kept", got.APIKey, got.APISecret)
	}
}

// Nothing on disk is not an error: a fresh install has never contributed.
func TestNoFileYetReadsAsOff(t *testing.T) {
	got, err := NewConfigStore(t.TempDir()).Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got.Enabled || got.APIKey != "" {
		t.Errorf("got %+v, want the zero config", got)
	}
}

// The config directory is a shared volume and every replica writes to it. Two
// saves at once must leave one whole file, not halves of both.
func TestConcurrentWritesLeaveOneWholeFile(t *testing.T) {
	dir := t.TempDir()
	s := NewConfigStore(dir)

	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			_ = s.Update(func(c *FeedsConfig) {
				c.Enabled = true
				c.APIKey = strings.Repeat(string(rune('a'+n)), 512)
			})
		}(i)
	}
	wg.Wait()

	got, err := s.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(got.APIKey) != 512 {
		t.Fatalf("api_key is %d chars, want one writer's whole value", len(got.APIKey))
	}
	if strings.Trim(got.APIKey, string(got.APIKey[0])) != "" {
		t.Error("two writers' values were mixed")
	}
}
