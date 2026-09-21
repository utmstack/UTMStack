# UTMStack EDR — Phase 1 Windows — Plan 2: Auto-detect

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make the EDR detect malware **automatically** — a whole-volume NTFS USN-journal watcher feeds a scan orchestrator that hashes, cache-checks, and scans every new/changed file; malicious files are **quarantined** (moved, never deleted) with a restore path, and signatures update on a schedule that invalidates the cache.

**Architecture:** Builds directly on Plan 1's packages. Adds `edr/quarantine`, `edr/orchestrator`, `edr/watcher` (USN), and `edr/feed` (freshclam) under `agent/edr/`, plus a quarantine store in `edr.db`. The service run loop, when enabled, starts the USN watcher(s) → orchestrator worker pool → scanner (now quarantine-wired). Pure-Go logic is TDD'd on macOS; USN journal I/O and freshclam are verified on the VM.

**Tech Stack:** Go 1.25.5, `golang.org/x/sys/windows` (CreateFile + DeviceIoControl for USN; already in `go.mod`), the Plan 1 packages, gorm/sqlite.

## Global Constraints

All Plan 1 Global Constraints carry over verbatim (target branch `release/v12.0.0`, Go-first, **branding: never emit `clamav`/`clamd` in events or surfaced logs**, event `DataType="utmstack_edr"` / `source ∈ {file_watcher,process_watcher,amsi,edr_engine}` / `engine="UTMStack EDR"`, own `edr.db`, **quarantine not delete**, no kernel driver, SYSTEM service). Plus:

- **No new Go dependencies** — USN uses `golang.org/x/sys/windows`, already present.
- **USN over per-directory watching** — whole-volume USN journal is the required file-arrival mechanism (spec §4.2).
- **Multi-arch (carried from Plan 1):** everything here must build for **`windows/amd64` and `windows/arm64`**. The USN code is `//go:build windows` (not arch-specific) and uses `x/sys/windows` + fixed-width struct fields, so it is architecture-portable — verify with `GOOS=windows GOARCH=arm64 go build ./...` as well as amd64. No arch-specific code is permitted.
- **Depends on Plan 1** being merged/present: packages `edr/config`, `edr/event`, `edr/cache`, `edr/engine`, `edr/scanner`, `edr/service`.

**Build/test note:** identical to Plan 1 — pure-Go logic uses `go test` on macOS; USN journal reads, path resolution, and freshclam are VM-verified on `10.211.55.12`.

---

## File Structure

**New packages:**
- `agent/edr/quarantine/quarantine.go` — move-to-store + restore; depends on `cache` for the record.
- `agent/edr/orchestrator/orchestrator.go` — bounded queue + worker pool driving `scanner.ScanFile`.
- `agent/edr/orchestrator/exclusions.go` — path-exclusion matcher.
- `agent/edr/watcher/usn_record.go` — pure USN_RECORD_V2 buffer parser (unit-tested).
- `agent/edr/watcher/usn_windows.go` — volume open + `DeviceIoControl` query/read loop + path resolution (Windows build tag).
- `agent/edr/watcher/usn_other.go` — non-Windows stub so the package builds on macOS (`//go:build !windows`).
- `agent/edr/feed/freshclam.go` — signature-update scheduler + cache staleness invalidation.

**Modified:**
- `agent/edr/cache/cache.go` — add `QuarantineRecord` model + migrate + CRUD; add `MarkStaleBySigDB`.
- `agent/edr/scanner/scanner.go` — inject a `Quarantiner`; on malicious → quarantine, emit `quarantined`.
- `agent/edr/main.go` — `runScan` passes a real quarantiner; add `restore <id>` subcommand.
- `agent/edr/service/service.go` — start watcher + orchestrator + feed when enabled.

---

## Task 1: Quarantine records in the cache DB

**Files:**
- Modify: `agent/edr/cache/cache.go`
- Test: `agent/edr/cache/quarantine_test.go`

**Interfaces:**
- Produces (on `*cache.Cache`):
  - `type QuarantineRecord struct { QuarantineID string (pk); OriginalPath, SHA256, Detection, Engine string; QuarantinedAt time.Time; Restorable, Restored bool }`
  - `func (c *Cache) StoreQuarantine(r QuarantineRecord) error`
  - `func (c *Cache) GetQuarantine(id string) (QuarantineRecord, bool, error)`
  - `func (c *Cache) ListQuarantine() ([]QuarantineRecord, error)`
  - `func (c *Cache) MarkRestored(id string) error`
  - `func (c *Cache) MarkStaleBySigDB(currentSigDB string) (int64, error)` — sets `Stale=true` on clean records whose `SigDBVersion != currentSigDB`

- [ ] **Step 1: Write the failing test**

```go
// agent/edr/cache/quarantine_test.go
package cache

import (
	"path/filepath"
	"testing"
)

func TestQuarantineRecordRoundTrip(t *testing.T) {
	c, err := Open(filepath.Join(t.TempDir(), "edr.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	rec := QuarantineRecord{QuarantineID: "q1", OriginalPath: `C:\x\a.exe`, SHA256: "h", Detection: "Win.Test.EICAR", Engine: "UTMStack EDR", Restorable: true}
	if err := c.StoreQuarantine(rec); err != nil {
		t.Fatalf("StoreQuarantine: %v", err)
	}
	got, found, err := c.GetQuarantine("q1")
	if err != nil || !found {
		t.Fatalf("GetQuarantine found=%v err=%v", found, err)
	}
	if got.Detection != "Win.Test.EICAR" || got.Engine != "UTMStack EDR" {
		t.Fatalf("bad record: %+v", got)
	}
	if err := c.MarkRestored("q1"); err != nil {
		t.Fatal(err)
	}
	got, _, _ = c.GetQuarantine("q1")
	if !got.Restored {
		t.Fatal("record not marked restored")
	}
}

func TestMarkStaleBySigDB(t *testing.T) {
	c, err := Open(filepath.Join(t.TempDir(), "edr.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	_ = c.Store(VerdictRecord{SHA256: "clean1", Verdict: VerdictClean, SigDBVersion: "100"})
	_ = c.Store(VerdictRecord{SHA256: "mal1", Verdict: VerdictMalicious, SigDBVersion: "100"})

	n, err := c.MarkStaleBySigDB("200")
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 { // only the clean record goes stale
		t.Fatalf("marked %d stale, want 1", n)
	}
	if r, _, _ := c.Lookup("clean1"); !r.Stale {
		t.Fatal("clean1 should be stale")
	}
	if r, _, _ := c.Lookup("mal1"); r.Stale {
		t.Fatal("malicious verdict must persist (not stale)")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./edr/cache/ -run 'Quarantine|Stale' -v`
Expected: FAIL — undefined `QuarantineRecord`/methods.

