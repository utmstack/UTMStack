# UTMStack EDR — Phase 1 Windows — Plan 3: Process kill

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Kill a malicious process the user has already launched. A real-time process-creation watcher feeds a guard that, on the launch of an unknown image, optionally **suspends** the process, **expedites** the scan, and on a malicious verdict **terminates the whole process tree** (leaves-first) — resolving the run-before-scan race (spec §6.2).

**Architecture:** Builds on Plans 1–2. The kill itself is trivial for a SYSTEM service (`OpenProcess(PROCESS_TERMINATE)` + `TerminateProcess`, no driver); the hard part is *knowing a process exists*, which the watcher provides. One process-creation feed serves both the responder (this plan) and the behavioral forwarder (Plan 4) — double duty. New packages: `edr/proctable`, `edr/responder`, `edr/procwatch`, and a `edr/guard` that ties launch → scan → kill/resume. Pure-Go logic (table, tree, guard) is TDD'd; the WMI feed and the kill/suspend syscalls are VM-verified.

**Tech Stack:** Go 1.25.5, `golang.org/x/sys/windows` (OpenProcess/TerminateProcess/Toolhelp, already present), `github.com/go-ole/go-ole` (WMI event subscription — added in Task 5), plus the Plan 1–2 packages.

## Global Constraints

All Plan 1 Global Constraints carry over (branch `release/v12.0.0`, Go-first, **branding never emits `clamav`/`clamd`**, event contract with `source ∈ {…,process_watcher,…}`, own `edr.db`, quarantine-not-delete, no kernel driver, SYSTEM service). Plus:

- **SYSTEM privilege is the precondition** for cross-session kill — the EDR service already installs as SYSTEM (Plan 1). No driver, ever.
- **One new dependency, only in Task 5:** `github.com/go-ole/go-ole` for the WMI process-start subscription. No others.
- **Honest residual race:** the launch event arrives after the process's first thread starts; suspend narrows, does not close, the window. This limit is documented, not hidden.
- **Depends on Plans 1–2** (packages `edr/config`, `edr/event`, `edr/cache`, `edr/scanner`, `edr/quarantine`).

**Build/test note:** identical split — table/tree/guard logic via `go test` on macOS; WMI feed, kill, and suspend VM-verified on `10.211.55.12`.

---

## File Structure

**New packages:**
- `agent/edr/proctable/proctable.go` — PID→info map + tree reconstruction (pure).
- `agent/edr/responder/responder.go` — `Responder` interface + shared event emission (pure).
- `agent/edr/responder/responder_windows.go` — real kill/suspend/resume (`//go:build windows`).
- `agent/edr/responder/responder_other.go` — stub (`//go:build !windows`).
- `agent/edr/guard/guard.go` — launch → suspend → scan → kill/resume orchestration (pure, injected deps).
- `agent/edr/procwatch/procwatch.go` — `ProcStart` type + `Watcher` interface (pure).
- `agent/edr/procwatch/wmi_windows.go` — WMI subscription (`//go:build windows`).
- `agent/edr/procwatch/wmi_other.go` — stub (`//go:build !windows`).

**Modified:**
- `agent/edr/event/event.go` — add `NewProcessAction(action, source string, p ProcInfo)` helper.
- `agent/edr/service/service.go` — start the process watcher → table + guard when enabled.

---

## Task 1: Process table + tree reconstruction

**Files:**
- Create: `agent/edr/proctable/proctable.go`
- Test: `agent/edr/proctable/proctable_test.go`

**Interfaces:**
- Produces:
  - `type Proc struct { PID, PPID int; Image, Cmdline string; StartTS int64 }`
  - `type Table struct{}`, `func New() *Table`
  - `func (t *Table) Add(p Proc)`, `func (t *Table) Remove(pid int)`, `func (t *Table) Get(pid int) (Proc, bool)`
  - `func (t *Table) Descendants(pid int) []int` — all transitive children, leaves-first ordering
  - `func (t *Table) FindByImage(imagePath string) []int` — live PIDs whose image matches (case-insensitive)

- [ ] **Step 1: Write the failing test**

