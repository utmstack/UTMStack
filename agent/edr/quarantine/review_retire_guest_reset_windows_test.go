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
	"testing"
	"time"

	"github.com/utmstack/UTMStack/agent/edr/cache"
)

type retirementResetContext struct {
	Version                        int
	Phase, Root, DB, Raw, NewStage string
	Inspection                     CancelledRetirementInspection
	Approval                       FreshRestoreApproval
	Artifacts                      map[string]cancelResetArtifact
	Directories                    map[string]string
	ReadyUTC                       time.Time
	ReadyFileID                    string
	FixtureSynced                  bool
}

func retirementResetPhase(phase string) bool {
	switch phase {
	case "receipt-created", "marker-created", "ticket-created", "receipt-written", "approved", "ticket-prepared", "marker-prepared", "marker-published", "before-hold-move", "hold-archived", "synced":
		return true
	}
	return false
}

func TestRetirementGuestResetPrepare(t *testing.T) {
	base := guestResetBase(t)
	phase := os.Getenv("OPENEDR_GUEST_RESET_PHASE")
	if !retirementResetPhase(phase) {
		t.Fatal("unsupported retirement reset boundary")
	}
	s, id, hold, stage, a := retirementInspectionFixture(t, "cancelled")
	r, err := InspectCancelledRetirement(s.dir, id, hold, stage)
	if err != nil {
		t.Fatal(err)
	}
	old, err := decodeRestoreManifest(r.Original.Manifest, id)
	if err != nil {
		t.Fatal(err)
	}
	ctx := retirementResetContext{Version: 1, Phase: phase, Root: s.dir, DB: filepath.Join(filepath.Dir(s.dir), "edr.db"), Raw: filepath.Join(old.Record.SourceCaptureDir, "source"), NewStage: filepath.Join(string(a.Inspection.Destination.ParentPath), ".edr-restore-"+a.OperationID), Inspection: r, Approval: a, Artifacts: map[string]cancelResetArtifact{}, Directories: map[string]string{}}
	if !freshResetInside(base, ctx.Root) {
		t.Fatal("retirement fixture escaped owned reset case")
	}
	if err = s.cache.Close(); err != nil {
		t.Fatal(err)
	}
	// Initial inert data and unknown fixture children must already be durable.
	// The subsequent checkpoint exercises only the retirement transition.
	err = filepath.Walk(filepath.Dir(s.dir), func(path string, info os.FileInfo, e error) error {
		if e != nil || !info.Mode().IsRegular() {
			return e
		}
		f, e := os.Open(path)
		if e != nil {
			return e
		}
		return errors.Join(admissionSyncSource(f), f.Close())
	})
	if err != nil {
		t.Fatal("initial retirement fixture flush failed", err)
	}
	ctx.FixtureSynced = true
	_, err = retireCancelledRecovery(s.dir, r, "retain every object across retirement reset", "", func(at string) error {
		if at != "retirement-"+phase {
			return nil
		}
		history := filepath.Join(s.dir, ".reviews", hold)
		fresh := filepath.Join(s.dir, ".fresh-restores", id, a.OperationID)
		for _, path := range []string{s.dir, history, filepath.Join(history, "restore"), fresh, filepath.Join(fresh, "restore"), stage, ctx.NewStage} {
			if !freshResetInside(base, path) {
				t.Fatal("retirement directory escaped reset case")
			}
			ctx.Directories[path] = freshResetDirectory(t, path)
		}
		holdPath := filepath.Join(s.dir, ".purges", id)
		if phase == "hold-archived" || phase == "synced" {
			holdPath = filepath.Join(history, "hold.json")
		}
		paths := map[string]string{
			"raw": ctx.Raw, "stored": filepath.Join(s.dir, id+".quarantined"),
			"old-stage": filepath.Join(stage, "snapshot"), "new-stage": filepath.Join(ctx.NewStage, "snapshot"),
			"hold": holdPath, "old-cancellation": filepath.Join(history, "manifest.json"), "old-journal": filepath.Join(history, "restore", "manifest.json"),
			"approval": filepath.Join(fresh, "manifest.json"), "ownership": filepath.Join(fresh, "ownership.json"),
			"fresh-ticket": filepath.Join(ctx.NewStage, freshTicketName), "fresh-cancellation": filepath.Join(fresh, "cancellation.json"),
			"fresh-journal": filepath.Join(fresh, "restore", "manifest.json"), "unknown": filepath.Join(fresh, "restore", "unknown-retirement-evidence"),
			"decision": filepath.Join(history, retirementReceiptName),
		}
		for name, path := range map[string]string{"prepared": filepath.Join(history, retirementPreparedName), "guard": filepath.Join(s.dir, ".restores", id), "stage-ticket": filepath.Join(stage, retirementTicketName)} {
			if _, e := os.Lstat(path); e == nil {
				paths[name] = path
			} else if !os.IsNotExist(e) {
				t.Fatal(e)
			}
		}
		for name, path := range paths {
			if !freshResetInside(base, path) {
				t.Fatal("retirement artifact escaped reset case", name)
			}
			ctx.Artifacts[name] = freshResetArtifact(t, path)
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
			t.Fatal("invalid retirement reset context", e)
		}
		_, e = ready.Write(body)
		e = errors.Join(e, ready.Sync(), ready.Close())
		if e != nil {
			t.Fatal(e)
		}
		if e = os.Rename(filepath.Join(base, "ready.tmp"), filepath.Join(base, "ready.json")); e != nil {
			t.Fatal(e)
		}
		// Only the external same-VM reset ends this live checkpoint.
		time.Sleep(20 * time.Minute)
		t.Fatal("owned retirement reset observation deadline")
		return nil
	})
	t.Fatal("retirement reset checkpoint returned", err)
}

