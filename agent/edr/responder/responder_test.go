package responder

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/utmstack/UTMStack/agent/edr/event"
	"github.com/utmstack/UTMStack/agent/edr/proctable"
)

type fakeKiller struct{ order []int }

func (f *fakeKiller) KillPID(pid int) error { f.order = append(f.order, pid); return nil }

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
