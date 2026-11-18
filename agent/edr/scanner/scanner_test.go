package scanner

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/utmstack/UTMStack/agent/edr/cache"
	"github.com/utmstack/UTMStack/agent/edr/config"
	"github.com/utmstack/UTMStack/agent/edr/engine"
	"github.com/utmstack/UTMStack/agent/edr/event"
	"github.com/utmstack/UTMStack/agent/edr/quarantine"
	"github.com/utmstack/UTMStack/agent/edr/signatures"
)

type pendingQuarantiner struct {
	Quarantiner
	identity os.FileInfo
}

func (q *pendingQuarantiner) QuarantineVerified(path, hash, detection string, info os.FileInfo) (string, error) {
	q.identity = info
	id, err := q.Quarantiner.QuarantineVerified(path, hash, detection, info)
	if err != nil {
		return id, err
	}
	return id, &quarantine.ContainmentError{QuarantineID: id, Err: errors.New("test finalization interruption")}
}

func TestScannerReportsPhysicalContainmentWithPendingFinalization(t *testing.T) {
	s, dir, _, spoolPath := storeFixture(t)
	p := filepath.Join(dir, "marker")
	if err := os.WriteFile(p, []byte("harmless detection marker"), 0600); err != nil {
		t.Fatal(err)
	}
	// Fstat pins Windows identity before containment moves the pathname.
	file, err := os.Open(p)
	if err != nil {
		t.Fatal(err)
	}
	before, err := file.Stat()
	closeErr := file.Close()
	if err != nil || closeErr != nil {
		t.Fatal(errors.Join(err, closeErr))
	}
	q := &pendingQuarantiner{Quarantiner: s.quarantiner}
	s.quarantiner = q
	s.scanSnapshot = snapshotForTest(s, func(context.Context, []byte) (bool, string, error) {
		return false, "Safe.Pending.Control", nil
	})
	verdict, _, err := s.ScanFile(p, event.SourceFileWatcher)
	if err != nil || verdict != cache.VerdictMalicious {
		t.Fatalf("detection failed: %s %v", verdict, err)
	}
	if q.identity == nil || !os.SameFile(before, q.identity) {
		t.Fatal("backend did not receive the scanned identity")
	}
	if _, err = os.Stat(p); !os.IsNotExist(err) {
		t.Fatal("physical containment did not happen", err)
	}
	lines := spoolLines(t, spoolPath)
	if len(lines) != 1 {
		t.Fatalf("events=%v", lines)
	}
	var got event.Event
	if err = json.Unmarshal([]byte(lines[0]), &got); err != nil {
		t.Fatal(err)
	}
	if got.Action != event.ActionQuarantined || got.ResponseStatus != "recovery_pending" {
		t.Fatalf("physical result hidden: %+v", got)
	}
}

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

	fq := &fakeQ{}
	s := New(config.Default(), c, sp, fq, engine.New(config.Default()))
	s.scanSnapshot = snapshotForTest(s, func(addr context.Context, data []byte) (bool, string, error) {
		return false, "Win.Test.EICAR_HDB-1", nil // pretend malicious
	})

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
	if !fq.called {
		t.Fatal("quarantiner was not invoked for malicious file")
	}
	if !strings.Contains(line, `"action":"quarantined"`) {
		t.Fatalf("expected quarantined action: %s", line)
	}
}

