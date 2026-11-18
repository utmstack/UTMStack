//go:build linux

package fanotify

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"golang.org/x/sys/unix"
)

func brokerPair(t *testing.T) (*os.File, *os.File) {
	t.Helper()
	fds, err := unix.Socketpair(unix.AF_UNIX, unix.SOCK_SEQPACKET|unix.SOCK_CLOEXEC, 0)
	if err != nil {
		t.Fatal(err)
	}
	a, b := os.NewFile(uintptr(fds[0]), "broker-parent"), os.NewFile(uintptr(fds[1]), "broker-owner")
	t.Cleanup(func() { a.Close(); b.Close() })
	return a, b
}

func TestBrokerDescriptorAndBytePath(t *testing.T) {
	a, b := brokerPair(t)
	s, err := brokerSocketFromFile(a)
	if err != nil {
		t.Fatal(err)
	}
	defer s.close()
	r, err := brokerSocketFromFile(b)
	if err != nil {
		t.Fatal(err)
	}
	defer r.close()
	path := t.TempDir() + "/original"
	if err := os.WriteFile(path, []byte("inert original"), 0600); err != nil {
		t.Fatal(err)
	}
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	want := Event{Path: "/fixture/\xff\xfe", Permission: true, Device: 9, Inode: 42}
	if err = s.send(ctx, brokerMessage{Kind: "scan", ID: 4, Event: want}, f); err != nil {
		t.Fatal(err)
	}
	if err = os.Rename(path, path+".old"); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(path, []byte("replacement"), 0600); err != nil {
		t.Fatal(err)
	}
	m, pinned, err := r.receive(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer pinned.Close()
	if m.Event != want || m.ID != 4 {
		t.Fatalf("metadata changed: %+v", m)
	}
	data, err := io.ReadAll(pinned)
	if err != nil || string(data) != "inert original" {
		t.Fatalf("reopened replacement: %q %v", data, err)
	}
	flags, err := unix.FcntlInt(pinned.Fd(), unix.F_GETFD, 0)
	if err != nil || flags&unix.FD_CLOEXEC == 0 {
		t.Fatal("received descriptor must be close-on-exec")
	}
}

func TestBrokerRejectsMalformedPacketsWithoutLeakingRights(t *testing.T) {
	cases := []struct {
		name  string
		data  []byte
		count int
	}{
		{"json", []byte("{"), 1},
		{"trailing", []byte(`{"version":1} {}`), 1},
		{"version", []byte(`{"version":2}`), 1},
		{"unknown", []byte(`{"version":1,"unknown":true}`), 1},
		{"multiple-rights", []byte(`{"version":1}`), 3},
		{"truncated-rights", []byte(`{"version":1}`), 100},
		{"oversized", bytes.Repeat([]byte("x"), brokerPacketLimit+1), 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			a, b := brokerPair(t)
			r, err := brokerSocketFromFile(b)
			if err != nil {
				t.Fatal(err)
			}
			defer r.close()
			f, err := os.Open("/dev/null")
			if err != nil {
				t.Fatal(err)
			}
			defer f.Close()
			rights := make([]int, tc.count)
			for i := range rights {
				rights[i] = int(f.Fd())
			}
			before, err := os.ReadDir("/proc/self/fd")
			if err != nil {
				t.Fatal(err)
			}
			if err = unix.Sendmsg(int(a.Fd()), tc.data, unix.UnixRights(rights...), nil, unix.MSG_NOSIGNAL); err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			_, received, err := r.receive(ctx)
			if received != nil {
				received.Close()
				t.Fatal("malformed message returned authority")
			}
			if err == nil {
				t.Fatal("malformed packet accepted")
			}
			after, err := os.ReadDir("/proc/self/fd")
			if err != nil {
				t.Fatal(err)
			}
			if len(before) != len(after) {
				t.Fatalf("descriptor leak: %d -> %d", len(before), len(after))
			}
		})
	}
}