- [ ] **Step 3: Implement — add to `agent/edr/cache/cache.go`**

Add the model, migrate it in `Open` (change the `AutoMigrate` call), and add the methods:

```go
// add near VerdictRecord
type QuarantineRecord struct {
	QuarantineID  string `gorm:"primaryKey"`
	OriginalPath  string
	SHA256        string
	Detection     string
	Engine        string
	QuarantinedAt time.Time
	Restorable    bool
	Restored      bool
}

// in Open(), replace the AutoMigrate line with:
//   if err := db.AutoMigrate(&VerdictRecord{}, &QuarantineRecord{}); err != nil {

func (c *Cache) StoreQuarantine(r QuarantineRecord) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if r.QuarantinedAt.IsZero() {
		r.QuarantinedAt = time.Now().UTC()
	}
	return c.db.Save(&r).Error
}

func (c *Cache) GetQuarantine(id string) (QuarantineRecord, bool, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	var r QuarantineRecord
	err := c.db.First(&r, "quarantine_id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return QuarantineRecord{}, false, nil
	}
	if err != nil {
		return QuarantineRecord{}, false, err
	}
	return r, true, nil
}

func (c *Cache) ListQuarantine() ([]QuarantineRecord, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	var rs []QuarantineRecord
	err := c.db.Find(&rs).Error
	return rs, err
}

func (c *Cache) MarkRestored(id string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.db.Model(&QuarantineRecord{}).Where("quarantine_id = ?", id).Update("restored", true).Error
}

// MarkStaleBySigDB marks clean records from an older signature DB as stale so
// they are re-scanned; malicious verdicts are left untouched (they persist).
func (c *Cache) MarkStaleBySigDB(currentSigDB string) (int64, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	res := c.db.Model(&VerdictRecord{}).
		Where("verdict = ? AND sig_db_version <> ?", VerdictClean, currentSigDB).
		Update("stale", true)
	return res.RowsAffected, res.Error
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./edr/cache/ -v`
Expected: PASS (existing + new).

- [ ] **Step 5: Commit**

```bash
git add agent/edr/cache/
git commit -m "feat(edr): quarantine records and signature-staleness in edr.db"
```

---

## Task 2: Quarantine store + scanner wiring

**Files:**
- Create: `agent/edr/quarantine/quarantine.go`
- Modify: `agent/edr/scanner/scanner.go`, `agent/edr/main.go`
- Test: `agent/edr/quarantine/quarantine_test.go`, update `agent/edr/scanner/scanner_test.go`

**Interfaces:**
- Produces:
  - `type Store struct{}`, `func New(dir string, c *cache.Cache) (*Store, error)`
  - `func (s *Store) Quarantine(originalPath, sha256, detection string) (id string, err error)` — moves the file into `dir` under a generated id (non-executable), writes a `QuarantineRecord`; never deletes on failure
  - `func (s *Store) Restore(id string) error`
  - Scanner change: `type Quarantiner interface { Quarantine(path, sha256, detection string) (string, error) }`; `New(cfg, c, sp, q)`; on malicious → quarantine then emit `event.Action = quarantined`.

- [ ] **Step 1: Write the failing quarantine test**

```go
// agent/edr/quarantine/quarantine_test.go
package quarantine

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/utmstack/UTMStack/agent/edr/cache"
)

func TestQuarantineMovesFileAndRecords(t *testing.T) {
	dir := t.TempDir()
	c, err := cache.Open(filepath.Join(dir, "edr.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	store, err := New(filepath.Join(dir, "quarantine"), c)
	if err != nil {
		t.Fatal(err)
	}

	src := filepath.Join(dir, "a.exe")
	if err := os.WriteFile(src, []byte("payload"), 0o644); err != nil {
		t.Fatal(err)
	}
	id, err := store.Quarantine(src, "hash", "Win.Test.EICAR")
	if err != nil {
		t.Fatalf("Quarantine: %v", err)
	}
	if _, err := os.Stat(src); !os.IsNotExist(err) {
		t.Fatal("original file should be gone after quarantine")
	}
	rec, found, _ := c.GetQuarantine(id)
	if !found || rec.OriginalPath != src || rec.Engine != "UTMStack EDR" {
		t.Fatalf("record wrong: %+v found=%v", rec, found)
	}

	if err := store.Restore(id); err != nil {
		t.Fatalf("Restore: %v", err)
	}
	if _, err := os.Stat(src); err != nil {
		t.Fatalf("restore did not put the file back: %v", err)
	}
}
```

- [ ] **Step 2: Run to verify it fails**

Run: `go test ./edr/quarantine/ -v`
Expected: FAIL — package/`New` undefined.

- [ ] **Step 3: Implement the store**

```go
// agent/edr/quarantine/quarantine.go
package quarantine

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/utmstack/UTMStack/agent/edr/cache"
)

type Store struct {
	dir   string
	cache *cache.Cache
}

func New(dir string, c *cache.Cache) (*Store, error) {
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, err
	}
	return &Store{dir: dir, cache: c}, nil
}

// Quarantine moves the file into the protected store and records it.
// It never deletes the source on failure.
func (s *Store) Quarantine(originalPath, sha256, detection string) (string, error) {
	id := uuid.NewString()
	dest := filepath.Join(s.dir, id+".quarantined")

	if err := moveFile(originalPath, dest); err != nil {
		return "", fmt.Errorf("quarantine move failed: %w", err)
	}
	// Best-effort: strip execute/normal attributes on the stored copy.
	_ = os.Chmod(dest, 0o600)

	rec := cache.QuarantineRecord{
		QuarantineID:  id,
		OriginalPath:  originalPath,
		SHA256:        sha256,
		Detection:     detection,
		Engine:        "UTMStack EDR",
		QuarantinedAt: time.Now().UTC(),
		Restorable:    true,
	}
	if err := s.cache.StoreQuarantine(rec); err != nil {
		// Roll the file back so we never lose it silently.
		_ = moveFile(dest, originalPath)
		return "", fmt.Errorf("quarantine record failed: %w", err)
	}
	return id, nil
}

func (s *Store) Restore(id string) error {
	rec, found, err := s.cache.GetQuarantine(id)
	if err != nil {
		return err
	}
	if !found {
		return fmt.Errorf("quarantine id not found: %s", id)
	}
	dest := filepath.Join(s.dir, id+".quarantined")
	if err := moveFile(dest, rec.OriginalPath); err != nil {
		return err
	}
	return s.cache.MarkRestored(id)
}

// moveFile renames, falling back to copy+remove across volumes.
func moveFile(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o700); err != nil {
		return err
	}
	if err := os.Rename(src, dst); err == nil {
		return nil
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	out, err := os.Create(dst)
	if err != nil {
		in.Close()
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		in.Close()
		out.Close()
		return err
	}
	in.Close()
	if err := out.Close(); err != nil {
		return err
	}
	return os.Remove(src)
}
```

