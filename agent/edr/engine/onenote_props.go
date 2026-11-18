package engine

import (
	"encoding/binary"
	"fmt"
)

// A bounded MS-ONESTORE property-set reader. It validates the object stream
// and nested property values before a referenced object can be used as an
// attachment container. The caller supplies only a declared ObjectData body.
type oneProps struct {
	data   []byte
	pos    int
	values int
}

func (p *oneProps) take(n int) ([]byte, bool) {
	if n < 0 || n > len(p.data)-p.pos {
		return nil, false
	}
	b := p.data[p.pos : p.pos+n]
	p.pos += n
	return b, true
}
func (p *oneProps) u16() (uint16, bool) {
	b, ok := p.take(2)
	if !ok {
		return 0, false
	}
	return binary.LittleEndian.Uint16(b), true
}
func (p *oneProps) u32() (uint32, bool) {
	b, ok := p.take(4)
	if !ok {
		return 0, false
	}
	return binary.LittleEndian.Uint32(b), true
}
func (p *oneProps) stream() (uint32, bool, bool, bool) {
	h, ok := p.u32()
	if !ok {
		return 0, false, false, false
	}
	n := h & 0xffffff
	if n > uint32((len(p.data)-p.pos)/4) {
		return 0, false, false, false
	}
	_, ok = p.take(int(n) * 4)
	return n, h&(1<<30) != 0, h&(1<<31) != 0, ok
}
func (p *oneProps) set(depth int) bool {
	if depth >= 16 || p.values > 1_000_000 {
		return false
	}
	count, ok := p.u16()
	if !ok {
		return false
	}
	ids, ok := p.take(int(count) * 4)
	if !ok {
		return false
	}
	p.values += int(count)
	for i := 0; i < int(count); i++ {
		id := binary.LittleEndian.Uint32(ids[4*i:])
		kind := (id >> 26) & 0x1f
		switch kind {
		case 1, 2, 8, 10, 12: // value carried in property ID or reference array
		case 3:
			_, ok = p.take(1)
		case 4:
			_, ok = p.take(2)
		case 5, 9, 11, 13:
			_, ok = p.take(4)
		case 6:
			_, ok = p.take(8)
		case 7:
			var n uint32
			n, ok = p.u32()
			if ok && n <= uint32(len(p.data)-p.pos) {
				_, ok = p.take(int(n))
			} else {
				ok = false
			}
		case 16:
			var n uint32
			n, ok = p.u32()
			if !ok || n > uint32((len(p.data)-p.pos)/2) {
				return false
			}
			_, ok = p.u32()
			if !ok {
				return false
			} // property ID of nested values
			for j := uint32(0); j < n; j++ {
				if !p.set(depth + 1) {
					return false
				}
			}
		case 17:
			ok = p.set(depth + 1)
		default:
			return false
		}
		if !ok {
			return false
		}
	}
	return true
}

func parseOneObjectProps(data []byte) (objectIDs, cellIDs int, err error) {
	p := oneProps{data: data}
	objects, _, noOSID, ok := p.stream()
	if !ok {
		return 0, 0, fmt.Errorf("object stream: %w", ErrMalformed)
	}
	var spaces, contexts uint32
	if !noOSID {
		var extended bool
		spaces, extended, _, ok = p.stream()
		if !ok {
			return 0, 0, fmt.Errorf("object space stream: %w", ErrMalformed)
		}
		if extended {
			contexts, _, _, ok = p.stream()
			if !ok {
				return 0, 0, fmt.Errorf("context stream: %w", ErrMalformed)
			}
		}
	}
	if !p.set(0) {
		return 0, 0, fmt.Errorf("property values at %d/%d prefix %x: %w", p.pos, len(data), data[:min(len(data), 16)], ErrMalformed)
	}
	return int(objects), int(spaces + contexts), nil
}

// All declarations of an object are keyed by the same ExGuid and partition.
// A live object needs metadata (4) and property data (1); partition 2 carries
// optional file data through a BLOB reference.
func validateFSSGroupObjects(group fssGroup) error {
	byID := map[fssID]map[uint64]fssObjectPart{}
	for _, part := range group.parts {
		parts := byID[part.id]
		if parts == nil {
			parts = map[uint64]fssObjectPart{}
			byID[part.id] = parts
		}
		parts[part.partition] = part
	}
	for id, parts := range byID {
		metadata, ok := parts[4]
		if !ok || metadata.kind != 0x16 || len(metadata.data) < 4 {
			return fmt.Errorf("object %v metadata: %w", id, ErrMalformed)
		}
		data, ok := parts[1]
		if !ok || data.kind != 0x16 {
			return fmt.Errorf("object %v data partition: %w", id, ErrMalformed)
		}
		if file, ok := parts[2]; ok && file.kind != 0x1c {
			return fmt.Errorf("object %v file partition: %w", id, ErrMalformed)
		}
		jcid := binary.LittleEndian.Uint32(metadata.data)
		if jcid == 0x00020030 { // validate cached title when present
			fields, err := oneParseObjectFields(data.data)
			if err != nil || oneOptionalUTF16(fields, 0x1c001cf3) != nil {
				return fmt.Errorf("object %v page metadata title: %w", id, ErrMalformed)
			}
		}
		// Other page/content objects may be superseded or unreferenced;
		// the semantic walker validates them when traversed.
		if jcid != 0x00120001 {
			continue
		}
		objects, cells, err := parseOneObjectProps(data.data)
		if err != nil {
			return fmt.Errorf("object %v property set: %w", id, err)
		}
		if objects < data.objectRefs || cells != data.cellRefs {
			return fmt.Errorf("object %v reference counts: %w", id, ErrMalformed)
		}
	}
	return nil
}
