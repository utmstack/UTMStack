package usecase

import (
	"os"
	"path/filepath"
	"sync"
	"syscall"
	"testing"
	"time"
)

// The backend runs more than one replica over a shared volume, so two of them
// can save the same framework at once.
func TestATenantMutationWaitsForAnotherProcess(t *testing.T) {
	dir := t.TempDir()

	peer, err := os.OpenFile(filepath.Join(dir, ".compliance.lock"), os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	if err := syscall.Flock(int(peer.Fd()), syscall.LOCK_EX); err != nil {
		t.Fatalf("flock: %v", err)
	}

	done := make(chan error, 1)
	go func() { done <- withTenantLock(dir, func() error { return nil }) }()

	select {
	case <-done:
		t.Fatal("the write went ahead while another process held the lock")
	case <-time.After(120 * time.Millisecond):
	}

	if err := syscall.Flock(int(peer.Fd()), syscall.LOCK_UN); err != nil {
		t.Fatalf("unlock: %v", err)
	}
	_ = peer.Close()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("withTenantLock: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("the write never resumed after the lock was released")
	}
}

// One tenant saving a framework has no reason to wait behind another's.
func TestTenantsDoNotWaitOnEachOther(t *testing.T) {
	root := t.TempDir()
	a := filepath.Join(root, "tenant-a")
	b := filepath.Join(root, "tenant-b")

	held := make(chan struct{})
	release := make(chan struct{})
	go func() {
		_ = withTenantLock(a, func() error {
			close(held)
			<-release
			return nil
		})
	}()
	<-held

	done := make(chan error, 1)
	go func() { done <- withTenantLock(b, func() error { return nil }) }()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("second tenant: %v", err)
		}
	case <-time.After(500 * time.Millisecond):
		t.Error("a tenant waited behind another tenant's lock")
	}
	close(release)
}

// The bug this replaces: a shared "<path>.tmp" let two writers interleave, and
// the rename then published whatever the other one had left there.
func TestConcurrentWritesNeverPublishEachOthersBytes(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "iso27001.yaml")

	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			body := make([]byte, 4096)
			for j := range body {
				body[j] = byte('a' + n)
			}
			_ = atomicWrite(target, body)
		}(i)
	}
	wg.Wait()

	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if len(got) != 4096 {
		t.Fatalf("file is %d bytes, want a whole one — a partial write was published", len(got))
	}
	for i, c := range got {
		if c != got[0] {
			t.Fatalf("byte %d is %q but the file starts with %q — two writers' bytes were mixed", i, c, got[0])
		}
	}
}

// A writer that dies mid-save must not wedge the tenant.
func TestACrashedWriterDoesNotWedgeTheTenant(t *testing.T) {
	dir := t.TempDir()

	peer, err := os.OpenFile(filepath.Join(dir, ".compliance.lock"), os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		t.Fatal(err)
	}
	if err := syscall.Flock(int(peer.Fd()), syscall.LOCK_EX); err != nil {
		t.Fatal(err)
	}
	_ = peer.Close() // process death, no explicit unlock

	start := time.Now()
	if err := withTenantLock(dir, func() error { return nil }); err != nil {
		t.Fatalf("withTenantLock: %v", err)
	}
	if elapsed := time.Since(start); elapsed > 100*time.Millisecond {
		t.Fatalf("waited %v for a lock whose holder was gone", elapsed)
	}
}
