package quarantine

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/utmstack/UTMStack/agent/edr/cache"
)

func TestAdmissionDoesNotRecursivelyQuarantineOwnedEvidence(t *testing.T) {
	s, c, path, dir := admissionFixture(t)
	id, err := s.Quarantine(path, payloadHash("malicious marker"), "Safe.Managed.Control")
	if err != nil {
		t.Fatal(err)
	}
	r, _, err := c.GetQuarantine(id)
	if err != nil {
		t.Fatal(err)
	}
	for _, owned := range []string{filepath.Join(s.dir, id+".quarantined"), filepath.Join(r.SourceCaptureDir, "source")} {
		if _, err = s.Quarantine(owned, payloadHash("malicious marker"), "Safe.Managed.Control"); !errors.Is(err, ErrAlreadyQuarantined) {
			t.Fatalf("managed evidence re-admitted: %s: %v", owned, err)
		}
	}
	fake := filepath.Join(dir, ".edr-quarantine-not-a-real-item")
	if err = os.Mkdir(fake, 0700); err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(fake, "marker")
	if err = os.WriteFile(p, []byte("malicious marker"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err = s.Quarantine(p, payloadHash("malicious marker"), "Safe.Managed.Control"); err != nil {
		t.Fatalf("reserved-looking name bypassed admission: %v", err)
	}
}

func TestAdmissionCopyFailureLeavesNoCapturedSource(t *testing.T) {
	s, c, path, dir := admissionFixture(t)
	s.admissionCopy = func(w io.Writer, r io.Reader) (int64, error) {
		n, _ := io.CopyN(w, r, 4)
		return n, io.ErrUnexpectedEOF
	}
	if _, err := s.Quarantine(path, payloadHash("malicious marker"), "Safe.Copy.Control"); err == nil {
		t.Fatal("partial copy admitted")
	}
	if b, err := os.ReadFile(path); err != nil || string(b) != "malicious marker" {
		t.Fatalf("source lost: %q %v", b, err)
	}
	if rows, err := c.ListQuarantine(); err != nil || len(rows) != 0 {
		t.Fatalf("partial admission recorded: %+v %v", rows, err)
	}
	for _, p := range []string{dir, filepath.Join(s.dir, ".admissions")} {
		entries, err := os.ReadDir(p)
		if err != nil {
			t.Fatal(err)
		}
		for _, e := range entries {
			if strings.HasPrefix(e.Name(), ".edr-quarantine-") || p != dir {
				t.Fatalf("partial admission artifacts retained: %s/%s", p, e.Name())
			}
		}
	}
}

func TestAdmissionSourceGrowthCannotExpandSnapshotCopy(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows denies the conflicting writer open")
	}
	s, _, path, _ := admissionFixture(t)
	var copied int64
	s.admissionCopy = func(w io.Writer, r io.Reader) (int64, error) {
		f, err := os.OpenFile(path, os.O_WRONLY, 0)
		if err != nil {
			return 0, err
		}
		err = f.Truncate(maxRestoreBytes + 1)
		_ = f.Close()
		if err != nil {
			return 0, err
		}
		copied, err = io.Copy(w, r)
		return copied, err
	}
	if _, err := s.Quarantine(path, payloadHash("malicious marker"), "Safe.Growth.Control"); !errors.Is(err, ErrAdmissionChanged) {
		t.Fatalf("growing source not refused: %v", err)
	}
	if copied != int64(len("malicious marker")) {
		t.Fatalf("snapshot exceeded pinned size: %d", copied)
	}
	if info, err := os.Stat(path); err != nil || info.Size() != maxRestoreBytes+1 {
		t.Fatalf("growing source changed or moved: %v %v", info, err)
	}
}

func TestAdmissionCorruptPendingSnapshotIsNotCommitted(t *testing.T) {
	s, c, path, _ := admissionFixture(t)
	s.admissionFault = func(phase string) error {
		if phase == "claimed" {
			return errors.New("simulated interruption")
		}
		return nil
	}
	id, err := s.Quarantine(path, payloadHash("malicious marker"), "Safe.Corruption.Control")
	if err == nil || id == "" {
		t.Fatalf("durable capture not reported: %q %v", id, err)
	}
	op := filepath.Join(s.dir, ".admissions", id)
	if err = os.WriteFile(filepath.Join(op, "snapshot"), []byte("corrupted bytes"), 0600); err != nil {
		t.Fatal(err)
	}
	results, err := s.Recover()
	if err == nil || len(results) != 1 || results[0].Status != "needs_attention" {
		t.Fatalf("corrupt data committed: %+v %v", results, err)
	}
	if rows, err := c.ListQuarantine(); err != nil || len(rows) != 0 {
		t.Fatalf("unverified snapshot recorded: %+v %v", rows, err)
	}
	if b, err := os.ReadFile(filepath.Join(filepath.Dir(path), ".edr-quarantine-"+id, "source")); err != nil || string(b) != "malicious marker" {
		t.Fatalf("original data not retained: %q %v", b, err)
	}
}

func TestAdmissionRejectsInvalidRecoveryMetadata(t *testing.T) {
	for _, kind := range []string{"path", "capture", "phase", "json", "oversized"} {
		t.Run(kind, func(t *testing.T) {
			s, c, path, _ := admissionFixture(t)
			s.admissionFault = func(phase string) error {
				if phase == "claimed" {
					return errors.New("interrupted commit")
				}
				return nil
			}
			id, err := s.Quarantine(path, payloadHash("malicious marker"), "Safe.Metadata.Control")
			if err == nil || id == "" {
				t.Fatal("interruption not reached", err)
			}
			journal := filepath.Join(s.dir, ".admissions", id, "manifest.json")
			body, err := os.ReadFile(journal)
			if err != nil {
				t.Fatal(err)
			}
			var m admissionManifest
			if err = json.Unmarshal(body, &m); err != nil {
				t.Fatal(err)
			}
			switch kind {
			case "path":
				m.Record.OriginalPath = path + ".different"
			case "capture":
				m.Record.SourceCaptureDir = filepath.Dir(path)
			case "phase":
				m.Phase = "unknown"
			}
			body, err = json.Marshal(m)
			if err != nil {
				t.Fatal(err)
			}
			if kind == "json" {
				body = []byte("{")
			} else if kind == "oversized" {
				body = make([]byte, maxAdmissionJournal+1)
			}
			if err = os.WriteFile(journal, body, 0600); err != nil {
				t.Fatal(err)
			}
			results, err := s.Recover()
			if err == nil || len(results) != 1 || results[0].Status != "needs_attention" {
				t.Fatalf("invalid metadata accepted: %+v %v", results, err)
			}
			if rows, err := c.ListQuarantine(); err != nil || len(rows) != 0 {
				t.Fatalf("invalid metadata committed: %+v %v", rows, err)
			}
			for _, p := range []string{filepath.Join(filepath.Dir(path), ".edr-quarantine-"+id, "source"), filepath.Join(s.dir, ".admissions", id, "snapshot")} {
				if b, err := os.ReadFile(p); err != nil || string(b) != "malicious marker" {
					t.Fatalf("invalid metadata discarded evidence: %q %v", b, err)
				}
			}
		})
	}
}

func TestAdmissionRecoveryRetainsUnknownEntriesAndManifest(t *testing.T) {
	s, c, path, _ := admissionFixture(t)
	s.admissionFault = func(phase string) error {
		if phase == "claimed" {
			return errors.New("interrupted commit")
		}
		return nil
	}
	id, err := s.Quarantine(path, payloadHash("malicious marker"), "Safe.Unknown.Control")
	if err == nil || id == "" {
		t.Fatal("interruption not reached", err)
	}
	op := filepath.Join(s.dir, ".admissions", id)
	unknown := filepath.Join(op, "operator-note")
	if err = os.WriteFile(unknown, []byte("retained operator data"), 0600); err != nil {
		t.Fatal(err)
	}
	results, err := s.Recover()
	if err == nil || len(results) != 1 || results[0].Status != "needs_attention" {
		t.Fatalf("unknown entry silently discarded: %+v %v", results, err)
	}
	if _, err = os.Stat(filepath.Join(op, "manifest.json")); err != nil {
		t.Fatal("cleanup removed the recovery manifest before validating all entries", err)
	}
	if b, err := os.ReadFile(unknown); err != nil || string(b) != "retained operator data" {
		t.Fatalf("unknown data lost: %q %v", b, err)
	}
	if rows, err := c.ListQuarantine(); err != nil || len(rows) != 1 {
		t.Fatalf("verified captured bytes were not recorded: %+v %v", rows, err)
	}
	if err = os.Remove(unknown); err != nil {
		t.Fatal(err)
	}
	if _, err = s.Recover(); err != nil {
		t.Fatal("retry after resolving unknown entry failed", err)
	}
}

func TestAdmissionReplayPreservesLaterRestoreState(t *testing.T) {
	s, c, path, _ := admissionFixture(t)
	s.admissionFault = func(phase string) error {
		if phase == "recorded" {
			return errors.New("cleanup interrupted")
		}
		return nil
	}
	id, err := s.Quarantine(path, payloadHash("malicious marker"), "Safe.Replay.Control")
	if err == nil || id == "" {
		t.Fatalf("committed operation not reported: %q %v", id, err)
	}
	if err = s.Restore(id); err != nil {
		t.Fatal(err)
	}
	if _, err = s.Recover(); err != nil {
		t.Fatal(err)
	}
	r, found, err := c.GetQuarantine(id)
	if err != nil || !found || !r.Restored {
		t.Fatalf("replay reset restored state: %+v %v", r, err)
	}
}

func TestAdmissionRecoveryRollsBackUnverifiedCapture(t *testing.T) {
	for _, replacement := range []bool{false, true} {
		t.Run(fmt.Sprint(replacement), func(t *testing.T) {
			s, c, path, _ := admissionFixture(t)
			s.admissionFault = func(phase string) error {
				if phase == "claimed-before-check" {
					return errors.New("interrupted before captured identity verification")
				}
				return nil
			}
			if _, err := s.Quarantine(path, payloadHash("malicious marker"), "Safe.Recovery.Control"); err == nil {
				t.Fatal("interruption not reached")
			}
			entries, err := os.ReadDir(filepath.Join(s.dir, ".admissions"))
			if err != nil || len(entries) != 1 {
				t.Fatalf("missing recovery record: %+v %v", entries, err)
			}
			raw := filepath.Join(filepath.Dir(path), ".edr-quarantine-"+entries[0].Name(), "source")
			if replacement {
				if err = os.Rename(raw, path+".old"); err != nil {
					t.Fatal(err)
				}
			}
			if err = os.WriteFile(raw, []byte("unverified benign data"), 0600); err != nil {
				t.Fatal(err)
			}
			results, err := s.Recover()
			if err != nil || len(results) != 1 || results[0].Status != "aborted" {
				t.Fatalf("unverified capture not rolled back: %+v %v", results, err)
			}
			if b, err := os.ReadFile(path); err != nil || string(b) != "unverified benign data" {
				t.Fatalf("unverified data lost: %q %v", b, err)
			}
			if rows, err := c.ListQuarantine(); err != nil || len(rows) != 0 {
				t.Fatalf("unverified source committed: %+v %v", rows, err)
			}
			if replacement {
				if b, err := os.ReadFile(path + ".old"); err != nil || string(b) != "malicious marker" {
					t.Fatalf("original inode lost: %q %v", b, err)
				}
			}
		})
	}
}

func TestConcurrentStoresAdmitOneSourceOnlyOnce(t *testing.T) {
	s, c, path, _ := admissionFixture(t)
	other, err := New(s.dir, c)
	if err != nil {
		t.Fatal(err)
	}
	start := make(chan struct{})
	out := make(chan error, 2)
	for _, store := range []*Store{s, other} {
		go func(store *Store) {
			<-start
			_, err := store.Quarantine(path, payloadHash("malicious marker"), "Safe.Concurrent.Control")
			out <- err
		}(store)
	}
	close(start)
	success := 0
	for i := 0; i < 2; i++ {
		if <-out == nil {
			success++
		}
	}
	if success != 1 {
		t.Fatalf("successful source admissions=%d", success)
	}
	rows, err := c.ListQuarantine()
	if err != nil || len(rows) != 1 {
		t.Fatalf("duplicate records: %+v %v", rows, err)
	}
	if b, err := os.ReadFile(filepath.Join(s.dir, rows[0].QuarantineID+".quarantined")); err != nil || string(b) != "malicious marker" {
		t.Fatalf("winning snapshot lost: %q %v", b, err)
	}
	if _, err = s.Recover(); err != nil {
		t.Fatal(err)
	}
}

func TestAdmissionRequiresTheOriginalScanIdentity(t *testing.T) {
	s, _, path, _ := admissionFixture(t)
	// Match the scanner's descriptor-derived identity. On Windows os.Stat can
	// defer file-ID lookup until SameFile, after this pathname has been reused.
	scanned, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	old, statErr := scanned.Stat()
	closeErr := scanned.Close()
	if err := errors.Join(statErr, closeErr); err != nil {
		t.Fatal(err)
	}
	if err = os.Rename(path, path+".old"); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(path, []byte("malicious marker"), 0600); err != nil {
		t.Fatal(err)
	}
	// Equal bytes, length and timestamp must not substitute a different file ID.
	if err = os.Chtimes(path, old.ModTime(), old.ModTime()); err != nil {
		t.Fatal(err)
	}
	if _, err = s.QuarantineVerified(path, payloadHash("malicious marker"), "Safe.Identity.Control", old); !errors.Is(err, ErrAdmissionChanged) {
		t.Fatalf("replacement identity accepted: %v", err)
	}
	if b, err := os.ReadFile(path); err != nil || string(b) != "malicious marker" {
		t.Fatalf("replacement moved: %q %v", b, err)
	}
}

func TestAdmissionCaptureRacePreservesReplacement(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows denies replacement while the source handle is held; tested separately")
	}
	for _, collision := range []bool{false, true} {
		t.Run(fmt.Sprint(collision), func(t *testing.T) {
			s, c, path, _ := admissionFixture(t)
			s.admissionFault = func(phase string) error {
				if phase == "ready" {
					if err := os.Rename(path, path+".old"); err != nil {
						return err
					}
					return os.WriteFile(path, []byte("benign replacement"), 0644)
				}
				if phase == "claimed-before-check" && collision {
					return os.WriteFile(path, []byte("newest benign file"), 0600)
				}
				return nil
			}
			if _, err := s.Quarantine(path, payloadHash("malicious marker"), "Safe.Identity.Control"); err == nil {
				t.Fatal("replacement admitted")
			}
			want := "benign replacement"
			if collision {
				want = "newest benign file"
			}
			if b, err := os.ReadFile(path); err != nil || string(b) != want {
				t.Fatalf("replacement lost: %q %v", b, err)
			}
			if b, err := os.ReadFile(path + ".old"); err != nil || string(b) != "malicious marker" {
				t.Fatalf("original inode changed: %q %v", b, err)
			}
			if rows, err := c.ListQuarantine(); err != nil || len(rows) != 0 {
				t.Fatalf("wrong object recorded: %+v %v", rows, err)
			}
			if collision {
				results, err := s.Recover()
				if err == nil || len(results) != 1 || results[0].Status != "needs_attention" {
					t.Fatalf("rollback conflict hidden: %+v %v", results, err)
				}
				if err = os.Rename(path, path+".newest"); err != nil {
					t.Fatal(err)
				}
				results, err = s.Recover()
				if err != nil || len(results) != 1 || results[0].Status != "aborted" {
					t.Fatalf("safe rollback recovery: %+v %v", results, err)
				}
				if b, err := os.ReadFile(path); err != nil || string(b) != "benign replacement" {
					t.Fatalf("captured benign entry lost: %q %v", b, err)
				}
			}
		})
	}
}

