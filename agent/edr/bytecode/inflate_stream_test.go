package bytecode

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"compress/zlib"
	"encoding/binary"
	"errors"
	"hash/crc32"
	"io"
	"math/rand"
	"testing"
)

func packedForInflate(t *testing.T, raw []byte, level int, wrapped bool) []byte {
	t.Helper()
	var buf bytes.Buffer
	if wrapped {
		w, err := zlib.NewWriterLevel(&buf, level)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := w.Write(raw); err != nil {
			t.Fatal(err)
		}
		if err := w.Close(); err != nil {
			t.Fatal(err)
		}
	} else {
		w, err := flate.NewWriter(&buf, level)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := w.Write(raw); err != nil {
			t.Fatal(err)
		}
		if err := w.Close(); err != nil {
			t.Fatal(err)
		}
	}
	return buf.Bytes()
}

func decodeInflateChunks(t *testing.T, bits int32, packed []byte, inputChunk, outputChunk int) ([]byte, int, int, error) {
	t.Helper()
	d, ok := newStreamInflater(bits)
	if !ok {
		t.Fatal("valid window bits rejected")
	}
	var output []byte
	consumed := 0
	for step := 0; step < 200000; step++ {
		end := min(len(packed), consumed+inputChunk)
		buf := make([]byte, outputChunk)
		read, wrote, status, err := d.process(packed[consumed:end], buf)
		consumed += read
		output = append(output, buf[:wrote]...)
		if err != nil || status == inflateStreamEnd {
			return output, consumed, status, err
		}
		if read == 0 && wrote == 0 && consumed == len(packed) {
			return output, consumed, status, nil
		}
	}
	t.Fatal("decoder did not converge")
	return nil, 0, 0, nil
}

func TestStreamInflaterRawAndZlibAcrossWindows(t *testing.T) {
	inputs := [][]byte{
		[]byte("hello"),
		bytes.Repeat([]byte("AAAAABBBBBCCCCCDDDDDEEEEE1234567890"), 20),
		bytes.Repeat([]byte("abcXYZ789!"), 700),
	}
	for _, wrapped := range []bool{false, true} {
		for _, level := range []int{flate.NoCompression, flate.BestSpeed, flate.DefaultCompression, flate.HuffmanOnly} {
			for _, raw := range inputs {
				packed := packedForInflate(t, raw, level, wrapped)
				windowBits := int32(-15)
				if wrapped {
					windowBits = 15
				}
				for _, chunk := range [][2]int{{1, 1}, {3, 7}, {17, 23}, {8192, 8192}} {
					got, used, status, err := decodeInflateChunks(t, windowBits, packed, chunk[0], chunk[1])
					if err != nil || status != inflateStreamEnd || used != len(packed) || !bytes.Equal(got, raw) {
						t.Fatalf("wrapped=%t level=%d raw=%d chunks=%v: output=%d input=%d/%d status=%d err=%v", wrapped, level, len(raw), chunk, len(got), used, len(packed), status, err)
					}
				}
			}
		}
	}
}

func TestStreamInflaterTrailerAndPartialInput(t *testing.T) {
	raw := []byte("bounded streaming input")
	packed := packedForInflate(t, raw, flate.DefaultCompression, true)
	withTail := append(bytes.Clone(packed), []byte("TAIL")...)
	got, used, status, err := decodeInflateChunks(t, 0, withTail, 3, 5)
	if err != nil || status != inflateStreamEnd || used != len(packed) || !bytes.Equal(got, raw) {
		t.Fatalf("trailing bytes: output=%q input=%d/%d status=%d err=%v", got, used, len(packed), status, err)
	}
	broken := bytes.Clone(packed)
	broken[len(broken)-1] ^= 1
	_, _, _, err = decodeInflateChunks(t, 15, broken, 2, 4)
	if !errors.Is(err, errInflateData) {
		t.Fatalf("bad Adler checksum = %v", err)
	}
	_, used, status, err = decodeInflateChunks(t, 15, packed[:len(packed)-2], 1, 3)
	if err != nil || used != len(packed)-2 || status != inflateBufError {
		t.Fatalf("partial stream: input=%d status=%d err=%v", used, status, err)
	}
	for _, bits := range []int32{-7, -16, 1, 7, 16, 23, 48} {
		if _, ok := newStreamInflater(bits); ok {
			t.Fatalf("invalid windowBits %d accepted", bits)
		}
	}
}

