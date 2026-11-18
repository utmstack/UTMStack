package unpack

import "encoding/binary"

// RebuildSection is a section of a program the reference rebuilds after
// unpacking: its address and sizes as the reference writes them, and its
// data (at least RawSize bytes; missing bytes are zero).
type RebuildSection struct {
	RVA, VSize, RawSize uint32
	Data                []byte
	// Empty is ASPack's gap section: named "empty", no raw data and no file
	// offset, not counted in the .clamNN names.
	Empty bool
}

// Rebuild writes the program the reference rebuilds for its MEW, ASPack,
// PEtite, PESpin and Upack unpackers, byte for byte as its rebuilt files show
// it (docs/edr/native-engine/parity/unpackers.md): its DOS header and stub
// text, the PE header at 0xd0 with the time stamp "CLAM", fixed fields
// (characteristics 0x8f83, code and data sizes and bases 0x1000, section
// alignment 0x1000, file alignment 0x200, subsystem version 3.10, GUI
// subsystem, stack and heap 1 MiB/4 KiB), the given entry point, image base
// and resource directory, the header area rounded up to 0x200, the image
// size 0x1000 plus every section's virtual size rounded up to 0x1000, and
// sections named
// .clam01, .clam02, ... with every characteristic bit set, their data laid
// out one after another, each padded to 0x200.
func Rebuild(sections []RebuildSection, entry, imageBase, resourceRVA, resourceSize uint32) []byte {
	le := binary.LittleEndian
	n := len(sections)
	headerEnd := 0xd0 + 24 + 0xe0 + 40*n
	headers := (headerEnd + 0x1ff) &^ 0x1ff
	pad := func(v uint32) int { return int((uint64(v) + 0x1ff) &^ 0x1ff) }
	size := headers
	for _, s := range sections {
		size += pad(s.RawSize)
	}
	out := make([]byte, size)
	copy(out, rebuildDOSHeader)
	h := out[0xd0:]
	copy(h, "PE\x00\x00")
	le.PutUint16(h[4:], 0x14c)
	le.PutUint16(h[6:], uint16(n))
	copy(h[8:], "CLAM")
	le.PutUint16(h[20:], 0xe0)
	le.PutUint16(h[22:], 0x8f83)
	o := h[24:]
	le.PutUint16(o[0:], 0x10b)
	le.PutUint32(o[4:], 0x1000)  // size of code
	le.PutUint32(o[8:], 0x1000)  // size of initialized data
	le.PutUint32(o[16:], entry)  // entry point
	le.PutUint32(o[20:], 0x1000) // base of code
	le.PutUint32(o[24:], 0x1000) // base of data
	le.PutUint32(o[28:], imageBase)
	le.PutUint32(o[32:], 0x1000) // section alignment
	le.PutUint32(o[36:], 0x200)  // file alignment
	le.PutUint16(o[40:], 1)      // operating system version 1.0
	le.PutUint16(o[48:], 3)      // subsystem version 3.10
	le.PutUint16(o[50:], 10)
	image := uint64(0x1000)
	for _, s := range sections {
		image += (uint64(s.VSize) + 0xfff) &^ 0xfff
	}
	le.PutUint32(o[56:], uint32(image))
	le.PutUint32(o[60:], uint32(headers))
	le.PutUint16(o[68:], 2) // GUI subsystem
	le.PutUint32(o[72:], 0x100000)
	le.PutUint32(o[76:], 0x1000)
	le.PutUint32(o[80:], 0x100000)
	le.PutUint32(o[84:], 0x1000)
	le.PutUint32(o[92:], 16)
	le.PutUint32(o[96+16:], resourceRVA)
	le.PutUint32(o[96+20:], resourceSize)
	raw := headers
	number := 0
	for i, s := range sections {
		t := o[0xe0+40*i:]
		le.PutUint32(t[8:], s.VSize)
		le.PutUint32(t[12:], s.RVA)
		le.PutUint32(t[36:], 0xffffffff)
		if s.Empty {
			copy(t, "empty")
			continue
		}
		number++
		copy(t, ".clam")
		t[5], t[6] = '0'+byte(number/10%10), '0'+byte(number%10)
		le.PutUint32(t[16:], s.RawSize)
		le.PutUint32(t[20:], uint32(raw))
		copy(out[raw:raw+int(min(uint64(s.RawSize), uint64(len(s.Data))))], s.Data)
		raw += pad(s.RawSize)
	}
	return out
}
