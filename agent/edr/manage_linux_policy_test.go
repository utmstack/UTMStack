package main

import (
	"github.com/utmstack/UTMStack/agent/edr/config"
	"reflect"
	"testing"
)

func TestLinuxPolicyCLIIsAtomic(t *testing.T) {
	c := config.Default()
	for _, edit := range [][2]string{{"linux_protection.fanotify.mounts", `["/"]`}, {"linux_protection.fanotify.roots", `["/tmp/lab"]`}, {"linux_protection.fanotify.enabled", "true"}, {"linux_protection.fanotify.mode", "permission_exec"}, {"linux_protection.process_events", "connector"}} {
		if err := applySet(&c, edit[0], edit[1]); err != nil {
			t.Fatal(err)
		}
	}
	before := c.LinuxProtection
	for _, edit := range [][2]string{{"linux_protection.fanotify.workers", "99"}, {"linux_protection.fanotify.roots", `["relative"]`}, {"linux_protection.process_events", "fake"}, {"linux_protection.fanotify.unknown", "true"}} {
		if applySet(&c, edit[0], edit[1]) == nil {
			t.Fatal("invalid accepted")
		}
		if !reflect.DeepEqual(c.LinuxProtection, before) {
			t.Fatal("failed edit mutated policy")
		}
	}
}
