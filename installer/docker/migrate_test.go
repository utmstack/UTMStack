package docker

import (
	"testing"

	"github.com/utmstack/UTMStack/installer/system"
)

// TestMigratePersistedMemoryAllocation covers the rewrite of the legacy
// single "event-processor" persisted entry into per-container entries.
func TestMigratePersistedMemoryAllocation(t *testing.T) {
	Services = []system.ServiceConfig{
		{Name: "event-processor-manager", Priority: 1, MinMemory: 2560, MaxMemory: 60 * 1024},
		{Name: "event-processor-worker", Priority: 1, MinMemory: 3072, MaxMemory: 60 * 1024},
		{Name: "opensearch", Priority: 1, MinMemory: 5120, MaxMemory: 60 * 1024},
	}

	// No legacy key: nothing to do.
	unchanged := map[string]*system.ServiceConfig{
		"event-processor-manager": {Name: "event-processor-manager", AssignedMemory: 3000},
		"event-processor-worker":  {Name: "event-processor-worker", AssignedMemory: 3000},
	}
	if migratePersistedMemoryAllocation(unchanged) {
		t.Fatal("migratePersistedMemoryAllocation should return false when there is no legacy key")
	}

	// Legacy single entry: manager keeps the old limit, worker gets half.
	rsrcs := map[string]*system.ServiceConfig{
		"event-processor": {Name: "event-processor", AssignedMemory: 5689},
		"opensearch":      {Name: "opensearch", AssignedMemory: 5000},
	}
	if !migratePersistedMemoryAllocation(rsrcs) {
		t.Fatal("migratePersistedMemoryAllocation should return true for a legacy entry")
	}
	if _, ok := rsrcs["event-processor"]; ok {
		t.Error("legacy event-processor key should be removed")
	}
	mgr, ok := rsrcs["event-processor-manager"]
	if !ok {
		t.Fatal("event-processor-manager entry missing after migration")
	}
	if mgr.AssignedMemory != 5689 {
		t.Errorf("manager AssignedMemory = %d, want 5689", mgr.AssignedMemory)
	}
	if mgr.MinMemory != 2560 || mgr.MaxMemory != 60*1024 || mgr.Priority != 1 {
		t.Errorf("manager metadata not taken from Services[]: %+v", mgr)
	}
	wkr, ok := rsrcs["event-processor-worker"]
	if !ok {
		t.Fatal("event-processor-worker entry missing after migration")
	}
	if wkr.AssignedMemory != 3072 { // 5689/2 = 2844, clamped up to MinMemory
		t.Errorf("worker AssignedMemory = %d, want 3072", wkr.AssignedMemory)
	}
	// Other services untouched.
	if os := rsrcs["opensearch"]; os.AssignedMemory != 5000 {
		t.Errorf("opensearch changed: %d", os.AssignedMemory)
	}

	// Idempotent: running the migration again changes nothing.
	if migratePersistedMemoryAllocation(rsrcs) {
		t.Error("second migration should be a no-op")
	}

	// Small legacy value: worker clamped up to its MinMemory.
	small := map[string]*system.ServiceConfig{
		"event-processor": {Name: "event-processor", AssignedMemory: 2000},
	}
	migratePersistedMemoryAllocation(small)
	if w := small["event-processor-worker"]; w.AssignedMemory != 3072 {
		t.Errorf("worker clamped AssignedMemory = %d, want 3072 (MinMemory)", w.AssignedMemory)
	}
}
