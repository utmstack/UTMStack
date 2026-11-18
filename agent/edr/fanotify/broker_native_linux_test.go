//go:build linux

package fanotify

import (
	"bufio"
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	"golang.org/x/sys/unix"
)

func TestNativeBrokerOwnerFixture(t *testing.T) {
	if os.Getenv("OPENEDR_NATIVE_BROKER_OWNER") != "1" {
		return
	}
	f := os.NewFile(3, "inherited-owner-channel")
	defer f.Close()
	err := ServeBroker(context.Background(), f)
	if err != nil && !errors.Is(err, io.EOF) && !errors.Is(err, context.Canceled) && !errors.Is(err, unix.ECONNRESET) {
		t.Fatal(err)
	}
}

func nativeBroker(t *testing.T, root string, scan Scanner) (<-chan Outcome, context.CancelFunc) {
	cfg := testConfig()
	cfg.Mounts = []string{root}
	cfg.Roots = []string{root}
	cfg.FailClosed = true
	cfg.Deadline = 500 * time.Millisecond
	cfg.Workers = 2
	cfg.MaxPending = 8
	cfg.MaxFileBytes = 64 << 20
	return nativeBrokerConfig(t, cfg, scan)
}

func nativeBrokerConfig(t *testing.T, cfg Config, scan Scanner) (<-chan Outcome, context.CancelFunc) {
	t.Helper()
	a, b := brokerPair(t)
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cmd := exec.Command(exe, "-test.run=^TestNativeBrokerOwnerFixture$", "-test.timeout=10s")
	cmd.Env = append(os.Environ(), "OPENEDR_NATIVE_BROKER_OWNER=1")
	if binary := os.Getenv("OPENEDR_BROKER_SERVICE_BINARY"); binary != "" {
		cmd = exec.Command(binary, "__fanotify-broker")
	}
	cmd.ExtraFiles = []*os.File{b}
	var logs bytes.Buffer
	cmd.Stdout = &logs
	cmd.Stderr = &logs
	if err = cmd.Start(); err != nil {
		cancel()
		t.Fatal(err)
	}
	b.Close()
	ownerDone := make(chan error, 1)
	go func() { ownerDone <- cmd.Wait() }()
	outcomes := make(chan Outcome, 32)
	health := make(chan error, 32)
	clientDone := make(chan error, 1)
	go func() {
		clientDone <- RunBrokerClient(ctx, a, cfg, scan, func(o Outcome) { outcomes <- o }, func(_ Stats, err error) { health <- err })
		a.Close()
	}()
	t.Cleanup(func() {
		cancel()
		select {
		case <-clientDone:
		case <-time.After(time.Second):
			t.Error("client did not stop")
		}
		select {
		case err := <-ownerDone:
			if err != nil {
				t.Errorf("owner: %v %s", err, logs.String())
			}
		case <-time.After(time.Second):
			cmd.Process.Kill()
			<-ownerDone
			t.Errorf("owner did not stop: %s", logs.String())
		}
	})
	select {
	case err := <-health:
		if err != nil {
			t.Fatalf("broker did not activate: %v", err)
		}
	case err := <-ownerDone:
		ownerDone <- err
		t.Fatalf("broker startup failed: %v %s", err, logs.String())
	case <-time.After(3 * time.Second):
		t.Fatal("no broker readiness")
	}
	return outcomes, cancel
}

