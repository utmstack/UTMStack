package ransomware

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/utmstack/UTMStack/agent/edr/event"
	"github.com/utmstack/UTMStack/agent/edr/proctable"
)

// liveFunc adapts a func to the liveImager interface for tests.
type liveFunc func(int) (string, error)

func (f liveFunc) Image(pid int) (string, error) { return f(pid) }

func containedEvents(app *fakeAppender) int {
	n := 0
	for _, l := range app.lines {
		var e event.Event
		if json.Unmarshal([]byte(l), &e) == nil && e.Action == event.ActionRansomwareContained {
			n++
		}
	}
	return n
}

// The containment event must name the process even when the
// process table has not caught up with the PID. The guard resolves the image
// from the live snapshot (deps.LiveImage) in that case.
func TestGuardContainmentResolvesImageFromLiveSnapshot(t *testing.T) {
	const img = `C:\temp\crypter.exe`
	g, resp, app, quar, inc, _ := newTestGuard("kill", map[string]bool{`C:\canary\a`: true}, 120, 160)

	g.deps.LiveImage = liveFunc(func(int) (string, error) { return img, nil })

	// Two canary writes by a PID the table has never seen (table miss).
	fe := FileEvent{Path: `C:\canary\a`, PID: 4242, Op: OpWrite}
	g.OnFileEvent(fe)
	g.OnFileEvent(fe)

	if len(resp.killed) == 0 {
		t.Fatal("expected a kill, got none")
	}
	for _, l := range app.lines {
		var e event.Event
		if json.Unmarshal([]byte(l), &e) == nil && e.Action == event.ActionRansomwareContained {
			if e.Process == nil || e.Process.Image != img {
				t.Fatalf("containment event image = %v, want %q", e.Process, img)
			}
		}
	}
	if containedEvents(app) != 1 {
		t.Fatalf("containment events = %d, want 1", containedEvents(app))
	}
	// Quarantine must have received the resolved image (not skipped).
	if quar.calls != 1 {
		t.Fatalf("quarantine calls = %d, want 1 (image must be resolved before quarantine)", quar.calls)
	}
	// Incident must record the image.
	if len(inc.recs) == 0 || inc.recs[len(inc.recs)-1].Image != img {
		t.Fatalf("incident image not recorded: %+v", inc.recs)
	}
}

// A dead live snapshot (process already reaped) must not break containment:
// the event carries whatever the table had and the kill/quarantine behaviour
// is unchanged.
func TestGuardContainmentLiveImageFailureKeepsTableValue(t *testing.T) {
	const img = `/tmp/crypter`
	g, resp, app, _, inc, tab := newTestGuard("kill", map[string]bool{`/tmp/canary`: true}, 120, 160)

	g.deps.LiveImage = liveFunc(func(int) (string, error) { return "", errors.New("no such process") })

	// Table knows the image; live lookup is not consulted (table hit).
	tab.Add(proctable.Proc{PID: 7, PPID: 1, Image: img, Cmdline: "crypter -all"})

	fe := FileEvent{Path: `/tmp/canary`, PID: 7, Op: OpWrite}
	g.OnFileEvent(fe)
	g.OnFileEvent(fe)

	if len(resp.killed) == 0 {
		t.Fatal("expected a kill, got none")
	}
	var found bool
	for _, l := range app.lines {
		var e event.Event
		if json.Unmarshal([]byte(l), &e) == nil && e.Action == event.ActionRansomwareContained {
			found = true
			if e.Process == nil || e.Process.Image != img {
				t.Fatalf("containment event image = %v, want %q (table value)", e.Process, img)
			}
		}
	}
	if !found {
		t.Fatalf("no containment event in spool: %v", app.lines)
	}
	if len(inc.recs) == 0 || inc.recs[len(inc.recs)-1].Image != img {
		t.Fatalf("incident image not recorded: %+v", inc.recs)
	}
}

// A straggling file event arriving for a PID that was just
// contained must not produce a second containment event. The scorer forgets
// the PID at kill time, so the guard itself remembers contained PIDs for a
// bounded window.
func TestGuardRepeatTripContainsOnce(t *testing.T) {
	g, resp, app, _, _, tab := newTestGuard("kill", map[string]bool{`C:\canary\a`: true}, 120, 160)
	tab.Add(proctable.Proc{PID: 9, PPID: 1, Image: `/tmp/crypter`, Cmdline: "crypter -all"})

	fe := FileEvent{Path: `C:\canary\a`, PID: 9, Op: OpWrite}
	// First trip: two writes escalate to kill → containment.
	g.OnFileEvent(fe)
	g.OnFileEvent(fe)

	// Straggler queued before the kill: another write for the same PID right
	// after containment.
	g.OnFileEvent(fe)

	if got := containedEvents(app); got != 1 {
		t.Fatalf("containment events = %d, want 1 (rapid repeat trip must contain once); spool=%s", got, strings.Join(app.lines, "\n"))
	}
	if got := len(resp.killed); got != 1 {
		t.Fatalf("kills = %d, want 1", got)
	}
}

// The suppression is bounded: after containedTTL the same PID may be
// contained again (a real second incident, or PID reuse after a long window).
func TestGuardContainmentSuppressionExpires(t *testing.T) {
	g, _, app, _, _, tab := newTestGuard("kill", map[string]bool{`C:\canary\a`: true}, 120, 160)
	tab.Add(proctable.Proc{PID: 11, PPID: 1, Image: `/tmp/crypter`, Cmdline: "crypter -all"})

	now := int64(1000)
	g.deps.Now = func() time.Time { return time.Unix(now, 0) }

	fe := FileEvent{Path: `C:\canary\a`, PID: 11, Op: OpWrite}
	g.OnFileEvent(fe)
	g.OnFileEvent(fe)
	if got := containedEvents(app); got != 1 {
		t.Fatalf("first containment missing: %d", got)
	}

	// Advance past the suppression window.
	now += 31
	g.OnFileEvent(fe)
	g.OnFileEvent(fe)
	if got := containedEvents(app); got != 2 {
		t.Fatalf("containment events after TTL = %d, want 2 (suppression must expire)", got)
	}
}
