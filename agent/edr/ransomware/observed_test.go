package ransomware

import (
	"errors"
	"github.com/utmstack/UTMStack/agent/edr/fanotify"
	"github.com/utmstack/UTMStack/agent/edr/proctable"
	"strings"
	"testing"
	"time"
)

type boundResponse struct {
	*fakeResp
	seen    []int64
	current int64
}

func (b *boundResponse) SuspendIdentity(pid int, id int64) error {
	b.seen = append(b.seen, id)
	if id != b.current {
		return errors.New("reused")
	}
	return b.fakeResp.Suspend(pid)
}
func (b *boundResponse) ResumeIdentity(pid int, id int64) error {
	b.seen = append(b.seen, id)
	return b.fakeResp.Resume(pid)
}
func (b *boundResponse) KillTreeIdentity(pid int, id int64, sig string) ([]int, error) {
	b.seen = append(b.seen, id)
	if id != b.current {
		return nil, errors.New("reused")
	}
	return b.fakeResp.KillTree(pid, sig)
}
func TestNativeCanaryResponseNeverRebindsPIDOrQuarantinesReplacementImage(t *testing.T) {
	for _, current := range []int64{123, 124} {
		g, legacy, _, q, inc, tab := newTestGuard("kill", map[string]bool{"/scope/canary": true}, 50, 100)
		r := &boundResponse{fakeResp: legacy, current: current}
		g.deps.Resp = r
		tab.Add(proctable.Proc{PID: 700, Image: "/benign/replacement", OSIdentity: 124, IdentityRequired: true})
		g.deps.LiveProcess = func(int) (string, int, error) { t.Fatal("native event rebound through LiveProcess"); return "", 0, nil }
		g.OnFileEvent(FileEvent{PID: 700, Path: "/scope/canary", Op: OpWrite, OSIdentity: 123, IdentityRequired: true, Image: "/original/image"})
		if len(r.seen) != 1 || r.seen[0] != 123 || q.calls != 0 {
			t.Fatalf("unsafe native response: ids=%v q=%d", r.seen, q.calls)
		}
		if current == 124 && legacy.killCount() != 0 {
			t.Fatal("replacement PID killed")
		}
		if len(inc.recs) != 1 || inc.recs[0].Image != "/original/image" {
			t.Fatal("incident metadata rebound")
		}
		// Buffered observations of the original process remain deduplicated,
		// even when the table already contains a different PID occupant.
		g.OnFileEvent(FileEvent{PID: 700, Path: "/scope/canary", Op: OpWrite, OSIdentity: 123, IdentityRequired: true, Image: "/original/image"})
		if len(r.seen) != 1 || len(inc.recs) != 1 {
			t.Fatal("trailing original observation retried containment")
		}
	}
}
func TestMissingNativeIdentityCannotUseLegacyResponse(t *testing.T) {
	g, r, app, q, inc, _ := newTestGuard("kill", map[string]bool{"/canary": true}, 50, 100)
	g.OnFileEvent(FileEvent{PID: 700, Path: "/canary", Op: OpWrite, IdentityRequired: true})
	if r.killCount() != 0 || q.calls != 0 {
		t.Fatal("unidentified event acted on PID")
	}
	if len(app.lines) != 1 || len(inc.recs) != 1 || inc.recs[0].Action != "identity_unavailable" {
		t.Fatal("unattributed canary modification was silently lost")
	}
}
func TestFanotifyFeedRequiresActualWriteAndIdentity(t *testing.T) {
	got := []FileEvent{}
	f := &FanotifyFeed{sink: func(e FileEvent) { got = append(got, e) }}
	f.observe(fanotify.Outcome{Event: fanotify.Event{PID: 700, Path: "/canary", Mask: fanotify.CloseWrite, OSIdentity: 123}})
	f.observe(fanotify.Outcome{Event: fanotify.Event{PID: 700, Path: "/canary", Mask: fanotify.Modify}})
	f.observe(fanotify.Outcome{Event: fanotify.Event{PID: 700, Path: "/canary", Mask: fanotify.Modify, OSIdentity: 123, Image: "/image"}})
	if len(got) != 2 || got[0].OSIdentity != 0 || !got[0].IdentityRequired || !got[0].FileIdentityRequired || got[1].OSIdentity != 123 || !got[1].IdentityRequired || !got[1].FileIdentityRequired || got[1].Op != OpWrite {
		t.Fatal(got)
	}
}
func TestLinuxRecoveryCommandsUseToolAndVerbBoundaries(t *testing.T) {
	positives := map[string][2]string{"linux_zfs_destroy": {"/usr/sbin/zfs", "zfs destroy -r tank/home@snap"}, "linux_btrfs_subvolume_delete": {"/usr/bin/btrfs", "btrfs subvolume delete /snapshots/old"}, "linux_backup_service_stop": {"/usr/bin/systemctl", "systemctl stop borgmatic.service"}}
	for want, parts := range positives {
		if got, ok := MatchT1490(parts[0], parts[1], nil); !ok || got != want {
			t.Errorf("%v -> %s %v", parts, got, ok)
		}
	}
	for _, parts := range [][2]string{{"/usr/bin/echo", "echo zfs destroy tank/a@snap"}, {"/usr/sbin/zfs", "zfs destroy -nv tank/home@snap"}, {"/usr/sbin/zfs", "zfs list snapshots"}, {"/usr/bin/btrfs", "btrfs subvolume list /"}, {"/usr/bin/systemctl", "systemctl status restic.service"}, {"/usr/bin/systemctl", "systemctl stop unrelated.service"}, {"/usr/bin/systemctl", "systemctl stop not-restic.service"}} {
		if name, ok := MatchT1490(parts[0], parts[1], nil); ok {
			t.Errorf("benign command matched %q: %v", name, parts)
		}
	}
	if _, ok := MatchT1490("/usr/bin/systemctl", "systemctl stop borgmatic.service", []string{"stop borgmatic.service"}); ok {
		t.Fatal("sanctioned command allowlist ignored")
	}
}
func TestETWCreationConstraintRejectsReplacementAfterEvent(t *testing.T) {
	event := time.Date(2026, 10, 21, 0, 0, 0, 0, time.UTC)
	id := event.Unix()*10000000 + 116444736000000000
	if !creationPredatesEvent(id, event) || creationPredatesEvent(id+1, event) || creationPredatesEvent(id, time.Time{}) {
		t.Fatal("event creation boundary unsafe")
	}
}

