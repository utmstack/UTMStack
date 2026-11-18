package unpack

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func fsg2Fixture(t *testing.T, name string) []byte {
	t.Helper()
	p := filepath.Join("..", "..", "..", "docs", "edr", "native-engine", "audits", "2026-10-26-fsg2", "oracle", name)
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func TestFSG2ReferenceRebuild(t *testing.T) {
	for _, tc := range []struct{ input, reference, decoded string }{
		{"fsg2-harmless.exe", "reference-rebuilt.exe", "decoded.bin"},
		{"fsg2-mixed.exe", "reference-mixed-rebuilt.exe", "mixed-decoded.bin"},
		{"fsg2-zero-offset.exe", "reference-zero-offset-rebuilt.exe", "zero-offset-decoded.bin"},
	} {
		t.Run(tc.input, func(t *testing.T) {
			f := fsg2Fixture(t, tc.input)
			p := FSG2Program{File: f, ImageBase: 0x400000, EmptyRVA: 0x1000, EmptyVSize: 0x2000,
				NextRVA: 0x3000, NextRaw: 0x400, NextSize: 0x400, Entry: 0x400}
			got, child := FSG2(p)
			if got != FSG2Unpacked {
				t.Fatalf("outcome = %d", got)
			}
			want := fsg2Fixture(t, tc.reference)
			if !bytes.Equal(child, want) {
				t.Fatalf("rebuilt PE differs from ClamAV: got %d bytes, want %d bytes", len(child), len(want))
			}
			decoded := fsg2Fixture(t, tc.decoded)
			if !bytes.Equal(child[0x200:0x200+len(decoded)], decoded) {
				t.Fatal("decompressed bytes differ")
			}
		})
	}
}

func TestFSG2Bounds(t *testing.T) {
	f := fsg2Fixture(t, "fsg2-harmless.exe")
	p := FSG2Program{File: f, ImageBase: 0x400000, EmptyRVA: 0x1000, EmptyVSize: 0x2000,
		NextRVA: 0x3000, NextRaw: 0x400, NextSize: 0x400, Entry: 0x400}
	check := func(name string, edit func(*FSG2Program), want FSG2Outcome) {
		t.Helper()
		q := p
		q.File = bytes.Clone(f)
		edit(&q)
		got, child := FSG2(q)
		if got != want || child != nil {
			t.Fatalf("%s: outcome %d, child %d bytes; want %d and no child", name, got, len(child), want)
		}
	}
	check("bad pointer slot", func(q *FSG2Program) { q.File[0x402] = 0xff }, FSG2Failed)
	check("bad stack", func(q *FSG2Program) { q.File[0x500] = 0xff }, FSG2Failed)
	check("bad destination", func(q *FSG2Program) { q.File[0x520] = 0xff }, FSG2Failed)
	check("bad stream", func(q *FSG2Program) { q.File[0x524] = 0xff }, FSG2Unknown)
	check("bad helper", func(q *FSG2Program) { q.File[0x533] = 0x7f }, FSG2Failed)
	check("size mismatch", func(q *FSG2Program) { q.EmptyVSize = q.NextSize }, FSG2Returns)
	check("huge allocation", func(q *FSG2Program) { q.EmptyVSize = 65 << 20 }, FSG2Unknown)
	check("raw outside file", func(q *FSG2Program) { q.NextRaw = uint32(len(f)) }, FSG2Unknown)
}
