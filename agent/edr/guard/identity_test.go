package guard

import (
	"errors"
	"github.com/utmstack/UTMStack/agent/edr/proctable"
	"testing"
)

type identityResp struct {
	fakeResp
	identity int64
	seen     []int64
}

func (r *identityResp) KillTreeIdentity(pid int, id int64, sig string) ([]int, error) {
	r.seen = append(r.seen, id)
	if r.identity != id {
		return nil, errors.New("PID reused")
	}
	return r.fakeResp.KillTree(pid, sig)
}
func (r *identityResp) SuspendIdentity(pid int, id int64) error {
	r.seen = append(r.seen, id)
	if r.identity != id {
		return errors.New("PID reused")
	}
	return r.fakeResp.Suspend(pid)
}
func (r *identityResp) ResumeIdentity(pid int, id int64) error {
	r.seen = append(r.seen, id)
	if r.identity != id {
		return errors.New("PID reused")
	}
	return r.fakeResp.Resume(pid)
}
func TestGuardUsesObservedIdentityForEveryResponse(t *testing.T) {
	for _, malicious := range []bool{false, true} {
		r := &identityResp{identity: 123}
		verdict := "clean"
		if malicious {
			verdict = "malicious"
		}
		g := New(suspendCfg(), fakeScanner{verdict: verdict}, r)
		g.OnStart(proctable.Proc{PID: 42, Image: "/opt/inert", OSIdentity: 123})
		for _, id := range r.seen {
			if id != 123 {
				t.Fatal(r.seen)
			}
		}
		if len(r.seen) != 2 {
			t.Fatalf("missing identity-bound responses: %v", r.seen)
		}
	}
	r := &identityResp{identity: 124}
	g := New(suspendCfg(), fakeScanner{verdict: "malicious"}, r)
	g.OnStart(proctable.Proc{PID: 42, Image: "/opt/inert", OSIdentity: 123})
	if r.killed != 0 || r.suspended != 0 || r.resumed != 0 {
		t.Fatalf("acted on reused PID: %+v", r)
	}
	legacy := &fakeResp{}
	g = New(suspendCfg(), fakeScanner{verdict: "malicious"}, legacy)
	g.OnStart(proctable.Proc{PID: 42, Image: "/opt/inert", OSIdentity: 123})
	if legacy.killed != 0 || legacy.suspended != 0 {
		t.Fatal("fell back to unsafe numeric response")
	}
}

func TestLinuxSystemSuspensionRailHasPathBoundaries(t *testing.T) {
	for _, p := range []string{"/usr/bin/sh", "/sbin/init", "/usr/lib/systemd/systemd", "/bin/../bin/sh"} {
		if !linuxSystemImage(p) {
			t.Fatal(p)
		}
	}
	for _, p := range []string{"/usr/binary/payload", "/home/test/usr/bin/tool", "/USR/bin/tool", "/opt/tool"} {
		if linuxSystemImage(p) {
			t.Fatal(p)
		}
	}
}

func TestGuardNeverRespondsToUnidentifiedNativeObservation(t *testing.T) {
	r := &identityResp{identity: 0}
	g := New(suspendCfg(), fakeScanner{verdict: "malicious"}, r)
	g.OnStart(proctable.Proc{PID: 42, Image: "/opt/inert", IdentityRequired: true})
	if r.killed != 0 || r.suspended != 0 || r.resumed != 0 || len(r.seen) != 0 {
		t.Fatalf("unidentified process response: %+v", r)
	}
}

// The service must release its own suspension when containment cannot complete.
func TestGuardReleasesOwnedSuspensionAfterIdentityResponseFailure(t *testing.T) {
	r := &failingIdentityKill{identityResp: identityResp{identity: 123}}
	g := New(suspendCfg(), fakeScanner{verdict: "malicious"}, r)
	g.OnStart(proctable.Proc{PID: 42, Image: "/opt/inert", OSIdentity: 123, IdentityRequired: true})
	if r.suspended != 42 || r.resumed != 42 || r.killed != 0 {
		t.Fatalf("owned suspension leaked: %+v", r)
	}
}

type failingIdentityKill struct{ identityResp }

func (*failingIdentityKill) KillTreeIdentity(int, int64, string) ([]int, error) {
	return nil, errors.New("native snapshot unavailable")
}
