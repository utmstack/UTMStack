package engine

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"os"
	"slices"
	"testing"
	"time"
	"unicode/utf16"

	"github.com/utmstack/UTMStack/agent/edr/signatures"
)

const oleTestMarker = "OLE-HARMLESS-EXTRACTED-MARKER"

func oleTestState(t *testing.T) *scanState {
	t.Helper()
	db, err := signatures.Compile(map[string][]byte{"test.ndb": []byte("OLE.Test:0:*:" + hex.EncodeToString([]byte(oleTestMarker)) + "\n")})
	if err != nil {
		t.Fatal(err)
	}
	return &scanState{ctx: context.Background(), db: db, maxFile: 16 << 20, remaining: 32 << 20, maxFiles: 100, maxDepth: 10}
}

// Synthetic streams deliberately use a non-contiguous allocation order. This
// proves extraction, rather than an accidental match in the container bytes.
func oleFixture(t *testing.T, major int, name string, payload []byte, mini, vba, difat bool) []byte {
	t.Helper()
	sector := 512
	shift := uint16(9)
	if major == 4 {
		sector = 4096
		shift = 12
	}
	if mini && len(payload) >= 4096 || !mini && len(payload) < 4096 {
		t.Fatal("invalid fixture stream size")
	}
	ceil := func(n, d int) int { return (n + d - 1) / d }
	rootSize := 0
	streamCount := ceil(len(payload), sector)
	if mini {
		rootSize = ceil(len(payload), 64) * 64
		streamCount = ceil(rootSize, sector)
	}
	streamIDs := make([]uint32, streamCount)
	for i := range streamIDs {
		streamIDs[i] = uint32(2 + i)
	}
	if len(streamIDs) > 1 {
		streamIDs[0], streamIDs[1] = streamIDs[1], streamIDs[0]
	}
	n := 2 + streamCount
	miniFatID := oleEnd
	if mini {
		miniFatID = uint32(n)
		n++
	}
	difatID := oleEnd
	if difat {
		difatID = uint32(n)
		n++
	}
	b := make([]byte, (n+1)*sector)
	copy(b, oleMagic)
	p16 := func(off int, v uint16) { binary.LittleEndian.PutUint16(b[off:], v) }
	p32 := func(off int, v uint32) { binary.LittleEndian.PutUint32(b[off:], v) }
	p16(24, 0x3e)
	p16(26, uint16(major))
	p16(28, 0xfffe)
	p16(30, shift)
	p16(32, 6)
	if major == 4 {
		p32(40, 1)
	}
	p32(44, 1)
	p32(48, 1)
	p32(56, 4096)
	p32(60, miniFatID)
	p32(68, difatID)
	if mini {
		p32(64, 1)
	}
	if difat {
		p32(72, 1)
	}
	for off := 76; off < 512; off += 4 {
		p32(off, oleFree)
	}
	if !difat {
		p32(76, 0)
	}
	fat := b[sector : 2*sector]
	for off := 0; off < sector; off += 4 {
		binary.LittleEndian.PutUint32(fat[off:], oleFree)
	}
	setFAT := func(id, next uint32) { binary.LittleEndian.PutUint32(fat[int(id)*4:], next) }
	setFAT(0, oleFAT)
	setFAT(1, oleEnd)
	for i, id := range streamIDs {
		next := oleEnd
		if i+1 < len(streamIDs) {
			next = streamIDs[i+1]
		}
		setFAT(id, next)
	}
	writeEntry := func(id int, name string, typ byte, start uint32, size uint64, child uint32) {
		e := b[2*sector+id*128 : 2*sector+(id+1)*128]
		units := utf16.Encode([]rune(name))
		if len(units) > 31 {
			t.Fatal("fixture name too long")
		}
		for i, u := range units {
			binary.LittleEndian.PutUint16(e[i*2:], u)
		}
		binary.LittleEndian.PutUint16(e[64:], uint16((len(units)+1)*2))
		e[66] = typ
		e[67] = 1
		binary.LittleEndian.PutUint32(e[68:], oleFree)
		binary.LittleEndian.PutUint32(e[72:], oleFree)
		binary.LittleEndian.PutUint32(e[76:], child)
		binary.LittleEndian.PutUint32(e[116:], start)
		binary.LittleEndian.PutUint64(e[120:], size)
	}
	rootStart := oleEnd
	if mini {
		rootStart = streamIDs[0]
	}
	writeEntry(0, "Root Entry", 5, rootStart, uint64(rootSize), 1)
	streamDir := 1
	if vba {
		writeEntry(1, "VBA", 1, 0, 0, 2)
		streamDir = 2
	}
	if mini {
		setFAT(miniFatID, oleEnd)
		table := b[(int(miniFatID)+1)*sector : (int(miniFatID)+2)*sector]
		for off := 0; off < sector; off += 4 {
			binary.LittleEndian.PutUint32(table[off:], oleFree)
		}
		miniIDs := make([]uint32, ceil(len(payload), 64))
		for i := range miniIDs {
			miniIDs[i] = uint32(i)
		}
		if len(miniIDs) > 1 {
			miniIDs[0], miniIDs[1] = miniIDs[1], miniIDs[0]
		}
		root := make([]byte, rootSize)
		for i, id := range miniIDs {
			next := oleEnd
			if i+1 < len(miniIDs) {
				next = miniIDs[i+1]
			}
			binary.LittleEndian.PutUint32(table[int(id)*4:], next)
			copy(root[int(id)*64:], payload[i*64:min(len(payload), (i+1)*64)])
		}
		for i, id := range streamIDs {
			copy(b[(int(id)+1)*sector:], root[i*sector:min(len(root), (i+1)*sector)])
		}
		start := oleEnd
		if len(miniIDs) > 0 {
			start = miniIDs[0]
		}
		writeEntry(streamDir, name, 2, start, uint64(len(payload)), oleFree)
	} else {
		for i, id := range streamIDs {
			copy(b[(int(id)+1)*sector:], payload[i*sector:min(len(payload), (i+1)*sector)])
		}
		writeEntry(streamDir, name, 2, streamIDs[0], uint64(len(payload)), oleFree)
	}
	if difat {
		setFAT(difatID, oleDIFAT)
		d := b[(int(difatID)+1)*sector : (int(difatID)+2)*sector]
		for off := 0; off < sector; off += 4 {
			binary.LittleEndian.PutUint32(d[off:], oleFree)
		}
		binary.LittleEndian.PutUint32(d, 0)
		binary.LittleEndian.PutUint32(d[sector-4:], oleEnd)
	}
	return b
}

