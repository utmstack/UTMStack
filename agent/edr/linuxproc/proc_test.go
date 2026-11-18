package linuxproc

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func statFixture(pid, parent int, start int64, name string) string {
	return fmt.Sprintf("%d (%s) S %d %s%d 0 0\n", pid, name, parent, strings.Repeat("0 ", 17), start)
}

func TestStatFieldsAndHostileCommandNames(t *testing.T) {
	if p, err := ParseStat([]byte(statFixture(1, 0, 0, "init"))); err != nil || p.StartTicks != 0 {
		t.Fatalf("namespace init tick zero: %+v %v", p, err)
	}
	if p, err := ParseStat([]byte(statFixture(2, 0, 0, "kthreadd"))); err != nil || p.StartTicks != 0 {
		t.Fatalf("early kernel zero tick: %+v %v", p, err)
	}
	for _, name := range []string{"ordinary", "space in name", "closing ) ( name", "line\nbreak"} {
		p, err := ParseStat([]byte(statFixture(42, 7, 123456, name)))
		if err != nil || p.PID != 42 || p.PPID != 7 || p.StartTicks != 123456 || p.State != 'S' {
			t.Fatalf("%q: %+v %v", name, p, err)
		}
	}
	for _, data := range []string{"", "42 missing", "42 (cut) S 1", statFixture(-1, 7, 1, "x"), statFixture(42, -1, 1, "x"), statFixture(42, 1, -1, "x"), strings.Repeat("x", MaxStatBytes+1)} {
		if _, err := ParseStat([]byte(data)); err == nil {
			t.Fatalf("accepted invalid stat %q", data[:min(100, len(data))])
		}
	}
}

func fixtureProcess(t *testing.T, root string, pid, parent int, start int64) string {
	t.Helper()
	dir := filepath.Join(root, fmt.Sprint(pid))
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "stat"), []byte(statFixture(pid, parent, start, "safe process")), 0600); err != nil {
		t.Fatal(err)
	}
	// The synthetic proc tree is portable; its executable link must also be
	// absolute on the host running the fixture, including a Windows volume.
	image := filepath.Join(root, "inert-fixture")
	if err := os.Symlink(image, filepath.Join(dir, "exe")); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "cmdline"), []byte("fixture\x00--inert\x00"), 0600); err != nil {
		t.Fatal(err)
	}
	return image
}

func TestReadSnapshotAndBounds(t *testing.T) {
	root := t.TempDir()
	image := fixtureProcess(t, root, 42, 7, 123456)
	r := Reader{Root: root, MaxProcesses: 2}
	p, err := r.Read(42, true)
	if err != nil || p.Image != image || p.Cmdline != "fixture --inert" {
		t.Fatalf("%+v %v", p, err)
	}
	all, err := r.Snapshot(context.Background(), true)
	if err != nil || len(all) != 1 || all[0].StartTicks != 123456 {
		t.Fatalf("%+v %v", all, err)
	}
	fixtureProcess(t, root, 43, 7, 123457)
	fixtureProcess(t, root, 44, 7, 123458)
	if _, err := r.Snapshot(context.Background(), false); !errors.Is(err, ErrLimit) {
		t.Fatalf("process bound ignored: %v", err)
	}
	if _, err := r.Read(0, true); err == nil {
		t.Fatal("PID zero accepted")
	}
	if _, err := r.Read(999, true); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("gone process: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := r.Snapshot(ctx, false); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
}

func TestReadRejectsMismatchedPIDAndOversizedCommand(t *testing.T) {
	root := t.TempDir()
	fixtureProcess(t, root, 42, 7, 123456)
	p := filepath.Join(root, "42")
	if err := os.WriteFile(filepath.Join(p, "cmdline"), []byte(strings.Repeat("x", MaxCommandBytes+1)), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := (Reader{Root: root}).Read(42, true); !errors.Is(err, ErrLimit) {
		t.Fatalf("unbounded command accepted: %v", err)
	}
	if err := os.WriteFile(filepath.Join(p, "stat"), []byte(statFixture(43, 7, 123456, "x")), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := (Reader{Root: root}).Read(42, false); err == nil {
		t.Fatal("stat from another PID accepted")
	}
}

func TestSnapshotAggregateMetadataBudget(t *testing.T) {
	root := t.TempDir()
	fixtureProcess(t, root, 42, 1, 123)
	if _, err := (Reader{Root: root, MaxSnapshotBytes: 128}).Snapshot(context.Background(), true); !errors.Is(err, ErrLimit) {
		t.Fatalf("aggregate metadata unbounded: %v", err)
	}
	if _, err := (Reader{Root: root, MaxSnapshotBytes: DefaultMaxSnapshotBytes + 1}).Snapshot(context.Background(), true); !errors.Is(err, ErrLimit) {
		t.Fatalf("oversized budget accepted: %v", err)
	}
}
