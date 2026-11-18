package engine

import "encoding/binary"

// Compact IDs in an object property stream are positional references to the
// FSSHTTPB object/cell arrays carried by the corresponding ObjectData record.
// Property order determines each reference's offset in those arrays.
type oneField struct {
	id, kind, number      uint32
	raw                   []byte
	objectRefs, spaceRefs int
}
type oneFields struct {
	objects, spaces, contexts []uint32
	fields                    []oneField
}

func oneCompactStream(p *oneProps) ([]uint32, bool, bool, bool) {
	h, ok := p.u32()
	if !ok {
		return nil, false, false, false
	}
	n := int(h & 0xffffff)
	if n > (len(p.data)-p.pos)/4 {
		return nil, false, false, false
	}
	b, ok := p.take(n * 4)
	if !ok {
		return nil, false, false, false
	}
	out := make([]uint32, n)
	for i := range out {
		out[i] = binary.LittleEndian.Uint32(b[4*i:])
	}
	return out, h&(1<<30) != 0, h&(1<<31) != 0, true
}

func oneParseFields(p *oneProps, depth int) ([]oneField, bool) {
	if depth >= 16 || p.values > 1_000_000 {
		return nil, false
	}
	count, ok := p.u16()
	if !ok {
		return nil, false
	}
	ids, ok := p.take(int(count) * 4)
	if !ok {
		return nil, false
	}
	p.values += int(count)
	out := make([]oneField, 0, int(count))
	for i := 0; i < int(count); i++ {
		id := binary.LittleEndian.Uint32(ids[4*i:])
		kind := (id >> 26) & 0x1f
		field := oneField{id: id, kind: kind}
		switch kind {
		case 1, 2:
		case 3:
			field.raw, ok = p.take(1)
		case 4:
			field.raw, ok = p.take(2)
		case 5, 9, 11, 13:
			field.number, ok = p.u32()
			if kind == 9 {
				field.objectRefs = int(field.number)
			}
			if kind == 11 {
				field.spaceRefs = int(field.number)
			}
		case 6:
			field.raw, ok = p.take(8)
		case 7:
			var n uint32
			n, ok = p.u32()
			if !ok || n > uint32(len(p.data)-p.pos) {
				return nil, false
			}
			field.raw, ok = p.take(int(n))
		case 8:
			field.objectRefs = 1
		case 10:
			field.spaceRefs = 1
		case 12:
		case 16:
			n, valid := p.u32()
			if !valid || n > uint32((len(p.data)-p.pos)/2) {
				return nil, false
			}
			_, valid = p.u32()
			if !valid {
				return nil, false
			}
			for j := uint32(0); j < n; j++ {
				nested, valid := oneParseFields(p, depth+1)
				if !valid {
					return nil, false
				}
				for _, child := range nested {
					field.objectRefs += child.objectRefs
					field.spaceRefs += child.spaceRefs
				}
			}
		case 17:
			nested, valid := oneParseFields(p, depth+1)
			if !valid {
				return nil, false
			}
			for _, child := range nested {
				field.objectRefs += child.objectRefs
				field.spaceRefs += child.spaceRefs
			}
		default:
			return nil, false
		}
		if !ok {
			return nil, false
		}
		out = append(out, field)
	}
	return out, true
}

func oneParseObjectFields(data []byte) (oneFields, error) {
	if _, _, err := parseOneObjectProps(data); err != nil {
		return oneFields{}, err
	}
	p := oneProps{data: data}
	objects, _, noOSID, ok := oneCompactStream(&p)
	if !ok {
		return oneFields{}, ErrMalformed
	}
	var spaces, contexts []uint32
	if !noOSID {
		var extended bool
		spaces, extended, _, ok = oneCompactStream(&p)
		if !ok {
			return oneFields{}, ErrMalformed
		}
		if extended {
			contexts, _, _, ok = oneCompactStream(&p)
			if !ok {
				return oneFields{}, ErrMalformed
			}
		}
	}
	fields, ok := oneParseFields(&p, 0)
	if !ok {
		return oneFields{}, ErrMalformed
	}
	return oneFields{objects: objects, spaces: spaces, contexts: contexts, fields: fields}, nil
}

