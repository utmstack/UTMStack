package sevenzip

import (
	"bytes"
	"errors"
	"hash/crc32"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
)

// The archives in testdata were made with 7-Zip 25.01
// (docs/2026-10-15-sevenzip/mkfixtures.sh; unsupported_first.7z, an ARM
// Thumb member then an LZMA2 one, with docs/2026-10-15-sevenzip/mkorder.py);
// their members hold the marker.
const testMarker = "EXTRACTED-HARMLESS-PAYLOAD-MARKER"

var testLimits = Limits{MaxFiles: 100, MaxUnpack: 64 << 20}

func readFixture(t testing.TB, name string) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("testdata", name+".7z"))
	if err != nil {
		t.Fatal(err)
	}
	return b
}

// unpackAll decodes every member, checking its CRC; it returns the members
// by name.
func unpackAll(a *Archive) (map[string][]byte, error) {
	members := map[string][]byte{}
	for _, f := range a.Files {
		if !f.HasStream {
			members[f.Name] = nil
			continue
		}
		out, err := a.Unpack(f.Folder)
		if err != nil {
			return nil, err
		}
		if f.Offset+f.Size > uint64(len(out)) {
			return nil, ErrData
		}
		m := out[f.Offset : f.Offset+f.Size]
		if f.HasCRC && crc32.ChecksumIEEE(m) != f.CRC {
			return nil, errors.New("member CRC mismatch")
		}
		members[f.Name] = m
	}
	return members, nil
}

func TestDecodesEachMethod(t *testing.T) {
	for name, member := range map[string]string{
		"copy": "text.txt", "lzma": "text.txt", "lzma2": "text.txt", "header_plain": "text.txt",
		"ppmd": "text.txt", "ppmd_restart": "big.txt", "bcj_lzma": "payload.bin", "bcj2": "payload.bin",
		"arm": "payload.bin", "solid": "text.txt", "empty_first": "text.txt",
	} {
		t.Run(name, func(t *testing.T) {
			a, err := Open(readFixture(t, name), testLimits)
			if err != nil {
				t.Fatal(err)
			}
			members, err := unpackAll(a)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Contains(members[member], []byte(testMarker)) {
				t.Fatalf("%s lacks the marker (members %d)", member, len(members))
			}
		})
	}
}

func TestSolidAndEmptyMembers(t *testing.T) {
	a, err := Open(readFixture(t, "solid"), testLimits)
	if err != nil {
		t.Fatal(err)
	}
	if len(a.Folders) != 1 || len(a.Files) != 2 || a.Files[0].Folder != 0 || a.Files[1].Folder != 0 ||
		a.Files[1].Offset != a.Files[0].Size {
		t.Fatalf("solid layout: %d folders, files %+v", len(a.Folders), a.Files)
	}
	a, err = Open(readFixture(t, "empty_first"), testLimits)
	if err != nil {
		t.Fatal(err)
	}
	if len(a.Files) != 2 || a.Files[0].Name != "empty.txt" || a.Files[0].HasStream || !a.Files[1].HasStream {
		t.Fatalf("empty member: %+v", a.Files)
	}
}

// The reference decodes the x86 and ARM branch filters only.
func TestUndecodedCoders(t *testing.T) {
	for name, want := range map[string]error{"bzip2": ErrUnsupported, "delta": ErrUnsupported, "enc_data": ErrEncrypted,
		"armt": ErrUnsupported, "arm64": ErrUnsupported, "ppc": ErrUnsupported, "sparc": ErrUnsupported,
		"ia64": ErrUnsupported, "riscv": ErrUnsupported} {
		a, err := Open(readFixture(t, name), testLimits)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if len(a.Files) != 1 {
			t.Fatalf("%s: files %+v", name, a.Files)
		}
		if _, err := a.Unpack(0); !errors.Is(err, want) {
			t.Errorf("%s: %v, want %v", name, err, want)
		}
	}
	if _, err := Open(readFixture(t, "enc_header"), testLimits); !errors.Is(err, ErrEncrypted) {
		t.Errorf("encrypted header: %v", err)
	}
}

// TestPPMdModelRestarts checks that the small-memory archive makes the model
// run out of memory and start again (and compact its free memory first),
// so the decoded bytes cover those paths.
func TestPPMdModelRestarts(t *testing.T) {
	a, err := Open(readFixture(t, "ppmd_restart"), testLimits)
	if err != nil {
		t.Fatal(err)
	}
	f := &a.Folders[0]
	src, err := a.packStream(f.packStart)
	if err != nil {
		t.Fatal(err)
	}
	props := f.Coders[0].Props
	p := newPPMd7(uint32(props[0]), uint32(props[1])|uint32(props[2])<<8|uint32(props[3])<<16|uint32(props[4])<<24)
	if !p.rcInit(src) {
		t.Fatal("range decoder start")
	}
	restarts, glued := 0, 0
	out := make([]byte, 0, f.UnpackSizes[0])
	for uint64(len(out)) < f.UnpackSizes[0] {
		text, glue := p.text, p.glueCount
		sym := p.decodeSymbol()
		if sym < 0 {
			t.Fatalf("symbol %d at %d", sym, len(out))
		}
		out = append(out, byte(sym))
		if p.text == p.alignOffset && text > p.alignOffset+1 {
			restarts++
		}
		if p.glueCount == 255 && glue != 255 {
			glued++
		}
	}
	if restarts == 0 || glued == 0 || !bytes.Contains(out, []byte(testMarker)) {
		t.Fatalf("restarts %d, glued %d, marker %v", restarts, glued, bytes.Contains(out, []byte(testMarker)))
	}
	if crc32.ChecksumIEEE(out) != a.Files[0].CRC {
		t.Fatal("CRC mismatch")
	}
}

