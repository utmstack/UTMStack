package platform

import (
	"errors"
	"strings"
	"testing"
)

// ING-F6: a journal entry much larger than bufio.Scanner's old 4096-byte
// initial buffer (and its 64KB hard ceiling) must still be delivered whole,
// and a long line must not stop the lines after it from being delivered too.
func TestReadJournalLinesDeliversEntriesLargerThanScannerDefaults(t *testing.T) {
	long := strings.Repeat("A", 200*1024) // 200KB, well past bufio.MaxScanTokenSize (64KB)
	input := "short line\n" + long + "\n" + "line after the long one\n"

	var got []string
	if err := readJournalLines(strings.NewReader(input), func(line string) {
		got = append(got, line)
	}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []string{"short line", long, "line after the long one"}
	if len(got) != len(want) {
		t.Fatalf("got %d lines, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("line %d: got len %d, want len %d", i, len(got[i]), len(want[i]))
		}
	}
}

// A trailing line with no final newline (the stream ends mid-write, or
// journalctl is killed) must still be delivered, not silently dropped.
func TestReadJournalLinesDeliversFinalLineWithoutTrailingNewline(t *testing.T) {
	var got []string
	err := readJournalLines(strings.NewReader("one\ntwo"), func(line string) {
		got = append(got, line)
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 || got[0] != "one" || got[1] != "two" {
		t.Fatalf("got %v, want [one two]", got)
	}
}

// A real read error (not a clean EOF) must be surfaced to the caller.
func TestReadJournalLinesPropagatesReadErrors(t *testing.T) {
	boom := errors.New("boom")
	r := &errAfterReader{after: "partial", err: boom}

	var got []string
	err := readJournalLines(r, func(line string) { got = append(got, line) })
	if !errors.Is(err, boom) {
		t.Fatalf("err = %v, want %v", err, boom)
	}
}

type errAfterReader struct {
	after string
	err   error
	sent  bool
}

func (r *errAfterReader) Read(p []byte) (int, error) {
	if !r.sent {
		r.sent = true
		n := copy(p, r.after)
		return n, nil
	}
	return 0, r.err
}
