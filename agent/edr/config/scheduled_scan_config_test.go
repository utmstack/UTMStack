package config

import (
	"os"
	"testing"
)

func TestScheduledScanDefaultsDisabled(t *testing.T) {
	d := Default()
	if d.ScheduledScan.Enabled {
		t.Fatal("ScheduledScan must default to disabled")
	}
	if d.ScheduledScan.Schedule != "every:24h" {
		t.Fatalf("Schedule default = %q, want every:24h", d.ScheduledScan.Schedule)
	}
	if d.ScheduledScan.BatchSize != 256 {
		t.Fatalf("BatchSize default = %d, want 256", d.ScheduledScan.BatchSize)
	}
	if d.ScheduledScan.ThrottleMs != 0 {
		t.Fatalf("ThrottleMs default = %d, want 0", d.ScheduledScan.ThrottleMs)
	}
}

func TestLoadPicksUpScheduledScanBlock(t *testing.T) {
	orig := ConfigFile
	ConfigFile = t.TempDir() + "/edr.json"
	defer func() { ConfigFile = orig }()

	onDisk := `{
      "scheduled_scan": {
        "enabled": true,
        "schedule": "every:6h",
        "paths": ["D:", "E:\\data"],
        "throttle_ms": 50,
        "batch_size": 128
      }
    }`
	if err := os.WriteFile(ConfigFile, []byte(onDisk), 0o644); err != nil {
		t.Fatal(err)
	}
	c, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if !c.ScheduledScan.Enabled {
		t.Fatal("enabled not picked up from disk")
	}
	if c.ScheduledScan.Schedule != "every:6h" {
		t.Fatalf("Schedule = %q, want every:6h", c.ScheduledScan.Schedule)
	}
	if len(c.ScheduledScan.Paths) != 2 || c.ScheduledScan.Paths[1] != `E:\data` {
		t.Fatalf("Paths = %v", c.ScheduledScan.Paths)
	}
	if c.ScheduledScan.ThrottleMs != 50 || c.ScheduledScan.BatchSize != 128 {
		t.Fatalf("throttle/batch = %d/%d, want 50/128", c.ScheduledScan.ThrottleMs, c.ScheduledScan.BatchSize)
	}
}

func TestLoadBackfillsBatchSizeWhenBlockPresent(t *testing.T) {
	orig := ConfigFile
	ConfigFile = t.TempDir() + "/edr.json"
	defer func() { ConfigFile = orig }()

	// A block that enables the scan but omits batch_size: the default must be
	// backfilled rather than collapsing to 0.
	if err := os.WriteFile(ConfigFile, []byte(`{"scheduled_scan": {"enabled": true, "batch_size": 0}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	c, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if !c.ScheduledScan.Enabled {
		t.Fatal("enabled not picked up from disk")
	}
	if c.ScheduledScan.BatchSize != 256 {
		t.Fatalf("BatchSize = %d, want backfilled default 256", c.ScheduledScan.BatchSize)
	}
	if c.ScheduledScan.Schedule != "every:24h" {
		t.Fatalf("Schedule = %q, want backfilled default every:24h", c.ScheduledScan.Schedule)
	}
}