func TestPPMdModelLimit(t *testing.T) {
	if _, err := decodePPMd([]byte{6, 0, 0, 0, 0x20}, []byte{0, 0, 0, 0, 0}, 10, nil); !errors.Is(err, ErrLimit) {
		t.Errorf("512 MiB model: %v", err)
	}
	for _, props := range [][]byte{{1, 0, 0, 1, 0}, {65, 0, 0, 1, 0}, {6, 0, 4, 0, 0}, {6, 0, 0, 1}} {
		if _, err := decodePPMd(props, []byte{0, 0, 0, 0, 0}, 10, nil); !errors.Is(err, ErrData) {
			t.Errorf("props % x: %v", props, err)
		}
	}
}

func TestReferenceHeaderRules(t *testing.T) {
	good := readFixture(t, "lzma")
	edit := func(f func(b []byte) []byte) []byte { return f(append([]byte(nil), good...)) }
	// The start header's CRC and the minor version are not checked.
	for name, b := range map[string][]byte{
		"start CRC":     edit(func(b []byte) []byte { b[8] ^= 1; return b }),
		"minor version": edit(func(b []byte) []byte { b[7] = 9; return b }),
		"trailing data": edit(func(b []byte) []byte { return append(b, "trailing"...) }),
	} {
		if a, err := Open(b, testLimits); err != nil {
			t.Errorf("%s: %v", name, err)
		} else if _, err := unpackAll(a); err != nil {
			t.Errorf("%s: %v", name, err)
		}
	}
	// A major version, a header out of the file or a header CRC mismatch
	// is an archive the reference reads nothing of.
	for name, b := range map[string][]byte{
		"major version": edit(func(b []byte) []byte { b[6] = 1; return b }),
		"truncated":     good[:len(good)-10],
		"header CRC":    edit(func(b []byte) []byte { b[28] ^= 1; return b }),
	} {
		if _, err := Open(b, testLimits); !errors.Is(err, ErrNothing) {
			t.Errorf("%s: %v, want nothing read", name, err)
		}
	}
	// Damaged data fails the member.
	b := edit(func(b []byte) []byte { b[40] ^= 0x55; return b })
	a, err := Open(b, testLimits)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := unpackAll(a); err == nil {
		t.Error("damaged data decoded")
	}
}

func TestLimits(t *testing.T) {
	a, err := Open(readFixture(t, "lzma"), Limits{MaxFiles: 100, MaxUnpack: 100})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := a.Unpack(0); !errors.Is(err, ErrLimit) {
		t.Errorf("output beyond the limit: %v", err)
	}
	if _, err := Open(readFixture(t, "solid"), Limits{MaxFiles: 1, MaxUnpack: 1 << 20}); !errors.Is(err, ErrLimit) {
		t.Errorf("files beyond the limit: %v", err)
	}
	done := make(chan struct{})
	close(done)
	a, err = Open(readFixture(t, "ppmd_restart"), Limits{MaxFiles: 100, MaxUnpack: 1 << 20, Done: done})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := a.Unpack(0); !errors.Is(err, ErrAborted) {
		t.Errorf("stopped decoding: %v", err)
	}
}

// TestDamagedArchives runs every fixture with bytes changed and cut: the
// reader must return, without panics, within its limits.
func TestDamagedArchives(t *testing.T) {
	names, _ := filepath.Glob(filepath.Join("testdata", "*.7z"))
	r := rand.New(rand.NewSource(7))
	for _, n := range names {
		good, err := os.ReadFile(n)
		if err != nil {
			t.Fatal(err)
		}
		tries := 300
		if testing.Short() {
			tries = 30
		}
		for i := 0; i < tries; i++ {
			b := append([]byte(nil), good...)
			switch i % 3 {
			case 0:
				b[r.Intn(len(b))] ^= byte(1 + r.Intn(255))
			case 1:
				b = b[:r.Intn(len(b))]
			default:
				for k := 0; k < 4; k++ {
					b[32+r.Intn(len(b)-32)] = byte(r.Intn(256))
				}
			}
			if a, err := Open(b, testLimits); err == nil {
				unpackAll(a)
			}
		}
	}
}

func FuzzOpen(f *testing.F) {
	names, _ := filepath.Glob(filepath.Join("testdata", "*.7z"))
	for _, n := range names {
		b, _ := os.ReadFile(n)
		f.Add(b)
	}
	f.Fuzz(func(t *testing.T, b []byte) {
		if a, err := Open(b, Limits{MaxFiles: 100, MaxUnpack: 1 << 20}); err == nil {
			unpackAll(a)
		}
	})
}
