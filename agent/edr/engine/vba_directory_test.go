package engine

import (
	"context"
	"encoding/binary"
	"errors"
	"testing"
	"unicode/utf16"
)

func vbaDirectoryFixture(offset uint32) []byte {
	var b []byte
	record := func(id uint16, p []byte) {
		b = binary.LittleEndian.AppendUint16(b, id)
		b = binary.LittleEndian.AppendUint32(b, uint32(len(p)))
		b = append(b, p...)
	}
	for _, id := range []uint16{1, 2, 0x14} {
		record(id, make([]byte, 4))
	}
	record(3, []byte{0xe4, 4})
	record(4, []byte("Project"))
	record(5, nil)
	record(0x40, nil)
	record(6, nil)
	record(0x3d, nil)
	record(7, make([]byte, 4))
	record(8, make([]byte, 4))
	record(9, make([]byte, 4))
	b = append(b, 0, 0)
	record(0xf, []byte{1, 0})
	record(0x13, make([]byte, 2))
	record(0x19, []byte("Module1"))
	record(0x1a, []byte("Module1"))
	var name []byte
	for _, v := range utf16.Encode([]rune("Module1")) {
		name = binary.LittleEndian.AppendUint16(name, v)
	}
	record(0x32, name)
	record(0x1c, nil)
	record(0x48, nil)
	record(0x31, binary.LittleEndian.AppendUint32(nil, offset))
	record(0x1e, make([]byte, 4))
	record(0x2c, make([]byte, 2))
	record(0x21, nil)
	record(0x2b, nil)
	record(0x10, nil)
	return b
}

func TestVBADirectoryBoundsAndOffsets(t *testing.T) {
	b := vbaDirectoryFixture(25)
	p, err := parseVBADirectory(context.Background(), b, 10)
	if err != nil || len(p.modules) != 1 || p.modules["MODULE1"] != 25 {
		t.Fatal(p, err)
	}
	for i := 0; i < len(b); i++ {
		if _, err := parseVBADirectory(context.Background(), b[:i], 10); err == nil {
			t.Fatalf("accepted truncated directory at %d", i)
		}
	}
	if _, err := parseVBADirectory(context.Background(), b, 0); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := parseVBADirectory(ctx, b, 10); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
}
func TestOLEDirectoryWithoutVBAProjectIsNoProject(t *testing.T) {
	for _, source := range []string{"harmless source", oleTestMarker} {
		payload := append([]byte{0, 0, 0}, vbaLiteralFixture([]byte(source))...)
		b := oleFixture(t, 3, "Module1", payload, true, true, false)
		dir := vbaLiteralFixture(vbaDirectoryFixture(3))
		if len(payload) > 64 || len(dir) > 448 {
			t.Fatal("fixture overflow")
		}
		// Extend the existing root ministream within its one allocated sector.
		binary.LittleEndian.PutUint64(b[2*512+120:], 512)
		binary.LittleEndian.PutUint32(b[2*512+2*128+72:], 3)
		e := b[2*512+3*128 : 3*512]
		for i, c := range "dir" {
			binary.LittleEndian.PutUint16(e[i*2:], uint16(c))
		}
		binary.LittleEndian.PutUint16(e[64:], 8)
		e[66] = 2
		e[67] = 1
		for _, off := range []int{68, 72, 76} {
			binary.LittleEndian.PutUint32(e[off:], oleFree)
		}
		binary.LittleEndian.PutUint32(e[116:], 1)
		binary.LittleEndian.PutUint64(e[120:], uint64(len(dir)))
		copy(b[3*512+64:], dir)
		for i := 1; i <= (len(dir)+63)/64; i++ {
			next := uint32(i + 1)
			if i == (len(dir)+63)/64 {
				next = oleEnd
			}
			binary.LittleEndian.PutUint32(b[4*512+i*4:], next)
		}
		s := oleTestState(t)
		sig, err := s.ole(b, 0)
		// Without a _VBA_PROJECT stream the reference reads no project, so
		// the compressed source is not matched (vbaproject_test.go covers
		// projects).
		if sig != "" || err != nil {
			t.Fatal(source, sig, err)
		}
	}
}
func FuzzVBADirectory(f *testing.F) {
	f.Add(vbaDirectoryFixture(10))
	f.Add([]byte{})
	f.Fuzz(func(t *testing.T, b []byte) {
		if len(b) > 1<<20 {
			return
		}
		_, _ = parseVBADirectory(context.Background(), b, 100)
	})
}
