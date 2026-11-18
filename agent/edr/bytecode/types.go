package bytecode

// The wire kinds and layout below describe the ClamBC interpreter ABI, not Go
// structs or the host C ABI. In particular, the upstream interpreter packs both
// struct kinds by summing their field sizes and represents pointers in 8 bytes.
// References: libclamav/bytecode.c (parseTypes, typesize, typealign, calc_gepz)
// and libclamav/bytecode_api_decl.c in https://github.com/Cisco-Talos/clamav.
const (
	functionType      = 1
	packedStructType  = 2
	structType        = 3
	arrayType         = 4
	pointerType       = 5
	firstCompoundType = 65
	firstDynamicType  = 69
	maxTypes          = 4096
	maxTypeMembers    = 4096
	maxTypeReferences = 65536
	maxTypeDepth      = 128
	maxTypeBytes      = 16 << 20
)

// elements holds function return type followed by parameters, struct members,
// or the single array element/pointer target. count is the array length; for
// all other kinds it equals len(elements). offsets is a struct's field layout.
type bcType struct {
	kind      int
	elements  []int
	count     int
	byteSize  int
	alignment int
	offsets   []int
}

// types is indexed by type ID minus 65. IDs 0..64 denote void and integer bit
// widths. The serialized count includes one unused sentinel; it is deliberately
// absent from this slice and cannot be referenced by a validated program.
type typeTable struct {
	types []bcType
}

func parseTypeTable(line []byte, count int) (*typeTable, error) {
	if count < 5 {
		return nil, ErrMalformed
	}
	if count > maxTypes || len(line) > maxProgramBytes {
		return nil, ErrLimit
	}
	c := &cursor{b: line}
	c.expect('T')
	start := c.fixed(2)
	if c.err != nil {
		return nil, c.err
	}
	if start != firstDynamicType {
		return nil, ErrUnsupported
	}
	t := &typeTable{types: make([]bcType, count-1)}
	for i, element := range []int{8, 16, 32, 64} {
		t.types[i] = bcType{kind: pointerType, elements: []int{element}, count: 1, byteSize: 8, alignment: 8}
	}
	references := 4
	for i := firstDynamicType - firstCompoundType; i < len(t.types); i++ {
		kind := int(c.fixed(1))
		if c.err != nil {
			return nil, c.err
		}
		value := bcType{kind: kind}
		subtypes := 1
		switch kind {
		case functionType, packedStructType, structType:
			subtypes = c.count(maxTypeMembers)
			value.count = subtypes
			if c.err != nil {
				return nil, c.err
			}
			if kind == functionType && subtypes == 0 {
				return nil, ErrMalformed
			}
		case arrayType:
			value.count = c.count(maxTypeBytes)
		case pointerType:
			value.count = 1
		default:
			return nil, ErrMalformed
		}
		if c.err != nil {
			return nil, c.err
		}
		if subtypes > maxTypeReferences-references {
			return nil, ErrLimit
		}
		// Each encoded reference needs at least its length prefix. Check before
		// allocating so truncated declarations cannot allocate large tables.
		if subtypes > len(c.b)-c.pos {
			return nil, ErrMalformed
		}
		references += subtypes
		value.elements = make([]int, subtypes)
		for j := range value.elements {
			id := c.number()
			if c.err != nil {
				return nil, c.err
			}
			if id >= uint64(firstCompoundType+len(t.types)) {
				return nil, ErrMalformed
			}
			value.elements[j] = int(id)
		}
		t.types[i] = value
	}
	if err := c.end(); err != nil {
		return nil, err
	}
	// A function is only a callable prototype or a pointer target. Void is
	// valid as a function result (and an opaque pointer target), never as a
	// stored struct/array element or a parameter.
	for _, ty := range t.types {
		for i, id := range ty.elements {
			if ty.kind == pointerType {
				continue
			}
			if id == 0 && !(ty.kind == functionType && i == 0) {
				return nil, ErrMalformed
			}
			if id >= firstCompoundType && t.types[id-firstCompoundType].kind == functionType {
				return nil, ErrMalformed
			}
		}
	}
	state := make([]byte, len(t.types))
	for id := firstCompoundType; id < firstCompoundType+len(t.types); id++ {
		if _, err := t.layout(id, state, 0); err != nil {
			return nil, err
		}
	}
	return t, nil
}