func TestAdmissionSnapshotIsIndependentOfRetainedWriter(t *testing.T) {
	s, c, path, _ := admissionFixture(t)
	writer, err := os.OpenFile(path, os.O_RDWR, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer writer.Close()
	s.admissionFault = func(phase string) error {
		if phase == "claimed" {
			_, err := writer.WriteAt([]byte("changed by writer"), 0)
			return err
		}
		return nil
	}
	id, err := s.Quarantine(path, payloadHash("malicious marker"), "Safe.Writer.Control")
	if runtime.GOOS == "windows" {
		if err == nil {
			t.Fatal("Windows admitted an existing writable handle")
		}
		if _, err = os.Stat(path); err != nil {
			t.Fatal("busy source moved", err)
		}
		return
	}
	if err != nil {
		t.Fatal(err)
	}
	_ = writer.Close()
	r, found, err := c.GetQuarantine(id)
	if err != nil || !found {
		t.Fatalf("record: %+v %v", r, err)
	}
	if b, err := os.ReadFile(filepath.Join(s.dir, id+".quarantined")); err != nil || string(b) != "malicious marker" {
		t.Fatalf("retained writer changed immutable snapshot: %q %v", b, err)
	}
	if b, err := os.ReadFile(filepath.Join(r.SourceCaptureDir, "source")); err != nil || string(b) != "changed by writer" {
		t.Fatalf("writer's later data lost: %q %v", b, err)
	}
	if err = s.Restore(id); err != nil {
		t.Fatal(err)
	}
	if b, err := os.ReadFile(path); err != nil || string(b) != "malicious marker" {
		t.Fatalf("unverified original restored: %q %v", b, err)
	}
	if _, err = os.Stat(filepath.Join(r.SourceCaptureDir, "source")); err != nil {
		t.Fatal("restore discarded retained writer data", err)
	}
	if err = s.Purge(id); err != nil {
		t.Fatal(err)
	}
	if _, err = os.Stat(r.SourceCaptureDir); !os.IsNotExist(err) {
		t.Fatalf("deliberate purge retained raw source: %v", err)
	}
}

func TestAdmissionInterruptedMetadataCommitIsRecoverable(t *testing.T) {
	s, c, path, dir := admissionFixture(t)
	s.admissionFault = func(phase string) error {
		if phase == "published" {
			if err := os.WriteFile(path, []byte("new benign file"), 0600); err != nil {
				return err
			}
			return c.Close()
		}
		return nil
	}
	id, err := s.Quarantine(path, payloadHash("malicious marker"), "Safe.Commit.Control")
	var contained *ContainmentError
	if err == nil || !errors.As(err, &contained) || id == "" || contained.QuarantineID != id {
		t.Fatalf("physical result hidden: id=%q err=%v", id, err)
	}
	if r, found, recentErr := s.RecentQuarantineByPath(path, 5*time.Minute); recentErr != nil || !found || r.QuarantineID != id {
		t.Fatalf("durable verdict lost while SQL unavailable: %+v %v", r, recentErr)
	}
	reopened, err := cache.Open(filepath.Join(dir, "edr.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	s2, err := New(s.dir, reopened)
	if err != nil {
		t.Fatal(err)
	}
	r, found, err := reopened.GetQuarantine(id)
	if err != nil || !found || !r.Restorable {
		t.Fatalf("journal was not committed: %+v %v", r, err)
	}
	if b, err := os.ReadFile(path); err != nil || string(b) != "new benign file" {
		t.Fatalf("recovery overwrote replacement: %q %v", b, err)
	}
	if _, err = s2.Recover(); err != nil {
		t.Fatal("idempotent recovery", err)
	}
	if rows, err := reopened.ListQuarantine(); err != nil || len(rows) != 1 {
		t.Fatalf("duplicate recovery rows: %+v %v", rows, err)
	}
}

func TestAdmissionCrashHelper(t *testing.T) {
	if os.Getenv("EDR_ADMISSION_CRASH_HELPER") != "1" {
		t.Skip("subprocess crash helper")
	}
	dir := os.Getenv("EDR_ADMISSION_CRASH_DIR")
	phase := os.Getenv("EDR_ADMISSION_CRASH_PHASE")
	c, err := cache.Open(filepath.Join(dir, "edr.db"))
	if err != nil {
		t.Fatal(err)
	}
	s, err := New(filepath.Join(dir, "quarantine"), c)
	if err != nil {
		t.Fatal(err)
	}
	s.admissionFault = func(at string) error {
		if at == phase {
			if os.Getenv("EDR_ADMISSION_CRASH_KILL") == "1" {
				fmt.Println("BOUNDARY_READY")
				time.Sleep(time.Hour)
			}
			os.Exit(93)
		}
		return nil
	}
	_, err = s.Quarantine(filepath.Join(dir, "original"), payloadHash("malicious marker"), "Safe.Crash.Control")
	t.Fatalf("crash boundary was not reached: %v", err)
}

func TestAdmissionRecoversForcedProcessTermination(t *testing.T) {
	for _, phase := range []string{"source-synced", "intent", "snapshot-created", "snapshot", "ready", "claimed-before-check", "verified", "captured-synced", "claimed", "published", "recorded"} {
		t.Run(phase, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "original")
			if err := os.WriteFile(path, []byte("malicious marker"), 0600); err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestAdmissionCrashHelper$")
			cmd.Env = append(os.Environ(), "EDR_ADMISSION_CRASH_HELPER=1", "EDR_ADMISSION_CRASH_KILL=1", "EDR_ADMISSION_CRASH_DIR="+dir, "EDR_ADMISSION_CRASH_PHASE="+phase)
			pipe, err := cmd.StdoutPipe()
			if err != nil {
				t.Fatal(err)
			}
			if err = cmd.Start(); err != nil {
				t.Fatal(err)
			}
			line, readErr := bufio.NewReader(pipe).ReadString('\n')
			if readErr != nil || line != "BOUNDARY_READY\n" {
				_ = cmd.Process.Kill()
				_ = cmd.Wait()
				t.Fatalf("boundary not reached: %q %v", line, readErr)
			}
			if err = cmd.Process.Kill(); err != nil {
				_ = cmd.Wait()
				t.Fatal(err)
			}
			if err = cmd.Wait(); err == nil {
				t.Fatal("forced termination succeeded without an exit error")
			}
			c, err := cache.Open(filepath.Join(dir, "edr.db"))
			if err != nil {
				t.Fatal(err)
			}
			defer c.Close()
			s, err := New(filepath.Join(dir, "quarantine"), c)
			if err != nil {
				t.Fatal(err)
			}
			rows, err := c.ListQuarantine()
			if err != nil {
				t.Fatal(err)
			}
			claimed := phase == "claimed-before-check" || phase == "verified" || phase == "captured-synced" || phase == "claimed" || phase == "published" || phase == "recorded"
			if claimed {
				if len(rows) != 1 {
					t.Fatalf("captured data not recovered: %+v %+v", rows, s.RecoveryStatus())
				}
				if err = s.Restore(rows[0].QuarantineID); err != nil {
					t.Fatal(err)
				}
			} else if len(rows) != 0 {
				t.Fatalf("unclaimed file admitted: %+v", rows)
			}
			if b, err := os.ReadFile(path); err != nil || string(b) != "malicious marker" {
				t.Fatalf("termination lost data: %q %v", b, err)
			}
			if _, err = s.Recover(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestAdmissionRecoversActualProcessInterruption(t *testing.T) {
	for _, phase := range []string{"source-synced", "intent", "snapshot-created", "snapshot", "ready", "claimed-before-check", "verified", "captured-synced", "claimed", "published", "recorded"} {
		t.Run(phase, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "original")
			if err := os.WriteFile(path, []byte("malicious marker"), 0600); err != nil {
				t.Fatal(err)
			}
			cmd := exec.Command(os.Args[0], "-test.run=^TestAdmissionCrashHelper$")
			cmd.Env = append(os.Environ(), "EDR_ADMISSION_CRASH_HELPER=1", "EDR_ADMISSION_CRASH_DIR="+dir, "EDR_ADMISSION_CRASH_PHASE="+phase)
			output, err := cmd.CombinedOutput()
			var exit *exec.ExitError
			if !errors.As(err, &exit) || exit.ExitCode() != 93 {
				t.Fatalf("helper did not stop at %s: %v %s", phase, err, output)
			}
			c, err := cache.Open(filepath.Join(dir, "edr.db"))
			if err != nil {
				t.Fatal(err)
			}
			defer c.Close()
			s, err := New(filepath.Join(dir, "quarantine"), c)
			if err != nil {
				t.Fatal(err)
			}
			rows, err := c.ListQuarantine()
			if err != nil {
				t.Fatal(err)
			}
			claimed := phase == "claimed-before-check" || phase == "verified" || phase == "captured-synced" || phase == "claimed" || phase == "published" || phase == "recorded"
			if claimed {
				if len(rows) != 1 || !rows[0].Restorable {
					t.Fatalf("captured source not recovered: %+v; status=%+v", rows, s.RecoveryStatus())
				}
				if _, err = os.Stat(path); !os.IsNotExist(err) {
					t.Fatal("claimed source reappeared", err)
				}
				if err = s.Restore(rows[0].QuarantineID); err != nil {
					t.Fatal(err)
				}
			} else if len(rows) != 0 {
				t.Fatalf("unclaimed work was committed: %+v", rows)
			}
			if b, err := os.ReadFile(path); err != nil || string(b) != "malicious marker" {
				t.Fatalf("data lost across interruption: %q %v", b, err)
			}
			if _, err = s.Recover(); err != nil {
				t.Fatal("second recovery failed", err)
			}
			pending, err := os.ReadDir(filepath.Join(s.dir, ".admissions"))
			if phase == "source-synced" {
				if !os.IsNotExist(err) {
					t.Fatalf("source synchronization interruption created an admission journal: %v %v", pending, err)
				}
				return
			}
			if err != nil || len(pending) != 0 {
				t.Fatalf("unreconciled journal: %v %v", pending, err)
			}
		})
	}
}

func admissionFixture(t *testing.T) (*Store, *cache.Cache, string, string) {
	t.Helper()
	dir := t.TempDir()
	c, err := cache.Open(filepath.Join(dir, "edr.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = c.Close() })
	s, err := New(filepath.Join(dir, "quarantine"), c)
	if err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(dir, "original")
	if err = os.WriteFile(p, []byte("malicious marker"), 0640); err != nil {
		t.Fatal(err)
	}
	return s, c, p, dir
}

func TestAdmissionRejectsUnverifiedContent(t *testing.T) {
	for _, hash := range []string{payloadHash("different bytes"), "invalid", ""} {
		t.Run(hash, func(t *testing.T) {
			s, c, p, _ := admissionFixture(t)
			if _, err := s.Quarantine(p, hash, "Safe.Admission.Control"); err == nil {
				t.Fatal("unverified bytes admitted")
			}
			if b, err := os.ReadFile(p); err != nil || string(b) != "malicious marker" {
				t.Fatalf("source changed: %q %v", b, err)
			}
			if rows, err := c.ListQuarantine(); err != nil || len(rows) != 0 {
				t.Fatalf("failed admission recorded: %+v %v", rows, err)
			}
		})
	}
}

func TestAdmissionRejectsLeafSymlinkWithoutChangingTarget(t *testing.T) {
	s, c, p, dir := admissionFixture(t)
	link := filepath.Join(dir, "link")
	if err := os.Symlink(p, link); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	if _, err := s.Quarantine(link, payloadHash("malicious marker"), "Safe.Admission.Control"); err == nil {
		t.Fatal("source symlink admitted")
	}
	if to, err := os.Readlink(link); err != nil || to != p {
		t.Fatalf("link moved or changed: %q %v", to, err)
	}
	if b, err := os.ReadFile(p); err != nil || string(b) != "malicious marker" {
		t.Fatalf("target changed: %q %v", b, err)
	}
	if rows, err := c.ListQuarantine(); err != nil || len(rows) != 0 {
		t.Fatalf("symlink recorded: %+v %v", rows, err)
	}
}

func TestAdmissionRejectsChangedStoreDirectory(t *testing.T) {
	s, _, p, dir := admissionFixture(t)
	trap := filepath.Join(dir, "unprotected")
	if err := os.Mkdir(trap, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(s.dir, s.dir+".saved"); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(trap, s.dir); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	if _, err := s.Quarantine(p, payloadHash("malicious marker"), "Safe.Admission.Control"); err == nil {
		t.Fatal("redirected store accepted")
	}
	if b, err := os.ReadFile(p); err != nil || string(b) != "malicious marker" {
		t.Fatalf("source moved into substitute store: %q %v", b, err)
	}
	if entries, err := os.ReadDir(trap); err != nil || len(entries) != 0 {
		t.Fatalf("substitute store changed: %v %v", entries, err)
	}
}

func TestAdmissionRejectsNonregularAndOversizedSources(t *testing.T) {
	for _, kind := range []string{"directory", "oversized"} {
		t.Run(kind, func(t *testing.T) {
			s, _, p, _ := admissionFixture(t)
			if err := os.Remove(p); err != nil {
				t.Fatal(err)
			}
			if kind == "directory" {
				if err := os.Mkdir(p, 0700); err != nil {
					t.Fatal(err)
				}
			} else {
				f, err := os.Create(p)
				if err != nil {
					t.Fatal(err)
				}
				err = f.Truncate(maxRestoreBytes + 1)
				_ = f.Close()
				if err != nil {
					t.Fatal(err)
				}
			}
			if _, err := s.Quarantine(p, payloadHash("malicious marker"), "Safe.Admission.Control"); err == nil {
				t.Fatal("ineligible source admitted")
			}
			if _, err := os.Lstat(p); err != nil {
				t.Fatalf("source moved: %v", err)
			}
		})
	}
}
