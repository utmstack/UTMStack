package service

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/utmstack/UTMStack/agent/edr/config"
	"github.com/utmstack/UTMStack/agent/edr/orchestrator"
)

func TestLivePolicyStatusTracksAppliedValuesOnly(t *testing.T) {
	prevFile, prevDir := config.ConfigFile, config.InstallDir
	config.InstallDir = t.TempDir()
	config.ConfigFile = filepath.Join(config.InstallDir, "edr.json")
	defer func() { config.ConfigFile = prevFile; config.InstallDir = prevDir }()
	active := config.Default()
	active.Enabled = true
	onDisk := config.Default()
	onDisk.Enabled = false
	onDisk.Allowlist.Paths = []string{"*.inert"}
	onDisk.Ransomware.ResponseMode = "alert"
	onDisk.ScheduledScan = config.ScheduledScan{Enabled: true, Paths: []string{config.InstallDir}}
	if err := config.Save(onDisk); err != nil {
		t.Fatal(err)
	}
	p := &program{pathEx: orchestrator.NewExcluder(nil)}
	p.maybeReload(&active)
	if !active.Enabled || active.ScheduledScan.Enabled || active.Ransomware.ResponseMode != "alert" || len(active.Allowlist.Paths) != 1 || !p.pathEx.Excluded("a.inert") {
		t.Fatalf("incorrect live policy %+v", active)
	}
}

func TestStoppedServiceCannotRetainActiveProtectionStatus(t *testing.T) {
	prev := config.StatusFile
	config.StatusFile = filepath.Join(t.TempDir(), "status.json")
	defer func() { config.StatusFile = prev }()
	doc := statusDoc{Running: true, EngineHealthy: true, EngineResident: true, SensorFileWatcher: true, SensorAMSI: true, BlocklistEnabled: true, BlocklistEnforce: true, BlocklistActiveFilters: 5, SigDBVersion: "retained-evidence", SensorStates: map[string]sensorState{"amsi": {Supported: true, Requested: true, State: "active"}}}
	raw, _ := json.Marshal(doc)
	if err := os.WriteFile(config.StatusFile, raw, 0600); err != nil {
		t.Fatal(err)
	}
	queue := orchestrator.New(nil, nil, 1, 1)
	queue.Enqueue(orchestrator.FileEvent{Path: "waiting-at-cancellation"})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	queue.Start(ctx)
	queue.Stop()
	writeStoppedStatus(queue, nil)
	raw, err := os.ReadFile(config.StatusFile)
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(raw, &doc); err != nil {
		t.Fatal(err)
	}
	if doc.Running || doc.EngineHealthy || doc.SensorFileWatcher || doc.SensorAMSI || doc.BlocklistEnabled || doc.BlocklistEnforce || doc.BlocklistActiveFilters != 0 || doc.SigDBVersion != "retained-evidence" || doc.SensorStates["amsi"].State != "stopped" {
		t.Fatalf("stale protection status %+v", doc)
	}
	if doc.ScanQueue == nil || doc.ScanQueue.Abandoned != 1 || doc.ScanQueue.Active != 0 || doc.ScanQueue.Queued != 0 {
		t.Fatalf("final queue counters missing: %+v", doc.ScanQueue)
	}
}
