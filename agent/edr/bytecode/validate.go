package bytecode

import "fmt"

func (p *Program) scalarType(id int) bool {
	return integerWidth(id) || p.types.pointer(id) && id&0x8000 == 0
}

// valueType reports whether the pinned interpreter moves a value of this type
// at its full width. parseBB derives an operation's size class from its
// instruction type only for integers and the predefined i8* (type 65); every
// other pointer type falls back to the one-bit handlers. COPY, LOAD, STORE,
// SELECT, RET, comparisons and GEP1 dispatch on that class, so other pointer
// widths are unsupported rather than run as 64-bit values.
func valueType(id int) bool { return integerWidth(id) || id == 65 }

func parseAPIs(line []byte, types *typeTable) (map[int]apiDefinition, error) {
	c := &cursor{b: line}
	c.expect('E')
	max, calls := c.count(4096), c.count(4096)
	if c.err != nil {
		return nil, c.err
	}
	if calls > max {
		return nil, ErrMalformed
	}
	apis := make(map[int]apiDefinition, calls)
	memo := map[int]int{}
	for i := 0; i < calls; i++ {
		id, ty, name := c.count(max), c.count(65535), c.text()
		if c.err != nil {
			return nil, c.err
		}
		def, ok := builtinAPIs[id]
		if !ok {
			return nil, ErrUnsupported
		}
		if _, duplicate := apis[id]; duplicate || id < 1 || name != def.name {
			return nil, ErrMalformed
		}
		if _, err := types.normalize(ty); err != nil || ty < 65 || ty&0x8000 != 0 {
			return nil, ErrMalformed
		}
		proto := types.types[ty-65]
		if proto.kind != functionType || len(proto.elements) != len(def.args)+1 || !types.referenceEqual(proto.elements[0], def.result, memo) {
			return nil, ErrMalformed
		}
		for j, a := range def.args {
			if !types.referenceEqual(proto.elements[j+1], a, memo) {
				return nil, ErrMalformed
			}
		}
		// Instructions are checked against the program's own prototype,
		// whose struct pointers have the program's type IDs.
		apis[id] = apiDefinition{name: def.name, result: proto.elements[0], args: proto.elements[1:]}
	}
	return apis, c.end()
}

// Literal encoding describes storage bytes in a zero-initialized 64-bit slot.
// The operation/prototype determines how that slot is read. Preserve truncation
// to serialized storage before widening/narrowing the read (including i1).
func coerce(a *operand, want int) bool {
	if want >= firstCompoundType && a.global && a.value == 0 {
		a.width = want
		return true
	}
	if a.width == want {
		return true
	}
	if a.constant && integerWidth(want) {
		a.value &= bitMask(a.width)
		a.width = want
		return true
	}
	return false
}
func (p *Program) pointerArgument(a *operand, widths []int) bool {
	if !a.constant && !a.global && widths[a.value]&0x8000 != 0 {
		a.address = true
		return true
	}
	return p.types.pointer(a.width) || a.global && a.value == 0
}
func (p *Program) argument(a *operand, want int, widths []int) bool {
	if p.types.pointer(want) {
		return p.pointerArgument(a, widths)
	}
	return coerce(a, want)
}

