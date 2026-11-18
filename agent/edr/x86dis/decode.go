// Package x86dis decodes one 32-bit x86 instruction into the 64-byte record
// that the reference's bytecode call disasm_x86 fills in, so that bytecode
// programs inspecting code see the same operation numbers and operands.
package x86dis

import "encoding/binary"

// Record layout: operation number (u16 LE) @0, operand-size prefix flag @2,
// address-size prefix flag @3, segment-override register @4 (0 when none),
// three 10-byte operands @5, @15, @25, zeros @35-63.
const argBase, argSize = 5, 10

// Operand access kinds (operand byte 0).
const (
	accessImm = 1
	accessRel = 2
	accessReg = 3
	accessMem = 4
)

type decoder struct {
	buf    []byte
	pos    int
	opsize bool // operand-size prefix seen
	adsize bool // address-size prefix seen
	seg    byte
	modrm  int // -1 until read
	// zeroExtend makes 8- and 16-bit displacements unsigned, as the
	// reference reports them for x87 memory operands.
	zeroExtend bool
	rec        [64]byte
}

func (d *decoder) byte() (byte, bool) {
	if d.pos >= len(d.buf) {
		return 0, false
	}
	b := d.buf[d.pos]
	d.pos++
	return b, true
}

// signed reads an n-byte little-endian value and sign-extends it.
func (d *decoder) signed(n int) (int64, bool) {
	v, ok := d.unsigned(n)
	shift := 64 - 8*uint(n)
	return int64(v<<shift) >> shift, ok
}

func (d *decoder) unsigned(n int) (uint64, bool) {
	if len(d.buf)-d.pos < n {
		return 0, false
	}
	var v uint64
	for i := n - 1; i >= 0; i-- {
		v = v<<8 | uint64(d.buf[d.pos+i])
	}
	d.pos += n
	return v, true
}

// displacement reads an n-byte memory displacement (n may be 0).
func (d *decoder) displacement(n int) (int64, bool) {
	if d.zeroExtend {
		v, ok := d.unsigned(n)
		return int64(v), ok
	}
	return d.signed(n)
}

func (d *decoder) readModRM() bool {
	if d.modrm >= 0 {
		return true
	}
	b, ok := d.byte()
	if !ok {
		return false
	}
	d.modrm = int(b)
	return true
}

func (d *decoder) mod() int { return d.modrm >> 6 }
func (d *decoder) reg() int { return d.modrm >> 3 & 7 }
func (d *decoder) rm() int  { return d.modrm & 7 }

// Decode decodes one instruction from at most len(buf) bytes. ok is false
// where the reference refuses: invalid encodings, prefixes alone, and input
// that ends inside an instruction.
func Decode(buf []byte) (record [64]byte, n int, ok bool) {
	d := decoder{buf: buf, modrm: -1}
	var b byte
	for {
		if b, ok = d.byte(); !ok {
			return record, 0, false
		}
		switch b {
		case 0x66:
			d.opsize = true
		case 0x67:
			d.adsize = true
		case 0x26, 0x2e, 0x36, 0x3e:
			d.seg = regES + b>>3&3
		case 0x64, 0x65:
			d.seg = regES + 4 + b&1
		default:
			goto opcode
		}
	}
opcode:
	f := oneByte[b]
	switch f.kind {
	case escape:
		if b, ok = d.byte(); !ok {
			return record, 0, false
		}
		f = twoByte[b]
	case x87:
		if !d.fpu(b) {
			return record, 0, false
		}
		return d.finish()
	}
	if f.kind == group {
		if !d.readModRM() {
			return record, 0, false
		}
		f = groups[f.g][d.reg()]
	}
	if f.kind != plain {
		return record, 0, false
	}
	operation := f.op
	if d.opsize && f.op16 != 0 {
		operation = f.op16
	}
	binary.LittleEndian.PutUint16(d.rec[0:], operation)
	for i, a := range f.args {
		if a == __ {
			break
		}
		if !d.operand(i, a, b) {
			return record, 0, false
		}
	}
	return d.finish()
}

func (d *decoder) finish() ([64]byte, int, bool) {
	if d.opsize {
		d.rec[2] = 1
	}
	if d.adsize {
		d.rec[3] = 1
	}
	d.rec[4] = d.seg
	return d.rec, d.pos, true
}

