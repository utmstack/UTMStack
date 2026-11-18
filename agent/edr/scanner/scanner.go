package scanner

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"hash/fnv"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/utmstack/UTMStack/agent/edr/cache"
	"github.com/utmstack/UTMStack/agent/edr/config"
	"github.com/utmstack/UTMStack/agent/edr/engine"
	"github.com/utmstack/UTMStack/agent/edr/event"
	"github.com/utmstack/UTMStack/agent/edr/quarantine"
	"github.com/utmstack/UTMStack/shared/logger"
)

var errTargetChanged = errors.New("scan target changed")

// maxReports bounds the report deduplication memory; reportWindow is how long
// an identical report is suppressed.
const (
	maxReports   = 1024
	reportWindow = 10 * time.Minute
)

// defaultContainRetry spreads the post-kill quarantine attempts over about 8
// seconds, long enough for a terminated process to release its image.
var defaultContainRetry = []time.Duration{0, 100 * time.Millisecond, 250 * time.Millisecond,
	500 * time.Millisecond, time.Second, 2 * time.Second, 4 * time.Second}

// Quarantiner contains a malicious file. Injected so the scanner stays testable.
type Quarantiner interface {
	Quarantine(path, sha256, detection string) (string, error)
	QuarantineVerified(path, sha256, detection string, scanned os.FileInfo) (string, error)
}

type Scanner struct {
	cfg                  config.EDRConfig
	cache                *cache.Cache
	spool                *event.Spool
	quarantiner          Quarantiner
	scanSnapshot         func(context.Context, []byte) (bool, string, string, error)
	scanSnapshotWithPath func(context.Context, []byte, string) (bool, string, string, error)
	metadataSensitive    func() bool
	maxFileBytes         int64
	openTarget           func(string) (*os.File, error)
	slots                chan struct{}
	reportMu             sync.Mutex
	reported             map[[32]byte]time.Time
	// sigDBVersion supplies the active generation for cache lookup. A newly
	// scanned verdict is stamped with the generation returned by ScanSnapshot.
	sigDBVersion func() string
	// flights coalesces concurrent launch-guard scans of one image (a program
	// that starts copies of itself) into one analysis and one response.
	flightMu sync.Mutex
	flights  map[string]*flight
	// containRetry is the wait before each attempt to quarantine a launched
	// image after its process tree was killed: a running image cannot be moved
	// until the process has fully exited.
	containRetry []time.Duration
	// containMu serialises containment by path; see containLock.
	containMu [64]sync.Mutex
}

// flight is one launch-guard analysis shared by every concurrent launch of the
// same image. Its results are written before done is closed.
type flight struct {
	done         chan struct{}
	callers      int // launches sharing it
	verdict, sig string
	contain      func() string
	err          error
}