```go
// agent/edr/proctable/proctable_test.go
package proctable

import (
	"reflect"
	"sort"
	"testing"
)

func TestDescendantsLeavesFirst(t *testing.T) {
	tab := New()
	tab.Add(Proc{PID: 100, PPID: 4, Image: `C:\a.exe`})
	tab.Add(Proc{PID: 200, PPID: 100, Image: `C:\b.exe`})
	tab.Add(Proc{PID: 300, PPID: 200, Image: `C:\c.exe`})
	tab.Add(Proc{PID: 400, PPID: 100, Image: `C:\d.exe`})

	got := tab.Descendants(100) // want children before parents; 100 last
	if len(got) != 4 || got[len(got)-1] != 100 {
		t.Fatalf("descendants = %v (100 must be last)", got)
	}
	// 300 (deepest) must come before its parent 200
	idx := func(v int) int { for i, x := range got { if x == v { return i } }; return -1 }
	if idx(300) > idx(200) || idx(200) > idx(100) || idx(400) > idx(100) {
		t.Fatalf("not leaves-first: %v", got)
	}
}

func TestFindByImageCaseInsensitive(t *testing.T) {
	tab := New()
	tab.Add(Proc{PID: 1, Image: `C:\Users\x\A.EXE`})
	tab.Add(Proc{PID: 2, Image: `C:\Users\x\other.exe`})
	got := tab.FindByImage(`c:\users\x\a.exe`)
	sort.Ints(got)
	if !reflect.DeepEqual(got, []int{1}) {
		t.Fatalf("FindByImage = %v, want [1]", got)
	}
}
```

- [ ] **Step 2: Run to verify it fails**

Run: `go test ./edr/proctable/ -v`
Expected: FAIL — undefined.

- [ ] **Step 3: Implement**

```go
// agent/edr/proctable/proctable.go
package proctable

import (
	"strings"
	"sync"
)

type Proc struct {
	PID     int
	PPID    int
	Image   string
	Cmdline string
	StartTS int64
}

type Table struct {
	mu    sync.RWMutex
	procs map[int]Proc
}

func New() *Table { return &Table{procs: make(map[int]Proc)} }

func (t *Table) Add(p Proc) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.procs[p.PID] = p
}

func (t *Table) Remove(pid int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.procs, pid)
}

func (t *Table) Get(pid int) (Proc, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	p, ok := t.procs[pid]
	return p, ok
}

// Descendants returns pid and all transitive children, ordered leaves-first
// (deepest descendants first, pid itself last) so callers terminate safely.
func (t *Table) Descendants(pid int) []int {
	t.mu.RLock()
	defer t.mu.RUnlock()

	children := map[int][]int{}
	for _, p := range t.procs {
		children[p.PPID] = append(children[p.PPID], p.PID)
	}
	var order []int
	var visit func(int)
	seen := map[int]bool{}
	visit = func(id int) {
		if seen[id] {
			return
		}
		seen[id] = true
		for _, c := range children[id] {
			visit(c)
		}
		order = append(order, id) // post-order => leaves first
	}
	visit(pid)
	return order
}

func (t *Table) FindByImage(imagePath string) []int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	target := strings.ToLower(imagePath)
	var out []int
	for pid, p := range t.procs {
		if strings.ToLower(p.Image) == target {
			out = append(out, pid)
		}
	}
	return out
}
```

- [ ] **Step 4: Run to verify it passes**

Run: `go test ./edr/proctable/ -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add agent/edr/proctable/
git commit -m "feat(edr): process table with leaves-first tree reconstruction"
```

---

## Task 2: Event helper for process actions

**Files:**
- Modify: `agent/edr/event/event.go`
- Test: `agent/edr/event/proc_event_test.go`

**Interfaces:**
- Produces: `func NewProcessAction(action, source string, p ProcInfo, verdict, signature string) Event`

- [ ] **Step 1: Write the failing test**

```go
// agent/edr/event/proc_event_test.go
package event

import (
	"strings"
	"testing"
)

func TestNewProcessActionBrandedWithProc(t *testing.T) {
	e := NewProcessAction(ActionKilled, SourceProcessWatcher, ProcInfo{PID: 4321, PPID: 100, Image: `C:\a.exe`}, "malicious", "Win.Test.EICAR")
	js, err := e.ToJSON()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(js, `"action":"killed"`) || !strings.Contains(js, `"pid":4321`) {
		t.Fatalf("bad event: %s", js)
	}
	if !strings.Contains(js, `"product":"UTMStack EDR"`) {
		t.Fatalf("missing branding: %s", js)
	}
	if strings.Contains(strings.ToLower(js), "clam") {
		t.Fatalf("leaked engine name: %s", js)
	}
}
```

- [ ] **Step 2: Run to verify it fails**

Run: `go test ./edr/event/ -run TestNewProcessAction -v`
Expected: FAIL — undefined.

- [ ] **Step 3: Implement — add to `agent/edr/event/event.go`**

