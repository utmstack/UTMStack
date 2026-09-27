//go:build linux

// Live test: does this kernel deliver CN_PROC PROC_EVENT_EXEC events to the
// production watcher? Needs root (NETLINK_CONNECTOR is privileged). Skips
// cleanly otherwise so `go test -short` and CI stay green.
//
//	sudo go test -run TestCNProcLive -v ./edr/procwatch
package procwatch

import (
	"context"
	"os"
	"os/exec"
	"testing"
	"time"
)

// recordingHandler implements Handler and records calls into a channel.
type recordingHandler struct{ ch chan ProcStart }

func (r *recordingHandler) OnProcStart(ps ProcStart) { r.ch <- ps }

func TestCNProcLive(t *testing.T) {
	if testing.Short() {
		t.Skip("live test skipped in -short mode")
	}
	if os.Geteuid() != 0 {
		t.Skip("needs root")
	}

	// Handler that records OnProcStart calls.
	events := make(chan ProcStart, 16)
	w := New(&recordingHandler{ch: events})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan struct{})
	go func() {
		w.Run(ctx)
		close(done)
	}()

	// Give the subscription a moment, then spawn /bin/true and remember its
	// PID: the assertion must match this specific process, not any exec on a
	// busy machine (the original test accepted the first event, which was
	// often unrelated system activity).
	time.Sleep(500 * time.Millisecond)
	cmd := exec.Command("/bin/true")
	if err := cmd.Start(); err != nil {
		t.Fatalf("spawn /bin/true: %v", err)
	}
	wantPID := cmd.Process.Pid
	// Reap it; the exec event is already in flight (or in the socket
	// buffer) and carries the PID, so delivery does not depend on the
	// process still being alive.
	if err := cmd.Wait(); err != nil {
		t.Fatalf("wait /bin/true: %v", err)
	}

	// Drain events until we see wantPID. Other execs on a busy machine
	// arrive first; draining also keeps the recording handler from
	// blocking on its full channel.
	deadline := time.Now().Add(5 * time.Second)
	found := false
	for !found {
		select {
		case ps := <-events:
			if ps.PID == 0 {
				t.Fatalf("OnProcStart fired with PID 0")
			}
			if ps.PID == wantPID {
				t.Logf("live exec event for /bin/true: pid=%d ppid=%d image=%q cmdline=%q", ps.PID, ps.PPID, ps.Image, ps.Cmdline)
				found = true
			}
		case <-time.After(time.Until(deadline)):
			t.Fatalf("no OnProcStart for /bin/true (pid %d) within 5s — CN_PROC exec not delivered", wantPID)
		}
	}

	cancel()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("Run did not return after ctx cancel")
	}
}
