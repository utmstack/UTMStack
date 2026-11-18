package signatures

import (
	"testing"
	"unsafe"
)

func TestCompactOffsetPreservesBoundsAndRange(t *testing.T) {
	if got := unsafe.Sizeof(offset{}); got != 16 {
		t.Fatalf("compiled offset occupies %d bytes, want 16", got)
	}
	meta := &peMeta{valid: true, entry: 40, sections: []section{{raw: 20, size: 10}}}
	tests := []struct {
		text       string
		low, high  int
		applicable bool
	}{
		{"*", 0, 100, true},
		{"12", 12, 12, true},
		{"12,5", 12, 17, true},
		{"EOF-10", 90, 90, true},
		{"EP-5", 35, 35, true},
		{"SE0", 20, 29, true},
		{"S0+3", 23, 23, true},
		{"SL+3", 23, 23, true},
		{"2147483647,2147483647", 0, 0, false},
		{"EOF-2147483647", 0, 0, false},
		{"EP+2147483647", 0, 0, false},
		{"EP-2147483647", 0, 0, false},
		{"SE2147483647", 0, 0, false},
	}
	for _, tt := range tests {
		t.Run(tt.text, func(t *testing.T) {
			o, err := parseOffset(tt.text)
			if err != nil {
				t.Fatal(err)
			}
			lo, hi, ok := o.bounds(100, meta)
			if ok != tt.applicable || ok && (lo != tt.low || hi != tt.high) {
				t.Fatalf("bounds = (%d,%d,%t), want (%d,%d,%t)", lo, hi, ok, tt.low, tt.high, tt.applicable)
			}
		})
	}
	if _, err := parseOffset("2147483648"); err == nil {
		t.Fatal("offset above signed 32-bit range was accepted")
	}
}
