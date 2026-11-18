package engine

import (
	"bytes"
	"testing"
)

// Office and the reference engine ignore bytes after the last complete
// sector, such as a newline appended by a transfer tool. A sector the file
// ends inside is read zero-filled, as the reference reads it.
func TestOLETrailingPartialSectorIsIgnored(t *testing.T) {
	// Compressed, so only extraction of the stream can reveal the marker.
	payload := gzipFixture(t, []byte("prefix "+oleTestMarker))
	for _, extra := range []int{1, 100, 511} {
		b := append(oleFixture(t, 3, "Contents", payload, true, false, false), bytes.Repeat([]byte{'\n'}, extra)...)
		if sig, err := oleTestState(t).visit(b, 0); sig != "OLE.Test" || err != nil {
			t.Fatalf("%d trailing bytes: sig=%q err=%v", extra, sig, err)
		}
	}
	b := oleFixture(t, 3, "Contents", payload, true, false, false)
	if sig, err := oleTestState(t).visit(b[:len(b)-100], 0); sig != "OLE.Test" || err != nil {
		t.Fatalf("truncated final sector: sig=%q err=%v", sig, err)
	}
}
