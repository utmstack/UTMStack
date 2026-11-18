package bytecode

import (
	"context"
	"errors"
	"strings"
	"testing"
)

// Synthetic fixtures use the public ClamBC format; no upstream code or malware
// is embedded. Compatibility is also checked against upstream retmagic/inf files
// during development (provenance recorded in README.md).
func num(v uint64) string {
	var digits []byte
	for v != 0 {
		digits = append(digits, 0x60|byte(v&15))
		v >>= 4
	}
	return string(append([]byte{0x60 + byte(len(digits))}, digits...))
}
func fixed(v uint64, n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = 0x60 | byte(v&15)
		v >>= 4
	}
	return string(b)
}
func constant(v uint64, width int) string {
	b := []byte(num(v))
	b[0] &^= 0x20
	return string(b) + fixed(uint64(width/8), 1)
}
func fixture(widths []int, blocks ...string) []byte {
	header := "ClamBC" + num(7) + num(0) + "|" + num(0) + strings.Repeat(num(0), 5) + "|" + num(0) + num(5) + num(1) + num(0x53e5493e9f3d1c30) + fixed(42, 2) + ":4096"
	locals := "A" + fixed(0, 1) + num(32) + "L" + num(uint64(len(widths)))
	for _, width := range widths {
		locals += num(uint64(width)) + fixed(0, 1)
	}
	n := 0
	for _, block := range blocks {
		n += strings.Count(block, "T") // all additional test instructions start with an integer type.
		c := &cursor{b: []byte(block), pos: 1}
		for c.err == nil && c.peek() != 'T' {
			c.number()
			c.number()
			op := c.fixed(2)
			count := 2
			if op >= 14 && op <= 16 || op >= 46 && op <= 48 {
				count = 1
			}
			if op >= 21 && op <= 30 {
				c.number()
			}
			for i := 0; i < count; i++ {
				c.operand(widths)
			}
			n++
		}
	}
	locals += "F" + num(uint64(n)) + num(uint64(len(blocks)))
	lines := []string{header, "", "T" + fixed(69, 2), "E" + num(0) + num(0), "G" + num(0) + num(1) + num(0) + "@`", locals}
	for i, block := range blocks {
		if i == len(blocks)-1 {
			block += "E"
		}
		lines = append(lines, block)
	}
	return []byte(strings.Join(lines, "\n") + "\n")
}
func ret(op string) string { return "T" + fixed(19, 2) + num(32) + op }
func TestReturnConstantAndArithmetic(t *testing.T) {
	for _, tt := range []struct {
		op         int
		a, b, want uint64
	}{
		{1, 0xffffffff, 2, 1}, {2, 0, 1, 0xffffffff}, {3, 21, 2, 42}, {4, 43, 2, 21}, {5, 0xfffffff6, 2, 0xfffffffb},
		{6, 43, 2, 1}, {7, 0xfffffff6, 3, 0xffffffff}, {8, 1, 31, 0x80000000}, {9, 0x80000000, 31, 1}, {10, 0x80000000, 31, 0xffffffff},
		{11, 15, 10, 10}, {12, 5, 10, 15}, {13, 15, 10, 5},
	} {
		body := "B" + num(32) + num(0) + fixed(uint64(tt.op), 2) + constant(tt.a, 32) + constant(tt.b, 32) + ret(num(0))
		p, err := Parse(fixture([]int{32}, body))
		if err != nil {
			t.Fatal(tt.op, err)
		}
		value, err := p.Run(context.Background(), 0, nil, 100)
		if err != nil || value != tt.want {
			t.Fatal(tt.op, value, err)
		}
	}
	p, err := Parse(fixture(nil, "B"+ret(constant(0x1234f00d, 32))))
	if err != nil {
		t.Fatal(err)
	}
	if value, err := p.Run(context.Background(), 0, nil, 10); value != 0x1234f00d || err != nil {
		t.Fatal(value, err)
	}
}
func TestBytecodeBoundsAndRuntimeFaults(t *testing.T) {
	loop := "BT" + fixed(18, 2) + num(1)
	p, err := Parse(fixture(nil, loop, loop))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := p.Run(context.Background(), 0, nil, 100); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := p.Run(ctx, 0, nil, 100); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	for _, op := range []int{4, 5, 6, 7, 8, 9, 10} {
		rhs := uint64(0)
		if op >= 8 {
			rhs = 32
		}
		p, err := Parse(fixture([]int{32}, "B"+num(32)+num(0)+fixed(uint64(op), 2)+constant(1, 32)+constant(rhs, 32)+ret(num(0))))
		if err != nil {
			t.Fatal(err)
		}
		if _, err := p.Run(context.Background(), 0, nil, 100); !errors.Is(err, ErrRuntime) {
			t.Fatal(op, err)
		}
	}
	b := fixture([]int{32}, "B"+ret(num(0)))
	p, err = Parse(b)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := p.Run(context.Background(), 0, nil, 100); !errors.Is(err, ErrRuntime) {
		t.Fatal("uninitialized read", err)
	}
	for i := 0; i < len(b)-1; i++ {
		if _, err := Parse(b[:i]); err == nil {
			t.Fatalf("accepted truncation at %d", i)
		}
	}
}
func TestUnsupportedBytecodeRejectedBeforeExecution(t *testing.T) {
	for _, op := range []int{0, 33, 37, 51, 255} {
		b := fixture(nil, "BT"+fixed(uint64(op), 2))
		if _, err := Parse(b); !errors.Is(err, ErrUnsupported) {
			t.Fatal(op, err)
		}
	}
}
func FuzzBytecodeParserAndVM(f *testing.F) {
	f.Add(fixture(nil, "B"+ret(constant(42, 32))))
	f.Fuzz(func(t *testing.T, b []byte) {
		if len(b) > 65536 {
			return
		}
		p, err := Parse(b)
		if err == nil {
			_, _ = p.Run(context.Background(), 0, nil, 1000)
		}
	})
}
