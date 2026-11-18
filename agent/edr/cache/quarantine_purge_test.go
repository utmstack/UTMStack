package cache

import (
	"bytes"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

func purgeCacheFixture(t *testing.T) (*Cache, QuarantineRecord, string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "edr.db")
	c, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { c.Close() })
	r := QuarantineRecord{QuarantineID: "purge-fixture", OriginalPath: "/original", SHA256: "digest", Detection: "Safe.Fixture", Engine: "UTMStack EDR", Restorable: true, SourceCaptureDir: "/capture", SourceCaptureDirID: "directory-id", SourceFileID: "file-id", OriginalSecurity: "security"}
	if err = c.StoreQuarantine(r); err != nil {
		t.Fatal(err)
	}
	r, _, err = c.GetQuarantine(r.QuarantineID)
	if err != nil {
		t.Fatal(err)
	}
	return c, r, path
}

func TestPurgeReceiptAtomicRollbackAndReplay(t *testing.T) {
	c, rec, path := purgeCacheFixture(t)
	if err := c.db.Exec("CREATE TRIGGER reject_purge_receipt BEFORE INSERT ON quarantine_purge_receipts BEGIN SELECT RAISE(ABORT, 'safe fixture'); END").Error; err != nil {
		t.Fatal(err)
	}
	at := time.Date(2026, 1, 2, 3, 4, 5, 123456789, time.UTC)
	if err := c.CommitQuarantinePurge(rec, "operation-one", at); err == nil {
		t.Fatal("receipt failure was ignored")
	}
	current, _, err := c.GetQuarantine(rec.QuarantineID)
	if err != nil || current.Purged || !current.PurgedAt.IsZero() {
		t.Fatalf("failed receipt did not roll back the audit row: %+v %v", current, err)
	}
	var count int64
	if err = c.db.Model(&QuarantinePurgeReceipt{}).Count(&count).Error; err != nil || count != 0 {
		t.Fatalf("failed transaction left a receipt: %d %v", count, err)
	}
	if err = c.db.Exec("DROP TRIGGER reject_purge_receipt").Error; err != nil {
		t.Fatal(err)
	}
	if err = c.CommitQuarantinePurge(rec, "operation-one", at); err != nil {
		t.Fatal(err)
	}
	if err = c.CommitQuarantinePurge(rec, "operation-one", at.Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	current, _, err = c.GetQuarantine(rec.QuarantineID)
	if err != nil || !current.Purged || !current.PurgedAt.Equal(at) {
		t.Fatalf("replay replaced completion time: %+v %v", current, err)
	}
	c.Close()
	reopened, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	if committed, err := reopened.ValidateQuarantinePurge(rec, "operation-one"); err != nil || !committed {
		t.Fatalf("durable receipt did not survive reopen: %t %v", committed, err)
	}
	if err := reopened.CommitQuarantinePurge(rec, "different-operation", at); err == nil {
		t.Fatal("different operation took over the completed purge")
	}
}

func TestPurgeReceiptRefusesChangedApproval(t *testing.T) {
	mutations := map[string]func(*QuarantineRecord){
		"path":       func(r *QuarantineRecord) { r.OriginalPath += "-other" },
		"hash":       func(r *QuarantineRecord) { r.SHA256 += "-other" },
		"detection":  func(r *QuarantineRecord) { r.Detection += "-other" },
		"engine":     func(r *QuarantineRecord) { r.Engine += "-other" },
		"date":       func(r *QuarantineRecord) { r.QuarantinedAt = r.QuarantinedAt.Add(time.Second) },
		"restorable": func(r *QuarantineRecord) { r.Restorable = !r.Restorable },
		"restored":   func(r *QuarantineRecord) { r.Restored = !r.Restored },
		"capture":    func(r *QuarantineRecord) { r.SourceCaptureDir += "-other" },
		"capture-id": func(r *QuarantineRecord) { r.SourceCaptureDirID += "-other" },
		"file-id":    func(r *QuarantineRecord) { r.SourceFileID += "-other" },
		"security":   func(r *QuarantineRecord) { r.OriginalSecurity += "-other" },
	}
	for name, mutate := range mutations {
		t.Run(name, func(t *testing.T) {
			c, rec, _ := purgeCacheFixture(t)
			changed := rec
			mutate(&changed)
			if err := c.CommitQuarantinePurge(changed, "operation", time.Now()); err == nil {
				t.Fatal("changed approved metadata was accepted")
			}
			current, _, _ := c.GetQuarantine(rec.QuarantineID)
			if current.Purged {
				t.Fatal("conflict marked the record purged")
			}
		})
	}
}

func TestPurgeReceiptBindsNonceAndSurvivesMigration(t *testing.T) {
	c, rec, path := purgeCacheFixture(t)
	if err := c.db.Migrator().DropTable(&QuarantinePurgeReceipt{}); err != nil {
		t.Fatal(err)
	}
	c.Close()
	migrated, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer migrated.Close()
	if err = migrated.CommitQuarantinePurge(rec, "same-nonce", time.Now()); err != nil {
		t.Fatal(err)
	}
	other := rec
	other.QuarantineID = "other-record"
	if err = migrated.StoreQuarantine(other); err != nil {
		t.Fatal(err)
	}
	if err = migrated.CommitQuarantinePurge(other, "same-nonce", time.Now()); err == nil {
		t.Fatal("receipt nonce was reused for another quarantine")
	}
	if current, _, _ := migrated.GetQuarantine(other.QuarantineID); current.Purged {
		t.Fatal("nonce conflict changed the other quarantine")
	}
}

func TestPurgeApprovalEncodingIsLossless(t *testing.T) {
	for _, raw := range [][]byte{[]byte("ordinary"), {0, 0xff, 0x80, '\\', '/', '\n'}, bytes.Repeat([]byte{0xff}, 4096)} {
		for _, at := range []time.Time{time.Time{}, time.Unix(-12345678, 987654321), time.Date(12000, 2, 3, 4, 5, 6, 7, time.UTC)} {
			r := QuarantineRecord{QuarantineID: "fixture", OriginalPath: string(raw), SHA256: string(raw), Detection: string(raw), Engine: string(raw), QuarantinedAt: at, Restorable: true, Restored: true, SourceCaptureDir: string(raw), SourceCaptureDirID: string(raw), SourceFileID: string(raw), OriginalSecurity: string(raw)}
			body, err := EncodeQuarantinePurgeRecord(r)
			if err != nil {
				t.Fatal(err)
			}
			decoded, err := DecodeQuarantinePurgeRecord(body)
			r.QuarantinedAt = r.QuarantinedAt.UTC()
			if err != nil || !reflect.DeepEqual(decoded, r) {
				t.Fatalf("purge approval lost bytes/time: %v %+v %+v", err, decoded, r)
			}
			other := r
			other.OriginalPath += string([]byte{0xfe})
			a, _ := purgeRecordDigest(r)
			b, _ := purgeRecordDigest(other)
			if a == b {
				t.Fatal("different invalid UTF-8 paths share an approval identity")
			}
		}
	}
}

func FuzzPurgeApproval(f *testing.F) {
	f.Add([]byte{0, 0xff, '/', 0x80}, int64(-1234567), uint32(12345))
	f.Fuzz(func(t *testing.T, data []byte, sec int64, nanos uint32) {
		if len(data) > 32768 {
			t.Skip()
		}
		r := QuarantineRecord{QuarantineID: "fixture", OriginalPath: string(data), OriginalSecurity: string(data), QuarantinedAt: time.Unix(sec, int64(nanos%1e9)).UTC()}
		body, err := EncodeQuarantinePurgeRecord(r)
		if err != nil {
			t.Fatal(err)
		}
		decoded, err := DecodeQuarantinePurgeRecord(body)
		if err != nil || !reflect.DeepEqual(decoded, r) {
			t.Fatalf("purge approval round trip: %v", err)
		}
		_, _ = DecodeQuarantinePurgeRecord(data) // Malformed bounded encodings must not panic.
	})
}
