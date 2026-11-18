//go:build linux

package fanotify

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"golang.org/x/sys/unix"
)

type unlinkBeforeInspect struct {
	backend
	path          string
	device, inode uint64
	once          sync.Once
	err           error
}

func (b *unlinkBeforeInspect) Inspect(raw rawEvent) (Event, error) {
	var st unix.Stat_t
	if err := unix.Fstat(raw.fd, &st); err == nil && uint64(st.Dev) == b.device && st.Ino == b.inode {
		b.once.Do(func() { b.err = os.Remove(b.path) })
		if b.err != nil {
			return Event{}, b.err
		}
	}
	return b.backend.Inspect(raw)
}

// Literal deletion suffixes and real unlink races share Linux's descriptor
// spelling. Their parent directory still establishes exclusion from a configured
// directory scope; inside that scope the ambiguous name must remain incomplete.
func TestNativeAmbiguousDescriptorDirectoryScope(t *testing.T) {
	root := isolatedMount(t)
	for _, unlink := range []bool{false, true} {
		for _, inside := range []bool{false, true} {
			label := "literal"
			if unlink {
				label = "unlinked"
			}
			if inside {
				label += "-inside"
			} else {
				label += "-outside"
			}
			t.Run(label, func(t *testing.T) {
				scope := filepath.Join(root, label, "protected")
				outside := filepath.Join(root, label, "unprotected")
				for _, dir := range []string{scope, outside} {
					if err := os.MkdirAll(dir, 0755); err != nil {
						t.Fatal(err)
					}
				}
				dir := outside
				if inside {
					dir = scope
				}
				name := "owned (deleted)"
				if unlink {
					name = "owned"
				}
				path := filepath.Join(dir, name)
				if err := os.WriteFile(path, []byte("inert scoped descriptor fixture"), 0644); err != nil {
					t.Fatal(err)
				}
				var st unix.Stat_t
				if err := unix.Stat(path, &st); err != nil {
					t.Fatal(err)
				}
				cfg := testConfig()
				cfg.Mounts = []string{root}
				cfg.Roots = []string{scope}
				cfg.Deadline = time.Second
				cfg.FailClosed = true
				var scans atomic.Int32
				m, err := New(cfg, func(context.Context, *os.File, Event) (Decision, error) {
					scans.Add(1)
					return Decision{Clean: true}, nil
				}, nil)
				if err != nil {
					t.Fatal(err)
				}
				m.open = func(cfg Config) (backend, error) {
					b, err := openBackend(cfg)
					if err != nil {
						return nil, err
					}
					if unlink {
						return &unlinkBeforeInspect{backend: b, path: path, device: uint64(st.Dev), inode: st.Ino}, nil
					}
					return b, nil
				}
				ctx, cancel := context.WithCancel(context.Background())
				done := make(chan error, 1)
				go func() { done <- m.Run(ctx) }()
				t.Cleanup(func() {
					cancel()
					select {
					case err := <-done:
						if err != nil && !errors.Is(err, context.Canceled) {
							t.Error(err)
						}
					case <-time.After(2 * time.Second):
						t.Error("owned directory-scope monitor did not stop")
					}
				})
				deadline := time.Now().Add(2 * time.Second)
				for m.Health() != nil {
					if time.Now().After(deadline) {
						t.Fatal(m.Health())
					}
					time.Sleep(time.Millisecond)
				}
				out, err := ownedCommand(t, path, false).CombinedOutput()
				if err != nil {
					t.Fatalf("owned reader: %v %s", err, out)
				}
				want := "ALLOWED"
				if inside {
					want = "DENIED"
				}
				if !strings.Contains(string(out), want) {
					t.Fatalf("scope effect: want %s: %s", want, out)
				}
				if scans.Load() != 0 {
					t.Fatal("ambiguous path reached content scanner")
				}
				if inside {
					if m.Health() == nil {
						t.Fatal("ambiguous protected path reported healthy")
					}
				} else if err := m.Health(); err != nil {
					t.Fatalf("excluded directory degraded protection: %v", err)
				}
			})
		}
	}
}
