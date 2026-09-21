// edr/internal/mirror/status_test.go
package mirror

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func readStatus(t *testing.T, root string) map[string]any {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(root, "status.json"))
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}
	return m
}

func TestRecorderWritesFeedSuccessAndPreservesItOnFailure(t *testing.T) {
	root := t.TempDir()
	r := NewRecorder(root)
	r.FeedSuccess("level1/ip", 42, "abc123")
	m := readStatus(t, root)
	feeds := m["feeds"].(map[string]any)
	f := feeds["level1/ip"].(map[string]any)
	if f["indicators"].(float64) != 42 || f["sha256"] != "abc123" {
		t.Fatalf("feed status = %v", f)
	}
	lastSuccess := f["last_success"]

	r.FeedFailure("level1/ip", "upstream 503")
	m = readStatus(t, root)
	f = m["feeds"].(map[string]any)["level1/ip"].(map[string]any)
	if f["last_error"] != "upstream 503" {
		t.Fatalf("last_error = %v", f["last_error"])
	}
	if f["last_success"] != lastSuccess || f["indicators"].(float64) != 42 {
		t.Fatal("failure must preserve previous success fields")
	}
}

func TestRecorderPreservesSuccessAcrossRestart(t *testing.T) {
	root := t.TempDir()

	r := NewRecorder(root)
	r.FeedSuccess("level1/ip", 42, "abc123")
	r.SigSuccess(map[string]string{"daily.cvd": "27461"})

	m := readStatus(t, root)
	f := m["feeds"].(map[string]any)["level1/ip"].(map[string]any)
	feedSuccess := f["last_success"]
	sigSuccess := m["signatures"].(map[string]any)["last_success"]

	// Simulated process restart: a fresh Recorder over the same root.
	r2 := NewRecorder(root)
	r2.FeedFailure("level1/ip", "down")
	r2.SigFailure("cdn down")

	m = readStatus(t, root)
	f = m["feeds"].(map[string]any)["level1/ip"].(map[string]any)
	if f["last_error"] != "down" {
		t.Fatalf("feed last_error = %v", f["last_error"])
	}
	if f["last_success"] != feedSuccess {
		t.Fatalf("feed last_success not preserved across restart: got %v want %v", f["last_success"], feedSuccess)
	}
	if f["indicators"].(float64) != 42 {
		t.Fatalf("feed indicators not preserved across restart: got %v", f["indicators"])
	}
	if f["sha256"] != "abc123" {
		t.Fatalf("feed sha256 not preserved across restart: got %v", f["sha256"])
	}

	sig := m["signatures"].(map[string]any)
	if sig["last_error"] != "cdn down" {
		t.Fatalf("signature last_error = %v", sig["last_error"])
	}
	if sig["last_success"] != sigSuccess {
		t.Fatalf("signature last_success not preserved across restart: got %v want %v", sig["last_success"], sigSuccess)
	}
	if sig["databases"].(map[string]any)["daily.cvd"] != "27461" {
		t.Fatalf("signature databases not preserved across restart: got %v", sig["databases"])
	}
}

func TestRecorderWritesSignatureStatus(t *testing.T) {
	root := t.TempDir()
	r := NewRecorder(root)
	r.SigSuccess(map[string]string{"daily.cvd": "27461"})
	m := readStatus(t, root)
	sig := m["signatures"].(map[string]any)
	if sig["databases"].(map[string]any)["daily.cvd"] != "27461" {
		t.Fatalf("signatures = %v", sig)
	}
	r.SigFailure("cdn unreachable")
	m = readStatus(t, root)
	sig = m["signatures"].(map[string]any)
	if sig["last_error"] != "cdn unreachable" {
		t.Fatalf("last_error = %v", sig["last_error"])
	}
	if m["updated"] == nil {
		t.Fatal("updated must be stamped")
	}
}
