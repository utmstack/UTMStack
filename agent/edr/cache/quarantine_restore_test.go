package cache

import (
	"testing"
	"time"
)

func TestRestoreReceiptAtomicRollbackReplayAndMigration(t *testing.T) {
	c, rec, path := purgeCacheFixture(t)
	if err := c.db.Migrator().DropTable(&QuarantineRestoreReceipt{}); err != nil {
		t.Fatal(err)
	}
	c.Close()
	c, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	if err = c.db.Exec("CREATE TRIGGER reject_restore_receipt BEFORE INSERT ON quarantine_restore_receipts BEGIN SELECT RAISE(ABORT, 'safe fixture'); END").Error; err != nil {
		t.Fatal(err)
	}
	at := time.Date(2026, 1, 2, 3, 4, 5, 123456789, time.UTC)
	if err = c.CommitQuarantineRestore(rec, "operation", "parent", "output", at); err == nil {
		t.Fatal("receipt failure was ignored")
	}
	current, _, err := c.GetQuarantine(rec.QuarantineID)
	if err != nil || current.Restored {
		t.Fatal("failed receipt did not roll back audit row", err)
	}
	var n int64
	if err = c.db.Model(&QuarantineRestoreReceipt{}).Count(&n).Error; err != nil || n != 0 {
		t.Fatal("failed transaction left receipt", err)
	}
	if err = c.db.Exec("DROP TRIGGER reject_restore_receipt").Error; err != nil {
		t.Fatal(err)
	}
	if err = c.CommitQuarantineRestore(rec, "operation", "parent", "output", at); err != nil {
		t.Fatal(err)
	}
	if err = c.CommitQuarantineRestore(rec, "operation", "parent", "output", at.Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	var receipt QuarantineRestoreReceipt
	if err = c.db.First(&receipt, "operation_id = ?", "operation").Error; err != nil || !receipt.CompletedAt.Equal(at) {
		t.Fatal("replay changed completion time", err)
	}
	c.Close()
	reopened, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	if ok, err := reopened.ValidateQuarantineRestore(rec, "operation", "parent", "output"); err != nil || !ok {
		t.Fatal("receipt did not survive reopen", err)
	}
	for _, ids := range [][3]string{{"other", "parent", "output"}, {"operation", "other", "output"}, {"operation", "parent", "other"}} {
		if err = reopened.CommitQuarantineRestore(rec, ids[0], ids[1], ids[2], at); err == nil {
			t.Fatal("different approval adopted restore", ids)
		}
	}
}

func TestRestoreReceiptRejectsMetadataAndNonceConflicts(t *testing.T) {
	mutations := map[string]func(*QuarantineRecord){
		"path":       func(r *QuarantineRecord) { r.OriginalPath += "-other" },
		"hash":       func(r *QuarantineRecord) { r.SHA256 += "-other" },
		"detection":  func(r *QuarantineRecord) { r.Detection += "-other" },
		"engine":     func(r *QuarantineRecord) { r.Engine += "-other" },
		"date":       func(r *QuarantineRecord) { r.QuarantinedAt = r.QuarantinedAt.Add(time.Second) },
		"restorable": func(r *QuarantineRecord) { r.Restorable = false },
		"restored":   func(r *QuarantineRecord) { r.Restored = true },
		"purged":     func(r *QuarantineRecord) { r.Purged = true },
		"capture":    func(r *QuarantineRecord) { r.SourceCaptureDir += "-other" },
		"capture-id": func(r *QuarantineRecord) { r.SourceCaptureDirID += "-other" },
		"source-id":  func(r *QuarantineRecord) { r.SourceFileID += "-other" },
		"security":   func(r *QuarantineRecord) { r.OriginalSecurity += "-other" },
	}
	for name, mutate := range mutations {
		t.Run(name, func(t *testing.T) {
			c, rec, _ := purgeCacheFixture(t)
			changed := rec
			mutate(&changed)
			if err := c.CommitQuarantineRestore(changed, "nonce", "parent", "output", time.Now()); err == nil {
				t.Fatal("changed metadata committed")
			}
			if r, _, err := c.GetQuarantine(rec.QuarantineID); err != nil || r.Restored {
				t.Fatal("conflict changed audit", err)
			}
		})
	}
	c, rec, _ := purgeCacheFixture(t)
	if err := c.CommitQuarantineRestore(rec, "nonce", "parent", "output", time.Now()); err != nil {
		t.Fatal(err)
	}
	other := rec
	other.QuarantineID = "other"
	if err := c.StoreQuarantine(other); err != nil {
		t.Fatal(err)
	}
	if err := c.CommitQuarantineRestore(other, "nonce", "parent", "output", time.Now()); err == nil {
		t.Fatal("nonce reused for another item")
	}
	if r, _, err := c.GetQuarantine(other.QuarantineID); err != nil || r.Restored {
		t.Fatal("nonce conflict changed other item", err)
	}
}