func TestOLEExtractsFragmentedNormalAndMiniStreams(t *testing.T) {
	for _, tc := range []struct {
		name        string
		major       int
		mini, difat bool
	}{{"v3 normal", 3, false, false}, {"v4 normal", 4, false, false}, {"mini", 3, true, false}, {"v4 mini", 4, true, false}} {
		t.Run(tc.name, func(t *testing.T) {
			length, boundary := 8192, 512
			if tc.major == 4 {
				boundary = 4096
			}
			if tc.mini {
				length, boundary = 128, 64
			}
			payload := bytes.Repeat([]byte{'x'}, length)
			copy(payload[boundary-8:], oleTestMarker)
			b := oleFixture(t, tc.major, "payload.bin", payload, tc.mini, false, tc.difat)
			if bytes.Contains(b, []byte(oleTestMarker)) {
				t.Fatal("fixture must require stream reconstruction")
			}
			sig, err := oleTestState(t).ole(b, 0)
			if err != nil || sig != "OLE.Test" {
				t.Fatalf("sig=%q err=%v", sig, err)
			}
		})
	}
}

func TestOLEOrdinaryExtractedStreamsMayBeClean(t *testing.T) {
	b := oleFixture(t, 3, "Résumé_文書", []byte("ordinary harmless data"), true, false, false)
	if sig, err := oleTestState(t).ole(b, 0); sig != "" || err != nil {
		t.Fatalf("sig=%q err=%v", sig, err)
	}
}

