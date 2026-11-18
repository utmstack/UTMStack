package signatures

import "encoding/binary"

// parseELF reads the header information the reference uses for ELF offsets
// and descriptors: the entry point as a file offset and the section table.
// It returns nil where the reference's header reader fails, which disables
// every relative offset on the object.
//
// Byte 4 must be 1 (32-bit) or 2 (64-bit); byte 5 equal to 1 is
// little-endian and any other value big-endian. The entry is mapped through
// the first program header of any type that contains it; an entry of 0 or no
// program headers gives 0. Every section header entry counts, the null one
// included; the entry size must be 40 or 64 bytes even when there are none.
func parseELF(data []byte) *peMeta {
	if len(data) < 52 || string(data[:4]) != "\x7fELF" {
		return nil
	}
	wide := false
	switch data[4] {
	case 1:
	case 2:
		wide = true
		if len(data) < 64 {
			return nil
		}
	default:
		return nil
	}
	var order binary.ByteOrder = binary.BigEndian
	if data[5] == 1 {
		order = binary.LittleEndian
	}
	var entry, phoff, shoff uint64
	var phentsize, phnum, shentsize, shnum uint16
	if wide {
		entry, phoff, shoff = order.Uint64(data[24:]), order.Uint64(data[32:]), order.Uint64(data[40:])
		phentsize, phnum, shentsize, shnum = order.Uint16(data[54:]), order.Uint16(data[56:]), order.Uint16(data[58:]), order.Uint16(data[60:])
	} else {
		entry, phoff, shoff = uint64(order.Uint32(data[24:])), uint64(order.Uint32(data[28:])), uint64(order.Uint32(data[32:]))
		phentsize, phnum, shentsize, shnum = order.Uint16(data[42:]), order.Uint16(data[44:]), order.Uint16(data[46:]), order.Uint16(data[48:])
	}
	read := func(at uint64, n int) []byte {
		if at > uint64(len(data)) || uint64(len(data))-at < uint64(n) {
			return nil
		}
		return data[at : at+uint64(n)]
	}
	meta := &peMeta{valid: true}
	switch {
	case phnum > 128:
		return nil
	case phnum == 0 || entry == 0:
	default:
		size := uint16(32)
		if wide {
			size = 56
		}
		if phentsize != size {
			return nil
		}
		found := false
		for i := uint64(0); i < uint64(phnum); i++ {
			h := read(phoff+i*uint64(size), int(size))
			if h == nil {
				return nil
			}
			if found {
				continue // every header must still be readable
			}
			if wide {
				offset, vaddr, memsz := order.Uint64(h[8:]), order.Uint64(h[16:]), order.Uint64(h[40:])
				if vaddr <= entry && entry < vaddr+memsz {
					meta.entry, found = int(uint32(entry-vaddr+offset)), true
				}
			} else {
				offset, vaddr, memsz := order.Uint32(h[4:]), order.Uint32(h[8:]), order.Uint32(h[20:])
				if e := uint32(entry); vaddr <= e && e < vaddr+memsz {
					meta.entry, found = int(e-vaddr+offset), true
				}
			}
		}
		if !found {
			return nil
		}
	}
	size := uint16(40)
	if wide {
		size = 64
	}
	if shnum > 256 || shentsize != size {
		return nil
	}
	meta.sections = make([]section, 0, shnum)
	for i := uint64(0); i < uint64(shnum); i++ {
		h := read(shoff+i*uint64(size), int(size))
		if h == nil {
			return nil
		}
		if wide {
			meta.sections = append(meta.sections, section{rva: uint32(order.Uint64(h[16:])), raw: uint32(order.Uint64(h[24:])), size: uint32(order.Uint64(h[32:]))})
		} else {
			meta.sections = append(meta.sections, section{rva: order.Uint32(h[12:]), raw: order.Uint32(h[16:]), size: order.Uint32(h[20:])})
		}
	}
	return meta
}
