package engine

import (
	"bytes"
	"compress/flate"
	"compress/zlib"
	"context"
	"crypto/md5"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"strconv"
	"testing"

	"github.com/ulikunitz/xz/lzma"
	"github.com/utmstack/UTMStack/agent/edr/signatures"
)

const payloadMarker = "EXTRACTED-HARMLESS-PAYLOAD-MARKER"

// payloadState detects the marker by body, or whole objects by MD5 when
// provided, which proves that an embedded object was extracted with exact bounds.
func payloadState(t *testing.T, hashed ...[]byte) *scanState {
	t.Helper()
	files := map[string][]byte{}
	if len(hashed) == 0 {
		files["test.ndb"] = []byte("Payload.Test:0:*:" + hex.EncodeToString([]byte(payloadMarker)) + "\n")
	}
	var hdb bytes.Buffer
	for i, b := range hashed {
		sum := md5.Sum(b)
		fmt.Fprintf(&hdb, "%x:%d:Payload.Hash%d\n", sum, len(b), i)
	}
	if hdb.Len() > 0 {
		files["test.hdb"] = hdb.Bytes()
	}
	db, err := signatures.Compile(files)
	if err != nil {
		t.Fatal(err)
	}
	return &scanState{ctx: context.Background(), db: db, maxFile: 16 << 20, remaining: 64 << 20, maxFiles: 1000, maxDepth: 12}
}

type peFixtureSection struct {
	name       string
	va, vsize  uint32
	rawPointer uint32 // 0: placed after the previous section
	raw        []byte
}

type peFixture struct {
	lfanew    uint32
	headers   uint32 // SizeOfHeaders; 0: 0x200
	imageSize uint32 // 0: derived
	entry     uint32
	optional  uint16 // SizeOfOptionalHeader; 0: 0xe0
	sections  []peFixtureSection
	overlay   []byte
	cert      []byte // appended after the overlay and named by the security directory
	symbols   [2]uint32
	fileAlign uint32
}

// build writes a PE32 image with the given layout. Section data is placed at
// 0x200-aligned offsets unless a raw pointer is given.
func (f peFixture) build() []byte {
	o := binary.LittleEndian
	if f.lfanew == 0 {
		f.lfanew = 0x80
	}
	if f.optional == 0 {
		f.optional = 0xe0
	}
	if f.headers == 0 {
		f.headers = 0x200
	}
	if f.fileAlign == 0 {
		f.fileAlign = 0x200
	}
	// The header bytes written are 0x200 long whatever SizeOfHeaders says.
	b := make([]byte, max(0x200, int(f.lfanew)+24+int(f.optional)+40*len(f.sections)))
	copy(b, "MZ")
	o.PutUint32(b[0x3c:], f.lfanew)
	h := b[f.lfanew:]
	copy(h, "PE\x00\x00")
	o.PutUint16(h[4:], 0x14c)
	o.PutUint16(h[6:], uint16(len(f.sections)))
	o.PutUint32(h[12:], f.symbols[0])
	o.PutUint32(h[16:], f.symbols[1])
	o.PutUint16(h[20:], f.optional)
	opt := h[24:]
	o.PutUint16(opt, 0x10b)
	o.PutUint32(opt[16:], f.entry)
	o.PutUint32(opt[28:], 0x400000)
	o.PutUint32(opt[32:], 0x1000)
	o.PutUint32(opt[36:], f.fileAlign)
	o.PutUint32(opt[60:], f.headers)
	o.PutUint32(opt[92:], 16)
	table := int(f.lfanew) + 24 + int(f.optional)
	imageSize := uint32(0x1000)
	pos := len(b)
	for i, s := range f.sections {
		t := b[table+40*i:]
		copy(t[:8], s.name)
		o.PutUint32(t[8:], s.vsize)
		o.PutUint32(t[12:], s.va)
		o.PutUint32(t[16:], uint32(len(s.raw)))
		imageSize = max(imageSize, s.va+max(s.vsize, 0x1000))
		if len(s.raw) == 0 {
			continue
		}
		pointer := s.rawPointer
		if pointer == 0 {
			pos = (pos + 0x1ff) &^ 0x1ff
			pointer = uint32(pos)
		}
		o.PutUint32(t[20:], pointer)
		if need := int(pointer) + len(s.raw); need > len(b) {
			b = append(b, make([]byte, need-len(b))...)
		}
		copy(b[pointer:], s.raw)
		pos = max(pos, int(pointer)+len(s.raw))
	}
	if f.imageSize != 0 {
		imageSize = f.imageSize
	}
	// The headers buffer may have grown; rewrite fields that live in it.
	o.PutUint32(b[int(f.lfanew)+24+56:], imageSize)
	b = append(b, f.overlay...)
	if f.cert != nil {
		dir := b[int(f.lfanew)+24+96+32:]
		o.PutUint32(dir, uint32(len(b)))
		o.PutUint32(dir[4:], uint32(len(f.cert)))
		b = append(b, f.cert...)
	}
	return b
}

