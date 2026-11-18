//go:build linux

package watcher

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

type streamPressureSink struct {
	blocked  atomic.Bool
	discard  atomic.Bool
	refused  chan string
	accepted chan string
}

func TestLinuxLiveAdmissionStillReportsKernelOverflow(t *testing.T) {
	data, err := os.ReadFile("/proc/sys/fs/inotify/max_queued_events")
	if err != nil {
		t.Fatal(err)
	}
	capacity, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil || capacity < 1 {
		t.Fatalf("kernel queue capacity %q: %v", data, err)
	}
	if capacity > 65536 {
		t.Skipf("bounded overflow fixture supports at most 65536 kernel events, got %d", capacity)
	}
	root := t.TempDir()
	sink := &streamPressureSink{refused: make(chan string, 1), accepted: make(chan string, 16)}
	w := New([]string{root}, sink, nil, nil)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- w.RunError(ctx) }()
	t.Cleanup(func() {
		cancel()
		select {
		case <-done:
		case <-time.After(time.Second):
			t.Error("watch stream did not stop")
		}
	})
	until := time.Now().Add(3 * time.Second)
	for w.Health() != nil {
		if time.Now().After(until) {
			t.Fatal("watch stream did not start")
		}
		time.Sleep(time.Millisecond)
	}
	sink.blocked.Store(true)
	if err := os.WriteFile(filepath.Join(root, "held.txt"), []byte("owned inert bytes"), 0600); err != nil {
		t.Fatal(err)
	}
	select {
	case <-sink.refused:
	case <-time.After(3 * time.Second):
		t.Fatal("admission did not block")
	}
	// Distinct file names prevent adjacent inotify event coalescing. The
	// blocked reader retains its existing batch, so this exceeds the actual
	// kernel queue without allocating an unbounded application queue.
	for i := 0; i < capacity+128; i++ {
		if err := os.WriteFile(filepath.Join(root, fmt.Sprintf("overflow-%06d", i)), nil, 0600); err != nil {
			t.Fatal(err)
		}
	}
	sink.discard.Store(true)
	sink.blocked.Store(false)
	select {
	case err := <-done:
		done <- err
		if !errors.Is(err, ErrResync) || errors.Is(err, errLinuxAdmission) {
			t.Fatalf("kernel loss hidden: %v", err)
		}
	case <-time.After(15 * time.Second):
		t.Fatal("kernel overflow did not require reconciliation")
	}
	t.Logf("actual_kernel_queue_capacity=%d unique_overflow_files=%d reconciliation_required=true", capacity, capacity+128)
}

func (s *streamPressureSink) Enqueue(path, op string) { s.TryEnqueue(path, op) }
func (s *streamPressureSink) TryEnqueue(path, op string) bool {
	if s.blocked.Load() {
		select {
		case s.refused <- path:
		default:
		}
		return false
	}
	if s.discard.Load() {
		return true
	}
	select {
	case s.accepted <- path:
		return true
	default:
		return false
	}
}

// A user-space admission refusal need not close the kernel watch and trigger
// another full sweep. Keep only the current bounded read batch, retain degraded
// health while blocked, and still honor cancellation. Kernel overflow remains
// a separate loss condition requiring reconciliation.
func TestLinuxLiveAdmissionRetainsStreamAndCancels(t *testing.T) {
	for _, cancelBlocked := range []bool{false, true} {
		t.Run(map[bool]string{false: "resume", true: "cancel"}[cancelBlocked], func(t *testing.T) {
			root := t.TempDir()
			sink := &streamPressureSink{refused: make(chan string, 1), accepted: make(chan string, 16)}
			w := New([]string{root}, sink, nil, nil)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			done := make(chan error, 1)
			go func() { done <- w.RunError(ctx) }()
			t.Cleanup(func() {
				cancel()
				select {
				case <-done:
				case <-time.After(time.Second):
					t.Error("watch stream did not stop")
				}
			})
			until := time.Now().Add(3 * time.Second)
			for w.Health() != nil {
				if time.Now().After(until) {
					t.Fatal("watch stream did not start")
				}
				time.Sleep(time.Millisecond)
			}
			sink.blocked.Store(true)
			path := filepath.Join(root, "retained-inert.txt")
			if err := os.WriteFile(path, []byte("owned inert queued bytes"), 0600); err != nil {
				t.Fatal(err)
			}
			select {
			case <-sink.refused:
			case <-time.After(3 * time.Second):
				t.Fatal("admission pressure not observed")
			}
			until = time.Now().Add(time.Second)
			for w.Health() == nil {
				if time.Now().After(until) {
					t.Fatal("admission gap not visible")
				}
				time.Sleep(time.Millisecond)
			}
			select {
			case err := <-done:
				done <- err
				t.Fatalf("admission reset the live stream: %v", err)
			case <-time.After(100 * time.Millisecond):
			}
			if cancelBlocked {
				cancel()
				select {
				case err := <-done:
					done <- err
					if !errors.Is(err, context.Canceled) {
						t.Fatal(err)
					}
				case <-time.After(time.Second):
					t.Fatal("blocked admission ignored cancellation")
				}
				return
			}
			sink.blocked.Store(false)
			select {
			case got := <-sink.accepted:
				if got != path {
					t.Fatalf("received %q", got)
				}
			case <-time.After(3 * time.Second):
				t.Fatal("retained event not admitted after capacity returned")
			}
			until = time.Now().Add(3 * time.Second)
			for w.Health() != nil {
				if time.Now().After(until) {
					t.Fatalf("drained stream remains degraded: %v", w.Health())
				}
				time.Sleep(time.Millisecond)
			}
		})
	}
}