func TestStreamInflaterGzipAndAutoHeaders(t *testing.T) {
	raw := bytes.Repeat([]byte("gzip header and distance history "), 80)
	var ordinary bytes.Buffer
	w := gzip.NewWriter(&ordinary)
	if _, err := w.Write(raw); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	for _, bits := range []int32{31, 47} {
		for _, chunk := range [][2]int{{1, 1}, {7, 11}, {8192, 8192}} {
			got, used, status, err := decodeInflateChunks(t, bits, ordinary.Bytes(), chunk[0], chunk[1])
			if err != nil || status != inflateStreamEnd || used != ordinary.Len() || !bytes.Equal(got, raw) {
				t.Fatalf("gzip bits=%d chunk=%v: output=%d used=%d/%d status=%d err=%v", bits, chunk, len(got), used, ordinary.Len(), status, err)
			}
		}
	}
	zlibData := packedForInflate(t, raw, flate.DefaultCompression, true)
	got, used, status, err := decodeInflateChunks(t, 47, zlibData, 1, 3)
	if err != nil || status != inflateStreamEnd || used != len(zlibData) || !bytes.Equal(got, raw) {
		t.Fatalf("auto zlib: output=%d used=%d/%d status=%d err=%v", len(got), used, len(zlibData), status, err)
	}
	// Construct all RFC 1952 optional header fields, including FHCRC, from
	// standard-library raw DEFLATE output. The header checksum is CRC32's
	// low 16 bits and the trailer uses little-endian CRC32 and ISIZE.
	header := []byte{0x1f, 0x8b, 8, 0x1e, 0, 0, 0, 0, 0, 255, 3, 0, 'x', 'y', 'z', 'n', 'a', 'm', 'e', 0, 'c', 'o', 'm', 'm', 'e', 'n', 't', 0}
	var checksum [2]byte
	binary.LittleEndian.PutUint16(checksum[:], uint16(crc32.ChecksumIEEE(header)))
	withOptions := append(bytes.Clone(header), checksum[:]...)
	withOptions = append(withOptions, packedForInflate(t, raw, flate.DefaultCompression, false)...)
	var trailer [8]byte
	binary.LittleEndian.PutUint32(trailer[:4], crc32.ChecksumIEEE(raw))
	binary.LittleEndian.PutUint32(trailer[4:], uint32(len(raw)))
	withOptions = append(withOptions, trailer[:]...)
	got, used, status, err = decodeInflateChunks(t, 31, withOptions, 1, 2)
	if err != nil || status != inflateStreamEnd || used != len(withOptions) || !bytes.Equal(got, raw) {
		t.Fatalf("optional gzip header: output=%d used=%d/%d status=%d err=%v", len(got), used, len(withOptions), status, err)
	}
	broken := bytes.Clone(withOptions)
	broken[len(header)] ^= 1
	_, _, _, err = decodeInflateChunks(t, 31, broken, 1, 2)
	if !errors.Is(err, errInflateData) {
		t.Fatalf("bad gzip FHCRC = %v", err)
	}
	broken = bytes.Clone(withOptions)
	broken[len(broken)-8] ^= 1
	_, _, _, err = decodeInflateChunks(t, 31, broken, 7, 23)
	if !errors.Is(err, errInflateData) {
		t.Fatalf("bad gzip trailer = %v", err)
	}
}

