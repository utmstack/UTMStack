package engine

import (
	"bytes"
	"compress/flate"
	"context"
	"encoding/binary"
	"errors"
	"io"
	"math/rand"
	"testing"

	"github.com/utmstack/UTMStack/agent/edr/mszip/msziptest"
)

type cabFixtureFile struct {
	name        string
	start, size uint32
}

// MS-CAB checksums are used by the fixture encoder. The scanner follows
// ClamAV's enabled salvage mode and does not reject mismatches.
func cabChecksum(data []byte) uint32 {
	var sum uint32
	for len(data) >= 4 {
		sum ^= binary.LittleEndian.Uint32(data[:4])
		data = data[4:]
	}
	var tail uint32
	for _, b := range data {
		tail = tail<<8 | uint32(b)
	}
	return sum ^ tail
}

// Encode a single-folder cabinet from the field widths in MS-CAB. Tests keep
// this builder separate from the production parser and use real DEFLATE blocks.
func cabFixture(t testing.TB, method uint16, blocks [][]byte, files []cabFixtureFile, reserve bool) []byte {
	t.Helper()
	b := make([]byte, 36)
	copy(b, "MSCF")
	b[24], b[25] = 3, 1
	binary.LittleEndian.PutUint16(b[26:28], 1)
	binary.LittleEndian.PutUint16(b[28:30], uint16(len(files)))
	if reserve {
		binary.LittleEndian.PutUint16(b[30:32], 4)
		b = append(b, 2, 0, 1, 1, 0xaa, 0xbb)
	}
	folderPosition := len(b)
	b = append(b, make([]byte, 8)...)
	if reserve {
		b = append(b, 0xcc)
	}
	binary.LittleEndian.PutUint32(b[16:20], uint32(len(b)))
	for _, file := range files {
		entry := make([]byte, 16)
		binary.LittleEndian.PutUint32(entry[0:4], file.size)
		binary.LittleEndian.PutUint32(entry[4:8], file.start)
		b = append(b, entry...)
		b = append(b, file.name...)
		b = append(b, 0)
	}
	binary.LittleEndian.PutUint32(b[folderPosition:folderPosition+4], uint32(len(b)))
	binary.LittleEndian.PutUint16(b[folderPosition+4:folderPosition+6], uint16(len(blocks)))
	binary.LittleEndian.PutUint16(b[folderPosition+6:folderPosition+8], method)
	history := make([]byte, 32768)
	for _, block := range blocks {
		payload := block
		if method == 1 {
			payload = msziptest.Encode(block, history)
		}
		header := make([]byte, 8)
		binary.LittleEndian.PutUint16(header[4:6], uint16(len(payload)))
		binary.LittleEndian.PutUint16(header[6:8], uint16(len(block)))
		checksum := cabChecksum(header[4:8]) ^ cabChecksum(payload)
		binary.LittleEndian.PutUint32(header[:4], checksum)
		b = append(b, header...)
		if reserve {
			b = append(b, 0xdd)
		}
		b = append(b, payload...)
		copy(history, block) // MSZIP restarts each frame at window offset zero
	}
	binary.LittleEndian.PutUint32(b[8:12], uint32(len(b)))
	return b
}

func TestCABChecksumRemainderAndCorruption(t *testing.T) {
	if got := cabChecksum([]byte{1, 2, 3, 4, 5, 6, 7}); got != 0x04060406 {
		t.Fatalf("checksum=%08x", got)
	}
	e := nativeTestEngine(t)
	content := gzipFixture(t, []byte(testMarker))
	archive := cabFixture(t, 0, [][]byte{content}, []cabFixtureFile{{"a.gz", 0, uint32(len(content))}}, false)
	if bytes.Contains(archive, []byte(testMarker)) {
		t.Fatal("fixture exposes marker without extraction")
	}
	if clean, sig, err := e.Scan(context.Background(), archive); clean || sig != "Native.Test" || err != nil {
		t.Fatal(clean, sig, err)
	}
	if clean, sig, err := e.Scan(context.Background(), append(peValidationFixture(), archive...)); clean || sig != "Native.Test" || err != nil {
		t.Fatalf("CAB PE overlay: %v %q %v", clean, sig, err)
	}
	if clean, sig, err := e.Scan(context.Background(), append(append([]byte(nil), archive...), []byte("unclassified trailer")...)); clean || sig != "Native.Test" || err != nil {
		t.Fatalf("CAB with trailer hid contained detection: %v %q %v", clean, sig, err)
	}
	withoutChecksum := append([]byte(nil), archive...)
	blockStart := int(binary.LittleEndian.Uint32(withoutChecksum[36:40]))
	binary.LittleEndian.PutUint32(withoutChecksum[blockStart:blockStart+4], 0)
	if clean, sig, err := e.Scan(context.Background(), withoutChecksum); clean || sig != "Native.Test" || err != nil {
		t.Fatalf("CAB without optional checksum: %v %q %v", clean, sig, err)
	}
	archive[len(archive)-1] ^= 1
	if clean, sig, err := e.Scan(context.Background(), archive); clean || sig != "Native.Test" || err != nil {
		t.Fatalf("bad checksum hid recoverable nested payload: %v %q %v", clean, sig, err)
	}
}

