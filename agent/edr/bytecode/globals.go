package bytecode

import "encoding/binary"

const (
	maxGlobals          = 4096
	maxGlobalBytes      = 16 << 20
	maxGlobalComponents = 65536
	// These wire IDs and fixed layouts are declared in bytecode_api.h and
	// clambc.h; the detection-name table (0x8002) and the PE header record
	// (0x8003) are in pe.go.
	globalMatchCounts  = 0x8000
	globalKind         = 0x8001
	globalFileSize     = 0x8004
	globalMatchOffsets = 0x8005
)

// specialGlobalSize is the complete supported scanner-global set. Runtime
// exposes these fixed-size objects as read-only little-endian VM regions.
func specialGlobalSize(id int) (int, bool) {
	switch id {
	case globalMatchCounts, globalMatchOffsets:
		return 64 * 4, true
	case globalKind:
		return 2, true
	case globalFileSize:
		return 4, true
	case globalPEData:
		return peRecordSize, true
	default:
		return 0, false
	}
}

// NeedsMatchContext reports whether execution requires actual logical-signature
// match arrays from the scanner. A declared but unused capability needs no
// context. Kind and file size derive from program metadata and input bytes.
func (p *Program) NeedsMatchContext() bool {
	if p == nil {
		return false
	}
	for _, g := range p.globals {
		for _, ref := range g.pointers {
			if ref.target == globalMatchCounts || ref.target == globalMatchOffsets {
				return true
			}
		}
	}
	return false
}

// data contains array initializers and zeroed pointer slots. Each pointer is
// resolved against a fresh virtual memory allocation for every Run; these
// records never contain host addresses. Global zero is the null object.
// unresolved marks a pointer global whose interpreter value has no native
// equivalent: its bytes stay uninitialized, so reading them faults.
type global struct {
	typeID     int
	data       []byte
	pointers   []globalPointer
	unresolved bool
}

type globalPointer struct {
	offset       int
	target       int
	targetOffset uint64
}

// parseGlobals decodes the flattened constant initializer format described by
// libclamav/bytecode.c (type_components, readConstant, parseGlobals), then
// applies it as that interpreter does (see initializeGlobal). Integers use the
// little-endian ClamBC data layout, independently of the host byte order.
func parseGlobals(line []byte, tt *typeTable) ([]global, error) {
	if tt == nil {
		return nil, ErrMalformed
	}
	if len(line) > maxProgramBytes {
		return nil, ErrLimit
	}
	c := &cursor{b: line}
	c.expect('G')
	maxGlobal := c.number()
	if c.err != nil {
		return nil, c.err
	}
	// The reference only refuses a declared global above the highest one it
	// knows (bytecode.c parseGlobals).
	if maxGlobal != 0 && (maxGlobal < globalMatchCounts || maxGlobal > globalMatchOffsets) {
		return nil, ErrUnsupported
	}
	n := c.count(maxGlobals)
	if c.err != nil {
		return nil, c.err
	}
	// Every global needs a type ID and an initializer terminator.
	if n > (len(c.b)-c.pos)/2 {
		return nil, ErrMalformed
	}
	globals := make([]global, n)
	counts := make(map[int]int)
	totalBytes, totalComponents := 0, 0
	for i := range globals {
		rawID := c.number()
		if c.err != nil {
			return nil, c.err
		}
		if rawID >= uint64(firstCompoundType+len(tt.types)) {
			return nil, ErrMalformed
		}
		id := int(rawID)
		if i == 0 && id != 0 || i != 0 && id == 0 {
			return nil, ErrMalformed
		}
		size, err := tt.size(id)
		if err != nil {
			return nil, err
		}
		if size > maxGlobalBytes-totalBytes {
			return nil, ErrLimit
		}
		components, err := tt.globalComponents(id, counts, 0)
		if err != nil {
			return nil, err
		}
		if components > maxGlobalComponents-totalComponents {
			return nil, ErrLimit
		}
		totalBytes += size
		totalComponents += components
		globals[i] = global{typeID: id, data: make([]byte, size)}
		if c.peek() == '@' && c.pos+1 < len(c.b) && c.b[c.pos+1] == '`' {
			c.pos += 2
			// A compact zero initializer stays zero, except that the
			// interpreter composes a zero pointer global like (0, 0).
			globals[i].unresolved = tt.pointer(id)
			continue
		}
		if i == 0 {
			return nil, ErrMalformed // the reserved null object must be zero-init
		}
		values := make([]uint64, components)
		for j := range values {
			prefix := c.byte()
			if prefix < 0x40 || prefix > 0x50 {
				return nil, ErrMalformed
			}
			values[j] = c.fixed(int(prefix - 0x40))
			if c.err != nil {
				return nil, c.err
			}
		}
		c.expect('`')
		if c.err != nil {
			return nil, c.err
		}
		if err := tt.initializeGlobal(&globals[i], id, values); err != nil {
			return nil, err
		}
	}
	if err := c.end(); err != nil {
		return nil, err
	}
	// Forward and self references are valid. The target must exist and offsets
	// may reach its one-past address, but can never wrap or leave its allocation.
	for _, g := range globals {
		for _, p := range g.pointers {
			if p.target >= 0x8000 {
				if uint64(p.target) > maxGlobal {
					return nil, ErrMalformed
				}
				if p.target == globalVirusNames {
					continue // never registered: any offset composes, no access succeeds
				}
				size, known := specialGlobalSize(p.target)
				if !known {
					return nil, ErrUnsupported
				}
				if p.targetOffset > uint64(size) {
					return nil, ErrMalformed
				}
				continue
			}
			if p.target >= len(globals) || p.targetOffset > uint64(len(globals[p.target].data)) {
				return nil, ErrMalformed
			}
		}
	}
	return globals, nil
}