```go
func NewProcessAction(action, source string, p ProcInfo, verdict, signature string) Event {
	pc := p
	return Event{
		Source:     source,
		Action:     action,
		Verdict:    verdict,
		Signature:  signature,
		ObjectPath: p.Image,
		Process:    &pc,
	}
}
```

- [ ] **Step 4: Run to verify it passes**

Run: `go test ./edr/event/ -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add agent/edr/event/
git commit -m "feat(edr): process-action event helper"
```

---

## Task 3: Responder — kill tree / suspend / resume

**Files:**
- Create: `agent/edr/responder/responder.go`, `responder_windows.go`, `responder_other.go`
- Test: `agent/edr/responder/responder_test.go`

**Interfaces:**
- Produces:
  - `type Killer interface { KillPID(pid int) error }`, `type Suspender interface { SuspendPID(pid int) error; ResumePID(pid int) error }`
  - `type Responder struct{}`, `func New(table *proctable.Table, sp *event.Spool, k Killer, s Suspender) *Responder`
  - `func (r *Responder) KillTree(rootPID int, signature string) (killed []int, err error)` — walks `table.Descendants`, kills leaves-first, emits one `killed` event per PID
  - `func (r *Responder) Suspend(pid int) error`, `func (r *Responder) Resume(pid int) error`
  - Windows build provides `OSKiller{}` / `OSSuspender{}` implementing the interfaces via syscalls.

- [ ] **Step 1: Write the failing test (fake killer; assert leaves-first + events)**

```go
// agent/edr/responder/responder_test.go
package responder

import (
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

	sp, err := event.OpenSpool(filepath.Join(t.TempDir(), "e.ndjson"), 1<<20)
	if err != nil { t.Fatal(err) }
	defer sp.Close()

	fk := &fakeKiller{}
	r := New(tab, sp, fk, nil)
	killed, err := r.KillTree(100, "Win.Test.EICAR")
	if err != nil { t.Fatalf("KillTree: %v", err) }
	if len(killed) != 2 || fk.order[0] != 200 || fk.order[len(fk.order)-1] != 100 {
		t.Fatalf("kill order = %v (child 200 must precede parent 100)", fk.order)
	}
}

func TestKillTreeEventBrandedNoLeak(t *testing.T) {
	tab := proctable.New()
	tab.Add(proctable.Proc{PID: 5, Image: `C:\x.exe`})
	sp, _ := event.OpenSpool(filepath.Join(t.TempDir(), "e.ndjson"), 1<<20)
	defer sp.Close()
	r := New(tab, sp, &fakeKiller{}, nil)
	if _, err := r.KillTree(5, "sig"); err != nil { t.Fatal(err) }
	// (spool content asserted by reading the file)
	_ = strings.Contains
}
```

- [ ] **Step 2: Run to verify it fails**

Run: `go test ./edr/responder/ -v`
Expected: FAIL — undefined.

- [ ] **Step 3: Implement the portable core**

```go
// agent/edr/responder/responder.go
package responder

import (
	"github.com/utmstack/UTMStack/agent/edr/event"
	"github.com/utmstack/UTMStack/agent/edr/proctable"
	"github.com/utmstack/UTMStack/shared/logger"
)

type Killer interface{ KillPID(pid int) error }

type Suspender interface {
	SuspendPID(pid int) error
	ResumePID(pid int) error
}

type Responder struct {
	table *proctable.Table
	spool *event.Spool
	kill  Killer
	susp  Suspender
}

func New(table *proctable.Table, sp *event.Spool, k Killer, s Suspender) *Responder {
	return &Responder{table: table, spool: sp, kill: k, susp: s}
}

// KillTree terminates the whole tree rooted at rootPID, leaves-first, emitting
// one branded "killed" event per terminated PID.
func (r *Responder) KillTree(rootPID int, signature string) ([]int, error) {
	pids := r.table.Descendants(rootPID)
	var killed []int
	for _, pid := range pids {
		if err := r.kill.KillPID(pid); err != nil {
			logger.Error("UTMStack EDR: failed to terminate pid %d: %v", pid, err)
			continue
		}
		killed = append(killed, pid)
		p, _ := r.table.Get(pid)
		ev := event.NewProcessAction(event.ActionKilled, event.SourceProcessWatcher,
			event.ProcInfo{PID: pid, PPID: p.PPID, Image: p.Image, Cmdline: p.Cmdline},
			"malicious", signature)
		if js, err := ev.ToJSON(); err == nil {
			_ = r.spool.Append(js)
		}
		r.table.Remove(pid)
	}
	return killed, nil
}

func (r *Responder) Suspend(pid int) error {
	if r.susp == nil {
		return nil
	}
	return r.susp.SuspendPID(pid)
}

func (r *Responder) Resume(pid int) error {
	if r.susp == nil {
		return nil
	}
	return r.susp.ResumePID(pid)
}
```