func (f oneFields) find(id uint32) (oneField, int, bool) {
	needle := id & 0x3ffffff
	for i := len(f.fields) - 1; i >= 0; i-- {
		if f.fields[i].id&0x3ffffff == needle {
			return f.fields[i], i, true
		}
	}
	return oneField{}, 0, false
}
func (f oneFields) objectOffset(index int) int {
	n := 0
	for i := 0; i < index; i++ {
		n += f.fields[i].objectRefs
	}
	return n
}
func (f oneFields) spaceOffset(index int) int {
	n := 0
	for i := 0; i < index; i++ {
		n += f.fields[i].spaceRefs
	}
	return n
}

func oneResolveObject(ids []uint32, targets []fssID, index int) (fssID, bool) {
	if index >= len(ids) {
		return fssID{}, false
	}
	compact := ids[index]
	var found fssID
	matches := 0
	for j := 0; j < len(targets) && j < len(ids); j++ {
		if ids[j] == compact {
			found = targets[j]
			matches++
		}
	}
	if matches == 1 {
		return found, true
	}
	if index < len(targets) && ids[index] == compact {
		return targets[index], true
	}
	return fssID{}, false
}
func oneResolveCell(ids []uint32, targets []fssCell, index int) (fssCell, bool) {
	if index >= len(ids) {
		return fssCell{}, false
	}
	compact := ids[index]
	var found fssCell
	matches := 0
	for j := 0; j < len(targets) && j < len(ids); j++ {
		if ids[j] == compact {
			found = targets[j]
			matches++
		}
	}
	if matches == 1 {
		return found, true
	}
	if index < len(targets) && ids[index] == compact {
		return targets[index], true
	}
	return fssCell{}, false
}

func (f oneFields) objectReferences(obj fssObject, property uint32) ([]fssID, error) {
	field, index, ok := f.find(property)
	if !ok {
		return nil, nil
	}
	if field.kind != 8 && field.kind != 9 {
		return nil, ErrMalformed
	}
	offset := f.objectOffset(index)
	if offset > len(f.objects) {
		return nil, nil
	}
	// The OneStore object stream is authoritative. An array count can be
	// larger than its remaining compact IDs; the upstream OneNote reader
	// takes only IDs that exist in the stream. Preserve slots for later
	// single-reference fields, which cannot be part of this array.
	available := len(f.objects) - offset
	if field.objectRefs > available {
		for _, later := range f.fields[index+1:] {
			if later.kind == 8 {
				available--
			}
		}
		if available < 0 {
			return nil, ErrMalformed
		}
	}
	count := min(field.objectRefs, available)
	out := make([]fssID, 0, count)
	for i := 0; i < count; i++ {
		if id, ok := oneResolveObject(f.objects, obj.objectIDs, offset+i); ok {
			out = append(out, id)
		}
	}
	return out, nil
}
func (f oneFields) spaceReferences(obj fssObject, space fssSpace, property uint32) ([]fssCell, error) {
	field, index, ok := f.find(property)
	if !ok {
		return nil, nil
	}
	if field.kind != 10 && field.kind != 11 {
		return nil, ErrMalformed
	}
	offset := f.spaceOffset(index)
	if field.spaceRefs > len(f.spaces)-offset {
		return nil, ErrMalformed
	}
	var targets []fssCell
	for _, cell := range obj.cellIDs {
		if cell.second != space.cell.second {
			targets = append(targets, cell)
		}
	}
	out := make([]fssCell, 0, field.spaceRefs)
	for i := 0; i < field.spaceRefs; i++ {
		cell, ok := oneResolveCell(f.spaces, targets, i)
		if !ok {
			return nil, ErrMalformed
		}
		out = append(out, cell)
	}
	return out, nil
}