func TestNativeBrokerOpenAllowDenyPinnedSubstitution(t *testing.T) {
	root := isolatedMount(t)
	path := filepath.Join(root, "broker-owned-marker")
	if err := os.WriteFile(path, []byte("inert-deny-marker"), 0600); err != nil {
		t.Fatal(err)
	}
	started := make(chan struct{}, 1)
	release := make(chan struct{})
	var releaseOnce sync.Once
	defer releaseOnce.Do(func() { close(release) })
	var first atomic.Bool
	outcomes, _ := nativeBroker(t, root, func(ctx context.Context, f *os.File, ev Event) (Decision, error) {
		if first.CompareAndSwap(false, true) {
			started <- struct{}{}
			select {
			case <-release:
			case <-ctx.Done():
				return Decision{}, ctx.Err()
			}
		}
		data, err := io.ReadAll(f)
		detected := strings.Contains(string(data), "inert-deny-marker")
		return Decision{Clean: !detected, Detected: detected, Signature: map[bool]string{true: "Safe.Broker.Marker"}[detected], Generation: "native-fixture-generation"}, err
	})
	actor := ownedCommand(t, path, false)
	type result struct {
		out []byte
		err error
	}
	done := make(chan result, 1)
	go func() { out, err := actor.CombinedOutput(); done <- result{out, err} }()
	select {
	case <-started:
	case <-time.After(2 * time.Second):
		t.Fatal("no pinned request")
	}
	if err := os.Rename(path, path+".held"); err != nil {
		t.Fatal(err)
	}
	// A second worker handles the replacement's open while the first still
	// holds the original descriptor. Reopening the path would now scan clean.
	if err := os.WriteFile(path, []byte("ordinary harmless replacement"), 0600); err != nil {
		t.Fatal(err)
	}
	releaseOnce.Do(func() { close(release) })
	select {
	case r := <-done:
		if r.err != nil || !strings.Contains(string(r.out), "DENIED") {
			t.Fatalf("original escaped deny: %s %v", r.out, r.err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("opener stuck")
	}
	deadline := time.After(time.Second)
	for {
		select {
		case o := <-outcomes:
			if !o.Decision.Detected {
				continue
			}
			if o.Allowed || o.Decision.Generation != "native-fixture-generation" {
				t.Fatalf("wrong original decision: %+v", o)
			}
			goto detected
		case <-deadline:
			t.Fatal("missing deny outcome")
		}
	}
detected:
	// The path is replaced after the original decision, and the next event
	// must independently scan the new inode and allow its harmless bytes.
	out, err := ownedCommand(t, path, false).CombinedOutput()
	if err != nil || !strings.Contains(string(out), "ALLOWED") {
		t.Fatalf("harmless replacement denied: %s %v", out, err)
	}
}

const nativeBrokerExecRoot = "/usr/local/bin/openedr-broker-fixture"

func TestNativeBrokerExecActorFixture(t *testing.T) {
	if os.Getenv("OPENEDR_BROKER_EXEC_ACTOR") != "1" {
		return
	}
	if os.Getenv("OPENEDR_BROKER_ROOT_SCOPE") != "1" || os.Getenv("OPENEDR_ISOLATED_VM") != "1" {
		t.Fatal("requires explicit diskless guest scope")
	}
	decoder, encoder := json.NewDecoder(os.Stdin), json.NewEncoder(os.Stdout)
	for i := 0; i < 2; i++ {
		var path string
		if err := decoder.Decode(&path); err != nil {
			t.Fatal(err)
		}
		if path != nativeBrokerExecRoot+"/allow" && path != nativeBrokerExecRoot+"/deny" {
			t.Fatal("unowned helper request")
		}
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		cmd := exec.CommandContext(ctx, path, "-test.run=^TestNativeOwnedExecFixture$")
		cmd.Env = append(os.Environ(), "OPENEDR_FANOTIFY_EXEC_CHILD=1")
		out, err := cmd.CombinedOutput()
		cancel()
		r := execFixtureReply{Output: string(out), Denied: errors.Is(err, os.ErrPermission)}
		if err != nil {
			r.Error = err.Error()
		}
		if err = encoder.Encode(r); err != nil {
			t.Fatal(err)
		}
	}
}

func TestNativeBrokerExecutableMountAllowAndDeny(t *testing.T) {
	_ = isolatedMount(t)
	if os.Getenv("OPENEDR_BROKER_ROOT_SCOPE") != "1" {
		t.Skip("requires explicit diskless guest executable-mount scope")
	}
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	contents, err := os.ReadFile(exe)
	if err != nil {
		t.Fatal(err)
	}
	token := make([]byte, 16)
	if _, err = rand.Read(token); err != nil {
		t.Fatal(err)
	}
	marker := []byte("OWNED_BROKER_EXEC_" + hex.EncodeToString(token))
	if bytes.Contains(contents, marker) {
		t.Fatal("marker already in helper")
	}
	if err = os.MkdirAll(nativeBrokerExecRoot, 0700); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"allow", "deny"} {
		data := contents
		if name == "deny" {
			data = append(append([]byte(nil), contents...), marker...)
		}
		if err = os.WriteFile(nativeBrokerExecRoot+"/"+name, data, 0700); err != nil {
			t.Fatal(err)
		}
	}
	launcher := exec.Command(exe, "-test.run=^TestNativeBrokerExecActorFixture$", "-test.timeout=10s")
	launcher.Env = append(os.Environ(), "OPENEDR_BROKER_EXEC_ACTOR=1")
	input, err := launcher.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}
	defer input.Close()
	output, err := launcher.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err = launcher.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() { launcher.Process.Kill(); launcher.Wait() }()
	cfg := testConfig()
	cfg.Mode = ModeExec
	cfg.Mounts = []string{"/"}
	cfg.Roots = []string{"/"}
	cfg.FailClosed = true
	cfg.Deadline = time.Second
	cfg.Workers = 2
	cfg.MaxPending = 8
	cfg.MaxFileBytes = 64 << 20
	if b, err := openBackend(cfg); !errors.Is(err, ErrUnsupported) {
		if b != nil {
			b.Close()
		}
		t.Fatalf("in-process executable-mount guard lost: %v", err)
	}
	_, cancel := nativeBrokerConfig(t, cfg, func(ctx context.Context, f *os.File, e Event) (Decision, error) {
		runtime.GC()
		data, err := io.ReadAll(io.LimitReader(f, 64<<20+1))
		detected := bytes.Contains(data, marker)
		return Decision{Clean: !detected, Detected: detected, Signature: map[bool]string{true: "Safe.Broker.Exec"}[detected], Generation: "native-exec-generation"}, err
	})
	defer cancel()
	encoder, decoder := json.NewEncoder(input), json.NewDecoder(output)
	for _, name := range []string{"allow", "deny"} {
		if err = encoder.Encode(nativeBrokerExecRoot + "/" + name); err != nil {
			t.Fatal(err)
		}
		type result struct {
			reply execFixtureReply
			err   error
		}
		done := make(chan result, 1)
		go func() { var r execFixtureReply; e := decoder.Decode(&r); done <- result{r, e} }()
		select {
		case r := <-done:
			if r.err != nil {
				t.Fatal(r.err)
			}
			if name == "allow" && (r.reply.Error != "" || !strings.Contains(r.reply.Output, "OWNED_EXEC_ALLOWED")) {
				t.Fatalf("harmless execution failed: %+v", r.reply)
			}
			if name == "deny" && (!r.reply.Denied || strings.Contains(r.reply.Output, "OWNED_EXEC_ALLOWED")) {
				t.Fatalf("marker execution escaped denial: %+v", r.reply)
			}
		case <-time.After(4 * time.Second):
			t.Fatal("executable permission caller stuck")
		}
	}
}

