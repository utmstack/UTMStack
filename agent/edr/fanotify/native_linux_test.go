//go:build linux

package fanotify

import (
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
	"testing"
	"time"

	"golang.org/x/sys/unix"
)

func TestNativeOwnedIOFixture(t *testing.T) {
	path := os.Getenv("OPENEDR_FANOTIFY_CHILD_PATH")
	if path == "" {
		return
	}
	if os.Getenv("OPENEDR_FANOTIFY_CHILD_WRITE") == "1" {
		if err := os.WriteFile(path, []byte("ordinary owned write fixture"), 0600); err != nil {
			t.Fatal(err)
		}
		time.Sleep(300 * time.Millisecond)
		return
	}
	f, err := os.Open(path)
	if errors.Is(err, os.ErrPermission) {
		os.Stdout.WriteString("DENIED\n")
		return
	}
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if _, err = io.ReadAll(f); err != nil {
		t.Fatal(err)
	}
	os.Stdout.WriteString("ALLOWED\n")
}
func isolatedMount(t *testing.T) string {
	t.Helper()
	root := os.Getenv("OPENEDR_FANOTIFY_TEST_ROOT")
	if os.Getenv("OPENEDR_ISOLATED_VM") != "1" || root == "" {
		t.Skip("native fanotify requires explicit isolated-VM tmpfs fixture opt-in")
	}
	if os.Geteuid() != 0 || !strings.HasPrefix(root, "/tmp/openedr-fanotify-") || filepath.Clean(root) != root {
		t.Fatal("native fixture must be root on a dedicated /tmp/openedr-fanotify-* mount")
	}
	var fs unix.Statfs_t
	if err := unix.Statfs(root, &fs); err != nil || uint64(fs.Type) != uint64(unix.TMPFS_MAGIC) {
		t.Fatalf("native fixture requires tmpfs: %v", err)
	}
	mounts, err := mountPoints()
	if err != nil || mounts[root] == 0 {
		t.Fatalf("native fixture is not an actual mountpoint: %v", err)
	}
	return root
}
func nativeMonitor(t *testing.T, root, mode string, scan Scanner, observe func(Outcome)) (*Monitor, context.CancelFunc) {
	t.Helper()
	m, err := New(Config{Enabled: true, Mounts: []string{root}, Roots: []string{root}, Mode: mode, FailClosed: true, Deadline: time.Second, Workers: 2, MaxPending: 8, MaxFileBytes: 128 << 20}, scan, observe)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	failedStart := false
	go func() { done <- m.Run(ctx) }()
	t.Cleanup(func() {
		cancel()
		if failedStart {
			return
		}
		select {
		case err := <-done:
			if !errors.Is(err, context.Canceled) {
				t.Errorf("native monitor ended: %v", err)
			}
		case <-time.After(2 * time.Second):
			t.Error("native monitor failed shutdown")
		}
	})
	deadline := time.Now().Add(2 * time.Second)
	for m.Health() != nil {
		select {
		case err := <-done:
			failedStart = true
			t.Fatalf("fanotify explicit native activation failed: %v", err)
		default:
		}
		if time.Now().After(deadline) {
			t.Fatal(m.Health())
		}
		time.Sleep(5 * time.Millisecond)
	}
	return m, cancel
}
func ownedCommand(t *testing.T, path string, write bool) *exec.Cmd {
	t.Helper()
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	t.Cleanup(cancel)
	cmd := exec.CommandContext(ctx, exe, "-test.run=^TestNativeOwnedIOFixture$")
	cmd.Env = append(os.Environ(), "OPENEDR_FANOTIFY_CHILD_PATH="+path)
	if write {
		cmd.Env = append(cmd.Env, "OPENEDR_FANOTIFY_CHILD_WRITE=1")
	}
	return cmd
}
func TestNativePermissionRepliesUsePinnedOriginalFile(t *testing.T) {
	root := isolatedMount(t)
	path := filepath.Join(root, "inert-marker.txt")
	if err := os.WriteFile(path, []byte("inert-deny-marker"), 0600); err != nil {
		t.Fatal(err)
	}
	started, release := make(chan struct{}, 1), make(chan struct{})
	defer close(release)
	_, _ = nativeMonitor(t, root, ModeOpen, func(ctx context.Context, f *os.File, e Event) (Decision, error) {
		started <- struct{}{}
		select {
		case <-release:
		case <-ctx.Done():
			return Decision{}, ctx.Err()
		}
		data, err := io.ReadAll(f)
		return Decision{Clean: !strings.Contains(string(data), "inert-deny-marker"), Detected: strings.Contains(string(data), "inert-deny-marker"), Signature: "Safe.Fanotify.Marker"}, err
	}, nil)
	cmd := ownedCommand(t, path, false)
	done := make(chan struct {
		out []byte
		err error
	}, 1)
	go func() {
		out, err := cmd.CombinedOutput()
		done <- struct {
			out []byte
			err error
		}{out, err}
	}()
	select {
	case <-started:
	case <-time.After(2 * time.Second):
		t.Fatal("permission scanner did not receive pinned file")
	}
	if err := os.Rename(path, path+".held"); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("benign replacement"), 0600); err != nil {
		t.Fatal(err)
	}
	release <- struct{}{}
	select {
	case result := <-done:
		if result.err != nil || !strings.Contains(string(result.out), "DENIED") {
			t.Fatalf("pinned malicious original escaped decision: %s %v", result.out, result.err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("owned opener was not released")
	}
}
func TestNativeNotifyAttributesOwnedWrite(t *testing.T) {
	root := isolatedMount(t)
	path := filepath.Join(root, "owned-write.txt")
	events := make(chan Outcome, 32)
	_, _ = nativeMonitor(t, root, ModeNotify, nil, func(o Outcome) { events <- o })
	cmd := ownedCommand(t, path, true)
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = cmd.Process.Kill(); _ = cmd.Wait() }()
	deadline := time.After(3 * time.Second)
	for {
		select {
		case o := <-events:
			if o.Event.PID == cmd.Process.Pid && o.Event.Mask&Modify != 0 {
				if o.Event.OSIdentity <= 0 || o.Event.PPID != os.Getpid() || o.Event.Image == "" || o.Event.Path != path {
					t.Fatalf("invalid native attribution: %+v", o)
				}
				return
			}
		case <-deadline:
			t.Fatal("owned modifying process was not attributed")
		}
	}
}

