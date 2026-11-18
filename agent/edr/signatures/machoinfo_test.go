package signatures

import (
	"context"
	"encoding/binary"
	"testing"
)

func nil2ctx() context.Context { return context.Background() }

// macho64 builds a little-endian 64-bit Mach-O with one LC_SEGMENT_64 of two
// sections and an LC_MAIN command.
func macho64(cpu uint32) []byte {
	b := make([]byte, 4096)
	le := binary.LittleEndian
	le.PutUint32(b[0:], 0xfeedfacf)
	le.PutUint32(b[4:], cpu)
	le.PutUint32(b[16:], 2) // ncmds
	pos := 32
	le.PutUint32(b[pos:], 0x19)      // LC_SEGMENT_64
	le.PutUint32(b[pos+4:], 72+2*80) // cmdsize
	le.PutUint32(b[pos+64:], 2)      // nsects
	for i := 0; i < 2; i++ {
		s := b[pos+72+80*i:]
		le.PutUint64(s[32:], uint64(0x100000000+0x1000*i)) // addr
		le.PutUint64(s[40:], 0x40)                         // size
		le.PutUint32(s[48:], uint32(0x400+0x100*i))        // offset
		le.PutUint32(s[52:], 4)                            // align 2^4
	}
	pos += 72 + 2*80
	le.PutUint32(b[pos:], 0x80000028) // LC_MAIN
	le.PutUint32(b[pos+4:], 24)
	return b
}

func TestMachOHeaderInformation(t *testing.T) {
	meta := parseMachO(macho64(0x01000007))
	if meta == nil || meta.entry != 0 || len(meta.sections) != 2 || meta.sections[1].raw != 0x500 || meta.sections[1].size != 0x40 {
		t.Fatalf("x86-64 Mach-O: %+v", meta)
	}
	for name, b := range map[string][]byte{
		"universal":   []byte("\xca\xfe\xba\xbe\x00\x00\x00\x02" + string(make([]byte, 64))),
		"no commands": func() []byte { b := macho64(7); binary.LittleEndian.PutUint32(b[16:], 0); return b }(),
		"too short":   []byte("\xcf\xfa\xed\xfe"),
	} {
		if meta := parseMachO(b); meta != nil {
			t.Errorf("%s: parsed %+v", name, meta)
		}
	}
	d := compileTest(t, "ndb", "Macho.Ep:9:EP+20:00000000\n")
	b := macho64(0x01000007)
	if res, err := d.MatchObject(nil2ctx(), b, MatchContext{Type: TypeMachO}); err != nil || res.Name != "Macho.Ep" {
		t.Fatalf("EP+20 is file offset 20 on x86-64: %q %v", res.Name, err)
	}
	// A universal binary always fails its header: no relative offsets.
	d = compileTest(t, "ndb", "End:0:EOF-4:5441494c\n")
	fat := append([]byte("\xca\xfe\xba\xbe\x00\x00\x00\x01"), make([]byte, 100)...)
	copy(fat[len(fat)-4:], "TAIL")
	if res, err := d.MatchObject(nil2ctx(), fat, MatchContext{Type: TypeMachOUnibin}); err != nil || res.Name != "" {
		t.Fatalf("EOF on a universal binary: %q %v", res.Name, err)
	}
}