func simplePE(code []byte) peFixture {
	return peFixture{entry: 0x1000, sections: []peFixtureSection{{name: ".text", va: 0x1000, vsize: 0x1000, raw: code}}}
}

func TestPELoaderLayoutsAreNotMalformed(t *testing.T) {
	code := bytes.Repeat([]byte{0x90}, 0x200)
	cases := map[string]struct {
		f      peFixture
		packed bool
	}{
		"plain": {simplePE(code), false},
		// Packers put the PE header inside the DOS header.
		"header inside DOS header": {func() peFixture { f := simplePE(code); f.lfanew = 0x10; return f }(), true},
		// Section data inside the headers is loadable and a packer layout.
		"section data in headers": {func() peFixture {
			f := simplePE(code[:0x40])
			f.sections[0].rawPointer = 0x1c0
			f.headers = 0x400
			return f
		}(), true},
		// SizeOfHeaders beyond a small file is loadable.
		"headers beyond file": {func() peFixture {
			f := simplePE(code)
			f.headers = 0x1000
			f.sections[0].rawPointer = 0x200
			return f
		}(), true},
		// The loader ignores COFF symbol-table fields of an image.
		"garbage symbol table": {func() peFixture { f := simplePE(code); f.symbols = [2]uint32{0xff4d0000, 0x7cea0076}; return f }(), false},
		// Raw data beyond the virtual size up to the section-aligned image.
		"raw beyond virtual size": {func() peFixture {
			f := simplePE(bytes.Repeat([]byte{0x90}, 0x3400))
			f.sections[0].vsize = 0x32c3
			f.imageSize = 0x1000 + 0x32c3
			return f
		}(), false},
		// The loader rounds raw pointers down to 512 bytes.
		"unaligned raw pointer": {func() peFixture { f := simplePE(code); f.sections[0].rawPointer = 0x201; return f }(), false},
		// A virtual extent past SizeOfImage reads nothing beyond the file.
		"virtual extent past the image": {func() peFixture { f := simplePE(code); f.imageSize = 0x1000; return f }(), false},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			b := c.f.build()
			img, err := parsePEImage(b)
			if err != nil {
				t.Fatalf("err=%v", err)
			}
			if img.packed != c.packed || img.overlay > len(b) {
				t.Fatalf("packed=%v overlay=%d len=%d", img.packed, img.overlay, len(b))
			}
			_, err = validatePE(b)
			if c.packed && !errors.Is(err, ErrUnsupported) || !c.packed && err != nil {
				t.Fatalf("validatePE err=%v", err)
			}
		})
	}
	for name, f := range map[string]peFixture{
		"section data past the end": func() peFixture { f := simplePE(code); f.sections[0].rawPointer = 0x10000; return f }(),
	} {
		t.Run(name, func(t *testing.T) {
			b := f.build()
			if name == "section data past the end" {
				b = b[:0x400]
			}
			if _, err := parsePEImage(b); !errors.Is(err, ErrMalformed) {
				t.Fatalf("err=%v", err)
			}
		})
	}
}