// Damaged structures are read as the reference reads them: what can be read
// is kept and scanned, so these files are decided rather than incomplete.
// TestOLEReadsAsReference checks the exact bytes on 326 cases.
func TestOLEDamagedStructuresReadAsReference(t *testing.T) {
	const dir = 1024 // directory sector; entry 1 is the stream
	base := func() []byte {
		payload := append([]byte(oleTestMarker), bytes.Repeat([]byte{'x'}, 8192)...)
		return oleFixture(t, 3, "payload.bin", payload, false, false, false)
	}
	first := func(b []byte) uint32 { return binary.LittleEndian.Uint32(b[dir+128+116:]) }
	cases := []struct {
		name   string
		mutate func([]byte)
		found  bool
	}{
		// The chain ends early: the first sector, which holds the marker, is kept.
		{"fat cycle", func(b []byte) { binary.LittleEndian.PutUint32(b[512+int(first(b))*4:], first(b)) }, true},
		{"truncated chain", func(b []byte) { binary.LittleEndian.PutUint32(b[512+int(first(b))*4:], oleEnd) }, true},
		{"chain past the end", func(b []byte) { binary.LittleEndian.PutUint32(b[512+int(first(b))*4:], 999999) }, true},
		// Neither the FAT count nor the extra header FAT numbers are read.
		{"fat count", func(b []byte) { binary.LittleEndian.PutUint32(b[44:], 0) }, true},
		{"unused fat numbers", func(b []byte) { binary.LittleEndian.PutUint32(b[80:], 999999) }, true},
		// No FAT: only a stream's first sector is read.
		{"fat sector past the end", func(b []byte) { binary.LittleEndian.PutUint32(b[76:], 999999) }, true},
		{"negative fat count", func(b []byte) { binary.LittleEndian.PutUint32(b[44:], 0x80000000) }, true},
		{"bad name length", func(b []byte) { binary.LittleEndian.PutUint16(b[dir+128+64:], 65) }, true},
		{"unpaired surrogate", func(b []byte) { binary.LittleEndian.PutUint16(b[dir+128:], 0xd800) }, true},
		{"high size word", func(b []byte) { binary.LittleEndian.PutUint32(b[dir+128+124:], 1) }, true},
		{"storage with a size", func(b []byte) { binary.LittleEndian.PutUint32(b[dir+120:], 99) }, true},
		{"byte order", func(b []byte) { binary.LittleEndian.PutUint16(b[28:], 0) }, true},
		// The stream starts in the directory sector.
		{"aliased directory sector", func(b []byte) { binary.LittleEndian.PutUint32(b[dir+128+116:], 1) }, false},
		// Unreachable or unreadable entries are not read.
		{"unreachable stream", func(b []byte) { binary.LittleEndian.PutUint32(b[dir+76:], oleFree) }, false},
		{"directory bounds", func(b []byte) { binary.LittleEndian.PutUint32(b[dir+76:], 999999) }, false},
		{"unused entry", func(b []byte) { b[dir+128+66] = 0 }, false},
		{"unknown entry type", func(b []byte) { b[dir+128+66] = 3 }, false},
		// Header values that make the reference read nothing.
		{"reserved bytes", func(b []byte) { b[34] = 1 }, false},
		{"mini shift above the sector shift", func(b []byte) { binary.LittleEndian.PutUint16(b[32:], 10) }, false},
		{"sector shift 31", func(b []byte) { binary.LittleEndian.PutUint16(b[30:], 31) }, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			b := base()
			tc.mutate(b)
			sig, err := oleTestState(t).ole(b, 0)
			if err != nil || (sig == "OLE.Test") != tc.found {
				t.Fatalf("sig=%q err=%v, want found=%v", sig, err, tc.found)
			}
		})
	}
	// A directory loop ends the whole walk; the stream after it is not read.
	b := base()
	binary.LittleEndian.PutUint32(b[dir+76:], 0)
	if sig, err := oleTestState(t).ole(b, 0); sig != "" || err != nil {
		t.Fatalf("loop: sig=%q err=%v", sig, err)
	}
	// Sector shifts below 9 are left undecided.
	b = base()
	binary.LittleEndian.PutUint16(b[30:], 8)
	if sig, err := oleTestState(t).ole(b, 0); sig != "" || unknownReasonOf(err) != signatures.UnknownCompoundFile {
		t.Fatalf("shift 8: sig=%q err=%v", sig, err)
	}
	// Fewer than 517 bytes: the reference reads nothing.
	for _, end := range []int{0, 7, 511, 516, 1024} {
		if sig, err := oleTestState(t).ole(base()[:end], 0); sig != "" || err != nil {
			t.Fatalf("%d bytes: sig=%q err=%v", end, sig, err)
		}
	}
}

