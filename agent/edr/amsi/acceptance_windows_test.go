//go:build windows && edr_acceptance

package amsi

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/utmstack/UTMStack/agent/edr/config"
	"github.com/utmstack/UTMStack/agent/edr/engine"
	"golang.org/x/sys/windows"
)

const acceptanceMarker = "UTMSTACK_AMSI_ACCEPTANCE_2026"

func acceptanceGate(t *testing.T) bool {
	t.Helper()
	if os.Getenv("EDR_WINDOWS_ACCEPTANCE") != "isolated-vm" {
		t.Skip("explicit isolated Windows VM acceptance only")
	}
	var token windows.Token
	if err := windows.OpenProcessToken(windows.CurrentProcess(), windows.TOKEN_QUERY, &token); err != nil {
		t.Fatal(err)
	}
	defer token.Close()
	user, err := token.GetTokenUser()
	if err != nil {
		t.Fatal(err)
	}
	system := user.User.Sid.IsWellKnown(windows.WinLocalSystemSid)
	if want := os.Getenv("EDR_ACCEPTANCE_IDENTITY"); want != "" {
		if system != (want == "system") {
			t.Fatalf("identity %s does not match requested %s", user.User.Sid, want)
		}
		if want == "standard" && token.IsElevated() {
			t.Fatal("standard-user acceptance unexpectedly elevated")
		}
	}
	t.Logf("identity SID=%s elevated=%v system=%v", user.User.Sid, token.IsElevated(), system)
	return system
}

// Executed as a separate SYSTEM process by windows-acceptance.ps1. The same
// production pipe and actual native Go engine serve both SYSTEM and standard
// user clients. It never installs a provider or changes Windows protection.
func TestWindowsAcceptancePipeService(t *testing.T) {
	if !acceptanceGate(t) {
		t.Fatal("the production pipe owner must be LocalSystem")
	}
	ready, stop := os.Getenv("EDR_ACCEPTANCE_READY"), os.Getenv("EDR_ACCEPTANCE_STOP")
	if ready == "" || stop == "" {
		t.Fatal("readiness and stop paths are required")
	}
	dir := t.TempDir()
	var wide []byte
	for _, c := range acceptanceMarker {
		wide = append(wide, byte(c), 0)
	}
	rules := fmt.Sprintf("UTM.Acceptance.ASCII:0:*:%s\nUTM.Acceptance.UTF16:0:*:%s\n", hex.EncodeToString([]byte(acceptanceMarker)), hex.EncodeToString(wide))
	if err := os.WriteFile(filepath.Join(dir, "acceptance.ndb"), []byte(rules), 0600); err != nil {
		t.Fatal(err)
	}
	cfg := config.Default()
	cfg.SignatureDBDir, cfg.FailMode = dir, os.Getenv("EDR_ACCEPTANCE_FAIL_MODE")
	if cfg.FailMode != "open" && cfg.FailMode != "closed" {
		t.Fatal("explicit fail mode required")
	}
	eng := engine.New(cfg)
	if err := eng.EnsureRunning(); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Minute)
	defer cancel()
	done := make(chan struct{})
	go func() { NewPipeServer(NewScanner(cfg, nil, eng)).Run(ctx); close(done) }()
	for {
		h, err := acceptanceConnect(ctx)
		if err == nil {
			windows.CloseHandle(h)
			break
		}
		select {
		case <-done:
			t.Fatal("pipe service exited before ready")
		case <-ctx.Done():
			t.Fatal(ctx.Err())
		case <-time.After(50 * time.Millisecond):
		}
	}
	if err := os.WriteFile(ready, []byte("ready"), 0600); err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(stop); err == nil {
			break
		}
		select {
		case <-done:
			t.Fatal("pipe service exited prematurely")
		case <-ctx.Done():
			t.Fatal(ctx.Err())
		case <-time.After(50 * time.Millisecond):
		}
	}
	cancel()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("pipe shutdown did not cancel the pending connection")
	}
}