func TestNativeOwnedExecFixture(t *testing.T) {
	if os.Getenv("OPENEDR_FANOTIFY_EXEC_CHILD") == "1" {
		fmt.Println("OWNED_EXEC_ALLOWED")
	}
}

type execFixtureReply struct {
	Output string
	Error  string
	Denied bool
}

func TestNativeOwnedExecLauncher(t *testing.T) {
	root := os.Getenv("OPENEDR_FANOTIFY_EXEC_LAUNCHER_ROOT")
	if root == "" {
		return
	}
	if os.Getenv("OPENEDR_ISOLATED_VM") != "1" || !strings.HasPrefix(root, "/tmp/openedr-fanotify-") {
		t.Fatal("launcher requires dedicated isolated fixture scope")
	}
	decoder, encoder := json.NewDecoder(os.Stdin), json.NewEncoder(os.Stdout)
	for i := 0; i < 2; i++ {
		var path string
		if err := decoder.Decode(&path); err != nil {
			t.Fatal(err)
		}
		if path != filepath.Join(root, "owned-exec-allow") && path != filepath.Join(root, "owned-exec-deny") {
			t.Fatal("launcher accepts only owned copied helper fixture paths")
		}
		ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
		cmd := exec.CommandContext(ctx, path, "-test.run=^TestNativeOwnedExecFixture$")
		cmd.Env = append(os.Environ(), "OPENEDR_FANOTIFY_EXEC_CHILD=1")
		out, err := cmd.CombinedOutput()
		cancel()
		reply := execFixtureReply{Output: string(out), Denied: errors.Is(err, os.ErrPermission)}
		if err != nil {
			reply.Error = err.Error()
		}
		if err := encoder.Encode(reply); err != nil {
			t.Fatal(err)
		}
	}
}

