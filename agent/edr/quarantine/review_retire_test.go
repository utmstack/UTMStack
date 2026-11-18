package quarantine

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/utmstack/UTMStack/agent/edr/cache"
)

func retireFixture(t *testing.T, state string) (*Store, string, string, FreshRestoreApproval, CancelledRetirementInspection, CancelledRecoveryRetirement) {
	t.Helper()
	s, id, hold, stage, a := retirementInspectionFixture(t, state)
	r, err := InspectCancelledRetirement(s.dir, id, hold, stage)
	if err != nil {
		t.Fatal(err)
	}
	m, err := RetireCancelledRecovery(s.dir, r, "permanently retire reviewed cancellation; retain all data", "")
	if err != nil {
		t.Fatal(err)
	}
	return s, id, stage, a, r, m
}

func TestRetirementArchivesSameHoldAndNeverDeletesOrPublishes(t *testing.T) {
	for _, state := range []string{"approval-only", "committed", "cancelled", "cancelled-malformed"} {
		t.Run(state, func(t *testing.T) {
			s, id, hold, stage, a := retirementInspectionFixture(t, state)
			r, err := InspectCancelledRetirement(s.dir, id, hold, stage)
			if err != nil {
				t.Fatal(err)
			}
			before, found, err := s.cache.GetQuarantine(id)
			if err != nil || !found {
				t.Fatal(err)
			}
			old, err := decodeRestoreManifest(r.Original.Manifest, id)
			if err != nil {
				t.Fatal(err)
			}
			oldStageSnapshot := filepath.Join(stage, "snapshot")
			originalJournal, err := os.ReadFile(filepath.Join(s.dir, ".reviews", hold, "restore", "manifest.json"))
			if err != nil {
				t.Fatal(err)
			}
			m, err := RetireCancelledRecovery(s.dir, r, "permanently retire reviewed cancellation; retain all data", "")
			if err != nil {
				t.Fatal(err)
			}
			if m.Kind != "cancelled-restore-retirement" || m.RootID != r.RootID || m.StageTicketID == "" || m.MarkerID == "" || m.ActorID == "" {
				t.Fatal("incomplete retirement decision", m)
			}
			if _, err = os.Lstat(filepath.Join(s.dir, ".purges", id)); !os.IsNotExist(err) {
				t.Fatal("active original fence was not archived", err)
			}
			guard, err := os.Lstat(filepath.Join(s.dir, ".restores", id))
			if err != nil || !guard.Mode().IsRegular() {
				t.Fatal("permanent legacy guard missing", err)
			}
			f, err := os.Open(filepath.Join(s.dir, ".reviews", hold, "hold.json"))
			if err != nil {
				t.Fatal(err)
			}
			fid, err := admissionFileID(f)
			f.Close()
			if err != nil || fid != r.Hold.ManifestID {
				t.Fatal("retirement archived another fence", err)
			}
			for _, p := range []string{filepath.Join(s.dir, id+".quarantined"), filepath.Join(old.Record.SourceCaptureDir, "source"), oldStageSnapshot} {
				requirePurgePayload(t, p, "payload")
			}
			journal, err := os.ReadFile(filepath.Join(s.dir, ".reviews", hold, "restore", "manifest.json"))
			if err != nil || !bytes.Equal(originalJournal, journal) {
				t.Fatal("original journal changed", err)
			}
			if state == "cancelled" || state == "cancelled-malformed" {
				p := filepath.Join(s.dir, ".fresh-restores", id, a.OperationID, "restore", "unknown-retirement-evidence")
				data, e := os.ReadFile(p)
				if e != nil || string(data) != "preserve unknown child" {
					t.Fatal("unknown fresh archive evidence was changed", e)
				}
			}
			after, found, err := s.cache.GetQuarantine(id)
			if err != nil || !found || !reflect.DeepEqual(before, after) {
				t.Fatal("retirement changed SQLite disposition", err)
			}
			if err = s.Restore(id); !errors.Is(err, ErrRecoveryRetired) {
				t.Fatal("ordinary restore ignored retirement", err)
			}
			if err = s.Purge(id); !errors.Is(err, ErrRecoveryRetired) {
				t.Fatal("ordinary purge gained deletion permission", err)
			}
			if n, e := s.PurgeExpired(1, time.Now().AddDate(10, 0, 0)); e != nil || n != 0 {
				t.Fatal("automatic retention did not retain retirement data", n, e)
			}
			results, err := s.Recover()
			if err != nil {
				t.Fatal(err)
			}
			retired := false
			for _, result := range results {
				if result.QuarantineID == id && result.Operation == "retirement" && result.Status == "retired" {
					retired = true
				}
			}
			if !retired {
				t.Fatal("startup did not report permanent retirement", results)
			}
			beforeRetry := inspectionTree(t, filepath.Dir(s.dir))
			again, err := RetireCancelledRecovery(s.dir, r, m.Reason, "")
			if err != nil || !reflect.DeepEqual(m, again) || !reflect.DeepEqual(beforeRetry, inspectionTree(t, filepath.Dir(s.dir))) {
				t.Fatal("historical exact retirement retry changed evidence", err)
			}
		})
	}
}

