// Package bytecode implements a bounded ClamBC interpreter. Scanner hook and
// signature activation support must be completed before it can activate feeds.
package bytecode

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

var (
	ErrMalformed   = errors.New("malformed signature bytecode program")
	ErrUnsupported = errors.New("unsupported signature bytecode semantics")
	// ErrExtraction stops a scanning run that writes or finalizes extracted
	// output without a recursive child-scanner callback. Callers that cannot
	// analyze the child must never turn this into a clean verdict.
	ErrExtraction = errors.New("bytecode run needs extracted-output scanning")
	// ErrPDFObjects stops a run that asks for PDF parser state this engine
	// cannot establish for the current object and phase.
	ErrPDFObjects = errors.New("bytecode run needs the PDF parser's object model")
	ErrLimit      = errors.New("signature bytecode resource limit exceeded")
	ErrRuntime    = errors.New("signature bytecode runtime fault")
)

const maxProgramBytes = 4 << 20
const maxInstructions = 100000
const maxValues = 8192

type operand struct {
	constant bool
	global   bool
	address  bool
	value    uint64
	width    int
}
type instruction struct {
	op, width, dest, function int
	typeID, gepOffset         int
	args                      []operand
	branches                  []int
}
type function struct {
	args, result int
	widths       []int
	blocks       [][]instruction
	offsets      []int
	frameBytes   int
}
type Program struct {
	Format, Kind                       uint64
	LogicalSignature                   string
	MinFunctionality, MaxFunctionality uint64
	functions                          []function
	types                              *typeTable
	apis                               map[int]apiDefinition
	globals                            []global
	storageBytes                       uint64
}

// Cursor decoding is independent of host endianness. Numeric payloads store
// least-significant hexadecimal nibbles first in bytes 0x60..0x6f.
type cursor struct {
	b   []byte
	pos int
	err error
}

func (c *cursor) byte() byte {
	if c.err != nil {
		return 0
	}
	if c.pos >= len(c.b) {
		c.err = ErrMalformed
		return 0
	}
	v := c.b[c.pos]
	c.pos++
	return v
}
func (c *cursor) peek() byte {
	if c.pos >= len(c.b) {
		return 0
	}
	return c.b[c.pos]
}
func (c *cursor) expect(b byte) {
	if c.byte() != b {
		c.err = ErrMalformed
	}
}
func (c *cursor) fixed(n int) uint64 {
	var v uint64
	for i := 0; i < n; i++ {
		b := c.byte()
		if b&0xf0 != 0x60 {
			c.err = ErrMalformed
			return 0
		}
		v |= uint64(b&15) << uint(4*i)
	}
	return v
}
func (c *cursor) number() uint64 {
	n := int(c.byte()) - 0x60
	if n < 0 || n > 16 {
		c.err = ErrMalformed
		return 0
	}
	return c.fixed(n)
}
func (c *cursor) count(max int) int {
	n := c.number()
	if n > uint64(max) {
		c.err = ErrLimit
		return 0
	}
	return int(n)
}
func (c *cursor) text() string {
	c.expect('|')
	n := c.count(maxProgramBytes)
	if c.err != nil || n > (len(c.b)-c.pos)/2 {
		c.err = ErrMalformed
		return ""
	}
	b := make([]byte, n)
	for i := range b {
		b[i] = byte(c.fixed(2))
	}
	if n == 0 {
		return ""
	}
	if b[n-1] != 0 || bytes.IndexByte(b[:n-1], 0) >= 0 {
		c.err = ErrMalformed
		return ""
	}
	return string(b[:n-1])
}
func (c *cursor) end() error {
	if c.err != nil {
		return c.err
	}
	if c.pos != len(c.b) {
		return ErrMalformed
	}
	return nil
}
func (c *cursor) operand(widths []int) operand {
	b := c.peek()
	if b&0xf0 == 0x40 || b == 0x50 {
		c.byte()
		n := int(b|0x20) - 0x60
		v := c.fixed(n)
		width := int(c.fixed(1)) * 8
		if width == 0 {
			return operand{global: true, value: v}
		}
		if width != 8 && width != 16 && width != 32 && width != 64 {
			c.err = ErrUnsupported
		}
		return operand{constant: true, value: v, width: width}
	}
	id := c.count(maxValues)
	if id >= len(widths) {
		c.err = ErrMalformed
		return operand{}
	}
	return operand{value: uint64(id), width: widths[id] & 0x7fff}
}
func integerWidth(n int) bool { return n == 1 || n == 8 || n == 16 || n == 32 || n == 64 }

