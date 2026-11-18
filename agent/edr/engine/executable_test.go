package engine

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/md5"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/utmstack/UTMStack/agent/edr/config"
)

func peValidationFixture() []byte {
	b := make([]byte, 1024)
	copy(b, "MZ")
	o := binary.LittleEndian
	o.PutUint32(b[0x3c:], 0x80)
	copy(b[0x80:], []byte{'P', 'E', 0, 0})
	o.PutUint16(b[0x84:], 0x14c)
	o.PutUint16(b[0x86:], 1)
	o.PutUint16(b[0x94:], 224)
	o.PutUint16(b[0x98:], 0x10b)
	o.PutUint32(b[0x98+56:], 0x2000)
	o.PutUint32(b[0x98+60:], 512)
	o.PutUint32(b[0x98+92:], 16)
	h := b[0x178:]
	copy(h, ".text")
	o.PutUint32(h[8:], 512)
	o.PutUint32(h[12:], 0x1000)
	o.PutUint32(h[16:], 512)
	o.PutUint32(h[20:], 512)
	return b
}

func TestBoundedPEValidation(t *testing.T) {
	data := append(peValidationFixture(), []byte("overlay")...)
	if offset, err := validatePE(data); err != nil || offset != 1024 {
		t.Fatalf("offset=%d err=%v", offset, err)
	}
	for name, mutate := range map[string]func([]byte){
		"header offset":         func(b []byte) { binary.LittleEndian.PutUint32(b[0x3c:], ^uint32(0)) },
		"section extent":        func(b []byte) { binary.LittleEndian.PutUint32(b[0x178+20:], ^uint32(0)) },
		"short optional header": func(b []byte) { binary.LittleEndian.PutUint16(b[0x94:], 2) },
		"short headers":         func(b []byte) { binary.LittleEndian.PutUint32(b[0x98+60:], 1) },
	} {
		t.Run(name, func(t *testing.T) {
			b := peValidationFixture()
			mutate(b)
			if _, err := validatePE(b); !errors.Is(err, ErrMalformed) {
				t.Fatal(err)
			}
		})
	}
	// The loader ignores COFF symbol-table fields of an image; Upack stores
	// code in them.
	b := peValidationFixture()
	binary.LittleEndian.PutUint32(b[0x84+8:], 512)
	binary.LittleEndian.PutUint32(b[0x84+12:], ^uint32(0))
	if _, err := validatePE(b); err != nil {
		t.Fatal(err)
	}
	b = peValidationFixture()
	copy(b[0x178:0x180], []byte("UPX0\x00\x00\x00\x00"))
	if _, err := validatePE(b); !errors.Is(err, ErrUnsupported) {
		t.Fatal(err)
	}
	b = peValidationFixture()
	binary.LittleEndian.PutUint16(b[0x86:], maxExecutableSections+1)
	if _, err := validatePE(b); !errors.Is(err, ErrMalformed) {
		t.Fatal(err)
	}
}

func peManySectionsValidationFixture(count int) []byte {
	base := peValidationFixture()
	headers := (376 + count*40 + 511) &^ 511
	rva := (headers + 4095) &^ 4095
	b := make([]byte, headers+512)
	copy(b[:416], base[:416])
	copy(b[headers:], base[512:1024])
	binary.LittleEndian.PutUint16(b[0x86:], uint16(count))
	binary.LittleEndian.PutUint32(b[0x98+56:], uint32(rva+4096))
	binary.LittleEndian.PutUint32(b[0x98+60:], uint32(headers))
	binary.LittleEndian.PutUint32(b[0x178+12:], uint32(rva))
	binary.LittleEndian.PutUint32(b[0x178+20:], uint32(headers))
	return b
}

func TestPEValidationAcceptsFullCOFFSectionCount(t *testing.T) {
	for _, count := range []int{4097, 8192, 65535} {
		t.Run(strconv.Itoa(count), func(t *testing.T) {
			b := peManySectionsValidationFixture(count)
			if overlay, err := validatePE(b); err != nil || overlay != len(b) {
				t.Fatalf("overlay=%d err=%v", overlay, err)
			}
		})
	}
}

func TestLargeSectionPEOverlayArchiveScanned(t *testing.T) {
	marker := []byte("UTMStack inert large-section PE overlay marker")
	cfg := config.Default()
	cfg.SignatureDBDir = t.TempDir()
	rule := fmt.Sprintf("%x:%d:Overlay.Marker\n", md5.Sum(marker), len(marker))
	if err := os.WriteFile(filepath.Join(cfg.SignatureDBDir, "probe.hdb"), []byte(rule), 0o600); err != nil {
		t.Fatal(err)
	}
	var archive bytes.Buffer
	zw := zip.NewWriter(&archive)
	w, err := zw.Create("marker.txt")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := w.Write(marker); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	scanner := New(cfg)
	if err := scanner.EnsureRunning(); err != nil {
		t.Fatal(err)
	}
	pe := append(peManySectionsValidationFixture(4097), archive.Bytes()...)
	clean, name, _, err := scanner.ScanSnapshot(context.Background(), pe)
	if err != nil || clean || name != "Overlay.Marker" {
		t.Fatalf("clean=%t name=%q err=%v", clean, name, err)
	}
}

