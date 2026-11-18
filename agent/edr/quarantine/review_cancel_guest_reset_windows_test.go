//go:build windows && edr_acceptance

package quarantine

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/utmstack/UTMStack/agent/edr/cache"
)

type cancelResetArtifact struct {
	Path, FileID, SHA256 string
	Bytes                int64
}

type cancelResetContext struct {
	Phase, ID, Root, DB, Original, Stage, Raw, Stored string
	RawSynced                                         bool
	Approval                                          RecoveryReleaseApproval
	Artifacts                                         map[string]cancelResetArtifact
	ReadyUTC                                          time.Time
	ReadyFileID                                       string
}

func inspectCancelResetArtifact(t *testing.T, path string, file *os.File) cancelResetArtifact {
	t.Helper()
	if file == nil {
		var err error
		file, err = os.Open(path)
		if err != nil {
			t.Fatal(err)
		}
		defer file.Close()
	}
	id, err := admissionFileID(file)
	if err != nil {
		t.Fatal(err)
	}
	h := sha256.New()
	n, err := io.Copy(h, io.LimitReader(file, maxRecoveryRelease+1))
	if err != nil || n > maxRecoveryRelease {
		t.Fatal("invalid owned cancellation reset artifact", err)
	}
	return cancelResetArtifact{path, id, hex.EncodeToString(h.Sum(nil)), n}
}