func TestNativeExecPermissionAllowsAndDeniesOnlyOwnedCopiedHelper(t *testing.T) {
	root := isolatedMount(t)
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	contents, err := os.ReadFile(exe)
	if err != nil {
		t.Fatal(err)
	}
	if len(contents) > 64<<20 {
		t.Fatal("owned helper exceeds native fixture budget")
	}
	token := make([]byte, 16)
	if _, err = rand.Read(token); err != nil {
		t.Fatal(err)
	}
	marker := []byte("OPENEDR_SAFE_EXEC_OVERLAY_" + hex.EncodeToString(token))
	if bytes.Contains(contents, marker) {
		t.Fatal("fixture marker unexpectedly in original helper")
	}
	allow := filepath.Join(root, "owned-exec-allow")
	deny := filepath.Join(root, "owned-exec-deny")
	if err = os.WriteFile(allow, contents, 0700); err != nil {
		t.Fatal(err)
	}
	blocked := append(append([]byte(nil), contents...), marker...)
	if err = os.WriteFile(deny, blocked, 0700); err != nil {
		t.Fatal(err)
	}
	// Launch the independent actor before permission activation. Forking a
	// protected child from the monitor's own Go runtime can deadlock vfork/GC.
	launcher := exec.Command(exe, "-test.run=^TestNativeOwnedExecLauncher$")
	launcher.Env = append(os.Environ(), "OPENEDR_FANOTIFY_EXEC_LAUNCHER_ROOT="+root)
	stdin, err := launcher.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}
	defer stdin.Close()
	stdout, err := launcher.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := launcher.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = launcher.Process.Kill(); _ = launcher.Wait() }()
	encoder, decoder := json.NewEncoder(stdin), json.NewDecoder(stdout)
	_, cancelMonitor := nativeMonitor(t, root, ModeExec, func(ctx context.Context, f *os.File, e Event) (Decision, error) {
		runtime.GC() // Exercise a real monitor GC while another process waits in exec.
		b, err := io.ReadAll(io.LimitReader(f, 128<<20+1))
		hit := bytes.Contains(b, marker)
		return Decision{Clean: !hit, Detected: hit, Signature: "Safe.Owned.Exec.Marker"}, err
	}, nil)
	defer cancelMonitor() // Resolve pending requests before waiting for the launcher.
	run := func(path string) execFixtureReply {
		if err := encoder.Encode(path); err != nil {
			t.Fatal(err)
		}
		done := make(chan struct {
			reply execFixtureReply
			err   error
		}, 1)
		go func() {
			var reply execFixtureReply
			err := decoder.Decode(&reply)
			done <- struct {
				reply execFixtureReply
				err   error
			}{reply, err}
		}()
		select {
		case result := <-done:
			if result.err != nil {
				t.Fatal(result.err)
			}
			return result.reply
		case <-time.After(5 * time.Second):
			t.Fatal("independent owned executor did not receive permission resolution")
			return execFixtureReply{}
		}
	}
	if reply := run(allow); reply.Error != "" || !strings.Contains(reply.Output, "OWNED_EXEC_ALLOWED") {
		t.Fatalf("owned benign helper execution failed: %+v", reply)
	}
	if reply := run(deny); reply.Error == "" || !reply.Denied || strings.Contains(reply.Output, "OWNED_EXEC_ALLOWED") {
		t.Fatalf("owned marker helper execution escaped deny: %+v", reply)
	}
}

func TestNativePermissionRejectsDaemonExecutableMount(t *testing.T) {
	root := isolatedMount(t)
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	fd, err := directoryFD(filepath.Dir(exe))
	if err != nil {
		t.Fatal(err)
	}
	want, err := fdMountID(fd)
	unix.Close(fd)
	if err != nil {
		t.Fatal(err)
	}
	mounts, err := mountPoints()
	if err != nil {
		t.Fatal(err)
	}
	var sourceMount string
	for path, id := range mounts {
		if id == want {
			sourceMount = path
			break
		}
	}
	if sourceMount == "" || sourceMount == root {
		t.Fatal("fixture executable must reside outside protected data mount")
	}
	for _, mode := range []string{ModeExec, ModeOpen} {
		_, err := openBackend(Config{Mounts: []string{sourceMount}, Roots: []string{sourceMount}, Mode: mode})
		if !errors.Is(err, ErrUnsupported) {
			t.Fatalf("permission mode %s accepted monitor executable mount: %v", mode, err)
		}
	}
}
func TestNativePermissionTimeoutAndShutdownReleaseOwnedOpeners(t *testing.T) {
	for _, shutdown := range []bool{false, true} {
		t.Run(fmt.Sprint(shutdown), func(t *testing.T) {
			root := isolatedMount(t)
			path := filepath.Join(root, fmt.Sprintf("owned-wait-%v.txt", shutdown))
			if err := os.WriteFile(path, []byte("ordinary controlled fixture"), 0600); err != nil {
				t.Fatal(err)
			}
			started := make(chan struct{}, 1)
			release := make(chan struct{})
			defer close(release)
			_, cancel := nativeMonitor(t, root, ModeOpen, func(ctx context.Context, f *os.File, e Event) (Decision, error) {
				started <- struct{}{}
				<-release
				return Decision{Clean: true}, nil
			}, nil)
			cmd := ownedCommand(t, path, false)
			done := make(chan struct {
				out []byte
				err error
			}, 1)
			go func() {
				out, err := cmd.CombinedOutput()
				done <- struct {
					out []byte
					err error
				}{out, err}
			}()
			select {
			case <-started:
			case <-time.After(2 * time.Second):
				t.Fatal("owned open permission not observed")
			}
			if shutdown {
				cancel()
			}
			select {
			case result := <-done:
				if result.err != nil || !strings.Contains(string(result.out), "DENIED") {
					t.Fatalf("held opener not released with configured deny: %s %v", result.out, result.err)
				}
			case <-time.After(2 * time.Second):
				t.Fatal("kernel opener remained blocked")
			}
		})
	}
}
