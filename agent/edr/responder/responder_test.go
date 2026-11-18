package responder

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/utmstack/UTMStack/agent/edr/event"
	"github.com/utmstack/UTMStack/agent/edr/proctable"
)

type fakeKiller struct {
	mu    sync.Mutex
	order []int
	fail  map[int]error // pid -> error returned once by its next kill
}

func (f *fakeKiller) KillPID(pid int) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if err, ok := f.fail[pid]; ok {
		delete(f.fail, pid)
		return err
	}
	f.order = append(f.order, pid)
	return nil
}

func (f *fakeKiller) kills(pid int) int {
	f.mu.Lock()
	defer f.mu.Unlock()
	n := 0
	for _, p := range f.order {
		if p == pid {
			n++
		}
	}
	return n
}

// identityKiller is a Killer that also names process instances, like the
// Windows killer does with the process creation time.
type identityKiller struct {
	fakeKiller
	ids        map[int]int64 // pid -> identity of the process holding it now
	beforeKill func()        // runs between the lookup and the kill
	afterKill  func(pid int) // runs once the kill succeeded
	byIdentity []int64
}

func (f *identityKiller) ProcessIdentity(pid int) (int64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	id, ok := f.ids[pid]
	if !ok {
		return 0, ErrProcessGone
	}
	return id, nil
}

func (f *identityKiller) KillIdentity(pid int, identity int64) error {
	if f.beforeKill != nil {
		f.beforeKill()
	}
	f.mu.Lock()
	cur, ok := f.ids[pid]
	f.mu.Unlock()
	if !ok {
		return ErrProcessGone
	}
	if cur != identity {
		return errPIDReused
	}
	f.mu.Lock()
	f.byIdentity = append(f.byIdentity, identity)
	f.mu.Unlock()
	if err := f.KillPID(pid); err != nil {
		return err
	}
	if f.afterKill != nil {
		f.afterKill(pid)
	}
	return nil
}

func newTestSpool(t *testing.T) (*event.Spool, string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "e.ndjson")
	sp, err := event.OpenSpool(path, 1<<20)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { sp.Close() })
	return sp, path
}

func killedEvents(t *testing.T, path string) int {
	t.Helper()
	b, _ := os.ReadFile(path)
	return strings.Count(string(b), `"action":"killed"`)
}

func TestKillTreeLeavesFirstAndEmits(t *testing.T) {
	tab := proctable.New()
	tab.Add(proctable.Proc{PID: 100, PPID: 4, Image: `C:\a.exe`})
	tab.Add(proctable.Proc{PID: 200, PPID: 100, Image: `C:\b.exe`})

	spoolPath := filepath.Join(t.TempDir(), "e.ndjson")
	sp, err := event.OpenSpool(spoolPath, 1<<20)
	if err != nil {
		t.Fatal(err)
	}
	defer sp.Close()

	fk := &fakeKiller{}
	r := New(tab, sp, fk, nil)
	killed, err := r.KillTree(100, "Win.Test.EICAR")
	if err != nil {
		t.Fatalf("KillTree: %v", err)
	}
	if len(killed) != 2 || fk.order[0] != 200 || fk.order[len(fk.order)-1] != 100 {
		t.Fatalf("kill order = %v (child 200 must precede parent 100)", fk.order)
	}

	b, _ := os.ReadFile(spoolPath)
	out := string(b)
	if !strings.Contains(out, `"action":"killed"`) || !strings.Contains(out, `"product":"UTMStack EDR"`) {
		t.Fatalf("kill events not branded: %s", out)
	}
	if strings.Contains(strings.ToLower(out), "clam") {
		t.Fatalf("kill event leaked engine name: %s", out)
	}
}