func TestPEPackerStubsAreUnsupported(t *testing.T) {
	stubs := map[string][]byte{
		"UPX":     {0x60, 0xbe, 0, 0x60, 0x40, 0, 0x8d, 0xbe, 0, 0xb0, 0xff, 0xff, 0x57, 0x83, 0xcd, 0xff},
		"FSG 2":   {0x87, 0x25, 1, 2, 3, 4, 0x61, 0x94, 0x55, 0xa4, 0xb6, 0x80, 0xff, 0x13},
		"FSG 1.3": {0xbb, 1, 2, 3, 4, 0xbf, 1, 2, 3, 4, 0xbe, 1, 2, 3, 4, 0x53, 0xe8, 0x0a, 0, 0, 0, 0x02, 0xd2, 0x75, 0x05, 0x8a, 0x16, 0x46, 0x12, 0xd2, 0xc3},
		"PESpin":  {0xeb, 0x01, 0x68, 0x60, 0xe8, 0, 0, 0, 0, 0x8b, 0x1c, 0x24, 0x83, 0xc3},
		"yC":      {0x55, 0x8b, 0xec, 0x53, 0x56, 0x57, 0x60, 0xe8, 0, 0, 0, 0, 0x5d, 0x81, 0xed},
		"WWPack":  append([]byte{0x53, 0x55, 0x8b, 0xe8, 0x33, 0xdb, 0xeb, 0x60, 0x0d, 0x0a, 0x0d, 0x0a}, "WWPack32"...),
		"ASPack":  {0x60, 0xe8, 0x03, 0, 0, 0, 0xe9, 0xeb, 0x04, 0x5d, 0x45, 0x55, 0xc3, 0xe8, 0x01},
		"PEtite":  {0xb8, 0, 0x50, 0x40, 0, 0x68, 0x23, 0x21, 0x40, 0, 0x64, 0xff, 0x35, 0, 0, 0, 0, 0x64, 0x89, 0x25, 0, 0, 0, 0},
		"Upack":   {0xbe, 0xb0, 0x11, 0x40, 0, 0xad, 0x50, 0xff, 0x76, 0x34, 0xeb, 0x7c},
	}
	for name, stub := range stubs {
		t.Run(name, func(t *testing.T) {
			code := append(append([]byte(nil), stub...), bytes.Repeat([]byte{0x90}, 0x200-len(stub))...)
			if _, err := validatePE(simplePE(code).build()); !errors.Is(err, ErrUnsupported) {
				t.Fatalf("err=%v", err)
			}
		})
	}
	for _, name := range []string{"MEW\x00F\x12\xd2\xc3", ".WWP32", "yC", ".nsp1"} {
		f := simplePE(bytes.Repeat([]byte{0x90}, 0x200))
		f.sections[0].name = name
		if _, err := validatePE(f.build()); !errors.Is(err, ErrUnsupported) {
			t.Fatalf("section %q: err=%v", name, err)
		}
	}
	// Ordinary code at the entry point is not a packer.
	for _, near := range [][]byte{{0x60, 0xe8, 0x00}, {0x55, 0x8b, 0xec, 0x83, 0xec, 0x10}} {
		code := append(append([]byte(nil), near...), bytes.Repeat([]byte{0xc3}, 0x200-len(near))...)
		if _, err := validatePE(simplePE(code).build()); err != nil {
			t.Fatalf("%x: err=%v", near, err)
		}
	}
}

// A packed stub's overlay is still analyzed, as self-extractors often are.
func TestPackedExecutableOverlayIsStillAnalyzed(t *testing.T) {
	f := simplePE(bytes.Repeat([]byte{0x90}, 0x200))
	f.sections[0].name = "UPX1"
	f.overlay = zipFixture(t, gzipFixture(t, []byte(payloadMarker)))
	if sig, err := payloadState(t).visit(f.build(), 0); sig != "Payload.Test" || err != nil {
		t.Fatalf("sig=%q err=%v", sig, err)
	}
}

func TestPEOverlayArchiveNeedsMoreThanMagic(t *testing.T) {
	f := simplePE(bytes.Repeat([]byte{0x90}, 0x200))
	// Installer data holding a chance gzip magic is not a gzip stream.
	f.overlay = append([]byte("setup data \x1f\x8b\x42\x99 more data "), bytes.Repeat([]byte{0xa5}, 64)...)
	if sig, err := payloadState(t).visit(f.build(), 0); sig != "" || err != nil {
		t.Fatalf("chance magic: sig=%q err=%v", sig, err)
	}
	// A real gzip member after other data is found and analyzed.
	f.overlay = append([]byte("setup data \x1f\x8b\x42\x99 "), gzipFixture(t, []byte(payloadMarker))...)
	if sig, err := payloadState(t).visit(f.build(), 0); sig != "Payload.Test" || err != nil {
		t.Fatalf("gzip overlay: sig=%q err=%v", sig, err)
	}
	// The Authenticode table appended by signing tools is not overlay data,
	// even when its bytes look like an archive header.
	f.overlay = nil
	f.cert = authenticodeTable(append([]byte{0x1f, 0x8b, 8, 0}, bytes.Repeat([]byte{0xff}, 60)...), nil)
	if sig, err := payloadState(t).visit(f.build(), 0); sig != "" || err != nil {
		t.Fatalf("certificate: sig=%q err=%v", sig, err)
	}
}

