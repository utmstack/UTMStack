package agent

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDrainSpoolEmitsNewLinesAndAdvances(t *testing.T) {
	dir := t.TempDir()
	spool := filepath.Join(dir, "events.ndjson")
	offset := filepath.Join(dir, ".offset")
	if err := os.WriteFile(spool, []byte("line-a\nline-b\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	var got []string
	off, err := drainSpool(spool, offset, func(raw string) { got = append(got, raw) })
	if err != nil {
		t.Fatalf("drainSpool: %v", err)
	}
	if len(got) != 2 || got[0] != "line-a" || got[1] != "line-b" {
		t.Fatalf("emitted = %v", got)
	}

	// Append one more line; a second drain must emit only the new line.
	f, _ := os.OpenFile(spool, os.O_APPEND|os.O_WRONLY, 0o600)
	f.WriteString("line-c\n")
	f.Close()

	got = nil
	if _, err := drainSpool(spool, offset, func(raw string) { got = append(got, raw) }); err != nil {
		t.Fatalf("drainSpool 2: %v", err)
	}
	if len(got) != 1 || got[0] != "line-c" {
		t.Fatalf("second drain emitted = %v (offset start %d)", got, off)
	}
}
