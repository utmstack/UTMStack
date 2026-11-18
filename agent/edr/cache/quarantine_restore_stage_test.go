package cache

import (
	"bytes"
	"testing"
)

func TestRestoreStageLookupMigrationConflictAndExactRemoval(t *testing.T) {
	c, _, path := purgeCacheFixture(t)
	if err := c.db.Migrator().DropTable(&QuarantineRestoreStage{}); err != nil {
		t.Fatal(err)
	}
	c.Close()
	var err error
	c, err = Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	record := QuarantineRestoreStage{StageID: "volume:directory", OperationID: "operation", QuarantineID: "record", RootID: "root", RootPath: []byte{'/', 't', 0xff}}
	if err = c.TrackQuarantineRestoreStage(record); err != nil {
		t.Fatal(err)
	}
	if err = c.TrackQuarantineRestoreStage(record); err != nil {
		t.Fatal("idempotent replay", err)
	}
	got, found, err := c.QuarantineRestoreStageByIdentity(record.StageID)
	if err != nil || !found || !sameRestoreStage(got, record) || !bytes.Equal(got.RootPath, record.RootPath) {
		t.Fatalf("lossless lookup: %+v %t %v", got, found, err)
	}
	changed := record
	changed.OperationID = "other"
	if err = c.TrackQuarantineRestoreStage(changed); err == nil {
		t.Fatal("another operation overwrote stage ownership")
	}
	if err = c.ForgetQuarantineRestoreStage(changed); err != nil {
		t.Fatal(err)
	}
	if _, found, err = c.QuarantineRestoreStageByIdentity(record.StageID); err != nil || !found {
		t.Fatal("mismatched removal deleted another operation", err)
	}
	if err = c.ForgetQuarantineRestoreStage(record); err != nil {
		t.Fatal(err)
	}
	if _, found, err = c.QuarantineRestoreStageByIdentity(record.StageID); err != nil || found {
		t.Fatal("exact removal did not retire lookup", err)
	}
}
