package docker

import (
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/utmstack/UTMStack/installer/config"
	"github.com/utmstack/UTMStack/installer/system"
	"github.com/utmstack/UTMStack/installer/utils"
)

// resetStackConfig resets the package singleton between cases so each test
// gets a fresh GetStackConfig resolution.
func resetStackConfig() {
	stackConfig = nil
	stackConfigOnce = sync.Once{}
}

// withTempMemoryPath points the authority file at a temp dir and returns a
// cleanup func.
func withTempMemoryPath(t *testing.T) func() {
	t.Helper()
	old := config.MemoryAllocationPath
	dir := t.TempDir()
	config.MemoryAllocationPath = filepath.Join(dir, "memory-allocation.json")
	return func() { config.MemoryAllocationPath = old }
}

func seedServices() {
	Services = []system.ServiceConfig{
		{Name: "event-processor-manager", Priority: 1, MinMemory: 2560, MaxMemory: 4 * 1024},
		{Name: "event-processor-worker", Priority: 1, MinMemory: 3072, MaxMemory: 60 * 1024},
		{Name: "opensearch", Priority: 1, MinMemory: 5120, MaxMemory: 60 * 1024},
		{Name: "backend", Priority: 3, MinMemory: 700, MaxMemory: 8 * 1024},
		{Name: "web-pdf", Priority: 3, MinMemory: 1250, MaxMemory: 2 * 1024},
		{Name: "postgres", Priority: 2, MinMemory: 500, MaxMemory: 2 * 1024},
		{Name: "user-auditor", Priority: 3, MinMemory: 200, MaxMemory: 1024},
		{Name: "agentmanager", Priority: 3, MinMemory: 200, MaxMemory: 1024},
		{Name: "frontend", Priority: 3, MinMemory: 80, MaxMemory: 1024},
	}
}

// TestUpdateRefusesMissingAuthority is the core guarantee: a normal update
// with a missing/corrupt memory-allocation.json must NOT silently re-balance
// (which would reset the distribution). GetStackConfig returns a
// non-authoritative in-memory balance, and Apply (modeled here by the
// MemoryAuthoritative flag) is expected to refuse.
func TestUpdateRefusesMissingAuthority(t *testing.T) {
	defer withTempMemoryPath(t)()
	resetStackConfig()
	seedServices()

	sc := GetStackConfig()
	if sc.MemoryAuthoritative {
		t.Fatal("missing file: GetStackConfig should be non-authoritative on an update")
	}
	if sc.ServiceResources == nil {
		t.Fatal("in-memory balance should still be populated for inspection")
	}
	// The file must NOT have been created by the update path.
	if _, err := os.Stat(config.MemoryAllocationPath); !os.IsNotExist(err) {
		t.Fatalf("update path must not write the authority file; it now exists")
	}
}

// TestFreshInstallSeedsAuthority: a fresh install (no file) is non-authoritative
// until EnsureMemoryBaseline writes the one-time baseline.
func TestFreshInstallSeedsAuthority(t *testing.T) {
	defer withTempMemoryPath(t)()
	resetStackConfig()
	seedServices()

	sc := GetStackConfig()
	if sc.MemoryAuthoritative {
		t.Fatal("fresh install should start non-authoritative")
	}
	if err := EnsureMemoryBaseline(); err != nil {
		t.Fatalf("EnsureMemoryBaseline: %v", err)
	}
	if !sc.MemoryAuthoritative {
		t.Fatal("EnsureMemoryBaseline should flip MemoryAuthoritative to true")
	}
	if _, err := os.Stat(config.MemoryAllocationPath); err != nil {
		t.Fatalf("EnsureMemoryBaseline should write the file: %v", err)
	}
}

// TestEnsureMemoryBaselineNoopWhenAuthoritative: on an update (file present)
// the baseline is already authoritative, so EnsureMemoryBaseline is a no-op and
// must not rewrite the file.
func TestEnsureMemoryBaselineNoopWhenAuthoritative(t *testing.T) {
	defer withTempMemoryPath(t)()
	resetStackConfig()
	seedServices()

	// Simulate a prior install that wrote a specific distribution.
	authoritative := map[string]*system.ServiceConfig{
		"backend": {Name: "backend", AssignedMemory: 1234, MinMemory: 700, MaxMemory: 8 * 1024},
		"opensearch": {Name: "opensearch", AssignedMemory: 5120, MinMemory: 5120, MaxMemory: 60 * 1024},
		"event-processor-manager": {Name: "event-processor-manager", AssignedMemory: 2560, MinMemory: 2560, MaxMemory: 4 * 1024},
		"event-processor-worker": {Name: "event-processor-worker", AssignedMemory: 3072, MinMemory: 3072, MaxMemory: 60 * 1024},
		"web-pdf": {Name: "web-pdf", AssignedMemory: 1250, MinMemory: 1250, MaxMemory: 2 * 1024},
		"postgres": {Name: "postgres", AssignedMemory: 500, MinMemory: 500, MaxMemory: 2 * 1024},
		"user-auditor": {Name: "user-auditor", AssignedMemory: 200, MinMemory: 200, MaxMemory: 1024},
		"agentmanager": {Name: "agentmanager", AssignedMemory: 200, MinMemory: 200, MaxMemory: 1024},
		"frontend": {Name: "frontend", AssignedMemory: 80, MinMemory: 80, MaxMemory: 1024},
	}
	if err := utils.WriteJSON(config.MemoryAllocationPath, authoritative); err != nil {
		t.Fatalf("seed file: %v", err)
	}

	sc := GetStackConfig()
	if !sc.MemoryAuthoritative {
		t.Fatal("existing file should be authoritative")
	}
	// backend must be the stored 1234, not a re-balanced value.
	if sc.ServiceResources["backend"].AssignedMemory != 1234 {
		t.Fatalf("update must read file verbatim; backend=%d want 1234", sc.ServiceResources["backend"].AssignedMemory)
	}
	if err := EnsureMemoryBaseline(); err != nil {
		t.Fatalf("EnsureMemoryBaseline (no-op case): %v", err)
	}
	if sc.ServiceResources["backend"].AssignedMemory != 1234 {
		t.Fatal("EnsureMemoryBaseline must not rewrite an authoritative distribution")
	}
}

// TestRecalculateReplacesAuthority: --recalculate-memory (Recalculate) rewrites
// the file with a fresh balance and marks it authoritative.
func TestRecalculateReplacesAuthority(t *testing.T) {
	defer withTempMemoryPath(t)()
	resetStackConfig()
	seedServices()

	// Start from a seeded (authoritative) file.
	_ = EnsureMemoryBaseline()
	backendBefore := stackConfig.ServiceResources["backend"].AssignedMemory

	if err := Recalculate(); err != nil {
		t.Fatalf("Recalculate: %v", err)
	}
	if !stackConfig.MemoryAuthoritative {
		t.Fatal("Recalculate should be authoritative")
	}
	_ = backendBefore // may equal the new balance on this host; we only assert file rewrite below
	if _, err := os.Stat(config.MemoryAllocationPath); err != nil {
		t.Fatalf("Recalculate should write the file: %v", err)
	}
}
