package engine

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"hash/crc32"
	"testing"

	"github.com/klauspost/compress/zstd"
	"github.com/ulikunitz/xz"
)

func rarStoredFixture(data []byte, encrypted bool) []byte {
	block := func(kind byte, flags uint16, payload []byte) []byte {
		h := []byte{0, 0, kind}
		h = binary.LittleEndian.AppendUint16(h, flags)
		h = binary.LittleEndian.AppendUint16(h, uint16(len(payload)+7))
		h = append(h, payload...)
		binary.LittleEndian.PutUint16(h, uint16(crc32.ChecksumIEEE(h[2:])))
		return h
	}
	b := append([]byte("Rar!\x1a\x07\x00"), block(0x73, 0, make([]byte, 6))...)
	f := binary.LittleEndian.AppendUint32(nil, uint32(len(data)))
	f = binary.LittleEndian.AppendUint32(f, uint32(len(data)))
	f = append(f, 2)
	f = binary.LittleEndian.AppendUint32(f, crc32.ChecksumIEEE(data))
	f = append(f, 0, 0, 0, 0, 20, 0x30, 1, 0, 0, 0, 0, 0, 'x')
	flags := uint16(0x8000)
	if encrypted {
		flags |= 4
	}
	b = append(b, block(0x74, flags, f)...)
	b = append(b, data...)
	return append(b, block(0x7b, 0, nil)...)
}
func TestNativeRARStoredAndNested(t *testing.T) {
	e := nativeTestEngine(t)
	inner := gzipFixture(t, []byte(testMarker))
	for _, data := range [][]byte{rarStoredFixture(inner, false), rarStoredFixture(rarStoredFixture(inner, false), false)} {
		if clean, sig, err := e.Scan(context.Background(), data); clean || sig != "Native.Test" || err != nil {
			t.Fatal(clean, sig, err)
		}
	}
	if clean, _, err := e.Scan(context.Background(), rarStoredFixture([]byte("benign"), false)); !clean || err != nil {
		t.Fatal(clean, err)
	}
	if clean, _, err := e.Scan(context.Background(), rarStoredFixture(inner, true)); clean || !errors.Is(err, ErrUnsupported) {
		t.Fatal(clean, err)
	}
	s := scanState{ctx: context.Background(), db: e.current.Load().db, maxFile: 100, remaining: 1, maxFiles: 10, maxDepth: 8}
	if _, err := s.rar(rarStoredFixture(inner, false), 0); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
}

func xzFixture(t testing.TB, data []byte) []byte {
	t.Helper()
	var out bytes.Buffer
	w, err := (xz.WriterConfig{DictCap: 1 << 16, BlockSize: 64}).NewWriter(&out)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = w.Write(data); err != nil {
		t.Fatal(err)
	}
	if err = w.Close(); err != nil {
		t.Fatal(err)
	}
	return out.Bytes()
}

func TestNativeXZAndZstandard(t *testing.T) {
	e := nativeTestEngine(t)
	w, err := zstd.NewWriter(nil, zstd.WithEncoderConcurrency(1))
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()
	for name, data := range map[string][]byte{
		"xz":                 xzFixture(t, []byte(testMarker)),
		"xz-multiple-blocks": xzFixture(t, []byte(string(bytes.Repeat([]byte("."), 100))+testMarker)),
		"xz-concatenated":    append(xzFixture(t, []byte("benign")), xzFixture(t, []byte(testMarker))...),
		"zstd":               w.EncodeAll([]byte(testMarker), nil),
		"zstd-nested":        w.EncodeAll(gzipFixture(t, []byte(testMarker)), nil),
	} {
		t.Run(name, func(t *testing.T) {
			clean, sig, err := e.Scan(context.Background(), data)
			if clean || sig != "Native.Test" || err != nil {
				t.Fatalf("%v %q %v", clean, sig, err)
			}
		})
	}
	for name, data := range map[string][]byte{"xz": xzFixture(t, []byte("benign")), "zstd": w.EncodeAll([]byte("benign"), nil)} {
		t.Run(name+"-valid", func(t *testing.T) {
			if clean, _, err := e.Scan(context.Background(), data); !clean || err != nil {
				t.Fatal(clean, err)
			}
			for _, bad := range [][]byte{data[:len(data)-1], append(append([]byte(nil), data...), 1)} {
				if clean, _, err := e.Scan(context.Background(), bad); clean || err == nil {
					t.Fatal("invalid input reported clean", clean, err)
				}
			}
		})
	}
}

func TestXZRejectsExpansionAndDictionaryBeforeDecode(t *testing.T) {
	b := xzFixture(t, bytes.Repeat([]byte("A"), 4096))
	if err := preflightXZ(b, 4095); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	if err := preflightXZ(b, 4096); err != nil {
		t.Fatal(err)
	}
	// Standard first block header: size, flags, LZMA2 ID, props length, dict.
	b[16] = 40 // 4GiB dictionary; no allocation or valid CRC required to refuse it.
	if err := preflightXZ(b, 4096); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
}

func zip64Fixture(t *testing.T, data []byte) []byte {
	b := zipFixture(t, data)
	end := len(b) - 22
	z := make([]byte, 56+20)
	binary.LittleEndian.PutUint32(z, 0x06064b50)
	binary.LittleEndian.PutUint64(z[4:], 44)
	binary.LittleEndian.PutUint16(z[12:], 45)
	binary.LittleEndian.PutUint16(z[14:], 45)
	binary.LittleEndian.PutUint64(z[24:], uint64(binary.LittleEndian.Uint16(b[end+8:])))
	binary.LittleEndian.PutUint64(z[32:], uint64(binary.LittleEndian.Uint16(b[end+10:])))
	binary.LittleEndian.PutUint64(z[40:], uint64(binary.LittleEndian.Uint32(b[end+12:])))
	binary.LittleEndian.PutUint64(z[48:], uint64(binary.LittleEndian.Uint32(b[end+16:])))
	binary.LittleEndian.PutUint32(z[56:], 0x07064b50)
	binary.LittleEndian.PutUint64(z[64:], uint64(end))
	binary.LittleEndian.PutUint32(z[72:], 1)
	binary.LittleEndian.PutUint16(b[end+8:], 0xffff)
	binary.LittleEndian.PutUint16(b[end+10:], 0xffff)
	binary.LittleEndian.PutUint32(b[end+12:], 0xffffffff)
	binary.LittleEndian.PutUint32(b[end+16:], 0xffffffff)
	out := append([]byte(nil), b[:end]...)
	out = append(out, z...)
	return append(out, b[end:]...)
}

// The pinned reference does not follow ZIP64 end records. This engine also
// resolves their bounded directory offsets so valid members are analyzed.
func TestNativeZIP64(t *testing.T) {
	e := nativeTestEngine(t)
	b := zip64Fixture(t, []byte(testMarker))
	if m, err := zipMembers(b, 10); err != nil || len(m) != 1 {
		t.Fatal(len(m), err)
	}
	if clean, sig, err := e.Scan(context.Background(), b); clean || sig != "Native.Test" || err != nil {
		t.Fatal(clean, sig, err)
	}
}

func FuzzCompressedPreflight(f *testing.F) {
	f.Add(xzFixture(f, []byte("benign test")))
	f.Add([]byte("PK\x05\x06"))
	f.Fuzz(func(t *testing.T, b []byte) {
		if len(b) > 1<<20 {
			return
		}
		_, _ = zipMembers(b, 100)
		_ = preflightXZ(b, 1<<20)
	})
}