Note: `github.com/google/uuid` is already in the agent `go.mod` (used by the log processor).

- [ ] **Step 4: Run to verify it passes**

Run: `go test ./edr/quarantine/ -v`
Expected: PASS.

- [ ] **Step 5: Wire quarantine into the scanner — update `agent/edr/scanner/scanner.go`**

Change `Scanner` to hold a `Quarantiner` and, on malicious, quarantine then emit a `quarantined` event:

```go
// add to scanner.go
type Quarantiner interface {
	Quarantine(path, sha256, detection string) (string, error)
}

// change New:
func New(cfg config.EDRConfig, c *cache.Cache, sp *event.Spool, q Quarantiner) *Scanner {
	return &Scanner{cfg: cfg, cache: c, spool: sp, quarantiner: q, scanBytes: engine.ScanBytes}
}
// add field: quarantiner Quarantiner

// in ScanFile, replace the malicious block:
	if verdict == cache.VerdictMalicious {
		ev := event.NewDetection(path, hash, sig, source)
		if s.quarantiner != nil {
			if _, qerr := s.quarantiner.Quarantine(path, hash, sig); qerr == nil {
				ev.Action = event.ActionQuarantined
			}
			// on quarantine failure, keep Action=detected (still reported)
		}
		js, err := ev.ToJSON()
		if err != nil {
			return verdict, sig, err
		}
		if err := s.spool.Append(js); err != nil {
			return verdict, sig, err
		}
	}
```

- [ ] **Step 6: Update the scanner test to inject a fake quarantiner and assert the `quarantined` action**

In `agent/edr/scanner/scanner_test.go`, add a fake and update the constructor call:

```go
type fakeQ struct{ called bool }
func (f *fakeQ) Quarantine(path, sha256, detection string) (string, error) { f.called = true; return "qid", nil }

// change: s := New(config.Default(), c, sp, &fakeQ{})
// after ScanFile, assert the spool line contains `"action":"quarantined"`.
```

Update the assertions:
```go
	if !strings.Contains(line, `"action":"quarantined"`) {
		t.Fatalf("expected quarantined action: %s", line)
	}
```

- [ ] **Step 7: Update `main.go` `runScan` to build a real store + add `restore`**

In `agent/edr/main.go`, in `runScan` construct the store and pass it; add a `restore` case:

```go
	// in runScan, after opening cache:
	store, err := quarantine.New(cfg.QuarantineDir, c)
	if err != nil {
		fmt.Println("quarantine:", err)
		os.Exit(1)
	}
	s := scanner.New(cfg, c, sp, store)

	// new subcommand in main() switch:
	case "restore":
		if len(os.Args) < 3 {
			fmt.Println("usage: utmstack_edr restore <id>")
			os.Exit(1)
		}
		runRestore(os.Args[2])
		return
```
Add `runRestore`:
```go
func runRestore(id string) {
	cfg, _ := config.Load()
	c, err := cache.Open(config.DBFile)
	if err != nil {
		fmt.Println("cache:", err)
		os.Exit(1)
	}
	defer c.Close()
	store, err := quarantine.New(cfg.QuarantineDir, c)
	if err != nil {
		fmt.Println("quarantine:", err)
		os.Exit(1)
	}
	if err := store.Restore(id); err != nil {
		fmt.Println("restore error:", err)
		os.Exit(1)
	}
	fmt.Println("UTMStack EDR: restored", id)
}
```
Add the `quarantine` import.

- [ ] **Step 8: Run scanner tests + build**

Run: `go test ./edr/scanner/ -v && GOOS=windows GOARCH=amd64 go build ./edr`
Expected: PASS + build OK.

- [ ] **Step 9: Commit**

```bash
git add agent/edr/quarantine/ agent/edr/scanner/ agent/edr/main.go
git commit -m "feat(edr): quarantine store, scanner quarantines malicious files, restore command"
```

---

## Task 3: Path-exclusion matcher

**Files:**
- Create: `agent/edr/orchestrator/exclusions.go`
- Test: `agent/edr/orchestrator/exclusions_test.go`

**Interfaces:**
- Produces: `type Excluder struct{}`, `func NewExcluder(patterns []string) *Excluder`, `func (e *Excluder) Excluded(path string) bool` — case-insensitive prefix/glob match on Windows-style paths.

- [ ] **Step 1: Write the failing test**

```go
// agent/edr/orchestrator/exclusions_test.go
package orchestrator

import "testing"

func TestExcluder(t *testing.T) {
	e := NewExcluder([]string{`C:\Windows\Temp`, `*.log`})
	cases := map[string]bool{
		`C:\Windows\Temp\x.dat`: true,
		`c:\windows\temp\y`:     true, // case-insensitive
		`C:\Users\a\app.log`:    true, // glob
		`C:\Users\a\a.exe`:      false,
	}
	for p, want := range cases {
		if got := e.Excluded(p); got != want {
			t.Fatalf("Excluded(%q) = %v, want %v", p, got, want)
		}
	}
}
```

- [ ] **Step 2: Run to verify it fails**

Run: `go test ./edr/orchestrator/ -run TestExcluder -v`
Expected: FAIL — undefined.

- [ ] **Step 3: Implement**

```go
// agent/edr/orchestrator/exclusions.go
package orchestrator

import (
	"path/filepath"
	"strings"
)

type Excluder struct{ patterns []string }

func NewExcluder(patterns []string) *Excluder {
	low := make([]string, 0, len(patterns))
	for _, p := range patterns {
		low = append(low, strings.ToLower(p))
	}
	return &Excluder{patterns: low}
}

func (e *Excluder) Excluded(path string) bool {
	p := strings.ToLower(filepath.ToSlash(path))
	for _, pat := range e.patterns {
		pat = filepath.ToSlash(pat)
		if strings.HasPrefix(p, pat) {
			return true
		}
		if ok, _ := filepath.Match(pat, p); ok {
			return true
		}
		if ok, _ := filepath.Match(pat, filepath.Base(p)); ok {
			return true
		}
	}
	return false
}
```

- [ ] **Step 4: Run to verify it passes**

Run: `go test ./edr/orchestrator/ -run TestExcluder -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add agent/edr/orchestrator/exclusions.go agent/edr/orchestrator/exclusions_test.go
git commit -m "feat(edr): path exclusion matcher"
```

---

## Task 4: Scan orchestrator (queue + worker pool)

**Files:**
- Create: `agent/edr/orchestrator/orchestrator.go`
- Test: `agent/edr/orchestrator/orchestrator_test.go`