func TestEmbeddedExecutablesAreExtracted(t *testing.T) {
	inner := simplePE(bytes.Repeat([]byte{0xcc}, 0x200)).build()
	// Stored in a section, followed by other data: the hash covers exactly
	// the stored executable only up to the next executable or the end.
	outer := simplePE(append(append(bytes.Repeat([]byte{0x90}, 0x40), inner...), bytes.Repeat([]byte{0x90}, 0x40)...))
	b := outer.build()
	start := bytes.Index(b[1:], inner[:0x40]) + 1
	if sig, err := payloadState(t, b[start:]).visit(b, 0); sig != "Payload.Hash0" || err != nil {
		t.Fatalf("section: sig=%q err=%v", sig, err)
	}
	// Stored at the end of the overlay: the reference hashes it to the end.
	outer = simplePE(bytes.Repeat([]byte{0x90}, 0x200))
	outer.overlay = append([]byte("carrier data "), inner...)
	if sig, err := payloadState(t, inner).visit(outer.build(), 0); sig != "Payload.Hash0" || err != nil {
		t.Fatalf("overlay: sig=%q err=%v", sig, err)
	}
	// An embedded executable's own installer overlay is analyzed in turn.
	stored := []byte("a member only an installer parser can bound exactly")
	withPayload := simplePE(bytes.Repeat([]byte{0xcc}, 0x200))
	withPayload.overlay = installShieldOverlay(map[string][]byte{"a.txt": stored}, []string{"a.txt"})
	outer.overlay = append([]byte("carrier data "), withPayload.build()...)
	if sig, err := payloadState(t, stored).visit(outer.build(), 0); sig != "Payload.Hash0" || err != nil {
		t.Fatalf("nested overlay: sig=%q err=%v", sig, err)
	}
	// "MZ" without a loadable header is ordinary data.
	outer.overlay = append([]byte("MZ this is only text MZ"), bytes.Repeat([]byte{0}, 128)...)
	if sig, err := payloadState(t).visit(outer.build(), 0); sig != "" || err != nil {
		t.Fatalf("text: sig=%q err=%v", sig, err)
	}
}

// Many stored executables are analyzed side by side, not as a chain of
// ever-deeper remainders that would exhaust the recursion limit.
func TestManyEmbeddedExecutablesStayShallow(t *testing.T) {
	inner := simplePE(bytes.Repeat([]byte{0xcc}, 0x200)).build()
	outer := simplePE(bytes.Repeat([]byte{0x90}, 0x200))
	for i := 0; i < 40; i++ {
		outer.overlay = append(outer.overlay, inner...)
	}
	outer.overlay = append(outer.overlay, gzipFixture(t, []byte("tail"))...)
	s := payloadState(t)
	s.maxDepth = 4
	b := outer.build()
	if sig, err := s.visit(b, 0); sig != "" || err != nil {
		t.Fatalf("sig=%q err=%v", sig, err)
	}
	if spent := int64(64<<20) - s.remaining; spent > int64(3*len(b)) {
		t.Fatalf("charged %d bytes for a %d-byte file", spent, len(b))
	}
}

// nsisLZMA compresses like NSIS: LZMA properties and dictionary size, then
// data ending with the end marker, with no size field.
func nsisLZMA(t *testing.T, data []byte) []byte {
	t.Helper()
	var b bytes.Buffer
	w, err := lzma.WriterConfig{DictCap: 1 << 16}.NewWriter(&b)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := w.Write(data); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	out := b.Bytes()
	return append(append([]byte(nil), out[:5]...), out[13:]...)
}

func nsisDeflate(t *testing.T, data []byte) []byte {
	t.Helper()
	var b bytes.Buffer
	w, _ := flate.NewWriter(&b, flate.BestCompression)
	w.Write(data)
	w.Close()
	return b.Bytes()
}

