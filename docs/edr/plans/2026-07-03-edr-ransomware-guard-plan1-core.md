# EDR Ransomware Guard — v1 Core Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add driverless behavioral **ransomware detection-and-response** to the EDR: catch unknown/novel crypto-ransomware by fusing high-fidelity signals (canary-file tampering + T1490 recovery-destruction commands) into a per-process risk score that drives the existing suspend/kill/quarantine arm.

**Architecture:** A new `edr/ransomware/` package = a **detection brain** bolted onto the existing **response body**. Independent *sensors* emit PID-tagged weighted `Evidence` into a pure-Go decaying **scorer**; when a process crosses a threshold, a **guard** runs an escalation ladder (suspend → kill-tree → quarantine → branded event → incident record) via the already-built `responder`/`quarantine`. Per-process file attribution comes from a swappable **file-activity feed** interface (v1 impl = ETW `Microsoft-Windows-Kernel-File`, VM-verified); T1490 attribution comes from the existing WMI `procwatch` feed, no ETW needed.

**Tech Stack:** Go 1.25.5, `gorm`+`glebarez/sqlite` (pure-Go), `golang.org/x/sys/windows`, ETW consumer lib (see Global Constraints). Native Windows pieces sit behind `//go:build windows` with `!windows` stubs; all pure logic is TDD'd with `go test` on macOS.

## Global Constraints

Every task's requirements implicitly include these (copied from the design spec + project CLAUDE.md):