func TestNativeBrokerPendingShutdownAndGroupRelease(t *testing.T) {
	root := isolatedMount(t)
	path := filepath.Join(root, "broker-shutdown-owned")
	if err := os.WriteFile(path, []byte("ordinary harmless bytes"), 0600); err != nil {
		t.Fatal(err)
	}
	started := make(chan struct{}, 1)
	_, cancel := nativeBroker(t, root, func(ctx context.Context, _ *os.File, _ Event) (Decision, error) {
		started <- struct{}{}
		<-ctx.Done()
		return Decision{}, ctx.Err()
	})
	type result struct {
		out []byte
		err error
	}
	done := make(chan result, 1)
	go func() { out, err := ownedCommand(t, path, false).CombinedOutput(); done <- result{out, err} }()
	select {
	case <-started:
	case <-time.After(2 * time.Second):
		t.Fatal("no pending permission")
	}
	cancel()
	select {
	case r := <-done:
		if r.err != nil || !strings.Contains(string(r.out), "DENIED") {
			t.Fatalf("pending shutdown did not apply fail-closed: %s %v", r.out, r.err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("shutdown left caller stuck")
	}
	// The owner has closed its group. FailClosed is not persistent kernel
	// enforcement after owner shutdown; later access must be released.
	out, err := ownedCommand(t, path, false).CombinedOutput()
	if err != nil || !strings.Contains(string(out), "ALLOWED") {
		t.Fatalf("closed group retained caller: %s %v", out, err)
	}
}

func TestNativeBrokerProcessManagerReapsCrashedOwner(t *testing.T) {
	root := isolatedMount(t)
	path := filepath.Join(root, "broker-process-manager-owned")
	if err := os.WriteFile(path, []byte("ordinary owned bytes"), 0600); err != nil {
		t.Fatal(err)
	}
	cfg := testConfig()
	cfg.Mounts = []string{root}
	cfg.Roots = []string{root}
	cfg.FailClosed = true
	cfg.Deadline = 300 * time.Millisecond
	started := make(chan struct{}, 1)
	b, err := NewPermissionBroker(cfg, func(ctx context.Context, _ *os.File, _ Event) (Decision, error) {
		started <- struct{}{}
		<-ctx.Done()
		return Decision{}, ctx.Err()
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(exe, "-test.run=^TestNativeBrokerOwnerFixture$", "-test.timeout=10s")
	cmd.Env = append(os.Environ(), "OPENEDR_NATIVE_BROKER_OWNER=1")
	b.command = func() *exec.Cmd { return cmd }
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- b.Run(ctx) }()
	t.Cleanup(func() {
		cancel()
		select {
		case <-done:
		case <-time.After(2 * time.Second):
			t.Error("process manager did not stop")
		}
	})
	limit := time.Now().Add(3 * time.Second)
	for b.Health() != nil {
		if time.Now().After(limit) {
			t.Fatal("broker process never became healthy")
		}
		time.Sleep(5 * time.Millisecond)
	}
	type result struct {
		out []byte
		err error
	}
	opened := make(chan result, 1)
	go func() { out, err := ownedCommand(t, path, false).CombinedOutput(); opened <- result{out, err} }()
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("owner never received request")
	}
	if err = cmd.Process.Kill(); err != nil {
		t.Fatal(err)
	}
	select {
	case r := <-opened:
		if r.err != nil || !strings.Contains(string(r.out), "ALLOWED") {
			t.Fatalf("owner crash did not release kernel caller: %s %v", r.out, r.err)
		}
	case <-time.After(time.Second):
		t.Fatal("crashed owner left caller blocked")
	}
	select {
	case err := <-done:
		if err == nil {
			t.Fatal("owner crash did not degrade manager")
		}
		done <- err
	case <-time.After(2 * time.Second):
		t.Fatal("owner was not reaped")
	}
	if b.Health() == nil {
		t.Fatal("crashed owner still healthy")
	}
	if err = unix.Kill(cmd.Process.Pid, 0); !errors.Is(err, unix.ESRCH) {
		t.Fatalf("broker child remains: %v", err)
	}
	if err = b.Run(context.Background()); !errors.Is(err, ErrUsed) {
		t.Fatalf("manager allowed repeated worker pools: %v", err)
	}
}

func TestNativeBrokerOrdinaryUserAndRootDecisions(t *testing.T) {
	root := isolatedMount(t)
	for _, name := range []string{"ordinary-allow", "ordinary-deny"} {
		if err := os.WriteFile(filepath.Join(root, name), []byte(name), 0644); err != nil {
			t.Fatal(err)
		}
	}
	_, cancel := nativeBroker(t, root, func(_ context.Context, f *os.File, e Event) (Decision, error) {
		data, err := io.ReadAll(f)
		detected := strings.HasSuffix(string(data), "deny")
		if e.OSIdentity <= 0 {
			return Decision{}, ErrAttribution
		}
		return Decision{Clean: !detected, Detected: detected, Signature: map[bool]string{true: "Safe.Broker.User"}[detected], Generation: "native-user-generation"}, err
	})
	defer cancel()
	for _, uid := range []uint32{0, 65534} {
		for _, name := range []string{"ordinary-allow", "ordinary-deny"} {
			cmd := ownedCommand(t, filepath.Join(root, name), false)
			cmd.SysProcAttr = &syscall.SysProcAttr{Credential: &syscall.Credential{Uid: uid, Gid: uid}}
			out, err := cmd.CombinedOutput()
			want := "ALLOWED"
			if strings.HasSuffix(name, "deny") {
				want = "DENIED"
			}
			if err != nil || !strings.Contains(string(out), want) {
				t.Fatalf("uid=%d %s: %s %v", uid, name, out, err)
			}
		}
	}
}

func TestNativeBrokerOpenDeadlinePolicy(t *testing.T) {
	root := isolatedMount(t)
	path := filepath.Join(root, "broker-deadline-owned")
	if err := os.WriteFile(path, []byte("ordinary owned bytes"), 0600); err != nil {
		t.Fatal(err)
	}
	for _, closed := range []bool{false, true} {
		t.Run(fmt.Sprint(closed), func(t *testing.T) {
			cfg := testConfig()
			cfg.Mounts = []string{root}
			cfg.Roots = []string{root}
			cfg.Deadline = 100 * time.Millisecond
			cfg.FailClosed = closed
			_, cancel := nativeBrokerConfig(t, cfg, func(ctx context.Context, _ *os.File, _ Event) (Decision, error) {
				<-ctx.Done()
				return Decision{}, ctx.Err()
			})
			defer cancel()
			out, err := ownedCommand(t, path, false).CombinedOutput()
			want := "ALLOWED"
			if closed {
				want = "DENIED"
			}
			if err != nil || !strings.Contains(string(out), want) {
				t.Fatalf("deadline policy=%t: %s %v", closed, out, err)
			}
		})
	}
}

type nativeBrokerProcessIDs struct{ Owner, Actor int }

func TestNativeBrokerScannerFailureFixture(t *testing.T) {
	mode := os.Getenv("OPENEDR_NATIVE_BROKER_SCANNER_FAILURE")
	if mode == "" {
		return
	}
	root := isolatedMount(t)
	path := filepath.Join(root, "broker-scanner-failure-"+mode)
	announced := make(chan struct{})
	cfg := testConfig()
	cfg.Mounts = []string{root}
	cfg.Roots = []string{root}
	cfg.FailClosed = true
	cfg.Deadline = 200 * time.Millisecond
	b, err := NewPermissionBroker(cfg, func(ctx context.Context, _ *os.File, _ Event) (Decision, error) {
		<-announced
		if mode == "crash" {
			os.Exit(97)
		}
		if mode != "stop" {
			panic("unknown scanner failure fixture")
		}
		if err := unix.Kill(os.Getpid(), unix.SIGSTOP); err != nil {
			return Decision{}, err
		}
		<-ctx.Done()
		return Decision{}, ctx.Err()
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	owner := exec.Command(exe, "-test.run=^TestNativeBrokerOwnerFixture$")
	owner.Env = append(os.Environ(), "OPENEDR_NATIVE_BROKER_OWNER=1")
	if binary := os.Getenv("OPENEDR_BROKER_SERVICE_BINARY"); binary != "" {
		owner = exec.Command(binary, "__fanotify-broker")
	}
	b.command = func() *exec.Cmd { return owner }
	go b.Run(context.Background())
	limit := time.Now().Add(3 * time.Second)
	for b.Health() != nil {
		if time.Now().After(limit) {
			t.Fatal("failure fixture owner never ready")
		}
		time.Sleep(5 * time.Millisecond)
	}
	actor := ownedCommand(t, path, false)
	output := os.NewFile(4, "independent-actor-result")
	defer output.Close()
	actor.Stdout = output
	actor.Stderr = output
	if err = actor.Start(); err != nil {
		t.Fatal(err)
	}
	if err = json.NewEncoder(os.Stdout).Encode(nativeBrokerProcessIDs{Owner: owner.Process.Pid, Actor: actor.Process.Pid}); err != nil {
		t.Fatal(err)
	}
	close(announced)
	select {}
}

func TestNativeBrokerRealKernelScannerStopAndCrash(t *testing.T) {
	root := isolatedMount(t)
	// The test process adopts and explicitly reaps the fixture's descendants
	// after the scanner exits. Nothing is left for an unrelated system init.
	if err := unix.Prctl(unix.PR_SET_CHILD_SUBREAPER, 1, 0, 0, 0); err != nil {
		t.Fatal(err)
	}
	defer unix.Prctl(unix.PR_SET_CHILD_SUBREAPER, 0, 0, 0, 0)
	for _, mode := range []string{"stop", "crash"} {
		t.Run(mode, func(t *testing.T) {
			path := filepath.Join(root, "broker-scanner-failure-"+mode)
			if err := os.WriteFile(path, []byte("ordinary owned bytes"), 0600); err != nil {
				t.Fatal(err)
			}
			exe, err := os.Executable()
			if err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 6*time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, exe, "-test.run=^TestNativeBrokerScannerFailureFixture$", "-test.timeout=6s")
			cmd.Env = append(os.Environ(), "OPENEDR_NATIVE_BROKER_SCANNER_FAILURE="+mode)
			read, write, err := os.Pipe()
			if err != nil {
				t.Fatal(err)
			}
			defer read.Close()
			defer write.Close()
			cmd.ExtraFiles = []*os.File{nil, write}
			metadata, err := cmd.StdoutPipe()
			if err != nil {
				t.Fatal(err)
			}
			if err = cmd.Start(); err != nil {
				t.Fatal(err)
			}
			write.Close()
			defer func() { cmd.Process.Kill(); cmd.Wait() }()
			var ids nativeBrokerProcessIDs
			if err = json.NewDecoder(metadata).Decode(&ids); err != nil {
				t.Fatal(err)
			}
			line := make(chan string, 1)
			go func() { s, _ := bufio.NewReader(read).ReadString('\n'); line <- s }()
			select {
			case s := <-line:
				// Channel EOF cancels the independently owned loop, which denies
				// its already-read request before closing the group.
				valid := strings.Contains(s, "DENIED")
				if !valid {
					t.Fatalf("scanner %s OS effect: %q", mode, s)
				}
			case <-time.After(3 * time.Second):
				t.Fatal("scanner failure left kernel opener stuck")
			}
			if mode == "stop" {
				state, err := os.ReadFile(fmt.Sprintf("/proc/%d/status", cmd.Process.Pid))
				if err != nil || !strings.Contains(string(state), "State:\tT") {
					t.Fatalf("scanner was not stopped: %s %v", state, err)
				}
				cmd.Process.Kill()
			}
			err = cmd.Wait()
			if err == nil {
				t.Fatal("scanner failure fixture unexpectedly exited successfully")
			}
			if mode == "crash" {
				var e *exec.ExitError
				if !errors.As(err, &e) || e.ExitCode() != 97 {
					t.Fatalf("scanner did not crash at owned checkpoint: %v", err)
				}
			}
			for _, pid := range []int{ids.Owner, ids.Actor} {
				limit := time.Now().Add(2 * time.Second)
				for {
					var status unix.WaitStatus
					reaped, e := unix.Wait4(pid, &status, unix.WNOHANG, nil)
					if reaped == pid {
						break
					}
					if e != nil {
						t.Fatalf("reap descendant %d: %v", pid, e)
					}
					if time.Now().After(limit) {
						t.Fatalf("descendant %d remains", pid)
					}
					time.Sleep(5 * time.Millisecond)
				}
				if err := unix.Kill(pid, 0); !errors.Is(err, unix.ESRCH) {
					t.Fatalf("descendant still exists: %d %v", pid, err)
				}
			}
			out, err := ownedCommand(t, path, false).CombinedOutput()
			if err != nil || !strings.Contains(string(out), "ALLOWED") {
				t.Fatalf("dead owner retained later access: %s %v", out, err)
			}
		})
	}
}
