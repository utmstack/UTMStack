//go:build windows && edr_acceptance

package quarantine

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/utmstack/UTMStack/agent/edr/cache"
)

type releaseGuestResetContext struct {
	Phase     string
	ID        string
	Root      string
	DB        string
	Original  string
	Stored    string
	Raw       string
	RawID     string
	RawSynced bool
	Approval  RecoveryReleaseApproval
}

func TestReviewReleaseGuestResetDiagnose(t *testing.T) {
	base := guestResetBase(t)
	body, err := os.ReadFile(filepath.Join(base, "ready.json"))
	if err != nil || len(body) > 16384 {
		t.Fatal("invalid owned diagnostic context", err)
	}
	var ctx releaseGuestResetContext
	if err := json.Unmarshal(body, &ctx); err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(filepath.Clean(ctx.Root), filepath.Clean(base)+string(os.PathSeparator)) {
		t.Fatal("diagnostic root escaped owned case")
	}
	root, err := openAdmissionRoot(ctx.Root, false)
	if err != nil {
		t.Fatal(err)
	}
	rootID, err := admissionFileID(root.file)
	root.Close()
	if err != nil || rootID != ctx.Approval.Original.RootID {
		t.Fatal("diagnostic root identity changed", err)
	}
	for name, path := range map[string]string{"raw": ctx.Raw, "snapshot": ctx.Stored} {
		if !strings.HasPrefix(filepath.Clean(path), filepath.Clean(base)+string(os.PathSeparator)) {
			t.Fatal("diagnostic artifact escaped owned case")
		}
		f, err := os.Open(path)
		if err != nil {
			t.Fatal(err)
		}
		entry, readErr := inspectJournalFile(JournalInspection{}, f, maxAdmissionJournal)
		closeErr := f.Close()
		if err := errors.Join(readErr, closeErr); err != nil {
			t.Fatal(err)
		}
		t.Logf("read_only_artifact=%s file_id=%s original_raw_id_match=%t sha256=%s bytes=%q context_sha256=%x", name, entry.ManifestID, entry.ManifestID == ctx.RawID, entry.SHA256, entry.Manifest, sha256.Sum256(body))
	}
}