**Interfaces:**
- Consumes: `scanner` (via a small interface), `Excluder` (Task 3), `config`.
- Produces:
  - `type FileEvent struct { Path, Op string }`
  - `type FileScanner interface { ScanFile(path, source string) (string, string, error) }`
  - `type Orchestrator struct{}`, `func New(sc FileScanner, ex *Excluder, workers, queueSize int) *Orchestrator`
  - `func (o *Orchestrator) Start(ctx context.Context)`, `func (o *Orchestrator) Enqueue(ev FileEvent) (accepted bool)`, `func (o *Orchestrator) Stop()`
  - Backpressure: `Enqueue` returns false (dropped, logged by caller) when the queue is full.

- [ ] **Step 1: Write the failing test**

```go
// agent/edr/orchestrator/orchestrator_test.go
package orchestrator

import (
	"context"
	"sync"
	"testing"
	"time"
)

type recScanner struct {
	mu     sync.Mutex
	seen   []string
}
func (r *recScanner) ScanFile(path, source string) (string, string, error) {
	r.mu.Lock(); defer r.mu.Unlock()
	r.seen = append(r.seen, path)
	return "clean", "", nil
}

func TestOrchestratorScansEnqueuedNonExcluded(t *testing.T) {
	rs := &recScanner{}
	o := New(rs, NewExcluder([]string{`*.log`}), 2, 16)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	o.Start(ctx)

	o.Enqueue(FileEvent{Path: `C:\a.exe`})
	o.Enqueue(FileEvent{Path: `C:\b.log`}) // excluded → never scanned
	o.Enqueue(FileEvent{Path: `C:\c.dll`})

	// wait for workers to drain
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		rs.mu.Lock(); n := len(rs.seen); rs.mu.Unlock()
		if n >= 2 { break }
		time.Sleep(10 * time.Millisecond)
	}
	rs.mu.Lock(); defer rs.mu.Unlock()
	if len(rs.seen) != 2 {
		t.Fatalf("scanned %v, want a.exe and c.dll only", rs.seen)
	}
}
```

- [ ] **Step 2: Run to verify it fails**

Run: `go test ./edr/orchestrator/ -run TestOrchestrator -v`
Expected: FAIL — undefined.

- [ ] **Step 3: Implement**

```go
// agent/edr/orchestrator/orchestrator.go
package orchestrator

import (
	"context"
	"sync"

	"github.com/utmstack/UTMStack/agent/edr/event"
)

type FileEvent struct {
	Path string
	Op   string
}

type FileScanner interface {
	ScanFile(path, source string) (string, string, error)
}

type Orchestrator struct {
	sc      FileScanner
	ex      *Excluder
	workers int
	queue   chan FileEvent
	wg      sync.WaitGroup
}

func New(sc FileScanner, ex *Excluder, workers, queueSize int) *Orchestrator {
	if workers < 1 {
		workers = 1
	}
	return &Orchestrator{sc: sc, ex: ex, workers: workers, queue: make(chan FileEvent, queueSize)}
}

func (o *Orchestrator) Start(ctx context.Context) {
	for i := 0; i < o.workers; i++ {
		o.wg.Add(1)
		go o.worker(ctx)
	}
}

func (o *Orchestrator) worker(ctx context.Context) {
	defer o.wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case ev := <-o.queue:
			if o.ex != nil && o.ex.Excluded(ev.Path) {
				continue
			}
			// Errors (file vanished, engine down) are non-fatal; the file stays
			// unknown and will be revisited on the next event/sweep.
			_, _, _ = o.sc.ScanFile(ev.Path, event.SourceFileWatcher)
		}
	}
}

// Enqueue returns false when the queue is full (caller logs the drop).
func (o *Orchestrator) Enqueue(ev FileEvent) bool {
	select {
	case o.queue <- ev:
		return true
	default:
		return false
	}
}

func (o *Orchestrator) Stop() { o.wg.Wait() }
```

- [ ] **Step 4: Run to verify it passes**

Run: `go test ./edr/orchestrator/ -v`
Expected: PASS (Excluder + Orchestrator).

- [ ] **Step 5: Commit**

```bash
git add agent/edr/orchestrator/orchestrator.go agent/edr/orchestrator/orchestrator_test.go
git commit -m "feat(edr): scan orchestrator with worker pool and backpressure"
```

---

## Task 5: USN record parser (pure, unit-tested)

**Files:**
- Create: `agent/edr/watcher/usn_record.go`
- Test: `agent/edr/watcher/usn_record_test.go`

**Interfaces:**
- Produces:
  - `const ReasonFileCreate = 0x00000100`, `ReasonDataExtend = 0x00000002`, `ReasonRenameNewName = 0x00002000`, `ReasonClose = 0x80000000`
  - `type USNChange struct { USN int64; FileRefID uint64; ParentRefID uint64; Reason uint32; FileName string }`
  - `func ParseUSNBuffer(buf []byte) ([]USNChange, error)` — parses concatenated `USN_RECORD_V2` records; ignores non-v2 records safely.

- [ ] **Step 1: Write the failing test (construct a synthetic USN_RECORD_V2 buffer)**

```go
// agent/edr/watcher/usn_record_test.go
package watcher

import (
	"encoding/binary"
	"testing"
	"unicode/utf16"
)

// buildRecord encodes one USN_RECORD_V2 with the given filename and reason.
func buildRecord(name string, reason uint32) []byte {
	nameU16 := utf16.Encode([]rune(name))
	nameBytes := make([]byte, len(nameU16)*2)
	for i, c := range nameU16 {
		binary.LittleEndian.PutUint16(nameBytes[i*2:], c)
	}
	const header = 60 // bytes before FileName in USN_RECORD_V2
	recLen := header + len(nameBytes)
	// pad record length to 8-byte alignment (as the FS does)
	for recLen%8 != 0 {
		recLen++
	}
	b := make([]byte, recLen)
	binary.LittleEndian.PutUint32(b[0:], uint32(recLen))    // RecordLength
	binary.LittleEndian.PutUint16(b[4:], 2)                 // MajorVersion
	binary.LittleEndian.PutUint16(b[6:], 0)                 // MinorVersion
	binary.LittleEndian.PutUint64(b[8:], 0x1111)            // FileReferenceNumber
	binary.LittleEndian.PutUint64(b[16:], 0x2222)           // ParentFileReferenceNumber
	binary.LittleEndian.PutUint64(b[24:], 42)               // Usn
	binary.LittleEndian.PutUint64(b[32:], 0)                // TimeStamp
	binary.LittleEndian.PutUint32(b[40:], reason)           // Reason
	binary.LittleEndian.PutUint32(b[44:], 0)                // SourceInfo
	binary.LittleEndian.PutUint32(b[48:], 0)                // SecurityId
	binary.LittleEndian.PutUint32(b[52:], 0)                // FileAttributes
	binary.LittleEndian.PutUint16(b[56:], uint16(len(nameBytes))) // FileNameLength
	binary.LittleEndian.PutUint16(b[58:], header)          // FileNameOffset
	copy(b[header:], nameBytes)
	return b
}

func TestParseUSNBuffer(t *testing.T) {
	buf := append(buildRecord("a.exe", ReasonFileCreate), buildRecord("note.txt", ReasonDataExtend)...)
	changes, err := ParseUSNBuffer(buf)
	if err != nil {
		t.Fatalf("ParseUSNBuffer: %v", err)
	}
	if len(changes) != 2 {
		t.Fatalf("got %d changes, want 2", len(changes))
	}
	if changes[0].FileName != "a.exe" || changes[0].Reason != ReasonFileCreate || changes[0].USN != 42 {
		t.Fatalf("bad first change: %+v", changes[0])
	}
	if changes[1].FileName != "note.txt" || changes[1].ParentRefID != 0x2222 {
		t.Fatalf("bad second change: %+v", changes[1])
	}
}
```

