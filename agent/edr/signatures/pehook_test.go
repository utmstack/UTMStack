package signatures

import (
	"bytes"
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"
)

// hookPE builds a PE32 program of size bytes on the layout of testPE:
// e_lfanew 128, a 224-byte optional header at 152 with 16 data directories,
// section alignment 0x1000, file alignment 0x200, header size 0x400,
// subsystem 2, and one section-table entry per {virtual size, virtual
// address, raw size, raw pointer} at 376.
func hookPE(size int, entry uint32, sections ...[4]uint32) []byte {
	return hookPEOpt(224, size, entry, sections...)
}

// hookPEOpt is hookPE with the given optional-header size; the section table
// follows the optional header.
func hookPEOpt(optional uint16, size int, entry uint32, sections ...[4]uint32) []byte {
	b := make([]byte, size)
	le := binary.LittleEndian
	copy(b, "MZ")
	le.PutUint32(b[0x3c:], 128)
	copy(b[128:], "PE\x00\x00")
	le.PutUint16(b[132:], 0x14c)
	le.PutUint16(b[134:], uint16(len(sections)))
	le.PutUint16(b[148:], optional)
	le.PutUint16(b[150:], 0x102)
	le.PutUint16(b[152:], 0x10b)
	le.PutUint32(b[168:], entry)
	le.PutUint32(b[180:], 0x400000) // image base
	le.PutUint32(b[184:], 0x1000)
	le.PutUint32(b[188:], 0x200)
	le.PutUint32(b[212:], 0x400)
	le.PutUint16(b[220:], 2)
	le.PutUint32(b[244:], 16)
	le.PutUint32(b[248+16:], 0x3000) // resource directory address
	for i, s := range sections {
		h := b[152+int(optional)+40*i:]
		copy(h, ".sec")
		le.PutUint32(h[8:], s[0])
		le.PutUint32(h[12:], s[1])
		le.PutUint32(h[16:], s[2])
		le.PutUint32(h[20:], s[3])
		le.PutUint32(h[36:], 0x60000020)
	}
	return b
}

func TestFSGHookReachRebuiltChild(t *testing.T) {
	root := filepath.Join("..", "..", "..", "docs", "edr", "native-engine", "audits", "2026-10-26-fsg2", "oracle")
	for _, tc := range []struct{ input, rebuilt string }{
		{"fsg2-harmless.exe", "reference-rebuilt.exe"},
		{"fsg133-harmless.exe", "reference-133-rebuilt.exe"},
		{"fsg133-two-sections.exe", "reference-133-two-sections.exe"},
	} {
		t.Run(tc.input, func(t *testing.T) {
			input, err := os.ReadFile(filepath.Join(root, tc.input))
			if err != nil {
				t.Fatal(err)
			}
			want, err := os.ReadFile(filepath.Join(root, tc.rebuilt))
			if err != nil {
				t.Fatal(err)
			}
			h, status := readHookHeader(input, true)
			if status != hookParsed {
				t.Fatalf("PE parser status %d", status)
			}
			reach, unpacked := h.unpackerHookReach(input, dconfPEFSG)
			if reach != reachNever || !unpacked.tried || !bytes.Equal(unpacked.child, want) {
				t.Fatalf("reach %d, tried %v, rebuilt %d bytes (reference %d)", reach, unpacked.tried, len(unpacked.child), len(want))
			}
		})
	}
}

func TestNsPackHookReachRebuiltChild(t *testing.T) {
	root := filepath.Join("..", "..", "..", "docs", "edr", "native-engine", "audits", "2026-10-26-nspack", "oracle")
	input, err := os.ReadFile(filepath.Join(root, "nspack-zero.exe"))
	if err != nil {
		t.Fatal(err)
	}
	want, err := os.ReadFile(filepath.Join(root, "reference-rebuilt.exe"))
	if err != nil {
		t.Fatal(err)
	}
	h, status := readHookHeader(input, true)
	if status != hookParsed {
		t.Fatalf("PE parser status %d", status)
	}
	reach, result := h.unpackerHookReach(input, dconfPENsPack)
	if reach != reachNever || !result.tried || !bytes.Equal(result.child, want) {
		t.Fatalf("reach %d, tried %v, rebuilt %d bytes (reference %d)", reach, result.tried, len(result.child), len(want))
	}
}

