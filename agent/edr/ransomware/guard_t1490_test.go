//go:build windows || linux

package ransomware

import (
	"strings"
	"testing"
	"time"

	"github.com/utmstack/UTMStack/agent/edr/config"
	"github.com/utmstack/UTMStack/agent/edr/proctable"
)

// t1490Scenario returns a platform-appropriate trusted image, T1490 command
// image + cmdline, and trust predicate, so the portable guard tests exercise
// the T1490 trigger on both Windows (vssadmin) and Linux (btrfs subvolume
// delete). It is defined per platform in guard_t1490_windows_test.go and
// guard_t1490_linux_test.go; this file only calls it.

// TestGuard_T1490_SuspendMode_KillsParentAndQuarantines verifies the full
// escalation: a T1490 command attributed to its parent (the encryptor) scores
// to the kill threshold, suspends the culprit (suspend mode), kills the tree,
// quarantines the image, and records a contained incident.
func TestGuard_T1490_SuspendMode_KillsParentAndQuarantines(t *testing.T) {
	g, resp, app, quar, inc, tab := newTestGuard("suspend", nil, 50, 100)
	_, t1490Image, t1490Cmd, _ := t1490Scenario()
	// Encryptor (pid 100) spawns the T1490 command (pid 200).
	tab.Add(proctable.Proc{PID: 100, PPID: 4, Image: `C:\Users\x\enc.exe`})
	tab.Add(proctable.Proc{PID: 200, PPID: 100, Image: t1490Image})

	g.OnProcStart(200, 100, t1490Image, t1490Cmd, 1)

	if len(resp.suspended) == 0 || resp.suspended[0] != 100 {
		t.Fatalf("suspend-mode must suspend the culprit (parent 100): %v", resp.suspended)
	}
	if len(resp.killed) == 0 || resp.killed[0] != 100 {
		t.Fatalf("must kill the encryptor parent (100), got %v", resp.killed)
	}
	if quar.calls != 1 {
		t.Fatalf("must quarantine the encryptor image once, got %d", quar.calls)
	}
	if len(inc.recs) != 1 || inc.recs[0].Action != "contained" {
		t.Fatalf("incident not recorded: %+v", inc.recs)
	}
	// A ransomware_contained event was emitted.
	found := false
	for _, l := range app.lines {
		if contains(l, "ransomware_contained") {
			found = true
		}
	}
	if !found {
		t.Fatalf("no ransomware_contained event: %v", app.lines)
	}
}

// TestGuard_TrustedProcess_NotScored verifies that a trusted (allowlisted)
// process that legitimately runs a T1490 command AND touches a canary is never
// contained, while an untrusted encryptor doing the same IS contained.
func TestGuard_TrustedProcess_NotScored(t *testing.T) {
	tab := proctable.New()
	resp := &fakeResp{}
	cfg := config.Default()
	cfg.Ransomware.Enabled = true
	cfg.Ransomware.ResponseMode = "kill"
	cfg.Ransomware.SuspendThreshold = 50
	cfg.Ransomware.KillThreshold = 100
	trustedImage, t1490Image, t1490Cmd, trust := t1490Scenario()
	g := NewGuard(GuardDeps{
		Cfg: cfg, Table: tab, Resp: resp, Spool: &fakeAppender{}, Quar: &fakeQuar{}, Incidents: &fakeIncidents{},
		Canaries: &fakeCanaries{paths: map[string]bool{`C:\decoy\_00.xlsx`: true}},
		Hash:     func(string) (string, error) { return "x", nil },
		Now:      func() time.Time { return time.Unix(1000, 0) },
		NewID:    func() string { return "i" },
		Trusted:  trust,
	})
	// Trusted backup product (pid 300) runs the T1490 command AND touches a
	// canary — both max-weight signals — yet must not be contained.
	tab.Add(proctable.Proc{PID: 300, PPID: 4, Image: trustedImage})
	tab.Add(proctable.Proc{PID: 301, PPID: 300, Image: t1490Image})
	g.OnProcStart(301, 300, t1490Image, t1490Cmd, 1)
	g.OnFileEvent(FileEvent{PID: 300, Path: `C:\decoy\_00.xlsx`, Op: OpWrite})
	if len(resp.killed) != 0 || len(resp.suspended) != 0 {
		t.Fatalf("trusted process must not be contained: killed=%v suspended=%v", resp.killed, resp.suspended)
	}

	// Control: an UN-trusted encryptor touching a canary IS contained.
	resp2 := &fakeResp{}
	g2 := NewGuard(GuardDeps{
		Cfg: cfg, Table: tab, Resp: resp2, Spool: &fakeAppender{}, Quar: &fakeQuar{}, Incidents: &fakeIncidents{},
		Canaries: &fakeCanaries{paths: map[string]bool{`C:\decoy\_01.xlsx`: true}},
		Hash:     func(string) (string, error) { return "x", nil },
		Now:      func() time.Time { return time.Unix(1000, 0) },
		NewID:    func() string { return "i" }, Trusted: trust,
	})
	tab.Add(proctable.Proc{PID: 400, PPID: 4, Image: `C:\Users\x\enc.exe`})
	g2.OnFileEvent(FileEvent{PID: 400, Path: `C:\decoy\_01.xlsx`, Op: OpWrite})
	if len(resp2.killed) == 0 {
		t.Fatal("untrusted encryptor tampering a canary should be contained")
	}
}

// TestGuard_SetPolicy_HotAppliesCommandAllowlistAndMode verifies that a
// command allowlist entry hot-applied via SetPolicy exempts a T1490 command,
// and that switching to alert mode surfaces the escalation without acting.
func TestGuard_SetPolicy_HotAppliesCommandAllowlistAndMode(t *testing.T) {
	g, resp, _, _, _, tab := newTestGuard("kill", nil, 50, 100)
	_, t1490Image, t1490Cmd, _ := t1490Scenario()
	tab.Add(proctable.Proc{PID: 100, PPID: 4, Image: `C:\Users\x\enc.exe`})
	tab.Add(proctable.Proc{PID: 200, PPID: 100, Image: t1490Image})

	// Hot-apply an allowlist entry that exempts this exact command → no containment.
	g.SetPolicy("kill", []string{allowlistToken(t1490Cmd)})
	g.OnProcStart(200, 100, t1490Image, t1490Cmd, 1)
	if len(resp.killed) != 0 {
		t.Fatalf("allowlisted command (post-SetPolicy) must not be contained: killed=%v", resp.killed)
	}

	// Clear the allowlist and switch to alert mode → escalation surfaces but no kill.
	g.SetPolicy("alert", nil)
	tab.Add(proctable.Proc{PID: 300, PPID: 4, Image: `C:\Users\x\enc2.exe`})
	tab.Add(proctable.Proc{PID: 301, PPID: 300, Image: t1490Image})
	g.OnProcStart(301, 300, t1490Image, t1490Cmd, 2)
	if len(resp.killed) != 0 || len(resp.suspended) != 0 {
		t.Fatalf("alert mode (post-SetPolicy) must not suspend/kill: killed=%v suspended=%v", resp.killed, resp.suspended)
	}
}

// allowlistToken returns a distinctive substring of the T1490 command that the
// command allowlist can match (so the exemption is real, not a no-op).
func allowlistToken(cmdline string) string {
	fields := strings.Fields(cmdline)
	if len(fields) > 1 {
		return fields[1]
	}
	return fields[0]
}
