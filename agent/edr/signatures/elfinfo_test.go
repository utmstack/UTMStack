package signatures

import (
	"encoding/binary"
	"testing"
)

// elf64 builds a little-endian ELF64 header with program headers
// {type, offset, vaddr, memsz} and nsect 64-byte section headers.
func elf64(order binary.ByteOrder, dataByte byte, entry uint64, phdrs [][4]uint64, nsect int, shentsize uint16) []byte {
	b := make([]byte, 4096)
	copy(b, "\x7fELF")
	b[4], b[5] = 2, dataByte
	order.PutUint64(b[24:], entry)
	order.PutUint64(b[32:], 64) // e_phoff
	order.PutUint16(b[54:], 56)
	order.PutUint16(b[56:], uint16(len(phdrs)))
	shoff := uint64(64 + 56*len(phdrs))
	order.PutUint64(b[40:], shoff)
	order.PutUint16(b[58:], shentsize)
	order.PutUint16(b[60:], uint16(nsect))
	for i, p := range phdrs {
		h := b[64+56*i:]
		order.PutUint32(h[0:], uint32(p[0]))
		order.PutUint64(h[8:], p[1])
		order.PutUint64(h[16:], p[2])
		order.PutUint64(h[40:], p[3])
	}
	for i := 0; i < nsect; i++ {
		h := b[shoff+uint64(64*i):]
		order.PutUint64(h[16:], uint64(0x1000*i)) // sh_addr
		order.PutUint64(h[24:], uint64(0x100*i))  // sh_offset
		order.PutUint64(h[32:], 0x80)             // sh_size
	}
	return b
}

func TestELFEntryPointFollowsTheFirstContainingHeader(t *testing.T) {
	// A PT_NOTE (4) listed first also contains the entry: it decides.
	b := elf64(binary.LittleEndian, 1, 0x401010, [][4]uint64{{4, 0x300, 0x401000, 0x100}, {1, 0x800, 0x401000, 0x1000}}, 3, 64)
	meta := parseELF(b)
	if meta == nil || meta.entry != 0x310 {
		t.Fatalf("entry = %+v, want 0x310 from the PT_NOTE", meta)
	}
	if len(meta.sections) != 3 || meta.sections[2].raw != 0x200 || meta.sections[2].size != 0x80 {
		t.Fatalf("sections = %+v", meta.sections)
	}
	// Byte 5 other than 1 means big-endian.
	be := elf64(binary.BigEndian, 7, 0x401010, [][4]uint64{{1, 0x800, 0x401000, 0x1000}}, 1, 64)
	if meta := parseELF(be); meta == nil || meta.entry != 0x810 {
		t.Fatalf("big-endian entry = %+v", meta)
	}
	// Entry 0 gives EP 0 whatever the program headers are.
	if meta := parseELF(elf64(binary.LittleEndian, 1, 0, nil, 1, 64)); meta == nil || meta.entry != 0 {
		t.Fatalf("zero entry = %+v", meta)
	}
	// No containing header, a section entry size check with no sections,
	// and a bad class all fail.
	for name, b := range map[string][]byte{
		"no containing header": elf64(binary.LittleEndian, 1, 0x999999, [][4]uint64{{1, 0x800, 0x401000, 0x10}}, 1, 64),
		"bad section size":     elf64(binary.LittleEndian, 1, 0, nil, 0, 0),
		"class 3":              func() []byte { b := elf64(binary.LittleEndian, 1, 0, nil, 1, 64); b[4] = 3; return b }(),
		"too short":            []byte("\x7fELF\x02\x01"),
	} {
		if meta := parseELF(b); meta != nil {
			t.Errorf("%s: parsed %+v", name, meta)
		}
	}
}

func TestELFOffsetsInRules(t *testing.T) {
	b := elf64(binary.LittleEndian, 1, 0, nil, 13, 64)
	copy(b[0x100*12+0x20:], "SECTMARK")
	copy(b[len(b)-4:], "TAIL")
	d := compileTest(t, "ndb", "Ep:6:EP+0:7f454c46\nSect:6:S12+32:534543544d41524b\nOut:6:S13+0:7f454c46\nEnd:0:EOF-4:5441494c\n")
	res, err := d.MatchObject(nil2ctx(), b, MatchContext{Type: TypeELF})
	if err != nil || res.Name != "Ep" {
		t.Fatalf("EP+0 at offset 0: %q %v", res.Name, err)
	}
	d = compileTest(t, "ndb", "Sect:6:S12+32:534543544d41524b\n")
	if res, err := d.MatchObject(nil2ctx(), b, MatchContext{Type: TypeELF}); err != nil || res.Name != "Sect" {
		t.Fatalf("section offset: %q %v", res.Name, err)
	}
	d = compileTest(t, "ndb", "Out:6:S13+0:7f454c46\n")
	if res, err := d.MatchObject(nil2ctx(), b, MatchContext{Type: TypeELF}); err != nil || res.Name != "" {
		t.Fatalf("section index out of range: %q %v", res.Name, err)
	}
	// A failed header disables every relative offset, root 0 included.
	d = compileTest(t, "ndb", "End:0:EOF-4:5441494c\n")
	broken := elf64(binary.LittleEndian, 1, 0, nil, 0, 0)
	copy(broken[len(broken)-4:], "TAIL")
	if res, err := d.MatchObject(nil2ctx(), broken, MatchContext{Type: TypeELF}); err != nil || res.Name != "" {
		t.Fatalf("EOF on a failed header: %q %v", res.Name, err)
	}
	if res, err := d.MatchObject(nil2ctx(), b, MatchContext{Type: TypeELF}); err != nil || res.Name != "End" {
		t.Fatalf("EOF on a parsed header: %q %v", res.Name, err)
	}
}