// Some writers mark the FAT's own sectors as the end of a chain or free
// (PowerPoint files in Apache POI's test data); the reference reads them.
func TestOLEAcceptsFATSectorsMarkedEndOrFree(t *testing.T) {
	for _, marker := range []uint32{oleEnd, oleFree} {
		b := oleFixture(t, 3, "payload.bin", append(bytes.Repeat([]byte{'x'}, 4096), oleTestMarker...), false, false, false)
		binary.LittleEndian.PutUint32(b[512:], marker)
		if sig, err := oleTestState(t).ole(b, 0); sig != "OLE.Test" || err != nil {
			t.Fatalf("%x: sig=%q err=%v", marker, sig, err)
		}
	}
}

// The reference takes the first 109 FAT sectors only from the header: a FAT
// sector listed only in the extension chain is not used, so a stream's
// sectors after its first cannot be followed.
func TestOLEFATSectorOnlyInExtensionIsNotUsed(t *testing.T) {
	payload := bytes.Repeat([]byte{'x'}, 8192)
	copy(payload[512-8:], oleTestMarker)
	b := oleFixture(t, 3, "payload.bin", payload, false, false, true)
	if sig, err := oleTestState(t).ole(b, 0); sig != "" || err != nil {
		t.Fatalf("sig=%q err=%v", sig, err)
	}
}

// A mini chain that comes back to a block ends there; the blocks read are
// kept.
func TestOLEMiniCycleKeepsBlocksRead(t *testing.T) {
	payload := append([]byte(oleTestMarker), bytes.Repeat([]byte{'x'}, 200)...)
	b := oleFixture(t, 3, "payload", payload, true, false, false)
	tableID := binary.LittleEndian.Uint32(b[60:])
	start := binary.LittleEndian.Uint32(b[1024+128+116:])
	binary.LittleEndian.PutUint32(b[(int(tableID)+1)*512+int(start)*4:], start)
	if sig, err := oleTestState(t).ole(b, 0); sig != "OLE.Test" || err != nil {
		t.Fatalf("sig=%q err=%v", sig, err)
	}
}

