package fanotify

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"
	"time"
)

type metadataErrorBackend struct {
	*fakeBackend
	path string
	err  error
}

func (b *metadataErrorBackend) Inspect(raw rawEvent) (Event, error) {
	return Event{Path: b.path, PID: raw.pid, Permission: true}, b.err
}

func TestAmbiguousNameExclusionPreservesProtectionFailurePolicy(t *testing.T) {
	cases := []struct {
		name, path string
		err        error
		excluded   bool
		root       string
	}{
		{"outside", "/outside/file (deleted)", errAmbiguousName, true, ""},
		{"prefix-sibling", "/fixture-sibling/file (deleted)", errAmbiguousName, true, ""},
		{"ancestor-suffix", "/fixture (deleted)/file (deleted)", errAmbiguousName, true, ""},
		{"inside", "/fixture/file (deleted)", errAmbiguousName, false, ""},
		{"missing-path", "", errAmbiguousName, false, ""},
		{"relative", "outside/file (deleted)", errAmbiguousName, false, ""},
		{"noncanonical", "/outside/../fixture/file (deleted)", errAmbiguousName, false, ""},
		{"oversized", "/outside/" + strings.Repeat("x", 4096) + " (deleted)", errAmbiguousName, false, ""},
		{"missing-suffix", "/outside/file", errAmbiguousName, false, ""},
		{"other-error", "/outside/file (deleted)", errors.New("unidentified pinned descriptor"), false, ""},
		{"all-root", "/outside/file (deleted)", errAmbiguousName, false, "/"},
	}
	for _, c := range cases {
		for _, closed := range []bool{false, true} {
			label := c.name + "-open"
			if closed {
				label = c.name + "-closed"
			}
			t.Run(label, func(t *testing.T) {
				b := newFake(t)
				cfg := testConfig()
				cfg.FailClosed = closed
				if c.root != "" {
					cfg.Roots = []string{c.root}
				}
				m, err := New(cfg, func(context.Context, *os.File, Event) (Decision, error) {
					t.Error("unresolved metadata reached scanner")
					return Decision{Clean: true}, nil
				}, nil)
				if err != nil {
					t.Fatal(err)
				}
				m.open = func(Config) (backend, error) { return &metadataErrorBackend{b, c.path, c.err}, nil }
				ctx, cancel := context.WithCancel(context.Background())
				done := make(chan error, 1)
				go func() { done <- m.Run(ctx) }()
				t.Cleanup(func() {
					cancel()
					select {
					case <-done:
					case <-time.After(time.Second):
						t.Error("monitor did not stop")
					}
				})
				b.events <- []rawEvent{{fd: 71, pid: os.Getpid(), mask: OpenPermission}}
				r := nextResponse(t, b)
				if r.allow != (c.excluded || !closed) {
					t.Fatalf("unexpected failure policy: allowed=%t", r.allow)
				}
				if healthy := m.Health() == nil; healthy != c.excluded {
					t.Fatalf("incorrect coverage health: %v", m.Health())
				}
			})
		}
	}
}
