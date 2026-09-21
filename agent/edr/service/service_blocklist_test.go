package service

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestStatusDocHasBlocklist(t *testing.T) {
	// statusDoc must marshal blocklist fields; build one with values set.
	d := statusDoc{}
	d.BlocklistEnabled = true
	d.BlocklistEnforce = true
	d.BlocklistIndicators = 42
	b, err := json.Marshal(d)
	if err != nil {
		t.Fatal(err)
	}
	s := string(b)
	for _, k := range []string{"blocklist_enabled", "blocklist_enforce", "blocklist_indicators"} {
		if !strings.Contains(s, k) {
			t.Fatalf("status doc missing %q: %s", k, s)
		}
	}
	if strings.Contains(strings.ToLower(s), "threatwinds") {
		t.Fatal("status doc must not name the intel vendor")
	}
}
