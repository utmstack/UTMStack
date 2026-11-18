//go:build linux || darwin

package ransomware

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/utmstack/UTMStack/agent/edr/fanotify"
	"github.com/utmstack/UTMStack/agent/edr/proctable"
	"golang.org/x/sys/unix"
)

func TestCanaryPinnedIdentitySurvivesAliasButRejectsReplacement(t *testing.T) {
	dir := canaryTestDir(t)
	store := &fakeCanaryStore{}
	m := testCanaryManager(t, store)
	if n, err := m.Plant([]string{dir}, 1); n != 1 || err != nil {
		t.Fatalf("plant = %d, %v", n, err)
	}
	path := filepath.Join(dir, PlanCanaries([]string{dir}, 1)[0].Name)
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	id, ok := canaryFileID(info)
	if !ok || !m.ContainsFileIdentity(id.device, id.inode) {
		t.Fatal("planted decoy identity was not retained")
	}
	alias := filepath.Join(dir, "another-name.xlsx")
	if err := os.Link(path, alias); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(path, filepath.Join(dir, "renamed.xlsx")); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("ordinary replacement"), 0600); err != nil {
		t.Fatal(err)
	}
	replacement, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	otherID, ok := canaryFileID(replacement)
	if !ok || m.ContainsFileIdentity(otherID.device, otherID.inode) || !m.ContainsFileIdentity(id.device, id.inode) {
		t.Fatal("replacement inherited canary identity or alias lost it")
	}
	var got []FileEvent
	feed := &FanotifyFeed{sink: func(e FileEvent) { got = append(got, e) },
		interest: m.Contains, identityInterest: m.ContainsFileIdentity}
	feed.observe(fanotify.Outcome{Event: fanotify.Event{PID: 700, Path: alias,
		Device: id.device, Inode: id.inode, Mask: fanotify.Modify, OSIdentity: 123}})
	feed.observe(fanotify.Outcome{Event: fanotify.Event{PID: 700, Path: path,
		Device: otherID.device, Inode: otherID.inode, Mask: fanotify.Modify, OSIdentity: 123}})
	if len(got) != 2 || !got[0].FileIdentityRequired || !got[1].FileIdentityRequired {
		t.Fatalf("pinned event identity was not delivered: %+v", got)
	}
	g := &Guard{deps: GuardDeps{Canaries: m}}
	if !g.isCanaryEvent(got[0]) || g.isCanaryEvent(got[1]) {
		t.Fatal("guard accepted replacement or missed original through alias")
	}
	if err := m.LoadCovered([]string{dir}); err != nil {
		t.Fatal(err)
	}
	if m.Count() != 0 || m.ContainsFileIdentity(id.device, id.inode) {
		t.Fatal("reconciliation retained a decoy whose recorded path was replaced")
	}
}

