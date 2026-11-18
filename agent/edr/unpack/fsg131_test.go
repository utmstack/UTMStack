package unpack

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func fsg131Fixture(t *testing.T, name string) []byte {
	t.Helper()
	p := filepath.Join("..", "..", "..", "docs", "edr", "native-engine", "audits", "2026-10-27-fsg131", "oracle", name)
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func TestFSG131ReferenceRebuild(t *testing.T) {
	for _, name := range []string{"fsg131-harmless", "fsg131-two-sections", "fsg131-call-stub"} {
		t.Run(name, func(t *testing.T) {
			f := fsg131Fixture(t, name+".exe")
			p := FSG131Program{File: f, ImageBase: 0x400000, HeaderSize: 0x1000,
				EmptyRVA: 0x1000, EmptyVSize: 0x5000, NextRVA: 0x3000,
				NextRaw: 0x400, NextSize: 0x4000, Entry: 0x3400, EntryRVA: 0x6000}
			got, child := FSG131(p)
			if got != FSG2Unpacked {
				t.Fatalf("outcome = %d", got)
			}
			want := fsg131Fixture(t, name+"-reference-rebuilt.exe")
			if !bytes.Equal(child, want) {
				t.Fatalf("rebuilt PE differs from ClamAV: got %d bytes, want %d bytes", len(child), len(want))
			}
		})
	}
}

func TestFSG131TruncatedSectionTable(t *testing.T) {
	f := fsg131Fixture(t, "fsg131-harmless.exe")
	p := FSG131Program{File: f, ImageBase: 0x400000, HeaderSize: 0x1000,
		EmptyRVA: 0x1000, EmptyVSize: 0x5000, NextRVA: 0x3000,
		NextRaw: 0x400, NextSize: 0x4000, Entry: 0x3400, EntryRVA: 0x6000}
	p.File = p.File[:0x200]
	if got, _ := FSG131(p); got != FSG2Unknown {
		t.Fatalf("truncated PE outcome = %d", got)
	}
}
