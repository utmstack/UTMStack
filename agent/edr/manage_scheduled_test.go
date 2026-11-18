package main

import (
	"encoding/json"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/utmstack/UTMStack/agent/edr/config"
)

func TestScheduledCLIValidatesAtomically(t *testing.T) {
	cfg := config.Default()
	if err := applySet(&cfg, "scheduled_scan.enabled", "true"); err == nil {
		t.Fatal("enabled without roots")
	}
	raw, _ := json.Marshal([]string{filepath.Clean(t.TempDir())})
	for _, kv := range [][2]string{{"scheduled_scan.paths", string(raw)}, {"scheduled_scan.enabled", "true"}, {"scheduled_scan.run_on_start", "true"}, {"scheduled_scan.max_files", "9"}, {"scheduled_scan.max_entries", "20"}, {"scheduled_scan.max_depth", "4"}, {"scheduled_scan.interval_minutes", "6"}, {"scheduled_scan.max_duration_minutes", "2"}} {
		if err := applySet(&cfg, kv[0], kv[1]); err != nil {
			t.Fatal(err)
		}
	}
	before := cfg.ScheduledScan
	for _, kv := range [][2]string{{"scheduled_scan.paths", `["relative"]`}, {"scheduled_scan.max_files", "-1"}, {"scheduled_scan.max_entries", "1000001"}, {"scheduled_scan.max_depth", "129"}, {"scheduled_scan.interval_minutes", "43201"}, {"scheduled_scan.run_on_start", "perhaps"}, {"scheduled_scan.unrecognized", "1"}} {
		if err := applySet(&cfg, kv[0], kv[1]); err == nil {
			t.Errorf("accepted %v", kv)
		}
		if !reflect.DeepEqual(before, cfg.ScheduledScan) {
			t.Fatalf("invalid change mutated configuration: %v", kv)
		}
	}
}