- **No kernel driver / minifilter, ever.** Attribution via ETW (user-mode), response via SYSTEM privilege. This is **detect-fast-then-kill**, not pre-block — state it honestly, never claim "prevents ransomware."
- **Go-first.** Only Go. Native ETW/file-attr code goes behind `//go:build windows` + a `!windows` stub with identical exported signatures, so the module builds/tests on macOS. No new non-Go artifact.
- **Branding (critical).** Every emitted event/log says **"UTMStack EDR"**. The strings `clam`/`clamav`/`clamd` must **never** appear in an event field or a surfaced log line. `event.Event.ToJSON()` already forces `Product`/`Engine = "UTMStack EDR"`. The labeling audit `grep -rniE '(logger\.(Info|Error|Fatal|Debug)|fmt\.(Print|Sprint))[^\n]*clam' edr/` must return nothing.
- **Own `edr.db`** (never the agent's `logs.db`). New tables go through `cache.Open`'s `AutoMigrate`.
- **Quarantine, never delete.** Reuse `quarantine.Store.Quarantine`.
- **Both arches.** All code must cross-build and vet for Windows **amd64 and arm64**.
- **ETW dependency (open decision §12.1/§12.5 of the spec):** v1 uses one ETW consumer library — plan pins `github.com/0xrawsec/golang-etw` as the concrete choice (precedent: Plan 3's `go-ole`). Confirm its exact API at the pinned version when implementing Task 11; the guard depends only on the `ransomware.FileActivityFeed` interface, so the library is swappable without touching tested code.
- **Git is user-driven.** Commit steps below mark the TDD rhythm; per project policy the **user performs all git operations** — the executor should surface each commit point and let the user commit (never revert via git).
- **Config default:** the feature ships **disabled** (`Ransomware.Enabled=false`), default `ResponseMode="suspend"`. Opt-in rollout.

## Module command reference (run from `utmstack-v12/agent/`)

```bash
go test ./edr/ransomware/... ./edr/...        # pure-Go unit tests (macOS host)
GOOS=windows GOARCH=amd64 go build ./...       # windows-tagged build (amd64)
GOOS=windows GOARCH=arm64 go build ./...       # windows-tagged build (arm64)
GOOS=windows GOARCH=amd64 go vet ./edr/...
grep -rniE '(logger\.(Info|Error|Fatal|Debug)|fmt\.(Print|Sprint))[^\n]*clam' edr/   # labeling audit → must be empty
```

## File Structure

**New package `edr/ransomware/`:**
| File | Responsibility | Tested on macOS? |
|---|---|---|
| `evidence.go` | `Evidence`, `SignalKind` consts, default weights | ✅ pure |
| `scorer.go` | Per-PID decaying score + thresholds + kind-diversity + generation keying → `Escalation` | ✅ pure |
| `rules.go` | T1490 command-rule table + matcher + allowlist | ✅ pure |
| `canary.go` | Canary planner (names/placement), `Manager` (plant/persist/membership) | ✅ pure logic |
| `canary_windows.go` / `canary_other.go` | `hideFile` (Hidden+System attrs) native / stub | native / stub |
| `feed.go` | `FileEvent`, `FileOp` consts, `FileActivityFeed` interface, portable `Consider` filter | ✅ pure |
| `feed_windows.go` / `feed_other.go` | ETW `Microsoft-Windows-Kernel-File` consumer / stub | native (VM) / stub |
| `guard.go` | `Guard`: `Run`, `OnProcStart`, `OnFileEvent`, escalation ladder | ✅ with fakes |

**Modified existing files:**
| File | Change |
|---|---|
| `edr/event/event.go` | add `ActionRansomwareSuspected`/`ActionRansomwareContained`, `NewRansomwareEvent` |
| `edr/config/config.go` | add `RansomwareConfig` + field + defaults + `Load` overlay |
| `edr/cache/cache.go` | add `CanaryRecord` + `RansomwareIncident` + `AutoMigrate` + accessors (new file `cache/ransomware.go`) |
| `edr/procwatch/dispatch.go` | add `onProc` callback + populate `StartTS` |
| `edr/service/service.go` | wire the guard + feed into `startPipeline`; add `statusDoc` fields |

---

### Task 1: Event actions + ransomware constructor

**Files:**
- Modify: `edr/event/event.go` (add consts near line 25; add constructor after `NewProcessAction`, ~line 88)
- Test: `edr/event/ransomware_event_test.go` (create)

**Interfaces:**
- Produces: `event.ActionRansomwareSuspected`, `event.ActionRansomwareContained` (string consts); `event.NewRansomwareEvent(action string, p ProcInfo, signal, severity string) Event`.

- [ ] **Step 1: Write the failing test**

```go
package event

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestNewRansomwareEvent_BrandedAndTagged(t *testing.T) {
	p := ProcInfo{PID: 1234, PPID: 10, Image: `C:\Users\x\enc.exe`, Cmdline: "enc.exe --all"}
	ev := NewRansomwareEvent(ActionRansomwareContained, p, "canary:00__accounts.xlsx", "critical")

	if ev.Source != SourceBehavioral {
		t.Fatalf("source = %q, want %q", ev.Source, SourceBehavioral)
	}
	if ev.Action != ActionRansomwareContained {
		t.Fatalf("action = %q", ev.Action)
	}
	if ev.Signature != "canary:00__accounts.xlsx" || ev.Severity != "critical" {
		t.Fatalf("signal/severity not set: %+v", ev)
	}
	if ev.Process == nil || ev.Process.PID != 1234 {
		t.Fatalf("process not carried: %+v", ev.Process)
	}
	js, err := ev.ToJSON()
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	_ = json.Unmarshal([]byte(js), &m)
	if m["product"] != Product || m["engine"] != Product {
		t.Fatalf("branding missing: %v", m)
	}
	if strings.Contains(strings.ToLower(js), "clam") {
		t.Fatalf("engine name leaked into event: %s", js)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./edr/event/ -run TestNewRansomwareEvent -v`
Expected: FAIL — `undefined: ActionRansomwareContained` / `NewRansomwareEvent`.

- [ ] **Step 3: Add the consts and constructor**

In `edr/event/event.go`, add to the action const block (after `ActionAllowed`):

```go
	ActionRansomwareSuspected = "ransomware_suspected"
	ActionRansomwareContained = "ransomware_contained"
```

Add after `NewProcessAction`:

```go
// NewRansomwareEvent builds a branded behavioral event for a ransomware
// escalation. signal is the top contributing detector (e.g.
// "canary:00__accounts.xlsx" or "t1490:vssadmin_delete_shadows"); severity is
// a coarse label ("high"|"critical"). Source is always behavioral so the
// platform parser routes it with the other behavioral telemetry.
func NewRansomwareEvent(action string, p ProcInfo, signal, severity string) Event {
	pc := p
	return Event{
		Source:     SourceBehavioral,
		Action:     action,
		ObjectPath: p.Image,
		Signature:  signal,
		Severity:   severity,
		Process:    &pc,
	}
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./edr/event/ -run TestNewRansomwareEvent -v`
Expected: PASS.

- [ ] **Step 5: Commit** (surface to user)

```bash
git add edr/event/event.go edr/event/ransomware_event_test.go
git commit -m "feat(edr): branded ransomware event actions + constructor"
```

---

### Task 2: Config block — `RansomwareConfig`

**Files:**
- Modify: `edr/config/config.go` (add struct + field ~line 55; defaults ~line 80; overlay ~line 114)
- Test: `edr/config/ransomware_config_test.go` (create)

**Interfaces:**
- Produces: `config.RansomwareConfig{Enabled bool; ResponseMode string; CanaryDirs []string; CanaryPerDir int; SuspendThreshold, KillThreshold, DecayHalfLifeMs int; CommandAllowlist []string; UseETW bool}`; `config.EDRConfig.Ransomware RansomwareConfig`.

- [ ] **Step 1: Write the failing test**

```go
package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestRansomwareDefaults(t *testing.T) {
	d := Default()
	if d.Ransomware.Enabled {
		t.Fatal("ransomware must default OFF")
	}
	if d.Ransomware.ResponseMode != "suspend" {
		t.Fatalf("default response_mode = %q, want suspend", d.Ransomware.ResponseMode)
	}
	if d.Ransomware.KillThreshold != 100 || d.Ransomware.SuspendThreshold != 50 {
		t.Fatalf("thresholds = %d/%d", d.Ransomware.SuspendThreshold, d.Ransomware.KillThreshold)
	}
	if !d.Ransomware.UseETW || d.Ransomware.CanaryPerDir != 1 || d.Ransomware.DecayHalfLifeMs != 10000 {
		t.Fatalf("bad defaults: %+v", d.Ransomware)
	}
}

func TestRansomwareOverlayRoundTrip(t *testing.T) {
	dir := t.TempDir()
	old := ConfigFile
	ConfigFile = filepath.Join(dir, "edr.json")
	defer func() { ConfigFile = old }()

	on := Default()
	on.Ransomware.Enabled = true
	on.Ransomware.ResponseMode = "kill"
	on.Ransomware.CanaryDirs = []string{`D:\shared`}
	b, _ := json.Marshal(on)
	if err := os.WriteFile(ConfigFile, b, 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if !got.Ransomware.Enabled || got.Ransomware.ResponseMode != "kill" ||
		len(got.Ransomware.CanaryDirs) != 1 || got.Ransomware.CanaryDirs[0] != `D:\shared` {
		t.Fatalf("overlay lost ransomware block: %+v", got.Ransomware)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./edr/config/ -run TestRansomware -v`
Expected: FAIL — `unknown field Ransomware`.

- [ ] **Step 3: Add the struct, field, defaults, and overlay**

In `edr/config/config.go`, add above `EDRConfig`:

```go
// RansomwareConfig configures the behavioral ransomware guard (v1 core).
type RansomwareConfig struct {
	Enabled          bool     `json:"enabled"`
	ResponseMode     string   `json:"response_mode"`      // "alert" | "suspend" | "kill"
	CanaryDirs       []string `json:"canary_dirs"`        // extra subtrees to seed (beyond volume roots + profile)
	CanaryPerDir     int      `json:"canary_per_dir"`     // decoys planted per directory
	SuspendThreshold int      `json:"suspend_threshold"`  // score to suspend a suspect
	KillThreshold    int      `json:"kill_threshold"`     // score to kill+quarantine
	DecayHalfLifeMs  int      `json:"decay_half_life_ms"` // per-PID score half-life
	CommandAllowlist []string `json:"command_allowlist"`  // benign admin command lines exempt from T1490 rules
	UseETW           bool     `json:"use_etw"`            // true = ETW file attribution (v1 default)
}
```

Add the field to `EDRConfig` (after `SigMirror`):

```go
	// Ransomware is the behavioral ransomware guard block (nested; copied wholesale
	// by Load when present on disk).
	Ransomware RansomwareConfig `json:"ransomware"`
```

In `Default()`, set the nested defaults inside the returned literal (add after `SigUpdateHours: 4,`):

```go
		Ransomware: RansomwareConfig{
			Enabled:          false, // opt-in rollout
			ResponseMode:     "suspend",
			CanaryPerDir:     1,
			SuspendThreshold: 50,
			KillThreshold:    100,
			DecayHalfLifeMs:  10000,
			UseETW:           true,
		},
```

In `Load()`, add one overlay before `return c, nil` (copy the nested struct wholesale when the on-disk block is present, keep defaults when absent). `RansomwareConfig` has slice fields (`CanaryDirs`, `CommandAllowlist`), so it is **not comparable** — detect "present" via non-zero scalars rather than `!=` against a zero struct:

```go
	if onDisk.Ransomware.ResponseMode != "" || onDisk.Ransomware.Enabled || onDisk.Ransomware.KillThreshold != 0 {
		c.Ransomware = onDisk.Ransomware
	}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./edr/config/ -run TestRansomware -v`
Expected: PASS (both tests).

- [ ] **Step 5: Commit** (surface to user)

```bash
git add edr/config/config.go edr/config/ransomware_config_test.go
git commit -m "feat(edr): ransomware guard config block + overlay"
```

---

### Task 3: Cache records — canaries + incidents

**Files:**
- Create: `edr/cache/ransomware.go`
- Modify: `edr/cache/cache.go:52` (extend `AutoMigrate` list)
- Test: `edr/cache/ransomware_test.go` (create)

**Interfaces:**
- Produces: `cache.CanaryRecord{Path string; SHA256 string; Volume string; PlacedAt time.Time}`; `cache.RansomwareIncident{ID string; PID int; Image, Cmdline, Signals string; Score int; Action string; DetectedAt time.Time; QuarantineID string}`; methods `(*Cache) StoreCanary(CanaryRecord) error`, `(*Cache) ListCanaries() ([]CanaryRecord, error)`, `(*Cache) StoreIncident(RansomwareIncident) error`.

- [ ] **Step 1: Write the failing test**

```go
package cache

import (
	"path/filepath"
	"testing"
)

func TestCanaryAndIncidentPersist(t *testing.T) {
	c, err := Open(filepath.Join(t.TempDir(), "edr.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	if err := c.StoreCanary(CanaryRecord{Path: `C:\Users\x\00__a.xlsx`, SHA256: "abc", Volume: "C:"}); err != nil {
		t.Fatal(err)
	}
	cs, err := c.ListCanaries()
	if err != nil || len(cs) != 1 || cs[0].Path != `C:\Users\x\00__a.xlsx` {
		t.Fatalf("canary not persisted: %v %v", cs, err)
	}
	if cs[0].PlacedAt.IsZero() {
		t.Fatal("PlacedAt not stamped")
	}

	if err := c.StoreIncident(RansomwareIncident{
		ID: "i1", PID: 42, Image: `C:\enc.exe`, Signals: "canary,t1490",
		Score: 200, Action: "contained", QuarantineID: "q1",
	}); err != nil {
		t.Fatal(err)
	}
	// second store with same ID must upsert, not error
	if err := c.StoreIncident(RansomwareIncident{ID: "i1", PID: 42, Action: "contained"}); err != nil {
		t.Fatalf("incident upsert failed: %v", err)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./edr/cache/ -run TestCanaryAndIncident -v`
Expected: FAIL — undefined `CanaryRecord`/`StoreCanary`.

- [ ] **Step 3: Create `edr/cache/ransomware.go`**

```go
package cache

import "time"

// CanaryRecord is a planted decoy file. A write/rename/delete on Path is a
// high-confidence ransomware signal.
type CanaryRecord struct {
	Path     string `gorm:"primaryKey"`
	SHA256   string
	Volume   string
	PlacedAt time.Time
}

// RansomwareIncident is the audit trail of a contained (or alerted) ransomware
// escalation.
type RansomwareIncident struct {
	ID           string `gorm:"primaryKey"`
	PID          int
	Image        string
	Cmdline      string
	Signals      string // comma-separated contributing signal kinds
	Score        int
	Action       string // "alerted" | "suspended" | "contained"
	DetectedAt   time.Time
	QuarantineID string
}

func (c *Cache) StoreCanary(r CanaryRecord) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if r.PlacedAt.IsZero() {
		r.PlacedAt = time.Now().UTC()
	}
	return c.db.Save(&r).Error
}

func (c *Cache) ListCanaries() ([]CanaryRecord, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	var rs []CanaryRecord
	err := c.db.Find(&rs).Error
	return rs, err
}

func (c *Cache) StoreIncident(r RansomwareIncident) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if r.DetectedAt.IsZero() {
		r.DetectedAt = time.Now().UTC()
	}
	return c.db.Save(&r).Error
}
```

- [ ] **Step 4: Extend `AutoMigrate`**

In `edr/cache/cache.go` line 52, change:

```go
	if err := db.AutoMigrate(&VerdictRecord{}, &QuarantineRecord{}, &USNCursor{}); err != nil {
```
to:
```go
	if err := db.AutoMigrate(&VerdictRecord{}, &QuarantineRecord{}, &USNCursor{}, &CanaryRecord{}, &RansomwareIncident{}); err != nil {
```

- [ ] **Step 5: Run test to verify it passes**

Run: `go test ./edr/cache/ -run TestCanaryAndIncident -v`
Expected: PASS.

- [ ] **Step 6: Commit** (surface to user)

```bash
git add edr/cache/ransomware.go edr/cache/cache.go edr/cache/ransomware_test.go
git commit -m "feat(edr): canary + ransomware-incident cache records"
```

---

### Task 4: Scorer + evidence model (the heart)

**Files:**
- Create: `edr/ransomware/evidence.go`, `edr/ransomware/scorer.go`
- Test: `edr/ransomware/scorer_test.go`

**Interfaces:**
- Produces:
  - `ransomware.SignalKind` (string) with `KindCanary`, `KindT1490` consts and `DefaultWeights map[SignalKind]float64`.
  - `ransomware.Evidence{PID int; Gen int64; Kind SignalKind; Weight float64; Detail string; TS time.Time}`.
  - `ransomware.Escalation int` with `EscNone`, `EscSuspend`, `EscKill`.
  - `ransomware.Scorer` via `NewScorer(suspendT, killT, halfLife float64, now func() time.Time) *Scorer`; `(*Scorer) Add(Evidence) Decision`; `(*Scorer) Forget(pid int)`.
  - `ransomware.Decision{Escalation Escalation; Score float64; TopSignal string; Kinds []SignalKind}`.

- [ ] **Step 1: Write the failing tests**

```go
package ransomware

import (
	"testing"
	"time"
)

func fixedClock(base time.Time) (*time.Time, func() time.Time) {
	t := base
	return &t, func() time.Time { return t }
}

func TestScorer_MaxWeightSignalKillsImmediately(t *testing.T) {
	now, clk := fixedClock(time.Unix(1000, 0))
	_ = now
	s := NewScorer(50, 100, 10, clk)
	d := s.Add(Evidence{PID: 7, Gen: 1, Kind: KindCanary, Weight: 100, Detail: "00__a.xlsx"})
	if d.Escalation != EscKill {
		t.Fatalf("canary touch must cross kill: got %v score=%v", d.Escalation, d.Score)
	}
	if d.TopSignal != "canary:00__a.xlsx" {
		t.Fatalf("top signal = %q", d.TopSignal)
	}
}

func TestScorer_EscalationDoesNotRefire(t *testing.T) {
	_, clk := fixedClock(time.Unix(1000, 0))
	s := NewScorer(50, 100, 10, clk)
	if s.Add(Evidence{PID: 7, Gen: 1, Kind: KindT1490, Weight: 100}).Escalation != EscKill {
		t.Fatal("first should kill")
	}
	// A second signal on the same PID+gen must not re-emit EscKill.
	if e := s.Add(Evidence{PID: 7, Gen: 1, Kind: KindCanary, Weight: 100}).Escalation; e != EscNone {
		t.Fatalf("re-fire = %v, want EscNone", e)
	}
}

func TestScorer_DecayLowersScore(t *testing.T) {
	base := time.Unix(1000, 0)
	cur := base
	s := NewScorer(50, 100, 10, func() time.Time { return cur })
	// One 60-weight signal → suspend but not kill.
	if s.Add(Evidence{PID: 5, Gen: 1, Kind: KindEntropyPlaceholder, Weight: 60}).Escalation != EscSuspend {
		t.Fatal("60 should suspend")
	}
	// Advance two half-lives (20s): 60 → ~15, well under suspend threshold.
	cur = base.Add(20 * time.Second)
	d := s.Add(Evidence{PID: 5, Gen: 1, Kind: KindEntropyPlaceholder, Weight: 1})
	if d.Score > 30 {
		t.Fatalf("score did not decay: %v", d.Score)
	}
}

func TestScorer_RecycledPIDResets(t *testing.T) {
	_, clk := fixedClock(time.Unix(1000, 0))
	s := NewScorer(50, 100, 10, clk)
	s.Add(Evidence{PID: 9, Gen: 1, Kind: KindCanary, Weight: 100})
	// Same PID, new generation → fresh score, a single small signal must not kill.
	d := s.Add(Evidence{PID: 9, Gen: 2, Kind: KindEntropyPlaceholder, Weight: 10})
	if d.Escalation != EscNone {
		t.Fatalf("recycled PID inherited old score: %v", d.Escalation)
	}
}

func TestScorer_KindDiversityBonus(t *testing.T) {
	_, clk := fixedClock(time.Unix(1000, 0))
	s := NewScorer(50, 100, 10, clk)
	// Two DIFFERENT 40-weight kinds: raw 80, with +25% diversity → 100 → kill.
	s.Add(Evidence{PID: 3, Gen: 1, Kind: KindEntropyPlaceholder, Weight: 40})
	d := s.Add(Evidence{PID: 3, Gen: 1, Kind: KindExtChurnPlaceholder, Weight: 40})
	if d.Escalation != EscKill {
		t.Fatalf("diversity bonus missing: score=%v esc=%v", d.Score, d.Escalation)
	}
}
```

> Note: `KindEntropyPlaceholder` / `KindExtChurnPlaceholder` are test-only stand-ins for v2 fuzzy kinds; they let us exercise the generic scorer with sub-max weights in v1. Define them in `evidence.go` now (the v2 plan renames/uses them for real sensors).

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./edr/ransomware/ -run TestScorer -v`
Expected: FAIL — package/types undefined.

- [ ] **Step 3: Implement `evidence.go`**

```go
package ransomware

import "time"

// SignalKind identifies a ransomware detector's output. v1 ships the two
// high-fidelity kinds; the *Placeholder kinds reserve weights for v2 fuzzy
// sensors and keep the scorer exercised with sub-max evidence today.
type SignalKind string

const (
	KindCanary SignalKind = "canary"
	KindT1490  SignalKind = "t1490"

	KindEntropyPlaceholder  SignalKind = "entropy"    // v2
	KindExtChurnPlaceholder SignalKind = "ext_churn"  // v2
)

// DefaultWeights maps a kind to its evidence weight. Canary and T1490 are
// max-weight (a single hit crosses the default kill threshold of 100).
var DefaultWeights = map[SignalKind]float64{
	KindCanary:              100,
	KindT1490:               100,
	KindEntropyPlaceholder:  40,
	KindExtChurnPlaceholder: 40,
}

// Evidence is one detector observation about a process. Gen is the process
// generation (proctable StartTS) used to invalidate a recycled PID.
type Evidence struct {
	PID    int
	Gen    int64
	Kind   SignalKind
	Weight float64
	Detail string
	TS     time.Time
}
```

- [ ] **Step 4: Implement `scorer.go`**

```go
package ransomware

import (
	"math"
	"sort"
	"sync"
	"time"
)

type Escalation int

const (
	EscNone Escalation = iota
	EscSuspend
	EscKill
)

// Decision is the scorer's verdict for one Add call.
type Decision struct {
	Escalation Escalation
	Score      float64
	TopSignal  string // "<kind>:<detail>" of the heaviest contributing evidence
	Kinds      []SignalKind
}

type pidState struct {
	gen        int64
	score      float64
	lastUpdate time.Time
	kinds      map[SignalKind]float64 // kind → heaviest weight seen (for diversity + top signal)
	topKind    SignalKind
	topDetail  string
	topWeight  float64
	maxEsc     Escalation // highest escalation already emitted (prevents re-firing)
}

// Scorer accumulates weighted, PID-tagged evidence with exponential time-decay
// and a kind-diversity bonus, mapping the running score to an escalation.
// Add/Forget are called concurrently (ETW file feed + WMI process dispatch), so
// the state map is mutex-guarded.
type Scorer struct {
	mu              sync.Mutex
	suspendT, killT float64
	halfLife        float64 // seconds
	now             func() time.Time
	states          map[int]*pidState
}

func NewScorer(suspendT, killT, halfLife float64, now func() time.Time) *Scorer {
	if now == nil {
		now = time.Now
	}
	return &Scorer{suspendT: suspendT, killT: killT, halfLife: halfLife, now: now, states: map[int]*pidState{}}
}

func (s *Scorer) Forget(pid int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.states, pid)
}

func (s *Scorer) Add(ev Evidence) Decision {
	s.mu.Lock()
	defer s.mu.Unlock()
	t := ev.TS
	if t.IsZero() {
		t = s.now()
	}
	st := s.states[ev.PID]
	if st == nil || (ev.Gen != 0 && st.gen != ev.Gen) {
		st = &pidState{gen: ev.Gen, lastUpdate: t, kinds: map[SignalKind]float64{}}
		s.states[ev.PID] = st
	}
	// Decay the existing score to `t`.
	if s.halfLife > 0 {
		dt := t.Sub(st.lastUpdate).Seconds()
		if dt > 0 {
			st.score *= math.Pow(2, -dt/s.halfLife)
		}
	}
	st.lastUpdate = t
	st.score += ev.Weight
	if ev.Weight > st.kinds[ev.Kind] {
		st.kinds[ev.Kind] = ev.Weight
	}
	if ev.Weight >= st.topWeight {
		st.topWeight, st.topKind, st.topDetail = ev.Weight, ev.Kind, ev.Detail
	}

	// Diversity bonus: +25% per distinct kind beyond the first.
	mult := 1 + 0.25*float64(len(st.kinds)-1)
	eff := st.score * mult

	esc := EscNone
	switch {
	case eff >= s.killT:
		esc = EscKill
	case eff >= s.suspendT:
		esc = EscSuspend
	}
	// Only emit an escalation that is strictly higher than one already emitted.
	out := EscNone
	if esc > st.maxEsc {
		st.maxEsc = esc
		out = esc
	}
	return Decision{Escalation: out, Score: eff, TopSignal: signalString(st.topKind, st.topDetail), Kinds: sortedKinds(st.kinds)}
}

func signalString(k SignalKind, detail string) string {
	if detail == "" {
		return string(k)
	}
	return string(k) + ":" + detail
}

func sortedKinds(m map[SignalKind]float64) []SignalKind {
	ks := make([]SignalKind, 0, len(m))
	for k := range m {
		ks = append(ks, k)
	}
	sort.Slice(ks, func(i, j int) bool { return ks[i] < ks[j] })
	return ks
}
```

- [ ] **Step 5: Run tests to verify they pass**

Run: `go test ./edr/ransomware/ -run TestScorer -v`
Expected: PASS (all five).

- [ ] **Step 6: Commit** (surface to user)

```bash
git add edr/ransomware/evidence.go edr/ransomware/scorer.go edr/ransomware/scorer_test.go
git commit -m "feat(edr): ransomware per-PID decaying scorer + evidence model"
```

---

### Task 5: T1490 recovery-tampering command rules

**Files:**
- Create: `edr/ransomware/rules.go`
- Test: `edr/ransomware/rules_test.go`

**Interfaces:**
- Produces: `ransomware.MatchT1490(image, cmdline string, allowlist []string) (ruleName string, matched bool)`.

- [ ] **Step 1: Write the failing tests**

```go
package ransomware

import "testing"

func TestMatchT1490_Positives(t *testing.T) {
	cases := []struct {
		img, cmd, want string
	}{
		{`C:\Windows\System32\vssadmin.exe`, `vssadmin  delete   shadows /all /quiet`, "vssadmin_delete_shadows"},
		{`C:\Windows\System32\wbem\WMIC.exe`, `wmic shadowcopy delete`, "wmic_shadowcopy_delete"},
		{`C:\Windows\System32\wbadmin.exe`, `wbadmin delete catalog -quiet`, "wbadmin_delete_catalog"},
		{`C:\Windows\System32\bcdedit.exe`, `bcdedit /set {default} recoveryenabled no`, "bcdedit_recovery_disable"},
		{`C:\Windows\System32\reagentc.exe`, `reagentc /disable`, "reagentc_disable"},
	}
	for _, c := range cases {
		name, ok := MatchT1490(c.img, c.cmd, nil)
		if !ok || name != c.want {
			t.Errorf("MatchT1490(%q) = %q,%v ; want %q,true", c.cmd, name, ok, c.want)
		}
	}
}

func TestMatchT1490_NegativesAndAllowlist(t *testing.T) {
	if _, ok := MatchT1490(`C:\Windows\System32\vssadmin.exe`, `vssadmin list shadows`, nil); ok {
		t.Error("listing shadows is benign")
	}
	if _, ok := MatchT1490(`C:\app\backup.exe`, `backup.exe --run`, nil); ok {
		t.Error("unrelated command matched")
	}
	// Allowlisted exact command is exempt (e.g. a sanctioned backup job).
	if _, ok := MatchT1490(`C:\Windows\System32\wbadmin.exe`, `wbadmin delete catalog -quiet`,
		[]string{"wbadmin delete catalog -quiet"}); ok {
		t.Error("allowlisted command should not match")
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./edr/ransomware/ -run TestMatchT1490 -v`
Expected: FAIL — `undefined: MatchT1490`.

- [ ] **Step 3: Implement `rules.go`**

```go
package ransomware

import "strings"

// t1490Rule matches a recovery/backup-destruction command (MITRE T1490). All
// tokens must be present (in order-independent substring form) in the
// whitespace-normalized, lower-cased command line.
type t1490Rule struct {
	name   string
	tokens []string
}

var t1490Rules = []t1490Rule{
	{"vssadmin_delete_shadows", []string{"vssadmin", "delete", "shadows"}},
	{"vssadmin_resize_shadowstorage", []string{"vssadmin", "resize", "shadowstorage"}},
	{"wmic_shadowcopy_delete", []string{"wmic", "shadowcopy", "delete"}},
	{"wbadmin_delete_catalog", []string{"wbadmin", "delete", "catalog"}},
	{"wbadmin_delete_backup", []string{"wbadmin", "delete", "backup"}},
	{"bcdedit_recovery_disable", []string{"bcdedit", "recoveryenabled", "no"}},
	{"bcdedit_ignore_failures", []string{"bcdedit", "bootstatuspolicy", "ignoreallfailures"}},
	{"reagentc_disable", []string{"reagentc", "/disable"}},
	{"diskshadow_delete", []string{"diskshadow", "delete"}},
}

func normCmd(s string) string { return strings.Join(strings.Fields(strings.ToLower(s)), " ") }

// MatchT1490 returns the first matching recovery-tampering rule name for the
// given process image + command line, or ("",false). An allowlisted exact
// (normalized) command line is exempt (sanctioned admin/backup tooling).
func MatchT1490(image, cmdline string, allowlist []string) (string, bool) {
	nc := normCmd(cmdline)
	for _, a := range allowlist {
		if normCmd(a) == nc {
			return "", false
		}
	}
	for _, r := range t1490Rules {
		all := true
		for _, tok := range r.tokens {
			if !strings.Contains(nc, tok) {
				all = false
				break
			}
		}
		if all {
			return r.name, true
		}
	}
	return "", false
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./edr/ransomware/ -run TestMatchT1490 -v`
Expected: PASS.

- [ ] **Step 5: Commit** (surface to user)

```bash
git add edr/ransomware/rules.go edr/ransomware/rules_test.go
git commit -m "feat(edr): T1490 recovery-tampering command rules"
```

---

### Task 6: Canary manager (planner + membership + hideFile)

**Files:**
- Create: `edr/ransomware/canary.go`, `edr/ransomware/canary_windows.go`, `edr/ransomware/canary_other.go`
- Test: `edr/ransomware/canary_test.go`

**Interfaces:**
- Produces:
  - `ransomware.CanarySpec{Dir, Name string}`; `ransomware.PlanCanaries(dirs []string, perDir int) []CanarySpec`.
  - `ransomware.canaryStore` interface `{StoreCanary(cache.CanaryRecord) error; ListCanaries() ([]cache.CanaryRecord, error)}` (satisfied by `*cache.Cache`).
  - `ransomware.NewManager(store canaryStore) *Manager`; `(*Manager) Plant(dirs []string, perDir int) (int, error)`; `(*Manager) Load() error`; `(*Manager) Contains(path string) bool`; `(*Manager) Count() int`.
  - `hideFile(path string) error` (windows sets Hidden+System; stub no-op).

- [ ] **Step 1: Write the failing tests**

```go
package ransomware

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/utmstack/UTMStack/agent/edr/cache"
)

type fakeCanaryStore struct{ recs []cache.CanaryRecord }

func (f *fakeCanaryStore) StoreCanary(r cache.CanaryRecord) error {
	f.recs = append(f.recs, r)
	return nil
}
func (f *fakeCanaryStore) ListCanaries() ([]cache.CanaryRecord, error) { return f.recs, nil }

func TestPlanCanaries_NamesSortEarly(t *testing.T) {
	specs := PlanCanaries([]string{`C:\Users\x\Documents`}, 2)
	if len(specs) != 2 {
		t.Fatalf("got %d specs", len(specs))
	}
	for _, s := range specs {
		if s.Name[0] != '0' {
			t.Errorf("canary %q should sort early (lead with a digit)", s.Name)
		}
	}
	if specs[0].Name == specs[1].Name {
		t.Error("canary names must be distinct per dir")
	}
}

func TestManager_PlantPersistsAndMatches(t *testing.T) {
	dir := t.TempDir()
	fs := &fakeCanaryStore{}
	m := NewManager(fs)
	n, err := m.Plant([]string{dir}, 2)
	if err != nil || n != 2 {
		t.Fatalf("plant = %d, %v", n, err)
	}
	// Files exist on disk, are non-empty, and registered.
	entries, _ := os.ReadDir(dir)
	if len(entries) != 2 {
		t.Fatalf("expected 2 canary files, got %d", len(entries))
	}
	p := filepath.Join(dir, entries[0].Name())
	if !m.Contains(p) {
		t.Errorf("planted canary %q not recognized", p)
	}
	// Case-insensitive / separator-agnostic membership.
	if !m.Contains(filepath.ToSlash(p)) {
		t.Error("membership must be separator-agnostic")
	}
	if len(fs.recs) != 2 {
		t.Errorf("canaries not persisted: %d", len(fs.recs))
	}
}

func TestManager_LoadRehydratesMembership(t *testing.T) {
	fs := &fakeCanaryStore{recs: []cache.CanaryRecord{{Path: `C:\Users\x\00__a.xlsx`}}}
	m := NewManager(fs)
	if err := m.Load(); err != nil {
		t.Fatal(err)
	}
	if !m.Contains(`c:\users\x\00__A.xlsx`) {
		t.Error("Load did not rehydrate membership")
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./edr/ransomware/ -run 'TestPlanCanaries|TestManager' -v`
Expected: FAIL — undefined symbols.

- [ ] **Step 3: Implement `canary.go`**

```go
package ransomware

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/utmstack/UTMStack/agent/edr/cache"
)

// canaryExts and canaryWords generate decoy names that (a) look like real
// user documents so an encryptor targets them, and (b) sort early in a
// directory listing (lead with digits) so an alphabetical walk hits them
// before real data. We deliberately avoid the "~$" Office lock-file prefix,
// which many encryptors skip as a temp file.
var canaryExts = []string{"xlsx", "docx", "pdf"}
var canaryWords = []string{"accounts", "passwords", "backup_keys", "payroll", "clients", "contracts"}

// canaryBody is inert, low-entropy filler so a canary reads as an ordinary
// document (not itself flagged as high-entropy).
const canaryBody = "CONFIDENTIAL - internal records. Do not distribute. " +
	"This document is retained for archival purposes only.\n"

type CanarySpec struct {
	Dir  string
	Name string
}

// PlanCanaries produces perDir deterministic decoy specs for each dir.
func PlanCanaries(dirs []string, perDir int) []CanarySpec {
	if perDir < 1 {
		perDir = 1
	}
	var out []CanarySpec
	for _, d := range dirs {
		for i := 0; i < perDir; i++ {
			word := canaryWords[i%len(canaryWords)]
			ext := canaryExts[i%len(canaryExts)]
			// e.g. "00__accounts.xlsx", "01__passwords.docx"
			name := "0" + string(rune('0'+i%10)) + "__" + word + "." + ext
			out = append(out, CanarySpec{Dir: d, Name: name})
		}
	}
	return out
}

type canaryStore interface {
	StoreCanary(cache.CanaryRecord) error
	ListCanaries() ([]cache.CanaryRecord, error)
}

// Manager plants and tracks canary files and answers membership queries.
type Manager struct {
	store canaryStore
	mu    sync.RWMutex
	set   map[string]bool // normalized path → true
}

func NewManager(store canaryStore) *Manager {
	return &Manager{store: store, set: map[string]bool{}}
}

func normCanaryPath(p string) string {
	return strings.ToLower(filepath.ToSlash(strings.TrimRight(p, `/\`)))
}

// Load rehydrates the in-memory membership set from persisted records.
func (m *Manager) Load() error {
	recs, err := m.store.ListCanaries()
	if err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, r := range recs {
		m.set[normCanaryPath(r.Path)] = true
	}
	return nil
}

// Plant writes perDir canaries into each dir, hides them, persists them, and
// registers membership. Returns the number successfully planted. Directories
// that don't exist are skipped (best-effort; a volume may lack the path).
func (m *Manager) Plant(dirs []string, perDir int) (int, error) {
	planted := 0
	for _, spec := range PlanCanaries(dirs, perDir) {
		if _, err := os.Stat(spec.Dir); err != nil {
			continue
		}
		full := filepath.Join(spec.Dir, spec.Name)
		if err := os.WriteFile(full, []byte(canaryBody), 0o644); err != nil {
			continue
		}
		_ = hideFile(full)
		sum := sha256.Sum256([]byte(canaryBody))
		rec := cache.CanaryRecord{Path: full, SHA256: hex.EncodeToString(sum[:]), Volume: filepath.VolumeName(full)}
		if err := m.store.StoreCanary(rec); err != nil {
			return planted, err
		}
		m.mu.Lock()
		m.set[normCanaryPath(full)] = true
		m.mu.Unlock()
		planted++
	}
	return planted, nil
}

func (m *Manager) Contains(path string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.set[normCanaryPath(path)]
}

func (m *Manager) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.set)
}
```

- [ ] **Step 4: Implement `canary_windows.go` and `canary_other.go`**

`canary_windows.go`:
```go
//go:build windows

package ransomware

import "golang.org/x/sys/windows"

// hideFile sets Hidden+System so the decoy is invisible to ordinary users
// (any tamper is therefore attacker-driven, not an accidental user edit).
func hideFile(path string) error {
	p, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return err
	}
	return windows.SetFileAttributes(p, windows.FILE_ATTRIBUTE_HIDDEN|windows.FILE_ATTRIBUTE_SYSTEM)
}
```

`canary_other.go`:
```go
//go:build !windows

package ransomware

// hideFile is a no-op off Windows (canary logic is unit-tested on the host).
func hideFile(path string) error { return nil }
```

- [ ] **Step 5: Run tests to verify they pass**

Run: `go test ./edr/ransomware/ -run 'TestPlanCanaries|TestManager' -v`
Expected: PASS.

- [ ] **Step 6: Commit** (surface to user)

```bash
git add edr/ransomware/canary.go edr/ransomware/canary_windows.go edr/ransomware/canary_other.go edr/ransomware/canary_test.go
git commit -m "feat(edr): canary decoy manager (plant/persist/membership)"
```

---

### Task 7: File-activity feed interface + portable filter

**Files:**
- Create: `edr/ransomware/feed.go`
- Test: `edr/ransomware/feed_test.go`

**Interfaces:**
- Produces:
  - `ransomware.FileOp` with `OpCreate, OpWrite, OpRename, OpDelete, OpSetInfo`.
  - `ransomware.FileEvent{PID int; Path string; Op FileOp}`.
  - `ransomware.FileActivityFeed` interface: `Run(ctx context.Context, sink func(FileEvent)) error`.
  - `ransomware.Consider(ev FileEvent, isCanary func(string) bool, excluded func(string) bool, selfPID int) bool` — the portable pre-filter (drop reads/closes, self-PID, excluded paths; keep canary touches + user-data writes).

- [ ] **Step 1: Write the failing tests**

```go
package ransomware

import "testing"

func TestConsider_KeepsCanaryTamper(t *testing.T) {
	isCanary := func(p string) bool { return p == `C:\u\00__a.xlsx` }
	ex := func(string) bool { return false }
	ev := FileEvent{PID: 40, Path: `C:\u\00__a.xlsx`, Op: OpWrite}
	if !Consider(ev, isCanary, ex, 9) {
		t.Fatal("canary write must be considered")
	}
}

func TestConsider_DropsSelfExcludedAndReads(t *testing.T) {
	isCanary := func(string) bool { return false }
	ex := func(p string) bool { return p == `C:\Program Files\UTMStack\edr` }
	// self PID dropped
	if Consider(FileEvent{PID: 9, Path: `C:\u\a.txt`, Op: OpWrite}, isCanary, ex, 9) {
		t.Error("self PID must be dropped")
	}
	// excluded path dropped
	if Consider(FileEvent{PID: 5, Path: `C:\Program Files\UTMStack\edr`, Op: OpWrite}, isCanary, ex, 9) {
		t.Error("excluded path must be dropped")
	}
	// non-mutating op dropped
	if Consider(FileEvent{PID: 5, Path: `C:\u\a.txt`, Op: OpCreate}, isCanary, ex, 9) {
		t.Error("OpCreate alone (no write/rename/delete) is not a ransomware signal in v1")
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./edr/ransomware/ -run TestConsider -v`
Expected: FAIL — undefined symbols.

- [ ] **Step 3: Implement `feed.go`**

```go
package ransomware

import "context"

type FileOp int

const (
	OpCreate FileOp = iota
	OpWrite
	OpRename
	OpDelete
	OpSetInfo
)

// FileEvent is a per-process file operation. PID is the process that performed
// the op — the attribution USN cannot provide, which is why v1 uses an ETW
// feed for this stream.
type FileEvent struct {
	PID  int
	Path string
	Op   FileOp
}

// FileActivityFeed delivers per-process file operations to sink until ctx is
// cancelled. v1 implementation: ETW Microsoft-Windows-Kernel-File (Task 11).
// The interface makes the feed swappable (e.g. a USN-fallback impl) without
// touching the guard.
type FileActivityFeed interface {
	Run(ctx context.Context, sink func(FileEvent)) error
}

// Consider is the portable pre-filter applied before evidence is generated:
// keep only mutating ops (write/rename/delete/setinfo), drop the EDR's own PID
// and excluded paths, and always keep a canary tamper (highest signal). This
// keeps the guard's evidence stream small and self-noise-free.
func Consider(ev FileEvent, isCanary func(string) bool, excluded func(string) bool, selfPID int) bool {
	if ev.PID == selfPID {
		return false
	}
	// A canary tamper is always kept, even under an operator-excluded directory —
	// it is the highest-confidence signal and canary dirs can overlap exclusions.
	// (Checked BEFORE excluded so an excluded canary dir can't silently drop it.)
	if isCanary != nil && isCanary(ev.Path) {
		return true
	}
	if excluded != nil && excluded(ev.Path) {
		return false
	}
	switch ev.Op {
	case OpWrite, OpRename, OpDelete, OpSetInfo:
		return true
	default:
		return false
	}
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./edr/ransomware/ -run TestConsider -v`
Expected: PASS.

- [ ] **Step 5: Commit** (surface to user)

```bash
git add edr/ransomware/feed.go edr/ransomware/feed_test.go
git commit -m "feat(edr): file-activity feed interface + portable filter"
```

---

### Task 8: Guard — sensors → scorer → escalation ladder

**Files:**
- Create: `edr/ransomware/guard.go`
- Test: `edr/ransomware/guard_test.go`

**Interfaces:**
- Consumes: `config.RansomwareConfig`, `config.EDRConfig`; `proctable.Table` (concrete: `Get`, `Descendants`); `event.Spool` (via `appender`); `MatchT1490`, `Scorer`, `Manager`, `FileEvent`/`Consider`, `cache.RansomwareIncident`.
- Produces:
  - Local interfaces: `treeResponder{KillTree(int,string)([]int,error); Suspend(int)error; Resume(int)error}`, `appender{Append(string)error}`, `quarantiner{Quarantine(path,sha,det string)(string,error)}`, `incidentSink{StoreIncident(cache.RansomwareIncident) error}`, `canarySet{Contains(string)bool}`.
  - `ransomware.GuardDeps{Cfg config.EDRConfig; Table *proctable.Table; Resp treeResponder; Spool appender; Quar quarantiner; Incidents incidentSink; Canaries canarySet; Hash func(string)(string,error); Now func() time.Time; NewID func() string}`.
  - `ransomware.NewGuard(deps GuardDeps) *Guard`; `(*Guard) OnProcStart(pid, ppid int, image, cmdline string, gen int64)`; `(*Guard) OnFileEvent(ev FileEvent)`.

- [ ] **Step 1: Write the failing tests**

```go
package ransomware

import (
	"testing"
	"time"

	"github.com/utmstack/UTMStack/agent/edr/cache"
	"github.com/utmstack/UTMStack/agent/edr/config"
	"github.com/utmstack/UTMStack/agent/edr/proctable"
)

type fakeResp struct {
	killed    []int
	suspended []int
	resumed   []int
}

func (f *fakeResp) KillTree(root int, sig string) ([]int, error) { f.killed = append(f.killed, root); return []int{root}, nil }
func (f *fakeResp) Suspend(pid int) error                        { f.suspended = append(f.suspended, pid); return nil }
func (f *fakeResp) Resume(pid int) error                         { f.resumed = append(f.resumed, pid); return nil }

type fakeAppender struct{ lines []string }

func (f *fakeAppender) Append(s string) error { f.lines = append(f.lines, s); return nil }

type fakeQuar struct{ calls int }

func (f *fakeQuar) Quarantine(path, sha, det string) (string, error) { f.calls++; return "q-" + path, nil }

type fakeIncidents struct{ recs []cache.RansomwareIncident }

func (f *fakeIncidents) StoreIncident(r cache.RansomwareIncident) error { f.recs = append(f.recs, r); return nil }

type fakeCanaries struct{ paths map[string]bool }

func (f *fakeCanaries) Contains(p string) bool { return f.paths[p] }

func newTestGuard(mode string, canaries map[string]bool) (*Guard, *fakeResp, *fakeAppender, *fakeQuar, *fakeIncidents, *proctable.Table) {
	tab := proctable.New()
	resp := &fakeResp{}
	app := &fakeAppender{}
	quar := &fakeQuar{}
	inc := &fakeIncidents{}
	cfg := config.Default()
	cfg.Ransomware.Enabled = true
	cfg.Ransomware.ResponseMode = mode
	g := NewGuard(GuardDeps{
		Cfg: cfg, Table: tab, Resp: resp, Spool: app, Quar: quar, Incidents: inc,
		Canaries:  &fakeCanaries{paths: canaries},
		Hash:      func(string) (string, error) { return "deadbeef", nil },
		Now:       func() time.Time { return time.Unix(1000, 0) },
		NewID:     func() string { return "inc-1" },
	})
	return g, resp, app, quar, inc, tab
}

func TestGuard_T1490_SuspendMode_KillsParentAndQuarantines(t *testing.T) {
	g, resp, app, quar, inc, tab := newTestGuard("suspend", nil)
	// Encryptor (pid 100) spawns vssadmin (pid 200).
	tab.Add(proctable.Proc{PID: 100, PPID: 4, Image: `C:\Users\x\enc.exe`})
	tab.Add(proctable.Proc{PID: 200, PPID: 100, Image: `C:\Windows\System32\vssadmin.exe`})

	g.OnProcStart(200, 100, `C:\Windows\System32\vssadmin.exe`, "vssadmin delete shadows /all /quiet", 1)

	if len(resp.suspended) == 0 || resp.suspended[0] != 100 {
		t.Fatalf("suspend-mode must suspend the culprit (parent 100): %v", resp.suspended)
	}
	if len(resp.killed) == 0 || resp.killed[0] != 100 {
		t.Fatalf("must kill the encryptor parent (100), got %v", resp.killed)
	}
	if quar.calls != 1 {
		t.Fatalf("must quarantine the encryptor image once, got %d", quar.calls)
	}
	if len(inc.recs) != 1 || inc.recs[0].Action != "contained" {
		t.Fatalf("incident not recorded: %+v", inc.recs)
	}
	// A ransomware_contained event was emitted.
	found := false
	for _, l := range app.lines {
		if contains(l, "ransomware_contained") {
			found = true
		}
	}
	if !found {
		t.Fatalf("no ransomware_contained event: %v", app.lines)
	}
}

func TestGuard_Canary_KillMode_NoSuspend(t *testing.T) {
	g, resp, _, quar, _, tab := newTestGuard("kill", map[string]bool{`C:\u\00__a.xlsx`: true})
	tab.Add(proctable.Proc{PID: 300, PPID: 4, Image: `C:\Users\x\enc.exe`})
	g.OnFileEvent(FileEvent{PID: 300, Path: `C:\u\00__a.xlsx`, Op: OpWrite})

	if len(resp.suspended) != 0 {
		t.Fatalf("kill mode must NOT suspend: %v", resp.suspended)
	}
	if len(resp.killed) == 0 || resp.killed[0] != 300 {
		t.Fatalf("canary tamper must kill toucher 300: %v", resp.killed)
	}
	if quar.calls != 1 {
		t.Fatalf("must quarantine, got %d", quar.calls)
	}
}

func TestGuard_AlertMode_EmitsButNeverActs(t *testing.T) {
	g, resp, app, quar, inc, tab := newTestGuard("alert", map[string]bool{`C:\u\00__a.xlsx`: true})
	tab.Add(proctable.Proc{PID: 400, PPID: 4, Image: `C:\Users\x\enc.exe`})
	g.OnFileEvent(FileEvent{PID: 400, Path: `C:\u\00__a.xlsx`, Op: OpWrite})

	if len(resp.killed) != 0 || len(resp.suspended) != 0 || quar.calls != 0 {
		t.Fatalf("alert mode must not act: kill=%v susp=%v quar=%d", resp.killed, resp.suspended, quar.calls)
	}
	if len(inc.recs) != 1 || inc.recs[0].Action != "alerted" {
		t.Fatalf("alert incident missing: %+v", inc.recs)
	}
	if len(app.lines) == 0 {
		t.Fatal("alert mode must still emit an event")
	}
}

func TestGuard_Disabled_NoOp(t *testing.T) {
	g, resp, _, _, _, tab := newTestGuard("kill", map[string]bool{`C:\u\00__a.xlsx`: true})
	g.cfg.Ransomware.Enabled = false
	tab.Add(proctable.Proc{PID: 500, PPID: 4, Image: `C:\enc.exe`})
	g.OnFileEvent(FileEvent{PID: 500, Path: `C:\u\00__a.xlsx`, Op: OpWrite})
	if len(resp.killed) != 0 {
		t.Fatal("disabled guard must do nothing")
	}
}

func contains(s, sub string) bool { return len(s) >= len(sub) && (indexOf(s, sub) >= 0) }
func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./edr/ransomware/ -run TestGuard -v`
Expected: FAIL — `NewGuard`/`GuardDeps` undefined.

- [ ] **Step 3: Implement `guard.go`**

```go
package ransomware

import (
	"context"
	"strings"
	"time"

	"github.com/utmstack/UTMStack/agent/edr/cache"
	"github.com/utmstack/UTMStack/agent/edr/config"
	"github.com/utmstack/UTMStack/agent/edr/event"
	"github.com/utmstack/UTMStack/agent/edr/proctable"
	"github.com/utmstack/UTMStack/shared/logger"
)

type treeResponder interface {
	KillTree(rootPID int, signature string) ([]int, error)
	Suspend(pid int) error
	Resume(pid int) error
}
type appender interface{ Append(string) error }
type quarantiner interface {
	Quarantine(path, sha256, detection string) (string, error)
}
type incidentSink interface{ StoreIncident(cache.RansomwareIncident) error }
type canarySet interface{ Contains(path string) bool }

// GuardDeps bundles the guard's collaborators (interfaces where a fake is
// useful, concrete where trivial) so the escalation logic is unit-testable.
type GuardDeps struct {
	Cfg       config.EDRConfig
	Table     *proctable.Table
	Resp      treeResponder
	Spool     appender
	Quar      quarantiner
	Incidents incidentSink
	Canaries  canarySet
	Hash      func(string) (string, error)
	Now       func() time.Time
	NewID     func() string
}

// Guard fuses ransomware sensors into per-process risk and runs the escalation
// ladder against the existing response arm.
type Guard struct {
	cfg    config.EDRConfig
	deps   GuardDeps
	scorer *Scorer
}

func NewGuard(deps GuardDeps) *Guard {
	rc := deps.Cfg.Ransomware
	if deps.Now == nil {
		deps.Now = time.Now
	}
	half := float64(rc.DecayHalfLifeMs) / 1000.0
	sc := NewScorer(float64(rc.SuspendThreshold), float64(rc.KillThreshold), half, deps.Now)
	return &Guard{cfg: deps.Cfg, deps: deps, scorer: sc}
}

// OnProcStart runs the T1490 command-rule sensor for a new process. Attribution
// is to the culprit (the parent that spawned the recovery command, skipping
// shell hosts), because vssadmin/wbadmin/etc. are the encryptor's children.
func (g *Guard) OnProcStart(pid, ppid int, image, cmdline string, gen int64) {
	if !g.cfg.Ransomware.Enabled {
		return
	}
	rule, ok := MatchT1490(image, cmdline, g.cfg.Ransomware.CommandAllowlist)
	if !ok {
		return
	}
	culprit := g.culpritPID(ppid)
	// Attribute by the culprit's OWN generation (StartTS), not the child's `gen`,
	// so two sensors on the same culprit don't spuriously reset each other's score.
	ev := Evidence{PID: culprit, Gen: g.genOf(culprit, 0), Kind: KindT1490,
		Weight: DefaultWeights[KindT1490], Detail: rule, TS: g.deps.Now()}
	g.handle(g.scorer.Add(ev), culprit)
}

// OnFileEvent runs the canary sensor for a per-process file op.
func (g *Guard) OnFileEvent(fe FileEvent) {
	if !g.cfg.Ransomware.Enabled {
		return
	}
	if !g.deps.Canaries.Contains(fe.Path) {
		return // v1: only canary touches produce evidence from the file stream
	}
	base := fe.Path
	if i := strings.LastIndexAny(base, `/\`); i >= 0 {
		base = base[i+1:]
	}
	ev := Evidence{PID: fe.PID, Gen: g.genOf(fe.PID, 0), Kind: KindCanary,
		Weight: DefaultWeights[KindCanary], Detail: base, TS: g.deps.Now()}
	g.handle(g.scorer.Add(ev), fe.PID)
}

func (g *Guard) genOf(pid int, gen int64) int64 {
	if gen != 0 {
		return gen
	}
	if p, ok := g.deps.Table.Get(pid); ok {
		return p.StartTS
	}
	return 0
}

// culpritPID walks up past known shell/host interpreters so we blame the
// encryptor, not the cmd.exe/powershell.exe it used to run vssadmin.
func (g *Guard) culpritPID(pid int) int {
	shells := map[string]bool{"cmd.exe": true, "powershell.exe": true, "pwsh.exe": true, "conhost.exe": true, "wscript.exe": true, "cscript.exe": true}
	for hops := 0; hops < 4; hops++ {
		p, ok := g.deps.Table.Get(pid)
		if !ok {
			return pid
		}
		base := strings.ToLower(p.Image)
		if i := strings.LastIndexAny(base, `/\`); i >= 0 {
			base = base[i+1:]
		}
		if !shells[base] || p.PPID == 0 {
			return pid
		}
		pid = p.PPID
	}
	return pid
}

// handle executes the escalation ladder for a scorer decision on culpritPID.
func (g *Guard) handle(d Decision, pid int) {
	if d.Escalation == EscNone {
		return
	}
	p, _ := g.deps.Table.Get(pid)
	pinfo := event.ProcInfo{PID: pid, PPID: p.PPID, Image: p.Image, Cmdline: p.Cmdline}
	mode := normalizeMode(g.cfg.Ransomware.ResponseMode)

	// alert mode: surface the escalation but never suspend/kill/quarantine.
	if mode == "alert" {
		sev := "high"
		if d.Escalation == EscKill {
			sev = "critical"
		}
		g.emit(event.ActionRansomwareSuspected, pinfo, d, sev)
		g.record("alerted", pid, p, d, "")
		return
	}

	// suspend|kill mode, sub-kill escalation: suspend mode freezes the suspect;
	// kill mode takes no containment action until the kill threshold.
	if d.Escalation == EscSuspend {
		suspended := false
		if mode == "suspend" {
			_ = g.deps.Resp.Suspend(pid)
			suspended = true
		}
		g.emit(event.ActionRansomwareSuspected, pinfo, d, "high")
		if suspended {
			g.record("suspended", pid, p, d, "")
		} else {
			g.record("suspected", pid, p, d, "")
		}
		return
	}

	// EscKill in suspend|kill mode: contain.
	if mode == "suspend" {
		_ = g.deps.Resp.Suspend(pid) // freeze the tree to reduce further encryption during the kill
	}
	killed, killErr := g.deps.Resp.KillTree(pid, "ransomware:"+d.TopSignal)
	if killErr != nil {
		logger.Error("UTMStack EDR: ransomware kill tree %d: %v", pid, killErr)
	}
	qid := ""
	quarFailed := false
	if p.Image != "" {
		sha := ""
		if g.deps.Hash != nil {
			sha, _ = g.deps.Hash(p.Image)
		}
		if id, err := g.deps.Quar.Quarantine(p.Image, sha, "ransomware:"+d.TopSignal); err == nil {
			qid = id
		} else {
			quarFailed = true
			logger.Error("UTMStack EDR: ransomware quarantine %s: %v", p.Image, err)
		}
	}
	// Honest reporting (project rule): responder.KillTree logs+swallows per-PID kill
	// errors and returns nil, so the honest "terminated nothing" signal is an empty
	// `killed` slice — not killErr. Don't claim "contained" unless we killed something
	// and quarantine didn't fail.
	action := "contained"
	if killErr != nil || len(killed) == 0 || quarFailed {
		action = "contain_failed"
	}
	g.emit(event.ActionRansomwareContained, pinfo, d, "critical")
	g.record(action, pid, p, d, qid)
	g.scorer.Forget(pid)
}

// normalizeMode maps a configured response_mode to one of the three known modes,
// defaulting unknown/empty/mis-cased values to the LEAST destructive ("alert")
// so a typo can never cause auto-kill.
func normalizeMode(m string) string {
	switch strings.ToLower(strings.TrimSpace(m)) {
	case "kill":
		return "kill"
	case "suspend":
		return "suspend"
	default:
		return "alert"
	}
}

func (g *Guard) emit(action string, p event.ProcInfo, d Decision, sev string) {
	if js, err := event.NewRansomwareEvent(action, p, d.TopSignal, sev).ToJSON(); err == nil {
		_ = g.deps.Spool.Append(js)
	}
}

func (g *Guard) record(action string, pid int, p proctable.Proc, d Decision, qid string) {
	id := "incident"
	if g.deps.NewID != nil {
		id = g.deps.NewID()
	}
	kinds := make([]string, 0, len(d.Kinds))
	for _, k := range d.Kinds {
		kinds = append(kinds, string(k))
	}
	_ = g.deps.Incidents.StoreIncident(cache.RansomwareIncident{
		ID: id, PID: pid, Image: p.Image, Cmdline: p.Cmdline,
		Signals: strings.Join(kinds, ","), Score: int(d.Score), Action: action, QuarantineID: qid,
	})
}

// Run is the long-lived driver: subscribe to the file-activity feed and route
// each considered event to OnFileEvent. Process-start evidence arrives via
// OnProcStart (called from the procwatch dispatch hook). selfPID excludes the
// EDR's own file ops; excluded filters self/high-churn paths.
//
// Run SUPERVISES the feed: if it exits unexpectedly (e.g. an ETW session drop)
// Run restarts it after a bounded backoff, so the canary sensor cannot die
// silently while status still reports protection. It tracks liveness in the
// atomic feedHealthy so writeStatus can surface `ransomware_feed_healthy`.
func (g *Guard) Run(ctx context.Context, feed FileActivityFeed, selfPID int, excluded func(string) bool) error {
	sink := func(fe FileEvent) {
		if Consider(fe, g.deps.Canaries.Contains, excluded, selfPID) {
			g.OnFileEvent(fe)
		}
	}
	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		g.feedHealthy.Store(true)
		err := feed.Run(ctx, sink)
		g.feedHealthy.Store(false)
		if ctx.Err() != nil {
			return ctx.Err() // clean shutdown
		}
		logger.Error("UTMStack EDR: ransomware file feed exited (%v); restarting", err)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(feedRetryBackoff):
		}
	}
}
```

> **Post-review hardening (final whole-branch review).** Add to this file: `import "sync/atomic"`; a package var `var feedRetryBackoff = 5 * time.Second` (testable seam); a `feedHealthy atomic.Bool` field on `Guard`; and `func (g *Guard) FeedHealthy() bool { return g.feedHealthy.Load() }`. In `service.go` add a `statusDoc` field `RansomwareFeedHealthy bool` (`json:"ransomware_feed_healthy"`), a `rwGuard *ransomware.Guard` field on `program` set in `startPipeline`, and pass it to `writeStatus` (set `rwGuard != nil && rwGuard.FeedHealthy()`). This makes a persistently-dead feed visible instead of a silent protection gap.

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./edr/ransomware/ -run TestGuard -v`
Expected: PASS (all four).

- [ ] **Step 5: Run the whole package + labeling audit**

Run:
```bash
go test ./edr/ransomware/... -v
grep -rniE '(logger\.(Info|Error|Fatal|Debug)|fmt\.(Print|Sprint))[^\n]*clam' edr/ransomware/
```
Expected: all tests PASS; grep prints nothing.

- [ ] **Step 6: Commit** (surface to user)

```bash
git add edr/ransomware/guard.go edr/ransomware/guard_test.go
git commit -m "feat(edr): ransomware guard escalation ladder (suspend/kill/quarantine)"
```

---

### Task 9: procwatch dispatch hook + StartTS

**Files:**
- Modify: `edr/procwatch/dispatch.go`
- Test: `edr/procwatch/dispatch_test.go` (create)

**Interfaces:**
- Produces: `procwatch.NewDispatch(tab *proctable.Table, g *guard.Guard, sp *event.Spool, skip func(string) bool, onProc func(ProcStart)) *Dispatch` (adds trailing `onProc` param). `OnProcStart` now stamps `StartTS` and invokes `onProc` (if non-nil) after adding to the table.

- [ ] **Step 1: Write the failing test**

```go
package procwatch

import (
	"testing"

	"github.com/utmstack/UTMStack/agent/edr/proctable"
)

func TestDispatch_InvokesOnProcAndStampsStartTS(t *testing.T) {
	tab := proctable.New()
	var got ProcStart
	d := NewDispatch(tab, nil, nil, nil, func(ps ProcStart) { got = ps })

	d.OnProcStart(ProcStart{PID: 77, PPID: 4, Image: `C:\enc.exe`, Cmdline: "enc.exe"})

	if got.PID != 77 {
		t.Fatalf("onProc not invoked with the ProcStart: %+v", got)
	}
	p, ok := tab.Get(77)
	if !ok || p.StartTS == 0 {
		t.Fatalf("StartTS not stamped in table: %+v ok=%v", p, ok)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./edr/procwatch/ -run TestDispatch_InvokesOnProc -v`
Expected: FAIL — `NewDispatch` arity mismatch.

- [ ] **Step 3: Update `dispatch.go`**

Add `"time"` to imports. Change the struct, constructor, and `OnProcStart`:

```go
type Dispatch struct {
	table  *proctable.Table
	guard  *guard.Guard
	spool  *event.Spool
	skip   func(image string) bool
	onProc func(ProcStart) // optional: ransomware T1490 sensor hook
}

func NewDispatch(tab *proctable.Table, g *guard.Guard, sp *event.Spool, skip func(string) bool, onProc func(ProcStart)) *Dispatch {
	return &Dispatch{table: tab, guard: g, spool: sp, skip: skip, onProc: onProc}
}

func (d *Dispatch) OnProcStart(ps ProcStart) {
	p := proctable.Proc{PID: ps.PID, PPID: ps.PPID, Image: ps.Image, Cmdline: ps.Cmdline, StartTS: time.Now().UnixNano()}
	d.table.Add(p)

	// Ransomware T1490 sensor: match recovery-tampering commands on the raw
	// process start. Runs before the skip filter so vssadmin/wbadmin under
	// %SystemRoot% are still evaluated (the encryptor is their parent).
	if d.onProc != nil {
		d.onProc(ps)
	}

	if d.skip != nil && d.skip(p.Image) {
		return
	}

	if d.spool != nil {
		if js, err := behavioral.ProcTelemetry(p).ToJSON(); err == nil {
			_ = d.spool.Append(js)
		}
	}

	if d.guard != nil {
		go func() {
			defer func() {
				if r := recover(); r != nil {
					logger.Error("UTMStack EDR: panic in guard.OnStart: %v", r)
				}
			}()
			d.guard.OnStart(p)
		}()
	}
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./edr/procwatch/ -run TestDispatch_InvokesOnProc -v`
Expected: PASS.

- [ ] **Step 5: Commit** (surface to user)

```bash
git add edr/procwatch/dispatch.go edr/procwatch/dispatch_test.go
git commit -m "feat(edr): procwatch dispatch onProc hook + StartTS stamping"
```

---

### Task 10: Service wiring + status fields

**Files:**
- Modify: `edr/service/service.go` (statusDoc ~line 51; startPipeline ~line 183; writeStatus signature/body)
- Verify: build only (native feed is Task 11); no new unit test — covered by cross-build + Task 12 VM acceptance.

**Interfaces:**
- Consumes: `ransomware.NewGuard`, `ransomware.NewManager`, `ransomware.GuardDeps`, `ransomware.FileEvent`, `ransomware.NewFeed` (Task 11), `procwatch.NewDispatch` (5-arg, Task 9).
- Produces: guard running under `goSafe("ransomware", …)`; `statusDoc` fields `RansomwareEnabled bool`, `RansomwareMode string`, `CanaryCount int`.

- [ ] **Step 1: Add status fields to `statusDoc`**

In `edr/service/service.go`, add to `statusDoc` (after `Product`):

```go
	RansomwareEnabled bool   `json:"ransomware_enabled"`
	RansomwareMode    string `json:"ransomware_mode"`
	CanaryCount       int    `json:"ransomware_canary_count"`
```

- [ ] **Step 2: Thread canary count into `writeStatus`**

Change `writeStatus` to accept the manager and set the fields. Update its signature and the two call sites (lines ~95 and ~102) to pass a `*ransomware.Manager` (nil before the pipeline starts):

```go
func writeStatus(cfg config.EDRConfig, eng *engine.Engine, canaries *ransomware.Manager) {
	healthy := engine.Ping(cfg.ClamdAddr) == nil
	ver := ""
	if healthy {
		raw, _ := engine.Version(cfg.ClamdAddr)
		ver = sigDBVersion(raw)
	}
	tn := eng.Tuning()
	cc := 0
	if canaries != nil {
		cc = canaries.Count()
	}
	doc := statusDoc{
		Running: true, EngineHealthy: healthy, SigDBVersion: ver,
		UpdatedAt: time.Now().UTC().Format(time.RFC3339), Product: "UTMStack EDR",
		EngineTier: string(tn.Tier), EngineResident: tn.ResidentViable,
		EngineThreads: tn.MaxThreads, EngineMaxFileMB: tn.MaxFileSizeMB, EngineMaxScanMB: tn.MaxScanSizeMB,
		EngineReload: tn.ConcurrentReload, EngineTuningNote: tn.Reason,
		RansomwareEnabled: cfg.Ransomware.Enabled, RansomwareMode: cfg.Ransomware.ResponseMode,
		CanaryCount: cc,
	}
	b, _ := json.MarshalIndent(doc, "", "  ")
	_ = os.WriteFile(config.StatusFile, b, 0o644)
}
```

Because the manager is created inside `startPipeline`, hoist it to a package-level (or `program`-level) handle set by `startPipeline` so `run()`'s ticker can read it. Add a field to `program`:

```go
type program struct {
	cancel   context.CancelFunc
	eng      *engine.Engine
	canaries *ransomware.Manager // set by startPipeline; read by writeStatus
}
```

Update `run()`'s two `writeStatus` calls to `writeStatus(cfg, p.eng, p.canaries)`.

- [ ] **Step 3: Wire the guard into `startPipeline`**

Add `"github.com/utmstack/UTMStack/agent/edr/ransomware"` to the import block. Inside `startPipeline`, after the process-kill chain is built (after line ~183 where `pw` is created) and BEFORE `goSafe("feed", …)`, insert:

```go
	// Ransomware guard: canary tripwires + T1490 command rules → per-PID scorer →
	// suspend/kill/quarantine. Attribution for file signals comes from the ETW
	// file-activity feed; T1490 attribution comes from the procwatch dispatch hook.
	var rwGuard *ransomware.Guard
	var onProc func(procwatch.ProcStart)
	if cfg.Ransomware.Enabled {
		mgr := ransomware.NewManager(c)
		_ = mgr.Load()
		canaryDirs := append(defaultCanaryDirs(), cfg.Ransomware.CanaryDirs...)
		if n, err := mgr.Plant(canaryDirs, cfg.Ransomware.CanaryPerDir); err != nil {
			logger.Error("UTMStack EDR: canary plant: %v", err)
		} else {
			logger.Info("UTMStack EDR: planted %d canaries", n)
		}
		p.canaries = mgr
		rwGuard = ransomware.NewGuard(ransomware.GuardDeps{
			Cfg: cfg, Table: tab, Resp: resp, Spool: sp, Quar: store, Incidents: c,
			Canaries: mgr, Hash: scanner.SHA256File, Now: time.Now, NewID: uuid.NewString,
		})
		onProc = func(ps procwatch.ProcStart) {
			rwGuard.OnProcStart(ps.PID, ps.PPID, ps.Image, ps.Cmdline, 0)
		}
		selfPID := os.Getpid()
		feed := ransomware.NewFeed()
		goSafe("ransomware", func() { _ = rwGuard.Run(ctx, feed, selfPID, ex.Excluded) })
	}
	pw := procwatch.New(procwatch.NewDispatch(tab, g, sp, ex.Excluded, onProc))
	goSafe("procwatch", func() { pw.Run(ctx) })
```

Remove the OLD `pw := …` / `goSafe("procwatch", …)` lines (183–184) that this block replaces (the new block builds `pw` with the 5-arg `NewDispatch`). Add `"github.com/google/uuid"` to imports (already used elsewhere in the module).

- [ ] **Step 4: Add `defaultCanaryDirs` helper (portable)**

Create `edr/service/canarydirs.go`:

```go
package service

import (
	"os"
	"path/filepath"
)

// defaultCanaryDirs returns the baseline directories seeded with canaries:
// each watched volume root plus common user-data folders under the profile.
// Extra dirs from config are appended by the caller. defaultFixedVolumes() is
// the same package-local helper the watcher uses (volumes_windows.go / _other.go).
func defaultCanaryDirs() []string {
	var dirs []string
	// defaultFixedVolumes() returns bare drive specs ("C:") for the USN watcher.
	// Anchor each to the drive ROOT before planting: filepath.Join("C:", name) is
	// drive-RELATIVE ("C:name") on Windows, which would plant the decoy in the
	// service CWD and produce a stored path that never matches the ETW feed. With
	// a trailing separator, filepath.Join("C:\\", name) is the absolute root path.
	for _, v := range defaultFixedVolumes() {
		if v == "" {
			continue
		}
		if !os.IsPathSeparator(v[len(v)-1]) {
			v += string(os.PathSeparator)
		}
		dirs = append(dirs, v)
	}
	if up := os.Getenv("USERPROFILE"); up != "" {
		for _, sub := range []string{"Desktop", "Documents", "Downloads", "Pictures"} {
			dirs = append(dirs, filepath.Join(up, sub))
		}
	}
	return dirs
}
```

> Note: `defaultFixedVolumes()` already exists (`volumes_windows.go`/`volumes_other.go`). On macOS it returns a stub, so this compiles host-side; real roots come from the Windows impl.

- [ ] **Step 5: Cross-build both arches (native feed still stubbed until Task 11)**

Run:
```bash
GOOS=windows GOARCH=amd64 go build ./...
GOOS=windows GOARCH=arm64 go build ./...
go build ./...   # host build uses the !windows stubs
```
Expected: all succeed (the `ransomware.NewFeed()` symbol resolves to the stub from Task 11's `feed_other.go`; if Task 11 is not yet done, add a temporary stub — but sequence Task 11 before this build).

- [ ] **Step 6: Commit** (surface to user)

```bash
git add edr/service/service.go edr/service/canarydirs.go
git commit -m "feat(edr): wire ransomware guard + canaries into service pipeline"
```

---

### Task 11: ETW file-activity feed (native, VM-verified)

**Files:**
- Create: `edr/ransomware/feed_windows.go`, `edr/ransomware/feed_other.go`
- Verify: cross-build both arches; **runtime behavior verified on the Windows VM in Task 12** (ETW cannot run on macOS — per the project testing model, native pieces are compile-checked on the host and run-verified on the VM).

**Interfaces:**
- Produces: `ransomware.NewFeed() FileActivityFeed`. Windows impl consumes `Microsoft-Windows-Kernel-File`; stub blocks until ctx done.

- [ ] **Step 1: Implement the `!windows` stub `feed_other.go`**

```go
//go:build !windows

package ransomware

import "context"

type nopFeed struct{}

// NewFeed returns a no-op feed off Windows so the module builds/tests on the
// host. The real ETW feed is Windows-only.
func NewFeed() FileActivityFeed { return nopFeed{} }

func (nopFeed) Run(ctx context.Context, sink func(FileEvent)) error {
	<-ctx.Done()
	return ctx.Err()
}
```

- [ ] **Step 2: Implement the Windows ETW feed `feed_windows.go`**

> The ETW consumer uses `github.com/0xrawsec/golang-etw` (add via `go get github.com/0xrawsec/golang-etw@latest` from `utmstack-v12/agent/`). **Confirm the exact API against the pinned version** — the shape below matches its Consumer/Session model; adjust field/method names if the pinned release differs. This is the §12 open dependency; the guard depends only on `FileActivityFeed`, so a different lib (or a raw `golang.org/x/sys/windows` StartTrace/ProcessTrace implementation) can replace this file without touching tested code.

```go
//go:build windows

package ransomware

import (
	"context"
	"strings"

	"github.com/0xrawsec/golang-etw/etw"
	"github.com/utmstack/UTMStack/shared/logger"
)

// Microsoft-Windows-Kernel-File provider GUID.
const kernelFileProviderGUID = "{EDD08927-9CC4-4E65-B970-C2560FB5C289}"

// Kernel-File event IDs we care about (mutating ops only — reads/closes are
// dropped to keep the stream small):
//   12 = Create, 30 = SetInformation, 26 = DeletePath (rename/delete),
//   32 = Write. IDs vary slightly by OS build; map defensively.
func opForEventID(id uint16) (FileOp, bool) {
	switch id {
	case 32:
		return OpWrite, true
	case 30:
		return OpSetInfo, true
	case 26:
		return OpDelete, true
	case 27:
		return OpRename, true
	case 12:
		return OpCreate, true
	default:
		return 0, false
	}
}

type etwFeed struct{}

func NewFeed() FileActivityFeed { return &etwFeed{} }

func (f *etwFeed) Run(ctx context.Context, sink func(FileEvent)) error {
	session := etw.NewRealTimeSession("UTMStackEDR-RansomFile")
	defer session.Stop()

	prov, err := etw.ParseProvider(kernelFileProviderGUID)
	if err != nil {
		return err
	}
	// Enable only file-write/rename/delete/setinfo/create keywords where the
	// provider supports keyword filtering; otherwise we filter by event ID below.
	if err := session.EnableProvider(prov); err != nil {
		return err
	}

	c := etw.NewConsumer(ctx).FromSessions(session)
	c.EventCallback = func(e *etw.Event) error {
		op, ok := opForEventID(uint16(e.System.EventID))
		if !ok {
			return nil
		}
		pid := int(e.System.Execution.ProcessID)
		// Kernel-File carries the target in "FileName" (or "OpenPath"); try both.
		path, _ := e.GetPropertyString("FileName")
		if path == "" {
			path, _ = e.GetPropertyString("OpenPath")
		}
		if path == "" {
			return nil
		}
		sink(FileEvent{PID: pid, Path: normalizeKernelPath(path), Op: op})
		return nil
	}

	if err := c.Start(); err != nil {
		return err
	}
	logger.Info("UTMStack EDR: ransomware file-activity feed (ETW) started")
	<-ctx.Done()
	_ = c.Stop()
	return ctx.Err()
}

// normalizeKernelPath maps a kernel device path (e.g. "\\Device\\HarddiskVolume3\\...")
// toward a drive-letter path where possible; falls back to the raw path. The
// canary registry stores drive-letter paths, so best-effort normalization keeps
// membership matching working (Task 12 validates on the VM and tightens this).
func normalizeKernelPath(p string) string {
	// Placeholder-free minimal normalization: trim the NT prefix; VM acceptance
	// (Task 12) confirms whether device→letter mapping is needed for canary hits.
	return strings.TrimPrefix(p, `\??\`)
}
```

- [ ] **Step 3: Cross-build both arches**

Run:
```bash
go get github.com/0xrawsec/golang-etw@latest
GOOS=windows GOARCH=amd64 go build ./...
GOOS=windows GOARCH=arm64 go build ./...
go build ./...   # host uses feed_other.go stub
go vet ./edr/...
```
Expected: all succeed. If the library's API differs, fix `feed_windows.go` until the windows build passes (host build/tests are unaffected — they use the stub).

- [ ] **Step 4: Commit** (surface to user)

```bash
git add edr/ransomware/feed_windows.go edr/ransomware/feed_other.go go.mod go.sum
git commit -m "feat(edr): ETW Kernel-File activity feed (windows) + host stub"
```

---

### Task 12: VM acceptance — real ransomware behavior end-to-end

**Files:** none created in-repo. This task **runs on the Windows VM** (`10.211.55.12`, `ricardovald1d15\atlas`) per the project testing model, and verifies **final outputs**, not just logs. A throwaway synthetic encryptor is used against a sandbox tree — never real user data.

**Preconditions:** ClamAV `clamd` reachable at `127.0.0.1:3310` on the VM (per CLAUDE.md External deps); EDR built for the VM's arch; `edr.json` with `"ransomware": {"enabled": true, "response_mode": "suspend", "canary_per_dir": 1}`.

- [ ] **Step 1: Cross-build + deploy**

Run (from `utmstack-v12/agent/`):
```bash
GOOS=windows GOARCH=amd64 go build -o dist/utmstack_edr_windows_amd64.exe ./edr
```
Copy `dist/utmstack_edr_windows_amd64.exe` to the VM (via the established `prlctl`/share path from the Plan 3 memory), install/enable the SYSTEM service.

- [ ] **Step 2: Verify canaries were planted**

On the VM, confirm decoy files exist (Hidden+System) in volume roots + profile folders, and that `status.json` shows `"ransomware_enabled": true` and `"ransomware_canary_count" > 0`.
Expected: several `0X__word.ext` hidden files present; status reflects them.

- [ ] **Step 3: Canary tripwire → contain**

Build a throwaway "encryptor" (Go or PowerShell) that, in a sandbox dir, also opens and overwrites one planted canary file. Run it as a normal user.
Expected outputs (verify each):
- The encryptor process (and tree) is **killed** (`Get-Process` shows it gone).
- Its binary is **moved to the quarantine store** (present under `quarantine/<uuid>.quarantined`, absent from origin).
- A `RansomwareIncident` row exists in `edr.db` with `Action="contained"`, the canary in `Signals`.
- Branded events reached the spool: a `ransomware_contained` event with `product/engine="UTMStack EDR"`, plus `killed` events. No field contains `clam`.

- [ ] **Step 4: T1490 command → contain**

From a throwaway parent process, run `vssadmin delete shadows /all /quiet` (on the VM, where shadow copies are expendable).
Expected: the **parent** (culprit) process is contained (killed + quarantined); a `ransomware_contained` event with signal `t1490:vssadmin_delete_shadows`; incident row recorded. Confirm a benign `vssadmin list shadows` does **nothing**.

- [ ] **Step 5: Response-mode matrix**

Set `response_mode` to `"alert"`, restart, re-run Step 3's encryptor.
Expected: process is **not** killed/quarantined; a `ransomware_suspected` event is emitted; incident `Action="alerted"`. Restore `"suspend"` afterward.

- [ ] **Step 6: Labeling audit + both-arch build gate**

Run (host):
```bash
grep -rniE '(logger\.(Info|Error|Fatal|Debug)|fmt\.(Print|Sprint))[^\n]*clam' edr/
for arch in amd64 arm64; do
  GOOS=windows GOARCH=$arch go build ./...
  GOOS=windows GOARCH=$arch go vet ./edr/...
done
go test ./edr/... ./agent/
```
Expected: grep empty; both arches build+vet clean; all host tests pass.

- [ ] **Step 7: Commit acceptance notes** (surface to user)

Record VM results (screens/log excerpts of the final outputs) in the plan's task checklist or a short acceptance note, and commit any small fixes discovered (e.g. `normalizeKernelPath` device→letter mapping if canary hits didn't match). Update the `edr-plan*` memory with the new capability + any bugs fixed, mirroring the Plan 3/4 memory entries.

---

## Self-Review

**1. Spec coverage** (against `2026-07-02-edr-ransomware-guard-design.md`):
- §4.1 ETW file feed → Task 7 (interface/filter) + Task 11 (native). ✅
- §4.2 canary manager → Task 6. ✅
- §4.3 T1490 rules + procwatch hook → Task 5 + Task 9. ✅
- §4.4 scoring engine → Task 4. ✅
- §4.5 guard/ladder → Task 8. ✅
- §5.1 config → Task 2. §5.2 events → Task 1. §5.3 cache → Task 3. §5.4 status → Task 10. ✅
- §6 escalation ladder + response modes → Task 8 (unit) + Task 12 (VM matrix). ✅
- §7 honest limits, §8 branding → Global Constraints + labeling audit in Tasks 8 & 12. ✅
- §9 testing (pure macOS + VM final-output) → per-task tests + Task 12. ✅
- v2 items (entropy/extension/rate/ransom-note, registry, VSS rollback) are **out of scope** here (separate plan) — placeholder kinds reserved in Task 4. ✅

**2. Placeholder scan:** No "TBD/TODO/handle edge cases" left. The one honest deferral is `normalizeKernelPath` device→letter mapping, explicitly gated to Task 12 VM validation (it can't be finalized without observing real ETW paths on the VM) — this is a native-verification step, not a code placeholder.

**3. Type consistency:** `Escalation`/`Decision`/`Evidence`/`SignalKind` used consistently across Tasks 4→8. `GuardDeps` field names match the wiring in Task 10. `NewDispatch` 5-arg signature defined in Task 9 and consumed in Task 10. `NewFeed()`/`FileActivityFeed` defined in Task 7/11 and consumed in Task 10. `cache.CanaryRecord`/`RansomwareIncident` and accessors defined in Task 3 and used in Tasks 6/8/10. Consistent. ✅
