package quarantine

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/google/uuid"
)

func TestFreshJournalPreservesVersionOneWireBytes(t *testing.T) {
	s, id, _, _, _ := reviewHoldFixture(t, "verified")
	report, err := InspectRecovery(s.dir, id)
	if err != nil {
		t.Fatal(err)
	}
	body := report.Journals[1].Manifest
	m, err := decodeRestoreManifest(body, id)
	if err != nil {
		t.Fatal(err)
	}
	roundTrip, err := json.Marshal(m)
	if err != nil || !bytes.Equal(roundTrip, body) || bytes.Contains(roundTrip, []byte(`"FreshID"`)) {
		t.Fatal("fresh journal support changed existing v1 wire bytes", err)
	}
	m.Version, m.FreshID = 2, m.OperationID
	if err = validateRestoreManifest(m, id); err != nil {
		t.Fatal("valid fresh syntax rejected", err)
	}
	for _, changed := range []string{"v1-with-fresh-id", "v2-without-fresh-id", "wrong-fresh-id", "retired-source"} {
		bad := m
		switch changed {
		case "v1-with-fresh-id":
			bad.Version = 1
		case "v2-without-fresh-id":
			bad.FreshID = ""
		case "wrong-fresh-id":
			bad.FreshID = uuid.NewString()
		case "retired-source":
			bad.Phase, bad.Source.State = "committed", "claimed"
		}
		if err = validateRestoreManifest(bad, id); err == nil {
			t.Fatal("invalid fresh journal syntax accepted", changed)
		}
	}
}

func TestFreshJournalCannotAuthorizeOrdinaryRecoveryOrBackupRetirement(t *testing.T) {
	for _, phase := range []string{"copying", "verified", "publishing", "published", "committed"} {
		t.Run(phase, func(t *testing.T) {
			s, id, original, stored, _ := reviewHoldFixture(t, "verified")
			m := restoreFixtureManifest(t, s, id)
			m.Version, m.FreshID, m.Phase = 2, m.OperationID, phase
			root, err := s.openRoot()
			if err != nil {
				t.Fatal(err)
			}
			journals, op, err := openRestoreOperation(root, id)
			if err != nil {
				t.Fatal(err)
			}
			if err = writeRestoreManifest(op, m); err != nil {
				t.Fatal(err)
			}
			before := inspectionTree(t, filepath.Dir(s.dir))
			if _, err = s.Recover(); !errors.Is(err, ErrFreshRestoreApprovalRequired) {
				t.Fatal("ordinary startup accepted fresh syntax as execution authority", err)
			}
			if err = s.Restore(id); !errors.Is(err, ErrFreshRestoreApprovalRequired) {
				t.Fatal("normal restore executed a fresh journal", err)
			}
			if err = s.finishRestore(root, journals, op, &m); !errors.Is(err, ErrFreshRestoreApprovalRequired) {
				t.Fatal("normal finish adopted or retired a fresh backup", err)
			}
			stage := filepath.Join(string(m.ParentPath), m.stageName(), "snapshot")
			if qid, err := s.Quarantine(stage, payloadHash("payload"), "Safe.Unapproved.Fresh.Stage"); qid != "" || !errors.Is(err, ErrRestoreIntegrity) || errors.Is(err, ErrAlreadyQuarantined) {
				t.Fatal("fresh syntax alone granted stage ownership", qid, err)
			}
			if !reflect.DeepEqual(before, inspectionTree(t, filepath.Dir(s.dir))) {
				t.Fatal("refused ordinary operation changed fresh evidence")
			}
			requirePurgePayload(t, stored, "payload")
			requirePurgePayload(t, stage, "payload")
			if _, err = os.Lstat(original); !os.IsNotExist(err) {
				t.Fatal("fresh syntax caused publication", err)
			}
			rec, found, err := s.cache.GetQuarantine(id)
			if err != nil || !found || rec.Restored || rec.Purged {
				t.Fatal("fresh syntax claimed historical completion", rec, err)
			}
			op.Close()
			journals.Close()
			root.Close()
		})
	}
}
