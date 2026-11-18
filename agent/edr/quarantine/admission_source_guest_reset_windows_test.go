//go:build windows && edr_acceptance

package quarantine

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/utmstack/UTMStack/agent/edr/cache"
	"golang.org/x/sys/windows"
)

type sourceResetBarrier struct {
	FileID       string
	Microseconds int64
}
type sourceResetContext struct {
	Schema                                                        int
	Case, Phase, Root, DB, Original, SourceParent, SourceParentID string
	QuarantineRoot, QuarantineRootID, QuarantineID                string
	Raw, RawID, CaptureID, Snapshot, SnapshotID                   string
	Journal, JournalID, JournalSHA                                string
	FreshFixtureWrite, FixtureSourceSync                          bool
	Barriers                                                      []sourceResetBarrier
}

func sourceResetCase(t *testing.T) (name, phase string, other bool) {
	t.Helper()
	name = os.Getenv("OPENEDR_SOURCE_RESET_CASE")
	switch name {
	case "intent-normal":
		return name, "intent", false
	case "claimed-normal":
		return name, "claimed", false
	case "intent-readonly-other":
		return name, "intent", true
	case "claimed-readonly-other":
		return name, "claimed", true
	default:
		t.Fatal("unknown owned source reset case")
		return "", "", false
	}
}
func sourceResetPrivate(t *testing.T, path string) *restoreDirectory {
	t.Helper()
	if !filepath.IsAbs(path) || filepath.Clean(path) != path {
		t.Fatal("noncanonical source reset root")
	}
	d, e := openAdmissionRoot(path, false)
	if e != nil {
		t.Fatal("owned private source reset root required", e)
	}
	return d
}

