package repository

import (
	"os"
	"strings"
	"syscall"
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
}

func TestWithFileLockWaits(t *testing.T) {
	dir := t.TempDir()
	target := dir + "/f.yaml"

	// pre-hold the flock via a peer fd
	peer, err := os.OpenFile(target+".lock", os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	if err := syscall.Flock(int(peer.Fd()), syscall.LOCK_EX); err != nil {
		t.Fatalf("flock: %v", err)
	}

	done := make(chan error, 1)
	go func() {
		done <- withFileLock(target, func() error { return nil })
	}()

	select {
	case <-done:
		t.Fatal("withFileLock returned while flock was held")
	case <-time.After(120 * time.Millisecond):
	}

	if err := syscall.Flock(int(peer.Fd()), syscall.LOCK_UN); err != nil {
		t.Fatalf("unlock: %v", err)
	}
	_ = peer.Close()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("withFileLock: %v", err)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("withFileLock did not return after flock released")
	}
}

// TestWithFileLockKernelReleasesOnClose proves the crash-recovery property: if
// the peer holding the flock closes its fd (equivalent to process death), the
// kernel releases the lock and a new acquirer proceeds without any TTL/steal.
func TestWithFileLockKernelReleasesOnClose(t *testing.T) {
	dir := t.TempDir()
	target := dir + "/f.yaml"

	peer, err := os.OpenFile(target+".lock", os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	if err := syscall.Flock(int(peer.Fd()), syscall.LOCK_EX); err != nil {
		t.Fatalf("flock: %v", err)
	}
	// simulate crash: close fd without explicit unlock
	_ = peer.Close()

	start := time.Now()
	if err := withFileLock(target, func() error { return nil }); err != nil {
		t.Fatalf("withFileLock: %v", err)
	}
	if elapsed := time.Since(start); elapsed > 100*time.Millisecond {
		t.Fatalf("lock not released on close: %v", elapsed)
	}
}
