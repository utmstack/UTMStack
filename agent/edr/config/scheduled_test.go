package config

import (
	"encoding/json"
	"path/filepath"
	"reflect"
	"testing"
)

func TestScheduledConfigBoundsAndStrictKeys(t *testing.T) {
	for _, raw := range []string{`{"enable":true}`, `{"max_files":-1}`, `{"max_files":500001}`, `{"max_entries":1000001}`, `{"max_depth":129}`, `{"interval_minutes":43201}`, `{"max_duration_minutes":1441}`, `{"enabled":true}`, `{"paths":["relative"]}`} {
		var s ScheduledScan
		err := json.Unmarshal([]byte(raw), &s)
		if err == nil {
			err = s.Validate()
		}
		if err == nil {
			t.Errorf("accepted %s", raw)
		}
	}
	s := ScheduledScan{}
	if err := s.Validate(); err != nil {
		t.Fatal(err)
	}
	e := s.Effective()
	if e.Enabled || e.IntervalMinutes != 240 || e.MaxFiles != 100000 || e.MaxEntries != 200000 || e.MaxDepth != 64 || e.MaxDurationMinutes != 60 {
		t.Fatalf("defaults %+v", e)
	}
}

func TestScheduledConfigSurvivesSaveAndLoad(t *testing.T) {
	previousFile, previousDir := ConfigFile, InstallDir
	InstallDir = t.TempDir()
	ConfigFile = filepath.Join(InstallDir, "edr.json")
	defer func() { ConfigFile = previousFile; InstallDir = previousDir }()
	cfg := Default()
	cfg.ScheduledScan = ScheduledScan{Enabled: true, Paths: []string{InstallDir}, IntervalMinutes: 7, RunOnStart: true, MaxFiles: 4, MaxEntries: 8, MaxDepth: 2, MaxDurationMinutes: 1}
	if err := Save(cfg); err != nil {
		t.Fatal(err)
	}
	loaded, err := Load()
	if err != nil || !reflect.DeepEqual(loaded.ScheduledScan, cfg.ScheduledScan) {
		t.Fatalf("roundtrip %+v %v", loaded.ScheduledScan, err)
	}
}