// testdata/cfb/oracle.zip holds generated compound files with deliberate
// damage (docs/2026-10-14-cfb-oracle/mkcfb.py) and, for each, the streams the
// reference wrote out (name, size, SHA-256) at the recursion limit it ran
// with, or "undecided" where this engine leaves the file undecided.
func TestOLEReadsAsReference(t *testing.T) {
	z, err := zip.OpenReader("testdata/cfb/oracle.zip")
	if err != nil {
		t.Fatal(err)
	}
	defer z.Close()
	files := map[string][]byte{}
	for _, f := range z.File {
		r, err := f.Open()
		if err != nil {
			t.Fatal(err)
		}
		files[f.Name], err = io.ReadAll(r)
		r.Close()
		if err != nil {
			t.Fatal(err)
		}
	}
	var expected map[string]struct {
		Rec       int     `json:"rec"`
		Undecided bool    `json:"undecided"`
		Streams   [][]any `json:"streams"`
	}
	if err := json.Unmarshal(files["expected.json"], &expected); err != nil {
		t.Fatal(err)
	}
	if len(expected) != 326 {
		t.Fatalf("%d cases", len(expected))
	}
	for name, want := range expected {
		data := files[name]
		s := oleTestState(t)
		s.maxDepth, s.maxFile, s.remaining = want.Rec, 400<<20, 400<<20-int64(len(data))
		got, undecided, err := oleReadAll(s, data)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if want.Undecided || undecided {
			if want.Undecided != undecided {
				t.Errorf("%s: undecided=%v, want %v", name, undecided, want.Undecided)
			}
			continue
		}
		var w []string
		for _, s := range want.Streams {
			w = append(w, s[2].(string))
		}
		if slices.Sort(w); !slices.Equal(got, w) {
			t.Errorf("%s: read %d streams %v, the reference %d %v", name, len(got), got, len(w), want.Streams)
		}
	}
}

// oleReadAll returns the sorted SHA-256 of every stream readOLE reads (a
// stream over the scan budget, which the reference skips, is left out), and
// whether the file is undecided. Streams read before a limit ended the walk
// are compared too.
func oleReadAll(s *scanState, data []byte) ([]string, bool, error) {
	f, err := readOLE(s, data)
	if err != nil {
		return nil, false, err
	}
	out := []string{}
	for _, st := range f.streams {
		b, err := f.read(st)
		if errors.Is(err, ErrLimit) {
			continue
		}
		if err != nil {
			return nil, false, err
		}
		h := sha256.Sum256(b)
		out = append(out, hex.EncodeToString(h[:]))
	}
	slices.Sort(out)
	return out, f.undecided != nil, nil
}

