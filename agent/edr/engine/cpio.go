package engine

import (
	"bytes"
	"encoding/binary"
)

// cpio archives come in four header formats: portable ASCII ("odc", octal
// fields), "newc" and its checksummed "crc" variant (hexadecimal fields,
// four-byte alignment) and the old binary format in either byte order
// (two-byte alignment). Each header is followed by the NUL-terminated name
// and the member data; the member named TRAILER!!! ends the archive.
// https://man.freebsd.org/cgi/man.cgi?query=cpio&sektion=5
type cpioKind int

const (
	cpioNone cpioKind = iota
	cpioODC
	cpioNewc
	cpioBinaryLE
	cpioBinaryBE
)

const cpioMaxName = 4096

func cpioFormat(b []byte) cpioKind {
	switch {
	case bytes.HasPrefix(b, []byte("070707")):
		return cpioODC
	case bytes.HasPrefix(b, []byte("070701")), bytes.HasPrefix(b, []byte("070702")):
		return cpioNewc
	case len(b) >= 2 && b[0] == 0xc7 && b[1] == 0x71:
		return cpioBinaryLE
	case len(b) >= 2 && b[0] == 0x71 && b[1] == 0xc7:
		return cpioBinaryBE
	}
	return cpioNone
}

type cpioEntry struct {
	name       []byte
	data, next int // member data offset and the following header's offset
	size       int
}

// cpioParse reads the header at pos. ok is false when the bytes are not a
// well-formed header of this kind; the whole entry must fit in data.
func cpioParse(data []byte, pos int, kind cpioKind) (e cpioEntry, ok bool) {
	b := data[pos:]
	var header, nameSize, align int
	var size uint64
	switch kind {
	case cpioODC:
		header, align = 76, 1
		if len(b) < header || !bytes.HasPrefix(b, []byte("070707")) {
			return e, false
		}
		n, good1 := cpioNumber(b[59:65], 8)
		sz, good2 := cpioNumber(b[65:76], 8)
		for _, f := range [][]byte{b[6:12], b[12:18], b[18:24], b[24:30], b[30:36], b[36:42], b[42:48], b[48:59]} {
			if _, good := cpioNumber(f, 8); !good {
				return e, false
			}
		}
		if !good1 || !good2 {
			return e, false
		}
		nameSize, size = int(n), sz
	case cpioNewc:
		header, align = 110, 4
		if len(b) < header || !bytes.HasPrefix(b, []byte("070701")) && !bytes.HasPrefix(b, []byte("070702")) {
			return e, false
		}
		var fields [13]uint64
		for i := range fields {
			v, good := cpioNumber(b[6+8*i:14+8*i], 16)
			if !good {
				return e, false
			}
			fields[i] = v
		}
		nameSize, size = int(min(fields[11], cpioMaxName+1)), fields[6]
	case cpioBinaryLE, cpioBinaryBE:
		header, align = 26, 2
		var order binary.ByteOrder = binary.LittleEndian
		if kind == cpioBinaryBE {
			order = binary.BigEndian
		}
		if len(b) < header || order.Uint16(b) != 0o070707 {
			return e, false
		}
		nameSize = int(order.Uint16(b[20:22]))
		// The 32-bit size is stored as two 16-bit words, most significant first.
		size = uint64(order.Uint16(b[22:24]))<<16 | uint64(order.Uint16(b[24:26]))
	default:
		return e, false
	}
	if nameSize < 1 || nameSize > cpioMaxName || nameSize > len(b)-header {
		return e, false
	}
	name := b[header : header+nameSize]
	if name[nameSize-1] != 0 || bytes.IndexByte(name[:nameSize-1], 0) >= 0 {
		return e, false
	}
	start := cpioAlign(header+nameSize, align)
	if start > len(b) || size > uint64(len(b)-start) {
		return e, false
	}
	e = cpioEntry{name: name[:nameSize-1], data: pos + start, size: int(size)}
	e.next = pos + cpioAlign(start+int(size), align)
	return e, true
}

func cpioAlign(n, align int) int {
	return (n + align - 1) / align * align
}

func cpioNumber(field []byte, base uint64) (uint64, bool) {
	var v uint64
	for _, c := range field {
		var d uint64
		switch {
		case c >= '0' && c <= '9':
			d = uint64(c - '0')
		case base == 16 && c >= 'a' && c <= 'f':
			d = uint64(c-'a') + 10
		case base == 16 && c >= 'A' && c <= 'F':
			d = uint64(c-'A') + 10
		default:
			return 0, false
		}
		if d >= base {
			return 0, false
		}
		v = v*base + d
	}
	return v, true
}

// cpio analyzes every member that holds data. It reports whether data is a
// cpio archive at all: a first header that does not parse leaves the object
// to other analysis. The reference accepts EOF after a complete member even
// without TRAILER!!!; a full but invalid later header remains malformed.
func (s *scanState) cpio(data []byte, kind cpioKind, depth int) (sig string, recognized bool, err error) {
	first, ok := cpioParse(data, 0, kind)
	if !ok {
		return "", false, nil
	}
	if depth >= s.maxDepth {
		return "", true, ErrLimit
	}
	var incomplete error
	e := first
	for count := 0; ; count++ {
		if err := s.ctx.Err(); err != nil {
			return "", true, err
		}
		if string(e.name) == "TRAILER!!!" {
			return "", true, incomplete
		}
		if count >= s.maxFiles-s.files {
			return "", true, ErrLimit
		}
		if e.size > 0 {
			// Members are windows of the archive. The reference still runs
			// type recognition on those windows, including embedded archives.
			sig, err := s.visitObject(data[e.data:e.data+e.size], depth+1, objectOptions{})
			if sig != "" {
				return sig, true, nil
			}
			if err != nil {
				incomplete = err
			}
		}
		headerSize := 26
		switch kind {
		case cpioODC:
			headerSize = 76
		case cpioNewc:
			headerSize = 110
		}
		if e.next >= len(data) || len(data)-e.next < headerSize {
			// Every full member has been scanned. ClamAV's cpio readers
			// end their loop when a complete next header cannot be read;
			// they do not require a trailer member.
			return "", true, incomplete
		}
		if e, ok = cpioParse(data, e.next, kind); !ok {
			return "", true, ErrMalformed
		}
	}
}