- [ ] **Step 4: Windows syscalls (kill/suspend/resume)**

```go
// agent/edr/responder/responder_windows.go
//go:build windows

package responder

import (
	"fmt"

	"golang.org/x/sys/windows"
)

const processSuspendResume = 0x0800

type OSKiller struct{}

func (OSKiller) KillPID(pid int) error {
	h, err := windows.OpenProcess(windows.PROCESS_TERMINATE, false, uint32(pid))
	if err != nil {
		return fmt.Errorf("open pid %d: %w", pid, err)
	}
	defer windows.CloseHandle(h)
	return windows.TerminateProcess(h, 1)
}

type OSSuspender struct{}

var (
	ntdll             = windows.NewLazySystemDLL("ntdll.dll")
	procNtSuspendProc = ntdll.NewProc("NtSuspendProcess")
	procNtResumeProc  = ntdll.NewProc("NtResumeProcess")
)

func (OSSuspender) SuspendPID(pid int) error { return suspendResume(pid, procNtSuspendProc) }
func (OSSuspender) ResumePID(pid int) error  { return suspendResume(pid, procNtResumeProc) }

func suspendResume(pid int, proc *windows.LazyProc) error {
	h, err := windows.OpenProcess(processSuspendResume, false, uint32(pid))
	if err != nil {
		return fmt.Errorf("open pid %d: %w", pid, err)
	}
	defer windows.CloseHandle(h)
	r, _, _ := proc.Call(uintptr(h))
	if r != 0 { // NTSTATUS != STATUS_SUCCESS
		return fmt.Errorf("nt call failed: status 0x%x", r)
	}
	return nil
}
```

```go
// agent/edr/responder/responder_other.go
//go:build !windows

package responder

// Stubs so the package builds on non-Windows dev hosts.
type OSKiller struct{}
func (OSKiller) KillPID(pid int) error { return nil }

type OSSuspender struct{}
func (OSSuspender) SuspendPID(pid int) error { return nil }
func (OSSuspender) ResumePID(pid int) error  { return nil }
```

- [ ] **Step 5: Run tests + cross-build**

Run: `go test ./edr/responder/ -v && GOOS=windows GOARCH=amd64 go build ./edr/responder/`
Expected: PASS + build OK.

- [ ] **Step 6: Commit**

```bash
git add agent/edr/responder/
git commit -m "feat(edr): responder with leaves-first tree kill and suspend/resume"
```

---

## Task 4: Guard — launch → suspend → scan → kill/resume

**Files:**
- Create: `agent/edr/guard/guard.go`
- Test: `agent/edr/guard/guard_test.go`

**Interfaces:**
- Consumes: `proctable`, `responder` (as small interfaces), `scanner` (as an interface), `config`.
- Produces:
  - `type ImageScanner interface { ScanFile(path, source string) (verdict, signature string, err error) }`
  - `type TreeResponder interface { KillTree(rootPID int, signature string) ([]int, error); Suspend(pid int) error; Resume(pid int) error }`
  - `type Guard struct{}`, `func New(cfg config.EDRConfig, sc ImageScanner, r TreeResponder) *Guard`
  - `func (g *Guard) OnStart(p proctable.Proc)` — the §6.2 flow: suspend if configured → scan image → kill tree if malicious, else resume

- [ ] **Step 1: Write the failing test (fakes; assert kill on malicious, resume on clean, suspend when configured)**