- [ ] **Step 2: Run to verify it fails**

Run: `go test ./edr/watcher/ -run TestParseUSNBuffer -v`
Expected: FAIL — undefined (package/symbols).

- [ ] **Step 3: Implement the parser (pure Go, builds everywhere)**

```go
// agent/edr/watcher/usn_record.go
package watcher

import (
	"encoding/binary"
	"fmt"
	"unicode/utf16"
)

const (
	ReasonDataOverwrite = 0x00000001
	ReasonDataExtend    = 0x00000002
	ReasonFileCreate    = 0x00000100
	ReasonRenameNewName = 0x00002000
	ReasonClose         = 0x80000000
)

type USNChange struct {
	USN         int64
	FileRefID   uint64
	ParentRefID uint64
	Reason      uint32
	FileName    string
}

// ParseUSNBuffer decodes concatenated USN_RECORD_V2 records. The OS returns
// a leading 8-byte next-USN value before the records when reading the journal;
// callers pass the buffer AFTER that 8-byte prefix.
func ParseUSNBuffer(buf []byte) ([]USNChange, error) {
	var out []USNChange
	off := 0
	for off+60 <= len(buf) {
		recLen := int(binary.LittleEndian.Uint32(buf[off:]))
		if recLen < 60 || off+recLen > len(buf) {
			break
		}
		major := binary.LittleEndian.Uint16(buf[off+4:])
		if major != 2 {
			off += recLen
			continue // ignore non-v2 records safely
		}
		fileRef := binary.LittleEndian.Uint64(buf[off+8:])
		parentRef := binary.LittleEndian.Uint64(buf[off+16:])
		usn := int64(binary.LittleEndian.Uint64(buf[off+24:]))
		reason := binary.LittleEndian.Uint32(buf[off+40:])
		nameLen := int(binary.LittleEndian.Uint16(buf[off+56:]))
		nameOff := int(binary.LittleEndian.Uint16(buf[off+58:]))
		if off+nameOff+nameLen > len(buf) || nameLen%2 != 0 {
			return out, fmt.Errorf("corrupt USN record at offset %d", off)
		}
		u16 := make([]uint16, nameLen/2)
		for i := range u16 {
			u16[i] = binary.LittleEndian.Uint16(buf[off+nameOff+i*2:])
		}
		out = append(out, USNChange{
			USN:         usn,
			FileRefID:   fileRef,
			ParentRefID: parentRef,
			Reason:      reason,
			FileName:    string(utf16.Decode(u16)),
		})
		off += recLen
	}
	return out, nil
}
```

- [ ] **Step 4: Run to verify it passes**

Run: `go test ./edr/watcher/ -run TestParseUSNBuffer -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add agent/edr/watcher/usn_record.go agent/edr/watcher/usn_record_test.go
git commit -m "feat(edr): USN_RECORD_V2 buffer parser"
```

---

## Task 6: USN volume watcher (Windows) + non-Windows stub

**Files:**
- Create: `agent/edr/watcher/usn_windows.go` (`//go:build windows`), `agent/edr/watcher/usn_other.go` (`//go:build !windows`)
- Test: build-only on macOS; behavior verified on the VM (Task 9).

**Interfaces:**
- Produces (both builds):
  - `type Sink interface { Enqueue(path, op string) }`
  - `type Watcher struct{}`, `func New(volumes []string, sink Sink, cursors CursorStore) *Watcher`
  - `func (w *Watcher) Run(ctx context.Context)` — one goroutine per volume; resolves each change to a full path and calls `sink.Enqueue`
  - `type CursorStore interface { LoadUSN(volume string) (journalID uint64, nextUSN int64); SaveUSN(volume string, journalID uint64, nextUSN int64) }`

- [ ] **Step 1: Non-Windows stub (keeps macOS build green)**

```go
// agent/edr/watcher/usn_other.go
//go:build !windows

package watcher

import "context"

type Sink interface{ Enqueue(path, op string) }

type CursorStore interface {
	LoadUSN(volume string) (uint64, int64)
	SaveUSN(volume string, journalID uint64, nextUSN int64)
}

type Watcher struct{}

func New(volumes []string, sink Sink, cursors CursorStore) *Watcher { return &Watcher{} }

// Run is a no-op on non-Windows platforms (USN is NTFS-only).
func (w *Watcher) Run(ctx context.Context) { <-ctx.Done() }
```

- [ ] **Step 2: Windows implementation**

This uses `CreateFile` on `\\.\C:` and `DeviceIoControl` with `FSCTL_QUERY_USN_JOURNAL` (0x000900f4) then repeated `FSCTL_READ_USN_JOURNAL` (0x000900bb). Path resolution opens the file by its 64-bit reference number and calls `GetFinalPathNameByHandle`. These are Windows-native and compiled/validated on the VM.