func TestCurrentCommandSnapshotKeepsIdentityAndExplicitEvidenceSemantics(t *testing.T) {
	g, legacy, app, q, inc, table := newTestGuard("kill", nil, 50, 100)
	r := &boundResponse{fakeResp: legacy, current: 80}
	g.deps.Resp = r
	table.Add(proctable.Proc{PID: 700, Image: "/owned/original-parent", OSIdentity: 80, IdentityRequired: true})
	g.deps.LiveProcess = func(int) (string, int, error) {
		t.Fatal("current observation rebound through numeric PID lookup")
		return "", 0, nil
	}
	g.OnObservedSnapshot(proctable.Proc{PID: 701, PPID: 700, Image: "/usr/sbin/zfs", Cmdline: "zfs destroy tank/home@snapshot", OSIdentity: 100, IdentityRequired: true})
	if legacy.killCount() != 1 || len(r.seen) != 1 || r.seen[0] != 80 || q.calls != 0 {
		t.Fatalf("snapshot response lost captured parent identity: %v", r.seen)
	}
	if len(inc.recs) != 1 || len(app.lines) != 1 || !strings.Contains(app.lines[0], "current_command_observation:linux_zfs_destroy") {
		t.Fatal("current command mislabeled as a historical exec", app.lines)
	}
}

func TestUnidentifiedCurrentSnapshotCannotUseLegacyResponse(t *testing.T) {
	g, r, app, _, _, table := newTestGuard("kill", nil, 50, 100)
	table.Add(proctable.Proc{PID: 700, Image: "/parent", OSIdentity: 80, IdentityRequired: true})
	g.OnObservedSnapshot(proctable.Proc{PID: 701, PPID: 700, Image: "/usr/sbin/zfs", Cmdline: "zfs destroy tank/home@snapshot"})
	if r.killCount() != 0 || len(app.lines) != 0 {
		t.Fatal("unidentified current observation acquired response authority")
	}
}
