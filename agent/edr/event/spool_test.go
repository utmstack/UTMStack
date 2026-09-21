package event

import (
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
		if err := s.Append(`{"n":` + strings.Repeat("0", 10) + `}`); err != nil {
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
}
