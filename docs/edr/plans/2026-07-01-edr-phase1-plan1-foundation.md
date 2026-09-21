# UTMStack EDR — Phase 1 Windows — Plan 1: Foundation

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Stand up a self-contained, agent-controlled `utmstack_edr` SYSTEM service that can scan a file through clamd, cache the verdict in its own SQLite DB, and deliver a UTMStack-EDR-branded event to the platform via the agent — controllable with `enable-edr` / `disable-edr` / `edr-status`.

**Architecture:** A new set of Go packages under `agent/edr/` compiled into a separate binary `utmstack_edr.exe` (same Go module as the agent, its own `main` package — the `updater` is the exact template). It supervises `clamd` as a child process, keeps its own `edr.db`, and writes normalized events to a durable disk spool. The agent ships the binary via its `Dependency` mechanism, drains the spool into its existing `LogQueue` (an `EDRRelay` goroutine), and exposes cobra verbs to start/stop/status the EDR service.

**Tech Stack:** Go 1.25.5, `kardianos/service`, `gorm.io/gorm` + `github.com/glebarez/sqlite`, `github.com/utmstack/UTMStack/shared/{fs,http,exec,svc,logger}`, `github.com/threatwinds/go-sdk/plugins`, clamd (localhost TCP socket, INSTREAM protocol).

## Global Constraints

Copied verbatim from `2026-07-01-edr-phase1-windows-design.md`. Every task implicitly includes these.

- **Target:** agent branch `release/v12.0.0`, module `github.com/utmstack/UTMStack/agent`, Go **1.25.5**. Do not bump the Go version or add dependencies beyond those already in `go.mod` except where a task explicitly says so.
- **Go-first:** all code in Go. (The one native exception, the AMSI DLL, is Plan 4 — not here.)
- **Branding / labeling (mandatory):** every emitted event and every surfaced log line says **"UTMStack EDR"**. The strings `clamav` / `clamd` must **never** appear in an emitted event field or a surfaced log message. Internal Go identifiers may use `clamd`.
- **Event contract:** `DataType = "utmstack_edr"`; event `source ∈ {file_watcher, process_watcher, amsi, edr_engine}`; `engine`/`product = "UTMStack EDR"`.
- **Storage isolation:** the EDR uses its **own** `edr.db`, never the agent's `logs.db`.
- **Quarantine, not delete** (relevant from Plan 2; nothing here deletes files).
- **No kernel driver** anywhere.
- **Privilege:** the EDR service installs as a Windows **SYSTEM** service.
- **Shared install dir:** all EDR files live under `fs.GetExecutablePath()` (the agent install dir), so the agent and EDR share one root.
- **Windows architectures (multi-arch):** the EDR must build and ship for **every Windows architecture the agent supports — `windows/amd64` and `windows/arm64`**. All Go code stays architecture-portable (no arch-specific assumptions; `unsafe.Sizeof`/struct layouts are per-arch and fine). The `dependency` package is split per-arch (`deps_windows_amd64.go`, `deps_windows_arm64.go`), so the EDR dependency entry must be registered in **both**. The EDR binary name carries the arch via `runtime.GOARCH` (`EDRFile()`), producing `utmstack_edr_windows_amd64.exe` / `utmstack_edr_windows_arm64.exe`. (Windows `386` is not an agent target — no `deps_windows_386.go` exists — so it is out of scope unless the agent adds it.)

**Build & test environment note.** Development host is macOS; Windows-only behavior (service install, clamd spawn, end-to-end) is verified on the Parallels VM `10.211.55.12` (`ricardovald1d15\atlas`). Pure-Go logic is TDD'd with `go test` on any OS. Where a step is Windows-only it says so and gives VM commands with expected output instead of a `go test` cycle.

**Cross-compile commands (from macOS) — build BOTH architectures:**
```bash
cd agent
for arch in amd64 arm64; do
  GOOS=windows GOARCH=$arch go build -o dist/utmstack_edr_windows_$arch.exe ./edr
done
```

---

## File Structure

**New EDR binary + packages (all under `agent/edr/`):**
- `agent/edr/main.go` — `package main`; arg switch (`install`/`uninstall`/`run`/`scan`/`status`). Mirrors `updater/main.go`.
- `agent/edr/config/config.go` — `EDRConfig` struct, defaults, `Load()`/`Save()`, path constants.
- `agent/edr/service/config.go` — kardianos `*service.Config` (`UTMStackEDR`).
- `agent/edr/service/service.go` — `program{}` (Start/Stop/run), `RunService()`.
- `agent/edr/service/install.go` — `InstallService()` / `UninstallService()`.
- `agent/edr/event/event.go` — `Event` struct + constructors + `ToJSON()` (branding).
- `agent/edr/event/spool.go` — `Spool` writer (append ndjson, size cap + rotate).
- `agent/edr/cache/cache.go` — `VerdictRecord` model, `Cache` (gorm/`edr.db`), `Lookup`/`Store`.
- `agent/edr/engine/clamd.go` — clamd INSTREAM client + `Ping`.
- `agent/edr/engine/engine.go` — clamd discovery + child-process supervision.
- `agent/edr/scanner/scanner.go` — `ScanFile()` orchestration (hash→cache→clamd→verdict→event).

**Agent-side control/relay additions:**
- `agent/agent/edr_relay.go` — `EDRRelay(ctx)` goroutine: drain spool → `LogQueue`.
- `agent/serv/service.go` — MODIFY: launch `EDRRelay` via `p.goSafe`.
- `agent/dependency/deps_windows_amd64.go` — MODIFY: add the `utmstack_edr` dependency + `EDRFile()` helper (new file `agent/dependency/edr.go` for the shared helper + hooks).
- `agent/cmd/enable_edr.go`, `agent/cmd/disable_edr.go`, `agent/cmd/edr_status.go` — cobra verbs.

**Tests:** `_test.go` beside each pure-Go package (`config`, `event`, `cache`, `engine` (protocol encoding), `scanner`, and the agent-side `edr_relay`).

---

## Task 1: EDR config package (defaults + load/save)

**Files:**
- Create: `agent/edr/config/config.go`
- Test: `agent/edr/config/config_test.go`

**Interfaces:**
- Produces:
  - `type EDRConfig struct { ... }` (fields below)
  - `func Default() EDRConfig`
  - `func Load() (EDRConfig, error)` — reads `<install>/edr.json`, filling any missing field from `Default()`
  - `func Save(c EDRConfig) error`
  - Path vars: `InstallDir string`, `ConfigFile`, `StatusFile`, `DBFile`, `SpoolDir`, `SpoolFile`, `QuarantineDir`, `EngineDir` (all under `fs.GetExecutablePath()`)
  - `const ServiceName = "UTMStackEDR"`

- [ ] **Step 1: Write the failing test**

```go
// agent/edr/config/config_test.go
package config

import "testing"

func TestDefaultHasSaneValues(t *testing.T) {
	d := Default()
	if d.FailMode != "open" {
		t.Fatalf("FailMode = %q, want open", d.FailMode)
	}
	if !d.SuspendOnLaunch {
		t.Fatal("SuspendOnLaunch should default true")
	}
	if d.ClamdAddr != "127.0.0.1:3310" {
		t.Fatalf("ClamdAddr = %q", d.ClamdAddr)
	}
	if d.ScanConcurrency < 1 {
		t.Fatalf("ScanConcurrency = %d, want >=1", d.ScanConcurrency)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd agent && go test ./edr/config/ -run TestDefaultHasSaneValues -v`
Expected: FAIL — package/`Default` undefined (build error).

- [ ] **Step 3: Write minimal implementation**