func acceptanceConnect(ctx context.Context) (windows.Handle, error) {
	name, _ := windows.UTF16PtrFromString(PipeName)
	var last error
	for ctx.Err() == nil {
		h, err := windows.CreateFile(name, pipeClientRights, 0, nil, windows.OPEN_EXISTING, windows.FILE_FLAG_OVERLAPPED|windows.SECURITY_SQOS_PRESENT|windows.SECURITY_IDENTIFICATION, 0)
		if err == nil {
			return h, nil
		}
		last = err
		if err != windows.ERROR_PIPE_BUSY && err != windows.ERROR_FILE_NOT_FOUND {
			return 0, err
		}
		select {
		case <-ctx.Done():
			break
		case <-time.After(10 * time.Millisecond):
		}
	}
	return 0, fmt.Errorf("connect: %v (%w)", last, ctx.Err())
}

func acceptanceFrame(status byte, content []byte) []byte {
	body := append([]byte("UTA2"), status)
	body = append(body, []byte("inert-acceptance-\xe2\x98\x83\x00")...)
	body = append(body, content...)
	frame := make([]byte, 4)
	binary.LittleEndian.PutUint32(frame, uint32(len(body))|pipeVersionFlag)
	return append(frame, body...)
}

func TestWindowsAcceptancePipeClient(t *testing.T) {
	system := acceptanceGate(t)
	incomplete := byte(pipeIncompleteAllow)
	if os.Getenv("EDR_ACCEPTANCE_FAIL_MODE") == "closed" {
		incomplete = pipeIncompleteBlock
	}
	cases := []struct {
		name     string
		frame    []byte
		want     byte
		fragment int
	}{
		{"clean", acceptanceFrame(pipeComplete, []byte("inert clean buffer")), pipeAllow, 0},
		{"marker", acceptanceFrame(pipeComplete, []byte(acceptanceMarker)), pipeBlock, 0},
		{"fragmented-marker", acceptanceFrame(pipeComplete, []byte(acceptanceMarker)), pipeBlock, 1},
		{"size-incomplete", acceptanceFrame(pipeSizeLimit, nil), incomplete, 0},
		{"read-incomplete", acceptanceFrame(pipeReadFailure, nil), incomplete, 0},
		{"bad-magic", append([]byte{8, 0, 0, 128}, []byte("NOPE\x00a\x00x")...), incomplete, 0},
		{"oversized-length", []byte{255, 255, 255, 255}, incomplete, 0},
	}
	legacy := append([]byte("legacy\x00"), []byte(acceptanceMarker)...)
	header := make([]byte, 4)
	binary.LittleEndian.PutUint32(header, uint32(len(legacy)))
	cases = append(cases, struct {
		name     string
		frame    []byte
		want     byte
		fragment int
	}{"legacy-complete", append(header, legacy...), pipeBlock, 0})
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			h, err := acceptanceConnect(ctx)
			if err != nil {
				t.Fatal(err)
			}
			defer windows.CloseHandle(h)
			owner, err := windows.GetSecurityInfo(h, windows.SE_KERNEL_OBJECT, windows.OWNER_SECURITY_INFORMATION)
			if err != nil {
				t.Fatal(err)
			}
			sid, _, err := owner.Owner()
			if err != nil || !sid.IsWellKnown(windows.WinLocalSystemSid) {
				t.Fatalf("owner is not SYSTEM: %v %v", sid, err)
			}
			stream := pipeStream{ctx: ctx, h: h}
			for offset := 0; offset < len(tc.frame); {
				end := len(tc.frame)
				if tc.fragment > 0 && offset+tc.fragment < end {
					end = offset + tc.fragment
				}
				if _, err := stream.Write(tc.frame[offset:end]); err != nil {
					t.Fatal(err)
				}
				offset = end
			}
			var response [1]byte
			if _, err := io.ReadFull(stream, response[:]); err != nil {
				t.Fatal(err)
			}
			if response[0] != tc.want {
				t.Fatalf("reply=%d want=%d", response[0], tc.want)
			}
		})
	}
	t.Run("stalled-frame-reports-incomplete-and-recovers", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 35*time.Second)
		defer cancel()
		h, err := acceptanceConnect(ctx)
		if err != nil {
			t.Fatal(err)
		}
		stream := pipeStream{ctx: ctx, h: h}
		// Announce a complete request, then withhold its payload without closing
		// the connection. Production's 30s deadline must cancel the native read
		// and still deliver an explicit incomplete reply, never a clean result.
		if _, err := stream.Write([]byte{10, 0, 0, 128}); err != nil {
			windows.CloseHandle(h)
			t.Fatal(err)
		}
		var response [1]byte
		_, err = io.ReadFull(stream, response[:])
		windows.CloseHandle(h)
		if err != nil || response[0] != incomplete {
			t.Fatalf("timeout reply=%d err=%v want=%d", response[0], err, incomplete)
		}
		// A new valid client must work after the cancelled overlapped operation.
		h, err = acceptanceConnect(ctx)
		if err != nil {
			t.Fatal(err)
		}
		defer windows.CloseHandle(h)
		stream.h = h
		if _, err := stream.Write(acceptanceFrame(pipeComplete, []byte("inert after timeout"))); err != nil {
			t.Fatal(err)
		}
		if _, err := io.ReadFull(stream, response[:]); err != nil || response[0] != pipeAllow {
			t.Fatalf("recovery=%d err=%v", response[0], err)
		}
	})
	if !system {
		t.Run("standard-user-cannot-create-second-instance", func(t *testing.T) {
			name, _ := windows.UTF16PtrFromString(PipeName)
			h, err := windows.CreateNamedPipe(name, windows.PIPE_ACCESS_DUPLEX, windows.PIPE_TYPE_BYTE|windows.PIPE_REJECT_REMOTE_CLIENTS, 1, 256, 256, 0, nil)
			if err == nil {
				windows.CloseHandle(h)
				t.Fatal("standard user created a competing production pipe")
			}
			t.Logf("competing instance rejected: %v", err)
		})
	}
}

