package main

import (
	"testing"

	"github.com/utmstack/UTMStack/agent/edr/config"
)

func TestApplySetBlocklist(t *testing.T) {
	c := config.Default()

	if err := applySet(&c, "blocklist.enforce", "false"); err != nil {
		t.Fatal(err)
	}
	if c.Blocklist.Enforce {
		t.Fatal("enforce should be false")
	}

	if err := applySet(&c, "blocklist.refresh_hours", "12"); err != nil {
		t.Fatal(err)
	}
	if c.Blocklist.RefreshHours != 12 {
		t.Fatalf("refresh = %d", c.Blocklist.RefreshHours)
	}

	if err := applySet(&c, "blocklist.direction", "out"); err != nil {
		t.Fatal(err)
	}
	if c.Blocklist.Direction != "out" {
		t.Fatalf("direction = %q", c.Blocklist.Direction)
	}

	if err := applySet(&c, "blocklist.direction", "sideways"); err == nil {
		t.Fatal("bad direction must error")
	}
}
