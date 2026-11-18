// Package chm reads Microsoft compiled HTML help files (ITSF) after the
// public descriptions of the format: the ITSF header and its two header
// sections, the ITSP directory with its PMGL listing chunks (names and
// section, offset and length as variable-length integers), the stored
// content (section 0) and the LZX-compressed one (section 1, its stream,
// control data and reset table being system files of section 0).
package chm

import (
	"encoding/binary"
	"errors"
	"strings"

	"github.com/utmstack/UTMStack/agent/edr/lzx"
)

var (
	// ErrFormat: the file or its directory cannot be read.
	ErrFormat = errors.New("chm: malformed file")
	// ErrLimit: a size beyond the caller's limit.
	ErrLimit = errors.New("chm: size limit")
)

// File is a directory entry.
type File struct {
	Name    string
	Section uint64
	Offset  uint64
	Length  uint64
}

// Archive is an opened CHM file.
type Archive struct {
	data    []byte
	content uint64 // where section 0 starts
	// Files are the regular entries in directory order; System the entries
	// whose names start with "::".
	Files  []File
	System map[string]File

	limit     uint64
	section1  []byte
	sec1Err   error
	sec1Built bool
}

const (
	contentName = "::DataSpace/Storage/MSCompressed/Content"
	controlName = "::DataSpace/Storage/MSCompressed/ControlData"
	resetName   = "::DataSpace/Storage/MSCompressed/Transform/{7FC28940-9D31-11D0-9B27-00A0C91E9C7C}/InstanceData/ResetTable"
)

// Open reads the headers and directory of data; limit bounds the
// decompressed section and the number of entries times 64 bytes.
func Open(data []byte, limit uint64) (*Archive, error) {
	le := binary.LittleEndian
	if len(data) < 0x58 || string(data[:4]) != "ITSF" {
		return nil, ErrFormat
	}
	version := le.Uint32(data[4:])
	headerLen := le.Uint32(data[8:])
	if version < 2 || version > 3 || uint64(headerLen) > uint64(len(data)) {
		return nil, ErrFormat
	}
	dirOff, dirLen := le.Uint64(data[0x48:]), le.Uint64(data[0x50:])
	var content uint64
	if version >= 3 {
		if len(data) < 0x60 {
			return nil, ErrFormat
		}
		content = le.Uint64(data[0x58:])
	} else {
		content = dirOff + dirLen
	}
	if dirOff > uint64(len(data)) || dirLen > uint64(len(data))-dirOff || dirLen < 0x54 {
		return nil, ErrFormat
	}
	dir := data[dirOff : dirOff+dirLen]
	if string(dir[:4]) != "ITSP" {
		return nil, ErrFormat
	}
	itspLen := le.Uint32(dir[8:])
	chunkSize := le.Uint32(dir[0x10:])
	chunks := le.Uint32(dir[0x2c:])
	if itspLen < 0x54 || uint64(itspLen) > dirLen || chunkSize < 0x20 || chunkSize > 1<<20 ||
		uint64(chunks)*uint64(chunkSize) > dirLen-uint64(itspLen) {
		return nil, ErrFormat
	}
	a := &Archive{data: data, content: content, System: map[string]File{}, limit: limit}
	for c := uint32(0); c < chunks; c++ {
		chunk := dir[uint64(itspLen)+uint64(c)*uint64(chunkSize):][:chunkSize]
		if string(chunk[:4]) != "PMGL" {
			continue // index (PMGI) chunks
		}
		free := le.Uint32(chunk[4:])
		if free > chunkSize-0x14 {
			return nil, ErrFormat
		}
		entries := chunk[0x14 : chunkSize-free]
		for p := 0; p < len(entries); {
			n, ok := encint(entries, &p)
			if !ok || n > uint64(len(entries)-p) {
				return nil, ErrFormat
			}
			name := string(entries[p : p+int(n)])
			p += int(n)
			var vals [3]uint64
			for i := range vals {
				if vals[i], ok = encint(entries, &p); !ok {
					return nil, ErrFormat
				}
			}
			f := File{Name: name, Section: vals[0], Offset: vals[1], Length: vals[2]}
			switch {
			case strings.HasPrefix(name, "::"):
				a.System[name] = f
			case strings.HasSuffix(name, "/") && f.Length == 0:
				// directory entries
			default:
				if uint64(len(a.Files)+1)*64 > limit {
					return nil, ErrLimit
				}
				a.Files = append(a.Files, f)
			}
		}
	}
	return a, nil
}

// encint reads a CHM variable-length integer: seven bits per byte, most
// significant first, the high bit set on all but the last byte.
func encint(b []byte, p *int) (uint64, bool) {
	var v uint64
	for i := 0; i < 9; i++ {
		if *p >= len(b) {
			return 0, false
		}
		c := b[*p]
		*p++
		v = v<<7 | uint64(c&0x7f)
		if c&0x80 == 0 {
			return v, true
		}
	}
	return 0, false
}

// stored returns a section 0 file's bytes.
func (a *Archive) stored(f File) ([]byte, error) {
	start := a.content + f.Offset
	if start < a.content || start > uint64(len(a.data)) || f.Length > uint64(len(a.data))-start {
		return nil, ErrFormat
	}
	return a.data[start : start+f.Length], nil
}

// Read returns a file's bytes.
func (a *Archive) Read(f File) ([]byte, error) {
	switch f.Section {
	case 0:
		return a.stored(f)
	case 1:
		sec, err := a.compressed()
		if err != nil && sec == nil {
			return nil, err
		}
		if f.Offset > uint64(len(sec)) || f.Length > uint64(len(sec))-f.Offset {
			if err == nil {
				err = ErrFormat
			}
			return nil, err
		}
		return sec[f.Offset : f.Offset+f.Length], nil
	}
	return nil, ErrFormat
}

// Decoded is the number of section 1 bytes decoded so far.
func (a *Archive) Decoded() int { return len(a.section1) }

// compressed decodes section 1 once: its length from the reset table, its
// window and reset interval from the LZXC control data.
func (a *Archive) compressed() ([]byte, error) {
	if a.sec1Built {
		return a.section1, a.sec1Err
	}
	a.sec1Built = true
	le := binary.LittleEndian
	fail := func(err error) ([]byte, error) { a.sec1Err = err; return nil, err }
	cf, ok1 := a.System[contentName]
	ctl, ok2 := a.System[controlName]
	rt, ok3 := a.System[resetName]
	if !ok1 || !ok2 || !ok3 || cf.Section != 0 || ctl.Section != 0 || rt.Section != 0 {
		return fail(ErrFormat)
	}
	content, err := a.stored(cf)
	if err != nil {
		return fail(err)
	}
	control, err := a.stored(ctl)
	if err != nil || len(control) < 0x18 || string(control[4:8]) != "LZXC" {
		return fail(ErrFormat)
	}
	reset, err := a.stored(rt)
	if err != nil || len(reset) < 0x28 {
		return fail(ErrFormat)
	}
	version := le.Uint32(control[8:])
	interval, window := le.Uint32(control[12:]), le.Uint32(control[16:])
	if version == 2 {
		interval, window = interval*0x8000, window*0x8000
	}
	bits := 0
	for b := 15; b <= 21; b++ {
		if window == 1<<b {
			bits = b
		}
	}
	if bits == 0 || interval == 0 || interval%0x8000 != 0 {
		return fail(ErrFormat)
	}
	length := le.Uint64(reset[0x10:])
	if length > a.limit {
		return fail(ErrLimit)
	}
	out, err := lzx.Decode(content, bits, int(length), int(interval/0x8000))
	a.section1, a.sec1Err = out, err
	return out, err
}
