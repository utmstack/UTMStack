//go:build windows && edr_acceptance

package quarantine

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/utmstack/UTMStack/agent/edr/cache"
)

type guestResetContext struct {
	Phase          string
	ID             string
	Root           string
	DB             string
	Original       string
	Stored         string
	Raw            string
	RawID          string
	StageID        string
	OperationID    string
	ManifestSHA256 string
}

func guestResetBase(t *testing.T) string {
	t.Helper()
	root := os.Getenv("OPENEDR_GUEST_RESET_ROOT")
	if os.Getenv("EDR_WINDOWS_ACCEPTANCE") != "isolated-vm" || root == "" {
		t.Skip("requires the owned disposable guest-reset controller")
	}
	if !filepath.IsAbs(root) || filepath.Clean(root) != root {
		t.Fatal("invalid owned case root")
	}
	return root
}
func writeGuestResetContext(t *testing.T, path string, value guestResetContext) {
	t.Helper()
	body, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = f.Write(body); err == nil {
		err = f.Sync()
	}
	err = errors.Join(err, f.Close())
	if err != nil {
		t.Fatal(err)
	}
	parent, err := openRestoreDirectory(filepath.Dir(path), true)
	if err != nil {
		t.Fatal(err)
	}
	defer parent.Close()
	if err = admissionSyncDirectory(parent); err != nil {
		t.Fatal(err)
	}
}
func TestRestoreGuestResetPrepare(t *testing.T) {
	base := guestResetBase(t)
	phase := os.Getenv("OPENEDR_GUEST_RESET_PHASE")
	if phase != "stage-indexed" && phase != "stage-index-deleted" {
		t.Fatal("unsupported owned reset phase")
	}
	root, err := openAdmissionRoot(base, false)
	if err != nil {
		t.Fatal(err)
	}
	root.Close()
	db := filepath.Join(base, "edr.db")
	if _, err = os.Lstat(db); !os.IsNotExist(err) {
		t.Fatal("refuse existing case database", err)
	}
	c, err := cache.Open(db)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	s, err := New(filepath.Join(base, "quarantine"), c)
	if err != nil {
		t.Fatal(err)
	}
	original := filepath.Join(base, "original.txt")
	if err = os.WriteFile(original, []byte("owned guest reset payload"), 0600); err != nil {
		t.Fatal(err)
	}
	id, err := s.Quarantine(original, payloadHash("owned guest reset payload"), "Safe.Guest.Reset.Control")
	if err != nil {
		t.Fatal(err)
	}
	rec, found, err := c.GetQuarantine(id)
	if err != nil || !found {
		t.Fatal("missing owned record", err)
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
	reached := false
	s.responseFault = func(at string) error {
		if at != "restore-"+phase {
			return nil
		}
		reached = true
		m := restoreFixtureManifest(t, s, id)
		body, err := os.ReadFile(filepath.Join(s.dir, ".restores", id, "manifest.json"))
		if err != nil {
			t.Fatal(err)
		}
		digest := sha256.Sum256(body)
		context := guestResetContext{Phase: phase, ID: id, Root: s.dir, DB: db, Original: original, Stored: filepath.Join(s.dir, id+".quarantined"), Raw: raw, RawID: rawID, StageID: m.StageID, OperationID: m.OperationID, ManifestSHA256: hex.EncodeToString(digest[:])}
		writeGuestResetContext(t, filepath.Join(base, "ready.json"), context)
		// The controller forcibly resets this disposable guest while the transaction
		// is paused here. No shipped binary or environment-triggered crash hook exists.
		time.Sleep(20 * time.Minute)
		return errors.New("controller did not reset the guest within the test deadline")
	}
	err = s.Restore(id)
	t.Fatalf("reset checkpoint returned: reached=%t err=%v", reached, err)
}
func TestRestoreGuestResetRecover(t *testing.T) {
	base := guestResetBase(t)
	body, err := os.ReadFile(filepath.Join(base, "ready.json"))
	if err != nil {
		t.Fatal(err)
	}
	var ctx guestResetContext
	if err = json.Unmarshal(body, &ctx); err != nil {
		t.Fatal(err)
	}
	if ctx.DB != filepath.Join(base, "edr.db") || ctx.Root != filepath.Join(base, "quarantine") || ctx.Original != filepath.Join(base, "original.txt") || ctx.Stored != filepath.Join(ctx.Root, ctx.ID+".quarantined") || !ordinaryQuarantineID(ctx.ID) {
		t.Fatal("changed owned reset context")
	}
	published := ctx.Phase == "stage-index-deleted"
	if !published && ctx.Phase != "stage-indexed" {
		t.Fatal("invalid checkpoint phase")
	}
	journal, err := os.ReadFile(filepath.Join(ctx.Root, ".restores", ctx.ID, "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	m, err := decodeRestoreManifest(journal, ctx.ID)
	if err != nil || m.OperationID != ctx.OperationID || m.StageID != ctx.StageID || ctx.Raw != filepath.Join(m.Record.SourceCaptureDir, "source") {
		t.Fatal("owned reset journal identity differs", err)
	}
	digest := sha256.Sum256(journal)
	if hex.EncodeToString(digest[:]) != ctx.ManifestSHA256 {
		t.Fatal("journal changed during guest reset")
	}
	c, err := cache.Open(ctx.DB)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	_, indexed, err := c.QuarantineRestoreStageByIdentity(ctx.StageID)
	if err != nil || indexed == published {
		t.Fatalf("index durability differed: indexed=%t published=%t err=%v", indexed, published, err)
	}
	s, err := New(ctx.Root, c)
	if err != nil {
		t.Fatal(err)
	}
	results := s.RecoveryStatus()
	if len(results) != 1 || results[0].Status == "needs_attention" {
		t.Fatalf("guest recovery incomplete: %+v", results)
	}
	rec, found, err := c.GetQuarantine(ctx.ID)
	if err != nil || !found || rec.Restored != published {
		t.Fatal("incorrect historical restore state", rec, err)
	}
	if _, indexed, err = c.QuarantineRestoreStageByIdentity(ctx.StageID); err != nil || indexed {
		t.Fatal("stage lookup survived completed cleanup", err)
	}
	if published {
		requirePurgePayload(t, ctx.Original, "owned guest reset payload")
		if _, err = os.Lstat(ctx.Stored); !os.IsNotExist(err) {
			t.Fatal("committed backup not retired", err)
		}
	} else {
		requirePurgePayload(t, ctx.Stored, "owned guest reset payload")
		if _, err = os.Lstat(ctx.Original); !os.IsNotExist(err) {
			t.Fatal("unpublished reset produced original", err)
		}
	}
	requirePurgePayload(t, ctx.Raw, "owned guest reset payload")
	f, err := os.Open(ctx.Raw)
	if err != nil {
		t.Fatal(err)
	}
	rawID, err := admissionFileID(f)
	f.Close()
	if err != nil || rawID != ctx.RawID {
		t.Fatal("raw capture identity changed", err)
	}
	if _, err = os.Lstat(filepath.Join(base, ".edr-restore-"+ctx.OperationID)); !os.IsNotExist(err) {
		t.Fatal("stage remains or cannot be inspected", err)
	}
	if again, err := s.Recover(); err != nil || len(again) != 0 {
		t.Fatal("repeat recovery not empty", again, err)
	}
	t.Logf("forced_guest_reset phase=%s restored=%t stage_index_removed=true raw_identity_preserved=true", ctx.Phase, published)
}
