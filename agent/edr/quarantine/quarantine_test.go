package quarantine

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

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
	id, err := store.Quarantine(src, payloadHash("payload"), "Win.Test.EICAR")
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

	// The quarantined copy must preserve the original bytes (move, not delete).
	stored := filepath.Join(dir, "quarantine", id+".quarantined")
	if b, err := os.ReadFile(stored); err != nil || string(b) != "payload" {
		t.Fatalf("quarantined copy content = %q err=%v, want original bytes", string(b), err)
	}

	if err := store.Restore(id); err != nil {
		t.Fatalf("Restore: %v", err)
	}
	if b, err := os.ReadFile(src); err != nil || string(b) != "payload" {
		t.Fatalf("restore did not put the original bytes back: %q err=%v", string(b), err)
	}
	if _, err := os.Stat(stored); !os.IsNotExist(err) {
		t.Fatal("quarantined copy should be gone after restore")
	}
	// Re-restoring an already-restored id must fail cleanly, not corrupt anything.
	if err := store.Restore(id); err == nil {
		t.Fatal("re-restore of an already-restored id should error")
	} else if !strings.Contains(err.Error(), "already restored") {
		t.Fatalf("re-restore error = %q, want 'already restored'", err)
	}
}

func TestPurgeAndPurgeExpired(t *testing.T) {
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

	// Two quarantined items.
	mk := func(name string) string {
		src := filepath.Join(dir, name)
		_ = os.WriteFile(src, []byte("x"), 0o644)
		id, err := store.Quarantine(src, payloadHash("x"), "Win.Test")
		if err != nil {
			t.Fatal(err)
		}
		return id
	}
	idA := mk("a.exe")
	idB := mk("b.exe")

	// Manual purge of A: stored file gone, record marked purged, re-purge errors.
	if err := store.Purge(idA); err != nil {
		t.Fatalf("purge: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "quarantine", idA+".quarantined")); !os.IsNotExist(err) {
		t.Fatal("purged file should be deleted")
	}
	rec, _, _ := c.GetQuarantine(idA)
	if !rec.Purged {
		t.Fatal("record not marked purged")
	}
	if err := store.Purge(idA); err == nil {
		t.Fatal("re-purge should error")
	}

	// PurgeExpired with a 0-day retention is a no-op (keep forever semantics).
	if n, _ := store.PurgeExpired(0, time.Now()); n != 0 {
		t.Fatalf("retention 0 must be a no-op, purged %d", n)
	}
	// Back-date B and sweep with a 1-day retention → B is purged, A already was.
	if err := c.StoreQuarantine(cache.QuarantineRecord{
		QuarantineID: idB, OriginalPath: filepath.Join(dir, "b.exe"), Engine: "UTMStack EDR",
		QuarantinedAt: time.Now().Add(-48 * time.Hour),
	}); err != nil {
		t.Fatal(err)
	}
	n, err := store.PurgeExpired(1, time.Now())
	if err != nil || n != 1 {
		t.Fatalf("PurgeExpired = %d, %v; want 1 purged", n, err)
	}
}

// lockDir makes dir read-only so a file in it can be read and copied but not
// renamed or removed, the way Windows treats a running image or a file another
// process holds open. The returned func unlocks it again.
func lockDir(t *testing.T, dir string) func() {
	t.Helper()
	if runtime.GOOS == "windows" || os.Geteuid() == 0 {
		t.Skip("needs POSIX directory permissions enforced for a non-root user")
	}
	if err := os.Chmod(dir, 0o500); err != nil {
		t.Fatal(err)
	}
	unlock := func() { _ = os.Chmod(dir, 0o700) }
	t.Cleanup(unlock)
	return unlock
}

func storeFiles(t *testing.T, dir string) []string {
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

// When the source can be copied but not removed, the store copy is only a
// duplicate of a file that stays in place. It must not be left behind without
// a record: invisible to 'quarantine list', never restorable or swept, and one
// more on every re-scan.
func TestQuarantineOfUnremovableSourceLeavesNoCopy(t *testing.T) {
	dir := t.TempDir()
	c, err := cache.Open(filepath.Join(dir, "edr.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	qdir := filepath.Join(dir, "quarantine")
	store, err := New(qdir, c)
	if err != nil {
		t.Fatal(err)
	}
	held := filepath.Join(dir, "held")
	if err := os.Mkdir(held, 0o700); err != nil {
		t.Fatal(err)
	}
	src := filepath.Join(held, "running.exe")
	if err := os.WriteFile(src, []byte("payload"), 0o600); err != nil {
		t.Fatal(err)
	}
	lockDir(t, held)

	for i := 0; i < 3; i++ {
		if _, err := store.Quarantine(src, payloadHash("payload"), "Win.Test"); err == nil {
			t.Fatal("quarantine of a source that cannot be removed must fail")
		}
	}
	if left := storeFiles(t, qdir); len(left) != 0 {
		t.Fatalf("unrecorded copies left in the store: %v", left)
	}
	if b, err := os.ReadFile(src); err != nil || string(b) != "payload" {
		t.Fatalf("original must stay in place untouched: %q err=%v", b, err)
	}
	if recs, err := c.ListQuarantine(); err != nil || len(recs) != 0 {
		t.Fatalf("records for a failed quarantine: %+v err=%v", recs, err)
	}
}

// forceCopy makes every rename fail, as it does across volumes, so the
// copy-then-remove fallback runs.
func forceCopy(t *testing.T) {
	t.Helper()
	orig := renameFile
	renameFile = func(string, string) error { return errors.New("not same device") }
	t.Cleanup(func() { renameFile = orig })
}

// The cross-volume fallback still moves the file: one recorded store copy,
// source gone, and restore puts the bytes back.
func TestQuarantineAcrossVolumesMovesAndRestores(t *testing.T) {
	forceCopy(t)
	dir := t.TempDir()
	c, err := cache.Open(filepath.Join(dir, "edr.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	qdir := filepath.Join(dir, "quarantine")
	store, err := New(qdir, c)
	if err != nil {
		t.Fatal(err)
	}
	src := filepath.Join(dir, "a.exe")
	if err := os.WriteFile(src, []byte("payload"), 0o600); err != nil {
		t.Fatal(err)
	}
	id, err := store.Quarantine(src, payloadHash("payload"), "Win.Test")
	if err != nil {
		t.Fatalf("Quarantine: %v", err)
	}
	if _, err := os.Stat(src); !os.IsNotExist(err) {
		t.Fatal("source should be gone after a cross-volume move")
	}
	if left := storeFiles(t, qdir); len(left) != 1 || left[0] != id+".quarantined" {
		t.Fatalf("store = %v, want only %s.quarantined", left, id)
	}
	if err := store.Restore(id); err != nil {
		t.Fatalf("Restore: %v", err)
	}
	if b, err := os.ReadFile(src); err != nil || string(b) != "payload" {
		t.Fatalf("restored bytes %q err=%v", b, err)
	}
}

// A copy fallback only ever deletes a file it created itself: restoring onto
// a path that is occupied again must fail and leave both files untouched.
func TestRestoreAcrossVolumesNeverOverwrites(t *testing.T) {
	dir := t.TempDir()
	c, err := cache.Open(filepath.Join(dir, "edr.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	qdir := filepath.Join(dir, "quarantine")
	store, err := New(qdir, c)
	if err != nil {
		t.Fatal(err)
	}
	src := filepath.Join(dir, "a.exe")
	if err := os.WriteFile(src, []byte("payload"), 0o600); err != nil {
		t.Fatal(err)
	}
	id, err := store.Quarantine(src, payloadHash("payload"), "Win.Test")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(src, []byte("user file"), 0o600); err != nil {
		t.Fatal(err)
	}
	forceCopy(t)
	if err := store.Restore(id); err == nil {
		t.Fatal("restore onto an occupied path must fail in the copy fallback")
	}
	if b, _ := os.ReadFile(src); string(b) != "user file" {
		t.Fatalf("occupying file changed: %q", b)
	}
	if b, err := os.ReadFile(filepath.Join(qdir, id+".quarantined")); err != nil || string(b) != "payload" {
		t.Fatalf("stored copy lost: %q err=%v", b, err)
	}
	if rec, _, _ := c.GetQuarantine(id); rec.Restored {
		t.Fatal("a failed restore must not mark the record restored")
	}
}

// Exercise the native same-filesystem path rather than forceCopy: POSIX rename
// would otherwise overwrite these entries before the exclusive fallback runs.
func TestRestoreSameFilesystemRefusesOccupiedPath(t *testing.T) {
	for _, kind := range []string{"regular", "directory", "symlink", "dangling-symlink"} {
		t.Run(kind, func(t *testing.T) {
			dir := t.TempDir()
			c, err := cache.Open(filepath.Join(dir, "edr.db"))
			if err != nil {
				t.Fatal(err)
			}
			defer c.Close()
			qdir := filepath.Join(dir, "quarantine")
			store, err := New(qdir, c)
			if err != nil {
				t.Fatal(err)
			}
			src := filepath.Join(dir, "original")
			if err = os.WriteFile(src, []byte("quarantined payload"), 0600); err != nil {
				t.Fatal(err)
			}
			id, err := store.Quarantine(src, payloadHash("quarantined payload"), "Safe.Fixture")
			if err != nil {
				t.Fatal(err)
			}
			target := filepath.Join(dir, "innocent-target")
			switch kind {
			case "regular":
				err = os.WriteFile(src, []byte("benign replacement"), 0600)
			case "directory":
				err = os.Mkdir(src, 0700)
			default:
				if kind == "symlink" {
					if err = os.WriteFile(target, []byte("innocent target"), 0600); err != nil {
						t.Fatal(err)
					}
				}
				if err = os.Symlink(target, src); err != nil {
					t.Skipf("symlinks unavailable: %v", err)
				}
			}
			if err != nil {
				t.Fatal(err)
			}
			if err = store.Restore(id); err == nil {
				t.Fatal("restore overwrote occupied destination")
			}
			stored := filepath.Join(qdir, id+".quarantined")
			if b, e := os.ReadFile(stored); e != nil || string(b) != "quarantined payload" {
				t.Fatalf("stored source lost: %q %v", b, e)
			}
			switch kind {
			case "regular":
				if b, e := os.ReadFile(src); e != nil || string(b) != "benign replacement" {
					t.Fatalf("replacement changed: %q %v", b, e)
				}
			case "directory":
				if info, e := os.Stat(src); e != nil || !info.IsDir() {
					t.Fatal("destination directory replaced")
				}
			default:
				if link, e := os.Readlink(src); e != nil || link != target {
					t.Fatalf("destination symlink replaced: %q %v", link, e)
				}
				if kind == "symlink" {
					if b, e := os.ReadFile(target); e != nil || string(b) != "innocent target" {
						t.Fatal("symlink target changed")
					}
				} else if _, e := os.Lstat(target); !os.IsNotExist(e) {
					t.Fatal("dangling target created")
				}
			}
			if rec, found, e := c.GetQuarantine(id); e != nil || !found || rec.Restored {
				t.Fatalf("failed restore record changed: %+v %v", rec, e)
			}
		})
	}
}

func TestConcurrentMovesNeverReplaceWinningDestination(t *testing.T) {
	for _, copyFallback := range []bool{false, true} {
		t.Run(fmt.Sprint(copyFallback), func(t *testing.T) {
			if copyFallback {
				forceCopy(t)
			}
			dir := t.TempDir()
			dst := filepath.Join(dir, "winner")
			const competitors = 16
			start := make(chan struct{})
			type result struct {
				index int
				err   error
			}
			results := make(chan result, competitors)
			var wg sync.WaitGroup
			for i := 0; i < competitors; i++ {
				src := filepath.Join(dir, fmt.Sprintf("source-%02d", i))
				if err := os.WriteFile(src, []byte(fmt.Sprintf("payload-%02d", i)), 0600); err != nil {
					t.Fatal(err)
				}
				wg.Add(1)
				go func(i int, src string) { defer wg.Done(); <-start; results <- result{i, moveFile(src, dst)} }(i, src)
			}
			close(start)
			wg.Wait()
			close(results)
			winner := -1
			for r := range results {
				src := filepath.Join(dir, fmt.Sprintf("source-%02d", r.index))
				if r.err == nil {
					if winner != -1 {
						t.Fatal("multiple moves replaced the destination")
					}
					winner = r.index
					if _, e := os.Lstat(src); !os.IsNotExist(e) {
						t.Fatal("winning source not moved")
					}
				} else if b, e := os.ReadFile(src); e != nil || string(b) != fmt.Sprintf("payload-%02d", r.index) {
					t.Fatalf("losing source lost: %q %v", b, e)
				}
			}
			if winner == -1 {
				t.Fatal("no move won")
			}
			if b, e := os.ReadFile(dst); e != nil || string(b) != fmt.Sprintf("payload-%02d", winner) {
				t.Fatalf("winning destination replaced: %q %v", b, e)
			}
		})
	}
}

func TestQuarantineUnavailableMetadataDoesNotMoveSource(t *testing.T) {
	s, c, src, _ := admissionFixture(t)
	if err := c.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Quarantine(src, payloadHash("malicious marker"), "Safe.Fixture"); err == nil {
		t.Fatal("closed metadata store accepted")
	}
	if b, err := os.ReadFile(src); err != nil || string(b) != "malicious marker" {
		t.Fatalf("source changed: %q %v", b, err)
	}
	if files := storeFiles(t, s.dir); len(files) != 0 {
		t.Fatalf("unprepared admission left data: %v", files)
	}
}

func TestRestoreReportsActualLocationAfterRecordUpdateFailure(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "edr.db")
	c, err := cache.Open(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	store, err := New(filepath.Join(dir, "quarantine"), c)
	if err != nil {
		t.Fatal(err)
	}
	src := filepath.Join(dir, "original")
	if err = os.WriteFile(src, []byte("safe payload"), 0600); err != nil {
		t.Fatal(err)
	}
	id, err := store.Quarantine(src, payloadHash("safe payload"), "Safe.Fixture")
	if err != nil {
		t.Fatal(err)
	}
	stored := filepath.Join(dir, "quarantine", id+".quarantined")
	store.restoreAfterPublish = func() error { return c.Close() }
	err = store.Restore(id)
	if err == nil || !strings.Contains(err.Error(), fmt.Sprintf("%q", src)) || !strings.Contains(err.Error(), "was not updated") {
		t.Fatalf("actual restored location omitted: %v", err)
	}
	if b, e := os.ReadFile(src); e != nil || string(b) != "safe payload" {
		t.Fatalf("restored file lost: %q %v", b, e)
	}
	if b, e := os.ReadFile(stored); e != nil || string(b) != "safe payload" {
		t.Fatal("verified backup must survive database failure", e)
	}
	reopened, e := cache.Open(dbPath)
	if e != nil {
		t.Fatal(e)
	}
	defer reopened.Close()
	if rec, found, e := reopened.GetQuarantine(id); e != nil || !found || rec.Restored {
		t.Fatalf("unexpected metadata state: %+v %v", rec, e)
	}
	recovered, e := New(store.dir, reopened)
	if e != nil {
		t.Fatal(e)
	}
	if results := recovered.RecoveryStatus(); len(results) != 1 || results[0].Status != "restored" {
		t.Fatalf("database failure not recovered: %+v", results)
	}
	if rec, _, e := reopened.GetQuarantine(id); e != nil || !rec.Restored {
		t.Fatal("restore receipt not replayed", e)
	}
	if _, e := os.Lstat(stored); !os.IsNotExist(e) {
		t.Fatal("backup not retired after receipt", e)
	}
}