func TestHookHeaderRecord(t *testing.T) {
	le := binary.LittleEndian
	data := hookPE(0x800, 0x1010, [4]uint32{0x400, 0x1000, 0x400, 0x400})
	h, st := readHookHeader(data, true)
	if st != hookParsed {
		t.Fatalf("status %d", st)
	}
	r := h.record[:]
	checks := map[int]uint32{0: 0, 4: 0x410, 632: 128, 644: 0x1000}
	for off, want := range checks {
		if got := le.Uint32(r[off:]); got != want {
			t.Fatalf("record +%d = %#x, want %#x", off, got, want)
		}
	}
	if got := le.Uint16(r[8:]); got != 1 {
		t.Fatalf("record +8 = %d", got)
	}
	if !bytes.Equal(r[12:36], data[128:152]) || !bytes.Equal(r[36:132], data[152:248]) {
		t.Fatal("file header or optional header bytes differ")
	}
	if !bytes.Equal(r[264:360], data[152:248]) || !bytes.Equal(r[360:376], make([]byte, 16)) {
		t.Fatal("the 64-bit view of the optional header differs")
	}
	for _, off := range []int{132, 376, 504} {
		if !bytes.Equal(r[off:off+128], data[248:376]) {
			t.Fatalf("data directories at +%d differ", off)
		}
	}
	if le.Uint16(r[264:]) != 0x10b || h.epBytes != 0x800-0x410 || h.entry != 0x410 {
		t.Fatalf("magic %#x, entry bytes %d, entry %#x", le.Uint16(r[264:]), h.epBytes, h.entry)
	}
	// No overlay: the only section is the first, so the start stays 0.
	if le.Uint32(r[636:]) != 0 || le.Uint32(r[640:]) != 0x800 {
		t.Fatalf("overlay %#x %#x", le.Uint32(r[636:]), le.Uint32(r[640:]))
	}
	// Fewer directories than 16: the rest of the record's directories are zero.
	le.PutUint32(data[244:], 3)
	if h, _ = readHookHeader(data, true); !bytes.Equal(h.record[504+24:632], make([]byte, 128-24)) || le.Uint32(h.record[504+16:]) != 0x3000 {
		t.Fatal("directories beyond NumberOfRvaAndSizes are not zero")
	}
}