func TestBrokerSendBoundIncludesContendedWriter(t *testing.T) {
	a, _ := brokerPair(t)
	s, err := brokerSocketFromFile(a)
	if err != nil {
		t.Fatal(err)
	}
	defer s.close()
	if err = unix.SetsockoptInt(s.fd, unix.SOL_SOCKET, unix.SO_SNDBUF, 4096); err != nil {
		t.Fatal(err)
	}
	// Fill the socket without receiving, then hold one writer behind backpressure.
	for {
		err = unix.Sendmsg(s.fd, bytes.Repeat([]byte("x"), 2048), nil, nil, unix.MSG_DONTWAIT|unix.MSG_NOSIGNAL)
		if errors.Is(err, unix.EAGAIN) {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	first := make(chan error, 1)
	go func() { first <- s.send(ctx, brokerMessage{Kind: "status"}, nil) }()
	time.Sleep(20 * time.Millisecond)
	bounded, stop := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer stop()
	started := time.Now()
	err = s.send(bounded, brokerMessage{Kind: "status"}, nil)
	if !errors.Is(err, context.DeadlineExceeded) || time.Since(started) > 250*time.Millisecond {
		t.Fatalf("writer ignored own deadline: %v", err)
	}
	cancel()
	select {
	case <-first:
	case <-time.After(time.Second):
		t.Fatal("stopped writer did not exit")
	}
}

func TestBrokerStatusBackpressureKeepsPermissionOwner(t *testing.T) {
	a, b := brokerPair(t)
	owner, err := brokerSocketFromFile(b)
	if err != nil {
		t.Fatal(err)
	}
	defer owner.close()
	client, err := brokerSocketFromFile(a)
	if err != nil {
		t.Fatal(err)
	}
	defer client.close()
	if err = unix.SetsockoptInt(owner.fd, unix.SOL_SOCKET, unix.SO_SNDBUF, 4096); err != nil {
		t.Fatal(err)
	}
	cfg := testConfig()
	cfg.FailClosed = true
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	fake := newFake(t)
	opened := make(chan struct{})
	done := make(chan error, 1)
	go func() {
		done <- serveBroker(ctx, owner, func(Config) (backend, error) { close(opened); return fake, nil })
	}()
	if err = client.send(ctx, brokerMessage{Kind: "configure", Config: &cfg}, nil); err != nil {
		t.Fatal(err)
	}
	<-opened
	for {
		err = unix.Sendmsg(owner.fd, bytes.Repeat([]byte("x"), 2048), nil, nil, unix.MSG_DONTWAIT|unix.MSG_NOSIGNAL)
		if errors.Is(err, unix.EAGAIN) {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
	}
	select {
	case err := <-done:
		t.Fatalf("telemetry backpressure dropped owner: %v", err)
	case <-time.After(300 * time.Millisecond):
	}
	fake.events <- []rawEvent{{fd: 7, pid: 999, mask: OpenPermission}}
	select {
	case r := <-fake.responses:
		if r.allow {
			t.Fatal("backpressure escaped fail-closed permission deadline")
		}
	case <-ctx.Done():
		t.Fatal("backpressure blocked permission owner")
	}
	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("owner did not stop")
	}
}

func TestBrokerClientAllowsBoundedOwnerStartup(t *testing.T) {
	a, b := brokerPair(t)
	owner, err := brokerSocketFromFile(b)
	if err != nil {
		t.Fatal(err)
	}
	defer owner.close()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	ready := make(chan error, 1)
	done := make(chan error, 1)
	go func() {
		done <- RunBrokerClient(ctx, a, testConfig(), func(context.Context, *os.File, Event) (Decision, error) { return Decision{Clean: true}, nil }, nil, func(_ Stats, e error) { ready <- e })
	}()
	_, f, err := owner.receive(ctx)
	if f != nil {
		f.Close()
	}
	if err != nil {
		t.Fatal(err)
	}
	select {
	case err := <-done:
		t.Fatalf("owner startup expired before startup budget: %v", err)
	case <-time.After(1100 * time.Millisecond):
	}
	if err = owner.send(ctx, brokerMessage{Kind: "status"}, nil); err != nil {
		t.Fatal(err)
	}
	select {
	case e := <-ready:
		if e != nil {
			t.Fatal(e)
		}
	case <-ctx.Done():
		t.Fatal("owner readiness lost")
	}
	cancel()
	<-done
}

func TestBrokerDecisionsAndIndependentDeadline(t *testing.T) {
	for _, failClosed := range []bool{false, true} {
		for _, kind := range []string{"clean", "detected", "incomplete", "panic", "stalled", "changed"} {
			t.Run(fmt.Sprintf("%t/%s", failClosed, kind), func(t *testing.T) {
				a, b := brokerPair(t)
				s, err := brokerSocketFromFile(b)
				if err != nil {
					t.Fatal(err)
				}
				defer s.close()
				cfg := testConfig()
				cfg.FailClosed = failClosed
				cfg.Deadline = 100 * time.Millisecond
				fake := newFake(t)
				fake.changed = kind == "changed"
				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				defer cancel()
				server := make(chan error, 1)
				go func() { server <- serveBroker(ctx, s, func(Config) (backend, error) { return fake, nil }) }()
				outcomes := make(chan Outcome, 8)
				client := make(chan error, 1)
				go func() {
					client <- RunBrokerClient(ctx, a, cfg, func(ctx context.Context, f *os.File, e Event) (Decision, error) {
						data, err := io.ReadAll(f)
						if err != nil || string(data) != "ordinary inert bytes" {
							return Decision{}, fmt.Errorf("bad pinned data: %q %v", data, err)
						}
						switch kind {
						case "detected":
							return Decision{Detected: true, Signature: "Safe.Broker.Marker", Generation: "pinned-generation"}, nil
						case "incomplete":
							return Decision{}, ErrIncomplete
						case "panic":
							panic("owned test panic")
						case "stalled":
							<-ctx.Done()
							return Decision{}, ctx.Err()
						}
						return Decision{Clean: true, Generation: "pinned-generation"}, nil
					}, func(o Outcome) { outcomes <- o }, nil)
				}()
				fake.events <- []rawEvent{{fd: 7, pid: 999, mask: OpenPermission}}
				select {
				case response := <-fake.responses:
					want := kind == "clean" || kind != "detected" && !failClosed
					if response.allow != want {
						t.Fatalf("decision: got %t want %t", response.allow, want)
					}
				case <-ctx.Done():
					t.Fatal("broker did not resolve permission")
				}
				select {
				case o := <-outcomes:
					if kind == "detected" && (o.Decision.Generation != "pinned-generation" || !o.Decision.Detected) {
						t.Fatalf("lost generation or detection: %+v", o)
					}
				case <-ctx.Done():
					t.Fatal("missing outcome")
				}
				cancel()
				select {
				case <-server:
				case <-time.After(time.Second):
					t.Fatal("broker did not stop")
				}
				select {
				case <-client:
				case <-time.After(time.Second):
					t.Fatal("client did not stop")
				}
			})
		}
	}
}

func TestBrokerRejectsInvalidSetupAndPeer(t *testing.T) {
	f, err := os.Open("/dev/null")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err = ServeBroker(context.Background(), f); err == nil {
		t.Fatal("accepted non-socket")
	}
	a, b := brokerPair(t)
	if err = ServeBroker(context.Background(), b); err == nil {
		t.Fatal("accepted socket from own process as parent channel")
	}
	s, err := brokerSocketFromFile(b)
	if err != nil {
		t.Fatal(err)
	}
	defer s.close()
	p, err := brokerSocketFromFile(a)
	if err != nil {
		t.Fatal(err)
	}
	defer p.close()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err = p.send(ctx, brokerMessage{Kind: "decision", ID: 1}, nil); err != nil {
		t.Fatal(err)
	}
	if err = serveBroker(ctx, s, func(Config) (backend, error) { t.Fatal("activated before valid setup"); return nil, nil }); err == nil {
		t.Fatal("accepted invalid setup")
	}
}

func TestBrokerConfigAndHealthBounds(t *testing.T) {
	cfg := testConfig()
	cfg.Roots = []string{"/fixture/\xff"}
	if _, err := NewPermissionBroker(cfg, func(context.Context, *os.File, Event) (Decision, error) { return Decision{}, nil }, nil); err == nil {
		t.Fatal("scope encoding silently changed byte path")
	}
	cfg = testConfig()
	for i := 0; i < 8; i++ {
		cfg.Roots = append(cfg.Roots, "/"+strings.Repeat("x", 4000))
	}
	if _, err := NewPermissionBroker(cfg, func(context.Context, *os.File, Event) (Decision, error) { return Decision{}, nil }, nil); err == nil {
		t.Fatal("unbounded configuration admitted")
	}
	b, err := NewPermissionBroker(testConfig(), func(context.Context, *os.File, Event) (Decision, error) { return Decision{Clean: true}, nil }, nil)
	if err != nil {
		t.Fatal(err)
	}
	if b.Health() == nil {
		t.Fatal("unstarted broker claims health")
	}
	b.running = true
	b.lastStatus = time.Now().Add(-2 * time.Second)
	if b.Health() == nil {
		t.Fatal("stale status claims health")
	}
	b.lastStatus = time.Now()
	if b.Health() != nil {
		t.Fatal("fresh ready status rejected")
	}
	b.failure = ErrIncomplete
	if !errors.Is(b.Health(), ErrIncomplete) {
		t.Fatal("degraded status lost")
	}
}

func TestBrokerLateDetectionHasNoSecondKernelResponse(t *testing.T) {
	a, b := brokerPair(t)
	s, err := brokerSocketFromFile(b)
	if err != nil {
		t.Fatal(err)
	}
	defer s.close()
	cfg := testConfig()
	// A late result requires a scan that actually started before expiry.
	// Leave dispatch headroom when the full suite shares a CPU quota; the
	// callback stays blocked until the owner has enforced its real deadline.
	cfg.Deadline = time.Second
	cfg.FailClosed = true
	fake := newFake(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	server := make(chan error, 1)
	go func() { server <- serveBroker(ctx, s, func(Config) (backend, error) { return fake, nil }) }()
	outcomes := make(chan Outcome, 8)
	started := make(chan struct{})
	release := make(chan struct{})
	defer func() {
		select {
		case <-release:
		default:
			close(release)
		}
	}()
	client := make(chan error, 1)
	go func() {
		client <- RunBrokerClient(ctx, a, cfg, func(context.Context, *os.File, Event) (Decision, error) {
			close(started)
			<-release
			return Decision{Detected: true, Signature: "Safe.Late.Broker", Generation: "late-generation"}, nil
		}, func(o Outcome) { outcomes <- o }, nil)
	}()
	fake.events <- []rawEvent{{fd: 7, pid: 999, mask: OpenPermission}}
	select {
	case <-started:
	case <-ctx.Done():
		t.Fatal("scanner did not start before the permission deadline")
	}
	select {
	case r := <-fake.responses:
		if r.allow {
			t.Fatal("deadline escaped fail-closed")
		}
	case <-ctx.Done():
		t.Fatal("missing deadline response")
	}
	close(release)
	for {
		select {
		case o := <-outcomes:
			if o.Reason != "late_scan" {
				continue
			}
			if !o.Decision.Detected || o.Allowed || o.Decision.Generation != "late-generation" {
				t.Fatalf("lost late result: %+v", o)
			}
			goto observed
		case <-ctx.Done():
			t.Fatal("late detection lost")
		}
	}
observed:
	select {
	case <-fake.responses:
		t.Fatal("late result responded twice")
	case <-time.After(50 * time.Millisecond):
	}
	cancel()
	<-client
	<-server
}

func TestBrokerExpiredRequestSkipsUnstartedScanner(t *testing.T) {
	a, b := brokerPair(t)
	s, err := brokerSocketFromFile(b)
	if err != nil {
		t.Fatal(err)
	}
	defer s.close()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	called := make(chan int, 2)
	client := make(chan error, 1)
	cfg := testConfig()
	cfg.Deadline = time.Second
	go func() {
		client <- RunBrokerClient(ctx, a, cfg, func(_ context.Context, _ *os.File, ev Event) (Decision, error) {
			called <- ev.PID
			return Decision{Clean: true, Generation: "fresh"}, nil
		}, nil, nil)
	}()
	configure, f, err := s.receive(ctx)
	if f != nil {
		f.Close()
	}
	if err != nil || configure.Kind != "configure" || f != nil {
		t.Fatalf("client setup: %+v %v", configure, err)
	}
	// The last writer's closure proves that the worker finished handling the
	// request, without relying on a sleep or on invoking the scanner itself.
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	defer writer.Close()
	if err := s.send(ctx, brokerMessage{Kind: "scan", ID: 1, Event: Event{PID: 1, Permission: true}, Deadline: time.Now().Add(-time.Second).UnixNano()}, writer); err != nil {
		t.Fatal(err)
	}
	writer.Close()
	if err := reader.SetReadDeadline(time.Now().Add(2 * time.Second)); err != nil {
		t.Fatal(err)
	}
	var buf [1]byte
	if n, err := reader.Read(buf[:]); n != 0 || !errors.Is(err, io.EOF) {
		t.Fatalf("expired request retained its descriptor: n=%d err=%v", n, err)
	}
	expired, f, err := s.receive(ctx)
	if f != nil {
		f.Close()
	}
	if err != nil || f != nil || expired.Kind != "decision" || expired.ID != 1 || expired.Decision != (Decision{}) || expired.Error != context.DeadlineExceeded.Error() {
		t.Fatalf("expired request did not release owner reply slot: %+v %v", expired, err)
	}
	fresh, err := os.CreateTemp(t.TempDir(), "pinned")
	if err != nil {
		t.Fatal(err)
	}
	defer fresh.Close()
	if err := s.send(ctx, brokerMessage{Kind: "scan", ID: 2, Event: Event{PID: 2, Permission: true}, Deadline: time.Now().Add(cfg.Deadline).UnixNano()}, fresh); err != nil {
		t.Fatal(err)
	}
	decision, f, err := s.receive(ctx)
	if f != nil {
		f.Close()
	}
	if err != nil || f != nil || decision.Kind != "decision" || decision.ID != 2 || !decision.Decision.Clean || decision.Decision.Generation != "fresh" || decision.Error != "" {
		t.Fatalf("fresh request after expiry: %+v %v", decision, err)
	}
	if pid := <-called; pid != 2 {
		t.Fatalf("already expired request started scanner for pid%d", pid)
	}
	select {
	case pid := <-called:
		t.Fatalf("unexpected extra scanner call for pid%d", pid)
	default:
	}
	cancel()
	<-client
}

// A raw scanner peer deliberately sends its decision after the kernel response.
// This exercises the owner's late-result path without a client-local callback
// that could hide a dropped transport completion.
func TestBrokerOwnerReportsCompletionAfterDeadline(t *testing.T) {
	a, b := brokerPair(t)
	peer, err := brokerSocketFromFile(a)
	if err != nil {
		t.Fatal(err)
	}
	defer peer.close()
	owner, err := brokerSocketFromFile(b)
	if err != nil {
		t.Fatal(err)
	}
	defer owner.close()
	cfg := testConfig()
	cfg.Deadline = time.Second
	cfg.FailClosed = true
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	fake := newFake(t)
	server := make(chan error, 1)
	go func() { server <- serveBroker(ctx, owner, func(Config) (backend, error) { return fake, nil }) }()
	if err := peer.send(ctx, brokerMessage{Kind: "configure", Config: &cfg}, nil); err != nil {
		t.Fatal(err)
	}
	fake.events <- []rawEvent{{fd: 7, pid: 999, mask: OpenPermission}}
	var request brokerMessage
	for {
		m, f, err := peer.receive(ctx)
		if f != nil {
			f.Close()
		}
		if err != nil {
			t.Fatal(err)
		}
		if m.Kind == "scan" {
			request = m
			break
		}
	}
	select {
	case r := <-fake.responses:
		if r.allow {
			t.Fatal("deadline escaped fail-closed")
		}
	case <-ctx.Done():
		t.Fatal("owner did not resolve kernel deadline")
	}
	want := Decision{Detected: true, Signature: "Safe.Late.Owner", Generation: "owner-generation"}
	if err := peer.send(ctx, brokerMessage{Kind: "decision", ID: request.ID, Decision: want}, nil); err != nil {
		t.Fatal(err)
	}
	for {
		m, f, err := peer.receive(ctx)
		if f != nil {
			f.Close()
		}
		if err != nil {
			t.Fatalf("late completion lost: %v", err)
		}
		if m.Kind == "outcome" && m.Reason == "late_scan" {
			if m.Decision != want || m.Allowed || m.Error != context.DeadlineExceeded.Error() {
				t.Fatalf("incorrect late completion: %+v", m)
			}
			break
		}
	}
	select {
	case <-fake.responses:
		t.Fatal("late completion responded to kernel twice")
	case <-time.After(50 * time.Millisecond):
	}
	cancel()
	<-server
}

// This process fixture substitutes only the kernel backend, not the broker,
// protocol or scanner processes. It establishes scheduling independence under
// SIGSTOP; actual kernel allow/deny requires a separate gated native campaign.
func TestBrokerProcessFixture(t *testing.T) {
	role := os.Getenv("OPENEDR_BROKER_PROCESS_FIXTURE")
	if role == "" {
		return
	}
	f := os.NewFile(3, "broker-inherited")
	defer f.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if role == "owner" {
		s, err := brokerSocketFromFile(f)
		if err != nil {
			t.Fatal(err)
		}
		defer s.close()
		fake := newFake(t)
		go func() { r := <-fake.responses; json.NewEncoder(os.Stdout).Encode(r.allow); cancel() }()
		err = serveBroker(ctx, s, func(Config) (backend, error) {
			fake.events <- []rawEvent{{fd: 7, pid: 999, mask: OpenPermission}}
			return fake, nil
		})
		if err != nil && !errors.Is(err, context.Canceled) {
			t.Fatal(err)
		}
		return
	}
	if role != "scanner" {
		t.Fatal("invalid fixture role")
	}
	cfg := testConfig()
	cfg.FailClosed = true
	cfg.Deadline = 200 * time.Millisecond
	_ = RunBrokerClient(ctx, f, cfg, func(scanCtx context.Context, _ *os.File, _ Event) (Decision, error) {
		// SIGSTOP freezes every scanner thread, including its Go runtime and
		// transport. The owner must still resolve without this process resuming.
		if err := unix.Kill(os.Getpid(), unix.SIGSTOP); err != nil {
			return Decision{}, err
		}
		// Delivery to the thread group is asynchronous. Do not race its stop
		// with a clean result from the thread that sent the signal.
		<-scanCtx.Done()
		return Decision{}, scanCtx.Err()
	}, nil, nil)
}

func TestBrokerOwnerResolvesWithScannerProcessStopped(t *testing.T) {
	a, b := brokerPair(t)
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	owner := exec.CommandContext(ctx, exe, "-test.run=^TestBrokerProcessFixture$", "-test.timeout=5s")
	owner.Env = append(os.Environ(), "OPENEDR_BROKER_PROCESS_FIXTURE=owner")
	owner.ExtraFiles = []*os.File{b}
	stdout, err := owner.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	var ownerErr bytes.Buffer
	owner.Stderr = &ownerErr
	if err = owner.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { owner.Process.Kill(); owner.Wait() })
	scanner := exec.CommandContext(ctx, exe, "-test.run=^TestBrokerProcessFixture$", "-test.timeout=5s")
	scanner.Env = append(os.Environ(), "OPENEDR_BROKER_PROCESS_FIXTURE=scanner")
	scanner.ExtraFiles = []*os.File{a}
	if err = scanner.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { scanner.Process.Kill(); scanner.Wait() })
	a.Close()
	b.Close()
	var allowed bool
	if err = json.NewDecoder(stdout).Decode(&allowed); err != nil {
		t.Fatalf("owner response: %v %s", err, ownerErr.String())
	}
	if allowed {
		t.Fatal("stopped scanner escaped fail-closed timeout")
	}
	state, err := os.ReadFile(fmt.Sprintf("/proc/%d/status", scanner.Process.Pid))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(state), "State:\tT") {
		t.Fatalf("scanner was not stopped when owner resolved: %s", state)
	}
	if err = owner.Wait(); err != nil {
		t.Fatalf("owner did not exit cleanly: %v %s", err, ownerErr.String())
	}
}