// TestOLECorpusAsReference compares readOLE with the reference on real or
// damaged files measured in the lab: EDR_CFB_ORACLE names a refole.py result
// (docs/2026-10-14-cfb-oracle/refole.py, run with the default recursion
// limit 17), whose paths must be readable here.
func TestOLECorpusAsReference(t *testing.T) {
	path := os.Getenv("EDR_CFB_ORACLE")
	if path == "" {
		t.Skip("EDR_CFB_ORACLE not set")
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var ref map[string]struct {
		Dumps [][]any `json:"dumps"`
	}
	if err := json.Unmarshal(raw, &ref); err != nil {
		t.Fatal(err)
	}
	names := make([]string, 0, len(ref))
	for name := range ref {
		names = append(names, name)
	}
	slices.Sort(names)
	equal, undecided, differ := 0, 0, 0
	for _, name := range names {
		data, err := os.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		s := oleTestState(t)
		s.maxDepth, s.maxFile, s.remaining, s.maxFiles = 17, 400<<20, 400<<20-int64(len(data)), 10000
		got, und, err := oleReadAll(s, data)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if und {
			undecided++
			continue
		}
		w := []string{}
		for _, d := range ref[name].Dumps {
			w = append(w, d[2].(string))
		}
		slices.Sort(w)
		if slices.Equal(got, w) {
			equal++
			continue
		}
		differ++
		t.Logf("%s: read %d streams, the reference %d", name, len(got), len(w))
	}
	t.Logf("%d files: %d as the reference, %d undecided, %d different", len(names), equal, undecided, differ)
}

func TestOLELimitsAndCancellation(t *testing.T) {
	b := oleFixture(t, 4, "payload", bytes.Repeat([]byte{'x'}, 8192), false, false, false)
	s := oleTestState(t)
	s.remaining = 4096
	if _, err := s.ole(b, 0); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	s = oleTestState(t)
	s.maxDepth = 0
	if _, err := s.ole(b, 0); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	s = oleTestState(t)
	s.maxFiles = 1
	if _, err := s.ole(b, 0); err != nil {
		t.Fatalf("one stream should fit one slot: %v", err)
	}
	s = oleTestState(t)
	s.maxFiles = 0
	if _, err := s.ole(b, 0); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	s = oleTestState(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	s.ctx = ctx
	if _, err := s.ole(b, 0); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
}

func TestOLEEncryptedIsIncompleteWithoutDetection(t *testing.T) {
	for _, name := range []string{"EncryptedPackage", "EncryptionInfo"} {
		b := oleFixture(t, 3, name, []byte("encrypted bytes"), true, false, false)
		if sig, err := oleTestState(t).ole(b, 0); sig != "" || !errors.Is(err, ErrUnsupported) {
			t.Fatalf("%s: %q %v", name, sig, err)
		}
	}
	// As in the reference, a VBA storage in a file without a _VBA_PROJECT
	// stream is no macro project: its streams are ordinary streams.
	b := oleFixture(t, 3, "Module1", []byte("no compressed source"), true, true, false)
	if sig, err := oleTestState(t).ole(b, 0); sig != "" || err != nil {
		t.Fatalf("%q %v", sig, err)
	}
}

func vbaLiteralFixture(data []byte) []byte {
	var body []byte
	for i := 0; i < len(data); i += 8 {
		body = append(body, 0)
		body = append(body, data[i:min(len(data), i+8)]...)
	}
	out := []byte{1, 0, 0}
	binary.LittleEndian.PutUint16(out[1:], uint16(len(body)-1)|0xb000)
	return append(out, body...)
}

func TestVBALiteralAndOverlappingCopyDecompression(t *testing.T) {
	literal := vbaLiteralFixture([]byte(oleTestMarker))
	out, err := decompressVBA(context.Background(), literal, 4096)
	if err != nil || string(out) != oleTestMarker {
		t.Fatalf("%q %v", out, err)
	}
	// One literal followed by a distance-1 overlapping copy expands to 4096 A's.
	compressed := []byte{1, 3, 0xb0, 2, 'A', 0xfc, 0x0f}
	out, err = decompressVBA(context.Background(), compressed, 4096)
	if err != nil || !bytes.Equal(out, bytes.Repeat([]byte{'A'}, 4096)) {
		t.Fatalf("len=%d err=%v", len(out), err)
	}
	if _, err := decompressVBA(context.Background(), compressed, 4095); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	raw := append([]byte{1, 0xff, 0x3f}, bytes.Repeat([]byte{'B'}, 4096)...)
	out, err = decompressVBA(context.Background(), raw, 4096)
	if err != nil || !bytes.Equal(out, bytes.Repeat([]byte{'B'}, 4096)) {
		t.Fatal(len(out), err)
	}
	multiple := append(append([]byte(nil), raw...), literal[1:]...)
	out, err = decompressVBA(context.Background(), multiple, 8192)
	if err != nil || !bytes.HasSuffix(out, []byte(oleTestMarker)) {
		t.Fatal(len(out), err)
	}
}

func TestVBACompressedSourceOutsideAProjectIsNotDecompressed(t *testing.T) {
	// The reference decompresses module source only for modules of a
	// project (vbaproject_test.go); compressed data elsewhere is scanned as
	// it is.
	compressed := vbaLiteralFixture([]byte(oleTestMarker))
	if bytes.Contains(compressed, []byte(oleTestMarker)) {
		t.Fatal("fixture must need decompression")
	}
	payload := append(bytes.Repeat([]byte{0x44}, 97), compressed...)
	b := oleFixture(t, 3, "Module1", payload, true, true, false)
	sig, err := oleTestState(t).ole(b, 0)
	if sig != "" || err != nil {
		t.Fatalf("sig=%q err=%v", sig, err)
	}
}

func TestVBAMalformedChunksAndTokensAreRejected(t *testing.T) {
	for _, data := range [][]byte{{}, {1}, {1, 0, 0}, {1, 0, 0xb0, 0}, {1, 3, 0xb0, 1, 0, 0, 0}, {1, 3, 0xb0, 2, 'A', 0xfc, 0x1f}, {1, 0, 0x30, 'a'}, {1, 0xff, 0xbf, 0}} {
		if _, err := decompressVBA(context.Background(), data, 8192); err == nil {
			t.Fatalf("accepted malformed container %x", data)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := decompressVBA(ctx, vbaLiteralFixture([]byte("data")), 1024); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
}

func FuzzOLEParserNeverPanics(f *testing.F) {
	f.Add([]byte("ordinary"))
	f.Add(append(append([]byte(nil), oleMagic...), make([]byte, 504)...))
	// Macro projects and workbooks with pictures (vbaproject_test.go).
	if files, err := vbaOracleFiles(); err == nil {
		for _, name := range []string{"gen/basic.doc", "gen/many.doc", "p3/fail-then-good.doc", "p4/where-two.doc", "dl/029-raw.doc"} {
			f.Add(files[name])
		}
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 1<<20 {
			t.Skip()
		}
		s := oleTestState(t)
		s.maxFile = 1 << 20
		s.remaining = 2 << 20
		_, _ = s.ole(data, 0)
	})
}

func FuzzVBADecompressNeverPanics(f *testing.F) {
	f.Add([]byte{1, 3, 0xb0, 2, 'A', 0xfc, 0x0f})
	f.Add(vbaLiteralFixture([]byte("harmless")))
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 65536 {
			t.Skip()
		}
		_, _ = decompressVBA(context.Background(), data, 65536)
	})
}

// FuzzOLERead checks that any input is read within its budget and without
// failing: seeds are the generated cases of testdata/cfb/oracle.zip.
func FuzzOLERead(f *testing.F) {
	z, err := zip.OpenReader("testdata/cfb/oracle.zip")
	if err != nil {
		f.Fatal(err)
	}
	for i, file := range z.File {
		if i%8 != 0 || file.UncompressedSize64 > 64<<10 {
			continue
		}
		r, err := file.Open()
		if err != nil {
			f.Fatal(err)
		}
		b, err := io.ReadAll(r)
		r.Close()
		if err != nil {
			f.Fatal(err)
		}
		f.Add(b)
	}
	z.Close()
	f.Fuzz(func(t *testing.T, data []byte) {
		start := time.Now()
		s := oleTestState(t)
		s.remaining = 1 << 20
		if _, _, err := oleReadAll(s, data); err != nil {
			t.Fatal(err)
		}
		if _, err := oleTestState(t).ole(data, 0); errors.Is(err, context.Canceled) {
			t.Fatal(err)
		}
		if d := time.Since(start); d > time.Second {
			t.Fatalf("%d bytes read in %v", len(data), d)
		}
	})
}

// A walk the recursion limit ends is a limit, as the reference reports it
// (Heuristics.Limits.Exceeded.MaxRecursion); the streams read before it are
// still scanned.
func TestOLEDeepTreeIsALimit(t *testing.T) {
	z, err := zip.OpenReader("testdata/cfb/oracle.zip")
	if err != nil {
		t.Fatal(err)
	}
	defer z.Close()
	for _, name := range []string{"more2/p_nest16.bin", "more2/p_nest17.bin"} {
		r, err := z.Open(name)
		if err != nil {
			t.Fatal(err)
		}
		data, err := io.ReadAll(r)
		r.Close()
		if err != nil {
			t.Fatal(err)
		}
		s := oleTestState(t)
		s.maxDepth = 17
		_, err = s.ole(data, 0)
		if deep := name == "more2/p_nest17.bin"; errors.Is(err, ErrLimit) != deep {
			t.Fatalf("%s: %v", name, err)
		}
	}
}
