package quarantine

import (
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"
)

func TestReviewCancelEarlierRuntimeCompatibility(t *testing.T) {
	program := os.Getenv("OPENEDR_CANCEL_EARLIER_PROGRAM")
	if program == "" {
		t.Skip("requires a pinned earlier production API with the owned cancellation worker")
	}
	for _, phase := range []string{"verified", "publication-intent"} {
		t.Run(phase, func(t *testing.T) {
			s, id, original, stored, a := reviewReleaseFixture(t, phase)
			m := restoreFixtureManifest(t, s, id)
			if _, err := CancelRecovery(s.dir, id, a, "retain cancelled operation across older API"); err != nil {
				t.Fatal(err)
			}
			stage := filepath.Join(filepath.Dir(original), m.stageName(), "snapshot")
			if err := s.cache.Close(); err != nil {
				t.Fatal(err)
			}
			before := inspectionTree(t, s.dir)
			cmd := exec.Command(program, "-test.run=^TestEarlierCancelledOperationWorker$", "-test.count=1", "-test.v")
			cmd.Env = append(os.Environ(), "OPENEDR_EARLIER_CANCEL_WORKER=1", "OPENEDR_EARLIER_CANCEL_ROOT="+s.dir, "OPENEDR_EARLIER_CANCEL_ID="+id, "OPENEDR_EARLIER_CANCEL_STAGE="+stage)
			output, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("earlier cancellation API worker: %v %s", err, output)
			}
			if !reflect.DeepEqual(before, inspectionTree(t, s.dir)) {
				t.Fatal("earlier runtime changed cancelled operation evidence")
			}
			root, err := s.openRoot()
			if err != nil {
				t.Fatal(err)
			}
			history, originalOp, err := openArchivedRestoreProof(root, id)
			root.Close()
			if err != nil {
				t.Fatal("earlier API changed exact fence, receipt or original identities", err)
			}
			history.Close()
			originalOp.Close()
			requirePurgePayload(t, stored, "payload")
			requirePurgePayload(t, stage, "payload")
			if _, err := os.Lstat(original); !os.IsNotExist(err) {
				t.Fatal("earlier runtime published original destination", err)
			}
			t.Logf("earlier production API retained cancellation: %s", output)
		})
	}
}
