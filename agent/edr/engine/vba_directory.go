package engine

import (
	"context"
	"encoding/binary"
	"strings"
	"unicode/utf16"
)

// MS-OVBA 2.3.4.2. The directory is a structured sequence, not a search for
// MODULEOFFSET-looking bytes. In particular PROJECTVERSION and references have
// fields outside their nominal size and cannot be skipped as generic TLVs.
type oleVBAProject struct {
	modules map[string]uint32
	seen    map[string]bool
}
type vbaDirReader struct {
	ctx          context.Context
	b            []byte
	pos, records int
	err          error
}

func (r *vbaDirReader) take(n int) []byte {
	if r.err != nil {
		return nil
	}
	if n < 0 || n > len(r.b)-r.pos {
		r.err = ErrMalformed
		return nil
	}
	b := r.b[r.pos : r.pos+n]
	r.pos += n
	return b
}
func (r *vbaDirReader) u16() uint16 {
	b := r.take(2)
	if b == nil {
		return 0
	}
	return binary.LittleEndian.Uint16(b)
}
func (r *vbaDirReader) u32() uint32 {
	b := r.take(4)
	if b == nil {
		return 0
	}
	return binary.LittleEndian.Uint32(b)
}
func (r *vbaDirReader) blob() []byte {
	n := r.u32()
	if uint64(n) > uint64(len(r.b)-r.pos) {
		r.err = ErrMalformed
		return nil
	}
	return r.take(int(n))
}
func (r *vbaDirReader) peek() uint16 {
	if len(r.b)-r.pos < 2 {
		r.err = ErrMalformed
		return 0
	}
	return binary.LittleEndian.Uint16(r.b[r.pos:])
}
func (r *vbaDirReader) record(id uint16, size int) []byte {
	if r.err != nil {
		return nil
	}
	r.records++
	if r.records > 65536 {
		r.err = ErrLimit
		return nil
	}
	if err := r.ctx.Err(); err != nil {
		r.err = err
		return nil
	}
	if r.u16() != id {
		r.err = ErrMalformed
		return nil
	}
	b := r.blob()
	if size >= 0 && len(b) != size {
		r.err = ErrMalformed
	}
	return b
}
func (r *vbaDirReader) paired(id, unicodeID uint16) []byte {
	r.record(id, -1)
	b := r.record(unicodeID, -1)
	if len(b)%2 != 0 {
		r.err = ErrMalformed
	}
	return b
}
func (r *vbaDirReader) reference() {
	if r.peek() == 0x16 {
		r.paired(0x16, 0x3e)
	}
	id := r.u16()
	switch id {
	case 0x33: // REFERENCEORIGINAL precedes a REFERENCECONTROL.
		r.blob()
		if r.peek() != 0x2f {
			r.err = ErrMalformed
			return
		}
		r.reference()
	case 0x0d:
		r.u32()
		r.blob()
		r.take(6)
	case 0x0e:
		r.u32()
		r.blob()
		r.blob()
		r.take(6)
	case 0x2f:
		r.u32()
		r.blob()
		r.take(6)
		if r.peek() == 0x16 {
			r.paired(0x16, 0x3e)
		}
		if r.u16() != 0x30 {
			r.err = ErrMalformed
			return
		}
		r.u32()
		r.blob()
		r.take(26)
	default:
		r.err = ErrUnsupported
	}
}

func parseVBADirectory(ctx context.Context, data []byte, maxModules int) (*oleVBAProject, error) {
	r := &vbaDirReader{ctx: ctx, b: data}
	r.record(0x01, 4)
	if r.peek() == 0x4a {
		r.record(0x4a, 4)
	}
	r.record(0x02, 4)
	r.record(0x14, 4)
	r.record(0x03, 2)
	r.record(0x04, -1)
	r.paired(0x05, 0x40)
	r.record(0x06, -1)
	r.record(0x3d, -1)
	r.record(0x07, 4)
	r.record(0x08, 4)
	r.record(0x09, 4)
	r.take(2) // VersionMinor follows the four-byte VersionMajor.
	if r.peek() == 0x0c {
		r.paired(0x0c, 0x3c)
	}
	for r.err == nil && r.peek() != 0x0f {
		r.records++
		if r.records > 65536 {
			r.err = ErrLimit
			break
		}
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		r.reference()
	}
	b := r.record(0x0f, 2)
	if r.err != nil {
		return nil, r.err
	}
	n := int(binary.LittleEndian.Uint16(b))
	if n > maxModules {
		return nil, ErrLimit
	}
	r.record(0x13, 2)
	project := &oleVBAProject{modules: make(map[string]uint32, n), seen: make(map[string]bool, n)}
	for i := 0; i < n && r.err == nil; i++ {
		r.record(0x19, -1)
		if r.peek() == 0x47 {
			r.record(0x47, -1)
		}
		nameData := r.paired(0x1a, 0x32)
		name, err := vbaStreamName(nameData)
		if err != nil {
			return nil, err
		}
		r.paired(0x1c, 0x48)
		offset := r.record(0x31, 4)
		if r.err != nil {
			return nil, r.err
		}
		if _, exists := project.modules[name]; exists {
			return nil, ErrMalformed
		}
		project.modules[name] = binary.LittleEndian.Uint32(offset)
		r.record(0x1e, 4)
		r.record(0x2c, 2)
		switch r.peek() {
		case 0x21, 0x22:
			r.record(r.peek(), 0)
		default:
			r.err = ErrMalformed
		}
		if r.peek() == 0x25 {
			r.record(0x25, 0)
		}
		if r.peek() == 0x28 {
			r.record(0x28, 0)
		}
		r.record(0x2b, 0)
	}
	r.record(0x10, 0)
	if r.err != nil {
		return nil, r.err
	}
	if r.pos != len(data) {
		return nil, ErrMalformed
	}
	return project, nil
}
func vbaStreamName(b []byte) (string, error) {
	if len(b) == 0 || len(b) > 62 || len(b)%2 != 0 {
		return "", ErrMalformed
	}
	u := make([]uint16, len(b)/2)
	for i := range u {
		u[i] = binary.LittleEndian.Uint16(b[i*2:])
		if u[i] == 0 {
			return "", ErrMalformed
		}
	}
	for i := 0; i < len(u); i++ {
		if u[i] >= 0xd800 && u[i] <= 0xdbff {
			if i+1 >= len(u) || u[i+1] < 0xdc00 || u[i+1] > 0xdfff {
				return "", ErrMalformed
			}
			i++
		} else if u[i] >= 0xdc00 && u[i] <= 0xdfff {
			return "", ErrMalformed
		}
	}
	name := string(utf16.Decode(u))
	if strings.ContainsAny(name, "/\\:!") {
		return "", ErrMalformed
	}
	return strings.ToUpper(name), nil
}
