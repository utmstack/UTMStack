package event

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSpoolAppendAndRotate(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "events.ndjson")
	s, err := OpenSpool(path, 40) // tiny cap to force rotation
	if err != nil {
		t.Fatalf("OpenSpool: %v", err)
	}
	defer s.Close()
	for i := 0; i < 5; i++ {
		_, err = DrainSpool(context.Background(), path, filepath.Join(dir, ".offset"), func(string) bool { return false })
		if err != nil {
			t.Fatal(err)
		}
		if err := s.Append(`{"n":` + strings.Repeat("0", 10) + `}`); i == 4 {
			if !errors.Is(err, ErrSpoolFull) {
				t.Fatalf("full buffer: %v", err)
			}
		} else if err != nil {
			t.Fatalf("Append: %v", err)
		}
	}
	// Current spool file must exist and stay under a bounded size.
	fi, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if fi.Size() > 200 {
		t.Fatalf("spool not rotating, size=%d", fi.Size())
	}
	var delivered int
	_, err = DrainSpool(context.Background(), path, filepath.Join(dir, ".offset"), func(string) bool { delivered++; return true })
	if err != nil || delivered != 4 {
		t.Fatalf("retained records=%d err=%v", delivered, err)
	}
	if stats := s.Snapshot(); stats.Accepted != 4 || stats.FullRejections != 1 || !stats.Degraded {
		t.Fatal(stats)
	}
}
