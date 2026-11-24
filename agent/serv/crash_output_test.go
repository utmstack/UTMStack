package serv

import (
	"os"
	"path/filepath"
	"runtime/debug"
	"testing"
)

func TestSetupCrashOutputCreatesCrashLog(t *testing.T) {
	defer func() {
		_ = debug.SetCrashOutput(nil, debug.CrashOptions{})
	}()

	logDir := filepath.Join(t.TempDir(), "logs")
	logFile := filepath.Join(logDir, "utmstack_agent.log")

	if err := setupCrashOutput(logFile); err != nil {
		t.Fatalf("setupCrashOutput: %v", err)
	}

	crashPath := filepath.Join(logDir, "utmstack_agent_crash.log")
	if _, err := os.Stat(crashPath); err != nil {
		t.Fatalf("expected crash log %q to exist: %v", crashPath, err)
	}
}
