package engine

import (
	"archive/tar"
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"hash/crc32"
	"runtime"
	"testing"
	"time"
)

// measureAllocation reports the bytes allocated while fn runs.
func measureAllocation(fn func()) uint64 {
	runtime.GC()
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)
	fn()
	runtime.ReadMemStats(&after)
	return after.TotalAlloc - before.TotalAlloc
}

// Bytes decompressed by an expansion that then fails (over the limit, or with
// trailing or malformed input) are charged to the aggregate expansion budget.
// Otherwise every stream gets a fresh full-size budget and total
// decompression work is not bounded by MaxScanSize.
func TestFailedExpansionsAreChargedToScanBudget(t *testing.T) {
	const maxFile, budget, streams = 1 << 20, 4 << 20, 60
	e := nativeTestEngine(t)
	g := e.current.Load()
	bomb := zlibFixture(t, make([]byte, maxFile+16))
	gzBomb := gzipFixture(t, make([]byte, maxFile+16))
	trailing := append(zlibFixture(t, make([]byte, maxFile/2)), "junk"...)

	pdfOf := func(stream []byte, filter string) []byte {
		var pdf bytes.Buffer
		pdf.WriteString("%PDF-1.7\n")
		for i := 1; i <= streams; i++ {
			fmt.Fprintf(&pdf, "%d 0 obj\n<< /Length %d /Filter %s >>\nstream\n", i, len(stream), filter)
			pdf.Write(stream)
			pdf.WriteString("\nendstream\nendobj\n")
		}
		pdf.WriteString("%%EOF\n")
		return pdf.Bytes()
	}
	// Two input bytes repeat one byte 128 times: the stream fills the
	// per-object limit, then lacks its end-of-data marker.
	runLength := bytes.Repeat([]byte{129, 'A'}, maxFile/128)

	// A compressed Flash header declares a small movie; its body inflates
	// past the per-object limit.
	swfBomb := append([]byte{'C', 'W', 'S', 10, 100, 0, 0, 0}, bomb...)
	tarOf := func(member []byte) []byte {
		var tarball bytes.Buffer
		tw := tar.NewWriter(&tarball)
		for i := 0; i < streams; i++ {
			if err := tw.WriteHeader(&tar.Header{Name: fmt.Sprintf("m%02d", i), Mode: 0600, Size: int64(len(member)), Format: tar.FormatUSTAR}); err != nil {
				t.Fatal(err)
			}
			if _, err := tw.Write(member); err != nil {
				t.Fatal(err)
			}
		}
		if err := tw.Close(); err != nil {
			t.Fatal(err)
		}
		return tarball.Bytes()
	}

	// The requested MSZIP member, not unused folder padding, exceeds the
	// per-object limit. Streaming may stop before unreferenced folder bytes.
	cabBomb := cabFixture(t, 1, cabChunks(make([]byte, maxFile+32768)), []cabFixtureFile{{"a", 0, maxFile + 32768}}, false)

	for name, data := range map[string][]byte{
		"tar-of-cab-over-limit": tarOf(cabBomb),
		"pdf-flate-over-limit":  pdfOf(bomb, "/FlateDecode"),
		"pdf-flate-trailing":    pdfOf(trailing, "/FlateDecode"),
		// in a chain the reference extracts (it does not extract RunLength alone)
		"pdf-runlength-no-eod":    pdfOf(runLength, "[/RunLengthDecode /FlateDecode]"),
		"pdf-chain-over-limit":    pdfOf(bomb, "[/FlateDecode /ASCIIHexDecode]"),
		"tar-of-gzip-over-limit":  tarOf(gzBomb),
		"tar-of-flash-over-limit": tarOf(swfBomb),
	} {
		t.Run(name, func(t *testing.T) {
			s := scanState{ctx: context.Background(), db: g.db, maxFile: maxFile, remaining: budget, maxFiles: 10000, maxDepth: 10}
			var sig string
			var err error
			allocated := measureAllocation(func() { sig, err = s.visit(data, 0) })
			t.Logf("input=%d bytes sig=%q err=%v allocated=%d MiB remaining=%d", len(data), sig, err, allocated>>20, s.remaining)
			if sig != "" || err == nil {
				t.Fatalf("expected incomplete, got sig=%q err=%v", sig, err)
			}
			// The aggregate budget is 4 MiB; allow generous slack for buffer
			// growth and decoder state. 60 uncharged expansions exceed it.
			if allocated > 16*budget {
				t.Fatalf("scan allocated %d MiB against a %d MiB aggregate expansion budget: failed expansions are not charged", allocated>>20, budget>>20)
			}
		})
	}
}

