package quarantine

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/utmstack/UTMStack/agent/edr/cache"
)

func TestFreshExecutionEarlierRuntimeRetainsActiveStageWithDurableHint(t *testing.T) {
	program := os.Getenv("OPENEDR_CANCEL_EARLIER_PROGRAM")
	if program == "" {
		t.Skip("requires the pinned earlier production cancellation API worker")
	}
	s, id, destination, stored, _, body := freshExecutionFixture(t)
	_, err := executeFreshFixture(s, body, func(at string) error {
		if at == "fresh-execution-verified" {
			return errors.New("owned fresh stage checkpoint")
		}
		return nil
	})
	if err == nil {
		t.Fatal("fresh stage checkpoint not reached")
	}
	m := restoreFixtureManifest(t, s, id)
	if _, found, e := s.cache.QuarantineRestoreStageByIdentity(m.StageID); e != nil || !found {
		t.Fatal("fresh output lacks durable earlier-reader locator", e)
	}
	if err = s.cache.Close(); err != nil {
		t.Fatal(err)
	}
	stage := filepath.Join(filepath.Dir(destination), m.stageName(), "snapshot")
	before := inspectionTree(t, s.dir)
	cmd := exec.Command(program, "-test.run=^TestEarlierCancelledOperationWorker$", "-test.count=1", "-test.v")
	cmd.Env = append(os.Environ(), "OPENEDR_EARLIER_CANCEL_WORKER=1", "OPENEDR_EARLIER_CANCEL_ROOT="+s.dir, "OPENEDR_EARLIER_CANCEL_ID="+id, "OPENEDR_EARLIER_CANCEL_STAGE="+stage)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("earlier API changed or admitted a fresh operation: %v %s", err, output)
	}
	if !reflect.DeepEqual(before, inspectionTree(t, s.dir)) {
		t.Fatal("earlier API changed fresh control evidence")
	}
	requirePurgePayload(t, stored, "payload")
	requirePurgePayload(t, stage, "payload")
	c, err := cache.Open(filepath.Join(filepath.Dir(s.dir), "edr.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { c.Close() })
	s.cache = c
	if _, err = executeFreshFixture(s, body, nil); err != nil {
		t.Fatal("earlier read prevented exact fresh retry", err)
	}
	requirePurgePayload(t, destination, "payload")
}