func TestRetirementGuestResetRecover(t *testing.T) {
	base := guestResetBase(t)
	f, err := os.Open(filepath.Join(base, "ready.json"))
	if err != nil {
		t.Fatal(err)
	}
	id, idErr := admissionFileID(f)
	body, err := io.ReadAll(io.LimitReader(f, maxFreshCancellation+1))
	err = errors.Join(err, idErr, f.Close())
	hash := sha256.Sum256(body)
	if err != nil || len(body) > maxFreshCancellation || os.Getenv("OPENEDR_FRESH_RESET_CONTEXT_SHA") != hex.EncodeToString(hash[:]) {
		t.Fatal("retirement context differs from external pre-reset checkpoint", err)
	}
	var ctx retirementResetContext
	if err = json.Unmarshal(body, &ctx); err != nil || ctx.Version != 1 || !ctx.FixtureSynced || ctx.ReadyFileID != id || ctx.ReadyUTC.IsZero() || !retirementResetPhase(ctx.Phase) || ctx.DB != filepath.Join(filepath.Dir(ctx.Root), "edr.db") || len(ctx.Artifacts) < 14 {
		t.Fatal("invalid retirement reset context", err)
	}
	for _, path := range []string{ctx.Root, ctx.DB, ctx.Raw, ctx.NewStage, string(ctx.Inspection.OldStage.Path)} {
		if !freshResetInside(base, path) {
			t.Fatal("retirement context escaped reset case")
		}
	}
	for path, want := range ctx.Directories {
		if !freshResetInside(base, path) || freshResetDirectory(t, path) != want {
			t.Fatal("retirement reset directory identity changed", path)
		}
	}
	for name, want := range ctx.Artifacts {
		if !freshResetInside(base, want.Path) || freshResetArtifact(t, want.Path) != want {
			t.Fatal("retirement reset artifact changed before retry", name)
		}
	}
	r := ctx.Inspection
	partial := ctx.Phase == "receipt-created" || ctx.Phase == "marker-created" || ctx.Phase == "ticket-created"
	m, err := RetireCancelledRecovery(ctx.Root, r, "retain every object across retirement reset", "")
	if (err != nil) != partial {
		t.Fatal("retirement reset adopted partial decision or refused exact complete retry", err)
	}
	if !partial {
		again, e := RetireCancelledRecovery(ctx.Root, r, m.Reason, "")
		if e != nil || again.FileID != m.FileID || again.MarkerID != m.MarkerID || again.StageTicketID != m.StageTicketID || again.DecisionID != m.DecisionID {
			t.Fatal("retirement reset retry changed its decision", e)
		}
	}
	for name, want := range ctx.Artifacts {
		if !partial {
			switch name {
			case "stage-ticket", "prepared":
				continue // Only these own prepared controls may gain canonical bytes.
			case "hold":
				want.Path = filepath.Join(ctx.Root, ".reviews", r.HoldID, "hold.json")
			}
		}
		if freshResetArtifact(t, want.Path) != want {
			t.Fatal("retirement retry changed retained evidence", name)
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
	wantErr := ErrRecoveryRetired
	if partial {
		wantErr = ErrRecoveryHeld
	}
	if err = s.Restore(r.QuarantineID); !errors.Is(err, wantErr) {
		t.Fatal("retirement reset lost restore protection", err)
	}
	if err = s.Purge(r.QuarantineID); !errors.Is(err, wantErr) {
		t.Fatal("retirement reset permitted deletion", err)
	}
	if !partial {
		empty, e := cache.Open(filepath.Join(base, "empty-after-reset.db"))
		if e != nil {
			t.Fatal(e)
		}
		defer empty.Close()
		other, e := New(filepath.Join(base, "other-after-reset"), empty)
		if e != nil {
			t.Fatal(e)
		}
		for _, path := range []string{filepath.Join(string(r.OldStage.Path), "snapshot"), filepath.Join(ctx.NewStage, "snapshot")} {
			if q, e := other.Quarantine(path, payloadHash("payload"), "Safe.Retirement.Reset"); q != "" || !errors.Is(e, ErrAlreadyQuarantined) {
				t.Fatal("retirement reset stage lost independent ownership", q, e)
			}
		}
	}
	destination := filepath.Join(string(ctx.Approval.Inspection.Destination.ParentPath), string(ctx.Approval.Inspection.Destination.Leaf))
	if _, err = os.Lstat(destination); !os.IsNotExist(err) {
		t.Fatal("retirement reset published a destination", err)
	}
	for _, path := range []string{ctx.Raw, filepath.Join(ctx.Root, r.QuarantineID+".quarantined"), filepath.Join(string(r.OldStage.Path), "snapshot"), filepath.Join(ctx.NewStage, "snapshot")} {
		requirePurgePayload(t, path, "payload")
	}
	t.Logf("retirement checkpoint retained all original/fresh data: phase=%s partial=%t decision=%s", ctx.Phase, partial, m.DecisionID)
}
