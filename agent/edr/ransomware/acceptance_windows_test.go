//go:build windows && edr_acceptance

package ransomware

import (
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/utmstack/UTMStack/agent/edr/config"
	"github.com/utmstack/UTMStack/agent/edr/proctable"
	"github.com/utmstack/UTMStack/agent/edr/responder"
)

func TestWindowsAcceptanceRansomwareOwnedWriter(t *testing.T) {
	if path := os.Getenv("EDR_WINDOWS_RR_WRITER_PATH"); path != "" {
		// The parent captures this exact child's creation identity before allowing
		// its harmless write. No unrelated process or executable is targeted.
		var gate [1]byte
		if _, err := io.ReadFull(os.Stdin, gate[:]); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte("ordinary controlled canary edit"), 0600); err != nil {
			t.Fatal(err)
		}
		time.Sleep(30 * time.Second)
	}
}

type acceptanceETWFeed struct {
	*etwFeed
	observed chan FileEvent
}

func (f acceptanceETWFeed) Run(ctx context.Context, sink func(FileEvent)) error {
	return f.etwFeed.Run(ctx, func(fe FileEvent) {
		select {
		case f.observed <- fe:
		default:
		}
		sink(fe)
	})
}

type acceptanceRansomwareEvents struct{ contained chan string }

func (s acceptanceRansomwareEvents) Append(line string) error {
	if strings.Contains(line, "ransomware_contained") {
		select {
		case s.contained <- line:
		default:
		}
	}
	return nil
}

func TestWindowsAcceptanceETWCanaryContainsOnlyOwnedChild(t *testing.T) {
	if os.Getenv("EDR_WINDOWS_ACCEPTANCE") != "isolated-vm" {
		t.Skip("disposable Windows VM only")
	}
	canary := filepath.Join(canaryTestDir(t), "owned-etw-canary.txt")
	if err := os.WriteFile(canary, []byte("ordinary inert canary fixture"), 0600); err != nil {
		t.Fatal(err)
	}
	feed := acceptanceETWFeed{etwFeed: NewFeed().(*etwFeed), observed: make(chan FileEvent, 32)}
	var diagnosticsMu sync.Mutex
	ids := map[uint16]uint64{}
	samples := []kernelFileRecord{}
	var ownedPID atomic.Int64
	feed.nativeRecord = func(id uint16) {
		diagnosticsMu.Lock()
		defer diagnosticsMu.Unlock()
		if len(ids) < 64 || ids[id] > 0 {
			ids[id]++
		}
	}
	feed.nativeDecoded = func(r kernelFileRecord) {
		if r.PID != int(ownedPID.Load()) && !strings.Contains(strings.ToLower(r.Path), "owned-etw-canary") {
			return
		}
		diagnosticsMu.Lock()
		defer diagnosticsMu.Unlock()
		if len(samples) < 16 {
			if len(r.Path) > 512 {
				r.Path = r.Path[:512]
			}
			samples = append(samples, r)
		}
	}
	t.Cleanup(func() {
		diagnosticsMu.Lock()
		defer diagnosticsMu.Unlock()
		t.Logf("bounded ETW diagnostics: native_ids=%v owned_records=%+v", ids, samples)
	})
	table := proctable.New()
	contained := make(chan string, 1)
	q := &fakeQuar{}
	cfg := config.Default()
	cfg.Ransomware.Enabled = true
	cfg.Ransomware.ResponseMode = "kill"
	canaries := NewManager(nil)
	canaries.set[normCanaryPath(canary)] = true
	g := NewGuard(GuardDeps{Cfg: cfg, Table: table, Resp: responder.New(table, nil, responder.OSKiller{}, responder.OSSuspender{}), Spool: acceptanceRansomwareEvents{contained}, Quar: q, Incidents: &fakeIncidents{}, Canaries: canaries})
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- g.Run(ctx, feed, os.Getpid(), nil) }()
	defer func() {
		cancel()
		select {
		case err := <-done:
			if !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded) {
				t.Errorf("ETW guard shutdown: %v", err)
			}
		case <-time.After(3 * time.Second):
			t.Error("ETW guard failed shutdown")
		}
	}()
	for !g.FeedHealthy() {
		// Trigger only owned harmless I/O so readiness proves native delivery,
		// independently of unrelated background activity on an otherwise idle VM.
		if err := os.WriteFile(filepath.Join(filepath.Dir(canary), "owned-warmup.tmp"), []byte("inert warmup"), 0600); err != nil {
			t.Fatal(err)
		}
		select {
		case <-ctx.Done():
			t.Fatal("native ETW startup", feed.Health())
		case <-time.After(10 * time.Millisecond):
		}
	}
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	child := exec.Command(exe, "-test.run=^TestWindowsAcceptanceRansomwareOwnedWriter$")
	child.Env = append(os.Environ(), "EDR_WINDOWS_RR_WRITER_PATH="+canary)
	gate, err := child.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}
	defer gate.Close()
	if err := child.Start(); err != nil {
		t.Fatal(err)
	}
	ownedPID.Store(int64(child.Process.Pid))
	defer func() { _ = child.Process.Kill(); _ = child.Wait() }()
	identity, err := (responder.OSKiller{}).ProcessIdentity(child.Process.Pid)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := gate.Write([]byte{1}); err != nil {
		t.Fatal(err)
	}
	select {
	case fe := <-feed.observed:
		if fe.PID != child.Process.Pid || fe.OSIdentity != identity || !fe.IdentityRequired || !strings.EqualFold(fe.Path, canary) || fe.Image == "" {
			t.Fatalf("ETW original attribution invalid: %+v; own pid=%d creation=%d", fe, child.Process.Pid, identity)
		}
		t.Logf("native Kernel-File ETW observed owned child pid=%d creation=%d op=%v", fe.PID, fe.OSIdentity, fe.Op)
	case <-ctx.Done():
		t.Fatal("native ETW did not observe owned canary modification", feed.Health())
	}
	select {
	case message := <-contained:
		if !strings.Contains(message, "canary:") {
			t.Fatalf("unexpected response signal: %s", message)
		}
	case <-ctx.Done():
		t.Fatal("owned canary writer not contained", feed.Health())
	}
	if err := child.Wait(); err == nil {
		t.Fatal("owned child was not terminated")
	}
	if q.calls != 0 {
		t.Fatal("behavior response quarantined a non-pinned executable pathname")
	}
}