func TestHookHeaderStatuses(t *testing.T) {
	good := func() []byte { return hookPE(0x800, 0x1000, [4]uint32{0x400, 0x1000, 0x400, 0x400}) }
	if _, st := readHookHeader(append([]byte("MZ no PE header here, just text"), bytes.Repeat([]byte{'.'}, 100)...), true); st != hookNone {
		t.Fatalf("no PE header: %d", st)
	}
	plus := good()
	binary.LittleEndian.PutUint16(plus[152:], 0x20b)
	if _, st := readHookHeader(plus, true); st != hookPE32Plus {
		t.Fatalf("PE32+: %d", st)
	}
	// A section starting past the end of the file is removed; if it is the
	// only one the reading fails.
	two := hookPE(0x800, 0x1000, [4]uint32{0x400, 0x1000, 0x400, 0x400}, [4]uint32{0x400, 0x2000, 0x200, 0x1000})
	h, st := readHookHeader(two, true)
	if st != hookParsed || len(h.sections) != 1 || binary.LittleEndian.Uint16(h.record[8:]) != 1 {
		t.Fatalf("section past the end: status %d", st)
	}
	one := hookPE(0x800, 0x100, [4]uint32{0x400, 0x1000, 0x200, 0x1000})
	if _, st := readHookHeader(one, true); st != hookFailed {
		t.Fatalf("only section past the end: %d", st)
	}
	// Strict mode: a misaligned section address, a first section not right
	// after the header, a gap between sections.
	for name, data := range map[string][]byte{
		"misaligned": hookPE(0x800, 0x1100, [4]uint32{0x400, 0x1100, 0x400, 0x400}),
		"first":      hookPE(0x800, 0x2000, [4]uint32{0x400, 0x2000, 0x400, 0x400}),
		"gap":        hookPE(0x800, 0x1000, [4]uint32{0x400, 0x1000, 0x200, 0x400}, [4]uint32{0x200, 0x3000, 0x200, 0x600}),
	} {
		if _, st := readHookHeader(data, true); st != hookFailed {
			t.Fatalf("%s, strict: %d", name, st)
		}
		if _, st := readHookHeader(data, false); st != hookParsed {
			t.Fatalf("%s, not strict: %d", name, st)
		}
	}
	// Alignment rules apply only to programs that are not native.
	bad := good()
	binary.LittleEndian.PutUint32(bad[188:], 0x100)
	if _, st := readHookHeader(bad, true); st != hookFailed {
		t.Fatalf("file alignment 0x100, GUI: %d", st)
	}
	binary.LittleEndian.PutUint16(bad[220:], 1)
	if _, st := readHookHeader(bad, true); st != hookParsed {
		t.Fatalf("file alignment 0x100, native: %d", st)
	}
	// An unmappable entry point fails the reading.
	if _, st := readHookHeader(hookPE(0x800, 0x9000, [4]uint32{0x400, 0x1000, 0x400, 0x400}), true); st != hookFailed {
		t.Fatalf("unmappable entry point: %d", st)
	}
}

func TestHookHeaderOverlay(t *testing.T) {
	le := binary.LittleEndian
	data := hookPE(0x1000, 0x1000, [4]uint32{0x1000, 0x1000, 0x400, 0x400}, [4]uint32{0x1000, 0x2000, 0x200, 0x800})
	h, st := readHookHeader(data, true)
	if st != hookParsed {
		t.Fatalf("status %d", st)
	}
	if le.Uint32(h.record[636:]) != 0xa00 || le.Uint32(h.record[640:]) != 0x1000-0xa00 {
		t.Fatalf("overlay %#x %#x", le.Uint32(h.record[636:]), le.Uint32(h.record[640:]))
	}
	pe := h.bytecodePE()
	if len(pe.Record) != 648 || pe.EntryPoint != 0x400 || pe.HeaderSize != 0x1000 || len(pe.Sections) != 2 || pe.Sections[1].Raw != 0x800 {
		t.Fatalf("bytecode view %+v", pe)
	}
}

// nrv2bLiterals is an NRV2B stream of the bytes as literals and the end
// marker: control bits in 32-bit little-endian words, most significant
// first; the marker is a 0 bit, the offset prefix 0x1000002 (bits after the
// leading one, each followed by 1 on the last) and the byte 0xff.
func nrv2bLiterals(data []byte) []byte {
	var out []byte
	at, left, word := 0, 0, uint32(0)
	bit := func(b uint32) {
		if left == 0 {
			if len(out) > 0 {
				binary.LittleEndian.PutUint32(out[at:], word)
			}
			at, left, word = len(out), 32, 0
			out = append(out, 0, 0, 0, 0)
		}
		left--
		word |= b << uint(left)
	}
	for _, c := range data {
		bit(1)
		out = append(out, c)
	}
	bit(0)
	const v = 0xffffff + 3
	for i := 23; i >= 0; i-- {
		bit(v >> uint(i) & 1)
		bit(map[bool]uint32{true: 1, false: 0}[i == 0])
	}
	out = append(out, 0xff)
	binary.LittleEndian.PutUint32(out[at:], word)
	return out
}

