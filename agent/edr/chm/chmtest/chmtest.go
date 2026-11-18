// Package chmtest builds small Compiled HTML Help files for tests: an ITSF
// version 3 header, an ITSP directory with one PMGL chunk, stored files
// (section 0) and an LZX-compressed section 1 (literal-only, window 64 KiB,
// reset every two frames: keep section 1 under 64 KiB).
package chmtest

import (
	"encoding/binary"

	"github.com/utmstack/UTMStack/agent/edr/lzx/lzxtest"
)

// Entry is a file to store: Section 0 is stored, 1 compressed.
type Entry struct {
	Name    string
	Section int
	Data    []byte
}

func encint(v uint64) []byte {
	var b []byte
	for {
		b = append([]byte{byte(v & 0x7f)}, b...)
		v >>= 7
		if v == 0 {
			break
		}
	}
	for i := 0; i < len(b)-1; i++ {
		b[i] |= 0x80
	}
	return b
}

// Build writes the file.
func Build(entries []Entry) []byte {
	le := binary.LittleEndian
	type dirEntry struct {
		name                    string
		section, offset, length uint64
	}
	var dir []dirEntry
	var content, sec1 []byte
	for _, e := range entries {
		if e.Section == 0 {
			dir = append(dir, dirEntry{e.Name, 0, uint64(len(content)), uint64(len(e.Data))})
			content = append(content, e.Data...)
		} else {
			dir = append(dir, dirEntry{e.Name, 1, uint64(len(sec1)), uint64(len(e.Data))})
			sec1 = append(sec1, e.Data...)
		}
	}
	compressed := lzxtest.Literals(sec1, 16)
	control := make([]byte, 0x18)
	le.PutUint32(control[0:], 6)
	copy(control[4:], "LZXC")
	le.PutUint32(control[8:], 2)
	le.PutUint32(control[12:], 2) // reset every two frames
	le.PutUint32(control[16:], 2) // 64 KiB window
	le.PutUint32(control[20:], 2)
	reset := make([]byte, 0x30)
	le.PutUint32(reset[0:], 2)
	le.PutUint32(reset[4:], 1)
	le.PutUint32(reset[8:], 8)
	le.PutUint32(reset[12:], 0x28)
	le.PutUint64(reset[0x10:], uint64(len(sec1)))
	le.PutUint64(reset[0x18:], uint64(len(compressed)))
	le.PutUint64(reset[0x20:], 0x8000)
	for _, sys := range []struct {
		name string
		data []byte
	}{
		{"::DataSpace/Storage/MSCompressed/Content", compressed},
		{"::DataSpace/Storage/MSCompressed/ControlData", control},
		{"::DataSpace/Storage/MSCompressed/Transform/{7FC28940-9D31-11D0-9B27-00A0C91E9C7C}/InstanceData/ResetTable", reset},
	} {
		dir = append(dir, dirEntry{sys.name, 0, uint64(len(content)), uint64(len(sys.data))})
		content = append(content, sys.data...)
	}
	const chunkSize = 0x1000
	chunk := make([]byte, chunkSize)
	copy(chunk, "PMGL")
	le.PutUint32(chunk[0x0c:], 0xffffffff)
	le.PutUint32(chunk[0x10:], 0xffffffff)
	p := 0x14
	for _, d := range dir {
		var e []byte
		e = append(e, encint(uint64(len(d.name)))...)
		e = append(e, d.name...)
		e = append(e, encint(d.section)...)
		e = append(e, encint(d.offset)...)
		e = append(e, encint(d.length)...)
		p += copy(chunk[p:], e)
	}
	le.PutUint32(chunk[4:], uint32(chunkSize-p))
	itsp := make([]byte, 0x54)
	copy(itsp, "ITSP")
	le.PutUint32(itsp[4:], 1)
	le.PutUint32(itsp[8:], 0x54)
	le.PutUint32(itsp[0x0c:], 0x0a)
	le.PutUint32(itsp[0x10:], chunkSize)
	le.PutUint32(itsp[0x14:], 2)
	le.PutUint32(itsp[0x18:], 1)
	le.PutUint32(itsp[0x1c:], 0xffffffff)
	le.PutUint32(itsp[0x28:], 0xffffffff)
	le.PutUint32(itsp[0x2c:], 1)
	le.PutUint32(itsp[0x44:], 0x54)
	directory := append(itsp, chunk...)
	head := make([]byte, 0x60+0x18)
	copy(head, "ITSF")
	le.PutUint32(head[4:], 3)
	le.PutUint32(head[8:], 0x60)
	le.PutUint32(head[0x0c:], 1)
	le.PutUint64(head[0x38:], 0x60)
	le.PutUint64(head[0x40:], 0x18)
	dirOff := uint64(len(head))
	le.PutUint64(head[0x48:], dirOff)
	le.PutUint64(head[0x50:], uint64(len(directory)))
	le.PutUint64(head[0x58:], dirOff+uint64(len(directory)))
	le.PutUint32(head[0x60:], 0x1fe)
	out := append(append(head, directory...), content...)
	le.PutUint64(out[0x68:], uint64(len(out)))
	return out
}