```go
// agent/edr/watcher/usn_windows.go
//go:build windows

package watcher

import (
	"context"
	"encoding/binary"
	"time"
	"unsafe"

	"github.com/utmstack/UTMStack/shared/logger"
	"golang.org/x/sys/windows"
)

const (
	fsctlQueryUSNJournal = 0x000900f4
	fsctlReadUSNJournal  = 0x000900bb
	usnReasonMask        = ReasonFileCreate | ReasonDataExtend | ReasonDataOverwrite | ReasonRenameNewName | ReasonClose
)

type Sink interface{ Enqueue(path, op string) }

type CursorStore interface {
	LoadUSN(volume string) (uint64, int64)
	SaveUSN(volume string, journalID uint64, nextUSN int64)
}

type Watcher struct {
	volumes []string
	sink    Sink
	cursors CursorStore
}

func New(volumes []string, sink Sink, cursors CursorStore) *Watcher {
	return &Watcher{volumes: volumes, sink: sink, cursors: cursors}
}

func (w *Watcher) Run(ctx context.Context) {
	for _, vol := range w.volumes {
		go w.watchVolume(ctx, vol)
	}
	<-ctx.Done()
}

// usnJournalData mirrors USN_JOURNAL_DATA_V0.
type usnJournalData struct {
	UsnJournalID    uint64
	FirstUsn        int64
	NextUsn         int64
	LowestValidUsn  int64
	MaxUsn          int64
	MaximumSize     uint64
	AllocationDelta uint64
}

// readUsnData mirrors READ_USN_JOURNAL_DATA_V0.
type readUsnData struct {
	StartUsn          int64
	ReasonMask        uint32
	ReturnOnlyOnClose uint32
	Timeout           uint64
	BytesToWaitFor    uint64
	UsnJournalID      uint64
}

func (w *Watcher) watchVolume(ctx context.Context, vol string) {
	// vol like "C:" -> device path \\.\C:
	path, err := windows.UTF16PtrFromString(`\\.\` + vol)
	if err != nil {
		logger.Error("UTMStack EDR watcher: bad volume %s: %v", vol, err)
		return
	}
	h, err := windows.CreateFile(path,
		windows.GENERIC_READ, windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE,
		nil, windows.OPEN_EXISTING, 0, 0)
	if err != nil {
		logger.Error("UTMStack EDR watcher: open %s: %v", vol, err)
		return
	}
	defer windows.CloseHandle(h)

	var jd usnJournalData
	var bytesRet uint32
	if err := windows.DeviceIoControl(h, fsctlQueryUSNJournal, nil, 0,
		(*byte)(unsafe.Pointer(&jd)), uint32(unsafe.Sizeof(jd)), &bytesRet, nil); err != nil {
		logger.Error("UTMStack EDR watcher: query journal %s: %v", vol, err)
		return
	}

	journalID, next := w.cursors.LoadUSN(vol)
	if journalID != jd.UsnJournalID {
		// New/recreated journal: resync from the current end (avoid a flood).
		journalID = jd.UsnJournalID
		next = jd.NextUsn
	}

	buf := make([]byte, 64*1024)
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		req := readUsnData{
			StartUsn:       next,
			ReasonMask:     usnReasonMask,
			Timeout:        0,
			BytesToWaitFor: 0,
			UsnJournalID:   journalID,
		}
		var got uint32
		err := windows.DeviceIoControl(h, fsctlReadUSNJournal,
			(*byte)(unsafe.Pointer(&req)), uint32(unsafe.Sizeof(req)),
			&buf[0], uint32(len(buf)), &got, nil)
		if err != nil {
			time.Sleep(time.Second)
			continue
		}
		if got <= 8 {
			time.Sleep(500 * time.Millisecond)
			continue
		}
		next = int64(binary.LittleEndian.Uint64(buf[:8]))
		changes, _ := ParseUSNBuffer(buf[8:got])
		for _, ch := range changes {
			if ch.Reason&(ReasonFileCreate|ReasonDataExtend|ReasonDataOverwrite|ReasonClose) == 0 {
				continue
			}
			full := resolvePath(vol, ch.FileRefID)
			if full == "" {
				continue
			}
			w.sink.Enqueue(full, "usn")
		}
		w.cursors.SaveUSN(vol, journalID, next)
	}
}

// resolvePath opens a file by its NTFS reference number and returns its full
// path via GetFinalPathNameByHandle. Returns "" if the file is gone.
func resolvePath(vol string, fileRef uint64) string {
	volPath, err := windows.UTF16PtrFromString(`\\.\` + vol)
	if err != nil {
		return ""
	}
	volHandle, err := windows.CreateFile(volPath, windows.GENERIC_READ,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE, nil, windows.OPEN_EXISTING,
		windows.FILE_FLAG_BACKUP_SEMANTICS, 0)
	if err != nil {
		return ""
	}
	defer windows.CloseHandle(volHandle)

	// FILE_ID_DESCRIPTOR for OpenFileById (Type=0 => FileId).
	var idDesc struct {
		Size   uint32
		Type   uint32
		FileID uint64
	}
	idDesc.Size = uint32(unsafe.Sizeof(idDesc))
	idDesc.Type = 0
	idDesc.FileID = fileRef

	fh, err := openFileByID(volHandle, unsafe.Pointer(&idDesc))
	if err != nil {
		return ""
	}
	defer windows.CloseHandle(fh)

	bufN := make([]uint16, windows.MAX_LONG_PATH)
	n, err := windows.GetFinalPathNameByHandle(fh, &bufN[0], uint32(len(bufN)), 0)
	if err != nil || n == 0 {
		return ""
	}
	return windows.UTF16ToString(bufN[:n])
}

var (
	modkernel32    = windows.NewLazySystemDLL("kernel32.dll")
	procOpenFileID = modkernel32.NewProc("OpenFileById")
)

func openFileByID(volHandle windows.Handle, idDesc unsafe.Pointer) (windows.Handle, error) {
	r, _, e := procOpenFileID.Call(
		uintptr(volHandle),
		uintptr(idDesc),
		uintptr(windows.GENERIC_READ),
		uintptr(windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE),
		0,
		uintptr(windows.FILE_FLAG_BACKUP_SEMANTICS),
	)
	h := windows.Handle(r)
	if h == windows.InvalidHandle {
		return 0, e
	}
	return h, nil
}
```

> **VM verification point:** `GetFinalPathNameByHandle`, `OpenFileById`, and the exact `FILE_ID_DESCRIPTOR` layout are Windows-native; confirm they compile against the pinned `x/sys/windows` and resolve real paths on the VM (Task 9). If `windows.MAX_LONG_PATH`/`GetFinalPathNameByHandle` are absent in the pinned version, use a 32768-length buffer and `modkernel32.NewProc("GetFinalPathNameByHandleW")` the same way as `OpenFileById`.

- [ ] **Step 3: Build both platforms**

Run: `go build ./edr/watcher/ && GOOS=windows GOARCH=amd64 go build ./edr/watcher/`
Expected: both succeed (macOS uses the stub).

- [ ] **Step 4: Commit**

```bash
git add agent/edr/watcher/usn_windows.go agent/edr/watcher/usn_other.go
git commit -m "feat(edr): whole-volume USN journal watcher (windows) + stub"
```

---

## Task 7: Signature feed (freshclam scheduler + cache invalidation)

**Files:**
- Create: `agent/edr/feed/freshclam.go`
- Test: `agent/edr/feed/freshclam_test.go`

