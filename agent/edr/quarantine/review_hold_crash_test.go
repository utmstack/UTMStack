package quarantine

import (
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/utmstack/UTMStack/agent/edr/cache"
)

func TestReviewHoldCrashWorker(t *testing.T) {
	if os.Getenv("OPENEDR_REVIEW_WORKER") != "1" {
		t.Skip("owned process-crash worker")
	}
	var approval RecoveryApproval
	if err := json.Unmarshal([]byte(os.Getenv("OPENEDR_REVIEW_APPROVAL")), &approval); err != nil {
		t.Fatal(err)
	}
	_, err := holdRecovery(os.Getenv("OPENEDR_REVIEW_ROOT"), os.Getenv("OPENEDR_REVIEW_ID"), approval, "owned process crash review", func(at string) error {
		if at == os.Getenv("OPENEDR_REVIEW_PHASE") {
			os.Exit(97)
		}
		return nil
	})
	t.Fatalf("crash checkpoint not reached: %v", err)
}

func TestReviewHoldActualProcessCrashes(t *testing.T) {
	self, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	for _, phase := range []string{"review-temp-created", "review-prepared", "review-published", "review-synced"} {
		t.Run(phase, func(t *testing.T) {
			s, id, original, stored, approval := reviewHoldFixture(t, "publication-intent")
			before := restoreFixtureManifest(t, s, id)
			if err := s.cache.Close(); err != nil {
				t.Fatal(err)
			}
			encoded, _ := json.Marshal(approval)
			cmd := exec.Command(self, "-test.run=^TestReviewHoldCrashWorker$", "-test.count=1")
			cmd.Env = append(os.Environ(), "OPENEDR_REVIEW_WORKER=1", "OPENEDR_REVIEW_ROOT="+s.dir, "OPENEDR_REVIEW_ID="+id, "OPENEDR_REVIEW_APPROVAL="+string(encoded), "OPENEDR_REVIEW_PHASE="+phase)
			output, err := cmd.CombinedOutput()
			status, ok := err.(*exec.ExitError)
			if !ok || status.ExitCode() != 97 {
				t.Fatalf("worker did not exit at checkpoint: %v %s", err, output)
			}
			c, err := cache.Open(filepath.Join(filepath.Dir(s.dir), "edr.db"))
			if err != nil {
				t.Fatal(err)
			}
			defer c.Close()
			recovered, err := New(s.dir, c)
			if err != nil {
				t.Fatal(err)
			}
			published := phase == "review-published" || phase == "review-synced"
			results := recovered.RecoveryStatus()
			if len(results) == 0 {
				t.Fatal("interrupted decision was lost")
			}
			for _, result := range results {
				want := "needs_attention"
				if published {
					want = "operator_held"
				}
				if result.Status != want {
					t.Fatalf("unproven decision advanced recovery: %+v", results)
				}
			}
			r, found, err := c.GetQuarantine(id)
			if err != nil || !found || r.Purged || r.Restored {
				t.Fatal("interrupted hold claimed completion", r, err)
			}
			requirePurgePayload(t, stored, "payload")
			requirePurgePayload(t, filepath.Join(r.SourceCaptureDir, "source"), "payload")
			requirePurgePayload(t, filepath.Join(filepath.Dir(original), before.stageName(), "snapshot"), "payload")
			if _, err := os.Lstat(original); !os.IsNotExist(err) {
				t.Fatal("interrupted hold published destination", err)
			}
			if err := recovered.Purge(id); err == nil || published && !errors.Is(err, ErrRecoveryHeld) {
				t.Fatal("interrupted decision allowed purge", err)
			}
			// Repetition must preserve the retained objects and fence.
			_, _ = recovered.Recover()
			requirePurgePayload(t, stored, "payload")
			t.Logf("process exit 97 at %s; published=%t; retained without completion", phase, published)
		})
	}
}