// These fixtures are bzip2 level-9 blocks converted to the NSIS wire form:
// BZh9, the five trailing block-magic bytes, block CRC and randomization bit
// are removed, and the standard end marker and CRC become a single 0x17 byte.
// The fixtures exercise the actual shortened bit stream, not a BZh wrapper.
func nsisBzip2Fixture(t *testing.T, encoded string) []byte {
	t.Helper()
	b, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

// nsisOverlay builds the first header followed by body and a CRC field.
func nsisOverlay(headerSize int, body []byte) []byte {
	b := make([]byte, 28)
	copy(b[4:], nsisMagic)
	binary.LittleEndian.PutUint32(b[20:], uint32(headerSize))
	binary.LittleEndian.PutUint32(b[24:], uint32(28+len(body)+4))
	return append(append(b, body...), 0xde, 0xad, 0xbe, 0xef)
}

func nsisBlock(data []byte, compressed bool) []byte {
	n := uint32(len(data))
	if compressed {
		n |= 0x80000000
	}
	return append(binary.LittleEndian.AppendUint32(nil, n), data...)
}

func TestNSISInstallerBlocks(t *testing.T) {
	script := bytes.Repeat([]byte("install script "), 20)
	payload := append(bytes.Repeat([]byte("installed file "), 30), payloadMarker...)
	pe := func(overlay []byte) []byte {
		f := simplePE(bytes.Repeat([]byte{0x90}, 0x200))
		f.overlay = append(make([]byte, 0), overlay...)
		return f.build()
	}
	detect := map[string][]byte{
		"LZMA blocks":    nsisOverlay(len(script), append(nsisBlock(nsisLZMA(t, script), true), nsisBlock(nsisLZMA(t, payload), true)...)),
		"deflate blocks": nsisOverlay(len(script), append(nsisBlock(nsisDeflate(t, script), true), nsisBlock(nsisDeflate(t, payload), true)...)),
		"bzip2 blocks": nsisOverlay(len(script), append(nsisBlock(nsisBzip2Fixture(t, "MQAAYyMAgABQSrgAQAChADAKVQaZqYTROiYTwmE9JomxNk0ThMJwn4uA"), true),
			nsisBlock(nsisBzip2Fixture(t, "MQABLi8AgARcnbjATkoYAEAA0GSIBptqT0gA3qqjEwEwDUa9wSWIJyCTvuqVays6ptE54yz0GtW4iWHwJMQUkH2DuD0hnSCuCkHuD9B/g8Fw"), true)...)),
		"stored block": nsisOverlay(len(script), append(nsisBlock(nsisDeflate(t, script), true), nsisBlock(gzipFixture(t, payload), false)...)),
		"solid LZMA": nsisOverlay(len(script), nsisLZMA(t, append(append(binary.LittleEndian.AppendUint32(nil, uint32(len(script))), script...),
			append(binary.LittleEndian.AppendUint32(nil, uint32(len(payload))), payload...)...))),
		"solid bzip2": nsisOverlay(len(script), nsisBzip2Fixture(t, "MQAAaq8CwACADFyduMBeSrgAEABAARAyhADQHknqAaBSqCE9QyB6hkYnqODhCbMCTMkB5BPYBD0iFeiarps6SyRb7mhDdjp4aOOX3zxz9kWxNVWpV6KuqrpiFhsTMndVgq7IsVDuTeT+RbE8kWX+Lg==")),
	}
	for name, overlay := range detect {
		t.Run(name, func(t *testing.T) {
			if bytes.Contains(overlay, []byte(payloadMarker)) {
				t.Fatal("fixture leaks the marker")
			}
			if sig, err := payloadState(t).visit(pe(overlay), 0); sig != "Payload.Test" || err != nil {
				t.Fatalf("sig=%q err=%v", sig, err)
			}
		})
	}
	good := nsisBlock(nsisLZMA(t, script), true)
	for name, c := range map[string]struct {
		overlay []byte
		want    error
	}{
		"corrupt NSIS bzip2":      {nsisOverlay(len(script), nsisBlock([]byte("1AY&SY compressed data"), true)), ErrMalformed},
		"truncated total":         {nsisOverlay(len(script), good)[:40], ErrMalformed},
		"corrupt deflate":         {nsisOverlay(len(script), nsisBlock([]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff}, true)), ErrMalformed},
		"huge dictionary":         {nsisOverlay(len(script), nsisBlock([]byte{0x5d, 0, 0, 0, 0x40, 1, 2, 3}, true)), ErrLimit},
		"solid huge member":       {nsisOverlay(len(script), nsisLZMA(t, binary.LittleEndian.AppendUint32(nil, 0x7fffffff))), ErrLimit},
		"LZMA without end marker": {nsisOverlay(len(script), nsisBlock(func() []byte { b := nsisLZMA(t, script); return b[:len(b)-6] }(), true)), ErrMalformed},
	} {
		t.Run(name, func(t *testing.T) {
			if sig, err := payloadState(t).visit(pe(c.overlay), 0); sig != "" || !errors.Is(err, c.want) {
				t.Fatalf("sig=%q err=%v want %v", sig, err, c.want)
			}
		})
	}
	// A solid stream expanding past the budget is charged and stops.
	s := payloadState(t)
	s.maxFile, s.remaining = 4096, 4096
	big := bytes.Repeat([]byte{'A'}, 1<<16)
	overlay := nsisOverlay(len(script), nsisLZMA(t, append(binary.LittleEndian.AppendUint32(nil, uint32(len(big))), big...)))
	if sig, err := s.visit(pe(overlay), 0); sig != "" || !errors.Is(err, ErrLimit) {
		t.Fatalf("budget: sig=%q err=%v", sig, err)
	}
}