func TestKillTreeKillsEachPidOnce(t *testing.T) {
	tab := proctable.New()
	tab.Add(proctable.Proc{PID: 100, PPID: 4, Image: `C:\a.exe`})
	tab.Add(proctable.Proc{PID: 200, PPID: 100, Image: `C:\a.exe`})
	sp, _ := event.OpenSpool(filepath.Join(t.TempDir(), "e.ndjson"), 1<<20)
	defer sp.Close()

	fk := &fakeKiller{}
	r := New(tab, sp, fk, nil)
	r.KillTree(100, "sig") // kills 200 (child) then 100
	r.KillTree(200, "sig") // redundant: 200 already claimed → skipped

	count := 0
	for _, p := range fk.order {
		if p == 200 {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("pid 200 terminated %d times, want exactly 1 (kill-once)", count)
	}
}

type fakeSusp struct{ suspended, resumed int }

func (f *fakeSusp) SuspendPID(pid int) error { f.suspended = pid; return nil }
func (f *fakeSusp) ResumePID(pid int) error  { f.resumed = pid; return nil }

func TestSuspendResumeDelegates(t *testing.T) {
	fs := &fakeSusp{}
	r := New(proctable.New(), nil, &fakeKiller{}, fs)
	_ = r.Suspend(55)
	_ = r.Resume(55)
	if fs.suspended != 55 || fs.resumed != 55 {
		t.Fatalf("suspend/resume not delegated: %+v", fs)
	}
}

func TestKillTreeRetriesAfterFailedKill(t *testing.T) {
	tab := proctable.New()
	tab.Add(proctable.Proc{PID: 100, PPID: 4, Image: `C:\a.exe`, StartTS: 1})
	sp, path := newTestSpool(t)
	fk := &fakeKiller{fail: map[int]error{100: errors.New("access denied")}}
	r := New(tab, sp, fk, nil)

	if killed, _ := r.KillTree(100, "sig"); len(killed) != 0 {
		t.Fatalf("first attempt reported %v killed, but the kill failed", killed)
	}
	killed, _ := r.KillTree(100, "sig")
	if len(killed) != 1 || killed[0] != 100 {
		t.Fatalf("retry after a failed kill: killed = %v, want [100]", killed)
	}
	if n := killedEvents(t, path); n != 1 {
		t.Fatalf("%d killed events, want 1", n)
	}
}

func TestKillTreeKillsReusedPIDByTableGeneration(t *testing.T) {
	tab := proctable.New()
	tab.Add(proctable.Proc{PID: 200, PPID: 4, Image: `C:\a.exe`, StartTS: 1})
	sp, path := newTestSpool(t)
	fk := &fakeKiller{}
	r := New(tab, sp, fk, nil)

	if killed, _ := r.KillTree(200, "sig"); len(killed) != 1 {
		t.Fatalf("first process not killed: %v", killed)
	}
	// A new process reuses PID 200 and the process table records it.
	tab.Add(proctable.Proc{PID: 200, PPID: 4, Image: `C:\b.exe`, StartTS: 2})
	if killed, _ := r.KillTree(200, "sig"); len(killed) != 1 {
		t.Fatalf("process reusing a killed PID was not killed: %v", killed)
	}
	if n := fk.kills(200); n != 2 {
		t.Fatalf("pid 200 terminated %d times, want 2 (two processes)", n)
	}
	if n := killedEvents(t, path); n != 2 {
		t.Fatalf("%d killed events, want 2", n)
	}
}

func TestKillTreeKillsReusedPIDByOSIdentity(t *testing.T) {
	sp, _ := newTestSpool(t)
	ik := &identityKiller{ids: map[int]int64{300: 111}}
	r := New(proctable.New(), sp, ik, nil)

	if killed, _ := r.KillTree(300, "sig"); len(killed) != 1 {
		t.Fatalf("first process not killed: %v", killed)
	}
	ik.mu.Lock()
	ik.ids[300] = 222 // a new process now holds PID 300
	ik.mu.Unlock()
	if killed, _ := r.KillTree(300, "sig"); len(killed) != 1 {
		t.Fatalf("process reusing a killed PID was not killed: %v", killed)
	}
	if fmt.Sprint(ik.byIdentity) != "[111 222]" {
		t.Fatalf("killed identities = %v, want [111 222]", ik.byIdentity)
	}
}

func TestKillTreeSameIdentityKilledOnce(t *testing.T) {
	sp, path := newTestSpool(t)
	ik := &identityKiller{ids: map[int]int64{300: 111}}
	r := New(proctable.New(), sp, ik, nil)

	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = r.KillTree(300, "sig")
		}()
	}
	wg.Wait()
	// The process is still listed (it has not exited yet): a later request for
	// the same process within the claim window is also skipped.
	_, _ = r.KillTree(300, "sig")
	if n := ik.kills(300); n != 1 {
		t.Fatalf("process terminated %d times, want exactly 1 (kill-once)", n)
	}
	if n := killedEvents(t, path); n != 1 {
		t.Fatalf("%d killed events, want 1", n)
	}
}

func TestKillTreeUnknownIdentityClaimExpires(t *testing.T) {
	tab := proctable.New()
	tab.Add(proctable.Proc{PID: 400, PPID: 4, Image: `C:\a.exe`}) // generation unknown
	sp, _ := newTestSpool(t)
	fk := &fakeKiller{}
	r := New(tab, sp, fk, nil)
	now := time.Unix(1_000_000, 0)
	r.now = func() time.Time { return now }

	_, _ = r.KillTree(400, "sig")
	now = now.Add(claimTTL / 2)
	_, _ = r.KillTree(400, "sig") // same PID, identity unknown: still claimed
	if n := fk.kills(400); n != 1 {
		t.Fatalf("pid 400 terminated %d times inside the claim window, want 1", n)
	}
	now = now.Add(claimTTL)
	_, _ = r.KillTree(400, "sig")
	if n := fk.kills(400); n != 2 {
		t.Fatalf("pid 400 terminated %d times after the claim expired, want 2", n)
	}
}