```go
// agent/edr/config/config.go
package config

import (
	"path/filepath"
	"runtime"

	"github.com/utmstack/UTMStack/shared/fs"
)

const ServiceName = "UTMStackEDR"

var (
	InstallDir    = fs.GetExecutablePath()
	ConfigFile    = filepath.Join(InstallDir, "edr.json")
	StatusFile    = filepath.Join(InstallDir, "status.json")
	DBFile        = filepath.Join(InstallDir, "edr.db")
	SpoolDir      = filepath.Join(InstallDir, "edr-spool")
	SpoolFile     = filepath.Join(SpoolDir, "events.ndjson")
	QuarantineDir = filepath.Join(InstallDir, "quarantine")
	EngineDir     = filepath.Join(InstallDir, "engine")
	LogFile       = filepath.Join(InstallDir, "logs", "utmstack_edr.log")
)

// EDRConfig is the operational configuration read by the EDR service and
// writable by the agent / central management.
type EDRConfig struct {
	Enabled          bool     `json:"enabled"`
	Server           string   `json:"server"`             // written by the agent on enable
	SkipCertValidate bool     `json:"skip_cert_validate"`
	WatchVolumes     []string `json:"watch_volumes"`      // Plan 2; empty = all fixed NTFS
	Exclusions       []string `json:"exclusions"`
	FailMode         string   `json:"fail_mode"`          // open | closed
	SuspendOnLaunch  bool     `json:"suspend_on_launch"`  // Plan 3
	SuspendTimeoutMs int      `json:"suspend_timeout_ms"` // Plan 3
	QuarantineDir    string   `json:"quarantine_dir"`
	QuarantineDays   int      `json:"quarantine_retention_days"`
	ClamdAddr        string   `json:"clamd_addr"`
	ClamdPath        string   `json:"clamd_path"` // optional explicit path for testing
	ScanConcurrency  int      `json:"scan_concurrency"`
	SigUpdateHours   int      `json:"sig_update_hours"`
}

func Default() EDRConfig {
	conc := runtime.NumCPU()
	if conc > 4 {
		conc = 4
	}
	if conc < 1 {
		conc = 1
	}
	return EDRConfig{
		Enabled:          false,
		FailMode:         "open",
		SuspendOnLaunch:  true,
		SuspendTimeoutMs: 5000,
		QuarantineDir:    QuarantineDir,
		QuarantineDays:   30,
		ClamdAddr:        "127.0.0.1:3310",
		ScanConcurrency:  conc,
		SigUpdateHours:   4,
	}
}

// Load reads edr.json, filling missing/zero fields from Default().
func Load() (EDRConfig, error) {
	c := Default()
	if !fs.Exists(ConfigFile) {
		return c, nil
	}
	var onDisk EDRConfig
	if err := fs.ReadJSON(ConfigFile, &onDisk); err != nil {
		return c, err
	}
	// Overlay non-zero on-disk values onto defaults.
	if onDisk.FailMode != "" {
		c.FailMode = onDisk.FailMode
	}
	if onDisk.ClamdAddr != "" {
		c.ClamdAddr = onDisk.ClamdAddr
	}
	if onDisk.ScanConcurrency > 0 {
		c.ScanConcurrency = onDisk.ScanConcurrency
	}
	if onDisk.QuarantineDir != "" {
		c.QuarantineDir = onDisk.QuarantineDir
	}
	c.Enabled = onDisk.Enabled
	c.Server = onDisk.Server
	c.SkipCertValidate = onDisk.SkipCertValidate
	c.SuspendOnLaunch = onDisk.SuspendOnLaunch
	c.ClamdPath = onDisk.ClamdPath
	c.WatchVolumes = onDisk.WatchVolumes
	c.Exclusions = onDisk.Exclusions
	return c, nil
}

func Save(c EDRConfig) error {
	if err := fs.CreateDirIfNotExist(InstallDir); err != nil {
		return err
	}
	return fs.WriteJSON(ConfigFile, c)
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `cd agent && go test ./edr/config/ -run TestDefaultHasSaneValues -v`
Expected: PASS.

- [ ] **Step 5: Add a round-trip test and re-run**

```go
// append to agent/edr/config/config_test.go
func TestLoadReturnsDefaultWhenNoFile(t *testing.T) {
	// Point ConfigFile at a non-existent temp path.
	orig := ConfigFile
	ConfigFile = t.TempDir() + "/nope.json"
	defer func() { ConfigFile = orig }()
	c, err := Load()
	if err != nil {
		t.Fatalf("Load err: %v", err)
	}
	if c.FailMode != "open" {
		t.Fatalf("expected default FailMode, got %q", c.FailMode)
	}
}
```

Run: `cd agent && go test ./edr/config/ -v`
Expected: PASS (both tests).

- [ ] **Step 6: Commit**

```bash
git add agent/edr/config/
git commit -m "feat(edr): config package with defaults and load/save"
```

---

## Task 2: Event model + durable spool

**Files:**
- Create: `agent/edr/event/event.go`, `agent/edr/event/spool.go`
- Test: `agent/edr/event/event_test.go`, `agent/edr/event/spool_test.go`

**Interfaces:**
- Consumes: `config` paths (Task 1).
- Produces:
  - `const DataType = "utmstack_edr"`, `const Product = "UTMStack EDR"`
  - `type Event struct { Timestamp, HostID, OS, Product, Source, Action, ObjectPath, FileHash, Verdict, Signature, Engine, Severity string; Process *ProcInfo }`
  - `func NewDetection(objectPath, sha256, signature, source string) Event`
  - `func (e Event) ToJSON() (string, error)` — sets Product/Engine/OS/HostID/Timestamp if empty
  - `type Spool struct{}`, `func OpenSpool(path string, maxBytes int64) (*Spool, error)`, `func (s *Spool) Append(line string) error`, `func (s *Spool) Close() error`

- [ ] **Step 1: Write the failing test (branding + no forbidden strings)**

```go
// agent/edr/event/event_test.go
package event

import (
	"strings"
	"testing"
)