// Parse validates a whole program before execution. Unsupported declarations
// and instructions fail even when unreachable or unused.
func Parse(data []byte) (*Program, error) {
	if len(data) > maxProgramBytes {
		return nil, ErrLimit
	}
	scanner := bufio.NewScanner(bytes.NewReader(data))
	scanner.Buffer(make([]byte, 4096), 1<<20)
	var lines [][]byte
	for scanner.Scan() {
		lines = append(lines, bytes.Clone(bytes.TrimSuffix(scanner.Bytes(), []byte{'\r'})))
		if len(lines) > maxInstructions {
			return nil, ErrLimit
		}
	}
	if scanner.Err() != nil {
		return nil, ErrLimit
	}
	if len(lines) < 7 || !bytes.HasPrefix(lines[0], []byte("ClamBC")) {
		return nil, ErrMalformed
	}
	c := &cursor{b: lines[0], pos: 6}
	p := &Program{Format: c.number()}
	if p.Format != 6 && p.Format != 7 {
		return nil, ErrUnsupported
	}
	c.number()
	c.text()
	exclude := c.number()
	p.Kind = c.number()
	p.MinFunctionality = c.number()
	p.MaxFunctionality = c.number()
	if p.Kind > 65535 || p.MaxFunctionality != 0 && p.MinFunctionality > p.MaxFunctionality {
		return nil, ErrMalformed
	}
	c.number()
	c.text()
	types := c.count(4096)
	functions := c.count(1024)
	if c.number() != 0x53e5493e9f3d1c30 || c.fixed(2) != 42 {
		return nil, ErrMalformed
	}
	c.expect(':')
	if c.err != nil {
		return nil, c.err
	}
	lineMax, err := strconv.Atoi(string(c.b[c.pos:]))
	if err != nil || lineMax < 1 || lineMax > 1<<20 {
		return nil, ErrLimit
	}
	for _, line := range lines[1:] {
		if len(line) >= lineMax {
			return nil, ErrLimit
		}
	}
	if exclude != 0 {
		return nil, ErrUnsupported
	}
	if functions < 1 || types < 5 {
		return nil, ErrMalformed
	}
	p.LogicalSignature = string(lines[1])
	p.types, err = parseTypeTable(lines[2], types)
	if err != nil {
		return nil, err
	}
	p.apis, err = parseAPIs(lines[3], p.types)
	if err != nil {
		return nil, err
	}
	p.globals, err = parseGlobals(lines[4], p.types)
	if err != nil {
		return nil, err
	}
	pos, total := 5, 0
	for fn := 0; fn < functions; fn++ {
		if pos >= len(lines) {
			return nil, ErrMalformed
		}
		c = &cursor{b: lines[pos]}
		pos++
		c.expect('A')
		f := function{args: int(c.fixed(1)), result: c.count(65535)}
		if f.result != 0 && !valueType(f.result) {
			return nil, ErrUnsupported
		}
		c.expect('L')
		locals := c.count(maxValues)
		if locals+f.args > maxValues {
			return nil, ErrLimit
		}
		for i := 0; i < locals+f.args; i++ {
			width := c.count(65535)
			addressed := c.fixed(1)
			if addressed > 1 || width&0x8000 != 0 || width == 0 {
				return nil, ErrMalformed
			}
			if _, err := p.types.normalize(width); err != nil {
				return nil, err
			}
			if !p.scalarType(width) && (addressed == 0 || width < 65 || p.types.types[width-65].kind == functionType) {
				return nil, ErrUnsupported
			}
			if i < f.args && addressed != 0 {
				return nil, ErrUnsupported
			}
			size, _ := p.types.size(width)
			align, _ := p.types.align(width)
			f.frameBytes = (f.frameBytes + align - 1) & ^(align - 1)
			if size > maxTypeBytes-f.frameBytes {
				return nil, ErrLimit
			}
			f.offsets = append(f.offsets, f.frameBytes)
			f.frameBytes += size
			if addressed != 0 {
				width |= 0x8000
			}
			f.widths = append(f.widths, width)
		}
		c.expect('F')
		declared := c.count(maxInstructions)
		blocks := c.count(4096)
		if err := c.end(); err != nil {
			return nil, err
		}
		if declared < 1 || blocks < 1 {
			return nil, ErrMalformed
		}
		if total+declared > maxInstructions {
			return nil, ErrLimit
		}
		total += declared
		actual := 0
		for bb := 0; bb < blocks; bb++ {
			if pos >= len(lines) {
				return nil, ErrMalformed
			}
			body, err := p.parseBlock(lines[pos], f.widths, blocks, functions, bb == blocks-1)
			pos++
			if err != nil {
				return nil, fmt.Errorf("function %d block %d: %w", fn, bb, err)
			}
			actual += len(body)
			if actual > declared {
				return nil, ErrMalformed
			}
			f.blocks = append(f.blocks, body)
		}
		if actual != declared {
			return nil, ErrMalformed
		}
		p.functions = append(p.functions, f)
	}
	if pos < len(lines) && len(lines[pos]) > 0 && lines[pos][0] == 'S' {
		pos = len(lines)
	} // encoded source is inert, already size bounded
	if pos != len(lines) {
		return nil, ErrMalformed
	}
	for fi := range p.functions {
		f := &p.functions[fi]
		for bi := range f.blocks {
			for ii := range f.blocks[bi] {
				inst := &f.blocks[bi][ii]
				if inst.op == 32 {
					target := p.functions[inst.function]
					if len(inst.args) != target.args || inst.width != target.result {
						return nil, ErrMalformed
					}
					for i := range inst.args {
						// The interpreter copies direct-call arguments out of the
						// caller's value slots (READOLD). Its mapped global and
						// addressed-local operands always fail that copy.
						a := &inst.args[i]
						if a.global || !a.constant && f.widths[a.value]&0x8000 != 0 {
							return nil, ErrUnsupported
						}
						if !p.argument(a, target.widths[i], f.widths) {
							return nil, ErrMalformed
						}
					}
				}
				if inst.op == 19 && inst.width != f.result {
					return nil, ErrMalformed
				}
				if inst.op == 20 && f.result != 0 {
					return nil, ErrMalformed
				}
			}
		}
	}
	// Conservative resident-size estimate for a publisher/database-wide budget.
	// Serialized source is not retained, but its multiple covers metadata and
	// allocation overhead in addition to the explicitly materialized tables.
	p.storageBytes = uint64(len(data)) * 4
	for _, g := range p.globals {
		p.storageBytes += uint64(len(g.data)) + uint64(len(g.pointers))*48 + 64
	}
	for _, ty := range p.types.types {
		p.storageBytes += 128 + uint64(len(ty.elements)+len(ty.offsets))*16
	}
	for _, f := range p.functions {
		p.storageBytes += uint64(len(f.widths))*32 + 128
		for _, b := range f.blocks {
			p.storageBytes += 32
			for _, in := range b {
				p.storageBytes += 128 + uint64(len(in.args))*64 + uint64(len(in.branches))*16
			}
		}
	}
	return p, nil
}

// StorageBytes conservatively estimates retained compiled tables and globals;
// callers loading many programs must impose an aggregate compilation budget.
func (p *Program) StorageBytes() uint64 { return p.storageBytes }

// String describes input requirements without implying scanner hook support.
func (p *Program) String() string {
	return fmt.Sprintf("signature bytecode format %d, %d functions, kind %d, trigger %q", p.Format, len(p.functions), p.Kind, strings.TrimSpace(p.LogicalSignature))
}