func elfValidationFixture(class byte, order binary.ByteOrder) []byte {
	b := make([]byte, 768)
	copy(b, []byte{0x7f, 'E', 'L', 'F', class, 1, 1})
	if order == binary.BigEndian {
		b[5] = 2
	}
	order.PutUint16(b[16:], 2)
	order.PutUint16(b[18:], 3)
	order.PutUint32(b[20:], 1)
	if class == 1 {
		order.PutUint32(b[28:], 52)
		order.PutUint16(b[40:], 52)
		order.PutUint16(b[42:], 32)
		order.PutUint16(b[44:], 1)
		order.PutUint32(b[52:], 1)
		order.PutUint32(b[52+16:], uint32(len(b)))
		order.PutUint32(b[52+20:], uint32(len(b)))
		order.PutUint32(b[52+28:], 4096)
	} else {
		order.PutUint64(b[32:], 64)
		order.PutUint16(b[52:], 64)
		order.PutUint16(b[54:], 56)
		order.PutUint16(b[56:], 1)
		order.PutUint32(b[64:], 1)
		order.PutUint64(b[64+32:], uint64(len(b)))
		order.PutUint64(b[64+40:], uint64(len(b)))
		order.PutUint64(b[64+48:], 4096)
	}
	return b
}

func TestBoundedELFValidation(t *testing.T) {
	for _, class := range []byte{1, 2} {
		for _, order := range []binary.ByteOrder{binary.LittleEndian, binary.BigEndian} {
			if err := validateELF(elfValidationFixture(class, order)); err != nil {
				t.Fatalf("class=%d byteOrder=%s: %v", class, order, err)
			}
		}
	}
	for name, mutate := range map[string]func([]byte){
		"program offset": func(b []byte) { binary.LittleEndian.PutUint64(b[32:], ^uint64(0)) },
		"segment extent": func(b []byte) { binary.LittleEndian.PutUint64(b[64+32:], ^uint64(0)) },
		"segment memory": func(b []byte) { binary.LittleEndian.PutUint64(b[64+40:], 1) },
		"alignment":      func(b []byte) { binary.LittleEndian.PutUint64(b[64+48:], 3) },
		"entry size":     func(b []byte) { binary.LittleEndian.PutUint16(b[54:], 1) },
	} {
		t.Run(name, func(t *testing.T) {
			b := elfValidationFixture(2, binary.LittleEndian)
			mutate(b)
			if err := validateELF(b); !errors.Is(err, ErrMalformed) {
				t.Fatal(err)
			}
		})
	}
	b := elfValidationFixture(2, binary.LittleEndian)
	binary.LittleEndian.PutUint16(b[56:], 0xffff)
	if err := validateELF(b); !errors.Is(err, ErrUnsupported) {
		t.Fatal(err)
	}
	b = elfValidationFixture(2, binary.LittleEndian)
	binary.LittleEndian.PutUint64(b[40:], 128)
	if err := validateELF(b); !errors.Is(err, ErrUnsupported) {
		t.Fatal(err)
	}
}

func elfWithSectionsFixture() []byte {
	b := elfValidationFixture(2, binary.LittleEndian)
	o := binary.LittleEndian
	o.PutUint64(b[40:], 0x180)
	o.PutUint16(b[58:], 64)
	o.PutUint16(b[60:], 3)
	o.PutUint16(b[62:], 2)
	copy(b[0x140:], []byte("\x00.text\x00.shstrtab\x00"))
	h := b[0x1c0:]
	o.PutUint32(h, 1)
	o.PutUint32(h[4:], 1)
	o.PutUint64(h[24:], 0x100)
	o.PutUint64(h[32:], 0x20)
	h = b[0x200:]
	o.PutUint32(h, 7)
	o.PutUint32(h[4:], 3)
	o.PutUint64(h[24:], 0x140)
	o.PutUint64(h[32:], 17)
	return b
}

func TestELFSectionsAreBoundedAndCompressedContentUnknown(t *testing.T) {
	b := elfWithSectionsFixture()
	if err := validateELF(b); err != nil {
		t.Fatal(err)
	}
	binary.LittleEndian.PutUint64(b[0x1c0+24:], ^uint64(0))
	if err := validateELF(b); !errors.Is(err, ErrMalformed) {
		t.Fatal(err)
	}
	b = elfWithSectionsFixture()
	binary.LittleEndian.PutUint64(b[0x1c0+8:], 0x800)
	if err := validateELF(b); !errors.Is(err, ErrUnsupported) {
		t.Fatal(err)
	}
	b = elfWithSectionsFixture()
	copy(b[0x141:], "UPX0\x00")
	if err := validateELF(b); !errors.Is(err, ErrUnsupported) {
		t.Fatal(err)
	}
}

func FuzzExecutableValidation(f *testing.F) {
	f.Add(peValidationFixture())
	f.Add(elfValidationFixture(1, binary.BigEndian))
	f.Add(elfWithSectionsFixture())
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 1<<20 {
			t.Skip()
		}
		_, _ = validatePE(data)
		if img, err := parsePEScanImage(data); err == nil {
			if img.overlay < 0 || img.overlay > len(data) {
				t.Fatal("PE scan overlay escaped physical input")
			}
			for _, section := range img.sections {
				if section.raw > section.rawEnd || section.rawEnd > uint64(len(data)) {
					t.Fatal("PE scan section escaped physical input")
				}
			}
		}
		_ = validateELF(data)
	})
}
