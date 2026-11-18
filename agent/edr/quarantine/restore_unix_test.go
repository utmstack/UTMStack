//go:build linux || darwin

package quarantine

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/sys/unix"
)

func noRestoreStages(t *testing.T, paths ...string) {
	t.Helper()
	for _, path := range paths {
		entries, err := os.ReadDir(path)
		if err != nil {
			t.Fatal(err)
		}
		for _, entry := range entries {
			if strings.HasPrefix(entry.Name(), ".edr-restore-") {
				t.Fatalf("private stage leaked in %s: %s", path, entry.Name())
			}
		}
	}
}

func TestRestoreRejectsFIFOWithoutBlocking(t *testing.T) {
	s, _, id, original, stored := restoreFixture(t)
	if err := os.Remove(stored); err != nil {
		t.Fatal(err)
	}
	if err := unix.Mkfifo(stored, 0600); err != nil {
		t.Fatal(err)
	}
	if err := s.Restore(id); !errors.Is(err, ErrRestoreIntegrity) {
		t.Fatalf("FIFO accepted: %v", err)
	}
	if _, err := os.Lstat(original); !os.IsNotExist(err) {
		t.Fatal("FIFO published", err)
	}
}

func TestRestoreRejectsOversizedSparseSource(t *testing.T) {
	s, _, id, original, stored := restoreFixture(t)
	if err := os.Truncate(stored, maxRestoreBytes+1); err != nil {
		t.Fatal(err)
	}
	if err := s.Restore(id); !errors.Is(err, ErrRestoreIntegrity) {
		t.Fatalf("oversized input accepted: %v", err)
	}
	if _, err := os.Lstat(original); !os.IsNotExist(err) {
		t.Fatal("oversized input published", err)
	}
	noRestoreStages(t, s.dir, filepath.Dir(original))
}

func TestRestoreCopyFailuresNeverPublishPartialFile(t *testing.T) {
	for _, kind := range []string{"copy-error", "source-mutated-during-copy", "short-copy"} {
		t.Run(kind, func(t *testing.T) {
			s, c, id, original, stored := restoreFixture(t)
			s.restoreCopy = func(out io.Writer, in io.Reader) (int64, error) {
				switch kind {
				case "copy-error":
					n, err := io.CopyN(out, in, 3)
					if err != nil {
						return n, err
					}
					return n, io.ErrUnexpectedEOF
				case "short-copy":
					return io.CopyN(out, in, 3)
				default:
					if err := os.WriteFile(stored, []byte("altered"), 0600); err != nil {
						return 0, err
					}
					return io.Copy(out, in)
				}
			}
			if err := s.Restore(id); err == nil {
				t.Fatal("unsafe copy succeeded")
			}
			if _, err := os.Lstat(original); !os.IsNotExist(err) {
				t.Fatal("partial output published", err)
			}
			if _, err := os.Stat(stored); err != nil {
				t.Fatal("source lost", err)
			}
			if rec, _, err := c.GetQuarantine(id); err != nil || rec.Restored {
				t.Fatal("failed copy marked restored", err)
			}
			noRestoreStages(t, s.dir, filepath.Dir(original))
		})
	}
}

