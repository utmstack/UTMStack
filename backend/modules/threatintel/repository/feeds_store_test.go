package repository

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"gopkg.in/yaml.v3"
)

// The event processor's feed matcher and the backend's contribution task read this file and the
// backend writes it. Its shape is the contract: plugins.<key>.enabled, the same wrapper every other
// plugin config uses, with one key for each of the two.
func TestTheFileIsShapedTheWayItIsRead(t *testing.T) {
	dir := t.TempDir()
	s := NewConfigStore(dir)

	if err := s.Update(func(c *FeedsConfig) {
		c.FeedsEnabled.Enabled = true
		c.FeedsContributionEnabled.Enabled = true
	}); err != nil {
		t.Fatalf("Update: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "system_plugins_feeds.yaml"))
	if err != nil {
		t.Fatalf("read: %v", err)
	}

	var parsed struct {
		Plugins map[string]struct {
			Enabled bool `yaml:"enabled"`
		} `yaml:"plugins"`
	}

	if err := yaml.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("the file is not yaml: %v\n%s", err, data)
	}

	for _, key := range []string{"feeds", "feeds_contribution"} {
		if got, ok := parsed.Plugins[key]; !ok || !got.Enabled {
			t.Errorf("plugins.%s.enabled must be true in:\n%s", key, data)
		}
	}

	if strings.Contains(string(data), "api_key") || strings.Contains(string(data), "api_secret") {
		t.Errorf("the file no longer carries credentials:\n%s", data)
	}
}

// Each switch is its own: turning the contribution off must not turn the feed matcher off, and the
// other way round. An update re-reads what is on disk instead of writing a config it did not see.
func TestSwitchingOneOffLeavesTheOtherAlone(t *testing.T) {
	s := NewConfigStore(t.TempDir())

	if err := s.Update(func(c *FeedsConfig) {
		c.FeedsEnabled.Enabled = true
		c.FeedsContributionEnabled.Enabled = true
	}); err != nil {
		t.Fatal(err)
	}

	if err := s.Update(func(c *FeedsConfig) { c.FeedsContributionEnabled.Enabled = false }); err != nil {
		t.Fatal(err)
	}

	got, err := s.Load()
	if err != nil {
		t.Fatal(err)
	}

	if got.FeedsContributionEnabled.Enabled {
		t.Error("the contribution stayed on")
	}

	if !got.FeedsEnabled.Enabled {
		t.Error("turning the contribution off turned the feed matcher off too")
	}

	if err := s.Update(func(c *FeedsConfig) { c.FeedsEnabled.Enabled = false }); err != nil {
		t.Fatal(err)
	}

	got, _ = s.Load()
	if got.FeedsEnabled.Enabled || got.FeedsContributionEnabled.Enabled {
		t.Errorf("got %+v, want both off", got)
	}
}

// Nothing on disk is not an error: a fresh install has never contributed.
func TestNoFileYetReadsAsOff(t *testing.T) {
	got, err := NewConfigStore(t.TempDir()).Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if got.FeedsEnabled.Enabled || got.FeedsContributionEnabled.Enabled {
		t.Errorf("got %+v, want the zero config", got)
	}
}

// A file that is not valid yaml is reported, not read as "off": the task must not quietly stop
// because someone edited the file badly.
func TestABrokenFileIsAnError(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "system_plugins_feeds.yaml"), []byte("plugins: [not a map"), 0o600); err != nil {
		t.Fatal(err)
	}

	if _, err := NewConfigStore(dir).Load(); err == nil {
		t.Error("a file that cannot be parsed must be an error")
	}
}

// The config directory is a shared volume and every replica writes to it. Two saves at once must
// leave one whole file, not halves of both: every writer sets the two switches to the same value, so
// a mixed result means two writers were interleaved.
func TestConcurrentWritesLeaveOneWholeFile(t *testing.T) {
	dir := t.TempDir()
	s := NewConfigStore(dir)

	var wg sync.WaitGroup

	for i := 0; i < 8; i++ {
		wg.Add(1)

		go func(n int) {
			defer wg.Done()

			on := n%2 == 0
			_ = s.Update(func(c *FeedsConfig) {
				c.FeedsEnabled.Enabled = on
				c.FeedsContributionEnabled.Enabled = on
			})
		}(i)
	}

	wg.Wait()

	got, err := s.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if got.FeedsEnabled.Enabled != got.FeedsContributionEnabled.Enabled {
		t.Errorf("two writers' values were mixed: %+v", got)
	}
}
