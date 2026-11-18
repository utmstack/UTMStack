package engine

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"hash/crc32"
	"math/rand"
	"testing"

	"github.com/klauspost/compress/zstd"
	"github.com/ulikunitz/xz"
)

// archive/zip reports an entry as a directory from its external attributes as
// well as from a trailing slash. Extractors that decide by name still write
// such a member's data, so it must be analyzed instead of being skipped.
func TestZIPAttributeDirectoryMemberWithDataIsScanned(t *testing.T) {
	e := nativeTestEngine(t)
	build := func(name string, creator uint16, attrs uint32, data []byte) []byte {
		var b bytes.Buffer
		w := zip.NewWriter(&b)
		f, err := w.CreateHeader(&zip.FileHeader{Name: name, Method: zip.Deflate, CreatorVersion: creator << 8, ExternalAttrs: attrs})
		if err != nil {
			t.Fatal(err)
		}
		if _, err := f.Write(data); err != nil {
			t.Fatal(err)
		}
		if err := w.Close(); err != nil {
			t.Fatal(err)
		}
		if bytes.Contains(b.Bytes(), []byte(testMarker)) {
			t.Fatal("fixture exposes the marker without decompression")
		}
		return b.Bytes()
	}
	payload := []byte(testMarker)
	if clean, sig, err := e.Scan(context.Background(), build("payload.bin", 0, 0x20, payload)); clean || sig != "Native.Test" || err != nil {
		t.Fatalf("control: clean=%v sig=%q err=%v", clean, sig, err)
	}
	for name, data := range map[string][]byte{
		"msdos-directory-bit": build("payload.bin", 0, 0x10, payload),
		"unix-s-ifdir":        build("payload.bin", 3, 0o040755<<16, payload),
	} {
		r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
		if err != nil || len(r.File) != 1 || !r.File[0].FileInfo().IsDir() {
			t.Fatalf("%s fixture: err=%v", name, err)
		}
		if clean, sig, err := e.Scan(context.Background(), data); clean || sig != "Native.Test" || err != nil {
			t.Errorf("%s: member with data was not analyzed: clean=%v sig=%q err=%v", name, clean, sig, err)
		}
	}
	// A name ending in '/' that still declares data cannot be extracted as a
	// directory; it must not be skipped as one.
	named := bytes.ReplaceAll(build("payloadX", 0, 0x20, payload), []byte("payloadX"), []byte("payload/"))
	if clean, sig, err := e.Scan(context.Background(), named); clean || err == nil && sig == "" {
		t.Fatalf("slash-named member with data reported clean: sig=%q err=%v", sig, err)
	}
	// Genuine empty directory entries stay skipped.
	var b bytes.Buffer
	w := zip.NewWriter(&b)
	if _, err := w.Create("folder/"); err != nil {
		t.Fatal(err)
	}
	if f, err := w.Create("folder/a.txt"); err != nil {
		t.Fatal(err)
	} else if _, err := f.Write([]byte("benign")); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	if clean, sig, err := e.Scan(context.Background(), b.Bytes()); !clean || sig != "" || err != nil {
		t.Fatalf("archive with an empty directory entry: clean=%v sig=%q err=%v", clean, sig, err)
	}
}