func TestRestorePublishesVerifiedSnapshotDespiteRetainedWriter(t *testing.T) {
	s, c, id, original, stored := restoreFixture(t)
	writer, err := os.OpenFile(stored, os.O_WRONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer writer.Close()
	s.restoreBeforePublish = func() error { _, err := writer.WriteAt([]byte("altered"), 0); return err }
	if err := s.Restore(id); err != nil {
		t.Fatal(err)
	}
	if body, err := os.ReadFile(original); err != nil || string(body) != "payload" {
		t.Fatalf("unverified bytes published: %q %v", body, err)
	}
	if rec, _, err := c.GetQuarantine(id); err != nil || !rec.Restored {
		t.Fatal("published snapshot missing record", err)
	}
	if _, err := os.Stat(stored); !os.IsNotExist(err) {
		t.Fatal("original stored inode not retired", err)
	}
	noRestoreStages(t, s.dir, filepath.Dir(original))
}

func TestRestorePreservesRacedStoredReplacement(t *testing.T) {
	s, c, id, original, stored := restoreFixture(t)
	displaced := filepath.Join(s.dir, "displaced-test-source")
	s.restoreBeforePublish = func() error {
		if err := os.Rename(stored, displaced); err != nil {
			return err
		}
		return os.WriteFile(stored, []byte("foreign entry"), 0600)
	}
	if err := s.Restore(id); !errors.Is(err, ErrRestoreIntegrity) {
		t.Fatalf("replacement race not reported: %v", err)
	}
	for path, want := range map[string]string{original: "payload", stored: "foreign entry", displaced: "payload"} {
		if body, err := os.ReadFile(path); err != nil || string(body) != want {
			t.Fatalf("file lost/replaced: %s %q %v", path, body, err)
		}
	}
	if rec, _, err := c.GetQuarantine(id); err != nil || !rec.Restored {
		t.Fatal("actual publication not recorded", err)
	}
	noRestoreStages(t, s.dir, filepath.Dir(original))
}

func TestRestoreDestinationAppearingAtCommitSurvives(t *testing.T) {
	s, c, id, original, stored := restoreFixture(t)
	s.restoreBeforePublish = func() error { return os.WriteFile(original, []byte("new benign file"), 0600) }
	if err := s.Restore(id); err == nil {
		t.Fatal("occupied destination accepted")
	}
	for path, want := range map[string]string{original: "new benign file", stored: "payload"} {
		if body, err := os.ReadFile(path); err != nil || string(body) != want {
			t.Fatalf("file changed: %s %q %v", path, body, err)
		}
	}
	if rec, _, err := c.GetQuarantine(id); err != nil || rec.Restored {
		t.Fatal("failed publication marked restored", err)
	}
	noRestoreStages(t, s.dir, filepath.Dir(original))
}

func TestRestoreRejectsRenamedDestinationParent(t *testing.T) {
	s, c, id, original, stored := restoreFixture(t)
	base := filepath.Dir(original)
	parent := filepath.Join(base, "destination")
	if err := os.Mkdir(parent, 0700); err != nil {
		t.Fatal(err)
	}
	rec, _, err := c.GetQuarantine(id)
	if err != nil {
		t.Fatal(err)
	}
	rec.OriginalPath = filepath.Join(parent, "restored.txt")
	if err := c.StoreQuarantine(rec); err != nil {
		t.Fatal(err)
	}
	moved := parent + "-moved"
	s.restoreBeforePublish = func() error {
		if err := os.Rename(parent, moved); err != nil {
			return err
		}
		if err := os.Mkdir(parent, 0700); err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(parent, "sentinel"), []byte("benign"), 0600)
	}
	if err := s.Restore(id); err == nil {
		t.Fatal("renamed parent accepted")
	}
	for _, dir := range []string{parent, moved} {
		if _, err := os.Lstat(filepath.Join(dir, "restored.txt")); !os.IsNotExist(err) {
			t.Fatal("unexpected publication", dir, err)
		}
	}
	if body, err := os.ReadFile(stored); err != nil || string(body) != "payload" {
		t.Fatal("source lost", err)
	}
	if body, err := os.ReadFile(filepath.Join(parent, "sentinel")); err != nil || string(body) != "benign" {
		t.Fatal("replacement directory affected", err)
	}
	noRestoreStages(t, s.dir, parent, moved)
}

func TestRestoreRefusesUnprotectedQuarantineDirectory(t *testing.T) {
	s, _, id, original, _ := restoreFixture(t)
	if err := os.Chmod(s.dir, 0777); err != nil {
		t.Fatal(err)
	}
	if err := s.Restore(id); !errors.Is(err, ErrRestoreIntegrity) {
		t.Fatalf("unprotected store accepted: %v", err)
	}
	if _, err := os.Stat(original); !os.IsNotExist(err) {
		t.Fatal("unprotected restore published", err)
	}
}

func TestRestoreLinkFallbackCommitsWithoutReplacing(t *testing.T) {
	for _, occupied := range []bool{false, true} {
		t.Run(fmt.Sprint(occupied), func(t *testing.T) {
			parent, err := openRestoreDirectory(t.TempDir(), false)
			if err != nil {
				t.Fatal(err)
			}
			defer parent.Close()
			stage, name, err := parent.staging()
			if err != nil {
				t.Fatal(err)
			}
			f, err := stage.open("snapshot", unix.O_WRONLY|unix.O_CREAT|unix.O_EXCL, 0600)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := f.Write([]byte("verified")); err != nil {
				t.Fatal(err)
			}
			f.Close()
			if occupied {
				if err := os.WriteFile(filepath.Join(parent.path, "target"), []byte("benign"), 0600); err != nil {
					t.Fatal(err)
				}
			}
			moved, err := publishRestoreSnapshot(stage, "snapshot", parent, "target", func(*restoreDirectory, string, *restoreDirectory, string) error { return unix.ENOSYS })
			if moved || occupied != (err != nil) {
				t.Fatalf("incorrect link commit result: moved=%v error=%v", moved, err)
			}
			if err := cleanupRestoreStage(parent, stage, name, "snapshot"); err != nil {
				t.Fatal(err)
			}
			want := "verified"
			if occupied {
				want = "benign"
			}
			if body, err := os.ReadFile(filepath.Join(parent.path, "target")); err != nil || string(body) != want {
				t.Fatalf("published target lost/overwritten: %q %v", body, err)
			}
			noRestoreStages(t, parent.path)
		})
	}
}

func TestRestorePreservesOrdinaryPOSIXFilenameCharacters(t *testing.T) {
	s, c, id, original, _ := restoreFixture(t)
	rec, _, err := c.GetQuarantine(id)
	if err != nil {
		t.Fatal(err)
	}
	rec.OriginalPath = filepath.Join(filepath.Dir(original), `report:part\one.txt`)
	if err := c.StoreQuarantine(rec); err != nil {
		t.Fatal(err)
	}
	if err := s.Restore(id); err != nil {
		t.Fatal(err)
	}
	if body, err := os.ReadFile(rec.OriginalPath); err != nil || string(body) != "payload" {
		t.Fatal("ordinary POSIX filename lost", err)
	}
}
