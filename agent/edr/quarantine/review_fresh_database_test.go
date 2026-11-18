package quarantine

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFreshExecutionDatabaseCannotAliasRecoveryEvidence(t *testing.T) {
	for _, which := range []string{"snapshot", "snapshot-hardlink", "raw", "approval", "original-journal", "cancellation", "hold", "published-output", "fresh-journal-hardlink", "ownership-hardlink", "ticket-hardlink"} {
		t.Run(which, func(t *testing.T) {
			s, id, destination, stored, a, body := freshExecutionFixture(t)
			var hold RecoveryHold
			if err := json.Unmarshal(a.Inspection.Hold.Manifest, &hold); err != nil {
				t.Fatal(err)
			}
			old, err := decodeRestoreManifest(a.Inspection.Original.Manifest, id)
			if err != nil {
				t.Fatal(err)
			}
			path := stored
			switch which {
			case "snapshot-hardlink":
				path = filepath.Join(filepath.Dir(s.dir), "owned-snapshot-alias.db")
				if err = os.Link(stored, path); err != nil {
					t.Fatal(err)
				}
			case "raw":
				path = filepath.Join(old.Record.SourceCaptureDir, "source")
			case "approval":
				path = filepath.Join(s.dir, ".fresh-restores", id, a.OperationID, "manifest.json")
			case "original-journal":
				path = filepath.Join(s.dir, ".reviews", hold.DecisionID, "restore", "manifest.json")
			case "cancellation":
				path = filepath.Join(s.dir, ".reviews", hold.DecisionID, "manifest.json")
			case "hold":
				path = filepath.Join(s.dir, ".purges", id)
			case "published-output":
				_, err = executeFreshFixture(s, body, func(at string) error {
					if at == "fresh-execution-published" {
						return errors.New("owned published checkpoint")
					}
					return nil
				})
				if err == nil {
					t.Fatal("published checkpoint not reached")
				}
				path = destination
			}
			if strings.HasSuffix(which, "-hardlink") && which != "snapshot-hardlink" {
				_, err = executeFreshFixture(s, body, func(at string) error {
					if at == "fresh-execution-verified" {
						return errors.New("owned alias guard checkpoint")
					}
					return nil
				})
				if err == nil {
					t.Fatal("verified checkpoint not reached")
				}
				var target string
				switch which {
				case "fresh-journal-hardlink":
					target = filepath.Join(s.dir, ".restores", id, "manifest.json")
				case "ownership-hardlink":
					target = filepath.Join(s.dir, ".fresh-restores", id, a.OperationID, "ownership.json")
				case "ticket-hardlink":
					target = filepath.Join(filepath.Dir(destination), ".edr-restore-"+a.OperationID, freshTicketName)
				}
				path = filepath.Join(filepath.Dir(s.dir), "owned-current-control-alias.db")
				if err = os.Link(target, path); err != nil {
					t.Fatal(err)
				}
			}
			before, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			_, err = ExecuteFreshRestore(s.dir, path, body, "execute reviewed new restore")
			if strings.HasSuffix(which, "-hardlink") && (err == nil || !strings.Contains(err.Error(), "aliases protected recovery evidence")) {
				t.Fatal("hard-link alias was not refused before opening SQLite", err)
			}
			if err == nil {
				t.Fatal("recovery evidence was opened as a database", which)
			}
			after, err := os.ReadFile(path)
			if err != nil || !bytes.Equal(before, after) {
				t.Fatal("database guard changed protected evidence", which, err)
			}
			requirePurgePayload(t, stored, "payload")
		})
	}
}