func scannerFixture(t *testing.T, content []byte) (*Scanner, *cache.Cache, string, *fakeQ) {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "sample.bin")
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatal(err)
	}
	c, err := cache.Open(filepath.Join(dir, "edr.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = c.Close() })
	q := &fakeQ{}
	s := New(config.Default(), c, nil, q, engine.New(config.Default()))
	s.sigDBVersion = func() string { return "generation-1" }
	return s, c, path, q
}

func TestCacheRequiresMatchingGeneration(t *testing.T) {
	s, c, path, _ := scannerFixture(t, []byte("unchanged"))
	hash := mustHash(t, path)
	if err := c.Store(cache.VerdictRecord{SHA256: hash, Verdict: cache.VerdictClean, SigDBVersion: "generation-0"}); err != nil {
		t.Fatal(err)
	}
	calls := 0
	s.scanSnapshot = snapshotForTest(s, func(_ context.Context, _ []byte) (bool, string, error) { calls++; return true, "", nil })
	for i := 0; i < 2; i++ {
		if _, _, err := s.ScanFile(path, event.SourceEngine); err != nil {
			t.Fatal(err)
		}
	}
	if calls != 1 {
		t.Fatalf("expected old clean generation rescanned once, got %d", calls)
	}
}

func TestMetadataSensitiveScanBypassesHashOnlyCache(t *testing.T) {
	s, c, safePath, _ := scannerFixture(t, []byte{0x7f, 0x80, 0, 1, 0xff, 0xfe, 0xaa, 0xbb})
	triggerPath := filepath.Join(filepath.Dir(safePath), "trigger.bin")
	data, err := os.ReadFile(safePath)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(triggerPath, data, 0o600); err != nil {
		t.Fatal(err)
	}
	hash := mustHash(t, safePath)
	if err := c.Store(cache.VerdictRecord{SHA256: hash, Verdict: cache.VerdictClean, SigDBVersion: "generation-1"}); err != nil {
		t.Fatal(err)
	}
	s.metadataSensitive = func() bool { return true }
	s.scanSnapshot = func(context.Context, []byte) (bool, string, string, error) {
		t.Fatal("metadata-sensitive scan used byte-only entry point")
		return false, "", "", nil
	}
	calls := 0
	s.scanSnapshotWithPath = func(_ context.Context, _ []byte, path string) (bool, string, string, error) {
		calls++
		if path == triggerPath {
			return false, "Native.Metadata.Filename", "generation-1", nil
		}
		return true, "", "generation-1", nil
	}
	for _, tc := range []struct {
		path, wantVerdict, wantSig string
	}{
		{safePath, cache.VerdictClean, ""},
		{triggerPath, cache.VerdictMalicious, "Native.Metadata.Filename"},
		{safePath, cache.VerdictClean, ""},
	} {
		verdict, sig, _, _, err := s.analyze(tc.path, event.SourceEngine)
		if err != nil || verdict != tc.wantVerdict || sig != tc.wantSig {
			t.Fatalf("%s: verdict=%s sig=%q err=%v", tc.path, verdict, sig, err)
		}
	}
	if calls != 3 {
		t.Fatalf("same bytes at different names were scanned %d times, want 3", calls)
	}
	if rec, found, err := c.Lookup(hash); err != nil || !found || rec.Verdict != cache.VerdictClean {
		t.Fatalf("metadata-sensitive scan changed hash-only cache entry: %+v found=%v err=%v", rec, found, err)
	}
}

func TestLogicalJSONRootNameBypassesRealVerdictCache(t *testing.T) {
	data := []byte("harmless UTMSTACK-BYTECODE-HARMLESS-MARKER content")
	_, c, safePath, q := scannerFixture(t, data)
	triggerPath := filepath.Join(filepath.Dir(safePath), "root.file")
	if err := os.WriteFile(triggerPath, data, 0o600); err != nil {
		t.Fatal(err)
	}
	program, err := os.ReadFile("../engine/testdata/metadata_root_name.cbc")
	if err != nil {
		t.Fatal(err)
	}
	cfg := config.Default()
	cfg.CollectMetadata = true
	cfg.SignatureDBDir = t.TempDir()
	eng := engine.New(cfg)
	if _, err := eng.ActivateGeneration(map[string][]byte{"name.cbc": program}, 1, strings.Repeat("1", 64)); err != nil {
		t.Fatal(err)
	}
	s := New(cfg, c, nil, q, eng)
	hash := mustHash(t, safePath)
	// Both a clean and a malicious content-only cache record must be ignored
	// by the real scanner/engine wiring. Equal bytes at these two names have
	// different native-validated logical bytecode verdicts.
	for _, cached := range []string{cache.VerdictClean, cache.VerdictMalicious} {
		if err := c.Store(cache.VerdictRecord{SHA256: hash, Verdict: cached, Signature: "Stale.Path.Label", SigDBVersion: eng.Version()}); err != nil {
			t.Fatal(err)
		}
		for _, tc := range []struct{ path, verdict, label string }{
			{safePath, cache.VerdictClean, ""},
			{triggerPath, cache.VerdictMalicious, "Bytecode.Native.Detected"},
			{safePath, cache.VerdictClean, ""},
		} {
			verdict, label, _, _, err := s.analyze(tc.path, event.SourceEngine)
			if err != nil || verdict != tc.verdict || label != tc.label {
				t.Fatalf("cached=%s path=%s verdict=%s label=%q err=%v", cached, tc.path, verdict, label, err)
			}
		}
		if row, found, err := c.Lookup(hash); err != nil || !found || row.Verdict != cached || row.Signature != "Stale.Path.Label" {
			t.Fatalf("path-specific scan modified the content-only cache: row=%+v found=%v err=%v", row, found, err)
		}
	}
}

func TestWithdrawnMaliciousRuleIsReevaluated(t *testing.T) {
	s, c, path, q := scannerFixture(t, []byte("false positive"))
	if err := c.Store(cache.VerdictRecord{SHA256: mustHash(t, path), Verdict: cache.VerdictMalicious, Signature: "Withdrawn", SigDBVersion: "generation-0"}); err != nil {
		t.Fatal(err)
	}
	s.scanSnapshot = snapshotForTest(s, func(_ context.Context, _ []byte) (bool, string, error) { return true, "", nil })
	v, _, err := s.ScanFile(path, event.SourceEngine)
	if err != nil || v != cache.VerdictClean || q.called {
		t.Fatalf("withdrawn rule persisted: verdict=%s err=%v quarantined=%v", v, err, q.called)
	}
}

func TestErrorAndOversizeNeverCacheClean(t *testing.T) {
	s, c, path, q := scannerFixture(t, []byte("payload"))
	hash := mustHash(t, path)
	calls := 0
	s.scanSnapshot = snapshotForTest(s, func(_ context.Context, _ []byte) (bool, string, error) {
		calls++
		return true, "", errors.New("incomplete")
	})
	if _, _, err := s.ScanFile(path, event.SourceEngine); err == nil {
		t.Fatal("expected scan error")
	}
	s.maxFileBytes = 3
	if _, _, err := s.ScanFile(path, event.SourceEngine); err == nil {
		t.Fatal("expected input limit error")
	}
	if calls != 1 {
		t.Fatal("oversized input reached engine")
	}
	if _, found, err := c.Lookup(hash); err != nil || found {
		t.Fatalf("incomplete result cached: found=%v err=%v", found, err)
	}
	if q.called {
		t.Fatal("incomplete scan caused containment")
	}
}

// A replacement written during analysis must not be moved by the path-based
// quarantine, but the confirmed match on the analyzed bytes is still reported.
func TestChangedFileIsNotQuarantinedButStillDetected(t *testing.T) {
	s, _, path, q := scannerFixture(t, []byte("malicious snapshot"))
	s.scanSnapshot = snapshotForTest(s, func(_ context.Context, data []byte) (bool, string, error) {
		if string(data) != "malicious snapshot" {
			t.Fatalf("wrong analyzed bytes: %q", data)
		}
		if err := os.WriteFile(path, []byte("benign replacement"), 0o600); err != nil {
			t.Fatal(err)
		}
		return false, "Test.Malicious", nil
	})
	v, sig, err := s.ScanFile(path, event.SourceEngine)
	if err != nil || v != cache.VerdictMalicious || sig != "Test.Malicious" || q.called {
		t.Fatalf("replacement handling: verdict=%s sig=%s err=%v quarantined=%v", v, sig, err, q.called)
	}
}

// The file watcher (sharing this scanner) can quarantine a dropped image while
// the process guard is still analysing it. The guard must still receive the
// malicious verdict so it kills the launched process, and the quarantining
// scan's own event is not duplicated.
func TestConfirmedMatchSurvivesConcurrentQuarantine(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "dropped.exe")
	if err := os.WriteFile(path, []byte("malicious image bytes"), 0o600); err != nil {
		t.Fatal(err)
	}
	c, err := cache.Open(filepath.Join(dir, "edr.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = c.Close() })
	store, err := quarantine.New(filepath.Join(dir, "quarantine"), c)
	if err != nil {
		t.Fatal(err)
	}
	spoolPath := filepath.Join(dir, "events.ndjson")
	sp, err := event.OpenSpool(spoolPath, 1<<20)
	if err != nil {
		t.Fatal(err)
	}
	defer sp.Close()
	s := New(config.Default(), c, sp, store, engine.New(config.Default()))
	s.sigDBVersion = func() string { return "generation-1" }
	s.scanSnapshot = snapshotForTest(s, func(_ context.Context, data []byte) (bool, string, error) {
		sum := mustHash(t, path)
		if _, err := store.Quarantine(path, sum, "Test.Malicious"); err != nil {
			t.Fatal(err)
		}
		return false, "Test.Malicious", nil
	})
	v, sig, err := s.ScanFile(path, event.SourceProcessWatcher)
	if v != cache.VerdictMalicious || sig != "Test.Malicious" || err != nil {
		t.Fatalf("process guard got verdict=%q sig=%q err=%v; the launched process would not be killed", v, sig, err)
	}
	if b, _ := os.ReadFile(spoolPath); len(b) != 0 {
		t.Fatalf("already-contained detection reported again: %s", b)
	}
}

// A running image renamed during analysis cannot be contained by path, but
// the confirmed match must still be returned (so the guard kills the PID it
// holds) and reported as a detection.
func TestConfirmedMatchReportedWhenTargetRenamed(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "launched.exe")
	if err := os.WriteFile(path, []byte("malicious image bytes"), 0o600); err != nil {
		t.Fatal(err)
	}
	c, err := cache.Open(filepath.Join(dir, "edr.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = c.Close() })
	spoolPath := filepath.Join(dir, "events.ndjson")
	sp, err := event.OpenSpool(spoolPath, 1<<20)
	if err != nil {
		t.Fatal(err)
	}
	defer sp.Close()
	q := &fakeQ{}
	s := New(config.Default(), c, sp, q, engine.New(config.Default()))
	s.sigDBVersion = func() string { return "generation-1" }
	s.scanSnapshot = snapshotForTest(s, func(_ context.Context, data []byte) (bool, string, error) {
		if err := os.Rename(path, filepath.Join(dir, "renamed.exe")); err != nil {
			t.Fatal(err)
		}
		return false, "Test.Malicious", nil
	})
	v, sig, err := s.ScanFile(path, event.SourceProcessWatcher)
	b, _ := os.ReadFile(spoolPath)
	if v != cache.VerdictMalicious || sig != "Test.Malicious" || err != nil {
		t.Fatalf("confirmed malicious bytes reported as verdict=%q sig=%q err=%v; spool=%s", v, sig, err, b)
	}
	line := strings.TrimSpace(string(b))
	if strings.Count(line, "\n") != 0 || !strings.Contains(line, `"action":"detected"`) ||
		!strings.Contains(line, `"verdict":"malicious"`) || !strings.Contains(line, `"file_hash":"`) {
		t.Fatalf("no single detection event for confirmed malicious bytes: %s", b)
	}
	if q.called {
		t.Fatal("path-based quarantine must still be skipped when the target changed")
	}
}

// A recent quarantine at the same path only stands in for this detection when
// it holds these exact bytes. A different earlier file there must not hide a
// new drop that vanished during analysis.
func TestVanishedTargetReportedWhenEarlierQuarantineDiffers(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "dropped.exe")
	c, err := cache.Open(filepath.Join(dir, "edr.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = c.Close() })
	store, err := quarantine.New(filepath.Join(dir, "quarantine"), c)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("earlier dropped bytes"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Quarantine(path, mustHash(t, path), "Test.Earlier"); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("new malicious bytes"), 0o600); err != nil {
		t.Fatal(err)
	}
	spoolPath := filepath.Join(dir, "events.ndjson")
	sp, err := event.OpenSpool(spoolPath, 1<<20)
	if err != nil {
		t.Fatal(err)
	}
	defer sp.Close()
	s := New(config.Default(), c, sp, store, engine.New(config.Default()))
	s.sigDBVersion = func() string { return "generation-1" }
	s.scanSnapshot = snapshotForTest(s, func(_ context.Context, data []byte) (bool, string, error) {
		if err := os.Rename(path, filepath.Join(dir, "moved.exe")); err != nil {
			t.Fatal(err)
		}
		return false, "Test.Malicious", nil
	})
	v, sig, err := s.ScanFile(path, event.SourceProcessWatcher)
	b, _ := os.ReadFile(spoolPath)
	if v != cache.VerdictMalicious || sig != "Test.Malicious" || err != nil {
		t.Fatalf("verdict=%q sig=%q err=%v", v, sig, err)
	}
	if !strings.Contains(string(b), `"action":"detected"`) || !strings.Contains(string(b), "Test.Malicious") {
		t.Fatalf("new drop hidden by an unrelated earlier quarantine: %s", b)
	}
}

func TestSnapshotVersionDoesNotRaceReload(t *testing.T) {
	s, c, path, _ := scannerFixture(t, []byte("snapshot"))
	version := "generation-1"
	s.sigDBVersion = func() string { return version }
	s.scanSnapshot = func(context.Context, []byte) (bool, string, string, error) {
		version = "generation-2"
		return true, "", "generation-1", nil
	}
	if _, _, err := s.ScanFile(path, event.SourceEngine); err != nil {
		t.Fatal(err)
	}
	rec, found, err := c.Lookup(mustHash(t, path))
	if err != nil || !found || rec.SigDBVersion != "generation-1" {
		t.Fatalf("incorrect snapshot stamp: %+v found=%v err=%v", rec, found, err)
	}
}

func TestNativeEngineDetectsFileWithoutDaemon(t *testing.T) {
	s, c, path, q := scannerFixture(t, []byte("UTM harmless native test marker"))
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "local.ndb"), []byte("UTM.Test.Native:0:*:55544d206861726d6c657373206e61746976652074657374206d61726b6572\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	cfg := config.Default()
	cfg.SignatureDBDir = dir
	eng := engine.New(cfg)
	// The service loads the shared engine before scanning; scans never load it.
	if err := eng.EnsureRunning(); err != nil {
		t.Fatal(err)
	}
	s = New(cfg, c, nil, q, eng)
	v, sig, err := s.ScanFile(path, event.SourceEngine)
	if err != nil || v != cache.VerdictMalicious || sig != "UTM.Test.Native" || !q.called {
		t.Fatalf("native result: verdict=%s sig=%s err=%v quarantined=%v", v, sig, err, q.called)
	}
	rec, found, _ := c.Lookup(mustHash(t, path))
	if !found || rec.SigDBVersion == "" || rec.SigDBVersion != eng.Version() {
		t.Fatalf("native version not cached: %+v", rec)
	}
}

// writeBadDB leaves a rejected database in place and returns a func
// that replaces it with a valid rule matching content.
func writeBadDB(t *testing.T) (string, func(content string)) {
	t.Helper()
	db := t.TempDir()
	if err := os.WriteFile(filepath.Join(db, "bad.ndb"), []byte("Bad.Rule:0:*:zz\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	return db, func(content string) {
		rule := "UTM.Test:0:*:" + hex.EncodeToString([]byte(content)) + "\n"
		if err := os.WriteFile(filepath.Join(db, "bad.ndb"), []byte(rule), 0o600); err != nil {
			t.Fatal(err)
		}
	}
}

// While no generation is loaded, each scan must fail fast instead of re-reading
// and recompiling the whole signature directory under the engine lock.
// Engine.Supervise and explicit reloads own retries. Observable: after the
// initial failed load (service start), a later change to the loose database is
// not picked up by an ordinary file scan.
func TestScanDoesNotReloadSharedEngine(t *testing.T) {
	db, fix := writeBadDB(t)
	cfg := config.Default()
	cfg.SignatureDBDir = db
	eng := engine.New(cfg)
	if err := eng.EnsureRunning(); err == nil {
		t.Fatal("expected initial load failure")
	}
	s, _, target, q := scannerFixture(t, []byte("UTM harmless marker"))
	s = New(cfg, s.cache, nil, q, eng)
	fix("UTM harmless marker")
	for i := 0; i < 3; i++ {
		v, sig, err := s.ScanFile(target, event.SourceFileWatcher)
		if !errors.Is(err, engine.ErrUnavailable) || v == cache.VerdictMalicious || eng.Version() != "" {
			t.Fatalf("scan %d reloaded the signature directory: verdict=%q sig=%q err=%v version=%q", i, v, sig, err, eng.Version())
		}
	}
}

// Constructor and analysis never load even an explicitly supplied engine;
// only the service/CLI owner loads and supervises signature generations.
func TestConstructorNeverLoadsEngine(t *testing.T) {
	db, fix := writeBadDB(t)
	cfg := config.Default()
	cfg.SignatureDBDir = db
	s, _, target, q := scannerFixture(t, []byte("UTM harmless marker"))
	s = New(cfg, s.cache, nil, q, engine.New(cfg))
	if _, _, err := s.ScanFile(target, event.SourceEngine); !errors.Is(err, engine.ErrUnavailable) {
		t.Fatalf("first scan err=%v", err)
	}
	fix("UTM harmless marker")
	if v, _, err := s.ScanFile(target, event.SourceEngine); !errors.Is(err, engine.ErrUnavailable) || v == cache.VerdictMalicious {
		t.Fatalf("second scan reloaded the database: verdict=%q err=%v", v, err)
	}
}

// A down engine is one condition, not one per file: distinct inputs share a
// single Unavailable report per window instead of one spool record each.
func TestUnavailableEngineReportedOncePerWindow(t *testing.T) {
	s, _, _, _ := scannerFixture(t, nil)
	spoolPath := filepath.Join(t.TempDir(), "events.ndjson")
	sp, err := event.OpenSpool(spoolPath, 1<<20)
	if err != nil {
		t.Fatal(err)
	}
	defer sp.Close()
	s.spool = sp
	s.sigDBVersion = func() string { return "" }
	s.scanSnapshot = func(context.Context, []byte) (bool, string, string, error) {
		return false, "", "", engine.ErrUnavailable
	}
	dir := t.TempDir()
	for i := 0; i < 20; i++ {
		path := filepath.Join(dir, fmt.Sprintf("f%d.bin", i))
		if err := os.WriteFile(path, []byte(fmt.Sprint("distinct content ", i)), 0o600); err != nil {
			t.Fatal(err)
		}
		if _, _, err := s.ScanFile(path, event.SourceFileWatcher); !errors.Is(err, engine.ErrUnavailable) {
			t.Fatalf("err=%v", err)
		}
	}
	b, _ := os.ReadFile(spoolPath)
	if strings.Count(string(b), "\n") != 1 || !strings.Contains(string(b), "Analysis.Incomplete.Unavailable") {
		t.Fatalf("expected one Unavailable report for 20 files: %s", b)
	}
}

func TestIncompleteAnalysisReportsUnknownAndDeduplicates(t *testing.T) {
	s, c, path, q := scannerFixture(t, []byte("unsupported content"))
	spoolPath := filepath.Join(t.TempDir(), "events.ndjson")
	sp, err := event.OpenSpool(spoolPath, 1<<20)
	if err != nil {
		t.Fatal(err)
	}
	defer sp.Close()
	s.spool = sp
	version := "generation-1"
	s.sigDBVersion = func() string { return version }
	s.scanSnapshot = snapshotForTest(s, func(_ context.Context, _ []byte) (bool, string, error) { return false, "", engine.ErrUnsupported })
	for i := 0; i < 3; i++ {
		v, _, err := s.ScanFile(path, event.SourceFileWatcher)
		if err == nil || v != cache.VerdictUnknown {
			t.Fatalf("incomplete verdict=%s err=%v", v, err)
		}
	}
	if q.called {
		t.Fatal("incomplete analysis quarantined input")
	}
	if _, found, err := c.Lookup(mustHash(t, path)); found || err != nil {
		t.Fatalf("incomplete verdict cached: %v %v", found, err)
	}
	b, err := os.ReadFile(spoolPath)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(string(b), "\n") != 1 || !strings.Contains(string(b), `"verdict":"unknown"`) ||
		!strings.Contains(string(b), `"action":"health"`) || !strings.Contains(string(b), "Analysis.Incomplete.Unsupported") {
		t.Fatalf("incorrect incomplete health event: %s", b)
	}
	version = "generation-2"
	_, _, _ = s.ScanFile(path, event.SourceFileWatcher)
	b, _ = os.ReadFile(spoolPath)
	if strings.Count(string(b), "\n") != 2 {
		t.Fatal("new generation health suppressed")
	}
}

func TestIncompleteReportMemoryIsBounded(t *testing.T) {
	s, _, _, _ := scannerFixture(t, nil)
	sp, err := event.OpenSpool(filepath.Join(t.TempDir(), "events.ndjson"), 1<<20)
	if err != nil {
		t.Fatal(err)
	}
	defer sp.Close()
	s.spool = sp
	for i := 0; i < maxReports+10; i++ {
		s.noteIncomplete("sample", fmt.Sprint(i), "v1", event.SourceEngine, engine.ErrLimit)
	}
	if len(s.reported) != maxReports {
		t.Fatalf("unbounded report deduplication: %d", len(s.reported))
	}
}

type fakeQ struct{ called bool }

func (f *fakeQ) QuarantineVerified(path, hash, detection string, _ os.FileInfo) (string, error) {
	return f.Quarantine(path, hash, detection)
}
func (f *fakeQ) Quarantine(path, sha256, detection string) (string, error) {
	f.called = true
	return "qid", nil
}

func mustHash(t *testing.T, p string) string {
	t.Helper()
	h, err := SHA256File(p)
	if err != nil {
		t.Fatal(err)
	}
	return h
}

// The scanner must stamp the signature-DB version a verdict was produced under,
// so the feed only re-scans a cached entry when signatures actually change.
func TestScanFileStampsSigDBVersion(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "clean.bin")
	if err := os.WriteFile(target, []byte("harmless"), 0o644); err != nil {
		t.Fatal(err)
	}
	c, err := cache.Open(filepath.Join(dir, "edr.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	sp, err := event.OpenSpool(filepath.Join(dir, "events.ndjson"), 1<<20)
	if err != nil {
		t.Fatal(err)
	}
	defer sp.Close()

	s := New(config.Default(), c, sp, &fakeQ{}, engine.New(config.Default()))
	s.scanSnapshot = snapshotForTest(s, func(addr context.Context, data []byte) (bool, string, error) { return true, "", nil }) // clean
	s.sigDBVersion = func() string { return "27000" }

	if _, _, err := s.ScanFile(target, event.SourceEngine); err != nil {
		t.Fatalf("ScanFile: %v", err)
	}
	rec, found, _ := c.Lookup(mustHash(t, target))
	if !found || rec.SigDBVersion != "27000" {
		t.Fatalf("record not stamped with sig-db version: %+v (found=%v)", rec, found)
	}

	// A record stamped with the *current* version must NOT be re-invalidated.
	if n, _ := c.MarkStaleBySigDB("27000"); n != 0 {
		t.Fatalf("stamped record invalidated at same version (n=%d) — over-invalidation not fixed", n)
	}
	// It IS invalidated when the version actually changes.
	if n, _ := c.MarkStaleBySigDB("28000"); n != 1 {
		t.Fatalf("stamped record not invalidated on version change (n=%d)", n)
	}
}

// The verdict cache is an optimization. A failing cache database must neither
// stop analysis nor drop a confirmed malicious verdict and its response.
func TestCacheFailureDoesNotLoseMaliciousVerdict(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "a.exe")
	if err := os.WriteFile(target, []byte("payload"), 0o644); err != nil {
		t.Fatal(err)
	}
	c, err := cache.Open(filepath.Join(dir, "edr.db"))
	if err != nil {
		t.Fatal(err)
	}
	sp, err := event.OpenSpool(filepath.Join(dir, "events.ndjson"), 1<<20)
	if err != nil {
		t.Fatal(err)
	}
	defer sp.Close()
	fq := &fakeQ{}
	s := New(config.Default(), c, sp, fq, engine.New(config.Default()))
	scanned := false
	s.scanSnapshot = snapshotForTest(s, func(addr context.Context, data []byte) (bool, string, error) {
		scanned = true
		return false, "Win.Test.Cache", nil
	})
	_ = c.Close() // every lookup and store now fails
	verdict, sig, err := s.ScanFile(target, event.SourceEngine)
	if err != nil || verdict != cache.VerdictMalicious || sig != "Win.Test.Cache" {
		t.Fatalf("verdict=%q sig=%q err=%v", verdict, sig, err)
	}
	if !scanned || !fq.called {
		t.Fatalf("analysis or quarantine skipped: scanned=%v quarantined=%v", scanned, fq.called)
	}
}

// storeFixture wires a scanner to a real quarantine store and event spool.
func storeFixture(t *testing.T) (s *Scanner, dir, qdir, spoolPath string) {
	t.Helper()
	dir = t.TempDir()
	c, err := cache.Open(filepath.Join(dir, "edr.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = c.Close() })
	qdir = filepath.Join(dir, "quarantine")
	store, err := quarantine.New(qdir, c)
	if err != nil {
		t.Fatal(err)
	}
	spoolPath = filepath.Join(dir, "events.ndjson")
	sp, err := event.OpenSpool(spoolPath, 1<<20)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sp.Close() })
	s = New(config.Default(), c, sp, store, engine.New(config.Default()))
	s.sigDBVersion = func() string { return "generation-1" }
	return s, dir, qdir, spoolPath
}

// heldFile writes a file into its own directory and makes that directory
// read-only: the file can be read and copied but not renamed or removed, the
// way Windows treats a running image or a log another process holds open. The
// returned func releases it.
func heldFile(t *testing.T, dir, name, content string) (string, func()) {
	t.Helper()
	if runtime.GOOS == "windows" || os.Geteuid() == 0 {
		t.Skip("needs POSIX directory permissions enforced for a non-root user")
	}
	held := filepath.Join(dir, "held")
	if err := os.Mkdir(held, 0o700); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(held, name)
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(held, 0o500); err != nil {
		t.Fatal(err)
	}
	release := func() { _ = os.Chmod(held, 0o700) }
	t.Cleanup(release)
	return path, release
}

func spoolLines(t *testing.T, path string) []string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		t.Fatal(err)
	}
	if s := strings.TrimSpace(string(b)); s != "" {
		return strings.Split(s, "\n")
	}
	return nil
}

func dirNames(t *testing.T, dir string) []string {
	t.Helper()
	ents, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, e := range ents {
		if strings.HasPrefix(e.Name(), ".item-lock-") && len(e.Name()) == len(".item-lock-")+2 {
			continue
		}
		if e.Name() == ".admissions" {
			pending, readErr := os.ReadDir(filepath.Join(dir, e.Name()))
			if readErr == nil && len(pending) == 0 {
				continue
			}
		}

		names = append(names, e.Name())
	}
	return names
}

type quarantinerFunc func(path, sha256, detection string) (string, error)

func (f quarantinerFunc) Quarantine(path, hash, detection string) (string, error) {
	return f(path, hash, detection)
}
func (f quarantinerFunc) QuarantineVerified(path, sha256, detection string, _ os.FileInfo) (string, error) {
	return f(path, sha256, detection)
}

// A detection in a file that cannot be moved (a log another process holds
// open, re-detected on every write) is reported as detected once per window
// and leaves nothing in the store. Containment is never suppressed: once the
// file can be moved it is quarantined and reported.
func TestUncontainableDetectionReportedOnceWithoutStoreCopies(t *testing.T) {
	s, dir, qdir, spoolPath := storeFixture(t)
	path, release := heldFile(t, dir, "operational.evtx", "record")
	s.scanSnapshot = snapshotForTest(s, func(_ context.Context, _ []byte) (bool, string, error) { return false, "Test.Marker", nil })
	for i := 0; i < 3; i++ {
		// Every write changes the content, and so the hash.
		if err := os.WriteFile(path, []byte(fmt.Sprint("record ", i)), 0o600); err != nil {
			t.Fatal(err)
		}
		if v, _, err := s.ScanFile(path, event.SourceFileWatcher); v != cache.VerdictMalicious || err != nil {
			t.Fatalf("write %d: verdict=%q err=%v", i, v, err)
		}
	}
	lines := spoolLines(t, spoolPath)
	if len(lines) != 1 || !strings.Contains(lines[0], `"action":"detected"`) {
		t.Fatalf("want one detected event for a file that cannot be contained, got %d: %v", len(lines), lines)
	}
	if left := dirNames(t, qdir); len(left) != 0 {
		t.Fatalf("store holds copies without a record: %v", left)
	}
	if b, err := os.ReadFile(path); err != nil || string(b) != "record 2" {
		t.Fatalf("file must stay in place: %q err=%v", b, err)
	}

	release()
	if err := os.WriteFile(path, []byte("record 3"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, _, err := s.ScanFile(path, event.SourceFileWatcher); err != nil {
		t.Fatal(err)
	}
	lines = spoolLines(t, spoolPath)
	if len(lines) != 2 || !strings.Contains(lines[1], `"action":"quarantined"`) {
		t.Fatalf("containment after release not reported: %v", lines)
	}
	if left := dirNames(t, qdir); len(left) != 1 {
		t.Fatalf("store = %v, want the one quarantined file", left)
	}
}

// A program that starts a copy of itself makes the launch guard scan one image
// twice at once. The scans share one analysis and one response: one engine
// pass, one quarantine and one event.
func TestConcurrentLaunchScansOfOneImageShareOneResponse(t *testing.T) {
	s, dir, qdir, spoolPath := storeFixture(t)
	path := filepath.Join(dir, "spawner.exe")
	if err := os.WriteFile(path, []byte("self spawning image"), 0o600); err != nil {
		t.Fatal(err)
	}
	entered, release := make(chan struct{}), make(chan struct{})
	var calls atomic.Int32
	s.scanSnapshot = snapshotForTest(s, func(_ context.Context, _ []byte) (bool, string, error) {
		if calls.Add(1) == 1 {
			close(entered)
		}
		<-release
		return false, "Test.Spawner", nil
	})
	type result struct {
		verdict, sig string
		contain      func() string
		err          error
	}
	results := make(chan result, 2)
	scan := func() {
		v, sig, contain, err := s.ScanImage(path, event.SourceProcessWatcher)
		results <- result{v, sig, contain, err}
	}
	go scan()
	<-entered
	go scan()
	deadline := time.Now().Add(5 * time.Second)
	for !s.sharedBy(path, 2) {
		if time.Now().After(deadline) {
			t.Fatal("second scan of the same image did not join the first")
		}
		time.Sleep(time.Millisecond)
	}
	close(release)
	for i := 0; i < 2; i++ {
		r := <-results
		if r.verdict != cache.VerdictMalicious || r.sig != "Test.Spawner" || r.err != nil || r.contain == nil {
			t.Fatalf("scan %d: verdict=%q sig=%q err=%v contain=%v", i, r.verdict, r.sig, r.err, r.contain != nil)
		}
		if a := r.contain(); a != event.ActionQuarantined {
			t.Fatalf("scan %d: action %q, want quarantined", i, a)
		}
	}
	if n := calls.Load(); n != 1 {
		t.Fatalf("image analysed %d times, want once", n)
	}
	if lines := spoolLines(t, spoolPath); len(lines) != 1 || !strings.Contains(lines[0], `"action":"quarantined"`) {
		t.Fatalf("want one quarantined event, got %v", lines)
	}
	if left := dirNames(t, qdir); len(left) != 1 {
		t.Fatalf("store = %v, want one quarantined file", left)
	}
}

// A running image (on another volume, or held open) cannot be removed until
// its process has exited, so the launch guard contains it only after the kill.
// The attempt made before the process is fully gone fails without leaving a
// copy, the retry quarantines it, and only then is the single event written,
// with the action that happened.
func TestLaunchImageQuarantinedByRetryAfterKill(t *testing.T) {
	s, dir, qdir, spoolPath := storeFixture(t)
	path, release := heldFile(t, dir, "running.exe", "running image")
	s.scanSnapshot = snapshotForTest(s, func(_ context.Context, _ []byte) (bool, string, error) { return false, "Test.Running", nil })
	s.containRetry = []time.Duration{0, time.Millisecond, time.Millisecond}
	store := s.quarantiner
	attempts := 0
	s.quarantiner = quarantinerFunc(func(p, h, d string) (string, error) {
		attempts++
		id, err := store.Quarantine(p, h, d)
		if err != nil {
			release() // the terminated process lets go of its image
		}
		return id, err
	})
	v, sig, contain, err := s.ScanImage(path, event.SourceProcessWatcher)
	if v != cache.VerdictMalicious || sig != "Test.Running" || err != nil || contain == nil {
		t.Fatalf("verdict=%q sig=%q err=%v contain=%v", v, sig, err, contain != nil)
	}
	if lines := spoolLines(t, spoolPath); len(lines) != 0 {
		t.Fatalf("reported before the process was terminated: %v", lines)
	}
	if a := contain(); a != event.ActionQuarantined {
		t.Fatalf("action %q, want quarantined", a)
	}
	if attempts != 2 {
		t.Fatalf("%d containment attempts, want 2", attempts)
	}
	if lines := spoolLines(t, spoolPath); len(lines) != 1 || !strings.Contains(lines[0], `"action":"quarantined"`) {
		t.Fatalf("want one quarantined event, got %v", lines)
	}
	if left := dirNames(t, qdir); len(left) != 1 {
		t.Fatalf("store = %v, want exactly the quarantined image", left)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatal("image still in place after quarantine")
	}
	// Every sharing launch calls contain; the response happens once.
	if a := contain(); a != event.ActionQuarantined || attempts != 2 || len(spoolLines(t, spoolPath)) != 1 {
		t.Fatalf("second contain acted again: action=%q attempts=%d", a, attempts)
	}
}

// Retries are bounded. An image that never becomes movable is reported once as
// detected (not contained) and the store holds nothing for it.
func TestLaunchImageLeftInPlaceAfterBoundedRetries(t *testing.T) {
	s, dir, qdir, spoolPath := storeFixture(t)
	path, _ := heldFile(t, dir, "protected.exe", "protected image")
	s.scanSnapshot = snapshotForTest(s, func(_ context.Context, _ []byte) (bool, string, error) { return false, "Test.Protected", nil })
	s.containRetry = []time.Duration{0, time.Millisecond, 2 * time.Millisecond}
	store := s.quarantiner
	attempts := 0
	s.quarantiner = quarantinerFunc(func(p, h, d string) (string, error) {
		attempts++
		return store.Quarantine(p, h, d)
	})
	_, _, contain, err := s.ScanImage(path, event.SourceProcessWatcher)
	if err != nil || contain == nil {
		t.Fatalf("err=%v contain=%v", err, contain != nil)
	}
	if a := contain(); a != event.ActionDetected {
		t.Fatalf("action %q, want detected", a)
	}
	if attempts != 3 {
		t.Fatalf("%d attempts, want 3", attempts)
	}
	if lines := spoolLines(t, spoolPath); len(lines) != 1 || !strings.Contains(lines[0], `"action":"detected"`) {
		t.Fatalf("want one detected event, got %v", lines)
	}
	if left := dirNames(t, qdir); len(left) != 0 {
		t.Fatalf("store holds copies without a record: %v", left)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("image must stay in place: %v", err)
	}
}

// When the file watcher contains and reports the image while the guard is
// still killing the process, the guard's containment adds no second event.
func TestLaunchContainmentSkipsImageAlreadyReported(t *testing.T) {
	s, dir, qdir, spoolPath := storeFixture(t)
	path := filepath.Join(dir, "dropped.exe")
	if err := os.WriteFile(path, []byte("dropped and run"), 0o600); err != nil {
		t.Fatal(err)
	}
	s.scanSnapshot = snapshotForTest(s, func(_ context.Context, _ []byte) (bool, string, error) { return false, "Test.Dropped", nil })
	_, _, contain, err := s.ScanImage(path, event.SourceProcessWatcher)
	if err != nil || contain == nil {
		t.Fatalf("err=%v contain=%v", err, contain != nil)
	}
	if _, _, err := s.ScanFile(path, event.SourceFileWatcher); err != nil {
		t.Fatal(err)
	}
	if a := contain(); a != "" {
		t.Fatalf("guard reported an already reported quarantine as %q", a)
	}
	if lines := spoolLines(t, spoolPath); len(lines) != 1 || !strings.Contains(lines[0], `"source":"file_watcher"`) {
		t.Fatalf("want only the file watcher's event, got %v", lines)
	}
	if left := dirNames(t, qdir); len(left) != 1 {
		t.Fatalf("store = %v", left)
	}
}

// pauseFirstQuarantine makes the first quarantine move the file into qdir,
// close moved, and record the move only after a pause: the window in which
// another scan finds the file gone with no record yet. Later calls use the
// real store.
func pauseFirstQuarantine(s *Scanner, qdir string, moved chan<- struct{}) {
	store := s.quarantiner
	var started atomic.Bool
	s.quarantiner = quarantinerFunc(func(p, h, d string) (string, error) {
		if !started.CompareAndSwap(false, true) {
			return store.Quarantine(p, h, d)
		}
		id := "paused"
		if err := os.Rename(p, filepath.Join(qdir, id+".quarantined")); err != nil {
			return "", err
		}
		close(moved)
		time.Sleep(100 * time.Millisecond)
		return id, s.cache.StoreQuarantine(cache.QuarantineRecord{QuarantineID: id, OriginalPath: p,
			SHA256: h, Detection: d, Restorable: true})
	})
}

// Two launches of one image whose analyses did not overlap each hold a
// containment. They must not interleave: the second sees the first's move
// together with its record and adds no event, rather than reporting the image
// as vanished (a duplicate detected event beside the quarantined one).
func TestSeparateLaunchContainmentsOfOneImageReportOnce(t *testing.T) {
	s, dir, qdir, spoolPath := storeFixture(t)
	path := filepath.Join(dir, "spawner.exe")
	if err := os.WriteFile(path, []byte("self spawning image"), 0o600); err != nil {
		t.Fatal(err)
	}
	s.scanSnapshot = snapshotForTest(s, func(_ context.Context, _ []byte) (bool, string, error) { return false, "Test.Spawner", nil })
	moved := make(chan struct{})
	pauseFirstQuarantine(s, qdir, moved)
	_, _, contain1, err1 := s.ScanImage(path, event.SourceProcessWatcher)
	_, _, contain2, err2 := s.ScanImage(path, event.SourceProcessWatcher)
	if err1 != nil || err2 != nil || contain1 == nil || contain2 == nil {
		t.Fatalf("err1=%v err2=%v contain1=%v contain2=%v", err1, err2, contain1 != nil, contain2 != nil)
	}
	first := make(chan string, 1)
	go func() { first <- contain1() }()
	<-moved
	if a := contain2(); a != "" {
		t.Fatalf("second containment reported %q while the first was moving the image", a)
	}
	if a := <-first; a != event.ActionQuarantined {
		t.Fatalf("first containment: action %q, want quarantined", a)
	}
	if lines := spoolLines(t, spoolPath); len(lines) != 1 || !strings.Contains(lines[0], `"action":"quarantined"`) {
		t.Fatalf("want one quarantined event, got %v", lines)
	}
	if left := dirNames(t, qdir); len(left) != 1 {
		t.Fatalf("store = %v, want one quarantined file", left)
	}
}

// A launch scanned while another scan is moving its image into quarantine must
// still get the malicious verdict so the guard kills it, although the image is
// already gone and its move not yet recorded.
func TestLaunchScanDuringQuarantineStillDetects(t *testing.T) {
	s, dir, qdir, spoolPath := storeFixture(t)
	path := filepath.Join(dir, "relaunched.exe")
	if err := os.WriteFile(path, []byte("relaunched image"), 0o600); err != nil {
		t.Fatal(err)
	}
	s.scanSnapshot = snapshotForTest(s, func(_ context.Context, _ []byte) (bool, string, error) { return false, "Test.Relaunch", nil })
	moved := make(chan struct{})
	pauseFirstQuarantine(s, qdir, moved)
	_, _, contain, err := s.ScanImage(path, event.SourceProcessWatcher)
	if err != nil || contain == nil {
		t.Fatalf("err=%v contain=%v", err, contain != nil)
	}
	first := make(chan string, 1)
	go func() { first <- contain() }()
	<-moved
	v, sig, contain2, err := s.ScanImage(path, event.SourceProcessWatcher)
	if v != cache.VerdictMalicious || sig != "Test.Relaunch" || err != nil {
		t.Fatalf("verdict=%q sig=%q err=%v; the launched process would not be killed", v, sig, err)
	}
	if contain2 != nil {
		t.Fatal("an image already being contained needs no second containment")
	}
	if a := <-first; a != event.ActionQuarantined {
		t.Fatalf("action %q, want quarantined", a)
	}
	if lines := spoolLines(t, spoolPath); len(lines) != 1 {
		t.Fatalf("want one event, got %v", lines)
	}
}

// sharedBy reports whether n launch scans currently share path's analysis.
func (s *Scanner) sharedBy(path string, n int) bool {
	s.flightMu.Lock()
	defer s.flightMu.Unlock()
	f := s.flights[pathKey(path)]
	return f != nil && f.callers == n
}

func TestIncompleteReportNamesAnalysisNotPerformed(t *testing.T) {
	s, _, path, _ := scannerFixture(t, []byte("html content"))
	spoolPath := filepath.Join(t.TempDir(), "events.ndjson")
	sp, err := event.OpenSpool(spoolPath, 1<<20)
	if err != nil {
		t.Fatal(err)
	}
	defer sp.Close()
	s.spool = sp
	s.sigDBVersion = func() string { return "generation-1" }
	cause := fmt.Errorf("%w: %w", engine.ErrUnsupported, &signatures.UnknownError{Reason: signatures.UnknownHTML})
	s.scanSnapshot = snapshotForTest(s, func(_ context.Context, _ []byte) (bool, string, error) { return false, "", cause })
	if v, _, err := s.ScanFile(path, event.SourceFileWatcher); err == nil || v != cache.VerdictUnknown {
		t.Fatalf("verdict=%s err=%v", v, err)
	}
	b, _ := os.ReadFile(spoolPath)
	if !strings.Contains(string(b), "Analysis.Incomplete.Unsupported.html-normalization") {
		t.Fatalf("event lacks the analysis not performed: %s", b)
	}
}