// New uses the owner's shared engine. The service or CLI loads and supervises
// it explicitly; constructing a scanner never creates or loads another engine.
func New(cfg config.EDRConfig, c *cache.Cache, sp *event.Spool, q Quarantiner, eng *engine.Engine) *Scanner {
	if eng == nil {
		panic("scanner.New requires a shared native engine")
	}

	tuning := eng.Tuning()
	maxFileBytes := int64(tuning.MaxFileSizeMB) << 20
	if maxFileBytes <= 0 {
		maxFileBytes = 25 << 20
	}
	workers := max(1, min(16, tuning.MaxThreads))
	if cfg.ScanConcurrency > 0 {
		workers = min(workers, cfg.ScanConcurrency)
	}
	return &Scanner{
		cfg: cfg, cache: c, spool: sp, quarantiner: q,
		maxFileBytes:         maxFileBytes,
		slots:                make(chan struct{}, workers),
		reported:             make(map[[32]byte]time.Time),
		flights:              make(map[string]*flight),
		containRetry:         defaultContainRetry,
		sigDBVersion:         eng.Version,
		scanSnapshot:         eng.ScanSnapshot,
		scanSnapshotWithPath: eng.ScanSnapshotWithPath,
		metadataSensitive:    eng.MetadataSensitive,
	}
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

// ScanFile reads a bounded snapshot, hashes it, checks the cache, and scans the
// same bytes through the native engine on a miss,
// stores the verdict, and — on a malicious verdict (cached OR freshly scanned) —
// quarantines the file when it still holds the analyzed bytes and emits a
// branded detection event.
func (s *Scanner) ScanFile(path, source string) (string, string, error) {
	verdict, sig, hash, identity, err := s.admitAndAnalyze(path, source)
	if verdict != cache.VerdictMalicious || hash == "" {
		return verdict, sig, err
	}
	// Act on a malicious verdict for THIS file even when the verdict came from the
	// cache: a re-dropped known-malicious file (same hash as one seen before) must
	// still be quarantined and reported, not silently left in place. The verdict
	// cache exists to skip the engine round-trip, not the response. An already-
	// quarantined file is gone, so its re-scan fails hashing above and never reaches
	// here — no double action.
	r, qerr := s.tryContain(path, hash, sig, identity)
	_, err = s.report(path, hash, sig, source, r, qerr)
	return verdict, sig, err
}

// ScanImage analyses the image of a newly launched process for the launch
// guard. Concurrent calls for one image (a program that starts copies of
// itself) share a single analysis and a single response. On a malicious verdict
// nothing is moved or reported yet: a running image usually cannot be moved, so
// the guard terminates the process tree first and then calls contain, which
// retries the quarantine with backoff and writes the one detection event with
// the action that happened. Every sharing caller may call contain; the response
// runs once and later calls return its action. contain is nil when no response
// is due.
func (s *Scanner) ScanImage(path, source string) (verdict, sig string, contain func() string, err error) {
	key := pathKey(path)
	s.flightMu.Lock()
	if f, ok := s.flights[key]; ok {
		f.callers++
		s.flightMu.Unlock()
		<-f.done
		return f.verdict, f.sig, f.contain, f.err
	}
	f := &flight{done: make(chan struct{}), callers: 1}
	s.flights[key] = f
	s.flightMu.Unlock()
	defer func() {
		s.flightMu.Lock()
		delete(s.flights, key)
		s.flightMu.Unlock()
		close(f.done)
	}()

	var hash string
	var identity os.FileInfo
	f.verdict, f.sig, hash, identity, f.err = s.admitAndAnalyze(path, source)
	if f.verdict == cache.VerdictMalicious && hash != "" {
		f.contain = s.containAfterKill(path, hash, f.sig, source, identity)
	}
	return f.verdict, f.sig, f.contain, f.err
}

// admitAndAnalyze runs analyze under the shared admission bound.
func (s *Scanner) admitAndAnalyze(path, source string) (string, string, string, os.FileInfo, error) {
	// Launch-guard requests share this admission bound with file watchers so
	// parked process scans cannot each allocate a full input buffer at once.
	s.slots <- struct{}{}
	defer func() { <-s.slots }()
	return s.analyze(path, source)
}

// analyze returns the verdict for the bytes path holds now, and their hash when
// a response is still due.
func (s *Scanner) analyze(path, source string) (verdict, sig, hash string, identity os.FileInfo, err error) {
	data, identity, err := readBoundedSnapshot(path, s.maxFileBytes, s.openTarget)
	if err != nil {
		// The image is unreadable — commonly because the file-watcher path just
		// quarantined (moved) it out from under a still-running process. If it was
		// quarantined as malicious moments ago, report that verdict so the caller
		// (the process guard) can still terminate the process. Don't re-quarantine
		// or re-emit here — that already happened on the original detection.
		if os.IsNotExist(err) {
			// Wait out a move into quarantine still in progress: it is
			// recorded before its lock is released.
			mu := s.containLock(path)
			mu.Lock()
			rec, ok, qerr := s.recentQuarantine(path)
			mu.Unlock()
			if qerr == nil && ok {
				return cache.VerdictMalicious, rec.Detection, "", nil, nil
			}
		}
		if errors.Is(err, engine.ErrLimit) || errors.Is(err, errTargetChanged) {
			s.noteIncomplete(path, "", s.sigDBVersion(), source, err)
		}
		return "", "", "", nil, err
	}
	sum := sha256.Sum256(data)
	hash = hex.EncodeToString(sum[:])
	currentVersion := s.sigDBVersion()

	// The verdict cache only saves engine work: a failing cache database is
	// treated as a miss and must never stop analysis or response.
	metadataActive := s.metadataSensitive != nil && s.metadataSensitive()
	if !metadataActive {
		rec, found, lookupErr := s.cache.Lookup(hash)
		if lookupErr != nil {
			logger.Error("UTMStack EDR: verdict cache lookup failed; analyzing without it: %v", lookupErr)
			found = false
		}
		if found && !rec.Stale && currentVersion != "" && rec.SigDBVersion == currentVersion &&
			(rec.Verdict == cache.VerdictClean || rec.Verdict == cache.VerdictMalicious) &&
			(s.metadataSensitive == nil || !s.metadataSensitive()) && s.sigDBVersion() == currentVersion {
			return rec.Verdict, rec.Signature, hash, identity, nil
		}
	}
	var clean bool
	var s2, usedVersion string
	metadataActive = metadataActive || s.metadataSensitive != nil && s.metadataSensitive()
	if metadataActive {
		if s.scanSnapshotWithPath == nil {
			err = engine.ErrUnsupported
		} else {
			clean, s2, usedVersion, err = s.scanSnapshotWithPath(context.Background(), data, path)
		}
	} else {
		clean, s2, usedVersion, err = s.scanSnapshot(context.Background(), data)
	}

	if err != nil {
		s.noteIncomplete(path, hash, usedVersion, source, err)
		return cache.VerdictUnknown, "", "", nil, err
	}
	verdict = cache.VerdictClean
	if !clean {
		verdict = cache.VerdictMalicious
	}
	if !metadataActive && (s.metadataSensitive == nil || !s.metadataSensitive()) {
		if err := s.cache.Store(cache.VerdictRecord{SHA256: hash, Verdict: verdict, Signature: s2, SigDBVersion: usedVersion}); err != nil {
			logger.Error("UTMStack EDR: verdict cache store failed; keeping the %s verdict: %v", verdict, err)
		}
	}
	return verdict, s2, hash, identity, nil
}

// containAfterKill returns the launch guard's containment for one confirmed
// detection. The first call makes one attempt per containRetry entry, waiting
// that long before it, until the image is moved or cannot be moved by path at
// all, then reports once. Later calls wait for it and return the same action.
func (s *Scanner) containAfterKill(path, hash, sig, source string, identity os.FileInfo) func() string {
	var once sync.Once
	var action string
	return func() string {
		once.Do(func() {
			r, qerr := containSkipped, error(nil)
			for _, wait := range s.containRetry {
				time.Sleep(wait)
				if r, qerr = s.tryContain(path, hash, sig, identity); r != containFailed {
					break
				}
			}
			var err error
			if action, err = s.report(path, hash, sig, source, r, qerr); err != nil {
				logger.Error("UTMStack EDR: detection event for %s not written: %v", path, err)
			}
		})
		return action
	}
}

// containResult is the outcome of one attempt to quarantine a detection.
type containResult int

const (
	containMoved    containResult = iota // moved into quarantine by this attempt
	containReported                      // a concurrent scan already moved and reported these bytes
	containSkipped                       // nothing to move: no store, or the target changed or vanished
	containFailed                        // the move failed; it may succeed once the file is released
)

// tryContain checks the original scan identity, then delegates content proof
// and namespace mutation to the store's pinned admission operation. A changed
// target skips containment without changing the confirmed malware verdict.
func (s *Scanner) tryContain(path, hash, sig string, identity os.FileInfo) (containResult, error) {
	if s.quarantiner == nil {
		return containSkipped, nil
	}
	mu := s.containLock(path)
	mu.Lock()
	defer mu.Unlock()
	current, err := os.Stat(path)
	switch {
	case os.IsNotExist(err) && s.recentlyQuarantined(path, hash):
		return containReported, nil
	case os.IsNotExist(err), err == nil && (identity == nil || !os.SameFile(identity, current) || identity.Size() != current.Size() || !identity.ModTime().Equal(current.ModTime())):
		return containSkipped, nil
	case err != nil:
		return containFailed, err
	}
	if _, err = s.quarantiner.QuarantineVerified(path, hash, sig, identity); err != nil {
		var contained *quarantine.ContainmentError
		if errors.As(err, &contained) {
			return containMoved, err
		}
		if errors.Is(err, quarantine.ErrAlreadyQuarantined) {
			return containReported, nil
		}
		if errors.Is(err, quarantine.ErrAdmissionChanged) {
			return containSkipped, nil
		}
		if _, serr := os.Stat(path); os.IsNotExist(serr) && s.recentlyQuarantined(path, hash) {
			return containReported, nil
		}
		return containFailed, err
	}

	return containMoved, nil
}

// report writes the one detection event for the outcome of containment and
// returns its action, or "" when a concurrent scan already reported it. A
// detection that stays in place is reported as detected, and at most once per
// window for a path and signature: a file that cannot be contained (a log
// another process holds open) is re-detected on every write, and every write
// changes its hash. A quarantine is always reported.
func (s *Scanner) report(path, hash, sig, source string, r containResult, qerr error) (string, error) {
	if r == containReported {
		return "", nil
	}
	ev := event.NewDetection(path, hash, sig, source)
	var key [32]byte
	var at time.Time
	if r == containMoved {
		ev.Action = event.ActionQuarantined
		if qerr != nil {
			ev.ResponseStatus = "recovery_pending"
			logger.Error("UTMStack EDR: file durably contained; finalization needs recovery: %v", qerr)
		}
	} else {
		key = sha256.Sum256([]byte("detection\x00" + pathKey(path) + "\x00" + sig))
		at = time.Now()
		if !s.admitReport(key, at) {
			return ev.Action, nil
		}
		if r == containFailed {
			logger.Error("UTMStack EDR: detection at %s could not be quarantined and stays in place: %v", path, qerr)
		}
	}
	js, err := ev.ToJSON()
	if err == nil && s.spool != nil {
		err = s.spool.Append(js)
	}
	if err != nil && !at.IsZero() {
		s.retractReport(key, at) // not delivered; let a later detection retry
	}
	return ev.Action, err
}

func (s *Scanner) recentQuarantine(path string) (cache.QuarantineRecord, bool, error) {
	if q, ok := s.quarantiner.(interface {
		RecentQuarantineByPath(string, time.Duration) (cache.QuarantineRecord, bool, error)
	}); ok {
		return q.RecentQuarantineByPath(path, 5*time.Minute)
	}
	return s.cache.RecentQuarantineByPath(path, 5*time.Minute)
}

// recentlyQuarantined reports whether these exact bytes were moved out of path
// moments ago by another scan.
func (s *Scanner) recentlyQuarantined(path, hash string) bool {
	rec, ok, err := s.recentQuarantine(path)
	return err == nil && ok && rec.SHA256 == hash
}

// containLock returns the lock that serialises containment of path. While one
// scan moves a file into quarantine and records the move, another scan of that
// path cannot find the file gone without a record (and report it as vanished,
// or lose the verdict for a process still running it), and its revalidation
// read or copy cannot hold the file open, which on Windows makes the move fail.
// Paths share a fixed set of locks, so memory stays bounded.
func (s *Scanner) containLock(path string) *sync.Mutex {
	h := fnv.New32a()
	_, _ = h.Write([]byte(pathKey(path)))
	return &s.containMu[h.Sum32()%uint32(len(s.containMu))]
}

// pathKey folds the spellings of one path together; Windows paths are
// case-insensitive.
func pathKey(path string) string {
	p := filepath.Clean(path)
	if runtime.GOOS == "windows" {
		p = strings.ToLower(p)
	}
	return p
}

// noteIncomplete surfaces coverage gaps without treating them as detections.
// Repeated attempts of identical content under one generation/reason emit at
// most once per ten minutes. Bounded deduplication prevents an untrusted stream
// of unique inputs from growing service memory without limit.
func (s *Scanner) noteIncomplete(path, hash, version, source string, err error) {
	if s.spool == nil {
		return
	}
	reason := engine.IncompleteReason(err)
	if reason == "Error" && errors.Is(err, errTargetChanged) {
		reason = "TargetChanged"
	}
	identity := hash
	if identity == "" {
		identity = path // oversized/unstable inputs have no verified content hash
	}
	if reason == "Unavailable" {
		// No generation is loaded: one engine-wide condition, not a property of
		// this input. Report it once per window rather than once per file.
		identity, path, hash = "", "", ""
	}
	key := sha256.Sum256([]byte(identity + "\x00" + version + "\x00" + reason))
	now := time.Now()
	if !s.admitReport(key, now) {
		return
	}
	ev := event.Event{Source: source, Action: event.ActionHealth, ObjectPath: path,
		FileHash: hash, Verdict: cache.VerdictUnknown, Signature: "Analysis.Incomplete." + reason, Severity: "medium"}
	js, marshalErr := ev.ToJSON()
	if marshalErr != nil || s.spool.Append(js) != nil {
		// A failed spool write is not a delivered report; allow a future retry.
		s.retractReport(key, now)
	}
}

// admitReport reports whether an event keyed by key may be written at now: at
// most one per reportWindow. The memory is bounded (oldest entry evicted) so an
// untrusted stream of unique inputs cannot grow service memory without limit.
func (s *Scanner) admitReport(key [32]byte, now time.Time) bool {
	s.reportMu.Lock()
	defer s.reportMu.Unlock()
	if last, ok := s.reported[key]; ok && now.Sub(last) < reportWindow {
		return false
	}
	if len(s.reported) >= maxReports {
		var oldestKey [32]byte
		var oldest time.Time
		for k, t := range s.reported {
			if oldest.IsZero() || t.Before(oldest) {
				oldestKey, oldest = k, t
			}
		}
		delete(s.reported, oldestKey)
	}
	s.reported[key] = now
	return true
}

// retractReport forgets a report admitted at `at` whose event was not written.
func (s *Scanner) retractReport(key [32]byte, at time.Time) {
	s.reportMu.Lock()
	if s.reported[key] == at {
		delete(s.reported, key)
	}
	s.reportMu.Unlock()
}

// readBoundedFile refuses special files and enforces a cap before allocation and
// while reading, including files that grow after Stat. The metadata checks catch
// ordinary writers and path replacements rather than caching an unstable read.
func readBoundedSnapshot(path string, limit int64, opener func(string) (*os.File, error)) ([]byte, os.FileInfo, error) {
	if limit <= 0 {
		limit = 25 << 20
	}
	before, err := os.Stat(path)
	if err != nil {
		return nil, nil, err
	}
	if !before.Mode().IsRegular() {
		return nil, nil, fmt.Errorf("scan target is not a regular file")
	}
	if before.Size() > limit {
		return nil, nil, fmt.Errorf("%w: input exceeds %d bytes", engine.ErrLimit, limit)
	}
	if opener == nil {
		opener = openScannerTarget
	}
	f, err := opener(path)
	if err != nil {
		return nil, nil, err
	}
	defer f.Close()
	opened, err := f.Stat()
	if err != nil {
		return nil, nil, err
	}
	if !opened.Mode().IsRegular() || !os.SameFile(before, opened) {
		return nil, nil, fmt.Errorf("%w while opening", errTargetChanged)
	}
	data, err := io.ReadAll(io.LimitReader(f, limit+1))
	if err != nil {
		return nil, nil, err
	}
	if int64(len(data)) > limit {
		return nil, nil, fmt.Errorf("%w: input exceeds %d bytes", engine.ErrLimit, limit)
	}
	after, err := f.Stat()
	if err != nil {
		return nil, nil, err
	}
	current, err := os.Stat(path)
	if err != nil {
		return nil, nil, err
	}
	if !os.SameFile(opened, current) || opened.Size() != after.Size() ||
		!opened.ModTime().Equal(after.ModTime()) || after.Size() != int64(len(data)) ||
		current.Size() != after.Size() || !current.ModTime().Equal(after.ModTime()) {
		return nil, nil, fmt.Errorf("%w while reading", errTargetChanged)
	}
	return data, after, nil
}