func (d *decoder) slot(i int) []byte { return d.rec[argBase+argSize*i : argBase+argSize*(i+1)] }

func (d *decoder) setReg(i int, r int) {
	s := d.slot(i)
	s[0], s[1] = accessReg, byte(r)
}

func (d *decoder) setValue(i int, access, size byte, v int64) {
	s := d.slot(i)
	s[0], s[1] = access, size
	binary.LittleEndian.PutUint64(s[2:], uint64(v))
}

// vsize is the size code of a word-or-dword operand.
func (d *decoder) vsize() byte {
	if d.opsize {
		return sizeWord
	}
	return sizeDword
}

func (d *decoder) vreg(r int) int {
	if d.opsize {
		return regAX + r
	}
	return regEAX + r
}

func byteReg(r int) int {
	if r < 4 {
		return regAL + r
	}
	return regAH + r - 4
}

// operand decodes argument i of kind a; opcode is the last opcode byte.
func (d *decoder) operand(i int, a arg, opcode byte) bool {
	switch a {
	case Eb, Ew, Ev, Ep, M, Mb, Ms, Mq, Rd, Gb, Gw, Gv, Cd, Dd, Sw:
		if !d.readModRM() {
			return false
		}
	}
	switch a {
	case Eb:
		return d.rmOperand(i, sizeByte, byteReg(d.rm()))
	case Ew:
		return d.rmOperand(i, sizeWord, regAX+d.rm())
	case Ev:
		return d.rmOperand(i, d.vsize(), d.vreg(d.rm()))
	case Ep:
		size := byte(sizeSix)
		if d.opsize {
			size = sizeDword
		}
		return d.rmOperand(i, size, d.vreg(d.rm()))
	case M, Mb, Ms, Mq:
		if d.mod() == 3 {
			return false
		}
		size := map[arg]byte{M: d.vsize(), Mb: sizeByte, Ms: sizeSix, Mq: sizeQword}[a]
		return d.memory(i, size)
	case Rd:
		d.setReg(i, regEAX+d.rm())
	case Gb:
		d.setReg(i, byteReg(d.reg()))
	case Gw:
		d.setReg(i, regAX+d.reg())
	case Gv:
		d.setReg(i, d.vreg(d.reg()))
	case Cd:
		switch d.reg() {
		case 0, 2, 3, 4:
			d.setReg(i, regCR0+d.reg())
		default:
			return false
		}
	case Dd:
		switch d.reg() {
		case 0, 1, 2, 3, 6, 7:
			d.setReg(i, regDR0+d.reg())
		default:
			return false
		}
	case Sw:
		if d.reg() > 5 {
			return false
		}
		d.setReg(i, regES+d.reg())
	case Ib:
		return d.immediate(i, accessImm, sizeByte, 1)
	case Iw:
		return d.immediate(i, accessImm, sizeWord, 2)
	case Iz:
		if d.opsize {
			return d.immediate(i, accessImm, sizeWord, 2)
		}
		return d.immediate(i, accessImm, sizeDword, 4)
	case Ap:
		if d.opsize {
			return d.immediate(i, accessImm, sizeDword, 4)
		}
		return d.immediate(i, accessImm, sizeSix, 6)
	case Jb:
		return d.immediate(i, accessRel, sizeByte, 1)
	case Jz:
		if d.opsize {
			return d.immediate(i, accessRel, sizeWord, 2)
		}
		return d.immediate(i, accessRel, sizeDword, 4)
	case Ob, Ov:
		size := byte(sizeByte)
		if a == Ov {
			size = d.vsize()
		}
		width := 4
		if d.adsize {
			width = 2 // zero-extended
		}
		disp, ok := d.unsigned(width)
		if !ok {
			return false
		}
		d.setMemory(i, size, regNone, regNone, 0, int64(disp))
	case One:
		d.setValue(i, accessImm, sizeWord, 1)
	case AL:
		d.setReg(i, regAL)
	case CL:
		d.setReg(i, regAL+1)
	case DX:
		d.setReg(i, regAX+2)
	case EAX:
		d.setReg(i, d.vreg(0))
	case Zb:
		d.setReg(i, byteReg(int(opcode&7)))
	case Zv:
		d.setReg(i, d.vreg(int(opcode&7)))
	case Zd:
		d.setReg(i, regEAX+int(opcode&7))
	case ES, CS, SS, DS, FS, GS:
		d.setReg(i, regES+int(a-ES))
	default:
		return false
	}
	return true
}