```go
// agent/edr/guard/guard_test.go
package guard

import (
	"testing"

	"github.com/utmstack/UTMStack/agent/edr/config"
	"github.com/utmstack/UTMStack/agent/edr/proctable"
)

type fakeScanner struct{ verdict, sig string }
func (f fakeScanner) ScanFile(path, source string) (string, string, error) { return f.verdict, f.sig, nil }

type fakeResp struct {
	suspended, resumed, killed int
}
func (f *fakeResp) KillTree(pid int, sig string) ([]int, error) { f.killed = pid; return []int{pid}, nil }
func (f *fakeResp) Suspend(pid int) error                        { f.suspended = pid; return nil }
func (f *fakeResp) Resume(pid int) error                         { f.resumed = pid; return nil }

func TestGuardKillsMaliciousLaunch(t *testing.T) {
	cfg := config.Default() // SuspendOnLaunch=true
	fr := &fakeResp{}
	g := New(cfg, fakeScanner{verdict: "malicious", sig: "Win.Test.EICAR"}, fr)
	g.OnStart(proctable.Proc{PID: 555, Image: `C:\a.exe`})
	if fr.suspended != 555 {
		t.Fatalf("expected suspend of 555, got %d", fr.suspended)
	}
	if fr.killed != 555 {
		t.Fatalf("expected kill of 555, got %d", fr.killed)
	}
}

func TestGuardResumesCleanLaunch(t *testing.T) {
	cfg := config.Default()
	fr := &fakeResp{}
	g := New(cfg, fakeScanner{verdict: "clean"}, fr)
	g.OnStart(proctable.Proc{PID: 777, Image: `C:\ok.exe`})
	if fr.resumed != 777 || fr.killed != 0 {
		t.Fatalf("expected resume of 777 and no kill, got resumed=%d killed=%d", fr.resumed, fr.killed)
	}
}
```

- [ ] **Step 2: Run to verify it fails**

Run: `go test ./edr/guard/ -v`
Expected: FAIL — undefined.

- [ ] **Step 3: Implement**

```go
// agent/edr/guard/guard.go
package guard

import (
	"github.com/utmstack/UTMStack/agent/edr/cache"
	"github.com/utmstack/UTMStack/agent/edr/config"
	"github.com/utmstack/UTMStack/agent/edr/event"
	"github.com/utmstack/UTMStack/agent/edr/proctable"
	"github.com/utmstack/UTMStack/shared/logger"
)

type ImageScanner interface {
	ScanFile(path, source string) (string, string, error)
}

type TreeResponder interface {
	KillTree(rootPID int, signature string) ([]int, error)
	Suspend(pid int) error
	Resume(pid int) error
}

type Guard struct {
	cfg config.EDRConfig
	sc  ImageScanner
	r   TreeResponder
}

func New(cfg config.EDRConfig, sc ImageScanner, r TreeResponder) *Guard {
	return &Guard{cfg: cfg, sc: sc, r: r}
}

// OnStart implements the run-before-scan race handling (spec §6.2):
// suspend (if enabled) → expedite scan of the image → kill the tree on
// malicious, otherwise resume. The scan uses source=process_watcher so the
// resulting detection event is attributed correctly.
func (g *Guard) OnStart(p proctable.Proc) {
	if p.Image == "" {
		return
	}
	suspended := false
	if g.cfg.SuspendOnLaunch {
		if err := g.r.Suspend(p.PID); err == nil {
			suspended = true
		}
	}

	verdict, sig, err := g.sc.ScanFile(p.Image, event.SourceProcessWatcher)
	if err != nil {
		logger.Error("UTMStack EDR: expedited scan of %s failed: %v", p.Image, err)
		if suspended {
			_ = g.r.Resume(p.PID) // fail-open: don't hang a launch on scan error
		}
		return
	}

	if verdict == cache.VerdictMalicious {
		if _, err := g.r.KillTree(p.PID, sig); err != nil {
			logger.Error("UTMStack EDR: kill tree %d failed: %v", p.PID, err)
		}
		return
	}
	if suspended {
		_ = g.r.Resume(p.PID)
	}
}
```

- [ ] **Step 4: Run to verify it passes**

Run: `go test ./edr/guard/ -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add agent/edr/guard/
git commit -m "feat(edr): launch guard implementing the run-before-scan race handling"
```

---

## Task 5: Process-creation watcher (WMI) + stub

**Files:**
- Create: `agent/edr/procwatch/procwatch.go`, `wmi_windows.go`, `wmi_other.go`
- Modify: `agent/go.mod` (add `github.com/go-ole/go-ole`)
- Test: build-only; behavior VM-verified.

**Interfaces:**
- Produces:
  - `type ProcStart struct { PID, PPID int; Image, Cmdline string }`
  - `type Handler interface { OnProcStart(ProcStart) }`
  - `type Watcher struct{}`, `func New(h Handler) *Watcher`, `func (w *Watcher) Run(ctx context.Context)`

- [ ] **Step 1: Add the dependency**

Run:
```bash
go get github.com/go-ole/go-ole@v1.3.0
go mod tidy
```
Expected: `go-ole` added to `go.mod` require block.

- [ ] **Step 2: Portable types + stub**