func TestRetirementProtectsOldAndFreshStagesAfterLostOrStaleDatabase(t *testing.T) {
	for _, state := range []string{"committed", "cancelled", "cancelled-malformed"} {
		t.Run(state, func(t *testing.T) {
			s, id, stage, a, r, _ := retireFixture(t, state)
			old, err := decodeRestoreManifest(r.Original.Manifest, id)
			if err != nil {
				t.Fatal(err)
			}
			for _, stale := range []bool{false, true} {
				c, e := cache.Open(filepath.Join(t.TempDir(), "recreated.db"))
				if e != nil {
					t.Fatal(e)
				}
				defer c.Close()
				if stale {
					if e = c.CommitQuarantine(old.Record); e != nil {
						t.Fatal(e)
					}
				}
				store, e := New(s.dir, c)
				if e != nil {
					t.Fatal(e)
				}
				if e = store.Restore(id); !errors.Is(e, ErrRecoveryRetired) {
					t.Fatal("stale/missing DB re-enabled v1 restore", e)
				}
				if e = store.Purge(id); !errors.Is(e, ErrRecoveryRetired) {
					t.Fatal("stale/missing DB permitted purge", e)
				}
				oldPath := filepath.Join(stage, "snapshot")
				if q, e := store.Quarantine(oldPath, payloadHash("payload"), "Safe.Retired.Stage"); q != "" || !errors.Is(e, ErrAlreadyQuarantined) {
					t.Fatal("original stage lost bounded ownership without SQL", q, e)
				}
				if state != "committed" {
					freshPath := filepath.Join(string(a.Inspection.Destination.ParentPath), ".edr-restore-"+a.OperationID, "snapshot")
					if q, e := store.Quarantine(freshPath, payloadHash("payload"), "Safe.Retired.Fresh"); q != "" || !errors.Is(e, ErrAlreadyQuarantined) {
						t.Fatal("retired fresh stage lost ownership without SQL", q, e)
					}
				}
			}
			if _, err = os.Stat(filepath.Join(s.dir, id+".quarantined")); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestRetirementOwnershipFollowsRenamedStagesAcrossAnotherStore(t *testing.T) {
	s, id, stage, a, _, _ := retireFixture(t, "cancelled")
	renamed := filepath.Join(filepath.Dir(stage), "original-retired-stage-renamed")
	if e := os.Rename(stage, renamed); e != nil {
		t.Fatal(e)
	}
	fresh := filepath.Join(string(a.Inspection.Destination.ParentPath), ".edr-restore-"+a.OperationID)
	freshRenamed := filepath.Join(filepath.Dir(fresh), "fresh-retired-stage-renamed")
	if e := os.Rename(fresh, freshRenamed); e != nil {
		t.Fatal(e)
	}
	c, e := cache.Open(filepath.Join(t.TempDir(), "other.db"))
	if e != nil {
		t.Fatal(e)
	}
	defer c.Close()
	other, e := New(filepath.Join(t.TempDir(), "another-store"), c)
	if e != nil {
		t.Fatal(e)
	}
	for _, p := range []string{filepath.Join(renamed, "snapshot"), filepath.Join(freshRenamed, "snapshot")} {
		if q, e := other.Quarantine(p, payloadHash("payload"), "Safe.Retired.Other"); q != "" || !errors.Is(e, ErrAlreadyQuarantined) {
			t.Fatal("another store or stage rename erased retirement ownership", q, e)
		}
	}
	if _, e = os.Stat(filepath.Join(s.dir, ".restores", id)); e != nil {
		t.Fatal(e)
	}
}

func TestRetirementPartialAndCompletePreparationRetries(t *testing.T) {
	for _, phase := range []string{"receipt-created", "marker-created", "ticket-created", "receipt-written", "approved", "ticket-prepared", "marker-prepared", "marker-published", "before-hold-move", "hold-archived", "synced"} {
		t.Run(phase, func(t *testing.T) {
			s, id, hold, stage, _ := retirementInspectionFixture(t, "cancelled")
			r, e := InspectCancelledRetirement(s.dir, id, hold, stage)
			if e != nil {
				t.Fatal(e)
			}
			_, e = retireCancelledRecovery(s.dir, r, "retain every object during retirement", "", func(at string) error {
				if at == "retirement-"+phase {
					return errors.New("owned retirement interruption")
				}
				return nil
			})
			if e == nil {
				t.Fatal("fault phase not reached")
			}
			_, e = RetireCancelledRecovery(s.dir, r, "retain every object during retirement", "")
			partial := phase == "receipt-created" || phase == "marker-created" || phase == "ticket-created"
			if (e != nil) != partial {
				t.Fatal("retirement adopted incomplete decision or refused exact durable retry", phase, e)
			}
			if partial {
				if _, e = os.Stat(filepath.Join(s.dir, ".purges", id)); e != nil {
					t.Fatal("partial retirement released original fence", e)
				}
			} else {
				if e = s.Restore(id); !errors.Is(e, ErrRecoveryRetired) {
					t.Fatal("complete retry did not fence v1 replay", e)
				}
			}
			requirePurgePayload(t, filepath.Join(s.dir, id+".quarantined"), "payload")
			requirePurgePayload(t, filepath.Join(stage, "snapshot"), "payload")
		})
	}
}

func TestRetirementPreparedControlRequiresItsBoundIdentityAndCanonicalPrefix(t *testing.T) {
	for _, control := range []string{"ticket", "marker"} {
		for _, damage := range []string{"empty", "prefix", "different", "suffix", "copy", "missing"} {
			t.Run(control+"/"+damage, func(t *testing.T) {
				s, id, hold, stage, _ := retirementInspectionFixture(t, "cancelled")
				r, e := InspectCancelledRetirement(s.dir, id, hold, stage)
				if e != nil {
					t.Fatal(e)
				}
				m, e := retireCancelledRecovery(s.dir, r, "complete only this reviewed preparation", "", func(at string) error {
					if at == "retirement-approved" {
						return errors.New("owned prepared control checkpoint")
					}
					return nil
				})
				if e == nil || m.FileID == "" {
					t.Fatal("complete decision checkpoint missing", e)
				}
				root, e := openAdmissionRoot(s.dir, false)
				if e != nil {
					t.Fatal(e)
				}
				path := filepath.Join(stage, retirementTicketName)
				_, expected, e := expectedRetirementTicket(root, m)
				if control == "marker" {
					path = filepath.Join(s.dir, ".reviews", hold, retirementPreparedName)
					_, expected, e = expectedRetirementMarker(root, m)
				}
				root.Close()
				if e != nil {
					t.Fatal(e)
				}
				body := []byte{}
				switch damage {
				case "prefix":
					body = expected[:len(expected)/2]
				case "different":
					body = []byte("different preparation")
				case "suffix":
					body = append(append([]byte{}, expected...), 'x')
				case "copy", "missing":
					if e = os.Rename(path, path+".retained"); e != nil {
						t.Fatal(e)
					}
				}
				if damage != "missing" {
					if e = os.WriteFile(path, body, 0600); e != nil {
						t.Fatal(e)
					}
				}
				_, e = RetireCancelledRecovery(s.dir, r, m.Reason, "")
				accepted := damage == "empty" || damage == "prefix"
				if (e == nil) != accepted {
					t.Fatal("retirement accepted changed preparation or refused its exact prefix", e)
				}
				if !accepted {
					if _, e = os.Stat(filepath.Join(s.dir, ".purges", id)); e != nil {
						t.Fatal("changed preparation removed the original fence", e)
					}
				}
				requirePurgePayload(t, filepath.Join(stage, "snapshot"), "payload")
				requirePurgePayload(t, filepath.Join(s.dir, id+".quarantined"), "payload")
			})
		}
	}
}

func TestRetirementDoesNotOpenOrRepairDatabase(t *testing.T) {
	s, id, hold, stage, _ := retirementInspectionFixture(t, "cancelled")
	r, e := InspectCancelledRetirement(s.dir, id, hold, stage)
	if e != nil {
		t.Fatal(e)
	}
	if e = s.cache.Close(); e != nil {
		t.Fatal(e)
	}
	db := filepath.Join(filepath.Dir(s.dir), "edr.db")
	invalid := []byte("retirement must not open this invalid database")
	if e = os.WriteFile(db, invalid, 0600); e != nil {
		t.Fatal(e)
	}
	if _, e = RetireCancelledRecovery(s.dir, r, "retain all data without opening database", ""); e != nil {
		t.Fatal(e)
	}
	body, e := os.ReadFile(db)
	if e != nil || !bytes.Equal(body, invalid) {
		t.Fatal("retirement changed the unusable database", e)
	}
}

func TestRetirementDoesNotRequireAnyFreshApproval(t *testing.T) {
	s, id, _, _, a := freshInspectionFixture(t)
	body, e := os.ReadFile(filepath.Join(s.dir, ".reviews", a.HoldDecisionID, "restore", "manifest.json"))
	if e != nil {
		t.Fatal(e)
	}
	old, e := decodeRestoreManifest(body, id)
	if e != nil {
		t.Fatal(e)
	}
	stage := filepath.Join(string(old.ParentPath), old.stageName())
	r, e := InspectCancelledRetirement(s.dir, id, a.HoldDecisionID, stage)
	if e != nil || len(r.Fresh) != 0 {
		t.Fatal("original-only retirement review requires unrelated fresh history", e)
	}
	if _, e = RetireCancelledRecovery(s.dir, r, "retain cancelled original without a fresh restore", ""); e != nil {
		t.Fatal(e)
	}
	if e = s.Restore(id); !errors.Is(e, ErrRecoveryRetired) {
		t.Fatal("original-only retirement allowed ordinary replay", e)
	}
	requirePurgePayload(t, filepath.Join(stage, "snapshot"), "payload")
}

func TestRetirementRefusesStaleInspectionBeforePreparation(t *testing.T) {
	for _, changed := range []string{"saved-root", "saved-stage", "saved-hold", "snapshot-bytes", "raw-bytes", "ticket-collision", "reason"} {
		t.Run(changed, func(t *testing.T) {
			s, id, hold, stage, _ := retirementInspectionFixture(t, "committed")
			r, e := InspectCancelledRetirement(s.dir, id, hold, stage)
			if e != nil {
				t.Fatal(e)
			}
			reason := "retire exactly reviewed original"
			switch changed {
			case "saved-root":
				r.RootID += "-changed"
			case "saved-stage":
				r.OldStage.DirectoryID += "-changed"
			case "saved-hold":
				r.Hold.ManifestID += "-changed"
			case "snapshot-bytes":
				if e = os.WriteFile(filepath.Join(s.dir, id+".quarantined"), []byte("changed"), 0600); e != nil {
					t.Fatal(e)
				}
			case "raw-bytes":
				old, e := decodeRestoreManifest(r.Original.Manifest, id)
				if e != nil {
					t.Fatal(e)
				}
				if e = os.WriteFile(filepath.Join(old.Record.SourceCaptureDir, "source"), []byte("changed"), 0600); e != nil {
					t.Fatal(e)
				}
			case "ticket-collision":
				if e = os.WriteFile(filepath.Join(stage, retirementTicketName), []byte("unknown evidence"), 0600); e != nil {
					t.Fatal(e)
				}
			case "reason":
				reason = " \t"
			}
			before := inspectionTree(t, filepath.Dir(s.dir))
			if _, e = RetireCancelledRecovery(s.dir, r, reason, ""); e == nil {
				t.Fatal("stale retirement evidence accepted")
			}
			if !reflect.DeepEqual(before, inspectionTree(t, filepath.Dir(s.dir))) {
				t.Fatal("stale refusal changed evidence")
			}
		})
	}
}

func TestRetirementRevalidatesStageTicketImmediatelyBeforeFenceMove(t *testing.T) {
	s, id, hold, stage, _ := retirementInspectionFixture(t, "cancelled")
	r, e := InspectCancelledRetirement(s.dir, id, hold, stage)
	if e != nil {
		t.Fatal(e)
	}
	_, e = retireCancelledRecovery(s.dir, r, "retire after exact revalidation", "", func(at string) error {
		if at == "retirement-before-hold-move" {
			return os.WriteFile(filepath.Join(stage, retirementTicketName), []byte("changed ownership ticket"), 0600)
		}
		return nil
	})
	if e == nil {
		t.Fatal("retirement moved fence after ticket changed")
	}
	if _, e = os.Stat(filepath.Join(s.dir, ".purges", id)); e != nil {
		t.Fatal("original fence moved without durable stage ownership", e)
	}
}

func TestRetirementCorruptBoundControlRetainsDataWithoutHints(t *testing.T) {
	for _, changed := range []string{"guard-bytes", "guard-copy", "guard-missing", "receipt-bytes", "original-archive", "fresh-approval", "fresh-owner", "fresh-journal"} {
		t.Run(changed, func(t *testing.T) {
			s, id, stage, a, _, m := retireFixture(t, "cancelled")
			oldPath := filepath.Join(stage, "snapshot")
			freshPath := filepath.Join(string(a.Inspection.Destination.ParentPath), ".edr-restore-"+a.OperationID, "snapshot")
			target := oldPath
			copyFile := func(p string) {
				t.Helper()
				b, e := os.ReadFile(p)
				if e != nil {
					t.Fatal(e)
				}
				if e = os.Rename(p, p+".retained"); e != nil {
					t.Fatal(e)
				}
				if e = os.WriteFile(p, b, 0600); e != nil {
					t.Fatal(e)
				}
			}
			switch changed {
			case "guard-bytes":
				if e := os.WriteFile(filepath.Join(s.dir, ".restores", id), []byte("{partial guard"), 0600); e != nil {
					t.Fatal(e)
				}
			case "guard-copy":
				copyFile(filepath.Join(s.dir, ".restores", id))
			case "guard-missing":
				if e := os.Rename(filepath.Join(s.dir, ".restores", id), filepath.Join(s.dir, ".restores", "retained-guard")); e != nil {
					t.Fatal(e)
				}
			case "receipt-bytes":
				if e := os.WriteFile(filepath.Join(s.dir, ".reviews", m.HoldID, retirementReceiptName), []byte("{partial decision"), 0600); e != nil {
					t.Fatal(e)
				}
			case "original-archive":
				copyFile(filepath.Join(s.dir, ".reviews", m.HoldID, "restore", "manifest.json"))
			case "fresh-approval":
				copyFile(filepath.Join(s.dir, ".fresh-restores", id, a.OperationID, "manifest.json"))
				target = freshPath
			case "fresh-owner":
				copyFile(filepath.Join(s.dir, ".fresh-restores", id, a.OperationID, "ownership.json"))
				target = freshPath
			case "fresh-journal":
				copyFile(filepath.Join(s.dir, ".fresh-restores", id, a.OperationID, "restore", "manifest.json"))
				target = freshPath
			}
			c, e := cache.Open(filepath.Join(t.TempDir(), "empty.db"))
			if e != nil {
				t.Fatal(e)
			}
			defer c.Close()
			other, e := New(filepath.Join(t.TempDir(), "other"), c)
			if e != nil {
				t.Fatal(e)
			}
			if q, e := other.Quarantine(target, payloadHash("payload"), "Safe.Retired.Corrupt"); q != "" || !errors.Is(e, ErrRestoreIntegrity) {
				t.Fatal("damaged bound retirement metadata permitted admission", q, e)
			}
			requirePurgePayload(t, target, "payload")
		})
	}
}

func TestRetirementPreparedStageTicketSurvivesLostDatabaseBeforeGuardPublication(t *testing.T) {
	for _, phase := range []string{"ticket-prepared", "marker-prepared", "marker-published"} {
		t.Run(phase, func(t *testing.T) {
			s, id, hold, stage, _ := retirementInspectionFixture(t, "cancelled")
			r, e := InspectCancelledRetirement(s.dir, id, hold, stage)
			if e != nil {
				t.Fatal(e)
			}
			_, e = retireCancelledRecovery(s.dir, r, "retain data during guard publication", "", func(at string) error {
				if at == "retirement-"+phase {
					return errors.New("owned pre-publication interruption")
				}
				return nil
			})
			if e == nil {
				t.Fatal("publication checkpoint not reached")
			}
			c, e := cache.Open(filepath.Join(t.TempDir(), "empty.db"))
			if e != nil {
				t.Fatal(e)
			}
			defer c.Close()
			other, e := New(filepath.Join(t.TempDir(), "other"), c)
			if e != nil {
				t.Fatal(e)
			}
			path := filepath.Join(stage, "snapshot")
			if q, e := other.Quarantine(path, payloadHash("payload"), "Safe.Prepared.Retirement"); q != "" || !errors.Is(e, ErrAlreadyQuarantined) {
				t.Fatal("prepared retirement ticket lost protection without SQLite", q, e)
			}
			requirePurgePayload(t, path, "payload")
			if _, e = os.Stat(filepath.Join(s.dir, ".purges", id)); e != nil {
				t.Fatal("pre-publication interruption moved the original hold", e)
			}
		})
	}
}

func TestRetirementHistoricalRetryDoesNotAdoptLaterStageOrRequireOldBackups(t *testing.T) {
	s, id, stage, _, r, m := retireFixture(t, "committed")
	old, e := decodeRestoreManifest(r.Original.Manifest, id)
	if e != nil {
		t.Fatal(e)
	}
	renamed := stage + "-retained"
	if e = os.Rename(stage, renamed); e != nil {
		t.Fatal(e)
	}
	if e = os.Mkdir(stage, 0700); e != nil {
		t.Fatal(e)
	}
	if e = os.WriteFile(filepath.Join(stage, "foreign"), []byte("later directory"), 0600); e != nil {
		t.Fatal(e)
	}
	// A historical ownership proof must survive an independently retired data
	// backup. This test removes its own fixture copies; the API never does so.
	if e = os.Remove(filepath.Join(s.dir, id+".quarantined")); e != nil {
		t.Fatal(e)
	}
	if e = os.Remove(filepath.Join(old.Record.SourceCaptureDir, "source")); e != nil {
		t.Fatal(e)
	}
	before := inspectionTree(t, filepath.Dir(s.dir))
	again, e := RetireCancelledRecovery(s.dir, r, m.Reason, "")
	if e != nil || !reflect.DeepEqual(m, again) || !reflect.DeepEqual(before, inspectionTree(t, filepath.Dir(s.dir))) {
		t.Fatal("historical retirement retry adopted later data", e)
	}
	c, e := cache.Open(filepath.Join(t.TempDir(), "empty.db"))
	if e != nil {
		t.Fatal(e)
	}
	defer c.Close()
	other, e := New(filepath.Join(t.TempDir(), "other"), c)
	if e != nil {
		t.Fatal(e)
	}
	if q, e := other.Quarantine(filepath.Join(renamed, "snapshot"), payloadHash("payload"), "Safe.Retired.History"); q != "" || !errors.Is(e, ErrAlreadyQuarantined) {
		t.Fatal("old-stage ownership depended on live backup or SQL", q, e)
	}
	if _, e = json.Marshal(again); e != nil {
		t.Fatal(e)
	}
}