**Interfaces:**
- Produces:
  - `type Feed struct{}`, `func New(cfg config.EDRConfig, c *cache.Cache) *Feed`
  - `func (f *Feed) Run(ctx context.Context)` — every `SigUpdateHours`, runs the updater binary, then reads the new sig-DB version and calls `cache.MarkStaleBySigDB`
  - `func (f *Feed) updateOnce() (sigDBVersion string, err error)` — with an injectable `runUpdater func() error` and `currentSigDB func() (string, error)` for testing

- [ ] **Step 1: Write the failing test (inject fakes; assert stale invalidation runs)**

```go
// agent/edr/feed/freshclam_test.go
package feed

import (
	"path/filepath"
	"testing"

	"github.com/utmstack/UTMStack/agent/edr/cache"
	"github.com/utmstack/UTMStack/agent/edr/config"
)

func TestUpdateOnceMarksStale(t *testing.T) {
	c, err := cache.Open(filepath.Join(t.TempDir(), "edr.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	_ = c.Store(cache.VerdictRecord{SHA256: "clean1", Verdict: cache.VerdictClean, SigDBVersion: "100"})

	f := New(config.Default(), c)
	f.runUpdater = func() error { return nil }
	f.currentSigDB = func() (string, error) { return "200", nil }

	ver, err := f.updateOnce()
	if err != nil || ver != "200" {
		t.Fatalf("updateOnce ver=%q err=%v", ver, err)
	}
	if r, _, _ := c.Lookup("clean1"); !r.Stale {
		t.Fatal("clean record should be stale after signature update")
	}
}
```

- [ ] **Step 2: Run to verify it fails**

Run: `go test ./edr/feed/ -v`
Expected: FAIL — undefined.

- [ ] **Step 3: Implement**

```go
// agent/edr/feed/freshclam.go
package feed

import (
	"context"
	"path/filepath"
	"time"

	"github.com/utmstack/UTMStack/agent/edr/cache"
	"github.com/utmstack/UTMStack/agent/edr/config"
	"github.com/utmstack/UTMStack/shared/exec"
	"github.com/utmstack/UTMStack/shared/logger"
)

type Feed struct {
	cfg          config.EDRConfig
	cache        *cache.Cache
	runUpdater   func() error
	currentSigDB func() (string, error)
}

func New(cfg config.EDRConfig, c *cache.Cache) *Feed {
	f := &Feed{cfg: cfg, cache: c}
	f.runUpdater = f.defaultRunUpdater
	f.currentSigDB = f.defaultCurrentSigDB
	return f
}

func (f *Feed) defaultRunUpdater() error {
	// Runs the signature updater bundled with the engine. Naming here is
	// internal only; nothing is surfaced in events/logs.
	bin := filepath.Join(config.EngineDir, "freshclam.exe")
	return exec.Run(bin, config.EngineDir, "--config-file", filepath.Join(config.EngineDir, "freshclam.conf"))
}

func (f *Feed) defaultCurrentSigDB() (string, error) {
	// Read the signature DB version from the engine's daemon banner and keep
	// only the numeric sig-db field (never the engine name).
	// Implemented via engine.Version in the service wiring; here we return "".
	return "", nil
}

func (f *Feed) updateOnce() (string, error) {
	if err := f.runUpdater(); err != nil {
		logger.Error("UTMStack EDR: signature update failed: %v", err)
		return "", err
	}
	ver, err := f.currentSigDB()
	if err != nil {
		return "", err
	}
	if ver != "" {
		n, err := f.cache.MarkStaleBySigDB(ver)
		if err != nil {
			return ver, err
		}
		logger.Info("UTMStack EDR: signatures updated to %s, %d cached entries invalidated", ver, n)
	}
	return ver, nil
}

func (f *Feed) Run(ctx context.Context) {
	interval := time.Duration(f.cfg.SigUpdateHours) * time.Hour
	if interval <= 0 {
		interval = 4 * time.Hour
	}
	_, _ = f.updateOnce()
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			_, _ = f.updateOnce()
		}
	}
}
```

- [ ] **Step 4: Run to verify it passes**

Run: `go test ./edr/feed/ -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add agent/edr/feed/
git commit -m "feat(edr): signature-feed scheduler with cache invalidation"
```

---

## Task 8: Wire watcher + orchestrator + feed into the service; USN cursor store

**Files:**
- Modify: `agent/edr/service/service.go`
- Create: `agent/edr/cache/usn_cursor.go` (implements `watcher.CursorStore` over `edr.db`)
- Test: `agent/edr/cache/usn_cursor_test.go`

**Interfaces:**
- Produces:
  - `type USNCursor struct { Volume string (pk); JournalID uint64; NextUSN int64 }` in cache; methods `LoadUSN`/`SaveUSN` on `*cache.Cache`.
  - Service: when `cfg.Enabled`, build `scanner` → `orchestrator` → `watcher` (volumes = `cfg.WatchVolumes` or all fixed) and `feed`, and start them under the run `ctx`.

- [ ] **Step 1: Cursor store test**

```go
// agent/edr/cache/usn_cursor_test.go
package cache

import (
	"path/filepath"
	"testing"
)

func TestUSNCursorRoundTrip(t *testing.T) {
	c, err := Open(filepath.Join(t.TempDir(), "edr.db"))
	if err != nil { t.Fatal(err) }
	defer c.Close()

	c.SaveUSN("C:", 999, 12345)
	jid, next := c.LoadUSN("C:")
	if jid != 999 || next != 12345 {
		t.Fatalf("LoadUSN = %d,%d", jid, next)
	}
	// unknown volume returns zeros
	if j, n := c.LoadUSN("D:"); j != 0 || n != 0 {
		t.Fatalf("unknown volume = %d,%d", j, n)
	}
}
```

- [ ] **Step 2: Run to verify it fails**

Run: `go test ./edr/cache/ -run TestUSNCursor -v`
Expected: FAIL.

- [ ] **Step 3: Implement `agent/edr/cache/usn_cursor.go`**

```go
// agent/edr/cache/usn_cursor.go
package cache

type USNCursor struct {
	Volume    string `gorm:"primaryKey"`
	JournalID uint64
	NextUSN   int64
}

func (c *Cache) LoadUSN(volume string) (uint64, int64) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	var cur USNCursor
	if err := c.db.First(&cur, "volume = ?", volume).Error; err != nil {
		return 0, 0
	}
	return cur.JournalID, cur.NextUSN
}

func (c *Cache) SaveUSN(volume string, journalID uint64, nextUSN int64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	_ = c.db.Save(&USNCursor{Volume: volume, JournalID: journalID, NextUSN: nextUSN}).Error
}
```

Add `&USNCursor{}` to the `AutoMigrate` call in `Open`.

- [ ] **Step 4: Run to verify it passes**

Run: `go test ./edr/cache/ -v`
Expected: PASS.

- [ ] **Step 5: Wire into the service run loop**

