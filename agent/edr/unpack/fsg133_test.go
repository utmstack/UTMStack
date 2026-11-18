package unpack

import (
	"bytes"
	"testing"
)

func TestFSG133ReferenceRebuild(t *testing.T) {
	for _, tc := range []struct{ input, reference string }{
		{"fsg133-harmless.exe", "reference-133-rebuilt.exe"},
		{"fsg133-two-sections.exe", "reference-133-two-sections.exe"},
	} {
		t.Run(tc.input, func(t *testing.T) {
			f := fsg2Fixture(t, tc.input)
			p := FSG133Program{File: f, ImageBase: 0x400000, HeaderSize: 0x1000,
				EmptyRVA: 0x1000, EmptyVSize: 0x2000, NextRVA: 0x3000,
				NextRaw: 0x400, NextSize: 0x400, Entry: 0x400, EntryRVA: 0x3000}
			got, child := FSG133(p)
			if got != FSG2Unpacked {
				t.Fatalf("outcome = %d", got)
			}
			want := fsg2Fixture(t, tc.reference)
			if !bytes.Equal(child, want) {
				t.Fatalf("rebuilt PE differs from ClamAV: got %d bytes, want %d bytes", len(child), len(want))
			}
		})
	}
}
