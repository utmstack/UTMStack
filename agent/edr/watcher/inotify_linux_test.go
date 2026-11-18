//go:build linux

package watcher

import (
	"context"
	"errors"
	"fmt"
	"golang.org/x/sys/unix"
	"os"
	"path/filepath"
	"testing"
	"time"
)

type channelSink struct{ paths chan string }

func (s channelSink) Enqueue(path, op string) { s.paths <- path }
func TestLinuxInotifySafeFixture(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	if err := os.Symlink(outside, filepath.Join(root, "not-followed")); err != nil {
		t.Fatal(err)
	}
	sink := channelSink{make(chan string, 128)}
	w := New([]string{root}, sink, nil, nil)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- w.RunError(ctx) }()
	deadline := time.Now().Add(3 * time.Second)
	for w.Health() != nil {
		if time.Now().After(deadline) {
			t.Fatalf("monitor did not start: %v", w.Health())
		}
		time.Sleep(10 * time.Millisecond)
	}
	sub := filepath.Join(root, "new-dir")
	if err := os.Mkdir(sub, 0700); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(sub, "inert.txt")
	if err := os.WriteFile(target, []byte("ordinary safe fixture"), 0600); err != nil {
		t.Fatal(err)
	}
	timer := time.NewTimer(3 * time.Second)
	defer timer.Stop()
	found := false
	for !found {
		select {
		case p := <-sink.paths:
			found = p == target
		case err := <-done:
			t.Fatalf("monitor failed: %v", err)
		case <-timer.C:
			t.Fatal("new nested file was not observed")
		}
	}
	if err := os.WriteFile(filepath.Join(outside, "external.txt"), []byte("not in scope"), 0600); err != nil {
		t.Fatal(err)
	}
	cancel()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("monitor did not cancel")
	}
}
func TestLinuxInotifyRejectsInvalidScope(t *testing.T) {
	for _, root := range []string{"relative", "/proc", filepath.Join(t.TempDir(), "missing")} {
		w := New([]string{root}, channelSink{make(chan string, 1)}, nil, nil)
		if err := w.RunError(context.Background()); err == nil {
			t.Fatalf("scope %q accepted", root)
		}
	}
}

func TestLinuxReconciliationStreamsWideDirectoryAndCountsExcludedEntries(t *testing.T) {
	root := t.TempDir()
	for i := 0; i < 512; i++ {
		if err := os.WriteFile(filepath.Join(root, fmt.Sprintf("safe-%04d.txt", i)), []byte("inert"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	fd, err := unix.InotifyInit1(unix.IN_NONBLOCK | unix.IN_CLOEXEC)
	if err != nil {
		t.Fatal(err)
	}
	defer unix.Close(fd)
	sink := channelSink{make(chan string, 8)}
	for _, excludeFiles := range []bool{false, true} {
		w := New([]string{root}, sink, nil, func(p string) bool { return excludeFiles && p != root })
		scope := inotifyScope{w: w, fd: fd, paths: map[int32]string{}, entries: maxLinuxReconcileEntries - 4}
		err := scope.addTree(context.Background(), root)
		if !errors.Is(err, ErrWatchLimit) || scope.entries != maxLinuxReconcileEntries {
			t.Fatalf("wide scope was not bounded: entries=%d files=%d err=%v", scope.entries, scope.files, err)
		}
		want := 3
		if excludeFiles {
			want = 0
		}
		if scope.files != want {
			t.Fatalf("regular files enqueued %d want %d", scope.files, want)
		}
	}
	cancelled, cancel := context.WithCancel(context.Background())
	cancel()
	scope := inotifyScope{w: New([]string{root}, sink, nil, nil), fd: fd, paths: map[int32]string{}}
	if err := scope.addTree(cancelled, root); !errors.Is(err, context.Canceled) || scope.entries != 0 {
		t.Fatalf("cancelled traversal did work: %d %v", scope.entries, err)
	}
}