func TestCABMSZIPHistoryAndFolderFileOffsets(t *testing.T) {
	e := nativeTestEngine(t)
	random := rand.New(rand.NewSource(23))
	first := make([]byte, 32768)
	if _, err := random.Read(first); err != nil {
		t.Fatal(err)
	}
	second := append(append([]byte(nil), first[len(first)-2048:]...), []byte(testMarker)...)
	archive := cabFixture(t, 1, [][]byte{first, second}, []cabFixtureFile{
		{"first.bin", 0, uint32(len(first))},
		{"second.bin", uint32(len(first)), uint32(len(second))},
	}, true)
	if bytes.Contains(archive, []byte(testMarker)) {
		t.Fatal("fixture exposes marker without MSZIP decoding")
	}
	// The second block is encoded against the first block's history. Reading
	// it with a fresh DEFLATE stream must fail, so this exercises dictionary
	// continuity rather than two independent compressed blocks.
	firstBlock := int(binary.LittleEndian.Uint32(archive[42:46]))
	secondBlock := firstBlock + 8 + 1 + int(binary.LittleEndian.Uint16(archive[firstBlock+4:firstBlock+6]))
	secondSize := int(binary.LittleEndian.Uint16(archive[secondBlock+4 : secondBlock+6]))
	secondPayload := archive[secondBlock+8+1 : secondBlock+8+1+secondSize]
	if !bytes.HasPrefix(secondPayload, []byte("CK")) {
		t.Fatal("fixture lacks MSZIP signature")
	}
	r := flate.NewReader(bytes.NewReader(secondPayload[2:]))
	_, independentErr := io.ReadAll(r)
	_ = r.Close()
	if independentErr == nil {
		t.Fatal("second block did not depend on folder history")
	}
	if clean, sig, err := e.Scan(context.Background(), archive); clean || sig != "Native.Test" || err != nil {
		t.Fatal(clean, sig, err)
	}
}

func TestCABUnsupportedAndMalformedNeverClean(t *testing.T) {
	e := nativeTestEngine(t)
	base := cabFixture(t, 0, [][]byte{[]byte("benign")}, []cabFixtureFile{{"a", 0, 6}}, false)
	for index, mutate := range []func([]byte) []byte{
		func(b []byte) []byte { binary.LittleEndian.PutUint16(b[42:44], 3); return b },      // LZX
		func(b []byte) []byte { binary.LittleEndian.PutUint16(b[42:44], 2); return b },      // Quantum
		func(b []byte) []byte { return cabHeaderSet(b, 1, 0xfffd) },                         // unavailable preceding volume
		func(b []byte) []byte { binary.LittleEndian.PutUint16(b[52:54], 0xffff); return b }, // split file
		func(b []byte) []byte { binary.LittleEndian.PutUint32(b[36:40], 0xffffffff); return b },
		func(b []byte) []byte { return b[:len(b)-1] },
		func(b []byte) []byte {
			b = b[:len(b)-1]
			binary.LittleEndian.PutUint32(b[8:12], uint32(len(b)))
			return b
		},
	} {
		b := mutate(append([]byte(nil), base...))
		clean, sig, err := e.Scan(context.Background(), b)
		if clean || sig != "" || err == nil {
			t.Fatalf("invalid CAB case %d became clean: %v %q %v", index, clean, sig, err)
		}
	}
	if clean, sig, err := e.Scan(context.Background(), base); !clean || sig != "" || err != nil {
		t.Fatalf("benign CAB rejected: %v %q %v", clean, sig, err)
	}
	if clean, sig, err := e.Scan(context.Background(), append(append([]byte(nil), base...), []byte("unclassified trailer")...)); clean || sig != "" || !errors.Is(err, ErrUnsupported) {
		t.Fatalf("CAB trailer was treated as complete: %v %q %v", clean, sig, err)
	}
	s := &scanState{ctx: context.Background(), db: e.current.Load().db, maxFile: int64(len(base)), remaining: int64(len(base) + 1), maxFiles: 10, maxDepth: 8}
	if sig, err := s.visit(base, 0); sig != "" || !errors.Is(err, ErrLimit) {
		t.Fatalf("CAB expansion exceeded remaining budget: %q %v", sig, err)
	}
}
