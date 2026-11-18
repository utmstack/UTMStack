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
	"golang.org/x/sys/windows"
)

type freshResetContext struct {
	Version                                                           int
	Phase, ID, Root, DB, Destination, Stored, Raw, OldStage, NewStage string
	Approval                                                          FreshRestoreApproval
	ApprovalJSON                                                      []byte
	Cancellation                                                      FreshCancellationInspection
	HasCancellation, RawSynced                                        bool
	OperationDirectoryID, StageID, OutputID                           string
	Directories                                                       map[string]string
	Artifacts                                                         map[string]cancelResetArtifact
	ReadyUTC                                                          time.Time
	ReadyFileID                                                       string
}

func freshResetInside(base, path string) bool {
	relative, err := filepath.Rel(base, path)
	return err == nil && filepath.IsLocal(relative)
}

// Only the test observer permits sharing with the checkpoint's open writer.
// Production metadata, source and output pinning remains unchanged.
func freshResetArtifact(t *testing.T, path string) cancelResetArtifact {
	t.Helper()
	parent, err := platformRestoreDirectory(filepath.Dir(path))
	if err != nil {
		t.Fatal(err)
	}
	defer parent.Close()
	f, err := admissionNTFile(windows.Handle(parent.file.Fd()), filepath.Base(path), windows.FILE_GENERIC_READ, windows.FILE_OPEN, windows.FILE_NON_DIRECTORY_FILE, nil, windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	return inspectCancelResetArtifact(t, path, f)
}

func freshResetDirectory(t *testing.T, path string) string {
	t.Helper()
	d, err := openAdmissionRoot(path, false)
	if err != nil {
		t.Fatal(err)
	}
	defer d.Close()
	id, err := admissionFileID(d.file)
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func freshResetPhase(phase string) bool {
	switch phase {
	case "approval-synced", "ownership-written", "verified", "publication-intent", "published-filesystem", "database-committed", "fresh-cancellation-approved", "fresh-cancellation-archived":
		return true
	}
	return false
}

func TestFreshGuestResetPrepare(t *testing.T) {
	base := guestResetBase(t)
	phase := os.Getenv("OPENEDR_GUEST_RESET_PHASE")
	if !freshResetPhase(phase) {
		t.Fatal("unsupported fresh reset boundary")
	}
	s, id, destination, stored, a, approvalJSON := freshExecutionFixture(t)
	if !freshResetInside(base, s.dir) {
		t.Fatal("fixture escaped owned reset case")
	}
	old, err := decodeRestoreManifest(a.Inspection.Original.Manifest, id)
	if err != nil {
		t.Fatal(err)
	}
	var hold RecoveryHold
	if err = json.Unmarshal(a.Inspection.Hold.Manifest, &hold); err != nil {
		t.Fatal(err)
	}
	ctx := freshResetContext{Version: 1, Phase: phase, ID: id, Root: s.dir, DB: filepath.Join(filepath.Dir(s.dir), "edr.db"), Destination: destination, Stored: stored, Raw: filepath.Join(old.Record.SourceCaptureDir, "source"), OldStage: filepath.Join(string(old.ParentPath), old.stageName(), "snapshot"), NewStage: filepath.Join(filepath.Dir(destination), ".edr-restore-"+a.OperationID), Approval: a, ApprovalJSON: approvalJSON, Artifacts: map[string]cancelResetArtifact{}, Directories: map[string]string{}}
	raw, err := os.Open(ctx.Raw)
	if err != nil {
		t.Fatal(err)
	}
	err = errors.Join(admissionSyncSource(raw), raw.Close())
	if err != nil {
		t.Fatal("cannot flush the independent initial raw fixture", err)
	}
	ctx.RawSynced = true
	capture := func() {
		t.Helper()
		decision := filepath.Join(s.dir, ".fresh-restores", id, a.OperationID)
		op := filepath.Join(s.dir, ".restores", id)
		if phase == "fresh-cancellation-archived" {
			op = filepath.Join(decision, "restore")
		}
		paths := map[string]string{"raw": ctx.Raw, "stored": stored, "old-stage": ctx.OldStage, "fence": filepath.Join(s.dir, ".purges", id), "old-cancellation": filepath.Join(s.dir, ".reviews", hold.DecisionID, "manifest.json"), "old-journal": filepath.Join(s.dir, ".reviews", hold.DecisionID, "restore", "manifest.json"), "approval": filepath.Join(decision, "manifest.json")}
		if phase != "approval-synced" {
			ctx.OperationDirectoryID = freshResetDirectory(t, op)
			ctx.StageID = freshResetDirectory(t, ctx.NewStage)
			ctx.Directories[op] = ctx.OperationDirectoryID
			ctx.Directories[ctx.NewStage] = ctx.StageID
			paths["fresh-journal"] = filepath.Join(op, "manifest.json")
			paths["ownership"] = filepath.Join(decision, "ownership.json")
			paths["ticket"] = filepath.Join(ctx.NewStage, freshTicketName)
			if phase == "ownership-written" {
				paths["ticket"] = filepath.Join(op, freshTicketTemporary)
			} else {
				paths["new-output"] = filepath.Join(ctx.NewStage, "snapshot")
				if phase == "published-filesystem" || phase == "database-committed" {
					paths["new-output"] = destination
				}
			}
		}
		if ctx.HasCancellation {
			paths["fresh-cancellation"] = filepath.Join(decision, "cancellation.json")
			paths["unknown"] = filepath.Join(op, "unknown-fresh-reset-evidence")
		}
		for name, path := range paths {
			if !freshResetInside(base, path) {
				t.Fatal("artifact escaped reset case", name)
			}
			ctx.Artifacts[name] = freshResetArtifact(t, path)
		}
		if output, ok := ctx.Artifacts["new-output"]; ok {
			ctx.OutputID = output.FileID
		}
		if err = s.cache.Close(); err != nil {
			t.Fatal(err)
		}
		readyPath := filepath.Join(base, "ready.json")
		if _, e := os.Lstat(readyPath); !os.IsNotExist(e) {
			t.Fatal("refuse existing reset checkpoint", e)
		}
		ready, e := os.OpenFile(filepath.Join(base, "ready.tmp"), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
		if e != nil {
			t.Fatal(e)
		}
		ctx.ReadyFileID, e = admissionFileID(ready)
		if e != nil {
			ready.Close()
			t.Fatal(e)
		}
		ctx.ReadyUTC = time.Now().UTC()
		body, e := json.Marshal(ctx)
		if e != nil || len(body) > maxFreshCancellation {
			ready.Close()
			t.Fatal("invalid reset context size", e)
		}
		_, e = ready.Write(body)
		e = errors.Join(e, ready.Sync(), ready.Close())
		if e != nil {
			t.Fatal(e)
		}
		if e = os.Rename(filepath.Join(base, "ready.tmp"), readyPath); e != nil {
			t.Fatal(e)
		}
		// Returning would clean the fixture. Only the external same-VM reset
		// ends this live checkpoint; a controller timeout cannot restart it.
		time.Sleep(20 * time.Minute)
		t.Fatal("owned fresh reset observation deadline")
	}
	if phase == "approval-synced" {
		capture()
		return
	}
	if strings.HasPrefix(phase, "fresh-cancellation-") {
		reached := false
		_, err = executeFreshFixture(s, approvalJSON, func(at string) error {
			if at == "fresh-execution-publication-intent" {
				reached = true
				return errors.New("owned pre-cancellation intent")
			}
			return nil
		})
		if err == nil || !reached {
			t.Fatal("fresh intent checkpoint not reached", err)
		}
		unknown := filepath.Join(s.dir, ".restores", id, "unknown-fresh-reset-evidence")
		f, e := os.OpenFile(unknown, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
		if e != nil {
			t.Fatal(e)
		}
		_, e = f.Write([]byte("retain fresh reset unknown child"))
		e = errors.Join(e, f.Sync(), f.Close())
		if e != nil {
			t.Fatal(e)
		}
		ctx.Cancellation, err = InspectFreshCancellation(s.dir, id, a.OperationID)
		if err != nil {
			t.Fatal(err)
		}
		ctx.HasCancellation = true
		_, err = cancelFreshRestore(s.dir, id, a.OperationID, ctx.Cancellation, "owned fresh reset cancellation", func(at string) error {
			if at == phase {
				capture()
			}
			return nil
		})
	} else {
		_, err = executeFreshFixture(s, approvalJSON, func(at string) error {
			if at == "fresh-execution-"+phase {
				capture()
			}
			return nil
		})
	}
	t.Fatal("fresh reset checkpoint returned", err)
}

func TestFreshGuestResetRecover(t *testing.T) {
	base := guestResetBase(t)
	f, err := os.Open(filepath.Join(base, "ready.json"))
	if err != nil {
		t.Fatal(err)
	}
	readyID, idErr := admissionFileID(f)
	body, err := io.ReadAll(io.LimitReader(f, maxFreshCancellation+1))
	err = errors.Join(err, idErr, f.Close())
	if err != nil || len(body) > maxFreshCancellation {
		t.Fatal("invalid reset context", err)
	}
	hash := sha256.Sum256(body)
	if expected := os.Getenv("OPENEDR_FRESH_RESET_CONTEXT_SHA"); len(expected) != 64 || expected != hex.EncodeToString(hash[:]) {
		t.Fatal("context differs from external pre-reset checkpoint")
	}
	var ctx freshResetContext
	if err = json.Unmarshal(body, &ctx); err != nil {
		t.Fatal(err)
	}
	if ctx.Version != 1 || !ctx.RawSynced || ctx.ReadyFileID != readyID || ctx.ReadyUTC.IsZero() || !freshResetPhase(ctx.Phase) || !ordinaryQuarantineID(ctx.ID) || ctx.ID != ctx.Approval.QuarantineID || ctx.DB != filepath.Join(filepath.Dir(ctx.Root), "edr.db") || ctx.Stored != filepath.Join(ctx.Root, ctx.ID+".quarantined") || len(ctx.Artifacts) < 7 {
		t.Fatal("changed reset context identity")
	}
	canonical, err := json.Marshal(ctx.Approval)
	if err != nil || string(canonical) != string(ctx.ApprovalJSON) {
		t.Fatal("saved reset approval changed", err)
	}
	for _, path := range []string{ctx.Root, ctx.DB, ctx.Destination, ctx.Stored, ctx.Raw, ctx.OldStage, ctx.NewStage} {
		if !freshResetInside(base, path) {
			t.Fatal("context escaped reset case")
		}
	}
	// All external and private observations are checked before SQLite/replay.
	for path, want := range ctx.Directories {
		if !freshResetInside(base, path) || freshResetDirectory(t, path) != want {
			t.Fatal("reset directory identity changed", path)
		}
	}
	for name, want := range ctx.Artifacts {
		if !freshResetInside(base, want.Path) || freshResetArtifact(t, want.Path) != want {
			t.Fatal("reset artifact changed before recovery", name)
		}
	}
	if ctx.Artifacts["raw"].FileID == ctx.Artifacts["stored"].FileID || ctx.OutputID != "" && (ctx.OutputID == ctx.Artifacts["raw"].FileID || ctx.OutputID == ctx.Artifacts["stored"].FileID || ctx.OutputID == ctx.Artifacts["old-stage"].FileID) {
		t.Fatal("reset adopted an older output identity")
	}
	completed := false
	if ctx.HasCancellation {
		first, e := CancelFreshRestore(ctx.Root, ctx.ID, ctx.Approval.OperationID, ctx.Cancellation, "owned fresh reset cancellation")
		if e != nil {
			t.Fatal(e)
		}
		again, e := CancelFreshRestore(ctx.Root, ctx.ID, ctx.Approval.OperationID, ctx.Cancellation, first.Reason)
		if e != nil || again.FileID != first.FileID || again.DecisionID != first.DecisionID {
			t.Fatal("reset cancellation retry changed", e)
		}
		if _, e = ExecuteFreshRestore(ctx.Root, ctx.DB, ctx.ApprovalJSON, "execute reviewed new restore"); e == nil {
			t.Fatal("cancelled reset operation resumed")
		}
	} else if ctx.Phase == "publication-intent" {
		if _, e := ExecuteFreshRestore(ctx.Root, ctx.DB, ctx.ApprovalJSON, "execute reviewed new restore"); e == nil {
			t.Fatal("intent-only reset was adopted")
		}
		if _, e := os.Lstat(ctx.Destination); !os.IsNotExist(e) {
			t.Fatal("intent-only reset published a destination", e)
		}
	} else {
		first, e := ExecuteFreshRestore(ctx.Root, ctx.DB, ctx.ApprovalJSON, "execute reviewed new restore")
		if e != nil || first.Phase != "committed" {
			t.Fatal("fresh reset exact retry failed", e)
		}
		again, e := ExecuteFreshRestore(ctx.Root, ctx.DB, ctx.ApprovalJSON, "execute reviewed new restore")
		if e != nil || again.OutputID != first.OutputID || again.OperationID != first.OperationID {
			t.Fatal("fresh completion retry changed", e)
		}
		if ctx.OutputID != "" && first.OutputID != ctx.OutputID {
			t.Fatal("reset completion adopted a replacement")
		}
		completed = true
		requirePurgePayload(t, ctx.Destination, "payload")
	}
	for name, want := range ctx.Artifacts {
		switch name {
		case "fresh-journal", "ticket":
			continue // mutable or intentionally retired
		case "new-output":
			if completed {
				want.Path = ctx.Destination
			}
		case "unknown":
			if ctx.Phase == "fresh-cancellation-approved" {
				want.Path = filepath.Join(ctx.Root, ".fresh-restores", ctx.ID, ctx.Approval.OperationID, "restore", filepath.Base(want.Path))
			}
		}
		if freshResetArtifact(t, want.Path) != want {
			t.Fatal("recovery changed retained artifact", name)
		}
	}
	if ctx.HasCancellation && freshResetDirectory(t, filepath.Join(ctx.Root, ".fresh-restores", ctx.ID, ctx.Approval.OperationID, "restore")) != ctx.OperationDirectoryID {
		t.Fatal("cancellation archived a replacement")
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
	if err = s.Restore(ctx.ID); !errors.Is(err, ErrRecoveryHeld) {
		t.Fatal("reset released original restore fence", err)
	}
	if err = s.Purge(ctx.ID); !errors.Is(err, ErrRecoveryHeld) {
		t.Fatal("reset released purge fence", err)
	}
	for _, path := range []string{ctx.Raw, ctx.Stored, ctx.OldStage} {
		requirePurgePayload(t, path, "payload")
	}
	if !completed && ctx.Phase != "approval-synced" && ctx.Phase != "ownership-written" {
		if qid, e := s.Quarantine(filepath.Join(ctx.NewStage, "snapshot"), payloadHash("payload"), "Safe.Reset.Fresh.Stage"); qid != "" || !errors.Is(e, ErrAlreadyQuarantined) {
			t.Fatal("retained fresh stage was re-admitted", qid, e)
		}
	}
	rec, found, err := c.GetQuarantine(ctx.ID)
	if err != nil || !found || rec.Purged || rec.Restored != completed {
		t.Fatal("reset cache claimed wrong disposition", err)
	}
	t.Logf("fresh reset verified: phase=%s completed=%t operation=%s output=%s fence=%s", ctx.Phase, completed, ctx.Approval.OperationID, ctx.OutputID, ctx.Artifacts["fence"].FileID)
}