func (tt *typeTable) globalComponents(id int, memo map[int]int, depth int) (int, error) {
	if id < firstCompoundType {
		return 1, nil
	}
	if count, ok := memo[id]; ok {
		return count, nil
	}
	if depth >= maxTypeDepth {
		return 0, ErrLimit
	}
	ty := &tt.types[id-firstCompoundType]
	count := 0
	switch ty.kind {
	case pointerType:
		count = 2
	case packedStructType, structType:
		for _, child := range ty.elements {
			childCount, err := tt.globalComponents(child, memo, depth+1)
			if err != nil {
				return 0, err
			}
			if childCount > maxGlobalComponents-count {
				return 0, ErrLimit
			}
			count += childCount
		}
	case arrayType:
		childCount, err := tt.globalComponents(ty.elements[0], memo, depth+1)
		if err != nil {
			return 0, err
		}
		if childCount != 0 && ty.count > maxGlobalComponents/childCount {
			return 0, ErrLimit
		}
		count = ty.count * childCount
	case functionType:
		return 0, ErrUnsupported
	default:
		return 0, ErrMalformed
	}
	memo[id] = count
	return count, nil
}

// initializeGlobal applies decoded initializer components the way the pinned
// interpreter's cli_bytecode_prepare_interpreter does. That is the reference
// even where it leaves initialization unimplemented: integer and struct
// globals stay zero, and a pointer global is composed from its (offset,
// global) pair. The interpreter composes the (0, 0) null spelling into a
// pointer to the start of its shared global block, which has no native
// equivalent, so that value is left unreadable. An array whose elements are
// 1, 2, 4 or 8 bytes receives its first count raw components truncated to the
// element size, so an array of pointers holds raw offset and ID numbers, not
// usable pointers. Arrays of other element sizes stay zero.
func (tt *typeTable) initializeGlobal(g *global, id int, values []uint64) error {
	if id < firstCompoundType {
		return nil
	}
	ty := &tt.types[id-firstCompoundType]
	switch ty.kind {
	case pointerType:
		if len(values) != 2 || len(g.data) != 8 {
			return ErrMalformed
		}
		address, target := values[0], values[1]
		if target >= 0x8000 {
			if target > globalMatchOffsets {
				return ErrUnsupported
			}
			if _, known := specialGlobalSize(int(target)); !known && target != globalVirusNames {
				return ErrUnsupported
			}
		}
		if address == 0 && target == 0 {
			g.unresolved = true
		} else {
			g.pointers = append(g.pointers, globalPointer{target: int(target), targetOffset: address})
		}
	case arrayType:
		size, err := tt.size(ty.elements[0])
		if err != nil {
			return err
		}
		if size != 1 && size != 2 && size != 4 && size != 8 {
			return nil
		}
		// Every sized element has at least one component, so this loop is
		// bounded by the global component budget as well as by count.
		var encoded [8]byte
		for i := 0; i < ty.count && i < len(values); i++ {
			binary.LittleEndian.PutUint64(encoded[:], values[i])
			copy(g.data[i*size:(i+1)*size], encoded[:size])
		}
	}
	return nil
}
