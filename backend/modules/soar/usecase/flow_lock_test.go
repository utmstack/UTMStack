package usecase

import (
	"github.com/utmstack/utmstack/backend/modules/soar/domain"
	"path/filepath"
	"sync"
	"testing"
)

const lockTenant = "8f1c1b8e-0000-4000-8000-000000000001"

func newStore(t *testing.T) *FlowStore {
	t.Helper()
	root := t.TempDir()
	return NewFlowStore(filepath.Join(root, "system"), filepath.Join(root, "user"))
}

// Each replica decides against its own in-memory index, and they share the
// directory. Without a lock other processes see, two creates of the same name
// each miss it in their own index and each write the file, so one silently
// replaces the other.
func TestConcurrentCreatesOfTheSameNameCollideOnce(t *testing.T) {
	root := t.TempDir()

	const replicas = 8
	var wg sync.WaitGroup
	created := make(chan bool, replicas)

	for range replicas {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// A store each: the index is what a second replica does not share.
			s := NewFlowStore(filepath.Join(root, "system"), filepath.Join(root, "user"))
			_, err := s.Create(lockTenant, domain.Flow{Name: "Isolate host"})
			created <- err == nil
		}()
	}
	wg.Wait()
	close(created)

	wins := 0
	for ok := range created {
		if ok {
			wins++
		}
	}
	if wins != 1 {
		t.Fatalf("%d replicas created the same flow, want exactly 1", wins)
	}
}

// A flow the tenant disabled still occupies its name.
func TestADisabledFlowStillCollides(t *testing.T) {
	s := newStore(t)

	sf, err := s.Create(lockTenant, domain.Flow{Name: "Isolate host"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if err := s.SetEnabled(lockTenant, sf.RelPath, false); err != nil {
		t.Fatalf("SetEnabled: %v", err)
	}

	other := NewFlowStore(s.systemDir, s.userDir)
	if _, err := other.Create(lockTenant, domain.Flow{Name: "Isolate host"}); err == nil {
		t.Error("a second create succeeded over a disabled flow of the same name")
	}
}

// Two tenants may each have a flow of the same name; they are different files.
func TestTwoTenantsMayShareAName(t *testing.T) {
	s := newStore(t)

	if _, err := s.Create(lockTenant, domain.Flow{Name: "Isolate host"}); err != nil {
		t.Fatalf("first tenant: %v", err)
	}
	if _, err := s.Create("another-tenant", domain.Flow{Name: "Isolate host"}); err != nil {
		t.Errorf("second tenant: %v, want it allowed", err)
	}
}

const otherTenantID = "8f1c1b8e-0000-4000-8000-0000000000b0"

// seedShipped writes a flow into the directory the product ships. There is one
// copy for the whole install — no tenant has their own.
func seedShipped(t *testing.T, s *FlowStore, relPath string, flow domain.Flow) {
	t.Helper()
	if err := writeFlowFile(filepath.Join(s.systemDir, relPath), flow); err != nil {
		t.Fatalf("seed: %v", err)
	}
}

// Nothing runs until a tenant asks for it. These flows shut machines down, so
// installing the product must not be how that starts.
func TestAShippedFlowIsOffUntilSwitchedOn(t *testing.T) {
	s := newStore(t)
	seedShipped(t, s, "isolate-host.yaml", domain.Flow{Name: "Isolate host"})
	if err := s.Load(); err != nil {
		t.Fatalf("Load: %v", err)
	}

	if sf := s.Get(lockTenant, "isolate-host.yaml"); sf == nil || sf.Active() {
		t.Error("a shipped flow is running before anybody switched it on")
	}
}

// Switching one on is a decision for the tenant that made it. The file is
// shared, so the choice cannot live in the file.
func TestEnablingASystemFlowIsPerTenant(t *testing.T) {
	s := newStore(t)
	seedShipped(t, s, "isolate-host.yaml", domain.Flow{Name: "Isolate host"})
	if err := s.Load(); err != nil {
		t.Fatalf("Load: %v", err)
	}

	if err := s.SetEnabled(otherTenantID, "isolate-host.yaml", true); err != nil {
		t.Fatalf("SetEnabled: %v", err)
	}

	if sf := s.Get(otherTenantID, "isolate-host.yaml"); sf == nil || !sf.Active() {
		t.Error("off for the tenant that switched it on")
	}
	if sf := s.Get(lockTenant, "isolate-host.yaml"); sf == nil || sf.Active() {
		t.Error("another tenant got it switched on too")
	}
}

// A release replaces the one shipped file. The tenant's decision is recorded
// somewhere else, so it survives an update that knows nothing about it.
func TestAnUpdateKeepsTheTenantsChoice(t *testing.T) {
	s := newStore(t)
	seedShipped(t, s, "isolate-host.yaml", domain.Flow{Name: "Isolate host"})
	if err := s.Load(); err != nil {
		t.Fatalf("Load: %v", err)
	}
	if err := s.SetEnabled(otherTenantID, "isolate-host.yaml", true); err != nil {
		t.Fatalf("SetEnabled: %v", err)
	}

	// What the next release does: overwrite what it ships.
	seedShipped(t, s, "isolate-host.yaml", domain.Flow{Name: "Isolate host", Description: "v2"})
	if err := s.Load(); err != nil {
		t.Fatalf("reload: %v", err)
	}

	sf := s.Get(otherTenantID, "isolate-host.yaml")
	if sf == nil {
		t.Fatal("the flow disappeared after the update")
	}
	if !sf.Active() {
		t.Error("the update switched the flow back off")
	}
	if sf.Description != "v2" {
		t.Errorf("description = %q, want the updated content", sf.Description)
	}
}

// A shipped flow is not the tenant's to rewrite or remove: it lives in a
// directory this store never writes to.
func TestAShippedFlowIsReadOnly(t *testing.T) {
	s := newStore(t)
	seedShipped(t, s, "isolate-host.yaml", domain.Flow{Name: "Isolate host"})
	if err := s.Load(); err != nil {
		t.Fatalf("Load: %v", err)
	}

	if _, err := s.Update(otherTenantID, "isolate-host.yaml", domain.Flow{Name: "mine now"}); err == nil {
		t.Error("a tenant rewrote a shipped flow")
	}
	if err := s.Delete(otherTenantID, "isolate-host.yaml"); err == nil {
		t.Error("a tenant deleted a shipped flow")
	}
}

func TestEnablingAnUnknownFlowFails(t *testing.T) {
	s := newStore(t)

	if err := s.SetEnabled(otherTenantID, "nope.yaml", true); err == nil {
		t.Error("switching on a flow that does not exist succeeded")
	}
}
