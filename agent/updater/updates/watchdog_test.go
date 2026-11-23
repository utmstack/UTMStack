package updates

import (
	"errors"
	"os"
	"testing"

	"github.com/utmstack/UTMStack/shared/logger"
)

func TestMain(m *testing.M) {
	logger.Init(os.DevNull, 200)
	os.Exit(m.Run())
}

func TestEnsureAgentRunningRestartsStoppedAgent(t *testing.T) {
	origExists, origIsActive, origStart := fsExistsFunc, svcIsActiveFunc, svcStartFunc
	defer func() {
		fsExistsFunc, svcIsActiveFunc, svcStartFunc = origExists, origIsActive, origStart
	}()

	started := false
	fsExistsFunc = func(string) bool { return true }
	svcIsActiveFunc = func(string) (bool, error) { return false, nil }
	svcStartFunc = func(string) error {
		started = true
		return nil
	}

	ensureAgentRunning(t.TempDir())
	if !started {
		t.Fatal("expected ensureAgentRunning to start stopped agent service")
	}
}

func TestEnsureAgentRunningNoopWhenActiveOrMissingOrError(t *testing.T) {
	origExists, origIsActive, origStart := fsExistsFunc, svcIsActiveFunc, svcStartFunc
	defer func() {
		fsExistsFunc, svcIsActiveFunc, svcStartFunc = origExists, origIsActive, origStart
	}()

	cases := []struct {
		name     string
		exists   bool
		active   bool
		checkErr error
	}{
		{"already active", true, true, nil},
		{"binary missing", false, false, nil},
		{"status check error", true, false, errors.New("scm unavailable")},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			started := false
			fsExistsFunc = func(string) bool { return tc.exists }
			svcIsActiveFunc = func(string) (bool, error) { return tc.active, tc.checkErr }
			svcStartFunc = func(string) error {
				started = true
				return nil
			}
			ensureAgentRunning(t.TempDir())
			if started {
				t.Fatalf("case %q: expected no start attempt", tc.name)
			}
		})
	}
}