func TestDetectionJSONIsBrandedAndClean(t *testing.T) {
	e := NewDetection(`C:\Users\x\a.exe`, "abc123", "Win.Test.EICAR", SourceEngine)
	js, err := e.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON: %v", err)
	}
	if !strings.Contains(js, `"product":"UTMStack EDR"`) {
		t.Fatalf("missing product branding: %s", js)
	}
	if !strings.Contains(js, `"engine":"UTMStack EDR"`) {
		t.Fatalf("missing engine branding: %s", js)
	}
	low := strings.ToLower(js)
	if strings.Contains(low, "clamav") || strings.Contains(low, "clamd") {
		t.Fatalf("event leaked engine name: %s", js)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd agent && go test ./edr/event/ -run TestDetectionJSONIsBrandedAndClean -v`
Expected: FAIL — undefined `NewDetection`/`SourceEngine`.

- [ ] **Step 3: Write minimal implementation**

```go
// agent/edr/event/event.go
package event

import (
	"encoding/json"
	"os"
	"time"
)

const (
	DataType = "utmstack_edr"
	Product  = "UTMStack EDR"

	SourceFileWatcher    = "file_watcher"
	SourceProcessWatcher = "process_watcher"
	SourceAMSI           = "amsi"
	SourceEngine         = "edr_engine"

	ActionDetected     = "detected"
	ActionQuarantined  = "quarantined"
	ActionKilled       = "killed"
	ActionScanComplete = "scan_complete"
	ActionHealth       = "health"
)

type ProcInfo struct {
	PID     int    `json:"pid"`
	PPID    int    `json:"ppid"`
	Image   string `json:"image"`
	Cmdline string `json:"command_line,omitempty"`
	Session int    `json:"session,omitempty"`
	User    string `json:"user,omitempty"`
}

type Event struct {
	Timestamp  string    `json:"timestamp"`
	HostID     string    `json:"host_id"`
	OS         string    `json:"os"`
	Product    string    `json:"product"`
	Source     string    `json:"source"`
	Action     string    `json:"action"`
	ObjectPath string    `json:"object_path,omitempty"`
	FileHash   string    `json:"file_hash,omitempty"`
	Verdict    string    `json:"verdict,omitempty"`
	Signature  string    `json:"signature,omitempty"`
	Engine     string    `json:"engine"`
	Severity   string    `json:"severity,omitempty"`
	Process    *ProcInfo `json:"process,omitempty"`
}

func NewDetection(objectPath, sha256, signature, source string) Event {
	return Event{
		Source:     source,
		Action:     ActionDetected,
		ObjectPath: objectPath,
		FileHash:   sha256,
		Verdict:    "malicious",
		Signature:  signature,
	}
}

func (e Event) ToJSON() (string, error) {
	if e.Product == "" {
		e.Product = Product
	}
	if e.Engine == "" {
		e.Engine = Product
	}
	if e.OS == "" {
		e.OS = "windows"
	}
	if e.Timestamp == "" {
		e.Timestamp = time.Now().UTC().Format(time.RFC3339Nano)
	}
	if e.HostID == "" {
		if h, err := os.Hostname(); err == nil {
			e.HostID = h
		}
	}
	b, err := json.Marshal(e)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `cd agent && go test ./edr/event/ -run TestDetectionJSONIsBrandedAndClean -v`
Expected: PASS.

- [ ] **Step 5: Write the spool test**

```go
// agent/edr/event/spool_test.go
package event

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSpoolAppendAndRotate(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "events.ndjson")
	s, err := OpenSpool(path, 40) // tiny cap to force rotation
	if err != nil {
		t.Fatalf("OpenSpool: %v", err)
	}
	defer s.Close()
	for i := 0; i < 5; i++ {
		if err := s.Append(`{"n":` + strings.Repeat("0", 10) + `}`); err != nil {
			t.Fatalf("Append: %v", err)
		}
	}
	// Current spool file must exist and stay under a bounded size.
	fi, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if fi.Size() > 200 {
		t.Fatalf("spool not rotating, size=%d", fi.Size())
	}
}
```

- [ ] **Step 6: Implement the spool, run, verify pass**

```go
// agent/edr/event/spool.go
package event

import (
	"os"
	"path/filepath"
	"sync"
)

// Spool is an append-only newline-delimited JSON event buffer with a size cap.
// When the current file exceeds maxBytes it is rotated to "<path>.1" (single
// generation; the agent relay drains fast enough that one backup suffices).
type Spool struct {
	path     string
	maxBytes int64
	mu       sync.Mutex
	f        *os.File
}

func OpenSpool(path string, maxBytes int64) (*Spool, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return nil, err
	}
	return &Spool{path: path, maxBytes: maxBytes, f: f}, nil
}

func (s *Spool) Append(line string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if fi, err := s.f.Stat(); err == nil && fi.Size() >= s.maxBytes {
		if err := s.rotate(); err != nil {
			return err
		}
	}
	_, err := s.f.WriteString(line + "\n")
	return err
}

func (s *Spool) rotate() error {
	if err := s.f.Close(); err != nil {
		return err
	}
	_ = os.Rename(s.path, s.path+".1") // best-effort; overwrites prior backup
	f, err := os.OpenFile(s.path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return err
	}
	s.f = f
	return nil
}

func (s *Spool) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.f.Close()
}
```

Run: `cd agent && go test ./edr/event/ -v`
Expected: PASS (both tests).

- [ ] **Step 7: Commit**

```bash
git add agent/edr/event/
git commit -m "feat(edr): branded event model and durable disk spool"
```

---

## Task 3: Verdict cache (own SQLite)

**Files:**
- Create: `agent/edr/cache/cache.go`
- Test: `agent/edr/cache/cache_test.go`

**Interfaces:**
- Produces:
  - `type VerdictRecord struct { SHA256 string; Verdict, Signature string; SigDBVersion string; FirstSeen, LastSeen time.Time; Stale bool }`
  - `type Cache struct{}`, `func Open(dbFile string) (*Cache, error)`, `func (c *Cache) Close() error`
  - `func (c *Cache) Lookup(sha256 string) (VerdictRecord, bool, error)`
  - `func (c *Cache) Store(r VerdictRecord) error`
  - `const VerdictClean = "clean"`, `VerdictMalicious = "malicious"`, `VerdictUnknown = "unknown"`

- [ ] **Step 1: Write the failing test**

```go
// agent/edr/cache/cache_test.go
package cache

import (
	"path/filepath"
	"testing"
)

func TestStoreThenLookup(t *testing.T) {
	c, err := Open(filepath.Join(t.TempDir(), "edr.db"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer c.Close()

	if err := c.Store(VerdictRecord{SHA256: "hash1", Verdict: VerdictMalicious, Signature: "Win.Test.EICAR", SigDBVersion: "27000"}); err != nil {
		t.Fatalf("Store: %v", err)
	}
	got, found, err := c.Lookup("hash1")
	if err != nil || !found {
		t.Fatalf("Lookup found=%v err=%v", found, err)
	}
	if got.Verdict != VerdictMalicious || got.Signature != "Win.Test.EICAR" {
		t.Fatalf("bad record: %+v", got)
	}
	if _, found, _ := c.Lookup("missing"); found {
		t.Fatal("unexpected hit for missing hash")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd agent && go test ./edr/cache/ -run TestStoreThenLookup -v`
Expected: FAIL — undefined symbols.

- [ ] **Step 3: Write minimal implementation**

```go
// agent/edr/cache/cache.go
package cache

import (
	"errors"
	"sync"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const (
	VerdictClean     = "clean"
	VerdictMalicious = "malicious"
	VerdictUnknown   = "unknown"
)

type VerdictRecord struct {
	SHA256       string `gorm:"primaryKey"`
	Verdict      string
	Signature    string
	SigDBVersion string
	FirstSeen    time.Time
	LastSeen     time.Time
	Stale        bool
}

type Cache struct {
	db *gorm.DB
	mu sync.RWMutex
}

func Open(dbFile string) (*Cache, error) {
	db, err := gorm.Open(sqlite.Open(dbFile), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, err
	}
	if err := db.AutoMigrate(&VerdictRecord{}); err != nil {
		return nil, err
	}
	return &Cache{db: db}, nil
}

func (c *Cache) Lookup(sha256 string) (VerdictRecord, bool, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	var r VerdictRecord
	err := c.db.First(&r, "sha256 = ?", sha256).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return VerdictRecord{}, false, nil
	}
	if err != nil {
		return VerdictRecord{}, false, err
	}
	return r, true, nil
}

func (c *Cache) Store(r VerdictRecord) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	now := time.Now().UTC()
	if r.FirstSeen.IsZero() {
		r.FirstSeen = now
	}
	r.LastSeen = now
	return c.db.Save(&r).Error
}

func (c *Cache) Close() error {
	sqlDB, err := c.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `cd agent && go test ./edr/cache/ -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add agent/edr/cache/
git commit -m "feat(edr): verdict cache backed by its own SQLite edr.db"
```

---

## Task 4: clamd INSTREAM client + engine supervision

**Files:**
- Create: `agent/edr/engine/clamd.go`, `agent/edr/engine/engine.go`
- Test: `agent/edr/engine/clamd_test.go`

**Interfaces:**
- Consumes: `config` (Task 1) for `ClamdAddr`, `ClamdPath`, `EngineDir`.
- Produces:
  - `func ScanBytes(addr string, data []byte) (clean bool, signature string, err error)` — clamd INSTREAM
  - `func Ping(addr string) error`
  - `func Version(addr string) (string, error)`
  - `type Engine struct{}`, `func New(cfg config.EDRConfig) *Engine`, `func (e *Engine) EnsureRunning() error`, `func (e *Engine) Stop() error`

- [ ] **Step 1: Write the failing test (INSTREAM framing against a fake server)**

The INSTREAM wire format is: send `zINSTREAM\0`, then for each chunk a 4-byte big-endian length prefix followed by the bytes, then a zero-length (`\0\0\0\0`) terminator. clamd replies `stream: OK\0` or `stream: <sig> FOUND\0`. The test stands up a TCP server that speaks this protocol and returns a FOUND response, verifying our client parses the signature.

```go
// agent/edr/engine/clamd_test.go
package engine

import (
	"bufio"
	"encoding/binary"
	"io"
	"net"
	"strings"
	"testing"
)

func fakeClamd(t *testing.T, reply string) (addr string, stop func()) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		r := bufio.NewReader(conn)
		// read "zINSTREAM\0"
		_, _ = r.ReadString(0)
		// drain chunks until zero-length terminator
		for {
			var n uint32
			if err := binary.Read(r, binary.BigEndian, &n); err != nil {
				break
			}
			if n == 0 {
				break
			}
			if _, err := io.CopyN(io.Discard, r, int64(n)); err != nil {
				break
			}
		}
		_, _ = conn.Write([]byte(reply))
	}()
	return ln.Addr().String(), func() { ln.Close() }
}