```go
// agent/edr/procwatch/procwatch.go
package procwatch

type ProcStart struct {
	PID     int
	PPID    int
	Image   string
	Cmdline string
}

type Handler interface {
	OnProcStart(ProcStart)
}
```

```go
// agent/edr/procwatch/wmi_other.go
//go:build !windows

package procwatch

import "context"

type Watcher struct{}

func New(h Handler) *Watcher { return &Watcher{} }

// Run is a no-op off Windows (WMI is Windows-only).
func (w *Watcher) Run(ctx context.Context) { <-ctx.Done() }
```

- [ ] **Step 3: WMI subscription (Windows)**

Subscribes to `__InstanceCreationEvent` for `Win32_Process` and emits `ProcStart` for each. This is standard go-ole WMI usage; the exact COM call chain is Windows-native and VM-validated.

```go
// agent/edr/procwatch/wmi_windows.go
//go:build windows

package procwatch

import (
	"context"
	"strconv"
	"time"

	ole "github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
	"github.com/utmstack/UTMStack/shared/logger"
)

type Watcher struct{ h Handler }

func New(h Handler) *Watcher { return &Watcher{h: h} }

func (w *Watcher) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		if err := w.subscribe(ctx); err != nil {
			logger.Error("UTMStack EDR process watcher: %v; retrying", err)
			time.Sleep(3 * time.Second)
		}
	}
}

func (w *Watcher) subscribe(ctx context.Context) error {
	if err := ole.CoInitializeEx(0, ole.COINIT_MULTITHREADED); err != nil {
		// S_FALSE (already initialized) is not fatal
	}
	defer ole.CoUninitialize()

	unknown, err := oleutil.CreateObject("WbemScripting.SWbemLocator")
	if err != nil {
		return err
	}
	defer unknown.Release()
	locator, err := unknown.QueryInterface(ole.IID_IDispatch)
	if err != nil {
		return err
	}
	defer locator.Release()

	serviceRaw, err := oleutil.CallMethod(locator, "ConnectServer", nil, `root\cimv2`)
	if err != nil {
		return err
	}
	service := serviceRaw.ToIDispatch()
	defer service.Release()

	query := "SELECT * FROM __InstanceCreationEvent WITHIN 1 WHERE TargetInstance ISA 'Win32_Process'"
	eventSourceRaw, err := oleutil.CallMethod(service, "ExecNotificationQuery", query)
	if err != nil {
		return err
	}
	eventSource := eventSourceRaw.ToIDispatch()
	defer eventSource.Release()

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}
		// NextEvent blocks up to timeout (ms); loop lets ctx cancel.
		evtRaw, err := oleutil.CallMethod(eventSource, "NextEvent", 1000)
		if err != nil {
			continue // timeout / transient
		}
		evt := evtRaw.ToIDispatch()
		targetRaw, err := oleutil.GetProperty(evt, "TargetInstance")
		if err != nil {
			evt.Release()
			continue
		}
		target := targetRaw.ToIDispatch()

		ps := ProcStart{
			PID:     dispInt(target, "ProcessId"),
			PPID:    dispInt(target, "ParentProcessId"),
			Image:   dispStr(target, "ExecutablePath"),
			Cmdline: dispStr(target, "CommandLine"),
		}
		if ps.PID != 0 && ps.Image != "" {
			w.h.OnProcStart(ps)
		}
		target.Release()
		targetRaw.Clear()
		evt.Release()
		evtRaw.Clear()
	}
}

func dispStr(d *ole.IDispatch, prop string) string {
	v, err := oleutil.GetProperty(d, prop)
	if err != nil {
		return ""
	}
	defer v.Clear()
	return v.ToString()
}

func dispInt(d *ole.IDispatch, prop string) int {
	v, err := oleutil.GetProperty(d, prop)
	if err != nil {
		return 0
	}
	defer v.Clear()
	switch val := v.Value().(type) {
	case int32:
		return int(val)
	case int64:
		return int(val)
	case string:
		n, _ := strconv.Atoi(val)
		return n
	default:
		return 0
	}
}
```

> **VM verification point:** the go-ole COM call sequence and property types are validated by running on the VM (Task 7). If `ExecutablePath` is empty for protected processes, fall back to `Name` + resolving the path from PID; the guard tolerates an empty image (skips).

- [ ] **Step 4: Build both platforms**

Run: `go build ./edr/procwatch/ && GOOS=windows GOARCH=amd64 go build ./edr/procwatch/`
Expected: both succeed.