func TestStreamInflaterRejectsInvalidHuffmanAlphabets(t *testing.T) {
	for _, tc := range []struct {
		name        string
		lengths     []uint8
		allowEmpty  bool
		allowSingle bool
		valid       bool
	}{
		{"complete", []uint8{1, 1}, false, false, true},
		{"single literal", []uint8{1, 0}, false, true, true},
		{"single code length", []uint8{1, 0}, false, false, false},
		{"incomplete deep", []uint8{2, 2}, false, true, false},
		{"oversubscribed", []uint8{1, 1, 1}, false, true, false},
		{"empty distance", []uint8{0, 0}, true, true, true},
	} {
		var tree huffmanTree
		err := tree.build(tc.lengths, tc.allowEmpty, tc.allowSingle)
		if (err == nil) != tc.valid {
			t.Fatalf("%s: err=%v, want valid=%t", tc.name, err, tc.valid)
		}
	}
	// A final dynamic block that advertises 287 literal codes is rejected
	// before the decoder reads a potentially hostile alphabet.
	_, _, _, err := decodeInflateChunks(t, -15, []byte{0xfd, 0x00, 0x00}, 1, 1)
	if !errors.Is(err, errInflateData) {
		t.Fatalf("excess literal alphabet: %v", err)
	}
}

func TestStreamInflaterGeneratedCorpusAcrossSmallWindows(t *testing.T) {
	rng := rand.New(rand.NewSource(185))
	for i := 0; i < 128; i++ {
		raw := make([]byte, rng.Intn(4096))
		if _, err := rng.Read(raw); err != nil {
			t.Fatal(err)
		}
		if i%3 == 0 {
			copy(raw, bytes.Repeat([]byte("repeat"), len(raw)/6))
		}
		wrapped := i%2 == 0
		level := []int{flate.NoCompression, flate.BestSpeed, flate.DefaultCompression, flate.BestCompression, flate.HuffmanOnly}[i%5]
		packed := packedForInflate(t, raw, level, wrapped)
		bits := int32(-15)
		if wrapped {
			bits = 15
		}
		inputChunk, outputChunk := 1+rng.Intn(37), 1+rng.Intn(41)
		got, used, status, err := decodeInflateChunks(t, bits, packed, inputChunk, outputChunk)
		if err != nil || status != inflateStreamEnd || used != len(packed) || !bytes.Equal(got, raw) {
			t.Fatalf("case=%d input=%d output=%d status=%d err=%v", i, used, len(got), status, err)
		}
	}
}

func TestStreamInflaterMutationsNeverBecomeFalseValid(t *testing.T) {
	rng := rand.New(rand.NewSource(1852026))
	raw := bytes.Repeat([]byte("DEFLATE block and checksum probe "), 64)
	for _, wrap := range []string{"raw", "zlib", "gzip"} {
		var packed []byte
		bits := int32(-15)
		switch wrap {
		case "raw":
			packed = packedForInflate(t, raw, flate.DefaultCompression, false)
		case "zlib":
			packed = packedForInflate(t, raw, flate.DefaultCompression, true)
			bits = 15
		case "gzip":
			var buf bytes.Buffer
			writer := gzip.NewWriter(&buf)
			if _, err := writer.Write(raw); err != nil {
				t.Fatal(err)
			}
			if err := writer.Close(); err != nil {
				t.Fatal(err)
			}
			packed, bits = buf.Bytes(), 31
		}
		for trial := 0; trial < 128; trial++ {
			mutated := bytes.Clone(packed)
			at := rng.Intn(len(mutated))
			mutated[at] ^= byte(1 << uint(rng.Intn(8)))
			got, _, status, err := decodeInflateChunks(t, bits, mutated, 1+rng.Intn(19), 1+rng.Intn(31))
			if err != nil || status != inflateStreamEnd {
				continue
			}
			var reader io.ReadCloser
			switch wrap {
			case "raw":
				reader = flate.NewReader(bytes.NewReader(mutated))
			case "zlib":
				reader, err = zlib.NewReader(bytes.NewReader(mutated))
			case "gzip":
				reader, err = gzip.NewReader(bytes.NewReader(mutated))
			}
			if err != nil {
				t.Fatalf("%s mutation %d accepted despite standard reader rejecting header: %v", wrap, trial, err)
			}
			want, readErr := io.ReadAll(reader)
			closeErr := reader.Close()
			if readErr != nil || closeErr != nil || !bytes.Equal(got, want) {
				t.Fatalf("%s mutation %d mismatch: output=%d/%d read=%v close=%v", wrap, trial, len(got), len(want), readErr, closeErr)
			}
		}
	}
}