func TestScanBytesParsesFound(t *testing.T) {
	addr, stop := fakeClamd(t, "stream: Win.Test.EICAR_HDB-1 FOUND\x00")
	defer stop()
	clean, sig, err := ScanBytes(addr, []byte("dummy"))
	if err != nil {
		t.Fatalf("ScanBytes: %v", err)
	}
	if clean {
		t.Fatal("expected not clean")
	}
	if !strings.Contains(sig, "EICAR") {
		t.Fatalf("signature = %q", sig)
	}
}

func TestScanBytesParsesOK(t *testing.T) {
	addr, stop := fakeClamd(t, "stream: OK\x00")
	defer stop()
	clean, _, err := ScanBytes(addr, []byte("dummy"))
	if err != nil {
		t.Fatalf("ScanBytes: %v", err)
	}
	if !clean {
		t.Fatal("expected clean")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd agent && go test ./edr/engine/ -run TestScanBytes -v`
Expected: FAIL — undefined `ScanBytes`.

- [ ] **Step 3: Implement the clamd client**

```go
// agent/edr/engine/clamd.go
package engine

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"net"
	"strings"
	"time"
)

const dialTimeout = 5 * time.Second

// ScanBytes streams data to clamd over INSTREAM and returns the verdict.
// NOTE: "clamd" here is an internal identifier only; nothing from this
// function is surfaced in events or logs (see branding rules).
func ScanBytes(addr string, data []byte) (bool, string, error) {
	conn, err := net.DialTimeout("tcp", addr, dialTimeout)
	if err != nil {
		return false, "", err
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(60 * time.Second))

	if _, err := conn.Write([]byte("zINSTREAM\x00")); err != nil {
		return false, "", err
	}
	const chunk = 8192
	for i := 0; i < len(data); i += chunk {
		end := i + chunk
		if end > len(data) {
			end = len(data)
		}
		var sz [4]byte
		binary.BigEndian.PutUint32(sz[:], uint32(end-i))
		if _, err := conn.Write(sz[:]); err != nil {
			return false, "", err
		}
		if _, err := conn.Write(data[i:end]); err != nil {
			return false, "", err
		}
	}
	if _, err := conn.Write([]byte{0, 0, 0, 0}); err != nil {
		return false, "", err
	}

	resp, err := bufio.NewReader(conn).ReadString(0)
	if err != nil {
		return false, "", err
	}
	resp = strings.TrimRight(resp, "\x00\n ")
	if strings.HasSuffix(resp, "OK") {
		return true, "", nil
	}
	if strings.HasSuffix(resp, "FOUND") {
		// "stream: <signature> FOUND"
		s := strings.TrimPrefix(resp, "stream: ")
		s = strings.TrimSuffix(s, " FOUND")
		return false, s, nil
	}
	return false, "", fmt.Errorf("unexpected engine response: %q", resp)
}

func Ping(addr string) error {
	conn, err := net.DialTimeout("tcp", addr, dialTimeout)
	if err != nil {
		return err
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(dialTimeout))
	if _, err := conn.Write([]byte("zPING\x00")); err != nil {
		return err
	}
	resp, err := bufio.NewReader(conn).ReadString(0)
	if err != nil {
		return err
	}
	if !strings.HasPrefix(strings.TrimSpace(resp), "PONG") {
		return fmt.Errorf("unexpected ping response: %q", resp)
	}
	return nil
}

func Version(addr string) (string, error) {
	conn, err := net.DialTimeout("tcp", addr, dialTimeout)
	if err != nil {
		return "", err
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(dialTimeout))
	if _, err := conn.Write([]byte("zVERSION\x00")); err != nil {
		return "", err
	}
	resp, err := bufio.NewReader(conn).ReadString(0)
	if err != nil {
		return "", err
	}
	return strings.TrimRight(resp, "\x00\n"), nil
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `cd agent && go test ./edr/engine/ -run TestScanBytes -v`
Expected: PASS (both).

- [ ] **Step 5: Implement engine supervision (Windows child process)**

The engine locates clamd (explicit `ClamdPath`, else `<EngineDir>/clamd.exe`), starts it as a child bound to `ClamdAddr`, and waits for `Ping` to succeed. Production signature download is deferred; for Plan 1 the engine assumes clamd binaries are present (installed on the VM or shipped later) and focuses on supervise + health.

```go
// agent/edr/engine/engine.go
package engine

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/utmstack/UTMStack/agent/edr/config"
	"github.com/utmstack/UTMStack/shared/fs"
	"github.com/utmstack/UTMStack/shared/logger"
)

type Engine struct {
	cfg config.EDRConfig
	cmd *exec.Cmd
}

func New(cfg config.EDRConfig) *Engine { return &Engine{cfg: cfg} }

func (e *Engine) binaryPath() string {
	if e.cfg.ClamdPath != "" {
		return e.cfg.ClamdPath
	}
	return filepath.Join(config.EngineDir, "clamd.exe")
}

// EnsureRunning starts the scan engine child process if it is not already
// answering on ClamdAddr, then blocks until it is healthy or times out.
func (e *Engine) EnsureRunning() error {
	if Ping(e.cfg.ClamdAddr) == nil {
		logger.Info("UTMStack EDR: scan engine already running")
		return nil
	}
	bin := e.binaryPath()
	if !fs.Exists(bin) {
		return fmt.Errorf("UTMStack EDR scan engine binary not found at %s", bin)
	}
	// clamd reads its listen address from its own config; we pass a config file
	// path under EngineDir. Generating that config is part of engine setup.
	cmd := exec.Command(bin, "--config-file", filepath.Join(config.EngineDir, "clamd.conf"), "--foreground")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("starting UTMStack EDR scan engine: %w", err)
	}
	e.cmd = cmd

	deadline := time.Now().Add(60 * time.Second)
	for time.Now().Before(deadline) {
		if Ping(e.cfg.ClamdAddr) == nil {
			logger.Info("UTMStack EDR: scan engine healthy")
			return nil
		}
		time.Sleep(time.Second)
	}
	return fmt.Errorf("UTMStack EDR scan engine did not become healthy")
}

func (e *Engine) Stop() error {
	if e.cmd != nil && e.cmd.Process != nil {
		return e.cmd.Process.Kill()
	}
	return nil
}
```

- [ ] **Step 6: Verify the package builds (engine.go has no unit test; supervision is VM-verified in Task 9)**

Run: `cd agent && GOOS=windows GOARCH=amd64 go build ./edr/engine/`
Expected: builds with no output.

- [ ] **Step 7: Commit**

```bash
git add agent/edr/engine/
git commit -m "feat(edr): clamd INSTREAM client and child-process engine supervision"
```

---

## Task 5: Scanner orchestration + on-demand `scan` command

**Files:**
- Create: `agent/edr/scanner/scanner.go`
- Test: `agent/edr/scanner/scanner_test.go`
- (main wiring lands in Task 6)

**Interfaces:**
- Consumes: `cache` (Task 3), `engine.ScanBytes` (Task 4), `event` + `Spool` (Task 2), `config` (Task 1).
- Produces:
  - `type Scanner struct{}`
  - `func New(cfg config.EDRConfig, c *cache.Cache, sp *event.Spool) *Scanner`
  - `func (s *Scanner) ScanFile(path, source string) (verdict string, signature string, err error)` — hashes, checks cache, scans via engine, stores verdict, and on malicious appends a branded event to the spool
  - `func SHA256File(path string) (string, error)`
  - A function field `scanBytes func(addr string, data []byte) (bool, string, error)` on `Scanner` so tests can inject a fake engine (defaults to `engine.ScanBytes`).

- [ ] **Step 1: Write the failing test (inject a fake engine; assert quarantine-less detection path emits a spool event)**

```go
// agent/edr/scanner/scanner_test.go
package scanner

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/utmstack/UTMStack/agent/edr/cache"
	"github.com/utmstack/UTMStack/agent/edr/config"
	"github.com/utmstack/UTMStack/agent/edr/event"
)

func TestScanFileMaliciousEmitsBrandedEvent(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "a.exe")
	if err := os.WriteFile(target, []byte("payload"), 0o644); err != nil {
		t.Fatal(err)
	}
	c, err := cache.Open(filepath.Join(dir, "edr.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	spoolPath := filepath.Join(dir, "events.ndjson")
	sp, err := event.OpenSpool(spoolPath, 1<<20)
	if err != nil {
		t.Fatal(err)
	}
	defer sp.Close()

	s := New(config.Default(), c, sp)
	s.scanBytes = func(addr string, data []byte) (bool, string, error) {
		return false, "Win.Test.EICAR_HDB-1", nil // pretend malicious
	}

	verdict, sig, err := s.ScanFile(target, event.SourceEngine)
	if err != nil {
		t.Fatalf("ScanFile: %v", err)
	}
	if verdict != cache.VerdictMalicious || !strings.Contains(sig, "EICAR") {
		t.Fatalf("verdict=%q sig=%q", verdict, sig)
	}
	// cache now holds the malicious verdict
	if rec, found, _ := c.Lookup(mustHash(t, target)); !found || rec.Verdict != cache.VerdictMalicious {
		t.Fatalf("verdict not cached: %+v found=%v", rec, found)
	}
	// spool holds one branded, clamav-free line
	b, _ := os.ReadFile(spoolPath)
	line := strings.TrimSpace(string(b))
	if !strings.Contains(line, `"product":"UTMStack EDR"`) {
		t.Fatalf("spool line not branded: %s", line)
	}
	if strings.Contains(strings.ToLower(line), "clam") {
		t.Fatalf("spool line leaked engine name: %s", line)
	}
}

func mustHash(t *testing.T, p string) string {
	t.Helper()
	h, err := SHA256File(p)
	if err != nil {
		t.Fatal(err)
	}
	return h
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd agent && go test ./edr/scanner/ -run TestScanFileMalicious -v`
Expected: FAIL — undefined `New`/`ScanFile`/`SHA256File`.

- [ ] **Step 3: Implement the scanner**

```go
// agent/edr/scanner/scanner.go
package scanner

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"

	"github.com/utmstack/UTMStack/agent/edr/cache"
	"github.com/utmstack/UTMStack/agent/edr/config"
	"github.com/utmstack/UTMStack/agent/edr/engine"
	"github.com/utmstack/UTMStack/agent/edr/event"
)

type Scanner struct {
	cfg       config.EDRConfig
	cache     *cache.Cache
	spool     *event.Spool
	scanBytes func(addr string, data []byte) (bool, string, error)
}

func New(cfg config.EDRConfig, c *cache.Cache, sp *event.Spool) *Scanner {
	return &Scanner{cfg: cfg, cache: c, spool: sp, scanBytes: engine.ScanBytes}
}

func SHA256File(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// ScanFile hashes the file, checks the cache, scans via the engine on a miss,
// stores the verdict, and emits a branded detection event on malicious.
func (s *Scanner) ScanFile(path, source string) (string, string, error) {
	hash, err := SHA256File(path)
	if err != nil {
		return "", "", err
	}
	if rec, found, err := s.cache.Lookup(hash); err != nil {
		return "", "", err
	} else if found && !rec.Stale {
		return rec.Verdict, rec.Signature, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return "", "", err
	}
	clean, sig, err := s.scanBytes(s.cfg.ClamdAddr, data)
	if err != nil {
		return "", "", err
	}

	verdict := cache.VerdictClean
	if !clean {
		verdict = cache.VerdictMalicious
	}
	if err := s.cache.Store(cache.VerdictRecord{SHA256: hash, Verdict: verdict, Signature: sig}); err != nil {
		return "", "", err
	}

	if verdict == cache.VerdictMalicious {
		ev := event.NewDetection(path, hash, sig, source)
		js, err := ev.ToJSON()
		if err != nil {
			return verdict, sig, err
		}
		if err := s.spool.Append(js); err != nil {
			return verdict, sig, err
		}
	}
	return verdict, sig, nil
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `cd agent && go test ./edr/scanner/ -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add agent/edr/scanner/
git commit -m "feat(edr): scanner orchestration (hash->cache->engine->verdict->event)"
```

---

## Task 6: EDR service + `main.go` (install/uninstall/run/scan/status)

**Files:**
- Create: `agent/edr/service/config.go`, `agent/edr/service/service.go`, `agent/edr/service/install.go`, `agent/edr/main.go`
- Test: none new (build + VM verification in Task 9); logic is covered by Tasks 1–5.

**Interfaces:**
- Consumes: everything above.
- Produces: the `utmstack_edr.exe` binary with subcommands `install | uninstall | run | scan <path> | status`.

- [ ] **Step 1: Service config (mirror updater)**

```go
// agent/edr/service/config.go
package service

import "github.com/kardianos/service"

func GetConfigServ() *service.Config {
	return &service.Config{
		Name:        "UTMStackEDR",
		DisplayName: "UTMStack EDR",
		Description: "UTMStack Endpoint Detection & Response",
	}
}
```

- [ ] **Step 2: Service program + run loop**

The Plan-1 run loop opens the cache + spool + engine, keeps the engine healthy, and idles until stopped. (Watchers arrive in Plan 2/3.) It writes `status.json` periodically.

```go
// agent/edr/service/service.go
package service

import (
	"context"
	"encoding/json"
	"os"
	"time"

	"github.com/kardianos/service"
	"github.com/utmstack/UTMStack/agent/edr/cache"
	"github.com/utmstack/UTMStack/agent/edr/config"
	"github.com/utmstack/UTMStack/agent/edr/engine"
	"github.com/utmstack/UTMStack/shared/logger"
)

type program struct {
	cancel context.CancelFunc
	eng    *engine.Engine
}

func (p *program) Start(s service.Service) error { go p.run(); return nil }

func (p *program) Stop(s service.Service) error {
	if p.cancel != nil {
		p.cancel()
	}
	if p.eng != nil {
		_ = p.eng.Stop()
	}
	return nil
}

type statusDoc struct {
	Running       bool   `json:"running"`
	EngineHealthy bool   `json:"engine_healthy"`
	SigDBVersion  string `json:"sigdb_version"`
	UpdatedAt     string `json:"updated_at"`
	Product       string `json:"product"`
}

func (p *program) run() {
	cfg, err := config.Load()
	if err != nil {
		logger.Error("UTMStack EDR: config load: %v", err)
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	p.cancel = cancel

	c, err := cache.Open(config.DBFile)
	if err != nil {
		logger.Error("UTMStack EDR: cache open: %v", err)
		return
	}
	defer c.Close()

	p.eng = engine.New(cfg)
	if err := p.eng.EnsureRunning(); err != nil {
		logger.Error("UTMStack EDR: %v", err)
	}

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	writeStatus(cfg)
	for {
		select {
		case <-ctx.Done():
			logger.Info("UTMStack EDR: stopping")
			return
		case <-ticker.C:
			writeStatus(cfg)
		}
	}
}

func writeStatus(cfg config.EDRConfig) {
	healthy := engine.Ping(cfg.ClamdAddr) == nil
	ver := ""
	if healthy {
		ver, _ = engine.Version(cfg.ClamdAddr)
	}
	doc := statusDoc{Running: true, EngineHealthy: healthy, SigDBVersion: ver, UpdatedAt: time.Now().UTC().Format(time.RFC3339), Product: "UTMStack EDR"}
	b, _ := json.MarshalIndent(doc, "", "  ")
	_ = os.WriteFile(config.StatusFile, b, 0o644)
}

func RunService() {
	prg := new(program)
	s, err := service.New(prg, GetConfigServ())
	if err != nil {
		logger.Fatal("UTMStack EDR: create service: %v", err)
	}
	if err := s.Run(); err != nil {
		logger.Fatal("UTMStack EDR: run service: %v", err)
	}
}
```

Note: `engine.Version` returns clamd's raw version banner; keep it in `status.json` only (a local operational file), never in an emitted event. If product policy requires hiding it entirely, replace with a signature-DB date — tracked as a follow-up, not required for Plan 1 acceptance.

- [ ] **Step 3: Install/uninstall (mirror updater)**

```go
// agent/edr/service/install.go
package service

import (
	"fmt"
	"os"

	"github.com/kardianos/service"
)

func InstallService() {
	prg := new(program)
	s, err := service.New(prg, GetConfigServ())
	if err != nil {
		fmt.Println("Error creating service:", err)
		os.Exit(1)
	}
	if err := s.Install(); err != nil {
		fmt.Println("Error installing service:", err)
		os.Exit(1)
	}
}

func UninstallService() {
	prg := new(program)
	s, err := service.New(prg, GetConfigServ())
	if err != nil {
		fmt.Println("Error creating service:", err)
		os.Exit(1)
	}
	_ = s.Stop()
	if err := s.Uninstall(); err != nil {
		fmt.Println("Error uninstalling service:", err)
		os.Exit(1)
	}
}
```

- [ ] **Step 4: `main.go` with subcommands**

```go
// agent/edr/main.go
package main

import (
	"fmt"
	"os"

	"github.com/utmstack/UTMStack/agent/edr/cache"
	"github.com/utmstack/UTMStack/agent/edr/config"
	"github.com/utmstack/UTMStack/agent/edr/event"
	"github.com/utmstack/UTMStack/agent/edr/scanner"
	"github.com/utmstack/UTMStack/agent/edr/service"
	"github.com/utmstack/UTMStack/shared/fs"
	"github.com/utmstack/UTMStack/shared/logger"
)

func main() {
	logger.Init(config.LogFile, logger.LevelInfo)

	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "install":
			service.InstallService()
			fmt.Println("UTMStackEDR service installed")
			return
		case "uninstall":
			service.UninstallService()
			fmt.Println("UTMStackEDR service uninstalled")
			return
		case "scan":
			if len(os.Args) < 3 {
				fmt.Println("usage: utmstack_edr scan <path>")
				os.Exit(1)
			}
			runScan(os.Args[2])
			return
		case "status":
			if fs.Exists(config.StatusFile) {
				lines, _ := fs.ReadLines(config.StatusFile)
				for _, l := range lines {
					fmt.Println(l)
				}
			} else {
				fmt.Println("UTMStack EDR: not running")
			}
			return
		}
	}
	service.RunService()
}

func runScan(path string) {
	cfg, err := config.Load()
	if err != nil {
		fmt.Println("config:", err)
		os.Exit(1)
	}
	c, err := cache.Open(config.DBFile)
	if err != nil {
		fmt.Println("cache:", err)
		os.Exit(1)
	}
	defer c.Close()
	sp, err := event.OpenSpool(config.SpoolFile, 8<<20)
	if err != nil {
		fmt.Println("spool:", err)
		os.Exit(1)
	}
	defer sp.Close()

	s := scanner.New(cfg, c, sp)
	verdict, sig, err := s.ScanFile(path, event.SourceEngine)
	if err != nil {
		fmt.Println("scan error:", err)
		os.Exit(1)
	}
	fmt.Printf("UTMStack EDR verdict: %s %s\n", verdict, sig)
}
```

- [ ] **Step 5: Build for Windows (both architectures)**

Run: `cd agent && for arch in amd64 arm64; do GOOS=windows GOARCH=$arch go build -o dist/utmstack_edr_windows_$arch.exe ./edr; done`
Expected: both `dist/utmstack_edr_windows_amd64.exe` and `dist/utmstack_edr_windows_arm64.exe` produced.

- [ ] **Step 6: Commit**

```bash
git add agent/edr/service/ agent/edr/main.go
git commit -m "feat(edr): SYSTEM service + main with install/uninstall/run/scan/status"
```

---

## Task 7: Agent-side EDR event relay (spool → LogQueue)

**Files:**
- Create: `agent/agent/edr_relay.go`
- Modify: `agent/serv/service.go` (launch the relay goroutine)
- Test: `agent/agent/edr_relay_test.go`

**Interfaces:**
- Consumes: the EDR spool file at `edr-spool/events.ndjson`; the existing `agent.LogQueue` (`chan *plugins.Log`).
- Produces:
  - `func EDRRelay(ctx context.Context)` — tails the spool, wrapping each line as `*plugins.Log{DataType:"utmstack_edr", Raw:<line>, DataSource:<host>}` and enqueuing on `LogQueue`, advancing a persisted byte offset in `<install>/edr-spool/.offset`.
  - `func drainSpool(path, offsetPath string, emit func(raw string)) (newOffset int64, err error)` — the pure, unit-tested core.

- [ ] **Step 1: Write the failing test for `drainSpool`**

```go
// agent/agent/edr_relay_test.go
package agent

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDrainSpoolEmitsNewLinesAndAdvances(t *testing.T) {
	dir := t.TempDir()
	spool := filepath.Join(dir, "events.ndjson")
	offset := filepath.Join(dir, ".offset")
	if err := os.WriteFile(spool, []byte("line-a\nline-b\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	var got []string
	off, err := drainSpool(spool, offset, func(raw string) { got = append(got, raw) })
	if err != nil {
		t.Fatalf("drainSpool: %v", err)
	}
	if len(got) != 2 || got[0] != "line-a" || got[1] != "line-b" {
		t.Fatalf("emitted = %v", got)
	}

	// Append one more line; a second drain must emit only the new line.
	f, _ := os.OpenFile(spool, os.O_APPEND|os.O_WRONLY, 0o600)
	f.WriteString("line-c\n")
	f.Close()

	got = nil
	if _, err := drainSpool(spool, offset, func(raw string) { got = append(got, raw) }); err != nil {
		t.Fatalf("drainSpool 2: %v", err)
	}
	if len(got) != 1 || got[0] != "line-c" {
		t.Fatalf("second drain emitted = %v (offset start %d)", got, off)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd agent && go test ./agent/ -run TestDrainSpool -v`
Expected: FAIL — undefined `drainSpool`.

- [ ] **Step 3: Implement the relay**

```go
// agent/agent/edr_relay.go
package agent

import (
	"bufio"
	"context"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/threatwinds/go-sdk/plugins"
	"github.com/utmstack/UTMStack/agent/utils"
	"github.com/utmstack/UTMStack/shared/fs"
)

var (
	edrSpoolFile  = filepath.Join(fs.GetExecutablePath(), "edr-spool", "events.ndjson")
	edrOffsetFile = filepath.Join(fs.GetExecutablePath(), "edr-spool", ".offset")
	edrDataType   = "utmstack_edr"
)

// drainSpool reads unread lines from spool (starting at the persisted byte
// offset), calls emit for each, and persists the advanced offset.
func drainSpool(spoolPath, offsetPath string, emit func(raw string)) (int64, error) {
	start := readOffset(offsetPath)

	f, err := os.Open(spoolPath)
	if err != nil {
		if os.IsNotExist(err) {
			return start, nil
		}
		return start, err
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		return start, err
	}
	if fi.Size() < start {
		// spool rotated/truncated; restart from the beginning
		start = 0
	}
	if _, err := f.Seek(start, 0); err != nil {
		return start, err
	}

	r := bufio.NewReader(f)
	pos := start
	for {
		line, err := r.ReadString('\n')
		if len(line) > 0 && strings.HasSuffix(line, "\n") {
			pos += int64(len(line))
			trimmed := strings.TrimRight(line, "\n")
			if trimmed != "" {
				emit(trimmed)
			}
		}
		if err != nil {
			break // EOF or read error: stop; partial (no newline) line stays unread
		}
	}
	writeOffset(offsetPath, pos)
	return pos, nil
}

func readOffset(p string) int64 {
	b, err := os.ReadFile(p)
	if err != nil {
		return 0
	}
	n, _ := strconv.ParseInt(strings.TrimSpace(string(b)), 10, 64)
	return n
}

func writeOffset(p string, n int64) {
	_ = os.MkdirAll(filepath.Dir(p), 0o755)
	_ = os.WriteFile(p, []byte(strconv.FormatInt(n, 10)), 0o644)
}

// EDRRelay drains the EDR spool into the agent LogQueue every second until ctx is done.
func EDRRelay(ctx context.Context) {
	host, _ := os.Hostname()
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			_, err := drainSpool(edrSpoolFile, edrOffsetFile, func(raw string) {
				select {
				case LogQueue <- &plugins.Log{DataType: edrDataType, Raw: raw, DataSource: host}:
				default:
					utils.Logger.LogF(100, "EDR relay: LogQueue full, retrying next tick")
				}
			})
			if err != nil {
				utils.Logger.LogF(100, "EDR relay drain error: %v", err)
			}
		}
	}
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `cd agent && go test ./agent/ -run TestDrainSpool -v`
Expected: PASS.

- [ ] **Step 5: Launch the relay from the agent run loop**

Modify `agent/serv/service.go` — in `run()`, alongside the other `p.goSafe(...)` calls (right before `startOSCollectors(ctx, pb.LogQueue)`), add:

```go
	p.goSafe("EDRRelay", func() {
		pb.EDRRelay(ctx)
	})
```

- [ ] **Step 6: Build the agent to verify wiring**

Run: `cd agent && GOOS=windows GOARCH=amd64 go build ./...`
Expected: builds with no output.

- [ ] **Step 7: Commit**

```bash
git add agent/agent/edr_relay.go agent/serv/service.go
git commit -m "feat(agent): relay EDR spool events into LogQueue"
```

---

## Task 8: Agent CLI verbs — enable / disable / status

**Files:**
- Create: `agent/cmd/enable_edr.go`, `agent/cmd/disable_edr.go`, `agent/cmd/edr_status.go`
- Test: none (thin wrappers over `shared/svc`; verified on VM in Task 9)

**Interfaces:**
- Consumes: `shared/svc` (`Start`/`Stop`/`Status`), `edrconfig.ServiceName` = `"UTMStackEDR"`, `edrconfig.{ConfigFile,StatusFile}`.
- Produces: `utmstack_agent enable-edr | disable-edr | edr-status`.

- [ ] **Step 1: `enable-edr`**

`enable-edr` writes `enabled=true` + the agent's server/skip-cert into `edr.json` (so the EDR can later fetch signatures), then starts the service.

```go
// agent/cmd/enable_edr.go
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	edrconfig "github.com/utmstack/UTMStack/agent/edr/config"
	"github.com/utmstack/UTMStack/agent/config"
	"github.com/utmstack/UTMStack/shared/svc"
)

var enableEDRCmd = &cobra.Command{
	Use:     "enable-edr",
	Short:   "Enable and start the UTMStack EDR module",
	PreRunE: requireInstalled,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := edrconfig.Load()
		if err != nil {
			return err
		}
		if agentCnf, err := config.GetCurrentConfig(); err == nil {
			c.Server = agentCnf.Server
			c.SkipCertValidate = agentCnf.SkipCertValidation
		}
		c.Enabled = true
		if err := edrconfig.Save(c); err != nil {
			return err
		}
		if err := svc.Start(edrconfig.ServiceName); err != nil {
			return fmt.Errorf("starting EDR service: %w", err)
		}
		fmt.Println("UTMStack EDR enabled")
		return nil
	},
}

func init() { rootCmd.AddCommand(enableEDRCmd) }
```

- [ ] **Step 2: `disable-edr`**

```go
// agent/cmd/disable_edr.go
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	edrconfig "github.com/utmstack/UTMStack/agent/edr/config"
	"github.com/utmstack/UTMStack/shared/svc"
)

var disableEDRCmd = &cobra.Command{
	Use:     "disable-edr",
	Short:   "Disable and stop the UTMStack EDR module",
	PreRunE: requireInstalled,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := svc.Stop(edrconfig.ServiceName); err != nil {
			return fmt.Errorf("stopping EDR service: %w", err)
		}
		c, err := edrconfig.Load()
		if err == nil {
			c.Enabled = false
			_ = edrconfig.Save(c)
		}
		fmt.Println("UTMStack EDR disabled")
		return nil
	},
}

func init() { rootCmd.AddCommand(disableEDRCmd) }
```

- [ ] **Step 3: `edr-status`**

```go
// agent/cmd/edr_status.go
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	edrconfig "github.com/utmstack/UTMStack/agent/edr/config"
	"github.com/utmstack/UTMStack/shared/fs"
	"github.com/utmstack/UTMStack/shared/svc"
)

var edrStatusCmd = &cobra.Command{
	Use:     "edr-status",
	Short:   "Show UTMStack EDR module status",
	PreRunE: requireInstalled,
	RunE: func(cmd *cobra.Command, args []string) error {
		st, _ := svc.Status(edrconfig.ServiceName)
		fmt.Printf("Service: %s\n", st)
		if fs.Exists(edrconfig.StatusFile) {
			lines, _ := fs.ReadLines(edrconfig.StatusFile)
			for _, l := range lines {
				fmt.Println(l)
			}
		}
		return nil
	},
}

func init() { rootCmd.AddCommand(edrStatusCmd) }
```

- [ ] **Step 4: Build**

Run: `cd agent && GOOS=windows GOARCH=amd64 go build ./...`
Expected: builds with no output.

- [ ] **Step 5: Commit**

```bash
git add agent/cmd/enable_edr.go agent/cmd/disable_edr.go agent/cmd/edr_status.go
git commit -m "feat(agent): enable-edr / disable-edr / edr-status verbs"
```

---

## Task 9: Ship the EDR binary via the dependency mechanism + end-to-end VM acceptance

**Files:**
- Create: `agent/dependency/edr.go` (helper + hooks)
- Modify: `agent/dependency/deps_windows_amd64.go` (add the dependency entry)
- Test: end-to-end on the Windows VM.

**Interfaces:**
- Consumes: the `Dependency` struct, `config.DependUrl`, `config.DependenciesPort`, `shared/exec`.
- Produces: `func EDRFile(suffix string) string`, hooks `configureEDR`, `uninstallEDR`, and a new entry in `GetDependencies()`.

- [ ] **Step 1: Dependency helper + hooks**

```go
// agent/dependency/edr.go
package dependency

import (
	"fmt"
	"path/filepath"
	"runtime"

	"github.com/utmstack/UTMStack/agent/config"
	"github.com/utmstack/UTMStack/shared/exec"
	"github.com/utmstack/UTMStack/shared/fs"
)

const EDRVersion = "12.0.0"

// EDRFile returns the EDR binary name with OS/arch suffix, matching UpdaterFile.
func EDRFile(suffix string) string {
	name := fmt.Sprintf("utmstack_edr_%s_%s%s", runtime.GOOS, runtime.GOARCH, suffix)
	if runtime.GOOS == "windows" {
		return name + ".exe"
	}
	return name
}

func configureEDR() error {
	edrPath := filepath.Join(fs.GetExecutablePath(), EDRFile(""))
	return exec.Run(edrPath, fs.GetExecutablePath(), "install")
}

func uninstallEDR() error {
	edrPath := filepath.Join(fs.GetExecutablePath(), EDRFile(""))
	if !fs.Exists(edrPath) {
		return nil
	}
	return exec.Run(edrPath, fs.GetExecutablePath(), "uninstall")
}

func edrDownloadURL(server string) string {
	return fmt.Sprintf(config.DependUrl, server, config.DependenciesPort, EDRFile(""))
}
```

- [ ] **Step 2: Register the dependency (BOTH Windows architectures)**

The `dependency` package is split per-arch, so add the identical `edr` entry to the slice returned by `GetDependencies()` in **both** `agent/dependency/deps_windows_amd64.go` **and** `agent/dependency/deps_windows_arm64.go` (after the `updater` entry). The `EDRFile`/`configureEDR`/`uninstallEDR`/`edrDownloadURL` helpers live in the untagged `agent/dependency/edr.go`, so they are shared — only the slice entry is duplicated:

```go
		{
			Name:        "edr",
			Version:     EDRVersion,
			BinaryPath:  filepath.Join(basePath, EDRFile("")),
			DownloadURL: edrDownloadURL,
			Critical:    false, // agent runs even if EDR is unavailable
			Configure:   configureEDR,
			Uninstall:   uninstallEDR,
		},
```

- [ ] **Step 3: Build the binaries for the VM (match the VM's architecture)**

Build for both supported architectures; deploy the one matching the VM (amd64 for x64 Windows, arm64 for Windows-on-ARM):
```bash
cd agent
for arch in amd64 arm64; do
  GOOS=windows GOARCH=$arch go build -o dist/utmstack_edr_windows_$arch.exe ./edr
  GOOS=windows GOARCH=$arch go build -o dist/utmstack_agent_windows_$arch.exe .
done
```
Expected: `utmstack_edr_windows_{amd64,arm64}.exe` and `utmstack_agent_windows_{amd64,arm64}.exe` in `agent/dist/`.

- [ ] **Step 4: Prepare the VM — install ClamAV (test provisioning only)**

Since the server does not yet host clamd (external dependency §12.1), install ClamAV on the VM manually for testing and generate a `clamd.conf` that binds a **TCP** socket on `127.0.0.1:3310`. Point the EDR at it via `edr.json` `clamd_path`, or run the ClamAV Windows service directly. Suggested VM PowerShell (run as Administrator):

```powershell
# Download and unzip the official ClamAV Windows build to C:\clamav (manual, one-time)
# Then create C:\clamav\clamd.conf with at least:
#   TCPSocket 3310
#   TCPAddr 127.0.0.1
#   DatabaseDirectory C:\clamav\database
# Fetch signatures once:
C:\clamav\freshclam.exe --config-file C:\clamav\freshclam.conf
# Start the engine:
Start-Process C:\clamav\clamd.exe -ArgumentList '--config-file C:\clamav\clamd.conf'
```

Then set the EDR to use it (create `<install>\edr.json`): `{ "clamd_addr": "127.0.0.1:3310", "clamd_path": "C:\\clamav\\clamd.exe" }`.

Expected: `Test-NetConnection 127.0.0.1 -Port 3310` succeeds.

- [ ] **Step 5: Copy binaries to the VM and install the EDR service**

Copy `dist\utmstack_edr_windows_amd64.exe` into the agent install directory as `utmstack_edr_windows_amd64.exe`, then (Administrator):

```powershell
.\utmstack_edr_windows_amd64.exe install
sc query UTMStackEDR
```
Expected: `sc query` shows the `UTMStackEDR` service `RUNNING`.

- [ ] **Step 6: Acceptance — on-demand EICAR scan produces a branded, clamav-free event**

Create the EICAR test file and scan it:

```powershell
$eicar = 'X5O!P%@AP[4\PZX54(P^)7CC)7}$EICAR-STANDARD-ANTIVIRUS-TEST-FILE!$H+H*'
Set-Content -Path C:\Users\atlas\Downloads\eicar.com -Value $eicar -NoNewline
.\utmstack_edr_windows_amd64.exe scan C:\Users\atlas\Downloads\eicar.com
```
Expected stdout: `UTMStack EDR verdict: malicious <signature>` (signature contains `EICAR`).

Verify the spool line is branded and leaks no engine name:

```powershell
Get-Content .\edr-spool\events.ndjson
Select-String -Path .\edr-spool\events.ndjson -Pattern 'clam' -SimpleMatch
```
Expected: the JSON line contains `"product":"UTMStack EDR"` and `"engine":"UTMStack EDR"`; the `Select-String` for `clam` returns **nothing**.

- [ ] **Step 7: Acceptance — orchestration verbs**

```powershell
.\utmstack_agent.exe disable-edr
sc query UTMStackEDR        # STOPPED
.\utmstack_agent.exe enable-edr
sc query UTMStackEDR        # RUNNING
.\utmstack_agent.exe edr-status
```
Expected: service transitions STOPPED→RUNNING; `edr-status` prints `Service: running` plus the `status.json` body with `"product": "UTMStack EDR"`.

- [ ] **Step 8: Acceptance — event reaches the platform**

With the agent running and `enable-edr` active, re-run the EICAR scan (Step 6). In the UTMStack platform, confirm a log/event of type `utmstack_edr` arrives, branded UTMStack EDR, containing the detection — and that it contains no `clamav`/`clamd` string. (Verify the final output in the platform UI/index, not just agent logs.)

Expected: the detection event is visible in the platform, correctly branded.

- [ ] **Step 9: Commit**

```bash
git add agent/dependency/edr.go agent/dependency/deps_windows_amd64.go
git commit -m "feat(agent): ship utmstack_edr via dependency mechanism"
```

---

## Self-Review (completed against the spec)

**Spec coverage (Plan 1's slice):**
- Engine host (clamd child + INSTREAM + health) → Task 4, 6. *(Signature auto-update/freshclam scheduling is deferred to Plan 2's config/feed work; Plan 1 provisions signatures manually on the VM — called out in Task 9 Step 4.)*
- Verdict cache (own SQLite) → Task 3. ✔ storage isolation honored.
- Event emitter + agent relay (branded, spool→LogQueue) → Task 2, 7. ✔
- Agent orchestration (enable/disable/status + dependency shipping) → Task 8, 9. ✔
- Labeling rule (no `clamav`/`clamd` in events/logs) → enforced by tests in Task 2 & 5 and the VM grep in Task 9 Step 6. ✔
- **Deferred (correctly, to later plans):** USN watcher, process watcher, responder/kill, AMSI, behavioral forwarder, quarantine, freshclam scheduling. These are Plans 2–4.

**Placeholder scan:** No TBD/TODO/"add error handling" left; every code step contains complete code. The two non-`go test` verification points (engine supervision, service install) are Windows-only and are given concrete VM commands with expected output, which is the correct substitute for a unit test cycle on those.

**Type consistency:** `ScanBytes(addr string, data []byte) (bool, string, error)` is defined in Task 4 and consumed identically in Task 5 (`scanBytes` field default). `event.Source*`/`Action*` constants defined in Task 2 are used in Task 5/6. `cache.VerdictRecord` fields defined in Task 3 are used unchanged in Task 5. `edrconfig.ServiceName = "UTMStackEDR"` matches the kardianos `service.Config.Name` in Task 6 and the `svc.*` calls in Task 8. Spool path (`edr-spool/events.ndjson`) is identical in `config` (Task 1), the EDR `main` (Task 6), and the agent relay (Task 7).

**Gaps intentionally carried forward:** production signature delivery (server hosting clamd, §12.1) and freshclam scheduling move to Plan 2; Plan 1 stays testable via manual VM provisioning.
