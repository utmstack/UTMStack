package orchestrator

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// Snapshot completion is deliberately held apart from worker completion. A
// close-write notification in that window must cover the new bytes too.
type rewriteScanner struct {
	seen    chan string
	release chan struct{}
}

func (s *rewriteScanner) ScanFile(path, source string) (string, string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", "", err
	}
	s.seen <- string(data)
	if string(data) == "before" {
		<-s.release
	}
	return "clean", "", nil
}

func TestRewriteDuringAnalysisGetsFollowUp(t *testing.T) {
	path := filepath.Join(t.TempDir(), "changed")
	if err := os.WriteFile(path, []byte("before"), 0600); err != nil {
		t.Fatal(err)
	}
	sc := &rewriteScanner{make(chan string, 4), make(chan struct{})}
	o := New(sc, nil, 1, 1)
	ctx, cancel := context.WithCancel(context.Background())
	defer func() { cancel(); o.Stop() }()
	o.Start(ctx)
	if !o.Enqueue(FileEvent{Path: path}) {
		t.Fatal("first notification rejected")
	}
	select {
	case got := <-sc.seen:
		if got != "before" {
			t.Fatal(got)
		}
	case <-time.After(time.Second):
		close(sc.release)
		t.Fatal("initial scan missing")
	}
	if err := os.WriteFile(path, []byte("after"), 0600); err != nil {
		close(sc.release)
		t.Fatal(err)
	}
	accepted := o.Enqueue(FileEvent{Path: path})
	close(sc.release)
	if !accepted {
		t.Fatal("rewrite notification rejected")
	}
	select {
	case got := <-sc.seen:
		if got != "after" {
			t.Fatalf("follow-up bytes=%q", got)
		}
	case <-time.After(time.Second):
		t.Fatal("accepted rewrite was lost while first snapshot was in flight")
	}
}
