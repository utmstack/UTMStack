package config

import (
	"encoding/json"
	"testing"
)

func TestLinuxProtectionScopeAndLimits(t *testing.T) {
	good := LinuxProtection{Fanotify: LinuxFanotify{Enabled: true, Mounts: []string{"/"}, Roots: []string{"/tmp/private"}, Mode: "permission_exec"}}
	if err := good.Validate(); err != nil {
		t.Fatal(err)
	}
	if good.Effective().NetworkEnforcement || good.Effective().Fanotify.FailClosed {
		t.Fatal("implicit policy escalation")
	}
	for _, mutate := range []func(*LinuxProtection){
		func(c *LinuxProtection) { c.Fanotify.Mounts = nil }, func(c *LinuxProtection) { c.Fanotify.Roots = nil },
		func(c *LinuxProtection) { c.Fanotify.Mounts = []string{"/home"} }, func(c *LinuxProtection) { c.Fanotify.Roots = []string{"/tmp/../etc"} },
		func(c *LinuxProtection) { c.Fanotify.DeadlineMS = 5001 }, func(c *LinuxProtection) { c.Fanotify.DeadlineMS = -1 },
		func(c *LinuxProtection) { c.Fanotify.Workers = 9 }, func(c *LinuxProtection) { c.Fanotify.MaxPending = 65 },
		func(c *LinuxProtection) { c.ProcessEvents = "fake" }, func(c *LinuxProtection) { c.Fanotify.Mode = "fake" },
	} {
		c := good
		mutate(&c)
		if err := c.Validate(); err == nil {
			t.Fatalf("invalid accepted %+v", c)
		}
	}
}
func TestLinuxProtectionStrictJSON(t *testing.T) {
	for _, raw := range []string{`{"process_event":"poll"}`, `{"fanotify":{"enable":true}}`} {
		var l LinuxProtection
		if json.Unmarshal([]byte(raw), &l) == nil {
			t.Fatal("unknown policy accepted")
		}
	}
	var l LinuxProtection
	if err := json.Unmarshal([]byte(`{"process_events":"poll","fanotify":{"enabled":false}}`), &l); err != nil {
		t.Fatal(err)
	}
	if l.Effective().ProcessEvents != "poll" || l.Fanotify.Enabled {
		t.Fatal(l)
	}
}
