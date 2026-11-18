//go:build linux

package dependency

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/utmstack/UTMStack/shared/fs"
)

func TestLinuxReleaseReadinessRequiresConfiguredNetworkEnforcement(t *testing.T) {
	u, start, status := linuxReadinessFixture(t, "notify")
	// Omitted blocklist fields inherit the loader's enabled/enforce defaults.
	for _, block := range []any{nil, map[string]any{}, map[string]any{"enabled": true, "enforce": true}} {
		cfg := map[string]any{"enabled": true, "linux_protection": map[string]any{"network_enforcement": true}}
		if block != nil {
			cfg["blocklist"] = block
		}
		writeLinuxReadinessJSON(t, filepath.Join(u.base, "edr.json"), cfg)
		for _, state := range []string{"degraded", "unsupported", "disabled", "runtime_unverified", ""} {
			status["sensor_states"] = map[string]any{"network_enforcement": map[string]any{"supported": true, "requested": true, "state": state}}
			writeLinuxReadinessJSON(t, filepath.Join(u.base, "status.json"), status)
			if err := u.checkReady(start); err == nil {
				t.Fatal("unavailable required network enforcement accepted", state, block)
			}
		}
		status["sensor_states"] = map[string]any{"network_enforcement": map[string]any{"supported": true, "requested": true, "state": "active"}}
		writeLinuxReadinessJSON(t, filepath.Join(u.base, "status.json"), status)
		if err := u.checkReady(start); err != nil {
			t.Fatal("active configured network protection refused", err)
		}
	}
	for _, block := range []map[string]any{{"enabled": false}, {"enforce": false}} {
		writeLinuxReadinessJSON(t, filepath.Join(u.base, "edr.json"), map[string]any{"enabled": true, "linux_protection": map[string]any{"network_enforcement": true}, "blocklist": block})
		delete(status, "sensor_states")
		writeLinuxReadinessJSON(t, filepath.Join(u.base, "status.json"), status)
		if err := u.checkReady(start); err != nil {
			t.Fatal("intentionally disabled network enforcement refused", err)
		}
	}
	status["sensor_states"] = map[string]any{"network_enforcement": map[string]any{"requested": true}}
	writeLinuxReadinessJSON(t, filepath.Join(u.base, "status.json"), status)
	if err := os.Remove(filepath.Join(u.base, "edr.json")); err != nil {
		t.Fatal(err)
	}
	if err := u.checkReady(start); err == nil {
		t.Fatal("requested networking without local configuration accepted")
	}
}

// Read-only gate against the actual owned foreground service's fresh status.
func TestNativeLinuxNetworkReadinessFixture(t *testing.T) {
	if os.Getenv("OPENEDR_NATIVE_NETWORK_READINESS") != "1" {
		t.Skip("requires owned native network acceptance")
	}
	base := "/opt/openedr-network"
	host, _ := os.Hostname()
	marker, _ := os.ReadFile("/run/openedr-disposable-lab")
	if os.Geteuid() != 0 || host != "openedr-linux-owned" || fs.GetExecutablePath() != base || strings.TrimSpace(string(marker)) != "OPENEDR_DISPOSABLE_LAB_ONLY" {
		t.Fatal("network readiness fixture requires the exact owned guest")
	}
	self, _ := os.Readlink("/proc/self/ns/net")
	initial, _ := os.Readlink("/proc/1/ns/net")
	if self == "" || self == initial {
		t.Fatal("network readiness fixture requires the owned private namespace")
	}
	f, err := os.Open(filepath.Join(base, EDRFile("")))
	if err != nil {
		t.Fatal(err)
	}
	digest := sha256.New()
	_, copyErr := io.Copy(digest, f)
	closeErr := f.Close()
	if copyErr != nil || closeErr != nil {
		t.Fatal(copyErr, closeErr)
	}
	u := newReleaseEDRUpdate(base)
	u.expectedServiceSHA = hex.EncodeToString(digest.Sum(nil))
	err = u.checkReady(time.Time{})
	wantReady := os.Getenv("OPENEDR_NETWORK_WANT_READY") == "1"
	if (err == nil) != wantReady {
		t.Fatal("actual network status violated release readiness", wantReady, err)
	}
	if !wantReady && !strings.Contains(err.Error(), "network_enforcement") {
		t.Fatal("unready for a different reason", err)
	}
	result := map[string]any{"want_ready": wantReady, "architecture": runtime.GOARCH}
	if err != nil {
		result["error"] = err.Error()
	}
	data, _ := json.Marshal(result)
	fmt.Printf("NATIVE_RELEASE_NETWORK_GATE %s\n", data)
}