func (d *decoder) immediate(i int, access, size byte, width int) bool {
	v, ok := d.signed(width)
	if !ok {
		return false
	}
	d.setValue(i, access, size, v)
	return true
}

// rmOperand decodes the ModRM r/m operand: the register reg when mod is 3,
// otherwise memory of the given size.
func (d *decoder) rmOperand(i int, size byte, reg int) bool {
	if d.mod() == 3 {
		d.setReg(i, reg)
		return true
	}
	return d.memory(i, size)
}

func (d *decoder) setMemory(i int, size byte, scaled, added int, scale byte, disp int64) {
	s := d.slot(i)
	s[0], s[1], s[2], s[3], s[4], s[5] = accessMem, size, byte(scaled), byte(added), scale, 0
	binary.LittleEndian.PutUint32(s[6:], uint32(disp))
}

// memory decodes a ModRM memory operand (mod 0-2) with its SIB byte and
// displacement.
func (d *decoder) memory(i int, size byte) bool {
	if d.adsize {
		return d.memory16(i, size)
	}
	mod, rm := d.mod(), d.rm()
	scaled, added, scale := regEAX+rm, regNone, byte(1)
	if rm == 4 {
		sib, ok := d.byte()
		if !ok {
			return false
		}
		index, base := int(sib>>3&7), int(sib&7)
		scale = 1 << (sib >> 6)
		scaled, added = regEAX+index, regEAX+base
		if index == 4 {
			scaled, added = regEAX+base, regNone
			scale = 1
		}
		if base == 5 && mod == 0 {
			if index == 4 {
				scaled, scale = regNone, 0
			}
			added = regNone
			mod = 2
		}
	} else if rm == 5 && mod == 0 {
		scaled, scale = regNone, 0
		mod = 2
	}
	disp, ok := d.displacement([3]int{0, 1, 4}[mod])
	if !ok {
		return false
	}
	d.setMemory(i, size, scaled, added, scale, disp)
	return true
}

var pairs16 = [8][2]int{
	{regAX + 3, regAX + 6}, // bx+si
	{regAX + 3, regAX + 7}, // bx+di
	{regAX + 5, regAX + 6}, // bp+si
	{regAX + 5, regAX + 7}, // bp+di
	{regAX + 6, regNone},   // si
	{regAX + 7, regNone},   // di
	{regAX + 5, regNone},   // bp
	{regAX + 3, regNone},   // bx
}

// memory16 decodes a ModRM memory operand under the address-size prefix.
func (d *decoder) memory16(i int, size byte) bool {
	mod, rm := d.mod(), d.rm()
	scaled, added, scale := pairs16[rm][0], pairs16[rm][1], byte(1)
	width := mod
	if mod == 0 && rm == 6 {
		// The direct address reports register 0 as its added register.
		scaled, added, scale, width = regNone, regEAX, 0, 2
	}
	disp, ok := d.displacement(width)
	if !ok {
		return false
	}
	d.setMemory(i, size, scaled, added, scale, disp)
	return true
}

// fpu decodes an x87 instruction whose first opcode byte is b (D8-DF).
func (d *decoder) fpu(b byte) bool {
	if !d.readModRM() {
		return false
	}
	row := b - 0xd8
	if d.mod() != 3 {
		d.zeroExtend = true
		m := fpuMemory[row][d.reg()]
		if m.op == 0 {
			return false
		}
		binary.LittleEndian.PutUint16(d.rec[0:], m.op)
		return d.memory(0, m.size)
	}
	r := fpuRegister[row][d.reg()][d.rm()]
	if r.op == 0 {
		return false
	}
	binary.LittleEndian.PutUint16(d.rec[0:], r.op)
	for i, a := range r.args {
		switch a {
		case ST0:
			d.setReg(i, regST0)
		case STi:
			d.setReg(i, regST0+d.rm())
		}
	}
	return true
}