func TestCanaryPinsSurviveUnlinkAndRetireOnRefreshOrClose(t *testing.T) {
	dir := canaryTestDir(t)
	m := testCanaryManager(t, &fakeCanaryStore{})
	if n, err := m.Plant([]string{dir}, 1); n != 1 || err != nil {
		t.Fatalf("plant = %d, %v", n, err)
	}
	path := filepath.Join(dir, PlanCanaries([]string{dir}, 1)[0].Name)
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	id, ok := canaryFileID(info)
	if !ok {
		t.Fatal("missing canary identity")
	}
	pin := m.ids[id]
	if pin == nil {
		t.Fatal("identity has no live descriptor")
	}
	flags, err := unix.FcntlInt(pin.Fd(), unix.F_GETFD, 0)
	if err != nil || flags&unix.FD_CLOEXEC == 0 {
		t.Fatalf("canary pin is not close-on-exec: %d, %v", flags, err)
	}
	if err := m.LoadCovered([]string{dir}); err != nil || m.ids[id] != pin {
		t.Fatalf("unchanged refresh replaced its live pin: %v", err)
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	if _, err := pin.Stat(); err != nil || !m.ContainsFileIdentity(id.device, id.inode) {
		t.Fatalf("unlinked decoy was not kept live: %v", err)
	}
	if err := m.LoadCovered([]string{dir}); err != nil {
		t.Fatal(err)
	}
	if _, err := pin.Stat(); !errors.Is(err, os.ErrClosed) || m.ContainsFileIdentity(id.device, id.inode) || m.Count() != 0 {
		t.Fatalf("refresh did not retire removed decoy: %v", err)
	}
	if n, err := m.Plant([]string{dir}, 1); n != 1 || err != nil {
		t.Fatalf("replant = %d, %v", n, err)
	}
	for _, p := range m.ids {
		pin = p
	}
	if err := m.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := pin.Stat(); !errors.Is(err, os.ErrClosed) || m.Count() != 0 {
		t.Fatalf("close left membership or a live descriptor: %v", err)
	}
	if err := m.Close(); err != nil {
		t.Fatal(err)
	}
	if err := m.Load(); !errors.Is(err, os.ErrClosed) {
		t.Fatalf("load after close: %v", err)
	}
	if _, err := m.Plant([]string{dir}, 1); !errors.Is(err, os.ErrClosed) {
		t.Fatalf("plant after close: %v", err)
	}
}

func TestCanaryMembershipLimitClosesRejectedPin(t *testing.T) {
	dir := canaryTestDir(t)
	path := filepath.Join(dir, "candidate")
	if err := os.WriteFile(path, []byte(canaryBody), 0600); err != nil {
		t.Fatal(err)
	}
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	id, pin, err := prepareCanaryPin(f)
	f.Close()
	if err != nil {
		t.Fatal(err)
	}
	m := testCanaryManager(t, nil)
	for i := 0; i < maxActiveCanaries; i++ {
		m.set[fmt.Sprintf("/occupied/%d", i)] = true
	}
	m.opmu.Lock()
	err = m.remember(path, id, pin)
	m.opmu.Unlock()
	if err == nil || m.ContainsFileIdentity(id.device, id.inode) || m.Count() != maxActiveCanaries {
		t.Fatalf("membership bound not enforced: %v", err)
	}
	if _, err := pin.Stat(); !errors.Is(err, os.ErrClosed) {
		t.Fatalf("rejected pin left open: %v", err)
	}
}

func TestGuardFeedKeepsPinnedAliasUnderExcludedPath(t *testing.T) {
	g, response, _, _, incidents, table := newTestGuard("kill", nil, 50, 100)
	g.deps.Resp = &boundResponse{fakeResp: response, current: 123}
	g.deps.Canaries = &fakeCanaries{
		paths: map[string]bool{"/decoy": true},
		ids:   map[canaryFileIdentity]bool{{device: 42, inode: 77}: true},
	}
	table.Add(proctable.Proc{PID: 700, Image: "/inert/writer", OSIdentity: 123, IdentityRequired: true})
	ready := make(chan func(FileEvent), 1)
	feed := admissionFeedFunc(func(ctx context.Context, sink func(FileEvent)) error {
		ready <- sink
		<-ctx.Done()
		return ctx.Err()
	})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- g.Run(ctx, feed, 999, func(string) bool { return true }) }()
	sink := <-ready
	sink(FileEvent{PID: 700, Path: "/decoy", Op: OpWrite, Device: 42, Inode: 78,
		FileIdentityRequired: true, OSIdentity: 123, IdentityRequired: true})
	if response.killCount() != 0 || len(incidents.recs) != 0 {
		t.Fatal("excluded replacement triggered response")
	}
	sink(FileEvent{PID: 700, Path: "/alias", Op: OpWrite, Device: 42, Inode: 77,
		FileIdentityRequired: true, OSIdentity: 123, IdentityRequired: true})
	if response.killCount() != 1 || len(incidents.recs) != 1 {
		t.Fatal("excluded alias of pinned canary was not contained")
	}
	cancel()
	if err := <-done; !errors.Is(err, context.Canceled) {
		t.Fatalf("guard exit: %v", err)
	}
}

type pathOnlyCanaries struct{}

func (pathOnlyCanaries) Contains(string) bool { return true }

func TestFanotifyGuardRefusesPathOnlyCanarySet(t *testing.T) {
	g, _, _, _, _, _ := newTestGuard("kill", nil, 50, 100)
	g.deps.Canaries = pathOnlyCanaries{}
	if err := g.Run(context.Background(), &FanotifyFeed{}, 999, nil); err == nil {
		t.Fatal("fanotify guard started without pinned canary identities")
	}
}

func TestGuardRespondsOnlyToPinnedCanaryFileIdentity(t *testing.T) {
	g, response, app, _, incidents, table := newTestGuard("kill", nil, 50, 100)
	g.deps.Resp = &boundResponse{fakeResp: response, current: 123}
	g.deps.Canaries = &fakeCanaries{
		paths: map[string]bool{"/decoy": true},
		ids:   map[canaryFileIdentity]bool{{device: 42, inode: 77}: true},
	}
	table.Add(proctable.Proc{PID: 700, Image: "/inert/writer", OSIdentity: 123, IdentityRequired: true})
	g.OnFileEvent(FileEvent{PID: 700, Path: "/decoy", Op: OpWrite,
		Device: 42, Inode: 78, FileIdentityRequired: true,
		OSIdentity: 123, IdentityRequired: true})
	if response.killCount() != 0 || len(app.lines) != 0 || len(incidents.recs) != 0 {
		t.Fatal("replacement file at canary path triggered response")
	}
	g.OnFileEvent(FileEvent{PID: 700, Path: "/alias", Op: OpWrite,
		Device: 42, Inode: 77, FileIdentityRequired: true,
		OSIdentity: 123, IdentityRequired: true})
	if response.killCount() != 1 || len(incidents.recs) != 1 {
		t.Fatal("pinned canary reached through alias failed response")
	}
}