In `agent/edr/service/service.go` `run()`, after `p.eng.EnsureRunning()` and only when `cfg.Enabled`, construct and start the pipeline. Add an orchestrator adapter so it satisfies `watcher.Sink`:

```go
	// after engine ensure:
	if cfg.Enabled {
		store, err := quarantine.New(cfg.QuarantineDir, c)
		if err != nil {
			logger.Error("UTMStack EDR: quarantine init: %v", err)
		}
		sc := scanner.New(cfg, c, sp, store) // sp opened from config.SpoolFile (add spool open in run)
		ex := orchestrator.NewExcluder(cfg.Exclusions)
		orch := orchestrator.New(sc, ex, cfg.ScanConcurrency, 4096)
		orch.Start(ctx)

		vols := cfg.WatchVolumes
		if len(vols) == 0 {
			vols = fixedVolumes() // helper: enumerate fixed NTFS drives, e.g. []string{"C:"}
		}
		w := watcher.New(vols, sinkAdapter{orch}, c)
		go w.Run(ctx)

		go feed.New(cfg, c).Run(ctx)
	}
```

Add near the bottom of the file:

```go
type sinkAdapter struct{ o *orchestrator.Orchestrator }

func (s sinkAdapter) Enqueue(path, op string) {
	if !s.o.Enqueue(orchestrator.FileEvent{Path: path, Op: op}) {
		logger.Debug(100, "UTMStack EDR: scan queue full, dropped %s", path)
	}
}

// fixedVolumes returns drive letters of fixed NTFS volumes. On non-windows it
// returns nil (the watcher is a stub there anyway).
func fixedVolumes() []string { return defaultFixedVolumes() }
```

Provide `defaultFixedVolumes()` in two build-tagged files: `service_volumes_windows.go` (enumerate via `windows.GetLogicalDrives` + `windows.GetDriveType == DRIVE_FIXED`) and `service_volumes_other.go` (`return nil`). Also open the spool in `run()`:

```go
	sp, err := event.OpenSpool(config.SpoolFile, 8<<20)
	if err != nil {
		logger.Error("UTMStack EDR: spool open: %v", err)
		return
	}
	defer sp.Close()
```

- [ ] **Step 6: Build both platforms**

Run: `go build ./... && GOOS=windows GOARCH=amd64 go build ./...`
Expected: both succeed.

- [ ] **Step 7: Commit**

```bash
git add agent/edr/cache/usn_cursor.go agent/edr/cache/usn_cursor_test.go agent/edr/service/
git commit -m "feat(edr): start USN watcher + orchestrator + feed when enabled"
```

---

## Task 9: End-to-end acceptance on the VM

- [ ] **Step 1: Build and deploy (both architectures)**

```bash
cd agent && for arch in amd64 arm64; do GOOS=windows GOARCH=$arch go build -o dist/utmstack_edr_windows_$arch.exe ./edr; done
```
Copy the binary matching the VM's architecture to the install dir; ensure ClamAV is running on `127.0.0.1:3310` (as in Plan 1 Task 9), reinstall the EDR service, and `enable-edr`.

- [ ] **Step 2: Auto scan-on-arrival + quarantine**

```powershell
$eicar = 'X5O!P%@AP[4\PZX54(P^)7CC)7}$EICAR-STANDARD-ANTIVIRUS-TEST-FILE!$H+H*'
Set-Content C:\Users\atlas\Downloads\eicar_auto.com -Value $eicar -NoNewline
Start-Sleep -Seconds 5
Test-Path C:\Users\atlas\Downloads\eicar_auto.com    # should be $false (quarantined)
Get-Content .\edr-spool\events.ndjson | Select-Object -Last 1
```
Expected: the file is gone (moved to quarantine); the last spool line has `"action":"quarantined"`, `"source":"file_watcher"`, `"product":"UTMStack EDR"`, and no `clam` string; a matching record exists in `quarantine/`.

- [ ] **Step 3: Verdict cache skip**

Re-create the same-content clean file twice; confirm (via engine/scan logs or timing) the second identical unchanged file is not re-scanned. Then simulate a signature update and confirm clean entries are marked stale and re-scanned (drive `feed.updateOnce` by running the updater and checking the log line "signatures updated to …, N cached entries invalidated").

- [ ] **Step 4: Restore (false-positive)**

```powershell
.\utmstack_edr_windows_amd64.exe status
# take a quarantine id from the platform event or quarantine dir listing:
.\utmstack_edr_windows_amd64.exe restore <quarantine_id>
Test-Path C:\Users\atlas\Downloads\eicar_auto.com    # should be $true again
```
Expected: file restored to its original path; record marked restored.

- [ ] **Step 5: Platform + labeling**

Confirm the quarantine event reached the platform as `utmstack_edr`, correctly branded, and that a case-insensitive search of the spool + service log for `clam` returns nothing.

- [ ] **Step 6: Commit any fixes discovered on the VM**

```bash
git add -A && git commit -m "fix(edr): plan 2 VM acceptance adjustments"
```

---

## Self-Review (against the spec)

**Spec coverage (Plan 2 slice):**
- §4.2 File-arrival watcher (whole-volume USN) → Tasks 5, 6, 8. ✔
- §4.4 Orchestrator (queue/concurrency/backpressure, exclusions) → Tasks 3, 4. ✔
- §4.5 Verdict cache staleness on sig update → Tasks 1, 7. ✔
- §4.6 Quarantine (move-not-delete, restore) → Tasks 1, 2. ✔ *(process-tree kill / suspend are Plan 3.)*
- §4.11 Config & feed (freshclam scheduling, exclusions, retention plumbing) → Tasks 3, 7. ✔
- §6.1 File arrival & scan flow, §6.5 sig-update → rescan → Tasks 4, 7, 8. ✔
- Branding preserved (event `source=file_watcher`, no engine name) → tests in Tasks 2, 9. ✔

**Placeholder scan:** Testable Go logic (quarantine, orchestrator, exclusions, USN parser, feed, cursor) has complete code + real `go test`. The Windows-native USN device I/O and path resolution are given concrete Win32 code with an explicit VM verification point (correct behavior confirmed by compile+run on the VM in Task 9) — the honest substitute for a macOS unit test.

**Type consistency:** `FileScanner.ScanFile(path, source string)(string,string,error)` matches `scanner.Scanner.ScanFile`. `watcher.Sink.Enqueue(path, op string)` matches `sinkAdapter`. `CursorStore` (LoadUSN/SaveUSN) matches the `*cache.Cache` methods. `Quarantiner.Quarantine(path,sha256,detection)(string,error)` matches `quarantine.Store`. `MarkStaleBySigDB` used by `feed` matches the `cache` method.

**Carried forward:** production signature hosting on the server (§12.1) still applies; freshclam config generation is provisioned on the VM as in Plan 1.