func TestNSISBzip2MultiblockAndLimits(t *testing.T) {
	// Generated from a 1,800,192-byte bzip2 level-9 stream by removing
	// BZh9, per-block magic suffixes, CRCs and randomization bits, then
	// replacing the end marker with 0x17. It contains three NSIS blocks.
	packed, err := os.ReadFile("testdata/nsis_bzip2_multiblock.bin")
	if err != nil {
		t.Fatal(err)
	}
	want := bytes.Repeat(func() []byte {
		b := make([]byte, 256)
		for i := range b {
			b[i] = byte(i)
		}
		return b
	}(), 7032)
	s := payloadState(t)
	got, err := s.nsisDecode(packed)
	if err != nil || !bytes.Equal(got, want) {
		t.Fatalf("three blocks: len=%d err=%v", len(got), err)
	}
	s = payloadState(t)
	s.maxFile, s.remaining = 1<<20, 1<<20
	if _, err := s.nsisDecode(packed); !errors.Is(err, ErrLimit) {
		t.Fatalf("expansion limit: %v", err)
	}
	s = payloadState(t)
	if _, err := s.nsisDecode(packed[:len(packed)-1]); !errors.Is(err, ErrMalformed) {
		t.Fatalf("missing final marker: %v", err)
	}
}

func installShieldOverlay(files map[string][]byte, order []string) []byte {
	var b []byte
	for _, name := range order {
		b = append(b, name+"\x00Disk1\\"+name+"\x000.0.0.0\x00"+strconv.Itoa(len(files[name]))+"\x00"...)
		b = append(b, files[name]...)
	}
	return b
}

func TestInstallShieldOverlay(t *testing.T) {
	payload := gzipFixture(t, []byte(payloadMarker))
	f := simplePE(bytes.Repeat([]byte{0x90}, 0x200))
	f.overlay = installShieldOverlay(map[string][]byte{"setup.ini": []byte("[Startup]\r\n"), "data.bin": payload}, []string{"setup.ini", "data.bin"})
	if sig, err := payloadState(t).visit(f.build(), 0); sig != "Payload.Test" || err != nil {
		t.Fatalf("sig=%q err=%v", sig, err)
	}
	// The stored file is extracted with its exact size.
	stored := []byte("an installer member of exact size")
	f.overlay = installShieldOverlay(map[string][]byte{"a.txt": stored, "b.txt": []byte("second")}, []string{"a.txt", "b.txt"})
	if sig, err := payloadState(t, stored).visit(f.build(), 0); sig != "Payload.Hash0" || err != nil {
		t.Fatalf("exact member: sig=%q err=%v", sig, err)
	}
	// Records must look like InstallShield's; other overlays are not parsed.
	for _, overlay := range []string{"a\x00b\x00c\x0010\x00", "name\x00path\x001.0\x00999999\x00short", "na/me\x00p\x001\x001\x00x"} {
		if _, _, ok, _ := payloadState(t).installShield([]byte(overlay), 0); ok {
			t.Fatalf("%q recognized", overlay)
		}
	}
}

