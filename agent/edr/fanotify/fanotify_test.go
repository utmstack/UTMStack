package fanotify

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func testConfig() Config {
	return Config{Enabled: true, Mounts: []string{"/"}, Roots: []string{"/fixture"}, Mode: ModeOpen, Deadline: 60 * time.Millisecond, Workers: 1, MaxPending: 2, MaxFileBytes: 1024}
}

type response struct {
	fd    int
	allow bool
}
type fakeBackend struct {
	events      chan []rawEvent
	responses   chan response
	file        *os.File
	mu          sync.Mutex
	released    map[int]int
	closed      bool
	changed     bool
	responseErr error
}

func newFake(t *testing.T) *fakeBackend {
	t.Helper()
	p := filepath.Join(t.TempDir(), "inert")
	if e := os.WriteFile(p, []byte("ordinary inert bytes"), 0600); e != nil {
		t.Fatal(e)
	}
	f, e := os.Open(p)
	if e != nil {
		t.Fatal(e)
	}
	t.Cleanup(func() { f.Close() })
	return &fakeBackend{events: make(chan []rawEvent, 8), responses: make(chan response, 64), file: f, released: map[int]int{}}
}
func (b *fakeBackend) Read(ctx context.Context) ([]rawEvent, error) {
	select {
	case e := <-b.events:
		return e, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}
func (b *fakeBackend) Inspect(raw rawEvent) (Event, error) {
	return Event{Path: "/fixture/inert", PID: raw.pid, Size: 20, Mask: raw.mask}, nil
}
func (b *fakeBackend) Attribute(raw rawEvent, e *Event) error { e.OSIdentity = 123; return nil }
func (b *fakeBackend) Unchanged(raw rawEvent, e Event) error {
	if b.changed {
		return errors.New("fixture changed")
	}
	return nil
}
func (b *fakeBackend) Duplicate(raw rawEvent) (*os.File, error) {
	return duplicateFixtureFile(b.file)
}
func (b *fakeBackend) Respond(raw rawEvent, allow bool) error {
	b.responses <- response{raw.fd, allow}
	return b.responseErr
}
func (b *fakeBackend) Release(raw rawEvent) { b.mu.Lock(); defer b.mu.Unlock(); b.released[raw.fd]++ }
func (b *fakeBackend) Close() error         { b.mu.Lock(); defer b.mu.Unlock(); b.closed = true; return nil }
func fakeMonitor(t *testing.T, cfg Config, b *fakeBackend, scan Scanner, observe func(Outcome)) (*Monitor, context.CancelFunc, chan error) {
	t.Helper()
	m, e := New(cfg, scan, observe)
	if e != nil {
		t.Fatal(e)
	}
	m.open = func(Config) (backend, error) { return b, nil }
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- m.Run(ctx) }()
	t.Cleanup(func() {
		cancel()
		select {
		case <-done:
		case <-time.After(time.Second):
			t.Error("monitor failed to stop")
		}
	})
	return m, cancel, done
}
func nextResponse(t *testing.T, b *fakeBackend) response {
	t.Helper()
	select {
	case r := <-b.responses:
		return r
	case <-time.After(time.Second):
		t.Fatal("permission request left unresolved")
		return response{}
	}
}
func TestPermissionUsesPinnedDescriptorAndDeniesConfirmedDetection(t *testing.T) {
	b := newFake(t)
	scan := func(ctx context.Context, f *os.File, e Event) (Decision, error) {
		data, err := io.ReadAll(f)
		if err != nil || string(data) != "ordinary inert bytes" {
			t.Errorf("wrong pinned content: %q %v", data, err)
		}
		return Decision{Detected: true, Signature: "Safe.Marker"}, nil
	}
	_, _, _ = fakeMonitor(t, testConfig(), b, scan, nil)
	b.events <- []rawEvent{{fd: 17, pid: 99999, mask: OpenPermission}}
	if r := nextResponse(t, b); r.fd != 17 || r.allow {
		t.Fatalf("malicious request allowed: %+v", r)
	}
}
func TestUncooperativeScanCannotHoldPermissionPastDeadline(t *testing.T) {
	for _, closed := range []bool{false, true} {
		t.Run(fmt.Sprint(closed), func(t *testing.T) {
			b := newFake(t)
			release := make(chan struct{})
			defer close(release)
			entered := make(chan struct{})
			cfg := testConfig()
			cfg.FailClosed = closed
			m, _, _ := fakeMonitor(t, cfg, b, func(context.Context, *os.File, Event) (Decision, error) {
				close(entered)
				<-release
				return Decision{Clean: true}, nil
			}, nil)
			b.events <- []rawEvent{{fd: 21, pid: 99999, mask: OpenPermission}}
			<-entered
			start := time.Now()
			r := nextResponse(t, b)
			if time.Since(start) > 500*time.Millisecond || r.allow == closed {
				t.Fatalf("deadline/fail policy violated: %+v", r)
			}
			if m.Health() == nil || m.Stats().Timeouts != 1 {
				t.Fatal("timeout lost from health")
			}
		})
	}
}
func TestBoundedAdmissionResolvesOverflowAndShutdown(t *testing.T) {
	b := newFake(t)
	release := make(chan struct{})
	defer close(release)
	cfg := testConfig()
	cfg.Deadline = time.Second
	cfg.FailClosed = true
	m, cancel, _ := fakeMonitor(t, cfg, b, func(context.Context, *os.File, Event) (Decision, error) { <-release; return Decision{Clean: true}, nil }, nil)
	b.events <- []rawEvent{{fd: 31, pid: 99999, mask: OpenPermission}, {fd: 32, pid: 99999, mask: OpenPermission}, {fd: 33, pid: 99999, mask: OpenPermission}}
	if r := nextResponse(t, b); r.fd != 33 || r.allow {
		t.Fatalf("unresolved admission overflow: %+v", r)
	}
	cancel()
	a, c := nextResponse(t, b), nextResponse(t, b)
	if a.allow || c.allow || a.fd == c.fd {
		t.Fatal("shutdown did not resolve owned requests")
	}
	if m.Health() == nil || m.Stats().Overflows < 1 {
		t.Fatal("admission loss hidden")
	}
	if err := m.Run(context.Background()); !errors.Is(err, ErrUsed) {
		t.Fatalf("restart could create another worker pool: %v", err)
	}
}
func TestChangedFileNeverEarnsCleanPermission(t *testing.T) {
	b := newFake(t)
	b.changed = true
	cfg := testConfig()
	cfg.FailClosed = true
	fakeMonitor(t, cfg, b, func(context.Context, *os.File, Event) (Decision, error) { return Decision{Clean: true}, nil }, nil)
	b.events <- []rawEvent{{fd: 41, pid: 99999, mask: OpenPermission}}
	if r := nextResponse(t, b); r.allow {
		t.Fatal("mutated file declared clean")
	}
}
func TestQueueOverflowClosesGroupAndReleasesBatch(t *testing.T) {
	b := newFake(t)
	m, e := New(testConfig(), func(context.Context, *os.File, Event) (Decision, error) { return Decision{Clean: true}, nil }, nil)
	if e != nil {
		t.Fatal(e)
	}
	m.open = func(Config) (backend, error) { return b, nil }
	b.events <- []rawEvent{{fd: -1, pid: -1, mask: QueueOverflow}, {fd: 51, pid: 99999, mask: OpenPermission}}
	if err := m.Run(context.Background()); !errors.Is(err, ErrOverflow) {
		t.Fatal(err)
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	if !b.closed || b.released[51] != 1 {
		t.Fatal("overflow leaked already-read descriptors")
	}
}
func TestPermissionReplyFailureClosesGroup(t *testing.T) {
	b := newFake(t)
	b.responseErr = errors.New("injected reply failure")
	m, e := New(testConfig(), func(context.Context, *os.File, Event) (Decision, error) { return Decision{Clean: true}, nil }, nil)
	if e != nil {
		t.Fatal(e)
	}
	m.open = func(Config) (backend, error) { return b, nil }
	b.events <- []rawEvent{{fd: 61, pid: 99999, mask: OpenPermission}}
	if err := m.Run(context.Background()); err == nil {
		t.Fatal("reply failure hidden")
	}
	if !b.closed {
		t.Fatal("failed group not closed")
	}
}
func TestConfigRequiresExplicitBoundedOptIn(t *testing.T) {
	cfg := testConfig()
	if e := Validate(cfg); e != nil {
		t.Fatalf("Linux POSIX paths must validate on every host: %v", e)
	}
	cfg.Enabled = false
	if _, e := New(cfg, nil, nil); !errors.Is(e, ErrDisabled) {
		t.Fatal(e)
	}
	for _, change := range []func(*Config){func(c *Config) { c.Roots = []string{"relative"} }, func(c *Config) { c.Roots = []string{"/proc"} }, func(c *Config) { c.Workers = 9 }, func(c *Config) { c.MaxPending = 65 }, func(c *Config) { c.Deadline = 6 * time.Second }, func(c *Config) { c.Mounts = []string{"/else"} }} {
		cfg = testConfig()
		change(&cfg)
		if e := Validate(cfg); e == nil {
			t.Fatal("unsafe config accepted")
		}
	}
}
func metadataFixture() []byte {
	b := make([]byte, 32)
	binary.NativeEndian.PutUint32(b, 32)
	b[4] = 3
	binary.NativeEndian.PutUint16(b[6:], 24)
	binary.NativeEndian.PutUint64(b[8:], OpenPermission)
	binary.NativeEndian.PutUint32(b[16:], 17)
	binary.NativeEndian.PutUint32(b[20:], 42)
	b[24] = 4
	binary.NativeEndian.PutUint16(b[26:], 8)
	binary.NativeEndian.PutUint32(b[28:], 18)
	return b
}
func TestMetadataDecoderBoundaries(t *testing.T) {
	b := metadataFixture()
	events, e := decodeEvents(b)
	if e != nil || len(events) != 1 || events[0].fd != 17 || events[0].pidfd != 18 || events[0].pid != 42 {
		t.Fatalf("decoded=%+v err=%v", events, e)
	}
	for _, bad := range [][]byte{b[:23], b[:31], append([]byte{}, b...)} {
		if len(bad) == 32 {
			bad[4] = 99
		}
		if _, e := decodeEvents(bad); e == nil {
			t.Fatal("bad metadata accepted")
		}
	}
}
func FuzzFanotifyMetadata(f *testing.F) {
	f.Add(metadataFixture())
	f.Add([]byte("short"))
	f.Fuzz(func(t *testing.T, b []byte) {
		if len(b) > 8192 {
			t.Skip()
		}
		_, _ = decodeEvents(b)
	})
}

func TestLateDetectionIsReportedWithoutASecondPermissionReply(t *testing.T) {
	b := newFake(t)
	release := make(chan struct{})
	observed := make(chan Outcome, 4)
	m, _, _ := fakeMonitor(t, testConfig(), b, func(context.Context, *os.File, Event) (Decision, error) {
		<-release
		return Decision{Detected: true, Signature: "Late.Safe.Marker"}, nil
	}, func(o Outcome) { observed <- o })
	b.events <- []rawEvent{{fd: 71, pid: 99999, mask: OpenPermission}}
	if r := nextResponse(t, b); !r.allow {
		t.Fatal("test requires fail-open deadline")
	}
	close(release)
	timer := time.After(time.Second)
	for {
		select {
		case o := <-observed:
			if o.Reason == "late_scan" {
				if !o.Decision.Detected || !o.Allowed {
					t.Fatal("late result misstates completed permission")
				}
				if m.Stats().Resolved != 1 {
					t.Fatal("permission resolved twice")
				}
				return
			}
		case <-timer:
			t.Fatal("late detection silently discarded")
		}
	}
}
func TestIncompleteDecisionCannotBecomeCleanWithoutError(t *testing.T) {
	b := newFake(t)
	cfg := testConfig()
	cfg.FailClosed = true
	m, _, _ := fakeMonitor(t, cfg, b, func(context.Context, *os.File, Event) (Decision, error) { return Decision{}, nil }, nil)
	b.events <- []rawEvent{{fd: 81, pid: 99999, mask: OpenPermission}}
	if r := nextResponse(t, b); r.allow {
		t.Fatal("missing scanner verdict allowed as clean")
	}
	if !errors.Is(m.Health(), ErrIncomplete) {
		t.Fatal(m.Health())
	}
}
func TestNotificationDetectionNeverClaimsBlocking(t *testing.T) {
	b := newFake(t)
	cfg := testConfig()
	cfg.Mode = ModeNotify
	cfg.FailClosed = true
	out := make(chan Outcome, 1)
	fakeMonitor(t, cfg, b, func(context.Context, *os.File, Event) (Decision, error) {
		return Decision{Detected: true, Signature: "Safe.Marker"}, nil
	}, func(o Outcome) { out <- o })
	b.events <- []rawEvent{{fd: 91, pid: 99999, mask: Modify}}
	select {
	case o := <-out:
		if !o.Allowed || !o.Decision.Detected || o.Event.Permission {
			t.Fatalf("notification falsely claims blocking: %+v", o)
		}
	case <-time.After(time.Second):
		t.Fatal("notification not delivered")
	}
	if len(b.responses) != 0 {
		t.Fatal("notification wrote a permission response")
	}
}