func TestWindowsAcceptanceNativeUntrustedOwner(t *testing.T) {
	if acceptanceGate(t) {
		t.Fatal("untrusted-owner test requires a standard user")
	}
	probe, dll := os.Getenv("EDR_ACCEPTANCE_NATIVE_PROBE"), os.Getenv("EDR_ACCEPTANCE_DLL")
	if probe == "" || dll == "" {
		t.Fatal("native probe and DLL required")
	}
	name, _ := windows.UTF16PtrFromString(PipeName)
	h, err := windows.CreateNamedPipe(name, windows.PIPE_ACCESS_DUPLEX|windows.FILE_FLAG_OVERLAPPED|windows.FILE_FLAG_FIRST_PIPE_INSTANCE, windows.PIPE_TYPE_BYTE|windows.PIPE_REJECT_REMOTE_CLIENTS, 1, 256, 256, 0, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer windows.CloseHandle(h)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	readDone := make(chan error, 1)
	go func() {
		_, err := pipeOperation(ctx, h, func(o *windows.Overlapped, _ *uint32) error {
			e := windows.ConnectNamedPipe(h, o)
			if e == windows.ERROR_PIPE_CONNECTED {
				return nil
			}
			return e
		})
		if err == nil {
			var b [1]byte
			n, e := (pipeStream{ctx: ctx, h: h}).Read(b[:])
			if n != 0 {
				err = fmt.Errorf("untrusted pipe received content: %x", b[:n])
			} else if e != nil && e != io.EOF {
				err = e
			}
		}
		readDone <- err
	}()
	out, err := exec.CommandContext(ctx, probe, dll, "transport-failure").CombinedOutput()
	if err != nil {
		t.Fatalf("native owner validation: %v\n%s", err, out)
	}
	if !bytes.Contains(out, []byte("PASS transport-failure")) {
		t.Fatalf("missing native owner assertion: %s", out)
	}
	if err := <-readDone; err != nil && !strings.Contains(err.Error(), "pipe is being closed") {
		t.Fatal(err)
	}
	t.Logf("%s", out)
}
