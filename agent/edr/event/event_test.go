package event

import (
	"encoding/json"
	"runtime"
	"strings"
	"testing"
)

func TestDetectionJSONIsBrandedAndClean(t *testing.T) {
	e := NewDetection(`C:\Users\x\a.exe`, "abc123", "Win.Test.EICAR", SourceEngine)
	js, err := e.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON: %v", err)
	}
	if !strings.Contains(js, `"product":"UTMStack EDR"`) {
		t.Fatalf("missing product branding: %s", js)
	}
	if !strings.Contains(js, `"engine":"UTMStack EDR"`) {
		t.Fatalf("missing engine branding: %s", js)
	}
	low := strings.ToLower(js)
	if strings.Contains(low, "clamav") || strings.Contains(low, "clamd") {
		t.Fatalf("event leaked engine name: %s", js)
	}
}

func TestEventOSDefaultsToRuntime(t *testing.T) {
	e := NewDetection(`/tmp/x`, "abc123", "Win.Test.EICAR", SourceFileWatcher)
	if e.OS != "" {
		t.Fatalf("fresh event must have empty OS, got %q", e.OS)
	}
	js, err := e.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON: %v", err)
	}
	var doc map[string]string
	if err := json.Unmarshal([]byte(js), &doc); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if doc["os"] != runtime.GOOS {
		t.Fatalf("os field = %q, want %q (the runtime OS, not a hardcode)", doc["os"], runtime.GOOS)
	}
}