func (p *Program) parseBlock(line []byte, widths []int, blocks, functions int, last bool) ([]instruction, error) {
	c := &cursor{b: line}
	c.expect('B')
	var out []instruction
	for c.err == nil {
		terminal := c.peek() == 'T'
		inst := instruction{gepOffset: -1}
		if terminal {
			c.byte()
		} else {
			inst.width = c.count(65535)
			inst.dest = c.count(maxValues)
		}
		inst.op = int(c.fixed(2))
		count := 0
		switch {
		case inst.op >= 1 && inst.op <= 13:
			count = 2
		case inst.op >= 14 && inst.op <= 16:
			count = 1
		case inst.op == 17:
			inst.args = append(inst.args, c.operand(widths))
			inst.branches = []int{c.count(blocks), c.count(blocks)}
		case inst.op == 18:
			inst.branches = []int{c.count(blocks)}
		case inst.op == 19:
			inst.width = c.count(65535)
			count = 1
		case inst.op == 20:
		case inst.op >= 21 && inst.op <= 30:
			inst.width = c.count(65535)
			count = 2
		case inst.op == 31:
			count = 3
		case inst.op == 32 || inst.op == 33:
			count = int(c.fixed(1))
			inst.function = c.count(4096)
			if inst.op == 32 {
				inst.function--
				if inst.function < 0 || inst.function >= functions {
					return nil, ErrMalformed
				}
			} else if _, ok := p.apis[inst.function]; !ok {
				return nil, ErrUnsupported
			}
		case inst.op == 34 || inst.op == 38 || inst.op == 49:
			count = 2
		case inst.op == 35 || inst.op == 36:
			inst.typeID = c.count(65535)
			count = 2
		case inst.op == 39 || inst.op == 50 || inst.op >= 46 && inst.op <= 48:
			count = 1
		case inst.op >= 40 && inst.op <= 43:
			count = 3
		case inst.op == 44 || inst.op == 45:
		default:
			return nil, ErrUnsupported
		}
		if inst.op != 45 && terminal != (inst.op >= 17 && inst.op <= 20) {
			return nil, fmt.Errorf("instruction %d opcode %d terminator: %w", len(out), inst.op, ErrMalformed)
		}
		for _, b := range inst.branches {
			if b <= 0 || b >= blocks {
				return nil, ErrMalformed
			}
		}
		for j := 0; j < count; j++ {
			inst.args = append(inst.args, c.operand(widths))
		}
		if c.err != nil {
			return nil, fmt.Errorf("instruction %d opcode %d operands: %w", len(out), inst.op, c.err)
		}
		for i := range inst.args {
			a := &inst.args[i]
			if a.global {
				if a.value >= uint64(len(p.globals)) {
					return nil, ErrMalformed
				}
				a.width = p.globals[a.value].typeID
			}
		}
		if err := p.validateInstruction(&inst, widths, terminal); err != nil {
			return nil, fmt.Errorf("instruction %d opcode %d: %w", len(out), inst.op, err)
		}
		out = append(out, inst)
		if len(out) > maxInstructions {
			return nil, ErrLimit
		}
		if terminal {
			break
		}
	}
	if last {
		c.expect('E')
	}
	if err := c.end(); err != nil {
		return nil, err
	}
	return out, nil
}