func ustarMember(t *testing.T, flag byte, data []byte) []byte {
	t.Helper()
	var b bytes.Buffer
	tw := tar.NewWriter(&b)
	if err := tw.WriteHeader(&tar.Header{Name: "x", Mode: 0600, Size: int64(len(data)), Typeflag: flag, Format: tar.FormatUSTAR}); err != nil {
		t.Fatal(err)
	}
	if _, err := tw.Write(data); err != nil {
		t.Fatal(err)
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	return b.Bytes()
}

// An old GNU sparse member ('S') with one data region covering the file.
func gnuSparseTar(data []byte) []byte {
	h := make([]byte, 512)
	octal := func(field []byte, v int) { copy(field, fmt.Sprintf("%0*o\x00", len(field)-1, v)) }
	copy(h, "sparse")
	octal(h[100:108], 0o644)
	octal(h[108:116], 0)
	octal(h[116:124], 0)
	octal(h[124:136], len(data))
	octal(h[136:148], 0)
	h[156] = tar.TypeGNUSparse
	copy(h[257:265], "ustar  \x00")
	octal(h[386:398], 0)
	octal(h[398:410], len(data))
	octal(h[483:495], len(data))
	copy(h[148:156], "        ")
	sum := 0
	for _, c := range h {
		sum += int(c)
	}
	copy(h[148:156], fmt.Sprintf("%06o\x00 ", sum))
	out := append(h, data...)
	out = append(out, make([]byte, (512-len(data)%512)%512)...)
	return append(out, make([]byte, 1024)...)
}

// Tar members with data-bearing typeflags other than '0' must be analyzed, and
// a multi-volume continuation is only part of a file, so it is incomplete.
func TestTarDataBearingTypeflagsAreScanned(t *testing.T) {
	e := nativeTestEngine(t)
	payload := gzipFixture(t, []byte(testMarker))
	if clean, sig, err := e.Scan(context.Background(), ustarMember(t, tar.TypeReg, payload)); clean || sig != "Native.Test" || err != nil {
		t.Fatalf("control: clean=%v sig=%q err=%v", clean, sig, err)
	}
	for _, flag := range []byte{tar.TypeCont, 'Z', 'M', 'V', 'D'} {
		if clean, sig, err := e.Scan(context.Background(), ustarMember(t, flag, payload)); clean || sig != "Native.Test" || err != nil {
			t.Errorf("typeflag %q: data member was not analyzed: clean=%v sig=%q err=%v", flag, clean, sig, err)
		}
	}
	sparse := gnuSparseTar(payload)
	r := tar.NewReader(bytes.NewReader(sparse))
	if h, err := r.Next(); err != nil || h.Typeflag != tar.TypeGNUSparse {
		t.Fatalf("sparse fixture: %v", err)
	}
	if clean, sig, err := e.Scan(context.Background(), sparse); clean || sig != "Native.Test" || err != nil {
		t.Errorf("old GNU sparse member was not analyzed: clean=%v sig=%q err=%v", clean, sig, err)
	}
	if clean, sig, err := e.Scan(context.Background(), ustarMember(t, 'M', []byte("benign"))); clean || sig != "" || !errors.Is(err, ErrUnsupported) {
		t.Errorf("multi-volume continuation reported complete: clean=%v sig=%q err=%v", clean, sig, err)
	}
	for _, flag := range []byte{tar.TypeCont, 'Z'} {
		if clean, sig, err := e.Scan(context.Background(), ustarMember(t, flag, []byte("benign"))); !clean || sig != "" || err != nil {
			t.Errorf("benign typeflag %q: clean=%v sig=%q err=%v", flag, clean, sig, err)
		}
	}
	var b bytes.Buffer
	tw := tar.NewWriter(&b)
	for _, h := range []*tar.Header{
		{Name: "dir/", Typeflag: tar.TypeDir, Mode: 0755, Format: tar.FormatUSTAR},
		{Name: "link", Typeflag: tar.TypeSymlink, Linkname: "dir/a", Format: tar.FormatUSTAR},
	} {
		if err := tw.WriteHeader(h); err != nil {
			t.Fatal(err)
		}
	}
	if err := tw.WriteHeader(&tar.Header{Name: "dir/a", Mode: 0600, Size: 6, Format: tar.FormatUSTAR}); err != nil {
		t.Fatal(err)
	}
	if _, err := tw.Write([]byte("benign")); err != nil {
		t.Fatal(err)
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if clean, sig, err := e.Scan(context.Background(), b.Bytes()); !clean || sig != "" || err != nil {
		t.Fatalf("header-only entries: clean=%v sig=%q err=%v", clean, sig, err)
	}
	// archive/tar reads the bytes a directory entry declares as the next
	// headers; readers that skip them as the entry's data see other members.
	sized := ustarMember(t, tar.TypeReg, make([]byte, 1024))
	sized[156] = tar.TypeDir
	copy(sized[148:156], "        ")
	sum := 0
	for _, c := range sized[:512] {
		sum += int(c)
	}
	copy(sized[148:156], fmt.Sprintf("%06o\x00 ", sum))
	if h, err := tar.NewReader(bytes.NewReader(sized)).Next(); err != nil || h.Typeflag != tar.TypeDir || h.Size != 1024 {
		t.Fatalf("sized directory fixture: %v", err)
	}
	if clean, sig, err := e.Scan(context.Background(), sized); clean || sig != "" || !errors.Is(err, ErrMalformed) {
		t.Fatalf("directory entry declaring data: clean=%v sig=%q err=%v", clean, sig, err)
	}
}

// A Zstandard stream may begin with skippable frames, which decoders ignore.
func TestZstdSkippableFramePrefixIsDecoded(t *testing.T) {
	e := nativeTestEngine(t)
	enc, err := zstd.NewWriter(nil)
	if err != nil {
		t.Fatal(err)
	}
	frame := enc.EncodeAll(gzipFixture(t, []byte(testMarker)), nil)
	_ = enc.Close()
	skippable := func(magic uint32, n int) []byte {
		b := binary.LittleEndian.AppendUint32(nil, magic)
		b = binary.LittleEndian.AppendUint32(b, uint32(n))
		return append(b, make([]byte, n)...)
	}
	for name, data := range map[string][]byte{
		"one":      append(skippable(0x184D2A50, 4), frame...),
		"several":  append(append(skippable(0x184D2A5F, 0), skippable(0x184D2A53, 9)...), frame...),
		"trailing": append(append(append([]byte(nil), frame...), skippable(0x184D2A51, 3)...), frame...),
	} {
		dec, err := zstd.NewReader(nil)
		if err != nil {
			t.Fatal(err)
		}
		out, err := dec.DecodeAll(data, nil)
		dec.Close()
		if err != nil || len(out) == 0 {
			t.Fatalf("%s: reference decoder: len=%d err=%v", name, len(out), err)
		}
		if clean, sig, err := e.Scan(context.Background(), data); clean || sig != "Native.Test" || err != nil {
			t.Errorf("%s: clean=%v sig=%q err=%v", name, clean, sig, err)
		}
	}
	// Skippable frames alone decode to nothing; a truncated frame is malformed.
	if clean, sig, err := e.Scan(context.Background(), skippable(0x184D2A50, 16)); !clean || sig != "" || err != nil {
		t.Errorf("skippable-only stream: clean=%v sig=%q err=%v", clean, sig, err)
	}
	truncated := skippable(0x184D2A50, 16)
	if clean, _, err := e.Scan(context.Background(), truncated[:12]); clean || err == nil {
		t.Errorf("truncated skippable frame reported clean: %v", err)
	}
	if clean, _, err := e.Scan(context.Background(), append(skippable(0x184D2A50, 4), 1, 2, 3, 4, 5)); clean || err == nil {
		t.Errorf("skippable frame followed by an invalid frame reported clean: %v", err)
	}
}

func cabChunks(b []byte) [][]byte {
	var blocks [][]byte
	for len(b) > 0 {
		n := min(len(b), 32768)
		blocks = append(blocks, b[:n])
		b = b[n:]
	}
	return blocks
}

// Stored folders are held and charged in full. Streamed MSZIP produces only
// the requested member views, so unreferenced padding is not retained or
// charged as output. Both paths must stop when no output budget remains.
func TestCABFolderOutputIsChargedToScanBudget(t *testing.T) {
	e := nativeTestEngine(t)
	db := e.current.Load().db
	const padding = 256 << 10
	for _, method := range []uint16{0, 1} {
		cab := []byte("benign")
		var folders, views int64
		for level := 0; level < 6; level++ {
			content := append(append([]byte(nil), cab...), make([]byte, padding)...)
			folders += int64(len(content))
			views += int64(len(cab))
			cab = cabFixture(t, method, cabChunks(content), []cabFixtureFile{{"n", 0, uint32(len(cab))}}, false)
		}
		const budget = 64 << 20
		s := &scanState{ctx: context.Background(), db: db, maxFile: budget, remaining: budget, maxFiles: 100, maxDepth: 16}
		if sig, err := s.visit(cab, 0); sig != "" || err != nil {
			t.Fatalf("large budget: sig=%q err=%v", sig, err)
		}
		if charged := budget - s.remaining; method == 0 && charged < folders {
			t.Fatalf("decoded %d folder bytes but charged only %d", folders, charged)
		} else if method == 1 && charged != int64(len(cab))+views {
			t.Fatalf("streamed member views charged %d, want %d", charged, int64(len(cab))+views)
		}
		s = &scanState{ctx: context.Background(), db: db, maxFile: budget, remaining: int64(len(cab)), maxFiles: 100, maxDepth: 16}
		if sig, err := s.visit(cab, 0); sig != "" || !errors.Is(err, ErrLimit) {
			t.Fatalf("nested folders beyond the budget: sig=%q err=%v", sig, err)
		}
	}
}

func rarStoredMembers(members ...[]byte) []byte {
	block := func(kind byte, flags uint16, payload []byte) []byte {
		h := []byte{0, 0, kind}
		h = binary.LittleEndian.AppendUint16(h, flags)
		h = binary.LittleEndian.AppendUint16(h, uint16(len(payload)+7))
		h = append(h, payload...)
		binary.LittleEndian.PutUint16(h, uint16(crc32.ChecksumIEEE(h[2:])))
		return h
	}
	b := append([]byte("Rar!\x1a\x07\x00"), block(0x73, 0, make([]byte, 6))...)
	for i, data := range members {
		f := binary.LittleEndian.AppendUint32(nil, uint32(len(data)))
		f = binary.LittleEndian.AppendUint32(f, uint32(len(data)))
		f = append(f, 2)
		f = binary.LittleEndian.AppendUint32(f, crc32.ChecksumIEEE(data))
		f = append(f, 0, 0, 0, 0, 20, 0x30, 1, 0, 0, 0, 0, 0, 'a'+byte(i))
		b = append(b, block(0x74, 0x8000, f)...)
		b = append(b, data...)
	}
	return append(b, block(0x7b, 0, nil)...)
}

// RAR members are buffered before recursion. Every buffered member stays in
// memory while an earlier nested member is analyzed, so the nested archive
// must see a budget that already excludes them.
func TestRARBufferedMembersAreChargedDuringNestedAnalysis(t *testing.T) {
	e := nativeTestEngine(t)
	db := e.current.Load().db
	payload := gzipFixture(t, []byte(testMarker))
	nested := rarStoredMembers(payload)
	held := bytes.Repeat([]byte("B"), 4096)
	outer := rarStoredMembers(nested, held)
	// Alive at detection: the outer input, both buffered outer members, the
	// buffered nested member and its decompressed content.
	need := int64(len(outer) + len(nested) + len(held) + len(payload) + len(testMarker))
	scan := func(budget int64) (string, error) {
		s := scanState{ctx: context.Background(), db: db, maxFile: 1 << 20, remaining: budget, maxFiles: 100, maxDepth: 8}
		return s.visit(outer, 0)
	}
	if sig, err := scan(need); sig != "Native.Test" || err != nil {
		t.Fatalf("sufficient budget: sig=%q err=%v", sig, err)
	}
	if sig, err := scan(need - 1); sig != "" || !errors.Is(err, ErrLimit) {
		t.Fatalf("budget excluding a held member: sig=%q err=%v", sig, err)
	}
}

// slowDeflate consumes about 5*n input bytes to produce the single byte "x".
func slowDeflate(n int) []byte {
	return append(bytes.Repeat([]byte{0, 0, 0, 0xff, 0xff}, n), 1, 1, 0, 0xfe, 0xff, 'x')
}

// zipSharedEntries points every central-directory entry at one local record.
func zipSharedEntries(t *testing.T, compressed []byte, flags []uint16) []byte {
	t.Helper()
	var b bytes.Buffer
	w := zip.NewWriter(&b)
	f, err := w.CreateRaw(&zip.FileHeader{Name: "a", Method: zip.Deflate, CRC32: crc32.ChecksumIEEE([]byte("x")),
		CompressedSize64: uint64(len(compressed)), UncompressedSize64: 1})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.Write(compressed); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	data := b.Bytes()
	eocd := data[len(data)-22:]
	off := binary.LittleEndian.Uint32(eocd[16:20])
	header := data[off : len(data)-22]
	out := append([]byte(nil), data[:off]...)
	for _, flag := range flags {
		entry := append([]byte(nil), header...)
		binary.LittleEndian.PutUint16(entry[8:10], flag)
		out = append(out, entry...)
	}
	end := append([]byte(nil), eocd...)
	binary.LittleEndian.PutUint16(end[8:10], uint16(len(flags)))
	binary.LittleEndian.PutUint16(end[10:12], uint16(len(flags)))
	binary.LittleEndian.PutUint32(end[12:16], uint32(len(header)*len(flags)))
	return append(out, end...)
}

// Entries sharing one local record would otherwise decompress the same input
// once per entry. They are one member to the reference (the local header
// gives flags, method and sizes), decoded once; members whose data overlap
// may consume at most the archive's own size in compressed input.
func TestZIPSharedEntriesAreDecodedWithinInputBudget(t *testing.T) {
	e := nativeTestEngine(t)
	compressed := slowDeflate(20000)
	for name, flags := range map[string][]uint16{"duplicates": make([]uint16, 500), "central flags differ": {0, 0x800, 0x2, 0x4}} {
		data := zipSharedEntries(t, compressed, flags)
		if m, err := zipMembers(data, 1000); err != nil || len(m) != len(flags) {
			t.Fatalf("%s fixture: %d %v", name, len(m), err)
		}
		if clean, sig, err := e.Scan(context.Background(), data); !clean || sig != "" || err != nil {
			t.Fatalf("%s: clean=%v sig=%q err=%v", name, clean, sig, err)
		}
	}
	// A member whose data holds a second member's header and data.
	local := func(csize int) []byte {
		h := binary.LittleEndian.AppendUint32(nil, 0x04034b50)
		h = binary.LittleEndian.AppendUint16(h, 20)
		h = binary.LittleEndian.AppendUint16(h, 0)
		h = binary.LittleEndian.AppendUint16(h, 8)
		h = append(h, 0, 0, 0, 0, 0, 0, 0, 0)
		h = binary.LittleEndian.AppendUint32(h, uint32(csize))
		h = binary.LittleEndian.AppendUint32(h, 1)
		h = binary.LittleEndian.AppendUint16(h, 1)
		h = binary.LittleEndian.AppendUint16(h, 0)
		return append(h, 'a')
	}
	central := func(csize, offset int) []byte {
		h := binary.LittleEndian.AppendUint32(nil, 0x02014b50)
		h = append(h, 20, 0, 20, 0, 0, 0, 8, 0, 0, 0, 0, 0, 0, 0, 0, 0)
		h = binary.LittleEndian.AppendUint32(h, uint32(csize))
		h = binary.LittleEndian.AppendUint32(h, 1)
		h = append(h, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0)
		h = binary.LittleEndian.AppendUint32(h, uint32(offset))
		return append(h, 'a')
	}
	inner := append(local(len(compressed)), compressed...)
	body := append(local(len(inner)), inner...)
	directory := append(central(len(inner), 0), central(len(compressed), 31)...)
	end := binary.LittleEndian.AppendUint32(nil, 0x06054b50)
	end = append(end, 0, 0, 0, 0, 2, 0, 2, 0)
	end = binary.LittleEndian.AppendUint32(end, uint32(len(directory)))
	end = binary.LittleEndian.AppendUint32(end, uint32(len(body)))
	overlapping := append(append(body, directory...), append(end, 0, 0)...)
	if clean, sig, err := e.Scan(context.Background(), overlapping); clean || sig != "" || !errors.Is(err, ErrLimit) {
		t.Fatalf("overlapping members: clean=%v sig=%q err=%v", clean, sig, err)
	}
}

// Referenced folders that share CFDATA blocks could decode the same input
// repeatedly; their aggregate work is bounded by the cabinet's input bytes.
func TestCABFoldersSharingBlocksAreDecodedWithinInputBudget(t *testing.T) {
	e := nativeTestEngine(t)
	// Random letters stay compressible enough for MSZIP's 32 KiB block bound.
	random := rand.New(rand.NewSource(6))
	content := make([]byte, 64<<10)
	for i := range content {
		content[i] = 'a' + byte(random.Intn(16))
	}
	base := cabFixture(t, 1, cabChunks(content), []cabFixtureFile{{"a", 0, 6}}, false)
	if clean, sig, err := e.Scan(context.Background(), base); !clean || sig != "" || err != nil {
		t.Fatalf("control: clean=%v sig=%q err=%v", clean, sig, err)
	}
	const folderAt, copies = 36, 3
	fileAt := int(binary.LittleEndian.Uint32(base[16:20]))
	dataAt := int(binary.LittleEndian.Uint32(base[36:40]))
	recordSize := dataAt - fileAt
	shift := uint32((8 + recordSize) * copies)
	shared := append([]byte(nil), base[:folderAt+8]...)
	for i := 0; i < copies; i++ {
		shared = append(shared, base[folderAt:folderAt+8]...)
	}
	for i := 0; i <= copies; i++ {
		record := append([]byte(nil), base[fileAt:dataAt]...)
		binary.LittleEndian.PutUint16(record[8:10], uint16(i))
		shared = append(shared, record...)
	}
	shared = append(shared, base[dataAt:]...)
	binary.LittleEndian.PutUint32(shared[8:12], uint32(len(shared)))
	binary.LittleEndian.PutUint32(shared[16:20], uint32(fileAt+8*copies))
	binary.LittleEndian.PutUint16(shared[26:28], 1+copies)
	binary.LittleEndian.PutUint16(shared[28:30], 1+copies)
	for i := 0; i <= copies; i++ {
		p := folderAt + 8*i
		binary.LittleEndian.PutUint32(shared[p:p+4], binary.LittleEndian.Uint32(shared[p:p+4])+shift)
	}
	if clean, sig, err := e.Scan(context.Background(), shared); clean || sig != "" || !errors.Is(err, ErrLimit) {
		t.Fatalf("folders sharing blocks: clean=%v sig=%q err=%v", clean, sig, err)
	}
}

// xzEmptyBlocks builds a stream of n empty LZMA2 blocks, each declaring the
// dictionary encoded by dictByte. The decoder allocates that dictionary for
// every block, whatever the block's size.
func xzEmptyBlocks(n int, dictByte byte) []byte {
	crc := func(b []byte) []byte { return binary.LittleEndian.AppendUint32(nil, crc32.ChecksumIEEE(b)) }
	out := []byte{0xfd, '7', 'z', 'X', 'Z', 0, 0, 0}
	out = append(out, crc([]byte{0, 0})...)
	header := []byte{2, 0, 0x21, 1, dictByte, 0, 0, 0}
	for i := 0; i < n; i++ {
		out = append(out, header...)
		out = append(out, crc(header)...)
		out = append(out, 0, 0, 0, 0) // LZMA2 end marker, then block padding
	}
	index := []byte{0}
	index = binary.AppendUvarint(index, uint64(n))
	for i := 0; i < n; i++ {
		index = append(index, 13, 0)
	}
	for len(index)%4 != 0 {
		index = append(index, 0)
	}
	out = append(append(out, index...), crc(index)...)
	footer := binary.LittleEndian.AppendUint32(nil, uint32((len(index)+4)/4-1))
	footer = append(footer, 0, 0)
	return append(append(append(out, crc(footer)...), footer...), 'Y', 'Z')
}

// Each XZ block allocates its declared dictionary even when it holds no data.
// The dictionary memory that no block output accounts for is capped.
func TestXZPreflightBoundsDictionaryAllocationAcrossBlocks(t *testing.T) {
	small := xzEmptyBlocks(3, 0)
	r, err := xz.NewReader(bytes.NewReader(small))
	if err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if _, err := out.ReadFrom(r); err != nil || out.Len() != 0 {
		t.Fatalf("fixture rejected by decoder: len=%d err=%v", out.Len(), err)
	}
	if err := preflightXZ(small, 1<<20); err != nil {
		t.Fatal(err)
	}
	const dict64MiB = 28
	if err := preflightXZ(xzEmptyBlocks(1, dict64MiB), 1<<20); err != nil {
		t.Fatalf("one oversized dictionary: %v", err)
	}
	if err := preflightXZ(xzEmptyBlocks(2000, dict64MiB), 1<<20); !errors.Is(err, ErrLimit) {
		t.Fatalf("2000 empty 64 MiB-dictionary blocks: %v", err)
	}
	if err := preflightXZ(xzEmptyBlocks(5, dict64MiB), 1<<20); !errors.Is(err, ErrLimit) {
		t.Fatalf("five empty 64 MiB-dictionary blocks: %v", err)
	}
}