// RAR members are read completely before they are analyzed. A member whose
// read fails, here on its checksum after all of its bytes were produced, is
// charged like any other failed expansion.
func TestFailedRARMemberReadIsCharged(t *testing.T) {
	e := nativeTestEngine(t)
	member := bytes.Repeat([]byte("R"), 64<<10)
	rar := rarStoredMembers(member)
	crc := bytes.Index(rar, binary.LittleEndian.AppendUint32(nil, crc32.ChecksumIEEE(member)))
	if crc < 0 {
		t.Fatal("fixture: member checksum not found")
	}
	// Corrupt the file checksum, then reseal the file header (whose own
	// checksum covers it) so that only the member read fails.
	rar[crc] ^= 0xff
	block := crc - 16
	size := int(binary.LittleEndian.Uint16(rar[block+5:]))
	binary.LittleEndian.PutUint16(rar[block:], uint16(crc32.ChecksumIEEE(rar[block+2:block+size])))
	s := scanState{ctx: context.Background(), db: e.current.Load().db, maxFile: 1 << 20, remaining: 4 << 20, maxFiles: 100, maxDepth: 8}
	sig, err := s.visit(rar, 0)
	if sig != "" || err == nil {
		t.Fatalf("bad member checksum: sig=%q err=%v", sig, err)
	}
	if charged := int64(4<<20) - s.remaining; charged < int64(len(rar)+len(member)) {
		t.Fatalf("charged %d bytes for a %d-byte archive and a %d-byte failed member read", charged, len(rar), len(member))
	}
}

// The LZFu RawSize header is attacker-controlled. The decoder does not
// allocate RawSize bytes before checking that the compressed input can
// produce it (LZFu expands by at most 8x), so 16-byte properties cannot each
// force a scan-limit-sized allocation.
func TestRTFRawSizeIsValidatedBeforeAllocation(t *testing.T) {
	const maxFile, props = 16 << 20, 500
	e := nativeTestEngine(t)
	g := e.current.Load()
	value := make([]byte, 16)
	binary.LittleEndian.PutUint32(value[0:], 12)         // COMPSIZE: header only
	binary.LittleEndian.PutUint32(value[4:], maxFile)    // RAWSIZE: at the limit
	binary.LittleEndian.PutUint32(value[8:], 0x75465a4c) // LZFu
	list := make([][]byte, props)
	for i := range list {
		list[i] = tnefProp(0x0102, 0x1009, value)
	}
	data := tnefFixture(tnefAttr(1, 0x00069003, tnefPropList(list...)))

	s := scanState{ctx: context.Background(), db: g.db, maxFile: maxFile, remaining: 2 * maxFile, maxFiles: 10000, maxDepth: 10}
	var sig string
	var err error
	start := time.Now()
	allocated := measureAllocation(func() { sig, err = s.visit(data, 0) })
	t.Logf("input=%d bytes sig=%q err=%v allocated=%d MiB elapsed=%v", len(data), sig, err, allocated>>20, time.Since(start))
	if sig != "" || err == nil {
		t.Fatalf("expected incomplete, got sig=%q err=%v", sig, err)
	}
	if allocated > 64<<20 {
		t.Fatalf("%d-byte TNEF caused %d MiB of allocation: RawSize is allocated before it is validated against the compressed input", len(data), allocated>>20)
	}

	// An ordinary LZFu stream within the expansion bound still decodes.
	raw := []byte(testMarker)
	compressed := rtfLiteralFixture(raw)
	if got, err := s.decompressRTF(compressed); err != nil || string(got) != testMarker {
		t.Fatalf("literal LZFu stream: %q %v", got, err)
	}

	// Groups of eight maximal 17-byte references to the preloaded dictionary
	// expand eightfold; only the end reference's three bytes produce nothing.
	const groups = 4
	var contents []byte
	for group := 0; group < groups; group++ {
		contents = append(contents, 0xff)
		for i := 0; i < 8; i++ {
			contents = binary.BigEndian.AppendUint16(contents, 15) // offset 0, length 17
		}
	}
	contents = append(contents, 1)
	contents = binary.BigEndian.AppendUint16(contents, uint16(207+groups*136)<<4)
	header := make([]byte, 16)
	binary.LittleEndian.PutUint32(header[:4], uint32(len(contents)+12))
	binary.LittleEndian.PutUint32(header[4:8], groups*136)
	binary.LittleEndian.PutUint32(header[8:12], 0x75465a4c)
	binary.LittleEndian.PutUint32(header[12:16], rtfCRC(contents))
	got, err := s.decompressRTF(append(header, contents...))
	if err != nil || !bytes.Equal(got, bytes.Repeat([]byte(rtfPrebuffer[:17]), groups*8)) {
		t.Fatalf("maximal LZFu expansion: %d bytes, %v", len(got), err)
	}
	binary.LittleEndian.PutUint32(header[4:8], 8*uint32(len(contents))+1)
	if _, err := s.decompressRTF(append(header, contents...)); !errors.Is(err, ErrMalformed) {
		t.Fatalf("RawSize above the LZFu bound: %v", err)
	}
}