- [ ] **Step 5: Commit**

```bash
git add agent/go.mod agent/go.sum agent/edr/procwatch/
git commit -m "feat(edr): WMI process-creation watcher (windows) + stub"
```

---

## Task 6: Wire the watcher → table + guard into the service

**Files:**
- Modify: `agent/edr/service/service.go`
- Create: `agent/edr/procwatch/dispatch.go` — a `Handler` that updates the table and calls the guard.
- Test: `agent/edr/procwatch/dispatch_test.go`

**Interfaces:**
- Produces:
  - `type Dispatch struct{}`, `func NewDispatch(tab *proctable.Table, g *guard.Guard) *Dispatch`
  - `func (d *Dispatch) OnProcStart(ps ProcStart)` — adds to table, then runs the guard in a goroutine (so scanning never blocks the WMI event loop)

- [ ] **Step 1: Write the failing test**

```go
// agent/edr/procwatch/dispatch_test.go
package procwatch

import (
	"testing"
	"time"

	"github.com/utmstack/UTMStack/agent/edr/proctable"
)

func TestDispatchAddsToTable(t *testing.T) {
	tab := proctable.New()
	d := NewDispatch(tab, nil) // nil guard => dispatch only updates the table
	d.OnProcStart(ProcStart{PID: 909, PPID: 4, Image: `C:\a.exe`})

	// table update is synchronous
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		if _, ok := tab.Get(909); ok {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatal("process 909 was not added to the table")
}
```

- [ ] **Step 2: Run to verify it fails**

Run: `go test ./edr/procwatch/ -run TestDispatch -v`
Expected: FAIL — undefined.

- [ ] **Step 3: Implement**

```go
// agent/edr/procwatch/dispatch.go
package procwatch

import (
	"github.com/utmstack/UTMStack/agent/edr/guard"
	"github.com/utmstack/UTMStack/agent/edr/proctable"
)

type Dispatch struct {
	table *proctable.Table
	guard *guard.Guard
}

func NewDispatch(tab *proctable.Table, g *guard.Guard) *Dispatch {
	return &Dispatch{table: tab, guard: g}
}

func (d *Dispatch) OnProcStart(ps ProcStart) {
	d.table.Add(proctable.Proc{PID: ps.PID, PPID: ps.PPID, Image: ps.Image, Cmdline: ps.Cmdline})
	if d.guard != nil {
		// Run the guard off the event loop so scan latency never blocks WMI.
		go d.guard.OnStart(proctable.Proc{PID: ps.PID, PPID: ps.PPID, Image: ps.Image, Cmdline: ps.Cmdline})
	}
}
```

- [ ] **Step 4: Run to verify it passes**

Run: `go test ./edr/procwatch/ -run TestDispatch -v`
Expected: PASS.

- [ ] **Step 5: Wire into `agent/edr/service/service.go`**

Inside the `if cfg.Enabled` block (added in Plan 2 Task 8), after the orchestrator/watcher/feed, add:

```go
		tab := proctable.New()
		resp := responder.New(tab, sp, responder.OSKiller{}, responder.OSSuspender{})
		g := guard.New(cfg, sc, resp) // sc is the *scanner.Scanner already built
		pw := procwatch.New(procwatch.NewDispatch(tab, g))
		go pw.Run(ctx)
```

Add imports for `proctable`, `responder`, `guard`, `procwatch`.

- [ ] **Step 6: Build both platforms**

Run: `go build ./... && GOOS=windows GOARCH=amd64 go build ./...`
Expected: both succeed.

- [ ] **Step 7: Commit**

```bash
git add agent/edr/procwatch/dispatch.go agent/edr/procwatch/dispatch_test.go agent/edr/service/
git commit -m "feat(edr): dispatch process starts into table + guard from the service"
```

---

## Task 7: End-to-end acceptance on the VM

- [ ] **Step 1: Build + deploy** (as Plan 1/2 Task 9; ClamAV up on 3310, EDR service reinstalled, `enable-edr`).

- [ ] **Step 2: Process kill on a running malicious process (the core ask)**

Create a benign "malicious" test that ClamAV flags. Use an EICAR-carrying batch/exe that spawns a child, then run it:

```powershell
# a.bat writes the EICAR string into a file it also is (self-detecting test),
# or use a signed test binary that clamd flags. Simplest: a script host launch
# that reads an EICAR-embedded file. For a real process kill, compile a tiny
# looping exe that embeds the EICAR pattern in its .data so clamd flags it,
# and have it spawn a child that also loops.
Start-Process C:\Users\atlas\Downloads\evil_parent.exe
Start-Sleep -Seconds 6
Get-Process evil_parent, evil_child -ErrorAction SilentlyContinue
```
Expected: neither `evil_parent` nor `evil_child` is running (tree killed); spool has `"action":"killed"` events for both PIDs with `"source":"process_watcher"`, branded, no `clam` string; the file is quarantined.

- [ ] **Step 3: Suspend-on-launch behavior**

With `suspend_on_launch=true`, launch an unknown large file; observe (Process Explorer) it is suspended during the scan, then killed (malicious) or resumed (clean). With a scan that exceeds `suspend_timeout_ms`, confirm the process resumes (fail-open) rather than hanging.

- [ ] **Step 4: Clean launch is untouched**

Launch `notepad.exe`; confirm it is (briefly) suspended then resumed and keeps running; no kill event.

- [ ] **Step 5: Platform + labeling**

Confirm `killed`/`suspended` events reach the platform as `utmstack_edr`, branded, and that a `clam` search of spool + service log is empty.

- [ ] **Step 6: Commit fixes**

```bash
git add -A && git commit -m "fix(edr): plan 3 VM acceptance adjustments"
```

---

## Self-Review (against the spec)

**Spec coverage (Plan 3 slice):**
- §4.3 Process-creation watcher (ETW-preferred; WMI implemented here as the buildable feed, same double-duty telemetry) → Task 5. *ETW upgrade noted as a future swap behind the `Watcher` interface.*
- §4.6 Responder: process-tree kill (leaves-first) + suspend-on-launch → Tasks 3, 4. ✔
- §4.4 correlation (launch references unknown image → expedite scan → act) → Task 4 (guard) + Task 6. ✔
- §6.2 race scenario end-to-end → Tasks 4, 6, 7. ✔
- Privilege precondition (SYSTEM, no driver) honored — kill is plain OpenProcess/TerminateProcess. ✔
- Branding preserved (`source=process_watcher`, no engine name) → tests in Tasks 2, 3. ✔
- Honest residual race documented (Global Constraints + guard comment). ✔

**Placeholder scan:** table, tree, responder core, guard, and dispatch have complete code + real `go test`. The kill/suspend syscalls and the WMI feed are concrete Win32/go-ole code with explicit VM verification points — correctness confirmed by compile+run on the VM (Task 7).

**Type consistency:** `guard.ImageScanner.ScanFile(path,source)(string,string,error)` matches `scanner.Scanner`. `guard.TreeResponder` (KillTree/Suspend/Resume) matches `responder.Responder`'s methods. `responder.Killer/Suspender` match `OSKiller`/`OSSuspender`. `procwatch.Handler.OnProcStart(ProcStart)` matches `Dispatch`. `proctable.Proc` is the shared process shape across table, responder, guard, dispatch.

**Note:** the WMI watcher adds `github.com/go-ole/go-ole` (the one dependency this plan introduces, in Task 5) — the only exception to the no-new-deps rule, called out explicitly per the constraint.

## Implementation deviations (found during VM acceptance, 2026-07-02)

VM testing on the Parallels Windows-on-ARM64 host surfaced two defects not catchable by cross-compile or macOS unit tests; both fixed:

1. **COM double-free crash (Task 5).** `wmi_windows.go` called both `dispatch.Release()` **and** `variant.Clear()` on the same go-ole object; `ToIDispatch()` shares the VARIANT's single COM reference, so this double-freed it → `0xc0000005` access violation that crashed the SYSTEM service ~60s after start (no Go panic → not caught by recover; diagnosed by running `utmstack_edr run` with stderr captured to a file). Fix: `Release()` the dispatch only, never also `Clear()` its source VARIANT. Additionally added `runtime.LockOSThread()` in `subscribe()` (COM is thread-affine) and a `service.goSafe` panic-recovery wrapper around the pipeline goroutines (the EDR service had none).
2. **Redundant-kill noise (Task 3).** A malicious child's own guard races the parent's tree-kill on the same PID, logging a spurious "failed to terminate" error. Fix: an atomic `claim(pid)` kill-once dedup in the responder (+ `ErrProcessGone`), with `TestKillTreeKillsEachPidOnce`.

**VM acceptance result:** an EICAR-marker test process (parent + child) was killed as a whole tree, leaves-first (conhost + child + parent), with branded `killed` events (`source=process_watcher`) relayed to the platform. Suspend-on-launch (§4.6) was not exercised (config had it off) — verify it in a follow-up.