func TestReviewCancelGuestResetPrepare(t *testing.T) {
	base := guestResetBase(t)
	phase := os.Getenv("OPENEDR_GUEST_RESET_PHASE")
	if phase != "cancel-approved" && phase != "cancel-archived" {
		t.Fatal("unsupported cancellation reset boundary")
	}
	s, id, original, stored, a := reviewReleaseFixture(t, "publication-intent")
	if !strings.HasPrefix(filepath.Clean(s.dir), base+string(os.PathSeparator)) {
		t.Fatal("cancellation fixture escaped owned case")
	}
	manifest := restoreFixtureManifest(t, s, id)
	rec, found, err := s.cache.GetQuarantine(id)
	if err != nil || !found {
		t.Fatal("missing cancellation fixture", err)
	}
	raw := filepath.Join(rec.SourceCaptureDir, "source")
	f, err := os.Open(raw)
	if err != nil {
		t.Fatal(err)
	}
	// Isolate cancellation metadata from source admission. This control has
	// an explicitly flushed raw starting artifact and claims no new admission
	// barrier or hardware power-loss qualification.
	err = admissionSyncSource(f)
	err = errors.Join(err, f.Close())
	if err != nil {
		t.Fatal("cannot establish durable raw starting artifact", err)
	}
	unknown := filepath.Join(s.dir, ".restores", id, "unknown-cancel-evidence")
	marker, err := os.OpenFile(unknown, os.O_CREATE|os.O_EXCL|os.O_RDWR, 0600)
	if err != nil {
		t.Fatal(err)
	}
	_, err = marker.Write([]byte("preserve unknown original child"))
	err = errors.Join(err, marker.Sync(), marker.Close())
	if err != nil {
		t.Fatal(err)
	}
	if err := s.cache.Close(); err != nil {
		t.Fatal(err)
	}
	ctx := cancelResetContext{Phase: phase, ID: id, Root: s.dir, DB: filepath.Join(filepath.Dir(s.dir), "edr.db"), Original: original, Stage: filepath.Join(filepath.Dir(original), manifest.stageName(), "snapshot"), Raw: raw, Stored: stored, RawSynced: true, Approval: a, Artifacts: map[string]cancelResetArtifact{}}
	_, err = cancelRecovery(s.dir, id, a, "owned guest reset explicit cancellation", func(at string) error {
		if at != phase {
			return nil
		}
		journal := filepath.Join(s.dir, ".restores", id, "manifest.json")
		if phase == "cancel-archived" {
			journal = filepath.Join(s.dir, ".reviews", a.HoldDecisionID, "restore", "manifest.json")
			unknown = filepath.Join(filepath.Dir(journal), "unknown-cancel-evidence")
		}
		paths := map[string]string{"raw": raw, "stored": stored, "stage": ctx.Stage, "original": journal, "unknown": unknown, "fence": filepath.Join(s.dir, ".purges", id)}
		for name, path := range paths {
			ctx.Artifacts[name] = inspectCancelResetArtifact(t, path, nil)
		}
		root, err := s.openRoot()
		if err != nil {
			return err
		}
		defer root.Close()
		history, err := admissionOpenChild(root, ".reviews")
		if err != nil {
			return err
		}
		defer history.Close()
		op, err := admissionOpenChild(history, a.HoldDecisionID)
		if err != nil {
			return err
		}
		defer op.Close()
		// Accommodate the creator's still-open DELETE/write receipt handle.
		receipt, err := openRecoveryReleaseReceipt(op)
		if err != nil {
			return err
		}
		ctx.Artifacts["receipt"] = inspectCancelResetArtifact(t, filepath.Join(op.path, "manifest.json"), receipt)
		receipt.Close()
		ready, err := os.OpenFile(filepath.Join(base, "ready.tmp"), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
		if err != nil {
			return err
		}
		ctx.ReadyFileID, err = admissionFileID(ready)
		if err != nil {
			ready.Close()
			return err
		}
		ctx.ReadyUTC = time.Now().UTC()
		body, err := json.Marshal(ctx)
		if err != nil || len(body) > 65536 {
			ready.Close()
			return errors.Join(err, errors.New("invalid reset context size"))
		}
		_, err = ready.Write(body)
		err = errors.Join(err, ready.Sync(), ready.Close())
		if err != nil {
			return err
		}
		if err := os.Rename(filepath.Join(base, "ready.tmp"), filepath.Join(base, "ready.json")); err != nil {
			return err
		}
		time.Sleep(20 * time.Minute)
		return errors.New("owned cancellation reset observation deadline")
	})
	t.Fatalf("cancellation reset boundary returned: %v", err)
}

func TestReviewCancelGuestResetRecover(t *testing.T) {
	base := guestResetBase(t)
	ready, err := os.Open(filepath.Join(base, "ready.json"))
	if err != nil {
		t.Fatal(err)
	}
	readyID, idErr := admissionFileID(ready)
	body, err := io.ReadAll(io.LimitReader(ready, 65537))
	err = errors.Join(err, idErr, ready.Close())
	if err != nil || len(body) > 65536 {
		t.Fatal("invalid owned cancellation reset context", err)
	}
	sum := sha256.Sum256(body)
	if expected := os.Getenv("OPENEDR_CANCEL_RESET_CONTEXT_SHA"); len(expected) != 64 || expected != hex.EncodeToString(sum[:]) {
		t.Fatal("cancellation reset context differs from external pre-reset checkpoint")
	}
	var ctx cancelResetContext
	if err := json.Unmarshal(body, &ctx); err != nil {
		t.Fatal(err)
	}
	if !ctx.RawSynced || ctx.ReadyUTC.IsZero() || ctx.ReadyFileID != readyID || !ordinaryQuarantineID(ctx.ID) || validRecoveryReleaseApproval(ctx.Approval) != nil || len(ctx.Artifacts) != 7 || ctx.DB != filepath.Join(filepath.Dir(ctx.Root), "edr.db") || ctx.Stored != filepath.Join(ctx.Root, ctx.ID+".quarantined") || ctx.Phase != "cancel-approved" && ctx.Phase != "cancel-archived" {
		t.Fatal("changed cancellation reset context")
	}
	for _, path := range []string{ctx.Root, ctx.DB, ctx.Original, ctx.Stage, ctx.Stored, ctx.Raw} {
		if !strings.HasPrefix(filepath.Clean(path), base+string(os.PathSeparator)) {
			t.Fatal("cancellation reset context escaped owned root")
		}
	}
	// Validate every retained byte and identity before SQLite or recovery opens.
	for name, want := range ctx.Artifacts {
		if !strings.HasPrefix(filepath.Clean(want.Path), base+string(os.PathSeparator)) {
			t.Fatal("artifact escaped owned case", name)
		}
		if got := inspectCancelResetArtifact(t, want.Path, nil); got != want {
			t.Fatal("cancellation reset changed retained artifact", name, got, want)
		}
	}
	if ctx.Artifacts["raw"].FileID == ctx.Artifacts["stored"].FileID {
		t.Fatal("raw and durable snapshot unexpectedly share an identity")
	}
	root, err := openAdmissionRoot(ctx.Root, false)
	if err != nil {
		t.Fatal(err)
	}
	op, m, err := openCancellationProof(root, ctx.ID, ctx.Approval)
	if err != nil {
		root.Close()
		t.Fatal(err)
	}
	archived, err := cancellationOriginal(root, op, m)
	op.Close()
	root.Close()
	if err != nil || archived != (ctx.Phase == "cancel-archived") {
		t.Fatal("cancellation reset changed operation location", archived, err)
	}
	decision, err := CancelRecovery(ctx.Root, ctx.ID, ctx.Approval, "owned guest reset explicit cancellation")
	if err != nil {
		t.Fatal(err)
	}
	if again, err := CancelRecovery(ctx.Root, ctx.ID, ctx.Approval, decision.Reason); err != nil || again.DecisionID != decision.DecisionID || again.FileID != decision.FileID {
		t.Fatal("post-reset cancellation retry did not stabilize", err)
	}
	for name, want := range ctx.Artifacts {
		if ctx.Phase == "cancel-approved" && (name == "original" || name == "unknown") {
			want.Path = filepath.Join(ctx.Root, ".reviews", ctx.Approval.HoldDecisionID, "restore", filepath.Base(want.Path))
		}
		if got := inspectCancelResetArtifact(t, want.Path, nil); got != want {
			t.Fatal("cancellation retry changed retained artifact", name, got, want)
		}
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
	if err := s.Restore(ctx.ID); !errors.Is(err, ErrRecoveryHeld) {
		t.Fatal("post-reset cancellation allowed restore", err)
	}
	if err := s.Purge(ctx.ID); !errors.Is(err, ErrRecoveryHeld) {
		t.Fatal("post-reset cancellation allowed purge", err)
	}
	if qid, err := s.Quarantine(ctx.Stage, payloadHash("payload"), "Safe.Reset.Cancelled.Stage"); qid != "" || !errors.Is(err, ErrAlreadyQuarantined) {
		t.Fatal("post-reset archived stage was re-admitted", qid, err)
	}
	if rec, found, err := c.GetQuarantine(ctx.ID); err != nil || !found || rec.Restored || rec.Purged || !rec.Restorable {
		t.Fatal("post-reset cancellation claimed historical completion", err)
	}
	for _, path := range []string{ctx.Raw, ctx.Stored, ctx.Stage} {
		requirePurgePayload(t, path, "payload")
	}
	if _, err := os.Lstat(ctx.Original); !os.IsNotExist(err) {
		t.Fatal("cancellation reset published original destination", err)
	}
	t.Logf("cancellation reset verified: phase=%s original_directory=%s fence=%s receipt=%s raw=%s snapshot=%s", ctx.Phase, ctx.Approval.Original.DirectoryID, ctx.Artifacts["fence"].FileID, ctx.Artifacts["receipt"].FileID, ctx.Artifacts["raw"].FileID, ctx.Artifacts["stored"].FileID)
}