// Dynamic-configuration bits of the PE word for the built-in unpackers.
const allUnpackers = dconfPEUPX | dconfPEFSG | dconfPEMEW | dconfPEPetite | dconfPEPESpin | dconfPEYC | dconfPEWWPack | dconfPENsPack | dconfPEUpack | dconfPEASPack

// reachOf reads data as the hook phase does and asks whether the unpacker
// hooks are reached.
func reachOf(t *testing.T, data []byte, pe uint32) hookReach {
	t.Helper()
	h, st := readHookHeader(data, true)
	if st != hookParsed {
		t.Fatalf("header status %d", st)
	}
	reach, _ := h.unpackerHookReach(data, pe)
	return reach
}

// entryPE is a two-section program (a first section of 0x200 raw bytes and a
// second one of raw bytes starting at 0x600) whose entry point is the start
// of the second section, with code written there.
func entryPE(code []byte, secondRaw uint32) []byte {
	b := hookPE(0x600+int(secondRaw), 0x2000, [4]uint32{0x1000, 0x1000, 0x200, 0x400}, [4]uint32{secondRaw, 0x2000, secondRaw, 0x600})
	copy(b[0x600:], code)
	return b
}

func TestUnpackerHookReach(t *testing.T) {
	le := binary.LittleEndian
	plain := hookPE(0x800, 0x1000, [4]uint32{0x400, 0x1000, 0x400, 0x400})
	if got := reachOf(t, plain, allUnpackers); got != reachHooks {
		t.Fatalf("plain program: %d", got)
	}
	// Too few bytes at the entry point: the reference returns before the
	// unpacker hooks (fewer than 168, then fewer than 200).
	for _, left := range []uint32{124, 190} {
		short := hookPE(0x800, 0x1000+0x400-left, [4]uint32{0x400, 0x1000, 0x400, 0x400})
		if got := reachOf(t, short, allUnpackers); got != reachNever {
			t.Fatalf("%d entry bytes: %d", left, got)
		}
	}
	// The empty-section layout: UPX is tried. On zero bytes every decoder
	// fails, so the parser goes on to the hooks.
	empty := hookPE(0x800, 0x2000, [4]uint32{0x1000, 0x1000, 0, 0}, [4]uint32{0x400, 0x2000, 0x400, 0x400})
	if got := reachOf(t, empty, allUnpackers); got != reachHooks {
		t.Fatalf("empty section: %d", got)
	}
	// A stream one decoder reads to its end: the rebuild is not modeled.
	packed := hookPE(0x800, 0x2000, [4]uint32{0x1000, 0x1000, 0, 0}, [4]uint32{0x400, 0x2000, 0x400, 0x400})
	copy(packed[0x400:], nrv2bLiterals(bytes.Repeat([]byte("packed program "), 20)))
	if got := reachOf(t, packed, allUnpackers); got != reachUnknown {
		t.Fatalf("decodable stream: %d", got)
	}
	// "mov esi, imm32" at the entry point moves the source by 0x15 bytes:
	// zeros there fail every decoder; a stream there decodes (to no
	// program header, so the rebuild is not modeled).
	skewed := hookPE(0x800, 0x2000, [4]uint32{0x1000, 0x1000, 0, 0}, [4]uint32{0x400, 0x2000, 0x400, 0x400})
	copy(skewed[0x400:], []byte{0x60, 0xbe, 0x15, 0x20, 0x40, 0})
	if got := reachOf(t, skewed, allUnpackers); got != reachHooks {
		t.Fatalf("mov esi at the entry point, zeros: %d", got)
	}
	copy(skewed[0x415:], nrv2bLiterals(bytes.Repeat([]byte("skewed "), 20)))
	if got := reachOf(t, skewed, allUnpackers); got != reachUnknown {
		t.Fatalf("mov esi at the entry point, a stream: %d", got)
	}
	// Both aligned virtual sizes not larger than the next section's aligned
	// raw size: the reference reports a size mismatch and returns.
	mismatch := hookPE(0x2400, 0x2000, [4]uint32{0x1000, 0x1000, 0, 0}, [4]uint32{0x400, 0x2000, 0x2000, 0x400})
	if got := reachOf(t, mismatch, allUnpackers); got != reachNever {
		t.Fatalf("size mismatch: %d", got)
	}
	if got := reachOf(t, empty, allUnpackers&^(dconfPEUPX|dconfPEFSG|dconfPEMEW)); got != reachHooks {
		t.Fatalf("empty section, UPX, FSG and MEW off: %d", got)
	}
	// MEW is tried before the 168-byte exit.
	mew := hookPE(0x480, 0x2000, [4]uint32{0x1000, 0x1000, 0, 0}, [4]uint32{0x400, 0x2000, 0x400, 0x400})
	mew = mew[:0x400+124]
	mew[0x400] = 0xe9
	rel := int32(0x154 - 0x2000 - 5)
	le.PutUint32(mew[0x401:], uint32(rel))
	if got := reachOf(t, mew, dconfPEMEW); got != reachUnknown {
		t.Fatalf("MEW: %d", got)
	}
	if got := reachOf(t, mew, dconfPEUPX); got != reachNever {
		t.Fatalf("MEW off, 124 entry bytes: %d", got)
	}
	// Upack, two sections (with the empty section) and three sections (an
	// optional header of 0x148 bytes).
	upack2 := hookPE(0x800, 0x2000, [4]uint32{0x1000, 0x1000, 0, 0}, [4]uint32{0x400, 0x2000, 0x400, 0x400})
	copy(upack2[0x400:], []byte{0x60, 0xe8, 0x09, 0, 0, 0})
	if got := reachOf(t, upack2, dconfPEMEW); got != reachUnknown {
		t.Fatalf("Upack, 2 sections: %d", got)
	}
	upack3 := hookPEOpt(0x148, 0x1000, 0x3000, [4]uint32{0x1000, 0x1000, 0x200, 0x600}, [4]uint32{0x1000, 0x2000, 0x200, 0x800}, [4]uint32{0x800, 0x3000, 0x800, 0xa00})
	copy(upack3[0xa00:], []byte{0xbe, 0, 0x30, 0x40, 0, 0xad, 0x50})
	if got := reachOf(t, upack3, dconfPEUpack); got != reachUnknown {
		t.Fatalf("Upack, 3 sections: %d", got)
	}
	if got := reachOf(t, upack3, 0); got != reachHooks {
		t.Fatalf("Upack, 3 sections, off: %d", got)
	}
	// FSG entry codes with the empty section.
	for _, code := range [][]byte{{0xbe, 0x00, 0x08, 0x40, 0x00}} {
		fsg := hookPE(0x800, 0x2000, [4]uint32{0x1000, 0x1000, 0, 0}, [4]uint32{0x400, 0x2000, 0x400, 0x400})
		copy(fsg[0x400:], code)
		if got := reachOf(t, fsg, dconfPEFSG); got != reachUnknown {
			t.Fatalf("FSG % x: %d", code, got)
		}
	}
	fsg2 := hookPE(0x800, 0x2000, [4]uint32{0x1000, 0x1000, 0, 0}, [4]uint32{0x400, 0x2000, 0x400, 0x400})
	copy(fsg2[0x400:], []byte{0x87, 0x25})
	if got := reachOf(t, fsg2, dconfPEFSG); got != reachHooks {
		t.Fatalf("FSG 2.0 invalid pointer slot: %d", got)
	}
	// The third FSG form also needs the entry point far enough into the
	// section after the empty one.
	fsg3 := []byte{0xbb, 0x00, 0x08, 0x40, 0x00, 0xbf, 0, 0, 0, 0, 0xbe}
	far := hookPE(0x3400, 0x4000, [4]uint32{0x1000, 0x1000, 0, 0}, [4]uint32{0x3000, 0x2000, 0x3000, 0x400})
	copy(far[0x2400:], fsg3)
	// The reference ends the PE parser when the packed raw section is
	// larger than the empty destination, before reading support data.
	if got := reachOf(t, far, dconfPEFSG); got != reachNever {
		t.Fatalf("FSG, third form: %d", got)
	}
	near := hookPE(0x800, 0x2000, [4]uint32{0x1000, 0x1000, 0, 0}, [4]uint32{0x400, 0x2000, 0x400, 0x400})
	copy(near[0x400:], fsg3)
	if got := reachOf(t, near, dconfPEFSG); got != reachHooks {
		t.Fatalf("FSG, third form at the section start: %d", got)
	}
	// Petite: mov eax, <last or second-to-last section address>.
	petite := entryPE([]byte{0xb8, 0x00, 0x20, 0x40, 0x00}, 0x400)
	if got := reachOf(t, petite, dconfPEPetite); got != reachUnknown {
		t.Fatalf("Petite: %d", got)
	}
	if got := reachOf(t, petite, allUnpackers&^dconfPEPetite); got != reachHooks {
		t.Fatalf("Petite off: %d", got)
	}
	// PESpin: the entry point in the last section, far enough from its end,
	// is tried; a stub without the version's instruction bytes fails ("Not
	// spinned or bad version") and the parser goes on.
	spin := entryPE(append([]byte{0, 0, 0, 0}, 0xe8, 0, 0, 0, 0, 0x8b, 0x1c, 0x24, 0x83, 0xc3), 0x4000)
	if got := reachOf(t, spin, dconfPEPESpin); got != reachHooks {
		t.Fatalf("PESpin: %d", got)
	}
	// yC: the entry point 0x60 into the last section.
	yc := hookPE(0x1000, 0x2060, [4]uint32{0x1000, 0x1000, 0x200, 0x400}, [4]uint32{0x800, 0x2000, 0x800, 0x600})
	if got := reachOf(t, yc, dconfPEYC); got != reachUnknown {
		t.Fatalf("yC: %d", got)
	}
	// WWPack: the entry point at the last section with its two code marks.
	ww := entryPE([]byte{0x53, 0x55, 0x8b, 0xe8, 0x33, 0xdb, 0xeb}, 0x400)
	copy(ww[0x600+0x68:], []byte{0xe8, 0, 0, 0, 0, 0x58, 0x2d, 0x6d, 0, 0, 0, 0x50, 0x60, 0x33, 0xc9, 0x50, 0x58, 0x50, 0x50})
	if got := reachOf(t, ww, dconfPEWWPack); got != reachUnknown {
		t.Fatalf("WWPack: %d", got)
	}
	// ASPack and NsPack entry codes.
	if got := reachOf(t, entryPE([]byte{0x60, 0xe8, 0x03, 0, 0, 0, 0xe9, 0xeb}, 0x400), dconfPEASPack); got != reachUnknown {
		t.Fatalf("ASPack: %d", got)
	}
	if got := reachOf(t, entryPE([]byte{0x9c, 0x60, 0xe8, 0, 0, 0, 0, 0x5d, 0xb8, 0x07, 0, 0, 0}, 0x400), dconfPENsPack); got != reachUnknown {
		t.Fatalf("NsPack: %d", got)
	}
	// NsPack behind a jump at the entry point.
	jump := entryPE([]byte{0xe9, 0x0b, 0, 0, 0}, 0x400)
	copy(jump[0x610:], []byte{0x9c, 0x60, 0xe8, 0, 0, 0, 0, 0x5d, 0xb8, 0x07, 0, 0, 0})
	if got := reachOf(t, jump, dconfPENsPack); got != reachUnknown {
		t.Fatalf("NsPack behind a jump: %d", got)
	}
	if got := reachOf(t, entryPE([]byte{0x55, 0x8b, 0xec}, 0x400), allUnpackers); got != reachHooks {
		t.Fatalf("ordinary entry code: %d", got)
	}
}
