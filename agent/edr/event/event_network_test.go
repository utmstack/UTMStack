// edr/event/event_network_test.go
package event

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestNewNetworkEventBranding(t *testing.T) {
	e := NewNetworkEvent(ActionBlocked, "outbound", "1.2.3.4", 443, "", "1.2.3.4", nil)
	js, err := e.ToJSON()
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal([]byte(js), &m); err != nil {
		t.Fatal(err)
	}
	if m["source"] != "network_watcher" {
		t.Fatalf("source = %v", m["source"])
	}
	if m["engine"] != "UTMStack EDR" || m["product"] != "UTMStack EDR" {
		t.Fatalf("branding wrong: %v", m)
	}
	if m["remote_ip"] != "1.2.3.4" || m["direction"] != "outbound" {
		t.Fatalf("fields wrong: %v", m)
	}
	if strings.Contains(strings.ToLower(js), "threatwinds") ||
		strings.Contains(strings.ToLower(js), "clam") {
		t.Fatalf("event leaked a vendor string: %s", js)
	}
}
