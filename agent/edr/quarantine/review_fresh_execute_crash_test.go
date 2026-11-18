package quarantine

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/utmstack/UTMStack/agent/edr/cache"
)

func freshTestCacheOpener(database string) freshCacheOpener {
	return func(_ *freshRestorePermit, _ restoreManifest) (*cache.Cache, func() error, error) {
		c, err := cache.Open(database)
		if err != nil {
			return nil, nil, err
		}
		return c, c.Close, nil
	}
}

func TestFreshExecutionCrashWorker(t *testing.T) {
	if os.Getenv("OPENEDR_FRESH_EXECUTION_CRASH_WORKER") != "1" {
		t.Skip("owned fresh execution crash worker")
	}
	body, err := os.ReadFile(os.Getenv("OPENEDR_FRESH_EXECUTION_APPROVAL"))
	if err != nil {
		t.Fatal(err)
	}
	_, err = executeFreshRestore(os.Getenv("OPENEDR_FRESH_EXECUTION_ROOT"), body, "execute reviewed new restore", freshTestCacheOpener(os.Getenv("OPENEDR_FRESH_EXECUTION_DATABASE")), func(at string) error {
		if at == "fresh-execution-"+os.Getenv("OPENEDR_FRESH_EXECUTION_PHASE") {
			os.Exit(97)
		}
		return nil
	})
	t.Fatal("fresh execution crash checkpoint did not exit", err)
}

func TestFreshExecutionActualProcessCrashes(t *testing.T) {
	self, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	for _, phase := range []string{"operation-created", "ownership-created", "ownership-written", "ticket-published", "output-created-filesystem", "output-identified", "verified", "publication-intent", "published-filesystem", "database-committed", "ticket-retired", "archived"} {
		t.Run(phase, func(t *testing.T) {
			s, id, destination, stored, _, body := freshExecutionFixture(t)
			path := filepath.Join(filepath.Dir(s.dir), "owned-fresh-execution-approval.json")
			if err = os.WriteFile(path, body, 0600); err != nil {
				t.Fatal(err)
			}
			if err = s.cache.Close(); err != nil {
				t.Fatal(err)
			}
			database := filepath.Join(filepath.Dir(s.dir), "edr.db")
			cmd := exec.Command(self, "-test.run=^TestFreshExecutionCrashWorker$", "-test.count=1")
			cmd.Env = append(os.Environ(), "OPENEDR_FRESH_EXECUTION_CRASH_WORKER=1", "OPENEDR_FRESH_EXECUTION_ROOT="+s.dir, "OPENEDR_FRESH_EXECUTION_DATABASE="+database, "OPENEDR_FRESH_EXECUTION_APPROVAL="+path, "OPENEDR_FRESH_EXECUTION_PHASE="+phase)
			output, err := cmd.CombinedOutput()
			if status, ok := err.(*exec.ExitError); !ok || status.ExitCode() != 97 {
				t.Fatalf("fresh checkpoint did not exit97: %v %s", err, output)
			}
			c, err := cache.Open(database)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { c.Close() })
			s.cache = c
			if err = s.Restore(id); !errors.Is(err, ErrRecoveryHeld) {
				t.Fatal("crash released ordinary restore", err)
			}
			if err = s.Purge(id); !errors.Is(err, ErrRecoveryHeld) {
				t.Fatal("crash released purge", err)
			}
			_, err = executeFreshFixture(s, body, nil)
			refuse := phase == "operation-created" || phase == "ownership-created" || phase == "output-created-filesystem" || phase == "publication-intent"
			if (err != nil) != refuse {
				t.Fatal("fresh crash replay adopted unknown evidence or refused exact proof", phase, err)
			}
			rec, found, e := c.GetQuarantine(id)
			if e != nil || !found || rec.Purged || rec.Restored == refuse {
				t.Fatal("fresh crash replay claimed wrong history", rec, e)
			}
			requirePurgePayload(t, stored, "payload")
			if !refuse {
				requirePurgePayload(t, destination, "payload")
			} else if _, e = os.Lstat(destination); !os.IsNotExist(e) {
				t.Fatal("unproven replay published output", phase, e)
			}
		})
	}
}
