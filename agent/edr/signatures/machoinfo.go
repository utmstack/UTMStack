package signatures

import "encoding/binary"

// parseMachO reads the header information the reference uses for Mach-O
// offsets: the section list and the entry point, which only PowerPC thread
// states set (on x86, x86-64 and ARM the entry is 0, so EP+n is file offset
// n). It returns nil where the reference's reader fails, including every
// universal ("fat") binary, whose slices are analyzed on their own.
func parseMachO(data []byte) *peMeta {
	if len(data) < 28 {
		return nil
	}
	var order binary.ByteOrder
	wide := false
	switch binary.LittleEndian.Uint32(data) {
	case 0xfeedface:
		order = binary.LittleEndian
	case 0xcefaedfe:
		order = binary.BigEndian
	case 0xfeedfacf:
		order, wide = binary.LittleEndian, true
	case 0xcffaedfe:
		order, wide = binary.BigEndian, true
	default:
		return nil
	}
	cpu := order.Uint32(data[4:])
	ncmds := order.Uint32(data[16:])
	if ncmds < 1 || ncmds > 1024 {
		return nil
	}
	pos := uint64(28)
	if wide {
		pos = 32
	}
	read := func(n uint64) []byte {
		if pos > uint64(len(data)) || uint64(len(data))-pos < n {
			return nil
		}
		b := data[pos : pos+n]
		pos += n
		return b
	}
	const (
		cpuX86   = 7
		cpuPPC   = 18
		cpuPPC64 = 18 | 0x1000000
	)
	meta := &peMeta{valid: true}
	var ep uint32
	var vsizes []uint32
	for i := uint32(0); i < ncmds; i++ {
		h := read(8)
		if h == nil {
			return nil
		}
		cmd, cmdsize := order.Uint32(h), order.Uint32(h[4:])
		switch {
		case cmd == 0x1 && !wide || cmd == 0x19 && wide:
			body, sectSize, nsectsAt := uint64(48), uint64(68), 40
			if wide {
				body, sectSize, nsectsAt = 64, 80, 56
			}
			seg := read(body)
			if seg == nil {
				return nil
			}
			nsects := order.Uint32(seg[nsectsAt:])
			if nsects > 255 {
				return nil
			}
			for j := uint32(0); j < nsects; j++ {
				s := read(sectSize)
				if s == nil {
					return nil
				}
				var addr, size uint64
				var offset, align, flags uint32
				if wide {
					addr, size = order.Uint64(s[32:]), order.Uint64(s[40:])
					offset, align, flags = order.Uint32(s[48:]), order.Uint32(s[52:]), order.Uint32(s[64:])
				} else {
					addr, size = uint64(order.Uint32(s[32:])), uint64(order.Uint32(s[36:]))
					offset, align, flags = order.Uint32(s[40:]), order.Uint32(s[44:]), order.Uint32(s[56:])
				}
				if align > 31 || size > 0xffffffff {
					return nil
				}
				raw := uint64(0)
				if kind := flags & 0xff; kind != 0x1 && kind != 0xc && kind != 0x12 {
					unit := uint64(1) << align
					if raw = (size + unit - 1) / unit * unit; raw > 0xffffffff {
						return nil
					}
				}
				meta.sections = append(meta.sections, section{rva: uint32(addr), raw: offset, size: uint32(raw)})
				vsizes = append(vsizes, uint32(size))
			}
		case (cmd == 0x4 || cmd == 0x5) && (cpu == cpuX86 || cpu == cpuPPC || cpu == cpuPPC64):
			if read(8) == nil {
				return nil
			}
			state := uint64(64)
			switch cpu {
			case cpuPPC:
				state = 160
			case cpuPPC64:
				state = 312
			}
			s := read(state)
			if s == nil {
				return nil
			}
			switch cpu {
			case cpuPPC:
				ep = order.Uint32(s)
			case cpuPPC64:
				ep = uint32(order.Uint64(s))
			}
		case cmdsize > 8:
			if read(uint64(cmdsize-8)) == nil {
				return nil
			}
		}
	}
	if ep != 0 {
		found := false
		for i, s := range meta.sections {
			if s.rva <= ep && ep-s.rva < vsizes[i] {
				meta.entry, found = int(ep-s.rva+s.raw), true
				break
			}
		}
		if !found {
			return nil
		}
	}
	return meta
}
