package orchestrator

import (
	"context"
	"testing"
	"time"

	"github.com/utmstack/UTMStack/agent/edr/event"
)

type sourceScanner struct{ seen chan string }

func (s sourceScanner) ScanFile(path, source string) (string, string, error) {
	s.seen <- source
	return "clean", "", nil
}
func TestScheduledWorkRetainsSourceAndDefaultWatcherSource(t *testing.T) {
	sc := sourceScanner{make(chan string, 2)}
	o := New(sc, nil, 1, 2)
	ctx, cancel := context.WithCancel(context.Background())
	defer func() { cancel(); o.Stop() }()
	o.Start(ctx)
	for _, tc := range []struct{ path, source, want string }{{"scheduled", event.SourceScheduledScan, event.SourceScheduledScan}, {"watcher", "", event.SourceFileWatcher}} {
		if !o.Enqueue(FileEvent{Path: tc.path, Source: tc.source}) {
			t.Fatal("queue rejected")
		}
		select {
		case got := <-sc.seen:
			if got != tc.want {
				t.Fatalf("source=%s want%s", got, tc.want)
			}
		case <-time.After(time.Second):
			t.Fatal("not scanned")
		}
	}
}