func installShieldMSIOverlay(t testing.TB, names []string, contents [][]byte) []byte {
	t.Helper()
	b := []byte("InstallShield\x00")
	header := make([]byte, 0x20)
	binary.LittleEndian.PutUint32(header, uint32(len(names)))
	b = append(b, header...)
	for i, name := range names {
		var z bytes.Buffer
		w := zlib.NewWriter(&z)
		w.Write(contents[i])
		w.Close()
		key := make([]byte, len(name))
		for j := range key {
			key[j] = name[j] ^ installShieldMSIKey[j&3]
		}
		stored := z.Bytes()
		for j, c := range stored {
			c ^= key[(j&0x3ff)%len(key)]
			stored[j] = c>>4 | c<<4
		}
		record := make([]byte, installShieldMSIRecord)
		copy(record, name)
		binary.LittleEndian.PutUint32(record[0x104:], 6)
		binary.LittleEndian.PutUint64(record[0x10c:], uint64(len(stored)))
		b = append(append(b, record...), stored...)
	}
	return b
}

func TestInstallShieldMSIOverlay(t *testing.T) {
	// Incompressible and longer than 1024 bytes, so the stored data crosses
	// the point where the format restarts its key index.
	payload := make([]byte, 4096)
	for i, x := 0, uint32(1); i < len(payload); i++ {
		x = x*1664525 + 1013904223
		payload[i] = byte(x >> 24)
	}
	payload = append(payload, payloadMarker...)
	f := simplePE(bytes.Repeat([]byte{0x90}, 0x200))
	f.overlay = append([]byte("NB10 debug info "), installShieldMSIOverlay(t, []string{"0x0409.ini", "package.msi"}, [][]byte{[]byte("[Info]"), payload})...)
	if bytes.Contains(f.overlay, []byte(payloadMarker)) {
		t.Fatal("fixture leaks the marker")
	}
	if sig, err := payloadState(t).visit(f.build(), 0); sig != "Payload.Test" || err != nil {
		t.Fatalf("sig=%q err=%v", sig, err)
	}
	overlay := installShieldMSIOverlay(t, []string{"a.bin"}, [][]byte{payload})
	binary.LittleEndian.PutUint64(overlay[14+0x20+0x10c:], 1<<40)
	f.overlay = overlay
	if sig, err := payloadState(t).visit(f.build(), 0); sig != "" || !errors.Is(err, ErrMalformed) {
		t.Fatalf("oversized record: sig=%q err=%v", sig, err)
	}
}

func TestAutoItOverlayIsUnsupported(t *testing.T) {
	f := simplePE(bytes.Repeat([]byte{0x90}, 0x200))
	f.overlay = append(append([]byte(nil), autoItMagic...), "5 compiled script"...)
	if sig, err := payloadState(t).visit(f.build(), 0); sig != "" || !errors.Is(err, ErrUnsupported) {
		t.Fatalf("sig=%q err=%v", sig, err)
	}
}

// Candidate headers that pass the signature check but not the full parse
// cost a section-table walk each; their number is bounded.
func TestEmbeddedHeaderSearchIsBounded(t *testing.T) {
	broken := make([]byte, 128)
	copy(broken, "MZ")
	binary.LittleEndian.PutUint32(broken[0x3c:], 0x40)
	copy(broken[0x40:], "PE\x00\x00")
	binary.LittleEndian.PutUint16(broken[0x46:], 1) // one section, but no optional header
	outer := simplePE(bytes.Repeat([]byte{0x90}, 0x200))
	outer.overlay = bytes.Repeat(broken, maxEmbeddedCandidates+10)
	expectIncomplete(t, payloadState(t), outer.build(), ErrLimit)
	outer.overlay = bytes.Repeat(broken, 10)
	expectDetected(t, payloadState(t), append(outer.build(), payloadMarker...), "Payload.Test")
}

// Where this engine follows the reference's parser, packer evidence alone
// does not leave a program incomplete: the reference unpacks only with its
// built-in unpackers. A program the parser would hand to one this engine
// lacks is incomplete for that reason.
func TestPackedProgramsFollowTheReferenceParser(t *testing.T) {
	s := pipelineState(t, map[string]string{"a.ndb": "Unrelated.Rule:0:*:5a5a5a5a5a5a5a5a\n"})
	named := simplePE(bytes.Repeat([]byte{0x90}, 0x200))
	named.sections[0].name = ".nsp1" // NsPack's name; its unpacker needs its entry code
	expectScan(t, s, named.build(), "", "")
	aspack := simplePE(append([]byte{0x60, 0xe8, 0x03, 0, 0, 0, 0xe9, 0xeb}, bytes.Repeat([]byte{0x90}, 0x1f8)...))
	expectScan(t, s, aspack.build(), "", "pe-unpacker")
}
