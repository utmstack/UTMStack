package repository

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/utmstack/utmstack/backend/modules/integrations/domain"
)

func TestTenantStoreRoundtrip(t *testing.T) {
	dir := t.TempDir()
	s := NewTenantStore(dir)

	want := []domain.Tenant{
		{Name: "t1", Config: map[string]string{"access_key": "AK", "secret": "SK"}},
		{Name: "t2", Config: map[string]string{"access_key": "AK2"}},
	}
	if err := s.Save("AWS_IAM_USER", want); err != nil {
		t.Fatalf("Save: %v", err)
	}
	got, err := s.Load("AWS_IAM_USER")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(got) != len(want) {
		t.Fatalf("len(got)=%d want %d", len(got), len(want))
	}
	for i := range want {
		if got[i].Name != want[i].Name {
			t.Fatalf("[%d] name got %q want %q", i, got[i].Name, want[i].Name)
		}
		for k, v := range want[i].Config {
			if got[i].Config[k] != v {
				t.Fatalf("[%d] %s got %q want %q", i, k, got[i].Config[k], v)
			}
		}
	}

	data, err := os.ReadFile(s.path("AWS_IAM_USER"))
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	txt := string(data)
	if !strings.Contains(txt, "plugins:") || !strings.Contains(txt, "aws:") || !strings.Contains(txt, "tenants:") {
		t.Fatalf("unexpected file layout:\n%s", txt)
	}

	// lockfile must be removed after Save
	if _, err := os.Stat(s.path("AWS_IAM_USER") + ".lock"); !os.IsNotExist(err) {
		t.Fatalf("lockfile still present after Save: err=%v", err)
	}
}

func TestWithFileLockStealsStale(t *testing.T) {
	dir := t.TempDir()
	target := dir + "/f.yaml"
	lock := target + ".lock"

	if err := os.WriteFile(lock, nil, 0o600); err != nil {
		t.Fatalf("preheld: %v", err)
	}
	old := time.Now().Add(-2 * lockStaleness)
	if err := os.Chtimes(lock, old, old); err != nil {
		t.Fatalf("chtimes: %v", err)
	}

	start := time.Now()
	if err := withFileLock(target, func() error { return nil }); err != nil {
		t.Fatalf("withFileLock: %v", err)
	}
	if elapsed := time.Since(start); elapsed > 200*time.Millisecond {
		t.Fatalf("stale lock not stolen fast enough: %v", elapsed)
	}
}

func TestWithFileLockWaits(t *testing.T) {
	dir := t.TempDir()
	target := dir + "/f.yaml"
	lock := target + ".lock"

	// pre-hold the lock
	if err := os.WriteFile(lock, nil, 0o600); err != nil {
		t.Fatalf("preheld: %v", err)
	}

	done := make(chan error, 1)
	go func() {
		done <- withFileLock(target, func() error { return nil })
	}()

	select {
	case <-done:
		t.Fatal("withFileLock returned while lock was held")
	case <-time.After(120 * time.Millisecond):
	}

	// release; now it should complete
	_ = os.Remove(lock)
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("withFileLock: %v", err)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("withFileLock did not return after lock released")
	}
}