func TestKillTreeDoesNotKillProcessThatReplacedTarget(t *testing.T) {
	tab := proctable.New()
	sp, path := newTestSpool(t)
	ik := &identityKiller{ids: map[int]int64{500: 111}}
	ik.beforeKill = func() {
		// The target exits and a new process takes its PID between the
		// identity lookup and the kill; the table records the new process.
		ik.mu.Lock()
		ik.ids[500] = 222
		ik.mu.Unlock()
		tab.Add(proctable.Proc{PID: 500, PPID: 4, Image: `C:\new.exe`, StartTS: 9})
	}
	r := New(tab, sp, ik, nil)

	if killed, _ := r.KillTree(500, "sig"); len(killed) != 0 {
		t.Fatalf("killed %v: the process holding the PID is not the target", killed)
	}
	if n := ik.kills(500); n != 0 {
		t.Fatalf("replacement process terminated %d times", n)
	}
	if n := killedEvents(t, path); n != 0 {
		t.Fatalf("%d killed events, want 0", n)
	}
	if p, ok := tab.Get(500); !ok || p.Image != `C:\new.exe` {
		t.Fatalf("replacement process dropped from the table: %+v %v", p, ok)
	}
}

func TestKillTreeEventKeepsContextWhenRacingPathSeesProcessGone(t *testing.T) {
	tab := proctable.New()
	tab.Add(proctable.Proc{PID: 700, PPID: 4, Image: `C:\evil.exe`, Cmdline: "evil.exe -x", StartTS: 1})
	sp, path := newTestSpool(t)
	ik := &identityKiller{ids: map[int]int64{700: 111}}
	var r *Responder
	ik.afterKill = func(pid int) {
		ik.afterKill = nil
		// The killed process exits at once, and another path (e.g. the
		// child's own guard racing a parent's tree kill) reaches it before
		// this kill reports it: that path finds it gone and drops its entry.
		ik.mu.Lock()
		delete(ik.ids, pid)
		ik.mu.Unlock()
		_, _ = r.KillTree(pid, "sig")
	}
	r = New(tab, sp, ik, nil)

	if killed, _ := r.KillTree(700, "sig"); len(killed) != 1 {
		t.Fatalf("killed = %v, want [700]", killed)
	}
	b, _ := os.ReadFile(path)
	lines := strings.Split(strings.TrimSpace(string(b)), "\n")
	if len(lines) != 1 {
		t.Fatalf("%d events, want 1: %s", len(lines), b)
	}
	var ev event.Event
	if err := json.Unmarshal([]byte(lines[0]), &ev); err != nil {
		t.Fatal(err)
	}
	if ev.Process == nil || ev.Process.Image != `C:\evil.exe` || ev.Process.Cmdline != "evil.exe -x" || ev.Process.PPID != 4 {
		t.Fatalf("killed event lost the process context: %s", lines[0])
	}
}

func TestKillTreeSkipsProcessGoneBeforeLookup(t *testing.T) {
	tab := proctable.New()
	tab.Add(proctable.Proc{PID: 600, PPID: 4, Image: `C:\a.exe`, StartTS: 1})
	sp, path := newTestSpool(t)
	ik := &identityKiller{ids: map[int]int64{}}
	r := New(tab, sp, ik, nil)

	if killed, _ := r.KillTree(600, "sig"); len(killed) != 0 {
		t.Fatalf("killed %v, but the process was already gone", killed)
	}
	if n := ik.kills(600); n != 0 {
		t.Fatalf("kill attempted %d times for a gone process", n)
	}
	if n := killedEvents(t, path); n != 0 {
		t.Fatalf("%d killed events, want 0", n)
	}
	if _, ok := tab.Get(600); ok {
		t.Fatal("stale table entry for a gone process was kept")
	}
}

func TestClaimSetIsBounded(t *testing.T) {
	var c claimSet
	now := time.Unix(1_000_000, 0)
	for i := 0; i < maxClaims+10; i++ {
		now = now.Add(time.Millisecond)
		if !c.take(processKey{pid: i, id: 1, src: identityTable}, now) {
			t.Fatalf("fresh claim %d refused", i)
		}
	}
	if len(c.at) > maxClaims {
		t.Fatalf("%d claims kept, want at most %d", len(c.at), maxClaims)
	}
	if c.take(processKey{pid: maxClaims + 9, id: 1, src: identityTable}, now) {
		t.Fatal("newest claim was evicted")
	}
	if !c.take(processKey{pid: 0, id: 1, src: identityTable}, now) {
		t.Fatal("oldest claim was not evicted")
	}
}

func TestClaimSetDecisions(t *testing.T) {
	var c claimSet
	now := time.Unix(1_000_000, 0)
	a := processKey{pid: 7, id: 111, src: identityOS}
	if !c.take(a, now) || c.take(a, now) {
		t.Fatal("an identity must be claimable exactly once")
	}
	if !c.take(processKey{pid: 7, id: 222, src: identityOS}, now) {
		t.Fatal("a different process on the same PID must be claimable")
	}
	if c.take(processKey{pid: 7}, now) {
		t.Fatal("unknown identity must defer to a live claim on the same PID")
	}
	if !c.take(processKey{pid: 8}, now) {
		t.Fatal("unknown identity on an unclaimed PID must be claimable")
	}
	c.release(a)
	if !c.take(a, now) {
		t.Fatal("a released claim must be claimable again")
	}
	if c.take(a, now.Add(claimTTL-time.Nanosecond)) {
		t.Fatal("claim expired early")
	}
	if !c.take(a, now.Add(claimTTL)) {
		t.Fatal("claim did not expire")
	}
}
