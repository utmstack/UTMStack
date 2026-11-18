//go:build linux

package dependency

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"golang.org/x/sys/unix"
)

func TestEDRReleaseLinuxReadinessRequiresConfiguredFileProtection(t *testing.T) {
	for _, state := range []string{"degraded", "runtime_unverified", "unsupported", "disabled", ""} {
		t.Run(state, func(t *testing.T) {
			u, start, status := linuxReadinessFixture(t, "permission_open")
			status["sensor_states"] = map[string]any{
				"file_access_monitor": map[string]any{"supported": true, "requested": true, "state": "active"},
				"file_prevention":     map[string]any{"supported": true, "requested": true, "state": state, "error": "owned unavailable permission fixture"},
			}
			writeLinuxReadinessJSON(t, filepath.Join(u.base, "status.json"), status)
			if err := u.checkReady(start); err == nil {
				t.Fatal("healthy engine hid unavailable configured permission owner")
			}
		})
	}
}

func TestEDRReleaseLinuxReadinessScope(t *testing.T) {
	for _, mode := range []string{"notify", "permission_open", "permission_exec"} {
		t.Run(mode, func(t *testing.T) {
			u, start, status := linuxReadinessFixture(t, mode)
			active := func() map[string]any { return map[string]any{"supported": true, "requested": true, "state": "active"} }
			states := map[string]any{"file_access_monitor": active()}
			if mode != "notify" {
				states["file_prevention"] = active()
			}
			status["sensor_states"] = states
			writeLinuxReadinessJSON(t, filepath.Join(u.base, "status.json"), status)
			if err := u.checkReady(start); err != nil {
				t.Fatalf("complete configured protection refused: %v", err)
			}
			for _, key := range []string{"supported", "requested"} {
				access := active()
				access[key] = false
				states["file_access_monitor"] = access
				writeLinuxReadinessJSON(t, filepath.Join(u.base, "status.json"), status)
				if err := u.checkReady(start); err == nil {
					t.Fatalf("unavailable access monitor %s accepted", key)
				}
			}
			delete(states, "file_access_monitor")
			writeLinuxReadinessJSON(t, filepath.Join(u.base, "status.json"), status)
			if err := u.checkReady(start); err == nil {
				t.Fatal("missing requested access monitor accepted")
			}
		})
	}
}

func TestEDRReleaseLinuxReadinessDisabledAndMissingConfiguration(t *testing.T) {
	u, start, status := linuxReadinessFixture(t, "permission_exec")
	writeLinuxReadinessJSON(t, filepath.Join(u.base, "edr.json"), map[string]any{"enabled": false, "linux_protection": map[string]any{"fanotify": map[string]any{"enabled": true, "mode": "permission_exec", "mounts": []string{"/"}, "roots": []string{"/srv/owned"}}}})
	writeLinuxReadinessJSON(t, filepath.Join(u.base, "status.json"), status)
	if err := u.checkReady(start); err != nil {
		t.Fatalf("intentionally disabled protection refused: %v", err)
	}
	writeLinuxReadinessJSON(t, filepath.Join(u.base, "edr.json"), map[string]any{"enabled": true})
	if err := u.checkReady(start); err != nil {
		t.Fatalf("default-off fanotify refused: %v", err)
	}
	if err := os.Remove(filepath.Join(u.base, "edr.json")); err != nil {
		t.Fatal(err)
	}
	if err := u.checkReady(start); err != nil {
		t.Fatalf("legacy status without requested protection refused: %v", err)
	}
	status["sensor_states"] = map[string]any{"file_access_monitor": map[string]any{"supported": true, "requested": true, "state": "active"}}
	writeLinuxReadinessJSON(t, filepath.Join(u.base, "status.json"), status)
	if err := u.checkReady(start); err == nil {
		t.Fatal("lost configuration for reported requested protection accepted")
	}
}

func TestEDRReleaseLinuxReadinessRejectsUnsafeConfiguration(t *testing.T) {
	u, start, status := linuxReadinessFixture(t, "permission_open")
	writeLinuxReadinessJSON(t, filepath.Join(u.base, "status.json"), status)
	path := filepath.Join(u.base, "edr.json")
	for _, data := range []string{"{", strings.Repeat(" ", (1<<20)+1), `{"enabled":true,"linux_protection":{"fanotify":{"enabled":true,"mode":"unknown"}}}`} {
		if err := os.WriteFile(path, []byte(data), 0600); err != nil {
			t.Fatal(err)
		}
		if err := u.checkReady(start); err == nil {
			t.Fatal("invalid configuration accepted as ready")
		}
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	if err := unix.Mkfifo(path, 0600); err != nil {
		t.Fatal(err)
	}
	finished := make(chan error, 1)
	go func() { finished <- u.checkReady(start) }()
	select {
	case err := <-finished:
		if err == nil {
			t.Fatal("FIFO configuration accepted")
		}
	case <-time.After(time.Second):
		t.Fatal("FIFO parked readiness verifier")
	}
}

func linuxReadinessFixture(t *testing.T, mode string) (*releaseEDRUpdate, time.Time, map[string]any) {
	t.Helper()
	u := newReleaseEDRUpdate(t.TempDir())
	u.expectedServiceSHA = strings.Repeat("a", 64)
	start := time.Now().UTC()
	writeLinuxReadinessJSON(t, filepath.Join(u.base, "edr.json"), map[string]any{"enabled": true, "linux_protection": map[string]any{"fanotify": map[string]any{"enabled": true, "mode": mode, "mounts": []string{"/"}, "roots": []string{"/srv/owned"}}}})
	status := map[string]any{"instance_id": "owned-linux-fixture", "process_id": 1234, "started_at": start.Add(time.Second).Format(time.RFC3339Nano), "running": true, "engine_healthy": true, "engine_implementation": u.identity.Implementation, "service_executable_sha256": u.expectedServiceSHA}
	return u, start, status
}

func writeLinuxReadinessJSON(t *testing.T, path string, value any) {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
}