func TestReviewReleaseGuestResetPrepare(t *testing.T) {
	base := guestResetBase(t)
	phase := os.Getenv("OPENEDR_GUEST_RESET_PHASE")
	if phase != "release-approved" && phase != "release-published" {
		t.Fatal("unsupported release guest-reset phase")
	}
	// The controller sets TMP/TEMP/SystemTemp beneath this private root. Process
	// termination by guest reset leaves the fixture for a later boot to inspect.
	s, id, original, stored, a := reviewReleaseFixture(t, "verified")
	if !strings.HasPrefix(filepath.Clean(s.dir), filepath.Clean(base)+string(os.PathSeparator)) {
		t.Fatal("fixture escaped owned persistent reset root")
	}
	rec, found, err := s.cache.GetQuarantine(id)
	if err != nil || !found {
		t.Fatal("missing reset fixture", err)
	}
	raw := filepath.Join(rec.SourceCaptureDir, "source")
	f, err := os.Open(raw)
	if err != nil {
		t.Fatal(err)
	}
	rawID, err := admissionFileID(f)
	f.Close()
	if err != nil {
		t.Fatal(err)
	}
	// Isolate release durability from the admission fixture's earlier buffered
	// source write. A separate unflushed control lost raw-source data at reset
	// while retaining its durable snapshot; that admission gap remains open.
	stable, err := os.OpenFile(raw, os.O_RDWR, 0)
	if err != nil {
		t.Fatal(err)
	}
	stableID, idErr := admissionFileID(stable)
	syncErr := stable.Sync()
	closeErr := stable.Close()
	if err := errors.Join(idErr, syncErr, closeErr); err != nil || stableID != rawID {
		t.Fatal("cannot establish durable owned raw-source fixture", err)
	}
	if err := s.cache.Close(); err != nil {
		t.Fatal(err)
	}
	ctx := releaseGuestResetContext{Phase: phase, ID: id, Root: s.dir, DB: filepath.Join(filepath.Dir(s.dir), "edr.db"), Original: original, Stored: stored, Raw: raw, RawID: rawID, RawSynced: true, Approval: a}
	_, err = releaseRecovery(s.dir, id, a, "owned guest reset explicit release", func(at string) error {
		if at != phase {
			return nil
		}
		body, err := json.Marshal(ctx)
		if err != nil {
			return err
		}
		f, err := os.OpenFile(filepath.Join(base, "ready.json"), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
		if err != nil {
			return err
		}
		_, writeErr := f.Write(body)
		syncErr := f.Sync()
		closeErr := f.Close()
		if err := errors.Join(writeErr, syncErr, closeErr); err != nil {
			return err
		}
		// There is no environment-triggered fault in shipped programs. Only this
		// owned test pauses while the external controller forcibly resets its VM.
		time.Sleep(20 * time.Minute)
		return errors.New("guest-reset controller exceeded the owned test deadline")
	})
	t.Fatalf("release reset checkpoint returned: %v", err)
}

func TestReviewReleaseGuestResetRecover(t *testing.T) {
	base := guestResetBase(t)
	body, err := os.ReadFile(filepath.Join(base, "ready.json"))
	if err != nil || len(body) > 16384 {
		t.Fatal("invalid reset context", err)
	}
	var ctx releaseGuestResetContext
	if err := json.Unmarshal(body, &ctx); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{ctx.Root, ctx.DB, ctx.Original, ctx.Stored, ctx.Raw} {
		if !strings.HasPrefix(filepath.Clean(path), filepath.Clean(base)+string(os.PathSeparator)) {
			t.Fatal("reset context escaped owned root")
		}
	}
	if !ctx.RawSynced || !ordinaryQuarantineID(ctx.ID) || ctx.DB != filepath.Join(filepath.Dir(ctx.Root), "edr.db") || ctx.Stored != filepath.Join(ctx.Root, ctx.ID+".quarantined") || validRecoveryReleaseApproval(ctx.Approval) != nil {
		t.Fatal("changed owned reset context identity")
	}
	moved := ctx.Phase == "release-published"
	if !moved && ctx.Phase != "release-approved" {
		t.Fatal("invalid release checkpoint")
	}
	report, err := InspectRecovery(ctx.Root, ctx.ID)
	if err != nil || report.RootID != ctx.Approval.Original.RootID {
		t.Fatal("changed reset root", err)
	}
	original := report.Journals[1]
	if original.DirectoryID != ctx.Approval.Original.DirectoryID || original.ManifestID != ctx.Approval.Original.ManifestID || original.SHA256 != ctx.Approval.Original.SHA256 {
		t.Fatal("release or reset changed original restore journal")
	}
	if (report.Journals[2].Status == "absent") != moved {
		t.Fatal("unexpected fence state after reset", report.Journals[2])
	}
	c, err := cache.Open(ctx.DB)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	s, err := New(ctx.Root, c)
	if err != nil {
		t.Fatal(err)
	}
	if !moved {
		for _, result := range s.RecoveryStatus() {
			if result.Status != "operator_held" {
				t.Fatal("unreleased hold allowed recovery after reset", result)
			}
		}
	} else if _, err := os.Lstat(filepath.Join(ctx.Root, ".restores", ctx.ID)); !os.IsNotExist(err) {
		t.Fatal("released pre-publication operation did not recover", err)
	}
	decision, err := ReleaseRecovery(ctx.Root, ctx.ID, ctx.Approval, "owned guest reset explicit release")
	if err != nil {
		t.Fatal("release could not stabilize after reset", err)
	}
	if decision.Approval != ctx.Approval {
		t.Fatal("release changed its original approval")
	}
	rec, found, err := c.GetQuarantine(ctx.ID)
	if err != nil || !found || rec.Restored || rec.Purged || ctx.Raw != filepath.Join(rec.SourceCaptureDir, "source") {
		t.Fatal("reset release claimed original completion", rec, err)
	}
	requirePurgePayload(t, ctx.Stored, "payload")
	requirePurgePayload(t, ctx.Raw, "payload")
	f, err := os.Open(ctx.Raw)
	if err != nil {
		t.Fatal(err)
	}
	rawID, err := admissionFileID(f)
	f.Close()
	if err != nil || rawID != ctx.RawID {
		t.Fatal("raw capture identity changed across reset", err)
	}
	if _, err := os.Lstat(ctx.Original); !os.IsNotExist(err) {
		t.Fatal("release published a destination", err)
	}
	t.Logf("forced_guest_reset phase=%s moved_before_reset=%t original_completion=false raw_identity_preserved=true", ctx.Phase, moved)
}