// Inspection accommodates the continuously held production append/delete pin.
// It grants no write/delete access and does not relax the original source pin.
func sourceResetOpen(t *testing.T, path string) *os.File {
	t.Helper()
	name, err := windows.UTF16PtrFromString(path)
	if err != nil {
		t.Fatal(err)
	}
	h, err := windows.CreateFile(name, windows.GENERIC_READ,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE,
		nil, windows.OPEN_EXISTING,
		windows.FILE_FLAG_BACKUP_SEMANTICS|windows.FILE_FLAG_OPEN_REPARSE_POINT, 0)
	if err != nil {
		t.Fatal(err)
	}
	return os.NewFile(uintptr(h), path)
}
func sourceResetPayload(t *testing.T, path string) {
	t.Helper()
	f := sourceResetOpen(t, path)
	defer f.Close()
	body := make([]byte, len("payload")+1)
	n, err := f.ReadAt(body, 0)
	if err != io.EOF || string(body[:n]) != "payload" {
		t.Fatalf("source reset payload differs: %q %v", body[:n], err)
	}
}
func sourceResetID(t *testing.T, path string) string {
	t.Helper()
	f := sourceResetOpen(t, path)
	defer f.Close()
	id, e := admissionFileID(f)
	if e != nil {
		t.Fatal(e)
	}
	return id
}
func sourceResetArtifact(t *testing.T, path, expectedID string) {
	t.Helper()
	actualID := sourceResetID(t, path)
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("source_reset_artifact path=%s file_id=%s expected_id_match=%t sha256=%s bytes=%q", path, actualID, actualID == expectedID, sourceResetSHA(body), body)
	if actualID != expectedID || string(body) != "payload" {
		t.Fatal("source reset artifact identity/content differs")
	}
}
func sourceResetSHA(body []byte) string { h := sha256.Sum256(body); return hex.EncodeToString(h[:]) }
func sourceResetWrite(t *testing.T, d *restoreDirectory, name string, body []byte) {
	t.Helper()
	f, e := admissionCreateFile(d, name)
	if e != nil {
		t.Fatal(e)
	}
	_, e = f.Write(body)
	if e == nil {
		e = f.Sync()
	}
	e = errors.Join(e, f.Close())
	if e != nil {
		t.Fatal(e)
	}
}
func TestAdmissionSourceGuestResetPrepare(t *testing.T) {
	if os.Getenv("EDR_WINDOWS_ACCEPTANCE") != "isolated-vm" || os.Getenv("OPENEDR_SOURCE_RESET_MODE") != "prepare" {
		t.Skip("owned disposable Windows source reset worker")
	}
	name, phase, other := sourceResetCase(t)
	base := os.Getenv("OPENEDR_SOURCE_RESET_ROOT")
	private := sourceResetPrivate(t, base)
	defer private.Close()
	entries, e := os.ReadDir(base)
	if e != nil || len(entries) != 0 {
		t.Fatal("refuse nonempty owned source reset case", e)
	}
	parent := base
	if other {
		parent = os.Getenv("OPENEDR_SOURCE_RESET_OTHER")
		d := sourceResetPrivate(t, parent)
		d.Close()
		entries, e = os.ReadDir(parent)
		if e != nil || len(entries) != 0 {
			t.Fatal("refuse nonempty source data directory", e)
		}
	}
	db := filepath.Join(base, "edr.db")
	c, e := cache.Open(db)
	if e != nil {
		t.Fatal(e)
	}
	defer c.Close()
	qroot := filepath.Join(base, "quarantine")
	s, e := New(qroot, c)
	if e != nil {
		t.Fatal(e)
	}
	// All database/root preparation precedes this fresh write. The fixture never
	// calls Sync on the raw input: only the production source barriers may do so.
	original := filepath.Join(parent, "source.txt")
	if e = os.WriteFile(original, []byte("payload"), 0600); e != nil {
		t.Fatal(e)
	}
	if other {
		p, _ := windows.UTF16PtrFromString(original)
		if e = windows.SetFileAttributes(p, windows.FILE_ATTRIBUTE_READONLY); e != nil {
			t.Fatal(e)
		}
	}
	sourceID := sourceResetID(t, original)
	scan := sourceResetOpen(t, original)
	scanned, e := scan.Stat()
	e = errors.Join(e, scan.Close())
	if e != nil {
		t.Fatal(e)
	}
	var barriers []sourceResetBarrier
	s.admissionSync = func(f *os.File) error {
		id, e := admissionFileID(f)
		if e != nil || id != sourceID {
			return errors.Join(e, errors.New("source reset barrier lost identity"))
		}
		at := time.Now()
		e = admissionSyncSource(f)
		if e == nil {
			barriers = append(barriers, sourceResetBarrier{id, time.Since(at).Microseconds()})
		}
		return e
	}
	s.admissionFault = func(at string) error {
		if at != phase {
			return nil
		}
		count := 1
		if phase == "claimed" {
			count = 2
		}
		if len(barriers) != count {
			t.Fatal("missing production source barrier", barriers)
		}
		names, e := os.ReadDir(filepath.Join(qroot, ".admissions"))
		if e != nil || len(names) != 1 || !ordinaryQuarantineID(names[0].Name()) {
			t.Fatal("ambiguous source reset journal", e)
		}
		id := names[0].Name()
		journal := filepath.Join(qroot, ".admissions", id, "manifest.json")
		body, e := os.ReadFile(journal)
		if e != nil {
			t.Fatal(e)
		}
		var m admissionManifest
		if e = json.Unmarshal(body, &m); e != nil {
			t.Fatal(e)
		}
		want := "preparing"
		raw := original
		snapshot := ""
		if phase == "claimed" {
			want = "claimed"
			raw = filepath.Join(m.Record.SourceCaptureDir, "source")
			snapshot = filepath.Join(filepath.Dir(journal), "snapshot")
		}
		if m.Phase != want || m.Record.SourceFileID != sourceID {
			t.Fatal("wrong source reset boundary", m)
		}
		sourceResetPayload(t, raw)
		if sourceResetID(t, raw) != sourceID {
			t.Fatal("source reset raw identity changed")
		}
		ctx := sourceResetContext{Schema: 1, Case: name, Phase: phase, Root: base, DB: db, Original: original, SourceParent: parent, SourceParentID: m.ParentID, QuarantineRoot: qroot, QuarantineRootID: sourceResetID(t, qroot), QuarantineID: id, Raw: raw, RawID: sourceID, Snapshot: snapshot, SnapshotID: m.SnapshotID, Journal: journal, JournalID: sourceResetID(t, journal), JournalSHA: sourceResetSHA(body), FreshFixtureWrite: true, FixtureSourceSync: false, Barriers: barriers}
		if phase == "claimed" {
			ctx.CaptureID = m.Record.SourceCaptureDirID
			sourceResetPayload(t, snapshot)
			if sourceResetID(t, snapshot) != m.SnapshotID || m.SnapshotID == sourceID {
				t.Fatal("source snapshot is not independently identified")
			}
		}
		encoded, e := json.Marshal(ctx)
		if e != nil {
			t.Fatal(e)
		}
		encoded = append(encoded, '\n')
		sourceResetWrite(t, private, "source-context.json", encoded)
		ready := struct {
			Case, Phase, ContextSHA, ContextID string
			ReadyUTC                           time.Time
		}{name, phase, sourceResetSHA(encoded), sourceResetID(t, filepath.Join(base, "source-context.json")), time.Now().UTC()}
		encoded, e = json.Marshal(ready)
		if e != nil {
			t.Fatal(e)
		}
		sourceResetWrite(t, private, "source-ready.json", append(encoded, '\n'))
		fmt.Printf("OWNED_SOURCE_RESET_READY case=%s phase=%s production_barriers=%d fixture_source_sync=false\n", name, phase, len(barriers))
		time.Sleep(5 * time.Minute)
		return errors.New("owned source reset controller deadline expired")
	}
	_, e = s.QuarantineVerified(original, payloadHash("payload"), "Safe.Source.Reset.Control", scanned)
	t.Fatal("source reset checkpoint returned without forced reset", e)
}
func TestAdmissionSourceGuestResetCheck(t *testing.T) {
	if os.Getenv("EDR_WINDOWS_ACCEPTANCE") != "isolated-vm" || os.Getenv("OPENEDR_SOURCE_RESET_MODE") != "check" {
		t.Skip("owned disposable Windows source reset verifier")
	}
	name, phase, other := sourceResetCase(t)
	base := os.Getenv("OPENEDR_SOURCE_RESET_ROOT")
	private := sourceResetPrivate(t, base)
	defer private.Close()
	path := filepath.Join(base, "source-context.json")
	body, e := os.ReadFile(path)
	if e != nil {
		t.Fatal(e)
	}
	if sourceResetSHA(body) != os.Getenv("OPENEDR_SOURCE_RESET_CONTEXT_SHA") || sourceResetID(t, path) != os.Getenv("OPENEDR_SOURCE_RESET_CONTEXT_ID") {
		t.Fatal("source reset context identity changed")
	}
	var ctx sourceResetContext
	if e = json.Unmarshal(body, &ctx); e != nil {
		t.Fatal(e)
	}
	parent := base
	if other {
		parent = os.Getenv("OPENEDR_SOURCE_RESET_OTHER")
	}
	expectedRaw := filepath.Join(parent, "source.txt")
	expectedSnapshot := ""
	count := 1
	if phase == "claimed" {
		expectedRaw = filepath.Join(parent, ".edr-quarantine-"+ctx.QuarantineID, "source")
		expectedSnapshot = filepath.Join(base, "quarantine", ".admissions", ctx.QuarantineID, "snapshot")
		count = 2
	}
	if ctx.Schema != 1 || ctx.Case != name || ctx.Phase != phase || ctx.Root != base || ctx.DB != filepath.Join(base, "edr.db") || ctx.QuarantineRoot != filepath.Join(base, "quarantine") || ctx.Original != filepath.Join(parent, "source.txt") || ctx.SourceParent != parent || ctx.Raw != expectedRaw || ctx.Snapshot != expectedSnapshot || ctx.Journal != filepath.Join(base, "quarantine", ".admissions", ctx.QuarantineID, "manifest.json") || !ordinaryQuarantineID(ctx.QuarantineID) || !ctx.FreshFixtureWrite || ctx.FixtureSourceSync || len(ctx.Barriers) != count {
		t.Fatal("source reset context differs from owned case")
	}
	for _, barrier := range ctx.Barriers {
		if barrier.FileID != ctx.RawID {
			t.Fatal("source barrier identity differs")
		}
	}
	qroot := sourceResetPrivate(t, ctx.QuarantineRoot)
	qroot.Close()
	sourceParent := sourceResetPrivate(t, parent)
	sourceParent.Close()
	if sourceResetID(t, ctx.QuarantineRoot) != ctx.QuarantineRootID || sourceResetID(t, parent) != ctx.SourceParentID {
		t.Fatal("source reset directory identity changed")
	}
	journal, e := os.ReadFile(ctx.Journal)
	if e != nil || sourceResetSHA(journal) != ctx.JournalSHA || sourceResetID(t, ctx.Journal) != ctx.JournalID {
		t.Fatal("reset changed the paused admission journal", e)
	}
	// These checks precede SQLite initialization and recovery. A failed artifact
	// remains untouched for read-only diagnosis rather than being repaired here.
	sourceResetArtifact(t, ctx.Raw, ctx.RawID)
	if phase == "claimed" {
		if sourceResetID(t, filepath.Dir(ctx.Raw)) != ctx.CaptureID {
			t.Fatal("capture identity changed")
		}
		sourceResetArtifact(t, ctx.Snapshot, ctx.SnapshotID)
		if ctx.SnapshotID == ctx.RawID {
			t.Fatal("snapshot identity changed")
		}
	}
	c, e := cache.Open(ctx.DB)
	if e != nil {
		t.Fatal(e)
	}
	defer c.Close()
	s, e := New(ctx.QuarantineRoot, c)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = s.Recover(); e != nil {
		t.Fatal("source reset recovery failed", e)
	}
	rec, found, e := c.GetQuarantine(ctx.QuarantineID)
	if e != nil {
		t.Fatal(e)
	}
	if phase == "intent" {
		if found {
			t.Fatal("pre-capture intent became committed", rec)
		}
		requirePurgePayload(t, ctx.Original, "payload")
		if sourceResetID(t, ctx.Original) != ctx.RawID {
			t.Fatal("aborted source identity changed")
		}
	} else {
		if !found || !rec.Restorable || rec.Restored || rec.Purged || rec.SourceFileID != ctx.RawID {
			t.Fatal("wrong captured recovery result", rec)
		}
		if _, e = os.Lstat(ctx.Original); !os.IsNotExist(e) {
			t.Fatal("captured source was published as an original", e)
		}
		stored := filepath.Join(ctx.QuarantineRoot, ctx.QuarantineID+".quarantined")
		requirePurgePayload(t, stored, "payload")
		if sourceResetID(t, stored) != ctx.SnapshotID {
			t.Fatal("published snapshot identity changed")
		}
		requirePurgePayload(t, ctx.Raw, "payload")
		if sourceResetID(t, ctx.Raw) != ctx.RawID {
			t.Fatal("retained raw identity changed")
		}
	}
	if _, e = os.Lstat(filepath.Dir(ctx.Journal)); !os.IsNotExist(e) {
		t.Fatal("source operation did not stabilize", e)
	}
	if _, e = s.Recover(); e != nil {
		t.Fatal("source reset replay is not idempotent", e)
	}
	t.Logf("forced_source_reset case=%s phase=%s raw_id=%s production_barriers=%d fixture_source_sync=false no_restore_or_purge=true", name, phase, ctx.RawID, count)
}
