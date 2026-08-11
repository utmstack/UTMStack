package repository

import (
	"os"
	"path/filepath"
	"sync"
	"syscall"
	"testing"
	"time"
)

// The backend runs more than one replica over a shared volume. Two of them
// writing the same overlay is a read-modify-write race, and the in-process
// mutex each one holds says nothing about the other.
func TestAStoreMutationWaitsForAnotherProcess(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, ".rules")

	// A peer process holding the lock, simulated by a separate fd.
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
		t.Fatal("the mutation went ahead while another process held the lock")
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
	case <-time.After(time.Second):
		t.Fatal("the mutation never resumed after the lock was released")
	}
}

// A replica that dies mid-write must not wedge the others: the kernel drops a
// flock when the fd closes, so there is no TTL to wait out and no lock to steal.
func TestACrashedWriterDoesNotWedgeTheStore(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, ".rules")

	peer, err := os.OpenFile(target+".lock", os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	if err := syscall.Flock(int(peer.Fd()), syscall.LOCK_EX); err != nil {
		t.Fatalf("flock: %v", err)
	}
	_ = peer.Close() // process death, no explicit unlock

	start := time.Now()
	if err := withFileLock(target, func() error { return nil }); err != nil {
		t.Fatalf("withFileLock: %v", err)
	}
	if elapsed := time.Since(start); elapsed > 100*time.Millisecond {
		t.Fatalf("waited %v for a lock whose holder was gone", elapsed)
	}
}

// Two writers racing to create the same rule must not both believe they won:
// exactly one file, and the loser told it already exists.
func TestOnlyOneWriterCreatesAGivenPath(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "rule.yaml")
	lock := filepath.Join(dir, ".rules")

	var wg sync.WaitGroup
	created := make(chan bool, 8)
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := withFileLock(lock, func() error {
				if _, err := os.Stat(target); err == nil {
					return os.ErrExist
				}
				return os.WriteFile(target, []byte("x"), 0o644)
			})
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
		t.Errorf("%d writers believed they created the rule, want exactly 1", wins)
	}
}