func integerSize(id int) int {
	switch {
	case id == 0:
		return 0
	case id <= 8:
		return 1
	case id <= 16:
		return 2
	case id <= 32:
		return 4
	default:
		return 8
	}
}

// normalize accepts the addressable-local flag understood by upstream
// typesize/typealign, but rejects any other high bits or undefined type IDs.
func (t *typeTable) normalize(id int) (int, error) {
	if t == nil || id < 0 || id > 0xffff {
		return 0, ErrMalformed
	}
	id &= 0x7fff
	if id >= firstCompoundType+len(t.types) {
		return 0, ErrMalformed
	}
	return id, nil
}

func (t *typeTable) size(id int) (int, error) {
	id, err := t.normalize(id)
	if err != nil {
		return 0, err
	}
	if id < firstCompoundType {
		return integerSize(id), nil
	}
	return t.types[id-firstCompoundType].byteSize, nil
}

func (t *typeTable) align(id int) (int, error) {
	id, err := t.normalize(id)
	if err != nil {
		return 0, err
	}
	if id < firstCompoundType {
		if id == 0 {
			return 1, nil
		}
		return integerSize(id), nil
	}
	return t.types[id-firstCompoundType].alignment, nil
}

func (t *typeTable) pointer(id int) bool {
	id, err := t.normalize(id)
	return err == nil && id >= firstCompoundType && t.types[id-firstCompoundType].kind == pointerType
}

// element performs one typed GEP step. Pointer/array indices are signed and may
// form an out-of-bounds address, as in LLVM; the VM must validate the resulting
// virtual region and every access. Struct indices name a real field. No offset
// can exceed the maximum virtual object size in either direction.
func (t *typeTable) element(id, index int) (nextType, byteOffset int, err error) {
	id, err = t.normalize(id)
	if err != nil {
		return 0, 0, err
	}
	if id < firstCompoundType {
		return 0, 0, ErrMalformed
	}
	ty := &t.types[id-firstCompoundType]
	switch ty.kind {
	case pointerType, arrayType:
		nextType = ty.elements[0]
		if nextType == 0 || nextType >= firstCompoundType && t.types[nextType-firstCompoundType].kind == functionType {
			return 0, 0, ErrUnsupported
		}
		stride, err := t.size(nextType)
		if err != nil {
			return 0, 0, err
		}
		if stride == 0 {
			return nextType, 0, nil
		}
		if index > maxTypeBytes/stride || index < -maxTypeBytes/stride {
			return 0, 0, ErrLimit
		}
		return nextType, index * stride, nil
	case packedStructType, structType:
		if index < 0 || index >= len(ty.elements) {
			return 0, 0, ErrMalformed
		}
		return ty.elements[index], ty.offsets[index], nil
	default:
		return 0, 0, ErrUnsupported
	}
}

func (t *typeTable) layout(id int, state []byte, depth int) (int, error) {
	if id < firstCompoundType {
		return integerSize(id), nil
	}
	i := id - firstCompoundType
	if state[i] == 2 {
		return t.types[i].byteSize, nil
	}
	if state[i] == 1 {
		return 0, ErrMalformed // infinite aggregate-by-value recursion
	}
	if depth >= maxTypeDepth {
		return 0, ErrLimit
	}
	state[i] = 1
	ty := &t.types[i]
	switch ty.kind {
	case functionType, pointerType:
		ty.byteSize, ty.alignment = 8, 8
	case packedStructType, structType:
		ty.alignment = 8
		ty.offsets = make([]int, len(ty.elements))
		for j, child := range ty.elements {
			size, err := t.layout(child, state, depth+1)
			if err != nil {
				return 0, err
			}
			if size > maxTypeBytes-ty.byteSize {
				return 0, ErrLimit
			}
			ty.offsets[j] = ty.byteSize
			ty.byteSize += size
		}
	case arrayType:
		size, err := t.layout(ty.elements[0], state, depth+1)
		if err != nil {
			return 0, err
		}
		if size != 0 && ty.count > maxTypeBytes/size {
			return 0, ErrLimit
		}
		ty.byteSize = ty.count * size
		ty.alignment, err = t.align(ty.elements[0])
		if err != nil {
			return 0, err
		}
	default:
		return 0, ErrMalformed
	}
	state[i] = 2
	return ty.byteSize, nil
}
