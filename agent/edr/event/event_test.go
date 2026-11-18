package event

import (
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
