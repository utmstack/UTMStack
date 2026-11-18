package scheduled

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/utmstack/UTMStack/agent/edr/config"
)

func fixture(t *testing.T) string {
	t.Helper()
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"a.txt", "sub/b.txt", "excluded/c.txt"} {
		path := filepath.Join(root, name)
		if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte("inert safe fixture"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func TestEnumerationHonorsExclusionsAndSkipsSymbolicLinks(t *testing.T) {
	root := fixture(t)
	if err := os.Symlink(root, filepath.Join(root, "cycle")); err != nil {
		t.Logf("symlink unavailable: %v", err)
	}
	var got []string
	s, err := New(config.ScheduledScan{Paths: []string{root}}, func(_ context.Context, path string) error { got = append(got, path); return nil }, func(path string) bool { return path == filepath.Join(root, "excluded") })
	if err != nil {
		t.Fatal(err)
	}
	if err = s.Sweep(context.Background()); err != nil {
		t.Fatal(err)
	}
	sort.Strings(got)
	want := []string{filepath.Join(root, "a.txt"), filepath.Join(root, "sub", "b.txt")}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("queued %v want %v", got, want)
	}
	st := s.Status()
	if st.Running || st.Enqueued != 2 || st.Error != "" || st.Started.IsZero() || st.Finished.IsZero() {
		t.Fatalf("status %+v", st)
	}
}

func TestEnumerationLimitsAreIncompleteAndBounded(t *testing.T) {
	for _, tc := range []struct {
		name string
		cfg  config.ScheduledScan
		max  int
	}{
		{"files", config.ScheduledScan{MaxFiles: 1}, 1},
		{"entries", config.ScheduledScan{MaxEntries: 1}, 0},
		{"depth", config.ScheduledScan{MaxDepth: 1}, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := fixture(t)
			cfg := tc.cfg
			cfg.Paths = []string{root}
			count := 0
			s, _ := New(cfg, func(context.Context, string) error { count++; return nil }, nil)
			err := s.Sweep(context.Background())
			if !errors.Is(err, ErrBudget) || count > tc.max || s.Status().Error == "" {
				t.Fatalf("err=%v count=%d status=%+v", err, count, s.Status())
			}
		})
	}
}

func TestBackpressureCancelsAndSweepsNeverOverlap(t *testing.T) {
	root := fixture(t)
	entered := make(chan struct{})
	done := make(chan error, 1)
	s, _ := New(config.ScheduledScan{Paths: []string{root}}, func(ctx context.Context, _ string) error { close(entered); <-ctx.Done(); return ctx.Err() }, nil)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() { done <- s.Sweep(ctx) }()
	select {
	case <-entered:
	case <-time.After(time.Second):
		t.Fatal("sweep did not reach queue")
	}
	if err := s.Sweep(ctx); !errors.Is(err, ErrRunning) {
		t.Fatalf("overlap allowed: %v", err)
	}
	cancel()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("blocked after cancellation")
	}
	if s.Status().Running {
		t.Fatal("sweep remained running")
	}
}

func TestSchedulerStartsOnlyWhenRequestedAndReportsReadErrors(t *testing.T) {
	root := fixture(t)
	for _, enabled := range []bool{false, true} {
		ctx, cancel := context.WithCancel(context.Background())
		queued := make(chan string, 4)
		s, _ := New(config.ScheduledScan{Enabled: enabled, RunOnStart: true, Paths: []string{root}}, func(_ context.Context, p string) error { queued <- p; return nil }, nil)
		done := make(chan struct{})
		go func() { s.Run(ctx); close(done) }()
		if enabled {
			select {
			case <-queued:
			case <-time.After(time.Second):
				t.Fatal("no initial sweep")
			}
		} else {
			select {
			case <-done:
			case <-time.After(time.Second):
				t.Fatal("disabled scheduler running")
			}
			if len(queued) != 0 {
				t.Fatal("disabled enqueued")
			}
		}
		cancel()
		select {
		case <-done:
		case <-time.After(time.Second):
			t.Fatal("scheduler did not stop")
		}
	}
	s, _ := New(config.ScheduledScan{Paths: []string{filepath.Join(root, "absent"), filepath.Join(root, "a.txt")}}, func(context.Context, string) error { return nil }, nil)
	if err := s.Sweep(context.Background()); !errors.Is(err, os.ErrNotExist) || s.Status().Enqueued != 1 {
		t.Fatalf("read failure hidden or prevented remaining paths: %v %+v", err, s.Status())
	}
}

func TestDepthBoundDoesNotStarveOtherRoots(t *testing.T) {
	root := fixture(t)
	count := 0
	s, _ := New(config.ScheduledScan{Paths: []string{root, filepath.Join(root, "a.txt")}, MaxDepth: 1}, func(context.Context, string) error { count++; return nil }, nil)
	if err := s.Sweep(context.Background()); !errors.Is(err, ErrBudget) || count != 2 {
		t.Fatalf("depth starved next root: %v count=%d", err, count)
	}
}

func TestConfiguredSymbolicLinkAncestorsAreRejected(t *testing.T) {
	root := fixture(t)
	link := filepath.Join(filepath.Dir(root), filepath.Base(root)+"-link")
	if err := os.Symlink(root, link); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	defer os.Remove(link)
	s, _ := New(config.ScheduledScan{Paths: []string{filepath.Join(link, "a.txt")}}, func(context.Context, string) error { t.Fatal("followed symbolic-link ancestor"); return nil }, nil)
	if err := s.Sweep(context.Background()); err == nil {
		t.Fatal("symbolic-link ancestor unreported")
	}
}
