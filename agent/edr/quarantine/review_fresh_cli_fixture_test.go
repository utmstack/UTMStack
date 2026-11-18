package quarantine

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestFreshCLIFixtureWorker(t *testing.T) {
	if os.Getenv("OPENEDR_FRESH_CLI_FIXTURE_WORKER") != "1" {
		t.Skip("owned standalone CLI fixture worker")
	}
	control := os.Getenv("OPENEDR_FRESH_CLI_FIXTURE_CONTROL")
	if !filepath.IsAbs(control) {
		t.Fatal("owned CLI control path must be absolute")
	}
	s, id, destination, _, a, approval := freshExecutionFixture(t)
	if os.Getenv("OPENEDR_FRESH_CLI_FIXTURE_MODE") == "cancel" {
		reached := false
		_, err := executeFreshFixture(s, approval, func(at string) error {
			if at == "fresh-execution-verified" {
				reached = true
				return errors.New("owned CLI cancellation checkpoint")
			}
			return nil
		})
		if err == nil || !reached {
			t.Fatal("CLI cancellation checkpoint not reached", err)
		}
		if err = os.WriteFile(filepath.Join(s.dir, ".restores", id, "unknown-owned-cli-evidence"), []byte("preserve CLI evidence"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	if err := s.cache.Close(); err != nil {
		t.Fatal(err)
	}
	body, err := json.Marshal(map[string]any{"root": s.dir, "database": filepath.Join(filepath.Dir(s.dir), "edr.db"), "id": id, "destination": destination, "inspection": a.Inspection, "operation": a.OperationID})
	if err != nil {
		t.Fatal(err)
	}
	f, err := os.OpenFile(control, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = f.Write(body); err != nil {
		t.Fatal(err)
	}
	if err = f.Sync(); err != nil {
		t.Fatal(err)
	}
	if err = f.Close(); err != nil {
		t.Fatal(err)
	}
	// The parent supplied its own disposable TMP/TEMP/TMPDIR. Preserve these
	// inert fixtures for that parent's real CLI processes and final cleanup.
	os.Exit(0)
}
