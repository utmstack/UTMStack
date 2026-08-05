package usecase

import (
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
			_, err := s.Create(lockTenant, Flow{Name: "Isolate host"})
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

	sf, err := s.Create(lockTenant, Flow{Name: "Isolate host"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if err := s.SetEnabled(lockTenant, sf.RelPath, false); err != nil {
		t.Fatalf("SetEnabled: %v", err)
	}

	other := NewFlowStore(s.systemDir, s.userDir)
	if _, err := other.Create(lockTenant, Flow{Name: "Isolate host"}); err == nil {
		t.Error("a second create succeeded over a disabled flow of the same name")
	}
}

// Two tenants may each have a flow of the same name; they are different files.
func TestTwoTenantsMayShareAName(t *testing.T) {
	s := newStore(t)

	if _, err := s.Create(lockTenant, Flow{Name: "Isolate host"}); err != nil {
		t.Fatalf("first tenant: %v", err)
	}
	if _, err := s.Create("another-tenant", Flow{Name: "Isolate host"}); err != nil {
		t.Errorf("second tenant: %v, want it allowed", err)
	}
}

const otherTenantID = "8f1c1b8e-0000-4000-8000-0000000000b0"

// seedShipped puts a flow the product ships into a tenant's own copy of the
// system set, which is where they live now.
func seedShipped(t *testing.T, s *FlowStore, tenant, relPath string, flow Flow) {
	t.Helper()
	target := filepath.Join(s.userDir, tenant, SystemSubdir, relPath)
	if err := writeFlowFile(target, flow); err != nil {
		t.Fatalf("seed: %v", err)
	}
}

// Switching off a flow the product ships is a decision for the tenant that made
// it. Every tenant has its own copy, so the rename touches nobody else.
func TestDisablingASystemFlowIsPerTenant(t *testing.T) {
	s := newStore(t)
	seedShipped(t, s, lockTenant, "isolate-host.yaml", Flow{Name: "Isolate host"})
	seedShipped(t, s, otherTenantID, "isolate-host.yaml", Flow{Name: "Isolate host"})
	if err := s.Load(); err != nil {
		t.Fatalf("Load: %v", err)
	}

	if err := s.SetEnabled(otherTenantID, "isolate-host.yaml", false); err != nil {
		t.Fatalf("SetEnabled: %v", err)
	}

	if sf := s.Get(otherTenantID, "isolate-host.yaml"); sf == nil || sf.Active() {
		t.Error("still on for the tenant that switched it off")
	}
	if sf := s.Get(lockTenant, "isolate-host.yaml"); sf == nil || !sf.Active() {
		t.Error("another tenant lost a flow it never touched")
	}
}

// What a tenant changes about a shipped flow is the suffix, so a release can
// overwrite the content in place and the tenant stays switched off. This is
// what having a copy per tenant buys: an update is a write, not a search.
func TestAnUpdateKeepsTheTenantSwitchedOff(t *testing.T) {
	s := newStore(t)
	seedShipped(t, s, otherTenantID, "isolate-host.yaml", Flow{Name: "Isolate host"})
	if err := s.Load(); err != nil {
		t.Fatalf("Load: %v", err)
	}
	if err := s.SetEnabled(otherTenantID, "isolate-host.yaml", false); err != nil {
		t.Fatalf("SetEnabled: %v", err)
	}

	// What the next release does: replace the file that exists.
	disabled := filepath.Join(s.userDir, otherTenantID, SystemSubdir, "isolate-host.yaml"+DisabledSuffix)
	if err := writeFlowFile(disabled, Flow{Name: "Isolate host", Description: "v2"}); err != nil {
		t.Fatalf("update: %v", err)
	}
	if err := s.Load(); err != nil {
		t.Fatalf("reload: %v", err)
	}

	sf := s.Get(otherTenantID, "isolate-host.yaml")
	if sf == nil {
		t.Fatal("the flow disappeared after the update")
	}
	if sf.Active() {
		t.Error("the update switched the flow back on")
	}
	if sf.Description != "v2" {
		t.Errorf("description = %q, want the updated content to have reached a disabled flow", sf.Description)
	}
}

func TestDisablingAnUnknownFlowFails(t *testing.T) {
	s := newStore(t)

	if err := s.SetEnabled(otherTenantID, "nope.yaml", false); err == nil {
		t.Error("disabling a flow that does not exist succeeded")
	}
}
