//go:build linux

package ransomware

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/utmstack/UTMStack/agent/edr/config"
	"github.com/utmstack/UTMStack/agent/edr/fanotify"
	"github.com/utmstack/UTMStack/agent/edr/proctable"
	"github.com/utmstack/UTMStack/agent/edr/responder"
	"golang.org/x/sys/unix"
)

func TestNativeRansomwareOwnedWriter(t *testing.T) {
	if path := os.Getenv("OPENEDR_RANSOMWARE_CHILD_PATH"); path != "" {
		if err := os.WriteFile(path, []byte("ordinary controlled canary edit"), 0600); err != nil {
			t.Fatal(err)
		}
		time.Sleep(30 * time.Second)
	}
}

type containedChannel struct{ events chan string }

func (c containedChannel) Append(s string) error {
	if strings.Contains(s, "ransomware_contained") {
		select {
		case c.events <- s:
		default:
		}
	}
	return nil
}
func TestNativeFanotifyCanaryWriteContainsOnlyOwnedChild(t *testing.T) {
	root := os.Getenv("OPENEDR_FANOTIFY_TEST_ROOT")
	if os.Getenv("OPENEDR_ISOLATED_VM") != "1" || root == "" {
		t.Skip("native ransomware fixture requires explicit isolated-VM fanotify opt-in")
	}
	if os.Geteuid() != 0 || !strings.HasPrefix(root, "/tmp/openedr-fanotify-") {
		t.Fatal("fixture must use isolated root tmpfs scope")
	}
	var fs unix.Statfs_t
	if err := unix.Statfs(root, &fs); err != nil || uint64(fs.Type) != uint64(unix.TMPFS_MAGIC) {
		t.Fatal("fixture requires dedicated tmpfs")
	}
	canaries := testCanaryManager(t, &fakeCanaryStore{})
	if n, err := canaries.Plant([]string{root}, 1); err != nil || n != 1 {
		t.Fatalf("native canary planting = %d, %v", n, err)
	}
	canary := filepath.Join(root, PlanCanaries([]string{root}, 1)[0].Name)
	feed, err := NewFanotifyFeed(fanotify.Config{Enabled: true, Mounts: []string{root}, Roots: []string{root}, Mode: fanotify.ModeNotify, Deadline: time.Second, Workers: 1, MaxPending: 8, MaxFileBytes: 1 << 20})
	if err != nil {
		t.Fatal(err)
	}
	table := proctable.New()
	events := make(chan string, 1)
	q := &fakeQuar{}
	cfg := config.Default()
	cfg.Ransomware.Enabled = true
	cfg.Ransomware.ResponseMode = "kill"
	g := NewGuard(GuardDeps{Cfg: cfg, Table: table, Resp: responder.New(table, nil, responder.OSKiller{}, responder.OSSuspender{}), Spool: containedChannel{events}, Quar: q, Incidents: &fakeIncidents{}, Canaries: canaries})
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- g.Run(ctx, feed, os.Getpid(), nil) }()
	defer func() {
		cancel()
		select {
		case err := <-done:
			if !errors.Is(err, context.Canceled) {
				t.Errorf("guard ended: %v", err)
			}
		case <-time.After(2 * time.Second):
			t.Error("guard failed shutdown")
		}
	}()
	deadline := time.Now().Add(2 * time.Second)
	for !g.FeedHealthy() {
		if time.Now().After(deadline) {
			t.Fatalf("feed unavailable: %v", feed.Health())
		}
		time.Sleep(5 * time.Millisecond)
	}
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	child := exec.Command(exe, "-test.run=^TestNativeRansomwareOwnedWriter$")
	child.Env = append(os.Environ(), "OPENEDR_RANSOMWARE_CHILD_PATH="+canary)
	if err = child.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = child.Process.Kill(); _ = child.Wait() }()
	select {
	case message := <-events:
		if !strings.Contains(message, "canary:") {
			t.Fatalf("wrong containment signal: %s", message)
		}
	case <-time.After(3 * time.Second):
		t.Fatalf("owned canary writer was not contained; feed=%v", feed.Health())
	}
	if err = child.Wait(); err == nil {
		t.Fatal("owned child was not terminated")
	}
	if q.calls != 0 {
		t.Fatal("native response quarantined a non-pinned executable pathname")
	}
}