func (p *Program) validateInstruction(in *instruction, widths []int, terminal bool) error {
	a := in.args
	op := in.op
	// COPY and STORE encode dummy destinations; COPY's actual type is the
	// second operand, STORE's is its typed source when present.
	if op == 34 {
		if a[1].constant || a[1].global || !p.scalarType(a[1].width) {
			return ErrMalformed
		}
		in.width = a[1].width
		if !coerce(&a[0], in.width) {
			return ErrMalformed
		}
	} else if op == 38 && !a[0].constant && !a[0].global {
		in.width = a[0].width
	}
	if (op == 31 || op == 34 || op == 35 || op == 38 || op == 39) && p.scalarType(in.width) && !valueType(in.width) {
		return ErrUnsupported
	}
	noResult := terminal || op == 34 || op == 38 || op >= 40 && op <= 42 || op == 45 || (op == 32 || op == 33) && in.width == 0
	if !noResult {
		if in.dest >= len(widths) || !p.scalarType(in.width) {
			return ErrMalformed
		}
		want := in.width
		if op >= 21 && op <= 30 {
			want = 1
		}
		if widths[in.dest]&0x7fff != want {
			return ErrMalformed
		}
	}
	if op >= 1 && op <= 13 || op >= 21 && op <= 30 || op == 19 {
		if !integerWidth(in.width) && !(op == 19 || op == 21 || op == 22) {
			return ErrUnsupported
		}
		if !valueType(in.width) {
			return ErrUnsupported
		}
		for i := range a {
			if !coerce(&a[i], in.width) {
				return ErrMalformed
			}
		}
	}
	if op >= 14 && op <= 16 {
		if a[0].constant || a[0].global {
			return ErrUnsupported
		}
		if !integerWidth(a[0].width) || !integerWidth(in.width) {
			return ErrMalformed
		}
		if op == 14 && a[0].width <= in.width || op != 14 && a[0].width >= in.width {
			return ErrMalformed
		}
	}
	if op == 17 && !coerce(&a[0], 1) {
		return ErrMalformed
	}
	if op == 31 && (!coerce(&a[0], 1) || !coerce(&a[1], in.width) || !coerce(&a[2], in.width)) {
		return ErrMalformed
	}
	if op == 33 {
		def := p.apis[in.function]
		if in.width != def.result || len(a) != len(def.args) {
			return ErrMalformed
		}
		for i, want := range def.args {
			if !p.argument(&a[i], want, widths) {
				return ErrMalformed
			}
		}
	}
	if op == 35 || op == 36 {
		if in.typeID&0x8000 != 0 || !p.types.pointer(in.typeID) || !p.types.pointer(in.width) || !coerce(&a[1], 32) {
			return ErrMalformed
		}
		if op == 35 {
			if a[0].constant || a[0].global || widths[a[0].value]&0x8000 != 0 || !p.types.pointer(a[0].width) {
				return ErrUnsupported
			}
			if _, _, err := p.types.element(in.typeID, 0); err != nil {
				return err
			}
		} else {
			if a[0].global {
				return ErrUnsupported
			} // upstream's global GEPZ addressing is not a stable contract
			if !p.pointerArgument(&a[0], widths) {
				return ErrMalformed
			}
			child := p.types.types[in.typeID-65].elements[0]
			if child < 65 {
				return ErrMalformed
			}
			kind := p.types.types[child-65].kind
			if kind == structType || kind == packedStructType {
				if !a[1].constant {
					return ErrMalformed
				}
				_, offset, err := p.types.element(child, int(int32(a[1].value)))
				if err != nil {
					return err
				}
				in.gepOffset = offset
			}
		}
	}
	if op == 38 {
		if !p.scalarType(in.width) || !coerce(&a[0], in.width) || !p.types.pointer(a[1].width) {
			return ErrMalformed
		}
		// STORE reads the pointer value, not the address of an alloca slot.
	}
	if op == 39 && (!p.scalarType(in.width) || !p.pointerArgument(&a[0], widths)) {
		return ErrMalformed
	}
	if op >= 40 && op <= 43 {
		if !p.pointerArgument(&a[0], widths) {
			return ErrMalformed
		}
		if op == 40 {
			if !coerce(&a[1], 32) {
				return ErrMalformed
			}
		} else if !p.pointerArgument(&a[1], widths) {
			return ErrMalformed
		}
		// The wire interpreter uses i32 lengths for memcpy/memcmp and i64
		// for memmove/memset. Never truncate a negative/oversized length.
		// memmove/memset use only the low 32 bits of their 64-bit read, so an
		// i32 length variable is its own value there.
		lengthWidth := 32
		if (op == 40 || op == 42) && (a[2].constant || a[2].width != 32) {
			lengthWidth = 64
		}
		if !coerce(&a[2], lengthWidth) {
			return ErrMalformed
		}
		if op == 43 && in.width != 32 {
			return ErrMalformed
		}
	}
	if op == 44 && in.width != 1 {
		return ErrMalformed
	}
	if op >= 46 && op <= 48 && (in.width != 16<<(op-46) || !coerce(&a[0], in.width)) {
		return ErrMalformed
	}
	if op == 49 {
		if in.width != 32 || !p.pointerArgument(&a[0], widths) || !p.pointerArgument(&a[1], widths) {
			return ErrMalformed
		}
	}
	if op == 50 && (in.width != 64 || !p.pointerArgument(&a[0], widths)) {
		return ErrMalformed
	}
	return nil
}
